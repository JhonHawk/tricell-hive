## MODIFIED Requirements

### Requirement: Design record

The design record SHALL be a `DESIGN.md` at the root of the app that owns the UI, one per app in a monorepo, in the Google Labs DESIGN.md format plus a `## Hive` section holding the library, its installed version, the customization layer, the library tools, the product record, the recorded decisions and each source file with its fingerprint, and, when the user declares one, the app's design file with its tool and location. Every recorded value SHALL name its source file, and no value SHALL be invented; the design file is the one value the user gives instead of a source file, and it SHALL be written only at the user's instruction. The design file SHALL say where proposals for the app's screens are drawn and SHALL NOT by itself authorize a write to that file. The record SHALL be stale when the installed version differs, a source fingerprint changed or its file is gone, or a new customization piece appears; the design file carries no fingerprint and SHALL NOT make the record stale. A stale record SHALL be named with what changed and SHALL NOT be used silently. The record SHALL be written or refreshed only through a task of a UI change's plan or at the user's request, and SHALL NOT be overwritten without showing the change.

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
- **AND** a later session offers proposals in that file without asking which file to use
