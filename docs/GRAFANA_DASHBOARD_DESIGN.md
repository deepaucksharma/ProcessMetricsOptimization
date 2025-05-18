# NRDOT Process-Metrics Optimization - Grafana Dashboard Design

**Prerequisite:** This guide assumes the observability stack (Prometheus, Grafana, NRDOT collector exporters) is configured as described in [OBSERVABILITY_STACK_SETUP.md](OBSERVABILITY_STACK_SETUP.md). The dashboards described here rely on metrics exposed and collected through that setup.

To create **exhaustive and much better Grafana dashboards** for monitoring a New Relic Distribution of OpenTelemetry (NRDOT) pipeline, we need to focus on telling a clear story about the pipeline's performance, efficiency, and health. Below is a structured approach that outlines guiding principles, a proposed dashboard structure, and detailed panel designs to achieve this goal.

---

## Guiding Principles for Effective NRDOT Dashboards

1. **Pipeline-Focused Visualization**: Dashboards should visualize data flow through the entire pipeline, from receiver to exporter, highlighting how each processor transforms the data.

2. **Hierarchical Context**: Organize panels in a logical progression that moves from high-level system health to detailed processor-specific metrics.

3. **Optimization Impact**: Clearly demonstrate the impact of each optimization processor on data volume, emphasizing cost reduction metrics.

4. **Comparative Analysis**: Include before/after views to highlight the effectiveness of optimization strategies.

5. **Alertable Thresholds**: Incorporate visual indicators for acceptable ranges and warning thresholds.

6. **Algorithm Insights**: Provide detailed visibility into the internal workings and decision-making processes of each processor algorithm.

---

## Recommended Dashboard Structure

### 1. System Overview Dashboard

This top-level dashboard provides a holistic view of the entire NRDOT collector system.

#### Panels:

1. **Collector Health Status**
   - CPU Usage: `rate(process_cpu_seconds_total{job="nrdot-collector-self-telemetry"}[5m])`
   - Memory Usage: `process_resident_memory_bytes{job="nrdot-collector-self-telemetry"}`
   - Uptime: `process_uptime_seconds{job="nrdot-collector-self-telemetry"}`
   - Component Status: Health indicators for each processor, receiver, and exporter

2. **Pipeline Throughput Overview**
   - Input Rate: `rate(otelcol_receiver_accepted_metric_points{receiver="hostmetrics"}[1m])`
   - Output Rate: `rate(otelcol_exporter_sent_metric_points{exporter="otlphttp"}[1m])`
   - Optimization Ratio: `(rate(otelcol_receiver_accepted_metric_points{receiver="hostmetrics"}[1m]) - rate(otelcol_exporter_sent_metric_points{exporter="otlphttp"}[1m])) / rate(otelcol_receiver_accepted_metric_points{receiver="hostmetrics"}[1m]) * 100`

3. **Optimization Impact**
   - Data Volume Before Optimization
   - Data Volume After Optimization
   - Overall Reduction Percentage

4. **Alert Status**
   - Critical and Warning alerts related to collector performance
   - Pipeline health indicators

---

### 2. Process Metrics Collection Dashboard

This dashboard focuses on the collection of process metrics and their initial processing.

#### Panels:

1. **Host Metrics Collection**
   - Collection Rate: `rate(otelcol_receiver_accepted_metric_points{receiver="hostmetrics",transport=""}[1m])`
   - Collection Errors: `rate(otelcol_receiver_refused_metric_points{receiver="hostmetrics"}[1m])`
   - CPU Collection Time: `histogram_quantile(0.95, sum(rate(otelcol_receiver_collect_latency_bucket{receiver="hostmetrics",transport=""}[5m])) by (le))`

2. **Process Metrics Breakdown**
   - Process Counts by Host
   - Process Metrics Distribution by Type
   - Top Process Metrics by Volume

3. **PriorityTagger Effectiveness**
   - Critical Process Detection Rate: `rate(otelcol_otelcol_prioritytagger_critical_processes_tagged_total[1m])`
   - Critical Process Percentage: Custom ratio calculation
   - Process Tagging Performance: `histogram_quantile(0.95, sum(rate(otelcol_processor_prioritytagger_latency_bucket[5m])) by (le))`

---

### 3. Optimization Pipeline Dashboard

This dashboard visualizes the progressive optimization of metrics through the L0-L3 pipeline layers.

#### Panels:

1. **L0: PriorityTagger Performance**
   - Tagged vs. Untagged Processes
   - Processing Rate: `rate(otelcol_processor_prioritytagger_processed_metric_points[1m])`
   - Tag Distribution by Criteria (name, regex, CPU, memory)

