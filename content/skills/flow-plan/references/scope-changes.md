# Scope changes

## When to read

Read when, while work is open in any flow, the user asks for something that would change that work, an answer to a question adds to it, or the work finds it needs more than the request asked: before acting on it. This includes a request during research, planning, the build, a validation pause, the first rendered view, after integration and before close, at the close question, and during a promotion. It is the one home for classifying such a change; the open-topics rule in the global guidance points here. Reading this reference does not invoke the planning workflow.

## Classify by intent

The work's intent is the outcome its quoted request names, and, in a retained plan, its `AC<n>`. Classify each change against that intent into exactly one class:

1. **Correction:** it changes how the result does what the request or its criteria already ask. The global guidance calls this a topic that clarifies. It stays in the current stage.
2. **Added task:** it asks for new behavior that serves the same outcome. In a retained plan it becomes an added task, as [plan revision](plan-revision.md) sets. Without a retained plan, the work stays in [flow-build](../../flow-build/SKILL.md) only while it remains a small understood change; when the change adds another ticket, another repository, a contract, or stored data, stop dependent edits and route the work to [flow-plan](../SKILL.md).
3. **New change:** it changes the outcome the request named. Ask whether the current work continues to its own completion point or pauses as [flow-close](../../flow-close/SKILL.md) sets, and plan the new outcome as a separate change.
4. **Separate topic:** the current outcome does not need it, or it belongs to other work. Give it a destination under the open-topics rule in the global guidance, such as a ticket, a plan, watched, or dropped, and continue the work.

The global guidance's labels map to these classes without overlap: a topic that clarifies is a correction; a related topic is an added task or a new change; a separate topic is a separate topic.

Examples that mark the boundaries:

- On a save form, "the button should say Save" is a correction.
- On an export screen, "also export to CSV" is an added task.
- "Let's turn this into a theming system" is a new change.
- "Dark mode would be nice someday" is a separate topic.

## Say it, ask only when it matters

Before acting, state the class and where the work goes in one line. Ask only when two classes are both reasonable and lead to different deliveries, typically an added task against a new change; ask by following [questions](questions.md), with the recommended class first and its reason.

Under [unattended delegation](../../unattended-delegation/SKILL.md), a correction within its objective and scope keeps the delegation; an added task or a new change is a scope change that ends it, as that skill sets.

## Where the work returns

Apply the stage's own rule at each point; this list only names it:

- **Research:** a finding that changes the objective or scope stops the iterative route, as [flow-research](../../flow-research/SKILL.md) sets.
- **Plan:** a correction or added task revises the plan, as [plan revision](plan-revision.md) sets.
- **Build:** an added task follows class 2 above, including the route to `flow-plan`.
- **Validation pause:** a correction runs as a correction round of [flow-verify](../../flow-verify/SKILL.md); an added task follows class 2 and is verified as [plan revision](plan-revision.md) sets.
- **After integration, before close:** a defect fix of the same change goes as the follow-up pull request that [delivery decisions](delivery-decisions.md) covers. A new requirement is a new change: the integrated change's record closes normally and the new outcome is planned apart.
- **Close question:** an answer that adds to the work is classified here before the close continues.
- **Promotion:** the run's candidate and tickets are fixed when it starts. A request to add or remove tickets ends the current run at a safe point, a step boundary where nothing is half-applied, and the next run starts with the new set from the first step of [flow-deploy](../../flow-deploy/SKILL.md).

## Repeated corrections

Correction rounds have no fixed limit, since each one is the user's request. When a third correction round, during a validation pause or after the first rendered view, touches the same aspect as the two before it, ask by following [questions](questions.md) whether to return to the plan, revising it or opening a new change, before running that round; recommend returning when the corrections keep changing what the result should do rather than how it does it.
