# NRDOT Process-Metrics Optimization - Implementation Plan

This document outlines the detailed, iterative approach to implementing the NRDOT Process-Metrics Optimization pipeline from the initial "Hello-World" baseline to the full "Opt-Plus" production implementation.

## Overview

The plan follows a phased, incremental approach where each phase:
- Builds on the foundation of the previous phase
- Delivers tangible value or validated learning
- Has clear goals, deliverables, and validation criteria
- Maintains comprehensive test coverage and observability

## Phase 0: Foundation - "Hello-World" Custom Processor & Self-Observability
**Timeline: Complete**

### Goal
Establish the basic framework for building, deploying, and observing a custom NRDOT processor.

### Key Activities (Recap)
- [x] Set up go.mod, basic project structure
- [x] Implement helloworld processor (simple attribute mutation)
- [x] Integrate obsreport for standard self-metrics
- [x] Create main.go to register the custom processor
- [x] Develop Dockerfile for building the custom NRDOT collector
- [x] Create docker-compose.yaml for local collector + Prometheus + Grafana + Mock OTLP sink
- [x] Configure Prometheus to scrape collector self-metrics
- [x] Provide basic Grafana dashboard for helloworld processor's obsreport metrics
- [x] Enable zPages for local debugging
- [x] Simple Makefile for build/run

### Deliverable
- Runnable nrdot-hello starter kit

### Validation Criteria
- [x] `make compose-up` works successfully
- [x] helloworld processor correctly mutates metrics (verified in Mock OTLP sink or logging exporter)
- [x] obsreport metrics for helloworld (e.g., `otelcol_processor_helloworld_processed_metric_points`) are visible in Prometheus, Grafana, and zPages

### Status
Complete - Foundation established

## Phase 1: L0 - PriorityTagger Processor - "The Unskippables"
**Timeline: Week 1-2**

### Goal
Implement and validate the first optimization layer: tagging critical processes to ensure they are never dropped by subsequent layers.

### Key Activities
- [ ] **Configuration Design**
  - [ ] Define YAML configuration schema (critical_executables list, cpu_utilization_threshold, memory_rss_threshold_mib)
  - [ ] Create JSON schema for validation

- [ ] **Implementation**
  - [ ] Create `processors/prioritytagger/` directory structure
  - [ ] Implement Config and Factory structs
  - [ ] Implement processor logic to identify critical processes
  - [ ] Add nr.priority="critical" attribute to matched processes
  - [ ] Use internal/attributes package for efficient attribute manipulation
  - [ ] Integrate with internal/obsreport wrapper
  - [ ] Implement custom self-metric: `nrdot_prioritytagger_critical_processes_tagged_total`

- [ ] **Testing**
  - [ ] Unit tests for configuration validation
  - [ ] Unit tests for regex matching and threshold-based tagging
  - [ ] Unit tests for no-match scenarios and idempotent behavior
  - [ ] Benchmark tests for attribute manipulation performance

- [ ] **Integration**
  - [ ] Register in main.go
  - [ ] Integrate into config/base.yaml
  - [ ] Update Docker Compose configuration

- [ ] **Observability**
  - [ ] Add Grafana panel for critical process count
  - [ ] Add NRQL query example: `SELECT count(*) FROM Metric WHERE nr.priority = 'critical' FACET process.executable.name`

- [ ] **Documentation**
  - [ ] README.md for prioritytagger
  - [ ] Update main documentation to include L0 processor

### Deliverables
- NRDOT collector image with functional prioritytagger
- Updated dashboards showing critical processes

### Validation Criteria
- [ ] Critical processes (based on executable name) are consistently tagged with `nr.priority="critical"`
- [ ] Processes exceeding configured CPU/memory thresholds are correctly tagged
- [ ] The `nrdot_prioritytagger_critical_processes_tagged_total` self-metric accurately reflects tagging activity
- [ ] Non-matching processes remain unmodified

### SLO Gate
- [ ] Prometheus metric for `nr.priority` attribute is visible in dashboards

