# OpenTelemetry Collector Configuration Files

This directory contains various configuration files for the OpenTelemetry collector with different optimization levels and feature sets.

## Configuration Files

### Primary Configurations

- **entity-fixed-config.yaml**: Enhanced configuration with proper entity synthesis for New Relic
  - Includes host.id and entity.type attributes
  - Optimized for proper New Relic entity visualization
  - Includes process metrics with PID-level details
  - Recommended for most use cases

- **fixed-nr-config.yaml**: Alternative configuration focusing on trace-aware attributes
  - Simplified attribute structure
  - Optimized for performance
  - Reduced metric cardinality

- **trace-aware-config.yaml**: Basic configuration with trace-aware attributes
  - Minimal configuration for trace awareness
  - Compatible with standard OpenTelemetry collector

### Backup Configurations

The `backup` directory contains previous versions of configurations:

- **balanced-config.yaml**: Configuration with balanced performance and detail
- **config.yaml**: Original configuration
- **fixed-config.yaml**: Initial fixed configuration
- **nrdot-config.yaml**: Configuration for New Relic NRDOT collector
- **trace-aware-config.yaml**: Previous version of trace-aware configuration

## Configuration Selection

Choose the appropriate configuration based on your needs:

1. **For entity visualization in New Relic**: Use `entity-fixed-config.yaml`
2. **For minimal overhead**: Use `fixed-nr-config.yaml`
3. **For basic trace awareness**: Use `trace-aware-config.yaml`

## Environment Variables

All configurations support the following environment variables:

- `NEW_RELIC_LICENSE_KEY`: Your New Relic license key
- `OTEL_DEPLOYMENT_ENVIRONMENT`: Environment name (e.g., production)
- `OTEL_SERVICE_NAME`: Service name for the collector

## Usage

To use a specific configuration:

### With Docker:
```bash
docker-compose -f docker-compose-entity-fixed.yml up -d
```

### With Local Binary:
```bash
./otelcol --config=./config/entity-fixed-config.yaml
```

### With Run Script:
```bash
./run-entity-fixed.sh
```