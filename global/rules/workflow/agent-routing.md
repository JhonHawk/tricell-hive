---
alwaysApply: true
---

## Agent Routing

### Explicit Invocation Takes Precedence
If the user names a specific agent (`@agent-name`, "use the X agent"), invoke it directly. The rules below only apply when the main thread decides implicitly which agent to use.

### Single-Agent Disambiguation
When a task could map to multiple agents, use these signals. If a task hits multiple rows, don't pick one — chain them in this order: design → implementation → tests → review (see Multi-Agent Detection and Reference Chains below).

| Signal in task | Route to | NOT to |
|---|---|---|
| slow query, EXPLAIN, N+1, index tuning | performance-engineer | backend-developer |
| schema design, migration, ORM config | database-specialist | backend-developer |
| write tests as the primary objective (coverage push, new E2E suite) | test-engineer | the implementing agent |
| vulnerability, OWASP, secrets, auth bypass | security-reviewer | code-reviewer |
| prompt design, LLM integration, structured output | prompt-engineer | backend-developer |
| CI/CD, Docker, Terraform, deploy pipeline | devops-engineer | backend-developer, cloud-architect |
| cloud topology, landing zone, DR (RTO/RPO), migration planning (6Rs), FinOps cost architecture | cloud-architect | devops-engineer, system-designer |
| secret scanning, leaked credentials | secrets-auditor | security-reviewer |
| new API contract, cross-service schema design, service boundaries | system-designer | code-reviewer |
| Angular components, services, routing, angular.json present | angular-developer | nextjs-architecture-expert |
| Next.js, App Router, Server Components, next.config present | nextjs-architecture-expert | angular-developer |
| API endpoints, backend logic, microservices (no DB/perf focus) | backend-developer | database-specialist, performance-engineer |
| Kotlin Multiplatform (KMP), Android, Compose, shared mobile code | kotlin-multiplatform-developer | backend-developer |
| server-only Kotlin (Ktor/Spring, no Android or multiplatform target) | backend-developer | kotlin-multiplatform-developer |
| README, ADR, API docs, setup guide | technical-writer | the implementing agent |
| raw client requirements doc, project intake analysis (no quotation involved) | requirement-analyst | spec-quality-reviewer |
| spec/épica completeness, Gherkin verifiability, quality gate on written specs | spec-quality-reviewer | product-critic, code-reviewer |
| challenge necessity/scope/shape of a feature BEFORE implementation | product-critic | spec-quality-reviewer, architect-style review |
| UX friction in a live mock or deployed flow, navigation review | ux-flow-reviewer | code-reviewer, nextjs-architecture-expert |
| functional verification of a running app: walk ACs + adversarial/negative testing (double-click, invalid input, gated routes, mid-flow refresh) in a real browser | in-vivo-qa-tester | ux-flow-reviewer, test-engineer |
| workspace file hygiene, misplaced artifacts, ledger repair, unpromoted decisions | workspace-custodian | secrets-auditor |

### Skill Disambiguation (output rendering)
When choosing a skill for *how to render output* (distinct from agent selection above):

| Signal in output need | Route to | NOT to |
|---|---|---|
| static rich HTML report (plan, spec, audit, brief, code review writeup, research, deck, design tokens reference) | `flow-report` skill | `playground` |
| interactive HTML with live state (sliders, knobs, prototype, live editor, custom config form) | `playground` skill | `flow-report` |
| UI component, landing page, visual artifact for production | `frontend-design` or `canvas-design` skill | `flow-report`, `playground` |
| short conversational, single-section, agent-consumed, versioned doc | Markdown or inline prose | any HTML skill |

See `quality/communication-format.md` for the full trigger logic and carve-outs.

### Browser Tooling (which tool drives the browser)
Distinct from agent selection above: once an agent needs a browser, the default driver is the
`agent-browser` CLI (via Bash) — chrome-devtools MCP only for Lighthouse/perf-insight/heap,
playwright MCP only as fallback. `tools/browser-automation.md` owns the full rationale.

### Delegation Thresholds
Delegate on growing complexity, not only on explicit request. These are orientative gates against a monolithic main thread, not hard stops — apply with proportion:
- **Understanding a flow that spans 4+ files** → delegate a bounded exploration to `Explore` rather than reading them all in the main thread (extends the "3+ search queries" trigger in global `CLAUDE.md`).
- **Writing 2+ non-trivial files** → delegate one writer (the domain specialist per the table above), then verify in fresh context.
- **~20 tool calls, 5 exploratory reads, or 2 non-mechanical edits without delegating** → pause and re-plan instead of pushing the session further.

