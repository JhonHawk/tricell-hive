---
name: flow-plan
description: Prepare an implementation plan by resolving consequential decisions, defining affected contracts, and decomposing work into verifiable tasks. Use when planning is requested or needed before a multi-step change; small understood edits need no formal plan.
---

# Prepare an executable plan

1. Read the code, the `## Hive` settings and the existing specs the change touches.
2. Resolve with the user the decisions that are theirs, asked as [hive-core](../hive-core/SKILL.md) sets. For a change to a screen, decide the pattern it follows and where each new action goes, with its visual weight.
3. Write the change folder `<Specs>/changes/<change-id>/`:
   - `proposal.md`: the request, the scope, and numbered acceptance criteria (AC1, AC2…);
   - `tasks.md`: the tasks, each with the criteria it closes and how it is checked;
   - `design.md`: only when there are design decisions to record.
4. Ask once, in one round: the delivery (local only, commits, push and pull request, or merge), the code review mechanism, and how a UI change is verified.
5. When the plan changes a contract, stored data, security or a dependency, have one `hive-review-plan` child review it, and apply what holds.
6. Hand the plan to `flow-build`.
