package othersrollup

import (
	"errors"
	"strings"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

// AggregationType defines the type of aggregation to perform.
type AggregationType string

const (
	SumAggregation AggregationType = "sum"
	AvgAggregation AggregationType = "avg"
	// Future potential aggregation types
	// MinAggregation AggregationType = "min"
	// MaxAggregation AggregationType = "max"
)

// Config defines the configuration for the OthersRollup processor.
type Config struct {
	OutputPIDAttributeValue            string                     `mapstructure:"output_pid_attribute_value"`
	OutputExecutableNameAttributeValue string                     `mapstructure:"output_executable_name_attribute_value"`
	Aggregations                       map[string]AggregationType `mapstructure:"aggregations"`
	MetricsToRollup                    []string                   `mapstructure:"metrics_to_rollup"`
	PriorityAttributeName              string                     `mapstructure:"priority_attribute_name"`
	CriticalAttributeValue             string                     `mapstructure:"critical_attribute_value"`
	// TopKAttributeName string `mapstructure:"topk_attribute_name"` // If adaptivetopk adds a tag
}

var _ component.Config = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

func (cfg *Config) Validate() error {
	if cfg.OutputPIDAttributeValue == "" {
		return errors.New("output_pid_attribute_value cannot be empty")
	}
	if cfg.OutputExecutableNameAttributeValue == "" {
		return errors.New("output_executable_name_attribute_value cannot be empty")
	}
	if cfg.PriorityAttributeName == "" {
		return errors.New("priority_attribute_name must be specified to identify non-rollup candidates")
	}
	if cfg.CriticalAttributeValue == "" {
		return errors.New("critical_attribute_value must be specified")
	}

	for metric, agg := range cfg.Aggregations {
		if metric == "" {
			return errors.New("metric name in aggregations cannot be empty")
		}
		switch strings.ToLower(string(agg)) {
		case string(SumAggregation), string(AvgAggregation):
			// valid
		default:
			return errors.New("invalid aggregation type for metric " + metric + ": " + string(agg) + ". Supported: sum, avg")
		}
	}
	return nil
}

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		return nil
	}

	// Set default values
	cfg.OutputPIDAttributeValue = "-1"
	cfg.OutputExecutableNameAttributeValue = "_other_"
	cfg.Aggregations = map[string]AggregationType{
		"process.cpu.utilization": AvgAggregation,
		"process.memory.rss":      SumAggregation,
	}
	cfg.MetricsToRollup = []string{} // Default: rollup all compatible non-priority/TopK metrics
	cfg.PriorityAttributeName = "nr.priority"
	cfg.CriticalAttributeValue = "critical"

	return componentParser.Unmarshal(cfg)
}

// ProcessorType returns the processor type for metrics usage
func (cfg *Config) ProcessorType() string {
	return "othersrollup"
}
