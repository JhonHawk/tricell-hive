# Outside the repository

Read when an approach reads or changes an external system. Read it from [flow-plan](../../flow-plan/SKILL.md) step 2, after the approach is chosen. An external system is one provider account or service reached by its own credential or CLI, such as a cloud account, a GitHub organization or a third-party service. What the delivery already uses of this repository's remote, CI and tracker (the work branch, its pull request, its checks, the ticket's moves and the write of its business rules) does not count; anything else in them does, such as another branch's contents, a branch's protections or an environment's permissions. Reading this leaf authorizes no effect.

## External facts

Before writing the tasks:

1. List each external fact the approach depends on, such as a remote branch's contents, what an account or licence allows, a permission, a limit or an environment's configuration value. Measure each one with a read-only read. A read that spends money or paid quota follows the confirmation rule of [long checks](../build/checks.md#long-checks).
2. With one external system, the main thread measures. With two or more, launch side by side one read-only `hive-read-state` child per system, as [delegation](../delegation/delegation.md) sets, whose brief names the system, how it is reached, the facts and their targets, and asks for each fact the exact read-only command and a verbatim excerpt of its output. A fact whose read the child cannot run is not measured.
3. Cross-check each fact in the main thread and accept it only with that command and excerpt. Compare a fact about what an account allows, a limit or a permission with the provider's current documentation too, read as [verify current state](../principles/verify-current-state.md) sets. A fact on which the observed state and the documentation disagree, or that no read-only read can show, is written `not measured`, with the task that meets it first and what happens when it turns out false.
4. Record one line per fact in the proposal's verified-context section, or in the summary of a plan kept in the conversation: the fact, its command or source, its result and its date.
