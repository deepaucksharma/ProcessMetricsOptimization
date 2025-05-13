# New Relic DOT Demo Lab

A one-command demo environment for New Relic DOT (Distribution of Telemetry) with multiple collection profiles, showcasing Google's Online Boutique and Weaveworks' Sock Shop microservices.

> **ℹ️ Important Note**: This project uses New Relic DOT 1.1.0, which is based on OpenTelemetry Collector v0.125.0.

## Overview

This project provides a simple way to spin up a complete observability demo environment:

- **Demo Applications**:
  - Google's Online Boutique microservices demo (with built-in load generator)
  - Weaveworks' Sock Shop microservices demo (with built-in load generator)

- **New Relic DOT Collector** with five different collection profiles:
  - **Ultra**: Full fidelity metrics (5s interval)
  - **Balanced**: PID-aware metrics with reduced metrics (30s interval, ~8× reduction)
  - **Optimized**: Executable-aggregated metrics (60s interval, ~40-60× reduction)
  - **Lean**: SLO-focused metrics (120s interval, ~90× reduction)
  - **Micro**: Minimal SLO metrics (300s interval, ~250×+ reduction)

- **Deployment Options**:
  - Docker Compose (local development)
  - Kubernetes via kind (cloud-like deployment)
  - GitHub Actions (continuous monitoring)

## Profile Comparison

| Profile    | Collection Interval | Target Series/Host | Collector RAM | Points/Min | Key Features                                  |
|------------|---------------------|-------------------|--------------|------------|---------------------------------------------|
| Ultra      | 5s                  | 400-600           | 130-160 MB   | ~9,000     | Full fidelity metrics                       |
| Balanced   | 30s                 | ≤80               | 90-110 MB    | ~1,000     | Per-PID visibility                          |
| Optimized  | 60s                 | 30-40             | 70-80 MB     | ~500       | Executable aggregation (~40-60× reduction)  |
| Lean       | 120s                | 20-25             | 60-70 MB     | ~200       | SLO focus, higher thresholds (~90× reduction)|
| Micro      | 300s                | 15-20             | 55-65 MB     | ≤100       | Minimal SLO metrics (~250× reduction)       |

## Requirements

### Docker Mode
- Docker and Docker Compose
- New Relic License Key

### Kubernetes Mode
- Kind (Kubernetes in Docker)
- kubectl
- New Relic License Key

### GitHub Actions Mode
- GitHub repository
- GitHub Actions enabled
- New Relic License Key added as a repository secret named `NR_LICENSE_KEY`

## Quick Start

```bash
# Clone the repository
git clone https://github.com/your-org/nrdot-demo-lab
cd nrdot-demo-lab

# Set your New Relic license key
export NR_KEY=your_new_relic_license_key
# Or create a .env file with NR_KEY=your_key

# Validate configuration before running
make validate

# Start the environment with the default profile (balanced)
make up

# Follow collector logs
make logs

# Display NRQL for profile comparison
make query

# Get link to filtered dashboard
make dashboard

# When finished
make down
```

## Configuration Approaches

This demo lab provides two different approaches to configuring the NRDOT collector:

### 1. Multi-Pipeline Profile Configuration (Default)

The default approach uses a single YAML configuration file (`config.yaml`) with multiple conditional pipelines, each optimized for a specific profile. The appropriate pipeline is activated based on environment variables.

```bash
# Start with multi-pipeline configuration (balanced profile by default)
make up

# Change the profile
PROFILE=ultra make up
PROFILE=optimized make up
PROFILE=lean make up
PROFILE=micro make up
```

This approach is used by:
- `make up` (Docker)
- `MODE=kind make up` (Kubernetes)
- GitHub Actions continuous monitoring

### 2. Simplified Configuration

The simplified approach uses a more streamlined configuration file (`updated-config.yaml`) with a single pipeline that adapts based on environment variables.

```bash
# Start with simplified configuration
make simple-up

# Change the profile
PROFILE=ultra make simple-up
PROFILE=optimized make simple-up
PROFILE=lean make simple-up
PROFILE=micro make simple-up
```

### 3. All Profiles Simultaneously

For comparison purposes, you can run all five profiles in parallel:

```bash
# Run all profiles in Docker
make all-profiles

# Run all profiles in Kubernetes
make k8s-all-profiles
```

## Deployment Modes

```bash
# Docker mode (default)
MODE=docker make up

# Kubernetes (kind) mode
MODE=kind make up
```

## Quick NRQL to Compare Profiles

```sql
SELECT
  bytecountestimate()/1e6 AS "MB/5m",
  uniques(metricName)     AS "Series"
FROM   Metric
WHERE  metricName LIKE 'process.%'
FACET  benchmark.profile
SINCE 5 minutes AGO
```

## Accessing Demo Applications

- **Online Boutique**: http://localhost:8080
- **Sock Shop**: http://localhost:8079

## GitHub Actions Integration

Three approaches are provided based on your needs and GitHub plan:

### Approach 1: Free & Fully-Hosted (Cron-style re-launches)

Best for most users who want continuous monitoring within free tier limits.

