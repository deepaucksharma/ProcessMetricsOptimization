# NRDOT Process-Metrics Optimization - Implementation Plan

**Objective:** To incrementally build and validate a four-layer (L0 → L3) custom OpenTelemetry Collector pipeline designed to significantly reduce process-metric cardinality and ingest costs for New Relic, while preserving critical operational insight.

**Guiding Principle:** Each phase will deliver a testable and observable increment of value, building upon the previous one, and progressing towards the full "Opt-Plus" configuration. This plan focuses on the functional build-out of the processors and their integration.

---

## Phase 0 — Foundation & "Hello World" Processor ✅

**Status: Complete**

This foundational phase established the core project structure, development environment, and a working example of a custom OpenTelemetry processor.

| Deliverable                                                                | Status | Key Details & Validation                                                                                                                                                              |
|----------------------------------------------------------------------------|--------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Custom Collector with `helloworld` Processor**                             | ✅     | `helloworld` processor implemented, configurable, adds attributes, emits custom (`nrdot_helloworld_mutations_total`) and standard (`obsreport`) metrics. Validated via mock sink & Grafana. |
| **Dockerized Local Development Stack**                                       | ✅     | `docker-compose.yaml` provisions: Custom Collector, Prometheus, Grafana, Mock OTLP Sink.                                                                                                |
| **Standardized Service Ports & Access**                                      | ✅     | Collector zPages (`:55679`), Prometheus (`:9090`), Grafana (`:3000`). Validated by `test/url_check.sh`.                                                                                 |
| **CI/CD Basics**                                                             | ✅     | GitHub Actions for build, lint, unit tests (`go test ./...`), and `govulncheck`.                                                                                                        |
| **Core Documentation & Dashboards**                                          | ✅     | `README.md`, `CLAUDE.MD`, `DEVELOPING_PROCESSORS.md`, `NRDOT_PROCESSOR_SELF_OBSERVABILITY.md`, this `IMPLEMENTATION_PLAN.MD`, and "NRDOT Custom Processor Starter KPIs" Grafana dashboard. |
| **`Makefile` for Dev Workflow**                                              | ✅     | Comprehensive targets for build, test, Docker ops, linting, etc. (`make help` for details).                                                                                             |
| **Relocated Simple Demo**                                                    | ✅     | Original top-level `main.go` moved to `examples/simple_demo/main.go`.                                                                                                                 |

---

## Phase 1 — L0: PriorityTagger Processor ⏳

**Objective:** Implement and validate the `prioritytagger` processor, which identifies and tags metrics from critical processes to ensure their preservation throughout subsequent pipeline stages.

| Item                        | Target Specification & Key Considerations                                                                                                                                                                                                             |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Processor Location**      | `processors/prioritytagger/`                                                                                                                                                                                                                            |
| **Configuration** (`config.go`) | Define struct `Config` with options for: <br> - `CriticalExecutables`: `[]string` (e.g., "kubelet", "systemd") <br> - `CriticalExecutablePatterns`: `[]string` (regex list) <br> - (Optional) `CPUSteadyStateThreshold`: `float64` <br> - (Optional) `MemoryRSThresholdMiB`: `int64` <br> - `PriorityAttributeName`: `string` (default: "nr.priority") <br> - `CriticalAttributeValue`: `string` (default: "critical") <br> Implement `Validate()` and default setters. |
| **Core Logic** (`processor.go`) | Iterate `pmetric.Metrics`. For each process, check if `process.executable.name` matches `CriticalExecutables` or `CriticalExecutablePatterns`. <br> (Optional) If thresholds configured, check `process.cpu.utilization` (ensure it represents steady state) and `process.memory.rss`. <br> If criteria met, add/update the attribute (e.g., `nr.priority="critical"`) to relevant data points. |
| **Self-Observability**      | Implement `obsreport` for standard metrics. <br> Custom Metric: `nrdot_prioritytagger_critical_processes_tagged_total` (Counter, increments for each process tagged in a batch).                                                       |
| **Unit Testing** (`_test.go`) | Cover: <br> - Configuration validation (valid/invalid cases). <br> - Tagging by exact name, by pattern. <br> - (Optional) Tagging by CPU/memory thresholds. <br> - Correct attribute name/value. <br> - No-match scenarios (metrics pass through unmodified). <br> - Idempotency (already tagged metrics are not altered negatively). <br> Aim for ≥80% coverage. |
| **Integration & Validation**  | 1. Register factory in `cmd/collector/main.go`. <br> 2. Update `config/base.yaml` (or a test copy) to include `prioritytagger` in the pipeline: `hostmetrics -> prioritytagger -> helloworld -> mock_sink`. <br> 3. `make docker-build && make compose-up`. <br> 4. Verify tagged metrics in Mock OTLP Sink logs. <br> 5. Verify `nrdot_prioritytagger_critical_processes_tagged_total` in Prometheus/Grafana. |
| **Documentation**           | Create `processors/prioritytagger/README.md` detailing configuration and behavior. <br> Update main `README.md`, `CLAUDE.MD`, and this plan.                                                                                           |
| **Grafana Dashboard**       | Add a panel to "NRDOT Custom Processor Starter KPIs" for `nrdot_prioritytagger_critical_processes_tagged_total` (e.g., rate over time).                                                                                             |
| **Definition of Done**      | All above items complete. `make test-unit` and `make test-urls` pass. CI pipeline green. Metrics correctly tagged and visible in local dev stack.                                                                                 |

