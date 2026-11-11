## MODIFIED Requirements

### Requirement: Design record

The design record SHALL be a `DESIGN.md` at the root of the app that owns the UI, one per app in a monorepo, in the Google Labs DESIGN.md format plus a `## Hive` section holding the library, its installed version, the customization layer, the library tools, the product record, the recorded decisions and each source file with its fingerprint, and, when the user declares one, the app's design file with its tool and location. Every recorded value SHALL name its source file, and no value SHALL be invented; the design file is the one value the user gives instead of a source file, and it SHALL be written only at the user's instruction. When the stage that draws screen examples used a file the record does not declare, that stage SHALL ask the user once whether to record it, and a yes SHALL be that instruction: the stage SHALL keep it in the approved example's record, and the plan for that screen SHALL include the task that writes the field, delivered with the change. When the user corrects an example with a convention that applies beyond that screen, that stage SHALL ask once whether to record it among the recorded decisions, and a yes SHALL be the instruction to write it there with its date and, as its origin, where the example's record is kept: the change folder that carries it, otherwise the Figma file and the frame's node; the plan for that screen SHALL include that task, delivered with the change. The design file SHALL say where examples for the app's screens are drawn and SHALL NOT by itself authorize a write to that file. The record SHALL be stale when the installed version differs, a source fingerprint changed or its file is gone, or a new customization piece appears; the design file carries no fingerprint and SHALL NOT make the record stale. A stale record SHALL be named with what changed and SHALL NOT be used silently. The record SHALL be written or refreshed only through a task of a UI change's plan or at the user's request, and SHALL NOT be overwritten without showing the change.

#### Scenario: Missing record in a UI plan
- **GIVEN** a plan for a change to a screen in a project with no `DESIGN.md`
- **WHEN** the plan settles its UI decisions
- **THEN** the plan includes a task that writes the record, delivered with the change

#### Scenario: Theme file changed after the record
- **GIVEN** a `DESIGN.md` whose recorded fingerprint of the theme file no longer matches that file
- **WHEN** the skill reads the record
- **THEN** it names the changed file, proposes the refresh, and does not use the stale tokens silently

#### Scenario: No design system
- **GIVEN** a project with no tokens, theme or component library
- **WHEN** the skill writes the record
- **THEN** it states that no design system exists and labels any proposed starter values as a proposal, not as project tokens

#### Scenario: Design file declared by the user
- **GIVEN** the user names the design file of an app and asks to record it
- **WHEN** the record is written
- **THEN** its `## Hive` section holds the design file with its tool and location
- **AND** a later session recommends drawing examples in that file without asking which file to use

#### Scenario: Design file used but not declared
- **GIVEN** a stage that drew screen examples in a file the design record does not declare
- **WHEN** it asks the user to approve the example
- **THEN** it asks in the same round whether to record the file in the design record
- **AND** on a yes the plan for that screen includes the task that writes the design file, with its tool and location, in the `## Hive` section

#### Scenario: Convention set while correcting an example
- **GIVEN** a user who corrects an example with a convention for every screen of that kind in the app
- **WHEN** the stage asks the approval and the user says yes to recording it
- **THEN** the plan for that screen includes the task that writes the convention among the recorded decisions, with its date and, as its origin, the change folder that carries the example's record
