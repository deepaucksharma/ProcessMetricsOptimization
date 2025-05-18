package helloworld

import (
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

// Config defines the configuration for the Hello World processor.
type Config struct {
	// Message is the message to include in logs.
	Message string `mapstructure:"message"`

	// AddToResource determines whether to add the hello attribute to resources
	// in addition to metric data points.
	AddToResource bool `mapstructure:"add_to_resource"`
}

var _ component.Config = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Message == "" {
		return errors.New("message cannot be empty")
	}
	return nil
}

// Unmarshal implements confmap.Unmarshaler
func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		return nil
	}

	// Set defaults
	cfg.Message = "Hello from OpenTelemetry!"
	cfg.AddToResource = false

	// Unmarshal configuration
	return componentParser.Unmarshal(cfg)
}

// ProcessorType returns the processor type for metrics usage
func (cfg *Config) ProcessorType() string {
	return "helloworld"
}
