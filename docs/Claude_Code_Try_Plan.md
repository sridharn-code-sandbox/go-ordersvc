# Claude Code 2.0 Cookbook — Hands-On Trial Plan

**Go · Kubernetes · Docker · GitHub · Mac M4**

*A stepwise  plan to exercise all [16 cookbook recipes](Claude_Code_Cookbook.md) against a real Go microservice project.*

*Makefile-centric • CI tools for audits • Claude for architecture • ADRs for decisions*

*February 2026 — v3*

> **🎯 PROJECT: go-ordersvc**
>
> A production-grade Go microservice for order management. REST + gRPC APIs, PostgreSQL, Redis cache, Kafka events, Dockerized, deployed to local K8s (Docker Desktop). A Makefile is the single interface for all project actions. Architecture Decision Records (ADRs) trace every technical choice back to a business reason. Drift detection ensures code never silently diverges from decisions.

---

## Design Philosophy

### The Makefile is the Contract

Every project action lives in the Makefile. This gives you three things:

1. **Single source of truth.** A new engineer types `make help` and sees every action available. No tribal knowledge.
2. **Claude uses it too.** Instead of Claude guessing linter flags, it runs `make lint`. Same flags, same output, every time.
3. **CI/CD mirrors local.** Your GitHub Actions workflow calls the same make targets. If `make test` passes on your Mac, it passes in CI.

### Tools Do Audits. Claude Does Thinking.

Deterministic checks belong to dedicated tools. They're faster, cheaper, and consistent. Claude's value is in the things tools can't do.

| Check | Tool (in Makefile + CI) | Why Not Claude |
|-------|------------------------|----------------|
| Go lint + style | golangci-lint | Faster, free, consistent, catches 100+ patterns |
| Go security scan | gosec | CVE database, SARIF output for GitHub Security tab |
| Container CVEs | trivy image scan | Scans OS packages + Go deps against NVD |
| Secret detection | gitleaks | Regex + entropy detection, git-history-aware |
| K8s manifests | kubeconform + polaris | Schema validation + best-practice scoring |
| Helm charts | helm lint + template | Catches YAML errors, missing values |
| Dependency audit | govulncheck | Go's official vulnerability database |
| Code formatting | gofmt / goimports | Deterministic, zero-cost |
| ADR structural drift | grep / make drift-check | Import boundaries, hardcoded values |

### Decisions Are Documented, Not Verbal

Every architectural decision gets recorded as an ADR (Architecture Decision Record) in `docs/decisions/`. This solves three problems:

1. **Decisions evaporate.** A PM asks "why does caching work this way?" six months later. The answer is in ADR-0012, not a Slack thread nobody can find.
2. **Code drifts from intent.** Someone moves caching to the handler layer. The ADR constraint says service-layer only, `make drift-check` fails in CI.
3. **Business intent gets lost.** The PM wrote "orders page is slow." The ADR traces the Redis decision back to the NPS drop from 42 to 31.

**The workflow for every feature:**

```
PM writes Jira ticket     Engineer runs /refine     Claude generates ADR
(business language)   →   (translates to plan)  →   (records decisions)
                                                          │
  make drift-check         PR passes CI              Claude implements
  (enforces ADR rules) ←   (tools + drift check) ←   (reads ADR constraints)
```

> **🔑 TRACEABILITY CHAIN**
>
> Code line → test case → acceptance criteria → subtask → ADR constraint → ADR business context → Jira ticket → user problem.
>
> Every line of code is traceable back to a business decision.

> **💡 THE RULE OF THUMB**
>
> - Deterministic yes/no answer → put it in the Makefile and CI.
> - Requires judgment, context, or architectural understanding → Claude's job.
> - Decision made for a business reason → ADR with a CHECK rule.
>
> Claude's `/review` skill runs AFTER `make ci` passes, so it focuses on what machines miss.

---

<details>
<summary><h2>The Makefile</h2></summary>

*Created in Step 1. Used in every step after. Includes drift-check for ADR enforcement.*

```makefile
.PHONY: help build test lint sec scan fmt vet vuln docker docker-push
.PHONY: k8s-deploy k8s-delete k8s-status compose-up compose-down
.PHONY: proto docs clean ci drift-check

APP         := go-ordersvc
VERSION     := $(shell git describe --tags --always --dirty)
REGISTRY    := ghcr.io/$(shell gh api user -q .login)
IMAGE       := $(REGISTRY)/$(APP):$(VERSION)
HELM_CHART  := deploy/helm/$(APP)
K8S_NS      := default

# ──── Core ────

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-16s\033[0m %s\n", $$1, $$2}'

build: ## Build the Go binary
	CGO_ENABLED=0 GOARCH=arm64 go build -ldflags "-s -w \
	  -X main.version=$(VERSION)" -o bin/$(APP) ./cmd/server

run: build ## Run locally
	./bin/$(APP)

clean: ## Remove build artifacts
	rm -rf bin/ coverage.out

# ──── Quality (deterministic checks) ────

fmt: ## Format code (gofmt + goimports)
	gofmt -w .
	goimports -w .

vet: ## Go vet
	go vet ./...

lint: fmt vet ## Lint (golangci-lint)
	golangci-lint run ./...

sec: ## Security scan (gosec)
	gosec -quiet ./...

vuln: ## Dependency vulnerability check (govulncheck)
	govulncheck ./...

scan: docker ## Container image scan (trivy)
	trivy image --severity HIGH,CRITICAL $(IMAGE)

secrets: ## Detect leaked secrets (gitleaks)
	gitleaks detect --source . --verbose

# ──── Testing ────

test: ## Run tests with race detector
	go test -race -count=1 -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

test-integration: compose-up ## Run integration tests
	go test -race -tags=integration -count=1 ./...
	$(MAKE) compose-down

# ──── Docker ────

docker: ## Build Docker image (ARM64)
	docker build --platform linux/arm64 -t $(IMAGE) .

compose-up: ## Start local stack (Postgres + Redis + Kafka)
	docker compose up -d --wait

compose-down: ## Stop local stack
	docker compose down -v

# ──── Kubernetes ────

k8s-lint: ## Validate Helm chart
	helm lint $(HELM_CHART)
	helm template $(APP) $(HELM_CHART) | kubeconform -strict

k8s-deploy: docker ## Deploy to Docker Desktop K8s
	kubectl config use-context docker-desktop
	helm upgrade --install $(APP) $(HELM_CHART) \
	  --namespace $(K8S_NS) --set image.tag=$(VERSION)

k8s-status: ## Show pod status
	kubectl get pods,svc,hpa -n $(K8S_NS) -l app=$(APP)

# ──── Protobuf ────

proto: ## Generate Go code from proto files
	protoc --go_out=. --go-grpc_out=. api/proto/orders/v1/*.proto

# ──── Documentation & Drift ────

drift-check: ## Check deterministic ADR constraints
	@echo 'Checking ADR constraints...'
	@# Each line maps to a CONSTRAINT/CHECK in an ADR
	@# ADR-0001: Clean Architecture layer boundaries
	@if grep -r 'database/sql\|pgx' internal/handler/; then \
	  echo '\033[31mDRIFT: ADR-0001 - DB import in handler layer\033[0m'; \
	  exit 1; fi
	@echo '\033[32m✔ All deterministic ADR constraints pass\033[0m'

# ──── CI aggregate ────

ci: lint sec vuln test k8s-lint drift-check ## Run full CI suite
	@echo '\033[32m✔ All CI checks passed\033[0m'
```

