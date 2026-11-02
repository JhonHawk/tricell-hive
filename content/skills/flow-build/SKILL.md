---
name: flow-build
description: Implement or resume authorized code or configuration work; reconcile current state, build the candidate, and hand it to flow-verify. Use for execution rather than research-only or planning-only requests; a formal plan is not a prerequisite for a small understood change.
---

# Implement authorized work

At build entry, the primary producer follows [flow-walkthrough](../flow-walkthrough/SKILL.md), which sets when a viewer session applies, for the viewer stage; for a changed interface read [UI verification](../flow-verify/references/ui-verification.md) for capture scope.

Enter with a ready plan, or with a small understood change the user asked for (as the global guidance's «Proportionality» defines it). Without a plan, ask the delivery and review questions as the global guidance's «Delivery» section sets, before the first edit.

1. Check the plan's prerequisites. For each unmet one, ask the user for it and start the tasks that do not need it; mark it done once confirmed, without writing a secret's value. Move the tracked ticket to in progress. When the delivery makes commits, work on a branch named by the repository's convention.
2. For each task: implement it, or delegate it as [Delegation](#delegation) sets; run the checks that cover it, and commit it when the delivery makes commits. When a check fails, first decide whether the code, the criterion or the test is wrong, and fix that one: change, skip or delete an existing test only when the test itself is wrong, and say so in the task's line. Mark the task in `tasks.md` with one line: done, its commit, and anything left unverified; and write one progress line in the message, such as `✓ T3 · 3/5 · export button`.
3. When a task changes stored data, a migration, security, money, concurrency or a public contract, have one `hive-verify-task` child check it against its criteria.
4. When the change adds or rearranges a screen, follow the dependencies the plan declares rather than its listing order. Before asking the first-view question, decide what keeps going while the user reviews: each task that does not need that answer. Launch one as a child only when it meets the [Delegation](#delegation) criterion, since a child starts with a large base context and pays off only for work larger than its brief; keep the others in the main thread while the children work. Show the user its first rendered view with `flow-walkthrough` as soon as the UI task is done, compared against the plan's sketch; ask the user to accept it or request corrections, and adjust only the UI before continuing. Apply the first-view exception when [unattended delegation](../unattended-delegation/SKILL.md) is active. Skip the first rendered view when the UI task is the plan's last task: the walkthrough of `flow-verify` validates it once.
5. When every task is done, reproduce the repository's CI: read its workflow files and run locally every job that needs no external service, secret or deployment, for the packages the change touches and those that depend on them. Then hand the candidate to `flow-verify`, which checks the acceptance criteria in its walk. A mechanical change skips verify: check its criteria, if it has any, integrate as the delivery says, and run `flow-close`.

Keep a checkpoint in `tasks.md`: the stage and step, the last commit, the open findings with their decision, and the next action. After a compaction, resume from it, the active skill and one `git status`; read code again only before editing it.

When the host has a native task list, list the tasks in it at the start, and again with the remaining ones when you resume, such as after a compaction; mark each one in progress when it starts and completed when it is done. `tasks.md` still governs.

Repairing the local environment the work needs, such as reinstalling or relinking dependencies, regenerating a client or seeding local data, is part of the authorized work and needs no question: do it and name it in the final report.

Between tasks, continue with the next tool call instead of ending the turn with a progress report. End the turn only to ask a question that must be asked now, at a validation stop, at a blocker that stops all remaining work, or when the user stops the work.

## Delegation

The main thread coordinates and verifies; it keeps its own context small so the session does not keep compacting.

- Delegate a task when its interface is settled, its writes do not overlap files another task in progress writes, and its brief is shorter than the change itself. Otherwise keep it in the main thread: when its brief would not be shorter than its change, or it writes files another task in progress writes.
- Pick the role by surface: `hive-build-backend`, `hive-build-frontend`, `hive-build-data`, `hive-build-infra`, `hive-build-kmp`, `hive-write-tests` or `hive-write-docs`.
- Launch independent children in the same step, not in waves; tasks that share a file go one after another.
- The brief names the change folder and the task ID and tells the child to read the plan files before editing; it does not paraphrase the task or its criteria.
- Wait for children with the host's wait tool, which returns when a child finishes or sends a message; never with a loop of sleeps and status checks.
- Children stay in their working tree and files: they do not commit, push, open pull requests, write trackers or launch children. The main thread checks the evidence they return, runs the task's checks, commits and marks the task.

## Stored data

- Before choosing a technique, take a read-only count of the affected records in each environment the change reaches, and record it. Scale the technique to that count: a few records need one targeted update in an authorized window; a count of zero needs no migration.
- Migrate or remove the stored values the new code stops reading, as a task of the same change.
- Before running a change to stored data, define its cutover checks, the conditions that abort it, and its recovery path, and verify that the backup or restore it relies on exists. Rolling back code or schema does not restore transformed or deleted data; when reversal cannot recover it, choose a validated restore or a fix-forward.

## Security boundaries

When the change touches untrusted input, permissions, process execution or data exposure:

- Enforce authorization at the trusted execution boundary, including the denied and cross-owner cases.
- Keep data apart from executable syntax: parameterized queries and structured process arguments, never interpolation into shell, query, template or path expressions.
- Keep secrets and internal fields out of errors, responses, logs and evidence.
- Offer `hive-review-security` in one line; it runs only when the user asks for it.
