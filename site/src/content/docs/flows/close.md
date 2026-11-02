---
title: Close
description: Wrap up a finished, paused or abandoned piece of work, report what happened and clean up only what that work created.
---

Close is the last step of a piece of work. It makes sure the records match reality, removes the leftovers the task made, tells you where things stand and asks what comes next.

## When to use

Close starts when work reaches its completion point: the merge into the base branch, or, when a human reviewer is assigned to the pull request, the opening of that pull request. In that second case the ticket stays in progress, and a later session that finds the pull request merged runs the close. It also runs when you abandon or pause work, when you end the session, or when you ask for a cleanup.

Example requests:

```text wrap
Close this work item and clean up only the temporary resources it created.
Pause this work. Leave the plan on disk and tell me what is still open.
Clean up the session.
```

## What it asks

One question at the end, the close question. Its options are to end the session, to see the findings (only when the work left some), or to find the next ticket. The findings round asks, for each defect the work found but did not fix, whether to fix it now, open a ticket or ignore it. A ticket is never opened without your answer. Close runs once: after you answer, the agent acts on the answer and does not repeat the report.

## What it produces

- The ticket moved to the state your project uses for integrated work, and the change record folded into the project's current requirements and archived.
- Merged branches and their clean working copies removed, temporary files deleted and the processes the work started stopped. Anything with uncommitted work or uncertain ownership is kept.
- One final report of the whole work item: what was integrated, how each acceptance criterion was shown, what was skipped or could not run, local repairs, the findings and what remains.
- A page version of the report through [flow-report](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-report/SKILL.md), offered as an option when you want to keep or share it.

When the work had a viewer session, Close finishes it first; see [Viewer](/viewer/).

## How it ends

With the close question answered. Choosing to end the session is a valid answer. The exact steps, the severity scale for findings and the branch cleanup rules are in the [flow-close skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md).

## What it does not do

- It does not delete a branch with unmerged commits or a working copy with uncommitted files; it shows them and asks.
- It does not push, merge or deploy on its own.
- It does not archive a paused change; the folder stays on disk.

## Related flows

- [Verify](/flows/verify/) usually comes right before Close, because it integrates the change; [Build](/flows/build/) does when verify is skipped.
- [Deploy](/flows/deploy/) comes after integration. When it runs, only the cleanup of Close follows it.
- [Changes and loops](/flows/overview/) shows where the work goes when your answer adds to it.

## FAQ

### What if I only want to stop for now?
Say you are pausing. The plan stays where it is, and the report lists what is pending.

### Does Close mean the work is shipped?
It means the work reached its completion point, for example merged into the base branch. Promotion to other environments is [Deploy](/flows/deploy/).

Source: [content/skills/flow-close/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)
