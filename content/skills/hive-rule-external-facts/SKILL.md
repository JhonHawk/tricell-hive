---
name: hive-rule-external-facts
description: "Hive rule leaf: read in flow-plan step 2 when the chosen approach reads or changes an external system. List and measure each external fact before writing tasks."
disable-model-invocation: true
---

# Outside the repository

Read from [flow-plan](../flow-plan/SKILL.md) step 2, after the approach is chosen, when that approach reads or changes an external system. An external system is one provider account or service reached by its own credential or CLI, such as a cloud account, a GitHub organization or a third-party service. What the delivery already uses of this repository's remote, CI and tracker does not count: the work branch, its pull request, its checks, the ticket's moves and the write of its business rules. They count when the plan depends on anything else in them, such as another branch's contents, a branch's protections or an environment's permissions. The plan reviewer reads it to check what the plan records, not the steps the main thread ran. Reading this leaf authorizes no effect.

## External facts

Before writing the tasks:

1. List each external fact the approach depends on: a state the plan assumes, such as a remote branch's contents, what an account or licence allows, a permission, a limit, a service's availability or a configuration value of an environment.
2. Measure each one with a read-only read. A read that spends money or paid quota follows the confirmation rule of [long checks](../hive-rule-checks/SKILL.md#long-checks).
3. With one external system, the main thread measures. With two or more, launch side by side one read-only child per external system, as [delegation](../hive-rule-delegation/SKILL.md) sets: a `hive-read-state` child whose brief names the system, how it is reached, the facts and their targets, and asks, for each fact, for the exact read-only command and a verbatim excerpt of its output. A child that cannot run a provider's read command reports it, and that fact is not measured.
4. Cross-check each fact in the main thread. Accept a fact only with that command and that excerpt. Compare a fact about what an account allows, a limit or a permission with the provider's current documentation too, read as [verify current state](../hive-principle-verify-current-state/SKILL.md) sets. A fact on which the observed state and the documentation disagree, or that no read-only read can show, is written as `not measured`, with the task that meets it first and what happens when it turns out false.
5. Record one line per fact in the proposal's verified-context section, or in the summary of a plan kept in the conversation: the fact, its command or source, its result and its date. The plan summary also lists each fact not measured, as [plan handoff](../flow-plan/references/plan-handoff.md) sets.
