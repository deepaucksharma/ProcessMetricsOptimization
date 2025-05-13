package e2e

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/deepaucksharma/reservoir"
	"github.com/deepaucksharma/trace-aware-reservoir-otel/apps/collector/persistence"
	"go.uber.org/zap"
)

func TestReservoirWithCheckpointing(t *testing.T) {
	// Create a logger
	logger, _ := zap.NewDevelopment()
	
	// Create a temporary directory for badger
	tmpDir, err := os.MkdirTemp("", "badger-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)
	
	// Create a metrics reporter
	metrics := NewSimpleMetricsReporter()
	
	// Create checkpoint manager
	checkpointMgr, err := persistence.NewBadgerCheckpointManager(
		tmpDir,
		1024 * 1024, // 1MB target size
		metrics.GetCheckpointAgeGauge(),
		metrics.GetReservoirDbSizeGauge(),
		metrics.GetCompactionCountCounter(),
		logger,
	)
	if err != nil {
		t.Fatalf("Failed to create checkpoint manager: %v", err)
	}
	defer checkpointMgr.Close()
	
	// Create reservoir and window
	r := reservoir.NewAlgorithmR(10, metrics)
	window := reservoir.NewTimeWindow(2 * time.Second)
	window.SetRolloverCallback(func() {
		fmt.Println("Window rolled over")
	})
	
	// Create trace buffer with short timeout for testing
	traceBuffer := reservoir.NewTraceBuffer(100, 500*time.Millisecond, metrics)
	
	// Add spans to the reservoir
	fmt.Println("Adding spans to reservoir...")
	for i := 0; i < 20; i++ {
		span := createTestSpan(fmt.Sprintf("trace-%d", i/2), fmt.Sprintf("span-%d", i))
		added := r.AddSpan(span)
		fmt.Printf("Added span %s (trace: %s): %v\n", span.ID, span.TraceID, added)
	}
	
	// Add spans to trace buffer
	fmt.Println("\nAdding spans to trace buffer...")
	for i := 0; i < 10; i++ {
		traceID := fmt.Sprintf("trace-buffer-%d", i/3)
		span := createTestSpan(traceID, fmt.Sprintf("buffer-span-%d", i))
		traceBuffer.AddSpan(span)
		fmt.Printf("Added span %s to trace buffer (trace: %s)\n", span.ID, span.TraceID)
	}
	
	// Wait for traces to complete
	fmt.Println("\nWaiting for trace timeout...")
	time.Sleep(600 * time.Millisecond)
	completedTraces := traceBuffer.GetCompletedTraces()
	
	fmt.Printf("Got %d completed traces\n", len(completedTraces))
	for i, trace := range completedTraces {
		fmt.Printf("Trace %d: %d spans\n", i+1, len(trace))
		
		// Add these spans to the reservoir
		for _, span := range trace {
			r.AddSpan(span)
		}
	}
	
	// Get the state of the reservoir before checkpointing
	beforeSample := r.GetSample()
	fmt.Printf("\nReservoir before checkpoint contains %d spans\n", len(beforeSample))
	
	// Create a checkpoint
	windowID, startTime, endTime := window.Current()
	fmt.Printf("Creating checkpoint for window %d (%s - %s)\n", 
		windowID, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))
	
	// Convert spans for checkpointing
	spans := r.GetAllSpansWithKeys()
	spanMap := make(map[string]reservoir.SpanWithResource, len(spans))
	for k, v := range spans {
		spanMap[k] = reservoir.SpanWithResource{
			Span:     v,
			Resource: v.ResourceInfo,
			Scope:    v.ScopeInfo,
		}
	}
	
	// Checkpoint
	err = checkpointMgr.Checkpoint(windowID, startTime, endTime, 1, spanMap)
	if err != nil {
		t.Fatalf("Failed to create checkpoint: %v", err)
	}
	
	// Get size of checkpoint files
	var totalSize int64
	err = filepath.Walk(tmpDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			fmt.Printf("Checkpoint file: %s, size: %d bytes\n", path, info.Size())
			totalSize += info.Size()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk checkpoint directory: %v", err)
	}
	fmt.Printf("Total checkpoint size: %d bytes\n", totalSize)
	
	// Reset reservoir
	fmt.Println("\nResetting reservoir...")
	r.Reset()
	if r.Size() != 0 {
		t.Fatalf("Expected empty reservoir after reset, got size %d", r.Size())
	}
	
	// Load from checkpoint
	fmt.Println("Loading from checkpoint...")
	restoredWindowID, restoredStartTime, restoredEndTime, restoredCount, restoredSpans, err := 
		checkpointMgr.LoadCheckpoint()
	if err != nil {
		t.Fatalf("Failed to load checkpoint: %v", err)
	}
	
	fmt.Printf("Restored window: ID=%d, start=%s, end=%s, count=%d, spans=%d\n",
		restoredWindowID, restoredStartTime.Format(time.RFC3339), 
		restoredEndTime.Format(time.RFC3339), restoredCount, len(restoredSpans))
	
	// Restore the reservoir state
	window.SetState(restoredWindowID, restoredStartTime, restoredEndTime, restoredCount)
	for _, span := range restoredSpans {
		r.AddSpan(span.Span)
	}
	
	// Get sample after restoration
	afterSample := r.GetSample()
	fmt.Printf("Reservoir after restore contains %d spans\n", len(afterSample))
	
	// Verify sizes match
	if len(beforeSample) != len(afterSample) {
		t.Fatalf("Sample size mismatch: before=%d, after=%d", len(beforeSample), len(afterSample))
	}
	
	// Test DB compaction
	fmt.Println("\nTesting DB compaction...")
	err = checkpointMgr.Compact()
	if err != nil {
		t.Fatalf("Failed to compact DB: %v", err)
	}
	
	// Get final metrics
	fmt.Println("\nEnd-to-end test complete!")
	fmt.Println("Final metrics:")
	fmt.Printf("Reservoir size: %.0f\n", metrics.GetReservoirSize())
	fmt.Printf("Sampled spans: %.0f\n", metrics.GetSampledSpans())
	fmt.Printf("Trace buffer size: %.0f\n", metrics.GetTraceBufferSize())
	fmt.Printf("Evictions: %.0f\n", metrics.GetEvictions())
	fmt.Printf("Checkpoint age: %.1f seconds\n", metrics.GetCheckpointAge())
	fmt.Printf("DB size: %.0f bytes\n", metrics.GetDBSize())
	fmt.Printf("Compactions: %.0f\n", metrics.GetCompactions())
}