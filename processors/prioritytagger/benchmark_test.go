package prioritytagger

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

// BenchmarkProcessorWithManyProcesses benchmarks the processor with a large number of processes
func BenchmarkProcessorWithManyProcesses(b *testing.B) {
	// Test with different numbers of processes
	for _, numProcesses := range []int{10, 100, 1000} {
		b.Run(fmt.Sprintf("Processes-%d", numProcesses), func(b *testing.B) {
			// Create a realistic configuration
			cfg := &Config{
				CriticalExecutables:        []string{"systemd", "containerd", "kubelet"},
				CriticalExecutablePatterns: []string{"kube.*", "docker.*", "containerd.*"},
				CPUSteadyStateThreshold:    0.7,
				MemoryRSSThresholdMiB:      1024,
				PriorityAttributeName:      "nr.priority",
				CriticalAttributeValue:     "critical",
			}
			require.NoError(b, cfg.Validate())

			// Create metrics with many processes - about 10% will be "critical"
			md := createBenchmarkMetrics(numProcesses)

			// Create the processor
			mockConsumer := consumertest.NewNop()
			settings := component.TelemetrySettings{
				Logger:         zap.NewNop(),
				MeterProvider:  nil, // Using nil for testing
				TracerProvider: nil, // Using nil for testing
			}
			proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
			require.NoError(b, err)

			// Start the processor
			err = proc.Start(context.Background(), nil) // Using nil for NopHost
			require.NoError(b, err)

			// Reset the timer to exclude setup time
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				err = proc.ConsumeMetrics(context.Background(), md)
				require.NoError(b, err)
			}

			// Stop the timer and shutdown
			b.StopTimer()
			err = proc.Shutdown(context.Background())
			require.NoError(b, err)
		})
	}
}

// createBenchmarkMetrics creates a large number of process metrics for benchmarking
func createBenchmarkMetrics(numProcesses int) pmetric.Metrics {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()

	// Create CPU utilization metric
	cpuMetric := sm.Metrics().AppendEmpty()
	cpuMetric.SetName("process.cpu.utilization")
	cpuMetric.SetUnit("1")

	// Create memory RSS metric
	memMetric := sm.Metrics().AppendEmpty()
	memMetric.SetName("process.memory.rss")
	memMetric.SetUnit("bytes")

	// Create metrics for many processes
	for i := 0; i < numProcesses; i++ {
		// Determine process characteristics
		var procName string
		var cpuUtil float64
		var memoryBytes int64

		// Make about 10% of processes match our critical criteria
		if i%10 == 0 {
			// Critical process - either by name, high CPU, or high memory
			switch i % 3 {
			case 0:
				// Critical by name
				criticalNames := []string{"systemd", "containerd", "kubelet", "kube-proxy", "docker"}
				procName = criticalNames[i%len(criticalNames)]
				cpuUtil = 0.1
				memoryBytes = 100 * 1024 * 1024
			case 1:
				// Critical by CPU
				procName = fmt.Sprintf("high-cpu-process-%d", i)
				cpuUtil = 0.8
				memoryBytes = 100 * 1024 * 1024
			case 2:
				// Critical by memory
				procName = fmt.Sprintf("high-mem-process-%d", i)
				cpuUtil = 0.1
				memoryBytes = 2000 * 1024 * 1024
			}
		} else {
			// Regular, non-critical process
			procName = fmt.Sprintf("process-%d", i)
			cpuUtil = 0.1
			memoryBytes = 100 * 1024 * 1024
		}

		// Add CPU datapoint
		cpuDP := cpuMetric.SetEmptyGauge().DataPoints().AppendEmpty()
		cpuDP.SetDoubleValue(cpuUtil)
		cpuDP.Attributes().PutStr(processExecutableNameKey, procName)
		cpuDP.Attributes().PutStr("process.pid", fmt.Sprintf("%d", i))

		// Add memory datapoint
		memDP := memMetric.SetEmptyGauge().DataPoints().AppendEmpty()
		memDP.SetIntValue(memoryBytes)
		memDP.Attributes().PutStr(processExecutableNameKey, procName)
		memDP.Attributes().PutStr("process.pid", fmt.Sprintf("%d", i))
	}

	return md
}

// BenchmarkRegexMatching specifically benchmarks the regex matching performance
// since this could be a bottleneck with many processes and complex patterns
func BenchmarkRegexMatching(b *testing.B) {
	// Test with different numbers of regex patterns
	for _, numPatterns := range []int{1, 5, 10, 20} {
		b.Run(fmt.Sprintf("Patterns-%d", numPatterns), func(b *testing.B) {
			// Create a configuration with many regex patterns
			cfg := &Config{
				CriticalExecutablePatterns: make([]string, numPatterns),
				PriorityAttributeName:      "nr.priority",
				CriticalAttributeValue:     "critical",
			}

			// Generate patterns like "process-\\d+", "java-\\d+", etc.
			for i := 0; i < numPatterns; i++ {
				cfg.CriticalExecutablePatterns[i] = fmt.Sprintf("pattern%d-.*", i)
			}
			require.NoError(b, cfg.Validate())

			// Create metrics with process names that don't match (worst case)
			md := pmetric.NewMetrics()
			rm := md.ResourceMetrics().AppendEmpty()
			sm := rm.ScopeMetrics().AppendEmpty()
			metric := sm.Metrics().AppendEmpty()
			metric.SetName("process.cpu.utilization")
			for i := 0; i < 100; i++ {
				dp := metric.SetEmptyGauge().DataPoints().AppendEmpty()
				dp.SetDoubleValue(0.1)
				dp.Attributes().PutStr(processExecutableNameKey, fmt.Sprintf("nomatch-%d", i))
				dp.Attributes().PutStr("process.pid", fmt.Sprintf("%d", i))
			}

			// Create the processor
			mockConsumer := consumertest.NewNop()
			settings := component.TelemetrySettings{
				Logger:         zap.NewNop(),
				MeterProvider:  nil, // Using nil for testing
				TracerProvider: nil, // Using nil for testing
			}
			proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
			require.NoError(b, err)

			// Reset the timer to exclude setup time
			b.ResetTimer()

			// Run the benchmark
			for i := 0; i < b.N; i++ {
				err = proc.ConsumeMetrics(context.Background(), md)
				require.NoError(b, err)
			}
		})
	}
}
