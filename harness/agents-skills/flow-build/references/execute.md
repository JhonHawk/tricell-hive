# flow-build — Execute (`planned`/`building` → `built`)

Builds the portable plan's pending tasks. Reached when reconciliation has confirmed explicit
implementation authority and the state is `planned` or `building`; the `verify` subcommand skips
this stage.

1. **Open the authorized workspace.** Confirm the contract digest, implementation scope, session
   Git mode, required repositories and any task prerequisites. If the current tree contains a
   change outside the authorized scope, pause and surface it before writing.
2. **Mark execution.** Set the execution state to `building` and record the run start without
   touching the frozen contract or its digest. Before each meaningful task, delegation, fix,
   review round, or remote operation, append an open recovery attempt with its stable semantic
   scope, sequence, contract digest, and start time. Close it with an outcome and evidence when
   the operation settles. The coordinator is the only plan writer; workers return evidence.
3. **Build each pending task in plan order.** Dispatch to the specialist required by the task's
   routing row in a fresh context, or implement the recipe directly when the current harness has
   no specialist roster. Give dependent tasks the concrete outputs of their predecessors. Do not
   redo a task whose evidence has already been reconciled. Before repeating a meaningful
   operation, run `plan.py recovery-check --kind <kind> --scope <scope>` and reconcile any open
   or ambiguous attempt first. A legacy plan without a recovery block has unknown history and is
   initialized only after the observed workspace and delivery state are reconciled.

   Follow the task's `Test approach:` before changing the corresponding behavior. For `tdd`,
   write and run the intended check, record the observed RED failure, then implement the smallest
   change. For `characterization`, establish the green baseline before a pure refactor. For
   `not-applicable`, confirm the recorded reason and use the applicable readback, consistency or
   visual review. RED is implementation evidence, not a failed fix, review, delegation or remote
   attempt, and the normal cycle needs no approval between phases.
4. **Verify each task claim.** Run the task's `Verify:` command and compare its output with the
   expected result. Run the affected check after the implementation and after any refactor needed
   to establish GREEN; these approach-required runs are not duplicate boundary verification.
   Record the baseline, RED, GREEN, refactor result or `not-applicable` reason in the mutable
   `Test evidence` table, plus the command, result and relevant paths in the task's execution
   evidence. An unverified “it works” is not done. A task may be marked complete while changes
   remain in the working tree under `hold`; a commit is not a prerequisite for evidence.

   If implementation for a `tdd` task already exists without observed RED evidence, preserve the
   work and record the missing RED as an exception pending explicit user resolution. Do not invent
   the failure or force a revert. Until the exception is explicitly accepted, leave the task
   incomplete and do not advance the plan to `built` or `verified`, or perform delivery. An
   accepted exception still requires the actual fail-to-pass comparison, pass-to-pass
   evidence and applicable independent checks; use a safe isolated pre-change baseline when needed
   and record completion as `exception-accepted`, never strict TDD.
5. **Review at the applicable boundary.** Apply the review and in-vivo timing selected during
   reconciliation. Findings route back to the builder; the affected task or gate runs again.
   Review success proves quality for the covered scope and grants no new action permission.
6. **Respect bounded recovery.** Keep the existing limits attached to the semantic scope: one
   delegation rerun (two total dispatch attempts), three failed fixes for the same problem, two
   review correction rounds for the same review cycle (including a successful closure), and two
   failed remote operations. A successful review closes that exact review scope and cannot be
   reopened after a digest change; use a fresh semantic scope for an independent review. Successful
   remote operations may continue within their milestone while failed operations accumulate. One
   remote failure may be excluded only with a recorded `verified_transient` exception and evidence.
   An exhausted or completed scope stops and surfaces the attempts; it does not reset automatically.
7. **Close the build stage.** When every task has verified evidence, set `Status: built`. This
   state means implementation is complete for the plan's scope; it does not mean the change was
   committed, pushed, merged, deployed or otherwise delivered.

Delivery is reconciled only after the verification gate has advanced the observed plan to
`verified`, in CLOSE. A delivery grant on a plan still in `planned`, `building` or `built` never
permits a commit, push, PR, merge or deployment during Execute.

Three failed fixes on the same symptom = stop. Do not dispatch a fourth; surface the symptom,
attempts and architecture question to the user (`debugging.md`).
