package helloworld

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/metric"
)

// obsreportHelper encapsulates observability functionality for the Hello World processor.
type obsreportHelper struct {
	settings        component.TelemetrySettings
	processedPoints metric.Int64Counter
	droppedPoints   metric.Int64Counter
}

// newObsreportHelper creates a new observability helper for the Hello World processor.
func newObsreportHelper(settings component.TelemetrySettings) (*obsreportHelper, error) {
	meter := settings.MeterProvider.Meter("helloworld")

	processedPoints, err := meter.Int64Counter(
		"otelcol_otelcol_processor_helloworld_processed_metric_points",
		metric.WithDescription("Number of metric points processed by the hello world processor"),
	)
	if err != nil {
		return nil, err
	}

	droppedPoints, err := meter.Int64Counter(
		"otelcol_otelcol_processor_helloworld_dropped_metric_points",
		metric.WithDescription("Number of metric points dropped by the hello world processor"),
	)
	if err != nil {
		return nil, err
	}

	return &obsreportHelper{
		settings:        settings,
		processedPoints: processedPoints,
		droppedPoints:   droppedPoints,
	}, nil
}

// StartMetricsOp starts the metrics operation and returns the context.
func (orh *obsreportHelper) StartMetricsOp(ctx context.Context) context.Context {
	// Simply return the context as is
	return ctx
}

// EndMetricsOp ends the metrics operation and records the number of processed metrics.
func (orh *obsreportHelper) EndMetricsOp(ctx context.Context, numProcessedPoints int, numDroppedPoints int, err error) {
	// Record the number of processed points
	orh.processedPoints.Add(ctx, int64(numProcessedPoints))

	// Record dropped points
	if numDroppedPoints > 0 {
		orh.droppedPoints.Add(ctx, int64(numDroppedPoints))
	}
}
