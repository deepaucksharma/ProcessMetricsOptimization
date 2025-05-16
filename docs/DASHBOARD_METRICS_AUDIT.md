# Dashboard Metrics Audit

This document identifies uncertain or problematic metric references in our Grafana dashboards, updated with our actual findings from the Prometheus API.

## General Issues

1. **Metric Naming Inconsistencies**:
   - Some dashboards use `otelcol_otelcol_processor_*` pattern (double prefix)
   - Others use `otelcol_otelcol_otelcol_processor_*` pattern (triple prefix)
   - **CONFIRMED**: Actual metrics use the triple prefix pattern for processor metrics
   - **CONFIRMED**: Some generic metrics use just `otelcol_processor_*` without additional prefixes

2. **Missing Metrics**:
   - Many custom metrics referenced in dashboards don't exist in Prometheus
   - **CONFIRMED**: No metrics with `nrdot_` prefix exist at all
   - **CONFIRMED**: Many algorithm-specific metrics (e.g., fill ratios, k-values) are missing

## Available Metrics

After checking the Prometheus API, we found these metrics are actually available:

1. **Core Metrics**:
   - `otelcol_processor_accepted_metric_points` (with processor label)
   - `otelcol_processor_dropped_metric_points` (with processor label)
   - `otelcol_processor_refused_metric_points` (with processor label)

2. **Processor-Specific Metrics**:
   - `otelcol_otelcol_otelcol_processor_adaptivetopk_processed_metric_points`
   - `otelcol_otelcol_otelcol_processor_adaptivetopk_dropped_metric_points`
   - `otelcol_otelcol_otelcol_processor_prioritytagger_processed_metric_points`
   - `otelcol_otelcol_otelcol_processor_reservoirsampler_processed_metric_points`
   - `otelcol_otelcol_otelcol_adaptivetopk_topk_processes_selected_total`

3. **System Metrics**:
   - `otelcol_process_cpu_seconds`
   - `otelcol_process_memory_rss`
   - `otelcol_process_uptime`

## Dashboard-Specific Issues

### grafana-nrdot-optimization-pipeline.json

| Panel | Problematic Metric | Issue | Suggested Replacement |
|-------|-------------------|-------|----------------------|
| All Processors: Processed Metric Points Rate | `otelcol_otelcol_processor_othersrollup_processed_metric_points` | Incorrect prefix (uses double instead of triple) | `otelcol_otelcol_otelcol_processor_othersrollup_processed_metric_points` |
| OthersRollup: Original DP Throughput | `otelcol_otelcol_processor_othersrollup_processed_metric_points` | Incorrect prefix | `otelcol_otelcol_otelcol_processor_othersrollup_processed_metric_points` |
| OthersRollup: Original DP Throughput | `otelcol_otelcol_processor_othersrollup_dropped_metric_points` | Incorrect prefix | `otelcol_otelcol_otelcol_processor_othersrollup_dropped_metric_points` |
| OthersRollup: Rollup Counts | `otelcol_otelcol_othersrollup_aggregated_series_count_total` | Missing metric | Use `otelcol_processor_accepted_metric_points{processor="othersrollup"}` instead |
| OthersRollup: Rollup Counts | `otelcol_otelcol_othersrollup_input_series_rolled_up_total` | Missing metric | Use `otelcol_processor_dropped_metric_points{processor="othersrollup"}` instead |

### grafana-nrdot-reservoirsampler-algo.json

