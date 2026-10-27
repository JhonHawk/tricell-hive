---
title: Changes and loops
description: Where each flow iterates, where you can ask for a change, and where the work returns afterward.
---

Hive's flows are not a one-way line. Work loops back when a check fails, when a request changes, or when a promotion finds a defect. This page shows, per flow, where that happens. It does not repeat the rules; each row links to the file that holds them.

When you ask for something that would change work already open, the agent first says what kind of change it is and where the work goes. The classification (correction, added task, new change or separate topic) lives in the [scope changes reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/scope-changes.md).

| Flow | Where it iterates | Where you can ask for a change | Where the work returns |
| --- | --- | --- | --- |
| [Research](/flows/research/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/SKILL.md)) | Rounds of an iterative investigation | Any time while it runs; the agent classifies your request, and a finding can also change the objective or scope | Your request joins the options of the route question at the end; nothing is implemented before you choose a route |
| [Plan](/flows/plan/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md)) | Decision rounds and plan revisions | Any time the plan is open | The plan is revised; a change of intent becomes a new change |
| [Build](/flows/build/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md)) | Task by task toward a built candidate, and correction rounds after you reject the first rendered view | When you add to the request, at the first rendered view, or when the work finds it needs more | Stays in Build while small and understood; otherwise back to Plan |
| [Verify](/flows/verify/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md)) | Correction rounds during your validation pause | At the validation pause | A correction stays in Verify as one more round; an added task goes back to Build, or to Plan when it outgrows a small change; after repeated corrections the agent may propose replanning |
| [Close](/flows/close/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/SKILL.md)) | None; it ends the work item | In your answer to the close question, or with a requirement after integration | A correction or defect fix ships as a follow-up of the same delivery; the record still closes, then an added task goes to Build when it is small or to Plan otherwise, and a new outcome is planned as its own change |
| [Deploy](/flows/deploy/) ([skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/SKILL.md)) | Environment by environment | When you ask to add or remove tickets during a run | A defect goes back to Build; a ticket integrated later waits for the next run; to remove a ticket already on the base branch, the run waits while the agent asks whether to keep it, and the run continues, or to revert it as its own change, and then the run stops and starts over |

The diagram in the [README](https://github.com/JhonHawk/tricell-hive#a-workflow-with-room-for-judgment) draws the same loops.