## Phase 2: L1 - AdaptiveTopK Processor (Fixed K) - "Focus on the Hotspots"
**Timeline: Week 3-4**

### Goal
Implement the core Top-K selection logic, ensuring it correctly identifies and passes through the busiest processes based on a primary metric.

### Key Activities
- [ ] **Configuration Design**
  - [ ] Define YAML configuration for fixed K version (key_metric, key_metric_secondary, k_value)
  - [ ] Create JSON schema for validation

- [ ] **Implementation**
  - [ ] Create `processors/adaptivetopk/` directory structure
  - [ ] Implement Config and Factory structs
  - [ ] Implement min-heap or similar for Top-K tracking
  - [ ] Ensure critical-tagged processes (nr.priority="critical") bypass the Top-K selection
  - [ ] Pass through metrics for Top-K processes
  - [ ] Implement metric dropping for non-Top-K, non-critical processes
  - [ ] Integrate with internal/obsreport wrapper
  - [ ] Track and expose dropped_metric_points

- [ ] **Testing**
  - [ ] Unit tests for configuration validation
  - [ ] Unit tests for Top-K selection algorithm
  - [ ] Unit tests for critical process bypass logic
  - [ ] Unit tests for tie-breaking using secondary metric
  - [ ] Benchmark tests for Top-K selection performance

- [ ] **Integration**
  - [ ] Register in main.go
  - [ ] Place after prioritytagger in config/base.yaml
  - [ ] Update Docker Compose configuration

- [ ] **Observability**
  - [ ] Add Grafana panel for dropped metric points
  - [ ] Add NRQL query example: `FROM Metric SELECT uniqueCount(process.pid) SINCE 30 MINUTES`

- [ ] **Documentation**
  - [ ] README.md for adaptivetopk (fixed K version)
  - [ ] Update main documentation to include L1 processor

### Deliverables
- NRDOT collector image with functional prioritytagger and fixed-K adaptivetopk
- Updated dashboards showing Top-K processes

### Validation Criteria
- [ ] Only metrics from critical processes and the Top-K processes are exported
- [ ] The correct processes are selected as Top-K based on key_metric
- [ ] Critical processes bypass the Top-K selection
- [ ] `otelcol_processor_adaptivetopk_dropped_metric_points_total` accurately reflects dropped metrics

### SLO Gate
- [ ] uniqueCount(process.pid) on exported data ≤ k_value + actual critical count
- [ ] PromQL `otelcol_processor_adaptivetopk_dropped_metric_points_total` is visible and reporting sensible values

## Phase 3: L1 - AdaptiveTopK (Dynamic K & Hysteresis) - "Smarter Hotspot Focus"
**Timeline: Week 5-6**

### Goal
Enhance adaptivetopk with dynamic K adjustment based on host load and add hysteresis.

### Key Activities
- [ ] **Internal Support Package**
  - [ ] Implement `internal/banding/` package for host load to K-value mapping
  - [ ] Create unit tests for banding logic

- [ ] **Configuration Enhancement**
  - [ ] Extend YAML configuration (host_load_metric, load_bands, min_k, max_k, hysteresis_duration)
  - [ ] Update JSON schema

- [ ] **Implementation Enhancement**
  - [ ] Add functionality to read host load metric from batch
  - [ ] Integrate banding logic for dynamic K calculation
  - [ ] Implement hysteresis to prevent rapid changes in Top-K membership
  - [ ] Add custom self-metric: `nrdot_adaptivetopk_current_k`

- [ ] **Testing**
  - [ ] Unit tests for dynamic K adjustment
  - [ ] Unit tests for hysteresis behavior
  - [ ] Integration tests for host load scenarios

- [ ] **Integration**
  - [ ] Update config/base.yaml with dynamic K configuration
  - [ ] Create test scenarios for varying host loads

- [ ] **Observability**
  - [ ] Add Grafana panel for current K value
  - [ ] Add correlation panel with host load metric

