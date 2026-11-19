---
name: hive-rule-scope-changes
description: "Hive rule leaf: read when the user asks for something that changes open work, or before proposing a ticket. Classifies the change, states its destination and sets when a ticket is proposed."
disable-model-invocation: true
---

# Scope changes

Read when the user asks for something that would change open work, and before you propose a ticket for a separate topic, a finding or a research result.

When the user asks for something that would change open work, classify it against the work's request and criteria, and state the class and its destination in one line before acting:

1. **Correction:** it changes how the result does what was already asked; it stays in the current stage.
2. **Added task:** new behavior for the same outcome; it joins the plan, or, without one, stays in `flow-build` only while it stays a small understood change («Proportionality»), and otherwise goes to `flow-plan`.
3. **New change:** it changes the outcome; ask whether the current work finishes or pauses, and plan the new one apart.
4. **Separate topic:** the current outcome does not need it; give it a destination as the ticket rule below sets, and continue.

A correction or added task that changes a written plan revises it as [flow-plan](../flow-plan/SKILL.md) «Revise a plan in progress» sets.

A ticket records work that must be done but will not be done in this session or with its context: it lies outside the files and modules the current work touches, needs another repository, person, decision or access, or is larger than a small understood change («Proportionality»). This rule governs every proposal to open one, for a separate topic, a finding or a research result: recommend fixing now a finding whose fix changes only files the current work already changed and adds no task, dependency or effect, a ticket only for work this rule covers, and only listing a P3 finding. The current work needs no ticket; propose one for it only when the user asked to track it or the project guidance requires it. To check whether a ticket already exists before proposing one, to create one, to break work down into a parent ticket with children, or to validate one that exists, read and follow [create-issue](../create-issue/SKILL.md): it sets that search, the ticket's content and level, and how a new ticket relates to its origin.
