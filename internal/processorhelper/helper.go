package processorhelper

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/otel/metric"
)

// ObsHelper provides basic obsreport metrics for processors.
type ObsHelper struct {
	name      string
	settings  component.TelemetrySettings
	processed metric.Int64Counter
	dropped   metric.Int64Counter
}

// NewObsHelper creates a new ObsHelper for the given processor name.
func NewObsHelper(name string, settings component.TelemetrySettings) (*ObsHelper, error) {
	var processed metric.Int64Counter
	var dropped metric.Int64Counter
	if settings.MeterProvider != nil {
		meter := settings.MeterProvider.Meter(name)
		var err error
		processed, err = meter.Int64Counter(
			"otelcol_otelcol_processor_"+name+"_processed_metric_points",
			metric.WithDescription("Number of metric points processed by the "+name+" processor"),
		)
		if err != nil {
			return nil, err
		}
		dropped, err = meter.Int64Counter(
			"otelcol_otelcol_processor_"+name+"_dropped_metric_points",
			metric.WithDescription("Number of metric points dropped by the "+name+" processor"),
		)
		if err != nil {
			return nil, err
		}
	}
	return &ObsHelper{name: name, settings: settings, processed: processed, dropped: dropped}, nil
}

// StartMetricsOp begins processing observation.
func (o *ObsHelper) StartMetricsOp(ctx context.Context) context.Context {
	return ctx
}

// EndMetricsOp records the outcome of a metrics operation.
func (o *ObsHelper) EndMetricsOp(ctx context.Context, numProcessed int, err error) {
	if o.processed != nil {
		o.processed.Add(ctx, int64(numProcessed))
	}
	if err != nil && o.dropped != nil {
		o.dropped.Add(ctx, int64(numProcessed))
	}
}

// Helper embeds TelemetrySettings and provides default processor methods.
type Helper struct {
	component.TelemetrySettings
	Obsreport *ObsHelper
}

// NewHelper constructs a Helper with obsreport instrumentation.
func NewHelper(settings component.TelemetrySettings, name string) (*Helper, error) {
	obs, err := NewObsHelper(name, settings)
	if err != nil {
		return nil, err
	}
	return &Helper{TelemetrySettings: settings, Obsreport: obs}, nil
}

// Start implements component.Component.
func (h *Helper) Start(context.Context, component.Host) error { return nil }

// Shutdown implements component.Component.
func (h *Helper) Shutdown(context.Context) error { return nil }

// Capabilities indicates this processor mutates data by default.
func (h *Helper) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}

// StartMetricsOp delegates to the underlying ObsHelper.
func (h *Helper) StartMetricsOp(ctx context.Context) context.Context {
	return h.Obsreport.StartMetricsOp(ctx)
}

// EndMetricsOp delegates to the underlying ObsHelper.
func (h *Helper) EndMetricsOp(ctx context.Context, numProcessed int, err error) {
	h.Obsreport.EndMetricsOp(ctx, numProcessed, err)
}
