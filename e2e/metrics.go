package e2e

import (
	"time"

	"github.com/deepaucksharma/reservoir"
)

// SimpleMetricsReporter for testing
type SimpleMetricsReporter struct {
	reservoirSize   float64
	sampledSpans    float64
	traceBufferSize float64
	evictions       float64
	checkpointAge   float64
	dbSize          float64
	compactions     float64
}

func NewSimpleMetricsReporter() *SimpleMetricsReporter {
	return &SimpleMetricsReporter{}
}

func (m *SimpleMetricsReporter) ReportReservoirSize(size int) {
	m.reservoirSize = float64(size)
}

func (m *SimpleMetricsReporter) ReportSampledSpans(count int) {
	m.sampledSpans += float64(count)
}

func (m *SimpleMetricsReporter) ReportTraceBufferSize(size int) {
	m.traceBufferSize = float64(size)
}

func (m *SimpleMetricsReporter) ReportEvictions(count int) {
	m.evictions += float64(count)
}

func (m *SimpleMetricsReporter) ReportCheckpointAge(age time.Duration) {
	m.checkpointAge = age.Seconds()
}

func (m *SimpleMetricsReporter) ReportDBSize(sizeBytes int64) {
	m.dbSize = float64(sizeBytes)
}

func (m *SimpleMetricsReporter) ReportCompactions(count int) {
	m.compactions += float64(count)
}

func (m *SimpleMetricsReporter) GetReservoirSize() float64 {
	return m.reservoirSize
}

func (m *SimpleMetricsReporter) GetSampledSpans() float64 {
	return m.sampledSpans
}

func (m *SimpleMetricsReporter) GetTraceBufferSize() float64 {
	return m.traceBufferSize
}

func (m *SimpleMetricsReporter) GetEvictions() float64 {
	return m.evictions
}

func (m *SimpleMetricsReporter) GetCheckpointAge() float64 {
	return m.checkpointAge
}

func (m *SimpleMetricsReporter) GetDBSize() float64 {
	return m.dbSize
}

func (m *SimpleMetricsReporter) GetCompactions() float64 {
	return m.compactions
}

// Gauge functions for BadgerCheckpointManager
func (m *SimpleMetricsReporter) GetCheckpointAgeGauge() func(float64) {
	return func(v float64) {
		m.checkpointAge = v
	}
}

func (m *SimpleMetricsReporter) GetReservoirDbSizeGauge() func(float64) {
	return func(v float64) {
		m.dbSize = v
	}
}

func (m *SimpleMetricsReporter) GetCompactionCountCounter() func(float64) {
	return func(v float64) {
		m.compactions += v
	}
}

// Create a test span
func createTestSpan(traceID, spanID string) reservoir.SpanData {
	now := time.Now().UnixNano()
	return reservoir.SpanData{
		ID:        spanID,
		TraceID:   traceID,
		Name:      "test-span",
		StartTime: now,
		EndTime:   now + 1000000,
		Attributes: map[string]interface{}{
			"test.attribute": "value",
		},
		ResourceInfo: reservoir.ResourceInfo{
			Attributes: map[string]interface{}{
				"service.name": "test-service",
			},
		},
		ScopeInfo: reservoir.ScopeInfo{
			Name:    "test-scope",
			Version: "1.0",
		},
	}
}