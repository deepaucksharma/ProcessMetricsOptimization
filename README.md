# NRDOT Process-Metrics Optimization

A production-grade, **multi-layer** process-metrics pipeline built as a custom **New Relic Distribution of OpenTelemetry (NRDOT)** collector.
Its core mission is to **significantly reduce metric ingest costs (aiming for ≥ 90% reduction)** by intelligently refining process metric data, while preserving deep visibility into the processes that are critical for operational insight.

The pipeline will achieve this by progressively **Tagging → Filtering → Aggregating → Sampling** process metrics through a series of custom OpenTelemetry processors (L0 – L3). This project begins with a foundational "Hello World" processor (Phase 0) and will incrementally build out the full "Opt-Plus" optimization pipeline in subsequent phases.

---

## Project Status & Roadmap Snapshot

The project is developed in phases. **Currently, Phase 0 is complete.**

| Phase | Focus                                          | Status        |
|-------|------------------------------------------------|---------------|
| 0     | Foundation: Hello World processor & Local Dev Stack | ✅ **Complete** |
| 1     | **L0: PriorityTagger** – Tag critical processes  | ⏳ Pending     |
| 2     | **L1: AdaptiveTopK** – Keep busiest K processes | ⏳ Pending     |
| 3     | **L2: OthersRollup** – Aggregate the rest     | ⏳ Pending     |
| 4     | **L3: ReservoirSampler** – Sample the long tail  | ⏳ Pending     |
| 5     | Full "Opt-Plus" Pipeline (Hot + Cold Paths)    | ⏳ Pending     |

Detailed milestones and technical plans for each phase are documented in [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md).

---

## Quick Start

There are two primary ways to engage with this project:

### Option A ▪ Standalone Go Demo *(Illustrative Example Only)*

This option runs a minimal Go HTTP application. **It is NOT the OpenTelemetry collector.** It's a very simple, standalone example to conceptually demonstrate how a processor might modify data, using the original `main.go` content.

```bash
# 1. Clone the repository (if you haven't already)
git clone https://github.com/newrelic/nrdot-process-optimization.git
cd nrdot-process-optimization

# 2. Run the simple demo
go run main.go

# 3. Access in your browser:
#    → http://localhost:8080
```

### Option B ▪ Full Dockerized OpenTelemetry Collector *(Recommended for Development)*

This option builds and runs the actual custom OpenTelemetry collector (which includes the "Hello World" processor) within a Dockerized local development environment.

Prerequisites:
- Docker Engine installed and running.
- Docker Compose installed.

Steps:

```bash
# 1. Clone the repository (if you haven't already)
git clone https://github.com/newrelic/nrdot-process-optimization.git
cd nrdot-process-optimization

# 2. Build the custom OpenTelemetry Collector Docker image
#    (This uses build/Dockerfile and targets cmd/collector/main.go)
make docker-build

# 3. Start the local development stack using Docker Compose
#    (This uses build/docker-compose.yaml and config/base.yaml)
make compose-up

# 4. To view logs from all services (Collector, Mock Sink, etc.):
#    (Run in a new terminal or after detaching from compose-up if it wasn't run with -d)
make logs

# 5. When finished, stop the local development stack:
make compose-down
```

Local Development Stack Service URLs (when make compose-up is running):
| Service | URL | Purpose |
|---------|-----|---------|
| Collector zPages | http://localhost:55679 | Debugging collector internals (pipelines, etc.) |
| Prometheus UI | http://localhost:9090 | Querying metrics (collector & custom) |
| Grafana UI | http://localhost:3000 | Visualizing metrics (login: admin/admin) |
| Mock OTLP Sink Logs | View via make logs | Verifying OTLP data exported by the collector |

## "Hello World" Processor (Phase 0 Deliverable)

The initial phase of this project includes a fully functional "Hello World" processor integrated into the custom OpenTelemetry Collector. Its purpose is to demonstrate:

1. The fundamental structure of a custom OpenTelemetry processor (factory.go, config.go, processor.go).
2. Loading and using processor-specific configuration (from config/base.yaml).
3. Manipulating pmetric.Metrics data by adding custom attributes.
4. Core self-observability:
   - Using the standard obsreport package for common processor metrics.
   - Emitting custom processor-specific metrics (e.g., counters, gauges) via go.opentelemetry.io/otel/metric.

Functionality: As configured in config/base.yaml, the helloworld processor adds a hello.processor attribute (with a value like "Hello from NRDOT Process-Metrics Optimization!") to every metric data point. It can also be configured (add_to_resource: true) to add this attribute at the resource level.

