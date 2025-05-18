package prioritytagger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestConfigValidate(t *testing.T) {
	testCases := []struct {
		name        string
		cfg         Config
		expectError bool
	}{
		{
			name: "Valid configuration with executables",
			cfg: Config{
				CriticalExecutables:    []string{"kubelet", "systemd"},
				PriorityAttributeName:  "nr.priority",
				CriticalAttributeValue: "critical",
			},
			expectError: false,
		},
		{
			name: "Valid configuration with patterns",
			cfg: Config{
				CriticalExecutablePatterns: []string{"kube.*", ".*d$"},
				PriorityAttributeName:      "nr.priority",
				CriticalAttributeValue:     "critical",
			},
			expectError: false,
		},
		{
			name: "Invalid empty configuration",
			cfg: Config{
				PriorityAttributeName:   "nr.priority",
				CriticalAttributeValue:  "critical",
				CPUSteadyStateThreshold: -1.0, // Explicitly set to negative to disable
				MemoryRSSThresholdMiB:   -1,   // Explicitly set to negative to disable
			},
			expectError: true,
		},
		{
			name: "Invalid empty priority attribute name",
			cfg: Config{
				CriticalExecutables:    []string{"kubelet"},
				PriorityAttributeName:  "",
				CriticalAttributeValue: "critical",
			},
			expectError: true,
		},
		{
			name: "Invalid empty critical attribute value",
			cfg: Config{
				CriticalExecutables:    []string{"kubelet"},
				PriorityAttributeName:  "nr.priority",
				CriticalAttributeValue: "",
			},
			expectError: true,
		},
		{
			name: "Invalid regex pattern",
			cfg: Config{
				CriticalExecutablePatterns: []string{"[invalid"},
				PriorityAttributeName:      "nr.priority",
				CriticalAttributeValue:     "critical",
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cfg.Validate()
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProcessorTaggingByName(t *testing.T) {
	// Create test metrics with a process that should be tagged
	md := createTestMetrics("kubelet", 0.1, 100*1024*1024)

	// Create the processor with a configuration to tag kubelet
	cfg := &Config{
		CriticalExecutables:    []string{"kubelet", "systemd"},
		PriorityAttributeName:  "nr.priority",
		CriticalAttributeValue: "critical",
	}

	// Validate configuration and compile patterns
	err := cfg.Validate()
	require.NoError(t, err)

	// Create and run the processor
	mockConsumer := consumertest.NewNop()
	settings := component.TelemetrySettings{
		Logger:         zap.NewNop(),
		MeterProvider:  nil, // Using nil for testing
		TracerProvider: nil, // Using nil for testing
	}
	proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
	require.NoError(t, err)

	// Process metrics
	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	// Verify that the process was tagged as critical
	rm := md.ResourceMetrics().At(0)
	sm := rm.ScopeMetrics().At(0)
	metric := sm.Metrics().At(0)
	dp := metric.Gauge().DataPoints().At(0)

	priority, found := dp.Attributes().Get(cfg.PriorityAttributeName)
	assert.True(t, found)
	assert.Equal(t, cfg.CriticalAttributeValue, priority.Str())
}

func TestProcessorTaggingByPattern(t *testing.T) {
	// Create test metrics with a process that should match the pattern
	md := createTestMetrics("kubeproxy", 0.1, 100*1024*1024)

	// Create the processor with a configuration to tag processes matching a pattern
	cfg := &Config{
		CriticalExecutablePatterns: []string{"kube.*", "docker.*"},
		PriorityAttributeName:      "nr.priority",
		CriticalAttributeValue:     "critical",
	}

	// Validate configuration and compile patterns
	err := cfg.Validate()
	require.NoError(t, err)

	// Create and run the processor
	mockConsumer := consumertest.NewNop()
	settings := component.TelemetrySettings{
		Logger:         zap.NewNop(),
		MeterProvider:  nil, // Using nil for testing
		TracerProvider: nil, // Using nil for testing
	}
	proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
	require.NoError(t, err)

	// Process metrics
	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	// Verify that the process was tagged as critical
	rm := md.ResourceMetrics().At(0)
	sm := rm.ScopeMetrics().At(0)
	metric := sm.Metrics().At(0)
	dp := metric.Gauge().DataPoints().At(0)

	priority, found := dp.Attributes().Get(cfg.PriorityAttributeName)
	assert.True(t, found)
	assert.Equal(t, cfg.CriticalAttributeValue, priority.Str())
}

func TestProcessorTaggingByCPUThreshold(t *testing.T) {
	// Create test metrics with a process that has high CPU utilization
	md := createTestMetrics("chrome", 0.35, 100*1024*1024)

	// Create the processor with CPU threshold configuration only
	cfg := &Config{
		CPUSteadyStateThreshold: 0.3, // Tag if CPU util > 0.3
		PriorityAttributeName:   "nr.priority",
		CriticalAttributeValue:  "critical",
	}

	// Validate configuration and compile patterns
	err := cfg.Validate()
	require.NoError(t, err)

	// Create and run the processor
	mockConsumer := consumertest.NewNop()
	settings := component.TelemetrySettings{
		Logger:         zap.NewNop(),
		MeterProvider:  nil, // Using nil for testing
		TracerProvider: nil, // Using nil for testing
	}
	proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
	require.NoError(t, err)

	// Process metrics
	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	// Verify that the process was tagged as critical due to high CPU
	rm := md.ResourceMetrics().At(0)
	sm := rm.ScopeMetrics().At(0)
	metric := sm.Metrics().At(0)
	dp := metric.Gauge().DataPoints().At(0)

	priority, found := dp.Attributes().Get(cfg.PriorityAttributeName)
	assert.True(t, found)
	assert.Equal(t, cfg.CriticalAttributeValue, priority.Str())
}

func TestProcessorTaggingByMemoryThreshold(t *testing.T) {
	// Create test metrics with a process that has high memory usage (500 MiB)
	md := createTestMetrics("firefox", 0.1, 500*1024*1024)

	// Create the processor with memory threshold configuration only
	cfg := &Config{
		MemoryRSSThresholdMiB:  400, // Tag if memory > 400 MiB
		PriorityAttributeName:  "nr.priority",
		CriticalAttributeValue: "critical",
	}

	// Validate configuration and compile patterns
	err := cfg.Validate()
	require.NoError(t, err)

	// Create and run the processor
	mockConsumer := consumertest.NewNop()
	settings := component.TelemetrySettings{
		Logger:         zap.NewNop(),
		MeterProvider:  nil, // Using nil for testing
		TracerProvider: nil, // Using nil for testing
	}
	proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
	require.NoError(t, err)

	// Process metrics
	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	// Verify that the process was tagged as critical due to high memory
	rm := md.ResourceMetrics().At(0)
	sm := rm.ScopeMetrics().At(0)
	metric := sm.Metrics().At(0)
	dp := metric.Gauge().DataPoints().At(0)

	priority, found := dp.Attributes().Get(cfg.PriorityAttributeName)
	assert.True(t, found)
	assert.Equal(t, cfg.CriticalAttributeValue, priority.Str())
}

func TestProcessorNoMatch(t *testing.T) {
	// Create test metrics with a process that does not match
	md := createTestMetrics("notepad", 0.1, 10*1024*1024)

	// Create the processor with a configuration
	cfg := &Config{
		CriticalExecutables:        []string{"kubelet", "systemd"},
		CriticalExecutablePatterns: []string{"kube.*", "docker.*"},
		CPUSteadyStateThreshold:    0.3,
		MemoryRSSThresholdMiB:      400,
		PriorityAttributeName:      "nr.priority",
		CriticalAttributeValue:     "critical",
	}

	// Validate configuration and compile patterns
	err := cfg.Validate()
	require.NoError(t, err)

	// Create and run the processor
	mockConsumer := consumertest.NewNop()
	settings := component.TelemetrySettings{
		Logger:         zap.NewNop(),
		MeterProvider:  nil, // Using nil for testing
		TracerProvider: nil, // Using nil for testing
	}
	proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
	require.NoError(t, err)

	// Process metrics
	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	// Verify that the process was NOT tagged as critical
	rm := md.ResourceMetrics().At(0)
	sm := rm.ScopeMetrics().At(0)
	metric := sm.Metrics().At(0)
	dp := metric.Gauge().DataPoints().At(0)

	_, found := dp.Attributes().Get(cfg.PriorityAttributeName)
	assert.False(t, found)

	// We only care that the attribute wasn't found, not the actual value of an empty pcommon.Value
	// as implementations may differ
}

func TestProcessorIdempotency(t *testing.T) {
	// Create test metrics with a process that is already tagged as critical
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName("test.metric")
	dp := metric.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetDoubleValue(42)
	dp.Attributes().PutStr(processExecutableNameKey, "notepad")
	dp.Attributes().PutStr("nr.priority", "critical")

	// Create the processor
	cfg := &Config{
		CriticalExecutables:    []string{"kubelet", "systemd"},
		PriorityAttributeName:  "nr.priority",
		CriticalAttributeValue: "critical",
	}

	// Validate configuration and compile patterns
	err := cfg.Validate()
	require.NoError(t, err)

	// Create and run the processor
	mockConsumer := consumertest.NewNop()
	settings := component.TelemetrySettings{
		Logger:         zap.NewNop(),
		MeterProvider:  nil, // Using nil for testing
		TracerProvider: nil, // Using nil for testing
	}
	proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
	require.NoError(t, err)

	// Process metrics
	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	// Verify that the process is still tagged as critical
	rm = md.ResourceMetrics().At(0)
	sm = rm.ScopeMetrics().At(0)
	metric = sm.Metrics().At(0)
	dp = metric.Gauge().DataPoints().At(0)

	priority, found := dp.Attributes().Get(cfg.PriorityAttributeName)
	assert.True(t, found)
	assert.Equal(t, cfg.CriticalAttributeValue, priority.Str())
}

// Helper function to create test metrics with process info
func createTestMetrics(executableName string, cpuUtil float64, memRSSBytes int64) pmetric.Metrics {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("host.name", "test-host")

	sm := rm.ScopeMetrics().AppendEmpty()
	sm.Scope().SetName("test.scope")

	// Create a CPU metric with both CPU and memory attributes on the same datapoint
	// This is important for testing CPU and memory thresholds as the checker looks at individual datapoints
	cpuMetric := sm.Metrics().AppendEmpty()
	cpuMetric.SetName("process.cpu.utilization")
	cpuDP := cpuMetric.SetEmptyGauge().DataPoints().AppendEmpty()
	cpuDP.SetDoubleValue(cpuUtil)
	cpuDP.Attributes().PutStr(processExecutableNameKey, executableName)
	cpuDP.Attributes().PutStr("process.pid", "12345")
	cpuDP.Attributes().PutDouble(processCPUUtilizationKey, cpuUtil)
	cpuDP.Attributes().PutInt(processMemoryRSSKey, memRSSBytes)

	return md
}
