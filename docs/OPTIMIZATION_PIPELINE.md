# NRDOT Optimization Pipeline

This document provides comprehensive documentation for the NRDOT Process-Metrics Optimization Pipeline, including its architecture, configuration options, performance characteristics, and operational guidance.

## Overview

The NRDOT optimization pipeline is a multi-layered approach to reducing process metrics cardinality in OpenTelemetry data, designed to achieve >90% reduction in data volume while preserving visibility into critical system processes.

The pipeline consists of four custom processors that work together in sequence:

1. **L0: PriorityTagger** - Identifies and tags critical processes that must be preserved regardless of optimization
2. **L1: AdaptiveTopK** - Selects metrics from the K most resource-intensive processes, with dynamic K adjustment
3. **L3: ReservoirSampler** - Statistically samples a representative subset of non-critical, non-top processes
4. **L2: OthersRollup** - Aggregates all remaining processes into a single "_other_" process series

![Pipeline Architecture](../docs/PIPELINE_DIAGRAM.md)

## Pipeline Architecture

### Layer 0: PriorityTagger Processor

The PriorityTagger serves as the foundation of the optimization pipeline by identifying and tagging "critical" processes that must be preserved regardless of resource usage.

**Key Features:**
- Tags processes as critical based on exact name matches
- Supports regex pattern matching for flexible process identification
- Automatically identifies resource-intensive processes using configurable CPU/memory thresholds
- Preserves all metrics for critical processes regardless of downstream optimization

**Configuration Example:**
```yaml
prioritytagger:
  critical_executables:
    - systemd
    - kubelet
    - containerd
  critical_executable_patterns:
    - ".*java.*"
    - "kube.*"
  cpu_steady_state_threshold: 0.25
  memory_rss_threshold_mib: 400
  priority_attribute_name: "nr.priority"
  critical_attribute_value: "critical"
```

### Layer 1: AdaptiveTopK Processor

The AdaptiveTopK processor ensures visibility into the most active processes by selecting metrics from the top K resource-consuming processes.

**Key Features:**
- Dynamically adjusts K based on system load to capture more processes during high load
- Uses efficient min-heap algorithm to identify highest-resource processes
- Can operate in fixed K mode with a static configuration
- Preserves all critical processes tagged by the PriorityTagger regardless of resource usage

**Configuration Example:**
```yaml
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

  # Process selection configuration
  key_metric_name: "process.cpu.utilization"
  secondary_key_metric_name: "process.memory.rss"
  priority_attribute_name: "nr.priority"
  critical_attribute_value: "critical"
```

### Layer 3: ReservoirSampler Processor

The ReservoirSampler processor provides statistical visibility into the "long tail" of processes by sampling a representative subset.

**Key Features:**
- Implements Algorithm R variant for reservoir sampling
- Maintains a fixed-size reservoir of unique process identities
- Adds metadata for proper sample rate scaling during analysis
- Provides uniform probability representation of the process population

**Configuration Example:**
```yaml
reservoirsampler:
  reservoir_size: 125
  identity_attributes:
    - "process.pid"
    - "process.executable.name"
    - "process.command_line"
  sampled_attribute_name: "nr.process_sampled_by_reservoir"
  sampled_attribute_value: "true"
  sample_rate_attribute_name: "nr.sample_rate"
  priority_attribute_name: "nr.priority"
  critical_attribute_value: "critical"
```

### Layer 2: OthersRollup Processor

The OthersRollup processor aggregates all remaining non-critical, non-TopK, non-sampled processes into a single "_other_" series.

**Key Features:**
- Produces a single series for all remaining processes
- Supports different aggregation strategies (sum, average) per metric type
- Preserves metric type semantics (gauge vs. sum)
- Handles counter metrics appropriately to avoid resets

**Configuration Example:**
```yaml
othersrollup:
  output_pid_attribute_value: "-1"
  output_executable_name_attribute_value: "_other_"
  aggregations:
    "process.cpu.utilization": "avg"
    "process.memory.rss": "sum"
    "process.memory.virtual": "sum"
    "process.disk.io_read_bytes": "sum"
    "process.disk.io_write_bytes": "sum"
  priority_attribute_name: "nr.priority"
  critical_attribute_value: "critical"
  rollup_source_attribute_name: "nr.rollup_source"
  rollup_source_attribute_value: "othersrollup"
```

## Pipeline Configuration

The complete pipeline configuration is available in `config/opt-plus.yaml`. This section outlines key configuration decisions and considerations for tuning the pipeline.

### Receiver Configuration

The pipeline processes metrics from the OpenTelemetry hostmetrics receiver. Key considerations:

- Collection interval: Default is 60s, but can be adjusted based on needs
- Process metrics: Consider disabling metrics that aren't essential to reduce further cardinality
- Resource filters: Can be added at the receiver level to pre-filter processes

### Processor Sequence

The processor sequence is critical for proper pipeline function:

1. `memory_limiter` - Standard memory protection
2. `prioritytagger` - Must be first in the optimization chain
3. `adaptivetopk` - Selects top processes after critical ones are tagged
4. `reservoirsampler` - Samples from remaining processes
5. `othersrollup` - Aggregates all remaining processes
6. `batch` - Standard batching before export

