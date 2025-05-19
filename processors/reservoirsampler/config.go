package reservoirsampler

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

type Config struct {
	ReservoirSize           int      `mapstructure:"reservoir_size"`
	IdentityAttributes      []string `mapstructure:"identity_attributes"`
	SampledAttributeName    string   `mapstructure:"sampled_attribute_name"`
	SampledAttributeValue   string   `mapstructure:"sampled_attribute_value"`
	SampleRateAttributeName string   `mapstructure:"sample_rate_attribute_name"`
	PriorityAttributeName   string   `mapstructure:"priority_attribute_name"`
	CriticalAttributeValue  string   `mapstructure:"critical_attribute_value"`
	// TopKAttributeName      string   `mapstructure:"topk_attribute_name"`
}

var _ component.Config = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

func (cfg *Config) Validate() error {
	if cfg.ReservoirSize <= 0 {
		return errors.New("reservoir_size must be positive")
	}
	if len(cfg.IdentityAttributes) == 0 {
		return errors.New("identity_attributes must be specified")
	}
	for _, attr := range cfg.IdentityAttributes {
		if attr == "" {
			return errors.New("identity_attributes cannot contain empty strings")
		}
	}
	if cfg.SampledAttributeName == "" {
		return errors.New("sampled_attribute_name cannot be empty")
	}
	if cfg.SampledAttributeValue == "" {
		return errors.New("sampled_attribute_value cannot be empty")
	}
	if cfg.SampleRateAttributeName == "" {
		return errors.New("sample_rate_attribute_name cannot be empty")
	}
	if cfg.PriorityAttributeName == "" {
		return errors.New("priority_attribute_name must be specified")
	}
	if cfg.CriticalAttributeValue == "" {
		return errors.New("critical_attribute_value must be specified")
	}
	return nil
}

func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		return nil
	}
	// Set default values
	cfg.ReservoirSize = 100
	cfg.IdentityAttributes = []string{"process.pid"}
	cfg.SampledAttributeName = "nr.process_sampled_by_reservoir"
	cfg.SampledAttributeValue = "true"
	cfg.SampleRateAttributeName = "nr.sample_rate"
	cfg.PriorityAttributeName = "nr.priority"
	cfg.CriticalAttributeValue = "critical"

	return componentParser.Unmarshal(cfg)
}

// ProcessorType returns the processor type for metrics usage
func (cfg *Config) ProcessorType() string {
	return "reservoirsampler"
}
