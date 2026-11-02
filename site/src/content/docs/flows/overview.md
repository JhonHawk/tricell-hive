---
title: Changes and loops
description: A map of Hive's flows, where each one iterates, where you can ask for a change, and where the work returns afterward.
---

Hive's flows are not a one-way line. Work loops back when a check fails, when a request changes, or when a promotion finds a defect. This page maps the flows and shows, per flow, where that happens. It does not repeat the rules; each row links to the file that holds them.

## The flows

| Flow | What it does |
| --- | --- |
| [Status](/flows/status/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-status/SKILL.md)) | Reports the project's state and backlog, read-only, and recommends what to take next. |
| [Research](/flows/research/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/SKILL.md)) | Investigates a question, a failure or a ticket, and ends with a route for you to choose. |
| [Plan](/flows/plan/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md)) | Resolves the decisions and splits the work into verifiable tasks. |
| [Build](/flows/build/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md)) | Makes the change and hands over a built candidate. |
| [Verify](/flows/verify/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md)) | Checks the candidate with evidence, asks for your validation and integrates it. |
| [Close](/flows/close/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)) | Records the outcome, cleans up and asks what comes next. |
| [Deploy](/flows/deploy/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md)) | Promotes an integrated change through the environments and checks what is served. |

Around them:

- The [viewer](/viewer/) shows the stages, documents and captures of a flow as it runs, and the [flow-walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md) skill publishes to it and presents a built interface for your validation.
- [Unattended delegation](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md) lets work continue while you are away, within the objective and answers you already gave, and stops at anything irreversible.
- [Security review](/flows/security/) is a review you ask for, plus the advisory check.

## Where the work loops

When you ask for something that would change work already open, the agent first says what kind of change it is and where the work goes: a correction, an added task, a new change or a separate topic. That classification lives in the [global guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).

| Flow | Where it iterates | Where the work returns |
| --- | --- | --- |
| Research | Rounds of an iterative investigation | Your request joins the route question at the end; nothing is implemented before you choose a route |
| Plan | Decision rounds and plan revisions | The plan is revised; a change of intent becomes a new change |
| Build | Task by task toward a built candidate, and correction rounds after you reject the first rendered view | Stays in Build while the change is small and understood; otherwise back to Plan |
| Verify | One round of fixes after your validation and the reviews | A correction stays in Verify; an added task goes back to Build, or to Plan when it outgrows a small change |
| Close | None; it ends the work item | An added task goes to Build when small or to Plan otherwise, and a new outcome is planned as its own change |
| Deploy | Environment by environment | A defect goes back to Build, then the run starts again with the fix; a ticket integrated later waits for the next run |

The diagram in the [README](https://github.com/JhonHawk/tricell-hive#a-workflow-with-room-for-judgment) draws the same loops.
