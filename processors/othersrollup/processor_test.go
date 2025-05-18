package othersrollup

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
)

func TestOthersRollup_SimpleSum(t *testing.T) {
	cfg := &Config{
		OutputPIDAttributeValue:            "_other_pid_",
		OutputExecutableNameAttributeValue: "_other_exe_",
		Aggregations: map[string]AggregationType{
			"process.memory.rss": SumAggregation,
		},
		MetricsToRollup:        []string{"process.memory.rss"}, // Explicitly target this metric
		PriorityAttributeName:  "nr.priority",
		CriticalAttributeValue: "critical",
	}
	require.NoError(t, cfg.Validate())

	nextSink := new(consumertest.MetricsSink)
	settings := processor.CreateSettings{
		ID:                component.NewID(typeStr),
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}

	proc, err := newOthersRollupProcessor(settings, nextSink, cfg)
	require.NoError(t, err)

	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("host.name", "test-host-1")
	sm := rm.ScopeMetrics().AppendEmpty()

	// Metric to be rolled up
	memMetric := sm.Metrics().AppendEmpty()
	memMetric.SetName("process.memory.rss")
	memMetric.SetEmptySum().SetIsMonotonic(true) // Example: cumulative sum

	dp1 := memMetric.Sum().DataPoints().AppendEmpty()
	dp1.SetIntValue(100)
	dp1.Attributes().PutStr(processPIDKey, "10")
	dp1.Attributes().PutStr(processExecutableNameKey, "procA")

	dp2 := memMetric.Sum().DataPoints().AppendEmpty()
	dp2.SetIntValue(200)
	dp2.Attributes().PutStr(processPIDKey, "11")
	dp2.Attributes().PutStr(processExecutableNameKey, "procB")

	// Critical metric, should pass through
	cpuMetric := sm.Metrics().AppendEmpty()
	cpuMetric.SetName("process.cpu.utilization")
	cpuMetric.SetEmptyGauge()
	dpCrit := cpuMetric.Gauge().DataPoints().AppendEmpty()
	dpCrit.SetDoubleValue(0.5)
	dpCrit.Attributes().PutStr(processPIDKey, "1")
	dpCrit.Attributes().PutStr(processExecutableNameKey, "critical_proc")
	dpCrit.Attributes().PutStr(cfg.PriorityAttributeName, cfg.CriticalAttributeValue)

	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	processedMetrics := nextSink.AllMetrics()
	require.Len(t, processedMetrics, 1, "Should have received one batch of metrics")

	outputRm := processedMetrics[0].ResourceMetrics().At(0)
	outputSm := outputRm.ScopeMetrics().At(0)

	assert.Equal(t, 2, outputSm.Metrics().Len(), "Should have 2 metrics: one rolled-up, one critical pass-through")

	foundRolledUp := false
	foundCritical := false

	for i := 0; i < outputSm.Metrics().Len(); i++ {
		m := outputSm.Metrics().At(i)
		if m.Name() == "process.memory.rss" {
			foundRolledUp = true
			assert.Equal(t, pmetric.MetricTypeSum, m.Type(), "Type should be preserved for sum")
			dps := m.Sum().DataPoints()
			assert.Equal(t, 1, dps.Len(), "Should have 1 rolled up datapoint")
			rollupDp := dps.At(0)
			assert.Equal(t, 300.0, rollupDp.DoubleValue(), "Sum should be 100 + 200 = 300")
			pid, _ := rollupDp.Attributes().Get(processPIDKey)
			exe, _ := rollupDp.Attributes().Get(processExecutableNameKey)
			assert.Equal(t, cfg.OutputPIDAttributeValue, pid.Str(), "PID should be _other_pid_")
			assert.Equal(t, cfg.OutputExecutableNameAttributeValue, exe.Str(), "Executable name should be _other_exe_")
		} else if m.Name() == "process.cpu.utilization" {
			foundCritical = true
			assert.Equal(t, pmetric.MetricTypeGauge, m.Type(), "Type should be gauge")
			dps := m.Gauge().DataPoints()
			assert.Equal(t, 1, dps.Len(), "Should have 1 critical datapoint")
			critDpOut := dps.At(0)
			assert.Equal(t, 0.5, critDpOut.DoubleValue(), "Critical value should be preserved")
			pid, _ := critDpOut.Attributes().Get(processPIDKey)
			assert.Equal(t, "1", pid.Str(), "Critical PID should be preserved")
		}
	}
	assert.True(t, foundRolledUp, "Rolled up metric not found")
	assert.True(t, foundCritical, "Critical pass-through metric not found")
}

