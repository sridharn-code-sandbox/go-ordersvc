# ADR-0002: Order Details REST API Design

## Status

**Accepted**

**Date:** 2026-02-14

## Business Context

### Problem Statement
We need a RESTful API for managing orders that allows clients to create, read, update, and delete orders. The API must support pagination for listing orders, status filtering, and proper health checks for Kubernetes deployments.

### Success Criteria
- All CRUD operations available via REST endpoints
- Pagination support with configurable limits
- Status-based filtering for order lists
- Proper HTTP status codes (201, 200, 404, 409, etc.)
- Health endpoints compatible with Kubernetes probes
- Structured JSON logging for observability

### Constraints
- Must follow RESTful conventions
- Must use /api/v1 prefix for versioning
- Must return JSON responses
- Must log every request with structured slog

## Options Considered

### Option 1: RESTful Resource-Based API
**Pros:**
- Industry standard, well-understood
- Clear mapping: nouns to resources, verbs to HTTP methods
- Easy to document with OpenAPI
- Cacheable GET requests

**Cons:**
- Multiple round-trips for complex operations
- Status transitions require PATCH semantics

**Estimated Effort:** M

### Option 2: GraphQL API
**Pros:**
- Flexible queries
- Single endpoint
- Type-safe schema

**Cons:**
- Overkill for simple CRUD
- Learning curve for team
- Caching more complex

**Estimated Effort:** L

### Option 3: gRPC-only
**Pros:**
- High performance
- Strong typing with protobuf
- Bi-directional streaming

**Cons:**
- Not browser-friendly
- Debugging harder
- Requires gateway for REST clients

**Estimated Effort:** L

## Decision

**We will use RESTful Resource-Based API (Option 1) with Chi router.**

### Rationale
- REST is the simplest approach for CRUD operations
- Team already has REST experience
- Easy integration with frontend and third-party clients
- OpenAPI documentation tooling is mature
- Chi router is lightweight and idiomatic Go

## API Specification

### Endpoints

| Method | Path | Description | Success | Error |
|--------|------|-------------|---------|-------|
| POST | /api/v1/orders | Create order | 201 + Location | 400, 500 |
| GET | /api/v1/orders/:id | Get order by ID | 200 | 404, 500 |
| GET | /api/v1/orders | List orders (paginated) | 200 | 400, 500 |
| PATCH | /api/v1/orders/:id/status | Update status | 200 | 400, 404, 409, 500 |
| DELETE | /api/v1/orders/:id | Delete order | 204 | 404, 500 |
| GET | /healthz | Liveness probe | 200 | - |
| GET | /readyz | Readiness probe | 200 | 503 |

### Query Parameters for GET /api/v1/orders

| Parameter | Type | Default | Max | Description |
|-----------|------|---------|-----|-------------|
| limit | int | 20 | 100 | Items per page |
| offset | int | 0 | - | Pagination offset |
| status | string | - | - | Filter by status |

### Response Formats

**Order Response:**
```json
{
  "id": "uuid",
  "customer_id": "string",
  "items": [...],
  "status": "pending|confirmed|processing|shipped|delivered|cancelled",
  "total": 99.99,
  "version": 1,
  "created_at": "2026-02-14T12:00:00Z",
  "updated_at": "2026-02-14T12:00:00Z"
}
```

**List Response:**
```json
{
  "orders": [...],
  "total": 100,
  "limit": 20,
  "offset": 0
}
```

**Error Response:**
```json
{
  "error": "description",
  "code": "ERROR_CODE"
}
```

## Constraints

> **CRITICAL: All CONSTRAINT blocks are enforceable rules. Violations will break the build.**

### CONSTRAINT: All Endpoints Must Use /api/v1 Prefix
**BECAUSE:** API versioning allows backward-compatible evolution. The /api prefix distinguishes API routes from health checks.

**CHECK:** All order routes must start with `/api/v1/`. Verify with: `grep -r "r.Route\|r.Get\|r.Post\|r.Patch\|r.Delete" internal/handler/http/router.go | grep -v "/api/v1" | grep -v "healthz\|readyz"`

**Example:**
```go
// Good:
r.Route("/api/v1", func(r chi.Router) {
    r.Route("/orders", func(r chi.Router) {
        r.Post("/", h.CreateOrder)
    })
})

// Bad:
r.Post("/orders", h.CreateOrder)  // Missing /api/v1 prefix
```

### CONSTRAINT: POST Must Return 201 with Location Header
**BECAUSE:** HTTP semantics require 201 Created for successful resource creation, and Location header tells client where to find the new resource.

