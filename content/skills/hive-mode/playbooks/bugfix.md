# Playbook: bugfix

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the user reports a defect and asks for its fix, or an [investigate](investigate.md) confirms one and the user asks for the fix. A defect that cannot wait for the next promotion is [hotfix](hotfix.md).

**Stages.** `hive-mode`, then [flow-build](../../flow-build/SKILL.md) for a small understood change or [flow-plan](../../flow-plan/SKILL.md) when the fix touches a contract, stored data or more than a small change, then [flow-verify](../../flow-verify/SKILL.md) and [flow-close](../../flow-close/SKILL.md).

**Specific to a defect**
- Before the first edit, read and follow [defect fix](../../hive-core/build/defect-fix.md), which also sets the evidence that the defect is gone; for a check that fails, [tests](../../hive-core/build/tests.md).
- Delivery follows the mode in force ([delivery questions](../../hive-core/git/delivery-ask.md)).

**Handoffs.** A fix whose cause is stored data continues as [data-change](data-change.md); a restructuring the fix would need is separate [refactor](refactor.md) work, as [scope changes](../../hive-core/plan/scope-changes.md) sets.

**Leaves.** [verification floor](../../hive-core/build/verification-floor.md), [defect fix](../../hive-core/build/defect-fix.md), [tests](../../hive-core/build/tests.md), [no test runner](../../hive-core/build/tests-no-runner.md), [checks](../../hive-core/build/checks.md), [suppression review](../../hive-core/review/suppression-review.md).

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-core/close/close-report.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-core/messages/report-endings.md) that [close modes](../../hive-core/close/close-modes.md) sets.