> **KEY DESIGN DECISIONS**
>
> - `make ci` includes `drift-check` — ADR constraints are enforced on every CI run.
> - `drift-check` grows over time: each new ADR with a deterministic CHECK adds a grep line.
> - All quality targets are deterministic tools, not Claude. Claude's job is `/review` and `/drift` after `make ci` passes.
> - `make help` is self-documenting — every target has a `##` comment.

### Tell Claude About the Makefile and ADRs

Add these sections to CLAUDE.md so Claude uses Makefile targets and respects ADRs:

```markdown
## Project Commands (always use these, never invent alternatives)

make build          # Build binary
make test           # Run tests with race detector + coverage
make lint           # Format + vet + golangci-lint
make sec            # Security scan (gosec)
make vuln           # Dependency vulnerabilities (govulncheck)
make docker         # Build Docker image
make scan           # Trivy scan on Docker image
make compose-up     # Start Postgres + Redis + Kafka locally
make k8s-deploy     # Deploy to Docker Desktop K8s
make k8s-lint       # Validate Helm charts
make drift-check    # Verify code conforms to ADR constraints
make ci             # Full CI suite (run before every PR)

## Workflow
- After writing code: run `make lint` then `make test`
- Before creating a PR: run `make ci` (includes drift-check)
- Never run gofmt, golangci-lint, or gosec directly — use make

## Architecture Decision Records
- Before implementing any feature, check docs/decisions/ for ADRs
- Follow all CONSTRAINT rules in relevant ADRs
- If a constraint conflicts with the task, STOP and ask
- After implementation, update the ADR Traceability section
- ADR constraints are non-negotiable without engineer approval
```

> **💡 WHY THIS MATTERS**
>
> Without the ADR rules in CLAUDE.md, Claude might move caching from the service layer to the handler layer because "it's simpler." With the ADR rules, Claude reads the constraint first and keeps caching where the architectural decision says it belongs.

</details>

---

## Before You Start

*One-time setup. Do this before starting Step 1.*

### 1. Install prerequisites

```bash
# Claude Code (requires Node.js 18+)
npm install -g @anthropic-ai/claude-code
claude --version    # should show 2.x

# Go 1.22+
brew install go

# Docker Desktop (enable K8s in Settings > Kubernetes)
# Allocate: 4 CPUs, 8GB RAM minimum

# CLI tools
brew install kubectl helm gh protobuf
brew install golangci-lint gosec trivy gitleaks kubeconform
go install golang.org/x/vuln/cmd/govulncheck@latest
go install golang.org/x/tools/cmd/goimports@latest

# GitHub CLI auth
gh auth login
```

### 2. Create the GitHub repo

```bash
mkdir ~/projects/go-ordersvc && cd ~/projects/go-ordersvc
go mod init github.com/<your-user>/go-ordersvc
git init
gh repo create go-ordersvc --public --source=. --remote=origin
echo '# go-ordersvc' > README.md
git add . && git commit -m 'init' && git push -u origin main
```

### 3. Verify your Claude Code plan

You need at least Pro ($20/mo) for Opus 4.6 access. Max ($100/mo) recommended for heavy use.

```bash
claude
/usage    # check plan and remaining messages
/model    # verify Opus 4.6 available
```

> **⚠️ MAC M4 + DOCKER DESKTOP NOTES**
>
> Docker Desktop on Apple Silicon runs K8s via a lightweight Linux VM. ARM64 images build natively (no emulation). Allocate at least 4 CPUs and 8GB RAM to Docker Desktop for comfortable K8s + Postgres + Redis + Kafka.

---

## Plan Overview

| Step | Theme | Recipes | What You Build |
|------|-------|---------|----------------|
| 1 | Foundation + Makefile + ADRs | #01, #02, #03, #04 | CLAUDE.md, Makefile, ADR template, scaffold, TDD |
| 2 | Business Intent → Code | #05, #06, #07, #08 | /refine workflow, REST API, Docker, ADR-0002, docs |
| 3 | K8s + Agent Patterns | #09, #10, #11, #12 | K8s deploy, skills (incl. /drift), agents, MCP |
| 4 | 2.0 Features Deep Dive | #13, #14, #15, #16 | Checkpoints, hooks, context, drift-check in CI |
| 5 | Integration + Ship | All combined | gRPC, Kafka, ADRs for each, full CI + drift pipeline |

