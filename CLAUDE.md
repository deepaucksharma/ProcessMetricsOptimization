# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

NRDOT Process-Metrics Optimization is a production-grade, five-layer process-metrics pipeline for the New Relic Distribution of OpenTelemetry (NRDOT). The project is designed to significantly reduce the ingest cost of process metrics while maintaining visibility into the most important processes.

The pipeline consists of four key custom optimization processors (L0-L3) that progressively filter and optimize process metrics, leading to hot path export (L4) and optional cold path export (L5):
1. L0: PriorityTagger - Tags critical processes to ensure they're always monitored
2. L1: AdaptiveTopK - Selects top K resource-intensive processes based on host load
3. L2: OthersRollup - Aggregates metrics from non-priority, non-top processes
4. L3: ReservoirSampler - Takes a statistically representative sample of remaining processes

This approach reduces metric series from potentially 250k+ per day to under 100, providing substantial cost savings while maintaining observability.

## Development Commands

```bash
# Build the project
make build

# Run all tests
make test

# Run just unit tests
make test-unit

# Run integration tests
make test-integration

# Run end-to-end tests
make test-e2e

# Run benchmarks
make bench

# Run linting and static analysis
make lint

# Build Docker image
make docker-build

# Run local development stack with Docker Compose
make compose-up

# Generate SBOM
make sbom

# Vulnerability scanning
make vuln-scan

# Clean build artifacts
make clean
```

## Project Structure

```
📁 nrdot-process-optimization/
├── cmd/collector/        # Main application entrypoint
├── processors/           # Custom OTEL processors
│   ├── prioritytagger/   # L0: Critical process tagging
│   ├── adaptivetopk/     # L1: Top-K process selection
│   ├── othersrollup/     # L2: Non-priority/top process aggregation
│   ├── reservoirsampler/ # L3: Statistical sampling
│   └── (legacy_examples/helloworld/) # The initial starter kit
├── internal/             # Shared utilities
│   ├── attributes/       # Efficient pcommon.Map operations
│   ├── banding/          # Host load to K-value mapping for AdaptiveTopK
│   ├── obsreport/        # Standardized obsreport wrapper
│   ├── configvalidation/ # JSON schema validation
│   └── testhelpers/      # Utilities for testing (e.g., pmetric generation)
├── config/               # YAML configuration files
├── build/                # Docker, Compose, etc.
├── charts/               # Helm chart
└── test/                 # Test suites
```

## Configuration

The project uses a YAML configuration file (config/opt-plus.yaml) with the following key parameters, configurable via environment variables:

```yaml
processors:
  prioritytagger:
    critical_executables: [systemd, sshd, dockerd, kubelet, newrelic-infra]
    cpu_utilization_threshold: ${NR_PRIORITY_CPU:0.8}
  
  adaptivetopk:
    key_metric: process.cpu.utilization
    bands: { "0.2": 5, "0.5": 20, "0.8": 35, "1": 50 }
    hysteresis_duration: ${NR_TOPK_HYSTERESIS:15s}
  
  othersrollup:
    output_pid_attribute_value: "-1"
    output_executable_name_attribute_value: "_other_"
    aggregations:
      process.cpu.utilization: avg
      process.memory.usage: sum
      process.fd.open: sum
      process.threads: sum
  
  reservoirsampler:
    reservoir_size: ${NR_RESERVOIR:25}
    
  # (Standard processors like 'batch' and 'memory_limiter' are also present in the full config)
```

