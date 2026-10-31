---
name: flow-status
description: Report the project's current status and backlog, read-only, and recommend what to take next. Use when the user asks for the project's status, its backlog, unfinished work, or what to work on next, such as "¿qué sigue?" or "status del proyecto"; not for investigating one question or one ticket, which belongs to flow-research.
---

# Report the project's status

This skill is read-only and assesses no ticket in depth; `flow-research` does that once the user picks one.

1. Collect, in parallel:
   - the tracker's tickets in every non-terminal state. Run the listing in the main thread only when the query caps its output at 50 tickets and short fields (identifier, title, state, labels, priority, assignee); give any other to a `hive-read-state` child;
   - unfinished local work: the current branch, uncommitted changes, local branches not merged into the base branch, worktrees, and open pull requests.
2. Write the report as prose in the message, as [Report](#report) sets. Never replace it with a question card or a table.
3. When the user asked what to take next, end with the route question after the report, asked as [hive-core](../hive-core/SKILL.md) sets: each option a ticket or group to take, the recommended one first. Choosing one starts `flow-research` to assess that ticket. A plain status request ends with the recommendations, without a question.

## Report

In this order:

1. One line per non-terminal state with its count, in progress first, and the delivery state: the work integrated in the base branch and not yet promoted.
2. The tickets grouped by product module or by relation, taking modules from tracker labels or projects, otherwise from titles. Keep an initiative's umbrella and its phases together and in order. One line per ticket: its ID, its title, and its state as a suffix. Then the cross-cutting work (infrastructure, CI, tests, tooling, debt) grouped by type, the deferred tickets apart, and a ticket another session has in progress apart. Never a flat list, and never grouped by state, application, repository or layer. When the user asks for another cut, such as the tickets that fit one session or those blocking a release, group by that cut instead.
3. The unfinished local work, with what each item still needs.
4. The recommendations: what to take first and why, in this order of priority: production bugs, then security or data integrity, then promotion blockers, then debt; among equals, the smallest and most visible first. Never recommend a ticket another session has in progress.
