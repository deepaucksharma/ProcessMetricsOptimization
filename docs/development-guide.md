# Trace-Aware Reservoir Development Guide

This comprehensive guide provides everything you need to know to develop, build, deploy, and maintain the trace-aware-reservoir-otel project.

## Table of Contents
- [Prerequisites](#prerequisites)
- [Project Structure](#project-structure)
- [Building and Testing](#building-and-testing)
- [Deployment](#deployment)
- [Implementation Status](#implementation-status)
- [Troubleshooting](#troubleshooting)
- [Maintenance](#maintenance)

## Prerequisites

- **Go 1.21+** - For development and building
- **Docker** - For containerization and local Kubernetes
- **Kubernetes** - For deployment (e.g., Docker Desktop with Kubernetes enabled or KinD)
- **Helm** - For Kubernetes deployment
- **New Relic license key** (optional) - For sending data to New Relic

> **Note for Windows Users**: If you're on Windows 10/11, please refer to our [Windows Development Guide](windows-guide.md) for detailed setup instructions using WSL 2.

## Project Structure

```
trace-aware-reservoir-otel/
│
├── core/                      # Core library code
│   └── reservoir/             # Reservoir sampling implementation
│
├── apps/                      # Applications
│   ├── collector/             # OpenTelemetry collector integration
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

## Building and Testing

Our streamlined Makefile provides a comprehensive set of commands for all development tasks:

```bash
# Setup your development environment
make setup

# Run all tests
make test

# Run specific test suites
make test-core     # Core library tests
make test-apps     # Application tests
make test-bench    # Benchmark framework tests

# Run linters
make lint

# Build the collector application
make build

# Build all applications
make build-all
```

### Docker Image Building

```bash
# Build main Docker image
make image VERSION=v0.1.0

# Build benchmark Docker image
make image-bench VERSION=v0.1.0

# Build all Docker images
make images VERSION=v0.1.0
```

The multi-stage Dockerfile in `build/docker/Dockerfile.streamlined` handles all the build steps efficiently, including:
1. Dependency caching for faster builds
2. Multi-target builds for different use cases
3. Optimized runtime images

## Deployment

### Local Testing with KinD

```bash
# Create a KinD cluster
make kind

# Load Docker images into KinD
make kind-load

# Deploy to the KinD cluster
export NEW_RELIC_KEY="your_license_key_here"  # Optional
make deploy VERSION=v0.1.0
```

### Deployment to Kubernetes

```bash
# Deploy to Kubernetes
make deploy VERSION=v0.1.0 NAMESPACE=otel LICENSE_KEY=$NEW_RELIC_KEY
```

### Quick Development Cycle

For rapid development iterations:

```bash
# Complete development cycle: test, build, deploy
make dev VERSION=v0.1.0
```

### Operations

```bash
# Check deployment status
make status

# View collector logs
make logs

# Check metrics
make metrics

# Clean up resources
make clean
```

## Implementation Status

The project has undergone a major refactoring to improve modularity and maintainability:

### Completed Features

1. ✅ **Modular Architecture**: 
   - Core sampling logic is now separated from OpenTelemetry integration
   - Clean interfaces enable easier testing and extension

2. ✅ **Advanced Sampling Features**:
   - Statistically-sound reservoir sampling using Algorithm R
   - Time-based windowing for regular export cycles
   - Trace-aware buffering to keep spans from the same trace together
   - Persistent state using BadgerDB

3. ✅ **Benchmarking Framework**: 
   - Go-based benchmark runner provides robust orchestration
   - Profile-based configuration for easy comparison of different settings
   - KPI evaluation to measure performance objectively

4. ✅ **Deployment Automation**:
   - Streamlined Docker and Kubernetes deployment
   - Consolidated Helm chart for all deployment scenarios
   - Integration with New Relic OpenTelemetry Distribution

### Next Steps

1. **Comprehensive Unit Testing**:
   - Improve test coverage for core components
   - Add performance benchmarks for key algorithms

2. **CI/CD Enhancements**:
   - Update GitHub Actions workflows
   - Add automated release process

3. **Performance Tuning**:
   - Optimize memory usage in core algorithm
   - Improve trace buffer efficiency
   - Enhance checkpointing mechanism

## Troubleshooting

### Common Issues and Solutions

1. **Pod Startup Issues**
   - **Problem**: CrashLoopBackOff with Badger "permission denied"
   - **Solution**: Check the persistent volume permissions. The container runs as uid 10001 (otel user).

2. **Image Pull Errors**
   - **Problem**: imagePullBackOff
   - **Solution**: Verify the image exists and is publicly accessible or properly authenticated.

3. **Processor Not Found**
   - **Problem**: Processor not showing in metrics or components list
   - **Solution**: Check the processor registration in factory.go and ensure imports are correct.

4. **Benchmark Failures**
   - **Problem**: KPI evaluations failing
   - **Solution**: Check the KPI definitions in `bench/kpis/` and adjust thresholds if needed.

### Viewing Logs

```bash
# View collector logs
make logs

# Check specific metrics
make metrics | grep reservoir_sampler
```

## Maintenance

### Upgrading

To upgrade to a new version:

1. Update your code and tests
2. Create a new tag (e.g., v0.1.1)
3. Push the tag to GitHub
4. Build and push the new image or let GitHub Actions do it
5. Deploy the new version with `make deploy VERSION=v0.1.1`

### Reference Configuration

Example configuration from a benchmark profile:

```yaml
processors:
  memory_limiter:
    check_interval: 100ms
    limit_percentage: 90
    spike_limit_percentage: 95
  batch:
    timeout: 100ms
    send_batch_size: 10000
  reservoir_sampler:
    size_k: 15000                       # Reservoir size (in thousands)
    window_duration: 30s                # Time window for each reservoir
    checkpoint_path: /var/otelpersist/badger  # Persistence location
    checkpoint_interval: 5s             # How often to save state
    trace_aware: true                   # Buffer spans from the same trace
    trace_buffer_timeout: 5s            # How long to wait for spans from same trace
    trace_buffer_max_size: 500000       # Maximum buffer size
    db_compaction_schedule_cron: "0 */6 * * *"  # Compaction schedule
    db_compaction_target_size: 268435456  # Target size (256 MiB)
```

---

For benchmarking details, see the [Benchmarking Guide](benchmarking-guide.md).