---

## Phase 2 — L1: AdaptiveTopK Processor ⏳

**Objective:** Implement the `adaptivetopk` processor to select metrics from the 'K' most resource-intensive processes. This will be delivered in two sub-phases: Fixed K, then Dynamic K with Hysteresis.

### Sub-Phase 2a: Fixed K Functionality

| Item                        | Target Specification & Key Considerations                                                                                                                                                                                                                               |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Processor Location**      | `processors/adaptivetopk/`                                                                                                                                                                                                                                                |
| **Configuration** (`config.go`) | Options for: <br> - `KValue`: `int` (the fixed number of top processes to keep). <br> - `KeyMetricName`: `string` (e.g., "process.cpu.utilization", "process.memory.rss"). <br> - `SecondaryKeyMetricName` (Optional): `string` (for tie-breaking). <br> - `PriorityAttributeName`: `string` (default: "nr.priority"). <br> - `CriticalAttributeValue`: `string` (default: "critical"). |
| **Core Logic** (`processor.go`) | - For each batch of metrics: <br>   1. Identify processes already tagged as critical (e.g., `nr.priority="critical"`); these always pass through. <br>   2. From the remaining non-critical processes, select the Top 'K' based on `KeyMetricName` (using a min-heap is efficient). <br>   3. Metrics from critical processes and the selected Top K processes are passed to the next consumer. <br>   4. Metrics from all other non-critical, non-TopK processes are dropped. |
| **Self-Observability**      | Implement `obsreport` (crucially, `otelcol_processor_adaptivetopk_dropped_metric_points` will be high). <br> Custom Metric: `nrdot_adaptivetopk_topk_processes_selected_total` (Counter, increments by K, or actual number if less than K available).                     |
| **Unit Testing** (`_test.go`) | Cover: <br> - Correct selection of Top K. <br> - Correct handling of critical process bypass. <br> - Tie-breaking logic (if `SecondaryKeyMetricName` used). <br> - Scenarios with fewer than K processes available. <br> - Metric dropping logic.                              |
| **Integration & Validation**  | 1. Integrate into `config/base.yaml`: `... -> prioritytagger -> adaptivetopk (fixed K) -> helloworld -> mock_sink`. <br> 2. Validate that only critical + Top K metrics appear in mock sink. <br> 3. Verify `otelcol_processor_adaptivetopk_dropped_metric_points` and custom counter in Prometheus/Grafana. |
| **Documentation**           | Create/update `processors/adaptivetopk/README.md`.                                                                                                                                                                                                                      |

### Sub-Phase 2b: Dynamic K & Hysteresis

| Item                        | Target Specification & Key Considerations                                                                                                                                                                                                                         |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Internal Package**        | Consider `internal/banding/` for reusable logic to map a host load metric to a K value.                                                                                                                                                                             |
| **Configuration Update**    | Add/modify options for: <br> - `HostLoadMetricName`: `string` (e.g., "system.cpu.utilization" - this metric must be present in the input batch). <br> - `LoadBandsToKMap`: `map[float64]int` (e.g., `{0.2: 5, 0.5: 10, 0.8: 20}` where key is load threshold, value is K). <br> - `HysteresisDuration`: `time.Duration` (e.g., "15s", "1m"). <br> - `MinKValue`, `MaxKValue` (bounds for dynamic K). <br> Remove `KValue` (fixed). |
| **Core Logic Update**       | - Determine current host load from `HostLoadMetricName`. <br> - Use `internal/banding` (or inline logic) to find the target K based on `LoadBandsToKMap`. <br> - Implement hysteresis: once a process enters Top K, it stays for `HysteresisDuration` even if it drops below the threshold, unless evicted by a higher-ranking process. This prevents flapping. The K value itself should also adjust smoothly. |
| **Self-Observability Update** | Custom Metric: `nrdot_adaptivetopk_current_k_value` (Gauge).                                                                                                                                                                                                          |
| **Unit Testing Update**     | Cover: <br> - Dynamic K adjustment based on varying host load. <br> - Hysteresis behavior (processes staying in Top K, K value stabilization).                                                                                                                          |
| **Integration & Validation**  | Simulate varying host load (e.g., by manipulating a test metric generator or using a configurable test receiver) and verify that `nrdot_adaptivetopk_current_k_value` adapts as expected and the number of exported PIDs changes accordingly.                             |
| **Documentation Update**    | Update `processors/adaptivetopk/README.md` with dynamic K and hysteresis details.                                                                                                                                                                                     |
| **Definition of Done**      | All AdaptiveTopK features (Fixed K, Dynamic K, Hysteresis) implemented and validated. Docs updated. CI green.                                                                                                                                                         |

