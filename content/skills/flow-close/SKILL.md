---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when the user ends the session; or when the user asks for hygiene or cleanup, such as "haz higiene" or "clean up". Closes the change record, cleans merged branches, task temporaries and processes, and reports the outcome.
---

# Close a work item and the session

This skill grants no effect of its own. The global guidance authorizes the deletions it names (task temporaries, processes the task or its children started, merged work branches from this work or the user's earlier work and their clean worktrees) and the ticket state moves; everything else here follows the delivery answer already recorded for the work.

Before a report, read [reporting](references/reporting.md); before a user decision or manual validation request, read [questions](../hive-core/references/questions.md). Before cleaning support artifacts, read [artifact layout and retention](../workspace-conventions/references/artifact-layout.md). Before resolving settings needed for closure, read [project settings](../hive-core/references/project-settings.md).

## Entry and exit

Enter at the completion point (integration by [flow-verify](../flow-verify/SKILL.md), or flow-build when verify is skipped; a later completion point the repository's `Completion` declares follows [project settings](../hive-core/references/project-settings.md#later-completion-point)), on abandonment, a pause, a hygiene request or the session's end; any of these enters directly. [flow-deploy](../flow-deploy/SKILL.md) is a separate stage after integration; after a promotion, only the cleanup below runs. Exit with the record closed, the cleanup done and the report.

## When to run it

1. **Work item close:** the item reached its completion point (merged into the base branch, by a flow or at the user's request outside one, with any required post-merge CI passed; pushed to the base branch under `direct-base`; or the working tree verified under `hold`), or it is abandoned or paused. Under `hold`, the ticket stays in progress and the delivery the user chose not to make is a reminder in the report, not a pending item. When the delivery hands the open pull request to a human reviewer, opening it is the completion point: the ticket stays in progress until the merge, and the report gives the merge as a reminder; trigger 5 closes it later. Inside an iterative route that [flow-research](../flow-research/SKILL.md) defines, a verified round is not the completion point: report it in one line and continue; the route's stop is, also when nothing was integrated, and then the ticket moves only if the work was integrated. A validation stop or a question that waits is not a pause: it does not run this skill. Classify a requirement that arrives after integration as [scope changes](../hive-core/references/scope-changes.md) sets.
2. **Session end:** the user ends the session. A work item still open closes as a pause.
3. **Explicit request:** the user asks for hygiene or cleanup. Run the parts of 1 and 2 that apply to the current state.
4. **After a promotion:** run only [Clean up after a merge](#clean-up-after-a-merge) and the Temporaries step of the work-item close, without a completion report or close question.
5. **Later trigger:** a session that finds merged the pull request of an item closed under human review runs the work-item close for it; one that observes the condition of the final `Completion` move applies the move as [project settings](../hive-core/references/project-settings.md#later-completion-point) sets, and only that.

## Close a work item

Before these steps, follow [Close the current walkthrough](#close-the-current-walkthrough) when this work has one. Run these steps in order; the last one ends the turn. A mechanical change with no change folder and no finding to list runs only Tracker, Git cleanup, Temporaries and processes, and Completion report.

1. **Tracker.** When the work was integrated, confirm the move the global ticket-state rule made then, or apply it now, for the ticket and each sub-issue the same work covers; under `hold`, a human-review route or a pause, leave it in progress; when abandoned, report the move back as a reminder.
2. **Change record.** When the work has a change folder, handle it by its state as [Close the change record](#close-the-change-record) sets.
3. **Git cleanup.** Run [Clean up after a merge](#clean-up-after-a-merge) when you merged. When the change record goes to the base branch through a closing pull request, do this step after that publication, not before.
4. **Temporaries and processes.** Remove only reproducible temporary files this task created and no longer needs, and stop only the processes it or its children started, including those recorded when a child returned, as [check returned work](../flow-research/references/delegation.md#check-returned-work) sets; a recorded process is not of uncertain ownership. Keep prior material, unique evidence, and anything whose ownership or disposability is uncertain. Restore the local state the user relies on that this task changed, such as containers or services left pointing at a removed worktree, or a local development database behind migrations the work integrated: do it yourself rather than listing it for the user. First check that restoring loses none of the user's data; if it could, ask, stating what you checked.
5. **Incidental findings.** Before listing, asking about or recording incidental findings, read and follow [incidental findings](references/incidental-findings.md); it holds how they are listed, asked about, answered and recorded.
6. **Memory.** In one memory write, after the Incidental findings step and before the report, record that each worktree you removed, merged or abandoned was removed, so no later session resumes it, and the recurrence counts of declined findings; also when no worktree was removed but counts changed. The answers to the findings question are saved later, when the user gives them, as [incidental findings](references/incidental-findings.md) sets.
7. **Completion report.** Follow [reporting](references/reporting.md), with a cleanup line stating what was removed and what was kept, and why.
8. **Close question.** Follow [Close question](#close-question), after the report. Under unattended delegation, do not ask it: the report names the recommended next ticket and what stays open, and the delegation ends there.

## Close question

**Condition:** the turn completes a work item, after its report and closing items.

**Actions:** ask the close question by reading and following [questions](../hive-core/references/questions.md). This question replaces the Recommendations route question. Classify an answer that adds to the work as [scope changes](../hive-core/references/scope-changes.md) sets.

**Exceptions:** skip it when the user already answered it or explicitly asked to end, or the work was a mechanical change with no Pending item and no open topic without a destination. That report still carries its cleanup label, delivery state, and reminders. Investigation-only findings use [flow-research](../flow-research/SKILL.md#deliver-the-finding)'s route question instead.

For every applicable close question, name the next ticket and the reason, such as a blocker, a P0 or P1 defect found, or the next ticket of the same module. You choose it; an option that asks the user to name it does not count. Do not recommend as the next ticket a ticket opened from this work item's findings, during the build or this close, or a finding proposed in this close, unless it is P0 or P1 or blocks the next ticket or a pending item of the work.

Offer one option per route, each continue option naming the flow it starts and the recommended ticket's estimated size, which the size condition of [questions](../hive-core/references/questions.md) sets, given once for the question instead of on every option; when the next work continues a plan in progress, that size is its remaining tasks with their time range:

- Continue now with the recommended ticket when the loaded context helps it.
- Take that ticket to a new session when the context no longer helps.
- End the session, stating what stays open if it ends: pending items with their owner, and reminders.

When the round also asks about incidental findings, [incidental findings](references/incidental-findings.md) sets their option to resolve them in this work and when a route that depends on those answers waits for a later turn.

**Condition:** the next work continues a plan still in progress.

**Actions:** the continue option covers all of its remaining tasks, and names in one line the effects they need that are not yet authorized, grouped as [effect authorization](../hive-core/references/effect-authorization.md) sets, so one answer authorizes them. A critical effect, such as changing DNS or creating a resource that spends money, keeps its own question as that reference sets, and a production deployment keeps its own question unless the user already named it, and the build still stops for a blocking decision.

**Condition:** the next ticket reuses neither the closed item's files nor its decisions, or the session has used more than half of its context window.

**Actions:** recommend the new session, but never for a plan still in progress or after the user chose to continue it in this session; never split a flow that is still in progress.

## Close the current walkthrough

When completing, abandoning or pausing the current work with a `walkthrough.json`, or removing its source evidence unit after closure, read and follow [walkthrough close](references/walkthrough-close.md).

## Close the change record

While the work runs, `tasks.md` and the status in `proposal.md` are updated on disk only; nobody versions the folder before this step. Select the case before applying any of the integrated-change steps below:

- **Pause:** do not apply deltas or archive. Leave the folder on disk with its current status and report it as pending.
- **Abandon:** do not apply deltas. Archive the folder with its delta folder renamed with an `unapplied-` prefix (`unapplied-specs/`, `unapplied-product/`) so no check reads it as applied, record the reason in `proposal.md`, and version it by the chosen closing route; under `hold` leave it unversioned.
- **`hold`:** do not apply deltas or archive before integration. Keep the record unversioned and report it as a reminder, like the code.
- **Human-review route:** do not apply deltas or archive before the reviewer merges and required post-merge CI passes. Until then, report the record as a reminder with the open pull request.
- **Integrated:** after the code is integrated and any required post-merge CI has passed, apply the following steps in order.

1. For each delta file, apply it to the current-requirements home. Use the delta formats in [change records](../flow-plan/references/change-records.md#requirement-deltas) and, for a product map, [product map projects](../flow-plan/references/change-records.md#product-map-projects). In `<specs>/specs/`: apply ADDED requirements to `<specs>/specs/<capability>/spec.md`, creating it when absent, with the delta's preamble as its title and Purpose; replace each MODIFIED requirement block by name; delete each REMOVED block. Edit only those blocks, leave the rest of the file untouched, and review the resulting diff. Then check every delta: each ADDED or MODIFIED requirement appears in the spec with the same text, apart from whitespace, and no `### Requirement:` heading of a REMOVED one remains; on a product-map page, the new text of each ADDED or MODIFIED rule, without its marker or quoted current rule, is present and each REMOVED rule is gone. Fix the spec where it does not match.
2. Set the change's status in `proposal.md` to Completed only after that check passes, the final phase in [the plan format](../flow-plan/references/plan-format.md), citing the commit or pull request that integrated it.
3. Move the folder to `<specs>/changes/archive/YYYY-MM-DD-<change-id>`, using the closing date. Use `mv` when the folder was never tracked by Git, and `git mv` only when it is already versioned. Never recreate the files by rewriting them.
4. Update `<specs>/project.md` only if the phase, an open decision, or a blocker changed.
5. Version the record once, as the cases below say. Stage by named path, never a directory (see [git-workflow](../git-workflow/SKILL.md)).

Versioning cases:

- **Specs repository with `Delivery: direct-base`:** one commit and one push with the deltas applied and the folder already in `archive/`, after the code is integrated and, when it applies, the post-merge CI has passed.
- **Specs inside the code repository, or in a separate specs repository that does not declare `Delivery: direct-base`:** the commit goes to the base branch by the closing route the user chose in the delivery question: a direct push, or a small closing pull request without dedicated review, with the named merger. Act only on the recorded answer, as [delivery decisions](../flow-plan/references/delivery-decisions.md) sets; repository policy does not replace it. Order: return to the base branch and fast-forward it, make the record commit, publish it by the chosen route, then clean up as [Clean up after a merge](#clean-up-after-a-merge) sets, including the closing pull request's branch. If no closing route was recorded, ask for it before publishing.

## Clean up after a merge

After verifying a merge you performed, or after a promotion, read and follow [branch cleanup](references/branch-cleanup.md) before deleting any branch or worktree.

## End the session

1. Before deleting the UI review unit, follow [Close the current walkthrough](#close-the-current-walkthrough) when applicable. Apply [artifact layout and retention](../workspace-conventions/references/artifact-layout.md) for the user's selected walkthrough output and its dependent resources, optional HTML, individual files and paused work; remove or update links to material deleted under that rule. Name the retained unit's folder in the handoff.
2. Stop the task's processes that are still running, as the Temporaries and processes step of the work-item close sets.
3. List the reminders the user still controls, and what stays open with its owner.
