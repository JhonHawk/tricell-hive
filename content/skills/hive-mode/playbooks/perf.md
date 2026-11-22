# Playbook: perf

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the goal is a measured improvement in speed, size, memory or another resource, with no intended change of behavior. A slowness that is a defect against a stated requirement is a `bugfix`.

**Stages.** `flow-plan` with the measurement as part of the plan, then `flow-build`, `flow-verify` and `flow-close`.

**Specific to a performance goal**
- Fix the goal as a number before changing anything: the metric, the method that measures it, the baseline measured on the unchanged code, and the threshold that counts as done. A goal without a baseline and a threshold is not yet a criterion.
- Take the baseline and every later measurement the same way, on the real path with data of realistic volume, as [acceptance criteria](../../flow-plan/SKILL.md#acceptance-criteria) sets for a criterion's real effect.
- Change and measure in cycles as [loops](../../hive-core/plan/loops.md) sets, so each gain is attributed to one change.
- Report baseline, final value, threshold, method and what was not measured, and record each measurement as evidence in the checkpoint ([checkpoint](../../hive-core/plan/checkpoint.md)).
- The existing checks pass unchanged ([tests](../../hive-core/build/tests.md)): a speedup bought by dropping a check, a validation or a result is a behavior change.
- A measurement that loads or changes an outside system is an effect planned as the [effects round](../../hive-core/plan/external-effects.md) sets, since it can spend quota or disturb a shared resource; one that only reads is not.

**Handoffs.** A defect the profiling exposes is a finding or a `bugfix`; a restructuring the speedup needs is a `refactor`.
