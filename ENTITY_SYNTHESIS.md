# New Relic Entity Synthesis Guide for OpenTelemetry

This guide addresses the common situation where your data is arriving in New Relic (visible in NRQL queries) but not showing up in the pre-built dashboards, entity views (like Infrastructure UI), or specific charts.

## Understanding Entity Synthesis

New Relic uses a process called "Entity Synthesis" to create and populate entities like Hosts, Services, and other components in their UI. For this to work properly, your telemetry data must include specific attributes that New Relic looks for.

## Critical Attributes for Host Entity

| Attribute | Purpose | Required For |
|-----------|---------|-------------|
| `host.id` | Unique identifier for the host | Host entity creation |
| `host.name` | Display name for the host | UI display |
| `entity.type` | Type of entity (e.g., "HOST") | Entity categorization |
| `entity.guid` | Globally unique ID for entity | Entity linking |
| `service.name` | Name of the service | Service entity |
| `collector.name` | Source of telemetry | Data source identification |

## Metric Naming Conventions

New Relic dashboards often expect specific metric names:

| OpenTelemetry Standard | New Relic Expected | Used For |
|------------------------|------------------|----------|
| `process.cpu.utilization` | `process.cpu.pct` | Process CPU charts |
| `process.memory.usage` | `process.memory.usage` | Process memory charts |
| `system.cpu.utilization` | `system.cpu.utilization` | Host CPU charts |
| `system.memory.usage` | `system.memory.usage` | Host memory charts |

The `entity-fixed-config.yaml` maintains both original and transformed metric names for compatibility.

## Troubleshooting Steps

### 1. Verify Host ID and Name

Run this NRQL query:
```sql
FROM Metric SELECT count(*) FACET resource.host.id, resource.host.name 
WHERE instrumentation.provider = 'opentelemetry' 
SINCE 30 minutes ago
```

Expected: You should see a distinct host ID and name.

### 2. Check Entity Type Identification

Run this NRQL query:
```sql
FROM Metric SELECT count(*) FACET resource.entity.type 
WHERE instrumentation.provider = 'opentelemetry' 
SINCE 30 minutes ago
```

Expected: Should show "HOST" or other entity types you're expecting.

### 3. Verify Process Metrics Attributes

For process-level insights:
```sql
FROM Metric SELECT latest(process.cpu.pct) FACET process.pid, process.executable.name 
WHERE instrumentation.provider = 'opentelemetry' 
SINCE 30 minutes ago LIMIT 100
```

Expected: Should show process information with PIDs and executable names.

### 4. Verify System Metrics

For system-level metrics:
```sql
FROM Metric SELECT latest(system.cpu.utilization) TIMESERIES 
WHERE host.name = '[your-hostname]' 
SINCE 30 minutes ago
```

Expected: Should show CPU utilization data for your host.

### 5. Check Data Temporality

Run this NRQL query to check for negative deltas (indicates temporality issues):
```sql
FROM Metric SELECT min(process.cpu.pct), max(process.cpu.pct) 
WHERE instrumentation.provider = 'opentelemetry' 
SINCE 30 minutes ago
```

Expected: Min values should not be negative. If they are, temporality conversion may be misconfigured.

## Solutions for Common Issues

### Missing Host Entity in Infrastructure UI

1. **Problem**: No `host.id` or wrong format
   **Solution**: Use the `resourcedetection` processor with the system detector enabled and ensure it has access to read `/etc/machine-id` or equivalent

2. **Problem**: Missing `entity.type` = "HOST" attribute
   **Solution**: Add this attribute via the resource processor

3. **Problem**: Not sending required system metrics
   **Solution**: Enable `system.cpu.utilization`, `system.memory.usage`, and other core system metrics

### Process Information Not Visible

1. **Problem**: Missing process attributes
   **Solution**: Ensure process metrics include `process.pid`, `process.executable.name`, and optionally `process.command_line`

2. **Problem**: Process metrics getting filtered out
   **Solution**: Review any filter processors to ensure they're not removing process metrics

### Charts Show No Data

1. **Problem**: Metric naming mismatch
   **Solution**: Use the `metricstransform` processor to create both original and New Relic naming versions

2. **Problem**: Wrong aggregation functions in charts  
   **Solution**: Check if charts use sum() for counters or average() for gauges and ensure metrics have right types

## Environment Variables

Setting these environment variables can help with entity synthesis:

```bash
export OTEL_RESOURCE_ATTRIBUTES="host.name=$(hostname),entity.type=HOST"
export OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE=delta
```

## Testing Entity Synthesis

After making changes:

1. Stop any running collector
2. Update configuration and run script with fixes
3. Wait 2-5 minutes for data to propagate to New Relic
4. Check the Entities explorer in New Relic to see if your host appears
5. Verify in Infrastructure UI if processes are visible

## Advanced: Custom Entity Synthesis Rules

For persistent issues, you may need to create custom entity synthesis rules in New Relic:

1. Go to New Relic One > All capabilities > Settings > Entity synthesis rules
2. Create a new rule for "HOST" entity type
3. Set conditions based on metrics you're sending
4. Define required attributes for the entity

This approach can help when automatic synthesis fails despite proper configuration.