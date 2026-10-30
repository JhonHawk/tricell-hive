---
name: "hive-refute-claim"
description: "Challenge a specific claim or proposed finding with counterevidence. Use to test whether a reported finding or conclusion holds before acting on it."
model_profile: "inherit"
access_profile: "observe"
---

# hive-refute-claim

1. State the claim and the evidence that would confirm or refute it. Inspect the actual code path, configuration, versions, and assumptions; for a claim about documentation, process, or project state, inspect the cited sources and the live state instead.

2. Look for counterexamples, unreachable preconditions, existing safeguards, and missing context before agreeing with the claim.

3. Use focused checks only after inspecting their effects; do not mutate live state merely to prove a point.

4. Classify the result as confirmed, refuted, plausible, or unverifiable, with evidence and remaining uncertainty. Refute only what the evidence shows false: for a code claim, the claim is factually wrong, impossible given types or invariants, already handled, or only a matter of style; for a claim about documentation, process or state, the cited source or the live state contradicts it, or it describes an older version or another scope. A realistic state that is rare but reachable, such as a race or a boundary value, keeps a finding plausible; being hard to trigger is not a refutation. Do not silently fix the issue under review.

Stay within the parent’s assigned scope, write ownership, and output destination. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`, or in a separate worktree. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. Unless the brief says your dependency installation is not shared, run no command that installs, prunes, or relinks dependencies, such as `pnpm install`, `pnpm add`, or a filtered `pnpm --filter … exec`. Run a test suite that sets up, resets, or binds a local service, such as a test database, a cache, or a port, only on the instance and at the time the brief assigns; when the brief assigns none, first check whether another agent's run is using that service, and if one is, do not run the suite and report it as not run. Stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
