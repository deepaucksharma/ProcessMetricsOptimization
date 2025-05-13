# Trace-Aware Reservoir Benchmarking Guide

This comprehensive guide explains how to benchmark and evaluate the performance of the trace-aware reservoir sampling implementation for OpenTelemetry.

## Table of Contents
- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Benchmark Architecture](#benchmark-architecture)
- [Benchmark Profiles](#benchmark-profiles)
- [Advanced Usage](#advanced-usage)
- [Creating Custom Profiles](#creating-custom-profiles)
- [Result Analysis](#result-analysis)
- [Troubleshooting](#troubleshooting)

## Overview

The benchmark system uses a fan-out architecture to ensure identical traffic is sent to each profile being tested. This allows for fair comparison of different configurations since they all receive exactly the same trace data during evaluation.

Key features:
- **Fan-out topology** for parallel testing of multiple profiles
- **Profile-based configuration** for different use cases
- **Objective KPI evaluation** against defined thresholds
- **Flexible runtime options** for different test scenarios
- **Results collection** for performance analysis

## Prerequisites

| Tool         | Version | Purpose                          |
| ------------ | ------- | -------------------------------- |
| **Docker**   | ≥ 24    | Building images & running KinD   |
| **KinD**     | ≥ 0.20  | Kubernetes in Docker for testing |
| **kubectl**  | ≥ 1.28  | Interacting with K8s cluster     |
| **Helm**     | ≥ 3.12  | Deploying collector & load-gen   |
| **Go**       | ≥ 1.21  | Building the benchmark runner    |
| **GNU Make** | any     | Running Makefile targets         |

Optional:
* `NEW_RELIC_KEY` environment variable if you want to push spans to New Relic.

## Quick Start

We've streamlined the benchmarking process with two easy approaches:

### Option 1: Using the Makefile

```bash
# Build the Docker image
make image VERSION=bench

# Run full benchmarks
make bench PROFILES=max-throughput-traces,tiny-footprint-edge DURATION=10m

# Run a quick benchmark test with one profile
make bench-quick
```

### Option 2: Using the Benchmark Runner Script

For more flexibility, use our dedicated benchmark runner script:

```bash
# Run in KinD with multiple profiles
./scripts/run-benchmark.sh --mode kind --profiles max-throughput-traces,tiny-footprint-edge --duration 5m

# Run locally in simulator mode
./scripts/run-benchmark.sh --mode local --port 9090

# Run with an existing Kubernetes cluster
./scripts/run-benchmark.sh --mode existing --kubeconfig ~/.kube/my-cluster.yaml
```

## Benchmark Architecture

The benchmark system utilizes a sophisticated fan-out architecture with several key components:

```
┌────────────────┐     ┌────────────────┐     ┌────────────────┐
│  Load Generator│────▶│  Fan-out       │────▶│  Profile A     │
└────────────────┘     │  Collector     │     └────────────────┘
                       │                │     
                       │                │     ┌────────────────┐
                       │                │────▶│  Profile B     │
                       └────────────────┘     └────────────────┘
                                 │            
                                 │            ┌────────────────┐
                                 └───────────▶│  Profile C     │
                                              └────────────────┘
```

1. **Load Generator**: Creates synthetic trace traffic
2. **Fan-out Collector**: Receives traces and duplicates them to all profile collectors
3. **Profile Collectors**: Each profile gets its own collector with specific configuration
4. **KPI Evaluation**: Metrics are scraped and evaluated against profile-specific thresholds

The otel-bundle Helm chart supports these deployment modes:
- **collector**: Runs a collector with reservoir sampler
- **fanout**: Runs a fan-out collector that duplicates traffic
- **loadgen**: Runs a load generator that sends synthetic traffic

## Benchmark Profiles

The project includes pre-defined benchmark profiles for different use cases:

| Profile                | Description                                | Primary Focus                 |
| ---------------------- | ------------------------------------------ | ---------------------------- |
| max-throughput-traces  | Optimized for maximum trace throughput     | Performance under high load   |
| tiny-footprint-edge    | Optimized for minimal resource usage       | Resource efficiency          |

Each profile is defined in `bench/profiles/` as a YAML file:

```yaml
# Example: bench/profiles/max-throughput-traces.yaml
collector:
  replicaCount: 1
  configOverride:
    processors:
      reservoir_sampler:
        size_k: 15000
        window_duration: 30s
        # ...other settings
  resources:
    limits:
      cpu: 2000m
      memory: 4Gi
```

Each profile has corresponding KPI definitions in `bench/kpis/`:

```yaml
# Example: bench/kpis/max-throughput-traces.yaml
rules:
  - name: "Memory Usage"
    metric: "process_runtime_heap_alloc_bytes"
    min_value: 0
    max_value: 500000000 # 500MB
    critical: true
  # ...other KPI rules
```

## Advanced Usage

### Detailed Benchmark Process

The full benchmark process includes these steps:

1. **Setup**: Create a KinD cluster with our benchmark configuration
   ```bash
   kind create cluster --config infra/kind/bench-config.yaml
   ```

2. **Build & Load**: Build and load the image into KinD
   ```bash
   make image VERSION=bench
   kind load docker-image ghcr.io/deepaucksharma/trace-reservoir:bench
   ```

3. **Run**: Execute the benchmark with the Go-based runner
   ```bash
   cd bench && go run ./runner/main.go \
     -image ghcr.io/deepaucksharma/trace-reservoir:bench \
     -duration 10m \
     -profiles max-throughput-traces,tiny-footprint-edge
   ```

4. **Analyze**: Review metrics and KPI evaluation
   ```bash
   kubectl -n bench-<profile> logs deploy/collector-<profile>
   ```

### New Relic Integration Options

| Mode                         | How to enable                                                      | Results in New Relic                                           |
| ---------------------------- | ------------------------------------------------------------------ | -------------------------------------------------------------- |
| **Silent Bench** (no export) | Comment out the `otlphttp` exporter, use the `logging` exporter    | Nothing is sent; only local KPI CSVs                           |
| **Side-by-side in NR**       | Use `otlphttp` exporter with the `resource/add-profile` processor  | Filter by attribute `benchmark.profile` to see separate traces  |

## Creating Custom Profiles

To create a custom benchmark profile:

1. **Add Profile Configuration**
   - Create a new YAML file in `bench/profiles/` (e.g., `custom-profile.yaml`)
   - Define your processor configuration and resource requirements

   ```yaml
   # bench/profiles/custom-profile.yaml
   collector:
     replicaCount: 1
     configOverride:
       processors:
         reservoir_sampler:
           size_k: 10000
           window_duration: 45s
           checkpoint_path: /var/otelpersist/badger
           checkpoint_interval: 10s
           trace_aware: true
     resources:
       limits:
         cpu: 1000m
         memory: 2Gi
   ```

2. **Add KPI Definitions**
   - Create a matching file in `bench/kpis/` (e.g., `custom-profile.yaml`)
   - Define metrics and thresholds for evaluation

   ```yaml
   # bench/kpis/custom-profile.yaml
   rules:
     - name: "Memory Usage"
       description: "Memory usage should be below threshold"
       metric: "process_runtime_heap_alloc_bytes" 
       min_value: 0
       max_value: 750000000 # 750MB
       critical: true
       
     - name: "Reservoir Size"
       description: "Reservoir should maintain expected size"
       metric: "reservoir_size"
       min_value: 9000
       max_value: 11000
       critical: true
   ```

3. **Run the Benchmark**
   ```bash
   make bench PROFILES=custom-profile DURATION=10m
   ```

## Result Analysis

Benchmark results are stored in CSV files and summarized in Markdown format in the `benchmark-results-<timestamp>` directory.

The result summary includes:
- Configuration details for each profile
- KPI metrics and their values
- Pass/fail status for each metric
- Resource usage statistics

Example metrics to analyze:
- `reservoir_size`: Current reservoir size
- `sampled_spans_total`: Total number of spans sampled
- `process_runtime_heap_alloc_bytes`: Memory usage
- `reservoir_db_size_bytes`: Database size on disk

## Troubleshooting

| Issue                             | Solution                                                            |
| --------------------------------- | ------------------------------------------------------------------- |
| Can't load Docker image to KinD   | Verify image exists with `docker images`                            |
| Pod fails to start                | Check logs with `kubectl logs -n bench-<profile> pod/<pod-name>`     |
| KPI evaluation fails              | Verify metrics exist and check spelling in KPI YAML files           |
| No traffic sent to collectors     | Check fanout service with `kubectl -n fanout get pods,services`     |
| Fan-out exporter times out        | Ensure the service name in exporter configuration matches the deployed services |

To see specific metrics for a profile:

```bash
kubectl exec -n bench-<profile> -- curl localhost:8888/metrics | grep reservoir_sampler
```

## Cleaning Up

To clean up all benchmark resources:

```bash
# Using the Makefile
make bench-clean

# Or directly with KinD
kind delete cluster --name benchmark-kind
```

---

For more information about the development process, see the [Development Guide](development-guide.md).