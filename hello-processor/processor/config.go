package helloprocessor

import (
	"fmt"
)

// Config defines configuration for the Hello World processor
type Config struct {
	// Message is a simple configuration parameter for the greeting message
	Message string `mapstructure:"message"`
}

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Message == "" {
		return fmt.Errorf("message cannot be empty")
	}
	return nil
}