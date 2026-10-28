# Scope changes

## When to read

Read when, while work is open in any flow, the user asks for something that would change that work, an answer to a question adds to it, or the work finds it needs more than the request asked: before acting on it. This includes a request during research, planning, the build, a validation pause, the first rendered view, after integration and before close, at the close question, and during a promotion. It is the one home for classifying such a change; the open-topics rule in the global guidance points here. Reading this reference does not invoke the planning workflow.

## Classify by intent

The work's intent is the outcome its quoted request names, and, in a retained plan, its `AC<n>`. Classify each change against that intent into exactly one class:

1. **Correction:** it changes how the result does what the request or its criteria already ask. The global guidance calls this a topic that clarifies. It stays in the current stage.
2. **Added task:** it asks for new behavior that serves the same outcome. In a retained plan it becomes an added task, as [plan revision](../../flow-plan/references/plan-revision.md) sets. In a build without a retained plan, the work stays in [flow-build](../../flow-build/SKILL.md) only while it remains a small understood change; when the change adds another ticket, another repository, a contract, or stored data, stop dependent edits and route the work to [flow-plan](../../flow-plan/SKILL.md). Before any build, such as during research, the added task joins what that flow's route question offers; a classification never authorizes implementation. An added task, or a nearby fix proposed as an addition to a plan, that creates a new interaction surface, such as a focus stop, a control or a route, carries its own acceptance criterion and enters the UI checks that the `UI verification` value selects; when that value is none or not yet resolved, it enters the implementer's rendered check that [the UI gate](../../flow-verify/references/ui-gate.md) sets, or stays out of the work.
3. **New change:** it changes the outcome the request named. Ask whether the current work continues to its own completion point or pauses as [flow-close](../../flow-close/SKILL.md) sets, and plan the new outcome as a separate change. During research there is no work to continue or pause: the new change joins the route question that [flow-research](../../flow-research/SKILL.md) asks, as a route of its own.
4. **Separate topic:** the current outcome does not need it, or it belongs to other work. Give it a destination under the open-topics rule in the global guidance, such as a ticket, a plan, watched, or dropped, and continue the work.

The global guidance's labels map to these classes without overlap: a topic that clarifies is a correction; a related topic is an added task or a new change, resolved as scope before the work that depends on it; a separate topic is a separate topic, which waits for its destination. An answer to a pending question is not a topic, and a child does not apply the open-topics rule.

Examples that mark the boundaries:

- On a save form, "the button should say Save" is a correction.
- On an export screen, "also export to CSV" is an added task.
- "Let's turn this into a theming system" is a new change.
- "Dark mode would be nice someday" is a separate topic.

## Say it, ask only when it matters

Before acting, state the class and where the work goes in one line. Until an added task or a new change is resolved, hold the work that depends on it: an added task is resolved once it has its criteria in the plan or, without one, in the build's stated scope; a new change once the user chose whether the current work continues or pauses. Ask only when two classes are both reasonable and lead to different deliveries, typically an added task against a new change; ask by following [questions](questions.md), with the recommended class first and its reason.

Under [unattended delegation](../../unattended-delegation/SKILL.md), only a correction can be a message that keeps it; an added task or a new change changes its scope, which ends it as that skill sets for such a message.

## Where the work returns

Apply the stage's own rule at each point; this list only names it:

- **Research:** a change the user adds while the research runs joins the route question of [flow-research](../../flow-research/SKILL.md), which asks again when the findings change; a finding that changes the objective or scope stops the iterative route, as that skill sets.
- **Plan:** a correction or added task revises the plan, as [plan revision](../../flow-plan/references/plan-revision.md) sets.
- **Build:** an added task follows class 2 above, including the route to `flow-plan`. When the build returns to plan, the commits already made stay on the work branch, the branch and the ticket's in-progress state stay as they are, and the plan records that work as done and keeps what is still valid, as [plan revision](../../flow-plan/references/plan-revision.md) sets.
- **Validation pause:** a correction runs as a correction round of [flow-verify](../../flow-verify/SKILL.md); an added task follows class 2 and is verified as [plan revision](../../flow-plan/references/plan-revision.md) sets.
- **After integration, before close:** a correction or defect fix of the same change is a follow-up under the same delivery answer, as the follow-up rule of [delivery decisions](../../flow-plan/references/delivery-decisions.md) covers. The integrated change's record closes normally; an added task then goes to `flow-build` when it is a small understood change and otherwise to `flow-plan`, and a new change follows class 3.
- **Close question:** an answer that adds to the work is classified here before the close continues.
- **Promotion:** the run's candidate and tickets are fixed when it starts. Promotion carries the whole base branch, so a ticket already on it when the run started is already in the candidate, and one integrated after the start goes to the next run. A request to remove a ticket whose commits are on the base branch cannot be met by restarting: hold the run before its next effect, such as merging the promotion pull request or deploying, and ask whether to keep it in the promotion, and the run continues, or to revert it on the base branch as its own change, and then stop the run at a safe point, a step boundary where nothing is half-applied, and start the next run from the first step of [flow-deploy](../../flow-deploy/SKILL.md).

## Repeated corrections

Correction rounds have no fixed limit, since each one is the user's request. When a third correction round touches the same aspect as the two before it in the same stage, counting the rounds after the first rendered view and those of a validation pause separately, ask by following [questions](questions.md) whether to return to the plan, revising it or opening a new change, before running that round; recommend returning when the corrections keep changing what the result should do rather than how it does it.
