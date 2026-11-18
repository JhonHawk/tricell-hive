## ADDED Requirements

### Requirement: The close lists the project's determinism candidates

A work item's final report SHALL list, only when the work has one, each determinism candidate of the project: a repetition the work did not turn into a mechanism, namely an inline script or command chain the work ran two or more times for the same purpose, or a correction the user gave twice in the work, with what repeated, how many times and the proposed mechanism, without asking a new question. The close SHALL search the project's persistent memory for the candidate by its purpose and create or update its record, named after its tool and target, with the session and date. The close footer SHALL offer turning a candidate seen in two sessions into a mechanism as a reply of its own; a candidate is not a finding.

#### Scenario: Repeated inline script

- **GIVEN** a work item ran the same inline link-check script three times
- **WHEN** its close writes the final report
- **THEN** the report lists one candidate naming the script's purpose, the count and a proposed project mechanism, and the project's memory holds its record with this session

#### Scenario: Candidate seen in a second session

- **GIVEN** the project's memory already records a candidate from an earlier session
- **WHEN** a later work item's close finds the same candidate
- **THEN** the record counts two sessions and the close footer offers turning it into a mechanism

#### Scenario: No candidate

- **GIVEN** a work item repeated no script, command chain or correction
- **WHEN** its close writes the final report
- **THEN** the report has no candidates part and the close writes no candidate record
