---
name: hive-rule-checkpoint
description: "Hive rule leaf: read at the start of a build or verify, after a compaction, and when pausing. What the checkpoint holds, how tasks.md or the checkpoint file is written, the native task list."
disable-model-invocation: true
---

# Checkpoint

Read at the start of a build or verify, after a compaction, and when pausing.

Keep a checkpoint in the `## Checkpoint` section of `tasks.md`: the stage and step, the last commit, the open findings with their decision, the next action, the paths of its before captures or that it has none, as one `--check` line, and the work's worktree when it has one, written with `tasks.py <folder> checkpoint` (`--finding` per open finding, `--check` per check recorded with its command, result and commit, `--worktree "<path> · <branch>"`, both of which later updates keep, and `--worktree none` once the worktree is removed, and, when pausing, `--unwritten "<text>"` with what no file holds yet, such as a decision or a next step known only to this session, `--unwritten none` once it is written; another session finds the worktree with `git worktree list`), or by hand when it cannot run. With a plan kept in the conversation, or without a plan when the work has an approved example of [create-design](../create-design/SKILL.md), keep it in a file named after the work in the host's temporary location, written at the start and after each task, adding the decisions, criteria, pending task IDs, viewer work key if any, the worktree's path and branch if any, the approved example's record if any, and each resource the work created outside the repository with its location.

When the host has a native task list, list the tasks in it at the start, and again with the remaining ones when you resume and the host does not show it again, such as after a compaction; mark each one in progress when it starts and completed when it is done. The native list is a live view: `tasks.md`, or the checkpoint file of a plan kept in the conversation, governs. Tasks are never declared in the viewer.

In `flow-verify`, keep the same checkpoint with the reviewed commit and each finding's decision.
