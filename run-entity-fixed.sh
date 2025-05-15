#!/bin/bash

# Load machine-specific identification
if [ -f /etc/machine-id ]; then
  export HOST_ID=$(cat /etc/machine-id)
elif [ -f /var/lib/dbus/machine-id ]; then
  export HOST_ID=$(cat /var/lib/dbus/machine-id)
else
  # Fall back to a hash of the hostname for Mac
  export HOST_ID=$(hostname | md5sum | cut -d' ' -f1)
fi

# Set the New Relic license key
export NEW_RELIC_LICENSE_KEY=2a326cab47d49f5aab8db7ee009e4a57FFFFNRAL
# WARNING: Replace with your actual New Relic license key

# Set additional environment variables for entity synthesis
export OTEL_DEPLOYMENT_ENVIRONMENT=production
export OTEL_SERVICE_NAME=otel-collector-host
export OTEL_RESOURCE_ATTRIBUTES="host.name=$(hostname),entity.type=HOST"
export OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta

# Change to the directory containing this script
cd "$(dirname "$0")"

# Make sure we're using the entity-fixed configuration
CONFIG_FILE="./config/entity-fixed-config.yaml"

# Check if the configuration file exists
if [ ! -f "$CONFIG_FILE" ]; then
  echo "Error: Entity-fixed configuration file not found at $CONFIG_FILE"
  exit 1
fi

# Check if the collector binary exists
if [ ! -f "./otelcol" ]; then
  echo "Error: OpenTelemetry Collector binary (otelcol) not found"
  echo "Download it from https://github.com/open-telemetry/opentelemetry-collector-releases/releases"
  exit 1
fi

# Make sure the collector binary is executable
chmod +x ./otelcol

echo "Starting OpenTelemetry collector with entity-fixed configuration..."
echo "Host ID: $HOST_ID"

# Start the collector in the background with the entity-fixed configuration
./otelcol --config="$CONFIG_FILE" > otel-collector.log 2>&1 &

COLLECTOR_PID=$!

# Check if collector started successfully
sleep 2
if ! ps -p $COLLECTOR_PID > /dev/null; then
  echo "Error: Collector failed to start. Check the logs at $(pwd)/otel-collector.log"
  exit 1
fi

echo "Entity-fixed collector started with PID $COLLECTOR_PID. It will run in the background."
echo "To stop it, run: kill $COLLECTOR_PID"
echo "Logs are available in: $(pwd)/otel-collector.log"

# We'll let this run for 2 hours, then automatically shut down
(
  sleep 7200  # 2 hours
  if ps -p $COLLECTOR_PID > /dev/null; then
    echo "Auto-stopping collector after 2 hours" >> otel-collector.log
    kill $COLLECTOR_PID
  fi
) &

echo "The collector will automatically stop after 2 hours."

# Print NRQL queries to check for entity synthesis
cat << EOF

=================================================
USEFUL NRQL QUERIES TO CHECK ENTITY SYNTHESIS
=================================================

# Check for host entity related metrics:
FROM Metric SELECT count(*) WHERE instrumentation.provider = 'opentelemetry' AND entity.type = 'HOST' FACET host.id, host.name SINCE 30 minutes ago

# Check process metrics with PID details:
FROM Metric SELECT latest(process.cpu.pct) FACET host.name, process.executable.name, process.pid SINCE 30 minutes ago

# Verify trace-aware attributes:
FROM Metric SELECT count(*) WHERE trace.aware.collector = 'true' FACET collector.name SINCE 30 minutes ago

# Check system metrics for the Infrastructure UI:
FROM Metric SELECT latest(system.cpu.utilization) FACET host.name TIMESERIES SINCE 30 minutes ago

EOF

# If curl is available, test New Relic connectivity
if command -v curl &> /dev/null; then
  echo "Testing connectivity to New Relic..."
  if curl -s --head --connect-timeout 5 https://otlp.nr-data.net > /dev/null; then
    echo "✅ Successfully connected to New Relic endpoint"
  else
    echo "⚠️ Warning: Could not connect to New Relic endpoint. Check your internet connection."
  fi
fi