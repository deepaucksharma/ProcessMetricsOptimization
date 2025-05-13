#!/bin/bash
#
# Simplified Benchmark Runner for trace-aware-reservoir-otel
#
# This script provides an easy interface to run benchmarks with various configurations
# and profiles. It handles the setup of the benchmark environment, running the benchmarks,
# and cleanup afterwards.
#

set -eo pipefail

# Default values
IMAGE=""
VERSION="v0.1.0"
DURATION="10m"
PROFILES="max-throughput-traces"
LICENSE_KEY=""
KUBECONFIG=""
CLEANUP="true"
MODE="kind"  # Options: kind, existing, local
LOCAL_PORT="8888"
REGISTRY="ghcr.io"
ORG="deepaucksharma"
IMAGE_NAME="trace-reservoir"

# Colors
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print banner
echo -e "${BLUE}==================================================${NC}"
echo -e "${BLUE}   Trace-Aware Reservoir Benchmark Runner ${NC}"
echo -e "${BLUE}==================================================${NC}"

# Help function
function show_help {
    echo -e "${GREEN}Usage:${NC} $0 [options]"
    echo
    echo -e "${GREEN}Options:${NC}"
    echo "  -h, --help                 Show this help message"
    echo "  -i, --image IMAGE          Use specific image (default: auto-build)"
    echo "  -v, --version VERSION      Image version (default: $VERSION)"
    echo "  -d, --duration DURATION    Benchmark duration (default: $DURATION)"
    echo "  -p, --profiles PROFILES    Comma-separated list of profiles (default: $PROFILES)"
    echo "  -l, --license LICENSE_KEY  New Relic license key"
    echo "  -k, --kubeconfig FILE      Path to kubeconfig file for existing cluster"
    echo "  -m, --mode MODE            Benchmark mode: kind, existing, local (default: $MODE)"
    echo "  --port PORT                Local port for simulator mode (default: $LOCAL_PORT)"
    echo "  --no-cleanup               Don't clean up resources after benchmark"
    echo
    echo -e "${GREEN}Examples:${NC}"
    echo "  $0 --mode kind --profiles max-throughput-traces,tiny-footprint-edge --duration 5m"
    echo "  $0 --mode local --port 9090"
    echo "  $0 --mode existing --kubeconfig ~/.kube/my-cluster.yaml"
    echo
    exit 0
}

# Process command line arguments
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in
        -h|--help)
            show_help
            ;;
        -i|--image)
            IMAGE="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -d|--duration)
            DURATION="$2"
            shift 2
            ;;
        -p|--profiles)
            PROFILES="$2"
            shift 2
            ;;
        -l|--license)
            LICENSE_KEY="$2"
            shift 2
            ;;
        -k|--kubeconfig)
            KUBECONFIG="$2"
            shift 2
            ;;
        -m|--mode)
            MODE="$2"
            shift 2
            ;;
        --port)
            LOCAL_PORT="$2"
            shift 2
            ;;
        --no-cleanup)
            CLEANUP="false"
            shift
            ;;
        *)
            echo -e "${RED}Error:${NC} Unknown option $1"
            show_help
            ;;
    esac
done

# Verify dependencies
function check_dependencies {
    echo -e "${BLUE}[1/6]${NC} Checking dependencies..."
    
    # Check for required commands
    for cmd in "go" "docker" "kubectl" "kind" "helm"; do
        if ! command -v $cmd &> /dev/null; then
            echo -e "${RED}Error:${NC} $cmd is required but not installed."
            echo "Please install $cmd and try again."
            exit 1
        fi
    done
    
    echo -e "${GREEN}All dependencies satisfied.${NC}"
}

# Build images if needed
function build_images {
    echo -e "${BLUE}[2/6]${NC} Preparing Docker images..."
    
    if [ -z "$IMAGE" ]; then
        echo "No image specified, building locally..."
        
        # Set image name
        IMAGE="${REGISTRY}/${ORG}/${IMAGE_NAME}:${VERSION}"
        BENCH_IMAGE="${REGISTRY}/${ORG}/${IMAGE_NAME}-bench:${VERSION}"
        
        # Build images
        echo "Building main image: $IMAGE"
        docker build -t "$IMAGE" \
            --build-arg VERSION="$VERSION" \
            -f build/docker/Dockerfile.streamlined \
            --target production .
            
        echo "Building benchmark image: $BENCH_IMAGE"
        docker build -t "$BENCH_IMAGE" \
            --build-arg VERSION="$VERSION" \
            -f build/docker/Dockerfile.streamlined \
            --target benchmark .
    else
        echo "Using specified image: $IMAGE"
    fi
    
    echo -e "${GREEN}Images ready.${NC}"
}

# Setup kind cluster
function setup_kind {
    echo -e "${BLUE}[3/6]${NC} Setting up Kind cluster..."
    
    # Check if cluster exists, create if it doesn't
    if ! kind get clusters | grep -q "benchmark-kind"; then
        echo "Creating Kind cluster 'benchmark-kind'..."
        kind create cluster --name benchmark-kind --config infra/kind/bench-config.yaml
    else
        echo "Using existing Kind cluster 'benchmark-kind'."
    fi
    
    # Load images
    echo "Loading Docker images into cluster..."
    kind load docker-image "$IMAGE" --name benchmark-kind
    if [ -n "$BENCH_IMAGE" ]; then
        kind load docker-image "$BENCH_IMAGE" --name benchmark-kind
    fi
    
    echo -e "${GREEN}Kind cluster ready.${NC}"
}

