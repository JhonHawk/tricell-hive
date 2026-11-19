---
name: hive-principle-separate-before-serializing-shared-state
description: "Hive principle leaf: read when concurrent actors could write the same file, branch, key or object. Give each its own target first; add a lock or single writer only for a real shared invariant."
disable-model-invocation: true
---

# Separate before serializing shared state

Read when two or more actors (processes, agents, jobs, requests) could write the same file, branch, key or object.

1. Ask first whether they need one shared target or are each publishing their own facts. When they publish their own facts, give each its own file, key, branch or folder, and combine them only where the state is read. Two workers writing separate fields of one state file still share it.
2. Only when one shared write target is a real invariant, serialize access in the structure: a lock file, a single writer, sequential phases, or a compare-and-swap.
3. An instruction to actors to take turns is not concurrency control.

Check before you add a lock: could the sharing be removed instead?

Parallel children in one working tree follow [delegation](../hive-rule-delegation/SKILL.md): disjoint write paths, not a lock.
