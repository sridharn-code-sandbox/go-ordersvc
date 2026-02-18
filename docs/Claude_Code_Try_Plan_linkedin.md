<!-- LinkedIn Article Format — Copy/paste into LinkedIn "Write article" editor -->
<!-- Use LinkedIn's rich text editor. Apply formatting manually using the toolbar. -->
<!-- [H1] [H2] [H3] markers = heading levels to apply in LinkedIn's editor. -->
<!-- "Prompt:" sections = paste as-is (LinkedIn renders them as plain text). -->


[H1] Building a Go Microservice with Claude Code: A Hands-On Trial Plan

Go + Kubernetes + Docker + GitHub + Mac M4

A stepwise plan to exercise all 16 Claude Code cookbook recipes against a real Go microservice project.

Makefile-centric. CI tools for audits. Claude for architecture. ADRs for decisions.

---

[H2] The Project: go-ordersvc

A production-grade Go microservice for order management. REST + gRPC APIs, PostgreSQL, Redis cache, Kafka events, Dockerized, deployed to local K8s (Docker Desktop).

Three key principles drive this project:

1. A Makefile is the single interface for all project actions
2. Architecture Decision Records (ADRs) trace every technical choice back to a business reason
3. Drift detection ensures code never silently diverges from decisions

---

[H2] Design Philosophy

Three rules govern how Claude Code, tools, and humans work together in this project.


[H3] Rule 1: The Makefile is the Contract

Every project action lives in the Makefile. This gives you three things:

- Single source of truth. A new engineer types "make help" and sees every action available. No tribal knowledge.
- Claude uses it too. Instead of Claude guessing linter flags, it runs "make lint." Same flags, same output, every time.
- CI/CD mirrors local. Your GitHub Actions workflow calls the same make targets. If "make test" passes on your Mac, it passes in CI.


[H3] Rule 2: Tools Do Audits. Claude Does Thinking.

Deterministic checks belong to dedicated tools. They're faster, cheaper, and consistent. Claude's value is in the things tools can't do.

What tools handle (in Makefile + CI):
- Go lint + style → golangci-lint
- Go security scan → gosec
- Container CVEs → trivy
- Secret detection → gitleaks
- K8s manifests → kubeconform + polaris
- Dependency audit → govulncheck
- Code formatting → gofmt / goimports
- ADR structural drift → grep / make drift-check

What Claude handles:
- Architecture review (wrong abstraction boundaries)
- Business logic bugs
- Translating business tickets to implementation plans (/refine)
- ADR compliance judgment (/drift)
- Design alternatives


[H3] Rule 3: Decisions Are Documented, Not Verbal

Every architectural decision gets recorded as an ADR in docs/decisions/. This solves three problems:

1. Decisions evaporate. A PM asks "why does caching work this way?" six months later. The answer is in ADR-0012, not a Slack thread nobody can find.
2. Code drifts from intent. Someone moves caching to the handler layer. The ADR constraint says service-layer only, make drift-check fails in CI.
3. Business intent gets lost. The PM wrote "orders page is slow." The ADR traces the Redis decision back to the NPS drop from 42 to 31.

The traceability chain: Code line → test case → acceptance criteria → subtask → ADR constraint → ADR business context → Jira ticket → user problem. Every line of code is traceable back to a business decision.

The rule of thumb:
- Deterministic yes/no answer → put it in the Makefile and CI
- Requires judgment, context, or architectural understanding → Claude's job
- Decision made for a business reason → ADR with a CHECK rule

---

[H2] Plan Overview

The plan has 5 steps. Each step has 3-5 blocks. Each block maps to a cookbook recipe. You get: exact prompts, make targets to run, and a checkpoint.

