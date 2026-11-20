---
name: hive-rule-close-hygiene
description: "Hive rule leaf: read at steps 4 to 7 of a close. Records the pending archive, cleans branches and worktrees, and writes the pending-work records."
disable-model-invocation: true
---

# Record the archive, clean up and record pending work

Read at steps 4 to 7 of [flow-close](../flow-close/SKILL.md).

## Archive the change folder

With `Specs` in the work's repository, do not archive: write the memory record `archive-pending/<change-id>`, titled `Archive waiting: <change-id>`, with the folder, the pull request and merge commit, and `git restore` the folder's uncommitted edits that touch only its checkpoint. A hygiene request archives as [batch archive](../flow-close/references/batch-archive.md) sets. With `Specs` in another repository, archive it now as that reference sets for one folder.

## Clean up

Clean up the branches and worktrees of each repository the work touched as [branch cleanup](../flow-close/references/branch-cleanup.md) sets (using [cleanup_branches.py](../flow-close/scripts/cleanup_branches.py)), and stop and remove only what the global guidance's «Preservation» authorizes at close; keep the checkpoint while step 2 lists items as pending. Before removing a worktree, move the items step 2's answer keeps out of it to the same paths in the main checkout, since removing a worktree deletes its ignored files. List any dump or export made and ask before removing it. List remote resources created (storage objects, registry images, task definitions), each with its removal command, and ask before removing them: each is a critical effect. Keep anything with uncommitted work or uncertain ownership.

## Pending work

A work item leaves something open on purpose when a task of its plan runs after the merge, such as a measurement over later CI runs, or a later phase of the same work waits. A promotion along the declared environments is not one, since the ticket's state records it, unless a plan task promotes to an environment where the ticket or the request asks for the change; a promotion not yet made never keeps a branch, worktree, process or support item. Neither are this close's own deferred steps, which step 7 records apart, nor anything of an operational task.

At step 7, including a pause close, in this order:

1. Search persistent memory for `Pending:` records under `pending/`, never an `archive-pending/` one, that name this work's ticket or `<change-id>`, in the project «Persistent memory» names, with an explicit result limit. Remove from each record the items this work resolved, and soft-delete a record left with none, except this work's own record when step 2 rewrites it, with `engram delete <id>` using the ID the search returned, never `--hard`. An abandoned work soft-deletes its own record. Without Engram, remove the entry from the host's memory, or say in the report that it could not be removed.
2. When this work leaves something open on purpose, write its own record under the topic `pending/<change-id>`, titled `Pending: <change-id>`, replacing any earlier content, using the change folder's `<change-id>`, or for a plan without one the ticket key or a short slug, as [flow-plan](../hive-rule-change-folder/SKILL.md#change-folder) names it. For each open item, state what remains, why, who owns it, and the tickets or conditions that unblock it. Comment the same items on the ticket, when there is one.
