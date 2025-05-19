package adaptivetopk

import (
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap"
)

// Config defines the configuration for the AdaptiveTopK processor.
type Config struct {
	// KValue is the fixed number of top processes to keep (Sub-Phase 2a).
	// If HostLoadMetricName is set, KValue is ignored.
	KValue int `mapstructure:"k_value"`

	// KeyMetricName is the metric used to rank processes (e.g., "process.cpu.utilization").
	KeyMetricName string `mapstructure:"key_metric_name"`
	// SecondaryKeyMetricName is an optional metric for tie-breaking.
	SecondaryKeyMetricName string `mapstructure:"secondary_key_metric_name"`

	// PriorityAttributeName is the attribute identifying critical processes.
	PriorityAttributeName string `mapstructure:"priority_attribute_name"`
	// CriticalAttributeValue is the value indicating a critical process.
	CriticalAttributeValue string `mapstructure:"critical_attribute_value"`

	// --- Sub-Phase 2b: Dynamic K & Hysteresis ---
	// HostLoadMetricName is the metric for overall host load (e.g., "system.cpu.utilization").
	// If set, KValue is ignored, and dynamic K is used.
	HostLoadMetricName string `mapstructure:"host_load_metric_name"`
	// LoadBandsToKMap maps host load thresholds to K values.
	// Example: {0.2: 5, 0.5: 10, 0.8: 20} (load_threshold: K_value)
	LoadBandsToKMap map[float64]int `mapstructure:"load_bands_to_k_map"`
	// HysteresisDuration is how long a process stays in TopK after dropping below threshold.
	HysteresisDuration time.Duration `mapstructure:"hysteresis_duration"`
	// MinKValue is the minimum bound for dynamic K.
	MinKValue int `mapstructure:"min_k_value"`
	// MaxKValue is the maximum bound for dynamic K.
	MaxKValue int `mapstructure:"max_k_value"`
}

var _ component.Config = (*Config)(nil)
var _ confmap.Unmarshaler = (*Config)(nil)

// Validate checks if the processor configuration is valid
func (cfg *Config) Validate() error {
	if cfg.KeyMetricName == "" {
		return errors.New("key_metric_name must be specified")
	}
	if cfg.PriorityAttributeName == "" {
		return errors.New("priority_attribute_name must be specified")
	}
	if cfg.CriticalAttributeValue == "" {
		return errors.New("critical_attribute_value must be specified")
	}

	isDynamicK := cfg.HostLoadMetricName != ""
	isFixedK := cfg.KValue > 0

	if isDynamicK {
		if len(cfg.LoadBandsToKMap) == 0 {
			return errors.New("load_bands_to_k_map must be specified when host_load_metric_name is set")
		}
		if cfg.MinKValue <= 0 {
			return errors.New("min_k_value must be positive when host_load_metric_name is set")
		}
		if cfg.MaxKValue < cfg.MinKValue {
			return errors.New("max_k_value must be greater than or equal to min_k_value when host_load_metric_name is set")
		}
		// Further validation for LoadBandsToKMap keys and values can be added.
		for threshold, k := range cfg.LoadBandsToKMap {
			if threshold < 0 || threshold > 1.0 { // Assuming load is a utilization metric
				return fmt.Errorf("load threshold must be between 0.0 and 1.0, got %.2f", threshold)
			}
			if k <= 0 {
				return fmt.Errorf("k value in load_bands_to_k_map must be positive, got %d for threshold %.2f", k, threshold)
			}
		}
	} else if !isFixedK {
		return errors.New("either k_value (for fixed K) or host_load_metric_name (for dynamic K) must be configured")
	} else if isFixedK && cfg.KValue <= 0 {
		return errors.New("k_value must be positive if host_load_metric_name is not set")
	}

	return nil
}

// Unmarshal implements confmap.Unmarshaler
func (cfg *Config) Unmarshal(componentParser *confmap.Conf) error {
	if componentParser == nil {
		return nil
	}

	// Set defaults (Fixed K defaults)
	cfg.KValue = 10 // Default fixed K
	cfg.KeyMetricName = "process.cpu.utilization"
	cfg.PriorityAttributeName = "nr.priority"
	cfg.CriticalAttributeValue = "critical"

	// Dynamic K defaults (if user enables dynamic K by setting HostLoadMetricName)
	cfg.HostLoadMetricName = ""
	cfg.LoadBandsToKMap = make(map[float64]int)
	cfg.HysteresisDuration = 1 * time.Minute
	cfg.MinKValue = 5
	cfg.MaxKValue = 20

	return componentParser.Unmarshal(cfg)
}

// IsDynamicK returns true if the processor is configured for dynamic K.
func (cfg *Config) IsDynamicK() bool {
	return cfg.HostLoadMetricName != ""
}

// ProcessorType returns the processor type for metrics usage
func (cfg *Config) ProcessorType() string {
	return "adaptivetopk"
}
