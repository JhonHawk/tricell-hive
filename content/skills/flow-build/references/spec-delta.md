# Spec delta without a plan

Read before the first edit of a change built without a plan that changes a behavior an existing `<Specs>/specs/<capability>/spec.md` describes, once the delivery and review questions are answered. The spec itself is never edited by hand: a hygiene batch merges the change's delta into it, so the spec keeps a record of each change.

1. Write the change folder as [change folder](../../hive-core/plan/change-folder.md#change-folder) sets, before any other edit, keeping the proposal's `Status` since the hygiene batch needs it to archive the folder, with the `specs/<capability>/spec.md` delta that holds the changed requirement, written against the spec with the deltas of the change folders still under `<Specs>/changes/` on the base applied, since a hygiene batch has not merged them yet. Write no `design.md`.
2. Launch no `hive-review-plan` child for this folder, neither when it is written nor when a revision changes it: a change small enough to skip the plan also skips its review. A risky task still gets the task check of [flow-build](../SKILL.md) step 3.
3. Run the check of [Revise a plan in progress](../../hive-core/plan/change-folder.md#revise-a-plan-in-progress) step 3 on the folder before the first edit of the code or guidance.
4. Build the change as [flow-build](../SKILL.md) sets for work with a change folder: record each task with `tasks.py`, and leave the archive and the delta merge to the hygiene batch, as [batch archive](../../flow-close/references/batch-archive.md) sets.
