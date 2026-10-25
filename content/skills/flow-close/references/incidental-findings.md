# Incidental findings

## When to read

Read before listing, asking about or recording the incidental findings that the global guidance says to propose at a work item's close. Reading this reference does not invoke the closing workflow; [flow-close](../SKILL.md) runs it.

## Actions

List the findings the global guidance says to propose, each with its evidence, severity, and suggested action, for the close question, and mark which ones you recommend opening and why. Ask about every listed finding in that question, splitting its options across consecutive cards when they exceed the tool's limit, as [questions](../../flow-plan/references/questions.md) sets; a finding the user was not asked about stays pending, never declined. Each finding accepts three answers: open a ticket, whose body follows [ticket body](../../flow-research/references/ticket-body.md); decline (saved in persistent memory with a recurrence count); or "don't mention again" (saved in persistent memory as suppressed, under the project and a stable topic, so later searches skip it). A guidance finding about the workflow itself, which has no project ticket, accepts only the last two. Save each answer once the user gives it in the close question, as a user decision. On a recurrence of a declined finding, update its count in the memory write of [step 3](../SKILL.md), raise the severity when the recurrence widens the impact, and propose it again; never propose a suppressed one.

## Which findings are listed and how

This section narrows the incidental-findings rule in the global guidance at the close, and names each element it changes; every other element of that rule stands. It applies to every finding listed above, whatever its source, including the pre-existing findings that the build, review and UI gates hand over.

- **Remaining defect with a concrete failure scenario, proposed as an issue.** A finding is proposed only when its priority is P2 or higher. A P3 finding is not proposed as a ticket; [the P3 rule](../../flow-build/references/verification.md#gate-selection-and-timing) in the verification reference sets what happens to it. Style or wording nits stay in the report only, as the global guidance sets.
- **Severity.** Give every proposed finding a priority from P0 to P3 by its observable impact, with the definitions of the `hive-review-code` role: P0 blocks the release, an operation or a core flow for every input; P1 must be fixed before merge; P2 should be fixed but does not block; P3 is minor. When the source reports its own severity scale, map each finding to P0-P3 by the same impact first.
- **One proposal per finding.** Propose P2 findings that share an area or a cause as one grouped proposal that lists its members with their evidence, not one proposal per member. Findings of P1 or higher are proposed one by one.
- **Duplicate check.** Search the tracker for duplicates once per proposal, so once per group, not once per member.
- **Answer and recurrence.** Record the answer to a group for each of its members, as a separate memory entry per member. When one member of a declined group reappears alone, count it as a recurrence of that member, not of the group.
- **UI findings.** Propose a finding about a rendered interface as a ticket only after it was seen rendered, in a running screen or its capture. A finding seen only in the code stays in the report as unconfirmed.
