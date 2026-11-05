# Branch cleanup

Read by [flow-close](../SKILL.md) at its cleanup step, and by a hygiene request. It removes merged branches and their clean worktrees, which the global guidance's «Preservation» authorizes; it grants no other effect.

## Which repositories

- For a work item: each repository where the work created a branch, a worktree or a commit, or wrote its change folder, plus the repository that holds `Specs` when it is another one.
- For a hygiene request: the current repository, the repository that holds its `Specs`, and each repository a pending-close record in persistent memory names.

Run the steps below in each of them.

## The work's branch

1. Exclude the base and environment branches, and keep a branch that an open pull request uses as its base.
2. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the local head. A branch with unmerged commits is never deleted: report it with `git log <base>..<branch>`.
3. Return the checkout to the base branch, and remove the branch's clean worktree before deleting the branch; never force the removal. Keep a worktree with uncommitted or untracked files, and its branch.
4. Delete the merged remote branch, then the local one with `git branch -d`; use `git branch -D` only for a squash- or rebase-merged branch that passed step 2.

## The sweep

After the work's branch, check the other branches of the same repository:

- each local or remote branch merged into the base whose pull request the authenticated user (`gh api user`) authored, with a `headRefOid` equal to the local head;
- in a repository without pull requests, each local branch that `git branch --merged <base>` lists and whose head differs from the commit it was created at, the last line of `git reflog show --format=%H <branch>`; a branch without that reflog line is kept;
- the temporary or preview branches the work created.

The sweep never switches a checkout's branch: this narrows, at this point, git-workflow step 2's "never switch, reset or delete its branch" to branches no checkout uses. It only removes clean worktrees and deletes branches that no remaining checkout has checked out. Apply steps 1, 2 and 4 above to each branch, and keep it, reporting the reason, when any of these holds:

- it has unmerged commits, or no commit of its own, such as a branch just created whose head is still the commit it was created at;
- a checkout other than a clean worktree this sweep removes has it checked out;
- its worktree has uncommitted or untracked files, or ignored files outside dependency and build folders such as `node_modules/`, `.venv/`, `dist/` and `build/` (`git -C <worktree> status --ignored --porcelain`), since `git worktree remove` deletes ignored files;
- a running process has its working directory inside the worktree (`lsof -d cwd -Fn` lists it), or that check cannot run, such as when `lsof` is missing;
- persistent memory records the branch or its worktree for work whose close has been pending for less than 24 hours. When the pending close is older, the sweep removes them and updates that work's records as [Pending work](../SKILL.md#pending-work) sets.

A promotion that has not been made is never a reason to keep a branch or a worktree.

## Report

List the repositories checked, the branches and worktrees removed, and each one kept with its reason.
