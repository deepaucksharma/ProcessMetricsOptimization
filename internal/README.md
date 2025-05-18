# Internal Packages

This directory contains shared implementation logic that can be used across multiple processors. These packages are not meant to be exported outside this project.

## Current Packages

| Package | Description | Primary Users |
|---------|-------------|---------------|
| [`banding/`](./banding/) | Adaptive decision making based on host metrics | AdaptiveTopK processor |

## Package Principles

These packages follow specific design principles:

| Principle | Description |
|-----------|-------------|
| **Reusability** | Code should be generalizable across multiple processors |
| **Simplicity** | Interfaces should be clean and focused |
| **Testability** | All code must have comprehensive unit tests |
| **Documentation** | Public functions and types must be thoroughly documented |
| **Separation** | Processor-specific logic stays in processor packages |

## Development Guidelines

When working with internal packages:

```go
// Import internal packages using the full module path
import "github.com/newrelic/nrdot-process-optimization/internal/banding"

// Use clear interfaces
type LoadMapper interface {
    GetKValueForLoad(loadValue float64) int
}

// Add comprehensive documentation
// LoadBandMapper maps host load metrics to discrete K values
// with configurable thresholds and hysteresis
type LoadBandMapper struct {
    // ...
}
```

## Adding New Packages

To contribute a new internal package:

1. **Identify reuse cases** - Ensure at least two processors will benefit
2. **Create package directory** - Use a descriptive, concise name
3. **Add README.md** - Document purpose, interfaces, and usage examples
4. **Implement with tests** - Aim for >90% test coverage
5. **Review existing packages** - Follow established patterns