- [ ] **Documentation**
  - [ ] Update adaptivetopk README.md with dynamic K and hysteresis details
  - [ ] Add examples of different banding configurations

### Deliverables
- Enhanced adaptivetopk processor with dynamic K and hysteresis
- Updated dashboards showing K adaptation

### Validation Criteria
- [ ] `nrdot_adaptivetopk_current_k` self-metric changes according to simulated host load
- [ ] Hysteresis prevents rapid flapping of processes in/out of Top-K
- [ ] Number of exported PIDs correctly reflects the dynamic K value

### SLO Gate
- [ ] `nrdot_adaptivetopk_current_k` shows appropriate adaptation to host load changes
- [ ] Validation criteria from Phase 2 continue to be met, but with dynamic K

## Phase 4: L2 - OthersRollup Processor - "Accounting for the Rest"
**Timeline: Week 7-8**

### Goal
Implement and validate aggregation of non-Top-K, non-critical process metrics into an "Others" category.

### Key Activities
- [ ] **Configuration Design**
  - [ ] Define YAML configuration (output_attributes, aggregations map)
  - [ ] Create JSON schema for validation

- [ ] **Implementation**
  - [ ] Create `processors/othersrollup/` directory structure
  - [ ] Implement Config and Factory structs
  - [ ] Implement aggregation logic (sum for resources, avg for utilization)
  - [ ] Strip original PID attributes and apply "Others" attributes
  - [ ] Integrate with internal/obsreport wrapper
  - [ ] Add custom self-metric: `nrdot_othersrollup_aggregated_series_count_total`

- [ ] **Testing**
  - [ ] Unit tests for configuration validation
  - [ ] Unit tests for different aggregation types (sum, avg)
  - [ ] Create golden file fixtures for complex scenarios
  - [ ] Integration tests with previous processors

- [ ] **Integration**
  - [ ] Register in main.go
  - [ ] Place after adaptivetopk in config/base.yaml
  - [ ] Create initial opt-plus.yaml
  - [ ] Update Docker Compose configuration

- [ ] **Observability**
  - [ ] Add Grafana panel comparing "Others CPU%" vs "Host CPU%"
  - [ ] Add NRQL query examples for "Others" metrics

- [ ] **Documentation**
  - [ ] README.md for othersrollup with aggregation details
  - [ ] Update main documentation to include L2 processor

### Deliverables
- NRDOT collector image with functional L0, L1, and L2 processors
- Updated dashboards showing "Others" aggregation

### Validation Criteria
- [ ] "Others" series appears with correct attributes and aggregated values
- [ ] Sum Check: `sum("Others" CPU) + sum(TopK CPU) + sum(Critical CPU) ≈ Host System CPU`
- [ ] Aggregation correctly applies sum vs. avg based on metric type

### SLO Gate
- [ ] Sum check validation with difference < 2%
- [ ] Integration tests pass with various process distributions

## Phase 5: L3 - ReservoirSampler Processor - "A Glimpse of the Long Tail"
**Timeline: Week 9-10**

### Goal
Implement sampling for a representative subset of processes not covered by L0, L1, or L2.

### Key Activities
- [ ] **Configuration Design**
  - [ ] Define YAML configuration (reservoir_size, identity_attributes, time_window)
  - [ ] Create JSON schema for validation

- [ ] **Implementation**
  - [ ] Create `processors/reservoirsampler/` directory structure
  - [ ] Implement Config and Factory structs
  - [ ] Implement Algorithm R (or variant) for reservoir sampling
  - [ ] Add `nr.process_sampled_by_reservoir="true"` and `nr.sample_rate` attributes
  - [ ] Integrate with internal/obsreport wrapper
  - [ ] Add custom self-metrics: `nrdot_reservoirsampler_fill_ratio` and `nrdot_reservoirsampler_selected_identities_count`

- [ ] **Testing**
  - [ ] Unit tests for configuration validation
  - [ ] Unit tests for sampling logic and reservoir filling
  - [ ] Statistical tests for sampling uniformity
  - [ ] Integration tests with full processor chain

