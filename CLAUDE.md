# CLAUDE.md - AI Development Assistant Guide

This document provides specific context and guidelines for AI coding assistants (like Claude, ChatGPT, etc.) when working with the **NRDOT Process-Metrics Optimization** repository. Its aim is to help the AI understand the project's goals, structure, current state, and conventions, leading to more effective and aligned assistance.

---

## 1. Project Core Objective & Current Status

* **Primary Goal:** To build a custom OpenTelemetry (OTel) Collector distribution that significantly reduces process metric ingest costs (aiming for ≥90% reduction) for New Relic, while preserving essential visibility into host processes.
* **Mechanism:** A multi-layer pipeline of custom OTel processors (L0-L3) that progressively Tag, Filter, Aggregate, and Sample process metrics.
* **Current Status (Phase 0 - Complete):**
  * A foundational "Hello World" (`helloworld`) custom processor is implemented and functional.
  * A robust local development environment using Docker Compose is established, including the custom Collector, Prometheus, Grafana, and a Mock OTLP Sink.
  * Standardized ports and service interactions are defined.
  * Core documentation (`README.md`, this `CLAUDE.md`, `IMPLEMENTATION_PLAN.md`, processor development guides) is in place.
  * CI for build, lint, unit tests, and basic vulnerability checks is operational.
* **Next Steps:** Proceed with implementing the L0-L3 optimization processors as outlined in the [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md).

---

## 2. Key Project Files & Directory Structure

Familiarize yourself with this structure to locate relevant code and configurations:

```
📁 nrdot-process-optimization/
├── .github/                            # GitHub Actions CI/CD Workflows
├── build/                              # Docker & Docker Compose files
│   ├── Dockerfile                      # Builds the custom collector image
│   └── docker-compose.yaml             # Orchestrates the local dev stack
├── cmd/collector/                      # Main entry point for the custom OTel Collector
│   └── main.go                         # Registers all components (processors, receivers, etc.)
├── config/                             # Collector configuration files
│   └── base.yaml                       # Config for Phase 0 (hostmetrics + helloworld pipeline)
│   └── (opt-plus.yaml)                 # (Future) Config for the full L0-L3 optimization pipeline
├── dashboards/                         # Grafana dashboard JSON files
│   └── grafana-nrdot-custom-processor-starter-kpis.json
├── docs/                               # Developer documentation and guides
│   ├── DEVELOPING_PROCESSORS.md        # How to build new custom processors
│   └── NRDOT_PROCESSOR_SELF_OBSERVABILITY.md # Standards for processor metrics
├── main.go                             # Standalone "Hello World" demo application
├── processors/                         # Location for all custom OTel processors
│   └── helloworld/                     # Phase 0: Example "Hello World" processor
│       ├── config.go                   # Configuration struct and validation
│       ├── factory.go                  # Processor factory implementation
│       ├── obsreport.go                # Helper for obsreport (can be integrated directly)
│       ├── processor.go                # Core processor logic (ConsumeMetrics, etc.)
│       └── processor_test.go           # Unit tests for the processor
│   └── (prioritytagger/)               # (Future) L0 Processor
│   └── (adaptivetopk/)                 # (Future) L1 Processor
│   └── (othersrollup/)                 # (Future) L2 Processor
│   └── (reservoirsampler/)             # (Future) L3 Processor
├── test/                               # Test suites and helper scripts
│   └── url_check.sh                    # Checks local dev stack service URLs
├── Makefile                            # Central build and task automation script
├── go.mod / go.sum                     # Go module definitions
├── README.md                           # Main project overview and quick start
├── IMPLEMENTATION_PLAN.md              # Phased development roadmap
└── CLAUDE.md                           # This file
```

(Parentheses () denote components planned for future phases.)

## 3. Core Development Workflow & Makefile Commands

The Makefile is your primary interface for common tasks. Refer to it (`make help`) for a full list.

