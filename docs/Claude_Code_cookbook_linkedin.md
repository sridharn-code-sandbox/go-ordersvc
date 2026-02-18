

[H1] The Claude Code Cookbook: 16 Production Recipes for Claude Code 2.0+

Problem → Approach → Prompt → Outcome

---

[H2] TL;DR

Claude Code is an AI agent that lives in your terminal. Treat it like an orchestration framework, not a chatbot. CLAUDE.md gives it persistent project memory. Plan Mode prevents rework. TDD makes it self-correcting. Skills encode your team's standards. Hooks enforce them automatically. Checkpoints let you experiment fearlessly. Sub-agents handle specialist work. MCP connects it to your entire stack. The 16 recipes below are the patterns that compound.

Start here: Recipe #01 (CLAUDE.md) on day one. Recipe #02 (Plan Mode) before any big task. Recipe #03 (TDD) for every feature. Everything else builds on these three.

The mental model:
- CLAUDE.md = persistent state
- Plan Mode = prevents rework
- TDD = self-correcting loop
- Skills = process-as-code
- Hooks = deterministic enforcement
- Checkpoints = undo for everything
- Sub-agents = specialist delegation
- MCP = system integration

The mistake everyone makes: Reviewing every line Claude writes. Switch to acceptance-criteria verification (Recipe #08). 5 minutes checking outcomes beats 30 minutes reading code.

Cost control: Use /model to switch. Opus for architecture and debugging. Sonnet for mechanical edits. Token spend drops 40-60%.

---

[H2] Quick Reference

#01 CLAUDE.md — Persistent project memory. 50 lines that eliminate re-explaining.
#02 Plan Mode — Read-only analysis before touching code. Prevents rework.
#03 TDD — Tests first, implementation second. Self-correcting loop.
#04 Model Switch — Opus for thinking, Sonnet for doing. 40-60% cost reduction.
#05 Visual Debug — Paste screenshot, describe fix. 60-second turnaround.
#06 PR Automation — Branch, commit, changelog, PR. One prompt.
#07 Doc Generation — Docs from code, not memory. Always current.
#08 Acceptance Criteria — Verify outcomes, not lines. 5 min vs 30 min review.
#09 Parallel Dev — Git worktrees + multiple Claudes. Linear throughput gains.
#10 Skills — Reusable prompt patterns as /commands. One-word triggers.
#11 Sub-agents — Specialist AI with restricted access. Auto-delegation.
#12 MCP — Connect Claude to DB, browser, APIs. Full-stack agent.
#13 Checkpoints — Esc Esc to rewind any change. Fearless experimentation.
#14 Hooks — Auto-run lint/format/test on edits. Deterministic enforcement.
#15 Context Mgmt — /context + /compact. Sessions stay coherent for hours.
#16 Plugins — Bundle and share your entire setup. Zero onboarding friction.

---

[H2] Part 1: Foundations

The primitives everything else depends on.


[H3] #01 — Persistent Project Memory

Problem: Every session starts from zero. Claude doesn't know your stack, conventions, or deployment pipeline.

Approach: Create CLAUDE.md in project root. Claude loads it on startup. Include: architecture, conventions, stack, testing strategy. Nest CLAUDE.md in subdirectories for service-specific context. Add a "Compact Instructions" section to control what survives auto-compaction.

Prompt:
"Explore this codebase. Generate a CLAUDE.md covering: monorepo layout, service boundaries, tech stack, deployment pipeline, testing strategy, and naming conventions."

Outcome: Claude starts every session pre-loaded with project context. No re-explaining. Keep under 2K tokens. Update as the project evolves.


[H3] #02 — Plan Before You Build

Problem: Claude starts coding immediately. For small fixes, fine. For migrations or architecture changes, that's rework waiting to happen.

Approach: Plan Mode (Shift+Tab or /plan). Claude switches to read-only: maps dependencies, identifies blast radius, produces a reviewable plan. Approve before it writes code.

Prompt:
"Switch to Plan Mode. Analyze the payments service. Design a migration plan from REST to gRPC. Map downstream consumers, flag breaking changes, propose phased rollout with rollback."

Outcome: A reviewable plan with dependency graph before a single line changes. The feature that changed how I work — forced thinking before doing.


[H3] #03 — Test-Driven Everything

Problem: Claude generates code that looks correct but breaks on edge cases. You find out in staging or prod.

Approach: TDD: have Claude write comprehensive failing tests first (edge cases, concurrency, failure modes), then implement to pass them. Claude iterates until green.

Prompt:
"Write failing tests for the rate limiter: sliding window accuracy, concurrent requests, Redis failure fallback, header propagation, cluster sync. Then implement to pass all."

Outcome: Claude self-corrects during implementation. Tests become your living contract. TDD makes Claude a self-correcting system.


[H3] #04 — Model Switching as Resource Scheduling

Problem: Burning Opus-level compute on renaming variables. Token budget is finite and you're allocating it uniformly.

Approach: Use /model to switch. Opus 4.6 for architecture and complex debugging. Sonnet 4.5 for mechanical edits. Haiku 4.5 for lookups. Fast mode (Alt+Shift+F) for speed.

Prompt:
"/model sonnet — Rename all UserDTO to UserResponse across the codebase.
/model opus — Analyze the auth module for race conditions. Think hard."

Outcome: 40-60% token savings on routine tasks. Deep reasoning reserved for problems that need it.

---

[H2] Part 2: Workflows

Tighter feedback loops. Less manual toil.


[H3] #05 — Visual Debugging

Problem: You spend 5 minutes describing a UI bug that a screenshot shows instantly.

Approach: Pass screenshots directly to Claude Code. Point at the problem, describe expected behavior in one line.

Prompt:
"[paste screenshot] Mobile nav breaks below 375px — hamburger overlaps search. Fix with flexbox. Don't use absolute positioning."

Outcome: Fix in under 60 seconds instead of multi-message back-and-forth.


[H3] #06 — Automate the PR Lifecycle

Problem: Branching, PR descriptions, changelogs — same ceremony every time, scales linearly with feature velocity.

Approach: Claude integrates with GitHub via gh CLI. Branches, commits, generates context-aware PRs, builds changelogs. Use --from-pr to resume sessions linked to a PR.

Prompt:
"Review all commits since v3.1.0. Generate a changelog grouped by: Breaking Changes, New Endpoints, Bug Fixes, Infra Updates. Create a PR targeting release/3.2.0 with changelog as description."

Outcome: PR with structured changelog, ready for review. 20 minutes of ceremony in one prompt.


[H3] #07 — Docs That Don't Drift

Problem: API docs are three sprints behind. New engineers onboard by reading Slack history.

Approach: Generate docs from code, not memory. Claude reads your actual types, routes, and schemas. Re-run after every sprint. Documentation becomes a computed value.

Prompt:
"Generate API reference for /api/v2/ routes. For each endpoint: method, path, auth, request/response schemas (from TS types), rate limits, error codes. Output as markdown."

Outcome: Always-current docs generated in minutes. Re-run, don't maintain.


[H3] #08 — Verify Outcomes, Not Lines

Problem: You review every line Claude writes. 30 minutes per feature. You've become a human linter.

Approach: Define acceptance criteria upfront. Let Claude implement. Verify by running the app and checking tests. Use checkpoints (Esc Esc) as your safety net.

Prompt:
"Implement webhook delivery. Acceptance criteria: 1) Exponential backoff, max 5 retries. 2) Dead-letter queue after exhaustion. 3) HMAC-SHA256 signature verification. 4) Delivery logs queryable by event ID. Write tests for all criteria."

Outcome: 5 minutes checking outcomes vs. 30 minutes reading code. If something's wrong, Esc Esc to rewind.

---

[H2] Part 3: Advanced

Where Claude Code becomes an agent system.


[H3] #09 — Parallel Feature Development

Problem: Sequential feature development. Three independent features take 3x the calendar time.

Approach: Git worktrees + multiple Claude instances. Each worktree is isolated on its own branch. Launch background agents with & prefix. Merge when done.

How it works:
- Terminal 1: git worktree add ../proj-auth -b feat/auth && cd ../proj-auth && claude
- Terminal 2: git worktree add ../proj-search -b feat/search && cd ../proj-search && claude

Outcome: Concurrent development with zero cross-contamination. Near-linear throughput gains.


[H3] #10 — Skills (Evolved from Slash Commands)

Problem: You type the same 200-word prompt dozens of times a week.

Approach: Skills in .claude/skills/. Markdown file with optional YAML frontmatter becomes a /command. Claude can also auto-invoke skills by description. Skills support context:fork for isolated sub-agent execution.

Example (.claude/skills/review/SKILL.md):
"Review staged changes. Check: 1) Error handling on I/O, 2) Missing DB indexes, 3) N+1 patterns, 4) Secrets in code, 5) Missing retry/timeout on HTTP calls. Output: file | issue | severity | fix."

Outcome: One-word trigger: /review. Claude can also auto-detect when a skill is relevant.


[H3] #11 — Specialist Agent Teams

Problem: One generalist Claude handles everything. Context gets polluted with irrelevant history.

Approach: Sub-agents in .claude/agents/. Each has a focused prompt, restricted tools, optional memory scope. Experimental Agent Teams enable parallel agents on shared task lists.

Example (.claude/agents/security-auditor.md):
"Audit for: injection, auth bypass, secrets exposure, insecure deserialization, SSRF. Never modify files. Output: vulnerability | severity | file:line | remediation."

Outcome: Auto-delegation to specialists. Sub-agents run in isolated context. Security auditor reads but never writes.


[H3] #12 — Connect Claude to Your Stack (MCP)

Problem: Claude can read/write files but can't query your database, drive a browser, or hit internal APIs.

Approach: MCP (Model Context Protocol) servers. Connect Playwright (browser), Postgres (data), Figma (design), or custom servers. OAuth support for servers like Slack.

Setup:
- claude mcp add playwright npx @playwright/mcp@latest
- claude mcp add postgres npx @anthropic-ai/mcp-server-postgres

Outcome: Claude goes from filesystem-only to full-stack agent. One prompt spans browser, database, and internal APIs.

---

[H2] Part 4: New in 2.0+

Checkpoints, hooks, context management, plugins.


[H3] #13 — Checkpoints: Fearless Experimentation

Problem: You're afraid to let Claude attempt big refactors. If it goes wrong, undoing takes longer than doing it manually.

Approach: Checkpoints auto-save before every AI edit. Esc Esc or /rewind to roll back. Choose: revert code only, conversation only, or both. Retained 30 days.

Prompt:
"Refactor the entire auth module to use OAuth2 with PKCE. If the approach doesn't work after the first pass, I'll Esc Esc to rewind and try a different strategy."

Outcome: Experiment boldly. Large refactors with a safety net. Ship 3-5x more experimental work because the downside is near-zero.


[H3] #14 — Hooks: Deterministic Automation

Problem: You forget to lint. Tests don't run until PR time. Formatting is inconsistent. These are deterministic tasks.

Approach: Hooks run at lifecycle points: PreToolUse, PostToolUse, PermissionRequest, Stop, SubagentStart. Configure via /hooks or .claude/settings.json.

What you can automate:
- PostToolUse (Write|Edit) → auto-format with make fmt
- Stop → desktop notification when Claude finishes
- PreToolUse (git push) → lint gate before any push

Outcome: Auto-format on every edit. Lint gate before push. Desktop notification when Claude finishes. Best practices become enforced rules.


[H3] #15 — Context Management

Problem: After 45 minutes Claude "forgets" early instructions. Responses degrade. You restart and lose progress.

Approach: /context shows usage. /compact summarizes to free room (with optional focus). Auto-compaction at ~83%. Sub-agents get isolated context. Opus 4.6 beta supports 1M tokens.

Commands:
- /context — check what's using space
- /compact focus on the API migration changes — targeted compaction

Outcome: Sessions stay coherent for hours. Put persistent rules in CLAUDE.md (survives compaction). Run /context regularly.


[H3] #16 — Plugins: Share Your Setup

Problem: Great skills/hooks/agents, but sharing means copy-pasting files and hoping setups match.

Approach: Plugins bundle skills, hooks, agents into distributable packages via GitHub repos. Marketplaces are plugin collections. /plugins to discover and install.

Commands:
- /plugins — browse available
- /plugins add [repo URL] — install from GitHub

Outcome: New engineer runs /plugins add and inherits your entire setup. Zero onboarding friction.

---

[H2] Cheat Sheet

CLAUDE.md → Use on day 1, every project → Persistent state
Plan Mode → Before big refactors → Prevents rework
TDD prompts → New features → Self-correcting loop
/model switch → Every task → Resource scheduling
Screenshots → UI bugs → Fast feedback
gh integration → PR lifecycle → Toil elimination
Doc generation → After milestones → Computed docs
Acceptance criteria → Feature requests → Right abstraction
Git worktrees → Parallel features → Horizontal scaling
Skills → Repeated workflows → Process-as-code
Sub-agents → Specialist tasks → Agent microservices
MCP servers → External tools → System boundary
Checkpoints → Risky changes → Fearless experiments
Hooks → Enforce standards → Deterministic rules
Context mgmt → Long sessions → Resource management
Plugins → Team distribution → Portable setup

---

[H2] What I Learned

Highest ROI: CLAUDE.md. Not sub-agents, not MCP, not worktrees. A 50-line markdown file. Boring. Highest-leverage thing on day one.

2.0 game-changer: Checkpoints changed my risk tolerance. Before: small, safe tasks. Now: full module rewrites knowing I can Esc Esc. Ship 3-5x more experimental work.

What failed: Complex multi-agent orchestration on day one. 6-agent pipeline with custom MCP servers = debugging nightmare. Start simple. Add layers after extracting value from the current one.

The system: CLAUDE.md feeds Plan Mode. Tests feed verification. Skills feed consistency. Hooks feed enforcement. Checkpoints feed experimentation. Sub-agents feed throughput. The value is in how they compose.

---

[H2] Resources

Official Docs: https://code.claude.com/docs
Source Code: https://github.com/anthropics/claude-code
Changelog: https://github.com/anthropics/claude-code/blob/main/CHANGELOG.md
Awesome List: https://github.com/hesreallyhim/awesome-claude-code
Prompting Guide: https://docs.claude.com/en/docs/build-with-claude/prompt-engineering/overview

---

#ClaudeCode #AI #SoftwareEngineering #DevTools #Productivity #AIAgents #CodingWithAI #DeveloperExperience
