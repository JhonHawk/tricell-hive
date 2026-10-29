---
title: Close
description: Wrap up a finished, paused or abandoned piece of work and clean up only what that work created.
---

Close is the last step of a piece of work. It makes sure the records match reality, removes the leftovers the task made, and tells you where things stand.

Its reporting reference applies to findings, relayed results, explanations and
messages with external links or IDs. Greetings and
short direct answers need no reference read.
Questions and artifact cleanup have their own direct reference links. Those
procedures retain their existing obligations; reading a reference does not
invoke another complete flow.

## When to use

Use Close when work reaches its completion point, when you abandon it, when you pause it, or when you end the session or ask for a cleanup. The completion point is usually the merge into the base branch, after any checks that must pass there; under `hold` it is the verified working tree, and when the delivery hands the open pull request to a human reviewer, it is opening that pull request, and a later session that finds it merged closes the item then, at its next stopping point, as if it had merged it. An iterative research route runs Close when it stops, as Research sets. A validation stop or a question that waits is not a pause and does not run Close. When you ask for the merge of your work directly in conversation, it reaches that point too, so Close runs in full rather than only cleaning up branches. A project can declare a later completion point in the `Completion` key of its `## Hive` section; the work still closes at integration, as [project settings](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/project-settings.md#later-completion-point) sets. Ending the session with work still open closes it as a pause. A mechanical change with no change folder and nothing to list skips the record and the findings, and usually the closing question. Under unattended delegation the closing question is not asked: the report names the suggested next ticket, what stays open and the findings left pending. After a promotion only the branch and process cleanup runs, with no completion report or closing question.

Example requests:

```text wrap
Close this work item and clean up only the temporary resources it created.
Pause this work. Leave the plan on disk and tell me what is still open.
Clean up the session.
```

## What it produces

The agent goes through these steps in order:

For work with a walkthrough, it first records its known PR or commit delivery link
when available, preserving the captured candidate and other works in the session.
Then it marks only that work closed and publishes its JSON, or standalone HTML
when the binary is absent, before removing a working copy or review artifacts. Closed also covers
abandoned or paused work; cleanup by itself does not close work. The existing
viewer must confirm that it collected the exact current source and resources
before their source unit can be removed. If the viewer is offline, too old, has
no matching copy or does not confirm within 15 seconds, the agent keeps the
still-needed unit and reports its folder. It never starts or restarts the viewer
to force confirmation. The procedure lives in
[Close the current walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/references/walkthrough-close.md).

1. When the work is complete, confirms the ticket is in the state the merge moved it to, along with the sub-issues the same work covered, or moves it now. When your project declares a later completion point in `Completion`, that state is the intermediate one the key names. Under `hold`, or while a pull request waits for a human reviewer, the ticket stays in progress and the report gives the missing delivery or merge as a reminder. A paused or abandoned ticket does not move to finished.
2. Handles the change record, if there is one: after integration and required post-merge checks it folds the plan's requirement changes into the project's current requirements, confirms that each one matches what the plan said, marks the change completed and archives the folder. Under `hold` or while a pull request waits for a human reviewer, the record stays on disk without applying its requirements or archiving it. An abandoned change is archived without touching the requirements, with its unapplied changes kept apart so nothing reads them as applied.
3. Checks merged branches and their working copies, returns to the base branch and removes eligible clean working copies before deleting their branches. The order and safety checks live in [branch cleanup](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/references/branch-cleanup.md).
4. Removes temporary files the task made, stops processes it or its helpers started, and puts back local things the task changed that you rely on, such as your development containers or database schema, instead of leaving them to you.
5. Lists side findings that the running system can actually trigger, after searching the tracker and its memory for duplicates and for findings you already declined or asked not to hear about, recommends resolving in this work those whose fix is small and sits in the area the work already touches, and a ticket only for the rest, and lets you resolve each one in this work (naming Build for a small, understood change or Plan otherwise), open a ticket for it, decline it, or ask not to hear about it again. It asks about every finding it lists, over several question cards when they do not fit in one, and never declines one on your behalf. A suspicion it could not trace that far stays in the report as unconfirmed, and so does a finding about the interface that it saw only in the code. A minor finding of the lowest priority (P3) also stays in the report only. It groups P2 findings that share an area or a cause into one proposal with one answer, and it proposes a finding about the interface only after it saw the problem rendered. The rules live in [Incidental findings](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/references/incidental-findings.md).
6. Records in memory that the worktrees it removed are gone and the recurrence counts of declined findings, in one write.
7. Before the report, waits for the background work whose result the report states, such as a CI watcher, and stops the rest, so nothing speaks up after it; a late notice that changes nothing gets no message. Then writes a completion report, whole in one message even when a question or a push came during the close, that says what was done and what it excludes, each check not met, skipped or accepted without verification, any unresolved uncertainty, and what stays open; it adds the decisions it took without asking, with how to revert each, the first view's path when you were not shown it, records left outside this machine, and cleanup that failed or left something behind, such as a kept branch, only when there are any. It lists checks that passed only for a mechanical change, as the checks that cover it.
8. Asks what to do next: continue with a suggested ticket or the rest of the plan (it does not suggest a ticket it has just opened from that work's findings, unless the finding is critical or blocks the next step), take it to a new session, or end. The continue options give the estimated size of the suggested ticket once, or the remaining tasks of the plan, as the number of tasks and what that count rests on, with no time estimate. While a plan is still in progress, continuing covers all its remaining tasks and asks once for the permissions they still need, grouped by effect class (a critical effect, such as spending money or deleting data, is asked on its own); the agent does not suggest a new session then.

## What it does not do

- It does not delete a branch that has unmerged commits. It shows you that work and asks. A branch merged by squash or rebase counts as merged when its pull request shows it merged from the same last commit, so it is deleted with its clean working copy instead of being left behind.
- It does not delete a working copy that has uncommitted or untracked files.
- It does not remove material whose owner or purpose is unclear.
- It does not archive a paused change. The folder stays on disk and is reported as pending.
- It does not push, merge or deploy on its own. Versioning the change record follows the delivery choice you already made.

## Related flows

- [Verify](/flows/verify/) usually comes right before Close, because it integrates the change; [Build](/flows/build/) does when verify is skipped.
- [Deploy](/flows/deploy/) is a separate stage after integration. It never holds a close open: when your project declares a later completion point in `Completion`, Deploy moves the ticket to its final state once that condition holds and no functional walk is pending in that environment.
- [Changes and loops](/flows/overview/) shows where the work goes when your answer to the close question, or a requirement after integration, adds to it.
- [Plan](/flows/plan/) created the change record that Close archives.
- [Research](/flows/research/) is separate. A research-only session ends with its route question instead.

## FAQ

### What does Close delete?
Temporary files and processes the task or its helpers created, merged branches, including ones from earlier work, with their clean working copies, and, at the end of the session, the interface screenshots, flow recordings and walkthrough manifest and optional page that you did not ask to keep: keeping the manifest or page keeps its referenced images and recordings too; keeping only individual images or recordings drops the remaining unit. Paused work that continues in another session keeps them all, as does an unresolved confirmation of the viewer's current closed copy. Branches are deleted only when their pull request is yours and merged, and a remote branch only when its head is the one that was merged or already in the base branch, so commits pushed afterwards are not lost. A branch with unmerged commits is never deleted, and anything whose owner is unclear is kept.

### Does Close mean the work is shipped?
It means the work reached its completion point, for example merged into the base branch. Later promotion to other environments is tracked separately; if your project names a later point as done in `Completion`, the ticket sits in an intermediate state until it holds and no functional walk is pending there.

### What if I only want to stop for now?
Say you are pausing. The plan stays where it is, and the report lists what is pending and who owns it.

### Do I have to answer the closing question?
It is how the agent offers the next route. The message opens with a bold context line, stage · ticket · progress, with the running flow's untranslated name as the stage, such as Close or Research, so you can tell which piece of work is asking; a question card opens its question with the stage and ticket. The options are in a question card when your tool shows one that waits for your answer, and in the message otherwise, as the [questions reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/questions.md) sets. A question whose options depend on your answers to the findings is asked in a later turn, after them. A dismissed or timed-out card is not an answer, so the agent asks again in text. Choosing to end the session is a valid answer.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-close/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)
