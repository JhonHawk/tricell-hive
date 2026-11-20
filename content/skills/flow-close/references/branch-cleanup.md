# Branch cleanup

Read by [flow-close](../SKILL.md) at its cleanup step, and by a hygiene request. The global guidance’s «Preservation» authorizes removing merged branches and their clean worktrees, and the fast-forward of the base that the helper makes under the conditions below; this procedure grants no other effect.

## Scope and inputs

For a work item, cover each repository where it created a branch, worktree or commit, or wrote its change folder, plus the repository holding `Specs`. For a hygiene request, cover the current repository, its specs repository and those named in pending-close records.

Handle the work’s own branch first, then sweep earlier branches in the same repository. Before its own cleanup, stop only processes this work started, move out the material its support decision keeps, return its checkout to the base through [git-workflow](../../git-workflow/SKILL.md), and run from the main checkout. The sweep never switches another checkout’s branch.

For each pass, refresh Git refs and collect the following inputs for [cleanup_branches.py](../scripts/cleanup_branches.py); the helper does not query the tracker or persistent memory:

- Merged PRs: for the work’s branch, read its PR state, target base and `headRefOid`; for the sweep, list the user’s merged PRs into the base with `gh pr list --author @me --state merged --base <base> --json headRefName,headRefOid`. Supply only eligible branches as `--merged-pr <branch>=<headRefOid>`.
- Non-PR candidates: use `--candidate <branch>` only for temporary/preview branches this work created, or, in a repository without PRs, local branches whose ownership the user’s earlier plan, checkpoint or close record establishes. The helper checks ancestry; this flag never proves ownership or permits unmerged commits.
- Protected inputs: pass each open PR’s base as `--open-pr-base <branch>` and each branch named by unfinished work in memory, plans or checkpoints as `--active-branch <branch>`. This includes fresh branches with no commits beyond base, pending closes, and every active or ownership-uncertain assignment in the [E2E execution record](../../flow-plan/references/execution-record.md). When a pending close is older than 24 hours, resume its close steps, including its support decision, before its own cleanup; never sweep it away. Pass environment branches not already read from `AGENTS.md` with `--env-branch <branch>`.

If an input cannot be established, keep the affected candidates and report why; an unavailable PR or memory query is not an empty result. A promotion not yet made is not a reason to keep a branch.

## Plan and apply

Run `python3 -I <flow-close-dir>/scripts/cleanup_branches.py --repo <absolute-repo> --base <base>` with the inputs above, repeating each flag for each value. With no eligible candidates, there is nothing to apply. Read the returned `worktrees` and `branches`, including every `keep` reason, before repeating the invocation with `--apply` and the same inputs.

With `--apply`, the helper first fetches the base from `origin` and fast-forwards the base of the repository's main checkout, even when `--repo` names a linked worktree, only when that checkout is on the base, its base has no commits missing from `origin`, and no incoming file is changed, untracked or ignored there; otherwise `base_update` keeps it and gives the reason and the files. Never update the base another way, such as `git pull`, which can rebase, or a merge over local changes.

The helper owns the mechanical checks and removal order. It must retain uncommitted or untracked work, unique ignored material, a busy worktree or one whose process check fails, protected/active branches, and branches still checked out outside a removable worktree. Only ignored build/dependency files, allowlisted generated caches or copies already preserved in the main checkout are disposable; untracked files never qualify. Squash/rebase deletion requires the merged PR head to match each copy being removed; never force worktree removal or delete later/unmerged commits.

If the helper is unavailable, perform only the authorized cleanup whose inputs and preservation conditions above you can establish manually; a missing check keeps the affected work. Do not bypass a helper refusal. Preserve the remote-before-local branch deletion order, using safe local deletion except for a verified squash/rebase merge.

## Report

Report the repositories checked, `removed_worktrees`, `deleted_branches`, `base_update` (the commits the base advanced, or why it stayed with its files), each retained item with its reason, and any `errors`. A nonzero exit or partial result is not a completed cleanup.
