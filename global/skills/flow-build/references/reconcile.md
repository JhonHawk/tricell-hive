# flow-build — OPEN (reconcile)

The reconciler's entry, run on every `/flow-build` invocation: resolve which plan is being
executed, adopt it if it was born organic, and read the pending state from git rather than
from the plan header.

1. Read `<project>/_support/PROJECT.md` (missing → suggest applying the bootstrap-playbook,
   stop) and resolve the plan: the path at `$ARGUMENTS` (a part file inside an initiative's
   `plan/`, or a single `<slug>-plan.md`); **zero-arg → discover** the current session plan —
   the ledger's `## Current handoff` plan pointer, else the newest `sessions/*/*-plan.md` with
   `Status: planned|building`. Nothing found → report that and stop (native plan mode,
   captured by the plan-capture hook, produces one; a manually placed plan file works too).
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
   | `planned` | execute tasks from `T1` (`references/execute.md`) |
   | `building`, some `T<n>` already in git | **resume** from the first task with no commit — never re-run a landed task (`references/execute.md`) |
   | `built` | run the gate (`references/verify-gate.md`) |
   | `verified` | nothing pending — report and stop |

4. **In-vivo timing — infer, never ask** (relevant only when the plan has ≥1 `in-vivo: yes`
   or `design-review: yes` task; none → nothing to decide). Default: **defer all browser
   walks to `built`** — walks amortize over the change-group (`testing.md > Execution
   Scope`). Go **inline** (walk after each gated task) only on signal: multiple independent
   gated tasks where early walk feedback would redirect later ones, or the plan/user says
   so. State the choice in the start summary; it persists for the run until the user
   overrides it.
