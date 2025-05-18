package prioritytagger

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	// The value of "type" key in configuration.
	typeStr = "prioritytagger"

	// The stability level of the processor.
	stability = component.StabilityLevelDevelopment
)

// NewFactory creates a new factory for the PriorityTagger processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, stability),
	)
}

// createDefaultConfig creates the default configuration for the PriorityTagger processor.
func createDefaultConfig() component.Config {
	return &Config{
		CriticalExecutables:        []string{},
		CriticalExecutablePatterns: []string{},
		CPUSteadyStateThreshold:    -1.0, // Negative to disable by default
		MemoryRSSThresholdMiB:      -1,   // Negative to disable by default
		PriorityAttributeName:      "nr.priority",
		CriticalAttributeValue:     "critical",
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

	// Validate the configuration - important to catch regex compilation errors
	if err := pCfg.Validate(); err != nil {
		return nil, err
	}

	proc, err := newProcessor(pCfg, set.Logger, nextConsumer, set.TelemetrySettings)
	if err != nil {
		return nil, err
	}
	return proc, nil
}
