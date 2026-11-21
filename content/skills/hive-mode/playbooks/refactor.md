# Playbook: refactor

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work restructures code and keeps its behavior: renaming, moving, splitting, merging, simplifying or replacing an internal design. If any observable behavior is meant to change, it is a [feature](feature.md) or [bugfix](bugfix.md).

**Stages.** `hive-mode`, then [flow-plan](../../flow-plan/SKILL.md) when the restructuring crosses files, modules or a contract or is a risky change under the global guidance's «Proportionality»; otherwise straight to [flow-build](../../flow-build/SKILL.md). Then [flow-verify](../../flow-verify/SKILL.md) and [flow-close](../../flow-close/SKILL.md).

**Specific to a refactor**
- Behavior stays the same, and the proof is the same checks passing before and after. First make sure the behavior the move touches is pinned by a check through the real path; where none exists, add it before moving anything ([tests](../../hive-core/references/tests.md); with no runner, [no test runner](../../hive-core/references/tests-no-runner.md) sets the direct checks).
- Work in steps that each leave the checks passing, so any step can be reverted alone. Do not edit a test to make a step pass: a test that must change means the behavior changed, which stops the step.
- Search for every caller and literal use of what moves, change them all, and remove what nothing uses once they are moved. Do not leave both versions alive.
- Prefer what already exists in the project over a new abstraction ([reuse before writing](../../hive-core/references/reuse-before-writing.md)); the restructuring is justified by what it removes, not by what it could allow later.
- Never weaken a check, lint rule or type to finish a step ([weakened checks](../../hive-core/references/suppression-review.md)).
- A bug found along the way is not fixed inside the refactor: it becomes separate work, a finding or a [bugfix](bugfix.md), as [scope changes](../../hive-core/references/scope-changes.md) sets, so the diff stays behavior-neutral and reviewable.
- Each stop, push, pull request and merge happens under the delivery mode in force.

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-core/references/close-report.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-core/references/report-endings.md) that [close modes](../../hive-core/references/close-modes.md) sets.

**Leaves.** [verification floor](../../hive-core/references/verification-floor.md), [tests](../../hive-core/references/tests.md), [no test runner](../../hive-core/references/tests-no-runner.md), [reuse before writing](../../hive-core/references/reuse-before-writing.md), [checks](../../hive-core/references/checks.md), [weakened checks](../../hive-core/references/suppression-review.md), [scope changes](../../hive-core/references/scope-changes.md).