Key Self-Metrics (visible in Prometheus/Grafana via the local dev stack):
| Metric Name | Type | Description |
|-------------|------|-------------|
| otelcol_processor_helloworld_processed_metric_points | Counter | Standard obsreport: Count of metric points processed. |
| nrdot_helloworld_mutations_total | Counter | Custom metric: Total number of metric points modified. |
| otelcol_processor_helloworld_latency_bucket | Histogram | Standard obsreport: Latency of processing. |

## Local Development Environment Deep Dive

The make compose-up command, utilizing build/docker-compose.yaml, orchestrates a multi-container environment:

1. **Custom OpenTelemetry Collector** (otel-collector service):
   - Built using build/Dockerfile, which compiles cmd/collector/main.go.
   - Configured by config/base.yaml, which defines a pipeline:
     - hostmetrics receiver (collecting process metrics, CPU, memory, etc.).
     - memory_limiter processor.
     - helloworld processor (our custom component).
     - batch processor.
     - Exporters: otlphttp (to the mock sink) and prometheus (for its own metrics).
   - Its zPages are exposed on http://localhost:55679.
   - Its Prometheus-compatible metrics endpoints are scraped by the Prometheus service.

2. **Mock OTLP Sink** (mock-otlp-sink service):
   - A simple service that listens for OTLP/HTTP metric data.
   - It logs the received metrics to standard output, allowing developers to inspect the data processed and exported by the collector. View these logs with make logs.

3. **Prometheus** (prometheus service):
   - Pre-configured to scrape metrics from:
     - The custom Collector's telemetry metrics endpoint (e.g., otel-collector:8888/metrics inside the Docker network).
     - The custom Collector's Prometheus exporter endpoint (e.g., otel-collector:8889/metrics inside the Docker network).
   - Accessible at http://localhost:9090.

4. **Grafana** (grafana service):
   - Pre-configured with Prometheus as a data source.
   - Automatically provisions the "NRDOT Custom Processor Starter KPIs" dashboard from dashboards/grafana-nrdot-custom-processor-starter-kpis.json.
   - Accessible at http://localhost:3000.

## Development Workflow & Key Makefile Targets

The Makefile is the primary entry point for most development tasks.
| Target | Action |
|--------|--------|
| make help | Displays a list of all available targets and their descriptions. |
| make build | Builds the custom OpenTelemetry collector binary (./bin/otelcol). |
| make test | Runs all tests: unit, integration (future), e2e (future), URL checks. |
| make test-unit | Runs only Go unit tests (go test ./...). |
| make test-urls | Runs test/url_check.sh to verify local dev stack service URLs. |
| make lint | Runs go vet, go fmt, and golangci-lint (if installed). |
| make docker-build | Builds the Docker image for the custom collector. |
| make compose-up | Starts the local development stack via Docker Compose. |
| make compose-down | Stops the local development stack. |
| make logs | Follows logs from all services in the Docker Compose stack. |
| make run | Starts services using run.sh docker up (alternative to compose-up). |
| make run-demo | Runs the standalone Go demo via run.sh demo. |
| make sbom | Generates a Software Bill of Materials (SBOM) using Syft (if installed). |
| make vuln-scan | Runs Go dependency vulnerability scanning (govulncheck). |
| make clean | Removes build artifacts. |

## Project Structure Overview

