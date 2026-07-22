---
name: flow-build
description: >
  Execute an approved plan (the executing stage of the daily dev chain) as a state-driven
  reconciler — resumable, never re-doing landed work. `verify` jumps straight to the
  verification gate. Runs on any harness. Point it at a plan or part.
---

# /flow-build — execute the plan

Follow the flow contract (`~/.agents/skills/flow-core/SKILL.md`). This is the *doing* half of
the dev chain: it executes the plan `/flow-plan write` produced. It is a **reconciler** — it reads the
desired state (the plan) and the observed state (git), and converges toward the next pending
state. It never re-does landed work and never trusts a header over git.

**Runs on any harness.** Where specialist agents are available, dispatch each task to its
specialist (`agent-routing.md`) in a fresh context and keep the main thread orchestrating — the
flow-pack rule. Where they are not (a lighter harness), implement the task's recipe directly:
the plan is self-contained (`references/plan-format.md`) precisely so a cheaper executor can run
it. The in-vivo gate is **not** frontier-only — it runs here too; a teammate who only has this
harness still gates their work.

## OPEN — reconcile

1. Read `<project>/_support/PROJECT.md` (missing → suggest `/flow-start`, stop) and resolve
   the plan: the path at `$ARGUMENTS` (a part file inside an initiative's `plan/`, or a single
   `<slug>-plan.md`); **zero-arg → discover** the current session plan — the ledger's
   `## Current handoff` plan pointer, else the newest `sessions/*/*-plan.md` with `Status:
   planned|building`. Nothing found → report that and stop (plan mode or `/flow-plan` produce
   one; a manually placed plan file works too).
2. **ADOPT — organic plans.** A captured plan-mode plan carries no reconciler metadata; adopt
   it **additively** (never rewrite the approved content): derive `T<n>` task boundaries from
   its steps, add `Verify:`/expected-output per task where the plan implies them, and the
   `Commit:` tags. If adoption introduces **integration semantics the approved plan did not
   carry** (branching model, PR/merge flow, CI gates), that delta folds into this run's
   question block as a confirmation — the user approved a plan without those semantics, so
   they are asked, not assumed. Write the adopted metadata back to the plan file.
3. Read the plan's `Status` and `git log` (grep the task IDs `T<n>`). **Status is a claim,
   git is ground truth** — organic commits may carry no `T<n>` tags: when tags are absent,
   match tasks against the actual log/diff (files touched) before deciding anything landed.
   Determine the pending state:

   | `Status` + git | Do |
   |---|---|
   | `planned` | execute tasks from `T1` |
   | `building`, some `T<n>` already in git | **resume** from the first task with no commit — never re-run a landed task |
   | `built` | run the gate (below) |
   | `verified` | nothing pending — report and stop |

4. **In-vivo timing — infer, never ask** (relevant only when the plan has ≥1 `in-vivo: yes`
   or `design-review: yes` task; none → nothing to decide). Default: **defer all browser
   walks to `built`** — walks amortize over the change-group (`testing.md > Execution
   Scope`). Go **inline** (walk after each gated task) only on signal: multiple independent
   gated tasks where early walk feedback would redirect later ones, or the plan/user says
   so. State the choice in the start summary; it persists for the run until the user
   overrides it.

## Execute (`planned`/`building` → `built`)

