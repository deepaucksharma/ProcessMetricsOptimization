# Hello World OpenTelemetry Processor

A simple OpenTelemetry processor example that adds "Hello World" attributes to your metrics and sends them to New Relic.

## Overview

This project provides a simple Hello World processor for OpenTelemetry that:

1. Collects host metrics from your system
2. Adds custom "Hello World" attributes to all metrics
3. Sends the enriched data to New Relic
4. Includes Docker setup for easy deployment

## Quick Start

```bash
# Make the run script executable
chmod +x run.sh

# Start the collector
./run.sh
```

## Environment Variables

The processor uses the following environment variables (already configured in .env):

```
# New Relic License Key (REQUIRED)
NEW_RELIC_LICENSE_KEY=your_license_key

# Environment settings
OTEL_DEPLOYMENT_ENVIRONMENT=production
OTEL_SERVICE_NAME=otel-collector-host
OTEL_LOG_LEVEL=info

# Collection settings
COLLECTION_INTERVAL=30s
BATCH_SEND_SIZE=1000
BATCH_TIMEOUT=10s

# Memory settings
NEW_RELIC_MEMORY_LIMIT_MIB=250
NEW_RELIC_MEMORY_SPIKE_LIMIT_MIB=150
```

## Viewing Data in New Relic

Once your processor is running, you can view your data in New Relic by running the following NRQL query:

```sql
FROM Metric SELECT * WHERE hello.processor = 'Hello from OpenTelemetry!'
```

## Architecture

This processor demonstrates:

1. How to collect host metrics with OpenTelemetry
2. How to add custom attributes to telemetry data
3. How to send metrics to New Relic
4. Basic dockerization of an OpenTelemetry collector