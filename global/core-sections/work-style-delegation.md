---
order: 56
targets: [agents]
join: tight
---

- **Delegation gates are hard.** Any of these fires → delegate, or state in one line why inline is correct: 4+ files to understand · 2+ non-trivial files to write · ~20 undelegated calls · 3+ external reads (tracker, PR checks, deploy jobs, memory, third-party MCP) · 2+ fetches to pin one fact. External sweeps go to `status-fetch` or a research subagent; their payloads re-send every turn, so the cost is the rest of the session. **One route per change-group**, chosen at the first gate that fires and covering understanding AND writing: understanding inline past the gate and then delegating only the write is the named violation. Per-action delegation (tests, builds, installs, verifiers, reviewers) is orthogonal and never changes the route.
- **Dispatching a subagent:** name artifact destinations as ABSOLUTE paths — a relative `_support/` lands in the agent's own repo. A NAMED agent's final text never reaches the spawner, so instruct delivery via SendMessage to "main"; unnamed subagents auto-return.
- **Parallelize only independent work** — no dependency between tasks (an output another consumes, a co-written file, a pending decision); a split that would degrade a result serializes instead. Every dispatch prompt names how the agent asks you at a fork its prompt does not settle; no channel → it reports blocked.
- **Verification is adversarial**: refute via an executed check, never confirm from narration. Verifiers overlap independent work, their verdict collected at the session mode's gate (pre-merge on the PR, else the commit gate), never idle-polled.
- **Independent review scales by surface, not diff size.** Always: security, auth, payments, migrations, data integrity, and resource limits (pools, timeouts, retries, rate limits, TTLs). A trivial diff outside those skips it. A re-review scopes to the finding's blast radius.
- **A multi-repo change-group fans the review out:** one blind reviewer per non-trivially-changed repo, plus a cross-implementation reviewer ONLY when a shared contract changed (input: contract spec + diffs filtered to interface surfaces; mandate: the seams — shapes, status codes, auth/cookie semantics, event payloads, error propagation). One parallel dispatch, still ONE review layer and one verdict.
- **Budgets, so a cycle terminates:** a delegated stage whose output cannot feed the next gets one bounded re-run, then escalates; max two fix rounds per cycle; an approving review closes it and a new finding never resets the budget. A finding blocks only with a concrete failure scenario reproducible now (inputs → wrong outcome); coverage/hardening are follow-ups recorded at close.
- Spawn domain specialists yourself — standing-authorized, never wait for a flow skill or the user to name one (Codex: `spawn_agent` with the agent's name as `agent_type`; TOMLs in `~/.codex/agents/`). The gates above decide WHEN; this table decides WHO. Reviewers and verifiers never implement.

| Signal in the task | Route to | NOT to |
|---|---|---|
| Angular components, services, routing; `angular.json` present | angular-developer | react-developer |
| Next.js, App Router, Server Components; `next.config` present | react-developer | angular-developer |
| React with no Next.js (Vite, React Router, CRA) | react-developer | ts-backend-developer, the main thread |
| Node/TypeScript backend (NestJS, Express/Fastify, Prisma/Drizzle, BullMQ) | ts-backend-developer | backend-developer, database-specialist |
| Backend on any non-Node stack (Java/Spring, Kotlin server, Python) | backend-developer | ts-backend-developer, database-specialist |
| Kotlin Multiplatform, Android, Compose, shared mobile code | kotlin-multiplatform-developer | backend-developer |
| Server-only Kotlin (Ktor/Spring, no Android target) | backend-developer | kotlin-multiplatform-developer |
| Schema design, migration, ORM config | database-specialist | the backend agents |
| Slow query, EXPLAIN, N+1, index tuning | performance-engineer | the backend agents |
| CI/CD, Docker, Terraform, deploy pipeline | devops-engineer | the backend agents, cloud-architect |
| Cloud topology, landing zone, DR (RTO/RPO), FinOps architecture | cloud-architect | devops-engineer, sdd-design |
| New API contract, cross-service schema, service boundaries | sdd-design | review-code |
| Review a diff/PR for correctness and cleanup — the DEFAULT when no other row is primary | review-code | the implementing agent |
| Vulnerability, OWASP, secrets, auth bypass | review-security | review-code |
| Secret scanning, leaked credentials | secrets-auditor | review-security |
| Refute or adversarially verify a finding, claim or diagnosis | review-refuter | review-code, the main thread |
| Tests as the primary objective (coverage push, new E2E suite) | test-engineer | the implementing agent |
| Functional verification of a running app in a real browser: walk ACs + adversarial paths | sdd-verify | review-ux, review-code |
| UX friction in a live mock or deployed flow, navigation review | review-ux | review-code, react-developer |
| Redesign or visually polish a screen that already exists | visual-designer | the framework specialist, review-ux |
| Spec/épica completeness, Gherkin verifiability, quality gate | sdd-spec-reviewer | sdd-product-critic, review-code |
| Challenge necessity/scope/shape of a feature BEFORE implementation | sdd-product-critic | sdd-spec-reviewer |
| Raw client requirements, project intake analysis | sdd-spec-reviewer (intake mode) | the main thread |
| "Where is X / how does Y work", think through an idea, or a question only docs/web settle | sdd-explore | the implementing agent, the main thread's own fetches |
| Prompt design, LLM integration, structured output | prompt-engineer | the backend agents |
| README, ADR, API docs, setup guide — in-repo Markdown; draft or revise an épica/delta spec | sdd-spec-writer (`docs` / `spec`) | the implementing agent, the main thread |
| Workspace file hygiene, misplaced artifacts, ledger repair | workspace-custodian | secrets-auditor |
| Low-reasoning external state: tracker board, PR checks, deploy jobs, an APPROVED tracker batch | state-fetcher | a research subagent |

- Work depending on external tools/credentials/services: verify the full set (`which`, `--version`) — implementation, verification, promotion — before committing to a plan; missing items become explicit asks; promotion credentials documented durably.
- Product-specific slash commands, skills, and agent names are workflow references unless the harness exposes them; inspect or edit the underlying files when unavailable.
