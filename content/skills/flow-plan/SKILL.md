---
name: flow-plan
description: Prepare an implementation plan by resolving consequential decisions, defining affected contracts, and decomposing work into verifiable tasks. Use when planning is requested or needed before a multi-step change; small understood edits need no formal plan.
---

# Prepare an executable plan

1. Read the code, the `## Hive` settings and the existing specs the change touches.
2. Resolve with the user the decisions that are theirs, asked as the global guidance sets. A change to what a consumer of a public interface sends or receives is a contract change: decide in the plan whether it stays compatible. For a change to a screen, settle the UI decisions below.
3. Write the change folder from the templates in [assets/change-folder/](assets/change-folder/) as [Change folder](#change-folder) sets, with acceptance criteria as [Acceptance criteria](#acceptance-criteria) sets.
4. Ask once, in one round: the delivery mode below and the code review mechanism. A UI change is always verified as `flow-verify` sets: your own browser walk and one `hive-review-ux` child.
5. When the plan changes a contract, stored data, security or a dependency, have one `hive-review-plan` child review it, and apply what holds.
6. Read and follow [plan handoff](references/plan-handoff.md): write the plan summary in the message and end it with the build route. A plan does not authorize implementation: only the user's reply starts `flow-build`.

## UI decisions

For a change to a screen, settle the UI before the plan is approved:

1. Load the project's design system first: the UI skills, guide and component library the project guidance names, and the real component catalog, through the library's MCP server or its installed components folder.
2. Fix the data contract each changed screen consumes: the response shape, its states and its errors.
3. For each region that changes, decide the existing screen it follows (with its path), each component by its library name with its variant and size, where each action goes (the page header, a toolbar, a row menu or a dialog) with one primary action per region, the loading, empty, error and permission states, and the copy.
4. Draw each region before and after as an ASCII sketch, at most 60 characters wide, with the real labels, item counts and component names, and write the decisions and sketches in the `UI` section of `design.md`. When the user chooses between compositions, show each option as such a sketch, in the option preview of the host's question tool when it offers one, otherwise in a code block before the question.
5. When no existing pattern fits, settle the composition with one `hive-design-ui` child before the plan is approved.

For a new or rearranged screen, order the tasks: the contract in code (types, schema or client), then the UI against it with stub data typed by that contract and of realistic volume and text length, then the backend that fulfils the same contract.

## Acceptance criteria

Number them `AC1`, `AC2`…, one observable behavior each. Name for each the observation that would show it false, and check that the observation is false on the current base branch: a criterion already true there, or one that restates the request, grades nothing; rewrite or drop it. Every criterion appears in the `Closes:` line of at least one task, whose verification covers it.

## Change folder

Write `<Specs>/changes/<change-id>/`, with `<change-id>` in English kebab-case starting with the ticket key. Earlier change folders in the repository are context, not a model: the templates and this skill set the shape and the ceremony, such as one `hive-review-plan` child, even when earlier folders record more. Copy each template from [assets/change-folder/](assets/change-folder/), fill it, and delete its comments and the sections that do not apply:

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
