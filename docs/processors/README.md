# Processor Documentation

This directory contains detailed documentation for each of the custom processors in the NRDOT Process-Metrics Optimization pipeline.

## Processors

- [PriorityTagger (L0)](prioritytagger.md) - Tags critical processes for preservation in subsequent pipeline stages
- [AdaptiveTopK (L1)](adaptivetopk.md) - Selects the K most resource-intensive processes
- [OthersRollup (L2)](othersrollup.md) - Aggregates non-critical, non-TopK processes into a single "_other_" series
- [ReservoirSampler (L3)](reservoirsampler.md) - Samples a statistically representative subset of processes

## Processor Pipeline Order

The processors are designed to work together in a specific sequence:

```
┌───────────────┐    ┌─────────────────┐    ┌────────────────┐    ┌──────────────────┐    ┌────────────────┐
│ HOSTMETRICS   │    │ PRIORITYTAGGER  │    │ ADAPTIVETOPK   │    │ RESERVOIRSAMPLER │    │ OTHERSROLLUP   │
│ RECEIVER      │───>│ PROCESSOR (L0)  │───>│ PROCESSOR (L1) │───>│ PROCESSOR (L3)   │───>│ PROCESSOR (L2) │──> EXPORTERS
│ Process Data  │    │ Tag Critical    │    │ Keep Top K     │    │ Sample Long-Tail │    │ Aggregate Rest │
└───────────────┘    └─────────────────┘    └────────────────┘    └──────────────────┘    └────────────────┘
```

Each processor documentation includes:
- Configuration options
- Operation details
- Self-observability metrics
- Usage examples