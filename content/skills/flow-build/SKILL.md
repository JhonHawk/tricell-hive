---
name: flow-build
description: Implement or resume authorized code or configuration work; reconcile current state, build the candidate, and hand it to flow-verify. Use for execution rather than research-only or planning-only requests; a formal plan is not a prerequisite for a small understood change.
---

# Implement authorized work

At build entry, the primary producer follows [flow-walkthrough](../flow-walkthrough/SKILL.md), which sets when a viewer session applies, for the viewer stage; for a changed interface read [UI verification](../flow-verify/references/ui-verification.md) for capture scope.

Enter with a ready plan, in a change folder or kept in the conversation as [flow-plan](../flow-plan/SKILL.md#where-the-plan-lives) sets, or with a small understood change the user asked for (as the global guidance's «Proportionality» defines it), handed over by `flow-research` in the same turn or started by the user's typed command, or with a fix that the close footer or a findings round sends when that fix is a small understood change; other fixes go to `flow-plan`. Without a plan, ask the delivery and review questions as the global guidance's «Delivery» section sets, before the first edit.

1. Check the plan's prerequisites. For each unmet one, ask the user for it and start the tasks that do not need it; mark it done once confirmed, without writing a secret's value. Move the tracked ticket to in progress, except for an operational task, as the global guidance's «Scope and authorization» sets. When the delivery makes commits, work on a branch named by the repository's convention.
2. For each task: implement it, or delegate it as [delegation](references/delegation.md) sets, applying [Tests](#tests) before writing, changing or deleting a test; run the checks that cover it, reading their results and ordering the reruns of a round of fixes as [check reuse](../flow-verify/SKILL.md#check-reuse) sets, and commit it when the delivery makes commits. When a check fails, handle it as [Tests](#tests) sets. When the plan has a change folder, tick the task's boxes in `tasks.md` and add one line: done, its commit, and anything left unverified, with `python3 -I <flow-plan-dir>/scripts/tasks.py <folder> done T<n> --commit <sha> --unverified "<what>"` ([tasks](../flow-plan/scripts/tasks.py)), passing `none` as the commit when the delivery makes none and as the text when nothing stayed unverified, or by hand when it cannot run; and write one progress line in the message, such as `✓ T3 · 3/5 · export button`, which the script prints.
3. When a task of a plan with a change folder is a risky change, as the global guidance's «Proportionality» defines it, have one `hive-verify-task` child check it against its criteria, briefed as [delegation](references/delegation.md#reviewer-brief) sets.
4. When the change adds or rearranges a screen, follow the dependencies the plan declares rather than its listing order. Before asking the first-view question, decide what keeps going while the user reviews: each task that does not need that answer. Launch one as a child only when [delegation](references/delegation.md) sends it to one; keep the others in the main thread while the children work. Show the user its first rendered view with `flow-walkthrough` as soon as the UI task is done, compared against the plan's sketch; ask the user to accept it or request corrections, with the open-viewer option that [viewer publication](../flow-walkthrough/references/viewer-publication.md#open-the-viewer) sets, and adjust only the UI before continuing. Apply the first-view exception when [unattended delegation](../unattended-delegation/SKILL.md) is active. Skip the first rendered view when the UI task is the plan's last task: the walkthrough of `flow-verify` validates it once.
5. When every task is done, reproduce the repository's CI: read its workflow files and run locally every job that needs no external service, secret or deployment, for the packages the change touches and those that depend on them. Record the commands, relevant output, checked commit and any uncommitted changes, and the configuration and environment the checks depend on in the existing checkpoint. Then continue into `flow-verify` in the same turn, with no next-step block: read its skill whole, then hand it the candidate and that evidence. When the user asked to stop after build, end instead with the next-step block toward `flow-verify`; when they asked to skip verify, integrate as the delivery says and run `flow-close`. Verify owns [check reuse](../flow-verify/SKILL.md#check-reuse), checks the acceptance criteria in its walk, and holds the delivery's validation stop. A mechanical change skips verify: check its criteria, if it has any, integrate as the delivery says, and run `flow-close`. An operational task, as «Proportionality» defines it, skips the CI run and verify: check its criteria and run `flow-close`.

Keep a checkpoint in the `## Checkpoint` section of `tasks.md`: the stage and step, the last commit, the open findings with their decision, and the next action, written with `tasks.py <folder> checkpoint` (`--finding` per open finding, `--check` per check recorded with its command, result and commit, which later updates keep), or by hand when it cannot run. With a plan kept in the conversation, keep it in a file named after the work in the host's temporary location, written at the start and after each task, adding the decisions, criteria, pending task IDs, viewer work key if any, and each resource the work created outside the repository with its location. After a compaction, resume from it, the active skill and one `git status`; read code again only before editing it.

When the host has a native task list, list the tasks in it at the start, and again with the remaining ones when you resume and the host does not show it again, such as after a compaction; mark each one in progress when it starts and completed when it is done. The native list is a live view: `tasks.md`, or the checkpoint file of a plan kept in the conversation, governs. In the viewer, a build without plan tasks registers nothing by hand where the host reports its native task list (the viewer takes that list as the work's tasks); otherwise, and for the tasks of a plan, declare them as [viewer publication](../flow-walkthrough/references/viewer-publication.md#tasks) sets.

Repairing the local environment the work needs, such as reinstalling or relinking dependencies, regenerating a client or seeding local data (read the global guidance's «Test accounts» first when the data includes accounts), is part of the authorized work and needs no question: do it and name it in the final report, and record what it showed about this machine in `HIVE.local.md` as the global guidance's «Local instructions» sets. It never quits, restarts or kills a program other sessions may use, such as Docker Desktop, as the global guidance's «Preservation» sets.

Between tasks, continue with the next tool call instead of ending the turn with a progress report. End the turn only to ask a question that must be asked now, at a validation stop, at a blocker that stops all remaining work, or when the user stops the work.

## Tests

Read before writing, changing or deleting a test, when a check fails, when choosing a task's test approach in a plan, and, in a repository without a test runner, before the first edit of a defect fix or consequential logic.

**When a check fails**, first decide whether the code, the criterion or the test is wrong, and fix the code or the test: change, skip or delete an existing test only when the test itself is wrong, and say so in the task's line or your report. A wrong criterion is the user's decision: report it, or a child reports it to its parent, instead of changing it. Once the user decides, revise the plan as [Revise a plan in progress](../flow-plan/SKILL.md#revise-a-plan-in-progress) sets.

**When a test must fail first (`tdd`).** Write the test, run it and see it fail for the intended reason, then change the code until it passes, one behavior at a time. This applies only to:

- A defect you fix: one regression test, in the test that owns the broken behavior as point 2 of the filter below sets, seen failing on the code before the fix. Do not repeat it at every layer the defect crosses. When the failure cannot be reproduced before the fix, say so instead of claiming it.
- Consequential logic: permissions, money, data integrity, or a public or persisted contract.

Verify any other behavior change with a direct check, or by extending the test that already owns that behavior (`check`). A refactor that keeps behavior uses the existing tests, or first a test that pins the current behavior (`characterization`). A mechanical change adds no test.

**Before writing a test**, apply this filter; do not write a test that fails any point:

1. Name the realistic change to the code that would make it fail. If none exists, do not write it.
2. Find the test that already owns that behavior. Each contract has one owner, at the strongest boundary the repository's test harness runs cheaply, such as an integration test through HTTP and the database rather than a unit test with mocked collaborators. Extend the owner, its table or its fixtures, instead of adding a sibling test or the same assertion at another layer. Keep a lower-layer test only for a branch the owner cannot reach cheaply.
3. Replace with a test double only what crosses the system boundary: an external service, time, randomness, or, in a client's tests, the project's own server API through stub data typed by its contract. Do not mock the project's own modules in the same process, nor its own database when the repository has an integration harness: the test then stops covering their behavior.
4. Reject a test that asserts how the code is written instead of what it does, such as which internal functions were called or in what order; that uses a double to supply the behavior the test claims to check, so it tests the double rather than the code; that reads source code as text to check how it is written, unless the file's text is itself what the project delivers to its readers, such as guidance or documentation, never code or build configuration; that copies a list or inventory from the code; that computes its expected value with the code under test; that asserts something always true; or that needs an export, flag or wrapper only tests use.
5. Wait on an observable condition, never a fixed sleep, and isolate state shared between tests.

When the change handles a failure, a boundary value, a concurrent actor or a transaction, its verification covers that case too: the negative case, the limit, the ordering, or the rollback, in the owner test when there is one, otherwise in the direct check.

**Project or area without tests.** When the repository has a test runner but the package the change touches has no tests, the rules above apply, and the first test of a contract becomes its owner. When the repository has no test runner:

- Verify a change without consequences with a direct check, such as the build, the type checker, or running the changed path with a request, a script or the browser, and report "no tests; verified with <check>".
- For a defect or consequential logic, before editing, find every caller of the behavior by searching its literal values, such as names, routes, roles and hosts, and pin its current behavior for those callers with a check through the real path, adding no dependency. Keep that check where the project already keeps checks, or as evidence of the change when it has none.
- Adding test infrastructure is the user's decision: the main thread proposes it in the plan, or before the first edit when the work has no plan, the first time a defect fix or consequential logic reaches the project, with the package check of the global guidance, and saves the answer in persistent memory as a project decision so it is not asked again. Never install it without that answer; a child that needs a runner the project lacks reports it to its parent instead.

## Delegation

Before deciding whether a child does a task, briefing it or waiting for it, read [delegation](references/delegation.md): it sets when a task goes to a child, which role, its brief, the wait and what the main thread checks afterwards.

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
