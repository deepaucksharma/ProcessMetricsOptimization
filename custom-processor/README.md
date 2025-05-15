# Trace-Aware Reservoir OpenTelemetry Configuration

This directory contains a streamlined OpenTelemetry configuration that adds trace-aware attributes to metrics. It provides a ready-to-use setup for sending metrics with enhanced context to New Relic.

## 🔧 Features

- Simulated trace-aware processor using native attribute processor
- Adds trace-aware attributes at both resource and datapoint levels
- Uses official OpenTelemetry Collector Contrib Docker image
- Includes comprehensive host metrics collection
- Ready-to-use Docker containerization

## 🚀 Quick Start with Docker

The easiest way to run the collector is using the provided run script:

```bash
# Set your New Relic license key
export NEW_RELIC_LICENSE_KEY=your_license_key

# Run the collector
chmod +x run.sh
./run.sh
```

The script will:
1. Pull the OpenTelemetry Collector Contrib Docker image
2. Run it with our trace-aware configuration
3. Stream logs to the console

## 🔌 Configuration Options

You can customize the collector with these environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `NEW_RELIC_LICENSE_KEY` | (required) | Your New Relic license key |
| `OTEL_DEPLOYMENT_ENVIRONMENT` | production | Environment label for your metrics |
| `OTEL_LOG_LEVEL` | info | Log level (debug, info, warn, error) |

## 🔍 How It Works

Instead of a custom Go processor, this implementation uses the built-in attributes processor to add trace-aware context:

### Resource-level attributes:
- `collector.pipeline`: Set to "newrelic" to identify the pipeline
- `processor.traceaware.version`: Set to "1.0.0" for versioning
- `service.name`: Identifies the service generating metrics
- `service.instance.id`: Unique identifier for the collector instance

### Datapoint-level attributes:
- `processor.traceaware`: Set to "processed" on all datapoints

This approach achieves the same result as a custom processor but with a more streamlined implementation using standard OpenTelemetry components.

## 📈 Observing in New Relic

Once data is flowing to New Relic, you can query it using NRQL:

```sql
FROM Metric SELECT * WHERE processor.traceaware = 'processed' LIMIT 100
```

You can also query system metrics collected by the host metrics receiver:

```sql
FROM Metric SELECT latest(system.cpu.utilization) FACET cpu TIMESERIES 
```

## 🧩 Configuration Customization

For developers who want to modify the configuration:

1. Edit `config-test.yaml` to change metrics collection or processing
2. Add additional processors or receivers as needed
3. Modify the attributes added by the attributes processor

## 🔒 Security Considerations

This implementation follows security best practices:
- Uses the official OpenTelemetry Collector image
- Requires minimal configuration settings
- Does not expose unnecessary ports
- Uses resource limits to prevent container resource exhaustion