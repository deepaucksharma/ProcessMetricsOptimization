# NRDOT Processor Self-Observability

> **Guiding Philosophy:** Instrument once, visualize everywhere

All NRDOT custom processors must implement comprehensive self-observability to monitor performance, capacity, and behavior across multiple surfaces.

## Required Metrics

### Standard OTel Metrics (via obsreport)

| Metric | Type | Description |
|--------|------|-------------|
| `otelcol_processor_<name>_processed_metric_points` | Counter | Number of metric points processed |
| `otelcol_processor_<name>_dropped_metric_points` | Counter | Number of metric points dropped |
| `otelcol_processor_<name>_refused_metric_points` | Counter | Number of metric points refused |
| `otelcol_processor_<name>_latency` | Histogram | Processing time |

### Custom Processor KPIs

- All custom metrics **MUST** use the prefix `nrdot_<processor_name>_`
- Implement using `go.opentelemetry.io/otel/metric` via `component.TelemetrySettings`

**Examples:**
- `nrdot_prioritytagger_critical_processes_tagged_total` (counter)
- `nrdot_adaptivetopk_current_k` (gauge)
- `nrdot_othersrollup_aggregated_series_count_total` (counter)
- `nrdot_reservoirsampler_fill_ratio` (gauge)

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
# Build and start the local development stack
make docker-build && make compose-up

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