# Banding Package

> The `banding` package provides adaptive decision making based on system metrics.

## Overview

This package is part of the AdaptiveTopK processor's Dynamic K functionality (Sub-Phase 2b). The AdaptiveTopK processor has been implemented with Fixed K functionality (Sub-Phase 2a) and this package will be completed during the implementation of Dynamic K functionality.

The package will provide utilities for mapping continuous metric values (like system load) to discrete operational parameters (like the K value for TopK selection).

## Key Components

| Component | Description | Status |
|-----------|-------------|--------|
| `LoadBandMapper` | Maps host load to K values based on thresholds | Planned |
| `HysteresisController` | Prevents rapid fluctuations across thresholds | Planned |
| `BandTransition` | Manages smooth transitions between bands | Planned |

## Capabilities

- **Continuous → Discrete Mapping**: Convert continuous metrics (e.g., CPU utilization 0.0-1.0) to discrete operational values
- **Threshold-Based Decisions**: Configure multiple thresholds with corresponding values
- **Hysteresis Control**: Prevent "flapping" when values oscillate around thresholds
- **Thread Safety**: All components will be safe for concurrent use
- **State Management**: Track historical values for informed decision making

## Example (Coming in Phase 2)

```go
// Create a new mapper with thresholds and corresponding K values
mapper := banding.NewLoadBandMapper(map[float64]int{
    0.2: 5,   // When load < 0.2, K = 5
    0.5: 10,  // When 0.2 <= load < 0.5, K = 10
    0.8: 20,  // When 0.5 <= load < 0.8, K = 20
    1.0: 30,  // When load >= 0.8, K = 30
}, banding.WithHysteresis(time.Second*15))

// Get the K value based on current load with hysteresis
kValue := mapper.GetKValueForLoad(0.47) // Returns 10
```

## Implementation Details

The `LoadBandMapper` will:

1. Accept a map of thresholds to values during initialization
2. Sort thresholds for efficient lookup
3. Apply optional hysteresis configuration (time-based or count-based)
4. Provide thread-safe lookup methods
5. Support metrics emission for visibility into decisions

## Interface (Planned)

```go
type LoadBandMapper interface {
    // GetKValueForLoad returns the appropriate K value based on the current load
    GetKValueForLoad(load float64) int

    // GetBandBoundaries returns the configured thresholds
    GetBandBoundaries() []float64

    // GetCurrentBand returns the band the mapper is currently in
    GetCurrentBand() int
}
```

This package will be implemented in Phase 2 (AdaptiveTopK processor development).