---
name: flow-research
description: Investigate a technical question, a reported failure, or a ticket the user picked, by inspecting code, documentation, and evidence, comparing alternatives or resolving conflicting claims. Use for an investigation one bounded lookup cannot answer, for a reported failure, bug, error, or regression that comes without an explicit fix request, for a question whether or how to do work, such as "can you fix this record?", even when its cause is already known, and to assess a ticket before working on it; not for the project's status or backlog, which belongs to flow-status, a simple lookup, or implementation the user already authorized.
---

# Research a question

The primary producer follows [flow-walkthrough](../flow-walkthrough/SKILL.md), which sets when a viewer session applies, to start or resume the viewer stage and publish at each stage change, new capture or document snapshot; publication failure does not stop independent research.

Research does not change the project, tracker or shared data until the user chooses a route; the local viewer publication above records its progress.

## Entry and exit

Enter with a question, a reported failure or a proposal, including a question whether or how to do work; follow-up questions keep the research open and authorize nothing. When one bounded lookup (as the global guidance's «Proportionality» defines it) in the main thread answers the question and no work remains for the user to run, report the answer with its source and any uncertainty, with no delegation or route; a tracker query that lists tickets still follows step 2. Exit with findings and their evidence, and the route of step 5 when work remains that the user could run next. For a reported failure whose cause is confirmed, leave the delivery and review questions to `flow-build`.

## Steps

1. Name the question and the decision it informs. Ask the user only for intent that changes the conclusion.
2. Investigate. Delegate an aspect to a `hive-research` child when its brief is shorter than the lookup and the lookup needs more than 5 file or source reads; keep the rest in the main thread. Brief each child with the decision, the aspects to cover, what is settled, and when to stop. Run a tracker query that lists tickets in the main thread, capped at 50 tickets and short fields (identifier, title, state, labels, priority, assignee); use `hive-read-state` children only to run several independent collections at once. For a failure whose cause is not yet shown, list the plausible causes and check each one, in parallel when the checks are independent, before concluding; report the ones you ruled out with the evidence that ruled them out.
3. Check each claim that decides the answer by opening its cited source. Resolve conflicting sources by their dates, versions, scope and direct observation; neither the backlog nor the documentation takes precedence, and missing delivery evidence means delivery is not verified, not that it did not happen.
4. Report the answer, its evidence, and what remains uncertain. When the user asks to keep or share the result, or the report exceeds what fits on one screen, offer a `flow-report` page as one option of the question the message ends with.
5. When the findings leave work the user could run next, end the report with the route as the global guidance's «Questions and routes» section sets: the recommended flow (`flow-build` for a small understood change as «Proportionality» defines it, `flow-plan` for larger work) with its reason, and one line saying how to continue, keep investigating or stop.

The user's reply authorizes only the flow it names.

## Assess a chosen ticket

When the user picks a ticket to work on, read it and the code it touches, and write this assessment in the message:

- **Impact:** what uses what the ticket changes, found by searching its literal values (hosts, tables, variables, routes) and direct callers, and whether it touches production.
- **Complexity:** the files, modules and contracts it changes.
- **Readiness:** whether its acceptance criteria and decisions are settled, and the access, data or answers still missing.
- **Dependencies:** the tickets it waits on or blocks.

A ticket is not ready while its scope is open: its criteria are missing, it names more than one change, or the specs contradict it. Settle that in research: ask the scope decisions in rounds, as the global guidance's «Questions and routes» section sets, such as which operations enter or whether the spec or the ticket governs, and add the answers to the assessment. Then end with the route of step 5: `flow-build` for a small change, `flow-plan` for larger work. Plan settles how the change is done, never what enters it.
