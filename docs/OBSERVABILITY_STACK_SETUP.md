# NRDOT Collector - Observability Stack Setup Guide

This guide outlines the setup for a comprehensive observability stack for the entire NRDOT (New Relic Distribution of OpenTelemetry) collector, not just the custom processors. This includes Grafana for visualization, Prometheus for metrics collection, and leveraging `obsreport` for standard OpenTelemetry collector observability metrics. The focus is on end-to-end monitoring of the NRDOT collector, covering its health, performance, and the flow of telemetry data through the pipeline.

---

## Step-by-Step Setup for Grafana, Prometheus, and Obsreport with NRDOT Collector

### 1. Configure Prometheus to Scrape Metrics from NRDOT Collector
The NRDOT collector exposes metrics via its Prometheus exporter and a dedicated telemetry metrics endpoint. Prometheus can scrape these endpoints to gather data on collector health, processor performance, and telemetry data flow.

#### Key Actions:
- **Enable Prometheus Exporter in NRDOT Collector**:
  - Ensure the NRDOT collector's configuration (e.g., `config/base.yaml`) includes the `prometheus` exporter in a pipeline.
  - Example configuration snippet:
    ```yaml
    exporters:
      prometheus:
        endpoint: "0.0.0.0:8889" # Port for pipeline metrics
        namespace: otelcol # Optional namespace
    service:
      pipelines:
        metrics:
          exporters: [prometheus, otlphttp]
      telemetry:
        metrics:
          level: detailed
          address: "0.0.0.0:8888" # Collector telemetry endpoint
    ```

- **Configure Prometheus Scrape Jobs**:
  - Update `prometheus.yml` (often in `build/docker-compose.yaml`) to scrape both endpoints:
    ```yaml
    scrape_configs:
      - job_name: 'nrdot-collector-pipeline-metrics'
        static_configs:
          - targets: ['otel-collector:8889']
      - job_name: 'nrdot-collector-self-telemetry'
        static_configs:
          - targets: ['otel-collector:8888']
    ```

- **Define Alerting Rules**:
  - Add rules in Prometheus for critical conditions (e.g., high error rates):
    ```yaml
    groups:
    - name: nrdot-collector
      rules:
      - alert: HighProcessorDroppedPoints
        expr: rate(otelcol_processor_dropped_metric_points[5m]) > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High rate of dropped metric points in processor {{ $labels.processor }}"
    ```

---

### 2. Set Up Grafana to Use Prometheus as a Data Source
Grafana visualizes the metrics scraped by Prometheus, providing dashboards for monitoring.

#### Key Actions:
- **Add Prometheus Data Source**:
  - In Grafana, add Prometheus as a data source (e.g., URL: `http://prometheus:9090`).
  - This is typically pre-configured in `build/docker-compose.yaml`.

- **Import Dashboards**:
  - Start with existing OpenTelemetry Collector dashboards (e.g., Grafana ID 15983) and customize them for NRDOT-specific metrics.

---

### 3. Design Grafana Dashboards for End-to-End Monitoring
Dashboards should cover the entire collector pipeline. See `docs/GRAFANA_DASHBOARD_DESIGN.md` for detailed designs.

#### Key Metrics:
- **Collector Health**: CPU (`rate(process_cpu_seconds_total[1m])`), Memory (`process_resident_memory_bytes`).
- **Pipeline Throughput**: Receiver input (`otelcol_receiver_accepted_metric_points`), Exporter output (`otelcol_exporter_sent_metric_points`).
- **Processor Performance**: Processed points (`otelcol_processor_processed_metric_points`), Dropped points (`otelcol_processor_dropped_metric_points`).

---

### 4. Integrate Obsreport for Standard Metrics
`obsreport` provides standard metrics for receivers, processors, and exporters, exposed via the telemetry endpoint (`:8888`).

#### Key Actions:
- Ensure `service.telemetry.metrics.level: detailed` is set in the collector config.
- Verify Prometheus scrapes `:8888` for `otelcol_*` metrics.

---

### 5. Ensure End-to-End Visibility
Monitor the entire data flow using `obsreport` and custom metrics (e.g., `otelcol_otelcol_helloworld_mutations_total`).

#### Key Metrics:
- Receivers: `otelcol_receiver_accepted_metric_points`, `otelcol_receiver_refused_metric_points`.
- Processors: `otelcol_processor_processed_metric_points`, `otelcol_processor_dropped_metric_points`.
- Exporters: `otelcol_exporter_sent_metric_points`, `otelcol_exporter_send_failed_metric_points`.

---

## Conclusion
This setup ensures full visibility into the NRDOT collector's operation using Prometheus, Grafana, and `obsreport`. It complements dashboard designs in `docs/GRAFANA_DASHBOARD_DESIGN.md`.