---
title: Status
description: See where the project stands, what is unfinished, and what to take next, without changing anything.
---

Status answers "where are we?" in one screen. The agent reads your tracker and Git, groups what it finds, and recommends what to take next.

## When it starts

Use Status when you ask for the project's status, its backlog, unfinished work, or what to work on next.

```text wrap
What is the status of the project?
What should I take next?
```

To investigate one question or assess one ticket, use [Research](/flows/research/) instead.

## What it asks you

Nothing. The report is never replaced by a question card, and the flow is read-only.

## What it reads

- The tickets in every non-terminal state, as short summaries.
- Unfinished local work: the current branch, uncommitted changes, branches not yet merged, worktrees, open pull requests, and change folders not yet committed in the specs repository, even when it is a separate one.
- The work left open on purpose: the `Pending:` records [Close](/flows/close/) keeps in persistent memory, each marked stale when its ticket is done and its conditions are met, or its change folder is archived.

## What the report holds

One block, in this order:

1. A summary line: open tickets, how many wait for each promotion, how many are in progress.
2. The tickets grouped by product module or by relation, never as a flat list.
3. The unfinished local work and the work left open on purpose, each item with the action it needs.
4. Numbered recommendations, ordered by priority: production bugs, then security or data integrity, then promotion blockers, then debt. A ticket another session has in progress is never recommended.

## How it ends

A plain status request ends with the recommendations. When you asked what to take next, the report ends with the recommended ticket and the command that assesses it, which starts [Research](/flows/research/) for that ticket. Your reply is what starts it.

Source: [content/skills/flow-status/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-status/SKILL.md)
