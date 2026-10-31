---
name: flow-plan
description: Prepare an implementation plan by resolving consequential decisions, defining affected contracts, and decomposing work into verifiable tasks. Use when planning is requested or needed before a multi-step change; small understood edits need no formal plan.
---

# Prepare an executable plan

1. Read the code, the `## Hive` settings and the existing specs the change touches.
2. Resolve with the user the decisions that are theirs, asked as [hive-core](../hive-core/SKILL.md) sets. For a change to a screen, settle the UI decisions below.
3. Write the change folder from the templates in [assets/change-folder/](assets/change-folder/) as [Change folder](#change-folder) sets, with acceptance criteria as [Acceptance criteria](#acceptance-criteria) sets.
4. Ask once, in one round: the delivery mode below, the code review mechanism, and how a UI change is verified.
5. When the plan changes a contract, stored data, security or a dependency, have one `hive-review-plan` child review it, and apply what holds.
6. Show the plan summary below and ask the build route question. A plan does not authorize implementation: only the user's choice starts `flow-build`.

## Plan summary and build route

Summarize the plan in the conversation, because the user may not open the change folder; do not paste it whole:

- the objective in one or two sentences;
- the text of each `AC<n>`, and each requirement the delta adds, modifies or removes: they are the contract the user approves;
- the decisions taken, and the minor calls made without asking with their reasons;
- each task on one line with what it delivers;
- the prerequisites still unmet, with who provides each;
- the verification: tests, UI check and code review;
- the delivery mode, with its validation stop and endpoint;
- what remains open, and the change folder path.

Then ask the build route question, asked as [hive-core](../hive-core/SKILL.md) sets: start `flow-build` with the change folder, recommended unless something blocks it, or stop with the plan ready. The `flow-build` option says it approves the criteria by their IDs with a gloss, such as "approves AC1–AC4 (export, permissions and audit)". A correction to a criterion does not answer the question: revise the plan, show the revised criteria and ask again.

## UI decisions

- Decide the screen pattern the change follows, where each new action goes (the page header, a toolbar, a row menu or a dialog), and its visual weight: one primary action per region, the rest secondary.
- When the user chooses between compositions, show each option as an ASCII sketch of the affected screen or region, at most 60 characters wide, with its real labels and item count, in the option preview of the host's question tool when it offers one, otherwise in a code block before the question.

## Acceptance criteria

Number them `AC1`, `AC2`…, one observable behavior each. Name for each the observation that would show it false, and check that the observation is false on the current base branch: a criterion already true there, or one that restates the request, grades nothing; rewrite or drop it. Every criterion appears in the `Closes:` line of at least one task, whose verification covers it.

## Change folder

Write `<Specs>/changes/<change-id>/`, with `<change-id>` in English kebab-case starting with the ticket key. Copy each template from [assets/change-folder/](assets/change-folder/), fill it, and delete its comments and the sections that do not apply:

- `proposal.md`: the request and the ticket's criteria quoted, the scope, the numbered criteria, the prerequisites, and the delivery decisions;
- `tasks.md`: the tasks with checkboxes, each with its `Closes:` line and how it is checked, and the verification of the whole change;
- `design.md`: only when there are design decisions to record;
- `specs/<capability>/spec.md`: a delta, with `ADDED`, `MODIFIED` or `REMOVED Requirements`, when the change adds, alters or removes an observable behavior, a contract, a permission or stored data.

**Prerequisites** are what the build needs from outside the repository, usually from a third party or a person: an API key, an account or user in a platform, an access grant, test data. List each with a checkbox, who provides it and the task that needs it, or state that there are none. Never write a secret's value; name where it is stored.

## Delivery modes

- `interactive`: stop before the push for the user to validate, then push, open the pull request and merge.
- `automatic`: review, push, open the pull request and merge without stopping.
- `direct-base`: push straight to the base branch; the default for a repository whose `## Hive` declares `Delivery: direct-base`, which gets no delivery or review question.
- `hold`: no commit and no publication; the verified working tree is the result.
