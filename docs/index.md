# NRDOT Process-Metrics Optimization Documentation

Welcome to the documentation for the NRDOT Process-Metrics Optimization project. This documentation hub provides comprehensive information about the project's architecture, components, development guides, and operational procedures.

## Documentation Structure

### Architecture
- [Pipeline Overview](architecture/pipeline_overview.md) - High-level overview of the optimization pipeline
- [Pipeline Diagram](architecture/pipeline_diagram.md) - Visual representation of the pipeline flow
- [Metrics Schema](architecture/metrics_schema.md) - Details about the metrics data model

### Processors
- [PriorityTagger (L0)](processors/prioritytagger.md) - Tags critical processes
- [AdaptiveTopK (L1)](processors/adaptivetopk.md) - Selects top K resource-intensive processes
- [OthersRollup (L2)](processors/othersrollup.md) - Aggregates non-critical processes
- [ReservoirSampler (L3)](processors/reservoirsampler.md) - Samples representative processes

### Development
- [Developing Processors](development/developing_processors.md) - Guide for creating new processors
- [Processor Self-Observability](development/processor_self_observability.md) - Standards for processor metrics
- [Metric Naming Conventions](development/metric_naming_conventions.md) - Naming standards for metrics
- [Implementation Plan](development/implementation_plan.md) - Phased development roadmap

### Operations
- [Observability Stack Setup](operations/observability_stack_setup.md) - Setting up monitoring infrastructure
- [Dashboard Metrics Audit](operations/dashboard_metrics_audit.md) - Auditing metrics in dashboards
- [Completing Phase 5](operations/completing_phase_5.md) - Final integration steps

### Dashboards
- [Grafana Dashboard Design](dashboards/grafana_dashboard_design.md) - Dashboard design principles
- [Dashboard Overview](dashboards/dashboard_overview.md) - Overview of available dashboards

## Quick Links

- [Main README](../README.md) - Project overview and quick start
- [Implementation Plan](development/implementation_plan.md) - Detailed development roadmap
- [Full Pipeline Documentation](architecture/pipeline_overview.md) - Comprehensive pipeline details

## Contributing

For guidance on contributing to this project, please refer to the [Developing Processors](development/developing_processors.md) guide.