> **💡 HOW TO USE THIS PLAN**
>
> Each step has 3–5 blocks. Each block maps to a cookbook recipe. You get: exact prompts, make targets to run, and a checkpoint.
>
> - 📝 blocks are documentation steps — they generate or verify ADRs.
> - ⭐ blocks are essential if short on time.

---

## Step 1: Foundation + Makefile + ADRs

*Recipes #01, #02, #03, #04*

### ⭐ Block 1.1 — CLAUDE.md + Makefile + Doc Scaffolding ([Recipe #01](Claude_Code_Cookbook.md))

*Goal: Claude knows your project, your Makefile is the command contract, and docs/decisions/ is ready.*

**Step 1:** Launch Claude Code and generate CLAUDE.md, Makefile, and documentation scaffolding together.

```
cd ~/projects/go-ordersvc
claude

This is a new Go project. Generate these files:

1. CLAUDE.md covering:
   - Language: Go 1.22, module: github.com/<user>/go-ordersvc
   - Architecture: Clean Architecture (handler > service > repo)
   - Stack: Chi router, pgx for Postgres, go-redis, protobuf/gRPC
   - Infra: Docker Desktop K8s on Mac M4 (ARM64), Helm charts
   - Testing: table-driven tests with testify, race detector
   - 'Project Commands' section listing all Makefile targets.
     Rule: 'Always use make targets. Never run tools directly.'
   - 'Architecture Decision Records' section with rules:
     'Before implementing any feature, check docs/decisions/.
      Follow all CONSTRAINT rules. If a constraint conflicts,
      STOP and ask. After implementation, update Traceability.'
   - Compact Instructions: preserve architecture decisions and
     API contracts during compaction

2. Makefile with all targets from the plan (build, test, lint,
   sec, vuln, docker, scan, compose, k8s, proto, drift-check, ci).
   drift-check should be empty for now with a placeholder comment.
   ci target must include drift-check.

3. docs/decisions/TEMPLATE.md - ADR template with sections:
   Status, Date, Business Context, Options Considered,
   Decision, Constraints (CONSTRAINT/BECAUSE/CHECK format),
   Consequences, Traceability (Jira/PR/Subtasks).

4. .claude/skills/refine/SKILL.md - Skill to translate business
   Jira tickets into implementation plans + ADRs. Process:
   understand business outcome, analyze codebase, propose 2-3
   options, recommend one, break into subtasks with acceptance
   criteria, generate ADR after engineer approves.

5. .claude/skills/drift/SKILL.md - Skill to check all accepted
   ADRs' CONSTRAINT/CHECK rules against current code. Reports
   PASS, DRIFT (with business impact from BECAUSE), or UNCLEAR.
   Never modifies files. Read-only audit.
```

**Step 2:** Verify the scaffolding.

```bash
make help             # should show all targets including drift-check
ls docs/decisions/    # should show TEMPLATE.md
ls .claude/skills/    # should show refine/ and drift/
```

**Step 3:** Exit Claude, re-enter. Ask: "What should I do before implementing a feature?" — Claude should mention checking `docs/decisions/` for ADRs.

> ✅ **CHECKPOINT:** CLAUDE.md, Makefile, ADR template, /refine skill, /drift skill all exist. `make help` shows all targets including drift-check.

---

### ⭐ Block 1.2 — Plan Before You Build ([Recipe #02](Claude_Code_Cookbook.md))

*Goal: Architect the service before writing code.*

```
# Press Shift+Tab or type /plan

Design the project structure for go-ordersvc. This is an order
management microservice with:
- REST API: CRUD orders, list with pagination, filter by status
- Postgres: orders table (id, customer_id, items jsonb, status,
  total, created_at, updated_at)
- Redis: cache hot orders, rate limiting
- Docker: multi-stage Dockerfile for ARM64
- K8s: Helm chart for Docker Desktop cluster

Map directory structure (/cmd, /internal, /api, /deploy),
list all files, define interfaces between layers.
Include health checks at /healthz and /readyz.
Don't write code yet.
```

**Review and challenge the plan, then approve:**

```
# Exit Plan Mode (Shift+Tab), then:
Implement the project structure from the plan. Create all
directories and stub files with interfaces. No business logic -
just the skeleton. Make sure 'make build' passes.
```

> ✅ **CHECKPOINT:** Project structure exists. `make build` succeeds.

---

### 📝 Block 1.2b — First ADR: Architecture Decision

*Goal: Record the clean architecture decision before writing any business logic.*

> **WHY NOW?** The clean architecture decision was just made in Plan Mode. If you don't record it now, in three months someone will put SQL queries in a handler and nobody will remember why the layers exist.

```
Generate docs/decisions/ADR-0001-clean-architecture.md.

Business Context: We need a maintainable, testable codebase that
multiple engineers can work on without stepping on each other.
Jira: N/A (foundational decision).

Decision: Clean Architecture with handler > service > repo layers.

Constraints (each with BECAUSE and CHECK):
1. Handlers must not import repository or database packages.
   Because: handlers should only do HTTP concerns (parse request,
   call service, format response). Testable without DB.
   Check: no pgx/sql imports in internal/handler/.

2. Service layer must depend on repository interfaces, not
   concrete implementations. Because: enables mock-based testing
   and swapping implementations (e.g., Postgres to DynamoDB).
   Check: service files import interfaces, not pgx directly.

3. Repository layer must not import handler or service packages.
   Because: data access is independent of business logic.
   Check: no handler/service imports in internal/repository/.

Then add the first deterministic check to the Makefile drift-check
target: grep for database imports in internal/handler/.
```

**Verify:**

```bash
make drift-check     # should pass (no violations exist yet)
cat docs/decisions/ADR-0001-clean-architecture.md
```

> ✅ **CHECKPOINT:** ADR-0001 committed. `make drift-check` passes with at least one real constraint check.

---

### ⭐ Block 1.3 — Test-Driven Everything ([Recipe #03](Claude_Code_Cookbook.md))

*Goal: Tests exist before implementation. `make test` is the feedback loop.*

