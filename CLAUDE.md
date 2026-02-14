# CLAUDE.md - Project Intelligence

## Language & Module

- **Go 1.22** with module `github.com/<user>/go-ordersvc`
- Build flags: `CGO_ENABLED=0 GOARCH=arm64`

## Architecture

Clean Architecture with strict layer separation:

```
handler (HTTP/gRPC) → service (business logic) → repo (data access)
```

- **handler/**: HTTP handlers (Chi) and gRPC servers. Request validation, response formatting.
- **service/**: Business logic. No direct DB or external calls. Depends on repo interfaces.
- **repo/**: Data access layer. Implements interfaces defined by service layer.
- **domain/**: Core entities and value objects. No external dependencies.

Dependencies flow inward only. Never import handler from service or service from repo.

## Stack

| Component | Library |
|-----------|---------|
| HTTP Router | chi (go-chi/chi/v5) |
| PostgreSQL | pgx/v5 |
| Redis | go-redis/redis/v9 |
| Messaging | Kafka (segmentio/kafka-go) |
| RPC | protobuf + gRPC |
| Validation | go-playground/validator/v10 |

## Infrastructure

- **Runtime**: Docker Desktop Kubernetes on Mac M4 (ARM64)
- **Orchestration**: Helm charts in `deploy/helm/`
- **Local Dev**: Docker Compose with Postgres, Redis, Kafka
- **Registry**: ghcr.io

## Testing

- **Style**: Table-driven tests with github.com/stretchr/testify
- **Race Detection**: Always run with `-race` flag
- **Coverage**: Minimum 80% for service layer
- **Naming**: `TestFunctionName_Scenario_ExpectedBehavior`
- **Mocks**: Interfaces in `internal/mocks/`, generated or hand-written

```go
func TestCreateOrder_ValidInput_ReturnsOrder(t *testing.T) {
    tests := []struct {
        name    string
        input   CreateOrderRequest
        want    *Order
        wantErr bool
    }{
        // test cases
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // test body
        })
    }
}
```

## Conventions

| Type | Convention | Example |
|------|------------|---------|
| Files | lowercase_snake | `order_handler.go` |
| Go types | CamelCase | `OrderService` |
| Interfaces | -er suffix or I- prefix | `OrderRepository` |
| Packages | lowercase, short | `order`, `repo` |
| Env vars | UPPER_SNAKE | `DATABASE_URL` |

## API Contracts

- HTTP endpoints defined in `api/openapi/`
- gRPC services defined in `api/proto/`
- Breaking changes require version bump in path (`/v1/` → `/v2/`)

## Architecture Decision Records

> **Before implementing any feature, check `docs/decisions/`**

All architectural decisions are documented in `docs/decisions/` using the ADR format.

### Rules

1. **Check existing ADRs** before implementing any feature
2. **Follow all CONSTRAINT rules** defined in ADRs
3. **If a constraint conflicts** with requirements → STOP and ask user
4. **After implementation** → Update Traceability section in relevant ADR

### CONSTRAINT Format

ADRs use this format for enforceable rules:

```
CONSTRAINT: [Rule description]
BECAUSE: [Rationale]
CHECK: [How to verify compliance]
```

### Finding Relevant ADRs

- Search `docs/decisions/*.md` for keywords related to your feature
- Pay special attention to CONSTRAINT blocks
- Check Traceability section for related PRs/tasks

## Compact Instructions

When context is compacted, preserve:
1. Architecture layer boundaries (handler → service → repo)
2. API contracts and endpoint signatures
3. Database schema decisions and migrations
4. gRPC/protobuf service definitions
5. Integration points with external services
6. **Architecture Decision Records and their CONSTRAINT rules**

## Project Commands

> **Rule: Always use make targets. Never run tools directly.**

| Target | Description |
|--------|-------------|
| `make build` | Build binary for ARM64 |
| `make run` | Run the service locally |
| `make clean` | Remove build artifacts |
| `make fmt` | Format code (gofmt + goimports) |
| `make vet` | Run go vet |
| `make lint` | Run golangci-lint |
| `make sec` | Security scan with gosec |
| `make vuln` | Vulnerability check with govulncheck |
| `make secrets` | Scan for secrets with gitleaks |
| `make test` | Run tests with race detector and coverage |
| `make test-integration` | Run integration tests |
| `make cover` | Open coverage report in browser |
| `make docker` | Build Docker image (ARM64) |
| `make docker-push` | Push to ghcr.io |
| `make scan` | Scan Docker image with Trivy |
| `make compose-up` | Start Postgres + Redis + Kafka |
| `make compose-down` | Stop compose services |
| `make k8s-lint` | Lint Helm charts + kubeconform |
| `make k8s-deploy` | Deploy to Kubernetes |
| `make k8s-status` | Show deployment status |
| `make proto` | Generate protobuf/gRPC code |
| `make drift-check` | Verify no config drift from ADRs |
| `make ci` | Full validation (lint + sec + vuln + test + drift-check + k8s-lint) |
| `make help` | Show all targets |
