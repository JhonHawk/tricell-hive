---
alwaysApply: true
---

## Agent Routing

### Explicit Invocation Takes Precedence
If the user names a specific agent (`@agent-name`, "use the X agent"), invoke it directly. The rules below govern implicit routing only — they don't interfere with Agent Teams coordination or `flow-*` skills, whose bodies own their agent dispatch.

### Single-Agent Disambiguation
When a task could map to multiple agents, use these signals. If a task hits multiple rows, don't pick one — chain them in order: design → implementation → tests → review (see Multi-Agent Chains below).

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
| refute or adversarially verify a finding, claim, or diagnosis produced by another agent or investigation | finding-refuter | code-reviewer, the main thread |

### Skill Disambiguation (output rendering)
Static rich HTML report → `flow-report`; live interactive state (sliders, live re-render) → `playground`; production UI artifact → `frontend-design`/`canvas-design`; short, conversational, agent-consumed, or versioned doc → Markdown. Full trigger logic and carve-outs: `quality/communication-format.md` (canonical — not restated here).

### Browser Tooling
Once an agent needs a browser, the default driver is the `agent-browser` CLI (via Bash); chrome-devtools MCP only for Lighthouse/perf-insight/heap; playwright MCP as fallback. Rationale: `tools/browser-automation.md`.

### Delegation Gates
Delegate on growing complexity, not only on explicit request. These are hard gates, not suggestions: when one fires, delegate — or state in one visible line why inline is correct. Continuing silently past a fired gate is the violation. Governing question for anything unlisted: does this inflate main-thread context without need? Yes → delegate.
- **Understanding a flow that spans 4+ files** → delegate a bounded exploration to `Explore` (extends the "3+ search queries" trigger in global `CLAUDE.md`).
- **Writing 2+ non-trivial files** → delegate one writer (the domain specialist per the table above), then verify in fresh context.
- **~20 tool calls, 5 exploratory reads, or 2 non-mechanical edits without delegating** → re-plan in flight: delegate the remainder instead of pushing the session further.

Inline vs delegate — quick reference:

| Action | Inline | Delegate |
|---|---|---|
| Read 1-3 files to make a decision | ✓ | |
| Read 4+ files to understand a flow | | `Explore` |
| Mechanical single-file edit | ✓ | |
| Multi-file write or new logic | | one writer (specialist per the table) |
| State bash (`git status`, `gh pr view`) | ✓ | |
| Execution bash (tests, builds, installs) | | the writer/verifier runs them |
| Reading as preparation for an edit | | never split: the writer reads AND edits |

Delegation hygiene:
- **Executor subagents never orchestrate:** `Agent` stays out of their allowlist — Bash, MCP, and CLI access are unaffected.
- **Deduplicate launches:** before spawning, check no equivalent subagent is already running or launched — wait for its result instead of relaunching.
- **Parallel read-only agents are the default; parallel writers are not** — two writers only in isolated, explicitly approved worktrees.

### Delegation Anti-patterns
Each one inflates main-thread context without need:
- **Monolithic session:** grinding past a fired gate "to finish this part first".
- **Read-as-prep, edit separately:** paying for the same context twice — the writer reads and edits in one delegation.
- **Duplicate launch:** relaunching a subagent whose equivalent is already running or answered.
- **Reopening verification for fresh budget:** a failed fix cycle escalates (Chain Interruption below); it never restarts verification to reset the count.

### Multi-Agent Chains — declare, don't gate
A task spanning 2+ domains (design AND implementation, frontend AND backend, implies tests, crosses service boundaries) gets a chain of specialists, each receiving the previous agent's key outputs (spec paths, schemas, diffs). **Declare the chain in the start summary and execute** — fold it into the plan gate when one exists. A blocking presentation is signal-driven only: 2+ genuinely valid chains (ask which), an embedded stakeholder decision, or an irreversible/costly stage.

Reference chains — sequences, not mandatory pipelines; skip steps in proportion to the change (a trivial feature collapses to the implementation agent alone). Implementation agents write tests for their own code per `quality/testing.md`; test-engineer is for primary-task coverage work:
- **New feature**: system-designer → implementation agent(s) → (test-engineer if coverage push) → code-reviewer
- **Bug fix (root cause unknown)**: performance-engineer or Explore → implementing agent
- **Security audit**: security-reviewer → secrets-auditor → code-reviewer
- **New deployment**: implementation agent → devops-engineer → security-reviewer
- **Cloud migration / greenfield infra**: cloud-architect (topology/DR/cost spec) → devops-engineer (IaC + pipelines) → security-reviewer

### Verification runs in fresh context
Route review/verification to a separate subagent that did NOT implement the change — fresh-context verifiers outperform self-critique. The verifier's input is the actual change (the diff, the run output), never the implementer's report of it — a report travels as claims to check, not context to trust.
- **The verifier runs the tests, it does not read about them:** it re-establishes the verifiable test gate (`quality/testing.md`) from an actual run — fail-to-pass, pass-to-pass — and inspects the diff for test-gaming (the list lives in `testing.md`; a green report over weakened tests fails the gate).
- **The implementer's handoff carries evidence, not adjectives:** test paths added and the verify command with its actual output — `pnpm test messages.spec → 12 passing`, not "implemented with tests".
- **Adversarial triage is the default, not a request:** any "is this claim true / does this bug exist / verify this finding" task routes to `finding-refuter` without being asked. Budget by risk: trivial/cosmetic finding → no refuter; standard finding → one `finding-refuter` pass; hot path (auth, security, payments, migrations, data integrity) or >400 changed lines → 2-3 independent refuters, majority verdict (2-of-3).
- **Fix budget: two rounds per verification cycle.** A third failed fix escalates per Chain Interruption and the three-fix breaker in `quality/debugging.md` — the loop never silently extends, and restarting verification to reset the budget is the anti-pattern above.

### Chain Interruption
When an agent's output can't feed the next stage (incomplete spec, failing tests, ambiguous review): never silently skip the stage. A technically resolvable gap gets ONE bounded re-run with the clarified ask, reported in the close summary. Second failure — or a gap only the user can resolve — escalates with the specific issue: re-run with clarification, skip with acknowledgment, or adjust the plan.

### Workflow Tool vs Subagents vs Agent Teams
- **Default to a direct subagent.** Explore for search, the domain specialist (table above) for domain work, chained design → impl → tests → review. A single `Agent` call with the right specialist is the correct default for single-domain work — including under `ultracode`. Don't wrap one task in a workflow; don't delegate trivial edits at all.
- **Workflow tool only when** the agent set or execution order is data-driven — computed per run ("review every changed file", fan-out over an unknown-size list) — never when a fixed chain of named agents suffices.
- **Agent Teams only when** you need persistent parallel workers with shared task lists and P2P messaging, or the user asks.
- **Inside a workflow, route each stage to its specialist via `agentType`**, reusing the table above. The default subagent (no `agentType`) is for glue only — fan-out, dedup, synthesis. An unknown `agentType` throws; there is no fallback.
- **Workflows require explicit opt-in:** the `ultracode` keyword, a standing `/effort ultracode` session, an explicit user request, or `bypassPermissions` — never inferred intent alone. *(Version-sensitive: workflow/ultracode are research-preview, Claude Code 2.1.154+.)*
