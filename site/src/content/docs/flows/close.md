---
title: Close
description: Wrap up a finished, paused or abandoned piece of work, report what happened and clean up its resources and the merged branches left in the repositories it touched.
---

Close is the last step of a piece of work. It moves the ticket, tells you where things stand and ends with a footer of replies; the record archive and the cleanup of what the task made run when you choose to close.

## When to use

Close starts when work reaches its completion point: the merge into the base branch, or the open pull request when the delivery is `ready` (it stops before the merge for you) or a human reviewer is assigned to it. In that second case the ticket stays in progress, the viewer shows the work as in review, and a later session that finds the pull request merged finishes only what waited for the merge: the viewer close, the ticket, the change record and the branch cleanup. It also runs when a one-off operation that leaves nothing in the repository ends, such as a database dump, when you abandon or pause work, when you end the session, or when you ask for a cleanup. Work kept on hold, with nothing committed, is fully closed later, when it is delivered; until then it is paused: the processes and containers it started are stopped and the report lists what is pending. Work paused outside `hold`, such as a plan paused before its build, closes the same way and also versions its change folder.

Example requests:

```text wrap
Close this work item and clean up only the temporary resources it created.
Pause this work. Leave the plan on disk and tell me what is still open.
Clean up the session.
```

## What it asks

At the completion point it updates and reads back the viewer record when there is one, moves the ticket when integrated, and reports, then stops at a footer, not a card: it recommends closing and lists replies from the work's context. The support decision, record archive and cleanup run when you choose to close, start other work (the pending close runs first, in one line) or end the session; if the session just stops, the next close or cleanup request picks it up. When it runs, and the work left material in the project's support folder, such as research records, working files or evidence, it first asks what to do with that material, before it archives the change record and cleans up: each item comes with a proposed action (commit it, archive it, move a conclusion into the project's documentation, delete it, or keep it), and nothing is moved, committed or deleted before you answer. Paused work, work on hold and work waiting at an open pull request only list that material. The footer recommends closing, or seeing the findings when one is P1 or higher, or continuing a plan in progress. Its other replies come from the context: resolve named findings in the work's branch, see the findings, explain a finding or result, take the next ticket, or continue with what remains. The findings round asks, for each defect the work found but did not fix, whether to fix it now, open a ticket or ignore it. Fixing now is recommended for a small fix in the area the work touched; a ticket, for work that will be done later or by someone else, after checking that none exists and linked to the ticket it came from. A ticket is never opened without your answer. Under [unattended delegation](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md) the whole close runs at once and ends on the report, without the footer. Close runs once: after your reply, the agent acts on it and does not repeat the report.

## What it produces

- The ticket moved to the state your project uses for integrated work, in the tracker its key belongs to, and the change record folded into the project's current requirements and archived.
- Merged branches and their clean working copies removed in each repository the work touched, its specs repository included: the work's own branch and the earlier merged branches you authored, unless a working copy holds files that are not committed, a process is running in it, a pull request still uses the branch, or the branch belongs to work still in progress or whose close has waited less than a day; an older pending close is resumed rather than skipped. Temporary files deleted, the processes stopped, and the containers it started (with their volumes) and the images it built removed. A dump or export it made is listed and removed only if you say so. Resources it created outside your machine, such as storage objects or registry images, are listed with the command that removes each, and removed only if you say so. Anything with uncommitted work or uncertain ownership is kept.
- The work's support material handled as you answered: what you chose to commit, archive or move into the documentation is committed with the archived change record, or in a commit of their own on a new branch when there is none, delivered the way the work was; what you chose to keep stays where it is, moved out of a temporary working copy into your main checkout before that copy is removed.
- What the work leaves open on purpose, such as a measurement over later CI runs or a later phase, written as a comment on the ticket and as one `Pending:` record in persistent memory that [Status](/flows/status/) lists. A promotion along your environments is not recorded this way, since the ticket's state already shows it, and one that has not been made keeps no branch, working copy or file; only a promotion that the plan itself lists as a task, because the ticket or your request asked for that environment, is recorded. Each later close updates the records that name its ticket and deletes one with nothing left open.
- The change folder committed in a separate specs repository when the work reaches its completion point, so it never waits for the close to exist in Git. At a pause it is committed too, when that repository declares `Delivery: direct-base`; otherwise it stays uncommitted, is listed as pending and the report recommends declaring `direct-base` there.
- The rows of your project's account index, without any secret value, copied to persistent memory, so a later session or worktree knows which test accounts exist and where their secrets are, as the [test accounts rule](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md) sets.
- One final report of the whole work item: what was integrated, how each acceptance criterion was shown, what was skipped or could not run, local repairs, the work's support material with what was done to it, the findings and what remains.
- A page version of the report through [flow-report](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-report/SKILL.md), offered as an option when you want to keep or share it.

When the work had a viewer session, Close finishes it first; see [Viewer](/viewer/). That only marks the work's record: Close never stops the viewer or closes the window you have open, nor quits a program other sessions may use, such as Docker Desktop or a local database server. When the viewer is not running, Close goes on without it, keeps what the viewer close needs, and the report says which viewer session stayed open.

## How it ends

With the footer. Your reply, including choosing to close, decides what runs next. The exact steps, the severity scale for findings and the branch cleanup rules are in the [flow-close skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md) and its [branch cleanup reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/references/branch-cleanup.md).

## What it does not do

- It does not delete a branch with unmerged commits, a branch just created with no commits of its own, or a working copy with uncommitted files; it keeps them and reports why. It never switches a checkout's branch.
- It does not push, merge or deploy on its own.
- It does not archive a paused change; the folder stays on disk, committed or pending as described above.

## Related flows

- [Verify](/flows/verify/) usually comes right before Close, because it integrates the change; [Build](/flows/build/) does when verify is skipped. Both end with the footer.
- [Deploy](/flows/deploy/) comes after integration. When it runs, only the cleanup of Close follows it.
- [Changes and loops](/flows/overview/) shows where the work goes when your answer adds to it.

## FAQ

### What if I only want to stop for now?
Say you are pausing. The plan stays on disk, committed on the work's branch when its specs live in the same repository and that branch exists, or in a specs repository that declares `direct-base`; otherwise it is listed as pending, and the report lists what is pending.

### Does Close mean the work is shipped?
It means the work reached its completion point, for example merged into the base branch. Promotion to other environments is [Deploy](/flows/deploy/).

Source: [content/skills/flow-close/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)