### Most Frequently Used Targets:

| Command | Description |
|---------|-------------|
| `make build` | Compiles the custom OTel collector binary (`./bin/otelcol`). |
| `make docker-build` | Builds the Docker image for the custom collector using `build/Dockerfile`. |
| `make compose-up` | Starts the local development stack (Collector, Prometheus, Grafana, Mock Sink) via Docker Compose. |
| `make compose-down` | Stops the local development stack. |
| `make logs` | Tails logs from all services in the Docker Compose stack. |
| `make test` | Runs all available tests (unit, URL checks; future: integration, E2E). |
| `make test-unit` | Runs Go unit tests (`go test ./...`). |
| `make lint` | Runs Go static analysis tools (`go vet`, `go fmt`, `golangci-lint` if available). |
| `make run-demo` | Executes the simple Go demo in `main.go`. |

### Typical Local Development Loop for a Processor:

1. Make code changes within a `processors/<name>/` directory.
2. Write/update unit tests (`_test.go` files).
3. Run `make test-unit` and `make lint` frequently.
4. Rebuild the collector image: `make docker-build`.
5. Restart the local stack with the new image: `make compose-up` (it will stop and recreate the collector service).
6. Observe behavior:
   - Check logs: `make logs` (especially the `mock-otlp-sink` and `otel-collector`).
   - Inspect zPages: http://localhost:55679.
   - Query metrics in Prometheus: http://localhost:9090.
   - View dashboards in Grafana: http://localhost:3000.
7. Repeat until the processor behaves as expected.

## 4. Coding Standards & Processor Development Conventions

Adherence to these standards is crucial for consistency and maintainability.

### Go Version: 
Go 1.22 or higher.

### Formatting & Linting:
Code must pass `go fmt`, `go vet`. `golangci-lint` (configured in `.golangci.yml` if present, or via `make lint`) is highly encouraged.

### Processor Structure:
Each custom processor resides in its own sub-directory under `processors/` (e.g., `processors/myprocessor/`) and typically includes:

- **config.go**: Defines the processor's configuration struct (implementing `component.Config`, `config.Validator`, and `confmap.Unmarshaler`), including default values.
- **factory.go**: Implements `processor.Factory` (using `processor.NewFactory`) to create instances of the processor and its default configuration. Specifies the processor `typeStr` (e.g., "myprocessor") and stability level.
- **processor.go**: Contains the core logic, implementing `processor.Metrics` (primarily the `ConsumeMetrics` method). It also handles capabilities (`MutatesData`) and lifecycle hooks (`Start`, `Shutdown`).
- **processor_test.go**: Unit tests for the processor, covering configuration, metric transformation, edge cases, and error handling.
- **obsreport.go** (Optional): If a dedicated helper for obsreport is preferred. Otherwise, obsreport usage can be directly integrated into processor.go. The helloworld processor shows an example.

### Self-Observability:

- **Standard Metrics**: All processors MUST use `go.opentelemetry.io/collector/obsreport` (e.g., `obsreport.NewProcessor` and its `StartMetricsOp`/`EndMetricsOp` methods) to emit standard OTel processor metrics like `otelcol_processor_<name>_processed_metric_points`, `otelcol_processor_<name>_dropped_metric_points`, and latency histograms.
- **Custom Metrics**: Implement processor-specific Key Performance Indicators (KPIs) using `go.opentelemetry.io/otel/metric` obtained via `component.TelemetrySettings.MeterProvider.Meter("<processor_type_str>")`. Custom metric names MUST be prefixed with `nrdot_<processor_name>_` (e.g., `nrdot_helloworld_mutations_total`).

### pdata Manipulation:
Processors operate on `pmetric.Metrics`. Be mindful of efficient manipulation of these structures.

### Error Handling:
Log errors using the `zap.Logger` provided in `processor.CreateSettings`. Propagate errors up the call stack where appropriate.

