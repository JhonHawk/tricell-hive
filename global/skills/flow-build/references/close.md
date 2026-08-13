# flow-build — CLOSE

Runs at the end of every `/flow-build` invocation, after the run's last stage.

1. **Tracker** — move task states, comment evidence + merged-PR refs (via the declared access;
   mcp → dispatch the updates; manual → list them; none → update `tasks.md`).
2. **Git** — confirm nothing open: `gh pr list --state open` in every repo touched, output in the
   report; non-empty is a blocker NOW, never a "next step". No unmerged task branches.
3. **Ledger** — update PROJECT.md (phase/epic progress, artifacts, decisions, **Promotion
   prerequisites** for qa/prod). Finalize the session/initiative if created: mark `finalized` in
   the sessions index, seal back-references (`Session: <slug>` on `tasks.md` rows, `Implementado
   en: sessions/<slug>` on decisions/epics), promote durable findings, save the slug to Engram.
4. **Evidence** — versioned reports already written by the gate; **purge ephemeral raw** from
   `_support/evidence/` (confirm gitignored). The report survives, the screenshots do not.
5. **Docs impact** — state the change-group's specs/product-doc delta: files updated in the
   specs repo (product state, conventions, manual pages), or `none` with the reason. An
   operator-facing change (screen, role capability, business rule) with no delta is a gap to
   close in the change-group, never a "next step".
6. **Report** — tasks done vs pending, merges (PR#, CI results), docs impact (item 5), what
   was verified (paths) vs not, **executor substitutions** (a gate or `Agent:`-annotated task
   run inline instead of via its routing-row agent — name it with the one-line why; a silent
   substitution is a close defect), servers started/stopped, suggested next scope. Deferred tasks are a one-line count, never
   a pending list; promotion is never the suggested next work — if integration is ahead of `qa`,
   close with a one-line offer to run the QA promotion walk now
   (`flow-core/references/promotion-playbook.md`; the user decides; never expand its
   plan, never list it as a user to-do).