---

## Phase 3 — L2: OthersRollup Processor ⏳

**Objective:** Implement the `othersrollup` processor to aggregate metrics from non-priority, non-TopK processes into a single "_other_" category, significantly reducing cardinality.

| Item                        | Target Specification & Key Considerations                                                                                                                                                                                                                            |
|-----------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Processor Location**      | `processors/othersrollup/`                                                                                                                                                                                                                                             |
| **Configuration** (`config.go`) | Options for: <br> - `OutputPIDAttributeValue`: `string` (default: "-1"). <br> - `OutputExecutableNameAttributeValue`: `string` (default: "_other_"). <br> - `Aggregations`: `map[string]string` (metric name to aggregation function, e.g., `{"process.cpu.utilization": "avg", "process.memory.usage": "sum"}`). <br> - `MetricsToRollup`: `[]string` (list of metric names to apply rollup to; if empty, applies to all compatible). |
| **Core Logic** (`processor.go`) | - Identify metrics NOT tagged as critical AND NOT part of the Top K selection (requires careful pipeline ordering or passing state/tags from `adaptivetopk`). <br> - For these "other" metrics: <br>   1. Aggregate values based on the `Aggregations` map (sum, average). <br>   2. Create new metric data points with identifying attributes set to `OutputPIDAttributeValue`, `OutputExecutableNameAttributeValue`, etc. Original PID/Executable attributes are removed/replaced. <br>   3. Original "other" metrics are dropped. |
| **Self-Observability**      | `obsreport` for standard metrics. <br> Custom Metric: `nrdot_othersrollup_aggregated_series_count_total` (Counter, increments for each output "_other_" series generated per batch).                                                                               |
| **Unit Testing** (`_test.go`) | Cover: <br> - Different aggregation functions (sum, avg). <br> - Correct attribute replacement. <br> - Scenarios with no "other" metrics. <br> - Use of `MetricsToRollup` filter. <br> Golden file tests for complex aggregation scenarios are highly recommended.          |
| **Integration & Validation**  | 1. Integrate into `config/base.yaml`: `... -> adaptivetopk -> othersrollup -> helloworld -> mock_sink`. <br> 2. Verify that "_other_" series appears in mock sink with correctly aggregated values and attributes. <br> 3. Validate the "Sum Check": `Sum("Others" CPU) + Sum(TopK CPU) + Sum(Critical CPU) ≈ Host System CPU` (requires a way to measure host system CPU independently or ensure all input CPU metrics are accounted for). <br> 4. Verify custom counter. |
| **Documentation**           | Create `processors/othersrollup/README.md`.                                                                                                                                                                                                                            |
| **Definition of Done**      | `othersrollup` implemented and validated. Sum check holds within acceptable tolerance (e.g., <2-5%). Docs updated. CI green.                                                                                                                                      |

---

## Phase 4 — L3: ReservoirSampler Processor ⏳

**Objective:** Implement the `reservoirsampler` processor to select a statistically representative sample of metrics from the remaining "long-tail" processes (those not critical, not TopK, and not yet rolled up, or sampled from the input before rollup).

*Decision Point: Determine precise interaction with `OthersRollup`. Common approaches:
    1.  Sample from (Non-Critical AND Non-TopK) -> Rollup the rest of (Non-Critical AND Non-TopK AND Non-Sampled).
    2.  Rollup (Non-Critical AND Non-TopK) -> Sample from the original (Non-Critical AND Non-TopK) set (more complex state).
    Approach 1 is generally simpler to implement.*

