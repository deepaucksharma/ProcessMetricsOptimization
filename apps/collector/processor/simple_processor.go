package processor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
)

// SimpleReservoirProcessor is a simplified implementation for testing purposes
type SimpleReservoirProcessor struct {
	// Required interfaces for the processor
	component.StartFunc
	component.ShutdownFunc

	// Next consumer in the pipeline
	NextConsumer consumer.Traces
}

// ConsumeTraces processes trace data and samples it
func (p *SimpleReservoirProcessor) ConsumeTraces(ctx context.Context, traces ptrace.Traces) error {
	// For testing/compilation purposes, just pass the traces to the next consumer
	return p.NextConsumer.ConsumeTraces(ctx, traces)
}

// Start implements the Component interface
func (p *SimpleReservoirProcessor) Start(ctx context.Context, host component.Host) error {
	return nil
}

// Shutdown implements the Component interface
func (p *SimpleReservoirProcessor) Shutdown(ctx context.Context) error {
	return nil
}

// Capabilities implements the processor.Traces interface
func (p *SimpleReservoirProcessor) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: true}
}