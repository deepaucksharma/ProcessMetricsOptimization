.PHONY: build test test-unit test-integration test-e2e test-urls bench lint docker-build compose-up compose-down logs sbom vuln-scan clean run run-demo help

# Build variables
COLLECTOR_IMAGE := nrdot-process-optimization:latest
BUILD_INFO_IMPORT_PATH := github.com/newrelic/nrdot-process-optimization/cmd/collector
VERSION := 0.1.0
GIT_SHA := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Default target
.DEFAULT_GOAL := help

# Help target to show available commands
help:
	@echo "NRDOT Process-Metrics Optimization Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build & Run:"
	@echo "  build              Build the project"
	@echo "  docker-build       Build Docker image"
	@echo "  run                Start services with Docker"
	@echo "  run-demo           Run the simple demo app"
	@echo ""
	@echo "Docker Operations:"
	@echo "  compose-up         Start services with Docker Compose"
	@echo "  compose-down       Stop Docker Compose services"
	@echo "  logs               View logs from Docker services"
	@echo ""
	@echo "Testing:"
	@echo "  test               Run all tests"
	@echo "  test-unit          Run unit tests"
	@echo "  test-integration   Run integration tests"
	@echo "  test-e2e           Run end-to-end tests"
	@echo "  test-urls          Test all service URLs"
	@echo "  bench              Run benchmarks"
	@echo ""
	@echo "Quality & Security:"
	@echo "  lint               Run linting and static analysis"
	@echo "  sbom               Generate Software Bill of Materials"
	@echo "  vuln-scan          Run vulnerability scanning"
	@echo ""
	@echo "Cleanup:"
	@echo "  clean              Clean build artifacts"

# Build the project
build:
	go build -o ./bin/otelcol ./cmd/collector

# Run all tests
test: test-unit test-integration test-e2e test-urls

# Run just unit tests
test-unit:
	go test -race -v ./...

# Run integration tests
test-integration:
	go test -race -v -tags=integration ./test/integration/...

# Run end-to-end tests
test-e2e:
	go test -race -v -tags=e2e ./test/e2e/...

# Test service URLs
test-urls:
	@echo "Testing all service URLs..."
	@test/url_check.sh

# Run benchmarks
bench:
	go test -run=XXX -bench=. ./...

# Run linting and static analysis
lint:
	go vet ./...
	go fmt ./...
	# If golangci-lint is installed, use it
	which golangci-lint &>/dev/null && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

# Build Docker image
docker-build:
	docker build -t $(COLLECTOR_IMAGE) -f build/Dockerfile .

# Run local development stack with Docker Compose
compose-up:
	docker-compose -f build/docker-compose.yaml up -d

# Stop local development stack
compose-down:
	docker-compose -f build/docker-compose.yaml down

# Show logs from Docker Compose services
logs:
	docker-compose -f build/docker-compose.yaml logs -f

# Run the application using the unified script (Docker mode)
run:
	./run.sh up

# Run the simple demo app
run-demo:
	./run.sh demo

# Generate SBOM
sbom:
	@echo "Generating Software Bill of Materials (SBOM)"
	@which syft &>/dev/null && syft . -o spdx-json > sbom.spdx.json || echo "syft not installed, skipping SBOM generation"

# Vulnerability scanning
vuln-scan:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Clean build artifacts
clean:
	rm -rf ./bin
	rm -f sbom.spdx.json
