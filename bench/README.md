# Trace-Aware Reservoir Benchmark

This directory contains the benchmark harness for the Trace-Aware Reservoir Sampler. It provides tools for performance testing, comparison, and validation of different sampler configurations.

## Overview

The benchmark system uses a fan-out topology where identical traces are sent to multiple collector instances, each with a different configuration profile. This ensures fair comparison across profiles since they all receive exactly the same traffic.

## Key Components

- **profiles/**: YAML files with different performance-oriented configurations
- **kpis/**: Success criteria for each profile
- **runner/**: Go-based benchmark orchestrator
- **pte-kpi/**: KPI evaluation tool

## Quick Start

Use the Makefile targets from the repository root:

```bash
# Build the image
make image VERSION=bench

# Run benchmarks with all profiles
make bench DURATION=10m

# Run with specific profiles
make bench PROFILES=max-throughput-traces,tiny-footprint-edge DURATION=5m

# Run a quick benchmark test
make bench-quick
```

Alternatively, use the benchmark runner script:

```bash
./scripts/run-benchmark.sh --mode kind --profiles max-throughput-traces --duration 5m
```

## Available Profiles

- **max-throughput-traces**: Optimized for maximum trace throughput
- **tiny-footprint-edge**: Optimized for minimal resource usage in edge environments

## For More Information

See the comprehensive [Benchmarking Guide](../docs/benchmarking-guide.md) for detailed instructions on running, customizing, and analyzing benchmarks.