# Processor Helper

This internal package provides a small helper that all processors can embed for
common functionality. It embeds `component.TelemetrySettings` for access to
logging and metrics and exposes a shared obsreport helper that tracks the number
of processed and dropped metric points.

The helper also implements default `Start`, `Shutdown` and `Capabilities`
methods so individual processors only need to implement `ConsumeMetrics` and
processor specific logic.

### Usage

```go
import "github.com/newrelic/nrdot-process-optimization/internal/processorhelper"

// In your processor constructor
helper, err := processorhelper.NewHelper(settings.TelemetrySettings, "myprocessor")
if err != nil {
    return nil, err
}

p := &myProcessor{
    Helper: helper,
    // ... other fields
}
```

In your processor implementation, use `StartMetricsOp` and `EndMetricsOp` around
your processing logic:

```go
ctx = p.StartMetricsOp(ctx)
// perform mutations
p.EndMetricsOp(ctx, processedCount, nil)
```
