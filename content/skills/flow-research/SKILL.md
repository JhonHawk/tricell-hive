---
name: flow-research
description: Investigate a technical question, a reported failure, or a ticket the user picked, by inspecting code, documentation, and evidence, comparing alternatives or resolving conflicting claims. Use for substantive research and diagnosis, for a reported failure, bug, error, or regression that comes without an explicit fix request, for a question whether or how to do work, such as "can you fix this record?", even when its cause is already known, and to assess a ticket before working on it; not for the project's status or backlog, which belongs to flow-status, a simple lookup, or implementation the user already authorized.
---

# Research a question

Research is read-only: nothing changes until the user chooses a route.

## Entry and exit

Enter with a question, a reported failure or a proposal, including a question whether or how to do work; follow-up questions keep the research open and authorize nothing. When one bounded lookup in the main thread answers the question and no work remains for the user to run, report the answer with its source and any uncertainty, with no delegation or route question; a tracker query that lists tickets still follows step 2. Exit with findings and their evidence, and the route question of step 5 when work remains that the user could run next. For a reported failure whose cause is confirmed, leave the delivery and review questions to `flow-build`.

## Steps

1. Name the question and the decision it informs. Ask the user only for intent that changes the conclusion.
2. Investigate. Keep short lookups in the main thread; give larger ones to `hive-research` children, each briefed with the decision, the aspects to cover, what is settled, and when to stop. Run a tracker query that lists tickets in the main thread only when the query caps its output at 50 tickets and short fields (identifier, title, state, labels, priority, assignee); give any other to a `hive-read-state` child.
3. Check each claim that decides the answer by opening its cited source. Resolve conflicting sources by their dates, versions, scope and direct observation; neither the backlog nor the documentation takes precedence, and missing delivery evidence means delivery is not verified, not that it did not happen.
4. Report the answer, its evidence, and what remains uncertain.
5. When the findings leave work the user could run next, end with the route question, asked as [hive-core](../hive-core/SKILL.md) sets: the ways to act on the findings, each naming the flow it starts (`flow-build` for a small understood change, `flow-plan` for larger work), plus keep investigating and stop here, the recommended one first with its reason.

Choosing an option authorizes only the flow it names.

## Assess a chosen ticket

When the user picks a ticket to work on, read it and the code it touches, and report before the route question:

- **Impact:** what uses what the ticket changes, found by searching its literal values (hosts, tables, variables, routes) and direct callers, and whether it touches production.
- **Complexity:** the files, modules and contracts it changes.
- **Readiness:** whether its acceptance criteria and decisions are settled, and the access, data or answers still missing.
- **Dependencies:** the tickets it waits on or blocks.

Recommend the route from that assessment: `flow-build` for a ready, small change, `flow-plan` for ready, larger work, and the missing item when it is not ready.
