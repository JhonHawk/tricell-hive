---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when the user ends the session; or when the user asks for hygiene or cleanup, such as "haz higiene" or "clean up". Closes the change record, cleans merged branches, task temporaries and processes, and reports the outcome.
---

# Close a work item and the session

This skill grants no effect of its own. The global guidance authorizes the deletions it names (task temporaries, task-started processes, merged work branches and their clean worktrees) and the ticket state moves; everything else here follows the delivery answer already recorded for the work.

## When to run it

1. **Work item close:** the item reached its completion point (merged into the base branch with any required post-merge CI passed, pushed to the base branch under `direct-base`, or the working tree verified under `hold`), or it is abandoned or paused.
2. **Session end:** the user ends the session.
3. **Explicit request:** the user asks for hygiene or cleanup. Run the parts of 1 and 2 that apply to the current state.

## Close a work item

Run these steps in order; the last one ends the turn.

1. **Tracker.** Apply the global ticket-state rule to the ticket and to each sub-issue the same work covers.
2. **Change record.** When the work has a change folder, close it as described in [Close the change record](#close-the-change-record).
3. **Git cleanup.** Run [Clean up after a merge](#clean-up-after-a-merge) when you merged. When the change record goes to the base branch through a closing pull request, do this step after that publication, not before.
4. **Temporaries and processes.** Remove only reproducible temporary files this task created and no longer needs, and stop only the processes it started. Keep prior material, unique evidence, and anything whose ownership or disposability is uncertain.
5. **Incidental findings.** List the findings the global guidance says to propose, each with its evidence, severity, and suggested action, for the single close question. Each finding accepts three answers: open a ticket; decline (saved in persistent memory with a recurrence count); or "don't mention again" (saved in persistent memory as suppressed, under the project and a stable topic, so later searches skip it). A guidance finding about the workflow itself, which has no project ticket, accepts only the last two. On a recurrence of a declined finding, update its count, raise the severity when the recurrence widens the impact, and propose it again; never propose a suppressed one.
6. **Completion report.** Write it as the global guidance describes, with a cleanup line stating what was removed and what was kept, and why.
7. **Close question.** Ask the global close question, unless the user already answered it or explicitly asked to end, or the work was a mechanical change with no Pending item; that report still carries its cleanup label, delivery state, and reminders. Offer one option per route, each continue option naming the flow it starts, since this question replaces the Recommendations route question: continue now with the recommended ticket when the loaded context helps it; take that ticket to a new session when the context no longer helps; or end the session, stating what stays open if it ends: pending items with their owner, and reminders. Recommend the new session when the next ticket reuses neither the closed item's files nor its decisions, or when the session has used more than half of its context window; never split a flow that is still in progress. Name the ticket and the reason, such as a blocker, a defect found, or the next ticket of the same module; you choose it, and an option that asks the user to name it does not count.

## Close the change record

Run this after the code is integrated and any required post-merge CI has passed. While the work runs, `tasks.md` and the status in `proposal.md` are updated on disk only; nobody versions the folder before this step.

1. For each delta file, apply it to the current-requirements home. Use the delta formats in [change records](../flow-plan/references/change-records.md#requirement-deltas) and, for a product map, [product map projects](../flow-plan/references/change-records.md#product-map-projects). In `<specs>/specs/`: apply ADDED requirements to `<specs>/specs/<capability>/spec.md`, creating it when absent; replace each MODIFIED requirement block by name; delete each REMOVED block. Edit only those blocks, leave the rest of the file untouched, and review the resulting diff.
2. Set the change's status to closed in `proposal.md`, citing the commit or pull request that integrated it.
3. Move the folder to `<specs>/changes/archive/YYYY-MM-DD-<change-id>`, using the closing date. Use `mv` when the folder was never tracked by Git, and `git mv` only when it is already versioned. Never recreate the files by rewriting them.
4. Update `<specs>/project.md` only if the phase, an open decision, or a blocker changed.
5. Version the record once, as the cases below say. Stage by named path, never a directory (see [git-workflow](../git-workflow/SKILL.md)).

Versioning cases:

- **Specs repository with `Delivery: direct-base`:** one commit and one push with the deltas applied and the folder already in `archive/`, after the code is integrated and, when it applies, the post-merge CI has passed.
- **Specs inside the code repository, or in a separate specs repository that does not declare `Delivery: direct-base`:** the commit goes to the base branch by the closing route the user chose in the delivery question: a direct push, or a small closing pull request without dedicated review, with the named merger. Act only on that recorded answer; repository policy does not replace it. Order: return to the base branch and fast-forward it, make the record commit, publish it by the chosen route, then delete branches, including the closing pull request's branch. If no closing route was recorded, ask for it before publishing.
- **`hold`:** the record stays unversioned and is reported as pending, like the code.
- **Pause:** do not archive. Leave the folder on disk with its current status and report it as pending.
- **Abandon:** archive without applying deltas, with the reason in `proposal.md`, and version it by the same closing route as a normal close.

## Clean up after a merge

After verifying a merge you performed, clean up its branches. The global guidance grants the deletion of merged work branches and their clean worktrees. First confirm that no open pull request uses the work branch as its base; if one does, keep the branch and report it. `gh pr merge --delete-branch` skips that check, so delete separately. Delete the merged remote branch unless the host already removed it, then the local branch with `git branch -d`, which refuses unmerged work. Return the checkout to the base branch, run `git fetch --prune`, and fast-forward the base. After a promotion, also fast-forward each local environment branch it moved, such as `qa`, when it is behind its remote and has no commits of its own. Delete merged preview or temporary branches that the work created. Apply the same checks to this repository's other local and remote branches whose pull requests are merged into the base. A branch with unmerged commits is a decision, not noise: report its work (`git log <base>..<branch>`) and recommend integrating it or ask; never delete it. Report the branches checked, the deletions, the fast-forwards, and the unmerged branches left.

## End the session

1. Delete the UI review images this task created that the user did not explicitly select for retention, and remove or update their report links. The placement rule is in the global support-folder guidance.
2. Stop any task-started process that is still running.
3. List the reminders the user still controls, and what stays open with its owner.
