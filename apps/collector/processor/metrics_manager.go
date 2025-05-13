package processor

import (
	"context"
	"time"

	"go.uber.org/atomic"
)

// MetricsManager handles metrics reporting for the reservoir sampler
type MetricsManager struct {
	ctx context.Context
	
	// Gauges
	reservoirSize    *atomic.Float64
	checkpointAge    *atomic.Float64
	reservoirDbSize  *atomic.Float64
	traceBufferSize  *atomic.Float64
	
	// Counters
	sampledSpans     *atomic.Float64
	lruEvictions     *atomic.Float64
	compactionCount  *atomic.Float64
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager(ctx context.Context) *MetricsManager {
	return &MetricsManager{
		ctx:              ctx,
		reservoirSize:    atomic.NewFloat64(0),
		checkpointAge:    atomic.NewFloat64(0),
		reservoirDbSize:  atomic.NewFloat64(0),
		traceBufferSize:  atomic.NewFloat64(0),
		sampledSpans:     atomic.NewFloat64(0),
		lruEvictions:     atomic.NewFloat64(0),
		compactionCount:  atomic.NewFloat64(0),
	}
}

// RegisterMetrics registers all metrics with the meter - no-op implementation
func (m *MetricsManager) RegisterMetrics() error {
	// No-op implementation for simplified testing
	return nil
}

// Implement reservoir.MetricsReporter interface

// ReportReservoirSize reports the current size of the reservoir
func (m *MetricsManager) ReportReservoirSize(size int) {
	m.reservoirSize.Store(float64(size))
}

// ReportSampledSpans reports the number of spans that were sampled
func (m *MetricsManager) ReportSampledSpans(count int) {
	m.sampledSpans.Add(float64(count))
}

// ReportTraceBufferSize reports the current size of the trace buffer
func (m *MetricsManager) ReportTraceBufferSize(size int) {
	m.traceBufferSize.Store(float64(size))
}

// ReportEvictions reports the number of trace evictions from the buffer
func (m *MetricsManager) ReportEvictions(count int) {
	m.lruEvictions.Add(float64(count))
}

// ReportCheckpointAge reports the age of the last checkpoint
func (m *MetricsManager) ReportCheckpointAge(age time.Duration) {
	m.checkpointAge.Store(age.Seconds())
}

// ReportDBSize reports the size of the checkpoint storage
func (m *MetricsManager) ReportDBSize(sizeBytes int64) {
	m.reservoirDbSize.Store(float64(sizeBytes))
}

// ReportCompactions reports the number of storage compactions
func (m *MetricsManager) ReportCompactions(count int) {
	m.compactionCount.Add(float64(count))
}

// Getter methods for use by other components

// GetReservoirSizeGauge returns the gauge func for reservoir size
func (m *MetricsManager) GetReservoirSizeGauge() func(float64) {
	return func(v float64) {
		m.reservoirSize.Store(v)
	}
}

// GetCheckpointAgeGauge returns the gauge func for checkpoint age
func (m *MetricsManager) GetCheckpointAgeGauge() func(float64) {
	return func(v float64) {
		m.checkpointAge.Store(v)
	}
}

// GetReservoirDbSizeGauge returns the gauge func for DB size
func (m *MetricsManager) GetReservoirDbSizeGauge() func(float64) {
	return func(v float64) {
		m.reservoirDbSize.Store(v)
	}
}

// GetSampledSpansCounter returns the counter func for sampled spans
func (m *MetricsManager) GetSampledSpansCounter() func(float64) {
	return func(v float64) {
		m.sampledSpans.Add(v)
	}
}

// GetLruEvictionsCounter returns the counter func for LRU evictions
func (m *MetricsManager) GetLruEvictionsCounter() func(float64) {
	return func(v float64) {
		m.lruEvictions.Add(v)
	}
}

// GetCompactionCountCounter returns the counter func for DB compactions
func (m *MetricsManager) GetCompactionCountCounter() func(float64) {
	return func(v float64) {
		m.compactionCount.Add(v)
	}
}