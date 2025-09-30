# 5G Network Project Makefile
# Comprehensive build, test, and deployment automation

.PHONY: all clean build test deploy help

# Variables
PROJECT_NAME := 5g-network
REGISTRY := docker.io/5gnetwork
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
DOCKER_BUILD_ARGS := --build-arg VERSION=$(VERSION)

# Go variables
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
GOFMT := $(GOCMD) fmt
GOVET := $(GOCMD) vet

# Directories
BIN_DIR := bin
COVERAGE_DIR := coverage
DEPLOY_DIR := deploy

# Network Functions
NFS := nrf amf smf upf ausf udm udr pcf nssf nef nwdaf gnb-cu gnb-du gnb-ru

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[1;33m
RED := \033[0;31m
NC := \033[0m

##@ General

.DEFAULT_GOAL := help

help: ## Display this help message
	@echo "$(GREEN)5G Network Project - Makefile Commands$(NC)"
	@echo ""
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make $(YELLOW)<target>$(NC)\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  $(GREEN)%-25s$(NC) %s\n", $$1, $$2 } /^##@/ { printf "\n$(YELLOW)%s$(NC)\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

setup: ## Set up development environment
	@echo "$(GREEN)Setting up development environment...$(NC)"
	@bash scripts/setup-dev-env.sh

fmt: ## Format Go code
	@echo "$(GREEN)Formatting Go code...$(NC)"
	@$(GOFMT) ./...

lint: ## Run linters
	@echo "$(GREEN)Running linters...$(NC)"
	@golangci-lint run --timeout 5m ./...

vet: ## Run go vet
	@echo "$(GREEN)Running go vet...$(NC)"
	@$(GOVET) ./...

tidy: ## Tidy Go modules
	@echo "$(GREEN)Tidying Go modules...$(NC)"
	@$(GOMOD) tidy

generate: ## Generate code (mocks, protobuf, eBPF)
	@echo "$(GREEN)Generating code...$(NC)"
	@$(GOCMD) generate ./...
	@cd observability/ebpf && make all

##@ Building

build: build-nfs ## Build all components
	@echo "$(GREEN)✓ All components built$(NC)"

build-nfs: $(addprefix build-,$(NFS)) ## Build all network functions

build-nrf: ## Build NRF
	@echo "$(GREEN)Building NRF...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/nrf -ldflags="-X main.Version=$(VERSION)" ./nf/nrf/cmd

build-amf: ## Build AMF
	@echo "$(GREEN)Building AMF...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/amf -ldflags="-X main.Version=$(VERSION)" ./nf/amf/cmd

build-smf: ## Build SMF
	@echo "$(GREEN)Building SMF...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/smf -ldflags="-X main.Version=$(VERSION)" ./nf/smf/cmd

build-upf: ## Build UPF
	@echo "$(GREEN)Building UPF...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/upf -ldflags="-X main.Version=$(VERSION)" ./nf/upf/cmd

build-ausf: ## Build AUSF
	@echo "$(GREEN)Building AUSF...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/ausf -ldflags="-X main.Version=$(VERSION)" ./nf/ausf/cmd

build-udm: ## Build UDM
	@echo "$(GREEN)Building UDM...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/udm -ldflags="-X main.Version=$(VERSION)" ./nf/udm/cmd

build-udr: ## Build UDR
	@echo "$(GREEN)Building UDR...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/udr -ldflags="-X main.Version=$(VERSION)" ./nf/udr/cmd

build-pcf: ## Build PCF
	@echo "$(GREEN)Building PCF...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/pcf -ldflags="-X main.Version=$(VERSION)" ./nf/pcf/cmd

build-nssf: ## Build NSSF
	@echo "$(GREEN)Building NSSF...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/nssf -ldflags="-X main.Version=$(VERSION)" ./nf/nssf/cmd

build-nef: ## Build NEF
	@echo "$(GREEN)Building NEF...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/nef -ldflags="-X main.Version=$(VERSION)" ./nf/nef/cmd

build-nwdaf: ## Build NWDAF
	@echo "$(GREEN)Building NWDAF...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/nwdaf -ldflags="-X main.Version=$(VERSION)" ./nf/nwdaf/cmd

build-gnb-cu: ## Build gNodeB CU
	@echo "$(GREEN)Building gNodeB CU...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/gnb-cu -ldflags="-X main.Version=$(VERSION)" ./nf/gnb/cmd/cu

build-gnb-du: ## Build gNodeB DU
	@echo "$(GREEN)Building gNodeB DU...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/gnb-du -ldflags="-X main.Version=$(VERSION)" ./nf/gnb/cmd/du

build-gnb-ru: ## Build gNodeB RU (simulator)
	@echo "$(GREEN)Building gNodeB RU...$(NC)"
	@$(GOBUILD) -o $(BIN_DIR)/gnb-ru -ldflags="-X main.Version=$(VERSION)" ./nf/gnb/cmd/ru

build-webui: ## Build WebUI
	@echo "$(GREEN)Building WebUI...$(NC)"
	@cd webui/frontend && npm run build

##@ Docker

docker-build-all: $(addprefix docker-build-,$(NFS)) ## Build all Docker images

docker-build-%: ## Build Docker image for specific NF
	@echo "$(GREEN)Building Docker image for $*...$(NC)"
	@docker build $(DOCKER_BUILD_ARGS) -t $(REGISTRY)/$*:$(VERSION) -t $(REGISTRY)/$*:latest -f nf/$*/Dockerfile .

docker-push-all: $(addprefix docker-push-,$(NFS)) ## Push all Docker images

docker-push-%: ## Push Docker image for specific NF
	@echo "$(GREEN)Pushing Docker image for $*...$(NC)"
	@docker push $(REGISTRY)/$*:$(VERSION)
	@docker push $(REGISTRY)/$*:latest

##@ Testing

test: test-unit ## Run all tests

test-unit: ## Run unit tests
	@echo "$(GREEN)Running unit tests...$(NC)"
	@mkdir -p $(COVERAGE_DIR)
	@$(GOTEST) -v -race -coverprofile=$(COVERAGE_DIR)/coverage.out -covermode=atomic ./...

test-integration: ## Run integration tests
	@echo "$(GREEN)Running integration tests...$(NC)"
	@$(GOTEST) -v -tags=integration ./test/integration/...

test-e2e: ## Run end-to-end tests
	@echo "$(GREEN)Running end-to-end tests...$(NC)"
	@$(GOTEST) -v -tags=e2e -timeout=30m ./test/e2e/...

test-coverage: test-unit ## Generate test coverage report
	@echo "$(GREEN)Generating coverage report...$(NC)"
	@$(GOCMD) tool cover -html=$(COVERAGE_DIR)/coverage.out -o $(COVERAGE_DIR)/coverage.html
	@echo "$(GREEN)Coverage report: $(COVERAGE_DIR)/coverage.html$(NC)"

test-bench: ## Run benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@$(GOTEST) -bench=. -benchmem ./...

##@ Kubernetes

create-cluster: ## Create local Kubernetes cluster (kind)
	@echo "$(GREEN)Creating Kubernetes cluster...$(NC)"
	@kind create cluster --name 5g-network --config deploy/kind/config.yaml

delete-cluster: ## Delete local Kubernetes cluster
	@echo "$(YELLOW)Deleting Kubernetes cluster...$(NC)"
	@kind delete cluster --name 5g-network

load-images: ## Load Docker images into kind cluster
	@echo "$(GREEN)Loading images into kind cluster...$(NC)"
	@for nf in $(NFS); do \
		kind load docker-image $(REGISTRY)/$$nf:latest --name 5g-network; \
	done

##@ Deployment

deploy-infra: ## Deploy infrastructure (ClickHouse, Victoria Metrics, etc.)
	@echo "$(GREEN)Deploying infrastructure...$(NC)"
	@kubectl create namespace databases --dry-run=client -o yaml | kubectl apply -f -
	@kubectl create namespace observability --dry-run=client -o yaml | kubectl apply -f -
	@helm upgrade --install clickhouse deploy/helm/clickhouse --namespace databases --wait
	@helm upgrade --install victoria-metrics deploy/helm/victoria-metrics --namespace observability --wait
	@helm upgrade --install otel-collector deploy/helm/otel-collector --namespace observability --wait
	@helm upgrade --install tempo deploy/helm/tempo --namespace observability --wait
	@helm upgrade --install grafana deploy/helm/grafana --namespace observability --wait

deploy-core: ## Deploy 5G core network
	@echo "$(GREEN)Deploying 5G core network...$(NC)"
	@kubectl create namespace 5gc --dry-run=client -o yaml | kubectl apply -f -
	@helm upgrade --install 5g-core deploy/helm/5g-core --namespace 5gc --wait

deploy-all: deploy-infra deploy-core ## Deploy everything

undeploy-core: ## Undeploy 5G core network
	@echo "$(YELLOW)Undeploying 5G core network...$(NC)"
	@helm uninstall 5g-core --namespace 5gc || true

undeploy-infra: ## Undeploy infrastructure
	@echo "$(YELLOW)Undeploying infrastructure...$(NC)"
	@helm uninstall clickhouse --namespace databases || true
	@helm uninstall victoria-metrics --namespace observability || true
	@helm uninstall otel-collector --namespace observability || true
	@helm uninstall tempo --namespace observability || true
	@helm uninstall grafana --namespace observability || true

undeploy-all: undeploy-core undeploy-infra ## Undeploy everything

##@ Quick Start

quick-start: ## Quick start - sets up everything
	@bash scripts/quick-start.sh

demo: ## Run demo scenario
	@echo "$(GREEN)Running demo scenario...$(NC)"
	@bash scripts/demo.sh

##@ Observability

logs-amf: ## View AMF logs
	@kubectl logs -n 5gc -l app=amf --follow

logs-smf: ## View SMF logs
	@kubectl logs -n 5gc -l app=smf --follow

logs-upf: ## View UPF logs
	@kubectl logs -n 5gc -l app=upf --follow

logs-nrf: ## View NRF logs
	@kubectl logs -n 5gc -l app=nrf --follow

grafana-port-forward: ## Port-forward to Grafana
	@echo "$(GREEN)Forwarding Grafana to http://localhost:3000$(NC)"
	@kubectl port-forward -n observability svc/grafana 3000:80

webui-port-forward: ## Port-forward to WebUI
	@echo "$(GREEN)Forwarding WebUI to http://localhost:8080$(NC)"
	@kubectl port-forward -n 5gc svc/webui-frontend 8080:3000

##@ Database

clickhouse-shell: ## Open ClickHouse shell
	@kubectl exec -it -n databases clickhouse-0 -- clickhouse-client

load-test-data: ## Load test subscriber data
	@bash scripts/load-test-data.sh

##@ Utilities

clean: ## Clean build artifacts
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BIN_DIR)
	@rm -rf $(COVERAGE_DIR)
	@rm -rf vendor
	@find . -name '*.test' -delete
	@cd observability/ebpf && make clean

clean-all: clean ## Clean everything including dependencies
	@echo "$(YELLOW)Cleaning all artifacts...$(NC)"
	@$(GOMOD) clean -cache
	@docker system prune -af

verify: lint vet test-unit ## Verify code quality
	@echo "$(GREEN)✓ Code verification passed$(NC)"

ci: tidy generate verify build test ## CI pipeline
	@echo "$(GREEN)✓ CI pipeline completed successfully$(NC)"

status: ## Show cluster status
	@echo "$(GREEN)=== Cluster Status ===$(NC)"
	@kubectl get nodes
	@echo ""
	@echo "$(GREEN)=== 5G Core Pods ===$(NC)"
	@kubectl get pods -n 5gc
	@echo ""
	@echo "$(GREEN)=== Infrastructure Pods ===$(NC)"
	@kubectl get pods -n databases
	@kubectl get pods -n observability

##@ Documentation

docs: ## Generate documentation
	@echo "$(GREEN)Generating documentation...$(NC)"
	@godoc -http=:6060 &
	@echo "$(GREEN)Documentation server running at http://localhost:6060$(NC)"

docs-stop: ## Stop documentation server
	@pkill -f "godoc -http=:6060" || true