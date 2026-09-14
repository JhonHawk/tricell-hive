# Workspace: <group>/<project>
Read `_support/PROJECT.md` (project ledger) before working — current phase,
artifact index, and open questions live there.
File placement follows the flow-core file-routing rule: versioned material →
`<project>-specs/`; temporary/sensitive/raw → `_support/`.
Session artifacts (all harnesses): the portable plan lives at
`sessions/YYYY-MM-DD-<slug>/<slug>-plan.md` (session-capture layout) and contains a frozen
contract, explicit Authorization and mutable Execution evidence. `Status: planned` means the
contract is ready; it is not implementation consent. A plan carrying `Session: no` means the user
declined the session folder — don't create one and make no durable resume claim. Investigation
conclusions the user asks to keep go to `<slug>-findings.md` in the same layout. When session
artifacts are produced, update the ledger's `## Current handoff` when one exists and commit them
at close only when the session's Git mode authorizes it. `/flow-plan` creates or normalizes the
plan; `/flow-build` adopts and executes it after explicit implementation authority. Native plan
mode is an optional drafting surface, never the Hive approval mechanism. Trivial fixes, small
commits, and investigations without kept artifacts proceed ad-hoc with no session machinery;
epic-scoped formal work still uses `/flow-plan` (format and Preflight:
`flow-core/references/plan-format.md`).
Flow phase offering: offer the next live command (`/flow-plan` or `/flow-build`) or applicable
playbook per the ledger's `Current phase` / `Next suggested`; never execute a
flow command uninvited, and never offer production deploys automatically.
(Full directive lives in each harness's global layer.)

<client conventions block, if any — filled from the conventions step of Stage B>
