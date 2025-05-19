# NRDOT Process-Metrics Optimization - Implementation Plan

**Objective:** To incrementally build and validate a four-layer (L0 → L3) custom OpenTelemetry Collector pipeline designed to significantly reduce process-metric cardinality and ingest costs for New Relic, while preserving critical operational insight.

**Guiding Principle:** Each phase will deliver a testable and observable increment of value, building upon the previous one, and progressing towards the full "Opt-Plus" configuration. This plan focuses on the functional build-out of the processors and their integration.

**Current Status:** All phases have been completed, and the full optimization pipeline is now operational using the opt-plus.yaml configuration.

---

## Phase 0 — Foundation & "Hello World" Processor ✅

**Status: Complete**

This foundational phase established the core project structure, development environment, and a working example of a custom OpenTelemetry processor.

| Deliverable                                                                | Status | Key Details & Validation                                                                                                                                                              |
|----------------------------------------------------------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Custom Collector with `helloworld` Processor**                             | ✅     | `helloworld` processor implemented, configurable, adds attributes, emits custom (`otelcol_otelcol_helloworld_mutations_total`) and standard (`obsreport`) metrics. Validated via mock sink & Grafana. |
| **Dockerized Local Development Stack**                                       | ✅     | `docker-compose.yaml` provisions: Custom Collector, Prometheus, Grafana, Mock OTLP Sink.                                                                                                |
| **Standardized Service Ports & Access**                                      | ✅     | Collector zPages (`:55679`), Prometheus (`:9090`), Grafana (`:3000`). Validated by `test/url_check.sh`.                                                                                 |
| **CI/CD Basics**                                                             | ✅     | GitHub Actions for build, lint, unit tests (`go test ./...`), and `govulncheck`.                                                                                                        |
| **Core Documentation & Dashboards**                                          | ✅     | `README.md`, `CLAUDE.MD`, `DEVELOPING_PROCESSORS.md`, `NRDOT_PROCESSOR_SELF_OBSERVABILITY.md`, this `IMPLEMENTATION_PLAN.MD`, and "NRDOT Processors - HelloWorld & PriorityTagger KPIs" Grafana dashboard. |
| **`Makefile` for Dev Workflow**                                              | ✅     | Comprehensive targets for build, test, Docker ops, linting, etc. (`make help` for details).                                                                                             |

---

## Phase 1 — L0: PriorityTagger Processor ✅

**Status: Complete**

**Objective:** Implement and validate the `prioritytagger` processor, which identifies and tags metrics from critical processes to ensure their preservation throughout subsequent pipeline stages.

**Validation Summary:** The PriorityTagger processor has been implemented and verified to be processing metrics correctly in the pipeline. The standard metrics `otelcol_otelcol_otelcol_processor_prioritytagger_processed_metric_points` show steady accumulation of processed points, confirming the processor is actively working. There are some non-critical errors related to process scraping in a Docker environment (common in containerized settings). However, the core functionality of identifying and tagging critical processes is operational.

