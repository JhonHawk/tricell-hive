# Branch cleanup

Read by [flow-close](../SKILL.md) at its cleanup step, and by a hygiene request. [Close cleanup](../../hive-core/close/close-cleanup.md) authorizes removing merged branches and their clean worktrees, and the fast-forward of the base that the helper makes under the conditions below; this procedure grants no other effect.

## Scope and inputs

For a work item, cover each repository where it created a branch, worktree or commit. For a hygiene request, cover the current repository and those named in close records that list a waiting step ([flow-close](../SKILL.md) step 7).

Handle the work's own branch first, then sweep earlier branches in the same repository. Before its own cleanup, stop only processes this work started, move the material its inventory keeps as [support material](support-material.md) sets, return its checkout to the base through [git-workflow](../../git-workflow/SKILL.md), and run from the main checkout. The sweep never switches another checkout's branch.

For each pass, refresh Git refs and collect the inputs of [cleanup_branches.py](../scripts/cleanup_branches.py) (`--help` lists its flags); the helper queries neither the tracker nor persistent memory:

- **Merged PRs.** For the work's branch, read its PR state, target base and `headRefOid`; for the sweep, list the user's merged PRs into the base with `gh pr list --author @me --state merged --base <base> --json headRefName,headRefOid`. Pass only eligible branches as `--merged-pr <branch>=<headRefOid>`.
- **Non-PR candidates.** `--candidate <branch>` only for temporary or preview branches this work created, or, in a repository without PRs, local branches whose ownership the user's earlier plan, checkpoint or close record establishes. The helper checks ancestry; the flag never proves ownership or permits unmerged commits.
- **Protected inputs.** Each open PR's base as `--open-pr-base`, and each branch named by unfinished work in memory, plans or checkpoints as `--active-branch`: fresh branches with no commits beyond base, closes with a waiting step in their close record, and every active or ownership-uncertain assignment in the [E2E execution record](../../flow-plan/references/execution-record.md). When such a close record is older than 24 hours and is not a `Close pending` one, resume its close steps, including its support-material inventory, before its own cleanup; never sweep it away. Pass environment branches not already read from `AGENTS.md` with `--env-branch`.

If an input cannot be established, keep the affected candidates and report why: an unavailable PR or memory query is not an empty result. A promotion not yet made is not a reason to keep a branch.

## Plan and apply

Run `python3 -I <flow-close-dir>/scripts/cleanup_branches.py --repo <absolute-repo> --base <base>` with those inputs, repeating each flag per value; read the returned `worktrees` and `branches`, including every `keep` reason, before repeating it with `--apply` and the same inputs.

With `--apply`, the helper first fetches the base from `origin` and fast-forwards the base of the repository's main checkout, even when `--repo` names a linked worktree, only when that checkout is on the base, its base has no commits missing from `origin`, and no incoming file is changed, untracked or ignored there; otherwise `base_update` keeps it and gives the reason and the files. Never update the base another way, such as `git pull`, which can rebase, or a merge over local changes.

The helper retains uncommitted or untracked work, unique ignored material, a busy worktree or one whose process check fails, protected or active branches, and branches still checked out outside a removable worktree. Only ignored build or dependency files, allowlisted generated caches or copies already preserved in the main checkout are disposable. Squash or rebase deletion requires the merged PR head to match each copy being removed; never force worktree removal or delete later or unmerged commits.

If the helper is unavailable, perform only the authorized cleanup whose inputs and preservation conditions you can establish manually; a missing check keeps the affected work, a helper refusal is not bypassed, and the remote branch is deleted before the local one, using safe local deletion except for a verified squash or rebase merge.

## Report

Report the repositories checked, `removed_worktrees`, `deleted_branches`, `base_update` (the commits the base advanced, or why it stayed with its files), each retained branch as ``kept `<branch>`: <reason>``, each other retained item with its reason, `stray_clones` (full clones under the worktree folders that Git does not list: each with `path`, `branch`, `dirty` and `unpushed`, or `path` and `error` for one it could not read, which the helper never removes), and any `errors`. A nonzero exit or partial result is not a completed cleanup.
