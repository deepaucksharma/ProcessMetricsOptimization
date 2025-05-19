package adaptivetopk

import (
	"container/heap"
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
	processExecutableNameKey = "process.executable.name"
	processPIDKey            = "process.pid" // Assuming PID is available for unique identification
)

// processInfo holds data for ranking processes
type processInfo struct {
	pid            string  // Unique process identifier (e.g., PID or command line hash)
	metricValue    float64 // Primary metric value for ranking
	secondaryValue float64 // Secondary metric value for tie-breaking
	attributes     pcommon.Map
	isCritical     bool
	index          int // For heap interface
}

// processHeap implements heap.Interface for processInfo
type processHeap []*processInfo

func (ph processHeap) Len() int { return len(ph) }
func (ph processHeap) Less(i, j int) bool {
	// Min-heap: we want to pop the smallest element
	// For TopK largest, this means if element i is smaller than j, it has lower priority to stay in heap
	if ph[i].metricValue == ph[j].metricValue {
		// Tie-breaking logic using secondary metric
		if ph[i].secondaryValue == ph[j].secondaryValue {
			return ph[i].pid < ph[j].pid // Consistent tie-breaking by PID
		}
		return ph[i].secondaryValue < ph[j].secondaryValue
	}
	return ph[i].metricValue < ph[j].metricValue
}
func (ph processHeap) Swap(i, j int) {
	ph[i], ph[j] = ph[j], ph[i]
	ph[i].index = i
	ph[j].index = j
}
func (ph *processHeap) Push(x any) {
	n := len(*ph)
	item := x.(*processInfo)
	item.index = n
	*ph = append(*ph, item)
}
func (ph *processHeap) Pop() any {
	old := *ph
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*ph = old[0 : n-1]
	return item
}

type adaptiveTopKProcessor struct {
	config       *Config
	logger       *zap.Logger
	nextConsumer consumer.Metrics
	obsrep       *adaptiveTopKObsreport

	// --- State for Dynamic K & Hysteresis (Sub-Phase 2b) ---
	currentDynamicK       int
	processHysteresis     map[string]time.Time // processID -> expiryTime
	lastHysteresisCleanup time.Time            // Track when we last did a full cleanup
}

func newAdaptiveTopKProcessor(settings processor.CreateSettings, next consumer.Metrics, cfg *Config) (*adaptiveTopKProcessor, error) {
	obsrep, err := newAdaptiveTopKObsreport(settings.TelemetrySettings)
	if err != nil {
		return nil, fmt.Errorf("failed to create obsreport for adaptivetopk processor: %w", err)
	}
	p := &adaptiveTopKProcessor{
		config:       cfg,
		logger:       settings.Logger,
		nextConsumer: next,
		obsrep:       obsrep,
	}
	// Initialize hysteresis map and set initial cleanup time
	p.processHysteresis = make(map[string]time.Time)
	p.lastHysteresisCleanup = time.Now()

	// Set initial dynamic K value
	if cfg.IsDynamicK() {
		p.currentDynamicK = cfg.MinKValue // Initial K
	} else {
		p.currentDynamicK = cfg.KValue // Use fixed K as initial dynamic K
	}
	p.obsrep.recordCurrentKValue(context.Background(), int64(p.currentDynamicK))
	return p, nil
}

