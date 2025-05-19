# OthersRollup Processor

The `othersrollup` processor aggregates metrics from non-priority, non-TopK processes into a single "_other_" summary series. This significantly reduces cardinality while retaining a sense of overall system load from less significant processes.

It should typically be placed after `prioritytagger` and `adaptivetopk` processors in the pipeline.

## Configuration

```yaml
processors:
  othersrollup:
    # Attribute value to use for the process ID of the rolled-up series.
    output_pid_attribute_value: "-1" # or "_other_pid_"
    # Attribute value for the executable name of the rolled-up series.
    output_executable_name_attribute_value: "_other_"
    # Map of metric names to aggregation functions (e.g., "sum", "avg").
    # If a metric is not listed, its aggregation behavior might be default (e.g., sum for cumulative, avg for gauge) or skipped.
    aggregations:
      "process.cpu.utilization": "avg"
      "process.memory.rss": "sum"
      "process.disk.io_read_bytes": "sum"
    # List of specific metric names to apply rollup to. If empty, applies to all compatible metrics not belonging to priority/TopK.
    metrics_to_rollup:
      - "process.cpu.utilization"
      - "process.memory.rss"
    # Attribute name used to identify processes already tagged as critical (should not be rolled up).
    priority_attribute_name: "nr.priority"
    # Attribute value indicating a critical process.
    critical_attribute_value: "critical"
    # Attribute name that might be added by adaptivetopk (should not be rolled up if present).
    # This depends on how adaptivetopk marks its selections.
    # topk_attribute_name: "nr.topk_selected"
```

## How It Works

1. **Identify "Other" Metrics**: The processor identifies metric data points that do NOT belong to:
   - Critical processes (as tagged by priority_attribute_name).
   - Top K processes (this requires coordination with adaptivetopk, potentially by checking for an absence of a special TopK tag or by ensuring adaptivetopk has already filtered them out).

2. **Aggregate**: For the identified "other" metrics:
   - Values are aggregated based on the aggregations map (e.g., sum, average).
   - Aggregation happens per resource and per metric name.

3. **Create Rolled-up Series**: New metric data points are created for these aggregated values.
   - Identifying attributes like PID and executable name are replaced with output_pid_attribute_value and output_executable_name_attribute_value.
   - Other relevant attributes (like hostname) are preserved.

4. **Drop Originals**: The original metric data points that were rolled up are dropped.

5. **Pass Through**: Critical and TopK metrics (which should have been handled or passed through by previous processors) are passed through untouched by this processor.

## Metrics

| Metric Name | Type | Description |
|-------------|------|-------------|
| otelcol_otelcol_processor_othersrollup_processed_metric_points | Counter | Total number of original metric data points processed. |
| otelcol_otelcol_processor_othersrollup_dropped_metric_points | Counter | Total number of original metric data points dropped (after rollup). |
| otelcol_otelcol_othersrollup_aggregated_series_count_total | Counter | Number of new "other" series generated per batch. |
| otelcol_otelcol_othersrollup_input_series_rolled_up_total | Counter | Number of input series that were aggregated into an "other" series. |
