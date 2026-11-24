# Playbook: data-change

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work changes stored data: a schema, a migration, a backfill, a value the code stops reading, or a correction of records. A database dump or export that leaves no repository change is an operational task, which keeps its plan in the conversation and still asks each critical effect.

**Stages.** The full chain, with `flow-plan` always, since stored data is a risky change under the global guidance's «Proportionality», and `flow-deploy` for each environment the change reaches.

**Specific to stored data**
- Before planning, read and follow [stored data](../../hive-core/build/stored-data.md); list and measure each external fact the plan rests on with read-only reads ([external facts](../../hive-core/plan/external-facts.md)).
- A write to shared or production data is an effect planned as the [effects round](../../hive-core/plan/external-effects.md) sets, which keeps a write to data the task did not create in a non-local database out of any class.
- Checks that sign in or use records the environment holds read the account index first: [test accounts](../../hive-core/local/test-accounts.md).
- A dump or export the task made is kept as the global guidance's «Preservation» and [close modes](../../hive-core/close/close-modes.md) set.

**Handoffs.** A bug whose cause is bad data continues here from `bugfix`; a code restructuring the migration needs is a separate `refactor`.
