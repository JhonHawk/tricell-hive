---
title: Close
description: Wrap up a finished, paused or abandoned piece of work and clean up only what that work created.
---

Close is the last step of a piece of work. It makes sure the records match reality, removes the leftovers the task made, and tells you where things stand.

## When to use

Use Close when work reaches its completion point (usually merged into the base branch, after any checks that must pass there), when you abandon it, when you pause it, or when you end the session or ask for a cleanup.

Example requests:

```text wrap
Close this work item and clean up only the temporary resources it created.
Pause this work. Leave the plan on disk and tell me what is still open.
Clean up the session.
```

## What it produces

The agent goes through these steps in order:

1. When the work is complete, moves the ticket to its finished state, along with the sub-issues the same work covered. A paused or abandoned ticket does not move to finished.
2. Closes the change record, if there is one: for completed work it folds the plan's requirement changes into the project's current requirements, and it archives the folder. An abandoned change is archived without touching the requirements.
3. Deletes branches that are fully merged, and their clean working copies.
4. Removes temporary files the task made, stops processes it started, and puts back local things the task changed that you rely on, such as your development containers or database schema, instead of leaving them to you.
5. Lists side findings, so you can open a ticket for each, decline it, or ask not to hear about it again.
6. Writes a completion report that says what was done, what was cleaned and what stays open.
7. Asks what to do next: continue with a suggested ticket, take it to a new session, or end.

## What it does not do

- It does not delete a branch that has unmerged commits. It shows you that work and asks.
- It does not delete a working copy that has uncommitted or untracked files.
- It does not remove material whose owner or purpose is unclear.
- It does not archive a paused change. The folder stays on disk and is reported as pending.
- It does not push, merge or deploy on its own. Versioning the change record follows the delivery choice you already made.

## Related flows

- [Build](/flows/build/) usually comes right before Close.
- [Plan](/flows/plan/) created the change record that Close archives.
- [Research](/flows/research/) is separate. A research-only session closes with the offer to show recommendations instead.

## FAQ

### What does Close delete?
Temporary files and processes the task created, merged branches, including ones from earlier work, with their clean working copies, and, at the end of the session, interface screenshots you did not ask to keep, unless the work is paused and continues in another session. A branch with unmerged commits is never deleted, and anything whose owner is unclear is kept.

### Does Close mean the work is shipped?
It means the work reached its completion point, for example merged into the base branch. Later promotion to other environments is tracked separately.

### What if I only want to stop for now?
Say you are pausing. The plan stays where it is, and the report lists what is pending and who owns it.

### Do I have to answer the closing question?
It is how the agent offers the next route. The options are written in the message, which ends with the "Waiting for your decision:" line. If your tool also shows a question card, it is only a shortcut, so dismissing it loses nothing. Choosing to end the session is a valid answer.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-close/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)
