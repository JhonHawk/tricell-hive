# ui-criteria

## Purpose

How Hive detects a project's component library and keeps its design record, and which UI criteria design, build and review apply below the project's own design system.

## Requirements

### Requirement: Component library detection and precedence

The `hive-ui-criteria` skill SHALL identify, for any component library, the library the changed screens import, its installed version from the lockfile or the installed package rather than the manifest range, its customization layer with the path of each piece, and the library's MCP server or skill when installed, and SHALL read that version's documentation before stating an API or token. Its decisions SHALL follow this precedence: an accessibility requirement the project or the ticket declares, then recorded project decisions, then the customization layer, then the library defaults, then the skill's defaults. Its accessibility floors SHALL apply to what the change builds or customizes, and no recorded decision, customization or library default SHALL excuse a floor failure there. An element the change reuses unmodified that fails a floor SHALL NOT be a defect of the change: it SHALL be listed among the defects already there with its reason, and, when an accessibility requirement is declared, with its priority and marked as failing that requirement. Any change that alters a token or an installed component SHALL return as a pending decision unless the brief or a recorded decision names it.

#### Scenario: Manifest range differs from the installed version
- **GIVEN** a project whose manifest declares `^2.4.0` for its component library and whose lockfile resolves `2.8.5`
- **WHEN** the skill detects the library
- **THEN** it records `2.8.5` with the lockfile path as its source

#### Scenario: A library default meets a skill default
- **GIVEN** a library whose default button is 36 px high and the skill's default touch target of 44 px
- **WHEN** a screen is designed or reviewed
- **THEN** the library default holds and is not reported as a deviation

### Requirement: Design record

The design record SHALL be a `DESIGN.md` at the root of the app that owns the UI, one per app in a monorepo, in the Google Labs DESIGN.md format plus a `## Hive` section holding the library, its installed version, the customization layer, the library tools, the product record, the recorded decisions and each source file with its fingerprint, and, when the user declares them, the app's design file with its tool and location and the name of the design tool's published library that holds the app's components and variables. Every recorded value SHALL name its source file, and no value SHALL be invented; the design file and the design library are the values the user gives instead of a source file, and each SHALL be written only at the user's instruction. When the stage that draws screen examples used a file the record does not declare, that stage SHALL ask the user once whether to record it, and a yes SHALL be that instruction: the stage SHALL keep it in the approved example's record, and the plan for that screen SHALL include the task that writes the field, delivered with the change. When the user corrects an example with a convention that applies beyond that screen, that stage SHALL ask once whether to record it among the recorded decisions, and a yes SHALL be the instruction to write it there with its date and, as its origin, where the example's record is kept: the change folder that carries it, otherwise the Figma file and the frame's node; the plan for that screen SHALL include that task, delivered with the change. The design file SHALL say where examples for the app's screens are drawn and SHALL NOT by itself authorize a write to that file. The record SHALL be stale when the installed version differs, a source fingerprint changed or its file is gone, or a new customization piece appears; the design file and the design library carry no fingerprint and SHALL NOT make the record stale. A stale record SHALL be named with what changed and SHALL NOT be used silently. The record SHALL be written or refreshed only through a task of a UI change's plan or at the user's request, and SHALL NOT be overwritten without showing the change.

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

#### Scenario: Design library declared by the user
- **GIVEN** the user names the published library that holds an app's components and asks to record it
- **WHEN** the record is written
- **THEN** its `## Hive` section holds the design library by the name the design tool lists
- **AND** a later stage that draws examples for that app uses instances of that library's components and defines its own pieces only for what the library lacks

#### Scenario: Design file used but not declared
- **GIVEN** a stage that drew screen examples in a file the design record does not declare
- **WHEN** it asks the user to approve the example
- **THEN** it asks in the same round whether to record the file in the design record
- **AND** on a yes the plan for that screen includes the task that writes the design file, with its tool and location, in the `## Hive` section

#### Scenario: Convention set while correcting an example
- **GIVEN** a user who corrects an example with a convention for every screen of that kind in the app
- **WHEN** the stage asks the approval and the user says yes to recording it
- **THEN** the plan for that screen includes the task that writes the convention among the recorded decisions, with its date and, as its origin, the change folder that carries the example's record

### Requirement: Measured UI criteria in review

The UX review SHALL report as a P0–P3 defect, only when the reviewed change introduced it, each check the skill marks as measured: the accessibility floors and the measurable checks of collections, copy and extremes. Each floor SHALL name the standard that sets it, with its conformance level. An unmodified library default and a recorded decision SHALL follow the precedence of the detection requirement, and the screen mode and refinement guidance SHALL NOT be judged as defects.

#### Scenario: Filters without a reset
- **GIVEN** a change that adds filters to a list with no control to reset them
- **WHEN** the UX review checks the list
- **THEN** it reports the missing reset as a defect with its priority

#### Scenario: Library default control border
- **GIVEN** a control whose border contrast is the library default
- **WHEN** the UX review checks non-text contrast
- **THEN** it does not report the border as a defect

#### Scenario: Library default focus ring below the floor
- **GIVEN** a project with no declared accessibility requirement and an unmodified library focus ring whose contrast is below 3:1
- **WHEN** the UX review checks the changed region
- **THEN** it lists the focus ring among the defects already there with its reason, does not fail the change's verdict, and a fix that edits the theme or the installed component returns as a pending decision

#### Scenario: Declared accessibility requirement
- **GIVEN** a project whose `DESIGN.md` declares a WCAG 2.2 AA requirement and the same library focus ring
- **WHEN** the UX review checks the changed region
- **THEN** it lists the focus ring among the defects already there with its priority, marked as failing the declared requirement, and the change's verdict covers only the defects it introduced
