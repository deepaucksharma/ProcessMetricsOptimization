# ---------- user knobs ----------
REGISTRY ?= ghcr.io
NR_KEY   ?= replace_me
PROFILE  ?= balanced        # ultra / balanced / optimized / lean / micro
MODE     ?= docker          # docker | kind
DEMO_ID  ?= demo-$(shell hostname)-$(shell date +%Y%m%d)

# Include .env file if exists for license keys
-include .env

# ---------- exported to compose / kubectl ----------
export DEMO_ID
export NR_USE_ULTRA     := $(if $(filter ultra,$(PROFILE)),true,false)
export NR_USE_BALANCED  := $(if $(filter balanced,$(PROFILE)),true,false)
export NR_USE_OPTIMIZED := $(if $(filter optimized,$(PROFILE)),true,false)
export NR_USE_LEAN      := $(if $(filter lean,$(PROFILE)),true,false)
export NR_USE_MICRO     := $(if $(filter micro,$(PROFILE)),true,false)

# ---------- derived commands ----------
ifeq ($(MODE),docker)
UP   = docker compose up -d
DOWN = docker compose down -v
LOGS = docker compose logs -f collector
endif

ifeq ($(MODE),kind)
CL = kubectl -n observability
UP = \
 kind create cluster --name demo 2>/dev/null ;\
 $(CL) create secret generic newrelic-license \
       --from-literal=NEW_RELIC_LICENSE_KEY=$(NR_KEY) --dry-run=client -o yaml | $(CL) apply -f - ;\
 kubectl apply -k k8s/overlays/$(PROFILE) ;\
 $(CL) apply -f k8s/minimal-demo.yaml
DOWN = kind delete cluster --name demo
LOGS = kubectl -n observability logs -l app=nrdot-collector-host -f
endif

.PHONY: up down logs validate validate-otel validate-otel-simple clean dashboard query all-profiles k8s-all-profiles k8s-test

up:                  ## Spin everything up
	$(UP)
	@echo "🚀  Lab ready – profile=$(PROFILE) mode=$(MODE) demo_id=$(DEMO_ID)"

down:                ## Tear everything down
	$(DOWN)

logs:                ## Follow collector logs
	$(LOGS)

validate:            ## Syntax check the config.yaml file
	@echo "Validating collector configuration..."
	@if which yamllint > /dev/null; then \
		yamllint -d relaxed --no-warnings config.yaml && echo "✅ config.yaml syntax is valid"; \
	else \
		docker run --rm -v $(PWD)/config.yaml:/config.yaml cytopia/yamllint:latest \
			/config.yaml -d relaxed --no-warnings && echo "✅ config.yaml syntax is valid"; \
	fi

validate-otel:       ## Validate five-profile config.yaml
	@echo "Validating multi-pipeline OTel collector configuration..."
	@docker run --rm \
	  -v $(PWD)/config.yaml:/etc/nrdot-collector-host/config.yaml:ro \
	  -e NR_USE_BALANCED=true \
	  newrelic/nrdot-collector-host:1.1.0 \
	  --config /etc/nrdot-collector-host/config.yaml

validate-otel-simple: ## Validate updated-config.yaml
	@echo "Validating simplified OTel collector configuration..."
	@docker run --rm \
	  -v $(PWD)/updated-config.yaml:/etc/otel/config.yaml:ro \
	  -e COLLECTION_INTERVAL=30s \
	  -e INCLUDE_THREADS=false \
	  -e INCLUDE_FDS=false \
	  newrelic/nrdot-collector-host:1.1.0 \
	  --config /etc/otel/config.yaml

clean:               ## Remove dangling docker volumes / kind data
	docker system prune -f
	kind delete cluster --name demo 2>/dev/null || true

dashboard:           ## Echo NR link filtered by this demo_id
	@echo "https://one.newrelic.com/launcher/dashboards.launcher?query=benchmark.demo_id%20%3D%20'$(DEMO_ID)'"

query:               ## Show profile comparison NRQL
	@echo "SELECT"
	@echo "  bytecountestimate()/1e6 AS \"MB/5m\","
	@echo "  uniques(metricName)     AS \"Series\""
	@echo "FROM   Metric"
	@echo "WHERE  metricName LIKE 'process.%'"
	@echo "FACET  benchmark.profile"
	@echo "SINCE 5 minutes AGO"

all-profiles:        ## Run all profiles in parallel with docker
	docker-compose -f docker-compose-all-profiles.yml up -d
	@echo "🚀 All profile collectors started with demo_id=$(DEMO_ID)"

k8s-all-profiles:    ## Run all profiles in parallel in Kubernetes
	kind create cluster --name demo 2>/dev/null || true
	kubectl -n observability create secret generic newrelic-license \
        --from-literal=NEW_RELIC_LICENSE_KEY=$(NR_KEY) --dry-run=client -o yaml | kubectl apply -f -
	cd k8s && ./update-configmap.sh
	kubectl apply -f k8s/all-profiles-deployments.yaml
	@echo "🚀 All profile collectors deployed to Kubernetes with demo_id=$(DEMO_ID)"

k8s-test:            ## Test K8s connectivity to New Relic
	kubectl apply -f k8s/test-connectivity.yaml
	sleep 10
	kubectl -n observability logs -l job-name=test-nr-connectivity

simple-up:           ## Run with simplified config
	@echo "Setting up profile-specific environment variables..."
	@export CLEAN_DEMO_ID=$$(echo "$(DEMO_ID)" | tr -d ' '); \
	export COLLECTION_INTERVAL="30s"; \
	export INCLUDE_THREADS="false"; \
	export INCLUDE_FDS="false"; \
	case "$(PROFILE)" in \
		"ultra") \
			export COLLECTION_INTERVAL="5s"; \
			export INCLUDE_THREADS="true"; \
			export INCLUDE_FDS="true"; \
			;; \
		"balanced") \
			export COLLECTION_INTERVAL="30s"; \
			;; \
		"optimized") \
			export COLLECTION_INTERVAL="60s"; \
			;; \
		"lean") \
			export COLLECTION_INTERVAL="120s"; \
			;; \
		"micro") \
			export COLLECTION_INTERVAL="300s"; \
			;; \
		*) \
			echo "Unknown profile: $(PROFILE), using balanced"; \
			export COLLECTION_INTERVAL="30s"; \
			;; \
	esac; \
	DEMO_ID=$$CLEAN_DEMO_ID \
	COLLECTION_INTERVAL=$$COLLECTION_INTERVAL \
	INCLUDE_THREADS=$$INCLUDE_THREADS \
	INCLUDE_FDS=$$INCLUDE_FDS \
	docker compose -f docker-compose-simplified.yml up -d
	@echo "🚀 Simplified collector started with profile=$(PROFILE)"