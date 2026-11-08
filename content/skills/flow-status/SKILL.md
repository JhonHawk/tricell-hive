---
name: flow-status
description: Report the project's current status and backlog, read-only, and recommend what to take next. Use when the user asks for the project's status, its backlog, unfinished work, or what to work on next, such as "¿qué sigue?" or "status del proyecto"; not for investigating one question or one ticket, which belongs to flow-research.
---

# Report the project's status

This skill is read-only, except for the `HIVE.local.md` entries the global guidance's «Local instructions» authorizes, and assesses no ticket in depth; `flow-research` does that once the user picks one. Read only what the report needs: the `## Hive` section of the repository, the tracker, Git and the pending-work records in persistent memory.

1. Collect, in parallel:
   - each declared tracker's tickets in every non-terminal state, in the main thread, with queries that cap their output at 50 tickets and return only short fields (identifier, title, state, labels, priority, assignee, parent). Use `hive-read-state` children only to run several independent collections at once, such as more than one tracker or project, as [delegation](../flow-build/references/delegation.md) sets;
   - unfinished local work: the current branch, uncommitted changes, local branches not merged into the base branch, worktrees, open pull requests, and the change folders under `Specs` with work outside Git, read with `git status --short --untracked-files=normal -- <Specs>/changes` in each checkout and worktree of the repository that holds it, where a `??` entry ending in `/` directly under `changes/` is a folder never committed and any other entry an uncommitted edit;
   - the work left open on purpose: the `Pending:` records [flow-close](../flow-close/SKILL.md#pending-work) writes, searched in the project the global guidance's «Persistent memory» names, with an explicit result limit. A record is ready to resume when every ticket and condition that unblocks it is resolved; flow-status never deletes one, which only a close does.
2. Write the report as [Report](#report) sets. Never replace it with a question card.
3. When the user asked what to take next, end the report with the route as the global guidance's «Questions and routes» section sets: the recommended ticket or group and the command that assesses it, such as `/flow-research COM-833` in the host's syntax. The ticket the user names starts `flow-research`. A plain status request ends with the recommendations.

## Report

Write it in one code block so its columns align, in this order:

1. One summary line: the open tickets, how many wait for each environment's promotion, and how many are in progress. No pull request numbers, commits or dates.
2. The tickets grouped by product module or by relation, taking modules from tracker labels or projects, otherwise from titles; head each group with its name and its umbrella ticket. One line per ticket: the state first, padded so the column aligns, then the identifier, then a short title of three to six words, and only when it matters a blocker, the priority or the assignee. Keep an initiative's phases together and in order; then the cross-cutting work (infrastructure, CI, tests, tooling, debt) grouped by type, the deferred tickets apart, and a ticket another session has in progress apart. Never a flat list, and never grouped by state, application, repository or layer. When the user asks for another cut, such as the tickets that fit one session or those blocking a release, group by that cut instead.
3. The unfinished local work and the work left open on purpose, one line per item with the action it needs; a record ready to resume says so, with the evidence.
4. The recommendations, numbered, one line each with its reason, in this order of priority: production bugs, then security or data integrity, then promotion blockers, then debt; among equals, the ticket with the fewest tasks first, then the one with a user-visible outcome. Never recommend a ticket another session has in progress.

```
13 open · 4 waiting for QA · 10 waiting for production

Meta self-service (COM-826)
  To Deploy  COM-829  Number lifecycle
  Backlog    COM-832  Embedded Signup v4 · blocked by COM-835

Local work
  • 7 stray package lockfiles → delete

Recommendation
  1. Promote development → qa: closes COM-829.
```
