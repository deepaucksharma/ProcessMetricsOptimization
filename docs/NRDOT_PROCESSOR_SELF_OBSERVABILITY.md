# NRDOT Processor Self-Observability

## Guiding Philosophy

**Instrument once, visualize everywhere** - All NRDOT custom processors must implement a comprehensive self-observability framework that allows their performance, capacity, and behavior to be monitored across multiple surfaces. The setup of the overall observability stack is detailed in [docs/OBSERVABILITY_STACK_SETUP.md](docs/OBSERVABILITY_STACK_SETUP.md).

## Required Metrics

### Standard OTel Metrics (via obsreport)

| Metric | Type | Description |
|--------|------|-------------|
| `otelcol_otelcol_otelcol_processor_<name>_processed_metric_points` | Counter | Number of metric points processed |
| `otelcol_processor_dropped_metric_points` | Counter | Number of metric points dropped |
| `otelcol_processor_refused_metric_points` | Counter | Number of metric points refused |
| `otelcol_processor_latency_bucket` | Histogram | Processing time |

### Custom Processor KPIs

 - All custom metrics **MUST** use the prefix `otelcol_otelcol_<processor_name>_`
- Implement using `go.opentelemetry.io/otel/metric` via `component.TelemetrySettings`

**Implemented Examples:**

| Processor | Metric | Type | Description |
|-----------|--------|------|-------------|
| **HelloWorld** | `otelcol_otelcol_helloworld_mutations_total` | Counter | Number of metric points modified |
| **PriorityTagger** | `otelcol_otelcol_prioritytagger_critical_processes_tagged_total` | Counter | Number of unique processes tagged as critical |
| **AdaptiveTopK** | `otelcol_otelcol_adaptivetopk_topk_processes_selected_total` | Counter | Number of non-critical processes selected for TopK |
| **AdaptiveTopK** | `otelcol_otelcol_adaptivetopk_current_k_value` | Gauge | Current K value in use (for Dynamic K) |
| **OthersRollup** | `otelcol_otelcol_othersrollup_aggregated_series_count_total` | Counter | Number of new "_other_" series generated |
| **OthersRollup** | `otelcol_otelcol_othersrollup_input_series_rolled_up_total` | Counter | Number of input series aggregated |
| **ReservoirSampler** | `otelcol_otelcol_reservoirsampler_reservoir_fill_ratio` | Gauge | Current reservoir fill ratio (0.0 to 1.0) |
| **ReservoirSampler** | `otelcol_otelcol_reservoirsampler_selected_identities_count` | Gauge | Current count of unique process identities in reservoir |
| **ReservoirSampler** | `otelcol_otelcol_reservoirsampler_eligible_identities_seen_total` | Counter | Total unique eligible process identities encountered |

## Observability Surfaces

| Surface | URL | Purpose |
|---------|-----|---------|
| **zPages** | http://localhost:15679 | Shows pipelines, metric stats, and traces |
| **Prometheus** | http://localhost:19090 | Exposes all metrics for querying |
| **Grafana** | http://localhost:13000 | Visualizes dashboards for processor metrics |
| **Mock OTLP Sink** | Via `make logs` | Shows actual data after processing |

## Development & Testing Workflow

### 1. Local Testing Setup

```bash
# Build the Docker image
make docker-build

# Option 1: Start with base configuration (HelloWorld + PriorityTagger)
make compose-up

# Option 2: Start with full optimization pipeline (all processors)
make opt-plus-up

# Option 3: Run the automated pipeline test
./test/test_opt_plus_pipeline.sh

# View logs from all services
make logs

# Stop the stack when done
make compose-down
```

### 2. Validation Process

1. Check **zPages** to verify:
   - Processor is loaded correctly
   - Pipeline is functioning
   - Data is flowing through the processor

2. Check **Prometheus** to verify:
   - Standard metrics are being emitted
   - Custom metrics are being emitted
   - Values are changing as expected

3. Check **Grafana** to visualize:
   - Processor performance over time
   - Key metrics on the "NRDOT Processors" dashboard
   - Relationships between different metrics

4. Check **Mock Sink logs** to verify:
   - Processor is modifying data as expected
   - Output format is correct
   - No unexpected data loss

## Production Monitoring Guidelines

### Metrics Strategy

- All `otelcol_processor_*` metrics should be sent to monitoring backend
- Custom dashboard for visualizing processor-specific KPIs
- Alerts on critical thresholds and anomalies

### Logging Strategy

| Level | Usage |
|-------|-------|
| ERROR | All failures requiring immediate attention |
| WARNING | Potential issues that may require investigation |
| INFO | Significant state changes and operational events |
| DEBUG | Detailed tracking (disabled in production) |

## Certification Requirements

A processor is considered production-ready when it:

1. ✅ Correctly implements all standard obsreport metrics
2. ✅ Has at least 2-3 custom metrics for processor-specific KPIs
3. ✅ Includes dashboard panels in the shared Grafana dashboard
4. ✅ Works with the local Docker Compose stack
5. ✅ Has comprehensive documentation of its metrics