- [ ] **Integration**
  - [ ] Register in main.go
  - [ ] Update opt-plus.yaml to include reservoirsampler after adaptivetopk
  - [ ] Design processor order to ensure proper interaction between OthersRollup and ReservoirSampler

- [ ] **Observability**
  - [ ] Add Grafana panels for reservoir fill ratio and sampled identity count
  - [ ] Add NRQL query examples for sampled processes

- [ ] **Documentation**
  - [ ] README.md for reservoirsampler with sampling details
  - [ ] Update main documentation to include L3 processor
  - [ ] Document interaction between sampling and rollup

### Deliverables
- NRDOT collector image with complete L0-L3 processor chain
- Updated dashboards showing sampled processes

### Validation Criteria
- [ ] Approximately `reservoir_size` PIDs are exported with the `nr.process_sampled_by_reservoir` attribute
- [ ] `nr.sample_rate` attribute is present and accurate
- [ ] Self-metrics correctly reflect sampler activity

### SLO Gate
- [ ] Chi^2 uniform p > 0.05 (if tested statistically)
- [ ] Self-metric `nr.sample_rate` shows value ≈ (reservoir_size / eligible_for_sampling_count)

## Phase 6: Full Opt-Plus Pipeline Integration & Cold Path - "The Full Picture"
**Timeline: Week 11-12**

### Goal
Integrate all L0-L3 processors into the opt-plus.yaml hot path and implement the optional cold path for raw data.

### Key Activities
- [ ] **Pipeline Configuration**
  - [ ] Finalize opt-plus.yaml with hot path (L0-L3 processors)
  - [ ] Define cold path (minimal processing, raw data)
  - [ ] Implement JSON Schema validation for opt-plus.yaml

- [ ] **Cold Path Implementation**
  - [ ] Configure/implement S3/File exporter for cold path
  - [ ] Ensure Parquet format with correct schema
  - [ ] Parameterize S3 bucket, prefix, credentials

- [ ] **Testing**
  - [ ] End-to-end tests for full hot path
  - [ ] End-to-end tests for cold path
  - [ ] Validate no data loss between paths
  - [ ] Performance tests for both paths

- [ ] **Integration**
  - [ ] Create nrdot-configgen utility/option
  - [ ] Update Docker Compose for dual-path testing

- [ ] **Observability**
  - [ ] Add Grafana panels for both hot and cold path metrics
  - [ ] Add dashboard for analyzing cardinality reduction

- [ ] **Documentation**
  - [ ] Document opt-plus.yaml configuration options
  - [ ] Update main documentation with dual-path architecture
  - [ ] Create cold path setup guide

### Deliverables
- Complete opt-plus.yaml with hot and cold paths
- Functional Parquet/S3 cold path export
- nrdot-configgen utility

### Validation Criteria
- [ ] Hot path to New Relic meets cardinality targets (≤ 80-100 unique process series)
- [ ] Cold path correctly writes all raw process metrics
- [ ] No data loss or duplication between paths

### SLO Gate
- [ ] 10 GiB/day host → S3 (mock), zero back-pressure
- [ ] Parquet schema validation passes

## Phase 7: nrdot-doctor CLI & Hardening - "Polish and Production Readiness"
**Timeline: Week 13-14**

### Goal
Develop the nrdot-doctor CLI for diagnostics and harden the solution for production use.

### Key Activities
- [ ] **nrdot-doctor CLI**
  - [ ] Create `cmd/doctor/` directory structure
  - [ ] Implement permissions check for PID scrape
  - [ ] Add config validation against schema
  - [ ] Add self-metrics query/analysis
  - [ ] Implement health check recommendations

- [ ] **Security Hardening**
  - [ ] Configure distroless runtime image
  - [ ] Generate and validate SBOM
  - [ ] Implement vulnerability scanning in CI
  - [ ] Configure non-root user (UID 10001)

- [ ] **Performance Optimization**
  - [ ] Run benchmarks for critical code paths
  - [ ] Profile under heavy load
  - [ ] Optimize memory usage and garbage collection