Per task, in plan order — integration strictly serialized so sibling branches never coexist
unmerged (each task's gate validates the integrated state of every task before it):

1. **Branch** — checkout the integration branch, pull, cut the task branch (per the plan's naming;
   never from a sibling branch — task N+1 branches from the trunk that already contains N).
2. **Build the task** — dispatch to its specialist per the handoff protocol (intent + consumer,
   the task's Files/Interfaces/steps as bounded context, conventions, return contract), or
   implement the recipe directly on a harness without specialists. Dependent tasks receive the
   predecessor's concrete outputs.
3. **Verify the claim** — run the task's `Verify:` command and confirm it matches the expected
   output. An unverified "it works" is not done; the orchestrator re-checks via the shell.
4. **Per-task gate** — the two-stage review (below) on the task diff; if in-vivo timing is
   **inline** and the task is `in-vivo: yes`, its walk runs now. Findings route back to the
   builder; the affected stage re-runs.
5. **Commit, PR, CI, merge** — one commit per verified task with the `Commit: feat(<scope>):
   T<n> …` tag (the tag is how state is read from git — never batch tasks into one commit). A
   plan may declare **change-groups** (cohesive runs of small tasks): one branch/PR per group,
   tasks landing as sequential tagged commits on it, gates amortizing per `testing.md > Execution
   Scope`; undeclared → one PR per task. Push, open the PR, and **overlap the CI wait**
   (`gh pr checks <n> --watch`) with the task's non-integrative ceremony (ledger notes, evidence
   filing, next dispatch prep) — never merge red or pending; the session owns the wait, it is
   never handed to the user. CI failure → route to the builder, fix
   on the same PR, re-run the local gate on the affected subset BEFORE re-pushing. Then merge per
   the plan's mechanics, delete the task branch (local + remote), checkout the integration branch,
   pull.

- **Three failed fixes on the same symptom = stop.** Do not dispatch a fourth — the pattern is
  the suspect, not the hypothesis. Bring the symptom, the attempts, and the architecture question
  to the user (`debugging.md`).
- **Git is autonomous inside this flow** (invoking `/flow-build` is the authorization): per-task
  commits + the PR/CI/merge cycle. `qa`/prod promotions run via git conventions
  (`git-workflow.md`) with `flow-core/references/promotion-playbook.md`; the safety
  gates of `git-workflow.md` (protected branches, no force-push/rewrites) never relax.
- When every task is in git → set `Status: built`.

## Gate (`built` → `verified`) — also the `verify` subcommand

`verify` jumps straight here (assumes `built`); the default reaches here after Execute. Run, for
the tasks not yet gated:

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
   else `<repo>/_support/sessions/<slug>/reports/`) per `references/test-report-template.md`. **A
   `blocked` AC is not a pass** — unblock at the root (seed missing reference data, fix the
   precondition) and re-run; escalate only a genuinely external blocker as a ledger **Promotion
   prerequisite**, never closed as done. **Visually broken is a defect, not a cosmetic note:**
   anything detected visually broken during the walk — layout overflow, clipped or capped text,
   overlapping elements, content not filling its container — is fixed in this cycle like a failing
   AC, never deferred as polish. (The design gate below owns *craft*; this owns *breakage* and
   applies even when `design-review` is not set.)
3. **Design gate** for `design-review: yes` tasks (opt-in; user-facing UI tasks set the flag in the
   plan, mirroring `in-vivo: yes`): dispatch **ux-flow-reviewer** against the running app on the
   **Visual craft** rubric axis (`flow-core/references/ux-rubric.md` #11–17; criteria
   `languages/ui-visual-design.md`) — type scale, spacing system, color & WCAG-AA contrast, action
   hierarchy, elevation, borders restraint, component simplicity. Same evidence/report routing as
   the in-vivo gate. **A craft `blocker` is not a pass** — fix at the root and re-walk; `friction`/
   `polish` may pass with the user's recorded acknowledgement.
4. **Integrated smoke** when 2+ tasks merged or any conflict was resolved: serve a **production
   build per app** (global `Execution` rule — never the dev server, one app at a time) — it
   validates the state QA receives. Stop any server this flow started (verify per port:
   `lsof -nP -iTCP:<port> -sTCP:LISTEN`).
5. **test-engineer** only if the goal includes a coverage push — specialists test their own code.
- Gate passes → set `Status: verified`.

## CLOSE

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
5. **Report** — tasks done vs pending, merges (PR#, CI results), what was verified (paths) vs
   not, servers started/stopped, suggested next scope. Deferred tasks are a one-line count, never
   a pending list; promotion is never the suggested next work — if integration is ahead of `qa`,
   close with a one-line offer to run the QA promotion walk now
   (`flow-core/references/promotion-playbook.md`; the user decides; never expand its
   plan, never list it as a user to-do).
