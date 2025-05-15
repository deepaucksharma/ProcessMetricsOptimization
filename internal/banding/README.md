# Banding Package

The banding package provides functionality for adaptive decision making based on system metrics. It will be implemented as part of the AdaptiveTopK processor development.

## Purpose

This package will provide utilities to:

1. Map a continuous metric value (like CPU utilization) to discrete "bands" or ranges
2. Determine the appropriate value of K (number of processes to keep) based on system load
3. Implement hysteresis to prevent rapid changes in K value when system load fluctuates around band thresholds

## Planned Features

- `LoadBandMapper`: Maps a host load metric to a corresponding K value based on configured thresholds
- Hysteresis support to prevent "flapping" when values are near thresholds
- Smooth transitions between bands
- Thread-safe state management for history-based decisions

## Implementation Timeline

This package will be implemented during Phase 2 of the project as part of the AdaptiveTopK processor development.

## Usage Example (Future)

```go
// Example of how this package will be used (not yet implemented)
mapper := banding.NewLoadBandMapper(map[float64]int{
    0.2: 5,  // When load < 0.2, K = 5
    0.5: 10, // When 0.2 <= load < 0.5, K = 10
    0.8: 20, // When 0.5 <= load < 0.8, K = 20
    1.0: 30, // When load >= 0.8, K = 30
})

// Get the K value based on current load
kValue := mapper.GetKValueForLoad(0.47) // Returns 10
```