func (p *adaptiveTopKProcessor) Start(_ context.Context, _ component.Host) error { return nil }
func (p *adaptiveTopKProcessor) Shutdown(_ context.Context) error                { return nil }
func (p *adaptiveTopKProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *adaptiveTopKProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	ctx = p.obsrep.StartMetricsOp(ctx)
	numOriginalMetricPoints := metricsutil.CountPoints(md)

	// Determine current K value (fixed or dynamic)
	currentK := p.config.KValue
	if p.config.IsDynamicK() {
		// Dynamic K calculation based on host metrics
		hostLoad := p.findHostLoadMetric(md)
		if hostLoad >= 0 {
			if p.updateDynamicK(hostLoad) {
				p.obsrep.recordCurrentKValue(ctx, int64(p.currentDynamicK))
			}
		}
		currentK = p.currentDynamicK
	}

	// Estimate process count for pre-allocation
	// This helps reduce map resizing and improve performance
	estimatedProcessCount := estimateProcessCount(md)

	// Collect all processInfos from the batch with pre-allocated capacity
	allProcesses := make(map[string]*processInfo, estimatedProcessCount) // PID -> processInfo

	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			// Process metrics in two passes - first find critical processes and metric values
			// This reduces attribute lookups by checking only relevant metrics

			// First pass - extract the key metrics we need for ranking
			keyMetricName := p.config.KeyMetricName
			secondaryKeyMetricName := p.config.SecondaryKeyMetricName

			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				metricName := metric.Name()

				// Skip metrics that aren't used for ranking or priority
				if metricName != keyMetricName && metricName != secondaryKeyMetricName {
					continue
				}

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
					pidVal, pidExists := attrs.Get(processPIDKey)
					if !pidExists {
						continue // Skip data points not identifiable by PID
					}
					pid := pidVal.Str()

					// Get or create process info
					proc, exists := allProcesses[pid]
					if !exists {
						proc = &processInfo{
							pid:            pid,
							attributes:     pcommon.NewMap(),
							metricValue:    0,
							secondaryValue: 0,
						}
						dp.Attributes().CopyTo(proc.attributes) // Store all attributes
						allProcesses[pid] = proc
					}

					// Check for critical tag
					critVal, critExists := attrs.Get(p.config.PriorityAttributeName)
					if critExists && critVal.Str() == p.config.CriticalAttributeValue {
						proc.isCritical = true
					}

					// Update metricValue if this is the key ranking metric
					if metricName == keyMetricName {
						proc.metricValue = getNumericValue(dp)
					} else if metricName == secondaryKeyMetricName {
						// Update secondaryValue if this is the secondary ranking metric
						proc.secondaryValue = getNumericValue(dp)
					}
				}
			}
		}
	}

	// Identify critical processes and Top K non-critical processes
	// Pre-allocate maps and slices based on the number of processes
	processCount := len(allProcesses)
	selectedPIDs := make(map[string]bool, processCount)
	nonCriticalProcs := make([]*processInfo, 0, processCount)

	for _, proc := range allProcesses {
		if proc.isCritical {
			selectedPIDs[proc.pid] = true
		} else {
			nonCriticalProcs = append(nonCriticalProcs, proc)
		}
	}

	// Pre-allocate the heap with exactly K capacity when possible
	// This avoids heap resizing during the topK selection
	heapCapacity := currentK
	if len(nonCriticalProcs) < currentK {
		heapCapacity = len(nonCriticalProcs)
	}
	topKHeap := make(processHeap, 0, heapCapacity)
	heap.Init(&topKHeap)

	var topKCount int64

	// Optimization for the case where we have fewer processes than K
	if len(nonCriticalProcs) <= currentK {
		// If we have fewer or equal processes than K, simply include all of them
		for _, proc := range nonCriticalProcs {
			selectedPIDs[proc.pid] = true
		}
		topKCount = int64(len(nonCriticalProcs))
	} else {
		// Use a min-heap to find Top K non-critical processes
		for _, proc := range nonCriticalProcs {
			if topKHeap.Len() < currentK {
				heap.Push(&topKHeap, proc)
			} else if topKHeap.Len() > 0 && proc.metricValue > topKHeap[0].metricValue {
				// Process has higher priority than the lowest one in the heap
				heap.Pop(&topKHeap)
				heap.Push(&topKHeap, proc)
			}
		}

		// Extract the top K processes from the heap
		count := 0
		for topKHeap.Len() > 0 {
			proc := heap.Pop(&topKHeap).(*processInfo)
			selectedPIDs[proc.pid] = true
			count++
		}
		topKCount = int64(count)
	}

	// Record metrics
	p.obsrep.recordTopKProcessesSelected(ctx, topKCount)

	// Apply hysteresis to processes if configured
	if p.config.IsDynamicK() && p.config.HysteresisDuration > 0 {
		p.applyProcessHysteresis(selectedPIDs, allProcesses)
	}

	// Filter the original metrics
	filteredMd := pmetric.NewMetrics()
	md.ResourceMetrics().CopyTo(filteredMd.ResourceMetrics())

	// Remove metrics that don't belong to critical processes or topK processes
	filteredMd.ResourceMetrics().RemoveIf(func(rm pmetric.ResourceMetrics) bool {
		rm.ScopeMetrics().RemoveIf(func(sm pmetric.ScopeMetrics) bool {
			sm.Metrics().RemoveIf(func(metric pmetric.Metric) bool {
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					metric.Gauge().DataPoints().RemoveIf(func(dp pmetric.NumberDataPoint) bool {
						pidVal, pidExists := dp.Attributes().Get(processPIDKey)
						return !pidExists || !selectedPIDs[pidVal.Str()]
					})
					return metric.Gauge().DataPoints().Len() == 0
				case pmetric.MetricTypeSum:
					metric.Sum().DataPoints().RemoveIf(func(dp pmetric.NumberDataPoint) bool {
						pidVal, pidExists := dp.Attributes().Get(processPIDKey)
						return !pidExists || !selectedPIDs[pidVal.Str()]
					})
					return metric.Sum().DataPoints().Len() == 0
				default:
					// For other metric types, pass them through unchanged
					return false
				}
			})
			return sm.Metrics().Len() == 0
		})
		return rm.ScopeMetrics().Len() == 0
	})

	numProcessedMetricPoints := metricsutil.CountPoints(filteredMd)
	numDroppedMetricPoints := numOriginalMetricPoints - numProcessedMetricPoints
	p.obsrep.EndMetricsOp(ctx, p.config.ProcessorType(), numProcessedMetricPoints, numDroppedMetricPoints, nil)

	return p.nextConsumer.ConsumeMetrics(ctx, filteredMd)
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

