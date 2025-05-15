# Internal Packages

This directory contains internal packages that are used by multiple processors but are not meant to be exported outside this project.

## Directory Structure

- `banding/`: Contains logic for adaptive decision making based on host metrics, primarily used by the AdaptiveTopK processor to dynamically adjust the number of processes to retain based on system load.

## Usage Guidelines

1. Code in these packages should be generalizable and reusable across multiple processors
2. Keep implementation details that are specific to a particular processor in that processor's package
3. Add unit tests for all code in internal packages
4. Document public functions and types thoroughly

## Adding New Internal Packages

When adding a new internal package:

1. Ensure it serves a clear purpose that multiple processors will benefit from
2. Create a README.md in the package directory explaining its purpose and usage
3. Keep interfaces simple and focused
4. Follow the project's coding standards