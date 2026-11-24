# Clean up and record pending work

Read at steps 5 and 7 of [flow-close](../../flow-close/SKILL.md).

## Clean up

Clean up the branches and worktrees of each repository the work touched as [branch cleanup](../../flow-close/references/branch-cleanup.md) sets (using [cleanup_branches.py](../../flow-close/scripts/cleanup_branches.py)), and stop and remove only what [close cleanup](close-cleanup.md) authorizes. Keep the checkpoint while step 2 lists items as pending, and after the merge as [checkpoint](../plan/checkpoint.md) sets. Before removing a worktree, move what step 2 keeps as [support material](../../flow-close/references/support-material.md) sets. Keep each item that needs the user pending, with its question, as [close modes](close-modes.md) sets. List each remote resource the work created (storage objects, registry images, task definitions) with its removal command and its question: each is a critical effect, removed only when the user asks. Keep anything with uncommitted work or uncertain ownership.

## Pending work

A work item leaves something open on purpose when a task of its plan runs after the merge, such as a measurement over later CI runs, or a later phase of the same work waits. These are not: a promotion along the declared environments (the ticket's state records it) unless a plan task promotes to an environment where the ticket or the request asks for the change, and a promotion not yet made keeps no branch, worktree, process or support item; this close's own deferred steps, which [flow-close](../../flow-close/SKILL.md) step 7 lists in the close record; and anything of an operational task.

At step 7, including a pause close, in this order:

1. Search persistent memory, in the project «Persistent memory» names and with an explicit result limit, for `Pending:` records under `pending/` that name this work's ticket or `<work>`. Remove from each the items this work resolved, and soft-delete a record left with none with `engram delete <id>` using the ID the search returned, never `--hard`, which erases the record beyond recovery; this work's own record is the exception when item 2 rewrites it, and an abandoned work soft-deletes its own. Without Engram, use the host's memory, or say in the report that the entry could not be removed.
2. When this work leaves something open on purpose, write its own record under the topic `pending/<work>`, titled `Pending: <work>`, replacing any earlier content, with the `<work>` name of its plan file or, for a plan kept in the conversation, the ticket key or a short slug, as [plan record](../plan/plan-record.md#plan-file) names it. For each open item, state what remains, why, who owns it, and the tickets or conditions that unblock it. Comment the same items on the ticket, when there is one.
