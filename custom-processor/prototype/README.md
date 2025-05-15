# OpenTelemetry Hello World Processor Prototype

This is a simple "Hello World" prototype for an OpenTelemetry processor that demonstrates the basic lifecycle and structure of a processor before integrating it into an OpenTelemetry Collector.

## Overview

This prototype implements a minimal processor that logs messages when:
1. The processor starts
2. The processor receives telemetry items to process
3. The processor shuts down

This simplified implementation focuses on the processor lifecycle rather than complex processing logic.

## Processor Lifecycle

In OpenTelemetry, processors typically follow this lifecycle:

1. **Initialization**: The processor is created with configuration parameters
2. **Start**: Called when the processor is starting up, used for setup operations
3. **Processing**: The processor receives and processes telemetry data
4. **Shutdown**: Called when the processor is shutting down, used for cleanup

This prototype demonstrates these stages with simple logging.

## Running the Prototype

```bash
cd prototype
go mod tidy
go run main.go
```

This will:
1. Create a hello world processor with a custom message
2. Start the processor
3. Simulate processing three telemetry items
4. Shut down the processor

## Integration with OpenTelemetry

To integrate this into a real OpenTelemetry Collector processor:

1. The processor would need to be implemented using the OpenTelemetry SDK
2. The processor would implement the required interfaces (ProcessorFactory, etc.)
3. The processor would properly handle the telemetry data model (traces, metrics, logs)
4. The processor would be configured via YAML in the collector configuration

## Next Steps

1. Expand the processor to handle real trace/metric/log data
2. Implement the necessary interfaces for OpenTelemetry integration
3. Add configuration options
4. Test with the OpenTelemetry Collector

## References

- [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector)
- [Creating Custom Processors](https://opentelemetry.io/docs/collector/custom-components/)
- [OpenTelemetry Go SDK](https://github.com/open-telemetry/opentelemetry-go)