```
Write comprehensive failing tests for the order service layer.
Use table-driven tests with testify/assert. Cover:
1. CreateOrder: valid, missing customer_id, empty items
2. GetOrder: found, not found (ErrNotFound)
3. ListOrders: pagination, filter by status, empty results
4. UpdateOrderStatus: valid transitions (pending > confirmed >
   shipped > delivered), invalid transitions

Use a mock repository (interface-based, per ADR-0001 constraint #2).
Then implement the service layer to pass all tests.
Run 'make test' to verify.
```

**Observe:** Claude should reference ADR-0001 when choosing to use interface-based mocks. If it doesn't, the CLAUDE.md rules need strengthening.

> ✅ **CHECKPOINT:** `make test` passes. Coverage > 80%. Service uses repo interfaces (per ADR-0001).

---

### Block 1.4 — Model Switching ([Recipe #04](Claude_Code_Cookbook.md))

*Goal: Feel the cost/quality trade-off.*

```
/model sonnet
Add copyright headers to every .go file, then run make fmt.

/model opus
Review the service layer for potential race conditions if we
add concurrent order processing later. Think hard about this.
```

> ✅ **CHECKPOINT:** You understand when to use which model. Sonnet for mechanical, Opus for analysis.

---

**End of Step 1:**

```bash
make ci    # includes drift-check
git add -A && git commit -m 'step 1: scaffold + makefile + ADR-0001 + service + tests'
git push origin main
```

---

## Step 2: Business Intent → Code

*Recipes #05, #06, #07, #08*

### 📝⭐ Block 2.0 — Business Ticket → /refine → ADR (Full Workflow)

*Goal: Experience the complete business-to-code translation workflow.*

> **SIMULATING A PM'S TICKET:** Pretend you received this Jira ticket from your PM. It's written in pure business language — no mention of REST, handlers, or HTTP status codes.

```
/refine

ORD-100: Customers can't see their order details

User Problem:
Customers need to look up their orders - current status, what
they ordered, and when it was placed. Right now there's no way
for them to do this without calling support.

Business Impact:
40% of support tickets are 'where is my order' inquiries.
Each costs $8 to handle. That's $50K/month we can eliminate.

Desired Outcome:
Customers can see their order details instantly. They can see a
list of their recent orders and filter by status (pending, shipped,
delivered). They should never need to call support for basic
order info.

How We'll Measure Success:
'Where is my order' tickets drop by 80% within 30 days of launch.
```

**What happens:** Claude reads the codebase (sees the service layer from Step 1), proposes an implementation plan mapping business outcomes to technical changes, breaks it into subtasks with Given/When/Then acceptance criteria tied to the $50K/month impact.

**Review Claude's plan, then approve:**

```
Good analysis. Approve. Generate the ADR.
```

**Claude generates ADR-0002 with constraints like:**

- **CONSTRAINT:** GET /orders/:id must return 404 (not 500) for missing orders. **BECAUSE:** 500 errors trigger customer support calls — the exact problem we're solving. **CHECK:** test verifies 404 response for non-existent ID.
- **CONSTRAINT:** List endpoint must support status filter. **BECAUSE:** "where is my order" requires filtering by "shipped"/"delivered." **CHECK:** test verifies `?status=shipped` returns only shipped orders.

```bash
cat docs/decisions/ADR-0002-order-details-api.md
make drift-check     # verify new constraints added
```

> ✅ **CHECKPOINT:** ADR-0002 exists. Constraints trace to business outcomes. `make drift-check` passes.

---

### ⭐ Block 2.1 — REST API + Docker ([Recipe #08](Claude_Code_Cookbook.md): Acceptance Criteria)

*Goal: Implement the feature per ADR-0002 constraints.*

```
Implement the REST API for orders per ADR-0002.
Read docs/decisions/ADR-0002-order-details-api.md first.
Follow all CONSTRAINT rules.

Acceptance criteria (from the ADR subtasks):
1. POST /api/v1/orders creates order, returns 201 + Location
2. GET /api/v1/orders/:id returns order, 404 if missing
3. GET /api/v1/orders?status=pending&limit=20 paginates
4. PATCH /api/v1/orders/:id/status transitions correctly
5. GET /healthz returns 200, GET /readyz checks DB connection
6. Structured JSON logging with slog on every request

Then create a multi-stage Dockerfile:
- Build: golang:1.22-alpine, CGO_ENABLED=0, GOARCH=arm64
- Run: gcr.io/distroless/static-debian12

Verify with: make test && make docker && make drift-check
```

**Observe:** Claude should reference ADR-0002 constraints during implementation. If it puts DB access in a handler, `make drift-check` will catch the ADR-0001 violation.

> ✅ **CHECKPOINT:** `make docker` + `make scan` + `make drift-check` pass. `curl /healthz` returns 200.

---

### Block 2.2 — Visual Debugging ([Recipe #05](Claude_Code_Cookbook.md))

*Goal: Screenshot-driven fix.*

```
# Paste a screenshot of terminal error or docker logs
[paste screenshot]
This error appears when hitting /api/v1/orders without Postgres.
Add graceful error handling instead of a panic. Then make test.
```

---

### ⭐ Block 2.3 — PR Automation + docker-compose ([Recipe #06](Claude_Code_Cookbook.md))

*Goal: Claude creates branch, docker-compose, PR.*

```
Create branch feat/docker-compose. Add docker-compose.yml with:
- go-ordersvc (from Dockerfile)
- postgres:16-alpine with init SQL for orders table
- redis:7-alpine

Wire it into the Makefile: compose-up and compose-down targets.
Add integration tests that use the compose stack.
Verify with make compose-up && make test-integration.

Commit and create a PR. Include in the PR description:
- Which ADR constraints this satisfies
- Acceptance criteria status (pass/fail)
```

> ✅ **CHECKPOINT:** PR on GitHub with ADR references. `make compose-up && make test-integration` passes.

---

### 📝 Block 2.4 — Generated Docs + ADR Update ([Recipe #07](Claude_Code_Cookbook.md))

