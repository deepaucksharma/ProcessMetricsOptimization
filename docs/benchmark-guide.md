# Trace-Aware Reservoir Sampling Benchmark Guide

This guide provides step-by-step instructions for running benchmarks on the trace-aware reservoir sampling implementation for OpenTelemetry.

## Prerequisites

Make sure you have the following tools installed:

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

You can use the provided script to quickly set up and run a simulated benchmark:

```bash
# Make the script executable if needed
chmod +x scripts/setup-benchmark.sh

# Run with default parameters (bench tag, 5m duration, max-throughput-traces profile)
./scripts/setup-benchmark.sh

# Or customize parameters:
./scripts/setup-benchmark.sh custom-tag 10m tiny-footprint-edge
```

Note: The current implementation runs a simplified benchmark simulation. A full OpenTelemetry Collector integration will be implemented in future versions.

## Manual Benchmark Process

If you prefer to run each step manually, follow these instructions:

### 1. Build the Docker image

```bash
# From repo root
export IMAGE_TAG=bench  # or any tag you prefer
docker build -t ghcr.io/deepaucksharma/nrdot-reservoir:$IMAGE_TAG -f build/docker/Dockerfile.bench .
```

### 2. Create a KinD cluster

```bash
# Create a KinD cluster with our custom configuration
kind create cluster --config infra/kind/bench-config.yaml
```

### 3. Load the image into KinD

```bash
# Load the built image into KinD
kind load docker-image ghcr.io/deepaucksharma/nrdot-reservoir:$IMAGE_TAG
```

### 4. Run the benchmark

```bash
# From repo root
cd bench/runner
go run main.go \
  --image=ghcr.io/deepaucksharma/nrdot-reservoir:$IMAGE_TAG \
  --duration=5m \
  --profiles=max-throughput-traces
```

You can select specific profiles to run by separating them with commas:

```bash
go run main.go \
  --image=ghcr.io/deepaucksharma/nrdot-reservoir:$IMAGE_TAG \
  --duration=5m \
  --profiles=max-throughput-traces,tiny-footprint-edge
```

### 5. Inspecting benchmark results

Results will be output to the console, but you can also check the metrics directly:

```bash
kubectl -n bench-<profile> logs deploy/collector-<profile>
```

To see a specific metric:

```bash
kubectl exec -n bench-<profile> deploy/collector-<profile> -- curl localhost:8888/metrics | grep reservoir_sampler
```

## Available Benchmark Profiles

| Profile                | Description                                        | Primary Focus                   |
| ---------------------- | -------------------------------------------------- | ------------------------------- |
| max-throughput-traces  | Optimized for maximum trace throughput             | Performance under high load     |
| tiny-footprint-edge    | Optimized for minimal resource usage               | Resource efficiency             |

## Creating Custom Profiles

To create a custom benchmark profile:

1. Add a new YAML file in `bench/profiles/` (e.g., `custom-profile.yaml`)
2. Define your processor configuration and resource requirements
3. Add corresponding KPI rules in `bench/kpis/` (e.g., `custom-profile.yaml`)
4. Run the benchmark with your new profile

Example minimal profile:

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

Example KPI file:

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

## Cleaning Up

To clean up all benchmark resources:

```bash
# Delete the KinD cluster
kind delete cluster
```

## Troubleshooting

| Issue                             | Solution                                                            |
| --------------------------------- | ------------------------------------------------------------------- |
| Can't load Docker image to KinD   | Verify image exists with `docker images`                            |
| Pod fails to start                | Check logs with `kubectl logs -n bench-<profile> pod/<pod-name>`     |
| KPI evaluation fails              | Verify metrics exist and check spelling in kpi YAML files           |
| No traffic sent to collectors     | Check fanout service with `kubectl -n fanout get pods,services`     |

## Additional Resources

- [Core Benchmark Implementation Guide](benchmark-implementation.md)
- [Trace-Aware Reservoir Implementation](../core/reservoir/README.md)