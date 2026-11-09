## ADDED Requirements

### Requirement: Decisions carry labelled context

A decision `D<n>` shown to the user SHALL carry, under its title line, a `Situation:` line that says what happened or what was found, what was verified and what to check, with its target, in words the user has used or that the same message explains; under each option, an `If you choose it:` line that says what the user gets and what it costs them in time, money, risk or what stops working; and an `I recommend <letter>:` line whose reason is a consequence for the user and never a quotation of a rule or a skill. The three labels SHALL be written in the session language, while the IDs, letters and the `(★)` mark stay as they are. The round's waiting line SHALL also say how many known decisions are still to come. In a native question card, the question SHALL carry the situation, each option's description the text of its `If you choose it:` line, or the option's effect in one line when it has none, and the starred option's description SHALL end with the reason of the recommendation. The "Open the viewer" option that a validation question carries SHALL have no `If you choose it:` line. A finding `H<n>` asked in a round carries its title and option lines without the three labelled lines, and the replies of a next-step block keep their own format.

#### Scenario: A rule requires the question

- **GIVEN** a rule requires the agent to ask before an effect
- **WHEN** the agent asks it as a decision
- **THEN** the recommendation line states the effect that needs the answer, such as its cost or what it changes
- **AND** it quotes no rule or skill as its reason

#### Scenario: A name the user was not given

- **WHEN** the situation would name a task ID, a file, a branch or an internal name the user has not been given
- **THEN** the same message explains it, or the situation leaves it out

### Requirement: Criteria close on the real effect

Each acceptance criterion SHALL end with `Observed in: <place>`, the place where its real effect shows. For a screen or an API the project serves, the candidate served locally with the project's real build and seeded data of realistic volume SHALL count as the real effect. A mock, a stub or an emulation of a service the project does not own, and a dry run, are substitutes, and a check against one SHALL NOT close a criterion. Until a walk observes a criterion in its place, its task SHALL keep it in its `Closes:` line and report it as not verified. The walk of verify SHALL close each criterion whose place it reaches; a criterion whose place is a deployed environment SHALL stay not verified until that deployment's functional walk observes it, and naming the environment SHALL NOT declare that walk. A criterion SHALL name a deployed environment only when its effect exists nowhere else. A criterion without the marker SHALL be verified through its real path. `check_change.py` SHALL print one finding for each active criterion none of whose bullets, read with the lines indented under it, names a place after the marker, and SHALL NOT evaluate the marker with `--archive`.

#### Scenario: Permission checked against a mock

- **GIVEN** a criterion observed in a deployed account
- **WHEN** a task checks the permission with a local mock and the check passes
- **THEN** the task's line names the criterion as not verified until the deployment
- **AND** a task verifier returns `cannot verify` for it, naming the substitute, and does not return `not met` because of the plan

#### Scenario: Screen served locally

- **GIVEN** a criterion observed in a screen the repository serves
- **WHEN** verify walks the candidate served locally with the project's real build and seeded data of realistic volume
- **THEN** that walk closes the criterion

#### Scenario: Effect that exists only after deployment

- **GIVEN** a criterion whose place is an environment no build task reaches
- **WHEN** the build and verify finish
- **THEN** the task that builds the behavior names the criterion in `Closes:`, and every report names it as not verified until the deployment
- **AND** the deployment's report names it as not verified, until the functional walk of that environment runs and closes it

#### Scenario: Archiving a folder written before the marker

- **GIVEN** a finished change folder whose criteria carry no `Observed in:` marker
- **WHEN** it is checked with `--archive`
- **THEN** the missing marker produces no finding
