package helloworld

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/zap"
)

type helloWorldProcessor struct {
	config           *Config
	logger           *zap.Logger
	metricsConsumer  consumer.Metrics
	obsrecv          *obsreportHelper
	mutationsCounter metric.Int64Counter
}

func newProcessor(config *Config, logger *zap.Logger, mexp consumer.Metrics, settings component.TelemetrySettings) (*helloWorldProcessor, error) {
	obsrecv, err := newObsreportHelper(settings)
	if err != nil {
		return nil, err
	}

	// Create a metric for counting mutations
	meter := settings.MeterProvider.Meter("helloworld")
	mutationsCounter, err := meter.Int64Counter(
		"otelcol_otelcol_helloworld_mutations_total",
		metric.WithDescription("Total number of metrics modified by the hello world processor"),
	)
	if err != nil {
		return nil, err
	}

	return &helloWorldProcessor{
		config:           config,
		logger:           logger,
		metricsConsumer:  mexp,
		obsrecv:          obsrecv,
		mutationsCounter: mutationsCounter,
	}, nil
}

func (p *helloWorldProcessor) Start(_ context.Context, _ component.Host) error {
	return nil
}

func (p *helloWorldProcessor) Shutdown(_ context.Context) error {
	return nil
}

func (p *helloWorldProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

func (p *helloWorldProcessor) ConsumeMetrics(ctx context.Context, md pmetric.Metrics) error {
	// Start the metrics observation
	ctx = p.obsrecv.StartMetricsOp(ctx)

	// Add Hello World attribute to each metric
	metricCount := p.processMetrics(ctx, md)

	// Record the observation and the number of processed items
	p.obsrecv.EndMetricsOp(ctx, metricCount, 0, nil)

	// Increment our custom mutation counter metric
	p.mutationsCounter.Add(ctx, int64(metricCount))

	// Send the modified metrics to the next consumer
	return p.metricsConsumer.ConsumeMetrics(ctx, md)
}

func (p *helloWorldProcessor) processMetrics(ctx context.Context, md pmetric.Metrics) int {
	// This count tracks the number of data points we've processed
	metricCount := 0

	// Iterate through the resource metrics
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)

		// Add attributes to the resource level if configured
		if p.config.AddToResource {
			addHelloAttribute(rm.Resource().Attributes())
		}

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
						addHelloAttribute(pts.At(l).Attributes())
						metricCount++
					}
				case pmetric.MetricTypeSum:
					pts := metric.Sum().DataPoints()
					for l := 0; l < pts.Len(); l++ {
						addHelloAttribute(pts.At(l).Attributes())
						metricCount++
					}
				case pmetric.MetricTypeHistogram:
					pts := metric.Histogram().DataPoints()
					for l := 0; l < pts.Len(); l++ {
						addHelloAttribute(pts.At(l).Attributes())
						metricCount++
					}
				case pmetric.MetricTypeSummary:
					pts := metric.Summary().DataPoints()
					for l := 0; l < pts.Len(); l++ {
						addHelloAttribute(pts.At(l).Attributes())
						metricCount++
					}
				case pmetric.MetricTypeExponentialHistogram:
					pts := metric.ExponentialHistogram().DataPoints()
					for l := 0; l < pts.Len(); l++ {
						addHelloAttribute(pts.At(l).Attributes())
						metricCount++
					}
				}
			}
		}
	}

	p.logger.Debug("Hello World processor modified metrics",
		zap.Int("count", metricCount),
		zap.String("message", p.config.Message))

	return metricCount
}

func addHelloAttribute(attrs pcommon.Map) {
	attrs.PutStr("hello.processor", "Hello from OpenTelemetry!")
}
