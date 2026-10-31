---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when the user ends the session; or when the user asks for hygiene or cleanup, such as "haz higiene" or "clean up". Closes the change record, cleans merged branches, task temporaries and processes, and reports the outcome.
---

# Close a work item and the session

The completion point is the merge into the base branch. When a human reviewer is assigned to the pull request, opening it is the completion point: the ticket stays in progress, and a later session that finds it merged runs this close.

1. When the work is integrated into the base branch, move its ticket to the state the project uses for that.
2. Close an integrated change folder: merge each delta into `<Specs>/specs/<capability>/spec.md`, mark `proposal.md` completed, move the folder to `<Specs>/changes/archive/YYYY-MM-DD-<change-id>/`, and deliver that move as the delivery says.
3. When the work folder holds a `walkthrough.json`, mark it closed first with `python3 <flow-walkthrough dir>/scripts/render_walkthrough.py <work folder> --session-status closed --confirm-viewer`; when it exits 2, keep that folder. Clean up as [Branch cleanup](#branch-cleanup) sets, stop the processes this work started, and remove its reproducible temporary files. Keep anything with uncommitted work or uncertain ownership.
4. Report what was integrated and where, the ticket's state, and anything still pending with who owns it.
5. Write the outcome to persistent memory.
6. Ask the user whether to continue with other work or end the session.

## Branch cleanup

1. Exclude the base and environment branches, and keep a branch that an open pull request uses as its base.
2. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the local head. A branch with unmerged commits is never deleted: report it with `git log <base>..<branch>`.
3. Return the checkout to the base branch, and remove the branch's clean worktree before deleting the branch; never force the removal. Keep a worktree with uncommitted or untracked files, and its branch.
4. Delete the merged remote branch, then the local one with `git branch -d`; use `git branch -D` only for a squash- or rebase-merged branch that passed step 2.
