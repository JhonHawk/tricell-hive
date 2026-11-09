# Outside the repository

Read from [flow-plan](../SKILL.md) step 2, after the approach is chosen, when that approach reads or changes an external system. An external system is one provider account or service reached by its own credential or CLI, such as a cloud account, a GitHub organization or a third-party service. What the delivery already uses of this repository's remote, CI and tracker does not count: the work branch, its pull request, its checks and the ticket's moves. They count when the plan depends on anything else in them, such as another branch's contents, a branch's protections or an environment's permissions. Reading this reference authorizes no effect.

## External facts

Before writing the tasks:

1. List each external fact the approach depends on: a state the plan assumes, such as a remote branch's contents, what an account or licence allows, a permission, a limit, a service's availability or a configuration value of an environment.
2. Measure each one with a read-only read. A read that spends money or paid quota follows the confirmation rule of the global guidance's «Scope and authorization».
3. With one external system, the main thread measures. With two or more, launch in the same step one read-only child per external system, as [delegation](../../flow-build/references/delegation.md) sets: `hive-read-state` for a remote's Git state, CI, a tracker or a deployment, and `hive-research` for anything else. Each brief names the facts and their targets and asks, for each fact, for the exact read-only command and a verbatim excerpt of its output. A child that cannot run a provider's read command reports it, and that fact is not measured.
4. Cross-check each fact in the main thread. Accept a fact only with that command and that excerpt. Compare a fact about what an account allows, a limit or a permission with the provider's current documentation too, read as the global guidance's «Evidence» sets. A fact on which the observed state and the documentation disagree, or that no read-only read can show, is written as `not measured`, with the task that meets it first and what happens when it turns out false.
5. Record one line per fact in the proposal's verified-context section, or in the summary of a plan kept in the conversation: the fact, its command or source, its result and its date. The plan summary also lists each fact not measured, as [plan handoff](plan-handoff.md) sets.

## Effects round

At [flow-plan](../SKILL.md) step 4, for each plan:

1. List the effects on external systems that the tasks foresee, leaving out those the delivery answer or a rule already authorizes: commits, the push, the pull request, its merge and the ticket's moves. Group the rest into classes. A class of effect is one kind of action on one named target, such as access-policy changes in one cloud account or administrator merges in one repository, with its spending ceiling, zero when it spends nothing, and the tasks that use it.
2. Ask one decision per class, as the global guidance's «Questions and routes» formats a decision: in the delivery and review round when that round is asked, and in a round of its own when those answers are already in force. When the target is a production one, the decision's `Situation:` line says so.
3. The answer is the user's instruction for each effect of that class that the named tasks of this work item take, within the ceiling. It covers no other plan and no other work item.
4. No class covers these, and each keeps its own question: a production deployment, asked as [flow-deploy](../../flow-deploy/SKILL.md) step 4 sets; an effect that stops or interrupts a component serving users; an irreversible or hard-to-reverse effect; destroying or replacing infrastructure; changing or deleting data the task did not create in a non-local database; and creating accounts or records in a shared environment, as the global guidance's «Test accounts» sets.
5. Before each covered effect, record its target, the change and how to revert it in the task record; the final report lists them. An effect no answered class covers is asked when it arrives.
6. Record one line per class in the proposal's `## Effects` section, or in the summary of a plan kept in the conversation: the class, its target, its ceiling, its tasks and the answer. A plan with no effect on an external system adds no question. An operational task has no effects round: it asks each critical effect, as the global guidance's «Proportionality» sets.
