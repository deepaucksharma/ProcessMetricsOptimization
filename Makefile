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
 $(CL) apply -f k8s/namespace.yaml ;\
 $(CL) apply -f k8s/minimal-demo.yaml ;\
 $(CL) create secret generic newrelic-license \
       --from-literal=NEW_RELIC_LICENSE_KEY=$(NR_KEY) --dry-run=client -o yaml | $(CL) apply -f - ;\
 $(CL) apply -f k8s/collector-daemonset.yaml
DOWN = kind delete cluster --name demo
LOGS = kubectl -n observability logs -l app=nrdot-collector-host -f
endif

.PHONY: up down logs validate clean dashboard query all-profiles k8s-test

up:          ## Spin everything up
	$(UP)
	@echo "🚀  Lab ready – profile=$(PROFILE) mode=$(MODE) demo_id=$(DEMO_ID)"

down:        ## Tear everything down
	$(DOWN)

logs:        ## Follow collector logs
	$(LOGS)

validate:    ## Syntax check the config.yaml file
	@echo "Validating collector configuration..."
	@if which yamllint > /dev/null; then \
		yamllint -d relaxed --no-warnings config.yaml && echo "✅ config.yaml syntax is valid"; \
	else \
		docker run --rm -v $(PWD)/config.yaml:/config.yaml cytopia/yamllint:latest \
			/config.yaml -d relaxed --no-warnings && echo "✅ config.yaml syntax is valid"; \
	fi

validate-otel: ## OTel semantic config validation
	@echo "Validating OTel collector configuration..."
	@docker run --rm \
	  -v $(PWD)/config.yaml:/etc/nrdot-collector-host/config.yaml:ro \
	  -e NR_USE_BALANCED=true \
	  newrelic/nrdot-collector-host:1.1.0 \
	  --config /etc/nrdot-collector-host/config.yaml

clean:       ## Remove dangling docker volumes / kind data
	docker system prune -f
	kind delete cluster --name demo 2>/dev/null || true

dashboard:   ## Echo NR link filtered by this demo_id
	@echo "https://one.newrelic.com/launcher/dashboards.launcher?query=benchmark.demo_id%20%3D%20'$(DEMO_ID)'"

query:       ## Show profile comparison NRQL
	@echo "SELECT"
	@echo "  bytecountestimate()/1e6 AS \"MB/5m\","
	@echo "  uniques(metricName)     AS \"Series\""
	@echo "FROM   Metric"
	@echo "WHERE  metricName LIKE 'process.%'"
	@echo "FACET  benchmark.profile"
	@echo "SINCE 5 minutes AGO"

all-profiles:  ## Run all profiles in parallel with docker
	docker-compose -f docker-compose-all-profiles.yml up -d
	@echo "🚀 All profile collectors started with demo_id=$(DEMO_ID)"

k8s-test:    ## Test K8s connectivity to New Relic
	kubectl apply -f k8s/test-connectivity.yaml
	sleep 10
	kubectl -n observability logs -l job-name=test-nr-connectivity