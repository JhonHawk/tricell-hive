---
name: flow-build
description: Implement or resume authorized code or configuration work; reconcile current state, build the candidate, and hand it to flow-verify. Use for execution rather than research-only or planning-only requests; a formal plan is not a prerequisite for a small understood change.
---

# Implement authorized work

Enter with a ready plan, or with a small understood change the user asked for. Without a plan, ask the delivery question of `flow-plan` step 4 before the first edit.

1. Move the tracked ticket to in progress. When the delivery makes commits, work on a branch named by the repository's convention.
2. For each task: implement it, run the affected tests, lint and typecheck, and commit it when the delivery makes commits. Mark the task in `tasks.md` with one line: done, its commit, and anything left unverified.
3. When a task changes stored data, a migration, security, money, concurrency or a public contract, have one `hive-verify-task` child check it against its criteria.
4. When the change adds or rearranges a screen, show the user its first rendered view as soon as it exists, and adjust before continuing.
5. Before finishing, check each acceptance criterion against the result, then hand the candidate to `flow-verify`.
