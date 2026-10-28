# Persistent memory

## When to read

Read before writing persistent memory, before naming a project in a memory search, and when two memories, or a memory and a current source, disagree. The global guidance keeps what persistent memory means and its write cadence, the one home of that cadence. Reading this reference does not invoke a workflow.

## Where to write

Persistent memory means Engram when its tools work in this session, and otherwise the host's own memory: the store the host writes for itself and reloads across sessions, not its instruction files. When you write memory yourself and the host has both, write the project's decisions, restrictions, and work or delivery state to Engram, each under its own `topic_key`, and keep in the host's own memory only what concerns that host and, for a project fact, the pointer `engram:<project>/<topic_key>` instead of a copy. This holds over the host's own memory instructions, except when the user asks to save something in the host's own memory.

When an Engram call takes a project, in a write or a search, and in the `<project>` of that pointer, use the name Engram resolves: the `project_name` of the nearest `.engram/config.json` at or above the working directory, or, without one, the project Engram reports as current. Never replace it with a name that differs from it, such as the `## Hive` `Project` value, a tracker project or a directory name you derive yourself, even when it looks like the same project: a different name stores the record in a separate project that the other sessions do not search.

## Correct a stale entry

A host's automatic capture may still copy project facts. When two memories, or a memory and a current source, disagree, act on either, and correct an entry, only when a current source settles which one is stale; otherwise report both without editing. Correct a stale Engram entry by updating it, following Engram's conflict prompts when it raises them, and a stale fact in the host's own memory, where the host lets you edit it, by replacing only that fact with the pointer; this rule authorizes that edit. Change only the entries the task touches.
