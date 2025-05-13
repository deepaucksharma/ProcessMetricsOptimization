#!/bin/bash
# Simple connectivity tester for New Relic OTLP
# Usage: ./test-nr-connectivity.sh [LICENSE_KEY]

# Use provided key or attempt to read from env var
NR_LICENSE_KEY=${1:-$NR_KEY}

if [ -z "$NR_LICENSE_KEY" ]; then
  echo "Error: New Relic license key not provided"
  echo "Usage: ./test-nr-connectivity.sh [LICENSE_KEY]"
  echo "   or: NR_KEY=your_license_key ./test-nr-connectivity.sh"
  exit 1
fi

# Tests to run
echo "🧪 Testing connectivity to New Relic OTLP endpoints..."

echo "⏱️ Testing OTLP HTTP metrics endpoint (US)..."
curl -s -o /dev/null -w "%{http_code}\n" \
  -X POST https://otlp.nr-data.net/v1/metrics \
  -H "Content-Type: application/json" \
  -H "api-key: $NR_LICENSE_KEY" \
  -d '{"resourceMetrics":[{"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"test-service"}}]},"scopeMetrics":[{"metrics":[{"name":"test.metric","gauge":{"dataPoints":[{"timeUnixNano":"1643226945000000000","asDouble":1}]}}]}]}]}'

echo "⏱️ Testing OTLP HTTP traces endpoint (US)..."
curl -s -o /dev/null -w "%{http_code}\n" \
  -X POST https://otlp.nr-data.net/v1/traces \
  -H "Content-Type: application/json" \
  -H "api-key: $NR_LICENSE_KEY" \
  -d '{"resourceSpans":[{"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"test-service"}}]},"scopeSpans":[{"spans":[{"traceId":"0123456789abcdef0123456789abcdef","spanId":"0123456789abcdef","name":"test-span","kind":1,"startTimeUnixNano":"1643226945000000000","endTimeUnixNano":"1643226946000000000","status":{}}]}]}]}'

echo "⏱️ Testing OTLP HTTP logs endpoint (US)..."
curl -s -o /dev/null -w "%{http_code}\n" \
  -X POST https://otlp.nr-data.net/v1/logs \
  -H "Content-Type: application/json" \
  -H "api-key: $NR_LICENSE_KEY" \
  -d '{"resourceLogs":[{"resource":{"attributes":[{"key":"service.name","value":{"stringValue":"test-service"}}]},"scopeLogs":[{"logRecords":[{"timeUnixNano":"1643226945000000000","severityText":"INFO","body":{"stringValue":"This is a test log message"}}]}]}]}'

echo 
echo "🔍 If you see 200 responses for all three endpoints, connectivity is good!"
echo "   If you see 401/403, check your license key."
echo "   If you see connection errors, check your network settings."
echo
echo "✅ Test complete!"