### Configuration:
Define clear, well-documented configuration options in `config.go`. Provide sensible defaults. Ensure validation logic is robust.

### Testing:

- Aim for high unit test coverage (≥80%) for all processor logic.
- Use table-driven tests for varying inputs and conditions.
- Utilize `go.opentelemetry.io/collector/consumer/consumertest` (e.g., `consumertest.NewNop()`) to mock downstream consumers in tests.
- For pmetric generation in tests, use helpers from `go.opentelemetry.io/collector/pdata/testdata` or construct `pmetric.Metrics` objects directly.

### Idempotency:
Consider if a processor's actions should be idempotent, especially if metrics might be reprocessed under certain conditions (though rare in typical OTel flows).

### Documentation:
Each new processor should have a README.md in its directory explaining its purpose, configuration options, and any notable behaviors.

## 5. Local Development Environment Observability

When the local stack is running (`make compose-up`), use these endpoints for observation and debugging:

| Service | URL | Key Usage for Processor Development |
|---------|-----|--------------------------------------|
| Collector zPages | http://localhost:55679 | View active pipelines, component status, basic metric counts. Useful for checking if your processor is loaded and receiving data. |
| Prometheus UI | http://localhost:9090 | Query for standard `otelcol_processor_*` metrics and your custom `nrdot_<processor_name>_*` metrics. Example query: `rate(nrdot_helloworld_mutations_total[1m])`. |
| Grafana UI | http://localhost:3000 | View the "NRDOT Custom Processor Starter KPIs" dashboard. Add panels for your new processor's metrics to this dashboard or a new one. |
| Mock OTLP Sink Logs | Via `make logs` | Inspect the actual OTLP metric data being exported by the collector after it has passed through your processor. Verify attribute changes, filtering, aggregation, etc. |
| Collector Service Logs | Via `make logs` | Check for logs from your processor (e.g., debug statements, error messages). Filter for the `otel-collector` service. |

## 6. Configuration Files Overview

- **config/base.yaml**: This is the active configuration for local development (Phase 0). It defines a pipeline including the `hostmetrics` receiver, `helloworld` processor, and standard processors like `memory_limiter` and `batch`. It exports to the `mock-otlp-sink` and `prometheus`. When developing a new processor, you'll add it to a pipeline in this file (or a copy) for testing.

- **config/opt-plus.yaml** (Future): This will be the target configuration for the full L0-L3 optimization pipeline, intended for production-like scenarios. It will include `prioritytagger`, `adaptivetopk`, `othersrollup`, and `reservoirsampler`.

- **Environment Variables**: Key settings in `config/base.yaml` (and the future `opt-plus.yaml`) can be overridden by environment variables (e.g., `COLLECTION_INTERVAL`, `NEW_RELIC_LICENSE_KEY`, `NEW_RELIC_OTLP_ENDPOINT`). The `docker-compose.yaml` file often sets some of these, especially to route the OTLP exporter to the `mock-otlp-sink` service within the Docker network.

## 7. General Guidance for AI Prompts

- **Be Specific**: Instead of "write a processor," ask "Create the factory.go file for a new OpenTelemetry processor named 'myprocessor', ensuring it registers a default configuration struct named Config and uses processor.StabilityLevelDevelopment."

- **Provide Context**: Reference existing files (e.g., "Similar to processors/helloworld/processor.go, implement the ConsumeMetrics method for myprocessor...").

- **Iterate**: Ask for one part at a time (e.g., config struct, then factory, then processor logic).

- **Request Tests**: "Write unit tests for the Validate() method in processors/myprocessor/config.go."

- **Focus on OTel SDKs**: Emphasize usage of `go.opentelemetry.io/collector/component`, `pdata`, `consumer`, `processor` packages.

This guide helps ensure AI contributions align with project standards and accelerate development. Refer to specific `docs/*.md` files for deeper technical details on processor development and observability.