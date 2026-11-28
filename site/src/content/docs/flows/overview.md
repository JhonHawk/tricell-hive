---
title: Changes and loops
description: A map of Hive's flows, where each one iterates, where you can ask for a change, and where the work returns afterward.
---

Hive's flows are not a one-way line. Work loops back when a check fails, when a request changes, or when a promotion finds a defect. This page maps the flows and shows, per flow, where that happens. It does not repeat the rules; each row links to the file that holds them.

## The flows

| Flow | What it does |
| --- | --- |
| [Status](/flows/status/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-mode/SKILL.md)) | Reports the project's state and backlog, read-only, and recommends what to take next. |
| [Hive mode](/flows/hive-mode/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-mode/SKILL.md)) | The entry for every request to do work: settles what you need, then answers, continues into Build, or goes to Plan; an analysis ends at its findings. |
| [Create design](/flows/create-design/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-design/SKILL.md)) | Optional, between Hive mode and Plan: draws one example of a new screen in your Figma file, following your existing screens and reviewed before you see it, has you approve or correct it, and records the approved example, with a rule card beside each frame. |
| [Plan](/flows/plan/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md)) | Settles the how: presents approaches, resolves the decisions and splits the work into verifiable tasks. |
| [Build](/flows/build/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md)) | Makes the change and hands over a built candidate. |
| [Verify](/flows/verify/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md)) | Checks the candidate with evidence, asks for your validation and integrates it. |
| [Close](/flows/close/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)) | At integration moves the ticket and writes the work's business rules to it, then waits until you type `/flow-close`, ask for the close or end the session (unattended work closes on its own); the close then writes the one final report and ends as the close's situation sets: with a footer when you can still reply, with the report alone for a deploy's cleanup, abandoned work, a hygiene request, the session's end or unattended work; the cleanup runs on your reply, and with nothing left to decide or clean up it runs at once and the footer drops closing. Change records wait for a hygiene request, which archives them together. |
| [Create issue](/flows/create-issue/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-issue/SKILL.md)) | Writes a ticket someone else can work, breaks larger work down into a parent ticket with children or into phases of a project, or checks that an existing ticket is ready; Hive mode and Close follow it when a finding becomes a ticket. |
| [Create PRD](/flows/create-prd/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/create-prd/SKILL.md)) | Writes the requirement document of a module or feature, restructures one someone else wrote, or reviews it against the code; Hive mode hands such a request over. |
| [Deploy](/flows/deploy/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md)) | Promotes an integrated change through the environments and checks what is served. |

Around them:

- The [viewer](/viewer/) shows the UI captures of a work and the iteration each belongs to. Only Build and Verify publish to it, when they capture an interface; Plan, Hive mode and any flow without captures publish nothing. The [walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/ui/walkthrough.md) skill publishes the captures and presents a built interface for your validation. Hive's host hooks add the session's activity, and the host's task list where it has one, as [host-observed activity](/viewer/#host-observed-activity).
- [E2E coordination](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/references/e2e-coordination.md) is an explicitly delegated mode of the existing flows: the coordinator leads agreed deliverables through the authorized endpoint. You do not have to be away, and it does not by itself activate unattended delegation.
- [Swarm mode](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/swarm/SKILL.md) is a mode you start by its command or by naming it, for two to four tickets at once: related tickets are proposed as groups, and one `hive-ticket-owner` child per group researches, plans and builds in its own worktree, stopping after the plan and after the build, while the main thread asks you the decisions of every group together and verifies and publishes each group on its own pull request. It is not unattended delegation and your silence does not extend it.
- [Unattended delegation](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md) lets work continue while you are away, within the objective and answers you already gave, and stops at anything irreversible. It keeps a local decision log, one line per decision and per external effect with how to revert it, which the final report cites by path; the format is in the [skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md#decision-log).
- [Security review](/flows/security/) is recommended for a change that touches a security boundary and runs when you accept or ask for it, plus the advisory check.

## Where the work loops

When you ask for something that would change work already open, the agent first says what kind of change it is and where the work goes: a correction, an added task, a new change or a separate topic. That classification lives in the [scope changes rule](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/plan/scope-changes.md).

| Flow | Where it iterates | Where the work returns |
| --- | --- | --- |
| Hive mode | One open question per message, rounds of investigation, and, for a new or reorganized screen, a recommendation of the Create design stage in its next-step block | Your answer refines what is wanted; a complete request continues into Build in the same turn, a change of the how that needs your decision goes to Plan, and a new screen, or one whose composition is open, can go through Create design first |
| Create design | The example redrawn after each correction you give | You approve each screen's example, which goes to Plan, or to Build for a small settled change, as the screen's reference; a correction stays in the stage |
| Plan | Decision rounds and plan revisions | The plan is revised; a change of intent becomes a new change |
| Build | Task by task toward a built candidate, and correction rounds after you reject the first rendered view | Stays in Build while the how leaves you no decision; otherwise back to Plan |
| Verify | One round of fixes after your validation and the reviews | A correction stays in Verify; an added task goes back to Build, or to Plan when it brings a decision of the how for you |
| Close | None; it ends the work item | A reply to the footer can send findings to Build; an added task goes to Build when no decision of yours is open or to Plan otherwise, and a new outcome is planned as its own change |
| Deploy | Environment by environment | A defect goes back to Build, then the run starts again with the fix; a ticket integrated later waits for the next run |

The diagram in the [README](https://github.com/JhonHawk/tricell-hive#a-workflow-with-room-for-judgment) draws the same loops.
