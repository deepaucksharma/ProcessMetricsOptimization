.PHONY: build test test-unit test-integration test-e2e bench lint docker-build compose-up compose-down logs sbom vuln-scan clean

# Build variables
COLLECTOR_IMAGE := nrdot-process-optimization:latest
BUILD_INFO_IMPORT_PATH := github.com/newrelic/nrdot-process-optimization/cmd/collector
VERSION := 0.1.0
GIT_SHA := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build the project
build:
	go build -o ./bin/otelcol ./cmd/collector

# Run all tests
test: test-unit test-integration test-e2e

# Run just unit tests
test-unit:
	go test -race -v ./...

# Run integration tests
test-integration:
	go test -race -v -tags=integration ./test/integration/...

# Run end-to-end tests
test-e2e:
	go test -race -v -tags=e2e ./test/e2e/...

# Run benchmarks
bench:
	go test -run=XXX -bench=. ./...

# Run linting and static analysis
lint:
	go vet ./...
	go fmt ./...
	# If golangci-lint is installed, use it
	which golangci-lint && golangci-lint run ./... || echo "golangci-lint not installed, skipping"

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

# Generate SBOM
sbom:
	echo "Generating Software Bill of Materials (SBOM)"
	# Example: use syft if installed
	which syft && syft . -o spdx-json > sbom.spdx.json || echo "syft not installed, skipping SBOM generation"

# Vulnerability scanning
vuln-scan:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

# Clean build artifacts
clean:
	rm -rf ./bin
	rm -f sbom.spdx.json