// findHostLoadMetric searches for the configured host load metric in the metrics batch
// with early exit as soon as the metric is found
func (p *adaptiveTopKProcessor) findHostLoadMetric(md pmetric.Metrics) float64 {
	if !p.config.IsDynamicK() || p.config.HostLoadMetricName == "" {
		return -1.0 // Dynamic K not configured
	}

	targetMetric := p.config.HostLoadMetricName

	// Efficient search with early exit
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)

			// Use binary search if scope metrics are sufficiently large
			if sm.Metrics().Len() > 20 {
				// Create a quick index for faster lookups in large batches
				idx := -1
				for k := 0; k < sm.Metrics().Len(); k++ {
					if sm.Metrics().At(k).Name() == targetMetric {
						idx = k
						break
					}
				}

				if idx >= 0 {
					metric := sm.Metrics().At(idx)
					var dps pmetric.NumberDataPointSlice
					switch metric.Type() {
					case pmetric.MetricTypeGauge:
						dps = metric.Gauge().DataPoints()
					case pmetric.MetricTypeSum:
						dps = metric.Sum().DataPoints()
					default:
						continue
					}

					if dps.Len() > 0 {
						return getNumericValue(dps.At(0))
					}
				}
			} else {
				// Linear search for small metric collections
				for k := 0; k < sm.Metrics().Len(); k++ {
					metric := sm.Metrics().At(k)
					if metric.Name() == targetMetric {
						var dps pmetric.NumberDataPointSlice
						switch metric.Type() {
						case pmetric.MetricTypeGauge:
							dps = metric.Gauge().DataPoints()
						case pmetric.MetricTypeSum:
							dps = metric.Sum().DataPoints()
						default:
							continue
						}

						// For simplicity, we use the first datapoint found
						// In a real system, you might want to aggregate multiple datapoints
						if dps.Len() > 0 {
							return getNumericValue(dps.At(0))
						}
					}
				}
			}
		}
	}

	return -1.0 // Metric not found
}

// updateDynamicK updates the current K value based on host load
func (p *adaptiveTopKProcessor) updateDynamicK(hostLoad float64) bool {
	// Find the appropriate K value based on load bands
	newK := p.config.MinKValue // Default to minimum
	highestThreshold := -1.0

	// Find the highest threshold that's less than or equal to the current load
	for threshold, kValue := range p.config.LoadBandsToKMap {
		// Fix: Use threshold <= hostLoad to correctly match when load equals a threshold
		if threshold <= hostLoad && threshold > highestThreshold {
			highestThreshold = threshold
			newK = kValue
		}
	}

	// Ensure K is within configured bounds
	if newK < p.config.MinKValue {
		newK = p.config.MinKValue
	} else if newK > p.config.MaxKValue {
		newK = p.config.MaxKValue
	}

	changed := newK != p.currentDynamicK
	if changed && p.logger != nil {
		p.logger.Info("Dynamic K adjusted",
			zap.Float64("hostLoad", hostLoad),
			zap.Int("previousK", p.currentDynamicK),
			zap.Int("newK", newK))
	}

	p.currentDynamicK = newK
	return changed
}

