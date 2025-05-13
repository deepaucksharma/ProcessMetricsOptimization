package main

import (
	"fmt"
	"time"

	"github.com/deepaucksharma/reservoir"
)

func main() {
	fmt.Println("Verifying trace-aware reservoir sampling end-to-end")

	// Create a simple metrics reporter to track metrics
	metrics := NewSimpleMetricsReporter()

	// Create a reservoir sampler with 10 spans capacity
	r := reservoir.NewAlgorithmR(10, metrics)
	fmt.Printf("Created reservoir with size: %d\n", r.MaxSize())

	// Create a time window
	window := reservoir.NewTimeWindow(5 * time.Second)
	fmt.Printf("Created window with duration: %s\n", window.GetDuration())

	// Create a trace buffer
	traceBuffer := reservoir.NewTraceBuffer(100, 1*time.Second, metrics)
	fmt.Printf("Created trace buffer with size: %d\n", 100)

	// Add some spans directly to the reservoir
	fmt.Println("\nAdding spans directly to reservoir...")
	for i := 0; i < 20; i++ {
		span := createTestSpan(fmt.Sprintf("trace-%d", i/2), fmt.Sprintf("span-%d", i))
		added := r.AddSpan(span)
		fmt.Printf("Added span %s (trace: %s): %v\n", span.ID, span.TraceID, added)
	}
	
	fmt.Printf("\nReservoir size: %d (should be 10)\n", r.Size())
	fmt.Printf("Sampled spans count: %.0f\n", metrics.GetSampledSpans())

	// Add spans to the trace buffer
	fmt.Println("\nAdding spans to trace buffer...")
	for i := 0; i < 10; i++ {
		traceID := fmt.Sprintf("trace-buffer-%d", i/3)
		span := createTestSpan(traceID, fmt.Sprintf("buffer-span-%d", i))
		traceBuffer.AddSpan(span)
		fmt.Printf("Added span %s to trace buffer (trace: %s)\n", span.ID, span.TraceID)
	}

	fmt.Printf("\nTrace buffer size: %.0f\n", metrics.GetTraceBufferSize())

	// Get completed traces after waiting
	fmt.Println("\nWaiting for trace timeout...")
	time.Sleep(1500 * time.Millisecond)
	completedTraces := traceBuffer.GetCompletedTraces()
	
	fmt.Printf("Got %d completed traces\n", len(completedTraces))
	for i, trace := range completedTraces {
		fmt.Printf("Trace %d: %d spans\n", i+1, len(trace))
		
		// Add these spans to the reservoir
		for _, span := range trace {
			r.AddSpan(span)
		}
	}

	// Get and print sample
	sample := r.GetSample()
	fmt.Printf("\nFinal reservoir sample contains %d spans:\n", len(sample))
	for i, span := range sample {
		fmt.Printf("%d. %s (trace: %s)\n", i+1, span.ID, span.TraceID)
	}

	// Print final metrics
	fmt.Println("\nMetrics:")
	fmt.Printf("Reservoir size: %.0f\n", metrics.GetReservoirSize())
	fmt.Printf("Sampled spans: %.0f\n", metrics.GetSampledSpans())
	fmt.Printf("Trace buffer size: %.0f\n", metrics.GetTraceBufferSize())
	fmt.Printf("Evictions: %.0f\n", metrics.GetEvictions())

	fmt.Println("\nVerification complete!")
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