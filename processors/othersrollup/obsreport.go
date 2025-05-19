package othersrollup

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/metric"
)

const processorName = "othersrollup"

type othersRollupObsreport struct {
	settings                 component.TelemetrySettings
	processedPoints          metric.Int64Counter
	droppedPoints            metric.Int64Counter
	aggregatedSeriesCount    metric.Int64Counter
	inputSeriesRolledUpTotal metric.Int64Counter
}

func newOthersRollupObsreport(settings component.TelemetrySettings) (*othersRollupObsreport, error) {
	var processedPoints metric.Int64Counter
	var droppedPoints metric.Int64Counter
	var aggregatedSeriesCount metric.Int64Counter
	var inputSeriesRolledUpTotal metric.Int64Counter

	// Create metrics if MeterProvider is available
	if settings.MeterProvider != nil {
		meter := settings.MeterProvider.Meter(processorName)

		var err error
		processedPoints, err = meter.Int64Counter(
			"otelcol_otelcol_processor_othersrollup_processed_metric_points",
			metric.WithDescription("Number of metric points processed by the othersrollup processor"),
		)
		if err != nil {
			return nil, err
		}

		droppedPoints, err = meter.Int64Counter(
			"otelcol_otelcol_processor_othersrollup_dropped_metric_points",
			metric.WithDescription("Number of metric points dropped by the othersrollup processor"),
		)
		if err != nil {
			return nil, err
		}

		aggregatedSeriesCount, err = meter.Int64Counter(
			"otelcol_otelcol_othersrollup_aggregated_series_count_total",
			metric.WithDescription("Number of new _other_ series generated per batch."),
		)
		if err != nil {
			return nil, err
		}

		inputSeriesRolledUpTotal, err = meter.Int64Counter(
			"otelcol_otelcol_othersrollup_input_series_rolled_up_total",
			metric.WithDescription("Number of input series that were aggregated into an _other_ series."),
		)
		if err != nil {
			return nil, err
		}
	}

	return &othersRollupObsreport{
		settings:                 settings,
		processedPoints:          processedPoints,
		droppedPoints:            droppedPoints,
		aggregatedSeriesCount:    aggregatedSeriesCount,
		inputSeriesRolledUpTotal: inputSeriesRolledUpTotal,
	}, nil
}

// StartMetricsOp starts the metrics operation and returns the context
func (o *othersRollupObsreport) StartMetricsOp(ctx context.Context) context.Context {
	// Simply return the context as is
	return ctx
}

// EndMetricsOp ends the metrics operation and records the number of processed metrics
func (o *othersRollupObsreport) EndMetricsOp(ctx context.Context, _ string, numProcessedPoints int, numDroppedPoints int, err error) {
	// Record the number of processed points
	if o.processedPoints != nil {
		o.processedPoints.Add(ctx, int64(numProcessedPoints))
	}

	// Record dropped points
	if o.droppedPoints != nil && numDroppedPoints > 0 {
		o.droppedPoints.Add(ctx, int64(numDroppedPoints))
	}
}

func (o *othersRollupObsreport) recordAggregatedSeries(ctx context.Context, count int64) {
	if o.aggregatedSeriesCount != nil {
		o.aggregatedSeriesCount.Add(ctx, count)
	}
}

func (o *othersRollupObsreport) recordInputSeriesRolledUp(ctx context.Context, count int64) {
	if o.inputSeriesRolledUpTotal != nil {
		o.inputSeriesRolledUpTotal.Add(ctx, count)
	}
}
