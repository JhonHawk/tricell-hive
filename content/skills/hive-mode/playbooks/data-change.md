# Playbook: data-change

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work changes stored data: a schema, a migration, a backfill, a value the code stops reading, or a correction of records. A database dump or export that leaves no repository change is an operational task, which keeps its plan in the conversation and still asks each critical effect.

**Stages.** `hive-mode`, then [flow-plan](../../flow-plan/SKILL.md) always, since stored data is a risky change under the global guidance's «Proportionality», then [flow-build](../../flow-build/SKILL.md), [flow-verify](../../flow-verify/SKILL.md), [flow-deploy](../../flow-deploy/SKILL.md) for each environment the change reaches, and [flow-close](../../flow-close/SKILL.md).

**Specific to stored data**
- Before planning, read and follow [stored data](../../hive-core/build/stored-data.md).
- List and measure each external fact the plan rests on, with read-only reads, before writing tasks: [external facts](../../hive-core/plan/external-facts.md).
- A write to shared or production data is an effect: plan it as the [effects round](../../hive-core/plan/external-effects.md) sets, which keeps a write to data the task did not create in a non-local database out of any class.
- Checks that sign in or use records the environment holds read the account index first: [test accounts](../../hive-core/local/test-accounts.md). In QA and deployed environments use what exists; create nothing without asking.
- A dump or export the task made is kept as the global guidance's «Preservation» and [close modes](../../hive-core/close/close-modes.md) set.
- Delivery follows the mode in force ([delivery questions](../../hive-core/git/delivery-ask.md)); promotion follows `flow-deploy`.

**Handoffs.** A bug whose cause is bad data continues here from [bugfix](bugfix.md); a code restructuring the migration needs is separate [refactor](refactor.md) work, as [scope changes](../../hive-core/plan/scope-changes.md) sets.

**Leaves.** [verification floor](../../hive-core/build/verification-floor.md), [stored data](../../hive-core/build/stored-data.md), [external facts](../../hive-core/plan/external-facts.md), [effects round](../../hive-core/plan/external-effects.md), [test accounts](../../hive-core/local/test-accounts.md), [security boundaries](../../hive-core/build/security-boundaries.md) when the data is exposed or crosses a trust boundary.

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-core/close/close-report.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-core/messages/report-endings.md) that [close modes](../../hive-core/close/close-modes.md) sets; a promotion's cleanup close uses the deployment row of [close report](../../hive-core/close/close-report.md#variants).
