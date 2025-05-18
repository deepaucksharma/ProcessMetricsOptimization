package prioritytagger

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

const (
	// Process metric attribute names
	processExecutableNameKey = "process.executable.name"
	processCPUUtilizationKey = "process.cpu.utilization"
	processMemoryRSSKey      = "process.memory.rss"
)

type priorityTaggerProcessor struct {
	config          *Config
	logger          *zap.Logger
	metricsConsumer consumer.Metrics
	obsrecv         *obsreportHelper
}

func newProcessor(config *Config, logger *zap.Logger, mexp consumer.Metrics, settings component.TelemetrySettings) (*priorityTaggerProcessor, error) {
	obsrecv, err := newObsreportHelper(settings)
	if err != nil {
		return nil, err
	}

	return &priorityTaggerProcessor{
		config:          config,
		logger:          logger,
		metricsConsumer: mexp,
		obsrecv:         obsrecv,
	}, nil
}

func (p *priorityTaggerProcessor) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (p *priorityTaggerProcessor) Shutdown(_ context.Context) error {
	return nil
}

func (p *priorityTaggerProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *priorityTaggerProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// Start the metrics observation
	ctx = p.obsrecv.StartMetricsOp(ctx)

	// Process metrics, check if they belong to critical processes and tag them
	processedCount := p.processMetrics(ctx, md)

	// Record the observation and the number of processed items
	p.obsrecv.EndMetricsOp(ctx, p.config.ProcessorType(), processedCount, nil)

	// Send the modified metrics to the next consumer
	return p.metricsConsumer.ConsumeMetrics(ctx, md)
}

func (p *priorityTaggerProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) int {
	// This count tracks the total number of data points we've processed
	processedCount := 0

	// Track which process IDs have been tagged as critical
	// to avoid counting the same process multiple times in our custom metric
	taggedProcesses := make(map[string]bool)

	// Iterate through the resource metrics
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)

		// Iterate through the scope metrics
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)

			// Iterate through each metric
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)

				// Process the datapoints based on metric type
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					pts := metric.Gauge().DataPoints()
					for l := 0; l < pts.Len(); l++ {
						if isCriticalProcess(pts.At(l).Attributes(), p.config) {
							markAsCritical(pts.At(l).Attributes(), p.config)

							// Record the process as tagged for the custom metric
							processID := getProcessID(pts.At(l).Attributes())
							if processID != "" && !taggedProcesses[processID] {
								taggedProcesses[processID] = true
								p.obsrecv.RecordTaggedProcess(ctx)
							}
						}
						processedCount++
					}
				case pmetric.MetricTypeSum:
					pts := metric.Sum().DataPoints()
					for l := 0; l < pts.Len(); l++ {
						if isCriticalProcess(pts.At(l).Attributes(), p.config) {
							markAsCritical(pts.At(l).Attributes(), p.config)

							// Record the process as tagged for the custom metric
							processID := getProcessID(pts.At(l).Attributes())
							if processID != "" && !taggedProcesses[processID] {
								taggedProcesses[processID] = true
								p.obsrecv.RecordTaggedProcess(ctx)
							}
						}
						processedCount++
					}
				case pmetric.MetricTypeHistogram:
					pts := metric.Histogram().DataPoints()
					for l := 0; l < pts.Len(); l++ {
						if isCriticalProcess(pts.At(l).Attributes(), p.config) {
							markAsCritical(pts.At(l).Attributes(), p.config)

							// Record the process as tagged for the custom metric
							processID := getProcessID(pts.At(l).Attributes())
							if processID != "" && !taggedProcesses[processID] {
								taggedProcesses[processID] = true
								p.obsrecv.RecordTaggedProcess(ctx)
							}
						}
						processedCount++
					}
				case pmetric.MetricTypeSummary:
					pts := metric.Summary().DataPoints()
					for l := 0; l < pts.Len(); l++ {
						if isCriticalProcess(pts.At(l).Attributes(), p.config) {
							markAsCritical(pts.At(l).Attributes(), p.config)

							// Record the process as tagged for the custom metric
							processID := getProcessID(pts.At(l).Attributes())
							if processID != "" && !taggedProcesses[processID] {
								taggedProcesses[processID] = true
								p.obsrecv.RecordTaggedProcess(ctx)
							}
						}
						processedCount++
					}
				case pmetric.MetricTypeExponentialHistogram:
					pts := metric.ExponentialHistogram().DataPoints()
					for l := 0; l < pts.Len(); l++ {
						if isCriticalProcess(pts.At(l).Attributes(), p.config) {
							markAsCritical(pts.At(l).Attributes(), p.config)

							// Record the process as tagged for the custom metric
							processID := getProcessID(pts.At(l).Attributes())
							if processID != "" && !taggedProcesses[processID] {
								taggedProcesses[processID] = true
								p.obsrecv.RecordTaggedProcess(ctx)
							}
						}
						processedCount++
					}
				}
			}
		}
	}

	p.logger.Debug("PriorityTagger processor processed metrics",
		zap.Int("processed_count", processedCount),
		zap.Int("tagged_processes", len(taggedProcesses)))

	return processedCount
}

// isCriticalProcess determines if a process is critical based on configuration criteria
func isCriticalProcess(attrs pcommon.Map, cfg *Config) bool {
	// Check if it's already tagged as critical
	if value, exists := attrs.Get(cfg.PriorityAttributeName); exists {
		if value.Str() == cfg.CriticalAttributeValue {
			return true
		}
	}

	// Check if executable name matches any in the critical list
	exeName, exists := attrs.Get(processExecutableNameKey)
	if !exists {
		return false
	}

	// Check direct name matches
	for _, name := range cfg.CriticalExecutables {
		if exeName.Str() == name {
			return true
		}
	}

	// Check regex patterns
	for _, pattern := range cfg.GetCompiledPatterns() {
		if pattern.MatchString(exeName.Str()) {
			return true
		}
	}

	// Check CPU threshold if enabled
	if cfg.CPUSteadyStateThreshold >= 0 {
		if cpuUtil, exists := attrs.Get(processCPUUtilizationKey); exists {
			if cpuUtil.Double() > cfg.CPUSteadyStateThreshold {
				return true
			}
		}
	}

	// Check Memory RSS threshold if enabled
	if cfg.MemoryRSSThresholdMiB >= 0 {
		if memRSS, exists := attrs.Get(processMemoryRSSKey); exists {
			// Convert to MiB for comparison (assuming value is in bytes)
			var memRSSMiB int64
			var validType bool = true
			switch memRSS.Type() {
			case pcommon.ValueTypeInt:
				memRSSMiB = memRSS.Int() / (1024 * 1024)
			case pcommon.ValueTypeDouble:
				memRSSMiB = int64(memRSS.Double() / (1024 * 1024))
			default:
				// Skip if not a numeric type
				validType = false
			}

			if validType && memRSSMiB > cfg.MemoryRSSThresholdMiB {
				return true
			}
		}
	}

	return false
}

// markAsCritical adds the priority tag to the process attributes
func markAsCritical(attrs pcommon.Map, cfg *Config) {
	attrs.PutStr(cfg.PriorityAttributeName, cfg.CriticalAttributeValue)
}

// getProcessID extracts a unique process identifier from attributes
// This is used to count unique tagged processes for the custom metric
func getProcessID(attrs pcommon.Map) string {
	// Try to get process.pid first
	if pid, exists := attrs.Get("process.pid"); exists {
		return pid.Str()
	}

	// Fall back to executable name if pid doesn't exist
	if exeName, exists := attrs.Get(processExecutableNameKey); exists {
		return exeName.Str()
	}

	return ""
}
