# Developing NRDOT Custom Processors

This guide explains how to develop custom OpenTelemetry processors for the NRDOT Process-Metrics Optimization project.

## Processor Development Workflow

1. **Copy the Hello World template**: Start by copying the basic structure from the `processors/helloworld` directory
2. **Implement your logic**: Modify the copied code to implement your specific processor logic
3. **Add self-observability**: Ensure your processor has proper instrumentation
4. **Write tests**: Add comprehensive unit tests for your processor
5. **Test locally**: Run your processor locally to verify it works as expected
6. **Create a PR**: Submit your changes for review

## Basic Structure of a Processor

Every processor implementation consists of these key files:

- `config.go` - Configuration definition and validation
- `factory.go` - Factory for creating the processor
- `processor.go` - Main processor implementation
- `obsreport.go` - Observability helpers
- `processor_test.go` - Unit tests

## Self-Observability

The Hello World processor demonstrates the core self-observability patterns:

1. **OTel obsreport**: For standard OpenTelemetry processor metrics
   ```go
   // Create the obsreport helper
   obsrecv, err := newObsreportHelper(settings)
   if err != nil {
       return nil, err
   }
   
   // Start metrics observation
   ctx, numPoints := p.obsrecv.StartMetricsOp(ctx)
   
   // End metrics observation
   p.obsrecv.EndMetricsOp(ctx, p.config.ProcessorType(), metricCount, nil)
   ```

2. **Custom metrics**: For processor-specific KPIs
   ```go
   // Create a metric in the constructor
   meter := settings.MeterProvider.Meter("helloworld")
   mutationsCounter, err := meter.Int64Counter(
       "nrdot_helloworld_mutations_total",
       metric.WithDescription("Total number of metrics modified"),
   )
   
   // Use the metric in your code
   p.mutationsCounter.Add(ctx, int64(metricCount))
   ```

## Viewing Your Processor's Metrics

You can view your processor's self-metrics through:

1. **zPages**: Browse to http://localhost:55679 when running locally
2. **Prometheus**: Browse to http://localhost:9090 when running locally
3. **Grafana**: Browse to http://localhost:3000 when running locally
   - Use the pre-loaded "NRDOT Custom Processor Starter KPIs" dashboard

## Best Practices

1. **Naming conventions**: Use consistent naming for metrics and configuration
   - Custom metrics should be prefixed with `nrdot_<processor_name>_`
   - Configuration options should use snake_case
   
2. **Error handling**: Always check errors and gracefully recover from failures
   - Consider using `obsreport.EndMetricsOp()` with error to record failures

3. **Performance**: Be careful with memory allocations and heavy computations
   - Process-metrics pipelines can be memory and CPU intensive
   - Consider implementing a `WithMemoryLimiter()` option if processing large amounts of data

4. **Testing**: Write comprehensive tests for your processor
   - Test normal operation
   - Test error cases
   - Test edge cases (empty metrics, large numbers of metrics, etc.)

5. **Documentation**: Document your processor's behavior, configuration options, and metrics
   - Keep comments in your code up to date
   - Use descriptive variable names
   - Add info about new metrics to this document
   
## Processor Checklist

Before submitting a PR for review, check that your processor:

- [ ] Has clearly defined configuration with validation
- [ ] Handles all relevant metric types
- [ ] Emits standard obsreport metrics
- [ ] Implements custom metrics for processor-specific KPIs
- [ ] Includes comprehensive unit tests
- [ ] Has clear documentation
- [ ] Passes linting and static analysis
- [ ] Works with the Docker Compose setup
- [ ] Has a dashboard for its metrics
