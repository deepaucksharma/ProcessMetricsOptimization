# Trace-Aware Reservoir Sampling for OpenTelemetry

A trace-aware reservoir sampling implementation for OpenTelemetry Collector that intelligently samples traces based on their importance, maintaining a representative sample even under high load.

## Project Structure

The project is organized into these main components:

```
trace-aware-reservoir-otel/
│
├── core/                      # Core library code
│   └── reservoir/             # Reservoir sampling implementation
│
├── apps/                      # Applications
│   ├── collector/             # OpenTelemetry collector with reservoir sampling
│   └── tools/                 # Supporting tools
│
├── bench/                     # Benchmarking framework
│   ├── profiles/              # Benchmark configuration profiles
│   ├── kpis/                  # Key Performance Indicators
│   └── runner/                # Benchmark orchestration
│
├── infra/                     # Infrastructure code
│   ├── helm/                  # Helm charts
│   └── kind/                  # Kind cluster configurations
│
├── build/                     # Build configurations
│   ├── docker/                # Dockerfiles
│   └── scripts/               # Build scripts
│
├── scripts/                   # Operational scripts
│
└── docs/                      # Documentation
```

## Overview

This repository implements a statistically-sound reservoir sampling processor for the OpenTelemetry Collector that:

- Maintains an unbiased, representative sample even with unbounded data streams
- Preserves complete traces when operating in trace-aware mode
- Persists the reservoir state using Badger database for durability across restarts
- Integrates seamlessly with the New Relic OpenTelemetry Distribution (NR-DOT)

## Quick Start

### 1. Setup Development Environment

```bash
# Install dependencies and setup development environment
make setup
```

### 2. Build and Test

```bash
# Run all tests
make test

# Build the application
make build
```

### 3. Build Docker Images

```bash
# Build main Docker image
make image

# Build all images (main, benchmark)
make images
```

### 4. Deploy to Kubernetes

```bash
# Create Kind cluster and load images
make kind
make kind-load

# Deploy to Kubernetes
export NEW_RELIC_KEY="your_license_key_here"
make deploy
```

### 5. Run Benchmarks

```bash
# Run full benchmarks
make bench PROFILES=max-throughput-traces,tiny-footprint-edge DURATION=10m

# Run quick benchmark test
make bench-quick

# Alternatively, use the flexible benchmark runner script
./scripts/run-benchmark.sh --mode kind --profiles max-throughput-traces --duration 5m
```

## Implementation Details

The processor uses Algorithm R for reservoir sampling with these key characteristics:

- **Windowed Sampling**: Maintain separate reservoirs for configurable time windows
- **Trace Awareness**: Buffer and handle spans with the same trace ID together
- **Persistence**: Store reservoir state in Badger DB with configurable checkpointing
- **Metrics**: Expose performance and behavior metrics via Prometheus

### Architecture

```
┌─────────────┐     ┌───────────────────┐     ┌─────────────┐
│ OTLP Input  │────▶│ Reservoir Sampler │────▶│ OTLP Output │
└─────────────┘     └───────────────────┘     └─────────────┘
                            │
                            ▼
                    ┌───────────────┐
                    │ Badger DB     │
                    │ Persistence   │
                    └───────────────┘
```

## Configuration

Sample configuration in your collector config.yaml:

```yaml
processors:
  reservoir_sampler:
    size_k: 5000                         # Reservoir size (in thousands of traces)
    window_duration: 60s                 # Time window for each reservoir
    checkpoint_path: /var/otelpersist/badger  # Persistence location
    checkpoint_interval: 10s             # How often to save state
    trace_aware: true                    # Buffer spans from the same trace
    trace_buffer_timeout: 30s            # How long to wait for spans from same trace
    trace_buffer_max_size: 100000        # Maximum buffer size
    db_compaction_schedule_cron: "0 2 * * *"  # When to compact the database
    db_compaction_target_size: 134217728 # Target size for compaction (128 MiB)
```

## Benchmark Profiles

The project includes a benchmarking system with different performance profiles:

- **max-throughput-traces**: Optimized for maximum trace throughput
- **tiny-footprint-edge**: Optimized for minimal resource usage in edge environments

These profiles can be benchmarked simultaneously against identical traffic using our fan-out topology, allowing for direct comparison of different configurations.

## Streamlined Development Workflow

Our enhanced workflow simplifies the development experience:

### Unified Makefile Commands

```bash
# View all available commands with descriptions
make help

# Development tasks
make deps            # Install dependencies
make lint            # Run linters
make test            # Run all tests
make build           # Build the application

# Docker image building
make image           # Build main image
make image-bench     # Build benchmark image
make images          # Build all images

# Kubernetes deployment
make kind            # Create Kind cluster
make kind-load       # Load images into cluster
make deploy          # Deploy to Kubernetes
make quickrun        # Quick build and load

# Operations
make status          # Check deployment status
make logs            # Stream collector logs
make metrics         # Check metrics
make clean           # Clean up resources

# Benchmarking
make bench           # Run benchmarks
make bench-quick     # Run a quick benchmark
make bench-clean     # Clean up benchmark resources
```

### Flexible Benchmark Runner

For more control over benchmarks, use the benchmark runner script:

```bash
# Run in Kind cluster with multiple profiles
./scripts/run-benchmark.sh --mode kind --profiles max-throughput-traces,tiny-footprint-edge --duration 5m

# Run in local simulator mode
./scripts/run-benchmark.sh --mode local --port 9090

# Run with existing Kubernetes cluster
./scripts/run-benchmark.sh --mode existing --kubeconfig ~/.kube/my-cluster.yaml
```

## Prerequisites

- Docker
- Kubernetes cluster (e.g., Docker Desktop with Kubernetes enabled or KinD)
- Helm (for Kubernetes deployment)
- New Relic license key (optional, for sending data to New Relic)
- Go 1.21+

## Documentation

- [Streamlined Workflow](docs/streamlined-workflow.md) - Comprehensive guide to our improved development workflow
- [Implementation Guide](docs/implementation-guide.md) - Step-by-step guide for building and deploying
- [Core Library](core/reservoir/README.md) - Documentation for the core reservoir sampling library
- [Benchmark Implementation](docs/benchmark-implementation.md) - End-to-end benchmark guide with fan-out topology
- [Contributing Guide](CONTRIBUTING.md) - Guidelines for contributing to the project

## Windows Development 

For Windows 10/11 users, we recommend using WSL 2 (Windows Subsystem for Linux) with Ubuntu 22.04 to maintain full compatibility with the Linux-based tooling in this project.

See our [Windows Development Guide](docs/windows-guide.md) for detailed setup instructions.

## License

[Insert License Information]