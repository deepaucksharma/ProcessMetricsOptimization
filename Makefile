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
 helm repo add boutique https://microservices-demo.github.io/microservices-demo ;\
 helm repo add sockshop https://weaveworks.github.io/sock-shop/ ;\
 helm repo update ;\
 helm upgrade --install boutique boutique/microservices-demo -n observability -f k8s/boutique-helm-values.yaml ;\
 helm upgrade --install sockshop sockshop/sock-shop -n observability -f k8s/sockshop-helm-values.yaml ;\
 $(CL) create secret generic newrelic-license \
       --from-literal=NEW_RELIC_LICENSE_KEY=$(NR_KEY) --dry-run=client -o yaml | $(CL) apply -f - ;\
 $(CL) apply -f k8s/collector-daemonset.yaml
DOWN = kind delete cluster --name demo
LOGS = kubectl -n observability logs -l app=nrdot-collector-host -f
endif

.PHONY: up down logs validate clean dashboard query
up:          ## Spin everything up
	$(UP)
	@echo "🚀  Lab ready – profile=$(PROFILE) mode=$(MODE) demo_id=$(DEMO_ID)"

down:        ## Tear everything down
	$(DOWN)

logs:        ## Follow collector logs
	$(LOGS)

validate:    ## Static config lint (dry-run collector)
	docker run --rm -v $(PWD)/config.yaml:/cfg.yaml \
               newrelic/nrdot-collector-host:1.1.0 \
               --config /cfg.yaml --dry-run

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