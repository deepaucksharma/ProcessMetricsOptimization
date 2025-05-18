package prioritytagger

import (
	"errors"
	"regexp"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

// Config defines the configuration for the PriorityTagger processor.
type Config struct {
	// CriticalExecutables is a list of process executable names that are considered critical and will be tagged.
	CriticalExecutables []string `mapstructure:"critical_executables"`

	// CriticalExecutablePatterns is a list of regex patterns for matching process executable names that are considered critical.
	CriticalExecutablePatterns []string `mapstructure:"critical_executable_patterns"`

	// CPUSteadyStateThreshold is an optional threshold for CPU utilization. Processes with utilization above this threshold
	// will be tagged as critical. Set to a negative value to disable this check.
	CPUSteadyStateThreshold float64 `mapstructure:"cpu_steady_state_threshold"`

	// MemoryRSSThresholdMiB is an optional threshold for memory RSS in MiB. Processes with RSS above this threshold
	// will be tagged as critical. Set to a negative value to disable this check.
	MemoryRSSThresholdMiB int64 `mapstructure:"memory_rss_threshold_mib"`

	// PriorityAttributeName is the name of the attribute that will be added to tag critical processes.
	PriorityAttributeName string `mapstructure:"priority_attribute_name"`

	// CriticalAttributeValue is the value that will be set for the priority attribute to mark a process as critical.
	CriticalAttributeValue string `mapstructure:"critical_attribute_value"`

	// Compiled regex patterns (not part of mapstructure)
	patterns []*regexp.Regexp
}

var _ component.Config = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if len(cfg.CriticalExecutables) == 0 && len(cfg.CriticalExecutablePatterns) == 0 &&
		cfg.CPUSteadyStateThreshold < 0 && cfg.MemoryRSSThresholdMiB < 0 {
		return errors.New("at least one of critical_executables or critical_executable_patterns must be specified")
	}

	if cfg.PriorityAttributeName == "" {
		return errors.New("priority_attribute_name cannot be empty")
	}

	if cfg.CriticalAttributeValue == "" {
		return errors.New("critical_attribute_value cannot be empty")
	}

	// Validate regex patterns
	cfg.patterns = make([]*regexp.Regexp, 0, len(cfg.CriticalExecutablePatterns))
	for _, pattern := range cfg.CriticalExecutablePatterns {
		if pattern == "" {
			return errors.New("regex pattern cannot be empty")
		}
		re, err := regexp.Compile(pattern)
		if err != nil {
			return errors.New("invalid regex pattern: " + pattern + ", " + err.Error())
		}
		cfg.patterns = append(cfg.patterns, re)
	}

	return nil
}

// Unmarshal implements confmap.Unmarshaler
func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		return nil
	}

	// Set defaults
	cfg.CriticalExecutables = []string{}
	cfg.CriticalExecutablePatterns = []string{}
	cfg.CPUSteadyStateThreshold = -1.0 // Negative to disable by default
	cfg.MemoryRSSThresholdMiB = -1     // Negative to disable by default
	cfg.PriorityAttributeName = "nr.priority"
	cfg.CriticalAttributeValue = "critical"

	// Unmarshal configuration
	return componentParser.Unmarshal(cfg)
}

// ProcessorType returns the processor type for metrics usage
func (cfg *Config) ProcessorType() string {
	return "prioritytagger"
}

// GetCompiledPatterns returns the compiled regex patterns
func (cfg *Config) GetCompiledPatterns() []*regexp.Regexp {
	return cfg.patterns
}