*Goal: Docs generated. ADR-0002 updated with actual PR number.*

```
Generate docs/API.md from the actual handler code. For each
endpoint: method, path, request/response schemas, status codes,
example curl commands.

Generate docs/ARCHITECTURE.md explaining clean architecture
layers, how the Makefile targets map to the workflow, and where
ADRs live.

Update ADR-0002 Traceability section with the actual PR number.
If any constraints were refined during implementation, update
them in the ADR.
```

> ✅ **CHECKPOINT:** `docs/` exists. ADR-0002 Traceability has PR number. Docs match actual code.

---

**End of Step 2:**

```bash
gh pr merge --squash && git checkout main && git pull
```

---

## Step 3: K8s + Advanced Agent Patterns

*Recipes #09, #10, #11, #12*

### ⭐ Block 3.1 — Skills Including /drift ([Recipe #10](Claude_Code_Cookbook.md))

*Goal: Reusable skills that complement Makefile tools, including /drift for ADR compliance.*

**Key insight:** You already created /refine and /drift in Step 1. Now add /review and /k8s-review, then test /drift against your two ADRs.

```
Create two additional skills:

1. .claude/skills/review/SKILL.md - Architecture review skill:
   PREREQUISITE: make lint must pass before this runs.
   Focus on what linters miss: wrong abstraction boundaries,
   business logic errors, context.Context misuse.
   Output: file | concern | severity | recommendation

2. .claude/skills/k8s-review/SKILL.md - K8s architecture review:
   PREREQUISITE: make k8s-lint must pass first.
   Focus on: resource sizing for M4 Docker Desktop, HPA config,
   graceful shutdown, PDB config.

Then test the /drift skill against your existing ADRs:
```

```
/drift

# Expected output: checks ADR-0001 and ADR-0002 constraints
# Should report PASS for all constraints (no drift yet)
```

> ✅ **CHECKPOINT:** Four skills work. /drift reports PASS on all ADR-0001 + ADR-0002 constraints.

---

### ⭐ Block 3.2 — Deploy to K8s ([Recipe #08](Claude_Code_Cookbook.md) + [#10](Claude_Code_Cookbook.md))

*Goal: Service on Docker Desktop K8s.*

```
Create Helm chart in deploy/helm/go-ordersvc/. Include:
- Deployment: 2 replicas, resource limits (128Mi/250m),
  liveness + readiness probes on /healthz and /readyz
- Service: ClusterIP port 8080
- ConfigMap: DB connection, Redis URL
- HPA: 2-5 replicas on 70% CPU
- Postgres StatefulSet + Redis Deployment as subcharts

Validate with tools THEN Claude:
```

```bash
# Step 1: Deterministic validation (tools)
make k8s-lint

# Step 2: Architecture review (Claude)
/k8s-review

# Step 3: Deploy
make k8s-deploy
make k8s-status
```

> ✅ **CHECKPOINT:** `make k8s-lint` passes. `make k8s-deploy` succeeds. Pods running.

---

### Block 3.3 — Parallel Feature Dev ([Recipe #09](Claude_Code_Cookbook.md))

*Goal: Two features, two worktrees, simultaneous.*

```bash
# Terminal 1: Redis caching
git worktree add ../ordersvc-cache -b feat/redis-cache
cd ../ordersvc-cache && claude
> Add Redis caching to GetOrder. Cache 5 min. Invalidate on
> status update. Tests with miniredis. Verify: make test

# Terminal 2: Rate limiting
git worktree add ../ordersvc-ratelimit -b feat/rate-limit
cd ../ordersvc-ratelimit && claude
> Add Redis rate limiting middleware. 100 req/min per IP.
> 429 + Retry-After header. Tests. Verify: make test
```

**Merge both, run full validation:**

```bash
cd ~/projects/go-ordersvc
git merge feat/redis-cache && git merge feat/rate-limit
make ci    # full suite on merged code (includes drift-check)
```

> ✅ **CHECKPOINT:** Both merged. `make ci` passes on combined code.

---

### 📝 Block 3.3b — Record Feature Decisions (ADRs)

*Goal: Cache and rate limit decisions are recorded and enforceable.*

> **WHY RECORD THESE NOW?** You just made two non-obvious decisions: caching at service layer (not handler), and rate limiting at middleware level (not per-handler). Six months from now, someone will want to move caching to the handler "for simplicity." The ADR explains why it's at the service layer and the drift-check prevents the move.

```
Generate two ADRs:

1. docs/decisions/ADR-0003-redis-order-caching.md
   Business context: reduces DB load, improves response times
   for the order details feature (ADR-0002).
   Key constraints:
   - Caching MUST be at service layer, not handler
   - TTL MUST be configurable via CACHE_TTL_SECONDS env var
   - Redis failure MUST fall back to DB silently
   - 404s MUST NOT be cached
   - Cache MUST invalidate on UpdateOrderStatus
   Include CHECK rules for each. Add deterministic checks
   to make drift-check (e.g., no Redis imports in handler/).

2. docs/decisions/ADR-0004-rate-limiting.md
   Business context: protect the API from abuse and ensure
   fair access for all customers.
   Key constraints:
   - Rate limit at middleware, not per-handler
   - Limits MUST be configurable via env vars
   - 429 response MUST include Retry-After header
   Include CHECK rules and add to drift-check.

Then verify everything still holds:
```

```bash
make drift-check     # now checks ADR-0001 through ADR-0004
/drift               # Claude does full judgment audit
```

> ✅ **CHECKPOINT:** ADR-0003 + ADR-0004 committed. `make drift-check` has new constraint checks. /drift reports all PASS.

---

### Block 3.4 — Sub-agents + MCP ([Recipes #11, #12](Claude_Code_Cookbook.md))

*Goal: Claude reviews architecture. Tools handle the rest.*

**Create a security ARCHITECTURE reviewer (not scanner — gosec handles that):**

