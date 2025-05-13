package main

import (
	"fmt"
	"os"
	"time"

	"github.com/deepaucksharma/reservoir"
	"go.uber.org/zap"
)

func main() {
	// Create the logger
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()
	
	// Get configuration from environment variables or use defaults
	config := reservoir.DefaultConfig()
	
	// Log configuration
	logger.Info("Trace-Aware Reservoir Sampler for OpenTelemetry",
		zap.Int("reservoir_size", config.SizeK),
		zap.Duration("window_duration", config.WindowDuration),
		zap.Bool("trace_aware", config.TraceAware),
		zap.String("version", "0.1.0"),
	)
	
	// For benchmarking, we want to be able to run with any given configuration
	// This simulation mode just runs a sample reservoir for testing exports
	fmt.Println("Trace-Aware Reservoir Sampler is running in simulation mode")
	fmt.Println("Configuration:")
	fmt.Printf("  Reservoir Size: %d spans\n", config.SizeK)
	fmt.Printf("  Window Duration: %s\n", config.WindowDuration)
	fmt.Printf("  Trace-Aware Mode: %t\n", config.TraceAware)
	
	// Wait for a while to simulate running
	fmt.Println("\nSimulating running for 10 seconds...")
	time.Sleep(10 * time.Second)
	
	// Report metrics
	fmt.Println("\nSimulated reservoir metrics:")
	fmt.Printf("  Current Window: %d\n", time.Now().Unix() / 60)
	fmt.Printf("  Spans Sampled: %d\n", config.SizeK)
	fmt.Printf("  CPU Usage: %.2f%%\n", 15.5)
	fmt.Printf("  Memory Usage: %.2f MB\n", 256.3)
	
	fmt.Println("\nBenchmark is ready to be executed in a real environment")
	fmt.Println("Use the setup script to run a full benchmark with configured profiles")
}

