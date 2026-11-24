# Bounded E2E coordination

Read when the user explicitly delegates an objective through delivery, with local deliverable owners, inside the existing flows. This is a mode of those flows, not another skill entrypoint or a background scheduler. The coordinator selects and reads each required skill before entering its stage. Installing instructions does not prove native delegation, source reading or recovery works on a host.

## Bind the grant

Record the objective, acceptance criteria, delivery endpoint and mode, permitted effects, worktree and resource isolation authorization and stopping conditions in the plan, reusing answers already given. Resolve missing choices before their dependent action, continuing independent work. Under this explicit grant, continue covered stages without another human command; retain human validation when the chosen delivery requires it, and ask for changed scope or uncovered effects. E2E does not require the user's absence; when unattended authority applies, read [unattended delegation](../../unattended-delegation/SKILL.md) for its additional limits and expiry.

Before assignment, inspect the host's delegation, completion and stop capabilities: a role in a catalog is not evidence that it can run or write. If a required capability is unavailable, report the limitation and continue only work the available capabilities and grant cover; do not claim an owner or replace independent verification with self-review.

## Assign deliverables

Decompose the approved plan into units with stable IDs, task and criterion IDs, dependencies, fixed interfaces, allowed paths and required evidence. Run units concurrently only with disjoint writes and exclusive resources, and a later check across their interface. Follow [delegation](../../hive-core/delegation/delegation.md) for brief construction and reviewers. An explicitly approved local-owner unit uses `hive-build-owner` regardless of the ordinary delegation size threshold; other tasks keep that threshold and ordinary implementers.

The coordinator creates each authorized worktree and branch and assigns it with its repository, base SHA and resources; owners create none. The owner brief includes the plan, unit and assignment IDs, exact paths, effects and commands allowed, evidence destination, next checkpoint and stop conditions. The coordinator supplies needed account rows and secret sources under the global account rules, never values, and shares no writable service or credential beyond the assignment.

Use the operations of [tasks](../../flow-plan/scripts/tasks.py) for the multiunit record, as [execution record](../../flow-plan/references/execution-record.md) sets, linked from the plan's `## Checkpoint`. Keep local paths, native IDs and raw evidence in the work's ignored workspace, outside the plan file. The coordinator alone writes the record, the plan and the [decision log](../../unattended-delegation/SKILL.md#decision-log) of the work; the plan file stays in its checkout, not an owner worktree, and owner briefs give its absolute path without moving or copying it. Validate the record before dispatch and after each return with its expected generation, never overwriting a newer record to make an update pass. Record fields describe the grant and evidence, not authority or proof by themselves.

## Local owner boundary

Only an explicitly assigned `hive-build-owner` may build, test, fix and create local commits in its assigned worktree, branch and allowed paths. Use [scoped commit](../../git-workflow/scripts/git_state.py) `commit` with explicit paths and its secret scan; preserve unrelated staging. This exception changes only the no-commit and same-checkout defaults for that role under this grant. Owners also create no worktrees and integrate no other unit; the coordinator retains those operations and assigns independent verifiers.

An owner returns its exact full head, commands with exit results and evidence paths, criteria covered, limitations and next action. Before returning, it stops only processes it started and reports their cleanup, leaving work and native history intact. Partial work is a partial return, never an invented passing result.

## Check, correct and integrate

Wait for native completion or messages, not a loop that infers progress from elapsed time. On return, validate local Git identity, branch, full head, ancestry, clean state and allowed diff paths with `validate-handoff`; a rejected return stays unverified. Inspect the evidence files and actual outputs, since a structurally valid PASS field cannot establish acceptance.

Launch an independent verifier with the returned head, requirements and commands and results, following the reviewer brief. Assign it the owner's stopped, clean worktree for read-only inspection and checks at that exact head, which grants no edits, commits or worktree creation; the owner stays stopped during verification. Its verdict binds that head, criteria, command, result and available evidence, with an identity distinct from the current and historical owners. Send a demonstrated defect to the same owner when it can continue, following [defect fix](../../hive-core/build/defect-fix.md); a new head invalidates the old verdict. The stage's review and fix-round limits apply: E2E adds no retry allowance.

Only the coordinator integrates independently verified heads into the declared target, reads back the resulting head and runs the check spanning the units, recording `integration_scope` (`local` or `remote`), `integration_head` and the joint result before `integrated`. Local integration is not remote publication, which is recorded separately; this mode grants no effect beyond the recorded authorization.

## Replace and recover

At interruption, read the checkpoint and record, then reconcile native agent state, Git heads, dirty work and assigned resources. A stale ID, timeout or absent message does not prove the former writer stopped: require observed native completion or stop acknowledgment before assigning another writer. If ownership remains uncertain, keep the unit blocked and protect its resources, and continue independent units.

For replacement, preserve the unit ID, previous assignment, returned or partial head, evidence, limitations and next action, record the previous assignment's confirmed closure, then create a new assignment ID and invalidate its prior verdict. The new owner's brief carries that checkpoint. Never discard dirty work or force-reset it to fit a recorded SHA.

Before cleanup, feed all active or ownership-uncertain branches to the [branch cleanup](../../flow-close/references/branch-cleanup.md) protections: integrated status alone does not authorize deleting a dirty worktree or an unmerged branch. At the endpoint, report each criterion's observed evidence or unverified status, local integration separately from publication, and any resources or work deliberately retained.
