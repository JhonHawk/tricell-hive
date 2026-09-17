# Unattended Autonomy Mode — full mechanics

> Loaded through the `unattended-delegation` skill on every harness (Claude Code and Grok read it from `~/.claude/skills/unattended-delegation/references/`; Codex and opencode from `~/.agents/skills/`). The always-on stub at `rules/workflow/unattended-autonomy.md` carries only the activation guard and the gate pointers; this file is the mode itself — a run declared without these controls is not the mode.

> An explicitly-delegated unattended run: the user hands over control and leaves ("tienes control total esta noche", "no preguntes hasta que vuelva", "me voy a dormir, sigue tú"). ONE mode — no variants — with two explicit ways in: a per-run handover, or a **standing job** declared in the project ledger (below). Distinct from a *non-delegated* unattended turn (cron, workflow stage, background job with no declared job), which stays fail-closed per `quality/reporting-integrity.md > Fix at the Root`. The mode is a trade: removed confirmations are compensated by ADDED controls — decision log, reversible checkpoints, queued escalations, automatic expiry. Without those controls the delegation is not in effect.

## Activation — explicit only

- Only an explicit user declaration activates the mode. Silence, absence, or a long-running task NEVER activate it — a user who goes quiet has not delegated anything.
- On activation, declare the mode visibly before proceeding: the scope accepted, the absolute gates that stay closed, and where the decision log will live. Name the mode in every report produced under it.

## Widened scope — proceed-and-log

Applies only to an activated run (above); a *non-delegated* unattended turn stays fail-closed instead (`quality/reporting-integrity.md > Fix at the Root`).

- In-scope reversible technical decisions, resolvable blockers, dependency picks with a safe version (OSV tiers unchanged), test/build/verify loops, and already standing-authorized non-prod ops → proceed; never bounce these back as questions.
- **Integrating the run's OWN work is inside the scope, up to the production line.** The dedicated branch merges into the integration branch through its own PR once checks and the review pass are green, and promotion into a non-prod environment branch (`qa`) is declared, not asked (`git-workflow.md > Safety gates`) — a run told to land something in a non-prod environment cannot do it otherwise. The production-deploying branch is where it stops: a merge or promotion into it queues for the human, always.
- Unattended runs carry full verification rigor: the attended fast-feedback carve-out (`quality/testing.md > Execution Scope`) never applies without a user present to validate.
- Every widened decision lands in the **decision log**: what was decided, why, how to revert. The log is a structural component of the mode — proceeding without logging is outside the delegation.

## Tracker-scoped runs

- A handover bounded by the project's declared tracker ("work the sprint board while I'm away", "ve cerrando los tickets sin preguntarme") is this same mode — a scope parameter, not a variant. The declared tracker (`memory-routing.md > Tracker sync`) defines the work list; an undeclared tracker cannot scope a run.
- Execute item-by-item: one reversible change-group per ticket on the run's dedicated branch, one decision-log entry per ticket, per-ticket state in the close report (done / blocked / queued). A ticket requiring anything beyond the widened scope queues like any gated decision; tracker writes (status changes, comments) batch to the close per `memory-routing.md`.

## Standing jobs — declared recurring scope

- Recurring, signal-driven work (a schedule, PR events on a repo, a CI run going red on a branch, a channel) is this same mode with the scope declared ONCE in the ledger's `## Standing jobs` table (`flow-core/references/ledger-template.md`) instead of handed over per run — a scope parameter, like a tracker-scoped run, never a variant.
- **The declaration is the activation key.** A scheduled or background turn runs delegated only under a row whose trigger matches it; the turn reads the ledger at OPEN and names the row in every report. No matching row → non-delegated turn → fail closed (`quality/reporting-integrity.md > Fix at the Root`). Prompt-convention; no hook backs it yet.
- **Max action: propose, never land.** A standing job may push to its own branch, open or update PRs, comment, and open tickets in the declared tracker. It never merges, promotes, or touches a shared ref — narrower than a per-run handover (above), because nobody reads a daily job's decision log before its next run; landing is the human's or a later attended session's.
- **Review never decreases with volume.** Every PR the job opens takes the repo's declared review route (the recommendation stands in for the question no one is there to answer) exactly as a human-opened one; rising throughput never lowers the route. When the job meets the same defect class twice, it proposes the guard (a lint rule, a test) as its own PR instead of fixing instances forever (`development-principles.md > Fix the cause, not the check`).
- Same controls per run: one decision-log entry at the path the row names (a run with no signal still logs a no-op line), a dedicated branch, queued escalations in the run's close report.
- Expiry is the row's date or condition, never open-ended: the first run past it reports and stops. Editing or removing the row is the revocation.

## Absolute gates — never relax

Identical to attended mode; no gate's behavior depends on which mode is active. This list is enumerated here — not just pointed at — because activation requires declaring it, and because it overrides any repo-level standing authorization the individual owners allow:

- Destructive/irreversible operations · production (deploys, DNS, infra, data — including any merge or promotion INTO the production-deploying branch) · secrets · history rewrites / force-push · direct commits or pushes to protected refs · CRITICAL/HIGH supply-chain with no safe path · **data deletion**.
- Scope and mechanics live with the owners: `CLAUDE.md > Destructive Operations`, `workflow/git-workflow.md > Safety gates`, `quality/security.md`.
- Gates are inviolable constraints, never a judgment call to reason against. Where a deterministic layer exists (deny permissions, hooks), it backs them; the mode never argues past a denial.

## Gated decisions — pause-and-queue, else fail closed

- A gate or stakeholder decision hit mid-run is **queued with full context** — the decision, the options, a recommendation, and what it blocks — and the run continues with independent work.
- When nothing independent remains → fail closed: checkpoint the work, finish the log, stop. Never timeout-default-proceed — a sleeping user's silence is not consent.

## Reversibility bias

- All mode work happens on a dedicated branch with granular commits at natural seams; the only shared-ref mutations are the run's own PR merges and non-prod promotions, which stay revertible as merge commits. Reverting the entire run must be one branch reset or one revert chain, not archaeology.

## Expiry & revocation

- Run-bound for a per-run handover: the mode expires when the delegated run ends, and the user's first message revokes it instantly. It never carries into the next task or session. A standing job is date- or condition-bound by its ledger row (above).

## Close report

- The first interaction after the run leads with the decision log: widened decisions taken, queued gate decisions with their recommendations, and per-criterion verification state (`quality/reporting-integrity.md > Fix at the Root`). Queued decisions resolve before any merge or promotion of the run's branch.
