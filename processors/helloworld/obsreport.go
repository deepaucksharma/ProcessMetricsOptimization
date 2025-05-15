package helloworld

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/obsreport"
	"go.opentelemetry.io/collector/processor"
)

// obsreportHelper encapsulates the obsreport functionality for the Hello World processor.
type obsreportHelper struct {
	processor *obsreport.Processor
}

// newObsreportHelper creates a new obsreport helper for the Hello World processor.
func newObsreportHelper(settings component.TelemetrySettings) (*obsreportHelper, error) {
	proc, err := obsreport.NewProcessor(obsreport.ProcessorSettings{
		ProcessorID:             settings.ID,
		ProcessorCreateSettings: settings,
	})
	if err != nil {
		return nil, err
	}
	
	return &obsreportHelper{
		processor: proc,
	}, nil
}

// StartMetricsOp starts the metrics operation and returns the context and number of metrics.
func (orh *obsreportHelper) StartMetricsOp(ctx context.Context) (context.Context, int) {
	ctx = orh.processor.StartMetricsOp(ctx)
	return ctx, 0
}

// EndMetricsOp ends the metrics operation and records the number of processed metrics.
func (orh *obsreportHelper) EndMetricsOp(ctx context.Context, processorName string, numReceivedItems int, err error) {
	// This method doesn't use numReceivedItems - it just needs it to match the function signature.
	// What's actually used is numReceivedPoints.
	numReceivedPoints := numReceivedItems

	orh.processor.EndMetricsOp(ctx, processorName, numReceivedPoints, numReceivedPoints, err)
}