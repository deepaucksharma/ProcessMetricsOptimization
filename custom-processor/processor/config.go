package traceawareprocessor

import (
	"fmt"
)

// Config defines configuration for the Trace-Aware Reservoir processor
type Config struct {
	// Profile is the profile identifier (ultra, balanced, etc.)
	Profile string `mapstructure:"profile"`
}

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if cfg.Profile == "" {
		return fmt.Errorf("profile cannot be empty")
	}
	return nil
}