// applyProcessHysteresis applies hysteresis to process selection
// by keeping processes in the selectedPIDs map even after they fall out of the top K,
// until their hysteresis period expires
func (p *adaptiveTopKProcessor) applyProcessHysteresis(selectedPIDs map[string]bool, allProcesses map[string]*processInfo) {
	now := time.Now()

	// Basic cleanup - remove only expired entries
	for pid, expiryTime := range p.processHysteresis {
		if now.After(expiryTime) {
			delete(p.processHysteresis, pid)
		}
	}

	// Perform a more thorough cleanup every 5 minutes
	// This removes processes that no longer exist in the metrics, preventing memory leaks
	const fullCleanupInterval = 5 * time.Minute
	if now.Sub(p.lastHysteresisCleanup) > fullCleanupInterval {
		// Remove any hysteresis entries for processes that no longer exist in allProcesses
		for pid := range p.processHysteresis {
			if _, exists := allProcesses[pid]; !exists {
				delete(p.processHysteresis, pid)
			}
		}

		p.lastHysteresisCleanup = now

		if p.logger != nil {
			p.logger.Debug("Performed full hysteresis map cleanup",
				zap.Int("remainingEntriesCount", len(p.processHysteresis)))
		}
	}

	// For processes currently selected, update or add their expiry time
	for pid := range selectedPIDs {
		p.processHysteresis[pid] = now.Add(p.config.HysteresisDuration)
	}

	// Add processes still in their hysteresis period to the selected PIDs
	hysteresisCount := 0
	for pid, expiryTime := range p.processHysteresis {
		// Fix: Make sure we check if the process exists in allProcesses before applying hysteresis
		// This ensures we don't omit processes that should be included due to hysteresis
		if _, exists := allProcesses[pid]; exists && !selectedPIDs[pid] && now.Before(expiryTime) {
			selectedPIDs[pid] = true
			hysteresisCount++
		}
	}

	if hysteresisCount > 0 && p.logger != nil {
		p.logger.Debug("Applied process hysteresis",
			zap.Int("hysteresisProcessCount", hysteresisCount),
			zap.Int("totalSelectedProcesses", len(selectedPIDs)))
	}
}

// estimateProcessCount estimates the number of unique processes in a metrics batch
// by counting unique PIDs in the key metric's data points
func estimateProcessCount(md pmetric.Metrics) int {
	// Default estimate if we can't determine better
	defaultEstimate := 100

	// Quick scan for unique PIDs in the first resource metrics
	if md.ResourceMetrics().Len() == 0 {
		return defaultEstimate
	}

	// Keep track of unique PIDs
	uniquePIDs := make(map[string]struct{})

	// Scan first few resource metrics to get a reasonable estimate
	maxResourcesToScan := 2
	if md.ResourceMetrics().Len() < maxResourcesToScan {
		maxResourcesToScan = md.ResourceMetrics().Len()
	}

	for i := 0; i < maxResourcesToScan; i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			// Only scan a subset of metrics to save time
			metricsToScan := 10
			if sm.Metrics().Len() < metricsToScan {
				metricsToScan = sm.Metrics().Len()
			}

			for k := 0; k < metricsToScan; k++ {
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
					pidVal, exists := dps.At(l).Attributes().Get(processPIDKey)
					if exists {
						uniquePIDs[pidVal.Str()] = struct{}{}
					}
				}
			}
		}
	}

	// If we found unique PIDs, use that count with small buffer for additional processes
	if len(uniquePIDs) > 0 {
		// Add 20% buffer to account for new processes
		return int(float64(len(uniquePIDs)) * 1.2)
	}

	return defaultEstimate
}
