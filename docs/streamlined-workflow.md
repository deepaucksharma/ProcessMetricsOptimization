# Streamlined Development Workflow

Our enhanced modular architecture enables a highly streamlined development workflow. This guide outlines the updated development practices adopted for this project, which serve as a blueprint for similar projects.

---

## 1. Centralized Command Interface with Make

### 🛠 Enhanced Unified Makefile

Our enhanced root Makefile provides a comprehensive set of commands for all development and operational tasks:

```bash
# View all available commands
make help

# Development tasks
make deps            # Install development dependencies
make lint            # Run linters
make test            # Run all unit tests
make test-core       # Run core library tests only
make test-apps       # Run application tests only
make build           # Build the collector application
make build-all       # Build all applications
make setup           # Setup development environment

# Docker image building
make image           # Build main Docker image
make image-bench     # Build benchmark Docker image
make image-dev       # Build development Docker image
make images          # Build all Docker images
make push-image      # Push Docker image to registry

# Kubernetes deployment
make kind            # Create kind cluster
make kind-load       # Load Docker images into kind cluster
make deploy          # Deploy to Kubernetes
make dev             # Complete development cycle: test, build image, deploy
make quickrun        # Quick build and load for testing

# Operations
make status          # Check deployment status
make logs            # Stream collector logs
make metrics         # Check collector metrics
make clean           # Clean up resources

# Benchmarking
make bench-prep      # Prepare for benchmarking
make bench           # Run benchmarks
make bench-quick     # Run a quick benchmark test
make bench-clean     # Clean up benchmark resources

# Utility targets
make version         # Display current version
```

**Advantages**:
- Intuitive commands with consistent color-coded output
- Highly configurable via environment variables or make parameters
- Comprehensive help system with examples
- Logical grouping of related commands

**Examples**:

```bash
# Run tests with specific flags
make test GOFLAGS="-v -count=1"

# Build and deploy with custom version
make dev VERSION=v0.2.0

# Run benchmarks with custom profiles and duration
make bench PROFILES=max-throughput-traces,tiny-footprint-edge DURATION=5m
```

---

## 2. Streamlined Docker Build Process

### 🐳 Multi-Target Dockerfile (build/docker/Dockerfile.streamlined)

Our new streamlined Docker approach:

1. **Optimized Multi-stage Builds**:
   - Shared base image with dependencies
   - Separate builder stages for different components
   - Production vs benchmark targets from the same Dockerfile
   
2. **Layering Optimizations**:
   - Dependency caching for faster builds
   - Minimal artifact copying between stages
   - Slim runtime images (~15MB)

3. **Runtime Configurability**:
   - Comprehensive environment variable support
   - Automatic healthchecks
   - Non-root user by default

**Usage**:

```bash
# Build production image
docker build -t reservoir:prod -f build/docker/Dockerfile.streamlined --target production .

# Build benchmark image
docker build -t reservoir:bench -f build/docker/Dockerfile.streamlined --target benchmark .
```

The Makefile automatically selects the appropriate target based on the command.

---

## 3. Simplified Benchmark Runner

### 🚀 One-Stop Benchmarking Script (scripts/run-benchmark.sh)

A new comprehensive benchmark runner offers:

1. **Flexible Execution Modes**:
   - Kind cluster (automatic setup and teardown)
   - Existing Kubernetes cluster
   - Local simulator mode
   
2. **Advanced Configuration**:
   - Profile selection
   - Duration setting
   - Resource cleanup control
   
3. **Result Aggregation**:
   - Automatic collection of metrics
   - Generation of summary reports
   - Persistent results storage

**Usage Examples**:

```bash
# Run in Kind with multiple profiles
./scripts/run-benchmark.sh --mode kind --profiles max-throughput-traces,tiny-footprint-edge --duration 5m

# Run on local machine for quick testing
./scripts/run-benchmark.sh --mode local --port 9090

# Run on existing Kubernetes cluster
./scripts/run-benchmark.sh --mode existing --kubeconfig ~/.kube/my-cluster.yaml
```

