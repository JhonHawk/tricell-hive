# Branch cleanup

Read by [flow-close](../SKILL.md) at its cleanup step, and by a hygiene request. It removes merged branches and their clean worktrees, which the global guidance's «Preservation» authorizes; it grants no other effect.

## Which repositories

- For a work item: each repository where the work created a branch, a worktree or a commit, or wrote its change folder, plus the repository that holds `Specs` when it is another one.
- For a hygiene request: the current repository, the repository that holds its `Specs`, and each repository a pending-close record in persistent memory names.

Run the steps below in each of them.

## Keep conditions

Before checking the work's own worktree, stop the processes this work started inside it, by their process ID as the global guidance's «Preservation» sets, and leave it by running the remaining commands from the main checkout. Every worktree removal below keeps the worktree, its local branch and its remote branch, skipping step 4 for them, and reports the reason, when:

- it has uncommitted or untracked files, or an ignored entry (`git -C <worktree> status --ignored --porcelain`) other than the ones below, since `git worktree remove` deletes ignored files without refusing. These ignored entries do not keep it, because removing them loses nothing:
  - a dependency or build folder, or an entry in one, such as `node_modules/`, `.venv/`, `dist/`, `build/` and `.next/`;
  - an entry whose content the main checkout of the worktree's own repository already holds at the same path, checked with `cmp -s` for a file, or for a directory with `diff -rq <main-path> <worktree-path>` printing no line other than `Only in <main-path>…`, such as a `.env` copied in when the worktree was set up. The main checkout is the path on the first `worktree ` line of `git -C <worktree> worktree list --porcelain`; a repository with no main checkout, such as a bare one, has none of these entries;
  - a tool cache or generated file from this list only, at any depth, or an entry inside one of its folders: `*.tsbuildinfo`, `next-env.d.ts`, `*.pyc`, `.turbo/`, `__pycache__/`, `.pytest_cache/`, `.mypy_cache/`, `.ruff_cache/` and `.DS_Store`.

  These exemptions cover ignored entries only: an untracked one still keeps the worktree, which is never removed with `--force`.

  For the work's own worktree, move the items flow-close step 2's answer keeps out of it first, as flow-close step 5 sets;
- a running process has its working directory inside it (`lsof -d cwd -Fn` lists it), or that check cannot run, such as when `lsof` is missing.

## The work's branch

1. Exclude the base and environment branches, and keep a branch that an open pull request uses as its base.
2. A branch merged by squash or rebase counts as merged only when `gh pr view <branch> --json state,headRefOid` reports `MERGED` and its `headRefOid` equals the head of each copy you will delete: the local branch, and the remote one (`git rev-parse origin/<branch>` after `git fetch`). A copy with commits after that head, or with unmerged commits, is never deleted: report the local one with `git log <base>..<branch>` and the remote one with `git log <headRefOid>..origin/<branch>`.
3. Return the checkout to the base branch, and remove the branch's worktree, never forcing the removal, unless a keep condition holds.
4. Delete the merged remote branch, then the local one with `git branch -d`; use `git branch -D` only for a squash- or rebase-merged branch that passed step 2.

## The sweep

After the work's branch, check these other branches of the same repository:

- each local or remote branch whose pull request the user authored and merged into the base, listed once with `gh pr list --author @me --state merged --base <base> --json headRefName,headRefOid`;
- in a repository without pull requests, each local branch that `git branch --merged <base>` lists and that persistent memory records for the user's earlier work, such as a worktree record git-workflow step 3 wrote; keep any other branch there, since nothing shows it is the user's;
- the temporary or preview branches the work created.

The sweep never switches a checkout's branch: this narrows, at this point, git-workflow step 2's "never switch, reset or delete its branch" to branches no checkout uses. It applies steps 1, 2 and 4 above to each branch and removes its worktree unless a keep condition holds. It also keeps, reporting the reason, a branch:

- checked out in a checkout other than a clean worktree this sweep removes;
- that persistent memory records for work whose close is not recorded as done, such as work in progress or a close waiting for the user's reply, unless a pending-close record for that work is older than 24 hours. For such an older pending close, do not remove its branch or worktree here: resume that work's pending close steps in order, as flow-close sets for a hygiene request, which asks its support decision before its own cleanup.

A promotion that has not been made is never a reason to keep a branch or a worktree.

## Report

List the repositories checked, the branches and worktrees removed, and each one kept with its reason.
