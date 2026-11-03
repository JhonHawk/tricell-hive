---
title: Changes and loops
description: A map of Hive's flows, where each one iterates, where you can ask for a change, and where the work returns afterward.
---

Hive's flows are not a one-way line. Work loops back when a check fails, when a request changes, or when a promotion finds a defect. This page maps the flows and shows, per flow, where that happens. It does not repeat the rules; each row links to the file that holds them.

## The flows

| Flow | What it does |
| --- | --- |
| [Status](/flows/status/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-status/SKILL.md)) | Reports the project's state and backlog, read-only, and recommends what to take next. |
| [Research](/flows/research/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/SKILL.md)) | The entry for every request to do work: settles what you need, then answers, continues into Build, or goes to Plan; an analysis ends at its findings. |
| [Plan](/flows/plan/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md)) | Settles the how: presents approaches, resolves the decisions and splits the work into verifiable tasks. |
| [Build](/flows/build/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md)) | Makes the change and hands over a built candidate. |
| [Verify](/flows/verify/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md)) | Checks the candidate with evidence, asks for your validation and integrates it. |
| [Close](/flows/close/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)) | Moves the ticket and reports, then ends with a footer; the record archive and cleanup run on your reply. |
| [Deploy](/flows/deploy/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md)) | Promotes an integrated change through the environments and checks what is served. |

Around them:

- The [viewer](/viewer/) shows the stages, documents and captures of a work from Plan or Build on, and the [flow-walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md) skill publishes to it and presents a built interface for your validation.
- [Unattended delegation](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md) lets work continue while you are away, within the objective and answers you already gave, and stops at anything irreversible.
- [Security review](/flows/security/) is a review you ask for, plus the advisory check.

## Where the work loops

When you ask for something that would change work already open, the agent first says what kind of change it is and where the work goes: a correction, an added task, a new change or a separate topic. That classification lives in the [global guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).

| Flow | Where it iterates | Where the work returns |
| --- | --- | --- |
| Research | One open question per message, and rounds of investigation | Your answer refines what is wanted; a complete request continues into Build in the same turn, and a change of the how that needs your decision goes to Plan |
| Plan | Decision rounds and plan revisions | The plan is revised; a change of intent becomes a new change |
| Build | Task by task toward a built candidate, and correction rounds after you reject the first rendered view | Stays in Build while the how leaves you no decision; otherwise back to Plan |
| Verify | One round of fixes after your validation and the reviews | A correction stays in Verify; an added task goes back to Build, or to Plan when it brings a decision of the how for you |
| Close | None; it ends the work item | A reply to the footer can send findings to Build; an added task goes to Build when no decision of yours is open or to Plan otherwise, and a new outcome is planned as its own change |
| Deploy | Environment by environment | A defect goes back to Build, then the run starts again with the fix; a ticket integrated later waits for the next run |

The diagram in the [README](https://github.com/JhonHawk/tricell-hive#a-workflow-with-room-for-judgment) draws the same loops.
