#!/bin/bash

# Function to validate New Relic license key format
validate_license_key() {
  local key=$1
  # Basic check - NR keys are typically 40 chars without FFFFNRAL suffix
  if [[ ${#key} -lt 30 || "$key" == *"FFFFNRAL"* ]]; then
    return 1
  fi
  return 0
}

# Set the New Relic license key
# Replace this with your actual New Relic license key
export NEW_RELIC_LICENSE_KEY=2a326cab47d49f5aab8db7ee009e4a57FFFFNRAL

# Validate the license key
if ! validate_license_key "$NEW_RELIC_LICENSE_KEY"; then
  echo "Error: The New Relic license key appears to be invalid or is still using the placeholder."
  echo "Please replace it with your actual New Relic license key."
  echo "Edit this script and update the NEW_RELIC_LICENSE_KEY value before running again."
  exit 1
fi

# Set additional environment variables
export OTEL_DEPLOYMENT_ENVIRONMENT=production
export OTEL_SERVICE_NAME_HOST_METRICS=trace-aware-collector

# Change to the directory containing this script
cd "$(dirname "$0")"

# Make sure we're using the fixed NR configuration
CONFIG_FILE="./config/fixed-nr-config.yaml"

# Check if the configuration file exists
if [ ! -f "$CONFIG_FILE" ]; then
  echo "Error: Fixed NR configuration file not found at $CONFIG_FILE"
  exit 1
fi

# Check if the collector binary exists
if [ ! -f "./otelcol" ]; then
  echo "Error: OpenTelemetry Collector binary (otelcol) not found in the current directory."
  echo "Please download it from https://github.com/open-telemetry/opentelemetry-collector-releases/releases"
  exit 1
fi

# Make sure the collector binary is executable
chmod +x ./otelcol

# Verify connectivity to New Relic endpoint
echo "Verifying connectivity to New Relic..."
if ! curl -s --head --connect-timeout 10 https://otlp.nr-data.net > /dev/null; then
  echo "Warning: Cannot connect to New Relic endpoint. Check your internet connection."
  read -p "Continue anyway? (y/n) " -n 1 -r
  echo 
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    exit 1
  fi
fi

echo "Starting OpenTelemetry collector with trace-aware configuration..."

# Start the collector in the background with the trace-aware configuration
./otelcol --config="$CONFIG_FILE" > otel-collector.log 2>&1 &

COLLECTOR_PID=$!

# Check if collector started successfully
sleep 2
if ! ps -p $COLLECTOR_PID > /dev/null; then
  echo "Error: Collector failed to start. Check the logs at $(pwd)/otel-collector.log"
  exit 1
fi

echo "Trace-aware collector started with PID $COLLECTOR_PID. It will run in the background."
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
echo "Testing the configuration..."

# Give collector a moment to initialize
sleep 5

# Check for initial errors in the log
if grep -q "error\|failed\|exception" otel-collector.log; then
  echo "Warning: Errors detected in collector logs. Please check $(pwd)/otel-collector.log"
else
  echo "Collector started successfully with no immediate errors."
fi

echo ""
echo "To verify data is being sent to New Relic:"
echo "1. Log into your New Relic account"
echo "2. Go to the Metrics Explorer"
echo "3. Search for metrics with the attribute 'trace.aware.collector: true'"
echo "4. Data may take 1-2 minutes to appear"