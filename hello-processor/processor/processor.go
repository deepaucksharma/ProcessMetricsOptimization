package helloprocessor

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

type helloProcessor struct {
	logger       *zap.Logger
	config       *Config
	nextConsumer consumer.Metrics
	startTime    time.Time
}

func newHelloProcessor(logger *zap.Logger, cfg *Config, nextConsumer consumer.Metrics) (*helloProcessor, error) {
	return &helloProcessor{
		logger:       logger,
		config:       cfg,
		nextConsumer: nextConsumer,
		startTime:    time.Now(),
	}, nil
}

// Capabilities returns the capabilities of the processor.
func (p *helloProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

// Start is called when the processor is starting.
func (p *helloProcessor) Start(_ context.Context, _ component.Host) error {
	p.logger.Info("Hello World Processor started",
		zap.String("message", p.config.Message))
	return nil
}

// Shutdown is called when the processor is shutting down.
func (p *helloProcessor) Shutdown(_ context.Context) error {
	p.logger.Info("Hello World Processor stopped",
		zap.Duration("uptime", time.Since(p.startTime)))
	return nil
}

// ConsumeMetrics processes the metrics.
func (p *helloProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// Log receipt of metrics
	p.logger.Debug("Hello World Processor: Processing metrics",
		zap.Int("metrics_count", md.ResourceMetrics().Len()),
		zap.Int("datapoints", countDataPoints(md)))
	
	// Add hello world attribute to all metrics
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		
		// Add hello world resource attributes
		rm.Resource().Attributes().PutStr("hello.world", p.config.Message)
		
		// Add hello world attribute to each datapoint
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				addHelloAttribute(metric, p.config.Message)
			}
		}
	}
	
	// Pass the modified metrics to the next consumer in the pipeline
	return p.nextConsumer.ConsumeMetrics(ctx, md)
}

// Helper functions

func countDataPoints(md pmetric.Metrics) int {
	total := 0
	rms := md.ResourceMetrics()
	
	for i := 0; i < rms.Len(); i++ {
		sms := rms.At(i).ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			metrics := sms.At(j).Metrics()
			for k := 0; k < metrics.Len(); k++ {
				metric := metrics.At(k)
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

func addHelloAttribute(metric pmetric.Metric, message string) {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		addHelloToGauge(metric.Gauge(), message)
	case pmetric.MetricTypeSum:
		addHelloToSum(metric.Sum(), message)
	case pmetric.MetricTypeHistogram:
		addHelloToHistogram(metric.Histogram(), message)
	case pmetric.MetricTypeSummary:
		addHelloToSummary(metric.Summary(), message)
	case pmetric.MetricTypeExponentialHistogram:
		addHelloToExponentialHistogram(metric.ExponentialHistogram(), message)
	}
}

func addHelloToGauge(gauge pmetric.Gauge, message string) {
	for i := 0; i < gauge.DataPoints().Len(); i++ {
		dp := gauge.DataPoints().At(i)
		dp.Attributes().PutStr("hello.processor", message)
	}
}

func addHelloToSum(sum pmetric.Sum, message string) {
	for i := 0; i < sum.DataPoints().Len(); i++ {
		dp := sum.DataPoints().At(i)
		dp.Attributes().PutStr("hello.processor", message)
	}
}

func addHelloToHistogram(histogram pmetric.Histogram, message string) {
	for i := 0; i < histogram.DataPoints().Len(); i++ {
		dp := histogram.DataPoints().At(i)
		dp.Attributes().PutStr("hello.processor", message)
	}
}

func addHelloToSummary(summary pmetric.Summary, message string) {
	for i := 0; i < summary.DataPoints().Len(); i++ {
		dp := summary.DataPoints().At(i)
		dp.Attributes().PutStr("hello.processor", message)
	}
}

func addHelloToExponentialHistogram(eh pmetric.ExponentialHistogram, message string) {
	for i := 0; i < eh.DataPoints().Len(); i++ {
		dp := eh.DataPoints().At(i)
		dp.Attributes().PutStr("hello.processor", message)
	}
}