# Playbook: data-change

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work changes stored data: a schema, a migration, a backfill, a value the code stops reading, or a correction of records. A database dump or export that leaves no repository change is an operational task, which keeps its plan in the conversation and still asks each critical effect.

**Stages.** `hive-mode`, then [flow-plan](../../flow-plan/SKILL.md) always, since stored data is a risky change under the global guidance's «Proportionality», then [flow-build](../../flow-build/SKILL.md), [flow-verify](../../flow-verify/SKILL.md), [flow-deploy](../../flow-deploy/SKILL.md) for each environment the change reaches, and [flow-close](../../flow-close/SKILL.md).

**Specific to stored data**
- Count the affected records in each environment the change reaches before planning, migrate what the code stops reading, and define the cutover checks and the recovery: [stored data](../../hive-rule-stored-data/SKILL.md) sets each of these.
- List and measure each external fact the plan rests on, with read-only reads, before writing tasks: [external facts](../../hive-rule-external-facts/SKILL.md).
- A write to shared or production data, and any data the task did not create in a non-local database, needs the user's authorization as an effect: plan it in the [effects round](../../hive-rule-external-effects/SKILL.md), never take it in a task without its answer.
- Checks that sign in or use records the environment holds read the account index first: [test accounts](../../hive-rule-test-accounts/SKILL.md). In QA and deployed environments use what exists; create nothing without asking.
- Keep a dump or export the task made as evidence; ask before removing it.
- Each stop, push, pull request and merge happens under the delivery mode in force; promotion follows `flow-deploy` and its authorization per environment.

**Handoffs.** A bug whose cause is bad data continues here from [bugfix](bugfix.md); a code restructuring the migration needs is separate [refactor](refactor.md) work, as [scope changes](../../hive-rule-scope-changes/SKILL.md) sets.

**Leaves.** [verification floor](../../hive-rule-verification-floor/SKILL.md), [stored data](../../hive-rule-stored-data/SKILL.md), [external facts](../../hive-rule-external-facts/SKILL.md), [effects round](../../hive-rule-external-effects/SKILL.md), [test accounts](../../hive-rule-test-accounts/SKILL.md), [security boundaries](../../hive-rule-security-boundaries/SKILL.md) when the data is exposed or crosses a trust boundary.
