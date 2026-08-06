
## Unattended Autonomy Mode

> Always-on in Claude Code; Codex/opencode reach it through the `unattended-delegation` skill, which they must load BEFORE declaring the mode accepted. Two guards stand in front of it either way: activation is explicit-only (`CLAUDE.md > Destructive Operations`) and every absolute gate below is owned by an always-on rule (`> Absolute gates`).

> An explicitly-delegated unattended run: the user hands over control and leaves ("tienes control total esta noche", "no preguntes hasta que vuelva", "me voy a dormir, sigue tú"). ONE mode — no variants. Distinct from a *non-delegated* unattended turn (cron, workflow stage, background job), which stays fail-closed per `quality/development-principles.md > Fix at the Root`. The mode is a trade: removed confirmations are compensated by ADDED controls — decision log, reversible checkpoints, queued escalations, automatic expiry. Without those controls the delegation is not in effect.

### Activation — explicit only
- Only an explicit user declaration activates the mode. Silence, absence, or a long-running task NEVER activate it — a user who goes quiet has not delegated anything.
- On activation, declare the mode visibly before proceeding: the scope accepted, the absolute gates that stay closed, and where the decision log will live. Name the mode in every report produced under it.

### Widened scope — proceed-and-log
Applies only to an activated run (above); a *non-delegated* unattended turn stays fail-closed instead (`quality/development-principles.md > Fix at the Root`).
- In-scope reversible technical decisions, resolvable blockers, dependency picks with a safe version (OSV tiers unchanged), test/build/verify loops, and already standing-authorized non-prod ops → proceed; never bounce these back as questions.
- Unattended runs carry full verification rigor: the attended fast-feedback carve-out (`quality/testing.md > Execution Scope`) never applies without a user present to validate.
- Every widened decision lands in the **decision log**: what was decided, why, how to revert. The log is a structural component of the mode — proceeding without logging is outside the delegation.

### Tracker-scoped runs
- A handover bounded by the project's declared tracker ("work the sprint board while I'm away", "ve cerrando los tickets sin preguntarme") is this same mode — a scope parameter, not a variant. The declared tracker (`memory-routing.md > Tracker sync`) defines the work list; an undeclared tracker cannot scope a run.
- Execute item-by-item: one reversible change-group per ticket on the run's dedicated branch, one decision-log entry per ticket, per-ticket state in the close report (done / blocked / queued). A ticket requiring anything beyond the widened scope queues like any gated decision; tracker writes (status changes, comments) batch to the close per `memory-routing.md`.

### Absolute gates — never relax
Identical to attended mode; no gate's behavior depends on which mode is active. This list is enumerated here — not just pointed at — because activation requires declaring it, and because it overrides any repo-level standing authorization the individual owners allow:
- Destructive/irreversible operations · production (deploys, DNS, infra, data) · secrets · history rewrites / force-push · merge or push to shared/protected refs · CRITICAL/HIGH supply-chain with no safe path · **data deletion**.
- Scope and mechanics live with the owners: `CLAUDE.md > Destructive Operations`, `workflow/git-workflow.md > Safety gates`, `quality/security.md`.
- Gates are inviolable constraints, never a judgment call to reason against. Where a deterministic layer exists (deny permissions, hooks), it backs them; the mode never argues past a denial.

### Gated decisions — pause-and-queue, else fail closed
- A gate or stakeholder decision hit mid-run is **queued with full context** — the decision, the options, a recommendation, and what it blocks — and the run continues with independent work.
- When nothing independent remains → fail closed: checkpoint the work, finish the log, stop. Never timeout-default-proceed — a sleeping user's silence is not consent.

### Reversibility bias
- All mode work happens on a dedicated branch, granular commits at natural seams, zero mutation of shared refs. Reverting the entire run must be one branch reset, not archaeology.

### Expiry & revocation
- Run-bound: the mode expires when the delegated run ends, and the user's first message revokes it instantly. It never carries into the next task or session.

### Close report
- The first interaction after the run leads with the decision log: widened decisions taken, queued gate decisions with their recommendations, and per-criterion verification state (`quality/development-principles.md > Fix at the Root`). Queued decisions resolve before any merge or promotion of the run's branch.
