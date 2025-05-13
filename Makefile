# Unified Makefile for trace-aware-reservoir-otel

# Configuration variables
REGISTRY ?= ghcr.io
ORG ?= deepaucksharma
IMAGE_NAME ?= trace-reservoir
VERSION ?= v0.1.0
MAIN_IMAGE = $(REGISTRY)/$(ORG)/$(IMAGE_NAME):$(VERSION)
BENCH_IMAGE = $(REGISTRY)/$(ORG)/$(IMAGE_NAME)-bench:$(VERSION)
LICENSE_KEY ?= $(NEW_RELIC_KEY)
NAMESPACE ?= otel
TAG = $(shell git describe --tags --always --dirty)
BENCH_DURATION ?= 10m
BENCH_PROFILES ?= max-throughput-traces,tiny-footprint-edge
GOFLAGS ?= -v

# Colors for pretty output
BLUE := \033[1;34m
GREEN := \033[1;32m
YELLOW := \033[1;33m
RED := \033[1;31m
RESET := \033[0m

# Help command - lists all available targets with descriptions
.PHONY: help
help: ## Show this help
	@echo "$(BLUE)Trace-Aware Reservoir OpenTelemetry$(RESET)"
	@echo "$(GREEN)Usage:$(RESET) make [target] [options]"
	@echo ""
	@echo "$(GREEN)Available targets:$(RESET)"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(BLUE)%-20s$(RESET) %s\n", $$1, $$2}'
	@echo ""
	@echo "$(GREEN)Common commands:$(RESET)"
	@echo "  $(BLUE)make build$(RESET)       - Build the collector application"
	@echo "  $(BLUE)make test$(RESET)        - Run all unit tests"
	@echo "  $(BLUE)make image$(RESET)       - Build Docker image"
	@echo "  $(BLUE)make dev$(RESET)         - Complete development cycle: test, build, deploy"
	@echo "  $(BLUE)make bench$(RESET)       - Run benchmarks"
	@echo ""
	@echo "$(GREEN)Options:$(RESET)"
	@echo "  $(YELLOW)VERSION=$(RESET)version   - Set image version (default: $(VERSION))"
	@echo "  $(YELLOW)PROFILES=$(RESET)profile  - Set benchmark profiles (default: $(BENCH_PROFILES))"
	@echo "  $(YELLOW)DURATION=$(RESET)duration - Set benchmark duration (default: $(BENCH_DURATION))"
	@echo ""

#-----------------------------------------------------------
# Development tasks
#-----------------------------------------------------------

.PHONY: deps
deps: ## Install development dependencies
	@echo "$(GREEN)Installing development dependencies...$(RESET)"
	go mod download
	cd core/reservoir && go mod download

.PHONY: lint
lint: ## Run linters
	@echo "$(GREEN)Running linters...$(RESET)"
	go vet ./...
	cd core/reservoir && go vet ./...
	@echo "$(GREEN)Linting completed.$(RESET)"

.PHONY: test
test: ## Run all unit tests
	@echo "$(GREEN)Running all tests...$(RESET)"
	go test ./core/... ./apps/... ./bench/runner/... $(GOFLAGS)

.PHONY: test-core
test-core: ## Run core library tests only
	@echo "$(GREEN)Running core library tests...$(RESET)"
	cd core/reservoir && go test ./... $(GOFLAGS)

.PHONY: test-apps
test-apps: ## Run application tests only
	@echo "$(GREEN)Running application tests...$(RESET)"
	go test ./apps/... $(GOFLAGS)

.PHONY: test-bench
test-bench: ## Run benchmark framework tests only
	@echo "$(GREEN)Running benchmark framework tests...$(RESET)"
	go test ./bench/runner/... $(GOFLAGS)

.PHONY: build
build: ## Build the collector application
	@echo "$(GREEN)Building collector application...$(RESET)"
	go build -o bin/otelcol-reservoir ./apps/collector
	@echo "$(GREEN)Build completed: $(RESET)bin/otelcol-reservoir"

.PHONY: build-all
build-all: build ## Build all applications
	@echo "$(GREEN)Building all applications...$(RESET)"
	go build -o bin/bench-runner ./bench/runner
	go build -o bin/verify ./verify
	@echo "$(GREEN)All builds completed.$(RESET)"

#-----------------------------------------------------------
# Docker image building
#-----------------------------------------------------------

.PHONY: image
image: ## Build main Docker image
	@echo "$(GREEN)Building main Docker image: $(RESET)$(MAIN_IMAGE)"
	docker build -t $(MAIN_IMAGE) \
	  --build-arg VERSION=$(VERSION) \
	  -f build/docker/Dockerfile.multistage .

.PHONY: image-bench
image-bench: ## Build benchmark Docker image
	@echo "$(GREEN)Building benchmark Docker image: $(RESET)$(BENCH_IMAGE)"
	docker build -t $(BENCH_IMAGE) \
	  --build-arg VERSION=$(VERSION) \
	  -f build/docker/Dockerfile.bench .

.PHONY: image-dev
image-dev: ## Build development Docker image
	@echo "$(GREEN)Building development Docker image: $(RESET)$(MAIN_IMAGE)-dev"
	docker build -t $(MAIN_IMAGE)-dev \
	  --build-arg VERSION=$(VERSION) \
	  -f build/docker/Dockerfile.dev .

.PHONY: images
images: image image-bench ## Build all Docker images

.PHONY: push-image
push-image: ## Push Docker image to registry
	@echo "$(GREEN)Pushing Docker image: $(RESET)$(MAIN_IMAGE)"
	docker push $(MAIN_IMAGE)

#-----------------------------------------------------------
# Kubernetes deployment
#-----------------------------------------------------------

.PHONY: kind
kind: ## Create kind cluster if not exists
	@echo "$(GREEN)Creating Kind cluster...$(RESET)"
	kind create cluster --config infra/kind/kind-config.yaml || true

.PHONY: kind-load
kind-load: ## Load Docker images into kind cluster
	@echo "$(GREEN)Loading Docker images into Kind cluster...$(RESET)"
	kind load docker-image $(MAIN_IMAGE) || echo "$(YELLOW)Warning: Failed to load main image$(RESET)"
	kind load docker-image $(BENCH_IMAGE) || echo "$(YELLOW)Warning: Failed to load bench image$(RESET)"

.PHONY: deploy
deploy: ## Deploy to Kubernetes
	@echo "$(GREEN)Deploying to Kubernetes...$(RESET)"
	kubectl create namespace $(NAMESPACE) --dry-run=client -o yaml | kubectl apply -f -
	helm upgrade --install otel-reservoir ./infra/helm/otel-bundle \
	  -n $(NAMESPACE) \
	  --set mode=collector \
	  --set global.licenseKey="$(LICENSE_KEY)" \
	  --set image.repository="$(REGISTRY)/$(ORG)/$(IMAGE_NAME)" \
	  --set image.tag="$(VERSION)"
	@echo "$(GREEN)Deployment complete. Use 'make status' to check status.$(RESET)"

.PHONY: dev
dev: test build image kind-load deploy ## Complete development cycle: test, build image, deploy
	@echo "$(GREEN)Development cycle completed successfully.$(RESET)"

.PHONY: quickrun
quickrun: build image-bench kind-load ## Quick build and load for testing
	@echo "$(GREEN)Quick run completed. Images loaded into Kind cluster.$(RESET)"

#-----------------------------------------------------------
# Operations
#-----------------------------------------------------------

.PHONY: status
status: ## Check deployment status
	@echo "$(GREEN)Checking deployment status...$(RESET)"
	kubectl get pods -n $(NAMESPACE)

.PHONY: logs
logs: ## Stream collector logs
	@echo "$(GREEN)Streaming collector logs...$(RESET)"
	kubectl logs -f -n $(NAMESPACE) deployment/otel-reservoir-collector

.PHONY: metrics
metrics: ## Port-forward and check metrics
	@echo "$(GREEN)Port-forwarding to localhost:8888...$(RESET)"
	@kubectl port-forward -n $(NAMESPACE) svc/otel-reservoir-collector 8888:8888 & \
	PID=$$!; \
	echo "$(GREEN)Waiting for connection...$(RESET)"; \
	sleep 3; \
	echo "$(GREEN)Metrics for reservoir_sampler:$(RESET)"; \
	curl -s http://localhost:8888/metrics | grep reservoir_sampler; \
	kill $$PID

.PHONY: clean
clean: ## Clean up resources
	@echo "$(GREEN)Cleaning up resources...$(RESET)"
	helm uninstall otel-reservoir -n $(NAMESPACE) || true
	kubectl delete namespace $(NAMESPACE) || true
	rm -rf bin dist
	@echo "$(GREEN)Cleanup completed.$(RESET)"

#-----------------------------------------------------------
# Benchmarking
#-----------------------------------------------------------

.PHONY: bench-prep
bench-prep: image image-bench ## Prepare for benchmarking
	@echo "$(GREEN)Preparing for benchmarks...$(RESET)"

.PHONY: bench
bench: bench-prep ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(RESET)"
	cd bench && go run ./runner/main.go \
		-image $(MAIN_IMAGE) \
		-profiles $(BENCH_PROFILES) \
		-duration $(BENCH_DURATION) \
		$(if $(LICENSE_KEY),-nrLicense $(LICENSE_KEY),)

.PHONY: bench-quick
bench-quick: bench-prep ## Run a quick benchmark test
	@echo "$(GREEN)Running quick benchmark...$(RESET)"
	cd bench && go run ./runner/main.go \
		-image $(MAIN_IMAGE) \
		-profiles max-throughput-traces \
		-duration 2m
	@echo "$(GREEN)Quick benchmark completed.$(RESET)"

.PHONY: bench-clean
bench-clean: ## Clean up benchmark resources
	@echo "$(GREEN)Cleaning up benchmark resources...$(RESET)"
	kind delete cluster --name benchmark-kind || true
	@echo "$(GREEN)Benchmark cleanup completed.$(RESET)"

#-----------------------------------------------------------
# Utility targets
#-----------------------------------------------------------

.PHONY: version
version: ## Display current version
	@echo "$(GREEN)Project: $(RESET)trace-aware-reservoir-otel"
	@echo "$(GREEN)Version: $(RESET)$(VERSION)"
	@echo "$(GREEN)Commit: $(RESET)$(TAG)"

.PHONY: setup
setup: ## Setup development environment
	@echo "$(GREEN)Setting up development environment...$(RESET)"
	@mkdir -p bin
	go install golang.org/x/tools/cmd/goimports@latest
	go mod download
	cd core/reservoir && go mod download
	@echo "$(GREEN)Development environment setup completed.$(RESET)"

# Default target
.DEFAULT_GOAL := help