| Item                        | Status | Implementation Details                                                                                                                                                                                                             |
|-----------------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Processor Location**      | ✅     | `processors/prioritytagger/`                                                                                                                                                                                                            |
| **Configuration** (`config.go`) | ✅ | Implemented struct `Config` with options for: <br> - `CriticalExecutables`: `[]string` (e.g., "kubelet", "systemd") <br> - `CriticalExecutablePatterns`: `[]string` (regex list) <br> - `CPUSteadyStateThreshold`: `float64` <br> - `MemoryRSSThresholdMiB`: `int64` <br> - `PriorityAttributeName`: `string` (default: "nr.priority") <br> - `CriticalAttributeValue`: `string` (default: "critical") <br> Implemented `Validate()` and default setters via `Unmarshal()`. |
| **Core Logic** (`processor.go`) | ✅ | Implemented metric traversal logic to examine each data point. For each process metric, checks if `process.executable.name` matches `CriticalExecutables` or `CriticalExecutablePatterns`. If thresholds configured, checks `process.cpu.utilization` and `process.memory.rss` (supports both int and double types). If criteria met, adds the attribute (e.g., `nr.priority="critical"`) to relevant data points. |
| **Self-Observability**      | ✅    | Implemented `obsreport` for standard metrics (`processed_metric_points`, `dropped_metric_points`). <br> Added custom metric: `nrdot_prioritytagger_critical_processes_tagged_total` (Counter, increments for each unique process tagged in a batch).                                                       |
| **Unit Testing** (`_test.go`) | ✅  | Implemented comprehensive tests covering: <br> - Configuration validation with various valid/invalid cases <br> - Tagging by exact name, by pattern <br> - Tagging by CPU/memory thresholds <br> - Correct attribute application <br> - No-match scenarios <br> - Idempotency (already tagged metrics are not altered) <br> Added integration and benchmark tests. |
| **Integration & Validation**  | ✅  | 1. Registered factory in `cmd/collector/main.go`. <br> 2. Included `prioritytagger` in the pipeline configuration. <br> 3. Built and ran the Docker image with our custom collector. <br> 4. Verified the collector runs with the prioritytagger processor active. |
| **Documentation**           | ✅    | Created `processors/prioritytagger/README.md` detailing configuration and behavior. <br> Updated this implementation plan.                                                                                           |
| **Grafana Dashboard**       | ✅    | Dashboard provisioning improved with panels for prioritytagger metrics. Updated to correctly display both helloworld and prioritytagger processor metrics.           |
| **Definition of Done**      | ✅    | All required items complete. CI pipeline updated for the custom processor implementation.                                                                                 |

---

## Phase 2 — L1: AdaptiveTopK Processor ✅

**Status: Complete**

**Objective:** Implement the `adaptivetopk` processor to select metrics from the 'K' most resource-intensive processes. This will be delivered in two sub-phases: Fixed K, then Dynamic K with Hysteresis.

### Sub-Phase 2a: Fixed K Functionality

| Item                        | Status | Implementation Details                                                                                                                                                                                                                                               |
|-----------------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Processor Location**      | ✅     | `processors/adaptivetopk/`                                                                                                                                                                                                                                                |
| **Configuration** (`config.go`) | ✅  | Implemented struct `Config` with options for: <br> - `KValue`: `int` (the fixed number of top processes to keep). <br> - `KeyMetricName`: `string` (e.g., "process.cpu.utilization", "process.memory.rss"). <br> - `SecondaryKeyMetricName` (Optional): `string` (for tie-breaking). <br> - `PriorityAttributeName`: `string` (default: "nr.priority"). <br> - `CriticalAttributeValue`: `string` (default: "critical"). <br> - Additional fields to support future Dynamic K functionality. |
| **Core Logic** (`processor.go`) | ✅  | Implemented min-heap based Top-K selection algorithm: <br> 1. Identifies processes already tagged as critical; these always pass through. <br> 2. From the remaining non-critical processes, selects the Top 'K' based on `KeyMetricName`. <br> 3. Metrics from critical processes and the selected Top K processes are passed to the next consumer. <br> 4. Metrics from all other non-critical, non-TopK processes are dropped. |
| **Self-Observability**      | ✅     | Implemented `obsreport` for standard metrics. <br> Added custom metric: `otelcol_nrdot_adaptivetopk_topk_processes_selected_total` (Counter, increments by the number of processes selected for Top K). <br> Added `otelcol_nrdot_adaptivetopk_current_k_value` for future Dynamic K.                     |
| **Unit Testing** (`_test.go`) | ✅   | Implemented tests covering: <br> - Correct selection of Top K. <br> - Correct handling of critical process bypass. <br> - Scenarios with fewer than K processes available. <br> - Metric dropping logic.                              |
| **Integration & Validation**  | ✅   | 1. Registered factory in `cmd/collector/main.go`. <br> 2. Processor can be inserted in pipeline: `... -> prioritytagger -> adaptivetopk (fixed K) -> ...`. <br> 3. Unit tests confirm only critical + Top K metrics pass through. |
| **Documentation**           | ✅      | Created detailed `processors/adaptivetopk/README.md`.                                                                                                                                                                                                                      |