2. **L1: AdaptiveTopK Metrics**
   - Current K Value: `otelcol_otelcol_adaptivetopk_current_k_value`
   - TopK Selection Changes Over Time
   - Dropped Points: `rate(otelcol_processor_adaptivetopk_dropped_metric_points[1m])`
   - CPU Threshold Adjustments

3. **L2: OthersRollup Impact**
   - Aggregation Rate: `rate(otelcol_otelcol_othersrollup_aggregated_series_count_total[1m])`
   - Cardinality Reduction: Before/After Series Count
   - "Others" Category Statistics

4. **L3: ReservoirSampler Statistics**
   - Sampling Rate: Dynamic calculation
   - Reservoir Fill Status: `otelcol_otelcol_reservoirsampler_reservoir_fill_ratio`
   - Statistical Representation Accuracy
   - Long-tail Coverage

5. **Progressive Data Reduction**
   - Waterfall Chart: Data volume reduction at each pipeline stage
   - Cumulative Optimization Impact

---

### 4. Algorithm-Specific Dashboards

In addition to the high-level processor-specific dashboards, we've created specialized algorithm insight dashboards that provide deep visibility into the decision-making processes and internal workings of each processor.

#### 4.1 PriorityTagger Algorithm Dashboard (`grafana-nrdot-prioritytagger-algo.json`)

This dashboard visualizes how the PriorityTagger identifies and tags critical processes through different mechanisms.

1. **Tagging Decision Breakdown**
   - Critical process tagging rates by reason (exact match, pattern match, CPU threshold, memory threshold)
   - Distribution of tagging reasons as a pie chart

2. **Threshold-Based Tagging Insights**
   - CPU utilization of processes tagged as critical
   - Memory usage of processes tagged as critical
   - Comparison of tagged vs. non-tagged process resource usage

3. **Top Tagged Executables**
   - Table view of the most frequently tagged executables
   - Analysis of tagging patterns over time

#### 4.2 AdaptiveTopK Algorithm Dashboard (`grafana-nrdot-adaptivetopk-algo.json`)

This dashboard provides insights into the dynamic K selection process and the ranking of processes.

1. **K Value Dynamics**
   - Real-time visualization of the current K value
   - Correlation between system load and K value changes
   - Impact of hysteresis on K value stability

2. **TopK Selection Metrics**
   - Visualization of input vs. dropped metric rates
   - Distribution of key metric values used for ranking
   - Approximate cutoff value for the TopK selection

3. **TopK Process Selection**
   - Analysis of which processes are consistently in the TopK
   - Visual representation of the TopK cutoff threshold

#### 4.3 OthersRollup Algorithm Dashboard (`grafana-nrdot-othersrollup-algo.json`)

This dashboard shows how the OthersRollup aggregates non-critical, non-TopK processes.

1. **Rollup Performance**
   - Input series vs. aggregated "_other_" series rate
   - Cardinality reduction factor achieved by rollup

2. **Aggregated Metrics Analysis**
   - Visualization of the "_other_" process metrics (CPU, memory, etc.)
   - Patterns in aggregated data over time

3. **Aggregation Configuration**
   - Metrics being rolled up and their aggregation types
   - Impact of different aggregation strategies

#### 4.4 ReservoirSampler Algorithm Dashboard (`grafana-nrdot-reservoirsampler-algo.json`)

This dashboard visualizes the reservoir sampling algorithm's behavior and effectiveness.

1. **Reservoir State**
   - Reservoir fill ratio over time
   - Number of unique identities in the reservoir
   - Rate of new identities being added (churn)

2. **Sampling Impact**
   - Input vs. sampled vs. dropped metric rates
   - Visualization of sampling probability changes
   - Comparison of eligible vs. sampled process count

3. **Sample Distribution**
   - Analysis of which processes are most frequently sampled
   - Statistical representation quality metrics

---

## Panel Design Best Practices

### Processor Optimization Impact Panels

For each processor, create a "before/after" visualization:

1. **Before-After Time Series**
   ```
   // Before processor
   rate(otelcol_processor_<n>_received_metric_points[1m])

   // After processor
   rate(otelcol_processor_<n>_processed_metric_points[1m])
   ```

2. **Reduction Percentage Gauge**
   ```
   (
     sum(rate(otelcol_processor_<n>_received_metric_points[5m])) -
     sum(rate(otelcol_processor_<n>_processed_metric_points[5m]))
   ) /
   sum(rate(otelcol_processor_<n>_received_metric_points[5m])) * 100
   ```

3. **Processing Latency Heatmap**
   ```
   sum(rate(otelcol_processor_<n>_latency_bucket[1m])) by (le)
   ```

