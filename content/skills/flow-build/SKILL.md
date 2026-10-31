---
name: flow-build
description: Implement or resume authorized code or configuration work; reconcile current state, build the candidate, and hand it to flow-verify. Use for execution rather than research-only or planning-only requests; a formal plan is not a prerequisite for a small understood change.
---

# Implement authorized work

Enter with a ready plan, or with a small understood change the user asked for. Without a plan, ask the delivery question of `flow-plan` step 4 before the first edit.

1. Check the plan's prerequisites. For each unmet one, ask the user for it and start the tasks that do not need it; mark it done once confirmed, without writing a secret's value. Move the tracked ticket to in progress. When the delivery makes commits, work on a branch named by the repository's convention.
2. For each task: implement it, run the affected tests, lint and typecheck, and commit it when the delivery makes commits. Mark the task in `tasks.md` with one line: done, its commit, and anything left unverified.
3. When a task changes stored data, a migration, security, money, concurrency or a public contract, have one `hive-verify-task` child check it against its criteria.
4. When the change adds or rearranges a screen, show the user its first rendered view with `flow-walkthrough` as soon as it exists, and adjust before continuing.
5. Before finishing, check each acceptance criterion against the result, then hand the candidate to `flow-verify`. A mechanical change skips verify: run the local checks of `flow-verify` step 6, integrate as the delivery says, and run `flow-close`.

When the host has a native task list, list the tasks in it at the start, and again with the remaining ones when you resume, such as after a compaction; mark each one in progress when it starts and completed when it is done. `tasks.md` still governs.

Repairing the local environment the work needs, such as reinstalling or relinking dependencies, regenerating a client or seeding local data, is part of the authorized work and needs no question: do it and name it in the final report.

Between tasks, continue with the next tool call instead of ending the turn with a progress report. End the turn only to ask a question that must be asked now, at a validation stop, at a blocker that stops all remaining work, or when the user stops the work.

## Stored data

- Before choosing a technique, take a read-only count of the affected records in each environment the change reaches, and record it. Scale the technique to that count: a few records need one targeted update in an authorized window; a count of zero needs no migration.
- Migrate or remove the stored values the new code stops reading, as a task of the same change.

## Security boundaries

When the change touches untrusted input, permissions, process execution or data exposure:

- Enforce authorization at the trusted execution boundary, including the denied and cross-owner cases.
- Keep data apart from executable syntax: parameterized queries and structured process arguments, never interpolation into shell, query, template or path expressions.
- Keep secrets and internal fields out of errors, responses, logs and evidence.
- Offer `hive-review-security` in one line; it runs only when the user asks for it.
