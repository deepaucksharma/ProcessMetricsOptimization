# NRDOT Process-Metrics Optimization

A production-grade, **multi-layer** process-metrics pipeline built as a custom **New Relic Distribution of OpenTelemetry (NRDOT)** collector.
Its core mission is to **significantly reduce metric data volume (aiming for ≥ 90% reduction)** by intelligently refining process metric data, while preserving deep visibility into the processes that are critical for operational insight.

The pipeline achieves this by progressively **Tagging → Filtering → Aggregating → Sampling** process metrics through a series of custom OpenTelemetry processors (L0 – L3).

---

## Project Status & Roadmap Snapshot

The project is developed in phases. **All phases are now complete, including the integration of all processors in a single pipeline.**

| Phase | Focus                                          | Status        |
|-------|------------------------------------------------|---------------|
| 0     | Foundation: Hello World processor & Local Dev Stack | ✅ **Complete** |
| 1     | **L0: PriorityTagger** – Tag critical processes  | ✅ **Complete** |
| 2     | **L1: AdaptiveTopK** – Keep busiest K processes | ✅ **Complete** |
| 3     | **L2: OthersRollup** – Aggregate the rest     | ✅ **Complete** |
| 4     | **L3: ReservoirSampler** – Sample the long tail  | ✅ **Complete** |
| 5     | Full "Opt-Plus" Pipeline Integration & Testing  | ✅ **Complete** |

Detailed milestones and technical plans for each phase are documented in [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md).

---

## Quick Start

### Dockerized OpenTelemetry Collector

This option builds and runs the actual custom OpenTelemetry collector (which includes all processors) within a Dockerized local development environment.

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
#    (Uses build/docker-compose.yaml and config/opt-plus.yaml by default)
make compose-up
# To use a custom collector config:
# COLLECTOR_CONFIG=my-config.yaml make compose-up

# 4. To view logs from all services (Collector, Mock Sink, etc.):
#    (Run in a new terminal or after detaching from compose-up if it wasn't run with -d)
make logs

# 5. When finished, stop the local development stack:
make compose-down
```

### Testing the Full Optimization Pipeline

To test the complete optimization pipeline with all processors (PriorityTagger, AdaptiveTopK, ReservoirSampler, and OthersRollup), you can use either:

**Option 1:** Run the automated test script:
```bash
# Builds the collector, starts the stack with opt-plus.yaml, and verifies all processors
./test/test_opt_plus_pipeline.sh
# Use a custom config file
COLLECTOR_CONFIG=my-config.yaml ./test/test_opt_plus_pipeline.sh
```

**Option 2:** Use the Makefile target:
```bash
# Build the Docker image first
make docker-build

# Start the full optimization pipeline with opt-plus.yaml
make opt-plus-up
# With a custom config file
COLLECTOR_CONFIG=my-config.yaml make opt-plus-up

# View the logs to see the pipeline in action
make logs
```

This script will:
1. Build the collector with all processors
2. Start the stack with the opt-plus.yaml configuration
3. Verify that all services are accessible
4. Check Prometheus for processor metrics
5. Examine the OTLP sink for evidence of pipeline functionality
6. Calculate the cardinality reduction percentage (aiming for ≥90%)
7. Verify that critical processes are preserved

After the test completes successfully, you can explore the results in Grafana, Prometheus, and the mock OTLP sink logs.

### Customizing the Pipeline

The optimization pipeline can be customized using environment variables:

```bash
# Copy the example file
cp .env.example .env

# Edit the variables to match your requirements
vim .env