### Health Status Panels

Create color-coded health indicators:

1. **Memory Usage Status**
   ```
   // Green when < 70%, yellow when 70-85%, red when > 85%
   process_resident_memory_bytes{job="nrdot-collector-self-telemetry"} /
   process_resident_memory_bytes{job="nrdot-collector-self-telemetry", quantile="maximum"}
   ```

2. **Processor Error Rate Status**
   ```
   // Warning threshold at 0.1% error rate
   rate(otelcol_processor_<n>_dropped_metric_points[5m]) /
   rate(otelcol_processor_<n>_received_metric_points[5m])
   ```

### Volume Reduction Visualization

Demonstrate metric volume reduction:

1. **Daily Volume Reduction Gauge**
   ```
   // Volume reduction percentage
   (sum(increase(otelcol_receiver_accepted_metric_points[24h])) - sum(increase(otelcol_exporter_sent_metric_points[24h]))) / sum(increase(otelcol_receiver_accepted_metric_points[24h])) * 100
   ```

2. **Monthly Trend Projection**
   ```
   // Forecast based on current trend
   predict_linear(sum(increase(otelcol_exporter_sent_metric_points[24h]))[30d:1d], 86400 * 30)
   ```

### Algorithm Insight Panels

For algorithm-specific dashboards:

1. **Decision Process Visualization**
   ```
   // Example: PriorityTagger tagging by reason
   sum(increase(otelcol_otelcol_prioritytagger_critical_processes_tagged_total{reason="exact_match"}[5m]))
   ```

2. **Dynamic K Adjustment**
   ```
   // AdaptiveTopK K value changes with system load
   otelcol_otelcol_adaptivetopk_current_k_value vs system_cpu_utilization
   ```

3. **Reservoir Sampling Visualization**
   ```
   // ReservoirSampler fill ratio
   otelcol_otelcol_reservoirsampler_reservoir_fill_ratio
   ```

---

## Implementation Guidelines

### 1. Template Variables

Use Grafana template variables for flexible analysis:

- `$processor` - Select a specific processor
- `$time_window` - Adjust time window (1m, 5m, 1h)
- `$instance` - Filter by collector instance
- `$reduction_target` - Configurable reduction target percentage

### 2. Annotations

Add annotations for key events:

- Collector deployments/updates
- Configuration changes
- Scaling events

### 3. Alert Integration

Configure Grafana alerts for:

- Processor error rates exceeding thresholds
- Pipeline throughput anomalies
- Memory usage warnings
- Volume spikes

### 4. Visualization Types

Choose appropriate visualizations:

- **Time Series**: For throughput, error rates over time
- **Gauge**: For optimization percentages, health indicators
- **Bar Charts**: For comparing processor efficiency
- **Tables**: For detailed metrics breakdowns
- **Heatmaps**: For latency distributions
- **Stat Panels**: For key performance indicators
- **Pie Charts**: For cardinality distributions
- **Histograms**: For distribution of metric values

---

## Available Dashboard JSONs

See the `dashboards/` directory for complete ready-to-import dashboard JSONs:

- `grafana-nrdot-system-overview.json` - System-level overview dashboard
- `grafana-nrdot-custom-processor-starter-kpis.json` - Starter KPIs dashboard
- `grafana-nrdot-optimization-pipeline.json` - Full optimization pipeline dashboard
- `grafana-nrdot-prioritytagger-processor.json` - PriorityTagger-specific dashboard

### Algorithm Insight Dashboards:
- `grafana-nrdot-prioritytagger-algo.json` - PriorityTagger algorithm insights
- `grafana-nrdot-adaptivetopk-algo.json` - AdaptiveTopK algorithm insights
- `grafana-nrdot-othersrollup-algo.json` - OthersRollup algorithm insights
- `grafana-nrdot-reservoirsampler-algo.json` - ReservoirSampler algorithm insights

---

## Dashboard Update and Maintenance Strategy

1. **Version Control**: Store dashboard JSONs in the repository under version control.
2. **Automatic Provisioning**: Use Grafana's provisioning to automatically load dashboard updates.
3. **Standardized Naming**: Follow consistent naming conventions for metrics and dashboards.
4. **Documentation Links**: Include links to relevant documentation within dashboards.
5. **Regular Review**: Schedule quarterly reviews to ensure dashboards remain relevant and effective.
6. **Metric Enhancement**: Identify and implement additional metrics or labels needed for deeper algorithm insights.

By following these guidelines, your NRDOT collector dashboards will provide comprehensive visibility into the pipeline's performance, the effectiveness of your process-metrics optimization strategy, and detailed insights into the internal workings of each processor algorithm.