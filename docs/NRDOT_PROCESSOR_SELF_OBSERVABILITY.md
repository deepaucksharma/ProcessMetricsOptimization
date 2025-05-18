# NRDOT Processor Self-Observability & Validation Strategy

## Guiding Philosophy

**Instrument once, visualize everywhere** - All NRDOT custom processors must implement a comprehensive self-observability framework that allows their performance, capacity, and behavior to be monitored across multiple surfaces.

## Core Self-Observability Surfaces

1. **obsreport Standard Metrics** - Every processor MUST use obsreport to capture standard OTel metrics:
   - `otelcol_processor_<name>_processed_metric_points` - Number of metric points processed
   - `otelcol_processor_<name>_dropped_metric_points` - Number of metric points dropped
   - `otelcol_processor_<name>_refused_metric_points` - Number of metric points refused due to filters
   - `otelcol_processor_<name>_latency` - Processing time histogram

2. **Custom Processor KPIs** - Each processor MUST implement custom metrics relevant to its function:
   - Use `go.opentelemetry.io/otel/metric` via component.TelemetrySettings
   - All custom metrics should use the prefix `nrdot_<processor_name>_`
   - Examples:
     - `nrdot_prioritytagger_critical_processes_tagged_total` (counter)
     - `nrdot_adaptivetopk_current_k` (gauge)
     - `nrdot_othersrollup_aggregated_series_count_total` (counter)
     - `nrdot_reservoirsampler_fill_ratio` (gauge)

3. **zPages Debugging** - All processors get zpages by default when configured in collector
   - Accessible at http://localhost:55679 when running the collector
   - Shows pipelines, metric stats, and traces
   - Critical for local debugging

4. **Prometheus Scraping** - Metrics should be exposed via Prometheus exporter
   - Configure in base.yaml
   - Used for Grafana dashboards
   - Important for local validation

5. **Grafana Dashboards** - Custom dashboard for processor KPIs and behavior
   - Combined with standard OTel dashboard (15983/18309)
   - Created in JSON format in the dashboards/ directory
   - Visualizes both standard and custom metrics

## Local Validation Loop

For effective local development and testing:

1. **Collector Configuration**
   - Use the provided base.yaml which includes:
     - hostmetrics receiver for realistic process metrics
     - prometheus exporter for metrics collection
     - zpages extension for debugging
     - mock-nr endpoint for validating output
     - pipeline with the processor under test

2. **Docker Compose Stack**
   - Start with `make compose-up`
   - Includes Prometheus and Grafana
   - Automatically loads dashboards
   - Allows rapid iteration and validation

3. **Self-Observability Validation**
   - Check zPages for basic functionality and pipeline flow
   - Verify standard metrics in Prometheus
   - Verify custom metrics in Prometheus
   - Check dashboard visualizations
   - View processor output in mock-nr logs

4. **Iteration Process**
   1. Implement/modify processor
   2. Run `make docker-build && make compose-up`
   3. Check metrics and behavior in dashboards
   4. View logs with `make logs`
   5. Refine implementation
   6. Repeat

## Production Monitoring Strategy

When deployed in production:

1. **Standard Metrics Collection**
   - All otelcol_processor_* metrics sent to monitoring backend
   - Standard dashboards for all OTel components (Dashboard ID 15983/18309)

2. **KPI Dashboards**
   - Custom dashboard showing the key metrics for each processor
   - Alerts on critical thresholds
   - Anomaly detection for unusual behavior

3. **Logging Strategy**
   - ERROR level for all failures
   - WARNING level for potential issues
   - INFO level for significant state changes
   - DEBUG level for detailed tracking (disabled in production)

## Certification Requirements

Before a processor can be considered production-ready, it MUST:

1. Implement all standard obsreport metrics correctly
2. Have at least 2-3 custom metrics for processor-specific KPIs
3. Include a Grafana dashboard in JSON format
4. Be testable with the local Docker Compose stack
5. Have comprehensive documentation of its metrics and what they mean