Key environment variables:
- `NEW_RELIC_LICENSE_KEY`: Required for sending data to New Relic
- `NEW_RELIC_OTLP_ENDPOINT`: Endpoint for sending data (default: https://otlp.nr-data.net/v1/metrics or region-specific variants like https://otlp.eu01.nr-data.net/v1/metrics)
- `NR_INTERVAL`: Metric collection interval (default: 15s)
- `NR_PRIORITY_CPU`: CPU threshold for priority tagging (default: 0.8)
- `NR_TOPK_HYSTERESIS`: How long to keep a process in Top-K after it drops below threshold (default: 15s)
- `NR_RESERVOIR`: Size of the reservoir sample (default: 25)
- `S3_BUCKET`: Bucket name for cold storage export (if enabled)
- `S3_PREFIX`: Object key prefix for cold storage export
- `AWS_REGION`: Region for S3 cold storage
- `AWS_ROLE_ARN`: IAM role ARN for S3 access (if using IAM role authentication)

## Coding Standards and Best Practices

1. Go 1.22 or higher is required
2. Code must pass staticcheck, govet, and govulncheck
3. Follow idiomatic Go practices
4. No global variables in processors; share functionality via internal packages
5. All processors must emit self-metrics for observability
6. Use allocation-efficient methods from internal/attributes for pdata operations
7. Include thorough unit tests with ≥80% coverage
8. Add benchmarks for performance-critical code paths
9. Processors should aim for statelessness regarding global variables; all state managed via configuration or instance members
10. Concurrency safety must be ensured if internal state is modified across multiple goroutines (though typically OTel pipelines serialize calls to a single processor instance for a given signal type)

## Testing Strategy

1. Unit tests for each processor and utility package
2. Integration tests for processor chains
3. End-to-end tests using Docker Compose
4. Benchmark tests for performance-critical code
5. Golden file fixtures for verifying complex processor behavior

## Metrics and Observability

Each processor emits custom self-metrics:

- `nrdot_prioritytagger_critical_processes_tagged_total`
- `otelcol_processor_adaptivetopk_dropped_metric_points_total`
- `nrdot_adaptivetopk_current_k`
- `nrdot_othersrollup_aggregated_series_count_total`
- `nrdot_reservoirsampler_fill_ratio`
- `nrdot_reservoirsampler_selected_identities_count`

Standard obsreport metrics (e.g., `otelcol_processor_[processor_name]_processed_metric_points`, `otelcol_processor_[processor_name]_dropped_metric_points`, latency histograms) are also available for each processor and should be monitored via Grafana dashboard 15983/18309 or zPages.

Key Grafana dashboards are provided for monitoring these metrics.

## Security Considerations

1. Run as non-root (UID 10001)
2. Use distroless container images
3. Generate and validate SBOM
4. Scan for vulnerabilities in dependencies
5. No hardcoded secrets

## Development Tips

1. Refer to the OpenTelemetry Collector docs when implementing processors
2. Test changes with the provided Docker Compose setup
3. Benchmark before/after for performance-critical changes
4. Use NRQL queries from the docs/NRQL_EXAMPLES.md file to validate data output
5. Follow the sequence diagram in README.md to understand the data flow
6. Check config/opt-plus.yaml against schemas/opt-plus.schema.json when making changes
7. Understand the role of internal/obsreport wrapper for consistent self-metric emission
8. For Top-K and Reservoir Sampler, pay close attention to how process.pid is handled (retained for Top-K, sampled for Reservoir, stripped for Others) to manage cardinality

## Integration Testing

When developing or testing export to New Relic:
1. Ensure the correct endpoint is set: `https://otlp.nr-data.net/v1/metrics`
2. Use a valid `NEW_RELIC_LICENSE_KEY` in your .env file
3. Validate with NRQL: `FROM Metric SELECT uniqueCount(process.pid) FACET nr.priority SINCE 30 MINUTES`
4. Check overall cardinality reduction: `FROM Metric SELECT uniqueCount(timeseries) WHERE metricName = 'process.cpu.utilization' SINCE 30 MINUTES COMPARE WITH 1 DAY AGO`
5. Ensure only the critical and Top-K processes have individual PIDs: `FROM Metric SELECT count(*) WHERE process.pid != '-1' AND process.pid IS NOT NULL FACET process.executable.name SINCE 30 MINUTES LIMIT 100`