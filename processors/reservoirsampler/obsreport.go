package reservoirsampler

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/metric"
)

const processorName = "reservoirsampler"

type reservoirSamplerObsreport struct {
	settings                      component.TelemetrySettings
	processedPoints               metric.Int64Counter
	droppedPoints                 metric.Int64Counter
	reservoirFillRatio            metric.Float64Observable
	selectedIdentitiesCount       metric.Int64Observable
	eligibleIdentitiesSeenTotal   metric.Int64Counter
	newIdentitiesAddedToReservoir metric.Int64Counter
}

func newReservoirSamplerObsreport(settings component.TelemetrySettings) (*reservoirSamplerObsreport, error) {
	var processedPoints metric.Int64Counter
	var droppedPoints metric.Int64Counter
	var reservoirFillRatio metric.Float64Observable
	var selectedIdentitiesCount metric.Int64Observable
	var eligibleIdentitiesSeenTotal metric.Int64Counter
	var newIdentitiesAddedToReservoir metric.Int64Counter

	// Create metrics if MeterProvider is available
	if settings.MeterProvider != nil {
		meter := settings.MeterProvider.Meter(processorName)

		var err error
		processedPoints, err = meter.Int64Counter(
			"otelcol_otelcol_processor_reservoirsampler_processed_metric_points",
			metric.WithDescription("Number of metric points processed by the reservoirsampler processor"),
		)
		if err != nil {
			return nil, err
		}

		droppedPoints, err = meter.Int64Counter(
			"otelcol_otelcol_processor_reservoirsampler_dropped_metric_points",
			metric.WithDescription("Number of metric points dropped by the reservoirsampler processor"),
		)
		if err != nil {
			return nil, err
		}

		reservoirFillRatio, err = meter.Float64ObservableGauge(
			"otelcol_otelcol_reservoirsampler_reservoir_fill_ratio",
			metric.WithDescription("Current fill ratio of the reservoir (0.0 to 1.0)."),
		)
		if err != nil {
			return nil, err
		}

		selectedIdentitiesCount, err = meter.Int64ObservableGauge(
			"otelcol_otelcol_reservoirsampler_selected_identities_count",
			metric.WithDescription("Current number of unique process identities in the reservoir."),
		)
		if err != nil {
			return nil, err
		}

		eligibleIdentitiesSeenTotal, err = meter.Int64Counter(
			"otelcol_otelcol_reservoirsampler_eligible_identities_seen_total",
			metric.WithDescription("Total number of unique eligible process identities encountered since collector start."),
		)
		if err != nil {
			return nil, err
		}

		newIdentitiesAddedToReservoir, err = meter.Int64Counter(
			"otelcol_otelcol_reservoirsampler_new_identities_added_to_reservoir_total",
			metric.WithDescription("Total count of new identities added/replaced in the reservoir."),
		)
		if err != nil {
			return nil, err
		}
	}

	return &reservoirSamplerObsreport{
		settings:                      settings,
		processedPoints:               processedPoints,
		droppedPoints:                 droppedPoints,
		reservoirFillRatio:            reservoirFillRatio,
		selectedIdentitiesCount:       selectedIdentitiesCount,
		eligibleIdentitiesSeenTotal:   eligibleIdentitiesSeenTotal,
		newIdentitiesAddedToReservoir: newIdentitiesAddedToReservoir,
	}, nil
}

// StartMetricsOp starts the metrics operation and returns the context
func (o *reservoirSamplerObsreport) StartMetricsOp(ctx context.Context) context.Context {
	// Simply return the context as is
	return ctx
}

// EndMetricsOp ends the metrics operation and records the number of processed metrics
func (o *reservoirSamplerObsreport) EndMetricsOp(ctx context.Context, _ string, numProcessedPoints int, numDroppedPoints int, err error) {
	// Record the number of processed points
	if o.processedPoints != nil {
		o.processedPoints.Add(ctx, int64(numProcessedPoints))
	}

	// Record dropped points
	if o.droppedPoints != nil && numDroppedPoints > 0 {
		o.droppedPoints.Add(ctx, int64(numDroppedPoints))
	}
}

func (o *reservoirSamplerObsreport) recordReservoirFillRatio(ctx context.Context, ratio float64) {
	// This would normally use Observable registration, but for simplicity we'll log
	// Nothing to do here as this would be handled by registration callback
}

func (o *reservoirSamplerObsreport) recordSelectedIdentitiesCount(ctx context.Context, count int64) {
	// This would normally use Observable registration, but for simplicity we'll log
	// Nothing to do here as this would be handled by registration callback
}

func (o *reservoirSamplerObsreport) recordEligibleIdentitiesSeen(ctx context.Context) {
	if o.eligibleIdentitiesSeenTotal != nil {
		o.eligibleIdentitiesSeenTotal.Add(ctx, 1)
	}
}

func (o *reservoirSamplerObsreport) recordNewIdentityAdded(ctx context.Context) {
	if o.newIdentitiesAddedToReservoir != nil {
		o.newIdentitiesAddedToReservoir.Add(ctx, 1)
	}
}
