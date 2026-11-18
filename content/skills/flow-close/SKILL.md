---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when an operational task ends; when the user ends the session; or when the user asks for session hygiene, such as "haz higiene"; not to clean up, prune or reorganize existing docs or support material, which docs-cleanup does. Closes the change record, cleans merged branches, task temporaries and processes, proposes what to do with the work's support material, and reports the outcome with a close footer whose reply starts any cleanup left.
---

# Close a work item and the session

For explicit E2E work, read [coordination](../flow-build/references/e2e-coordination.md) before closing units or cleaning their resources.

The completion point is the merge into the base branch. When the delivery is `ready`, or a human reviewer's approval is pending, the open pull request is the completion point, reached at the stop [git-workflow step 8](../git-workflow/SKILL.md) sets: the ticket stays in progress, and a later session that finds it merged runs a follow-up close.

To clean up, prune, reorganize, curate or archive existing docs or support material, or to apply a layout migration, use [docs-cleanup](../docs-cleanup/SKILL.md) instead of these steps.

The close also runs when an operational task ends, as the global guidance's «Proportionality» defines it, and when the user abandons the work; a step with nothing to act on, such as the ticket move of work that was not integrated, is skipped. Work delivered under `hold` gets no full close until it is delivered later, at that delivery's completion point; until then, when its verified working tree is reported, it closes as a pause: step 1 pauses its viewer session, step 2 lists its support material without asking or acting on it, step 5 stops the processes and containers it started and keeps everything else, step 6 reports what is pending and step 7 records it, with no footer. Work paused outside `hold`, such as a plan paused before its build, closes the same way, and also commits its change folder as [git-workflow step 4](../git-workflow/SKILL.md) sets for a pause.

