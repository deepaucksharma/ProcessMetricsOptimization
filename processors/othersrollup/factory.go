package othersrollup

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	typeStr   = "othersrollup"
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
		OutputPIDAttributeValue:            "-1",
		OutputExecutableNameAttributeValue: "_other_",
		Aggregations: map[string]AggregationType{
			"process.cpu.utilization": AvgAggregation,
			"process.memory.rss":      SumAggregation,
		},
		MetricsToRollup:        []string{},
		PriorityAttributeName:  "nr.priority",
		CriticalAttributeValue: "critical",
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
	return newOthersRollupProcessor(set, nextConsumer, processorCfg)
}
