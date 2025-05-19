package othersrollup

import (
	"context"
	"fmt"
	"time"

	"github.com/newrelic/nrdot-process-optimization/internal/metricsutil"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/zap"
)

const (
	processPIDKey            = "process.pid"
	processExecutableNameKey = "process.executable.name"
)

type othersRollupProcessor struct {
	config       *Config
	logger       *zap.Logger
	nextConsumer consumer.Metrics
	obsrep       *othersRollupObsreport
}

// AggregationState holds the running sum and count for averaging.
type AggregationState struct {
	Sum   float64
	Count int64
	Type  AggregationType
}

func newOthersRollupProcessor(settings processor.CreateSettings, next consumer.Metrics, cfg *Config) (*othersRollupProcessor, error) {
	obsrep, err := newOthersRollupObsreport(settings.TelemetrySettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create obsreport for othersrollup processor: %w", err)
	}
	return &othersRollupProcessor{
		config:       cfg,
		logger:       settings.Logger,
		nextConsumer: next,
		obsrep:       obsrep,
	}, nil
}

func (p *othersRollupProcessor) Start(_ context.Context, _ component.Host) error { return nil }
func (p *othersRollupProcessor) Shutdown(_ context.Context) error                { return nil }
func (p *othersRollupProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *othersRollupProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	ctx = p.obsrep.StartMetricsOp(ctx)
	originalMetricPointCount := metricsutil.CountPoints(md)

	// This map will store aggregated values: resourceKey -> metric_name -> AggregationState
	rollupData := make(map[string]map[string]*AggregationState)
	// This map tracks which original PIDs were part of the rollup for a given metric name under a resource
	rolledUpPIDsPerMetric := make(map[string]map[string]map[string]bool) // resourceKey -> metricName -> PID -> true

	newMetrics := pmetric.NewMetrics() // Holds pass-through and new rolled-up metrics

	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		newRm := newMetrics.ResourceMetrics().AppendEmpty() // Create a new RM for the output
		rm.Resource().CopyTo(newRm.Resource())
		// Generate a resource key based on attributes
		resourceKey := resourceAttributesToString(rm.Resource().Attributes())

		// Initialize maps for the current resource if they don't exist
		if _, ok := rollupData[resourceKey]; !ok {
			rollupData[resourceKey] = make(map[string]*AggregationState)
		}
		if _, ok := rolledUpPIDsPerMetric[resourceKey]; !ok {
			rolledUpPIDsPerMetric[resourceKey] = make(map[string]map[string]bool)
		}

		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			newSm := newRm.ScopeMetrics().AppendEmpty() // Create a new SM for the output
			sm.Scope().CopyTo(newSm.Scope())

			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				// We create a new metric in newSm later if it has pass-through DPs or rolled-up DPs.
				// This avoids creating empty metrics if all DPs are rolled up and the metric itself doesn't get a rollup.

				// Function to determine if a metric datapoint should be rolled up
				shouldRollupDP := func(attrs pcommon.Map) bool {
					if prioVal, prioExists := attrs.Get(p.config.PriorityAttributeName); prioExists && prioVal.Str() == p.config.CriticalAttributeValue {
						return false
					}
					if len(p.config.MetricsToRollup) > 0 {
						isTargeted := false
						for _, mtr := range p.config.MetricsToRollup {
							if metric.Name() == mtr {
								isTargeted = true
								break
							}
						}
						if !isTargeted {
							return false // Not in the explicit list of metrics to rollup
						}
					}
					// If MetricsToRollup is empty, all non-priority DPs are candidates.
					// If MetricsToRollup is not empty, only those in the list are candidates.
					return true
				}

				passThroughDps := pmetric.NewNumberDataPointSlice() // Collect DPs that are not rolled up for this metric

				// Process and filter datapoints
				processDataPoints := func(originalDps pmetric.NumberDataPointSlice) {
					for l := 0; l < originalDps.Len(); l++ {
						dp := originalDps.At(l)
						if shouldRollupDP(dp.Attributes()) {
							metricName := metric.Name()
							aggState, exists := rollupData[resourceKey][metricName]
							if !exists {
								aggType := SumAggregation // Default aggregation type
								if configuredType, ok := p.config.Aggregations[metricName]; ok {
									aggType = configuredType
								} else if metric.Type() == pmetric.MetricTypeGauge { // Default for Gauge if not specified
									aggType = AvgAggregation
								}
								aggState = &AggregationState{Type: aggType}
								rollupData[resourceKey][metricName] = aggState
							}

							val := getNumericValue(dp)
							aggState.Sum += val
							aggState.Count++

							if pidVal, pidExists := dp.Attributes().Get(processPIDKey); pidExists {
								if _, ok := rolledUpPIDsPerMetric[resourceKey][metricName]; !ok {
									rolledUpPIDsPerMetric[resourceKey][metricName] = make(map[string]bool)
								}
								rolledUpPIDsPerMetric[resourceKey][metricName][pidVal.Str()] = true
							}
						} else {
							dp.CopyTo(passThroughDps.AppendEmpty()) // Add to pass-through slice
						}
					}
				}

				// Handle different metric types for collecting pass-through DPs and identifying roll-up candidates
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					processDataPoints(metric.Gauge().DataPoints())
					if passThroughDps.Len() > 0 {
						m := newSm.Metrics().AppendEmpty()
						metric.CopyTo(m)                                                                              // Copy metric metadata
						m.SetEmptyGauge().DataPoints().RemoveIf(func(_ pmetric.NumberDataPoint) bool { return true }) // Clear existing
						passThroughDps.CopyTo(m.Gauge().DataPoints())
					}
				case pmetric.MetricTypeSum:
					processDataPoints(metric.Sum().DataPoints())
					if passThroughDps.Len() > 0 {
						m := newSm.Metrics().AppendEmpty()
						metric.CopyTo(m)                                                                            // Copy metric metadata
						m.SetEmptySum().DataPoints().RemoveIf(func(_ pmetric.NumberDataPoint) bool { return true }) // Clear existing
						passThroughDps.CopyTo(m.Sum().DataPoints())
					}
				default:
					// For other metric types (Histogram, Summary, etc.), pass them through unchanged
					isTargetedForRollup := false
					if len(p.config.MetricsToRollup) > 0 {
						for _, mtr := range p.config.MetricsToRollup {
							if metric.Name() == mtr {
								isTargetedForRollup = true
								break
							}
						}
					} else { // If MetricsToRollup is empty, all are candidates if not priority
						isTargetedForRollup = true // but only gauge/sum are currently aggregated
					}

					if !isTargetedForRollup || (metric.Type() != pmetric.MetricTypeGauge && metric.Type() != pmetric.MetricTypeSum) {
						metric.CopyTo(newSm.Metrics().AppendEmpty()) // Pass through the whole metric
					}
				}
			} // End iterating original metrics (k loop)

			// Now, add the new rolled-up metrics to newSm for this resourceKey
			if resRollupData, ok := rollupData[resourceKey]; ok {
				aggregatedSeriesGenerated := int64(0)
				totalInputSeriesRolledUp := int64(0)

				for metricName, aggState := range resRollupData {
					if aggState.Count == 0 {
						continue
					}
					aggregatedSeriesGenerated++
					if pidsMap, found := rolledUpPIDsPerMetric[resourceKey][metricName]; found {
						totalInputSeriesRolledUp += int64(len(pidsMap))
					}

					rolledUpMetric := newSm.Metrics().AppendEmpty()
					rolledUpMetric.SetName(metricName)
					// Find original metric to copy metadata (Unit, Description, etc.)
					var originalMetricMetadata pmetric.Metric
					for k := 0; k < sm.Metrics().Len(); k++ {
						if sm.Metrics().At(k).Name() == metricName {
							originalMetricMetadata = sm.Metrics().At(k)
							break
						}
					}
					// Set metadata from original metric
					rolledUpMetric.SetDescription(originalMetricMetadata.Description())
					rolledUpMetric.SetUnit(originalMetricMetadata.Unit())

					newDp := rolledUpMetric.SetEmptyGauge().DataPoints().AppendEmpty() // Default to Gauge for rolled-up
					// If original was Sum and aggregation is Sum, new should be Sum
					// Check if the original metric was a Sum and we're using sum aggregation
					if originalMetricMetadata.Type() == pmetric.MetricTypeSum && aggState.Type == SumAggregation {
						rolledUpMetric.SetEmptySum().SetIsMonotonic(originalMetricMetadata.Sum().IsMonotonic())
						rolledUpMetric.Sum().SetAggregationTemporality(originalMetricMetadata.Sum().AggregationTemporality())
						newDp = rolledUpMetric.Sum().DataPoints().AppendEmpty()
					} else {
						// For Avg or if original was Gauge, new is Gauge
						rolledUpMetric.SetEmptyGauge()
						newDp = rolledUpMetric.Gauge().DataPoints().AppendEmpty()
					}

					if aggState.Type == AvgAggregation && aggState.Count > 0 {
						newDp.SetDoubleValue(aggState.Sum / float64(aggState.Count))
					} else { // Default to Sum
						newDp.SetDoubleValue(aggState.Sum)
					}
					newDp.SetTimestamp(pcommon.NewTimestampFromTime(time.Now())) // Or use latest timestamp from rolled DPs
					newDp.Attributes().PutStr(processPIDKey, p.config.OutputPIDAttributeValue)
					newDp.Attributes().PutStr(processExecutableNameKey, p.config.OutputExecutableNameAttributeValue)
					// Copy other non-process specific attributes from resource if needed
				}
				p.obsrep.recordAggregatedSeries(ctx, aggregatedSeriesGenerated)
				p.obsrep.recordInputSeriesRolledUp(ctx, totalInputSeriesRolledUp)
			}
			// If newSm has no metrics after filtering and rollup, it will be removed by the later check
			if newSm.Metrics().Len() == 0 {
				newRm.ScopeMetrics().RemoveIf(func(s pmetric.ScopeMetrics) bool {
					return s.Metrics().Len() == 0 // A simple check for empty scope
				})
			}
		} // End iterating scope metrics (j loop)
		if newRm.ScopeMetrics().Len() == 0 {
			newMetrics.ResourceMetrics().RemoveIf(func(r pmetric.ResourceMetrics) bool {
				return r.ScopeMetrics().Len() == 0 // Remove empty resource metrics
			})
		}
	} // End iterating resource metrics (i loop)

	finalMetricPointCount := metricsutil.CountPoints(newMetrics)
	droppedCount := originalMetricPointCount - finalMetricPointCount
	p.obsrep.EndMetricsOp(ctx, p.config.ProcessorType(), finalMetricPointCount, droppedCount, nil)

	if newMetrics.ResourceMetrics().Len() == 0 {
		p.logger.Debug("All metrics were rolled up or dropped, resulting in empty batch.")
		return nil // Consume an empty metrics struct, effectively dropping all
	}

	return p.nextConsumer.ConsumeMetrics(ctx, newMetrics)
}

func getNumericValue(dp pmetric.NumberDataPoint) float64 {
	switch dp.ValueType() {
	case pmetric.NumberDataPointValueTypeInt:
		return float64(dp.IntValue())
	case pmetric.NumberDataPointValueTypeDouble:
		return dp.DoubleValue()
	}
	return 0
}

// resourceAttributesToString converts resource attributes to a string for use as a map key
func resourceAttributesToString(attrs pcommon.Map) string {
	if attrs.Len() == 0 {
		return "empty_resource"
	}

	// Simple implementation - concatenate key/value pairs
	result := ""
	attrs.Range(func(k string, v pcommon.Value) bool {
		result += k + ":" + v.AsString() + ";"
		return true
	})

	return result
}
