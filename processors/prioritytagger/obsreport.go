package prioritytagger

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/metric"
)

// obsreportHelper encapsulates observability functionality for the PriorityTagger processor.
type obsreportHelper struct {
	settings                component.TelemetrySettings
	processedPoints         metric.Int64Counter
	droppedPoints           metric.Int64Counter
	criticalProcessesTagged metric.Int64Counter
}

// newObsreportHelper creates a new observability helper for the PriorityTagger processor.
func newObsreportHelper(settings component.TelemetrySettings) (*obsreportHelper, error) {
	var processedPoints metric.Int64Counter
	var droppedPoints metric.Int64Counter
	var criticalProcessesTagged metric.Int64Counter

	// Create metrics if MeterProvider is available
	if settings.MeterProvider != nil {
		meter := settings.MeterProvider.Meter("prioritytagger")

		var err error
		processedPoints, err = meter.Int64Counter(
			"otelcol_otelcol_processor_prioritytagger_processed_metric_points",
			metric.WithDescription("Number of metric points processed by the prioritytagger processor"),
		)
		if err != nil {
			return nil, err
		}

		droppedPoints, err = meter.Int64Counter(
			"otelcol_otelcol_processor_prioritytagger_dropped_metric_points",
			metric.WithDescription("Number of metric points dropped by the prioritytagger processor"),
		)
		if err != nil {
			return nil, err
		}

		criticalProcessesTagged, err = meter.Int64Counter(
			"otelcol_otelcol_prioritytagger_critical_processes_tagged_total",
			metric.WithDescription("Total number of processes tagged as critical by the prioritytagger processor"),
		)
		if err != nil {
			return nil, err
		}
	}

	return &obsreportHelper{
		settings:                settings,
		processedPoints:         processedPoints,
		droppedPoints:           droppedPoints,
		criticalProcessesTagged: criticalProcessesTagged,
	}, nil
}

// StartMetricsOp starts the metrics operation and returns the context.
func (orh *obsreportHelper) StartMetricsOp(ctx context.Context) context.Context {
	// Simply return the context as is
	return ctx
}

// EndMetricsOp ends the metrics operation and records the number of processed metrics.
func (orh *obsreportHelper) EndMetricsOp(ctx context.Context, processorName string, numReceivedItems int, err error) {
	// Record the number of processed points
	if orh.processedPoints != nil {
		orh.processedPoints.Add(ctx, int64(numReceivedItems))
	}

	// If there was an error, record dropped points
	if err != nil && orh.droppedPoints != nil {
		orh.droppedPoints.Add(ctx, int64(numReceivedItems))
	}
}

// RecordTaggedProcess increments the counter for tagged critical processes
func (orh *obsreportHelper) RecordTaggedProcess(ctx context.Context) {
	if orh.criticalProcessesTagged != nil {
		orh.criticalProcessesTagged.Add(ctx, 1)
	}
}
