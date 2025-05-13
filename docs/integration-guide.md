# Trace-Aware Reservoir Integration Guide

This guide explains how to integrate the trace-aware reservoir sampling processor with other systems, with a primary focus on the New Relic OpenTelemetry Distribution (NR-DOT).

## Table of Contents
- [Overview](#overview)
- [Integration with NR-DOT](#integration-with-nr-dot)
- [Core Library Usage](#core-library-usage)
- [Integration with Other OTel Distributions](#integration-with-other-otel-distributions)
- [Deployment Options](#deployment-options)
- [Verification](#verification)
- [Troubleshooting](#troubleshooting)
- [Performance Tuning](#performance-tuning)

## Overview

The trace-aware reservoir sampler can be integrated with various OpenTelemetry distributions and backends. This guide focuses on the most common integration scenario with the New Relic OpenTelemetry Distribution (NR-DOT), but the approaches can be adapted for other systems.

## Integration with NR-DOT

### 1. Repository Preparation

First, create a tag for your version of the Reservoir Sampler:

```bash
cd trace-aware-reservoir-otel
git tag v0.1.0
git push origin v0.1.0
```

### 2. NR-DOT Manifest Modification

You can either:

1. **Use our pre-built image**: This approach uses our Dockerfile that already includes the necessary manifest modifications
2. **Modify NR-DOT directly**: This approach involves forking and modifying the NR-DOT repository

#### Option 1: Using Pre-built Image (Recommended)

Our streamlined Dockerfile handles NR-DOT integration automatically:

```bash
# From the repo root
make image VERSION=v0.1.0
```

The Dockerfile:
1. Clones the NR-DOT repository
2. Patches the manifest to include our processor
3. Builds the integrated binary
4. Creates a minimal runtime image

#### Option 2: Direct NR-DOT Modification

If you prefer to modify NR-DOT directly:

1. Fork and clone the NR-DOT repository:
   ```bash
   git clone https://github.com/newrelic/opentelemetry-collector-releases.git
   cd opentelemetry-collector-releases
   git checkout -b feat/reservoir-sampler
   ```

2. Update the manifest files:

   In `distributions/nrdot-collector-host/manifest.yaml` and `distributions/nrdot-collector-k8s/manifest.yaml`, add to the processors section:

   ```yaml
   processors:
     # existing processors...
     - gomod: github.com/deepaucksharma/trace-aware-reservoir-otel/apps/collector/processor/reservoirsampler_with_badger v0.1.0
   ```

3. Build the distribution following NR-DOT's build instructions

### 3. Verification

Verify the processor is included in your build:

```bash
docker run --rm ghcr.io/deepaucksharma/trace-reservoir:v0.1.0 --config=none components | grep reservoir_sampler
```

## Core Library Usage

The core library can be used independently of the OpenTelemetry collector for custom applications:

```go
import (
    reservoir "github.com/deepaucksharma/trace-aware-reservoir-otel/core/reservoir"
)

func main() {
    // Create a configuration
    config := reservoir.DefaultConfig()
    config.SizeK = 5000
    config.WindowDuration = 60 * time.Second
    config.TraceAware = true
    
    // Create a metrics reporter
    reporter := reservoir.NewPrometheusMetricsReporter()
    
    // Create a window manager
    window := reservoir.NewTimeWindow(config.WindowDuration)
    
    // Create a reservoir
    res := reservoir.NewAlgorithmR(config.SizeK, reporter)
    
    // To add a span
    span := reservoir.SpanData{
        TraceID: "trace-123",
        SpanID: "span-456",
        Name: "example-span",
        // other fields...
    }
    
    // Add span to reservoir
    res.AddSpan(span)
    
    // When window rolls over, get sample
    sample := res.GetSample()
    
    // Process the sample...
}
```

## Integration with Other OTel Distributions

To integrate with other OpenTelemetry distributions:

1. Follow the distribution's documentation for adding custom processors
2. Add the module path: `github.com/deepaucksharma/trace-aware-reservoir-otel/apps/collector/processor/reservoirsampler_with_badger`
3. Configure the processor in your pipeline

Example configuration for standard OTel Collector:

```yaml
processors:
  reservoir_sampler:
    size_k: 5000                        
    window_duration: 60s                
    checkpoint_path: /var/otelpersist/badger
    checkpoint_interval: 10s            
    trace_aware: true                   

service:
  pipelines:
    traces:
      receivers: [otlp]
      processors: [memory_limiter, batch, reservoir_sampler]
      exporters: [otlp]
```

## Deployment Options

### Using Helm Chart

We provide a consolidated Helm chart that supports multiple deployment scenarios:

```bash
export NEW_RELIC_KEY="your_license_key_here"

# Deploy with default configuration
helm upgrade --install otel-reservoir ./infra/helm/otel-bundle \
  --namespace otel --create-namespace \
  --set mode=collector \
  --set global.licenseKey="$NEW_RELIC_KEY" \
  --set image.repository="ghcr.io/deepaucksharma/trace-reservoir" \
  --set image.tag="v0.1.0"
```

### Using a Specific Profile

You can also deploy with a specific benchmark profile:

```bash
helm upgrade --install otel-reservoir ./infra/helm/otel-bundle \
  --namespace otel --create-namespace \
  --set mode=collector \
  --set profile=max-throughput-traces \
  --set global.licenseKey="$NEW_RELIC_KEY" \
  --set image.repository="ghcr.io/deepaucksharma/trace-reservoir" \
  --set image.tag="v0.1.0" \
  -f bench/profiles/max-throughput-traces.yaml
```

### Kubernetes Deployment with Makefile

For convenience, use our Makefile:

```bash
# Create a Kind cluster (if needed)
make kind

# Load the image
make kind-load

# Deploy with our Helm chart
make deploy VERSION=v0.1.0 LICENSE_KEY=$NEW_RELIC_KEY
```

## Verification

After deployment, verify that everything is working correctly:

1. **Component Registration**:
   Check `http://localhost:8888/metrics` for `otelcol_build_info{features=...;processor_reservoir_sampler;}`

2. **Reservoir Metrics**:
   Monitor `reservoir_size`, `sampled_spans_total`, and other metrics

3. **Badger Database**:
   Check `reservoir_db_size_bytes` and `compaction_count_total`

4. **Window Rollover**:
   Look for "Exporting reservoir" log lines when a window rolls over

5. **Data in New Relic** (if applicable):
   Verify data is flowing to New Relic by checking the Spans explorer

```bash
# Check metrics with the Makefile
make metrics | grep reservoir
```

## Troubleshooting

### Common Issues

1. **Processor Not Found**: 
   - Check the import path in the manifest.yaml files
   - Verify the `set_config_value` in the Dockerfile is correct

2. **Database Issues**: 
   - Ensure the PVC is correctly mounted 
   - Check the permissions on the /var/otelpersist directory
   - Verify the checkpoint_path is accessible

3. **No Traces in New Relic**: 
   - Verify the OTLP exporter configuration
   - Check the API key is correctly set in global.licenseKey

4. **Memory Usage High**: 
   - Adjust the trace_buffer_max_size parameter
   - Consider lowering the reservoir size_k value

### Logs Check

```bash
# View collector logs
kubectl logs -n otel deploy/otel-reservoir-collector

# If using Makefile
make logs
```

## Performance Tuning

Our modular architecture makes performance tuning easier:

1. **Core Algorithm Tuning**:
   - Adjust the reservoir size based on expected trace volume
   - Configure window duration for appropriate sampling periods

2. **Persistence Optimization**:
   - Set appropriate checkpoint_interval for your durability needs
   - Configure DB compaction schedule based on usage patterns

3. **Memory Management**:
   - Tune the trace buffer timeout and max size
   - Monitor and adjust resource limits based on usage

Example tuned configuration:

```yaml
processors:
  reservoir_sampler:
    size_k: 10000                       # Larger reservoir for higher volume
    window_duration: 30s                # Shorter window for more frequent export
    checkpoint_path: /var/otelpersist/badger
    checkpoint_interval: 5s             # More frequent checkpoints
    trace_aware: true
    trace_buffer_timeout: 5s            # Shorter buffer timeout
    trace_buffer_max_size: 200000       # Reduced buffer size
    db_compaction_schedule_cron: "0 */4 * * *"  # More frequent compaction
    db_compaction_target_size: 134217728  # 128 MiB
```

---

For more information on deployment and development, see the [Development Guide](development-guide.md).