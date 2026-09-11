# flow-build — Execute (`planned`/`building` → `built`)

Builds the portable plan's pending tasks. Reached when reconciliation has confirmed explicit
implementation authority and the state is `planned` or `building`; the `verify` subcommand skips
this stage.

1. **Open the authorized workspace.** Confirm the contract digest, implementation scope, session
   Git mode, required repositories and any task prerequisites. If the current tree contains a
   change outside the authorized scope, pause and surface it before writing.
2. **Mark execution.** Set the execution state to `building` and record the run start without
   touching the frozen contract or its digest. The execution table is the only progress record;
   the ordinary contract steps remain unchanged.
3. **Build each pending task in plan order.** Dispatch to the specialist required by the task's
   routing row in a fresh context, or implement the recipe directly when the current harness has
   no specialist roster. Give dependent tasks the concrete outputs of their predecessors. Do not
   redo a task whose evidence has already been reconciled.
4. **Verify each task claim.** Run the task's `Verify:` command and compare its output with the
   expected result. An unverified “it works” is not done. Record the command, result and relevant
   paths in the task's execution evidence. A task may be marked complete while changes remain in
   the working tree under `hold`; a commit is not a prerequisite for evidence.
5. **Review at the applicable boundary.** Apply the review and in-vivo timing selected during
   reconciliation. Findings route back to the builder; the affected task or gate runs again.
   Review success proves quality for the covered scope and grants no new action permission.
6. **Close the build stage.** When every task has verified evidence, set `Status: built`. This
   state means implementation is complete for the plan's scope; it does not mean the change was
   committed, pushed, merged, deployed or otherwise delivered.

Delivery is reconciled only after the verification gate has advanced the observed plan to
`verified`, in CLOSE. A delivery grant on a plan still in `planned`, `building` or `built` never
permits a commit, push, PR, merge or deployment during Execute.

Three failed fixes on the same symptom = stop. Do not dispatch a fourth; surface the symptom,
attempts and architecture question to the user (`debugging.md`).
