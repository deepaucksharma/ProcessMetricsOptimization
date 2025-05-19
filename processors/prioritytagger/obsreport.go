package prioritytagger

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/otel/metric"
)

// obsreportHelper encapsulates observability functionality for the PriorityTagger processor.
type obsreportHelper struct {
	settings                component.TelemetrySettings
	criticalProcessesTagged metric.Int64Counter
}

// newObsreportHelper creates a new observability helper for the PriorityTagger processor.
func newObsreportHelper(settings component.TelemetrySettings) (*obsreportHelper, error) {
	var criticalProcessesTagged metric.Int64Counter

	// Create metrics if MeterProvider is available
	if settings.MeterProvider != nil {
		meter := settings.MeterProvider.Meter("prioritytagger")

		var err error
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
		criticalProcessesTagged: criticalProcessesTagged,
	}, nil
}

// StartMetricsOp starts the metrics operation and returns the context.
func (orh *obsreportHelper) StartMetricsOp(ctx context.Context) context.Context {
	// Simply return the context as is
	return ctx
}

// EndMetricsOp ends the metrics operation and records the number of processed metrics.
func (orh *obsreportHelper) EndMetricsOp(ctx context.Context) {}

// RecordTaggedProcess increments the counter for tagged critical processes
func (orh *obsreportHelper) RecordTaggedProcess(ctx context.Context) {
	if orh.criticalProcessesTagged != nil {
		orh.criticalProcessesTagged.Add(ctx, 1)
	}
}