func TestOthersRollup_Average(t *testing.T) {
	cfg := &Config{
		OutputPIDAttributeValue:            "_other_pid_",
		OutputExecutableNameAttributeValue: "_other_exe_",
		Aggregations: map[string]AggregationType{
			"process.cpu.utilization": AvgAggregation,
		},
		PriorityAttributeName:  "nr.priority",
		CriticalAttributeValue: "critical",
	}
	// MetricsToRollup is empty, so all non-priority gauges/sums are candidates
	require.NoError(t, cfg.Validate())

	nextSink := new(consumertest.MetricsSink)
	settings := processor.CreateSettings{
		ID:                component.NewID(typeStr),
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}
	proc, err := newOthersRollupProcessor(settings, nextSink, cfg)
	require.NoError(t, err)

	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()

	cpuMetric := sm.Metrics().AppendEmpty()
	cpuMetric.SetName("process.cpu.utilization")
	cpuMetric.SetEmptyGauge()

	dp1 := cpuMetric.Gauge().DataPoints().AppendEmpty()
	dp1.SetDoubleValue(0.2)
	dp1.Attributes().PutStr(processPIDKey, "20")

	dp2 := cpuMetric.Gauge().DataPoints().AppendEmpty()
	dp2.SetDoubleValue(0.4)
	dp2.Attributes().PutStr(processPIDKey, "21")

	dp3 := cpuMetric.Gauge().DataPoints().AppendEmpty()
	dp3.SetDoubleValue(0.6) // This one is critical, should pass through
	dp3.Attributes().PutStr(processPIDKey, "22")
	dp3.Attributes().PutStr(cfg.PriorityAttributeName, cfg.CriticalAttributeValue)

	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	processedMetrics := nextSink.AllMetrics()
	require.Len(t, processedMetrics, 1, "Should have received one batch of metrics")
	outputSm := processedMetrics[0].ResourceMetrics().At(0).ScopeMetrics().At(0)
	assert.Equal(t, 2, outputSm.Metrics().Len(), "Should have 2 metrics: rolled up + critical")

	foundRolledUpAvg := false
	foundCriticalPassThrough := false

	for i := 0; i < outputSm.Metrics().Len(); i++ {
		m := outputSm.Metrics().At(i)
		require.Equal(t, "process.cpu.utilization", m.Name(), "Should be cpu utilization metric")
		dps := m.Gauge().DataPoints()
		require.Equal(t, 1, dps.Len(), "Should have 1 datapoint")
		dp := dps.At(0)

		pidVal, _ := dp.Attributes().Get(processPIDKey)

		if pidVal.Str() == cfg.OutputPIDAttributeValue {
			foundRolledUpAvg = true
			assert.InDelta(t, 0.3, dp.DoubleValue(), 0.001, "Average should be (0.2+0.4)/2 = 0.3")
		} else if pidVal.Str() == "22" {
			foundCriticalPassThrough = true
			assert.Equal(t, 0.6, dp.DoubleValue(), "Critical value should be 0.6")
		}
	}
	assert.True(t, foundRolledUpAvg, "Rolled up average not found")
	assert.True(t, foundCriticalPassThrough, "Critical pass-through not found")
}
