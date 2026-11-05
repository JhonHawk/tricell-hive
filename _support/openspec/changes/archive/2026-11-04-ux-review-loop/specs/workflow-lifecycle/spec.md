# Delta for workflow-lifecycle

## ADDED Requirements

### Requirement: UX review in one pass

A UI change's `hive-review-ux` child SHALL receive, with the walk's captures, the means to reach each state its role always checks, or the states that cannot be reached. It SHALL list a state it cannot reach as not covered instead of failing its verdict. Its verdict SHALL judge only the defects the change introduced in the reviewed regions: passed when the reviewed states show no introduced defect of any priority, and not passed when no state could be reviewed. Defects already there SHALL be reported in a separate list. A correction of one of its findings SHALL be reread by a follow-up limited to the corrected regions, the layouts that contain them and their loading, empty and error states, with the automated audit and the layout, keyboard, focus and accessible-name checks on the longest values, never by a new full UX review. In a deployment walk, the reviewer SHALL get the same brief with the environment's URLs in place of the captures.

#### Scenario: State the brief cannot reach

- **GIVEN** a UI change whose error state needs a fixture the brief does not provide
- **WHEN** the UX reviewer cannot reach that state
- **THEN** it lists the state as not covered with what blocked it, and its verdict covers the reviewed states
- **AND** the producer covers that state in its own walk or reports it as not verified, in the validation round or, when the delivery does not stop for validation, in the final report

#### Scenario: Deployment walk

- **GIVEN** a deployment walk of a UI change
- **WHEN** the producer launches the UX reviewer alongside the functional walk
- **THEN** its brief follows the reviewer brief with the environment's URLs in place of the captures, and the reviewer takes its own captures

#### Scenario: Defect already there

- **GIVEN** a reviewed region with a P3 defect that the base branch already shows
- **WHEN** the reviewer finds no defect the change introduced
- **THEN** the verdict is passed and the P3 defect appears in the list of defects already there, with its priority
- **AND** the defect reaches the user in the validation round or, when the delivery does not stop for validation, in the close's findings

#### Scenario: No state could be reviewed

- **GIVEN** a UX review whose brief gives no reachable state
- **WHEN** the reviewer cannot review any state
- **THEN** the verdict is not passed and the report says the review did not run

#### Scenario: Correction of a UX finding

- **GIVEN** the producer fixed a defect the UX reviewer reported
- **WHEN** the fix is reread
- **THEN** the same child reviews only the corrected regions and states, the layouts that contain them and their loading, empty and error states, still running the automated audit and the measured checks with the longest values, and answers in one reply
- **AND** when the host has no tool to message that child, or the message returns an error, a new UX reviewer gets the corrected findings and only those regions and states