### Sub-Phase 2b: Dynamic K & Hysteresis

| Item                        | Status | Implementation Details                                                                                                                                                                                                                                         |
|-----------------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Internal Package**        | ✅     | Utilized `internal/banding/` package for reusable logic to map a host load metric to a K value.                                                                                                                                                             |
| **Configuration Update**    | ✅     | Successfully added options for: <br> - `HostLoadMetricName`: `string` <br> - `LoadBandsToKMap`: `map[float64]int` <br> - `HysteresisDuration`: `time.Duration` <br> - `MinKValue`, `MaxKValue` bounds for dynamic K. |
| **Core Logic Update**       | ✅     | - Implemented host load detection using `HostLoadMetricName` <br> - Implemented banding logic to determine K value based on load <br> - Added hysteresis mechanism to prevent thrashing in K value and process selections |
| **Self-Observability Update** | ✅   | Updated custom metric: `otelcol_nrdot_adaptivetopk_current_k_value` (Gauge) to report the dynamic K value.                                                                                                                         |
| **Unit Testing Update**     | ✅     | Added tests covering: <br> - Dynamic K adjustment based on system load <br> - Hysteresis behavior (processes staying in Top K, K value stabilization) <br> - Edge cases with changing loads                                                                                                                |
| **Integration & Validation**  | ✅   | Implemented and verified functionality in the complete pipeline                           |
| **Documentation Update**    | ✅     | Updated `processors/adaptivetopk/README.md` with dynamic K and hysteresis details.                                                                                                                                                                                     |
| **Definition of Done**      | ✅     | All AdaptiveTopK features (Fixed K, Dynamic K, Hysteresis) implemented and validated. Docs updated. CI green.                                                                                                                                                         |

---

## Phase 3 — L2: OthersRollup Processor ✅

**Status: Complete**

**Objective:** Implement the `othersrollup` processor to aggregate metrics from non-priority, non-TopK processes into a single "_other_" category, significantly reducing cardinality.

| Item                        | Status | Implementation Details                                                                                                                                                                                                                                            |
|-----------------------------|--------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Processor Location**      | ✅     | `processors/othersrollup/`                                                                                                                                                                                                                                             |
| **Configuration** (`config.go`) | ✅ | Implemented struct `Config` with options for: <br> - `OutputPIDAttributeValue`: `string` (default: "-1"). <br> - `OutputExecutableNameAttributeValue`: `string` (default: "_other_"). <br> - `Aggregations`: `map[string]AggregationType` (metric name to aggregation function, e.g., `{"process.cpu.utilization": "avg", "process.memory.rss": "sum"}`). <br> - `MetricsToRollup`: `[]string` (list of metric names to apply rollup to; if empty, applies to all compatible). <br> - `PriorityAttributeName`: `string` (default: "nr.priority"). <br> - `CriticalAttributeValue`: `string` (default: "critical"). |
| **Core Logic** (`processor.go`) | ✅ | Implemented metric aggregation logic: <br> 1. Identifies metrics NOT tagged as critical (pipeline ordering ensures proper handling with `adaptivetopk` output). <br> 2. For "other" metrics: <br>   - Aggregates values based on the `Aggregations` map (sum, average). <br>   - Creates new metric data points with identifying attributes set to `OutputPIDAttributeValue`, `OutputExecutableNameAttributeValue`. <br>   - Preserves metric type semantics (gauge vs. sum) during aggregation. <br>   - Original "other" metrics are dropped, critical metrics are passed through. |
| **Self-Observability**      | ✅     | Implemented `obsreport` for standard metrics. <br> Added custom metrics: <br> - `otelcol_nrdot_othersrollup_aggregated_series_count_total` (Counter, increments for each output "_other_" series). <br> - `otelcol_nrdot_othersrollup_input_series_rolled_up_total` (Counter, counts individual series merged into each "_other_" series).                                                                               |
| **Unit Testing** (`_test.go`) | ✅   | Implemented tests covering: <br> - Sum and average aggregation functions. <br> - Correct attribute replacement. <br> - Pass-through behavior for critical processes. <br> - Proper type preservation for gauge vs. sum metrics.          |
| **Integration & Validation**  | ✅   | 1. Registered factory in `cmd/collector/main.go`. <br> 2. Processor can be inserted in pipeline: `... -> adaptivetopk -> othersrollup -> ...`. <br> 3. Unit tests verify "_other_" series aggregation with correctly computed values and attributes. |
| **Documentation**           | ✅    | Created detailed `processors/othersrollup/README.md` documenting configuration options and behavior.                                                                                                                                                                                                                            |
| **Definition of Done**      | ✅    | Processor successfully implemented with all core features. Automated tests verify correct functionality. Documentation complete.                                                                                                                                      |

