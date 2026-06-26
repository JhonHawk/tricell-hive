---
alwaysApply: true
---

## Gap Resolution in Plans

> Owns **prerequisites** (missing-and-required). For **risks** (present-but-fragile, runtime failure modes), see `quality/critical-thinking.md`. The two rules don't overlap: gap = something missing now; risk = something that could break later.

**Principle:** A gap detected during planning that does not become an actionable task or escalated to the user is a defect in the plan.

### Gap vs Observation
- **Gap** — preexisting condition blocking the feature: missing dependency, incomplete interface, required migration, undefined contract, missing tests for modified code. Becomes a task.
- **Observation** — unrelated tech debt or nice-to-have. Surface with recommendation; user decides.

**Never silently discard.**

### Mandatory Gap Analysis

Runs after reading the spec and before writing the first plan task. Applies to plan mode (Shift+Tab) and any session designing implementation tasks.

**Step 1 — Investigate** with glob/grep/read. The scope spans the full flow — implementation, verification (incl. in-vivo), and what promotion to QA/prod will require — not just the implementation step:
- Files to modify — current state, tests, dependencies
- Interfaces/contracts — exist, complete?
- Dependencies — installed, correct version, configured?
- Infrastructure — migrations, env vars, CI/CD?
- External integrations & capabilities — API credentials, webhook endpoints, third-party tokens, and the CLIs/tools the flow needs (including verification and promotion steps): present? Verify with `which`/`--version`/a presence check — never assume absence, and never assume presence.

**Step 2 — Classify:**

| Classification | Criteria | Action |
|---|---|---|
| Inline gap | Prerequisite, ≤3 tasks to resolve | First tasks in plan |
| Prerequisite gap | Prerequisite, >3 tasks | Separate prerequisite plan |
| Observation | Non-blocking tech debt | Present with recommendation |

**Step 3 — Confirm:** Present gaps to the user before writing tasks when (a) any gap requires more than one inline task to resolve, OR (b) you found observations worth flagging. Otherwise, list any inline gaps in the plan and proceed — no separate confirmation round-trip needed.

### Exceptions
- **Simple single-task changes** (bug fixes, config tweaks, docs): skip 3-step process; principle still applies.
- **Isolated Mode** (worktree/feature branch): findings documented inline in the plan; checkpoints (`git-workflow.md`) cover progress review.

## Divergence Between Sources

> Owns **conflicts** (two present-but-disagreeing sources). Distinct from **gaps** (above: something missing) and **risks** (in `quality/critical-thinking.md`: present-but-fragile).

**Principle:** When two authoritative sources disagree about the same decision, never pick a side silently. The divergence is the signal — not a problem to route around.

### Common patterns

| Source A | Source B |
|---|---|
| Spec / wireframe / API contract | Task description, ticket, Linear issue |
| Code comment / docstring | Actual code behavior |
| README / docs | Current implementation |
| Test assertion | Implementation |
| Two specs / wireframes | Each other |

### Workflow

1. **Stop.** Do not silently follow either source.
2. **Report** to the user or orchestrator: cite both sources with paths, line numbers, and last-modified context.
3. **Wait for resolution**: declare which source is authoritative for this case, OR update one of them to match the other.
4. **Carve-out — proceed only when ALL three hold AND the decision is genuinely the implementer's to make (not a stakeholder/contract/spec decision):**
   - One source is unambiguously stale (older AND reflects pre-decision context).
   - The other clearly reflects current intent.
   - Following it is reversible (UI tweak in a mock — not a migration, not a contract change, not a security control).

   Even then: note the divergence in the commit message and in your handoff. Do not suppress.

### Anti-pattern

Catchphrases that hard-code a permanent priority between two sources ("spec over code", "tests over implementation") hide drift and skip the "flag and update" step. The rule is always "surface, then resolve" — never "follow source X by default".