- Add your New Relic license key as a repository secret named `NR_LICENSE_KEY`
- The workflow in `.github/workflows/scheduled-lab.yml` will run every 30 minutes
- Adjust the cron schedule as needed
- Manually trigger with different parameters via "Run workflow" button

### Approach 2: Budget / Pro-plan (Matrix fan-out + nightly soak)

For teams with paid GitHub plans who want to regularly compare multiple profiles.

- The workflow in `.github/workflows/matrix-lab.yml` runs all profiles simultaneously
- Scheduled to run daily at 03:00 UTC
- Good for regression testing and performance comparisons

### Approach 3: Truly Continuous (Self-hosted runner)

For 24/7 monitoring without limitations on a self-hosted runner.

1. Provision a small VM (2 vCPU + 2 GB RAM)
2. Install a self-hosted runner with tags: `self-hosted,linux,nrdot-lab`
3. Use the workflow in `.github/workflows/continuous-lab.yml` to run indefinitely

## Monitoring Your Monitoring

Add this New Relic alert to ensure your monitoring system is operating properly:

```sql
SELECT rate(sum(otelcol_exporter_sent_metric_points)) 
FROM Metric 
WHERE benchmark.profile IS NOT NULL 
FACET benchmark.profile
```

## VM Deployment Notes

For non-container deployments on VMs:

1. Create a systemd drop-in for the collector:
   ```
   mkdir -p /etc/systemd/system/nrdot-collector-host.service.d/
   cat > /etc/systemd/system/nrdot-collector-host.service.d/local.conf << EOL
   [Service]
   Environment="HOST_ROOT_PATH=/hostfs"
   
   # Multi-pipeline configuration
   Environment="PROFILE=lean"
   Environment="NR_USE_LEAN=true"
   
   # Or use simplified configuration
   # Environment="COLLECTION_INTERVAL=120s"
   # Environment="INCLUDE_THREADS=false"
   # Environment="INCLUDE_FDS=false"
   
   # Common settings
   Environment="MEM_LIMIT_MIB=256"
   Environment="NEW_RELIC_LICENSE_KEY=your-license-key"
   EOL
   ```

2. Mount hostfs if needed:
   ```
   mkdir -p /hostfs
   mount --rbind / /hostfs
   ```

3. Restart the collector:
   ```
   systemctl daemon-reload
   systemctl restart nrdot-collector-host
   ```

## Repository Structure

```
.
├── Makefile                      # Main orchestration with multiple profiles
├── docker-compose.yml            # Docker deployment with multi-pipeline config
├── docker-compose-all-profiles.yml # Deploy all profiles simultaneously
├── docker-compose-simplified.yml # Docker deployment with simplified config
├── config.yaml                   # Multi-pipeline configuration (5 profiles)
├── updated-config.yaml           # Simplified configuration using env vars
├── .env.example                  # Example environment variables
├── .github/                      # GitHub Actions integration
│   ├── actions/start-lab/        # Reusable composite action
│   │   └── action.yml
│   └── workflows/
│       ├── scheduled-lab.yml     # Approach 1: Cron-style relaunches
│       ├── matrix-lab.yml        # Approach 2: Matrix fan-out
│       └── continuous-lab.yml    # Approach 3: Self-hosted continuous
└── k8s/                          # Kubernetes configuration files
    ├── base/                     # Kustomize base configuration
    │   ├── kustomization.yaml    # Base kustomization file
    │   ├── namespace.yaml        # Observability namespace
    │   └── collector-daemonset.yaml # Base collector DaemonSet
    ├── overlays/                 # Kustomize profile-specific overlays
    │   ├── ultra/                # Ultra profile overlay
    │   ├── balanced/             # Balanced profile overlay
    │   ├── optimized/            # Optimized profile overlay
    │   ├── lean/                 # Lean profile overlay
    │   └── micro/                # Micro profile overlay
    ├── boutique-helm-values.yaml # Helm values for Online Boutique
    ├── sockshop-helm-values.yaml # Helm values for Sock Shop
    ├── minimal-demo.yaml         # Simple demo application (no Helm required)
    ├── all-profiles-deployments.yaml # Deploy all profiles to K8s
    └── test-connectivity.yaml    # Kubernetes job to test connectivity
```

## Security Note

This lab runs the collector with `pid: host` and mounts the host filesystem, which grants full host access. This is standard for monitoring but should be limited to trusted environments.

## Troubleshooting

### Common Issues

1. **403 Errors in Logs**: If you see HTTP 403 errors when the collector tries to send data to New Relic, check your license key.

2. **Filesystem Errors**: When running in a container, you may see errors about not being able to read filesystem usage. These are generally non-critical and won't prevent the collector from functioning properly.

3. **Missing Metrics**: If metrics aren't appearing in New Relic:
   - Check collector logs for errors with `make logs`
   - Verify the collection interval matches your profile (Ultra: 5s, Balanced: 30s, etc.)
   - Confirm your license key has the right permissions for metrics ingest

4. **Kubernetes Profile Selection**: If the profile isn't changing in Kubernetes mode, ensure you're using the latest version of kubectl with Kustomize support. The Kustomize approach is used to dynamically select the profile.

## License

This project is licensed under the Apache 2.0 License.