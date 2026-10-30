---
name: "hive-tune-performance"
description: "Investigate and improve performance using reproducible measurements. Use when something is slow or resource-heavy and the fix needs profiling and before-and-after measurements."
model_profile: "reasoning"
access_profile: "implement"
---

# hive-tune-performance

1. Define the user-visible metric, representative workload, environment, and baseline before optimizing.

2. Profile the relevant path and identify the limiting resource. Separate measured causes from hypotheses.

3. Change one consequential factor at a time when practical, compare under comparable conditions, and report variation, sample limits, and resource tradeoffs. Prefer a deterministic proxy, such as instruction, call, render, or query counts, over noisy timing only after showing that it moves with the user-visible metric; report that metric as the result.

4. Preserve correctness while verifying the improvement. Do not invent benchmark numbers or infer production gains from an unrepresentative local run. Propose a regression check that keeps the improvement, such as a CI threshold that may only tighten; add it only when the parent's scope includes it.

Stay within the parent’s assigned scope, write ownership, and output destination. Commit only in your assigned working tree and branch, and only when the brief allows it; never push, open or merge pull requests, write to trackers, or launch children of your own. Modify or delete only the paths the brief names, list planned deletions before applying them, and stop and report when the work needs a change outside them. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`, or in a separate worktree. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. Unless the brief says your dependency installation is not shared, run no command that installs, prunes, or relinks dependencies, such as `pnpm install`, `pnpm add`, or a filtered `pnpm --filter … exec`. Run a test suite that sets up, resets, or binds a local service, such as a test database, a cache, or a port, only on the instance and at the time the brief assigns; when the brief assigns none, first check whether another agent's run is using that service, and if one is, do not run the suite and report it as not run. Run installs, builds, and other long commands in the foreground with a bounded timeout rather than behind a background monitor, and stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
