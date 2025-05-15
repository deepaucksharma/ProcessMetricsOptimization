# Hello World OpenTelemetry Processor

A simple OpenTelemetry processor example that adds "Hello World" attributes to your metrics.

## 🔍 What It Does

This simple processor adds the following attributes to your metrics:

- At the resource level: `hello.world` = "Hello from OpenTelemetry!"
- At the datapoint level: `hello.processor` = "Hello from OpenTelemetry!"

These attributes can be used to identify metrics processed by this component and demonstrate how OpenTelemetry processors can modify telemetry data.

## 🚀 Quick Start

The easiest way to run this example is using Docker:

```bash
cd hello-processor
chmod +x run.sh
./run.sh
```

This will:
- Start a Docker container with the OpenTelemetry Collector
- Configure it to use the Hello World processor
- Collect basic host metrics and add the Hello World attributes
- Output the metrics to the logs for easy viewing

## 🔧 Configuration

The main configuration file is `config.yaml`. You can customize the greeting message:

```yaml
processors:
  hello:
    message: "Your custom message here"
```

## 💡 Learning OpenTelemetry Processors

This project demonstrates:

1. How to structure a basic OpenTelemetry processor
2. How to add attributes to metrics at both resource and datapoint levels
3. How to configure and run an OpenTelemetry collector with a custom processor

## 📂 Directory Structure

- `hello-processor/processor/` - The Go code for the processor implementation
  - `processor.go` - Core implementation
  - `config.go` - Configuration structure
  - `factory.go` - Factory for creating processor instances
- `hello-processor/config.yaml` - Example configuration
- `hello-processor/docker-compose.yml` - Docker setup
- `hello-processor/run.sh` - Convenience script to run everything