```bash
cat > .claude/agents/security-reviewer.md << 'EOF'
---
name: security-reviewer
description: Reviews architecture for security design issues
tools: Read, Grep, Glob
permissionMode: plan
---
You are a Go security architect. PREREQUISITE: make sec has
passed. Find what gosec CANNOT find:
- Auth bypass through handler logic
- TOCTOU races in status transitions
- Missing authorization on state-changing endpoints
- Violations of ADR constraints related to security
Never modify files. Report only.
EOF
```

**Add Postgres MCP for schema inspection:**

```bash
make compose-up    # start Postgres
claude mcp add postgres npx @anthropic-ai/mcp-server-postgres \
  -e POSTGRES_CONNECTION_STRING='postgresql://ordersvc:pass@localhost:5432/orders'
```

```
# In Claude:
> Query the orders table schema. Compare against the queries in
> ListOrders handler. Are we missing indexes for the WHERE +
> ORDER BY patterns? Cross-reference with ADR-0002 constraints
> about pagination and status filtering.
```

> ✅ **CHECKPOINT:** Security reviewer finds logic issues. MCP reads actual DB schema. ADR constraints referenced.

---

**End of Step 3:**

```bash
make ci
git add -A && git commit -m 'step 3: k8s + skills + agents + ADR-0003 + ADR-0004'
git push origin main
```

---

## Step 4: Claude Code 2.0 Features + Drift Enforcement

*Recipes #13, #14, #15, #16*

### ⭐ Block 4.1 — Checkpoints ([Recipe #13](Claude_Code_Cookbook.md))

*Goal: Experiment boldly, rewind safely.*

```
Refactor the repository layer from raw pgx queries to sqlc.
Generate sqlc.yaml, write .sql query files, regenerate Go code.
Verify: make test && make drift-check

# If it doesn't feel right or drift-check fails, rewind:
# Press Esc Esc (or /rewind)
# Choose: 'Code only' (revert files, keep conversation)

# Try a different approach with what you learned:
Actually, use sqlc for reads only. Keep pgx for writes that
need transactions. Implement that instead.
make test && make drift-check
```

> ✅ **CHECKPOINT:** You've rewound at least once. Final implementation passes `make test` + `make drift-check`.

---

### ⭐ Block 4.2 — Hooks + Drift Enforcement ([Recipe #14](Claude_Code_Cookbook.md))

*Goal: Hooks call Makefile targets including drift-check.*

```jsonc
// Use /hooks or add to .claude/settings.json:
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Write|Edit",
        "hooks": [{
          "type": "command",
          "command": "make fmt"
        }]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "make drift-check",
            "async": false
          },
          {
            "type": "command",
            "command": "osascript -e 'display notification \"Claude finished!\" with title \"Claude Code\"'",
            "async": true
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Bash(git push*)",
        "hooks": [{
          "type": "command",
          "command": "make lint && make drift-check"
        }]
      }
    ]
  }
}
```

> **WHAT THESE HOOKS DO**
>
> - **PostToolUse (Write|Edit) → `make fmt`:** Every file edit auto-formats via the Makefile.
> - **Stop → `make drift-check`:** Every time Claude finishes work, ADR constraints are verified automatically. You see violations immediately, not at PR time.
> - **PreToolUse (git push) → `make lint && make drift-check`:** Lint AND drift gate before any push.
>
> `make drift-check` in the Stop hook is the key addition. It means Claude can never silently violate an ADR constraint — you'll see the failure before Claude even finishes reporting what it did.

**Test it:** Intentionally create a drift. Ask Claude to add a pgx import to a handler file. The Stop hook should fire `make drift-check` and fail.

```
# Intentional drift test:
Add a direct database query to the health check handler in
internal/handler/health.go - import pgx and query Postgres.

# Watch: Stop hook fires make drift-check, which fails with:
# DRIFT: ADR-0001 - DB import in handler layer

# Then fix it:
Revert that change. Health check should use the service layer.
make drift-check    # should pass again
```

> ✅ **CHECKPOINT:** Hooks fire correctly. Intentional drift caught by Stop hook. Fixed and passing.

---

### Block 4.3 — Context Management ([Recipe #15](Claude_Code_Cookbook.md))

*Goal: Understand and control your context budget.*

```
/context     # see what's using space

# If high, compact with focus on what matters
/compact focus on K8s deployment, ADR constraints, and
the caching + rate limiting implementation

# Delegate heavy analysis to sub-agent (isolated context)
Have an Explore agent audit all error handling patterns
across the codebase and check them against ADR-0002's
constraint about 404 vs 500 responses.
```

> ✅ **CHECKPOINT:** `/context` shows token usage. Compacted once. Sub-agent returned summary.

---

### Block 4.4 — Plugins ([Recipe #16](Claude_Code_Cookbook.md))

*Goal: Understand your setup is already shareable.*

```
/plugins                # browse community plugins

# Your .claude/ directory IS a plugin:
# .claude/skills/refine/SKILL.md    (business > implementation)
# .claude/skills/drift/SKILL.md     (ADR compliance)
# .claude/skills/review/SKILL.md    (architecture review)
# .claude/skills/k8s-review/SKILL.md
# .claude/agents/security-reviewer.md
# .claude/settings.json (hooks with drift-check)
#
# A teammate clones this repo and gets EVERYTHING:
# the skills, the hooks, the ADRs, the drift checks.
```

> ✅ **CHECKPOINT:** Browsed /plugins. Your `.claude/` + `docs/decisions/` is a complete, shareable system.

---

**End of Step 4:**

```bash
make ci
git add -A && git commit -m 'step 4: checkpoints + hooks + drift enforcement'
git push origin main
```

---

## Step 5: Integration + Ship

*All recipes combined*

*Everything composes. gRPC, Kafka, full CI pipeline. Every recipe exercised. Every decision documented.*

### 📝⭐ Block 5.1 — gRPC Layer (Refine + Plan + TDD + ADR)

*Combines: /refine, #02 Plan, #03 TDD, #04 Model Switch, #13 Checkpoints*

**Start with business intent, not technical spec:**

