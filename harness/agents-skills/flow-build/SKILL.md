---
name: flow-build
description: >
  Execute an approved plan (the build half of F6 of the flow pack). A state-driven reconciler:
  it reads the plan's Status + git, executes the pending tasks, and runs the in-vivo gate.
  `verify` jumps straight to the gate. Runs on any harness. Point it at a plan or part.
---

# /flow-build — execute the plan

Follow the flow contract (`~/.agents/skills/flow-core/SKILL.md`). This is the *doing* half of
F6: it executes the plan `/flow-plan write` produced. It is a **reconciler** — it reads the
desired state (the plan) and the observed state (git), and converges toward the next pending
state. It never re-does landed work and never trusts a header over git.

**Runs on any harness.** Where specialist agents are available, dispatch each task to its
specialist (`agent-routing.md`) in a fresh context and keep the main thread orchestrating — the
flow-pack rule. Where they are not (a lighter harness), implement the task's recipe directly:
the plan is self-contained (`references/plan-format.md`) precisely so a cheaper executor can run
it. The in-vivo gate is **not** frontier-only — it runs here too; a teammate who only has this
harness still gates their work.

## OPEN — reconcile

1. Read `<project>/_support/PROJECT.md` (missing → suggest `/flow-kickoff`, stop) and the plan
   or part at `$ARGUMENTS` (a part file inside an initiative's `plan/`, or a single `<slug>-plan.md`).
2. Read the plan's `Status` and `git log` (grep the task IDs `T<n>`). Determine the pending state:

   | `Status` + git | Do |
   |---|---|
   | `planned` | execute tasks from `T1` |
   | `building`, some `T<n>` already in git | **resume** from the first task with no commit — never re-run a landed task |
   | `built` | run the gate (below) |
   | `verified` | nothing pending — report and stop |

3. **In-vivo timing — ask ONCE per session, before T1, only if the plan has ≥1 `in-vivo: yes`
   task.** *Run the in-vivo walk inline (after each gated task) or defer all walks to `built`?*
   The answer is a **session decision that persists** for the whole run — never re-asked per task,
   even with 10 tasks — until the user changes it (the once-per-session pattern of
   `git-workflow.md`). No in-vivo tasks → no question.

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
   T<n> …` tag (the tag is how state is read from git — never batch tasks into one commit). Push,
   open the PR, **wait for CI** (`gh pr checks <n> --watch`; never merge red or pending — the
   session owns the wait, it is never handed to the user). CI failure → route to the builder, fix
   on the same PR, re-run the local gate on the affected subset BEFORE re-pushing. Then merge per
   the plan's mechanics, delete the task branch (local + remote), checkout the integration branch,
   pull.

- **Three failed fixes on the same symptom = stop.** Do not dispatch a fourth — the pattern is
  the suspect, not the hypothesis. Bring the symptom, the attempts, and the architecture question
  to the user (`debugging.md`).
- **Git is autonomous inside this flow** (invoking `/flow-build` is the authorization): per-task
  commits + the PR/CI/merge cycle. Never touch protected branches (`qa`/prod promotions are
  `/flow-deploy`), never force-push or rewrite history.
- When every task is in git → set `Status: built`.

## Gate (`built` → `verified`) — also the `verify` subcommand

`verify` jumps straight here (assumes `built`); the default reaches here after Execute. Run, for
the tasks not yet gated:

1. **Two-stage review, strict order, fresh contexts** — (a) **spec compliance**: the task diff
   against the task text + ACs (missing, extra, misunderstood — nothing else); (b) **code
   quality** (code-reviewer) over the same diff, only after (a) passes. The diff is the input;
   the builder's report travels as claims to check, never as context to trust.
2. **In-vivo gate** for `in-vivo: yes` tasks (now, if timing was deferred): dispatch
   **in-vivo-qa-tester** against the running app — it walks the Gherkin ACs AND the
   negative/adversarial cases (double-click, invalid input, gated-route access, mid-flow refresh,
   expired session). Raw evidence → `_support/evidence/YYYY-MM-DD-<slug>/` (gitignored); the
   **versioned report** → the versioned layer (specs repo `<project>-specs/evidence/<epic-id>/`,
   else `<repo>/_support/sessions/<slug>/reports/`) per `references/test-report-template.md`. **A
   `blocked` AC is not a pass** — unblock at the root (seed missing reference data, fix the
   precondition) and re-run; escalate only a genuinely external blocker as a ledger **Promotion
   prerequisite**, never closed as done.
3. **Integrated smoke** when 2+ tasks merged or any conflict was resolved: serve a **production
   build per app** (`build` then `start`, never `next dev --turbopack`, never a monorepo-root
   start-all) — it validates the state QA receives. Stop any server this flow started (verify per
   port: `lsof -nP -iTCP:<port> -sTCP:LISTEN`, one port per call, stderr visible).
4. **test-engineer** only if the goal includes a coverage push — specialists test their own code.
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
   close with a one-line pointer to `/flow-deploy qa`.
