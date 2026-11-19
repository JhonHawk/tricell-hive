---
name: flow-close
description: Close a work item and its session hygiene. Use when a work item reaches its completion point, is abandoned or is paused; when an operational task ends; when the user ends the session; or when the user asks for session hygiene, such as "haz higiene"; not to clean up, prune or reorganize existing docs or support material, which project-layout does. Closes the change record, cleans merged branches, task temporaries and processes, proposes what to do with the work's support material, and reports the outcome with a close footer whose reply starts any cleanup left.
---

# Close a work item and the session

Read [close modes](../hive-rule-close-modes/SKILL.md) first: it sets which steps run now, which wait and how the close ends. For explicit E2E work, read [coordination](../flow-build/references/e2e-coordination.md) before closing units or cleaning their resources.

Read [ticket state](../hive-rule-ticket-state/SKILL.md) first: it defines the completion point and the state of the ticket at it, including the `ready` stop and the follow-up close of a pull request that was merged later.

To clean up, prune, reorganize, curate or archive existing docs or support material, or to check or migrate a project's layout, use [project-layout](../project-layout/SKILL.md) instead of these steps.

1. When this work has viewer captures, close it first, as [publication](../flow-verify/references/viewer-publication.md#close-and-input-hygiene) sets, and read it back: the work reads closed when it was integrated or abandoned, `in_review` when its open pull request waits for the user's merge or a human reviewer, and paused when it was paused. Closing it changes only its record: never stop the viewer service or close the viewer the user has open, even when this work started it. This step comes before any cleanup, because cleanup removes the inputs a failed close would need. When the viewer is not running, do not start it: keep the inputs its close needs, continue the rest of the close, and name in the report the viewer session left open with where those inputs are.
2. Take the inventory of this work's support material and ask its decision as [support material](references/support-material.md) sets, before steps 4–5: step 4 delivers what the answer commits, and step 5 removes the worktree and checkpoint the inventory reads.
3. When the work is integrated into the base branch, move its ticket as [ticket state](../hive-rule-ticket-state/SKILL.md) sets. At the completion point, commit its change folder in the specs repository when [git-workflow step 4](../git-workflow/SKILL.md) leaves that commit to the close. Close its business rules as [business rules](../flow-plan/references/business-rules.md) sets.
4. Close an integrated change folder as [close hygiene](../hive-rule-close-hygiene/SKILL.md#archive-the-change-folder) sets.
5. Clean up the branches and worktrees of each repository the work touched as [close hygiene](../hive-rule-close-hygiene/SKILL.md#clean-up) sets.
6. Write the final report as [Final report](../hive-rule-close-report/SKILL.md#final-report) sets.
7. Write the outcome to persistent memory, the account index rows as [test accounts](../hive-rule-test-accounts/SKILL.md) sets, the pending-work records as [Pending work](../hive-rule-close-hygiene/SKILL.md#pending-work) sets, and each candidate's record as [Candidates](../hive-rule-close-report/SKILL.md#candidates) sets.
8. End with the close footer as [close footer](../hive-rule-close-footer/SKILL.md) sets.
