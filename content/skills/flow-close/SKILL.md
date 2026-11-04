---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when an operational task ends; when the user ends the session; or when the user asks for hygiene or cleanup, such as "haz higiene" or "clean up". Closes the change record, cleans merged branches, task temporaries and processes, proposes what to do with the work's support material, and reports the outcome, ending with a close footer whose reply starts the cleanup.
---

# Close a work item and the session

The completion point is the merge into the base branch. When the delivery is `ready` or a human reviewer is assigned to the pull request, the open pull request is the completion point, reached when `ready` stops or, with a human reviewer, after [git-workflow step 7](../git-workflow/SKILL.md)'s read: the ticket stays in progress, and a later session that finds it merged runs a follow-up close: step 1's viewer close, step 2's support decision, the ticket move, the change-folder close, the cleanup, the memory write and a final report of those, without the close footer.

The close also runs when an operational task ends, as the global guidance's «Proportionality» defines it, and when the user abandons the work; a step with nothing to act on, such as the ticket move of work that was not integrated, is skipped. Work delivered under `hold` gets no full close until it is delivered later, at that delivery's completion point; until then, when its verified working tree is reported, it closes as a pause: step 1 pauses its viewer session, step 2 lists its support material without asking or acting on it, step 5 stops the processes and containers it started and keeps everything else, step 6 reports what is pending and step 7 records it, with no footer.

