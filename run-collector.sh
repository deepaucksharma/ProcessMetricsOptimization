#!/bin/bash

# Set the New Relic license key
export NEW_RELIC_LICENSE_KEY=2a326cab47d49f5aab8db7ee009e4a57FFFFNRAL

# Set additional environment variables
export OTEL_DEPLOYMENT_ENVIRONMENT=production
export OTEL_SERVICE_NAME_HOST_METRICS=trace-aware-collector

# Change to the directory containing this script
cd "$(dirname "$0")"

# Make sure we're using the trace-aware configuration
CONFIG_FILE="./config/trace-aware-config.yaml"

# Check if the configuration file exists
if [ ! -f "$CONFIG_FILE" ]; then
  echo "Error: Trace-aware configuration file not found at $CONFIG_FILE"
  exit 1
fi

# Start the collector in the background with the trace-aware configuration
./otelcol --config="$CONFIG_FILE" > otel-collector.log 2>&1 &

COLLECTOR_PID=$!

echo "Trace-aware collector started with PID $COLLECTOR_PID. It will run in the background."
echo "To stop it, run: kill $(pgrep -f otelcol)"
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