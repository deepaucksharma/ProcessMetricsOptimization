# Trace-Aware Reservoir OpenTelemetry Collector

This repository contains configuration and setup tools for running an OpenTelemetry collector with trace-aware attributes to improve metric visibility in New Relic.

## Features

- Trace-aware attribute addition to metrics
- Entity synthesis optimization for proper New Relic visualization
- Multi-environment support (Docker and Mac)
- Resource detection for proper host identification
- Process metrics collection with detailed attributes

## Getting Started

### Prerequisites

- Docker (for container deployment)
- New Relic account with a valid license key
- For local Mac execution: macOS with terminal access

### Setup

1. Clone this repository:
   ```
   git clone https://github.com/your-repo/trace-aware-reservoir-otel.git
   cd trace-aware-reservoir-otel
   ```

2. Copy the environment template:
   ```
   cp .env.example .env
   ```

3. Edit `.env` and add your New Relic license key:
   ```
   NEW_RELIC_LICENSE_KEY=your_license_key_here
   ```

4. Run the environment setup script:
   ```
   ./run-environments.sh
   ```

5. Choose your preferred environment:
   - Docker: For containerized deployment
   - Mac: For local execution on macOS
   - Both: Run in both environments simultaneously

## Configuration Files

- `config/entity-fixed-config.yaml`: Enhanced configuration with proper entity synthesis
- `config/fixed-nr-config.yaml`: Alternative configuration with different attribute structure
- `config/trace-aware-config.yaml`: Basic configuration with trace-aware attributes

## Troubleshooting

If metrics are not appearing correctly in New Relic:

1. Check `ENTITY_SYNTHESIS.md` for detailed troubleshooting steps
2. Follow the NRQL queries in the `run-environments.sh` output
3. Verify the license key is valid and connection to New Relic is working

## Guides

- `ENTITY_SYNTHESIS.md`: Detailed guide for New Relic entity synthesis
- `TROUBLESHOOTING.md`: General troubleshooting steps for collector issues

## Docker Configuration

The Docker setup uses the standard OpenTelemetry collector image with a custom configuration:

```
docker-compose -f docker-compose-entity-fixed.yml up -d
```

## Mac Configuration

The Mac setup downloads and runs the OpenTelemetry collector binary locally:

```
./run-entity-fixed.sh
```