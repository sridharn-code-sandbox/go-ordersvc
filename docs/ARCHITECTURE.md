# Architecture

This document explains the architecture of go-ordersvc, a Go microservice for order management.

## Clean Architecture

go-ordersvc follows Clean Architecture principles with strict layer separation. Dependencies flow inward only - outer layers can depend on inner layers, but never the reverse.

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP / gRPC                          │  ← Handlers (adapters)
├─────────────────────────────────────────────────────────┤
│                     Service                             │  ← Business Logic
├─────────────────────────────────────────────────────────┤
│                     Domain                              │  ← Core Entities
├─────────────────────────────────────────────────────────┤
│               Repository / Cache                        │  ← Data Access
└─────────────────────────────────────────────────────────┘
```

## Layer Responsibilities

### Domain Layer (`internal/domain/`)

The innermost layer containing core business entities with zero external dependencies.

**Files:**
- `order.go` - Order entity, OrderStatus enum, status transition rules
- `item.go` - OrderItem value object
- `errors.go` - Domain-specific errors
- `pagination.go` - Pagination types

**Key characteristics:**
- No imports from other internal packages
- No database or HTTP knowledge
- Contains validation logic (e.g., `CanTransitionTo()`)
- Defines domain errors used by all layers

```go
// Example: Status transition rules in domain layer
func (s OrderStatus) CanTransitionTo(newStatus OrderStatus) bool {
    validTransitions := map[OrderStatus][]OrderStatus{
        OrderStatusPending:    {OrderStatusConfirmed, OrderStatusCancelled},
        OrderStatusConfirmed:  {OrderStatusProcessing, OrderStatusCancelled},
        // ...
    }
    // ...
}
```

### Service Layer (`internal/service/`)

Contains business logic and orchestrates domain operations.

**Files:**
- `order_service.go` - Service interface definition
- `order_service_impl.go` - Implementation
- `dto.go` - Data Transfer Objects

**Key characteristics:**
- Depends on domain layer and repository interfaces
- Never imports concrete implementations (postgres, redis)
- Contains transaction orchestration
- Validates business rules

```go
type OrderService interface {
    CreateOrder(ctx context.Context, dto CreateOrderDTO) (*domain.Order, error)
    GetOrderByID(ctx context.Context, id string) (*domain.Order, error)
    UpdateOrder(ctx context.Context, id string, dto UpdateOrderDTO) (*domain.Order, error)
    DeleteOrder(ctx context.Context, id string) error
    ListOrders(ctx context.Context, req ListOrdersRequest) (*domain.PaginatedOrders, error)
    UpdateOrderStatus(ctx context.Context, id string, newStatus domain.OrderStatus) (*domain.Order, error)
}
```

### Handler Layer (`internal/handler/http/`)

HTTP adapters that translate HTTP requests/responses to service calls.

**Files:**
- `order_handler.go` - CRUD handlers for orders
- `health_handler.go` - Liveness/readiness probes
- `router.go` - Chi router setup
- `request.go` - HTTP request structs
- `response.go` - HTTP response structs
- `mapper.go` - Domain ↔ HTTP mapping

**Key characteristics:**
- Zero business logic - only request/response handling
- Cannot import database packages
- Maps HTTP errors to domain errors
- Uses Chi router for routing

```go
// Handlers only translate HTTP to service calls
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    json.NewDecoder(r.Body).Decode(&req)

    dto := service.CreateOrderDTO{...}
    order, err := h.service.CreateOrder(r.Context(), dto)

    w.Header().Set("Location", fmt.Sprintf("/api/v1/orders/%s", order.ID))
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(MapOrderToResponse(order))
}
```

### Repository Layer (`internal/repository/`)

Data access abstraction with concrete implementations.

**Files:**
- `order_repository.go` - Repository interface
- `postgres/order_repository_postgres.go` - PostgreSQL implementation
- `postgres/connection.go` - Database connection setup

**Key characteristics:**
- Interface defined separately from implementation
- Cannot import handler or service packages
- Handles database-specific concerns (SQL, transactions)
- Supports optimistic locking via version field

### Middleware Layer (`internal/middleware/`)

Cross-cutting concerns applied to all requests.

**Files:**
- `logging.go` - Structured request logging with slog

**Middleware stack (applied in order):**
1. `RequestID` - Generates unique request ID
2. `RealIP` - Extracts real client IP
3. `Logging` - Logs method, path, status, duration
4. `Recoverer` - Recovers from panics

## Dependency Injection

Dependencies flow from `main.go` down through constructors:

```go
// cmd/ordersvc/main.go
func main() {
    // Create dependencies bottom-up
    repo := postgres.NewOrderRepository(db)
    svc := service.NewOrderService(repo)
    handler := http.NewOrderHandler(svc)

    // Wire into router
    router := http.NewRouter(handler, healthHandler, logger)
    server.Start(router)
}
```

## ADR Constraints Enforcement

Architecture decisions are documented in `docs/decisions/` and enforced via `make drift-check`:

| Constraint | Check |
|------------|-------|
| Handlers must not import database packages | Grep for pgx/sql imports in handler/ |
| Service must not import postgres package | Grep for postgres imports in service/ |
| Repository must not import handler/service | Grep for handler/service imports in repository/ |
| API routes must use /api/v1 prefix | Grep for Route("/api/v1/orders" in handler |

## Makefile Workflow

The Makefile organizes the development workflow into logical groups:

### Build
```bash
make build          # Build ARM64 binary
make run            # Build and run locally
make clean          # Remove build artifacts
```

### Code Quality
```bash
make fmt            # Format code (gofmt + goimports)
make vet            # Run go vet
make lint           # Run golangci-lint
```

### Security
```bash
make sec            # Security scan (gosec)
make vuln           # Vulnerability check (govulncheck)
make secrets        # Secret detection (gitleaks)
```

### Testing
```bash
make test           # Unit tests with race detector
make test-integration  # Integration tests (requires docker-compose)
make cover          # Open coverage report in browser
```

### Docker
```bash
make docker         # Build Docker image (ARM64)
make docker-push    # Push to ghcr.io
make scan           # Trivy security scan
```

### Docker Compose
```bash
make compose-up     # Start Postgres + Redis + ordersvc
make compose-down   # Stop and remove volumes
make compose-logs   # View logs
```

### Kubernetes
```bash
make k8s-lint       # Lint Helm charts + kubeconform
make k8s-deploy     # Deploy with Helm
make k8s-status     # Show pods and services
```

### Validation
```bash
make drift-check    # Verify ADR constraint compliance
make ci             # Full CI pipeline (lint + sec + vuln + test + drift-check + k8s-lint)
```

## Directory Structure

```
go-ordersvc/
├── cmd/ordersvc/           # Application entry point
│   ├── main.go             # Startup, DI, server init
│   └── server.go           # HTTP server setup
├── internal/
│   ├── config/             # Configuration loading
│   ├── domain/             # Core entities (no deps)
│   ├── service/            # Business logic
│   ├── repository/         # Data access interfaces
│   │   └── postgres/       # PostgreSQL implementation
│   ├── handler/
│   │   └── http/           # Chi HTTP handlers
│   └── middleware/         # HTTP middleware
├── deploy/
│   ├── docker/             # Dockerfile, docker-compose
│   └── helm/ordersvc/      # Helm chart
├── db/migrations/          # SQL migrations
├── test/integration/       # Integration tests
├── docs/
│   ├── decisions/          # ADRs
│   ├── API.md              # API reference
│   └── ARCHITECTURE.md     # This file
└── Makefile
```

## ADR Location

Architecture Decision Records are stored in `docs/decisions/`:

| ADR | Title |
|-----|-------|
| ADR-0001 | Clean Architecture |
| ADR-0002 | Order Details REST API |
| ADR-0003 | Optimistic Locking |

Each ADR follows the format:
- **Status** - Accepted/Deprecated
- **Business Context** - Problem statement
- **Options Considered** - Alternatives evaluated
- **Decision** - Chosen approach with rationale
- **Constraints** - Enforceable rules (checked by drift-check)
- **Consequences** - Trade-offs
- **Traceability** - Related PRs and tasks

## Testing Strategy

### Unit Tests
- Located alongside source files (`*_test.go`)
- Mock dependencies using interfaces
- Table-driven tests for comprehensive coverage

### Integration Tests
- Located in `test/integration/`
- Build tag: `//go:build integration`
- Run against real services via docker-compose
- Test full API contracts

### Running Tests
```bash
# Unit tests only
make test

# Integration tests (start docker-compose first)
make compose-up
make test-integration
```

## Health Checks

Two endpoints for Kubernetes probe compatibility:

- `/healthz` (Liveness) - Returns 200 if server is running
- `/readyz` (Readiness) - Returns 200 if database is healthy, 503 otherwise

Health endpoints are mounted outside authentication middleware (ADR-0002 constraint).
