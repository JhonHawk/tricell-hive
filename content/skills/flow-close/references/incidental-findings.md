# Incidental findings

## When to read

Read before listing, asking about or recording the incidental findings that the global guidance says to propose at a work item's close. Reading this reference does not invoke the closing workflow; [flow-close](../SKILL.md) runs it.

## Actions

List the findings the global guidance says to propose, each with its evidence, severity, and suggested action, for the close question, and mark which ones you recommend opening and why. Ask about every listed finding in that question, splitting its options across consecutive cards when they exceed the tool's limit, as [questions](../../flow-plan/references/questions.md) sets; a finding the user was not asked about stays pending, never declined. Each finding accepts three answers: open a ticket, whose body follows [ticket body](../../flow-research/references/ticket-body.md); decline (saved in persistent memory with a recurrence count); or "don't mention again" (saved in persistent memory as suppressed, under the project and a stable topic, so later searches skip it). A guidance finding about the workflow itself, which has no project ticket, accepts only the last two. Save each answer once the user gives it in the close question, as a user decision. On a recurrence of a declined finding, update its count in the memory write of [step 3](../SKILL.md), raise the severity when the recurrence widens the impact, and propose it again; never propose a suppressed one.
