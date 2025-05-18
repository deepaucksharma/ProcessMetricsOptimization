# Metric Utility Package

`metricutil` provides small helper functions for working with `pmetric` data structures. These utilities are shared across processors to reduce code duplication.

## Functions

- **GetNumericValue** - Returns the numeric value from a `NumberDataPoint` as a `float64`, handling both int and double types.
- **MetricPointCount** - Counts the total number of data points contained in a `pmetric.Metrics` collection.

These helpers are internal-only and intended for use by processors within this repository.