```
/refine

ORD-201: Internal systems need real-time order access

User Problem:
Our warehouse system and notification service need to query and
watch order status changes. They currently poll the REST API
every 5 seconds, causing unnecessary load and delayed updates.

Business Impact:
Warehouse picks are delayed by up to 5 seconds. Notification
emails go out late. REST polling is 60% of our API traffic.

Desired Outcome:
Internal services get instant notification of order changes.
No polling. Warehouse picks start within 1 second of order
confirmation.
```

**After Claude proposes gRPC with server streaming, approve and generate ADR:**

```
Approve. Generate ADR-0005 for the gRPC decision.
Key constraints should cover: why gRPC over webhooks,
why server-streaming for watch, proto file location,
and that gRPC handlers follow the same clean architecture
layers as REST (per ADR-0001).

# Then implement
/model sonnet
Generate the proto file and run make proto.

/model opus
Write tests for the gRPC server including streaming tests
for WatchOrderStatus. Then implement.
Verify: make test && make drift-check
```

> ✅ **CHECKPOINT:** ADR-0005 committed. `make proto` generates code. gRPC tests pass. `make drift-check` green.

---

### 📝 Block 5.2 — Kafka Events (Refine + Parallel + ADR)

*Combines: /refine, #09 Parallel, #11 Sub-agents*

```
/refine

ORD-202: Downstream systems need order event feed

User Problem: Analytics, billing, and reporting systems need
to react to order lifecycle events. Currently they query our
DB directly with read replicas, causing tight coupling.

Desired Outcome: Order events published to a message bus.
Downstream systems subscribe and react independently.

# After approval:
Generate ADR-0006 for the Kafka decision.
Constraints should cover: event schema, at-least-once delivery,
idempotency requirements, and that event publishing happens
in the service layer (per ADR-0001).
```

```bash
git worktree add ../ordersvc-kafka -b feat/kafka-events
cd ../ordersvc-kafka && claude

Implement per ADR-0006. Read the ADR constraints first.
Add Kafka to docker-compose.yml. Use segmentio/kafka-go.
Write tests with embedded Kafka.
Verify: make compose-up && make test-integration && make drift-check

# After merge:
cd ~/projects/go-ordersvc
git merge feat/kafka-events
make sec             # gosec first (tools)
@security-reviewer   # then architecture review (Claude)
/drift               # full ADR compliance audit
```

> ✅ **CHECKPOINT:** ADR-0006 committed. Events publish. `make ci` passes. /drift reports all PASS across all 6 ADRs.

---

### ⭐ Block 5.3 — GitHub Actions CI (The Payoff)

*Goal: CI calls the same Makefile targets including drift-check.*

```
Create .github/workflows/ci.yml. The pipeline should call
Makefile targets - same commands we run locally:

Jobs:
  lint:       make lint
  security:   make sec && make vuln && make secrets
  test:       make test (upload coverage to codecov)
  build:      make docker && make docker-push
  k8s:        make k8s-lint
  scan:       make scan (trivy on the built image)
  drift:      make drift-check (ADR constraint enforcement)

Trigger on: push to main, PR to main.
Use Go 1.22, setup-go action, docker buildx for ARM64.

IMPORTANT: Every CI step must use a 'make' target.
The drift job ensures no PR can merge if it violates an ADR.
```

> **THE FULL LOOP**
>
> PM writes Jira ticket (business language)
> → Engineer runs /refine (Claude translates, proposes options)
> → Engineer approves, Claude generates ADR (decisions recorded)
> → Claude implements (reads ADR constraints)
> → Stop hook: `make drift-check` (immediate feedback)
> → `make ci` locally (full validation including drift)
> → GitHub Actions: same make targets (blocks PR on drift)
> → `/drift` quarterly (full judgment audit)
>
> Decisions are traceable. Drift is caught automatically. Business intent is preserved.

**Final push:**

```
# Regenerate docs to reflect the complete system
Regenerate docs/API.md (REST + gRPC endpoints) and
docs/ARCHITECTURE.md (include ADR workflow, drift detection,
and the full traceability chain).

# Full validation
make ci
/drift              # final judgment audit across all 6 ADRs
/review

# Create the final PR with changelog
git add -A && git checkout -b release/v1.0.0
Create a PR with changelog grouped by:
- API (REST + gRPC)
- Infrastructure (Docker + K8s + Kafka)
- Testing
- Documentation (ADRs, API docs, architecture docs)
- Drift Detection (Makefile checks, /drift skill)

gh pr merge --squash
git checkout main && git pull
make k8s-deploy
make k8s-status
```

> ✅ **CHECKPOINT:** CI green including drift. 6 ADRs committed. Docs current. Pods running on K8s.

---

## The Final Architecture

*How tools, Claude, the Makefile, and ADRs work together.*

```
┌─────────────────────────────────────────────────────────────┐
│                       MAKEFILE                              │
│              (Single Source of Truth)                        │
└───┬─────────┬───────────┬────────────┬────────────┬─────────┘
    │         │           │            │            │
 ┌──┴───┐  ┌──┴─────┐ ┌──┴──────┐ ┌──┴──────┐ ┌──┴──────┐
 │ You  │  │ Claude │ │ Hooks   │ │ GitHub  │ │ /drift  │
 └──────┘  └────────┘ └─────────┘ └─────────┘ └─────────┘
 make ci    make test  make fmt    make lint   reads ADRs
 /refine    make lint  make lint   make sec    checks code
            /review    drift-check make test   reports drift
            /drift     (on edit)   drift-check (judgment)
                       (on push)   (on every PR)

┌───────────────────────┬──────────────────────┬──────────┐
│  TOOLS (fast, free)   │ CLAUDE (judgment)     │ ADRs     │
├───────────────────────┼──────────────────────┼──────────┤
│ golangci-lint         │ Architecture review  │ Why this │
│ gosec, trivy          │ Business logic bugs  │ way and  │
│ gitleaks, govulncheck │ /refine translation  │ not that │
│ kubeconform, gofmt    │ /drift judgment      │ way      │
│ make drift-check      │ Design alternatives  │          │
└───────────────────────┴──────────────────────┴──────────┘
```