# Start the pipeline with your custom settings
make opt-plus-up
```

Key customization options include:
- Collection interval
- New Relic credentials
- Critical process criteria
- TopK value and dynamic K parameters
- Reservoir size
- Output attributes for rollup
- Collector configuration file (COLLECTOR_CONFIG)

Local Development Stack Service URLs (when make compose-up is running):
| Service | URL | Purpose |
|---------|-----|---------|
| Collector zPages | http://localhost:15679 | Debugging collector internals (pipelines, etc.) |
| Prometheus UI | http://localhost:19090 | Querying metrics (collector & custom) |
| Grafana UI | http://localhost:13000 | Visualizing metrics (login: admin/admin) |
| Mock OTLP Sink Logs | View via make logs | Verifying OTLP data exported by the collector |

## "Hello World" Processor (Phase 0 Deliverable)

The initial phase of this project includes a fully functional "Hello World" processor integrated into the custom OpenTelemetry Collector. Its purpose is to demonstrate:

1. The fundamental structure of a custom OpenTelemetry processor (factory.go, config.go, processor.go).
2. Loading and using processor-specific configuration.
3. Manipulating pmetric.Metrics data by adding custom attributes.
4. Core self-observability:
   - Using the standard obsreport package for common processor metrics.
   - Emitting custom processor-specific metrics (e.g., counters, gauges) via go.opentelemetry.io/otel/metric.

Functionality: The helloworld processor adds a hello.processor attribute (with a value like "Hello from NRDOT Process-Metrics Optimization!") to every metric data point. It can also be configured (add_to_resource: true) to add this attribute at the resource level.

Key Self-Metrics (visible in Prometheus/Grafana via the local dev stack):
| Metric Name | Type | Description |
|-------------|------|-------------|
| otelcol_otelcol_processor_helloworld_processed_metric_points | Counter | Standard obsreport: Count of metric points processed. |
| otelcol_otelcol_helloworld_mutations_total | Counter | Custom metric: Total number of metric points modified. |

## "PriorityTagger" Processor (Phase 1 Deliverable)

The second phase implements the first layer (L0) of our optimization pipeline - the `prioritytagger` processor. This processor:

1. Identifies and tags process metrics that are considered critical based on various criteria
2. Preserves full visibility for these critical processes in subsequent pipeline stages
3. Serves as the foundation for the selective optimization approach

Key Features:
- Tag processes as critical based on:
  - Exact executable name matches (e.g., "systemd", "kubelet")
  - Regular expression pattern matches (e.g., "kube.*", "docker.*")
  - CPU utilization thresholds
  - Memory RSS thresholds
- Add configurable attributes to mark critical processes (default: "nr.priority=critical")
- Support for both integer and double type metric values
- Comprehensive test coverage and observability

Configuration (opt-plus.yaml example):
```yaml
processors:
  prioritytagger:
    critical_executables:
      - systemd
      - kubelet
      - dockerd
      - containerd
      - chronyd
      - sshd
    critical_executable_patterns:
      - ".*java.*"
      - ".*node.*"
      - "kube.*"
      - ".*otelcol.*"
    cpu_steady_state_threshold: 0.25
    memory_rss_threshold_mib: 400
    priority_attribute_name: "nr.priority"
    critical_attribute_value: "critical"
```

Key Self-Metrics:
| Metric Name | Type | Description |
|-------------|------|-------------|
| otelcol_otelcol_processor_prioritytagger_processed_metric_points | Counter | Standard obsreport: Count of metric points processed. |
| otelcol_processor_dropped_metric_points | Counter | Standard obsreport: Count of metric points dropped due to errors. |
| otelcol_otelcol_prioritytagger_critical_processes_tagged_total | Counter | Custom: Count of processes tagged as critical. |

For detailed documentation, see [processors/prioritytagger/README.md](processors/prioritytagger/README.md).

## "AdaptiveTopK" Processor (Phase 2 Deliverable)

The third phase implements the second layer (L1) of our optimization pipeline - the `adaptivetopk` processor. This processor:

1. Selects metrics from the 'K' most resource-intensive processes (by CPU or memory usage)
2. Always passes through critical processes already tagged by the PriorityTagger
3. Dramatically reduces data volume by dropping metrics from less significant processes

Key Features:
- Configurable fixed K value (number of top processes to keep)
- Dynamic K adjustment based on system load
- Process hysteresis to prevent thrashing when process rankings change
- Flexible process ranking based on any metric (e.g., CPU utilization, memory usage)
- Optional secondary ranking metric for tie-breaking
- Efficient implementation using a min-heap algorithm

Configuration (example):
```yaml
processors:
  adaptivetopk:
    # Dynamic K configuration
    host_load_metric_name: "system.cpu.utilization"
    load_bands_to_k_map:
      0.15: 5    # Very low system load -> keep fewer processes
      0.3: 10    # Low system load
      0.5: 15    # Medium load
      0.7: 25    # High load
      0.85: 40   # Very high load
    hysteresis_duration: "30s"
    min_k_value: 5
    max_k_value: 50

    # Fixed K (not used when Dynamic K is enabled)
    # k_value: 10

    # The metric used to rank processes
    key_metric_name: "process.cpu.utilization"
    # Optional: Secondary metric for tie-breaking
    secondary_key_metric_name: "process.memory.rss"
    # Attribute name used to identify critical processes
    priority_attribute_name: "nr.priority"
    critical_attribute_value: "critical"
