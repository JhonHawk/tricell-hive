---
name: flow-plan
description: Prepare an implementation plan by resolving consequential decisions, defining affected contracts, and decomposing work into verifiable tasks. Use when planning is requested or needed before a multi-step change; small understood edits need no formal plan.
---

# Prepare an executable plan

At plan entry, the primary producer follows [flow-walkthrough](../flow-walkthrough/SKILL.md), which sets when a viewer session applies, for the viewer stage and the plan's document snapshots: the first at the approval stop, then one after each round of changes the user asks for.

1. Read the code, the `## Hive` settings and the existing specs the change touches.
2. Resolve with the user the decisions that are theirs, asked as the global guidance sets. A change to what a consumer of a public interface sends or receives is a contract change: decide in the plan whether it stays compatible. For a change to a screen, settle the UI decisions below.
3. Write the change folder from the templates in [assets/change-folder/](assets/change-folder/) as [Change folder](#change-folder) sets, with acceptance criteria as [Acceptance criteria](#acceptance-criteria) sets.
4. Once the change folder is saved, launch step 5's review when it applies, then ask the delivery mode and the code review mechanism while it runs, as the global guidance's «Delivery» section sets. A UI change is always verified as `flow-verify` sets: your own browser walk and one `hive-review-ux` child.
5. When the plan changes a contract, stored data, security or a dependency, have one `hive-review-plan` child review it, launched before step 4's question, and apply what holds before step 6.
6. Read and follow [plan handoff](references/plan-handoff.md): write the plan summary in the message and end it with the build route. A plan does not authorize implementation: only the user's reply starts `flow-build`.

## UI decisions

For a change to a screen, settle the UI before the plan is approved:

1. Load the project's design system first: the UI skills, guide and component library the project guidance names, and the real component catalog, through the library's MCP server or its installed components folder.
2. Fix the data contract each changed screen consumes: the response shape, its states and its errors.
3. For each region that changes, decide the existing screen it follows (with its path), each component by its library name with its variant and size, where each action goes (the page header, a toolbar, a row menu or a dialog) with one primary action per region, the loading, empty, error and permission states, and the copy.
4. Draw each region before and after as an ASCII sketch, at most 60 characters wide, with the real labels, item counts and component names, and write the decisions and sketches in the `UI` section of `design.md`. When the user chooses between compositions, show each option as such a sketch, in its preview as the global guidance's «Questions and routes» section sets.
5. When no existing pattern fits, settle the composition with one `hive-design-ui` child before the plan is approved.

For a new or rearranged screen, order the tasks by what each one needs in order to start. The contract in code (types, schema or client) usually comes first because both sides consume it; the UI against stub data typed by that contract, with realistic volume and text length, and the backend that fulfils it then need only the contract, so they can run side by side. A task's `Depends on` names only what it needs to start, such as a file, data or an answer; an ordering preference or a validation stop the task does not use, such as the first-view acceptance for a backend task, is not a dependency, because declaring it makes independent work wait.

## Acceptance criteria

Number them `AC1`, `AC2`…, one observable behavior each. Name for each the observation that would show it false, and check that the observation is false on the current base branch: a criterion already true there, or one that restates the request, grades nothing; rewrite or drop it. Every criterion appears in the `Closes:` line of at least one task, whose verification covers it.

## Change folder

Write `<Specs>/changes/<change-id>/`, with `<change-id>` in English kebab-case starting with the ticket key. Earlier change folders in the repository are context, not a model: the templates and this skill set the shape and the ceremony, such as one `hive-review-plan` child, even when earlier folders record more. Copy each template from [assets/change-folder/](assets/change-folder/), fill it, and delete its comments and the sections that do not apply:

- `proposal.md`: the request and the ticket's criteria quoted, the scope, the numbered criteria, the prerequisites, and the delivery decisions;
- `tasks.md`: the tasks with checkboxes, each with its `Closes:` line and how it is checked, and the verification of the whole change;
- `design.md`: only when there are design decisions to record;
- `specs/<capability>/spec.md`: a delta, with `ADDED`, `MODIFIED` or `REMOVED Requirements`, when the change adds, alters or removes an observable behavior, a contract, a permission or stored data.

**Prerequisites** are what the build needs from outside the repository, usually from a third party or a person: an API key, an account or user in a platform, an access grant, test data. List each with a checkbox, who provides it and the task that needs it, or state that there are none. Never write a secret's value; name where it is stored.