---

## Recipe Coverage Map

| # | Recipe | Step | Block | Documentation Step |
|---|--------|------|-------|--------------------|
| 01 | CLAUDE.md | 1 | 1.1 | ADR template + /refine + /drift skills created |
| 02 | Plan Mode | 1 | 1.2 | 📝 ADR-0001 generated from architecture plan |
| 03 | TDD | 1 | 1.3 | Tests reference ADR-0001 interface constraint |
| 04 | Model Switch | 1 | 1.4 | Sonnet for mechanical, Opus for analysis |
| 05 | Visual Debug | 2 | 2.2 | Screenshot → fix → make test |
| 06 | PR Automation | 2 | 2.3 | PR description references ADRs |
| 07 | Doc Generation | 2 | 2.4 | 📝 ADR-0002 updated with PR number |
| 08 | Acceptance Criteria | 2 | 2.0–2.1 | 📝 /refine generates ADR-0002 from business ticket |
| 09 | Parallel Dev | 3 | 3.3 | 📝 ADR-0003 + ADR-0004 after feature merge |
| 10 | Skills | 3 | 3.1 | /drift tested against live ADRs |
| 11 | Sub-agents | 3 | 3.4 | Security reviewer checks ADR constraints |
| 12 | MCP | 3 | 3.4 | DB schema cross-referenced with ADR-0002 |
| 13 | Checkpoints | 4 | 4.1 | drift-check validates after rewind |
| 14 | Hooks | 4 | 4.2 | 📝 drift-check in Stop hook catches violations |
| 15 | Context Mgmt | 4 | 4.3 | Sub-agent audits error handling vs ADR |
| 16 | Plugins | 4 | 4.4 | .claude/ + docs/decisions/ = shareable system |

---

## ADR Inventory

| ADR | Title | Created | Key Constraints | Drift Checks |
|-----|-------|---------|-----------------|--------------|
| 0001 | Clean Architecture | Step 1 | No DB in handlers, interface deps, layer isolation | grep: handler imports |
| 0002 | Order Details API | Step 2 | 404 not 500, status filter, pagination | grep: error responses |
| 0003 | Redis Caching | Step 3 | Service-layer only, configurable TTL, fallback to DB | grep: Redis in handler |
| 0004 | Rate Limiting | Step 3 | Middleware-level, configurable limits, Retry-After | grep: rate limit config |
| 0005 | gRPC API | Step 5 | Same layers as REST, proto file location, streaming | grep: gRPC handler imports |
| 0006 | Kafka Events | Step 5 | Service-layer publishing, at-least-once, idempotent | grep: Kafka in handler |

---

## Troubleshooting

| Problem | Fix |
|---------|-----|
| Claude runs `go test` instead of `make test` | Check CLAUDE.md has 'Always use make targets' rule. Re-enter session. |
| Claude ignores ADR constraints | Check CLAUDE.md has ADR section. Run `/drift` to verify compliance. |
| `make drift-check` false positive | Review the grep pattern. Adjust for test files: `grep -v _test.go`. |
| ADR gets stale after refactor | Run `/drift`. Update ADR status to Amended. Preserve original reasoning. |
| `make lint` fails: golangci-lint not found | `brew install golangci-lint`. Verify: `which golangci-lint` |
| Docker build fails on ARM64 | Ensure `--platform linux/arm64`. Check base images support ARM. |
| K8s pods OOMKilled | Docker Desktop > Settings > Resources > increase to 8GB+ RAM |
| Claude forgets instructions mid-session | `/context` to check. `/compact` to free space. Put rules in CLAUDE.md. |
| Hook blocks Claude from pushing | Intentional — `make lint` or `drift-check` failed. Fix issues first. |
| Too many ADRs to track | Run `/drift` for full audit. Deprecate ADRs for removed features. |

---

## What's Next

After completing all steps, consider adding:

- **Jira MCP server:** `claude mcp add jira` — /refine reads tickets directly instead of copy-paste.
- **ADR-aware PR template:** GitHub PR template that requires listing relevant ADRs and constraint status.
- **Quarterly /drift audit:** Schedule /drift runs before each release. Deprecate stale ADRs.
- **Monitoring:** Prometheus metrics + Grafana. Create ADR for observability strategy.
- **ArgoCD:** GitOps deployment. Helm chart in repo → ArgoCD watches → auto-deploy on merge.
- **Agent Teams:** Experimental Opus 4.6 feature. `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1`.
- **ADR dashboard:** Script that parses `docs/decisions/` and reports constraint count, drift-check coverage, and ADR age.

---

## Directory Structure

```
go-ordersvc/
├── CLAUDE.md                    # Points Claude to ADRs + Makefile
├── Makefile                     # build, test, lint, drift-check, ci
├── docs/
│   ├── decisions/
│   │   ├── TEMPLATE.md          # ADR template (CONSTRAINT/BECAUSE/CHECK)
│   │   ├── ADR-0001-clean-architecture.md
│   │   ├── ADR-0002-order-details-api.md
│   │   ├── ADR-0003-redis-order-caching.md
│   │   ├── ADR-0004-rate-limiting.md
│   │   ├── ADR-0005-grpc-api.md
│   │   └── ADR-0006-kafka-events.md
│   ├── API.md                   # Generated from code
│   └── ARCHITECTURE.md          # Generated from code
├── .claude/
│   ├── skills/
│   │   ├── refine/SKILL.md      # Business ticket → plan + ADR
│   │   ├── drift/SKILL.md       # ADR compliance audit
│   │   ├── review/SKILL.md      # Architecture review
│   │   └── k8s-review/SKILL.md
│   ├── agents/
│   │   └── security-reviewer.md
│   └── settings.json            # Hooks (fmt, lint, drift-check)
├── .github/
│   └── workflows/
│       └── ci.yml               # make ci + make drift-check
└── internal/                    # ... application code ...
```
