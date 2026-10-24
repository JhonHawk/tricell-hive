# Write a ticket body

## When to read

Read before writing the body of a ticket you open, whether it records follow-up work from research, an incidental finding, or any other route. It does not authorize opening a ticket: the rule that asks the user and checks for duplicates still applies. When the repository has an issue template, follow it and use this reference for what the template leaves open.

## What the body carries

Write it for someone who picks the ticket up later without this conversation.

- **Title:** the behavior to change, in the project's own vocabulary, not the fix you have in mind.
- **Current behavior:** what happens now, observably, with the conditions or input that produce it.
- **Expected behavior:** what should happen instead, observably.
- **Evidence:** the locators that show the current behavior, such as `path:line`, a command and its relevant output, or a link, each with the commit or date at which you read it. Evidence shows where the behavior was seen; it does not prescribe where to change the code, since lines move.
- **Acceptance criteria:** each one an observable behavior that is false in the current state, so that it can show when the ticket is done.
- **Out of scope:** what this ticket deliberately leaves out, with the reason when it is not obvious.

For an incidental finding, the body also carries what the global incidental-findings rule asks the proposal to state. Leave out a section that has nothing to say rather than filling it with boilerplate.
