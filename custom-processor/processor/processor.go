package traceawareprocessor

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type traceAwareProcessor struct {
	logger       *zap.Logger
	config       *Config
	nextConsumer consumer.Metrics
	startTime    time.Time
}

func newTraceAwareProcessor(logger *zap.Logger, cfg *Config, nextConsumer consumer.Metrics) (*traceAwareProcessor, error) {
	return &traceAwareProcessor{
		logger:       logger,
		config:       cfg,
		nextConsumer: nextConsumer,
		startTime:    time.Now(),
	}, nil
}

// Capabilities returns the capabilities of the processor.
func (p *traceAwareProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

// Start is called when the processor is starting.
func (p *traceAwareProcessor) Start(_ context.Context, _ component.Host) error {
	p.logger.Info("Trace-Aware processor started",
		zap.String("profile", p.config.Profile))
	return nil
}

// Shutdown is called when the processor is shutting down.
func (p *traceAwareProcessor) Shutdown(_ context.Context) error {
	p.logger.Info("Trace-Aware processor stopped",
		zap.String("profile", p.config.Profile),
		zap.Duration("uptime", time.Since(p.startTime)))
	return nil
}

// ConsumeMetrics processes the metrics.
func (p *traceAwareProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// Count metrics for logging
	totalDataPoints := countDataPoints(md)
	
	p.logger.Debug("Processing metrics",
		zap.String("profile", p.config.Profile),
		zap.Int("metrics_count", md.ResourceMetrics().Len()),
		zap.Int("datapoints", totalDataPoints))
	
	// Add profile-specific attributes to all metrics
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		
		// Add profile-aware resource attributes
		rm.Resource().Attributes().PutStr("collector.pipeline", p.config.Profile)
		rm.Resource().Attributes().PutStr("processor.traceaware.version", "1.0.0")
		
		// Generate unique entity identifier
		hostname, err := os.Hostname()
		if err != nil {
			p.logger.Warn("Failed to get hostname, using 'unknown' as fallback", zap.Error(err))
			hostname = "unknown"
		}
		entityID := fmt.Sprintf("%s-%s", hostname, p.config.Profile)
		rm.Resource().Attributes().PutStr("entity.id", entityID)
		
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				
				// Add processor identification to metrics based on type
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					addAttributeToGauge(metric.Gauge(), p.config.Profile)
				case pmetric.MetricTypeSum:
					addAttributeToSum(metric.Sum(), p.config.Profile)
				case pmetric.MetricTypeHistogram:
					addAttributeToHistogram(metric.Histogram(), p.config.Profile)
				case pmetric.MetricTypeSummary:
					addAttributeToSummary(metric.Summary(), p.config.Profile)
				case pmetric.MetricTypeExponentialHistogram:
					addAttributeToExponentialHistogram(metric.ExponentialHistogram(), p.config.Profile)
				}
			}
		}
	}
	
	// Pass the modified metrics to the next consumer in the pipeline
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

// Helper functions to count and add attributes to different metric types

func countDataPoints(md pmetric.Metrics) int {
	total := 0
	
	// Pre-allocate resource metrics array size
	rms := md.ResourceMetrics()
	rmsLen := rms.Len()
	
	for i := 0; i < rmsLen; i++ {
		// Pre-allocate scope metrics array size
		sms := rms.At(i).ScopeMetrics()
		smsLen := sms.Len()
		
		for j := 0; j < smsLen; j++ {
			// Pre-allocate metrics array size
			metrics := sms.At(j).Metrics()
			metricsLen := metrics.Len()
			
			for k := 0; k < metricsLen; k++ {
				metric := metrics.At(k)
				
				// Add datapoints count based on metric type
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					total += metric.Gauge().DataPoints().Len()
				case pmetric.MetricTypeSum:
					total += metric.Sum().DataPoints().Len()
				case pmetric.MetricTypeHistogram:
					total += metric.Histogram().DataPoints().Len()
				case pmetric.MetricTypeSummary:
					total += metric.Summary().DataPoints().Len()
				case pmetric.MetricTypeExponentialHistogram:
					total += metric.ExponentialHistogram().DataPoints().Len()
				}
			}
		}
	}
	
	return total
}

// addAttributesToDataPoints is a generic function that adds common attributes to any type of datapoints
// that implement the common attributes interface (all OTel metric datapoints do)
func addAttributesToDataPoints(dataPoints interface{}, profile string) {
	type dataPointsGetter interface {
		Len() int
		At(i int) pmetric.NumberDataPoint
	}
	
	type histogramDataPointsGetter interface {
		Len() int
		At(i int) pmetric.HistogramDataPoint
	}
	
	type expHistogramDataPointsGetter interface {
		Len() int
		At(i int) pmetric.ExponentialHistogramDataPoint
	}
	
	type summaryDataPointsGetter interface {
		Len() int
		At(i int) pmetric.SummaryDataPoint
	}
	
	switch pts := dataPoints.(type) {
	case dataPointsGetter:
		for i := 0; i < pts.Len(); i++ {
			dp := pts.At(i)
			dp.Attributes().PutStr("processor.traceaware", "processed")
			dp.Attributes().PutStr("collector.profile", profile)
		}
	case histogramDataPointsGetter:
		for i := 0; i < pts.Len(); i++ {
			dp := pts.At(i)
			dp.Attributes().PutStr("processor.traceaware", "processed")
			dp.Attributes().PutStr("collector.profile", profile)
		}
	case expHistogramDataPointsGetter:
		for i := 0; i < pts.Len(); i++ {
			dp := pts.At(i)
			dp.Attributes().PutStr("processor.traceaware", "processed")
			dp.Attributes().PutStr("collector.profile", profile)
		}
	case summaryDataPointsGetter:
		for i := 0; i < pts.Len(); i++ {
			dp := pts.At(i)
			dp.Attributes().PutStr("processor.traceaware", "processed")
			dp.Attributes().PutStr("collector.profile", profile)
		}
	}
}

func addAttributeToGauge(gauge pmetric.Gauge, profile string) {
	addAttributesToDataPoints(gauge.DataPoints(), profile)
}

func addAttributeToSum(sum pmetric.Sum, profile string) {
	addAttributesToDataPoints(sum.DataPoints(), profile)
}

func addAttributeToHistogram(histogram pmetric.Histogram, profile string) {
	addAttributesToDataPoints(histogram.DataPoints(), profile)
}

func addAttributeToSummary(summary pmetric.Summary, profile string) {
	addAttributesToDataPoints(summary.DataPoints(), profile)
}

func addAttributeToExponentialHistogram(eh pmetric.ExponentialHistogram, profile string) {
	addAttributesToDataPoints(eh.DataPoints(), profile)
}