Step 1: Foundation + Makefile + ADRs (Recipes #01-#04)
What you build: CLAUDE.md, Makefile, ADR template, scaffold, TDD

Step 2: Business Intent → Code (Recipes #05-#08)
What you build: /refine workflow, REST API, Docker, ADR-0002, docs

Step 3: K8s + Agent Patterns (Recipes #09-#12)
What you build: K8s deploy, skills (incl. /drift), agents, MCP

Step 4: 2.0 Features Deep Dive (Recipes #13-#16)
What you build: Checkpoints, hooks, context, drift-check in CI

Step 5: Integration + Ship (All combined)
What you build: gRPC, Kafka, ADRs for each, full CI + drift pipeline

---

[H2] Step 1: Foundation + Makefile + ADRs


[H3] Block 1.1 — CLAUDE.md + Makefile + Doc Scaffolding (Recipe #01)

Goal: Claude knows your project, your Makefile is the command contract, and docs/decisions/ is ready.

Launch Claude Code and generate CLAUDE.md, Makefile, and documentation scaffolding together. The CLAUDE.md should cover your architecture (Clean Architecture: handler > service > repo), stack (Chi router, pgx, go-redis, protobuf/gRPC), infra (Docker Desktop K8s on Mac M4), and testing strategy (table-driven tests with testify).

Key additions:
- A "Project Commands" section listing all Makefile targets with the rule: "Always use make targets. Never run tools directly."
- An "Architecture Decision Records" section with rules: "Before implementing any feature, check docs/decisions/. Follow all CONSTRAINT rules."
- Skills: /refine (translates business tickets to plans + ADRs) and /drift (checks ADR compliance)

Checkpoint: CLAUDE.md, Makefile, ADR template, /refine skill, /drift skill all exist. "make help" shows all targets including drift-check.


[H3] Block 1.2 — Plan Before You Build (Recipe #02)

Goal: Architect the service before writing code.

Use Plan Mode (Shift+Tab or /plan) to design the project structure. Map directory structure (/cmd, /internal, /api, /deploy), list all files, define interfaces between layers. Don't write code yet.

Review and challenge the plan, then approve and implement the skeleton.

Checkpoint: Project structure exists. "make build" succeeds.


[H3] Block 1.2b — First ADR: Architecture Decision

Goal: Record the clean architecture decision before writing any business logic.

Why now? The clean architecture decision was just made in Plan Mode. If you don't record it now, in three months someone will put SQL queries in a handler and nobody will remember why the layers exist.

Generate ADR-0001 with constraints like:
- Handlers must not import repository or database packages (CHECK: no pgx/sql imports in internal/handler/)
- Service layer must depend on repository interfaces, not concrete implementations
- Repository layer must not import handler or service packages

Add the first deterministic check to the Makefile drift-check target.

Checkpoint: ADR-0001 committed. "make drift-check" passes with at least one real constraint check.


[H3] Block 1.3 — Test-Driven Everything (Recipe #03)

Goal: Tests exist before implementation. "make test" is the feedback loop.

Write comprehensive failing tests for the order service layer using table-driven tests. Cover CreateOrder, GetOrder, ListOrders, UpdateOrderStatus with edge cases. Use mock repositories (interface-based, per ADR-0001).

Then implement the service layer to pass all tests.

Checkpoint: "make test" passes. Coverage > 80%. Service uses repo interfaces (per ADR-0001).


[H3] Block 1.4 — Model Switching (Recipe #04)

Goal: Feel the cost/quality trade-off.

Use /model sonnet for mechanical tasks (adding copyright headers). Use /model opus for deep analysis (reviewing for race conditions).

Checkpoint: You understand when to use which model. Sonnet for mechanical, Opus for analysis.

---

[H2] Step 2: Business Intent → Code


[H3] Block 2.0 — Business Ticket → /refine → ADR (Full Workflow)

Goal: Experience the complete business-to-code translation workflow.

Simulate a PM's Jira ticket written in pure business language:

"ORD-100: Customers can't see their order details. 40% of support tickets are 'where is my order' inquiries. Each costs $8 to handle. That's $50K/month we can eliminate."

Claude reads the codebase, proposes an implementation plan mapping business outcomes to technical changes, breaks it into subtasks with Given/When/Then acceptance criteria.

Claude generates ADR-0002 with constraints like:
- GET /orders/:id must return 404 (not 500) for missing orders. BECAUSE: 500 errors trigger customer support calls — the exact problem we're solving.
- List endpoint must support status filter. BECAUSE: "where is my order" requires filtering by "shipped"/"delivered."

Checkpoint: ADR-0002 exists. Constraints trace to business outcomes.


[H3] Block 2.1 — REST API + Docker (Recipe #08)

Goal: Implement the feature per ADR-0002 constraints.

Implement REST API with acceptance criteria: POST creates orders (201), GET returns order (404 if missing), GET with pagination and status filter, PATCH for status transitions, health checks. Create multi-stage Dockerfile for ARM64.

Observe: Claude should reference ADR-0002 constraints during implementation. If it puts DB access in a handler, make drift-check catches the ADR-0001 violation.

Checkpoint: "make docker" + "make scan" + "make drift-check" pass.


[H3] Block 2.3 — PR Automation + docker-compose (Recipe #06)

Goal: Claude creates branch, docker-compose, PR.

Create docker-compose.yml with go-ordersvc, postgres, and redis. Wire into Makefile. Add integration tests. Create PR with ADR constraint references.

Checkpoint: PR on GitHub with ADR references.


[H3] Block 2.4 — Generated Docs + ADR Update (Recipe #07)

Goal: Docs generated from code. ADR-0002 updated with actual PR number.

Generate docs/API.md from actual handler code. Generate docs/ARCHITECTURE.md. Update ADR-0002 Traceability section.

Checkpoint: Docs match actual code. ADR traceability complete.

---

[H2] Step 3: K8s + Advanced Agent Patterns


[H3] Block 3.1 — Skills Including /drift (Recipe #10)

Goal: Reusable skills that complement Makefile tools.

Add /review (architecture review after make lint passes) and /k8s-review (K8s architecture review after make k8s-lint passes). Test /drift against existing ADRs.

Checkpoint: Four skills work. /drift reports PASS on all constraints.


[H3] Block 3.2 — Deploy to K8s (Recipes #08 + #10)

Goal: Service on Docker Desktop K8s.

Create Helm chart with Deployment (2 replicas, resource limits, probes), Service, ConfigMap, HPA, Postgres and Redis subcharts.

Validate with tools first (make k8s-lint), then Claude (/k8s-review), then deploy (make k8s-deploy).

Checkpoint: "make k8s-lint" passes. Pods running.


[H3] Block 3.3 — Parallel Feature Dev (Recipe #09)

Goal: Two features, two worktrees, simultaneous.

Use git worktrees to develop Redis caching and rate limiting in parallel. Merge both, run full validation with make ci.

Checkpoint: Both merged. "make ci" passes on combined code.


[H3] Block 3.3b — Record Feature Decisions (ADRs)

Goal: Cache and rate limit decisions are recorded and enforceable.

Why record these now? You just made two non-obvious decisions: caching at service layer (not handler), and rate limiting at middleware level (not per-handler). Six months from now, someone will want to move caching to the handler "for simplicity." The ADR explains why and the drift-check prevents the move.

Generate ADR-0003 (Redis caching) and ADR-0004 (rate limiting) with CHECK rules. Add deterministic checks to make drift-check.

Checkpoint: ADR-0003 + ADR-0004 committed. /drift reports all PASS.


[H3] Block 3.4 — Sub-agents + MCP (Recipes #11, #12)

Goal: Claude reviews architecture. Tools handle the rest.

Create a security architecture reviewer sub-agent (read-only, finds what gosec can't). Add Postgres MCP for schema inspection — cross-reference actual DB schema with ADR-0002 constraints.

Checkpoint: Security reviewer finds logic issues. MCP reads actual DB schema.

---

[H2] Step 4: Claude Code 2.0 Features + Drift Enforcement


[H3] Block 4.1 — Checkpoints (Recipe #13)

Goal: Experiment boldly, rewind safely.

Refactor the repository layer. If it doesn't feel right, Esc Esc to rewind and try a different approach with what you learned.

Checkpoint: You've rewound at least once. Final implementation passes.


[H3] Block 4.2 — Hooks + Drift Enforcement (Recipe #14)

Goal: Hooks call Makefile targets including drift-check.

Configure hooks in .claude/settings.json:
- PostToolUse (Write|Edit) → make fmt (auto-format on every edit)
- Stop → make drift-check (ADR constraints verified every time Claude finishes)
- PreToolUse (git push) → make lint && make drift-check (lint + drift gate before push)

The Stop hook with make drift-check is the key addition. Claude can never silently violate an ADR constraint.

Test it: Intentionally add a pgx import to a handler. The Stop hook fires make drift-check and catches the ADR-0001 violation.

Checkpoint: Hooks fire correctly. Intentional drift caught and fixed.


[H3] Block 4.3 — Context Management (Recipe #15)

Goal: Understand and control your context budget.

Use /context to check usage, /compact to free space with targeted focus. Delegate heavy analysis to sub-agents with isolated context.

Checkpoint: Context managed. Sub-agent returned summary.


[H3] Block 4.4 — Plugins (Recipe #16)

Goal: Your setup is already shareable.

Your .claude/ directory IS a plugin: skills, agents, hooks. A teammate clones this repo and gets EVERYTHING.

Checkpoint: .claude/ + docs/decisions/ is a complete, shareable system.

---

[H2] Step 5: Integration + Ship


[H3] Block 5.1 — gRPC Layer

Start with business intent: "ORD-201: Internal systems need real-time order access. Warehouse picks are delayed by up to 5 seconds. REST polling is 60% of our API traffic."

Claude proposes gRPC with server streaming. Generate ADR-0005 with constraints ensuring gRPC handlers follow the same clean architecture layers as REST (per ADR-0001).

Checkpoint: ADR-0005 committed. gRPC tests pass. make drift-check green.


[H3] Block 5.2 — Kafka Events

Business intent: "ORD-202: Downstream systems need order event feed. Analytics and billing query our DB directly, causing tight coupling."

Generate ADR-0006 for Kafka. Implement in a worktree. Run security reviewer and /drift after merge.

Checkpoint: ADR-0006 committed. /drift reports all PASS across all 6 ADRs.


[H3] Block 5.3 — GitHub Actions CI (The Payoff)

Goal: CI calls the same Makefile targets including drift-check.

Create .github/workflows/ci.yml with jobs for lint, security, test, build, k8s, scan, and drift. Every CI step uses a make target. The drift job ensures no PR can merge if it violates an ADR.

The full loop:
PM writes Jira ticket (business language) → Engineer runs /refine (Claude translates) → Engineer approves, Claude generates ADR (decisions recorded) → Claude implements (reads ADR constraints) → Stop hook: make drift-check (immediate feedback) → make ci locally (full validation) → GitHub Actions: same make targets (blocks PR on drift) → /drift quarterly (full judgment audit)

Decisions are traceable. Drift is caught automatically. Business intent is preserved.

Checkpoint: CI green including drift. 6 ADRs committed. Docs current. Pods running on K8s.

---

[H2] The Final Architecture

How tools, Claude, the Makefile, and ADRs work together:

You / Claude / Hooks / GitHub / /drift — all flow through the Makefile as the single source of truth.

TOOLS (fast, free): golangci-lint, gosec, trivy, gitleaks, govulncheck, kubeconform, gofmt, make drift-check

CLAUDE (judgment): Architecture review, business logic bugs, /refine translation, /drift judgment, design alternatives

ADRs: Why this way and not that way

---

[H2] ADR Inventory (End State)

ADR-0001: Clean Architecture — No DB in handlers, interface deps, layer isolation
ADR-0002: Order Details API — 404 not 500, status filter, pagination
ADR-0003: Redis Caching — Service-layer only, configurable TTL, fallback to DB
ADR-0004: Rate Limiting — Middleware-level, configurable limits, Retry-After
ADR-0005: gRPC API — Same layers as REST, proto file location, streaming
ADR-0006: Kafka Events — Service-layer publishing, at-least-once, idempotent

---

[H2] What's Next

After completing all steps, consider adding:

- Jira MCP server — /refine reads tickets directly instead of copy-paste
- ADR-aware PR template — requires listing relevant ADRs and constraint status
- Quarterly /drift audit — schedule runs before each release
- Monitoring — Prometheus metrics + Grafana with an ADR for observability strategy
- ArgoCD — GitOps deployment, Helm chart auto-deploys on merge
- Agent Teams — experimental Opus 4.6 feature for parallel agents
- ADR dashboard — script that reports constraint count, drift-check coverage, and ADR age

---

[H2] Directory Structure

go-ordersvc/
  CLAUDE.md — Points Claude to ADRs + Makefile
  Makefile — build, test, lint, drift-check, ci
  docs/decisions/ — ADR-0001 through ADR-0006 + template
  docs/API.md — Generated from code
  docs/ARCHITECTURE.md — Generated from code
  .claude/skills/ — refine, drift, review, k8s-review
  .claude/agents/ — security-reviewer
  .claude/settings.json — Hooks (fmt, lint, drift-check)
  .github/workflows/ci.yml — make ci + make drift-check
  internal/ — application code

---

#ClaudeCode #Go #Golang #Kubernetes #Docker #Microservices #AI #SoftwareEngineering #DevOps #CleanArchitecture #ADR #DeveloperExperience #AIAgents #CodingWithAI
