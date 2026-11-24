# Playbook: ticket-status

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the user asks for the project's status, its backlog, unfinished work or what to work on next, such as "¿qué sigue?" or "status del proyecto". This report replaces the router's steps 0 to 5 and ends as step 3 sets; investigating one question or one ticket is the `investigate` playbook.

**Stages.** `hive-mode` alone, read-only except for the writes [local write](../../hive-core/local/local-write.md) authorizes. It assesses no ticket in depth: choosing a ticket from the recommendations starts `hive-mode` for that ticket, and the user's reply is what starts it. Read only what the report needs: the `## Hive` section of the repository ([project settings](../../hive-core/plan/project-settings.md)), the tracker, Git and the pending-work records in persistent memory.

1. Collect, in parallel:
   - each declared tracker's tickets in every non-terminal state, in the main thread, with queries that cap their output at 50 tickets and return only short fields (identifier, title, state, labels, priority, assignee, parent). Use `hive-read-state` children only to run several independent collections at once, such as more than one tracker or project, as [delegation](../../hive-core/delegation/delegation.md) sets;
   - unfinished local work: the current branch, uncommitted changes, local branches not merged into the base branch, worktrees, open pull requests, and the plan files `<support>/workspace/<YYYY-MM-DD>-<work>/<YYYY-MM-DD>-<work>.plan.md` in each checkout and worktree whose `Status` field is not `Completed`, found with a filtered search that reads only that field;
   - the work left open on purpose: the `Pending:` records [close hygiene](../../hive-core/close/close-hygiene.md#pending-work) writes, searched in the project the global guidance's «Persistent memory» names, with an explicit result limit. A record is ready to resume when every ticket and condition that unblocks it is resolved; this report never deletes one, which only a close does;
2. Write the report as [Report](#report) sets. Never replace it with a question card.
3. When the user asked what to take next, end the report with the route as [the next-step block](../../hive-core/messages/report-endings.md#the-next-step-block) sets: the recommended ticket or group and the command that assesses it, the `hive-mode` command with the ticket key, in the host's syntax that block sets. The ticket the user names starts `hive-mode`. A plain status request ends with the recommendations.

## Report

Write it in one code block so its columns align, in this order:

1. One summary line: the open tickets, how many wait for each environment's promotion, and how many are in progress. No pull request numbers, commits or dates.
2. The tickets grouped by product module or by relation (modules from tracker labels or projects, otherwise from titles), each group headed by its name and its umbrella ticket. One line per ticket: the state first, padded so the column aligns, then the identifier, then a title of three to six words, and a blocker, the priority or the assignee only when it matters. Keep an initiative's phases together and in order; then the cross-cutting work (infrastructure, CI, tests, tooling, debt) grouped by type, the deferred tickets apart, and a ticket another session has in progress apart. Never a flat list, and never grouped by state, application, repository or layer; when the user asks for another cut, such as the tickets that fit one session, group by that cut.
3. The unfinished local work, and the work left open on purpose, one line per item with the action it needs; a record ready to resume says so, with the evidence.
4. The recommendations, numbered, one line each with its reason, in this order of priority: production bugs, then security or data integrity, then promotion blockers, then debt; among equals, the ticket with the fewest tasks first, then the one with a user-visible outcome. Never recommend a ticket another session has in progress.

```
13 open · 4 waiting for QA · 10 waiting for production

Invoice exports (ABC-126)
  To Deploy  ABC-129  Export retry
  Backlog    ABC-132  CSV templates · blocked by ABC-135

Local work
  • 7 stray package lockfiles → delete

Recommendation
  1. Promote development → qa: closes ABC-129.
```

## Find the next ticket

When flow-close's footer reply **Find the next ticket** runs this report, use the cut of the tickets that reuse this session's context: the files, modules and decisions it touched. Leave out the tickets another session has in progress, and the tickets opened from this work's findings unless they are P0 or P1. Recommend the one that reuses the most context, with its task count (from its description or checklist); choosing it starts `hive-mode` to assess it. That report's next-step block carries the compaction reply for another work item that [the next-step block](../../hive-core/messages/report-endings.md#the-next-step-block) sets.
