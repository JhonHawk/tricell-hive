# flow-build — Gate (`built` → `verified`), also the `verify` subcommand

The verification gate. `verify` jumps straight here (assumes `built`); the default reaches here
after Execute. Run, for the tasks not yet gated:

1. **Two-stage review, fresh contexts, parallel dispatch** — (a) **spec compliance**: the task
   diff against the task text + ACs (missing, extra, misunderstood — nothing else); (b) **code
   quality** (code-reviewer) over the same diff. Neither consumes the other's output — dispatch
   both in ONE message and merge findings; a trivial diff outside hot surfaces
   (`agent-routing.md > Independent review scales by surface`) collapses to a single reviewer
   carrying both lenses. The diff is the input;
   the builder's report travels as claims to check, never as context to trust.
2. **In-vivo gate** for `in-vivo: yes` tasks (now, if timing was deferred): dispatch
   **in-vivo-qa-tester** against the running app — it walks the Gherkin ACs AND the
   negative/adversarial catalog the agent owns. Raw evidence → `_support/evidence/YYYY-MM-DD-<slug>/` (gitignored); the
   **versioned report** → the versioned layer (specs repo `<project>-specs/evidence/<epic-id>/`,
   else `<repo>/_support/sessions/<slug>/reports/`) per `flow-core/references/test-report-template.md`. **A
   `blocked` AC is not a pass** — unblock at the root (seed missing reference data, fix the
   precondition) and re-run; escalate only a genuinely external blocker as a ledger **Promotion
   prerequisite**, never closed as done. **Visually broken is a defect, not a cosmetic note:**
   anything detected visually broken during the walk — layout overflow, clipped or capped text,
   overlapping elements, content not filling its container — is fixed in this cycle like a failing
   AC, never deferred as polish. (The design gate below owns *craft*; this owns *breakage* and
   applies even when `design-review` is not set.)
3. **Design gate** for `design-review: yes` tasks (opt-in; user-facing UI tasks set the flag in the
   plan, mirroring `in-vivo: yes`): dispatch **ux-flow-reviewer** against the running app on the
   **Visual craft** rubric axis (`flow-core/references/ux-rubric.md` #11–18; criteria
   `languages/ui-visual-design.md`) — type scale, spacing system, color & WCAG-AA contrast, action
   hierarchy, elevation, borders restraint, component simplicity, net improvement. Same
   evidence/report routing as the in-vivo gate. **A craft `blocker` is not a pass** — fix at the
   root and re-walk; `friction`/`polish` may pass with the user's recorded acknowledgement.
   - **When the task changes an existing screen, capture the pre-change state BEFORE the run's
     first edit** — same viewports and themes the walk will use, each set explicitly per
     `tools/browser-automation.md`, into the run's raw-evidence
     folder — and hand both captures to the reviewer for dimension 18. Missing pre-change capture
     → the design gate reports **not-verified**, never pass.
4. **Integrated smoke** when 2+ tasks merged or any conflict was resolved: serve a **production
   build per app** (global `Execution` rule — never the dev server, one app at a time) — it
   validates the state QA receives. Stop any server this flow started (verify per port:
   `lsof -nP -iTCP:<port> -sTCP:LISTEN`).
5. **test-engineer** only if the goal includes a coverage push — specialists test their own code.
- Gate passes → set `Status: verified`.
