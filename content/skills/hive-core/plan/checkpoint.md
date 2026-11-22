# Checkpoint

Read at the start of a build or verify, after a compaction, and when pausing.

Keep a checkpoint in the `## Checkpoint` section of `tasks.md` with `tasks.py <folder> checkpoint` (flags in its `--help`), or by hand when it cannot run: the stage and step, the last commit, the open findings with their decision, the next action, the before captures' paths or that it has none (one `--check` line), each check recorded with its command, result and commit, and the work's worktree when it has one (another session finds it with `git worktree list`). When pausing, record with `--unwritten "<text>"` what no file holds yet, such as a decision or a next step known only to this session, and `--unwritten none` once it is written. Each time the flow reads a stage packet or a conditional leaf, record it with `--rules "<flow>/<packet>.md · extra: <leaf>"`, and each code principle that changed a decision with `--principles "<name>"`, or `--principles none`. With a plan kept in the conversation, or without a plan when the work has an approved example of [create-design](../../create-design/SKILL.md), keep the checkpoint in a file named after the work in the host's temporary location, written at the start and after each task, adding the decisions, criteria, pending task IDs, viewer work key if any, the worktree's path and branch if any, the approved example's record if any, and each resource the work created outside the repository with its location.

The checkpoint also names the work's `Playbook`, the `Playbook` field of `proposal.md` or the `- Playbook: <kind>` line of a checkpoint file, which `flow-build` and `flow-verify` read at entry; [hive-mode](../../hive-mode/SKILL.md#pick-the-playbook) sets who writes it.

A close after the merge writes and reads its checkpoint with `tasks.py checkpoint --out <host temp>/<repo>-<change-id>-checkpoint.md`, since the change folder is no longer edited then.

A host's native task list mirrors the tasks and is relisted after a compaction when the host does not show it; `tasks.md`, or the checkpoint file of a plan kept in the conversation, governs. Tasks are never declared in the viewer.

After a compaction during work, before you reply or end the turn, read the checkpoint, the `HIVE.local.md` files «Local instructions» names, the packets its `Rules:` line names (or the leaves its flow links when a packet is missing) and its extra leaves, then persistent memory, through the memory's read tools, such as `mem_context`; add the active skill and one `git status`, and read code again only before editing it. Continue from the checkpoint's next step; a request the history shows already done gets no second report.

A restriction in a compaction summary that names no source may be your own earlier inference: keep obeying it, but before it is the only reason to stop work or ask a question, look for its source in the checkpoint, the history and the rule it seems to apply, and follow that rule's own terms.

In `flow-verify`, keep the same checkpoint with the reviewed commit and each finding's decision.
