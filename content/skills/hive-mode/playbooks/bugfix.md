# Playbook: bugfix

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the user reports a defect and asks for its fix, or an [investigate](investigate.md) confirms one and the user asks for the fix. A defect that cannot wait for the next promotion is [hotfix](hotfix.md).

**Stages.** `hive-mode`, then [flow-build](../../flow-build/SKILL.md) for a small understood change or [flow-plan](../../flow-plan/SKILL.md) when the fix touches a contract, stored data or more than a small change, then [flow-verify](../../flow-verify/SKILL.md) and [flow-close](../../flow-close/SKILL.md).

**Specific to a defect**
- Show the failure first: a reproduction, or a regression test seen failing, before the first edit. The same check passes after the fix and is the evidence that the defect is gone.
- Name the cause before editing and fix it there, not where the failure surfaced: [defect fix](../../hive-rule-defect-fix/SKILL.md) sets the cause, the search for the same fault elsewhere, what a workaround is and the earliest layer that should catch the error.
- When a check fails, first decide whether the code or the check is wrong, as [tests](../../hive-rule-tests/SKILL.md) sets. In a repository or package without a test runner, [no test runner](../../hive-rule-tests-no-runner/SKILL.md) sets the direct checks.
- Fixing the cause may be a choice the user owns, such as a contract or stored data change: report it before writing a workaround.
- Other instances of the fault outside the files the work changes are findings, not edits.
- Each stop, push, pull request and merge happens under the delivery mode in force.

**Handoffs.** A fix whose cause is stored data continues as [data-change](data-change.md); a restructuring the fix would need is separate [refactor](refactor.md) work, as [scope changes](../../hive-rule-scope-changes/SKILL.md) sets.

**Leaves.** [verification floor](../../hive-rule-verification-floor/SKILL.md), [defect fix](../../hive-rule-defect-fix/SKILL.md), [tests](../../hive-rule-tests/SKILL.md), [no test runner](../../hive-rule-tests-no-runner/SKILL.md), [checks](../../hive-rule-checks/SKILL.md), [suppression review](../../hive-rule-suppression-review/SKILL.md).