```
📁 nrdot-process-optimization/
├── .github/                            # GitHub Actions CI/CD Workflows
│   └── workflows/
│       └── build.yml                   # Main CI workflow
├── build/                              # Files related to building the collector
│   ├── Dockerfile                      # For building the custom collector Docker image
│   └── docker-compose.yaml             # For the local development stack
├── cmd/collector/                      # Entry point for the custom OpenTelemetry Collector
│   └── main.go                         # Registers processors, receivers, exporters
├── config/                             # Collector configuration files
│   └── base.yaml                       # Configuration for Phase 0 (hostmetrics + helloworld)
│   └── (opt-plus.yaml)                 # (Future) Configuration for the full L0-L3 pipeline
├── dashboards/                         # Grafana dashboard JSON definitions
│   └── grafana-nrdot-custom-processor-starter-kpis.json
├── docs/                               # Project documentation
│   ├── DEVELOPING_PROCESSORS.md        # Guide for creating new custom processors
│   └── NRDOT_PROCESSOR_SELF_OBSERVABILITY.md # Standards for processor metrics
├── main.go                             # Standalone "Hello World" demo application
├── processors/                         # Custom OpenTelemetry processors
│   └── helloworld/                     # Phase 0: Example "Hello World" processor
│   └── (prioritytagger/)               # (Future) L0: Critical process tagging
│   └── (adaptivetopk/)                 # (Future) L1: Top-K process selection
│   └── (othersrollup/)                 # (Future) L2: Non-priority/top process aggregation
│   └── (reservoirsampler/)             # (Future) L3: Statistical sampling
├── test/                               # Test suites and helper scripts
│   └── url_check.sh                    # Script to check local dev stack service URLs
│   └── (integration/)                  # (Future) Integration tests
│   └── (e2e/)                          # (Future) End-to-end tests
├── CLAUDE.md                           # Guidance for AI-assisted development
├── IMPLEMENTATION_PLAN.md              # Detailed project implementation roadmap
├── Makefile                            # Makefile for development tasks
├── go.mod                              # Go module definition
├── go.sum                              # Go module checksums
├── README.md                           # This file
├── run-demo.sh                         # Legacy script, prefer `make run-demo` or `run.sh demo`
└── run.sh                              # Unified script for running demo or Docker stack
```

(Parentheses () indicate planned or future components not yet fully implemented in Phase 0.)

## Configuration Highlights

### config/base.yaml (Current - Phase 0)

This file configures the collector for the "Hello World" demonstration and local development:

- **Receivers**:
  - hostmetrics: Collects process metrics, CPU, memory, disk, network, etc., from the host where the collector runs. collection_interval is configurable.

- **Processors**:
  - memory_limiter: Prevents the collector from consuming excessive memory.
  - helloworld: Our custom processor, configured with its message.
  - batch: Batches metrics before exporting, configured with send_batch_size and timeout.

- **Exporters**:
  - otlphttp: Exports metrics to an OTLP/HTTP endpoint. In docker-compose.yaml, this is typically re-routed to the local mock-otlp-sink service. For actual New Relic export, NEW_RELIC_LICENSE_KEY and the correct NEW_RELIC_OTLP_ENDPOINT are required.
  - prometheus: Exposes collector metrics on an endpoint (e.g., 0.0.0.0:8889) that the Prometheus service scrapes.

- **Service Pipelines**: Defines how data flows from receivers through processors to exporters.

- **Environment Variable Overrides**: Many key parameters (e.g., COLLECTION_INTERVAL, NEW_RELIC_LICENSE_KEY, OTLP endpoint) can be set via environment variables, with defaults specified in base.yaml.

### config/opt-plus.yaml (Future)

This configuration file will be developed in later phases. It will define the full, multi-layer optimization pipeline:
hostmetrics -> L0:PriorityTagger -> L1:AdaptiveTopK -> L2:OthersRollup -> L3:ReservoirSampler -> batch -> otlphttp (to New Relic)
It may also include a "cold path" exporter for raw data (e.g., to Parquet/S3) for archival or deep analysis.

## The Optimization Pipeline (Future Processors)

The core of this project is the development of a series of custom OpenTelemetry processors designed to work in concert:

1. **L0: PriorityTagger**: Identifies and tags metrics belonging to critical processes (e.g., based on executable name or high resource usage). Tagged metrics can bypass or receive special handling in subsequent layers.

2. **L1: AdaptiveTopK**: Selects metrics from the top 'K' most resource-intensive processes (e.g., by CPU or memory). The value of 'K' can be dynamically adjusted based on overall host load.

3. **L2: OthersRollup**: Aggregates metrics from all other non-priority, non-TopK processes into a single summary series (e.g., _other_ process). This drastically reduces cardinality while retaining a sense of overall system load from less significant processes.

4. **L3: ReservoirSampler**: Applies statistical sampling (e.g., reservoir sampling) to the remaining long-tail of processes not captured by L0, L1, or L2. This provides representative insight into a subset of these processes without exporting all of them.

## Contributing

Contributions are highly welcome!

- To understand the current development focus and upcoming features, please review the [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md).
- For guidance on developing new custom processors, see docs/DEVELOPING_PROCESSORS.md.
- For standards on processor self-observability and metrics, refer to docs/NRDOT_PROCESSOR_SELF_OBSERVABILITY.md.
- If using AI for assistance, CLAUDE.md provides project-specific context.

Please feel free to open issues for bugs, feature requests, or ideas. Pull requests are encouraged.