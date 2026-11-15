## ADDED Requirements

### Requirement: Defects are fixed at their cause

Before the first edit of a fix for a defect, in a build task, a round of fixes or a child's task, the agent SHALL name the defect's cause, the first place the code does what it should not or fails to handle a state it should handle, shown by a reproduction, a regression test seen failing or the code read, and SHALL write it in the task's line or the report. It SHALL fix the defect at that cause with the smallest change that removes it, also when a review finding names only the place where the symptom surfaced, wherever it lies in the repository, and SHALL search the cause's literal values and callers for the same fault, fixing each instance in the files the work already changes and reporting the others as findings. A guard, retry, fallback, default value, clamp, swallowed error or discarded state added where the symptom surfaced SHALL NOT count as a fix, unless the code's contract, types, schema or documentation allow that state or it comes from input that crosses a trust boundary. When fixing the cause needs another repository, or a change to a contract or stored data, the agent SHALL report it as a blocker before writing any workaround, so the user chooses between fixing the cause now and a workaround. A workaround the user accepts, and a mitigation that changes code while an incident needs its impact reduced, SHALL be the smallest fix that holds, marked with one comment line that names the cause and named in the report as a workaround with the cause's location, and fixing the cause SHALL be proposed as a ticket unless the same work fixes it. A child SHALL fix only within the paths its brief names and SHALL report other instances, workarounds and outside causes to its parent. New code SHALL NOT copy an existing pattern that is a workaround, such as one a comment marks or a guard for a state the code's contract forbids, and SHALL name that pattern in the report. The code review SHALL check for a fix placed where the symptom surfaced instead of at its cause and for new code that copies a workaround, taking as their failure scenario the bad state the cause still produces for its other consumers or the workaround's spread, SHALL NOT rate a fix at the cause, with the other instances fixed with it, as a change the task did not ask for, and SHALL rate each finding by its observable impact. The rule SHALL live in a reference of `flow-build` that `flow-build`, its «Reuse before writing» section, the round of fixes of `flow-verify` and the urgent fixes of `flow-deploy` link with the condition that requires reading it.

#### Scenario: A null crash in a view

- **GIVEN** a view crashes on a null field that the API should never return
- **WHEN** the agent fixes the defect
- **THEN** it names the code that produces the null as the cause and fixes it there
- **AND** a null check added in the view alone is not reported as the fix

#### Scenario: A cause in another repository

- **GIVEN** the cause of a defect is in a service of another repository
- **WHEN** the agent is about to fix it
- **THEN** it reports the blocker before writing any workaround, and the user chooses
- **AND** a workaround the user accepts is marked with one comment line naming the cause, named as a workaround in the report, and a ticket is proposed for the cause

#### Scenario: A cause in an untouched module of the same repository

- **GIVEN** a view crashes because a helper in a module the work has not touched returns a value its contract forbids
- **WHEN** the agent fixes the defect
- **THEN** it fixes the helper, without a guard in the view and without proposing a ticket

#### Scenario: Copying an existing workaround

- **GIVEN** the surrounding code repeats a retry that a comment marks as a temporary workaround
- **WHEN** a task writes new code in that area
- **THEN** the new code does not copy the retry and the report names it