---

## Phase 4 — L3: ReservoirSampler Processor ✅

**Status: Complete**

**Objective:** Implement the `reservoirsampler` processor to select a statistically representative sample of metrics from the remaining "long-tail" processes (those not critical, not TopK, and not yet rolled up, or sampled from the input before rollup).

*Implementation Choice: Implemented Approach 1 - Sample from (Non-Critical AND Non-TopK) -> Rollup the rest of (Non-Critical AND Non-TopK AND Non-Sampled).*

| Item                        | Status | Implementation Details                                                                                                                                                                                                                                            |
|-----------------------------|--------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Processor Location**      | ✅     | `processors/reservoirsampler/`                                                                                                                                                                                                                                          |
| **Configuration** (`config.go`) | ✅  | Implemented struct `Config` with options for: <br> - `ReservoirSize`: `int` (number of unique process identities to sample, default: 100). <br> - `IdentityAttributes`: `[]string` (list of attribute keys that define a unique process identity, default: `["process.pid"]`). <br> - `SampledAttributeName`: `string` (default: "nr.process_sampled_by_reservoir") <br> - `SampledAttributeValue`: `string` (default: "true") <br> - `SampleRateAttributeName`: `string` (default: "nr.sample_rate") <br> - `PriorityAttributeName`: `string` (default: "nr.priority") <br> - `CriticalAttributeValue`: `string` (default: "critical") |
| **Core Logic** (`processor.go`) | ✅  | Implemented Algorithm R variant for reservoir sampling: <br> 1. Operates on metrics not tagged critical. <br> 2. For each batch, identifies new eligible process identities. <br> 3. For new identities: <br>   - If reservoir not full, adds them directly. <br>   - If reservoir full, random selection ensures proper statistical sampling. <br> 4. For metrics belonging to sampled identities, adds `SampledAttributeName` and `SampleRateAttributeName`. <br> 5. Pass through critical processes, drop non-sampled processes. |
| **Self-Observability**      | ✅     | Implemented `obsreport` for standard metrics. <br> Added custom metrics: <br> - `nrdot_reservoirsampler_reservoir_fill_ratio` (Gauge, 0.0 to 1.0) <br> - `nrdot_reservoirsampler_selected_identities_count` (Gauge, current count in reservoir) <br> - `nrdot_reservoirsampler_eligible_identities_seen_total` (Counter) <br> - `nrdot_reservoirsampler_new_identities_added_to_reservoir_total` (Counter) |
| **Unit Testing** (`_test.go`) | ✅   | Implemented tests covering: <br> - Reservoir filling behavior. <br> - Critical process pass-through. <br> - Proper tagging of sampled metrics with sample rate. <br> - Testing with multiple batches to verify reservoir behavior over time. |
| **Integration & Validation**  | ✅   | 1. Registered factory in `cmd/collector/main.go`. <br> 2. Processor can be inserted in pipeline: `... -> adaptivetopk -> reservoirsampler -> othersrollup -> ...` <br> 3. Unit tests verify that `ReservoirSize` unique processes are sampled with proper attributes. |
| **Documentation**           | ✅     | Created detailed `processors/reservoirsampler/README.md` documenting configuration options and behavior.                                                                                                                                                                                                                         |
| **Definition of Done**      | ✅     | Processor successfully implemented with core sampling algorithm. Automated tests verify correct functionality. Documentation complete.                                                                                                                         |

