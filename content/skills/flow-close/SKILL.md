---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when the user ends the session; or when the user asks for hygiene or cleanup, such as "haz higiene" or "clean up". Closes the change record, cleans merged branches, task temporaries and processes, and reports the outcome.
---

# Close a work item and the session

The completion point is the merge into the base branch. When a human reviewer is assigned to the pull request, opening it is the completion point: the ticket stays in progress, and a later session that finds it merged runs this close.

1. When the work is integrated into the base branch, move its ticket to the state the project uses for that.
2. Close an integrated change folder: merge each delta into `<Specs>/specs/<capability>/spec.md`, mark `proposal.md` completed, move the folder to `<Specs>/changes/archive/YYYY-MM-DD-<change-id>/`, and deliver that move as the delivery says.
3. When the work folder holds a `walkthrough.json`, mark it closed first with `python3 <flow-walkthrough dir>/scripts/render_walkthrough.py <work folder> --session-status closed --confirm-viewer`; when it exits 2, keep that folder. Clean up as [Branch cleanup](#branch-cleanup) sets, stop the processes this work started, and remove its reproducible temporary files. Keep anything with uncommitted work or uncertain ownership.
4. Write the final report as [Final report](#final-report) sets.
5. Write the outcome to persistent memory.
6. Ask the close question as [Close question](#close-question) sets.

## Final report

Report the whole work item in one message, each part only when it has content:

- **Integrated:** what changed and where: the pull request, the commits, the ticket's state.
- **Resolved:** each acceptance criterion with how it was shown, and the corrections the user asked for.
- **Omitted:** each check that was skipped or could not run, with its reason, including the CI-only jobs.
- **Repaired:** the local environment repairs made, such as reinstalled dependencies or seeded data.
- **Findings:** the defects not fixed, as [Findings](#findings) sets.
- **Pending:** what remains, with who owns it.

## Findings

A finding is a defect the work found but did not fix: a review finding below P1, a UI defect that was already there, a test that already failed or fails intermittently (keep its error output), or anything else noticed on the way. A finding of P2 or higher with a concrete failure scenario is proposed; the rest are listed in one line each in the final report. Never open a ticket without the user's answer.

## Close question

Ask once, as [hive-core](../hive-core/SKILL.md) sets, with these options:

- **End the session**: recommended when no proposed finding is P1 or higher and no plan is in progress.
- **See the findings**, with their count and highest severity in the label, such as "See the 3 findings (1 P1)", and only when there are proposed findings: recommended when one is P1 or higher.
- **Find the next ticket.**

**See the findings** shows one table, after searching the tracker for an existing ticket for each finding:

| ID | Finding | Detail and impact | Recommendation |
| --- | --- | --- | --- |
| H1 | <what, with its severity> | <where, how it fails, who it affects, the evidence> | <open a ticket, ignore, or fix now, with its reason> |

Recommend fixing now when the fix is small, in the area the work touched and has no new effect; a ticket when it is larger or belongs to another area; ignoring when it did not recur or has no impact. The user answers in free text by ID, such as "H1 ticket, H2 ignore", all at once or a few at a time. A ticket answer opens it unless one already exists; a fix-now answer goes through `flow-build` with the code review of the fix and its checks; an ignore answer is noted in the report. When every finding has an answer, ask the close question again without this option.

**Find the next ticket** runs `flow-status` with the cut of the tickets that reuse this session's context: the files, modules and decisions it touched. Leave out the tickets another session has in progress and those opened from this work's findings, unless they are P0 or P1. Recommend the one that reuses the most context, with its estimated size; choosing it starts `flow-research` to assess it.

## Branch cleanup

1. Exclude the base and environment branches, and keep a branch that an open pull request uses as its base.
2. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the local head. A branch with unmerged commits is never deleted: report it with `git log <base>..<branch>`.
3. Return the checkout to the base branch, and remove the branch's clean worktree before deleting the branch; never force the removal. Keep a worktree with uncommitted or untracked files, and its branch.
4. Delete the merged remote branch, then the local one with `git branch -d`; use `git branch -D` only for a squash- or rebase-merged branch that passed step 2.
