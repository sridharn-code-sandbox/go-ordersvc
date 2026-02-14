# Architecture Decision Records (ADRs)

## Purpose

This directory contains all architectural decisions for go-ordersvc. Every significant architectural choice must be documented here before implementation.

## When to Create an ADR

Create an ADR when:
- Choosing between multiple technical approaches
- Making decisions that are hard to reverse
- Setting constraints that affect future development
- Standardizing patterns across the codebase
- Changing existing architectural decisions

## Naming Convention

```
ADR-NNNN-short-descriptive-title.md
```

Examples:
- `ADR-0001-use-optimistic-locking.md`
- `ADR-0002-postgres-jsonb-for-order-items.md`
- `ADR-0003-chi-router-over-gin.md`

## Creating a New ADR

1. Copy `TEMPLATE.md` to a new file with the next sequential number
2. Fill in all sections, especially:
   - Business Context (the "why")
   - Options Considered (show your work)
   - CONSTRAINT blocks (enforceable rules)
   - Traceability (link to Jira/PR)
3. Submit for review as part of your PR
4. Update status to "Accepted" after approval

## CONSTRAINT Format

ADRs can define enforceable constraints using this format:

```markdown
### CONSTRAINT: [Rule Name]
**BECAUSE:** [Rationale]
**CHECK:** [How to verify]
```

**Example:**

```markdown
### CONSTRAINT: Service Layer Must Not Import HTTP Handlers
**BECAUSE:** Violates Clean Architecture dependency rules

**CHECK:** Run `go list -json ./... | jq -r '.Deps[]' | grep handler/http` in service package
```

## Integration with Development Workflow

### Before Implementation
1. Search existing ADRs for related decisions
2. Read all CONSTRAINT blocks in relevant ADRs
3. If constraints conflict with requirements → STOP and ask

### During Implementation
1. Follow all CONSTRAINT rules
2. Add tests to verify CHECK conditions
3. Update ADR Traceability section with PR link

### CI/CD
- `make drift-check` validates CONSTRAINT compliance (future)
- CI fails if constraints are violated

## Index

<!-- Maintain this index manually or with a script -->

| Number | Title | Status | Date |
|--------|-------|--------|------|
| 0001 | Use Clean Architecture with Handler → Service → Repository Layers | Accepted | 2026-02-14 |
| 0002 | Order Details REST API Design | Accepted | 2026-02-14 |
| 0003 | Use Optimistic Locking for Concurrent Updates | Accepted | 2026-02-14 |
| TEMPLATE | ADR Template | - | - |

## Resources

- [ADR GitHub](https://adr.github.io/)
- [Documenting Architecture Decisions](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions)
