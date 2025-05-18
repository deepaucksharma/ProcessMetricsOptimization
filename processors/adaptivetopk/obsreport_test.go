package adaptivetopk

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/processor"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

func TestCurrentKValueCallback(t *testing.T) {
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))

	settings := processor.CreateSettings{
		ID: component.NewID(typeStr),
		TelemetrySettings: component.TelemetrySettings{
			MeterProvider: mp,
		},
		BuildInfo: component.NewDefaultBuildInfo(),
	}

	obs, err := newAdaptiveTopKObsreport(settings.TelemetrySettings)
	require.NoError(t, err)

	obs.recordCurrentKValue(context.Background(), 7)

	var rm metricdata.ResourceMetrics
	require.NoError(t, reader.Collect(context.Background(), &rm))

	require.Len(t, rm.ScopeMetrics, 1)
	sm := rm.ScopeMetrics[0]
	require.Len(t, sm.Metrics, 1)
	gauge, ok := sm.Metrics[0].Data.(metricdata.Gauge[int64])
	require.True(t, ok)
	require.Len(t, gauge.DataPoints, 1)
	assert.Equal(t, int64(7), gauge.DataPoints[0].Value)
}
