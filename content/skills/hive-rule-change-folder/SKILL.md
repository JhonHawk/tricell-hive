---
name: hive-rule-change-folder
description: "Hive rule leaf: read when a plan needs a change folder, or when a written plan changes. The folder's files and templates, prerequisites, and revising a plan in progress."
disable-model-invocation: true
---

# Change folder and plan revisions

Read when the plan needs a change folder, and when a written plan changes during planning or the build.

## Change folder

Write `<Specs>/changes/<change-id>/`, with `<change-id>` in English kebab-case starting with the ticket key when the work has a ticket, otherwise a short slug of the work. Earlier change folders in the repository are context, not a model: the templates and this leaf set the shape and the ceremony, such as one `hive-review-plan` child, even when earlier folders record more. Copy each template from [assets/change-folder/](../flow-plan/assets/change-folder/), fill it, and delete its comments and the sections that do not apply:

- `proposal.md`: the request and the ticket's criteria quoted, the user's decisions in plain words, the business rules the work adopts, recorded as [business rules](../hive-rule-business-rules/SKILL.md) sets, the scope, the numbered criteria, the prerequisites, each external fact with its measurement, each class of effect with its answer, and the delivery decisions;
- `tasks.md`: the tasks with checkboxes, each with its `Closes:` line and how it is checked, and the verification of the whole change;
- `design.md`: only when there are design decisions to record;
- `specs/<capability>/spec.md`: a delta, with `ADDED`, `MODIFIED` or `REMOVED Requirements`, when the change adds, alters or removes an observable behavior, a contract, a permission or stored data.

**Prerequisites** are what the build needs from outside the repository, usually from a third party or a person: an API key, an account or user in a platform, an access grant, test data. Before listing an account or test data, read [test accounts](../hive-rule-test-accounts/SKILL.md). List each with a checkbox, who provides it and the task that needs it, or state that there are none. Never write a secret's value; name where it is stored.

## Revise a plan in progress

This applies when a written plan changes during planning or the build: a decision replaced, a criterion changed, added or dropped, a task added or changed, or a delta changed. A change of outcome is a new change, as [scope changes](../hive-rule-scope-changes/SKILL.md) sets.

1. Edit in the same pass every place in the change folder that states what changed: the proposal's sections and criteria, `design.md`, each task's text, `Closes:`, test approach and verification, and the delta. Mark the new decision or criterion with `**Replaces:** <old IDs>` on its line; keep a dropped criterion's line and add `**Retired:** <reason>` to it; never delete the old line. Both markers stay in English, like `Closes:` and `Observed in:`. A decision that adds, replaces or cancels a business rule also updates the rule record, as [business rules](../hive-rule-business-rules/SKILL.md) sets.
2. Reopen each `[x]` task the revision invalidates.
3. Run `python3 <flow-plan-dir>/scripts/check_change.py <change-folder>` ([script](../flow-plan/scripts/check_change.py)), which checks criteria coverage, `Observed in:` places, replaced/retired IDs and `## Business rules` grammar, and fix each line it prints until it exits 0.
4. When the revision changes a criterion or the delta, have one `hive-review-plan` child review the revised folder, briefed as [delegation](../hive-rule-reviewer-brief/SKILL.md) sets: one round per revision. A gap that survives it goes to the user as a decision.
5. Add one line to the progress section with the revision, its reason and what it invalidated, and tell the user in one line what changed.

A plan kept in the conversation applies steps 1, 2 and 5 to its checkpoint file.
