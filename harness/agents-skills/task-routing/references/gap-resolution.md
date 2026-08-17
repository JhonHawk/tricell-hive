
## Gap Resolution in Plans

> Owns **prerequisites** (missing-and-required). Risks (present-but-fragile) → `quality/critical-thinking.md`; conflicts between present-but-disagreeing sources → Divergence below. **Principle:** a gap detected during planning either becomes a plan task or reaches the user at the plan gate — a gap silently dropped is a plan defect.

**Gap vs observation.** A gap is a preexisting condition blocking the feature (missing dependency, incomplete interface, required migration, undefined contract, missing tests for modified code) — it becomes a task. An observation is unrelated tech debt — surface with a recommendation; the user decides. Never silently discard either.

### Gap analysis — investigate always, ritualize only on findings

- **Investigation is unconditional, scaled to the plan's surface.** Before writing tasks, verify what the plan assumes with glob/grep/read or a cheap presence check (`which`, `--version`): files' current state and tests, interfaces/contracts, dependency versions and config, infrastructure (migrations, env vars, CI/CD), and the credentials/CLIs the full flow needs — implementation, verification (incl. in-vivo), and promotion. Never assume presence or absence.
- **Findings drive the formality.** Clean check and no open decisions → write the plan and say the check was clean. Real gaps, or 3+ blocking decisions → classify and materialize (below).
- **Classify by size:** a gap resolvable in ≤3 tasks leads the plan (first tasks, visible at the plan gate); >3 tasks → propose a separate prerequisite plan. Observations ride along as recommendations.
- **The plan gate is THE consolidated block.** Gaps, stakeholder decisions, contract confirmations (`cross-service-workflow.md`), and the routing chain (`agent-routing.md`) fold into ONE interaction. A separate round-trip before the gate exists only when a gap changes the task's scope or is a stakeholder call.

### Decisions to close before executing

A decision that blocks task detail is resolved at the plan gate, never mid-execution. After the gap analysis, enumerate them and classify by owner:

- **Technical** (the implementer's to make): state the recommendation and proceed. Verify against a peer/tool only on signal — version-sensitive surface (context7), low reversibility, or genuine dispute — not as ritual for every choice.
- **Stakeholder/PO:** fold ALL into the plan gate so the user answers once. When ownership is unclear → treat as stakeholder.
- **3+ blocking decisions → materialize the table** (`Decision · What's decided · Type · Recommendation · Blocks`); 1-2 → plain prose in the plan gate. A decision that blocks nothing resolves inline.
- **Batching ≠ deferral.** The gate covers what is knowable at plan time; a blocker DISCOVERED mid-execution escalates in the moment — that it was discoverable earlier is a planning defect, but never a reason to sit on it.

### Exceptions
Simple single-task changes (bug fix, config tweak, docs): no ritual — the principle (investigate cheaply, never drop a gap) still applies.

## Divergence Between Sources

> Owns **conflicts**: two authoritative sources disagreeing on the same decision — spec vs ticket, comment vs code, test vs implementation, two specs vs each other. Never pick a side silently, and never hard-code a permanent winner ("spec over code", "tests over implementation") — hard-coded priority hides drift. The rule is surface, then resolve.

- **Default: surface it.** Cite both sources (paths, lines, last-modified context) and get a resolution — which source is authoritative for this case, or update one to match the other.
- **Resolve-by-evidence carve-out** — proceed without a round-trip only when ALL hold: (1) live state or the record trail proves one source verifiably stale (older AND pre-decision); (2) the decision is genuinely the implementer's — stakeholder, contract, migration, and security calls stay hard-blocking; (3) following the current source is reversible. **The carve-out is always reported** — commit message + handoff/close summary. An unreported resolution is a silent one.
- **Unattended runs fail closed:** no human present → report blocked; never self-authorize either side (`quality/development-principles.md > Fix at the Root`). An explicitly-delegated run (the `unattended-delegation` skill) queues the conflict with both sources cited and continues independent work instead.
