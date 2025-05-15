# NRDOT Process-Metrics Optimization

A production-grade, five-layer process-metrics pipeline for the New Relic Distribution of OpenTelemetry (NRDOT). The project is designed to significantly reduce the ingest cost of process metrics while maintaining visibility into the most important processes.

## Quick Start

To get started with the hello world processor:

### Simple Demo Version

```bash
# Clone the repository
git clone https://github.com/newrelic/nrdot-process-optimization.git
cd nrdot-process-optimization

# Option 1: Use the script (displays your machine's IP for remote access)
./run-demo.sh

# Option 2: Run directly
go run main.go

# Access in your browser:
# - If on the same machine: http://localhost:8080
# - From another device: http://<your-machine-ip>:8080
```

### Full OpenTelemetry Version

```bash
# Build the Docker image
make docker-build

# Start the local development stack
make compose-up

# View logs (in a new terminal)
make logs
```

## Hello World Processor

This starter kit includes a fully functional Hello World processor that demonstrates:

1. How to implement an OpenTelemetry processor
2. How to add custom attributes to metrics
3. How to instrument processors with self-observability metrics
4. How to integrate with the OpenTelemetry Collector

The Hello World processor adds a `hello.processor` attribute with the value "Hello from OpenTelemetry!" to all metrics that pass through it.

## Local Development Environment

The included Docker Compose setup provides:

- **OpenTelemetry Collector**: Configured with the Hello World processor
- **Mock New Relic Endpoint**: For local testing without sending data to New Relic
- **Prometheus**: For scraping and storing self-observability metrics
- **Grafana**: For visualization and dashboards

Access points:
- zPages: http://localhost:55679
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)
  - Pre-loaded dashboard: "NRDOT Custom Processor Starter KPIs"

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
│   ├── helloworld/       # Hello World example processor
│   ├── prioritytagger/   # L0: Critical process tagging (to be implemented)
│   ├── adaptivetopk/     # L1: Top-K process selection (to be implemented)
│   ├── othersrollup/     # L2: Non-priority/top process aggregation (to be implemented)
│   └── reservoirsampler/ # L3: Statistical sampling (to be implemented)
├── internal/             # Shared utilities
│   ├── attributes/       # Efficient pcommon.Map operations
│   ├── banding/          # Host load to K-value mapping for AdaptiveTopK
│   ├── obsreport/        # Standardized obsreport wrapper
│   ├── configvalidation/ # JSON schema validation
│   └── testhelpers/      # Utilities for testing
├── config/               # YAML configuration files
├── build/                # Docker, Compose, etc.
├── charts/               # Helm chart
└── test/                 # Test suites
```

## Configuration

The project uses a YAML configuration file (config/base.yaml) with the following key parameters, configurable via environment variables:

- `NEW_RELIC_LICENSE_KEY`: Required for sending data to New Relic
- `NEW_RELIC_OTLP_ENDPOINT`: Endpoint for sending data (default: https://otlp.nr-data.net/v1/metrics)
- `COLLECTION_INTERVAL`: Metric collection interval (default: 30s)
- `BATCH_SEND_SIZE`: Number of metrics to batch (default: 1000)
- `BATCH_TIMEOUT`: Batch sending timeout (default: 10s)
- `NEW_RELIC_MEMORY_LIMIT_MIB`: Memory limit (default: 250 MiB)
- `NEW_RELIC_MEMORY_SPIKE_LIMIT_MIB`: Memory spike limit (default: 150 MiB)

## Future Processors

The project will implement the following custom optimization processors:

1. **L0: PriorityTagger** - Tags critical processes to ensure they're always monitored
2. **L1: AdaptiveTopK** - Selects top K resource-intensive processes based on host load
3. **L2: OthersRollup** - Aggregates metrics from non-priority, non-top processes
4. **L3: ReservoirSampler** - Takes a statistically representative sample of remaining processes

This approach reduces metric series from potentially 250k+ per day to under 100, providing substantial cost savings while maintaining observability.

## Contributing

For detailed development guidance, see [CLAUDE.md](CLAUDE.md) and [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md).