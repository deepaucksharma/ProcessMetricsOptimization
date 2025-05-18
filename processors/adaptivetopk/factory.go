package adaptivetopk

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
)

const (
	typeStr   = "adaptivetopk"
	stability = component.StabilityLevelDevelopment
)

// NewFactory creates a new factory for the AdaptiveTopK processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		createDefaultConfig,
		processor.WithMetrics(createMetricsProcessor, stability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		KValue:                 10,
		KeyMetricName:          "process.cpu.utilization",
		PriorityAttributeName:  "nr.priority",
		CriticalAttributeValue: "critical",
		HostLoadMetricName:     "", // Dynamic K disabled by default
		LoadBandsToKMap:        make(map[float64]int),
		HysteresisDuration:     1 * time.Minute,
		MinKValue:              5,
		MaxKValue:              20,
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
	return newAdaptiveTopKProcessor(set, nextConsumer, processorCfg)
}
