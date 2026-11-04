---
title: Build
description: Make an approved change in your project and hand over a built candidate with the evidence of its own checks.
---

Build is where the work gets done. The agent changes code or configuration, runs the checks your project has, and hands a built candidate to [Verify](/flows/verify/). The [Build skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md) owns the steps.

## When it starts

Build starts when a plan is ready, when you type its command, when you pick a small fix from the close footer or a findings round, or from [Research](/flows/research/) in the same turn, with no command to type, when you asked for a change and its how leaves you no decision and can be finished from your request and the `git diff`. When the how has a decision for you, the agent goes to [Plan](/flows/plan/) first, unless you chose to skip the plan. The [viewer](/viewer/) session for the work starts here when it did not start at Plan.

```text wrap
Implement the agreed change locally and verify it. Do not commit or publish.
Continue the plan in the change folder from the next unfinished task.
Fix the date parsing bug and add a test that would have caught it.
```

## What it does

Task progress is published separately through semantic task and stage commands. The binary resolves technical IDs and versions; the agent still supplies the actual state, reason, ownership and blocking information. Editing the published plan checklist does not update live tasks. See the [viewer publication model](/viewer/#publish-markdown-on-demand).

- Checks the plan's prerequisites and asks you for any that are missing, while it starts the tasks that do not need them.
- Moves the tracked ticket to in progress, except for a one-off operation that leaves nothing in the repository, whose ticket moves only if you ask, and works on a branch when your delivery answer makes commits.
- Works task by task: implements it, or gives it to a helper when the task is settled, does not share files with another task in progress and needs more than five large reads or edits more than three files (independent tasks that write, such as a backend and a frontend task, run side by side only against a typed contract frozen beforehand, with an integration check after both); then runs the checks that cover it, commits when the delivery says so, and marks it in `tasks.md`, or in the message when the plan lives there. Helpers never commit, push or write to trackers; the main agent checks what they return.
- Writes few, meaningful tests. A test must fail before the code changes only for a bug it fixes, as one regression where the bug lives, and for permissions, money, data integrity or a public contract; other changes get a direct check or extend the test that already covers that behavior. Each contract has one owning test at the strongest boundary your test setup runs cheaply, and only external services, time, randomness and, in a client's tests, your own server API through typed stub data are faked. Tests also cover the failure, limit, ordering or rollback a change handles. In a project without tests, a bug fix or consequential change first pins what the code does today for everything that calls it, and the agent asks you once whether to add a test setup. The [Build skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md#tests) holds the rule.
- Has a separate helper check a risky task of a plan with a change folder: one that touches a contract, stored data or a migration, security or permissions, money, concurrency or a dependency. The helper gets the task, its criteria and the check outputs, never the builder's reasoning. Stored-data changes get a count of affected records and a recovery path before anything runs.
- Offers a security review in one line; it runs only if you ask.
- Repairs the local environment the work needs, such as reinstalling dependencies, without asking, and names it in the report.

## What it asks you

- The delivery mode and the code review, once, before the first edit, unless you already answered or the repository declares direct delivery. A mechanical change gets only the delivery question. See the [delivery rules](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).
- For a change that adds or rearranges a screen, to accept its first rendered view or ask for corrections before the rest continues. The view is published to the [viewer](/viewer/) through the [walkthrough skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md). When you have delegated the work to run [unattended](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md), the view is captured and published without waiting for you.

Between tasks it keeps going rather than reporting. It stops to ask you something, at a validation stop, or at a blocker.

## How it ends

When every task is done, the agent runs locally the CI jobs that need no external service, secret or deployment, then hands the candidate and its recorded check evidence to Verify. That handoff lets Verify reuse results that still cover the candidate rather than repeat them merely because the stage changed. A mechanical change skips Verify: its criteria are checked, it is integrated as your delivery says, and it ends with the [Close](/flows/close/) footer. A one-off operation skips the CI run and Verify: its criteria are checked and it ends with the footer. The progress and check evidence are kept in the existing checkpoint, so a long session can resume after a compaction.
