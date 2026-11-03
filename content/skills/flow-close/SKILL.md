---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when an operational task ends; when the user ends the session; or when the user asks for hygiene or cleanup, such as "haz higiene" or "clean up". Closes the change record, cleans merged branches, task temporaries and processes, proposes what to do with the work's support material, and reports the outcome.
---

# Close a work item and the session

The completion point is the merge into the base branch. When the delivery is `ready` or a human reviewer is assigned to the pull request, the open pull request is the completion point, reached when `ready` stops or, with a human reviewer, after [git-workflow step 7](../git-workflow/SKILL.md)'s read: the ticket stays in progress, and a later session that finds it merged runs a follow-up close: step 1's viewer close, the ticket move, the change-folder close, the cleanup, the memory write and a final report of those, without the close question; it asks only the decision of [Support material](#support-material) when that lists items to act on.

The close also runs when an operational task ends, as the global guidance's «Proportionality» defines it, and when the user abandons the work; a step with nothing to act on, such as the ticket move of work that was not integrated, is skipped. Work delivered under `hold` gets no full close until it is delivered later, at that delivery's completion point; until then, when its verified working tree is reported, it closes as a pause: step 1 pauses its viewer session, step 4 stops the processes and containers it started and keeps everything else, and steps 5–7 report what is pending, its support material included, listed without acting on it.

1. When this work has a viewer session, close it first, as [flow-walkthrough](../flow-walkthrough/SKILL.md) sets for its stage, session and evidence, and read it back: the work reads closed when it was integrated or abandoned, `in_review` when its open pull request waits for the user's merge or a human reviewer, and paused when it was paused. This step comes before any cleanup, because cleanup removes the inputs a failed close would need. When the viewer is not running, do not start it: keep the inputs its close needs, continue the rest of the close, and name in the final report the viewer session left open with where those inputs are.
2. When the work is integrated into the base branch, move its ticket to the state the project uses for that.
3. Steps 2 and 3 are independent: run them together. Close an integrated change folder: merge each delta into `<Specs>/specs/<capability>/spec.md`, mark `proposal.md` completed, move the folder to `<Specs>/changes/archive/YYYY-MM-DD-<change-id>/`, and deliver that move as the delivery says; when step 4's inventory proposes a commit, an archive or a promotion, deliver the move together with the answered items, after the close question's round.
4. After step 1, clean up as [Branch cleanup](#branch-cleanup) sets, stop the processes and containers this work started, and remove its reproducible temporary files and checkpoint file, the containers it started with their volumes and the images it built, as the global guidance's «Preservation» authorizes; list a dump or export it made and ask before removing it. List the remote resources it created, such as storage objects, registry images or task definitions, each with the command that removes it, and ask before removing them: each removal is a critical effect. Keep anything with uncommitted work or uncertain ownership. Then take the inventory of this work's support material as [Support material](#support-material) sets.
5. Write the final report as [Final report](#final-report) sets.
6. Write the outcome to persistent memory.
7. Ask the close question as [Close question](#close-question) sets.

Close a work item once; the follow-up close of a pull request that waited at its completion point, and the full close of `hold` work after its pause close, are part of that close, not a repeat. After the close question is answered, apply the support material answer first, then act on the close answer; do not run the close or repeat the final report again.

## Support material

List what this work created or changed under the project's support home: its folders under `sessions/`, `workspace/` and `evidence/`, found by the work's identifier as [artifact layout](../workspace-conventions/references/artifact-layout.md) names it, its change folder when step 3 did not archive it, and any other support path its checkpoint or `git status` shows it wrote. Leave out what step 4 removed; name material of other work or uncertain ownership as kept, without an action. Propose one action for each item, with its reason:

- **Commit:** a curated, shareable record under `sessions/` or `evidence/` that is not yet in Git; never a secret, a raw dump or reproducible output.
- **Archive:** a `sessions/` folder whose records all read closed, moved as [archive](../workspace-conventions/references/archive.md) sets; the user's answer is its selection.
- **Promote:** a conclusion that changes a durable document under `docs/`, as [workspace conventions](../workspace-conventions/SKILL.md) «Promote a finding» sets.
- **Delete:** an item under `workspace/` with positive evidence that it is reproducible or that a kept copy replaces it; never unique evidence, a dump or export, or uncommitted work.
- **Keep:** anything else, with where it stays.

List the items with their actions in the final report, and ask them as one decision of the close question's round, as the global guidance's «Questions and routes» sets: the starred option applies every proposed action, another keeps everything as it is, and a reply that names an item's path changes its action. Nothing is moved, committed or deleted before the answer. Commit the committed, archived and promoted items with step 3's move or, when no change folder is archived, in one commit that touches only the support home, delivered as the work's delivery mode sets. Work with no delivery mode, such as an operational task, proposes no commit or archive and keeps those items. Work that is paused, held under `hold`, or waiting at an open pull request lists its items as pending and acts on none; the full or follow-up close asks the decision.

## Final report

Report the whole work item in one message, each part only when it has content:

- **Integrated:** what changed and where: the pull request, the commits, the ticket's state; for a pull request left open, its head commit, the base commit its current-base check passed on, and the CI result.
- **Resolved:** each acceptance criterion with how it was shown, and the corrections the user asked for.
- **Omitted:** each check that was skipped or could not run, with its reason, including the CI-only jobs.
- **Repaired:** the local environment repairs made, such as reinstalled dependencies or seeded data.
- **Support material:** each item with its path and proposed action, as [Support material](#support-material) sets.
- **Findings:** the defects not fixed, as [Findings](#findings) sets.
- **Next:** what remains, with who owns it, and what comes next, such as the merge left to the user and the review bots still pending.

When the user asks to keep or share this report, or it exceeds what fits on one screen, offer a `flow-report` page as one option of the close question.

## Findings

A finding is a defect the work found but did not fix: a review finding below P1, a UI defect that was already there, a test that already failed or fails intermittently (keep its error output), or anything else noticed on the way. Every finding carries one priority by its observable impact, whatever scale its source used: P0 blocks the release, an operation or a core flow for every input; P1 must be fixed before merge; P2 should be fixed but does not block; P3 is minor. A finding of P2 or higher with a concrete failure scenario is proposed; the rest are listed in one line each in the final report. P2 findings that share an area or a cause are proposed as one group that lists each member with its evidence; findings of P1 or higher are proposed one by one. Never open a ticket without the user's answer.

## Close question

Ask once, as the global guidance sets, in the same round as the decision of [Support material](#support-material) when it lists items to act on, with these options:

- **End the session**: recommended when no proposed finding is P1 or higher and no plan is in progress.
- **See the findings**, with their count and highest severity in the label, such as "See the 3 findings (1 P1)", and only when there are proposed findings: recommended when one is P1 or higher.
- **Find the next ticket.**

**See the findings** shows one table, after searching the tracker for an existing ticket for each finding:

| ID | Finding | Detail and impact | Recommendation |
| --- | --- | --- | --- |
| H1 | <what, with its severity> | <where, how it fails, who it affects, the evidence> | <open a ticket, ignore, or fix now, with its reason> |

Recommend fixing now or a ticket as the global guidance's «Scope changes» ticket rule sets, which also governs the duplicate search and how a new ticket relates to its origin; recommend ignoring when the finding did not recur or has no impact. After the table, ask the findings as a round, as the global guidance's «Questions and routes» section sets: each finding is one question with the options fix now, ticket and ignore, its recommendation starred; the user may answer all at once or a few at a time, such as "ok" or "H1 ticket, H2 ignore". A ticket answer opens it unless one already exists; a fix-now answer goes through `flow-build` with the code review of the fix and its checks; an ignore answer is noted in the report. When every finding has an answer, ask the close question again without this option.

**Find the next ticket** runs `flow-status` with the cut of the tickets that reuse this session's context: the files, modules and decisions it touched. Leave out the tickets another session has in progress and those opened from this work's findings, unless they are P0 or P1. Recommend the one that reuses the most context, with its task count (from its description or checklist); choosing it starts `flow-research` to assess it. That report's next-step block carries the compaction reply for another work item that the global guidance's «Questions and routes» section sets.

## Branch cleanup

1. Exclude the base and environment branches, and keep a branch that an open pull request uses as its base.
2. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the local head. A branch with unmerged commits is never deleted: report it with `git log <base>..<branch>`.
3. Return the checkout to the base branch, and remove the branch's clean worktree before deleting the branch; never force the removal. Keep a worktree with uncommitted or untracked files, and its branch.
4. Delete the merged remote branch, then the local one with `git branch -d`; use `git branch -D` only for a squash- or rebase-merged branch that passed step 2.
