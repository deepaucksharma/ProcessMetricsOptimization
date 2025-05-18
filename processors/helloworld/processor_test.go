package helloworld

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.uber.org/zap"
)

func TestProcessor(t *testing.T) {
	// Create test metrics
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("resource.key", "resource-value")
	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName("test.metric")
	metric.SetDescription("test metric description")
	metric.SetUnit("count")
	dp := metric.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetIntValue(42)
	dp.SetTimestamp(pcommon.NewTimestampFromTime(componenttest.DefaultTime))
	dp.Attributes().PutStr("attribute.name", "attribute-value")

	// Create the processor
	cfg := &Config{
		Message:       "Test message",
		AddToResource: true,
	}

	mockConsumer := consumertest.NewNop()
	settings := component.TelemetrySettings{
		Logger:         zap.NewNop(),
		MeterProvider:  componenttest.NewNopMeterProvider(),
		TracerProvider: componenttest.NewNopTracerProvider(),
	}
	proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
	require.NoError(t, err)

	// Process metrics
	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	// Verify hello.processor attribute was added to the resource
	hello, found := rm.Resource().Attributes().Get("hello.processor")
	assert.True(t, found)
	assert.Equal(t, "Hello from OpenTelemetry!", hello.Str())

	// Verify hello.processor attribute was added to the datapoint
	hello, found = dp.Attributes().Get("hello.processor")
	assert.True(t, found)
	assert.Equal(t, "Hello from OpenTelemetry!", hello.Str())
}

func TestProcessorNoResourceAttributes(t *testing.T) {
	// Create test metrics
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	rm.Resource().Attributes().PutStr("resource.key", "resource-value")
	sm := rm.ScopeMetrics().AppendEmpty()
	metric := sm.Metrics().AppendEmpty()
	metric.SetName("test.metric")
	dp := metric.SetEmptyGauge().DataPoints().AppendEmpty()
	dp.SetIntValue(42)
	dp.Attributes().PutStr("attribute.name", "attribute-value")

	// Create the processor with AddToResource = false
	cfg := &Config{
		Message:       "Test message",
		AddToResource: false,
	}

	mockConsumer := consumertest.NewNop()
	settings := component.TelemetrySettings{
		Logger:         zap.NewNop(),
		MeterProvider:  componenttest.NewNopMeterProvider(),
		TracerProvider: componenttest.NewNopTracerProvider(),
	}
	proc, err := newProcessor(cfg, settings.Logger, mockConsumer, settings)
	require.NoError(t, err)

	// Process metrics
	err = proc.ConsumeMetrics(context.Background(), md)
	require.NoError(t, err)

	// Verify hello.processor attribute was NOT added to the resource
	hello, found := rm.Resource().Attributes().Get("hello.processor")
	assert.False(t, found)
	assert.Equal(t, pcommon.Value{}, hello)

	// Verify hello.processor attribute was added to the datapoint
	hello, found = dp.Attributes().Get("hello.processor")
	assert.True(t, found)
	assert.Equal(t, "Hello from OpenTelemetry!", hello.Str())
}
