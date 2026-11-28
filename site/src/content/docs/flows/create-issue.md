---
title: Create issue
description: Write a tracker ticket that someone who never saw your session can work, break larger work down into a parent ticket with children or into phases of a project, or check whether an existing ticket is ready.
---

Create issue writes the ticket for work that will be done later, by someone else, or in another session. A ticket states what is needed, why, and how anyone will know it is done; [Plan](/flows/plan/) still settles how it is built. The skill covers three uses: one ticket, a set of children under a parent ticket or inside a project, and a check of a ticket that already exists. It runs in the main conversation, since helper agents never write to a tracker, and it changes no file of your repository.

## When to use

Ask for a ticket, an issue, an epic, or a breakdown of work into tickets. Ask whether a ticket is ready, or to review or complete one. The skill also starts when you answer that a finding or a separate topic becomes a ticket, in [Hive mode](/flows/hive-mode/) or at the [close](/flows/close/). The agent proposes a ticket only for work that will not be done in the current session or with its context; work it can finish now is fixed instead.

It is not the way a ticket changes state, receives the work's pending-work comment, or receives its business rules. Those writes belong to the flow that causes them, and they are unchanged; so is the comment on a neighbor ticket whose work a decision moved, which [Build](/flows/build/) and [Close](/flows/close/) write.

Example requests:

```text wrap
Create a ticket for the duplicate invoice emails we found.
Break the export feature down into tickets we can assign.
Is ticket 214 ready for someone to pick up?
```

## What it asks

- Whether to create the ticket, once the draft is shown, when you used the skill on its own. When your answer already chose the ticket inside a flow, such as a finding answered "ticket", nothing more is asked.
- Only what changes the ticket and no read of the code or the tracker answers, each with a recommended answer. Every question that does not wait on another answer goes in the same round, in consecutive cards when they do not fit in one. What the agent does not ask is written in the ticket as an assumption.
- Which tracker and repository, only in a session without a repository, where no `## Hive` section declares a tracker.
- One approval for the whole set, when the work is broken down into a parent ticket with children. The agent shows every ticket before it writes any, with each milestone or label it would create, the child that holds each capability several children need, and which requirement rules each child covers.
- To confirm the comment or the edit it proposes on a ticket that already exists, before anything is posted.

## What it produces

- One ticket in the project's primary tracker, of one of four types (story, bug, task or spike), with the sections of its type, written in the session language unless the project sets another. When the repository has an issue template, the ticket fills it. The ticket carries no statement that an agent wrote it, unless the repository's contribution rules ask for one.
- Before the draft, a search of every tracker the project declares for a ticket with the same outcome, open or closed. When an open one exists, the agent reports it with its link and state and creates no other. When only a closed one exists and the work is still needed, as when a fixed defect returns, the new ticket is related to it.
- A check of the draft against a fixed list of eight readiness tests (backed by `check_issue.py`), run in the same conversation, before anything is written. A ticket that lacks something nobody in the session can supply, such as an undecided business rule or a bug's steps to reproduce, is still created, marked blocked, with the question and who it waits on, and is reported as blocked, never as ready. The agent never invents the missing part.
- For a breakdown, a parent ticket and its children, each of which is integrated and accepted on its own, written one at a time with the blocking relations the children need. A capability that several children need, such as sending email, goes in the first child that uses it, and the others are blocked by that child. When you ask for phases, or the set has more than seven children, the phases become milestones of the tracker (a version in Jira), created only if the approved set lists them, and in Linear a project you name takes the parent's place. When the set comes from a requirement document whose rules carry codes, each child names the codes it covers in a `Covers:` line, and a check (`check_coverage.py`) reports, before the approval, each rule no child covers, each rule several cite, each cited code the document lacks or cancelled, each line it cannot read, with the reason, and each child without a `Covers:` line. The agent stops at the first failed write and reports what exists. Where work is [unattended](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md), it creates no set and reports it as unresolved.
- A read-back of every ticket after it is written, compared with what was sent, since a tracker can drop a field without an error.

The procedure for each part lives in the skill's references: [ticket types](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-issue/references/ticket-types.md), [breakdown](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-issue/references/breakdown.md), [project structure](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-issue/references/project-structure.md), [trackers](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-issue/references/trackers.md) and [existing ticket](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-issue/references/existing-ticket.md).

## How it ends

Used on its own, it reports each ticket with its link, whether it is ready or blocked, and the assumptions it carries, then offers the next step as a reply, such as researching or planning the ticket. Creating a ticket starts no work on it. Followed from [Hive mode](/flows/hive-mode/) or [Close](/flows/close/), it returns to that flow at its next step.

## What it does not do

- It does not write a ticket without your answer, except one ticket for a finding while work runs unattended.
- It does not move a ticket's state, comment the work's pending items or write its business rules.
- It does not start Plan or Build for the ticket it creates.
- It does not edit your repository, and it creates no project, initiative, team or state: only the milestones and labels of an approved set.
- It does not audit, group or deduplicate the tickets already open; the [audit-issues](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/audit-issues/SKILL.md) skill does.

## Related flows

- [Hive mode](/flows/hive-mode/) checks for an existing ticket before it proposes one, and follows the skill once you answer that a finding becomes a ticket.
- [Close](/flows/close/) asks, for each defect the work found but did not fix, whether to open a ticket, and follows the skill when you answer yes.
- The [audit-issues](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/audit-issues/SKILL.md) skill audits the open tickets and uses this skill's parent tickets and their approval for the groups it forms.
- [Changes and loops](/flows/overview/) maps where each flow returns the work.

## FAQ

### Do I need a ticket to use Hive?
No. No flow needs one. The skill runs only when you ask for a ticket or answer that something becomes one.

### Can it check a ticket someone else wrote?
Yes. It reads the ticket against the same readiness list and reports what is missing. It posts or edits only after you confirm: a comment when your account is not the ticket's author, an edit when it is, and a report in the message when the account cannot write there.

### Where is the exact rule?
Source: [content/skills/create-issue/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-issue/SKILL.md)
