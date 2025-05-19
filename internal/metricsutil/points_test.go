package metricsutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

// TestCountPoints verifies that CountPoints correctly counts metric data points.
func TestCountPoints(t *testing.T) {
	md := pmetric.NewMetrics()

	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()

	// Gauge with two data points
	m1 := sm.Metrics().AppendEmpty()
	m1.SetEmptyGauge()
	dp := m1.Gauge().DataPoints().AppendEmpty()
	dp.SetIntValue(10)
	dp = m1.Gauge().DataPoints().AppendEmpty()
	dp.SetIntValue(20)

	// Sum with one data point
	m2 := sm.Metrics().AppendEmpty()
	m2.SetEmptySum()
	dp = m2.Sum().DataPoints().AppendEmpty()
	dp.SetIntValue(5)

	// Histogram with three data points
	m3 := sm.Metrics().AppendEmpty()
	m3.SetEmptyHistogram()
	for i := 0; i < 3; i++ {
		dp := m3.Histogram().DataPoints().AppendEmpty()
		dp.SetCount(uint64(i))
	}

	// Ensure total count is 6
	assert.Equal(t, 6, CountPoints(md))
}