# Run benchmarks
function run_benchmarks {
    echo -e "${BLUE}[4/6]${NC} Running benchmarks..."
    
    case $MODE in
        kind)
            # Run benchmarks in Kind
            echo "Running benchmarks in Kind cluster..."
            cd bench && go run ./runner/main.go \
                -image "$IMAGE" \
                -profiles "$PROFILES" \
                -duration "$DURATION" \
                ${LICENSE_KEY:+-nrLicense "$LICENSE_KEY"}
            ;;
        existing)
            # Run benchmarks in existing K8s cluster
            echo "Running benchmarks in existing Kubernetes cluster..."
            cd bench && go run ./runner/main.go \
                -image "$IMAGE" \
                -profiles "$PROFILES" \
                -duration "$DURATION" \
                ${LICENSE_KEY:+-nrLicense "$LICENSE_KEY"} \
                ${KUBECONFIG:+-kubeconfig "$KUBECONFIG"}
            ;;
        local)
            # Run in local simulator mode
            echo "Running in local simulator mode..."
            docker run -p "$LOCAL_PORT:8888" \
                --env RESERVOIR_SIZE=10000 \
                --env WINDOW_DURATION=60s \
                --env TRACE_AWARE=true \
                --env LOG_LEVEL=info \
                --rm "$IMAGE"
            ;;
        *)
            echo -e "${RED}Error:${NC} Unknown mode: $MODE"
            exit 1
            ;;
    esac
    
    echo -e "${GREEN}Benchmark execution completed.${NC}"
}

# Collect results
function collect_results {
    echo -e "${BLUE}[5/6]${NC} Collecting benchmark results..."
    
    # Create results directory
    RESULTS_DIR="benchmark-results-$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$RESULTS_DIR"
    
    # Collect CSV files
    if [ -d "/tmp" ] && [ "$(ls -A /tmp/kpi_*.csv 2>/dev/null)" ]; then
        echo "Copying KPI results to $RESULTS_DIR/"
        cp /tmp/kpi_*.csv "$RESULTS_DIR/" 2>/dev/null || true
    fi
    
    # Generate summary if in Kubernetes modes
    if [ "$MODE" = "kind" ] || [ "$MODE" = "existing" ]; then
        echo "Generating summary report..."
        echo "# Benchmark Results - $(date +%Y-%m-%d)" > "$RESULTS_DIR/summary.md"
        echo "" >> "$RESULTS_DIR/summary.md"
        echo "## Configuration" >> "$RESULTS_DIR/summary.md"
        echo "- Duration: $DURATION" >> "$RESULTS_DIR/summary.md"
        echo "- Profiles: $PROFILES" >> "$RESULTS_DIR/summary.md"
        echo "- Image: $IMAGE" >> "$RESULTS_DIR/summary.md"
        echo "" >> "$RESULTS_DIR/summary.md"
        echo "## Results" >> "$RESULTS_DIR/summary.md"
        
        # Add results from CSV files
        for f in "$RESULTS_DIR"/*.csv; do
            if [ -f "$f" ]; then
                PROFILE=$(basename "$f" | sed 's/kpi_\(.*\)\.csv/\1/')
                echo "### Profile: $PROFILE" >> "$RESULTS_DIR/summary.md"
                echo "" >> "$RESULTS_DIR/summary.md"
                echo "| Metric | Value | Threshold | Status |" >> "$RESULTS_DIR/summary.md"
                echo "| ------ | ----- | --------- | ------ |" >> "$RESULTS_DIR/summary.md"
                
                # Skip header, process each line
                tail -n +2 "$f" | while IFS=, read -r metric value min max status; do
                    if [ "$status" = "pass" ]; then
                        status_text="✅ PASS"
                    else
                        status_text="❌ FAIL"
                    fi
                    echo "| $metric | $value | $min - $max | $status_text |" >> "$RESULTS_DIR/summary.md"
                done
                echo "" >> "$RESULTS_DIR/summary.md"
            fi
        done
    fi
    
    echo -e "${GREEN}Results collected in $RESULTS_DIR/${NC}"
}

# Clean up resources
function cleanup_resources {
    if [ "$CLEANUP" = "true" ]; then
        echo -e "${BLUE}[6/6]${NC} Cleaning up resources..."
        
        if [ "$MODE" = "kind" ]; then
            echo "Deleting Kind cluster 'benchmark-kind'..."
            kind delete cluster --name benchmark-kind
        fi
        
        echo -e "${GREEN}Cleanup completed.${NC}"
    else
        echo -e "${YELLOW}Skipping cleanup as requested.${NC}"
    fi
}

# Main execution flow
check_dependencies
build_images

if [ "$MODE" = "kind" ]; then
    setup_kind
fi

run_benchmarks
collect_results

if [ "$MODE" = "kind" ] || [ "$MODE" = "existing" ]; then
    cleanup_resources
fi

echo -e "${BLUE}==================================================${NC}"
echo -e "${GREEN}Benchmark completed successfully!${NC}"
echo -e "${BLUE}==================================================${NC}"
echo "Results are available in: $RESULTS_DIR/"
echo

exit 0