# AdaptiveTopK Processor

The `adaptivetopk` processor selects metrics from the 'K' most resource-intensive processes, and ensures that already-prioritized processes are passed through. Its goal is to focus on the most significant processes while dropping metrics from less impactful ones.

'K' can be a fixed value or dynamically adjusted based on overall host load.

## Performance Optimizations

This processor includes several optimizations to improve performance and reduce memory usage:

1. **Efficient Host Load Metric Detection**:
   - Uses early exit algorithms to find system load metrics quickly
   - Employs different search strategies based on metrics collection size

2. **Memory-Efficient Data Structures**:
   - Pre-allocates maps and slices based on estimated process counts
   - Implements efficient min-heap operations for top-K selection
   - Uses fast path for cases with fewer processes than K

3. **Process Hysteresis Management**:
   - Implements periodic full cleanup to prevent memory leaks
   - Tracks process existence to avoid maintaining stale entries
   - Uses configurable cleanup intervals

## Configuration

### Sub-Phase 2a: Fixed K

```yaml
processors:
  adaptivetopk:
    # The fixed number of top processes to keep, in addition to critical ones.
    k_value: 10
    # The metric name used to rank processes (e.g., "process.cpu.utilization", "process.memory.rss").
    key_metric_name: "process.cpu.utilization"
    # Optional: Secondary metric name for tie-breaking.
    secondary_key_metric_name: "process.memory.rss"
    # Attribute name used to identify processes already tagged as critical.
    priority_attribute_name: "nr.priority"
    # Attribute value indicating a critical process.
    critical_attribute_value: "critical"
```

### Sub-Phase 2b: Dynamic K & Hysteresis (Future)

```yaml
processors:
  adaptivetopk:
    # Metric name representing overall host load (e.g., "system.cpu.utilization").
    host_load_metric_name: "system.cpu.utilization"
    # Map of host load thresholds to K values.
    # Example: {0.2: 5, 0.5: 10, 0.8: 20} (load: K)
    load_bands_to_k_map:
      0.2: 5
      0.5: 10
      0.8: 20
    # Duration for which a process, once selected for TopK, remains even if it drops below threshold.
    hysteresis_duration: "1m"
    # Minimum and maximum bounds for dynamically adjusted K.
    min_k_value: 5
    max_k_value: 50
    key_metric_name: "process.cpu.utilization"
    secondary_key_metric_name: "process.memory.rss"
    priority_attribute_name: "nr.priority"
    critical_attribute_value: "critical"
```

## How It Works

1. **Pass-Through Critical Processes**: Metrics from processes already tagged (e.g., by prioritytagger with nr.priority="critical") are always passed to the next consumer.

2. **Identify Top K**: From the remaining (non-critical) processes, it identifies the top 'K' processes based on the key_metric_name.
   - If k_value is configured, 'K' is fixed.
   - (Future) If host_load_metric_name and load_bands_to_k_map are configured, 'K' is determined dynamically.
   - (Future) Hysteresis will prevent rapid flapping of processes in and out of the Top K set.

3. **Forward Metrics**: Metrics belonging to critical processes and the selected Top K processes are forwarded.

4. **Drop Others**: Metrics from all other non-critical, non-TopK processes are dropped.

## Metrics

| Metric Name | Type | Description |
|-------------|------|-------------|
| otelcol_processor_adaptivetopk_processed_metric_points | Counter | Total number of metric data points processed. |
| otelcol_processor_adaptivetopk_dropped_metric_points | Counter | Total number of metric data points dropped. |
| otelcol_otelcol_adaptivetopk_topk_processes_selected_total | Counter | Total number of non-critical processes selected for Top K in each batch. |
| otelcol_otelcol_adaptivetopk_current_k_value (for Dynamic K) | Gauge | The current value of K being used for selection. |