### Multi-Agent Detection
Before acting on a broad task, evaluate whether it spans 2+ domains:
- Mentions both design AND implementation ("add payment processing")
- Requires both frontend AND backend changes
- Explicitly or implicitly needs testing after implementation
- Crosses service boundaries (see cross-service-workflow rule)

When multiple agents are needed:
1. Identify which agents are needed and in what order
2. Present the routing plan to the user before executing
3. Execute sequentially — pass each agent's key outputs (spec paths, API schemas) as context to the next

### Reference Chains
These are reference sequences, not mandatory pipelines. Skip steps in proportion to the change: a trivial feature collapses to the implementation agent alone; cross-service, security-sensitive, or high-risk changes use the full chain. Note that implementation agents write tests for their own code per `quality/testing.md` — test-engineer is reserved for primary-task coverage work.

- **New feature**: system-designer → implementation agent(s) → (test-engineer if coverage push needed) → code-reviewer
- **Bug fix (root cause unknown)**: performance-engineer or Explore → implementing agent
- **Security audit**: security-reviewer → secrets-auditor → code-reviewer
- **New deployment**: implementation agent → devops-engineer → security-reviewer
- **Cloud migration / greenfield infra**: cloud-architect (topology/DR/cost spec) → devops-engineer (IaC + pipelines) → security-reviewer

**Verification runs in fresh context.** Route review/verification stages to a separate subagent that did not implement the change — fresh-context verifiers outperform self-critique by the implementing thread. The verifier's input is the actual change (the diff, the run output), never the implementing agent's report of it — a report travels as claims to check, not as context to trust.
- **The verifier runs the tests, it does not read about them.** For any behavior change, it re-runs the suite to confirm the **verifiable test gate** (`quality/testing.md`): fail-to-pass (the new tests fail without the change) and pass-to-pass (the existing suite still passes), and it inspects the diff for test-gaming — `.skip`/`xit`, deleted assertions, `--no-verify`, runner-config edits. A green report over weakened tests fails the gate. This is why the test→code contract cannot cross the agent boundary on trust: the subagent can reward-hack it, so the verifier re-establishes it from the run.
- **The implementer's handoff carries the evidence, not the adjective.** It returns which tests it added (paths) and the verify command with its actual output — "implemented with tests" is a claim; `pnpm test messages.spec → 12 passing` is evidence the verifier can re-run.

### Chain Interruption
When an agent in a chain produces output that the next agent cannot consume (incomplete spec, failing tests, ambiguous review):
1. Do not silently skip or retry the agent
2. Present the gap to the user with the specific issue
3. Let the user decide: re-run with clarification, skip with acknowledgment, or adjust the plan

### Coexistence
This rule guides implicit routing only. It does not interfere with:
- Direct agent invocation by the user (`@agent-name`)
- Agent Teams coordination (teammates, SendMessage, shared tasks)
- `flow-*` skills (explicit phase gates — each skill's body owns its agent dispatch)
- `flow-report` skill (auto-invoked per `quality/communication-format.md` — handles output rendering, not routing)

### Workflow Tool vs Subagents vs Agent Teams
- **Default to a direct subagent.** Explore for search, the domain specialist (per Single-Agent Disambiguation above) for domain work, chained design → impl → tests → review. A single `Agent` call with the right specialist is the correct default for single-domain work — including under `ultracode`. Do not wrap one task in a workflow, and do not delegate trivial edits at all.
- **Reach for the Workflow tool only when** the agent set or execution order is data-driven — computed per run (e.g. "review every changed file", "audit every route", fan-out over an unknown-size list) — not when a fixed chain of named agents already suffices.
- **Reach for Agent Teams only when** you need persistent parallel workers with shared task lists and P2P messaging, or the user asks.
- **When you do author a workflow, route each stage to its specialist via `agentType`,** reusing the same Single-Agent Disambiguation table that governs direct routing — e.g. a Next.js implementation stage gets `agentType: "nextjs-architecture-expert"`, a security pass `"security-reviewer"`. The default subagent (no `agentType`) is for glue steps only — fan-out, dedup, synthesis — and loses any custom agent's tool/permission allowlist. An unknown `agentType` throws; there is no fallback.
- **Workflows require explicit opt-in.** They run only under the `ultracode` keyword, a standing `/effort ultracode` session, an explicit user request, or `bypassPermissions`; otherwise each run hits the permission gate. Never call the Workflow tool on inferred intent alone.
- *Version-sensitive (workflow/ultracode are research-preview, Claude Code 2.1.154+) — revisit the workflow-specific bullets if the API shifts.*
