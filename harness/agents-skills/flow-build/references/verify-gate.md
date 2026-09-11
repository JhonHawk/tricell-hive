# flow-build — Gate (`built` → `verified`), also the `verify` subcommand

The verification gate. `verify` jumps straight here (assumes `built`); the default reaches here
after Execute. Run the applicable checks against the integrated working-tree or committed state
for the tasks not yet gated:

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
   prerequisite**, never closed as done. **Visual judgment belongs to the design gate below** —
   this agent verifies function, not appearance, and takes no systematic per-screen captures.
   Its captures stay what they always were: repro evidence attached to a finding. Severity floor
   if it happens to see breakage while walking: `major` at least, never `minor` — it reports it
   and moves on, it does not go looking.

   **The two gates run in PARALLEL when the project can isolate them** — a separate account and
   tenant per agent, plus a distinct `agent-browser` session. Resolve that at the plan gate with
   the rest of the in-vivo prerequisites (`gap-resolution.md`), never at dispatch. Without
   isolation they SERIALIZE in this order: **design gate first, in-vivo second** — in-vivo
   deliberately destroys state (duplicate records, expired sessions, aborted requests) and a
   reviewer walking that wreckage reports defects that do not exist.
3. **Design gate — REQUIRED for any task with a user-facing surface**, not opt-in (canonical
   statement of the gate: `agent-routing.md > Verification runs in fresh context`, which binds a
   portable plan too — this file carries its mechanics). Dispatch **ui-reviewer** against the running app.
   **The axis scope is DERIVED from what a mock review actually covered for THIS feature — never
   from whether the project owns a mocks repo.** A mock review ran → the gate takes **Visual
   craft** (`flow-core/references/ux-rubric.md` #11–18) plus the five Flow dimensions a prototype
   cannot produce (#3 feedback, #4 error recovery, #5 empty states, #9 role coherence, #10
   responsive & a11y — each needs a real backend, real zero-result queries, real authz). **No mock
   review for this feature → the gate takes all 18**: half the portfolio has no mocks repo at all,
   and there the Flow axis is judged by nobody otherwise. Criteria `languages/ui-visual-design.md` — type scale, spacing system, color & WCAG-AA contrast, action
   hierarchy, elevation, borders restraint, component simplicity, net improvement — **plus
   BREAKAGE, which this gate owns**: layout overflow, clipped or capped text, overlapping
   elements, content not filling its container. Breakage is never `polish` — it fixes in-cycle
   like a failing AC. Same
   evidence/report routing as the in-vivo gate. **A craft `blocker` is not a pass** — fix at the
   root and re-walk; `friction`/`polish` may pass with the user's recorded acknowledgement.
   - **When the task changes an existing screen, capture the pre-change state BEFORE the run's
     first edit** — same viewports and themes the walk will use, each set explicitly per
     `tools/browser-automation.md`, into the run's raw-evidence
     folder — and hand both captures to the reviewer for dimension 18. Missing pre-change capture
     → the design gate reports **not-verified**, never pass.
3b. **Evidence appraisal — prepared at gate close, presented in the close report.** Partition the
   run's raw evidence into three, and resolve the DESTINATION by the established level, never by
   cwd: a workspace that already keeps dated `evidence/` folders at one level takes the new one at
   that SAME level (`project-structure.md`); only absent an established level does scope decide.
   - **discard** — captures of a finding now resolved, and intermediate states that produced no
     finding: reproducible, no communicative value.
   - **keep, always** — captures of findings still OPEN: they are the live bug's evidence.
   - **keep, proposed** — the final state of each screen the plan created or changed, on its happy
     path. These are the client-demo and user-manual material. The ORCHESTRATOR picks them from
     the plan's tasks, not the agent from what it happened to capture.
   Retained → `_support/evidence/YYYY-MM-DD-<slug>/` as lossless WebP (`support-artifacts.md`).
   **One confirmation covers the batch and nothing is deleted without it** — data deletion always
   confirms (`CLAUDE.md > Destructive Operations`), in every mode.

4. **Integrated smoke** when 2+ tasks merged or any conflict was resolved: serve a **production
   build per app** (global `Execution` rule — never the dev server, one app at a time) — it
   validates the state QA receives. Stop any server this flow started (verify per port:
   `lsof -nP -iTCP:<port> -sTCP:LISTEN`).
5. **test-engineer** only if the goal includes a coverage push — specialists test their own code.
- Gate passes → set `Status: verified`. This records the quality result only. Reconcile any
  separately authorized commit, push, PR, merge or deploy action afterward; the `verify`
  subcommand never performs publication.
