---
title: Plan
description: Turn a larger change into settled decisions and small tasks that each come with a way to check them.
---

Plan settles the how. [Research](/flows/research/) has already settled what you need; Plan sorts out the remaining choices with you first, so whoever builds it does not have to guess.

## When it starts

Research sends a request here only when its how has a decision that is yours, or when the work cannot be finished from your request and the `git diff` alone, such as work handed to another session or agent. You can also ask for a plan directly. A change whose how leaves you no decision needs no plan; Research continues it into [Build](/flows/build/).

Example requests:

```text wrap
Plan the API change, including compatibility and acceptance checks.
Plan moving our uploads to the new storage service. Ask me before choosing anything that affects existing data.
Break this feature into tasks I can review one at a time.
```

## What it asks you

The agent reads the code, your project settings and the existing specs first, then asks only for the decisions that are yours. Facts it can find, it looks up itself.

- When the how has more than one sensible way, two or three approaches in a message of their own before the tasks, each with its tradeoff and one recommended; you choose, and they are recorded in `design.md`. When only one way is sensible, the agent says so and why.
- Choices that change the outcome, asked in rounds and in the shared question format. A change to what an interface sends or receives counts as a contract change, and you decide whether it stays compatible.
- For a changed screen, the layout and its states, shown as before-and-after sketches you can choose between.
- Once the plan is written, how the work is delivered and which code review runs, unless you already answered. The recommended delivery follows the kind of work: automatic when its checks, review and verification can judge it, such as infrastructure or CI; interactive when it has a screen, copy or flow only you can judge. A one-off operation that leaves nothing in the repository, such as a database dump, gets neither question.

The rules for questions and for delivery are in the [shared guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).

## What it produces

Reviewable plan documents are published on demand from their retained source folder. The binary uploads the Markdown and returns a confirmed revision; requested edits publish another revision of the same work. Source files stay in the repository. See the [viewer publication model](/viewer/#publish-markdown-on-demand).

A plan stays in the message only when the work fits one session on this machine: it spans no other session, waits on nothing external and is not handed to helpers. Such a plan carries its decisions, criteria, prerequisites and tasks in the message, and the agent keeps a checkpoint file in the host's temporary folder with the decisions, the pending tasks and anything it created outside the repository, so a compacted session can pick up where it left off. The file is removed at Close. The agent writes a change folder in your specs directory when the plan fails that test, such as work that spans sessions or is handed to helpers, when you ask for one, when its deliveries need coordination, or when it changes a contract, stored data, security, money, concurrency or a dependency. A one-off operation always keeps its plan in the message; when one of those conditions applies, or you agree to add files to the repository for it, it is handled as a regular change. A plan in the message that changes a behavior your specs describe includes a task that updates them. Without a ticket, the folder is named by a short slug; a plan never waits for a ticket to exist.

The change folder is copied from the [templates](https://github.com/JhonHawk/tricell-hive/tree/master/content/skills/flow-plan/assets/change-folder):

- `proposal.md`: your request quoted, your decisions in plain words, the scope, numbered acceptance criteria, and what the build needs from outside the repository.
- `tasks.md`: small tasks, each saying which criteria it closes, what it depends on and how to check it.
- `design.md` and spec deltas, only when the change has design decisions or alters a behavior, contract, permission or stored data.

Every criterion is one observable behavior, with the observation that would show it false today. A criterion that is already true, or that only restates your request, is rewritten or dropped. When the plan has a change folder, one independent reviewer checks it before it is ready. The [viewer](/viewer/) session for the work starts here, and a running viewer can show the plan documents, as the [walkthrough skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md) describes.

A change folder stays uncommitted while the work runs; [Close](/flows/close/) versions it at the end.

## How it ends

The agent summarizes the plan in the message: the objective, each criterion, your decisions, the tasks on one line each, what is still missing, the checks and the delivery mode. It then gives the Build command as the next step, and offers to compact the session first, with what to keep, so the build starts without the research and planning history; after you compact, it shows the same next step again. A plan does not authorize implementation: typing that command is what approves the criteria shown and starts [Build](/flows/build/). A correction to a criterion is not approval; the agent revises and shows it again. The steps are in the [plan handoff reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/plan-handoff.md).

## What it does not do

- It does not write the implementation, or approve commits, pushes, pull requests, merges or deployments.
- It does not hide open technical doubts inside later tasks.

## Related flows

- Start from [Research](/flows/research/) when you still need facts before planning.
- Hand the finished plan to [Build](/flows/build/) by pointing it at the change folder.
- [Close](/flows/close/) archives the change folder when the work is integrated.
- If something new comes up while the plan is open, the agent classifies it first; see [changes and loops](/flows/overview/).

## FAQ

### Do I need a plan for a small change?
No. A change whose how leaves you no decision goes from Research to Build in the same turn.

### Does a ready plan mean the agent may start building?
No. Your reply starts Build, unless you already authorized it to proceed.

### Where is the exact rule?
Source: [content/skills/flow-plan/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md)
