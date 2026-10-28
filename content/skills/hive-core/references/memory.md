# Persistent memory

## When to read

Read before writing persistent memory, and when two memories, or a memory and a current source, disagree. The global guidance keeps what persistent memory means and when to write it. Reading this reference does not invoke a workflow.

## Where to write

Persistent memory means Engram when its tools work in this session, and otherwise the host's own memory: the store the host writes for itself and reloads across sessions, not its instruction files. When you write memory yourself and the host has both, write the project's decisions, restrictions, and work or delivery state to Engram, each under its own `topic_key`, and keep in the host's own memory only what concerns that host and, for a project fact, the pointer `engram:<project>/<topic_key>` instead of a copy. This holds over the host's own memory instructions, except when the user asks to save something in the host's own memory.

## When to write

The global guidance names the points; in full, write only when the user makes a decision; when a plan is ready, recording work state rather than a copy of the plan; at a work-item close, the durable point; when you create or remove a worktree, so another session can resume or skip it; and, when the user ends the session and the memory supports it, in a session summary. A session that ends without notice writes nothing more. This cadence holds over the save instructions of memory tools, such as a plugin's rule to save after every decision.

## Correct a stale entry

A host's automatic capture may still copy project facts. When two memories, or a memory and a current source, disagree, act on either, and correct an entry, only when a current source settles which one is stale; otherwise report both without editing. Correct a stale Engram entry by updating it, following Engram's conflict prompts when it raises them, and a stale fact in the host's own memory, where the host lets you edit it, by replacing only that fact with the pointer; this rule authorizes that edit. Change only the entries the task touches.
