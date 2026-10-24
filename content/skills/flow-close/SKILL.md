---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when the user ends the session; or when the user asks for hygiene or cleanup, such as "haz higiene" or "clean up". Closes the change record, cleans merged branches, task temporaries and processes, and reports the outcome.
---

# Close a work item and the session

This skill grants no effect of its own. The global guidance authorizes the deletions it names (task temporaries, processes the task or its children started, merged work branches from this work or earlier and their clean worktrees) and the ticket state moves; everything else here follows the delivery answer already recorded for the work.

Before a report, read [reporting](references/reporting.md); before a user decision or manual validation request, read [questions](../flow-plan/references/questions.md). Before cleaning support artifacts, read [artifact layout and retention](../workspace-conventions/references/artifact-layout.md). Before resolving settings needed for closure, read [project settings](../flow-plan/references/project-settings.md).

## When to run it

1. **Work item close:** the item reached its completion point (merged into the base branch with any required post-merge CI passed, pushed to the base branch under `direct-base`, or the working tree verified under `hold`), or it is abandoned or paused.
2. **Session end:** the user ends the session.
3. **Explicit request:** the user asks for hygiene or cleanup. Run the parts of 1 and 2 that apply to the current state.

## Close a work item

Before these steps, follow [Close the current walkthrough](#close-the-current-walkthrough) when this work has one. Run these steps in order; the last one ends the turn.

1. **Tracker.** Apply the global ticket-state rule to the ticket and to each sub-issue the same work covers.
2. **Change record.** When the work has a change folder, close it as described in [Close the change record](#close-the-change-record).
3. **Git cleanup.** Run [Clean up after a merge](#clean-up-after-a-merge) when you merged. When the change record goes to the base branch through a closing pull request, do this step after that publication, not before. Update the record of each worktree you removed, merged or abandoned, in the project's persistent memory, when one is available, to say it was removed, so no later session tries to resume it; write these records, with the recurrence counts of step 5, in one memory write after step 5 and before the report, also when no worktree was removed but counts changed.
4. **Temporaries and processes.** Remove only reproducible temporary files this task created and no longer needs, and stop only the processes it or its children started, including those recorded when a child returned, as [flow-build](../flow-build/SKILL.md) sets; a recorded process is not of uncertain ownership. Keep prior material, unique evidence, and anything whose ownership or disposability is uncertain. Restore the local state the user relies on that this task changed, such as containers or services left pointing at a removed worktree, or a local development database behind migrations the work integrated: do it yourself rather than listing it for the user. First check that restoring loses none of the user's data; if it could, ask, stating what you checked.
5. **Incidental findings.** List the findings the global guidance says to propose, each with its evidence, severity, and suggested action, for the close question, and mark which ones you recommend opening and why. Ask about every listed finding in that question, splitting its options across consecutive cards when they exceed the tool's limit, as [questions](../flow-plan/references/questions.md) sets; a finding the user was not asked about stays pending, never declined. Each finding accepts three answers: open a ticket, whose body follows [ticket body](../flow-research/references/ticket-body.md); decline (saved in persistent memory with a recurrence count); or "don't mention again" (saved in persistent memory as suppressed, under the project and a stable topic, so later searches skip it). A guidance finding about the workflow itself, which has no project ticket, accepts only the last two. Save each answer once the user gives it in the close question, as a user decision. On a recurrence of a declined finding, update its count in the memory write of step 3, raise the severity when the recurrence widens the impact, and propose it again; never propose a suppressed one.
6. **Completion report.** Follow [reporting](references/reporting.md), with a cleanup line stating what was removed and what was kept, and why.
7. **Close question.** Follow [Close question](#close-question), after the report.

## Close question

**Condition:** the turn completes a work item, after its report and closing items.

**Actions:** ask the close question by reading and following [questions](../flow-plan/references/questions.md). This question replaces the Recommendations route question.

**Exceptions:** skip it when the user already answered it or explicitly asked to end, or the work was a mechanical change with no Pending item. That report still carries its cleanup label, delivery state, and reminders. Investigation-only findings use [flow-research](../flow-research/SKILL.md#deliver-the-finding)'s route question instead.

For every applicable close question, name the next ticket and the reason, such as a blocker, a defect found, or the next ticket of the same module. You choose it; an option that asks the user to name it does not count.

Offer one option per route, each continue option naming the flow it starts:

- Continue now with the recommended ticket when the loaded context helps it.
- Take that ticket to a new session when the context no longer helps.
- End the session, stating what stays open if it ends: pending items with their owner, and reminders.

**Condition:** the next work continues a plan still in progress.

**Actions:** the continue option covers all of its remaining tasks, and names in one line the effects they need that are not yet authorized, such as cloud resources or DNS, so one answer authorizes them. A production deployment keeps its own question unless the user already named it, and the build still stops for a blocking decision.

**Condition:** the next ticket reuses neither the closed item's files nor its decisions, or the session has used more than half of its context window.

**Actions:** recommend the new session, but never for a plan still in progress or after the user chose to continue it in this session; never split a flow that is still in progress.

## Close the current walkthrough

**Condition:** completing, abandoning or pausing the current work with a `walkthrough.json`, or removing its source evidence unit after closure. Hygiene without closing that work does not change its status. Do not sweep other work folders or retrofit historical manifests from the same host session.

**Actions:** before rendering, follow the [manifest delivery metadata procedure](../flow-walkthrough/references/walkthrough-manifest.md) when a PR or commit is known for this work. Before removing its worktree, manifest, page, images or recordings, resolve the installed [flow-walkthrough renderer](../flow-walkthrough/scripts/render_walkthrough.py) through the host's skill catalog or explicit path, and run `python3 <skill-dir>/scripts/render_walkthrough.py <work folder> --session-status closed --confirm-viewer`. Use only this work's folder. The operation validates and prepares the current output under the [publication target](../flow-walkthrough/references/walkthrough-manifest.md#publication-target) rule, updates status explicitly with atomic file replacement, then asks the existing viewer to confirm its current JSON and resources, with a total confirmation bound of 15 seconds. It does not delete anything. Exit `0` confirms the exact canonical source, session, key and current content hash (including referenced resources), checked before and after the authenticated query; an old closed copy is insufficient. Elapsed time alone is never confirmation. Before a later deletion, rerun this operation if the source or its resources changed.

On any validation/publication error or unconfirmed copy (missing CLI, offline viewer, unsupported runtime, absent copy, ingestion failure or timeout), preserve the still-needed source unit and any worktree containing it, and report the pending confirmation and folder. Continue independent cleanup that cannot remove it. Never start or restart a viewer or open a parallel store to force confirmation; do not relay raw CLI output or logs. A successful confirmation permits only the cleanup already authorized under [artifact retention](../workspace-conventions/references/artifact-layout.md), including its pause and user-selected retention rules. Explicitly resuming this same work follows [walkthrough delivery](../flow-walkthrough/SKILL.md#deliver) to restore `in_progress`; merely rendering an old page does not reopen it.

## Close the change record

Run this after the code is integrated and any required post-merge CI has passed. While the work runs, `tasks.md` and the status in `proposal.md` are updated on disk only; nobody versions the folder before this step.

1. For each delta file, apply it to the current-requirements home. Use the delta formats in [change records](../flow-plan/references/change-records.md#requirement-deltas) and, for a product map, [product map projects](../flow-plan/references/change-records.md#product-map-projects). In `<specs>/specs/`: apply ADDED requirements to `<specs>/specs/<capability>/spec.md`, creating it when absent; replace each MODIFIED requirement block by name; delete each REMOVED block. Edit only those blocks, leave the rest of the file untouched, and review the resulting diff.
2. Set the change's status in `proposal.md` to Completed, the final phase in [the plan format](../flow-plan/references/plan-format.md), citing the commit or pull request that integrated it.
3. Move the folder to `<specs>/changes/archive/YYYY-MM-DD-<change-id>`, using the closing date. Use `mv` when the folder was never tracked by Git, and `git mv` only when it is already versioned. Never recreate the files by rewriting them.
4. Update `<specs>/project.md` only if the phase, an open decision, or a blocker changed.
5. Version the record once, as the cases below say. Stage by named path, never a directory (see [git-workflow](../git-workflow/SKILL.md)).

Versioning cases:

- **Specs repository with `Delivery: direct-base`:** one commit and one push with the deltas applied and the folder already in `archive/`, after the code is integrated and, when it applies, the post-merge CI has passed.
- **Specs inside the code repository, or in a separate specs repository that does not declare `Delivery: direct-base`:** the commit goes to the base branch by the closing route the user chose in the delivery question, or the one a declared `Delivery` value sets: a direct push, or a small closing pull request without dedicated review, with the named merger. Act only on the recorded answer, or on the route a declared `Delivery` value sets, with you as the merger, as [delivery decisions](../flow-plan/references/delivery-decisions.md) sets; repository policy does not replace either. Order: return to the base branch and fast-forward it, make the record commit, publish it by the chosen route, then delete branches, including the closing pull request's branch. If no closing route was recorded and none follows from a declared `Delivery` value, ask for it before publishing.
- **`hold`:** the record stays unversioned and is reported as pending, like the code.
- **Pause:** do not archive. Leave the folder on disk with its current status and report it as pending.
- **Abandon:** archive without applying deltas, with the reason in `proposal.md`, and version it by the same closing route as a normal close.

## Clean up after a merge

After verifying a merge you performed, clean up its branches in this order. The global guidance grants the deletion of merged work branches, from this work or earlier, and their clean worktrees.

1. **Check eligibility.** Exclude the base and environment branches. Confirm that no open pull request uses the work branch as its base; if one does, keep the branch and report it. `gh pr merge --delete-branch` skips that check, so delete separately. Inspect the branch's local head and every worktree that checks it out. Preserve worktrees with uncommitted or untracked files, or still-needed material; if one must remain on the work branch, keep that branch too. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the branch's local head. Otherwise, a branch with unmerged commits is a decision, not noise: report its work (`git log <base>..<branch>`) and recommend integrating it or ask; never delete it.
2. **Leave the work branch.** Return the retained checkout to the base branch without overwriting changes. If the work branch is checked out in a separate eligible clean worktree, leave that directory and remove that worktree before attempting to delete the branch. Do not force worktree removal. If a checkout or retained worktree still uses the branch, keep the branch and report why.
3. **Refresh the base.** Run `git fetch --prune` and fast-forward the base. After a promotion, also fast-forward each local environment branch it moved, such as `qa`, when it is behind its remote and has no commits of its own.
4. **Delete eligible branches.** Delete the merged remote branch unless the host already removed it, then the local branch with `git branch -d`, which refuses unmerged work. Use `git branch -D` only for a squash- or rebase-merged branch whose pull request and unchanged local head passed step 1; never rewrite history to make a branch eligible.
5. **Check the remaining branches.** Apply the same checks and order to merged preview or temporary branches the work created and this repository's other local and remote branches whose pull requests are merged into the base. Report the branches checked, the worktrees and branches removed, the fast-forwards, and the branches left with their reasons.

## End the session

1. Before deleting the UI review unit, follow [Close the current walkthrough](#close-the-current-walkthrough) when applicable. Apply [artifact layout and retention](../workspace-conventions/references/artifact-layout.md) for the user's selected walkthrough output and its dependent resources, optional HTML, individual files and paused work; remove or update links to material deleted under that rule. Name the retained unit's folder in the handoff.
2. Stop the task's processes that are still running, as step 4 of the work-item close sets.
3. List the reminders the user still controls, and what stays open with its owner.