```

Key Self-Metrics:
| Metric Name | Type | Description |
|-------------|------|-------------|
| otelcol_otelcol_processor_adaptivetopk_processed_metric_points | Counter | Count of metric points processed. |
| otelcol_processor_adaptivetopk_dropped_metric_points | Counter | Count of metric points dropped. |
| otelcol_otelcol_adaptivetopk_topk_processes_selected_total | Counter | Number of non-critical processes selected for Top K. |
| otelcol_otelcol_adaptivetopk_current_k_value | Gauge | Current K value when using dynamic K. |

For detailed documentation, see [processors/adaptivetopk/README.md](processors/adaptivetopk/README.md).

## "OthersRollup" Processor (Phase 3 Deliverable)

The fourth phase implements the third layer (L2) of our optimization pipeline - the `othersrollup` processor. This processor:

1. Aggregates metrics from non-priority, non-TopK processes into a single "_other_" summary series
2. Preserves overall system visibility while dramatically reducing metric cardinality
3. Allows flexible aggregation strategies (sum, average) per metric type

Key Features:
- Configurable output attributes for the rolled-up "_other_" series
- Metric-specific aggregation functions (e.g., average for CPU, sum for memory)
- Preserves metric type semantics (gauge vs. sum)
- Filter capability to target specific metrics for rollup
- Allows visibility into total system resource usage with minimal cardinality

Configuration (example):
```yaml
processors:
  othersrollup:
    # Attribute values for the rolled-up series
    output_pid_attribute_value: "-1"
    output_executable_name_attribute_value: "_other_"
    # Map of metric names to aggregation functions
    aggregations:
      "process.cpu.utilization": "avg"
      "process.memory.rss": "sum"
    # List of metrics to apply rollup to (if empty, applies to compatible metrics)
    metrics_to_rollup:
      - "process.cpu.utilization"
      - "process.memory.rss"
    # Attributes used to identify critical processes (should not be rolled up)
    priority_attribute_name: "nr.priority"
    critical_attribute_value: "critical"
```

Key Self-Metrics:
| Metric Name | Type | Description |
|-------------|------|-------------|
| otelcol_otelcol_processor_othersrollup_processed_metric_points | Counter | Count of metric points processed. |
| otelcol_processor_othersrollup_dropped_metric_points | Counter | Count of original points dropped after rollup. |
| otelcol_otelcol_othersrollup_aggregated_series_count_total | Counter | Number of new "_other_" series generated. |
| otelcol_otelcol_othersrollup_input_series_rolled_up_total | Counter | Count of input series aggregated into "_other_". |

For detailed documentation, see [processors/othersrollup/README.md](processors/othersrollup/README.md).

## "ReservoirSampler" Processor (Phase 4 Deliverable)

The fifth phase implements the fourth layer (L3) of our optimization pipeline - the `reservoirsampler` processor. This processor:

1. Selects a statistically representative sample of metrics from "long-tail" processes
2. Maintains a reservoir of unique process identities with proper statistical properties
3. Adds sampling information to enable accurate scaling during analysis

Key Features:
- Configurable reservoir size (number of unique processes to sample)
- Flexible process identity definition using any combination of attributes
- Implements Algorithm R variant for statistically sound sampling
- Adds sample rate metadata to enable proper scaling during analysis
- Pass-through behavior for critical processes already tagged

Configuration (example):
```yaml
processors:
  reservoirsampler:
    # The number of unique process identities to keep in the reservoir
    reservoir_size: 125
    # List of attribute keys that define a unique process identity
    identity_attributes:
      - "process.pid"
      - "process.executable.name"
      - "process.command_line"
    # Attributes to add to sampled metrics
    sampled_attribute_name: "nr.process_sampled_by_reservoir"
    sampled_attribute_value: "true"
    sample_rate_attribute_name: "nr.sample_rate"
    # Attributes used to identify critical processes
    priority_attribute_name: "nr.priority"
    critical_attribute_value: "critical"
