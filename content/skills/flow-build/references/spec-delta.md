# Spec delta without a plan

Read before the first edit of a change built without a plan that changes a behavior an existing `<Specs>/specs/<capability>/spec.md` describes. The spec itself is never edited by hand: the close merges the change's delta into it, so the spec keeps a record of each change.

1. Write the change folder as [change folder](../../hive-rule-change-folder/SKILL.md#change-folder) sets, with these files only: `proposal.md` with the request, the user's decisions, the `Playbook` field and the numbered criteria; `tasks.md` with the tasks; and the `specs/<capability>/spec.md` delta that holds the changed requirement. Write no `design.md`.
2. Launch no `hive-review-plan` child for this folder: a change small enough to skip the plan also skips its review.
3. Run `python3 -I <flow-plan-dir>/scripts/check_change.py <folder>` ([check](../../flow-plan/scripts/check_change.py)) and fix each line it prints until it exits 0, before the first edit of the code or guidance.
4. Build the change as [flow-build](../SKILL.md) sets for work with a change folder: record each task with `tasks.py`, and leave the archive and the delta merge to `flow-close`.
