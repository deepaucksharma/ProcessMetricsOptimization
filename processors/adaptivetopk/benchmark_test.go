package adaptivetopk

import (
	"context"
	"testing"
	"time"

	"github.com/newrelic/nrdot-process-optimization/internal/metricsutil"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
)

// generateTestMetrics creates a metrics batch with the specified number of processes
func generateTestMetrics(numProcesses int, numMetricsPerProcess int, hasCritical bool) pmetric.Metrics {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()

	// Add host load metric
	hostMetric := sm.Metrics().AppendEmpty()
	hostMetric.SetName("system.cpu.utilization")
	hostDP := hostMetric.SetEmptyGauge().DataPoints().AppendEmpty()
	hostDP.SetDoubleValue(0.5)

	// Create processes
	for i := 0; i < numProcesses; i++ {
		pid := string(rune(i + 48)) // Convert to ASCII digit

		for j := 0; j < numMetricsPerProcess; j++ {
			var metricName string
			switch j % 3 {
			case 0:
				metricName = "process.cpu.utilization"
			case 1:
				metricName = "process.memory.rss"
			case 2:
				metricName = "process.disk.io"
			}

			metric := sm.Metrics().AppendEmpty()
			metric.SetName(metricName)
			dp := metric.SetEmptyGauge().DataPoints().AppendEmpty()

			// Set different values per process
			if metricName == "process.cpu.utilization" {
				dp.SetDoubleValue(float64(i) * 0.01)
			} else if metricName == "process.memory.rss" {
				dp.SetIntValue(int64(i * 1000))
			} else {
				dp.SetDoubleValue(float64(i * 10))
			}

			dp.Attributes().PutStr(processPIDKey, pid)
			dp.Attributes().PutStr("process.executable.name", "proc"+pid)

			// Mark some processes as critical if required
			if hasCritical && i < numProcesses/10 { // Make 10% of processes critical
				dp.Attributes().PutStr("nr.priority", "critical")
			}
		}
	}

	return md
}

// BenchmarkAdaptiveTopKProcessor benchmarks the full processor with different process counts
func BenchmarkAdaptiveTopKProcessor(b *testing.B) {
	tests := []struct {
		name           string
		numProcesses   int
		metricsPerProc int
		kValue         int
		isDynamicK     bool
		hasHysteresis  bool
		hasCritical    bool
	}{
		{
			name:           "Small-Static-NoCritical",
			numProcesses:   50,
			metricsPerProc: 3,
			kValue:         10,
			isDynamicK:     false,
			hasHysteresis:  false,
			hasCritical:    false,
		},
		{
			name:           "Medium-Static-WithCritical",
			numProcesses:   200,
			metricsPerProc: 3,
			kValue:         20,
			isDynamicK:     false,
			hasHysteresis:  false,
			hasCritical:    true,
		},
		{
			name:           "Large-Static-WithCritical",
			numProcesses:   500,
			metricsPerProc: 3,
			kValue:         20,
			isDynamicK:     false,
			hasHysteresis:  false,
			hasCritical:    true,
		},
		{
			name:           "Medium-Dynamic-WithCritical",
			numProcesses:   200,
			metricsPerProc: 3,
			kValue:         20,
			isDynamicK:     true,
			hasHysteresis:  true,
			hasCritical:    true,
		},
		{
			name:           "Large-Dynamic-WithHysteresis",
			numProcesses:   500,
			metricsPerProc: 3,
			kValue:         20,
			isDynamicK:     true,
			hasHysteresis:  true,
			hasCritical:    true,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			// Create config
			cfg := &Config{
				KValue:                 tt.kValue,
				KeyMetricName:          "process.cpu.utilization",
				SecondaryKeyMetricName: "process.memory.rss",
				PriorityAttributeName:  "nr.priority",
				CriticalAttributeValue: "critical",
			}

			if tt.isDynamicK {
				cfg.HostLoadMetricName = "system.cpu.utilization"
				cfg.LoadBandsToKMap = map[float64]int{
					0.3: 10,
					0.6: 20,
					0.9: 30,
				}
				cfg.MinKValue = 5
				cfg.MaxKValue = 30
			}

			if tt.hasHysteresis {
				cfg.HysteresisDuration = 30 * time.Second
			}

			require.NoError(b, cfg.Validate())

			// Generate test data once for the entire benchmark
			md := generateTestMetrics(tt.numProcesses, tt.metricsPerProc, tt.hasCritical)

			// Reset timer for the actual benchmark
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				// Create a new processor for each iteration to avoid state contamination
				nextSink := new(consumertest.MetricsSink)
				settings := processor.CreateSettings{
					ID: component.NewID(typeStr),
					TelemetrySettings: component.TelemetrySettings{
						Logger:        nil, // No logger for benchmarks
						MeterProvider: nil, // No metrics for benchmarks
					},
					BuildInfo: component.NewDefaultBuildInfo(),
				}

				proc, err := newAdaptiveTopKProcessor(settings, nextSink, cfg)
				require.NoError(b, err)

				// Clone the metrics to avoid modifying the original
				mdClone := pmetric.NewMetrics()
				md.CopyTo(mdClone)

				// Run the processor
				err = proc.ConsumeMetrics(ctx, mdClone)
				require.NoError(b, err)
			}
		})
	}
}

// BenchmarkFindHostLoadMetric benchmarks the findHostLoadMetric function
func BenchmarkFindHostLoadMetric(b *testing.B) {
	proc := &adaptiveTopKProcessor{
		config: &Config{
			HostLoadMetricName: "system.cpu.utilization",
		},
	}

	// Generate metrics with different scales
	smallMetrics := generateTestMetrics(50, 3, false)
	mediumMetrics := generateTestMetrics(200, 3, false)
	largeMetrics := generateTestMetrics(500, 3, false)

	b.Run("SmallMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			proc.findHostLoadMetric(smallMetrics)
		}
	})

	b.Run("MediumMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			proc.findHostLoadMetric(mediumMetrics)
		}
	})

	b.Run("LargeMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			proc.findHostLoadMetric(largeMetrics)
		}
	})
}

// BenchmarkMetricPointCount benchmarks the CountPoints helper
func BenchmarkMetricPointCount(b *testing.B) {
	// Generate metrics with different scales
	smallMetrics := generateTestMetrics(50, 3, false)
	mediumMetrics := generateTestMetrics(200, 3, false)
	largeMetrics := generateTestMetrics(500, 3, false)

	b.Run("SmallMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metricsutil.CountPoints(smallMetrics)
		}
	})

	b.Run("MediumMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metricsutil.CountPoints(mediumMetrics)
		}
	})

	b.Run("LargeMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metricsutil.CountPoints(largeMetrics)
		}
	})
}

// BenchmarkEstimateProcessCount benchmarks the estimateProcessCount function
func BenchmarkEstimateProcessCount(b *testing.B) {
	// Generate metrics with different scales
	smallMetrics := generateTestMetrics(50, 3, false)
	mediumMetrics := generateTestMetrics(200, 3, false)
	largeMetrics := generateTestMetrics(500, 3, false)

	b.Run("SmallMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			estimateProcessCount(smallMetrics)
		}
	})

	b.Run("MediumMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			estimateProcessCount(mediumMetrics)
		}
	})

	b.Run("LargeMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			estimateProcessCount(largeMetrics)
		}
	})
}
