# Standing jobs — recurring delegated work (design decision)

Status: Accepted (2026-09-13)

## Decision

Recurring, signal-driven work ("gardening": follow new PRs, react to CI going red, run on a schedule, listen to a channel) enters Hive as a **scope parameter of the existing unattended mode**, not as a third contract. It is activated by a row in the project ledger (`PROJECT.md > Standing jobs`: trigger, scope, may-do, expiry, decision-log path) instead of a per-run handover. Max action is **propose, never land**: push its own branch, open/update PRs, comment, open tickets; never merge, promote, or touch a shared ref. Review never decreases with volume. Expiry is date- or condition-bound; editing or removing the row revokes.

## Why

- The product metaphor comes from Cursor Projects (2026-09-10; bibliography): a body of work with a coordinator, shared files, and subscriptions to signals. Hive already had the unit (ledger + portable plan + `/flow-build`) and the persistence model (state in git and memory, disposable session). What it lacked was a declared home for "what this body of work watches when no session is open".
- Before this, every non-delegated unattended turn (cron, background) was fail-closed and the delegated mode expired with the run, so signal-driven work had no legal shape. `ONE mode — no variants` is preserved by making the declaration the activation key, the same way tracker-scoped runs are a scope parameter.
- A prerequisite premise was measured first and rejected: the main thread does not over-edit during plan execution (`_support/archive/audits/2026-09-13-main-thread-edits.md`), so no mechanical write gate accompanies this change.

## Rejected from the source

- The persistent months-long thread as the unit's home (Hive: the session dies, state lives in git/Engram/ledger).
- Always-on without a declaration ("proactively manages work"): a turn with no matching row stays fail-closed.
- "As the fixes hold up, you review less": the review route is fixed per repo and never lowers with throughput; the job proposes the guard on the second occurrence of a defect class instead.
- Cloud-by-default is a product property, not a convention; Hive does not emulate it.

## Enforcement

Prompt-convention: the turn reads the ledger at OPEN (flow contract) and names the row in every report. Candidate deterministic backstop, not built: a SessionStart hook for scheduled turns that refuses to proceed without a matching row. Revisit when a second project runs a standing job.

## Files

`rules-situational/unattended-autonomy-mode.md` (mechanics), `rules/workflow/unattended-autonomy.md` (always-on activation guard), `quality/reporting-integrity.md` (fail-closed clause), core sections `on-demand-rules` and `destructive-operations`, skill `unattended-delegation`, `flow-core/references/ledger-template.md`, the flow manual.
