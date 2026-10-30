# Work-item boundary

## When to read

Read before the first message of a build, a verify stage or other authorized work while work of your own remains, before ending a turn while such work remains, and before ending a turn that completes a work item. It sets what those messages carry and when the turn may end. The message's format, citations and context line stay in [reporting](reporting.md).

## Between tasks

A work item is a ticket or the scope the user authorized, such as a plan, one of its phases, or an iterative route; a task inside it is not one.

While authorized work of your own remains, no rule requires a message between tasks; send any message with your next tool call, since a response without one can end the turn. A question that arrives meanwhile gets a short direct answer, one line when it asks about progress, such as how it is going or what comes next. A message that another rule shapes, such as a failed integration, an exposed secret or an explanation the user asked for, takes the form that rule or [reporting](reporting.md) sets; it still goes with that call. A notification that arrives meanwhile follows [the notification turn](reporting.md#report-a-running-or-completed-item).

Until that work is done, end the turn only:

- to ask the user a question that [questions](../../hive-core/references/questions.md) says to ask now, since a deferred one waits until no independent work remains;
- at a validation stop;
- to wait on a completion signal with no independent work left, as [delegation and waits](../../flow-research/references/delegation.md) sets;
- at a blocker that stops all remaining work;
- when the user asks you to stop, or another rule ends the work, such as an end condition of [unattended delegation](../../unattended-delegation/SKILL.md).

## When a work item completes

End a turn that completes a work item with a close question, after the report and its closing items; a turn that ends on the report is incomplete. Read [the close-question procedure](../SKILL.md#close-question) for its options, the fast research and exceptions. Investigation-only findings and follow-ups use [flow-research](../../flow-research/SKILL.md#deliver-the-finding)'s route question instead.
