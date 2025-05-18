package reservoirsampler

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	typeStr   = "reservoirsampler"
	stability = component.StabilityLevelDevelopment
)

func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		ReservoirSize:           100,
		IdentityAttributes:      []string{"process.pid"},
		SampledAttributeName:    "nr.process_sampled_by_reservoir",
		SampledAttributeValue:   "true",
		SampleRateAttributeName: "nr.sample_rate",
		PriorityAttributeName:   "nr.priority",
		CriticalAttributeValue:  "critical",
	}
}

func createMetricsProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (processor.Metrics, error) {
	processorCfg := cfg.(*Config)
	if err := processorCfg.Validate(); err != nil {
		return nil, err
	}
	return newReservoirSamplerProcessor(set, nextConsumer, processorCfg)
}
