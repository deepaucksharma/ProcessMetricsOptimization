# Troubleshooting New Relic Data Visibility

This guide helps troubleshoot why data might not be appearing in New Relic as expected, particularly focusing on trace-aware attributes.

## Common Issues

### 1. Invalid License Key

The most common cause is an invalid New Relic license key. The key in the scripts should be your actual New Relic ingest license key, not a placeholder.

To fix:
- Edit `run-fixed-collector.sh` and replace the placeholder key with your actual key
- New Relic license keys are typically 40 characters long
- Remove the `FFFFNRAL` suffix if present

### 2. Connection Issues

If the collector can't reach New Relic's endpoints, data won't appear.

To test connectivity:
```bash
curl -v https://otlp.nr-data.net
```

### 3. Configuration Format

New Relic expects data in a specific format. Our `fixed-nr-config.yaml` addresses common issues:

- Added standardized attribute naming: `collector.name`, `instrumentation.provider`
- Fixed resource attributes to match New Relic expectations
- Adjusted batch sizes and timeouts
- Corrected header format

### 4. Verify Data in New Relic

To see trace-aware metrics in New Relic:

1. Log in to your New Relic account
2. Navigate to **Metrics Explorer**
3. Use the NRQL query:
   ```
   FROM Metric SELECT * WHERE trace.aware.collector = 'true'
   ```
4. If no data appears, try:
   ```
   FROM Metric SELECT * WHERE collector.name = 'trace-aware-collector'
   ```

### 5. Check Logs for Errors

Examine the `otel-collector.log` file for errors:

```bash
grep -i "error\|failed" otel-collector.log | tail -20
```

Look for HTTP response codes - successful data transmission should show "200 OK" responses.

### 6. Restart with Debug Mode

If data still isn't appearing:

1. Stop any running collector processes
2. Edit the configuration to increase logging:
   ```yaml
   service:
     telemetry:
       logs:
         level: "debug"
   ```
3. Restart using the fixed script
4. Check logs again for detailed information

## New Relic Data Structure Requirements

For metrics to appear correctly in New Relic:

1. Make sure these attributes are properly set:
   - `service.name` - Critical for proper categorization
   - `deployment.environment` - Helps with filtering
   - `collector.name` - Helps identify the source

2. Use properly formatted metric names:
   - System metrics should use standard names (e.g., `system.cpu.utilization`)
   - Custom metrics should follow a hierarchy pattern

3. All attribute values should be strings, including boolean values:
   - Use `"true"` instead of `true`
   - Use `"1.0"` instead of `1.0`

## Next Steps

If problems persist:

1. Check New Relic's [metrics API docs](https://docs.newrelic.com/docs/apis/nerdgraph/examples/nerdgraph-metrics-api-tutorial/)
2. Try using the New Relic OTLP Example at https://github.com/newrelic/newrelic-opentelemetry-examples
3. Verify your account can receive custom metrics (some accounts may have restrictions)