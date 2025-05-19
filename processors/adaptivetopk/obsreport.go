package adaptivetopk

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/metric"
)

const processorName = "adaptivetopk"

// adaptiveTopKObsreport encapsulates observability functionality for the AdaptiveTopK processor
type adaptiveTopKObsreport struct {
	settings              component.TelemetrySettings
	processedPoints       metric.Int64Counter
	droppedPoints         metric.Int64Counter
	topKProcessesSelected metric.Int64Counter
	currentKValue         metric.Int64Observable // For Dynamic K
	currentVal            int64
}

func newAdaptiveTopKObsreport(settings component.TelemetrySettings) (*adaptiveTopKObsreport, error) {
	var processedPoints metric.Int64Counter
	var droppedPoints metric.Int64Counter
	var topKProcessesSelected metric.Int64Counter
	var currentKValue metric.Int64Observable

	// Create metrics if MeterProvider is available
	if settings.MeterProvider != nil {
		meter := settings.MeterProvider.Meter(processorName)

		var err error
		processedPoints, err = meter.Int64Counter(
			"otelcol_otelcol_processor_adaptivetopk_processed_metric_points",
			metric.WithDescription("Number of metric points processed by the adaptivetopk processor"),
		)
		if err != nil {
			return nil, err
		}

		droppedPoints, err = meter.Int64Counter(
			"otelcol_otelcol_processor_adaptivetopk_dropped_metric_points",
			metric.WithDescription("Number of metric points dropped by the adaptivetopk processor"),
		)
		if err != nil {
			return nil, err
		}

		topKProcessesSelected, err = meter.Int64Counter(
			"otelcol_otelcol_adaptivetopk_topk_processes_selected_total",
			metric.WithDescription("Total number of non-critical processes selected for Top K in each batch"),
		)
		if err != nil {
			return nil, err
		}

		// Using Int64Observable instead of deprecated Int64Gauge
		currentKValue, err = meter.Int64ObservableGauge(
			"otelcol_otelcol_adaptivetopk_current_k_value",
			metric.WithDescription("The current value of K being used for selection (for Dynamic K)"),
		)
		if err != nil {
			return nil, err
		}
	}

	o := &adaptiveTopKObsreport{
		settings:              settings,
		processedPoints:       processedPoints,
		droppedPoints:         droppedPoints,
		topKProcessesSelected: topKProcessesSelected,
		currentKValue:         currentKValue,
	}

	if settings.MeterProvider != nil {
		_, err := settings.MeterProvider.Meter(processorName).RegisterCallback(func(ctx context.Context, obs metric.Observer) error {
			obs.ObserveInt64(o.currentKValue, o.currentVal)
			return nil
		}, o.currentKValue)
		if err != nil {
			return nil, err
		}
	}

	return o, nil
}

// StartMetricsOp starts the metrics operation and returns the context
func (o *adaptiveTopKObsreport) StartMetricsOp(ctx context.Context) context.Context {
	// Simply return the context as is
	return ctx
}

// EndMetricsOp ends the metrics operation and records the number of processed metrics
func (o *adaptiveTopKObsreport) EndMetricsOp(ctx context.Context, _ string, numProcessedPoints int, numDroppedPoints int, err error) {
	// Record the number of processed points
	if o.processedPoints != nil {
		o.processedPoints.Add(ctx, int64(numProcessedPoints))
	}

	// Record dropped points
	if o.droppedPoints != nil && numDroppedPoints > 0 {
		o.droppedPoints.Add(ctx, int64(numDroppedPoints))
	}
}

// recordTopKProcessesSelected records the number of processes selected for Top K
func (o *adaptiveTopKObsreport) recordTopKProcessesSelected(ctx context.Context, count int64) {
	if o.topKProcessesSelected != nil {
		o.topKProcessesSelected.Add(ctx, count)
	}
}

// recordCurrentKValue records the current K value being used (for Dynamic K)
func (o *adaptiveTopKObsreport) recordCurrentKValue(ctx context.Context, kValue int64) {
	o.currentVal = kValue
}
