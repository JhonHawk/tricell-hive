---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when the user ends the session; or when the user asks for hygiene or cleanup, such as "haz higiene" or "clean up". Closes the change record, cleans merged branches, task temporaries and processes, and reports the outcome.
---

# Close a work item and the session

The completion point is the merge into the base branch. When a human reviewer is assigned to the pull request, opening it is the completion point: the ticket stays in progress, and a later session that finds it merged runs this close.

1. When this work has a viewer session or a legacy walkthrough, close it first, as [flow-walkthrough](../flow-walkthrough/SKILL.md) sets for its stage, session and evidence, and read it back: the work reads closed when it was integrated or abandoned, and paused when it was paused. This step comes before any cleanup, because cleanup removes the inputs a failed close would need.
2. When the work is integrated into the base branch, move its ticket to the state the project uses for that.
3. Steps 2 and 3 are independent: run them together. Close an integrated change folder: merge each delta into `<Specs>/specs/<capability>/spec.md`, mark `proposal.md` completed, move the folder to `<Specs>/changes/archive/YYYY-MM-DD-<change-id>/`, and deliver that move as the delivery says.
4. After the viewer close in step 1, clean up as [Branch cleanup](#branch-cleanup) sets, stop the processes this work started, and remove its reproducible temporary files. Keep anything with uncommitted work or uncertain ownership.
5. Write the final report as [Final report](#final-report) sets.
6. Write the outcome to persistent memory.
7. Ask the close question as [Close question](#close-question) sets.

Close a work item once. After the close question is answered, act on the answer; do not run the close or repeat the final report again.

## Final report

Report the whole work item in one message, each part only when it has content:

- **Integrated:** what changed and where: the pull request, the commits, the ticket's state.
- **Resolved:** each acceptance criterion with how it was shown, and the corrections the user asked for.
- **Omitted:** each check that was skipped or could not run, with its reason, including the CI-only jobs.
- **Repaired:** the local environment repairs made, such as reinstalled dependencies or seeded data.
- **Findings:** the defects not fixed, as [Findings](#findings) sets.
- **Next:** what remains, with who owns it, and what comes next.

When the user asks to keep or share this report, or it exceeds what fits on one screen, offer a `flow-report` page as one option of the close question.

## Findings

A finding is a defect the work found but did not fix: a review finding below P1, a UI defect that was already there, a test that already failed or fails intermittently (keep its error output), or anything else noticed on the way. Every finding carries one priority by its observable impact, whatever scale its source used: P0 blocks the release, an operation or a core flow for every input; P1 must be fixed before merge; P2 should be fixed but does not block; P3 is minor. A finding of P2 or higher with a concrete failure scenario is proposed; the rest are listed in one line each in the final report. P2 findings that share an area or a cause are proposed as one group that lists each member with its evidence; findings of P1 or higher are proposed one by one. Never open a ticket without the user's answer.

## Close question

Ask once, as the global guidance sets, with these options:

- **End the session**: recommended when no proposed finding is P1 or higher and no plan is in progress.
- **See the findings**, with their count and highest severity in the label, such as "See the 3 findings (1 P1)", and only when there are proposed findings: recommended when one is P1 or higher.
- **Find the next ticket.**

**See the findings** shows one table, after searching the tracker for an existing ticket for each finding:

| ID | Finding | Detail and impact | Recommendation |
| --- | --- | --- | --- |
| H1 | <what, with its severity> | <where, how it fails, who it affects, the evidence> | <open a ticket, ignore, or fix now, with its reason> |

Recommend fixing now when the fix changes only files this work already changed and adds no task, dependency or effect; a ticket when it is larger or belongs to another area; ignoring when it did not recur or has no impact. After the table, ask the findings as a round, as the global guidance's «Questions and routes» section sets: each finding is one question with the options fix now, ticket and ignore, its recommendation starred; the user may answer all at once or a few at a time, such as "ok" or "H1 ticket, H2 ignore". A ticket answer opens it unless one already exists; a fix-now answer goes through `flow-build` with the code review of the fix and its checks; an ignore answer is noted in the report. When every finding has an answer, ask the close question again without this option.

**Find the next ticket** runs `flow-status` with the cut of the tickets that reuse this session's context: the files, modules and decisions it touched. Leave out the tickets another session has in progress and those opened from this work's findings, unless they are P0 or P1. Recommend the one that reuses the most context, with its task count (from its description or checklist); choosing it starts `flow-research` to assess it.

## Branch cleanup

1. Exclude the base and environment branches, and keep a branch that an open pull request uses as its base.
2. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the local head. A branch with unmerged commits is never deleted: report it with `git log <base>..<branch>`.
3. Return the checkout to the base branch, and remove the branch's clean worktree before deleting the branch; never force the removal. Keep a worktree with uncommitted or untracked files, and its branch.
4. Delete the merged remote branch, then the local one with `git branch -d`; use `git branch -D` only for a squash- or rebase-merged branch that passed step 2.
