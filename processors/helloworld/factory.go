package helloworld

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	// The value of "type" key in configuration.
	typeStr = "helloworld"

	// The stability level of the processor.
	stability = component.StabilityLevelDevelopment
)

// NewFactory creates a new factory for the Hello World processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, stability),
	)
}

// createDefaultConfig creates the default configuration for the Hello World processor.
func createDefaultConfig() component.Config {
	return &Config{
		Message:       "Hello from OpenTelemetry!",
		AddToResource: false,
	}
}

// createMetricsProcessor creates a metrics processor based on this config.
func createMetricsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	pCfg := cfg.(*Config)
	proc, err := newProcessor(pCfg, set.Logger, nextConsumer, set.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	return proc, nil
}
