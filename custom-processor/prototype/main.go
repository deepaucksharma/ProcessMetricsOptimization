package main

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

// HelloWorldProcessor is a simplified processor that just logs messages
type HelloWorldProcessor struct {
	logger    *zap.Logger
	startTime time.Time
	message   string
}

// NewHelloWorldProcessor creates a new hello world processor
func NewHelloWorldProcessor(message string, logger *zap.Logger) *HelloWorldProcessor {
	if logger == nil {
		var err error
		logger, err = zap.NewDevelopment()
		if err != nil {
			panic(fmt.Sprintf("Failed to create logger: %v", err))
		}
	}

	return &HelloWorldProcessor{
		logger:    logger,
		startTime: time.Now(),
		message:   message,
	}
}

// Start is called when the processor is starting
func (p *HelloWorldProcessor) Start() error {
	p.logger.Info("Hello World Processor started!",
		zap.String("message", p.message),
		zap.Time("start_time", p.startTime))
	return nil
}

// Process would process telemetry data in a real implementation
func (p *HelloWorldProcessor) Process(itemID string) {
	p.logger.Info("Processing telemetry item",
		zap.String("item_id", itemID),
		zap.String("processor", "hello_world"),
		zap.String("custom_message", p.message))
}

// Shutdown is called when the processor is shutting down
func (p *HelloWorldProcessor) Shutdown() error {
	duration := time.Since(p.startTime)
	p.logger.Info("Hello World Processor is shutting down",
		zap.Duration("uptime", duration))
	return nil
}

func main() {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create a simple processor
	processor := NewHelloWorldProcessor("This is a custom processor for OpenTelemetry!", logger)
	
	// Start the processor
	processor.Start()
	
	// Simulate processing a few telemetry items
	processor.Process("span-1")
	processor.Process("metric-1")
	processor.Process("trace-1")
	
	// Shutdown the processor
	processor.Shutdown()
	
	fmt.Println("\nThis prototype demonstrates a simple OpenTelemetry processor.")
	fmt.Println("In a real implementation, this would be integrated into the OpenTelemetry Collector.")
}