| Item                        | Target Specification & Key Considerations (Assuming Approach 1 above)                                                                                                                                                                                                 |
|-----------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Processor Location**      | `processors/reservoirsampler/`                                                                                                                                                                                                                                          |
| **Configuration** (`config.go`) | Options for: <br> - `ReservoirSize`: `int` (number of unique process identities to sample). <br> - `IdentityAttributes`: `[]string` (list of attribute keys that define a unique process identity for sampling, e.g., `["process.executable.name", "process.command_line"]`). <br> - `SampledAttributeName`: `string` (default: "nr.process_sampled_by_reservoir") <br> - `SampledAttributeValue`: `string` (default: "true") <br> - `SampleRateAttributeName`: `string` (default: "nr.sample_rate") |
| **Core Logic** (`processor.go`) | - Operates on metrics not tagged critical and not selected by TopK. <br> - Implement Algorithm R (or a variant like L) for reservoir sampling to select `ReservoirSize` unique process identities over time. <br> - For metrics belonging to selected sampled identities, add `SampledAttributeName="true"` and `SampleRateAttributeName` (calculated based on total eligible identities seen vs. reservoir size). <br> - Metrics from non-critical, non-TopK, non-sampled processes are either dropped or passed to a subsequent `OthersRollup` (if `OthersRollup` is placed after `ReservoirSampler`). |
| **Self-Observability**      | `obsreport`. <br> Custom Metrics: `nrdot_reservoirsampler_reservoir_fill_ratio` (Gauge, 0.0 to 1.0). `nrdot_reservoirsampler_selected_identities_count` (Gauge, current number of unique identities in reservoir). `nrdot_reservoirsampler_eligible_identities_seen_total` (Counter). |
| **Unit Testing** (`_test.go`) | Cover: <br> - Reservoir filling and replacement logic. <br> - Correct attribute tagging for sampled items. <br> - Sample rate calculation. <br> - Statistical properties (e.g., run with a large synthetic dataset and check if known items appear with expected probability, if feasible). |
| **Integration & Validation**  | 1. Integrate into pipeline, e.g., `... -> adaptivetopk -> reservoirsampler -> othersrollup -> helloworld -> mock_sink`. <br> 2. Verify that approximately `ReservoirSize` unique process identities (with sampled attributes) appear in mock sink over time. <br> 3. Check `SampleRateAttributeName` for plausible values. <br> 4. Monitor self-metrics in Prometheus/Grafana. |
| **Documentation**           | Create `processors/reservoirsampler/README.md`.                                                                                                                                                                                                                         |
| **Definition of Done**      | `reservoirsampler` implemented and validated. Sampled metrics correctly tagged. Self-metrics reflect sampler state. Docs updated. CI green.                                                                                                                            |

---

## Phase 5 — "Opt-Plus" Pipeline Integration & Hot/Cold Paths ⏳

**Objective:** Assemble the full L0-L3 processor chain into a final "Opt-Plus" configuration. Optionally, implement a "cold path" for exporting raw, unprocessed data.

| Item                        | Target Specification & Key Considerations                                                                                                                                                                                                                                                                                                                                           |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Configuration** (`opt-plus.yaml`) | Create `config/opt-plus.yaml` defining the full "hot path" pipeline: <br> `hostmetrics -> L0:PriorityTagger -> L1:AdaptiveTopK -> L3:ReservoirSampler -> L2:OthersRollup -> memory_limiter -> batch -> otlphttp (to New Relic)`. <br> *Note: Order of L2 and L3 might be swapped based on design decisions from Phase 4.* <br> Parameterize thoroughly with environment variables. |
| **Cold Path (Optional)**    | If implemented: <br> - Add a parallel pipeline in `opt-plus.yaml`: `hostmetrics -> batch -> (e.g., awss3exporter for Parquet, or fileexporter)`. <br> - Configure exporter for format (e.g., Parquet), S3 bucket/prefix, or file path.                                                                                                                                                 |
| **End-to-End Testing**      | - Test the full hot path against the local mock sink, verifying data transformations at each stage and final output. <br> - (If sending to New Relic) Validate data in NR One, focusing on cardinality reduction (target: ≤100 unique process series/host/day for key metrics like `process.cpu.utilization`). <br> - If cold path implemented, test data export and integrity. |
| **Documentation**           | - Document `config/opt-plus.yaml` options extensively. <br> - Update main `README.md` to highlight the full pipeline. <br> - Provide NRQL examples for querying optimized data in `docs/NRQL_EXAMPLES.md`.                                                                                                                                                                             |
| **Grafana Dashboard Update**| Create or significantly update Grafana dashboards to visualize the behavior and effectiveness of the entire Opt-Plus pipeline, including overall cardinality reduction and resource consumption of each processor.                                                                                                                                                                      |
| **Definition of Done**      | `opt-plus.yaml` fully configured and tested. Hot path meets cardinality targets in test environments. Cold path (if any) is functional. Documentation and dashboards reflect the complete pipeline. CI green.                                                                                                                                                                      |

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

_This implementation plan is a living document. It will be updated as each phase is completed and as new insights or requirements emerge. The primary focus remains on delivering a functionally complete and validated sequence of processors._
