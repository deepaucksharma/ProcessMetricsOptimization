# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

This document provides specific context and guidelines for AI coding assistants when working with the **NRDOT Process-Metrics Optimization** repository. Its aim is to help the AI understand the project's goals, structure, current state, and conventions, leading to more effective and aligned assistance.

---

## 1. Project Core Objective & Current Status

* **Primary Goal:** To build a custom OpenTelemetry (OTel) Collector distribution that significantly reduces process metric data volume (aiming for ≥90% reduction) for New Relic, while preserving essential visibility into host processes.
* **Mechanism:** A multi-layer pipeline of custom OTel processors (L0-L3) that progressively Tag, Filter, Aggregate, and Sample process metrics.
* **Current Status:**
  * **Phase 0 - Complete:**
    * A foundational "Hello World" (`helloworld`) custom processor is implemented and functional.
    * A robust local development environment using Docker Compose is established, including the custom Collector, Prometheus, Grafana, and a Mock OTLP Sink.
    * Standardized ports and service interactions are defined.
    * Core documentation (`README.md`, this `CLAUDE.md`, `IMPLEMENTATION_PLAN.md`, processor development guides) is in place.
    * CI for build, lint, unit tests, and basic vulnerability checks is operational.
  * **Phase 1 - Complete:**
    * The L0 PriorityTagger (`prioritytagger`) processor is fully implemented.
    * It supports tagging critical processes by name, regex pattern, CPU utilization, and memory usage.
    * Handles both integer and double metric value types for proper threshold comparison.
    * Comprehensive unit tests and benchmarks are in place.
    * The processor is integrated into the custom collector and verified to be processing metrics.
    * Standard metric `otelcol_otelcol_processor_prioritytagger_processed_metric_points` confirms processing activity.
  * **Phase 2 - Complete:**
    * The L1 AdaptiveTopK (`adaptivetopk`) processor is fully implemented.
    * It selects metrics from the K most resource-intensive processes.
    * Uses an efficient min-heap algorithm to identify highest-resource processes.
    * Provides foundation for future dynamic K adjustment based on host load.
  * **Phase 3 - Complete:**
    * The L2 OthersRollup (`othersrollup`) processor is fully implemented.
    * Aggregates metrics from non-priority, non-TopK processes into a single summary series.
    * Supports flexible aggregation strategies (sum, average) per metric type.
    * Preserves metric type semantics (gauge vs. sum).
  * **Phase 4 - Complete:**
    * The L3 ReservoirSampler (`reservoirsampler`) processor is fully implemented.
    * Selects a statistically representative sample of metrics from "long-tail" processes.
    * Implements Algorithm R variant for reservoir sampling.
    * Adds sample rate metadata to enable proper scaling during analysis.
* **Phase 5 - Complete:**
    * The entire optimization pipeline is fully integrated and validated.
    * All processors (L0-L3) are working together in sequence.
    * Comprehensive dashboard metrics and visualizations are implemented.
    * The pipeline is fully operational and ready for production use.

---

## 2. Key Project Files & Directory Structure

Familiarize yourself with this structure to locate relevant code and configurations:

```
📁 nrdot-process-optimization/
├── .github/                            # GitHub Actions CI/CD Workflows
├── build/                              # Docker & Docker Compose files
│   ├── Dockerfile                      # Builds the custom collector image
│   └── docker-compose.yaml             # Orchestrates the local dev stack
├── cmd/collector/                      # Main entry point for the custom OTel Collector
│   └── main.go                         # Registers all components (processors, receivers, etc.)
├── config/                             # Collector configuration files
│   └── opt-plus.yaml                   # Config for the full L0-L3 optimization pipeline
├── dashboards/                         # Grafana dashboard JSON files
│   └── grafana-nrdot-custom-processor-starter-kpis.json
├── docs/                               # Developer documentation and guides
│   ├── DEVELOPING_PROCESSORS.md        # How to build new custom processors
│   ├── NRDOT_PROCESSOR_SELF_OBSERVABILITY.md # Standards for processor metrics
│   ├── GRAFANA_DASHBOARD_DESIGN.md     # Detailed guide for advanced Grafana dashboards
│   └── OBSERVABILITY_STACK_SETUP.md    # Guide for setting up Prometheus, Grafana with the collector
├── examples/                           # Standalone example code
│   └── README.md                       # Describes planned examples for the project
├── internal/                           # Internal shared packages
│   └── banding/                        # Support for AdaptiveTopK processor implementation
│       └── README.md                   # Documents planned functionality for banding package
├── processors/                         # Location for all custom OTel processors
│   └── helloworld/                     # Phase 0: Example "Hello World" processor
│       ├── config.go                   # Configuration struct and validation
│       ├── factory.go                  # Processor factory implementation
│       ├── obsreport.go                # Helper for obsreport (can be integrated directly)
│       ├── processor.go                # Core processor logic (ConsumeMetrics, etc.)
│       └── processor_test.go           # Unit tests for the processor
│   └── prioritytagger/                 # Phase 1: L0 Processor - Tag critical processes
│       ├── config.go                   # Configuration with critical process criteria
│       ├── factory.go                  # Factory implementation
│       ├── obsreport.go                # Metrics for processor monitoring
│       ├── processor.go                # Logic for identifying and tagging critical processes
│       ├── processor_test.go           # Unit tests
│       ├── integration_test.go         # Pipeline integration tests
│       ├── benchmark_test.go           # Performance benchmark tests
│       └── README.md                   # Processor documentation
│   └── adaptivetopk/                   # Phase 2: L1 Processor - Select top K resource-intensive processes
│       ├── config.go                   # Configuration for TopK selection
│       ├── factory.go                  # Factory implementation
│       ├── obsreport.go                # Metrics for processor monitoring
│       ├── processor.go                # TopK selection logic
│       └── processor_test.go           # Unit tests
│   └── othersrollup/                   # Phase 3: L2 Processor - Aggregate non-critical processes
│       ├── config.go                   # Configuration for aggregation
│       ├── factory.go                  # Factory implementation
│       ├── obsreport.go                # Metrics for processor monitoring
│       ├── processor.go                # Aggregation logic
│       └── processor_test.go           # Unit tests
│   └── reservoirsampler/               # Phase 4: L3 Processor - Sample remaining processes
│       ├── config.go                   # Configuration for sampling
│       ├── factory.go                  # Factory implementation
│       ├── obsreport.go                # Metrics for processor monitoring
│       ├── processor.go                # Sampling logic
│       └── processor_test.go           # Unit tests
├── test/                               # Test suites and helper scripts
│   └── url_check.sh                    # Checks local dev stack service URLs
├── Makefile                            # Central build and task automation script
├── go.mod / go.sum                     # Go module definitions
├── README.md                           # Main project overview and quick start
├── IMPLEMENTATION_PLAN.md              # Phased development roadmap
└── CLAUDE.md                           # This file
```

(Parentheses () denote components planned for future phases.)

[Rest of the file remains the same as before]