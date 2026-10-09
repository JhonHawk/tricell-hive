---
title: Build
description: Make an approved change in your project and check that it works before calling it done.
---

Build is where the work gets done. The agent changes code or configuration, runs the checks your project has, and reports what it verified and what it could not.

## When to use

Use Build when you have approved a change and the outcome is clear. That can be a saved plan, or just a sentence.

Example requests:

```text wrap
Implement the agreed change locally and verify it. Do not commit or publish.
Continue the plan in the change folder from the next unfinished task.
Fix the date parsing bug and add a test that would have caught it.
```

You do not need a plan first for a small change you understand.

## What it produces

A changed working tree, plus evidence:

- The agent checks the current state before editing. If a plan exists, it treats the plan as a claim to confirm, not a script to replay.
- It works in small steps, often with helper agents for pieces that do not overlap, and tells you how it split the work.
- For behavior that would hurt if it broke, such as calculations, stored data, permissions or public contracts, it writes the test first and watches it fail, when the project's tests can exercise that behavior.
- It runs the project's own tests, builds and linters and reports the results, including any that failed or could not run.
- For visible interface changes, independent reviewers look at the running result. The agent's own look does not count for that.
- Between tasks it reports progress in a line and keeps going; the completion report and the closing question come when the scope you authorized is done, not after every task.
- If you ask how it is going, it answers in a line and keeps working. Builds, deployments and CI runs it starts are watched with your CLI's monitor tool when it has one, until they finish, so a failure does not go unnoticed.
- When your CLI has a task list, the agent keeps the plan's tasks there, also when it continues a plan after a close, so you can see what is left.
- When a plan has numbered criteria, a separate verifier checks each task that is proven by a test or a named check. Other tasks are marked done on the evidence of their own check.

If Git delivery was not settled earlier, the agent asks once, before the first edit, how you want it delivered and reviewed.

## What it does not do

- It does not commit, push, open pull requests, merge or deploy unless you asked for that effect.
- It does not mark a check as passed when it could not run it.
- It does not edit a test you supplied as the acceptance check just to make it pass.
- It does not treat a successful build or a ticked box as proof by itself.

## Related flows

- [Plan](/flows/plan/) comes first for larger work and gives Build its tasks and checks.
- [Research](/flows/research/) helps when the cause of a failure is still unknown.
- [Close](/flows/close/) follows when the work is finished, abandoned or paused.

## FAQ

### Do I need a plan for a small change?
No. Give Build the outcome and its limits.

### Who commits and merges?
You decide. Say so in your request, or answer the delivery question the agent asks. Without that, the agent stops at a verified local result.

### What if the agent finds that the plan no longer matches the code?
It looks into the mismatch and updates the plan or asks you for the missing decision before it keeps editing.

### What happens to a user interface change?
The agent leaves the running result available for you to check, with steps to follow, and cleans up the temporary captures later.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-build/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md)
