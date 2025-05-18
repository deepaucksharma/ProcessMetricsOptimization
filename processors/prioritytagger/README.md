# PriorityTagger Processor

> **L0 Processor** - Tags critical processes for preservation in subsequent pipeline stages

## Overview

The `prioritytagger` processor identifies and tags metrics from critical processes based on:
- Process executable names
- Regex pattern matching
- CPU utilization thresholds
- Memory RSS thresholds

## Configuration

```yaml
processors:
  prioritytagger:
    # Critical process identification
    critical_executables: [kubelet, systemd, docker, containerd]
    critical_executable_patterns: [kube.*, docker.*, containerd.*]

    # Resource thresholds (set to negative to disable)
    cpu_steady_state_threshold: 0.8        # CPU utilization (0.0-1.0)
    memory_rss_threshold_mib: 1024         # Memory RSS in MiB

    # Tagging configuration
    priority_attribute_name: nr.priority   # Default: "nr.priority"
    critical_attribute_value: critical     # Default: "critical"
```

## Operation

1. **Input**: Process metrics from `hostmetrics` receiver
2. **Processing**:
   - Examines each metric datapoint with process information
   - Checks against all configured criteria
   - Adds the priority attribute to matching processes
3. **Output**: Original metrics with priority attributes added to critical processes

## Critical Process Selection Criteria

| Criterion | Configuration | Description |
|-----------|---------------|-------------|
| **Exact Name Match** | `critical_executables` | Process name exactly matches one in the list |
| **Pattern Match** | `critical_executable_patterns` | Process name matches regex pattern |
| **CPU Utilization** | `cpu_steady_state_threshold` | Process CPU usage exceeds threshold |
| **Memory Usage** | `memory_rss_threshold_mib` | Process memory RSS exceeds threshold in MiB |

## Usage in the Optimization Pipeline

In the full optimization pipeline (configured in opt-plus.yaml), the PriorityTagger processor is the first processor in the sequence (L0). It plays a critical role by identifying and tagging processes that should always be preserved throughout the pipeline, regardless of resource usage or sampling decisions.

```
┌───────────────┐    ┌─────────────────┐    ┌────────────────┐    ┌──────────────────┐    ┌────────────────┐
│ HOSTMETRICS   │    │ PRIORITYTAGGER  │    │ ADAPTIVETOPK   │    │ RESERVOIRSAMPLER │    │ OTHERSROLLUP   │
│ RECEIVER      │───>│ PROCESSOR (L0)  │───>│ PROCESSOR (L1) │───>│ PROCESSOR (L3)   │───>│ PROCESSOR (L2) │──> EXPORTERS
│ Process Data  │    │ Tag Critical    │    │ Keep Top K     │    │ Sample Long-Tail │    │ Aggregate Rest │
└───────────────┘    └─────────────────┘    └────────────────┘    └──────────────────┘    └────────────────┘
```

## Self-Observability

| Metric Name | Type | Description |
|-------------|------|-------------|
| `otelcol_otelcol_otelcol_processor_prioritytagger_processed_metric_points` | Counter | Total metric points processed |
| `otelcol_processor_dropped_metric_points` | Counter | Metric points dropped due to errors |
| `otelcol_otelcol_prioritytagger_critical_processes_tagged_total` | Counter | Unique processes tagged as critical |

## Pipeline Example

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

This configuration tags processes as critical if any of these apply:
- Name is `kubelet`, `systemd`, or `containerd`
- Name matches pattern `kube.*`
- CPU utilization > 50%
- Memory RSS > 2 GiB

Tagged processes receive the attribute `nr.priority="critical"`, which downstream processors like `adaptivetopk` can use to preserve these metrics.