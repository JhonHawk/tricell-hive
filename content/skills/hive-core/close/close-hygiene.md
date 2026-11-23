# Record the archive, clean up and record pending work

Read at steps 4, 5 and 7 of [flow-close](../../flow-close/SKILL.md).

## Archive the change folder

With `Specs` in the work's repository, do not archive: create or extend the memory record `archive-pending/<id>`, titled `Archive waiting: <id>` (`<id>` is the `<change-id>`, else the ticket key or a slug), with the folder, the pull request and the merge commit. A hygiene request archives as [batch archive](../../flow-close/references/batch-archive.md) sets. With `Specs` in another repository, archive now as that reference sets for one folder.

**When to offer hygiene.** Count, by reading only, the folders under `<Specs>/changes/` on the base, outside `archive/`, that have an `archive-pending/` record and that no open pull request carries. Offer hygiene, in the close footer's **Hygiene** reply, a close report's **Next** or the session's last message, and name the pending archives there, only when that count is 10 or more: fewer reads to the user as work left half done, while each record keeps its folder until a batch. The status report gives the count at any size and adds the hygiene reply only from 10.

## Clean up

Clean up the branches and worktrees of each repository the work touched as [branch cleanup](../../flow-close/references/branch-cleanup.md) sets (using [cleanup_branches.py](../../flow-close/scripts/cleanup_branches.py)), and stop and remove only what [close cleanup](close-cleanup.md) authorizes. Keep the checkpoint while step 2 lists items as pending, and after the merge as [checkpoint](../plan/checkpoint.md) sets. Before removing a worktree, move what step 2 keeps as [support material](../../flow-close/references/support-material.md) sets. Keep each item that needs the user pending, with its question, as [close modes](close-modes.md) sets. List each remote resource the work created (storage objects, registry images, task definitions) with its removal command and its question: each is a critical effect, removed only when the user asks. Keep anything with uncommitted work or uncertain ownership.

## Pending work

A work item leaves something open on purpose when a task of its plan runs after the merge, such as a measurement over later CI runs, or a later phase of the same work waits. These are not: a promotion along the declared environments (the ticket's state records it) unless a plan task promotes to an environment where the ticket or the request asks for the change, and a promotion not yet made keeps no branch, worktree, process or support item; this close's own deferred steps, which [flow-close](../../flow-close/SKILL.md) step 7 lists in the close record; and anything of an operational task.

At step 7, including a pause close, in this order:

1. Search persistent memory, in the project «Persistent memory» names and with an explicit result limit, for `Pending:` records under `pending/` (not `archive-pending/`) that name this work's ticket or `<change-id>`. Remove from each the items this work resolved, and soft-delete a record left with none with `engram delete <id>` using the ID the search returned, never `--hard`, which erases the record beyond recovery; this work's own record is the exception when item 2 rewrites it, and an abandoned work soft-deletes its own. Without Engram, use the host's memory, or say in the report that the entry could not be removed.
2. When this work leaves something open on purpose, write its own record under the topic `pending/<change-id>`, titled `Pending: <change-id>`, replacing any earlier content, with the change folder's `<change-id>` or, for a plan without one, the ticket key or a short slug, as [flow-plan](../plan/change-folder.md#change-folder) names it. For each open item, state what remains, why, who owns it, and the tickets or conditions that unblock it. Comment the same items on the ticket, when there is one.
