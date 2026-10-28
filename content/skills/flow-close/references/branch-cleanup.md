# Clean up after a merge

## When to read

Read after verifying a merge you performed, or after a promotion, before deleting any branch or worktree. Reading this reference does not invoke the closing workflow.

## Clean up the branches

After verifying a merge you performed, clean up its branches in this order. The global guidance grants the deletion of merged work branches, from this work or the user's earlier work, and their clean worktrees.

1. **Check eligibility.** Exclude the base and environment branches. Confirm that no open pull request uses the work branch as its base; if one does, keep the branch and report it. `gh pr merge --delete-branch` skips that check, so delete separately. Inspect the branch's local head and every worktree that checks it out. Preserve worktrees with uncommitted or untracked files, or still-needed material; if one must remain on the work branch, keep that branch too. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the branch's local head. Delete a remote branch only when its remote head equals that pull request's `headRefOid` or is an ancestor of the base, so commits pushed after the merge are not lost. Otherwise, a branch with unmerged commits is a decision, not noise: report its work (`git log <base>..<branch>`) and recommend integrating it or ask; never delete it.
2. **Leave the work branch.** Return the retained checkout to the base branch without overwriting changes. If the work branch is checked out in a separate eligible clean worktree, leave that directory and remove that worktree before attempting to delete the branch. Do not force worktree removal. If a checkout or retained worktree still uses the branch, keep the branch and report why.
3. **Refresh the base.** Run `git fetch --prune` and fast-forward the base. After a promotion, also fast-forward each local environment branch it moved, such as `qa`, when it is behind its remote and has no commits of its own.
4. **Delete eligible branches.** Delete the merged remote branch unless the host already removed it, then the local branch with `git branch -d`, which refuses unmerged work. Use `git branch -D` only for a squash- or rebase-merged branch whose pull request and unchanged local head passed step 1; never rewrite history to make a branch eligible.
5. **Check the remaining branches.** Apply the same checks and order to merged preview or temporary branches the work created and this repository's other local and remote branches whose pull requests the authenticated user authored (`gh api user`) and are merged into the base. Report the branches checked, the worktrees and branches removed, the fast-forwards, and the branches left with their reasons.