```

Key Self-Metrics:
| Metric Name | Type | Description |
|-------------|------|-------------|
| otelcol_otelcol_processor_reservoirsampler_processed_metric_points | Counter | Count of metric points processed. |
| otelcol_processor_reservoirsampler_dropped_metric_points | Counter | Count of points dropped (eligible but not sampled). |
| otelcol_otelcol_reservoirsampler_reservoir_fill_ratio | Gauge | Current fill ratio of the reservoir (0.0 to 1.0). |
| otelcol_otelcol_reservoirsampler_selected_identities_count | Gauge | Current count of unique identities in the reservoir. |
| otelcol_otelcol_reservoirsampler_eligible_identities_seen_total | Counter | Total unique eligible identities encountered. |
| otelcol_otelcol_reservoirsampler_new_identities_added_to_reservoir_total | Counter | Count of new identities added to the reservoir. |

For detailed documentation, see [processors/reservoirsampler/README.md](processors/reservoirsampler/README.md).

## Full "Opt-Plus" Pipeline Integration (Phase 5 Deliverable)

The final phase integrates all four processors into a complete optimization pipeline with comprehensive testing, monitoring, and documentation:

1. Completed implementation of Dynamic K evaluation for the AdaptiveTopK processor
2. Created comprehensive end-to-end testing for the entire pipeline
3. Developed detailed Grafana dashboards for monitoring the pipeline
4. Optimized processor configurations for maximum performance and cardinality reduction

Key Features:
- Dynamic K adjustment based on system load with configurable threshold bands
- Process hysteresis to prevent thrashing when rankings change
- Full pipeline validation with cardinality reduction measurement
- Algorithm-specific dashboards for deep insights into each processor
- Optimized pipeline order for maximum effectiveness

For comprehensive documentation of the complete optimization pipeline, see [docs/OPTIMIZATION_PIPELINE.md](docs/OPTIMIZATION_PIPELINE.md).

## Monitoring & Observability

### Grafana Dashboards

This project includes several Grafana dashboards for comprehensive monitoring:

1. **Full Pipeline Overview Dashboard**
   - Provides a holistic view of the entire optimization pipeline
   - Shows collector health, processor performance, and optimization impact
   - Displays cardinality reduction and cost savings metrics

2. **Algorithm-Specific Dashboards**
   - PriorityTagger Algorithm Dashboard
   - AdaptiveTopK Algorithm Dashboard
   - OthersRollup Algorithm Dashboard
   - ReservoirSampler Algorithm Dashboard

These dashboards show the internal decision-making processes of each processor algorithm, providing deep visibility into how cardinality reduction is achieved.

For detailed information on dashboard setup and usage, see [dashboards/README.md](dashboards/README.md) and [docs/GRAFANA_DASHBOARD_DESIGN.md](docs/GRAFANA_DASHBOARD_DESIGN.md).

## Local Development Environment Deep Dive

The make compose-up command, utilizing build/docker-compose.yaml, orchestrates a multi-container environment:

1. **Custom OpenTelemetry Collector** (otel-collector service):
   - Built using build/Dockerfile, which compiles cmd/collector/main.go.
   - Configured by config/opt-plus.yaml, which defines a pipeline:
     - hostmetrics receiver (collecting process metrics, CPU, memory, etc.).
     - memory_limiter processor.
     - prioritytagger processor (L0 - tagging critical processes).
     - adaptivetopk processor (L1 - keeping top K resource-intensive processes).
     - reservoirsampler processor (L3 - sampling a subset of remaining processes).
     - othersrollup processor (L2 - aggregating non-critical, non-TopK, non-sampled processes).
     - batch processor.
     - Exporters: otlphttp (to the mock sink) and prometheus (for its own metrics).
   - Its zPages are exposed on http://localhost:15679 (mapped from port 55679 in the container).
   - Its Prometheus-compatible metrics endpoints are scraped by the Prometheus service. The setup for this stack is detailed in [docs/OBSERVABILITY_STACK_SETUP.md](docs/OBSERVABILITY_STACK_SETUP.md).

2. **Mock OTLP Sink** (mock-otlp-sink service):
   - A simple service that listens for OTLP/HTTP metric data.
   - It logs the received metrics to standard output, allowing developers to inspect the data processed and exported by the collector. View these logs with make logs.

3. **Prometheus** (prometheus service):
   - Pre-configured to scrape metrics from:
     - The custom Collector's telemetry metrics endpoint (e.g., otel-collector:8888/metrics inside the Docker network).
     - The custom Collector's Prometheus exporter endpoint (e.g., otel-collector:8889/metrics inside the Docker network).
   - Accessible at http://localhost:19090. For details on integrating Prometheus and Grafana with the collector, see [docs/OBSERVABILITY_STACK_SETUP.md](docs/OBSERVABILITY_STACK_SETUP.md).

4. **Grafana** (grafana service):
   - Pre-configured with Prometheus as a data source.
   - Automatically provisions dashboards from the dashboards/ directory.
   - Accessible at http://localhost:13000.
   - For a more detailed guide on creating comprehensive dashboards, see [docs/GRAFANA_DASHBOARD_DESIGN.md](docs/GRAFANA_DASHBOARD_DESIGN.md). For details on integrating Prometheus and Grafana with the collector, see [docs/OBSERVABILITY_STACK_SETUP.md](docs/OBSERVABILITY_STACK_SETUP.md).

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
| make opt-plus-up | Starts the stack with the full optimization pipeline configuration. |
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
│   └── opt-plus.yaml                   # Configuration for the full L0-L3 optimization pipeline
├── dashboards/                         # Grafana dashboard JSON definitions
│   ├── README.md                       # Dashboard documentation and overview
│   ├── grafana-nrdot-custom-processor-starter-kpis.json
│   ├── grafana-nrdot-optimization-pipeline.json
│   ├── grafana-nrdot-prioritytagger-algo.json
│   ├── grafana-nrdot-adaptivetopk-algo.json
│   ├── grafana-nrdot-othersrollup-algo.json
│   ├── grafana-nrdot-reservoirsampler-algo.json
│   ├── grafana-nrdot-prioritytagger-processor.json
│   └── grafana-nrdot-system-overview.json
├── docs/                               # Project documentation
│   ├── COMPLETING_PHASE_5.md           # Guide for completing the full pipeline integration
│   ├── DEVELOPING_PROCESSORS.md        # Guide for creating new custom processors
│   ├── NRDOT_PROCESSOR_SELF_OBSERVABILITY.md # Standards for processor metrics
│   ├── GRAFANA_DASHBOARD_DESIGN.md     # Detailed guide for creating Grafana dashboards
│   ├── OBSERVABILITY_STACK_SETUP.md    # Guide for setting up Prometheus, Grafana with the collector
│   ├── OPTIMIZATION_PIPELINE.md        # Detailed documentation of the complete pipeline
│   └── PIPELINE_DIAGRAM.md             # Visual documentation of the pipeline architecture
├── examples/                           # Example code directory
│   └── README.md                       # Describes planned examples for the project
├── internal/                           # Internal shared packages
│   └── banding/                        # Support for AdaptiveTopK processor implementation
│       └── README.md                   # Documents functionality for banding package
├── processors/                         # Custom OpenTelemetry processors
│   └── helloworld/                     # Phase 0: Example "Hello World" processor
│   └── prioritytagger/                 # Phase 1: L0: Critical process tagging
│   └── adaptivetopk/                   # Phase 2: L1: Top-K process selection
│   └── othersrollup/                   # Phase 3: L2: Non-priority/top process aggregation
│   └── reservoirsampler/               # Phase 4: L3: Statistical sampling
├── test/                               # Test suites and helper scripts
│   └── url_check.sh                    # Script to check local dev stack service URLs
│   └── test_opt_plus_pipeline.sh       # End-to-end test for the full optimization pipeline
├── CLAUDE.md                           # Guidance for AI-assisted development
├── IMPLEMENTATION_PLAN.md              # Detailed project implementation roadmap
├── Makefile                            # Makefile for development tasks
├── go.mod                              # Go module definition
├── go.sum                              # Go module checksums
├── README.md                           # This file
└── run.sh                              # Unified script for running Docker stack
```

