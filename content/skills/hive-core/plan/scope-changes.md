# Scope changes

Read when the user changes open work, before proposing a ticket, or when the reads show that the code the work changes lies in an external project. The change is a request that would change open work, and the ticket is for a separate topic, a finding or a research result.

When the user asks for something that would change open work, classify it against the work's request and criteria, and state the class and its destination in one line before acting:

1. **Correction:** it changes how the result does what was already asked; it stays in the current stage.
2. **Added task:** new behavior for the same outcome; it joins the plan, or, without one, stays in `flow-build` only while it stays a small understood change ([proportionality](proportionality.md)), and otherwise goes to `flow-plan`.
3. **New change:** it changes the outcome; ask whether the current work finishes or pauses, and plan the new one apart.
4. **Separate topic:** the current outcome does not need it; give it a destination as the ticket rule below sets, and continue.

A correction or added task that changes a written plan revises it as [Revise a plan in progress](plan-record.md#revise-a-plan-in-progress) sets.

A ticket records work that must be done but will not be done in this session or with its context: it lies outside the files and modules the current work touches, needs another repository, person, decision or access, or is larger than a small understood change ([proportionality](proportionality.md)). This rule governs every proposal to open one, for a separate topic, a finding or a research result:

- recommend fixing now a finding whose fix changes only files the current work already changed and adds no task, dependency or effect;
- recommend a ticket only for work this rule covers;
- recommend only listing a P3 finding.

The current work needs no ticket; propose one for it only when the user asked to track it or the project guidance requires it. To search for an existing ticket, create one, break work into a parent ticket with children, or validate one, follow [create-issue](../../create-issue/SKILL.md).

## An external project

When the reads show that the code the work changes lies in an external project, ask as a decision, before the first plan, branch or edit there, whether to continue here or open a session in that project, recommending the new session so that project's instruction files and `## Hive` settings govern from the start. Ask once per project and session. "Continue here" is the request for that change. Unattended work reports that work and its question as unresolved.
