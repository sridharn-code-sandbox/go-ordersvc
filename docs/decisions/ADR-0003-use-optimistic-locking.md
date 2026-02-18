# ADR-0003: Use Optimistic Locking for Concurrent Order Updates

## Status

**Accepted**

**Date:** 2026-02-14

## Business Context

### Problem Statement
Multiple API requests can attempt to modify the same order simultaneously (e.g., changing status from "pending" to "confirmed" while another request cancels it). Without concurrency control, the last write wins, potentially creating invalid state transitions.

### Success Criteria
- No lost updates: All concurrent modifications are detected
- No invalid state transitions: Status machine rules are enforced atomically
- Performance: Minimal overhead compared to no locking
- Scalability: Works across multiple service instances

### Constraints
- Must work with PostgreSQL
- Cannot require distributed lock service (keep infra simple)
- Must have <5ms latency impact per operation

## Options Considered

### Option 1: Pessimistic Locking (SELECT FOR UPDATE)
**Pros:**
- Guaranteed to prevent conflicts
- Simple to understand

**Cons:**
- Locks held across network round-trips
- Deadlock risk with multiple orders
- Poor performance under high concurrency

**Estimated Effort:** S

### Option 2: Optimistic Locking (Version Field)
**Pros:**
- No locks - better concurrency
- Conflicts detected at write time
- Database-agnostic pattern
- Low latency impact (~1ms)

**Cons:**
- Requires retry logic in callers
- Version field in every table

**Estimated Effort:** S

### Option 3: Event Sourcing
**Pros:**
- Complete audit trail
- Natural concurrency handling

**Cons:**
- Major architectural shift
- Complexity overkill for current needs
- Query complexity increases

**Estimated Effort:** XL

## Decision

**We will use optimistic locking (Option 2) with a `version` integer field.**

### Rationale
- Optimistic locking provides the right balance of correctness and performance
- Version field is simple to implement and understand
- Caller retry logic is already needed for network failures
- Keeps infrastructure simple (no distributed locks)
- Meets latency requirements (<5ms measured in tests)

## Constraints

### CONSTRAINT: All Update Operations Must Check Version
**BECAUSE:** Skipping version checks defeats concurrency control and allows lost updates

**CHECK:**
```sql
-- Every UPDATE must include version in WHERE clause
UPDATE orders SET ..., version = version + 1
WHERE id = $1 AND version = $2
```

Verify with: `grep -r "UPDATE orders" internal/repository/postgres/`

**Example:**
```go
// Good:
UPDATE orders
SET status = $1, version = version + 1, updated_at = NOW()
WHERE id = $2 AND version = $3

// Bad:
UPDATE orders
SET status = $1, updated_at = NOW()
WHERE id = $2
-- Missing version check!
```

### CONSTRAINT: Service Layer Must Return ErrConcurrentModification
**BECAUSE:** Callers need to distinguish version conflicts from other errors to implement retry logic

**CHECK:** Repository Update() returns `domain.ErrConcurrentModification` when `RowsAffected() == 0`

Verify with unit tests in `internal/service/*_test.go`

**Example:**
```go
// Good:
if result.RowsAffected() == 0 {
    return domain.ErrConcurrentModification
}

// Bad:
if result.RowsAffected() == 0 {
    return errors.New("update failed")  // Loses semantic meaning
}
```

### CONSTRAINT: Version Field Must Be Non-Nullable Integer Starting at 1
**BECAUSE:** NULL versions break comparisons; starting at 1 makes "version 0" semantically mean "uninitialized"

**CHECK:** Database migration enforces `version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0)`

**Example:**
```sql
-- Good:
version INTEGER NOT NULL DEFAULT 1,
CONSTRAINT positive_version CHECK (version > 0)

-- Bad:
version INTEGER  -- Nullable!
```

## Consequences

### Positive
- Race conditions eliminated (verified by 5 new test cases)
- Low latency impact: 0.8ms average overhead measured
- Simple mental model for developers
- Works with existing PostgreSQL setup

### Negative
- API clients must handle 409 Conflict responses
- Version field in every entity table
- Read-modify-write pattern required (no direct field updates)

### Neutral
- Test complexity increases (need concurrent test scenarios)

### Mitigations
- **Client retry:** Provide SDK with automatic retry for 409 responses
- **Documentation:** Add retry examples to API docs
- **Test coverage:** 33 tests including 5 optimistic locking scenarios

## Traceability

### Related Work
- **Jira Ticket:** [PROJ-001](https://jira.example.com/PROJ-001) - "Prevent race conditions in order updates"
- **Pull Request:** [#2](https://github.com/sridharn-code-sandbox/go-ordersvc/pull/2) (day 1 commit)
- **Parent ADR:** ADR-0001 (Clean Architecture)
- **Related ADRs:** Builds on layer separation from ADR-0001

### Implementation Subtasks

- [x] **Task 1:** Add version field to domain.Order
  - **Acceptance:** `type Order struct` includes `Version int`
  - **Owner:** Backend Team
  - **Status:** Done (commit e45cc2a)

- [x] **Task 2:** Update database migration with version column
  - **Acceptance:** `000001_create_orders_table.up.sql` includes version with CHECK constraint
  - **Owner:** Backend Team
  - **Status:** Done (commit e45cc2a)

- [x] **Task 3:** Implement version checking in Postgres repository
  - **Acceptance:** UPDATE queries include `WHERE version = $N` and return ErrConcurrentModification
  - **Owner:** Backend Team
  - **Status:** Done (commit e45cc2a)

- [x] **Task 4:** Add optimistic locking tests
  - **Acceptance:** 5 tests cover concurrent modification scenarios
  - **Owner:** Backend Team
  - **Status:** Done (commit e45cc2a)

- [ ] **Task 5:** Add retry logic to API clients
  - **Acceptance:** SDK retries on 409 with exponential backoff
  - **Owner:** SDK Team
  - **Status:** Not Started

- [ ] **Task 6:** Document retry strategy in API docs
  - **Acceptance:** OpenAPI spec includes 409 response and retry guidance
  - **Owner:** Docs Team
  - **Status:** Not Started

### Updates
- **2026-02-14:** Initial creation during day 1 implementation
- **2026-02-14:** Status set to Accepted after implementation and testing
- **2026-02-14:** Renumbered from ADR-0001 to ADR-0002 to make room for Clean Architecture as foundational decision
- **2026-02-14:** Renumbered from ADR-0002 to ADR-0003 to make room for Order Details API