**CHECK:** POST handler sets `w.WriteHeader(http.StatusCreated)` and `w.Header().Set("Location", ...)`. Verify in handler tests.

**Example:**
```go
// Good:
w.Header().Set("Location", fmt.Sprintf("/api/v1/orders/%s", order.ID))
w.WriteHeader(http.StatusCreated)

// Bad:
w.WriteHeader(http.StatusOK)  // Wrong status code!
```

### CONSTRAINT: GET by ID Must Return 404 for Missing Orders
**BECAUSE:** HTTP 404 semantically means "resource not found". Using 200 with empty body or 500 is incorrect.

**CHECK:** Handler returns 404 when service returns ErrOrderNotFound. Verify with handler tests.

**Example:**
```go
// Good:
if errors.Is(err, domain.ErrOrderNotFound) {
    http.Error(w, "order not found", http.StatusNotFound)
    return
}

// Bad:
if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)  // Wrong for not found!
}
```

### CONSTRAINT: Every Request Must Be Logged with slog
**BECAUSE:** Structured logging is essential for debugging and observability. slog provides typed fields and JSON output.

**CHECK:** Logging middleware uses slog with method, path, status, duration. Verify middleware exists and is applied.

**Example:**
```go
// Good:
slog.Info("request completed",
    slog.String("method", r.Method),
    slog.String("path", r.URL.Path),
    slog.Int("status", status),
    slog.Duration("duration", duration),
)

// Bad:
log.Printf("request: %s %s")  // Unstructured!
```

### CONSTRAINT: Health Endpoints Must Not Require Authentication
**BECAUSE:** Kubernetes probes run without auth. Health checks must be accessible for infrastructure.

**CHECK:** /healthz and /readyz are mounted outside auth middleware. Verify in router setup.

**Example:**
```go
// Good:
r.Get("/healthz", h.Healthz)  // Outside auth middleware
r.Get("/readyz", h.Readyz)
r.Route("/api/v1", func(r chi.Router) {
    r.Use(authMiddleware)
    // ... protected routes
})

// Bad:
r.Group(func(r chi.Router) {
    r.Use(authMiddleware)
    r.Get("/healthz", h.Healthz)  // Health behind auth!
})
```

## Consequences

### Positive
- Clear, predictable API structure
- Easy to test with standard HTTP tools (curl, Postman)
- OpenAPI spec can be generated from handlers
- Health checks work with Kubernetes out of the box

### Negative
- Multiple round-trips for complex queries
- No real-time updates (would need WebSocket/SSE)

### Neutral
- Need to implement pagination logic

### Mitigations
- **Complex queries:** Add batch endpoints if needed in future
- **Real-time:** ADR-0004 will address event streaming if required

## Traceability

### Related Work
- **Jira Ticket:** N/A (foundational API design)
- **Pull Request:** [#1](https://github.com/nsridhar76/go-ordersvc/pull/1) - feat(docker): add docker-compose stack with integration tests
- **Parent ADR:** ADR-0001 (Clean Architecture)
- **Related ADRs:** ADR-0003 (Optimistic Locking for 409 responses)

### Implementation Subtasks

- [x] **Task 1:** Implement POST /api/v1/orders
  - **Acceptance:** Returns 201 + Location header, order JSON body
  - **Owner:** Backend Team
  - **Status:** Done

- [x] **Task 2:** Implement GET /api/v1/orders/:id
  - **Acceptance:** Returns 200 + order, or 404 if not found
  - **Owner:** Backend Team
  - **Status:** Done

- [x] **Task 3:** Implement GET /api/v1/orders with pagination
  - **Acceptance:** Supports ?status=pending&limit=20&offset=0
  - **Owner:** Backend Team
  - **Status:** Done

- [x] **Task 4:** Implement PATCH /api/v1/orders/:id/status
  - **Acceptance:** Valid transitions return 200, invalid return 400, conflicts return 409
  - **Owner:** Backend Team
  - **Status:** Done

- [x] **Task 5:** Implement health endpoints
  - **Acceptance:** /healthz returns 200, /readyz checks DB and returns 200 or 503
  - **Owner:** Backend Team
  - **Status:** Done

- [x] **Task 6:** Add structured logging middleware
  - **Acceptance:** Every request logged with slog (method, path, status, duration)
  - **Owner:** Backend Team
  - **Status:** Done

- [x] **Task 7:** Create multi-stage Dockerfile
  - **Acceptance:** Build with golang:1.24-alpine, run with distroless, arm64
  - **Owner:** Backend Team
  - **Status:** Done

### Updates
- **2026-02-14:** Initial creation for REST API design
- **2026-02-14:** All implementation subtasks completed
