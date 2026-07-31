# Workspace: <group>/<project>
Read `_support/PROJECT.md` (project ledger) before working — current phase,
artifact index, and open questions live there.
File placement follows the flow-core file-routing rule: versioned material →
`<project>-specs/`; temporary/sensitive/raw → `_support/`.
Session artifacts (all harnesses): an approved plan lives at
`sessions/YYYY-MM-DD-<slug>/<slug>-plan.md` (session-capture layout), first line
`Status: planned`; a plan carrying `Session: no` means the user declined the
session folder — don't create one. Investigation conclusions the user asks to
keep go to `<slug>-findings.md` in the same layout. When session artifacts are
produced, update the ledger's `## Current handoff` and commit them at close —
standing-authorized, `chore(sessions): <slug>`. `/flow-build` adopts and
executes any session plan. Trivial fixes, small commits, and investigations
without kept artifacts proceed ad-hoc with no session machinery; epic-scoped
formal work may still enter through `/flow-plan` (no command exposed → follow
the skill files directly).
Flow phase offering: offer the next `/flow-*` command per the ledger's
`Current phase` / `Next suggested`; never execute a flow command uninvited,
and never offer production deploys automatically. (Full directive lives in
each harness's global layer.)

<client conventions block, if any — filled from the conventions step of Stage B>