## The Optimization Pipeline

The core of this project is the development of a series of custom OpenTelemetry processors designed to work in concert. The diagram below shows the complete optimization pipeline:

```
┌───────────────┐    ┌─────────────────┐    ┌────────────────┐    ┌──────────────────┐    ┌────────────────┐
│ HOSTMETRICS   │    │ PRIORITYTAGGER  │    │ ADAPTIVETOPK   │    │ RESERVOIRSAMPLER │    │ OTHERSROLLUP   │
│ RECEIVER      │───>│ PROCESSOR (L0)  │───>│ PROCESSOR (L1) │───>│ PROCESSOR (L3)   │───>│ PROCESSOR (L2) │──> EXPORTERS
│ Process Data  │    │ Tag Critical    │    │ Keep Top K     │    │ Sample Long-Tail │    │ Aggregate Rest │
└───────────────┘    └─────────────────┘    └────────────────┘    └──────────────────┘    └────────────────┘
                           │                      │                       │                      │
                           ▼                      ▼                       ▼                      ▼
                     ┌─────────────────────────────────────────────────────────────────────────────────┐
                     │                        CARDINALITY REDUCTION EFFECT                              │
                     ├─────────────────┬────────────────┬──────────────────┬────────────────┬──────────┤
                     │ All Process     │ Critical       │ Top K Highest    │ Representative │ "_other_"│
                     │ Metrics         │ Process Metrics│ Resource Usage   │ Sample         │ Rollup   │
                     │ (100%)          │ (~5-15%)       │ (~5-10%)         │ (~1-5%)        │ (1 series)
                     └─────────────────┴────────────────┴──────────────────┴────────────────┴──────────┘
```

