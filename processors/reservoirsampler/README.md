# ReservoirSampler Processor

The `reservoirsampler` processor selects a statistically representative sample of metrics from "long-tail" processes. These are typically processes that are not tagged as critical by `prioritytagger` and not selected by `adaptivetopk`.

The goal is to provide insights into a subset of these less significant processes without incurring the cost of exporting all of them. This processor often works in conjunction with `othersrollup`, where this sampler might pick representatives *before* the rest are rolled up.

## Configuration

```yaml
processors:
  reservoirsampler:
    # The number of unique process identities to keep in the reservoir.
    reservoir_size: 100
    # List of attribute keys that together define a unique process identity for sampling.
    # Common choices: ["process.executable.name", "process.pid"] or just ["process.pid"]
    # or ["process.executable.name", "process.command_line"] for more specificity.
    identity_attributes:
      - "process.pid"
    # Attribute name to add to metrics from sampled processes.
    sampled_attribute_name: "nr.process_sampled_by_reservoir"
    # Attribute value for the sampled attribute.
    sampled_attribute_value: "true"
    # Attribute name for storing the calculated sample rate.
    # Sample rate = reservoir_size / total_eligible_identities_seen
    sample_rate_attribute_name: "nr.sample_rate"
    # Attribute name used to identify processes already tagged as critical (should not be sampled or counted as eligible).
    priority_attribute_name: "nr.priority"
    # Attribute value indicating a critical process.
    critical_attribute_value: "critical"
    # Optional: Attribute name that might be added by adaptivetopk.
    # Processes with this attribute should also not be sampled or counted as eligible.
    # topk_attribute_name: "nr.topk_selected"
```

## How It Works

1. **Identify Eligible Metrics**: The processor considers metric data points that do NOT belong to:
   - Critical processes (as tagged by priority_attribute_name).
   - Top K processes (if topk_attribute_name is configured and present).

2. **Extract Identity**: For each eligible metric data point, it constructs a unique process identity string based on the values of the identity_attributes.

3. **Reservoir Sampling (Algorithm R variant)**:
   - It maintains a reservoir of reservoir_size unique process identities.
   - When a new eligible process identity is encountered:
     - If the reservoir is not full, the identity is added.
     - If the reservoir is full, the new identity replaces an existing one with a probability that ensures each seen identity has an equal chance of being in the reservoir.

4. **Tag Sampled Metrics**: Metric data points belonging to process identities currently in the reservoir are:
   - Tagged with sampled_attribute_name set to sampled_attribute_value.
   - Tagged with sample_rate_attribute_name with the calculated sampling rate (e.g., if reservoir size is 100 and 1000 unique eligible PIDs have been seen, sample rate is 0.1).

5. **Pass Through / Drop**:
   - Metrics from critical/TopK processes are passed through untouched.
   - Metrics from sampled processes (now tagged) are passed through.
   - Metrics from eligible processes not in the reservoir are dropped by this processor.

## Metrics

| Metric Name | Type | Description |
|-------------|------|-------------|
| otelcol_processor_reservoirsampler_processed_metric_points | Counter | Total number of original metric data points processed. |
| otelcol_processor_reservoirsampler_dropped_metric_points | Counter | Total number of original metric data points dropped (eligible but not sampled). |
| otelcol_otelcol_reservoirsampler_reservoir_fill_ratio | Gauge | Current fill ratio of the reservoir (0.0 to 1.0). |
| otelcol_otelcol_reservoirsampler_selected_identities_count | Gauge | Current number of unique process identities in the reservoir. |
| otelcol_otelcol_reservoirsampler_eligible_identities_seen_total | Counter | Total number of unique eligible process identities encountered since collector start. |
| otelcol_otelcol_reservoirsampler_new_identities_added_to_reservoir_total | Counter | Total count of new identities added to the reservoir (either filling or replacing). |
