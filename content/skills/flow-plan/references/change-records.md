# Change records

Read when creating, updating, or closing a change folder. `<specs>` below is the `openspec` directory named by the repository's `Specs` setting; the global guidance defines its location and when a change folder is required. The layout follows OpenSpec conventions without requiring its CLI, so the project could adopt that tool later without migrating.

## Files

| File | Carries | From the [plan template](plan-format.md) |
| --- | --- | --- |
| `proposal.md` | Why the change exists, scope and exclusions, acceptance criteria, tracker issues, delivery decisions, current status | Control sheet, Objective, Scope and acceptance, Delivery |
| `design.md` | Verified context, selected approach, contracts, UI view decisions, migration constraints | Verified context, Proposed design |
| `tasks.md` | Checkbox tasks with execution, locations, and verification; shared gates; review status, progress, and next step | Tasks, Verification and human review, Review status and progress |
| `research.md` | Retained investigation for this change, only when retention is authorized | — |
| `specs/<capability>/spec.md` | Requirement deltas against the current specs | — |

`proposal.md` and `tasks.md` are required; add `design.md` when the change has design decisions to record. A reader resuming the change starts at `proposal.md`, whose status row names the next step.

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

## Close a change

After the change is integrated into its base branch:

1. For each delta file, apply ADDED requirements to `<specs>/specs/<capability>/spec.md`, creating it when absent; replace each MODIFIED requirement block by name; delete each REMOVED block. Edit only those blocks, leave the rest of the file untouched, and review the resulting diff.
2. Set the change's status to closed in `proposal.md` with the integrating commit or pull request.
3. Move the folder with `git mv <specs>/changes/<change-id> <specs>/changes/archive/YYYY-MM-DD-<change-id>`, using the closing date. Never recreate the files by rewriting them.
4. Update `<specs>/project.md` only if the phase, an open decision, or a blocker changed.

Deliver these edits as one versioned change under the Git rules of the repository that holds `<specs>`. A change abandoned before integration is archived the same way without merging its deltas, with the reason in `proposal.md`.
