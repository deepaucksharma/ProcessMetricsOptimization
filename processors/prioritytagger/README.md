# PriorityTagger Processor

The `prioritytagger` processor identifies and tags metrics from critical processes to ensure their preservation throughout subsequent pipeline stages. It examines process metrics and marks certain processes as critical based on various criteria including process name, regex patterns, CPU utilization, and memory consumption.

## Configuration

```yaml
processors:
  prioritytagger:
    # List of process executable names that are considered critical and will be tagged
    critical_executables: [kubelet, systemd, docker, containerd]
    
    # List of regex patterns for matching process executable names to be tagged as critical
    critical_executable_patterns: [kube.*, docker.*, containerd.*]
    
    # Optional: CPU utilization threshold above which a process is marked critical
    # Set to a negative value to disable this check (default: -1)
    cpu_steady_state_threshold: 0.8
    
    # Optional: Memory RSS threshold in MiB above which a process is marked critical
    # Set to a negative value to disable this check (default: -1)
    memory_rss_threshold_mib: 1024
    
    # Name of the attribute that will be added to tag critical processes
    # Default: "nr.priority"
    priority_attribute_name: nr.priority
    
    # Value that will be set for the priority attribute to mark a process as critical
    # Default: "critical"
    critical_attribute_value: critical
```

## How It Works

When the `prioritytagger` processor receives metrics, it:

1. Examines each metric datapoint that includes process information
2. Checks if the process is considered critical based on one or more of these criteria:
   - The process executable name matches one of the names in `critical_executables`
   - The process executable name matches one of the regex patterns in `critical_executable_patterns`
   - The process CPU utilization exceeds `cpu_steady_state_threshold` (if enabled)
   - The process memory RSS exceeds `memory_rss_threshold_mib` (if enabled)
3. If any criteria match, adds the configured priority attribute (e.g., `nr.priority="critical"`) to the datapoint

This tagging allows subsequent processors in the pipeline (like `adaptivetopk`) to identify and maintain critical processes regardless of other filtering rules.

## Metrics

The processor exposes the following metrics:

| Metric Name | Type | Description |
|-------------|------|-------------|
| `otelcol_processor_prioritytagger_processed_metric_points` | Counter | Total number of metric data points processed |
| `otelcol_processor_prioritytagger_dropped_metric_points` | Counter | Total number of metric data points dropped due to errors |
| `nrdot_prioritytagger_critical_processes_tagged_total` | Counter | Total number of unique processes tagged as critical |

## Example

```yaml
receivers:
  hostmetrics:
    collection_interval: 30s
    scrapers:
      process:

processors:
  prioritytagger:
    critical_executables: [kubelet, systemd, containerd]
    critical_executable_patterns: [kube.*]
    cpu_steady_state_threshold: 0.5
    memory_rss_threshold_mib: 2048

exporters:
  otlp:
    endpoint: otelcol:4317

service:
  pipelines:
    metrics:
      receivers: [hostmetrics]
      processors: [prioritytagger]
      exporters: [otlp]
```

In this example, the `prioritytagger` processor will tag as critical any process that:
- Has the name `kubelet`, `systemd`, or `containerd`
- Has a name matching the pattern `kube.*`
- Has a CPU utilization greater than 50%
- Has a memory RSS greater than 2 GiB

These critical processes will be given the attribute `nr.priority="critical"`, which can be used by downstream processors like `adaptivetopk` to ensure they're always preserved regardless of other selection criteria.