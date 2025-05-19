package reservoirsampler

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/newrelic/nrdot-process-optimization/internal/metricsutil"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

type reservoirSamplerProcessor struct {
	config       *Config
	logger       *zap.Logger
	nextConsumer consumer.Metrics
	obsrep       *reservoirSamplerObsreport

	// Reservoir state
	mu          sync.Mutex
	reservoir   map[string]bool // Stores unique process identities that are sampled
	streamCount int64           // Count of eligible items seen so far in the stream
	randSource  *rand.Rand      // For sampling algorithm
}

func newReservoirSamplerProcessor(settings processor.CreateSettings, next consumer.Metrics, cfg *Config) (*reservoirSamplerProcessor, error) {
	obsrep, err := newReservoirSamplerObsreport(settings.TelemetrySettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create obsreport for reservoirsampler: %w", err)
	}

	// Seed random source with current time for proper randomness in production
	rs := rand.NewSource(time.Now().UnixNano())

	return &reservoirSamplerProcessor{
		config:       cfg,
		logger:       settings.Logger,
		nextConsumer: next,
		obsrep:       obsrep,
		reservoir:    make(map[string]bool),
		randSource:   rand.New(rs),
		streamCount:  0,
	}, nil
}

func (p *reservoirSamplerProcessor) Start(_ context.Context, _ component.Host) error { return nil }
func (p *reservoirSamplerProcessor) Shutdown(_ context.Context) error                { return nil }
func (p *reservoirSamplerProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

// generateIdentity creates a unique string for a process based on configured attributes.
func (p *reservoirSamplerProcessor) generateIdentity(attrs pcommon.Map) (string, bool) {
	var identityParts []string
	for _, key := range p.config.IdentityAttributes {
		val, exists := attrs.Get(key)
		if !exists {
			return "", false // Cannot form identity if a key is missing
		}
		identityParts = append(identityParts, key+"="+val.AsString())
	}
	sort.Strings(identityParts) // Ensure consistent order for the hash

	// Hash the identity parts for consistency and to handle arbitrary attribute values
	h := sha256.New()
	h.Write([]byte(strings.Join(identityParts, ";")))
	return hex.EncodeToString(h.Sum(nil)), true
}

func (p *reservoirSamplerProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	ctx = p.obsrep.StartMetricsOp(ctx)
	numOriginalMetricPoints := metricsutil.CountPoints(md)

	// First pass: identify eligible DPs and perform sampling logic for their identities
	// We need to collect all unique eligible identities before modification
	eligibleIdentities := make(map[string]bool)

	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				var dps pmetric.NumberDataPointSlice
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					dps = metric.Gauge().DataPoints()
				case pmetric.MetricTypeSum:
					dps = metric.Sum().DataPoints()
				default:
					continue
				}

				for l := 0; l < dps.Len(); l++ {
					dp := dps.At(l)
					attrs := dp.Attributes()

					// Check if critical (skip critical processes)
					if prioVal, prioExists := attrs.Get(p.config.PriorityAttributeName); prioExists && prioVal.Str() == p.config.CriticalAttributeValue {
						continue // Skip critical
					}
					// Check if TopK (hypothetical, would skip TopK processes)
					// if topKVal, topKExists := attrs.Get(p.config.TopKAttributeName); topKExists {
					//     continue // Skip TopK
					// }

					identity, canIdentify := p.generateIdentity(attrs)
					if !canIdentify {
						continue // Cannot sample if identity cannot be formed
					}

					// Record the eligible identity for sampling
					eligibleIdentities[identity] = true
				}
			}
		}
	}

	// Process all eligible identities found in this batch
	for identity := range eligibleIdentities {
		// Is this identity new? (not yet seen in stream)
		if _, exists := p.reservoir[identity]; !exists {
			p.streamCount++                            // This is a new eligible identity in the logical stream
			p.obsrep.recordEligibleIdentitiesSeen(ctx) // Count unique eligible identities

			// Algorithm R variant for reservoir sampling
			if len(p.reservoir) < p.config.ReservoirSize {
				// Reservoir not full yet, add the identity
				p.reservoir[identity] = true
				p.obsrep.recordNewIdentityAdded(ctx)
			} else {
				// Reservoir is full, use random replacement
				// Choose a random number j from 0 to streamCount-1
				j := p.randSource.Int63n(p.streamCount)

				// If j < reservoirSize, replace an item at random
				if j < int64(p.config.ReservoirSize) {
					// Select a random item to replace
					reservoirItems := make([]string, 0, len(p.reservoir))
					for key := range p.reservoir {
						reservoirItems = append(reservoirItems, key)
					}
					idxToReplace := p.randSource.Intn(len(reservoirItems))
					delete(p.reservoir, reservoirItems[idxToReplace])

					// Add the new identity
					p.reservoir[identity] = true
					p.obsrep.recordNewIdentityAdded(ctx)
				}
			}
		}
	}

	// Update obsreport metrics
	currentSelectedCount := int64(len(p.reservoir))
	p.obsrep.recordSelectedIdentitiesCount(ctx, currentSelectedCount)

	fillRatio := 0.0
	if p.config.ReservoirSize > 0 {
		fillRatio = float64(currentSelectedCount) / float64(p.config.ReservoirSize)
	}
	p.obsrep.recordReservoirFillRatio(ctx, fillRatio)

	// Calculate sample rate
	sampleRate := 0.0
	if p.streamCount > 0 && len(p.reservoir) > 0 {
		sampleRate = float64(len(p.reservoir)) / float64(p.streamCount)
	}

	// Second pass: filter metrics (pass through critical & sampled, drop non-sampled eligible)
	newMd := pmetric.NewMetrics()
	md.ResourceMetrics().CopyTo(newMd.ResourceMetrics()) // Start with a copy

	newMd.ResourceMetrics().RemoveIf(func(rm pmetric.ResourceMetrics) bool {
		rm.ScopeMetrics().RemoveIf(func(sm pmetric.ScopeMetrics) bool {
			sm.Metrics().RemoveIf(func(metric pmetric.Metric) bool {
				var dps pmetric.NumberDataPointSlice
				isModifiableType := true
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					dps = metric.Gauge().DataPoints()
				case pmetric.MetricTypeSum:
					dps = metric.Sum().DataPoints()
				default:
					isModifiableType = false // Cannot easily filter/tag other types
				}

				if !isModifiableType {
					return false // Pass through unhandled metric types
				}

				dps.RemoveIf(func(dp pmetric.NumberDataPoint) bool {
					attrs := dp.Attributes()

					// Priority pass-through (critical processes)
					if prioVal, prioExists := attrs.Get(p.config.PriorityAttributeName); prioExists && prioVal.Str() == p.config.CriticalAttributeValue {
						return false // Keep
					}
					// TopK pass-through (hypothetical)
					// if topKVal, topKExists := attrs.Get(p.config.TopKAttributeName); topKExists {
					//     return false // Keep
					// }

					identity, canIdentify := p.generateIdentity(attrs)
					if !canIdentify {
						return true // Drop if cannot identify
					}

					if p.reservoir[identity] { // Is this identity in our current sample?
						// Tag it as sampled and add sample rate
						attrs.PutStr(p.config.SampledAttributeName, p.config.SampledAttributeValue)
						attrs.PutDouble(p.config.SampleRateAttributeName, sampleRate)
						return false // Keep
					}
					return true // Drop (eligible but not sampled)
				})
				return dps.Len() == 0 // Remove metric if all its DPs were dropped
			})
			return sm.Metrics().Len() == 0 // Remove scope if all its metrics were removed
		})
		return rm.ScopeMetrics().Len() == 0 // Remove resource if all its scopes were removed
	})

	numProcessedMetricPoints := metricsutil.CountPoints(newMd)
	numDroppedMetricPoints := numOriginalMetricPoints - numProcessedMetricPoints
	p.obsrep.EndMetricsOp(ctx, p.config.ProcessorType(), numProcessedMetricPoints, numDroppedMetricPoints, nil)

	if newMd.ResourceMetrics().Len() == 0 {
		p.logger.Debug("All metrics were dropped by reservoir sampler, resulting in empty batch.")
		return nil
	}
	return p.nextConsumer.ConsumeMetrics(ctx, newMd)
}
