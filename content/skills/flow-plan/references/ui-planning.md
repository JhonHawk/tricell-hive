# Plan interfaces within the product

Read when creating or changing a screen, application shell, layout, or reusable UI pattern. Apply this procedure in a formal plan or a bounded implementation brief; it does not require a document or design approval for every small visual edit.

## Establish the relevant scope

Identify the application, area, user task, and shell before choosing a reference. A monorepo may contain independent frontends; one application may use multiple shells. Inspect applicable project guidance, routes, layout components, tokens, and representative rendered screens. Prefer a suitable reference within the affected application; share across applications only when their visual and interaction requirements fit.

Distinguish the shell (navigation and persistent frame), screen pattern (list, detail, form or other task composition), and components/tokens. Reusing buttons alone does not establish consistency. Similar CRUD data does not necessarily imply the same workflow, density, actions, or layout.

## Resolve the convention

| Observed state | Required decision |
| --- | --- |
| Applicable guide and pattern exist | Read the relevant section, verify it against current sources, and identify what the new view reuses and changes |
| Guide is missing but a suitable implementation exists | Inspect code and rendered behavior, identify why it fits, and capture the minimum reusable convention within the authorized work |
| Guide, design, and implementation disagree | Record the discrepancy and resolve which behavior to preserve or change from current intent and evidence; none automatically takes precedence |
| No suitable pattern exists | Propose a composition and settle consequential visual/interaction choices before dependent implementation; use a sketch or mockup when it helps make the decision reviewable |

An existing screen is a candidate, not automatically an accepted standard. Distinguish proposed, accepted, and deprecated patterns using the project's decision process. Resolve material ambiguity with the user; routine reuse under an established convention needs no additional approval. Keep independent work moving while dependent visual decisions remain unresolved. An unavailable rendered reference is an explicit evidence limit, not permission to claim visual consistency from source alone.

## Record the view decisions

For each affected view or coherent family of views, record only the relevant details in the existing plan or brief:

- Application/area, route or entry point, and user task.
- Applicable guide and its observed currency; selected shell/pattern and inspected example, with source paths and a rendered/design reference when available.
- Components/composition to reuse, justified differences, and affected consumers if shared code changes. Prefer composition using the existing architecture over copying a screen or introducing a universal CRUD engine.
- Relevant loading, empty, error, permission, long-content and responsive states, including interaction and accessibility expectations.
- Concrete acceptance observations and any unresolved decision that blocks dependent work. The shared [verification reference](../../flow-build/references/verification.md) owns test depth and runtime/UI gates.

A compact view table is sufficient when decisions are settled. The plan must explain the composition, not merely promise to follow a guide or defer all design to build.

## Keep project knowledge current

Use the project's existing documentation home. If none exists and a reusable convention needs retention, use `_support/docs/frontend/ui-conventions.md` at the repository or declared shared-workspace support home. Start with sections per application/area and split into references only when useful. Do not create a separate guide for each feature or impose a shared shell across the monorepo.

When creating or restructuring that guide, read [the adaptable project UI guide](ui-conventions-template.md). Keep decisions, conditions, exceptions, and links there; component APIs and token values remain in their existing canonical sources. Reuse an existing design system, Storybook, Figma library, or equivalent rather than requiring new tooling. Check that linked material is accessible to the assigned implementer; availability is not proof of reading.

When a convention changes within the authorized task, update its guide and representative example with the implementation, identifying affected consumers and obsolete references. Keep the existing application owner/reviewer responsible for consequential changes; do not introduce a committee or an extra approval gate. Link the guide from the project's existing agent guidance with its frontend applicability condition when that guidance update is within scope; otherwise report the missing navigation. Creating conventions does not authorize redesigning unrelated screens.
