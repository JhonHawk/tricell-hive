# Bounded E2E coordination

Read when the user explicitly delegates an objective through delivery, with local deliverable owners, inside the existing flows. This is a mode of those flows, not another skill entrypoint or a background scheduler. The coordinator selects and reads each required skill before entering its stage. Installing instructions does not prove native delegation, source reading or recovery works on a host.

## Bind the grant

Record the objective, acceptance criteria, delivery endpoint and mode, permitted effects, worktree/resource isolation authorization and stopping conditions in the plan. Reuse answers already given. Resolve missing choices before their dependent action, continuing independent work. A research or planning request alone never authorizes implementation. Under this explicit grant, continue covered stages without another human command; retain human validation when the chosen delivery requires it, and ask for changed scope or uncovered effects. E2E does not require the user's absence; when unattended authority applies, read [unattended delegation](../../unattended-delegation/SKILL.md) for its additional limits and expiry.

Before assignment, inspect the host's available delegation, completion and stop capabilities. A role in a catalog is not evidence that it can run or write. If required capabilities are unavailable, report the limitation and continue only work the available capabilities and grant cover; do not silently claim an owner or replace independent verification with self-review.

## Assign deliverables

Decompose the approved plan into units with stable IDs, task and criterion IDs, dependencies, fixed interfaces, allowed paths and required evidence. Run units concurrently only with disjoint writes and exclusive resources, and a later check across their interface. Follow [delegation](../../hive-core/references/delegation.md) for brief construction and reviewers. An explicitly approved local-owner unit uses `hive-build-owner` regardless of the ordinary delegation size threshold; other tasks keep that threshold and ordinary implementers. Only the designated owner has the local-commit exception below.

The coordinator creates each authorized worktree/branch and assigns it with its repository, base SHA and resources; owners create none. The owner brief includes the plan, unit and assignment IDs, exact paths, effects and commands allowed, evidence destination, next checkpoint and stop conditions. The coordinator supplies needed account rows and secret sources under the global account rules, never values. Do not share writable services or credentials beyond the assignment.

Read [execution record](../../flow-plan/references/execution-record.md) and use the operations of [tasks](../../flow-plan/scripts/tasks.py) for the multiunit record, linked from `tasks.md`. Keep local paths, native IDs and raw evidence in the work's ignored workspace, not its versioned plan. The coordinator alone writes the record and plan, and the [decision log](../../unattended-delegation/SKILL.md#decision-log) of the work, whose format that section sets; the change folder stays in its checkout, not an owner worktree. Owner briefs point to that plan without moving or copying it. Validate it before dispatch and after each return; use its expected generation, never overwrite a newer record to make an update pass. Record fields describe the grant and evidence, not authority or proof by themselves.

## Local owner boundary

Only an explicitly assigned `hive-build-owner` may build, test, fix and create local commits in its assigned worktree, branch and allowed paths. Use [scoped commit](../../git-workflow/scripts/git_state.py) `commit` with explicit paths and its secret scan; preserve unrelated staging. This exception changes only the no-commit and same-checkout defaults for that role under this grant. The verifier checkout exception below grants no owner authority. Other child boundaries remain unchanged. Owners never publish, push, open or merge PRs, write trackers, create worktrees, launch children or integrate other units. The coordinator retains those authorized operations and assigns independent verifiers.

An owner returns its exact full head, commands with exit results and evidence paths, criteria covered, limitations and next action. Before returning, stop only processes it started and report their cleanup; leave work and native history intact. Partial work is a partial return, never an invented passing result.

## Check, correct and integrate

Wait for native completion or messages, not a loop that infers progress from elapsed time. On return, validate local Git identity, branch, full head, ancestry, clean state and allowed diff paths with `validate-handoff`. A rejected return stays unverified. Inspect the evidence files and actual outputs; a structurally valid PASS field cannot establish acceptance.

Launch an independent verifier with the returned head, requirements and commands/results, following the reviewer brief. Assign it the owner's stopped, clean worktree for read-only inspection and checks at that exact head; this changes only the ordinary same-checkout default, never grants edits, commits or worktree creation, and the owner stays stopped during verification. Its verdict binds that head, criteria, command, result and available evidence, with identity distinct from the current and historical owners. Send a demonstrated defect to the same owner when it can continue; follow [defect fix](../../hive-core/references/defect-fix.md). A new head invalidates the old verdict: obtain evidence for the new head before marking verified. Apply the existing review/fix-round limits of the stage; E2E adds no retry allowance.

Only the coordinator integrates independently verified heads into the declared target, reads back the resulting head and runs the check spanning the units. Record `integration_scope` (`local` or `remote`), `integration_head` and the joint result before `integrated`. Local integration is not remote publication. Continue the chosen verification, review and delivery procedures; this mode grants no effect beyond the recorded authorization. Record publication separately from local integration.

## Replace and recover

At interruption, read the checkpoint and record, then reconcile native agent state, Git heads, dirty work and assigned resources. A stale ID, timeout or absent message does not prove the former writer stopped. Require observed native completion or stop acknowledgment before assigning another writer. If ownership remains uncertain, keep the unit blocked and protect its resources; continue independent units.

For replacement, preserve the unit ID, previous assignment, returned or partial head, evidence, limitations and next action. Record the previous assignment's confirmed closure, then create a new assignment ID and invalidate its prior verdict. The new owner's brief carries that checkpoint so the human need not reconstruct the session. Never discard dirty work or force-reset it to fit a recorded SHA.

Before cleanup, feed all active or ownership-uncertain branches to the existing [branch cleanup](../../flow-close/references/branch-cleanup.md) protections. Integrated status alone does not authorize deleting a dirty worktree or an unmerged branch. At the endpoint, report each criterion's observed evidence or unverified status, local integration separately from publication, and any resources or work deliberately retained. Native history and unrelated sessions remain preserved.