| Panel | Problematic Metric | Issue | Suggested Replacement |
|-------|-------------------|-------|----------------------|
| Reservoir Fill Ratio | `nrdot_reservoirsampler_reservoir_fill_ratio` | **CONFIRMED MISSING** | Use `otelcol_processor_accepted_metric_points{processor="reservoirsampler"} / [reservoir_size_constant]` |
| Reservoir Population & Churn | `nrdot_reservoirsampler_selected_identities_count` | **CONFIRMED MISSING** | No direct replacement available; use generic metrics |
| Reservoir Population & Churn | `nrdot_reservoirsampler_new_identities_added_to_reservoir_total` | **CONFIRMED MISSING** | No direct replacement available; use generic metrics |
| Average Effective Sample Rate | `otelcol_processor_reservoirsampler_sample_rate` | **CONFIRMED MISSING** | Calculate from accepted vs dropped metrics |
| ReservoirSampler: Input vs. Sampled vs. Dropped Rate | `otelcol_processor_reservoirsampler_processed_metric_points` | Incorrect prefix | `otelcol_otelcol_otelcol_processor_reservoirsampler_processed_metric_points` |
| ReservoirSampler: Input vs. Sampled vs. Dropped Rate | `otelcol_processor_reservoirsampler_dropped_metric_points` | Incorrect prefix | Use `otelcol_processor_dropped_metric_points{processor="reservoirsampler"}` |

### grafana-nrdot-othersrollup-algo.json

| Panel | Problematic Metric | Issue | Suggested Replacement |
|-------|-------------------|-------|----------------------|
| OthersRollup: Aggregation Stats | `nrdot_othersrollup_aggregated_series_count_total` | **CONFIRMED MISSING** | Use `otelcol_processor_accepted_metric_points{processor="othersrollup"}` |
| OthersRollup: Aggregation Stats | `nrdot_othersrollup_input_series_rolled_up_total` | **CONFIRMED MISSING** | Use `otelcol_processor_dropped_metric_points{processor="othersrollup"}` |

### grafana-nrdot-prioritytagger-processor.json

| Panel | Problematic Metric | Issue | Suggested Replacement |
|-------|-------------------|-------|----------------------|
| PriorityTagger Processed Points Rate | `otelcol_otelcol_otelcol_processor_prioritytagger_processed_metric_points` | **CONFIRMED EXISTING** | Keep existing metric name, it's correct |

### grafana-nrdot-adaptivetopk-algo.json

| Panel | Problematic Metric | Issue | Suggested Replacement |
|-------|-------------------|-------|----------------------|
| Current K Value | `nrdot_adaptivetopk_current_k_value` | **CONFIRMED MISSING** | Use `otelcol_otelcol_otelcol_adaptivetopk_topk_processes_selected_total` (closest available metric) |
| K Value vs. System Load | `nrdot_adaptivetopk_current_k_value` | **CONFIRMED MISSING** | Use `otelcol_otelcol_otelcol_adaptivetopk_topk_processes_selected_total` (closest available metric) |
| K Value vs. System Load | `system_cpu_utilization` | **CONFIRMED MISSING** | Use `otelcol_process_cpu_seconds` |

## Implemented Fixes

1. **Fixed Prefixes**:
   - Updated all dashboards to use `otelcol_otelcol_otelcol_processor_*` for processor metrics
   - Used `otelcol_processor_*{processor="<name>"}` for generic metrics

2. **Missing Metrics Workarounds**:
   - For metrics that don't exist, used available metrics as proxies
   - Used calculations where possible (e.g., reservoir fill ratio = accepted points / capacity)
   - Created a new system-metrics dashboard to monitor basic system metrics

3. **Dashboard Documentation**:
   - Updated RECOMMENDATIONS.md with suggestions for future metric improvements
   - Added comments within dashboards to explain metric approximations

## Future Improvements

1. **Add Custom Metrics**:
   - Implement the missing metrics in each processor's observability code
   - Particularly needed: `current_k_value` for AdaptiveTopK, `reservoir_fill_ratio` for ReservoirSampler

2. **Standardize Metrics Schema**:
   - Document all metrics and their meanings in METRICS_SCHEMA.md
   - Use consistent prefixes for all metrics

3. **Create Alerting Rules**:
   - Add alerts for key processor metrics (high drop rates, abnormal processing rates)
   - Set up anomaly detection for processor performance