- [ ] **Documentation Finalization**
  - [ ] Complete all README files
  - [ ] Create CONTRIBUTING.md
  - [ ] Update CHANGELOG.md
  - [ ] Create architecture diagrams
  - [ ] Document all metrics and attributes

- [ ] **Dashboard Finalization**
  - [ ] Polish New Relic One dashboards
  - [ ] Finalize Grafana dashboards with alerting thresholds

### Deliverables
- nrdot-doctor CLI
- Hardened collector image
- Complete documentation set

### Validation Criteria
- [ ] nrdot-doctor correctly identifies permissions issues
- [ ] Security scans pass with no critical findings
- [ ] Performance benchmarks meet targets
- [ ] Documentation is complete and accurate

### SLO Gate
- [ ] nrdot-doctor fails if PID scrape is blocked
- [ ] Trivy scan shows 0 critical vulnerabilities

## Phase 8: Helm Chart & GA - "Deploy with Confidence"
**Timeline: Week 15-16**

### Goal
Package as a production-ready Helm chart and release v1.0.0.

### Key Activities
- [ ] **Helm Chart Development**
  - [ ] Create `charts/nrdot-hostmetrics-optimized/` directory structure
  - [ ] Parameterize all opt-plus.yaml settings
  - [ ] Configure resource limits and security contexts
  - [ ] Add cold store configuration options
  - [ ] Implement Helm test suite

- [ ] **Kubernetes Testing**
  - [ ] Test on EKS
  - [ ] Test on GKE
  - [ ] Test on AKS

- [ ] **Release Preparation**
  - [ ] Prepare OCI chart for publication
  - [ ] Generate final SBOMs
  - [ ] Create release notes
  - [ ] Tag v1.0.0

- [ ] **Documentation**
  - [ ] Create Helm chart documentation
  - [ ] Document Kubernetes deployment patterns
  - [ ] Create upgrade guide

### Deliverables
- Production-ready Helm chart
- OCI image published to registry
- Complete release artifacts (SBOMs, notes)
- v1.0.0 tag

### Validation Criteria
- [ ] `helm install` and `helm test` succeed on all target platforms
- [ ] All GA SLOs from previous phases are met
- [ ] Upgrade path from previous versions is validated

### SLO Gate
- [ ] helm test passes on EKS, GKE, AKS
- [ ] All documentation is complete and accurate

## Risk Management

| Risk ID | Description | Likelihood | Impact | Mitigation |
|---------|-------------|------------|--------|------------|
| R1 | Process scraping permissions blocked | Medium | High | nrdot-doctor preflight + documented requirements |
| R2 | Top-K selection becomes performance bottleneck | Medium | Medium | Benchmark early, optimize heap implementation |
| R3 | Aggregation math causes double-counting | Low | High | Golden fixture tests, extensive validation |
| R4 | Cold path export fails under high load | Medium | Medium | Batch export, backpressure testing |
| R5 | Security vulnerabilities in dependencies | Medium | High | Regular dependency updates, vulnerability scanning |
| R6 | Configuration drift between phases | Medium | Low | JSON schema validation, automated testing |
| R7 | Backward compatibility breaks | Low | Medium | Versioned configurations, thorough testing |

## Success Metrics

| Metric | Target | Validation Method |
|--------|--------|------------------|
| Process Metric Cardinality | ≤ 80-100 series per host | New Relic NRQL queries |
| Optimization Overhead | < 5% CPU, < 250MB memory | Performance profiling |
| Critical Process Coverage | 100% of defined critical processes | Integration tests |
| Sum Check Accuracy | < 2% difference | End-to-end tests |
| Reservoir Sample Statistical Validity | Chi^2 p > 0.05 | Statistical tests |
| Security Posture | 0 critical vulnerabilities | Automated scanning |

This plan provides a detailed roadmap from the initial hello-world processor to a production-grade, optimized process metrics pipeline. Each phase builds on the previous one, providing incremental value and validation along the way.