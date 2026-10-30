---
name: flow-research
description: Investigate a technical question or project and delivery state by inspecting code, documentation, and evidence, comparing alternatives or resolving conflicting claims. Use for substantive research and diagnosis, for a reported failure, bug, error, or regression that comes without an explicit fix request, for a question whether or how to do work, such as "can you fix this record?", even when its cause is already known, and for reviewing a pull request outside the current work; not for a simple lookup or implementation the user already authorized.
---

# Research a question

Research is read-only: nothing changes until the user chooses a route.

1. Name the question and the decision it informs. Ask the user only for intent that changes the conclusion.
2. Investigate. Keep short lookups in the main thread; give larger ones to `hive-research` children, each briefed with the decision, the aspects to cover, what is settled, and when to stop.
3. Check each claim that decides the answer by opening its cited source.
4. Report the answer, its evidence, and what remains uncertain.
5. When the findings leave work the user could run next, end with the route question, asked as [hive-core](../hive-core/SKILL.md) sets: the ways to act on the findings, each naming the flow it starts (`flow-build` for a small understood change, `flow-plan` for larger work), plus keep investigating and stop here, the recommended one first with its reason.

Choosing an option authorizes only the flow it names.
