# Change folder and plan revisions

Read when a plan needs a change folder, or a written plan changes.

## Change folder

Write `<Specs>/changes/<change-id>/` in English kebab-case, starting with the ticket key if any, else a short slug. Earlier change folders in the repository are context, not a model: the templates and this leaf set the shape and the ceremony, such as the plan review, even when earlier folders record more. Copy each template from [assets/change-folder/](../../flow-plan/assets/change-folder/), fill it, delete its comments and unused sections:

- `proposal.md`: the request and the ticket's criteria quoted, the user's decisions, the business rules the work adopts ([business rules](../business/business-rules.md)), the scope, the numbered criteria, the prerequisites, each external fact with its measurement, each class of effect with its answer, the delivery decisions;
- `tasks.md`: the tasks with checkboxes, each with its `Closes:` line and check, and the verification of the whole change;
- `design.md`: only when there are design decisions;
- `specs/<capability>/spec.md`: a delta (`ADDED`, `MODIFIED` or `REMOVED Requirements`) when the change adds, alters or removes an observable behavior, contract, permission or stored data.

**Delivery units.** A plan delivered in two or more units (PRs, commits to the base under `direct-base`, or candidates under `hold`) keeps the template's `## Delivery units` section, outside the `### T<n>` tasks; otherwise delete it. Write boxes as the template shows; `check_change.py` rejects a malformed one. [flow-verify](../../flow-verify/SKILL.md) ticks a box by hand and records the result under the checkpoint's `### Evidence`.

**Prerequisites** are what the build needs from outside the repository, usually from a third party or a person: an API key, an account, an access grant, test data. Before listing an account or test data, read [test accounts](../local/test-accounts.md). List each with a checkbox, its provider and the task that needs it, or state there are none. Never write a secret's value; name where it is stored.

## Revise a plan in progress

This applies when a written plan changes during planning or the build: a decision replaced, a criterion changed, added or dropped, a task added or changed, or a delta changed. A new outcome is a new change ([scope](scope-changes.md)).

1. Edit in the same pass every place in the folder that states what changed: the proposal's sections and criteria, `design.md`, each task's text, `Closes:`, test approach and verification, rerunning each changed verification command as [flow-plan](../../flow-plan/SKILL.md) step 3 sets, and the delta. Put `**Replaces:** <old IDs>` on the new decision's or criterion's own line; keep a dropped criterion's line and put `**Retired:** <reason>` on that same line; never delete the old line. Both markers stay English. A decision that adds, replaces or cancels a business rule also updates its record ([business rules](../business/business-rules.md)).
2. Reopen each `[x]` task it invalidates.
3. Run `python3 <flow-plan-dir>/scripts/check_change.py <change-folder>` ([script](../../flow-plan/scripts/check_change.py)): fix each line it prints until it exits 0.
4. When the revision changes a criterion or the delta, have the plan's reviewers reread the revised folder as [reviewer brief](../delegation/reviewer-brief.md) sets: one round per revision. A surviving gap goes to the user as a decision.
5. Add one line to the progress section with the revision, its reason and what it invalidated, and tell the user in a line.

A plan kept in the conversation applies steps 1, 2 and 5 to its checkpoint file.
