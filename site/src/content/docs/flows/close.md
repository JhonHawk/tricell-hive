---
title: Close
description: Wrap up a finished, paused or abandoned piece of work and clean up only what that work created.
---

Close is the last step of a piece of work. It makes sure the records match reality, removes the leftovers the task made, and tells you where things stand.

Its reporting reference applies to findings, relayed results, explanations and
messages with external links or IDs, including progress updates. Greetings and
short direct answers need no reference read.
Questions and artifact cleanup have their own direct reference links. Those
procedures retain their existing obligations; reading a reference does not
invoke another complete flow.

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

For work with a walkthrough, it first marks only that work closed and regenerates
its page before removing a working copy or review artifacts. Closed also covers
abandoned or paused work; cleanup by itself does not close work. The existing
viewer must confirm that it collected the exact current source and resources
before their source unit can be removed. If the viewer is offline, too old, has
no matching copy or does not confirm within 15 seconds, the agent keeps the
still-needed unit and reports its folder. It never starts or restarts the viewer
to force confirmation. The procedure lives in
[Close the current walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md#close-the-current-walkthrough).

1. When the work is complete, moves the ticket to its finished state, along with the sub-issues the same work covered. A paused or abandoned ticket does not move to finished.
2. Closes the change record, if there is one: for completed work it folds the plan's requirement changes into the project's current requirements, and it archives the folder. An abandoned change is archived without touching the requirements.
3. Deletes branches that are fully merged, and their clean working copies.
4. Removes temporary files the task made, stops processes it started, and puts back local things the task changed that you rely on, such as your development containers or database schema, instead of leaving them to you.
5. Lists side findings that the running system can actually trigger, marks which ones it recommends turning into tickets, and lets you open a ticket for each, decline it, or ask not to hear about it again. It asks about every finding it lists, over several question cards when they do not fit in one, and never declines one on your behalf. A suspicion it could not trace that far stays in the report as unconfirmed.
6. Writes a completion report that says what was done, what was cleaned, what stays open, and which implementation decisions it took without asking, with how to revert each.
7. Asks what to do next: continue with a suggested ticket or the rest of the plan, take it to a new session, or end. While a plan is still in progress, continuing covers all its remaining tasks and asks once for the permissions they still need; the agent does not suggest a new session then.

## What it does not do

- It does not delete a branch that has unmerged commits. It shows you that work and asks.
- It does not delete a working copy that has uncommitted or untracked files.
- It does not remove material whose owner or purpose is unclear.
- It does not archive a paused change. The folder stays on disk and is reported as pending.
- It does not push, merge or deploy on its own. Versioning the change record follows the delivery choice you already made.

## Related flows

- [Build](/flows/build/) usually comes right before Close.
- [Plan](/flows/plan/) created the change record that Close archives.
- [Research](/flows/research/) is separate. A research-only session ends with its route question instead.

## FAQ

### What does Close delete?
Temporary files and processes the task created, merged branches, including ones from earlier work, with their clean working copies, and, at the end of the session, the interface screenshots, flow recordings and walkthrough page with its JSON file that you did not ask to keep: keeping the page keeps every file it shows, and keeping single files drops the page with the rest. Paused work that continues in another session keeps them all, as does an unresolved confirmation of the viewer's current closed copy. A branch with unmerged commits is never deleted, and anything whose owner is unclear is kept.

### Does Close mean the work is shipped?
It means the work reached its completion point, for example merged into the base branch. Later promotion to other environments is tracked separately.

### What if I only want to stop for now?
Say you are pausing. The plan stays where it is, and the report lists what is pending and who owns it.

### Do I have to answer the closing question?
It is how the agent offers the next route. The message opens with a bold context line, stage · ticket · progress, so you can tell which piece of work is asking; a question card opens its question with the stage and ticket. With a question card that waits for your answer, the message keeps a summary line with the recommended option and the options are in the card; otherwise, or after a dismissal or timeout, the options are written in the message, which ends with the "Waiting for your decision:" line. A dismissed or timed-out card is not an answer, so the agent asks again in text. Choosing to end the session is a valid answer.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-close/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)
