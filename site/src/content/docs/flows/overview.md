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
| [Figma design](/flows/figma-design/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-figma-design/SKILL.md)) | Optional, between Research and Plan: draws one example of a new screen in your Figma file, following your existing screens and reviewed before you see it, has you approve or correct it, and records the approved example, with a rule card beside each frame. |
| [Plan](/flows/plan/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md)) | Settles the how: presents approaches, resolves the decisions and splits the work into verifiable tasks. |
| [Build](/flows/build/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md)) | Makes the change and hands over a built candidate. |
| [Verify](/flows/verify/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md)) | Checks the candidate with evidence, asks for your validation and integrates it. |
| [Close](/flows/close/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)) | Moves the ticket, writes the work's business rules to it and reports, then ends with a footer; the record archive and cleanup run on your reply. |
| [Deploy](/flows/deploy/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md)) | Promotes an integrated change through the environments and checks what is served. |

Around them:

- The [viewer](/viewer/) shows the UI captures of a work and the iteration each belongs to. Only Build and Verify publish to it, when they capture an interface; Plan, Research and any flow without captures publish nothing. The [flow-walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md) skill publishes the captures and presents a built interface for your validation. Hive's host hooks add the session's activity, and the host's task list where it has one, as [host-observed activity](/viewer/#host-observed-activity).
- [Unattended delegation](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md) lets work continue while you are away, within the objective and answers you already gave, and stops at anything irreversible.
- [Security review](/flows/security/) is a review you ask for, plus the advisory check.

## Where the work loops

When you ask for something that would change work already open, the agent first says what kind of change it is and where the work goes: a correction, an added task, a new change or a separate topic. That classification lives in the [global guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).

| Flow | Where it iterates | Where the work returns |
| --- | --- | --- |
| Research | One open question per message, rounds of investigation, and, for a new or reorganized screen, a recommendation of the Figma design stage in its next-step block | Your answer refines what is wanted; a complete request continues into Build in the same turn, a change of the how that needs your decision goes to Plan, and a new screen, or one whose composition is open, can go through Figma design first |
| Figma design | The example redrawn after each correction you give | You approve each screen's example, which goes to Plan, or to Build for a small settled change, as the screen's reference; a correction stays in the stage |
| Plan | Decision rounds and plan revisions | The plan is revised; a change of intent becomes a new change |
| Build | Task by task toward a built candidate, and correction rounds after you reject the first rendered view | Stays in Build while the how leaves you no decision; otherwise back to Plan |
| Verify | One round of fixes after your validation and the reviews | A correction stays in Verify; an added task goes back to Build, or to Plan when it brings a decision of the how for you |
| Close | None; it ends the work item | A reply to the footer can send findings to Build; an added task goes to Build when no decision of yours is open or to Plan otherwise, and a new outcome is planned as its own change |
| Deploy | Environment by environment | A defect goes back to Build, then the run starts again with the fix; a ticket integrated later waits for the next run |

The diagram in the [README](https://github.com/JhonHawk/tricell-hive#a-workflow-with-room-for-judgment) draws the same loops.