This multi-layer approach strategically reduces cardinality while preserving visibility into important processes:


1. **L0: PriorityTagger** ✅: Identifies and tags metrics belonging to critical processes (e.g., based on executable name, regex patterns, CPU utilization, or memory usage). Tagged metrics can bypass or receive special handling in subsequent layers. Supports both integer and double metric value types for proper threshold comparison.

2. **L1: AdaptiveTopK** ✅: Selects metrics from the top 'K' most resource-intensive processes (e.g., by CPU or memory). Uses an efficient min-heap algorithm to identify the highest-resource processes. Features dynamic K adjustment based on system load with hysteresis to prevent thrashing.

3. **L2: OthersRollup** ✅: Aggregates metrics from all other non-priority, non-TopK processes into a single summary series (e.g., _other_ process). This drastically reduces cardinality while retaining a sense of overall system load from less significant processes.

4. **L3: ReservoirSampler** ✅: Applies statistical sampling (e.g., reservoir sampling) to the remaining long-tail of processes. This provides representative insight into a subset of these processes without exporting all of them. Adds sample rate metadata to enable proper scaling during analysis.

For a detailed visualization of the pipeline and data transformations, see [docs/PIPELINE_DIAGRAM.md](docs/PIPELINE_DIAGRAM.md).

## Contributing

Contributions are highly welcome!

- To understand the current development focus and upcoming features, please review the [IMPLEMENTATION_PLAN.md](IMPLEMENTATION_PLAN.md).
- For guidance on developing new custom processors, see docs/DEVELOPING_PROCESSORS.md.
- For standards on processor self-observability and metrics, refer to docs/NRDOT_PROCESSOR_SELF_OBSERVABILITY.md.
- For detailed guidance on designing effective Grafana dashboards for this project, see [docs/GRAFANA_DASHBOARD_DESIGN.md](docs/GRAFANA_DASHBOARD_DESIGN.md).
- For setting up the overall observability stack with the NRDOT collector, refer to [docs/OBSERVABILITY_STACK_SETUP.md](docs/OBSERVABILITY_STACK_SETUP.md).
- If using AI for assistance, CLAUDE.md provides project-specific context.

Please feel free to open issues for bugs, feature requests, or ideas. Pull requests are encouraged.