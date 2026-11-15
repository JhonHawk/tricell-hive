## ADDED Requirements

### Requirement: Defects are fixed at their cause

Before the first edit of a fix for a defect, in a build task, a round of fixes or a child's task, the agent SHALL name the defect's cause, the first place the code does what it should not or fails to handle a state it should handle, shown by a reproduction, a regression test seen failing or the code read, and SHALL write it in the task's line or the report. It SHALL fix the defect at that cause with the smallest change that removes it, also when a review finding names only the place where the symptom surfaced, and SHALL search the cause's literal values and callers for the same fault, fixing each instance in the modules the work already touches and reporting the others as findings. A guard, retry, fallback, default value, clamp, swallowed error or discarded state added where the symptom surfaced SHALL NOT count as a fix, unless the system may legitimately reach that state. When the cause lies in a module the work does not touch, in another repository, or needs a change to a contract or stored data, and while an incident needs its impact reduced before the cause is fixed, the agent SHALL make the smallest fix inside the scope, mark it in the code with one comment line that names the cause, name it in the report as a workaround with the cause's location, and propose fixing the cause as a ticket; a change to a contract, stored data or the scope SHALL be the user's choice. A child SHALL fix only within the paths its brief names and SHALL report other instances, workarounds and outside causes to its parent. New code SHALL NOT copy an existing pattern that is a workaround, such as one a comment marks or a guard for a state that should never exist, and SHALL name that pattern in the report. The code review SHALL check for a fix placed where the symptom surfaced instead of at its cause and for new code that copies a workaround, taking as their failure scenario the bad state the cause still produces for its other consumers or the workaround's spread, SHALL NOT rate a fix at the cause as a change the task did not ask for, and SHALL rate each finding by its observable impact. The rule SHALL live in a reference of `flow-build` that `flow-build`, its «Reuse before writing» section, the round of fixes of `flow-verify` and the urgent fixes of `flow-deploy` link with the condition that requires reading it.

#### Scenario: A null crash in a view

- **GIVEN** a view crashes on a null field that the API should never return
- **WHEN** the agent fixes the defect
- **THEN** it names the code that produces the null as the cause and fixes it there
- **AND** a null check added in the view alone is not reported as the fix

#### Scenario: A cause in another team's service

- **GIVEN** the cause of a defect is in a service outside the work's scope
- **WHEN** the agent fixes it
- **THEN** it makes the smallest fix inside the scope, marks it with one comment line naming the cause, names it as a workaround in the report and proposes a ticket for the cause

#### Scenario: Copying an existing workaround

- **GIVEN** the surrounding code repeats a retry that a comment marks as a temporary workaround
- **WHEN** a task writes new code in that area
- **THEN** the new code does not copy the retry and the report names it
