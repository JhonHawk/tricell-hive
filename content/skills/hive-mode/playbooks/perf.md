# Playbook: perf

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the goal is a measured improvement in speed, size, memory or another resource, with no intended change of behavior. A slowness that is a defect against a stated requirement is a [bugfix](bugfix.md).

**Stages.** `hive-mode`, then [flow-plan](../../flow-plan/SKILL.md) with the measurement as part of the plan, then [flow-build](../../flow-build/SKILL.md), [flow-verify](../../flow-verify/SKILL.md) and [flow-close](../../flow-close/SKILL.md).

**Specific to a performance goal**
- Fix the goal as a number before changing anything: the metric, the method that measures it, the baseline measured on the unchanged code, and the threshold that counts as done. A goal without a baseline and a threshold is not yet a criterion.
- Measure where the cost is instead of guessing. Take the baseline and every later measurement the same way, on the real path with data of realistic volume, as [acceptance criteria](../../flow-plan/SKILL.md#acceptance-criteria) sets for a criterion's real effect.
- Change and measure in cycles as [loops](../../hive-rule-loops/SKILL.md) sets, so each gain is attributed to one change.
- A result is reported with its numbers: baseline, final value, threshold, method and what was not measured. Record each measurement as evidence in the checkpoint ([checkpoint](../../hive-rule-checkpoint/SKILL.md)).
- Behavior must stay the same: the existing checks pass unchanged ([tests](../../hive-rule-tests/SKILL.md)). A speedup bought by dropping a check, a validation or a result is a behavior change.
- Measurements that only read an outside system are read-only; one that loads or changes it is an effect planned as [effects round](../../hive-rule-external-effects/SKILL.md) sets, since it can spend quota or disturb a shared resource.
- Each stop, push, pull request and merge happens under the delivery mode in force.

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-rule-close-report/SKILL.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-rule-report-endings/SKILL.md) that [close modes](../../hive-rule-close-modes/SKILL.md) sets.

**Handoffs.** A defect the profiling exposes is separate work, a finding or a [bugfix](bugfix.md); a restructuring the speedup needs is a [refactor](refactor.md), as [scope changes](../../hive-rule-scope-changes/SKILL.md) sets.

**Leaves.** [verification floor](../../hive-rule-verification-floor/SKILL.md), [loops](../../hive-rule-loops/SKILL.md), [checks](../../hive-rule-checks/SKILL.md), [tests](../../hive-rule-tests/SKILL.md), [checkpoint](../../hive-rule-checkpoint/SKILL.md), [external effects](../../hive-rule-external-effects/SKILL.md).
