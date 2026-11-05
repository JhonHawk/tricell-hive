# Branch cleanup

Read by [flow-close](../SKILL.md) at its cleanup step, and by a hygiene request. It removes merged branches and their clean worktrees, which the global guidance's «Preservation» authorizes; it grants no other effect.

## Which repositories

- For a work item: each repository where the work created a branch, a worktree or a commit, or wrote its change folder, plus the repository that holds `Specs` when it is another one.
- For a hygiene request: the current repository, the repository that holds its `Specs`, and each repository a pending-close record in persistent memory names.

Run the steps below in each of them.

## Keep conditions

Every worktree removal below, the work's own included, keeps the worktree and its branch, and reports the reason, when:

- it has uncommitted or untracked files, or ignored files outside dependency and build folders such as `node_modules/`, `.venv/`, `dist/` and `build/` (`git -C <worktree> status --ignored --porcelain`), since `git worktree remove` deletes ignored files without refusing; move the items step 2's answer keeps out of it first, as flow-close step 5 sets;
- a running process has its working directory inside it (`lsof -d cwd -Fn` lists it), or that check cannot run, such as when `lsof` is missing.

## The work's branch

1. Exclude the base and environment branches, and keep a branch that an open pull request uses as its base.
2. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the head of each copy you will delete: the local branch, and the remote one (`git rev-parse origin/<branch>` after `git fetch`). A copy with commits after that head, or with unmerged commits, is never deleted: report it with `git log <base>..<branch>`.
3. Return the checkout to the base branch, and remove the branch's worktree, never forcing the removal, unless a keep condition holds.
4. Delete the merged remote branch, then the local one with `git branch -d`; use `git branch -D` only for a squash- or rebase-merged branch that passed step 2.

## The sweep

After the work's branch, check these other branches of the same repository:

- each local or remote branch whose pull request the authenticated user (`gh api user`) authored and that `gh pr view` reports `MERGED`;
- in a repository without pull requests, each local branch that `git branch --merged <base>` lists and that persistent memory records for the user's earlier work, such as a worktree record git-workflow step 3 wrote; keep any other branch there, since nothing shows it is the user's;
- the temporary or preview branches the work created.

The sweep never switches a checkout's branch: this narrows, at this point, git-workflow step 2's "never switch, reset or delete its branch" to branches no checkout uses. It applies steps 1, 2 and 4 above to each branch and removes its worktree unless a keep condition holds. It also keeps, reporting the reason, a branch:

- checked out in a checkout other than a clean worktree this sweep removes;
- that persistent memory records for work whose close is not recorded as done, such as work in progress or a close waiting for the user's reply, unless that close has been pending for more than 24 hours. For such an older pending close, do not remove its branch or worktree here: resume that work's pending close steps in order, as flow-close sets for a hygiene request, which asks its support decision before its own cleanup.

A promotion that has not been made is never a reason to keep a branch or a worktree.

## Report

List the repositories checked, the branches and worktrees removed, and each one kept with its reason.