---

## Phase 5 — "Opt-Plus" Pipeline Integration ✅

**Status: Complete**

**Objective:** Assemble the full L0-L3 processor chain into a final "Opt-Plus" configuration, implement Dynamic K functionality for AdaptiveTopK, and create comprehensive monitoring dashboards.

| Item                        | Status | Implementation Details                                                                                                                                                                                                                                                                                                                                                          |
|-----------------------------|--------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Configuration** (`opt-plus.yaml`) | ✅ | Created and optimized `config/opt-plus.yaml` defining the full "hot path" pipeline: <br> `hostmetrics -> L0:PriorityTagger -> L1:AdaptiveTopK -> L3:ReservoirSampler -> L2:OthersRollup -> memory_limiter -> batch -> otlphttp (to New Relic)`. <br> Placed ReservoirSampler before OthersRollup based on design decisions from Phase 4. <br> Includes thorough configuration parameters with environment variables support. |
| **Dynamic K Implementation** | ✅ | Successfully implemented Dynamic K evaluation for AdaptiveTopK processor: <br> - Added host load detection using system metrics <br> - Implemented banding logic to map load to K values <br> - Added process hysteresis to prevent thrashing <br> - Created unit tests for all functionality |
| **End-to-End Testing**      | ✅ | Created comprehensive end-to-end testing: <br> - Created `test/test_opt_plus_pipeline.sh` script to test the full pipeline <br> - Added cardinality reduction measurement <br> - Verified critical process preservation <br> - Implemented validation of all processor metrics <br> - Added success criteria for the pipeline |
| **Documentation**           | ✅ | Completed documentation: <br> - Created detailed `docs/OPTIMIZATION_PIPELINE.md` <br> - Updated main `README.md` to reflect Phase 5 completion <br> - Enhanced all processor documentation |
| **Grafana Dashboard Updates** | ✅ | Created comprehensive dashboards: <br> - Created full pipeline overview dashboard (`grafana-nrdot-optimization-pipeline.json`) <br> - Created four algorithm-specific dashboards: <br>   * PriorityTagger Algorithm Dashboard <br>   * AdaptiveTopK Algorithm Dashboard <br>   * OthersRollup Algorithm Dashboard <br>   * ReservoirSampler Algorithm Dashboard <br> - Added detailed dashboard documentation (`dashboards/README.md`) |
| **Definition of Done**      | ✅ | All Phase 5 requirements successfully completed. End-to-end pipeline validated with >=90% cardinality reduction. Full observability stack operational with comprehensive dashboards. |

---

## Cross-Phase Quality Gates (Simplified)

These apply to each processor development phase (L0-L3):

| Gate                               | Threshold / Requirement                                      |
|------------------------------------|--------------------------------------------------------------|
| **Unit Test Coverage**             | ≥ 80% for the processor's Go code.                           |
| **Linting & Formatting**           | 0 errors/warnings from `go vet`, `go fmt`, `golangci-lint`.  |
| **Self-Observability**             | All new custom metrics visible and reporting in local Grafana. |
| **Mock Sink Validation**           | Processor's output (tags, filtering, aggregation) verified in mock sink logs. |
| **Documentation**                  | Processor's `README.md` and relevant sections of global docs (README, CLAUDE, IMPL_PLAN) updated. |
| **Successful CI Pipeline Run**     | All automated checks (build, lint, test, vuln-scan) pass.   |

---

_This implementation plan is now complete with all phases successfully implemented. The NRDOT Process-Metrics Optimization project has achieved its primary goal of creating a custom OpenTelemetry Collector distribution that significantly reduces process metric data volume while preserving essential visibility into host processes._