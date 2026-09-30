# Change records

Read when creating or updating a change folder, or when closing one and you need the delta formats. `<specs>` below is the `openspec` directory named by the repository's `Specs` setting; the global guidance defines its location and when a change folder is required. The layout follows OpenSpec conventions without requiring its CLI, so the project could adopt that tool later without migrating.

## Files

| File | Carries | From the [plan template](plan-format.md) |
| --- | --- | --- |
| `proposal.md` | Why the change exists, scope and exclusions, constraints, numbered acceptance criteria (`AC<n>`), tracker issues, delivery decisions, current status | Control sheet, Objective, Scope and acceptance, Delivery |
| `design.md` | Verified context, selected approach, contracts, UI view decisions, migration constraints | Verified context, Proposed design |
| `tasks.md` | The change's base commit once the build starts; tasks with their markers, `Closes:` lines, execution, locations, and verification; shared gates; review status, progress, and next step | Tasks, Verification and human review, Review status and progress |
| `research.md` | Retained investigation for this change, only when retention is authorized | — |
| `specs/<capability>/spec.md` | Requirement deltas against the current specs | — |

`proposal.md` and `tasks.md` are required; add `design.md` when the change has design decisions to record. A reader resuming the change starts at `proposal.md`, whose status row names the next step.

Keep in `<specs>/project.md` only what no other source states: ticket state comes from the tracker, delivery state from Git, and the handoff from the active change.

## Requirement deltas

Write each delta file for one capability, using these sections as needed:

```markdown
## ADDED Requirements
### Requirement: <Requirement name>
The system SHALL <observable behavior>.
#### Scenario: <Scenario name>
- **WHEN** <condition or action>
- **THEN** <expected outcome>

## MODIFIED Requirements
### Requirement: <Existing requirement name, unchanged>
<Complete replacement text and its scenarios>

## REMOVED Requirements
### Requirement: <Existing requirement name>
<Reason for removal>
```

Name a capability after a product or system ability in English kebab-case, such as `template-submission` or `portal-status-contrast`, not after a ticket or epic. A MODIFIED or REMOVED requirement must match an existing requirement name in `<specs>/specs/`.

## Areas still described by older documents

When a change touches an area whose current rules live only in an older document, such as an epic that mixes current rules with delivery history, its deltas add those current rules as `ADDED` requirements (or a new product-map page) together with the change's own modifications. Cite the source document in `proposal.md`. At closure the area moves to the current-requirements home, and the older document stays as history for it. Do not rewrite older documents outside a change.

## Product map projects

When the project keeps a product map (`product/<module>/<view>.md` pages of business rules in force), it replaces `<specs>/specs/` as the current-requirements home. Keep the pages in their business voice and location; documentation sites and manuals reuse them.

- `proposal.md` lists the affected views with a one-line delta each.
- Write one delta file per affected view at `changes/<change-id>/product/<module>/<view>.md`, mirroring the page's section headings. Mark each rule `ADDED`, `MODIFIED` (quoting the current rule), or `REMOVED`, and add a new view page whole when the view does not exist yet.
- At closure, apply those rules to the page, add the change to the page's traceability line (for example `Influenciada por`), and leave the other sections untouched.

## Closing

The [flow-close](../../flow-close/SKILL.md#close-the-change-record) skill applies the deltas above, archives the folder, and versions the record once at close.
