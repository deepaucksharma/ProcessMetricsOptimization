# Architecture Documentation

This directory contains documentation related to the overall architecture of the NRDOT Process-Metrics Optimization pipeline.

## Contents

- [Pipeline Overview](pipeline_overview.md) - Comprehensive documentation of the complete optimization pipeline
- [Pipeline Diagram](pipeline_diagram.md) - Visual representations of the pipeline flow and data transformations
- [Metrics Schema](metrics_schema.md) - Details about the metrics data model and schema

## Key Concepts

The NRDOT optimization pipeline is a multi-layered approach to reducing process metrics cardinality in OpenTelemetry data, designed to achieve >90% reduction in data volume while preserving visibility into critical system processes.

The pipeline consists of four custom processors that work together in sequence:

1. **L0: PriorityTagger** - Identifies and tags critical processes that must be preserved regardless of optimization
2. **L1: AdaptiveTopK** - Selects metrics from the K most resource-intensive processes, with dynamic K adjustment 
3. **L3: ReservoirSampler** - Statistically samples a representative subset of non-critical, non-top processes
4. **L2: OthersRollup** - Aggregates all remaining processes into a single "_other_" process series