1. When this work has viewer captures, close it first, as [publication](../flow-verify/references/viewer-publication.md#close-and-input-hygiene) sets, and read it back: the work reads closed when it was integrated or abandoned, `in_review` when its open pull request waits for the user's merge or a human reviewer, and paused when it was paused. Closing it changes only its record: never stop the viewer service or close the viewer the user has open, even when this work started it. This step comes before any cleanup, because cleanup removes the inputs a failed close would need. When the viewer is not running, do not start it: keep the inputs its close needs, continue the rest of the close, and name in the report the viewer session left open with where those inputs are.
2. Take the inventory of this work's support material and ask its decision as [support material](references/support-material.md) sets, before steps 4–5: step 4 delivers what the answer commits, and step 5 removes the worktree and checkpoint the inventory reads.
3. When the work is integrated into the base branch, move its ticket to the state the project uses for that, in its tracker as the global guidance's «Project settings» sets. At the completion point, commit its change folder in the specs repository when [git-workflow step 4](../git-workflow/SKILL.md) leaves that commit to the close. Close its business rules as [business rules](../flow-plan/references/business-rules.md) sets.
4. Close an integrated change folder: run `python3 <flow-plan-dir>/scripts/check_change.py <folder> --archive` ([check](../flow-plan/scripts/check_change.py)) and fix each coherence line; for unchecked boxes, ask as a decision, with their count, whether to archive anyway; a no or an unattended close leaves it pending. Otherwise, on a new branch from the base, copy in the folder's uncommitted edits from the checkout or worktree holding them, run `python3 <flow-close-dir>/scripts/apply_deltas.py <folder> --specs <Specs> --detail "<PR, merge SHA>"` ([merge](scripts/apply_deltas.py)), fix each refusal, report each dropped scenario, then add `--apply`: it merges the deltas, marks the proposal completed and archives the folder. Once the archive commit holds those edits, `git restore` the folder in that source checkout; deliver the move with the items step 2's answer commits, archives or promotes, as the delivery says.
5. Clean up the branches and worktrees of each repository the work touched as [branch cleanup](references/branch-cleanup.md) sets (using [cleanup_branches.py](scripts/cleanup_branches.py)), and stop and remove only what the global guidance's «Preservation» authorizes at close; keep the checkpoint while step 2 lists items as pending. Before removing a worktree, move the items step 2's answer keeps out of it to the same paths in the main checkout, since removing a worktree deletes its ignored files. List any dump or export made and ask before removing it. List remote resources created (storage objects, registry images, task definitions), each with its removal command, and ask before removing them: each is a critical effect. Keep anything with uncommitted work or uncertain ownership.
6. Write the final report as [Final report](#final-report) sets.
7. Write the outcome to persistent memory, the account index rows as the global guidance's «Test accounts» sets, the pending-work records as [Pending work](#pending-work) sets, and each candidate's record as [Candidates](#candidates) sets.
8. End with the close footer as [Close footer](#close-footer) sets.

At the completion point, and when an operational task ends, first run steps 1, 3, 6 and 7, record steps 2, 4 and 5 as pending, and end with step 8; when step 2's inventory lists no item to act on, step 4 has no change folder and step 5 has nothing of this work to remove or ask about, run them at once instead, before step 6. Retain step 1’s readback and receipts and record their location for later input cleanup; do not close an already closed viewer again. Run the pending steps in order when the user chooses to close, starts another work item or ends the session, report what they did and any viewer limitation, and update memory. When another work starts, step 2’s question joins its first round; steps 4 and 5 wait for the answer. If the session stops without a reply, a later close or hygiene request resumes the pending steps from memory.

These closes defer no step, step 2's question included, and end on the report, without the footer: a follow-up close, an abandoned work item, a deploy's cleanup close, a hygiene request, the session's end, and any close under [unattended delegation](../unattended-delegation/SKILL.md). The pause close of `hold` work runs as its paragraph above sets.

Close a work item once; the follow-up close of a pull request that waited at its completion point, and the full close of `hold` work after its pause close, are part of that close, not a repeat. After the footer's reply, act on it; do not run the close or its final report again.

## Pending work

A work item leaves something open on purpose when a task of its plan runs after the merge, such as a measurement over later CI runs, or a later phase of the same work waits. A promotion along the declared environments is not one, since the ticket's state records it, unless a plan task promotes to an environment where the ticket or the request asks for the change; a promotion not yet made never keeps a branch, worktree, process or support item. Neither are this close's own deferred steps, which step 7 records apart, nor anything of an operational task.

At step 7, including a pause close, in this order:

1. Search persistent memory for `Pending:` records that name this work's ticket or `<change-id>`, in the project «Persistent memory» names, with an explicit result limit. Remove from each record the items this work resolved, and soft-delete a record left with none, except this work's own record when step 2 rewrites it, with `engram delete <id>` using the ID the search returned, never `--hard`. An abandoned work soft-deletes its own record. Without Engram, remove the entry from the host's memory, or say in the report that it could not be removed.
2. When this work leaves something open on purpose, write its own record under the topic `pending/<change-id>`, titled `Pending: <change-id>`, replacing any earlier content, using the change folder's `<change-id>`, or for a plan without one the ticket key or a short slug, as [flow-plan](../flow-plan/SKILL.md#change-folder) names it. For each open item, state what remains, why, who owns it, and the tickets or conditions that unblock it. Comment the same items on the ticket, when there is one.

## Final report

Report the whole work item in one message, each part only when it has content:

- **Integrated:** what changed and where: the pull request, the commits, the ticket's state; for a pull request left open, its head commit, the base commit its current-base check passed on, and the CI result.
- **Resolved:** each acceptance criterion with how it was shown, and the corrections the user asked for.
- **Omitted:** each check that was skipped or could not run, with its reason, including the CI-only jobs.
- **Repaired:** the local environment repairs made, such as reinstalled dependencies or seeded data.
- **Support material:** each item with its path and the action applied, or listed as pending, as [support material](references/support-material.md) sets.
- **Findings:** the defects not fixed, as [Findings](#findings) sets.
- **Candidates:** the work's repetitions, as [Candidates](#candidates) sets.
- **Next:** what remains, with who owns it, and what comes next, such as the merge left to the user and review bots or base CI still pending.

When the user asks to keep or share this report, or it exceeds what fits on one screen, offer a `flow-report` page as one reply of the close footer.

## Findings

A finding is a defect the work found but did not fix: a review finding below P1, a UI defect that was already there, a test that already failed or fails intermittently (keep its error output), or anything else noticed on the way. Every finding carries one priority by its observable impact, whatever scale its source used: P0 blocks the release, an operation or a core flow for every input; P1 is blocking, as [flow-verify step 3](../flow-verify/SKILL.md) defines it, and must be fixed before merge; P2 should be fixed but does not block; P3 is minor. A finding of P2 or higher with a concrete failure scenario is proposed; the rest get one line each in the report. P2 findings that share an area or a cause are proposed as one group that lists each member with its evidence; findings of P1 or higher are proposed one by one. Never open a ticket without the user's answer, which unattended delegation gives at its step 5.

## Candidates

A candidate is a repetition the work did not turn into a mechanism under [defect fix](../flow-build/references/defect-fix.md) point 5: an inline script or command chain the work ran two or more times for the same purpose, such as the same link check or hash comparison, or a correction the user gave twice in the work. The report lists each one on one line, with what repeated, how many times and the mechanism that point would choose, and asks nothing about it; a work with none has no such part.

At step 7, search the project's persistent memory for each candidate by the terms of its purpose, then create or update one record under the topic `candidate/<tool>-<target>` in kebab case, such as `candidate/link-check-docs`, with the sessions that saw it, their dates and its count. A candidate seen in two sessions or more is offered in the close footer; it is not a finding, since it is no defect.

## Close footer

End the final report with the next-step block of the global guidance's «Questions and routes», never with the question tool. Its heading recommends **close** with its reason, unless a proposed finding is P1 or higher, when it recommends seeing the findings, or a plan is in progress, when it recommends continuing it; with no step pending and neither case, its heading is **▶ Closed**, with no recommendation. Its replies come from this work's context, each only when it applies:

- **Close:** runs the pending steps above; only while one is pending.
- **Resolve findings now**, naming them, such as "resolve H2 and H3 in this branch": they enter `flow-build` on the work's branch, with the code review of the fix and its checks, or `flow-plan` when the fix is not a small understood change.
- **See the findings**, with their count and highest severity, such as "see the 3 findings (1 P1)".
- **Explain a finding or a result**, naming it.
- **Turn a candidate into a mechanism**, naming it: only for a candidate whose record counts two sessions or more; it enters `flow-build` as **Resolve findings now** does.
- **Find the next ticket**, as **Find the next ticket** below sets, run only when the user picks it.
- **Continue** with what remains of the work, naming it, such as the next phase of a plan in progress.

**See the findings** shows one table, after checking each finding for an existing ticket as [create-issue](../create-issue/SKILL.md) step 2 sets:

| ID | Finding | Detail and impact | Recommendation |
| --- | --- | --- | --- |
| H1 | <what, with its severity> | <where, how it fails, who it affects, the evidence> | <open a ticket, ignore, or fix now, with its reason> |

Recommend fixing now or a ticket as the global guidance's «Scope changes» ticket rule sets; recommend ignoring when the finding did not recur or has no impact. After the table, ask the findings as a round, as the global guidance's «Questions and routes» sets: each finding is one question with the options fix now, ticket and ignore, its recommendation starred; the user may answer all at once or a few at a time, such as "ok" or "H1 ticket, H2 ignore". A ticket answer creates it as that skill sets; a fix-now answer goes as **Resolve findings now** sets; an ignore answer is noted in the report. When every finding has an answer, end again with the close footer, without the findings replies.

**Find the next ticket** runs research's [status report](../flow-research/references/status.md#find-the-next-ticket) as it sets for that reply.
