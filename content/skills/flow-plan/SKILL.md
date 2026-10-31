---
name: flow-plan
description: Prepare an implementation plan by resolving consequential decisions, defining affected contracts, and decomposing work into verifiable tasks. Use when planning is requested or needed before a multi-step change; small understood edits need no formal plan.
---

# Prepare an executable plan

1. Read the code, the `## Hive` settings and the existing specs the change touches.
2. Resolve with the user the decisions that are theirs, asked as [hive-core](../hive-core/SKILL.md) sets. For a change to a screen, settle the UI decisions below.
3. Write the change folder as [Change folder](#change-folder) sets, with acceptance criteria as [Acceptance criteria](#acceptance-criteria) sets.
4. Ask once, in one round: the delivery mode below, the code review mechanism, and how a UI change is verified.
5. When the plan changes a contract, stored data, security or a dependency, have one `hive-review-plan` child review it, and apply what holds.
6. Hand the plan to `flow-build`.

## UI decisions

- Decide the screen pattern the change follows, where each new action goes (the page header, a toolbar, a row menu or a dialog), and its visual weight: one primary action per region, the rest secondary.
- When the user chooses between compositions, show each option as an ASCII sketch of the affected screen or region, at most 60 characters wide, with its real labels and item count, in the option preview of the host's question tool when it offers one, otherwise in a code block before the question.

## Acceptance criteria

Number them `AC1`, `AC2`…, one observable behavior each. Name for each the observation that would show it false, and check that the observation is false on the current base branch: a criterion already true there, or one that restates the request, grades nothing; rewrite or drop it. Every criterion appears in the `Closes:` line of at least one task, whose verification covers it.

## Change folder

Write `<Specs>/changes/<change-id>/`, with `<change-id>` in English kebab-case starting with the ticket key:

- `proposal.md`: the request and the ticket's criteria quoted, the scope and exclusions, the numbered criteria, and the delivery decisions;
- `tasks.md`: the tasks with checkboxes, each with its `Closes:` line and how it is checked;
- `design.md`: only when there are design decisions to record;
- `specs/<capability>/spec.md`: a delta, with `ADDED`, `MODIFIED` or `REMOVED Requirements`, when the change adds, alters or removes an observable behavior, a contract, a permission or stored data.

## Delivery modes

- `interactive`: stop before the push for the user to validate, then push, open the pull request and merge.
- `automatic`: review, push, open the pull request and merge without stopping.
- `direct-base`: push straight to the base branch; the default for a repository whose `## Hive` declares `Delivery: direct-base`, which gets no delivery or review question.
- `hold`: no commit and no publication; the verified working tree is the result.
