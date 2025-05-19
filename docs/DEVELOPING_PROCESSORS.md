# Developing NRDOT Custom Processors

This guide explains how to develop custom OpenTelemetry processors for the NRDOT Process-Metrics Optimization project.

## Quick Start

1. Copy the basic structure from an existing processor (see "Processor Examples" below)
2. Implement your specific processor logic
3. Add self-observability instrumentation
4. Write comprehensive unit tests
5. Test locally with the Docker Compose setup
6. Create a PR with documentation

## Processor Examples

The project includes several processors that can serve as examples:

| Processor | Purpose | Key Features | Location |
|-----------|---------|--------------|----------|
| **HelloWorld** | Demonstration | Attribute modification, basic instrumentation | `processors/helloworld/` |
| **PriorityTagger** | Tag critical processes | Pattern matching, resource thresholds | `processors/prioritytagger/` |
| **AdaptiveTopK** | Select top processes | Min-heap selection algorithm | `processors/adaptivetopk/` |
| **OthersRollup** | Aggregate metrics | Multiple aggregation strategies | `processors/othersrollup/` |
| **ReservoirSampler** | Statistical sampling | Reservoir algorithm, sampling metadata | `processors/reservoirsampler/` |

Choose the example that most closely matches your needs as a starting point.

## Processor Structure

Every processor consists of these key files:

| File | Purpose |
|------|---------|
| `config.go` | Configuration definition and validation |
| `factory.go` | Factory for creating the processor |
| `processor.go` | Main processor implementation |
| `obsreport.go` | Observability helpers |
| `processor_test.go` | Unit tests |
| `README.md` | Documentation for the processor |

## Aligning with OpenTelemetry

NRDOT processors follow the standard [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) design. Each processor implements the `component.Processor` interface and is created through a factory. The factory parses configuration into a `Config` struct and exposes `CreateMetricsProcessor` (and/or `CreateLogsProcessor`) for the collector to call.

Typical lifecycle:

1. **Factory creates the processor** – configuration is validated and an instance is returned.
2. **Start** – the processor allocates resources when the collector pipeline starts.
3. **Consume** – metrics or logs are processed for each call.
4. **Shutdown** – resources are released on collector shutdown.

For complex matching or transformation logic you can use the [OpenTelemetry Transformation Language (OTTL)](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/pkg/ottl). OTTL expressions allow attribute-based filters or modifications without additional code.

See the upstream [Collector documentation](https://opentelemetry.io/docs/collector/components/#processors) for further reference.

## Self-Observability Implementation

### Standard Metrics with obsreport

```go
// Create the obsreport helper
obsrecv, err := newObsreportHelper(settings)
if err != nil {
    return nil, err
}

// Start metrics observation
ctx = p.obsrecv.StartMetricsOp(ctx)

// End metrics observation
p.obsrecv.EndMetricsOp(ctx, metricCount, 0, nil)
```

### Custom Processor KPIs

```go
// Create a metric in the constructor
meter := settings.MeterProvider.Meter("helloworld")
mutationsCounter, err := meter.Int64Counter(
    "otelcol_otelcol_helloworld_mutations_total",
    metric.WithDescription("Total number of metrics modified"),
)

// Use the metric in your code
p.mutationsCounter.Add(ctx, int64(metricCount))
```

## Local Development Loop

### Basic Processor Development

1. Implement/modify your processor
2. Run `make docker-build && make compose-up`
3. Access observability tools:
   - **zPages**: http://localhost:15679
   - **Prometheus**: http://localhost:19090
   - **Grafana**: http://localhost:13000 (Use the "NRDOT Processors - HelloWorld & PriorityTagger KPIs" dashboard)
4. View logs with `make logs`
5. Refine implementation and repeat

### Full Pipeline Integration

To test your processor as part of the complete optimization pipeline:

1. Make sure your processor is registered in `cmd/collector/main.go`
2. Add your processor configuration to `config/opt-plus.yaml`
3. Run the full pipeline: `make docker-build && make opt-plus-up`
4. Verify processor interactions: `make logs`
5. Check metrics in Prometheus/Grafana
6. Run the automated pipeline test: `./test/test_opt_plus_pipeline.sh`

See [docs/COMPLETING_PHASE_5.md](COMPLETING_PHASE_5.md) for additional integration testing guidance.

## Best Practices

| Category | Recommendations |
|----------|----------------|
| **Naming Conventions** | - Prefix custom metrics with `nrdot_<processor_name>_`<br>- Use snake_case for configuration options |
| **Error Handling** | - Check all errors<br>- Use `obsreport.EndMetricsOp()` with error to record failures |
| **Performance** | - Minimize memory allocations<br>- Consider `WithMemoryLimiter()` for large data processing |
| **Testing** | - Test normal operation, error cases, and edge cases<br>- Add benchmark tests for performance-sensitive processors |
| **Documentation** | - Document behavior, configuration options, and metrics<br>- Keep code comments up to date |

## Pre-PR Checklist

- [ ] Configuration is well-defined with validation
- [ ] Handles all relevant metric types
- [ ] Emits standard obsreport metrics
- [ ] Implements custom metrics for processor-specific KPIs
- [ ] Includes comprehensive unit tests
- [ ] Has clear README.md documentation
- [ ] Passes linting and static analysis
- [ ] Works with the Docker Compose setup
- [ ] Adds dashboard panels for processor metrics

## Aligning with OpenTelemetry

Custom processors follow the standard `component.Processor` lifecycle defined by
the OpenTelemetry Collector. A processor is created by a factory, started by the
collector, receives telemetry via its `Consume` method, and is finally shut down
when the collector stops.

### Lifecycle and Factory Pattern

1. **Factory** – Each processor exposes a `NewFactory` function that returns a
   `processor.Factory`. The factory provides `createDefaultConfig` and a
   `create*Processor` function used by the collector to instantiate the
   component.
2. **Start/Shutdown** – Implement the `Start` and `Shutdown` methods from
   `component.Component` to allocate resources and clean them up.
3. **Consume** – Implement `ConsumeMetrics`, `ConsumeTraces`, or `ConsumeLogs`
   depending on the processor type. Use `Capabilities()` to declare whether the
   processor mutates data.

### Attribute Matching with OTTL

The **OpenTelemetry Transformation Language (OTTL)** provides a flexible way to
express attribute-based matching or transformations without hard‑coding logic in
Go. Processors can leverage OTTL statements to match on attributes or modify
telemetry in a declarative fashion.

For more details on the processor lifecycle and OTTL syntax, see the
[upstream OpenTelemetry documentation](https://github.com/open-telemetry/opentelemetry-collector/blob/main/processor/README.md).
