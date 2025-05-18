package metricutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pmetric"
)

func TestGetNumericValue(t *testing.T) {
	dps := pmetric.NewNumberDataPointSlice()
	intDP := dps.AppendEmpty()
	intDP.SetIntValue(10)
	assert.Equal(t, 10.0, GetNumericValue(intDP))

	doubleDP := dps.AppendEmpty()
	doubleDP.SetDoubleValue(1.5)
	assert.Equal(t, 1.5, GetNumericValue(doubleDP))
}

func TestMetricPointCount(t *testing.T) {
	md := pmetric.NewMetrics()
	rm := md.ResourceMetrics().AppendEmpty()
	sm := rm.ScopeMetrics().AppendEmpty()

	// Gauge with two points
	g := sm.Metrics().AppendEmpty()
	g.SetName("gauge")
	g.SetEmptyGauge().DataPoints().AppendEmpty()
	g.Gauge().DataPoints().AppendEmpty()

	// Sum with one point
	s := sm.Metrics().AppendEmpty()
	s.SetName("sum")
	s.SetEmptySum().DataPoints().AppendEmpty()

	// Histogram with one point
	h := sm.Metrics().AppendEmpty()
	h.SetName("hist")
	h.SetEmptyHistogram().DataPoints().AppendEmpty()

	// Exponential histogram with one point
	eh := sm.Metrics().AppendEmpty()
	eh.SetName("exp_hist")
	eh.SetEmptyExponentialHistogram().DataPoints().AppendEmpty()

	// Summary with one point
	sum := sm.Metrics().AppendEmpty()
	sum.SetName("summary")
	sum.SetEmptySummary().DataPoints().AppendEmpty()

	assert.Equal(t, 6, MetricPointCount(md))
}
