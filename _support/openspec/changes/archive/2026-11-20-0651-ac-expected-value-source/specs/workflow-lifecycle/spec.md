## MODIFIED Requirements

### Requirement: Criteria close on the real effect

Each acceptance criterion SHALL end with `Observed in: <place>`, the place where its real effect shows. For a screen or an API the project serves, the candidate served locally with the project's real build and seeded data of realistic volume SHALL count as the real effect. A mock, a stub or an emulation of a service the project does not own, and a dry run, are substitutes, and a check against one SHALL NOT close a criterion. Until a walk observes a criterion in its place, its task SHALL keep it in its `Closes:` line and report it as not verified. The walk of verify SHALL close each criterion whose place it reaches; a criterion whose place is a deployed environment SHALL stay not verified until that deployment's functional walk observes it, and naming the environment SHALL NOT declare that walk. A criterion SHALL name a deployed environment only when its effect exists nowhere else. A criterion without the marker SHALL be verified through its real path. A criterion whose observation reads data the change shows or produces, such as a count, a total, a set of records or a generated file, SHALL name the expected value and a source of it that does not run the code under test, such as the seed or fixture that defines it or a database query, and the walk SHALL compare what it observes with that value. `check_change.py` SHALL print one finding for each active criterion none of whose bullets, read with the lines indented under it, names a place after the marker, and SHALL NOT evaluate the marker with `--archive`.

#### Scenario: Permission checked against a mock

- **GIVEN** a criterion observed in a deployed account
- **WHEN** a task checks the permission with a local mock and the check passes
- **THEN** the task's line names the criterion as not verified until the deployment
- **AND** a task verifier returns `cannot verify` for it, naming the substitute, and does not return `not met` because of the plan

#### Scenario: Screen served locally

- **GIVEN** a criterion observed in a screen the repository serves
- **WHEN** verify walks the candidate served locally with the project's real build and seeded data of realistic volume
- **THEN** that walk closes the criterion

#### Scenario: Report whose data looks right

- **GIVEN** a criterion that a monthly report lists the completed tasks of the seeded workspace
- **WHEN** the plan writes the criterion
- **THEN** it names the expected tasks and their source, such as the seed or a database query, not the report under test
- **AND** verify's walk closes the criterion only when the report shows exactly those tasks

#### Scenario: Effect that exists only after deployment

- **GIVEN** a criterion whose place is an environment no build task reaches
- **WHEN** the build and verify finish
- **THEN** the task that builds the behavior names the criterion in `Closes:`, and every report names it as not verified until the deployment
- **AND** the deployment's report names it as not verified, until the functional walk of that environment runs and closes it

#### Scenario: Archiving a folder written before the marker

- **GIVEN** a finished change folder whose criteria carry no `Observed in:` marker
- **WHEN** it is checked with `--archive`
- **THEN** the missing marker produces no finding

