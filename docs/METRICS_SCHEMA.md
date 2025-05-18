# NRDOT Process-Metrics Optimization: Metrics Schema

This document outlines the metrics schema used in the NRDOT Process-Metrics Optimization project, focusing on the metric naming conventions, available metrics, and how to effectively query them in Grafana dashboards.

## Table of Contents
- [Metric Naming Conventions](#metric-naming-conventions)
- [Available Metrics](#available-metrics)
  - [Collector Metrics](#collector-metrics)
  - [Processor Metrics](#processor-metrics)
  - [Pipeline Metrics](#pipeline-metrics)
- [Grafana Dashboard Queries](#grafana-dashboard-queries)
- [Troubleshooting](#troubleshooting)

## Metric Naming Conventions

The project uses the following metric naming patterns:

1. **Standard OpenTelemetry Collector Metrics**:
   - Format: `otelcol_[component_type]_[metric_name]`
   - Example: `otelcol_process_cpu_seconds`, `otelcol_receiver_accepted_metric_points`

2. **Triple-Prefixed Processor Metrics**:
   - Format: `otelcol_otelcol_otelcol_processor_[processor_name]_[metric_name]`
   - Example: `otelcol_otelcol_otelcol_processor_prioritytagger_processed_metric_points`

3. **Processor-Specific Metrics**:
   - Format: `otelcol_otelcol_otelcol_[processor_name]_[metric_name]`
   - Example: `otelcol_otelcol_otelcol_adaptivetopk_topk_processes_selected_total`

## Available Metrics

### Collector Metrics

| Metric Name | Description | Type |
|-------------|-------------|------|
| `otelcol_process_cpu_seconds` | CPU usage in seconds | Counter |
| `otelcol_process_memory_rss` | Memory usage (RSS) | Gauge |
| `otelcol_process_uptime` | Collector uptime in seconds | Counter |
| `otelcol_process_runtime_heap_alloc_bytes` | Heap allocation in bytes | Gauge |
| `otelcol_process_runtime_total_alloc_bytes` | Total memory allocation in bytes | Counter |
| `otelcol_process_runtime_total_sys_memory_bytes` | Total system memory in bytes | Gauge |

### Processor Metrics

#### Common Processor Metrics

| Metric Name | Description | Type |
|-------------|-------------|------|
| `otelcol_processor_accepted_metric_points` | Number of metric points accepted by processor | Counter |
| `otelcol_processor_refused_metric_points` | Number of metric points refused by processor | Counter |
| `otelcol_processor_dropped_metric_points` | Number of metric points dropped by processor | Counter |
| `otelcol_processor_batch_metadata_cardinality` | Batch metadata cardinality | Gauge |

#### PriorityTagger Processor Metrics

| Metric Name | Description | Type |
|-------------|-------------|------|
| `otelcol_otelcol_otelcol_processor_prioritytagger_processed_metric_points` | Number of metric points processed by PriorityTagger | Counter |

#### AdaptiveTopK Processor Metrics

| Metric Name | Description | Type |
|-------------|-------------|------|
| `otelcol_otelcol_otelcol_processor_adaptivetopk_processed_metric_points` | Number of metric points processed by AdaptiveTopK | Counter |
| `otelcol_otelcol_otelcol_processor_adaptivetopk_dropped_metric_points` | Number of metric points dropped by AdaptiveTopK | Counter |
| `otelcol_otelcol_otelcol_adaptivetopk_topk_processes_selected_total` | Total number of top-K processes selected | Counter |

#### ReservoirSampler Processor Metrics

| Metric Name | Description | Type |
|-------------|-------------|------|
| `otelcol_otelcol_otelcol_processor_reservoirsampler_processed_metric_points` | Number of metric points processed by ReservoirSampler | Counter |

### Pipeline Metrics

| Metric Name | Description | Type |
|-------------|-------------|------|
| `otelcol_receiver_accepted_metric_points` | Number of metric points accepted by the receiver | Counter |
| `otelcol_receiver_refused_metric_points` | Number of metric points refused by the receiver | Counter |
| `otelcol_exporter_queue_size` | Size of the exporter queue | Gauge |
| `otelcol_exporter_queue_capacity` | Capacity of the exporter queue | Gauge |

## Grafana Dashboard Queries

### Rate-based Queries
For most counter metrics, use a rate function to visualize the change over time:

```promql
rate(otelcol_otelcol_otelcol_processor_prioritytagger_processed_metric_points{job="otel-collector-metrics", service="otel-collector-internal"}[1m])
```

### Processor Comparison Queries
To compare different processors, use a query like:

```promql
sum(rate(otelcol_otelcol_otelcol_processor_prioritytagger_processed_metric_points{job="otel-collector-metrics", service="otel-collector-internal"}[1m]) or
    rate(otelcol_otelcol_otelcol_processor_adaptivetopk_processed_metric_points{job="otel-collector-metrics", service="otel-collector-internal"}[1m]) or
    rate(otelcol_otelcol_otelcol_processor_reservoirsampler_processed_metric_points{job="otel-collector-metrics", service="otel-collector-internal"}[1m])
) by (processor)
```

### Pipeline Efficiency Queries
To measure pipeline efficiency:

```promql
(1 - (
  rate(otelcol_exporter_sent_metric_points{exporter="otlphttp", job="otel-collector-metrics"}[1m]) /
  rate(otelcol_receiver_accepted_metric_points{receiver="hostmetrics", job="otel-collector-metrics"}[1m])
)) * 100
```

## Troubleshooting

### Common Issues

1. **Missing Metrics**:
   - Verify the processor is correctly registered in the collector
   - Check the metric naming pattern (triple prefix vs. double prefix)
   - Confirm that the pipeline is properly configured in `opt-plus.yaml`

2. **Incorrect Values**:
   - Ensure the processor is actually processing metrics
   - Check for dropped metrics that might indicate issues

3. **Dashboard Visualization Issues**:
   - Verify Prometheus connectivity
   - Check that the prometheus datasource UID is correct
   - Ensure query expressions match the actual metric names

### Debugging Tips

1. Use Prometheus's `/api/v1/metadata?limit=1000` endpoint to view all available metrics
2. Use queries like `curl -s "http://localhost:19090/api/v1/label/__name__/values" | grep -i processor` to find specific metrics
3. For rate-based metrics, ensure adequate time windows (`[1m]` or `[5m]`)
4. Add debug logging to processors if metrics aren't being generated

## Updating Dashboard Metric References

When metric names change or new metrics are added, ensure that all dashboards are updated to use the correct names. Pay special attention to:

1. Triple vs. double prefix pattern
2. Correct label matching in queries (`{job="otel-collector-metrics", service="otel-collector-internal"}`)
3. Appropriate rate intervals based on the metric's expected change frequency