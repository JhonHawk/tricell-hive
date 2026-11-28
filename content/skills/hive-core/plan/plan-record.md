# Plan record and revisions

Read when a plan needs a plan file, or a written plan changes.

## Plan file

[flow-plan](../../flow-plan/SKILL.md#where-the-plan-lives) sets which plans get a file. Write it at `<support>/workspace/<YYYY-MM-DD>-<work>/<YYYY-MM-DD>-<work>.plan.md`: `<support>` is the project's support home as [layout](../../project-layout/references/layout.md) sets, the date is the day the plan is written, and `<work>` is English kebab-case, starting with the ticket key if any, else a short slug. That folder is the work's one home: the plan, its records `<YYYY-MM-DD>-<work>.<type>.<ext>`, `images/` and `evidence/`. The file is local and never committed; the pull request body carries the public record, as [git-workflow](../../git-workflow/SKILL.md) step 6 sets.

Earlier plans in the repository are context, not a model: the template and this leaf set the shape and the ceremony, such as the plan review, even when earlier plans record more. Copy [assets/plan.md](../../flow-plan/assets/plan.md), fill it, and delete its comments and unused sections. Write the sections in its order:

- the header table, `## Decisions` and `## Business rules`. Decisions carries the user's decisions and the assumptions taken without asking. Business rules carries the rules the work adopts ([business rules](../business/business-rules.md));
- `## Request` with the user's words and the ticket's criteria quoted, `## Objective`, and `## Scope and acceptance` with the numbered criteria;
- `## Prerequisites`, `## Verified context` with each external fact and its measurement, `## Effects` with each class of effect and its answer, and `## Delivery`;
- `## Design`, only with the design decisions: approaches, UI and copy, impact and recovery;
- `## Tasks`, each with its `Closes:` line and check, and `## Delivery units` when it applies;
- `## Checkpoint`, which `tasks.py` writes, and `## Review status and progress` last, one line per event.

A change to an observable contract, permission or stored data that a spec under `<Specs>/specs/` describes edits that spec in the same pull request, as a task of the plan; there is no delta file.

**Delivery units.** A plan delivered in two or more units (PRs, commits to the base under `direct-base`, or candidates under `hold`) keeps the template's `## Delivery units` section, outside the `### T<n>` tasks; otherwise delete it. [flow-verify](../../flow-verify/SKILL.md) ticks a box by hand and records the result under the checkpoint's `### Evidence`.

**Prerequisites** are what the build needs from outside the repository, such as an API key, an account, an access grant or test data: list each with a checkbox, its provider and the task that needs it, or state there are none. Read [test accounts](../local/test-accounts.md) before listing an account or test data. Never write a secret's value; name where it is stored.

**Children and sandboxes.** The plan stays in the main checkout. A child's brief gives its absolute path, and the child reads it as [child brief](../delegation/child-brief.md) sets. For an sbx sandbox, copy the plan into the worktree's ignored workspace; the main checkout's copy governs, and the copy is renewed after each revision.

**Close.** The plan stays as local support material, and the retention of [project-layout](../../project-layout/SKILL.md) prunes it.

## Revise a plan in progress

This applies when a written plan changes during planning or the build: a decision replaced, a criterion changed, added or dropped, or a task added or changed. A new outcome is a new work item ([scope](scope-changes.md)).

1. Edit in the same pass every place in the plan that states what changed: its decisions, criteria, design, each task's text, `Closes:`, test approach, `Risk:` and verification, rerunning each changed verification command as [Verification lines](../../flow-plan/SKILL.md#verification-lines) sets. Put `**Replaces:** <old IDs>` on the new decision's or criterion's own line; keep a dropped criterion's line and put `**Retired:** <reason>` on that same line; never delete the old line. A decision that adds, replaces or cancels a business rule also updates its record ([business rules](../business/business-rules.md)).
2. Reopen each `[x]` task it invalidates.
3. Run `python3 -I <flow-plan-dir>/scripts/check_change.py <plan-file>` ([script](../../flow-plan/scripts/check_change.py)) and fix each line it prints until it exits 0.
4. Check the revision, then reread it:
   - Before any reread, check what the revision wrote: each step, section, ID, `file:line`, size or limit a written passage names exists and says what the passage claims of it, and no other place in the plan, or in a file the passage cites, states the same thing differently. The reason is that in real sessions from 2026-11-10 to 2026-11-26, 71% of a reread's findings were defects in the revision's own text.
   - A decision the revision takes as an assumption that relaxes a standing rule or overrides a decision of the user goes to the user as a decision, not to the reread.
   - Then, when the revision changes one of these, have the plan's reviewers reread the revised plan:
     - a contract or an interface;
     - a criterion or a decision;
     - an assumption other than one it takes to the user;
     - an external fact or a verification line.

     They reread it as [reviewer brief](../delegation/reviewer-brief.md) sets, and, with several reviewers, those [plan review](../../flow-plan/references/plan-review.md) names reread it.
   - A reviewer does not reread a revision that applies its own proposal as proposed.
   - While the plan is being planned, it gets at most one reread round: a gap that survives it, or one a later revision brings, goes to the user as a decision before the full plan summary, with the affected criteria and tasks. During the build, each such revision gets one reread round, and a surviving gap goes to the user as a decision.
5. Add one line to the progress section with the revision, its reason and what it invalidated, renew the sandbox copy when there is one, and tell the user in a line.

A plan kept in the conversation applies steps 1, 2 and 5 to its checkpoint file.
