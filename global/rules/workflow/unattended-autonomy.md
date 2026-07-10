---
alwaysApply: true
---

## Unattended Autonomy Mode

> An explicitly-delegated unattended run: the user hands over control and leaves ("tienes control total esta noche", "no preguntes hasta que vuelva", "me voy a dormir, sigue tú"). ONE mode — no variants. Distinct from a *non-delegated* unattended turn (cron, workflow stage, background job), which stays fail-closed per `quality/development-principles.md > Fix at the Root`. The mode is a trade: removed confirmations are compensated by ADDED controls — decision log, reversible checkpoints, queued escalations, automatic expiry. Without those controls the delegation is not in effect.

### Activation — explicit only
- Only an explicit user declaration activates the mode. Silence, absence, or a long-running task NEVER activate it — a user who goes quiet has not delegated anything.
- On activation, declare the mode visibly before proceeding: the scope accepted, the absolute gates that stay closed, and where the decision log will live. Name the mode in every report produced under it.

### Widened scope — proceed-and-log
- In-scope reversible technical decisions, resolvable blockers, dependency picks with a safe version (OSV tiers unchanged), test/build/verify loops, and already standing-authorized non-prod ops → proceed; never bounce these back as questions.
- Every widened decision lands in the **decision log**: what was decided, why, how to revert. The log is a structural component of the mode — proceeding without logging is outside the delegation.

### Absolute gates — never relax
Identical to attended mode; no gate's behavior depends on which mode is active:
- Destructive/irreversible operations · production (deploys, DNS, infra, data) · secrets · history rewrites / force-push · merge or push to shared/protected refs · CRITICAL/HIGH supply-chain with no safe path · data deletion.
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
