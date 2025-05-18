package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// Metric represents a simple process metric
type Metric struct {
	Name       string
	Value      float64
	ProcessID  string
	Executable string
	Timestamp  time.Time
	Attributes map[string]string
}

// HelloWorldProcessor adds a "hello.processor" attribute to all metrics
type HelloWorldProcessor struct {
	Message string
}

// Process adds the hello attribute to the metric
func (p *HelloWorldProcessor) Process(metrics []Metric) []Metric {
	processedMetrics := make([]Metric, len(metrics))
	
	for i, metric := range metrics {
		// Make a copy of the metric
		processedMetric := metric
		
		// Initialize attributes map if it doesn't exist
		if processedMetric.Attributes == nil {
			processedMetric.Attributes = make(map[string]string)
		}
		
		// Add hello attribute
		processedMetric.Attributes["hello.processor"] = p.Message
		
		processedMetrics[i] = processedMetric
	}
	
	return processedMetrics
}

// generateSampleMetrics creates sample process metrics
func generateSampleMetrics() []Metric {
	return []Metric{
		{
			Name:       "process.cpu.utilization",
			Value:      0.35,
			ProcessID:  "1234",
			Executable: "nginx",
			Timestamp:  time.Now(),
			Attributes: map[string]string{
				"host.name": "web-server-1",
			},
		},
		{
			Name:       "process.memory.usage",
			Value:      256.5,
			ProcessID:  "1234",
			Executable: "nginx",
			Timestamp:  time.Now(),
			Attributes: map[string]string{
				"host.name": "web-server-1",
			},
		},
		{
			Name:       "process.cpu.utilization",
			Value:      0.85,
			ProcessID:  "5678",
			Executable: "mysql",
			Timestamp:  time.Now(),
			Attributes: map[string]string{
				"host.name": "db-server-1",
			},
		},
	}
}

// Simple HTTP handler to demonstrate the processor
func demoHandler(w http.ResponseWriter, r *http.Request) {
	// Generate sample metrics
	metrics := generateSampleMetrics()
	
	// Create and use the Hello World processor
	processor := &HelloWorldProcessor{
		Message: "Hello from NRDOT Process-Metrics Optimization!",
	}
	
	// Process the metrics
	processedMetrics := processor.Process(metrics)
	
	// Display the processed metrics
	fmt.Fprintln(w, "<h1>NRDOT Hello World Processor Demo</h1>")
	fmt.Fprintln(w, "<p>This demonstrates how the Hello World processor adds attributes to metrics.</p>")
	fmt.Fprintln(w, "<h2>Processed Metrics:</h2>")
	fmt.Fprintln(w, "<pre>")
	
	for _, metric := range processedMetrics {
		fmt.Fprintf(w, "Name: %s\n", metric.Name)
		fmt.Fprintf(w, "Value: %.2f\n", metric.Value)
		fmt.Fprintf(w, "Process ID: %s\n", metric.ProcessID)
		fmt.Fprintf(w, "Executable: %s\n", metric.Executable)
		fmt.Fprintf(w, "Timestamp: %s\n", metric.Timestamp.Format(time.RFC3339))
		fmt.Fprintln(w, "Attributes:")
		
		for k, v := range metric.Attributes {
			fmt.Fprintf(w, "  %s: %s\n", k, v)
		}
		
		fmt.Fprintln(w, "-------------------")
	}
	
	fmt.Fprintln(w, "</pre>")
}

func main() {
	// Set up HTTP server
	http.HandleFunc("/", demoHandler)
	
	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	// Get HOST from environment or use all interfaces
	host := os.Getenv("HOST")
	if host == "" {
		host = "0.0.0.0"
	}
	
	// Start server
	fmt.Printf("Starting NRDOT Hello World Processor Demo on %s:%s...\n", host, port)
	fmt.Printf("Access this demo from your browser using one of these URLs:\n")
	fmt.Printf("  - http://localhost:%s (if on the same machine)\n", port)
	fmt.Printf("  - Try your machine's IP address: http://<your-ip-address>:%s\n", port)
	
	log.Fatal(http.ListenAndServe(host+":"+port, nil))
}
