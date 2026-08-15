# flow-build — Execute (`planned`/`building` → `built`)

Builds the plan's pending tasks. Reached when the reconciled state is `planned` or
`building`; the `verify` subcommand skips this stage.

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
4. **Per-task gate** — the two-stage review (`references/verify-gate.md`) on the task diff; if
   in-vivo timing is **inline** and the task is `in-vivo: yes`, its walk runs now. Findings route
   back to the builder; the affected stage re-runs.
5. **Commit, PR, CI, merge** — one commit per verified task with the `Commit: feat(<scope>):
   T<n> …` tag (the tag is how state is read from git — never batch tasks into one commit). A
   plan may declare **change-groups** (cohesive runs of small tasks): one branch/PR per group,
   tasks landing as sequential tagged commits on it, gates amortizing per `testing.md > Execution
   Scope`; undeclared → one PR per task. Push, open the PR, and **overlap the CI wait**
   (`gh pr checks <n> --watch`) with the task's non-integrative ceremony (ledger notes, evidence
   filing, next dispatch prep) — never merge red or pending; the session owns the wait, it is
   never handed to the user. The merge also waits for the PR's bug-hunt pass
   (`git-workflow.md > PRs & promotion`, Phase B — the repo's reviewer app or the harness-native
   review; the verify-gate review already discharged Phase A). CI failure → route to the builder, fix
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
