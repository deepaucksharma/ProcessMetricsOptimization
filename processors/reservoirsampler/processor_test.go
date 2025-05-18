package reservoirsampler

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
)

func createSamplerTestMetrics(numProcs int, criticalIdx int, cfg *Config) pmetric.Metrics {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()

	for i := 0; i < numProcs; i++ {
		m := sm.Metrics().AppendEmpty()
		m.SetName(fmt.Sprintf("test.metric.%d", i))
		dp := m.SetEmptyGauge().DataPoints().AppendEmpty()
		dp.SetIntValue(int64(i))
		// Use PID as the identity attribute for simplicity in test
		dp.Attributes().PutStr("process.pid", strconv.Itoa(i))
		dp.Attributes().PutStr("process.executable.name", fmt.Sprintf("proc-%d", i))
		if i == criticalIdx {
			dp.Attributes().PutStr(cfg.PriorityAttributeName, cfg.CriticalAttributeValue)
		}
	}
	return md
}

func TestReservoirSampler_BasicSampling(t *testing.T) {
	cfg := &Config{
		ReservoirSize:           2, // Small reservoir size for testing
		IdentityAttributes:      []string{"process.pid"},
		SampledAttributeName:    "sampled",
		SampledAttributeValue:   "yes",
		SampleRateAttributeName: "rate",
		PriorityAttributeName:   "prio",
		CriticalAttributeValue:  "crit",
	}
	require.NoError(t, cfg.Validate())

	nextSink := new(consumertest.MetricsSink)
	settings := processor.CreateSettings{
		ID:                component.NewID(typeStr),
		TelemetrySettings: componenttest.NewNopTelemetrySettings(),
		BuildInfo:         component.NewDefaultBuildInfo(),
	}
	proc, err := newReservoirSamplerProcessor(settings, nextSink, cfg)
	require.NoError(t, err)

	// Send 5 processes, 1 critical (idx 0). Reservoir size is 2.
	// Expect critical (idx 0) + up to 2 sampled from (1,2,3,4)
	md := createSamplerTestMetrics(5, 0, cfg)

	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	processedMetrics := nextSink.AllMetrics()
	require.Len(t, processedMetrics, 1)

	outputMetricCount := 0
	sampledCount := 0
	criticalFound := false

	rms := processedMetrics[0].ResourceMetrics()
	for i := 0; i < rms.Len(); i++ {
		sms := rms.At(i).ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			ms := sms.At(j).Metrics()
			outputMetricCount += ms.Len()
			for k := 0; k < ms.Len(); k++ {
				metric := ms.At(k)
				var dps pmetric.NumberDataPointSlice
				if metric.Type() == pmetric.MetricTypeGauge {
					dps = metric.Gauge().DataPoints()
				} else {
					continue
				}

				for l := 0; l < dps.Len(); l++ {
					dp := dps.At(l)
					attrs := dp.Attributes()
					if IsCritical(attrs, cfg) {
						criticalFound = true
						pidVal, _ := attrs.Get("process.pid")
						assert.Equal(t, "0", pidVal.Str(), "Critical process PID should be 0")
					} else if IsSampled(attrs, cfg) {
						sampledCount++
						_, rateFound := attrs.Get(cfg.SampleRateAttributeName)
						assert.True(t, rateFound, "Sample rate attribute should be present")
					}
				}
			}
		}
	}

	assert.True(t, criticalFound, "Critical process should be passed through")
	assert.Equal(t, cfg.ReservoirSize, sampledCount, "Number of sampled processes should match reservoir size")
	assert.Equal(t, 1+cfg.ReservoirSize, outputMetricCount, "Total output metrics should be 1 critical + 2 sampled")

	// Send more data to see reservoir behavior with a new batch
	nextSink.Reset()
	mdMore := createSamplerTestMetrics(10, -1, cfg) // No criticals this time
	err = proc.ConsumeMetrics(context.Background(), mdMore)
	require.NoError(t, err)

	processedMetricsMore := nextSink.AllMetrics()
	require.Len(t, processedMetricsMore, 1, "Should have received one batch of metrics")

	sampledCountMore := 0
	rmsMore := processedMetricsMore[0].ResourceMetrics()
	for i := 0; i < rmsMore.Len(); i++ {
		sms := rmsMore.At(i).ScopeMetrics()
		for j := 0; j < sms.Len(); j++ {
			ms := sms.At(j).Metrics()
			for k := 0; k < ms.Len(); k++ {
				metric := ms.At(k)
				var dps pmetric.NumberDataPointSlice
				if metric.Type() == pmetric.MetricTypeGauge {
					dps = metric.Gauge().DataPoints()
				} else {
					continue
				}

				for l := 0; l < dps.Len(); l++ {
					if IsSampled(dps.At(l).Attributes(), cfg) {
						sampledCountMore++
					}
				}
			}
		}
	}

	assert.Equal(t, cfg.ReservoirSize, sampledCountMore, "Sampled count after more data should still be reservoir size")
}

func IsCritical(attrs pcommon.Map, cfg *Config) bool {
	val, exists := attrs.Get(cfg.PriorityAttributeName)
	return exists && val.Str() == cfg.CriticalAttributeValue
}

func IsSampled(attrs pcommon.Map, cfg *Config) bool {
	val, exists := attrs.Get(cfg.SampledAttributeName)
	return exists && val.Str() == cfg.SampledAttributeValue
}
