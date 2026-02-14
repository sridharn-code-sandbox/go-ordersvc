# go-ordersvc Makefile

BINARY_NAME := ordersvc
MODULE := github.com/<user>/go-ordersvc
REGISTRY := ghcr.io/<user>
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COVERAGE_FILE := coverage.out

# Build flags for ARM64 static binary
export CGO_ENABLED := 0
export GOARCH := arm64
export GOOS := darwin

LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: all build run clean fmt vet lint sec vuln secrets test test-integration cover \
        docker docker-push scan compose-up compose-down k8s-lint k8s-deploy k8s-status \
        proto drift-check ci help

all: build ## Default target

# ============================================================================
# Build
# ============================================================================

build: ## Build binary for ARM64
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

run: build ## Run the service locally
	./bin/$(BINARY_NAME)

clean: ## Remove build artifacts
	rm -rf bin/ $(COVERAGE_FILE)
	go clean -cache -testcache

# ============================================================================
# Code Quality
# ============================================================================

fmt: ## Format code (gofmt + goimports)
	gofmt -s -w .
	goimports -w .

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

# ============================================================================
# Security
# ============================================================================

sec: ## Security scan with gosec
	gosec -quiet ./...

vuln: ## Vulnerability check with govulncheck
	govulncheck ./...

secrets: ## Scan for secrets with gitleaks
	gitleaks detect --source . --verbose

# ============================================================================
# Testing
# ============================================================================

test: ## Run tests with race detector and coverage
	go test -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...

test-integration: ## Run integration tests
	go test -race -tags=integration -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...

cover: test ## Open coverage report in browser
	go tool cover -html=$(COVERAGE_FILE)

# ============================================================================
# Docker
# ============================================================================

docker: ## Build Docker image (ARM64)
	docker build --platform linux/arm64 -t $(REGISTRY)/$(BINARY_NAME):$(VERSION) .

docker-push: ## Push to ghcr.io
	docker push $(REGISTRY)/$(BINARY_NAME):$(VERSION)

scan: ## Scan Docker image with Trivy
	trivy image --severity HIGH,CRITICAL $(REGISTRY)/$(BINARY_NAME):$(VERSION)

# ============================================================================
# Docker Compose
# ============================================================================

compose-up: ## Start Postgres + Redis + Kafka
	docker compose up -d

compose-down: ## Stop compose services
	docker compose down -v

# ============================================================================
# Kubernetes
# ============================================================================

k8s-lint: ## Lint Helm charts + kubeconform
	helm lint deploy/helm/$(BINARY_NAME)
	helm template deploy/helm/$(BINARY_NAME) | kubeconform -strict -summary

k8s-deploy: ## Deploy to Kubernetes
	helm upgrade --install $(BINARY_NAME) deploy/helm/$(BINARY_NAME) \
		--set image.tag=$(VERSION) \
		--wait --timeout 5m

k8s-status: ## Show deployment status
	kubectl get pods -l app=$(BINARY_NAME)
	kubectl get svc -l app=$(BINARY_NAME)

# ============================================================================
# Protobuf
# ============================================================================

proto: ## Generate protobuf/gRPC code
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		api/proto/**/*.proto

# ============================================================================
# Architecture Validation
# ============================================================================

drift-check: ## Verify no config drift from ADRs
	@echo "==> Running ADR constraint checks..."
	@echo "Checking ADR-0001 CONSTRAINT: Handlers must not import database packages..."
	@grep -r "github.com/jackc/pgx\|database/sql" internal/handler/ && \
		{ echo "FAIL: Handler imports database packages (violates ADR-0001)"; exit 1; } || \
		echo "PASS: No database imports in handlers"
	@echo "==> All drift checks passed"

# ============================================================================
# CI
# ============================================================================

ci: lint sec vuln test drift-check k8s-lint ## Full validation (lint + sec + vuln + test + drift-check + k8s-lint)
	@echo "CI checks passed"

# ============================================================================
# Help
# ============================================================================

help: ## Show all targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