1. When this work has a viewer session, close it first, as [flow-walkthrough](../flow-walkthrough/SKILL.md) sets for its stage, session and evidence, and read it back: the work reads closed when it was integrated or abandoned, `in_review` when its open pull request waits for the user's merge or a human reviewer, and paused when it was paused. Closing the session changes only its record: never stop the viewer service or close the viewer the user has open, even when this work started it. This step comes before any cleanup, because cleanup removes the inputs a failed close would need. When the viewer is not running, do not start it: keep the inputs its close needs, continue the rest of the close, and name in the final report the viewer session left open with where those inputs are.
2. Take the inventory of this work's support material and ask its decision as [Support material](#support-material) sets, before steps 4–5: step 4 delivers what the answer commits, and step 5 removes the worktree and checkpoint the inventory reads.
3. When the work is integrated into the base branch, move its ticket to the state the project uses for that. At the completion point, commit its change folder in the specs repository when [git-workflow step 4](../git-workflow/SKILL.md) leaves that commit to the close.
4. Close an integrated change folder: merge each delta into `<Specs>/specs/<capability>/spec.md`, mark `proposal.md` completed, move the folder to `<Specs>/changes/archive/YYYY-MM-DD-<change-id>/` with the edits made to it since its last commit, taken from the checkout or worktree that holds them, then restore that copy with `git restore` once the archive commit holds those edits, and deliver that move, with the items step 2's answer commits, archives or promotes, as the delivery says.
5. Clean up as [Branch cleanup](#branch-cleanup) sets, stop the processes and containers this work started, never a program other sessions may use, as the global guidance's «Preservation» sets, and remove its reproducible temporary files and checkpoint file, the containers it started with their volumes and the images it built, as the global guidance's «Preservation» authorizes; keep the checkpoint while step 2 lists items as pending. Before removing a worktree, move the items step 2's answer keeps out of it to the same paths in the main checkout, since removing a worktree deletes the ignored files in it. List a dump or export it made and ask before removing it. List the remote resources it created, such as storage objects, registry images or task definitions, each with the command that removes it, and ask before removing them: each removal is a critical effect. Keep anything with uncommitted work or uncertain ownership.
6. Write the final report as [Final report](#final-report) sets.
7. Write the outcome to persistent memory, the account index rows as the global guidance's «Test accounts» sets, and the pending-work records as [Pending work](#pending-work) sets.
8. End with the close footer as [Close footer](#close-footer) sets.

At the completion point, and when an operational task ends, first run steps 1, 3, 6 and 7, record steps 2, 4 and 5 as pending, and end with step 8. Retain step 1’s readback and receipts and record their location for later input cleanup; do not close an already closed viewer again. Run the pending steps in order when the user chooses to close, starts another work item or ends the session, report what they did and any viewer limitation, and update memory. When another work starts, step 2’s question joins its first round; steps 4 and 5 wait for the answer. If the session stops without a reply, a later close or hygiene request resumes the pending steps from memory.

These closes run every step at once and end on the report, without the footer: a follow-up close, an abandoned work item, a deploy's cleanup close, a hygiene request, the session's end, and any close under [unattended delegation](../unattended-delegation/SKILL.md). The pause close of `hold` work runs as its paragraph above sets.

Close a work item once; the follow-up close of a pull request that waited at its completion point, and the full close of `hold` work after its pause close, are part of that close, not a repeat. After the footer's reply, act on it; do not run the close or repeat the final report again.

## Support material

List what this work created or changed under the project's support home, in its checkout or worktree: its folders under `sessions/`, `workspace/` and `evidence/`, found by the work's identifier as [artifact layout](../workspace-conventions/references/artifact-layout.md) names it, its change folder when the work was abandoned, and any other support path its checkpoint or its commits name. Its reproducible temporary files live in the host's temporary location, not here, and step 5 removes them. Name material of other work or uncertain ownership, such as an untracked folder no record ties to this work, as kept, without an action. Propose one action for each item, with its reason:

- **Commit:** a curated, shareable record under `sessions/` or `evidence/` that is not yet in Git; never a secret, a raw dump or reproducible output.
- **Archive:** a `sessions/` folder whose records all read closed and that holds nothing **Commit** excludes, moved as [archive](../workspace-conventions/references/archive.md) sets; the user's answer is its selection, and the commit below delivers the move, which that procedure leaves to its caller.
- **Promote:** a conclusion that changes a durable document under `docs/`, as [workspace conventions](../workspace-conventions/SKILL.md) «Promote a finding» sets.
- **Delete:** an item under `workspace/` with positive evidence that it is reproducible or that a kept copy replaces it; never unique evidence, a dump or export, or uncommitted work.
- **Keep:** anything else, with where it stays; an item in a worktree stays at the same path in the main checkout.

When items have an action other than keep, ask them as one decision, in its own round, as the global guidance's «Questions and routes» sets: the starred option applies every proposed action, another keeps everything as it is, and a reply that names an item's path changes its action. Nothing is moved, committed or deleted before the answer; apply it before steps 4–5 continue. Commit the committed, archived and promoted items with step 4's move or, when no change folder is archived, in one commit that touches only the support home; either goes on a new branch from the base as [git-workflow](../git-workflow/SKILL.md) sets, never on the base unless the delivery is `direct-base`, and is delivered as the work's delivery mode sets, before step 5 cleans up its branch with the work's. Work with no delivery mode, such as an operational task, proposes no commit, archive or promotion. Work that is paused, held under `hold`, or waiting at an open pull request lists its items as pending, asks nothing and acts on none; the full or follow-up close asks the decision.

## Pending work

A work item leaves something open on purpose when a task of its plan runs after the merge, such as a measurement over later CI runs, or a later phase of the same work waits. A promotion along the declared environments is not one, since the ticket's state records it; neither are this close's own deferred steps, which step 7 records apart, nor anything of an operational task.

At step 7, including a pause close, in this order:

1. Search persistent memory for `Pending:` records that name this work's ticket or `<change-id>`, in the project «Persistent memory» names, with an explicit result limit. Remove from each record the items this work resolved, and soft-delete a record left with none, except this work's own record when step 2 rewrites it, with `engram delete <id>` using the ID the search returned, never `--hard`. An abandoned work soft-deletes its own record. Without Engram, remove the entry from the host's memory, or say in the final report that it could not be removed.
2. When this work leaves something open on purpose, write its own record under the topic `pending/<change-id>`, titled `Pending: <change-id>`, replacing any earlier content, using the change folder's `<change-id>`, or for a plan without one the ticket key or a short slug, as [flow-plan](../flow-plan/SKILL.md#change-folder) names it. For each open item, state what remains, why, who owns it, and the tickets or conditions that unblock it. Comment the same items on the ticket, when there is one.

## Final report

Report the whole work item in one message, each part only when it has content:

- **Integrated:** what changed and where: the pull request, the commits, the ticket's state; for a pull request left open, its head commit, the base commit its current-base check passed on, and the CI result.
- **Resolved:** each acceptance criterion with how it was shown, and the corrections the user asked for.
- **Omitted:** each check that was skipped or could not run, with its reason, including the CI-only jobs.
- **Repaired:** the local environment repairs made, such as reinstalled dependencies or seeded data.
- **Support material:** each item with its path and the action applied, or listed as pending, as [Support material](#support-material) sets.
- **Findings:** the defects not fixed, as [Findings](#findings) sets.
- **Next:** what remains, with who owns it, and what comes next, such as the merge left to the user and the review bots still pending.

When the user asks to keep or share this report, or it exceeds what fits on one screen, offer a `flow-report` page as one reply of the close footer.

## Findings

A finding is a defect the work found but did not fix: a review finding below P1, a UI defect that was already there, a test that already failed or fails intermittently (keep its error output), or anything else noticed on the way. Every finding carries one priority by its observable impact, whatever scale its source used: P0 blocks the release, an operation or a core flow for every input; P1 must be fixed before merge; P2 should be fixed but does not block; P3 is minor. A finding of P2 or higher with a concrete failure scenario is proposed; the rest are listed in one line each in the final report. P2 findings that share an area or a cause are proposed as one group that lists each member with its evidence; findings of P1 or higher are proposed one by one. Never open a ticket without the user's answer.

## Close footer

End the final report with the next-step block of the global guidance's «Questions and routes», never with the question tool. Its heading recommends **close** with its reason, unless a proposed finding is P1 or higher, when it recommends seeing the findings, or a plan is in progress, when it recommends continuing it. Its replies come from this work's context, each only when it applies:

- **Close:** runs the remaining steps above.
- **Resolve findings now**, naming them, such as "resolve H2 and H3 in this branch": they go through `flow-build` on the work's branch, with the code review of the fix and its checks.
- **See the findings**, with their count and highest severity, such as "see the 3 findings (1 P1)".
- **Explain a finding or a result**, naming it.
- **Find the next ticket**, as **Find the next ticket** below sets, run only when the user picks it.
- **Continue** with what remains of the work, naming it, such as the next phase of a plan in progress.

**See the findings** shows one table, after searching the tracker for an existing ticket for each finding:

| ID | Finding | Detail and impact | Recommendation |
| --- | --- | --- | --- |
| H1 | <what, with its severity> | <where, how it fails, who it affects, the evidence> | <open a ticket, ignore, or fix now, with its reason> |

Recommend fixing now or a ticket as the global guidance's «Scope changes» ticket rule sets, which also governs the duplicate search and how a new ticket relates to its origin; recommend ignoring when the finding did not recur or has no impact. After the table, ask the findings as a round, as the global guidance's «Questions and routes» section sets: each finding is one question with the options fix now, ticket and ignore, its recommendation starred; the user may answer all at once or a few at a time, such as "ok" or "H1 ticket, H2 ignore". A ticket answer opens it unless one already exists; a fix-now answer goes through `flow-build` with the code review of the fix and its checks; an ignore answer is noted in the report. When every finding has an answer, end again with the close footer, without the findings replies.

**Find the next ticket** runs `flow-status` with the cut of the tickets that reuse this session's context: the files, modules and decisions it touched. Leave out the tickets another session has in progress and those opened from this work's findings, unless they are P0 or P1. Recommend the one that reuses the most context, with its task count (from its description or checklist); choosing it starts `flow-research` to assess it. That report's next-step block carries the compaction reply for another work item that the global guidance's «Questions and routes» section sets.

## Branch cleanup

1. Exclude the base and environment branches, and keep a branch that an open pull request uses as its base.
2. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the local head. A branch with unmerged commits is never deleted: report it with `git log <base>..<branch>`.
3. Return the checkout to the base branch, and remove the branch's clean worktree before deleting the branch; never force the removal. Keep a worktree with uncommitted or untracked files, and its branch.
4. Delete the merged remote branch, then the local one with `git branch -d`; use `git branch -D` only for a squash- or rebase-merged branch that passed step 2.
