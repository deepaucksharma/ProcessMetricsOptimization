# NRDOT Metric Naming Conventions

This document outlines the standardized naming conventions for metrics in the NRDOT Process-Metrics Optimization project. Following these conventions ensures consistency across processors and facilitates easier monitoring and debugging.

## General Metric Naming Pattern

All metrics in the NRDOT custom processors follow this pattern:

```
otelcol_otelcol_otelcol_[component_type]_[component_name]_[metric_name]
```

Where:
- `otelcol_otelcol_otelcol_` - Standard triple prefix for all OpenTelemetry Collector processor metrics
- `[component_type]` - Type of component (e.g., `processor`)
- `[component_name]` - Name of the specific processor (e.g., `prioritytagger`, `adaptivetopk`)
- `[metric_name]` - Specific metric name describing what is being measured

## Standard Metrics

Every processor implements these standard metrics:

| Metric Name | Type | Description |
|-------------|------|-------------|
| `otelcol_otelcol_otelcol_processor_[name]_processed_metric_points` | Counter | Number of metric points processed by the processor |
| `otelcol_processor_dropped_metric_points` | Counter | Number of metric points dropped by the processor |

## Processor-Specific Metrics

### PriorityTagger Processor

| Metric Name | Type | Description |
|-------------|------|-------------|
| `otelcol_nrdot_prioritytagger_critical_processes_tagged_total` | Counter | Total number of processes tagged as critical |

### AdaptiveTopK Processor

| Metric Name | Type | Description |
|-------------|------|-------------|
| `otelcol_nrdot_adaptivetopk_topk_processes_selected_total` | Counter | Total number of non-critical processes selected for Top K |
| `otelcol_nrdot_adaptivetopk_current_k_value` | Gauge | Current value of K being used for process selection |

### OthersRollup Processor

| Metric Name | Type | Description |
|-------------|------|-------------|
| `otelcol_nrdot_othersrollup_aggregated_series_count_total` | Counter | Number of new "_other_" series generated |
| `otelcol_nrdot_othersrollup_input_series_rolled_up_total` | Counter | Number of input series aggregated into "_other_" series |

### ReservoirSampler Processor

| Metric Name | Type | Description |
|-------------|------|-------------|
| `otelcol_nrdot_reservoirsampler_reservoir_fill_ratio` | Gauge | Current fill ratio of the reservoir (0.0 to 1.0) |
| `otelcol_nrdot_reservoirsampler_selected_identities_count` | Gauge | Current number of unique process identities in the reservoir |
| `otelcol_nrdot_reservoirsampler_eligible_identities_seen_total` | Counter | Total number of unique eligible process identities encountered |
| `otelcol_nrdot_reservoirsampler_new_identities_added_to_reservoir_total` | Counter | Total count of new identities added/replaced in the reservoir |

## Metric Types

- **Counter**: Cumulative metric that only increases over time (e.g., number of processed items)
- **Gauge**: Point-in-time measurement that can go up or down (e.g., current memory usage)
- **Histogram**: Distribution of values (e.g., request duration)

## Best Practices

1. Always use the full prefix `otelcol_otelcol_otelcol_` for standard processor metrics
2. Always use the full prefix `otelcol_nrdot_` for custom processor metrics
3. Use `_total` suffix for counter metrics that track cumulative values
4. Use descriptive metric names that clearly indicate what is being measured
5. Include the processor name in the metric for clear identification
6. Use standard metrics across all processors for consistent monitoring
7. Document all metrics in processor READMEs and in this central document