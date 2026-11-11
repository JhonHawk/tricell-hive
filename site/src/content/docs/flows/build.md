---
title: Build
description: Make an approved change in your project and hand over a built candidate with the evidence of its own checks.
---

Build is where the work gets done. The agent changes code or configuration, runs the checks your project has, and hands a built candidate to [Verify](/flows/verify/). The [Build skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md) owns the steps.

## When it starts

Build starts when a plan is ready, when you type its command, when you pick a small fix from the close footer or a findings round, or from [Research](/flows/research/) in the same turn, with no command to type, when you asked for a change and its how leaves you no decision and can be finished from your request and the `git diff`, unless it brings a new screen that Research first offers to [Figma design](/flows/figma-design/), which then ends with this command. When the how has a decision for you, the agent goes to [Plan](/flows/plan/) first, unless you chose to skip the plan. The [viewer](/viewer/) session for the work starts at the first UI capture, when the work has one.

```text wrap
Implement the agreed change locally and verify it. Do not commit or publish.
Continue the plan in the change folder from the next unfinished task.
Fix the date parsing bug and add a test that would have caught it.
```

## What it does

Build publishes only UI captures to the [viewer](/viewer/), with their iteration, and only when the work changes an interface; it declares no stages or tasks there. Task progress lives in `tasks.md` or the message. Tasks that the host's native list already reports (Claude Code, OpenCode) appear in the viewer marked as from the host, without the agent publishing them; with the viewer stopped, a capture is queued and sent later. See the [capture publication](/viewer/#publish-captures) and [host-observed activity](/viewer/#host-observed-activity).

- Checks the plan's prerequisites and asks you for any that are missing, while it starts the tasks that do not need them.
- Moves the tracked ticket to in progress, except for a one-off operation that leaves nothing in the repository, whose ticket moves only if you ask, and works on a branch when your delivery answer makes commits.
- Works task by task: implements it, or gives it to a helper when the task is settled, does not share files with another task in progress and needs more than five large reads or edits more than three files (independent tasks that write, such as a backend and a frontend task, run side by side only against a typed contract frozen beforehand, with an integration check after both); then runs the checks that cover it, commits when the delivery says so, and ticks its boxes in `tasks.md` with a bundled script that also writes the task's progress line and keeps the build checkpoint, with the checks recorded so far, in a `## Checkpoint` section, or marks it in the message when the plan lives there. When a check shows that a criterion is wrong, you decide, and the plan is then [revised](/flows/plan/#revising-a-plan-in-progress). Helpers never commit, push or write to trackers; the main agent checks what they return.
- Writes few, meaningful tests. A test must fail before the code changes only for a bug it fixes, as one regression where the bug lives, and for permissions, money, data integrity or a public contract; other changes get a direct check or extend the test that already covers that behavior. Each contract has one owning test at the strongest boundary your test setup runs cheaply, and only external services, time, randomness and, in a client's tests, your own server API through typed stub data are faked. Tests also cover the failure, limit, ordering or rollback a change handles. In a project without tests, a bug fix or consequential change first pins what the code does today for everything that calls it, and the agent asks you once whether to add a test setup. The [Build skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md#tests) holds the rule.
- Has a separate helper check a risky task of a plan with a change folder: one that touches a contract, stored data or a migration, security or permissions, money, concurrency or a dependency. The helper gets the task, its criteria and the check outputs, never the builder's reasoning. Stored-data changes get a count of affected records and a recovery path before anything runs.
- Offers a security review in one line; it runs only if you ask.
- Repairs the local environment the work needs, such as reinstalling dependencies, without asking, names it in the report, and records what it verified about your machine, such as the ports and commands that worked, in the project's `HIVE.local.md`, and what holds in any project, such as a missing tool, in the machine file `~/.config/hive/HIVE.local.md`. A helper only your machine needs, such as a script that creates a database per worktree, it writes as a local script in `_support/scripts/`, kept out of Git, names in `HIVE.local.md` and lists in its report.

## What it asks you

- The delivery mode and the code review, once per session and repository, before the first edit, unless the repository declares direct delivery or the work leaves nothing to commit. Later work reuses your answers and says which ones apply; you are asked again only to change them, when your answer has no validation stop or review and the new work calls for one, or when the mode would merge or publish commits that are not this work's. A mechanical change gets only the delivery question. See the [delivery rules](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).
- For a change that adds or rearranges a screen, to accept its first rendered view, compared against the plan's sketch or the approved example when the work has one, or ask for corrections before the rest continues. The capture is published to the [viewer](/viewer/) through the [walkthrough skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md). When you have delegated the work to run [unattended](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md), the view is captured and published without waiting for you.

Between tasks it keeps going rather than reporting. It stops to ask you something, at a validation stop, or at a blocker.

## How it ends

When every task is done, the agent runs locally the CI jobs that need no external service, secret or deployment, then continues into Verify in the same turn, with no command for you to type, handing over the candidate and its recorded check evidence; it stops there only if you asked it to stop after Build or to skip Verify. That handoff lets Verify reuse results that still cover the candidate rather than repeat them merely because the stage changed. A mechanical change skips Verify: its criteria are checked, it is integrated as your delivery says, and it ends with the [Close](/flows/close/) footer; when you chose to validate before the push, it still stops there and shows you what changed. A one-off operation skips the CI run and Verify: its criteria are checked and it ends with the footer. The progress and check evidence are kept in the existing checkpoint, so a long session can resume after a compaction.