### Dynamic K Configuration

The AdaptiveTopK processor's dynamic K feature adjusts based on system load:

- **Host Load Metric**: Use `system.cpu.utilization` for CPU-based adjustment
- **Load Bands**: Define thresholds and corresponding K values
- **Hysteresis**: Prevents thrashing between K values
- **Min/Max K**: Safety bounds for the dynamic adjustment

### Critical Process Selection

Critical process selection happens at the PriorityTagger level:

- **Name Matches**: Exact matches for essential system processes
- **Pattern Matches**: Regex patterns for classes of applications
- **Resource Thresholds**: Auto-detection of resource-intensive processes

## Performance Characteristics

The optimization pipeline is designed to achieve significant cardinality reduction while maintaining acceptable performance characteristics.

### Cardinality Reduction

In testing, the pipeline achieves:
- **90-95% reduction** in process metric cardinality
- Preservation of critical process visibility
- Effective representation of overall system process activity

### Memory Usage

Memory usage is influenced by:
- AdaptiveTopK heap size (proportional to process count)
- ReservoirSampler reservoir size (fixed, configurable)
- OthersRollup aggregation buffers

Typical memory overhead is minimal and scales well even with high process counts.

### CPU Usage

CPU usage is primarily impacted by:
- Total number of input processes
- Complexity of regex patterns in PriorityTagger
- Frequency of K value changes in AdaptiveTopK
- Total metric throughput

The pipeline is designed to minimize CPU overhead with efficient algorithms.

## Monitoring and Observability

The pipeline includes comprehensive observability features:

### Grafana Dashboards

The `dashboards/` directory contains preconfigured Grafana dashboards:

- `grafana-nrdot-optimization-pipeline.json` - Complete pipeline overview
- `grafana-nrdot-prioritytagger-processor.json` - Detailed PriorityTagger metrics
- `grafana-nrdot-system-overview.json` - Host system overview

### Prometheus Metrics

Each processor exports detailed metrics about its operation:

-**PriorityTagger Metrics:**
- `otelcol_otelcol_prioritytagger_critical_processes_tagged_total`
- `otelcol_processor_prioritytagger_processed_metric_points`

-**AdaptiveTopK Metrics:**
- `otelcol_otelcol_adaptivetopk_topk_processes_selected_total`
- `otelcol_otelcol_adaptivetopk_current_k_value`
- `otelcol_processor_adaptivetopk_processed_metric_points`
- `otelcol_processor_adaptivetopk_dropped_metric_points`

-**ReservoirSampler Metrics:**
- `otelcol_otelcol_reservoirsampler_reservoir_fill_ratio`
- `otelcol_otelcol_reservoirsampler_eligible_identities_seen_total`
- `otelcol_otelcol_reservoirsampler_new_identities_added_to_reservoir_total`
- `otelcol_otelcol_reservoirsampler_selected_identities_count`
- `otelcol_processor_reservoirsampler_processed_metric_points`
- `otelcol_processor_reservoirsampler_dropped_metric_points`

-**OthersRollup Metrics:**
- `otelcol_otelcol_othersrollup_aggregated_series_count_total`
- `otelcol_otelcol_othersrollup_input_series_rolled_up_total`
- `otelcol_processor_othersrollup_processed_metric_points`
- `otelcol_processor_othersrollup_dropped_metric_points`

## Troubleshooting

### Common Issues

**Cardinality Reduction Below Target:**
- Check PriorityTagger configuration for overly broad patterns
- Verify Dynamic K bands are appropriate for your environment
- Consider reducing ReservoirSampler size

**Missing Critical Processes:**
- Verify process names match exactly in critical_executables
- Check regex patterns for correctness
- Consider system-specific naming variations

**High Processor Latency:**
- Review metric collection interval
- Consider adjusted batch processor settings
- Check for regex pattern complexity issues

### Diagnostic Approaches

1. **Enable Debug Logging:**
   ```yaml
   logging:
     verbosity: detailed
   ```

2. **Monitor Processor Metrics:**
   Check processor-specific metrics to identify bottlenecks

3. **Review Pipeline Status:**
   Use zPages extension at `http://localhost:15679` to view pipeline status

## Operational Best Practices

### Deployment Recommendations

- Run the collector close to the metrics source
- Size the collector resources based on process count and metric volume
- Consider using separate collectors for metrics and traces

### Testing Changes

Always test configuration changes:
- Use the test script: `test/test_opt_plus_pipeline.sh`
- Verify cardinality reduction meets goals
- Check critical process preservation
- Monitor performance metrics

### Scaling Considerations

The pipeline scaling is primarily influenced by:
- Total number of processes being monitored
- Number of metrics per process
- Collection frequency

For high-scale environments, consider:
- Multiple collectors with load balancing
- Reducing collection frequency
- Pre-filtering metric sets

## Conclusion

The NRDOT Process-Metrics Optimization Pipeline provides a comprehensive solution to the challenge of high-cardinality process metrics. By using a layered approach with multiple specialized processors, it achieves significant data volume reduction while preserving visibility into critical system processes.

The pipeline demonstrates how custom OpenTelemetry components can be combined to create sophisticated data optimization solutions that maintain the value of telemetry data while controlling costs.