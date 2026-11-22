# Playbook: refactor

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work restructures code and keeps its behavior: renaming, moving, splitting, merging, simplifying or replacing an internal design. If any observable behavior is meant to change, it is a `feature` or a `bugfix`.

**Stages.** `flow-plan` unless the restructuring is a mechanical change or a small understood change as [proportionality](../../hive-core/plan/proportionality.md) defines them, which goes straight to `flow-build`; then `flow-verify` and `flow-close`.

**Specific to a refactor**
- The proof is the same checks passing before and after, so first make sure the behavior the move touches is pinned by a check through the real path, as [tests](../../hive-core/build/tests.md) sets (with no runner, [no test runner](../../hive-core/build/tests-no-runner.md) sets the direct checks).
- Work in steps that each leave the checks passing, so any step can be reverted alone; a test that must change stops the step.
- Search for every caller and literal use of what moves, change them all, and remove what nothing uses once they are moved; do not leave both versions alive.
- Prefer what the project already has over a new abstraction ([reuse before writing](../../hive-core/plan/reuse-before-writing.md)); the restructuring is justified by what it removes, not by what it could allow later.
- Never weaken a check, lint rule or type to finish a step ([weakened checks](../../hive-core/review/suppression-review.md)).
- A bug found along the way becomes separate work, a finding or a `bugfix` ([scope changes](../../hive-core/plan/scope-changes.md)), so the diff stays behavior-neutral and reviewable.
