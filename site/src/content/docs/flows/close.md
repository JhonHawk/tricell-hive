---
title: Close
description: Wrap up a finished, paused or abandoned piece of work, report what happened and clean up only what that work created.
---

Close is the last step of a piece of work. It makes sure the records match reality, removes the leftovers the task made, tells you where things stand and asks what comes next.

## When to use

Close starts when work reaches its completion point: the merge into the base branch, or the open pull request when the delivery is `ready` (it stops before the merge for you) or a human reviewer is assigned to it. In that second case the ticket stays in progress, the viewer shows the work as in review, and a later session that finds the pull request merged finishes only what waited for the merge: the viewer close, the ticket, the change record and the branch cleanup. It also runs when a one-off operation that leaves nothing in the repository ends, such as a database dump, when you abandon or pause work, when you end the session, or when you ask for a cleanup. Work kept on hold, with nothing committed, is fully closed later, when it is delivered; until then it is paused: the processes and containers it started are stopped and the report lists what is pending.

Example requests:

```text wrap
Close this work item and clean up only the temporary resources it created.
Pause this work. Leave the plan on disk and tell me what is still open.
Clean up the session.
```

## What it asks

One round at the end. When the work left material in the project's support folder, such as research records, working files or evidence, the round also asks what to do with it: each item comes with a proposed action (commit it, archive it, move a conclusion into the project's documentation, delete it, or keep it), and nothing is moved, committed or deleted before you answer. Paused work, work on hold and work waiting at an open pull request only list that material. The close question itself offers to end the session, to see the findings (only when the work left some), or to find the next ticket. The findings round asks, for each defect the work found but did not fix, whether to fix it now, open a ticket or ignore it. Fixing now is recommended for a small fix in the area the work touched; a ticket, for work that will be done later or by someone else, after checking that none exists and linked to the ticket it came from. A ticket is never opened without your answer. When you choose to find the next ticket, the status report that follows offers to compact the session before the next work starts. Close runs once: after you answer, the agent acts on the answer and does not repeat the report.

## What it produces

- The ticket moved to the state your project uses for integrated work, and the change record folded into the project's current requirements and archived.
- Merged branches and their clean working copies removed, temporary files deleted, the processes stopped, and the containers it started (with their volumes) and the images it built removed. A dump or export it made is listed and removed only if you say so. Resources it created outside your machine, such as storage objects or registry images, are listed with the command that removes each, and removed only if you say so. Anything with uncommitted work or uncertain ownership is kept.
- The work's support material handled as you answered: what you chose to commit, archive or move into the documentation is committed with the archived change record, or in a commit of their own on a new branch when there is none, delivered the way the work was; what you chose to keep stays where it is.
- One final report of the whole work item: what was integrated, how each acceptance criterion was shown, what was skipped or could not run, local repairs, the work's support material with its proposed actions, the findings and what remains.
- A page version of the report through [flow-report](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-report/SKILL.md), offered as an option when you want to keep or share it.

When the work had a viewer session, Close finishes it first; see [Viewer](/viewer/). When the viewer is not running, Close goes on without it, keeps what the viewer close needs, and the report says which viewer session stayed open.

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
