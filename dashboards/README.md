# NRDOT Dashboards

This directory contains Grafana dashboards for monitoring the NRDOT Process-Metrics Optimization pipeline. These dashboards provide comprehensive visibility into the pipeline's performance, health, and effectiveness.

## Dashboard Overview

### Core Dashboards

1. **`grafana-nrdot-system-overview.json`**
   - System-level metrics showing collector health and overall pipeline status
   - Focuses on CPU, memory, uptime, and general service availability

2. **`grafana-nrdot-custom-processor-starter-kpis.json`**
   - Basic KPIs for all custom processors
   - Serves as a starting point for new processor development

3. **`grafana-nrdot-optimization-pipeline.json`**
   - Comprehensive dashboard showing the full optimization pipeline
   - Displays data flow through all processors with cardinality reduction metrics
   - Provides cost impact analysis and optimization percentage tracking

4. **`grafana-nrdot-prioritytagger-processor.json`**
   - PriorityTagger-specific performance metrics
   - Shows tagging rates, patterns, and performance characteristics

### Algorithm Insight Dashboards

These specialized dashboards provide deep visibility into the internal workings and decision-making processes of each processor algorithm.

1. **`grafana-nrdot-prioritytagger-algo.json`**
   - Visualizes how the PriorityTagger identifies and tags critical processes
   - Shows tagging decisions broken down by reason (exact match, pattern, CPU/memory thresholds)
   - Provides insights into the resource usage patterns of tagged processes

2. **`grafana-nrdot-adaptivetopk-algo.json`**
   - Details the dynamic K selection process and how it adapts to system load
   - Visualizes the cutoff threshold for TopK selection
   - Shows which process types are consistently selected for the TopK

3. **`grafana-nrdot-othersrollup-algo.json`**
   - Demonstrates how non-critical, non-TopK processes are aggregated
   - Shows the cardinality reduction achieved through rollup
   - Visualizes the characteristics of the aggregated "_other_" series

4. **`grafana-nrdot-reservoirsampler-algo.json`**
   - Visualizes the reservoir sampling algorithm in action
   - Shows reservoir fill state, identity churn, and sampling rates
   - Provides visibility into which process types are being sampled

## Usage Instructions

### Importing Dashboards

1. Navigate to Grafana (default: http://localhost:13000)
2. Go to Dashboards > Import
3. Upload the JSON file or paste its contents
4. Select the appropriate Prometheus data source (usually named "Prometheus")
5. Click Import

### Dashboard Features

- **Time Range Selection**: Use the time picker in the upper right to adjust the time window
- **Refresh Rate**: Set your preferred auto-refresh interval
- **Hover Details**: Hover over any graph to see detailed metrics
- **Panel Options**: Click panel titles to see options for edit, export, view legend, etc.

### Metric Requirements

For all algorithm insight dashboards to function properly, the processors need to expose specific metrics:

#### PriorityTagger
- `otelcol_otelcol_prioritytagger_critical_processes_tagged_total`
- Ideally with a `reason` label indicating tagging mechanism

#### AdaptiveTopK
- `otelcol_otelcol_adaptivetopk_current_k_value`
- `otelcol_otelcol_adaptivetopk_topk_processes_selected_total`

#### OthersRollup
- `otelcol_otelcol_othersrollup_input_series_rolled_up_total`
- `otelcol_otelcol_othersrollup_aggregated_series_count_total`

#### ReservoirSampler
- `otelcol_otelcol_reservoirsampler_reservoir_fill_ratio`
- `otelcol_otelcol_reservoirsampler_selected_identities_count`
- `otelcol_otelcol_reservoirsampler_eligible_identities_seen_total`
- `otelcol_otelcol_reservoirsampler_new_identities_added_to_reservoir_total`

## Dashboard Customization

Feel free to modify these dashboards to better fit your specific environment:

1. Adjust thresholds to match your performance expectations
2. Add organization-specific annotations or variables
3. Integrate with additional data sources if available
4. Create specialized versions for different environments (dev, staging, prod)

## Maintenance

Ensure you keep these dashboards up-to-date as processors evolve:

1. Update metric names if they change in the code
2. Add panels for new functionality as it's implemented
3. Adjust thresholds as performance characteristics change
4. Regularly review for continued relevance and effectiveness

For detailed information on dashboard design principles and best practices, see the documentation in `/docs/GRAFANA_DASHBOARD_DESIGN.md`.