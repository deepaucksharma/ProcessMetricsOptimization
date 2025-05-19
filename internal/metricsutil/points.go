package metricsutil

import "go.opentelemetry.io/collector/pdata/pmetric"

// CountPoints counts the total number of data points contained in a Metrics collection.
func CountPoints(md pmetric.Metrics) int {
	count := 0
	for i := 0; i < md.ResourceMetrics().Len(); i++ {
		rm := md.ResourceMetrics().At(i)
		for j := 0; j < rm.ScopeMetrics().Len(); j++ {
			sm := rm.ScopeMetrics().At(j)
			for k := 0; k < sm.Metrics().Len(); k++ {
				metric := sm.Metrics().At(k)
				switch metric.Type() {
				case pmetric.MetricTypeGauge:
					count += metric.Gauge().DataPoints().Len()
				case pmetric.MetricTypeSum:
					count += metric.Sum().DataPoints().Len()
				case pmetric.MetricTypeHistogram:
					count += metric.Histogram().DataPoints().Len()
				case pmetric.MetricTypeSummary:
					count += metric.Summary().DataPoints().Len()
				case pmetric.MetricTypeExponentialHistogram:
					count += metric.ExponentialHistogram().DataPoints().Len()
				}
			}
		}
	}
	return count
}
