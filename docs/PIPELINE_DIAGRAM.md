# NRDOT Process-Metrics Optimization Pipeline Diagrams

This document provides visual representations of the optimization pipeline to help understand the data flow and transformations.

## Basic Pipeline Flow

The following diagram shows the basic flow of data through the optimization pipeline:

```mermaid
flowchart LR
    A[Hostmetrics\nReceiver] --> B[PriorityTagger\nL0]
    B --> C[AdaptiveTopK\nL1]
    C --> D[ReservoirSampler\nL3]
    D --> E[OthersRollup\nL2]
    E --> F[Batch\nProcessor]
    F --> G[OTLP\nExporter]

    style A fill:#f9f9f9,stroke:#333,stroke-width:2px
    style B fill:#d4f0ff,stroke:#0078d4,stroke-width:2px
    style C fill:#ddf4dd,stroke:#107c10,stroke-width:2px
    style D fill:#f7e8d5,stroke:#c57c38,stroke-width:2px
    style E fill:#f1dfff,stroke:#881798,stroke-width:2px
    style F fill:#f9f9f9,stroke:#333,stroke-width:2px
    style G fill:#f9f9f9,stroke:#333,stroke-width:2px
```

## Data Transformation Details

This diagram shows how the data is transformed at each step, with approximate cardinality reduction:

```mermaid
flowchart TD
    subgraph Input
        A[Raw Process\nMetrics\n100% Cardinality]
    end

    subgraph "L0: PriorityTagger"
        B1[Tag Critical\nProcesses]
    end

    subgraph "L1: AdaptiveTopK"
        C1[Keep Top K +\nCritical Processes]
        C2[Drop Other\nProcesses]
    end

    subgraph "L3: ReservoirSampler"
        D1[Sample Statistically\nRepresentative\nProcesses]
        D2[Drop Remaining\nProcesses]
    end

    subgraph "L2: OthersRollup"
        E1[Aggregate Non-Priority\nNon-TopK\nNon-Sampled]
        E2[Create '_other_'\nRollup Series]
    end

    subgraph Output
        F1[Critical\nProcesses\n~5-15%]
        F2[Top K\nProcesses\n~5-10%]
        F3[Sampled\nProcesses\n~1-5%]
        F4["_other_"\nRollup\n1 series]
    end

    A --> B1
    B1 --> C1
    C1 --> D1
    D1 --> E1

    C1 --> C2
    D1 --> D2

    E1 --> E2

    B1 --> F1
    C1 --> F2
    D1 --> F3
    E2 --> F4

    style A fill:#f9f9f9,stroke:#333,stroke-width:2px
    style B1 fill:#d4f0ff,stroke:#0078d4,stroke-width:2px
    style C1 fill:#ddf4dd,stroke:#107c10,stroke-width:2px
    style C2 fill:#ffd7d7,stroke:#d13438,stroke-width:2px
    style D1 fill:#f7e8d5,stroke:#c57c38,stroke-width:2px
    style D2 fill:#ffd7d7,stroke:#d13438,stroke-width:2px
    style E1 fill:#f1dfff,stroke:#881798,stroke-width:2px
    style E2 fill:#f1dfff,stroke:#881798,stroke-width:2px
    style F1 fill:#d4f0ff,stroke:#0078d4,stroke-width:2px
    style F2 fill:#ddf4dd,stroke:#107c10,stroke-width:2px
    style F3 fill:#f7e8d5,stroke:#c57c38,stroke-width:2px
    style F4 fill:#f1dfff,stroke:#881798,stroke-width:2px
```

## Cardinality Reduction Effect

This visualization shows the approximate cardinality reduction achieved by the pipeline:

```mermaid
pie
    title "Process Metrics Cardinality Breakdown"
    "Critical Processes" : 15
    "Top K Processes" : 10
    "Sampled Processes" : 5
    "'_other_' Rollup" : 1
    "Dropped Metrics" : 69
```

## Per-Process Attribute Effect

The table below shows how the pipeline affects the attributes of different process metrics:

| Process Type | PriorityTagger | AdaptiveTopK | ReservoirSampler | OthersRollup | Final Output |
|-------------|----------------|--------------|------------------|--------------|--------------|
| **Critical** | `nr.priority="critical"` | Passed through | Passed through | Passed through | All original metrics retained |
| **Top K** | Unchanged | Selected | Passed through | Passed through | All original metrics retained |
| **Sampled** | Unchanged | Dropped | `nr.process_sampled_by_reservoir="true"` + `nr.sample_rate="0.xx"` | Passed through | Sampled metrics with sampling metadata |
| **Others** | Unchanged | Dropped | Dropped | Aggregated into `process.pid="-1"` + `process.executable.name="_other_"` | Single aggregated series (sum or avg) |

This approach allows for significant metric volume reduction while maintaining visibility into both critical and representative processes.

## Note on Viewing Mermaid Diagrams

These diagrams use Mermaid syntax which may require a compatible Markdown viewer. In GitHub, they should render automatically. In other environments, you might need a Mermaid plugin or viewer.

You can also paste the Mermaid code into the [Mermaid Live Editor](https://mermaid.live/) to view and edit the diagrams.