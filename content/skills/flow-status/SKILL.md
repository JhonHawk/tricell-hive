---
name: flow-status
description: Report the project's current status and backlog, read-only, and recommend what to take next. Use when the user asks for the project's status, its backlog, unfinished work, or what to work on next, such as "¿qué sigue?" or "status del proyecto"; not for investigating one question or one ticket, which belongs to flow-research.
---

# Report the project's status

This skill is read-only and assesses no ticket in depth; `flow-research` does that once the user picks one.

1. Collect, in parallel:
   - the tracker's tickets in every non-terminal state. Run the listing in the main thread only when the query caps its output at 50 tickets and short fields (identifier, title, state, labels, priority, assignee); give any other to a `hive-read-state` child;
   - unfinished local work: the current branch, uncommitted changes, local branches not merged into the base branch, worktrees, and open pull requests.
2. Group five or more tickets by product module; never a flat list, and never grouped by state, application, repository or layer. When the user asks for another cut, such as the tickets that fit one session or those blocking a release, structure the answer by that cut instead; it replaces the default grouping and counts.
3. Report the unfinished local work apart from the tickets, with what each item still needs.
4. Close by recommending what to take first and why: production bugs, then security or data integrity, then promotion blockers, then debt. List a ticket another session has in progress apart, never as the recommendation.
5. When the user asked what to take next, end with the route question, asked as [hive-core](../hive-core/SKILL.md) sets: each option a ticket or group to take, the recommended one first. Choosing one starts `flow-research` to assess that ticket.