---

## 4. Modular Architecture Benefits

Our modular architecture unlocks significant workflow improvements:

```
trace-aware-reservoir-otel/
│
├── core/                     # Core library code
│   └── reservoir/            # Reservoir sampling implementation
├── apps/                     # Applications
│   ├── collector/            # OpenTelemetry collector integration
│   └── tools/                # Supporting tools
├── bench/                    # Benchmarking framework
│   ├── profiles/             # Benchmark profiles
│   ├── kpis/                 # Key Performance Indicators
│   └── runner/               # Go-based benchmark orchestrator
├── infra/                    # Infrastructure code
│   ├── helm/                 # Helm charts
│   └── kind/                 # Kind cluster configurations
├── build/                    # Build configurations
│   ├── docker/               # Dockerfiles
│   └── scripts/              # Build scripts
└── scripts/                  # Operational scripts
```

This structure delivers:

- **Independent Component Development**: Work on core, applications, or benchmarks separately
- **Clear Boundaries**: Well-defined interfaces between components
- **Focused Testing**: Test components in isolation with minimal dependencies
- **Dependency Management**: Core library with its own versioning

---

## 5. Streamlined Development Workflow

The enhanced workflow simplifies the development experience:

1. **Initial Setup** (one-time):
   ```bash
   make setup
   ```

2. **Development Cycle**:
   ```bash
   # Make changes to code
   make lint        # Check code quality
   make test        # Verify functionality
   make build       # Build application
   ```

3. **Quick Verification**:
   ```bash
   make quickrun
   ```

4. **Complete Verification**:
   ```bash
   make dev
   ```

5. **Benchmark Testing**:
   ```bash
   # Full benchmarks
   make bench
   
   # Quick benchmark test
   make bench-quick
   ```

6. **Cleanup**:
   ```bash
   make clean
   make bench-clean
   ```

---

## 6. CI/CD Integration

Our updated GitHub Actions workflows promote consistency between local and CI environments:

```yaml
# .github/workflows/ci.yml
jobs:
  test:
    steps:
      - uses: actions/checkout@v4
      - run: make test
  image:
    needs: test
    if: github.event_name == 'push' && startsWith(github.ref,'refs/tags/')
    steps:
      - uses: actions/checkout@v4
      - run: make image VERSION=${{ github.ref_name }}
      - run: make push-image
```

**Nightly Benchmarks**:
```yaml
# .github/workflows/bench.yml
jobs:
  benchmark:
    steps:
      - uses: actions/checkout@v4
      - run: make images VERSION=${{ github.sha }}
      - run: ./scripts/run-benchmark.sh --mode kind --duration 15m
      - uses: actions/upload-artifact@v3
        with:
          name: benchmark-results
          path: benchmark-results-*
```

---

## 7. Developer Experience Improvements

**Recommended additions**:

- **Devcontainer configuration**: Add a `.devcontainer` directory with VS Code settings
- **Pre-commit hooks**: Add `.pre-commit-config.yaml` for code quality checks
- **Contribution guide**: Clear instructions for new contributors
- **Example configurations**: Additional sample configurations for common use cases

---

## 8. Summary: Streamlined Workflow Benefits

Our enhanced architecture delivers these workflow improvements:

1. **Unified Interface**: Intuitive Makefile commands for all operations
2. **Containerization**: Optimized Docker builds for development and production
3. **Simplified Benchmarking**: Flexible benchmark runner with multiple modes
4. **Improved Modularity**: Clear component boundaries with well-defined interfaces
5. **Consistent Experience**: Same commands work locally and in CI pipelines
6. **Developer-Friendly**: Color-coded output, helpful documentation, and guided workflows

These improvements transform a complex project into a developer-friendly system with an efficient, streamlined workflow.