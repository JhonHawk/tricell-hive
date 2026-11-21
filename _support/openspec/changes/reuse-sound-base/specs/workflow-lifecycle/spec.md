## MODIFIED Requirements

### Requirement: Defects are fixed at their cause

Before the first edit of a fix for a defect, in a build task, a round of fixes, an urgent production fix or a child's task, once a failing check is judged to show wrong code, and when a plan weighs the approaches to such a fix, the agent SHALL name the defect's cause, the first place the code does what it should not or fails to handle a state it should handle, shown by a reproduction, a regression test seen failing or the code read, and SHALL write it in the task's line or the report. It SHALL fix the defect at that cause with the smallest change that removes it, also when a finding names only the place where the symptom surfaced, wherever it lies in the repository, and SHALL search the cause's literal values and callers for the same fault, fixing each instance in the files the work already changes and reporting the others as findings. A guard, retry, fallback, default value, clamp, swallowed error or discarded state added where the symptom surfaced SHALL NOT count as a fix, unless the code's documented contract, schema or specification allows that state or it comes from input that crosses a trust boundary. While an incident needs its impact reduced, the agent SHALL mitigate first within existing authority. Otherwise, when fixing the cause is a choice the global «Scope and authorization» makes a blocker for a defect that blocks work, such as a contract, stored data or a fix that is not a small understood change, the agent SHALL report it before writing any workaround, so the user chooses between fixing the cause now and a workaround; when the cause lies in a project outside the current one, the agent SHALL ask the question «Scope and authorization» sets for code in an external project, with a workaround here as the other option. A mitigation or a workaround the user accepts SHALL be the smallest fix that holds; a code change SHALL be marked with one comment line that names the cause; each SHALL be named in the report as a workaround with the cause's location, and fixing the cause SHALL be proposed as a ticket unless the same work fixes it. A child SHALL fix only within the paths its brief names and SHALL report other instances, workarounds and outside causes to its parent. The code review SHALL check for a fix placed where the symptom surfaced instead of at its cause, taking as its failure scenario the bad state the cause still produces for its other consumers, SHALL NOT rate a fix at the cause, with the other instances fixed with it, as a change the task did not ask for, and SHALL rate each finding by its observable impact. The rule SHALL live in the `hive-core` rule leaf `defect-fix`, which the global «Rule leaves» index, `flow-build`, the «Reuse before writing» leaf, the round of fixes of `flow-verify` and the urgent fixes of `flow-deploy` link with the condition that requires reading it.

#### Scenario: A null crash in a view

- **GIVEN** a view crashes on a null field that the API should never return
- **WHEN** the agent fixes the defect
- **THEN** it names the code that produces the null as the cause and fixes it there
- **AND** a null check added in the view alone is not reported as the fix

#### Scenario: A cause in another project

- **GIVEN** the cause of a defect is in a service of a project outside the current one
- **WHEN** the agent is about to fix it
- **THEN** before writing any workaround it asks whether to continue there from this session or open a session in that project, with a workaround here as the other option
- **AND** a workaround the user accepts is marked with one comment line naming the cause, named as a workaround in the report, and a ticket is proposed for the cause

#### Scenario: A cause in an untouched module of the same repository

- **GIVEN** a view crashes because a helper in a module the work has not touched returns a value its contract forbids
- **WHEN** the agent fixes the defect
- **THEN** it fixes the helper, without a guard in the view and without proposing a ticket

## ADDED Requirements

### Requirement: New code builds only on a sound base

Before writing new code, and when a plan weighs approaches, the agent SHALL reuse what the codebase already has, such as a helper, component or pattern, only when it is a sound base. A base SHALL NOT count as sound when it is a workaround, such as one a comment marks or a guard the «Defects are fixed at their cause» requirement does not count as a fix; when it breaks the repository's `CODING_STANDARDS.md`, its lint configuration or a convention its documentation states; when using it needs a boolean flag or a branch inside it that only the new code turns on; or when the same logic lives elsewhere in the repository with diverging behavior. The agent SHALL NOT copy or extend such a base, and SHALL name it in the report with the condition it meets. It SHALL fix the base first, within the paths a child's brief names, keeping what the base's other callers observe as a characterization test pins it, when that fix is a small understood change, and that fix SHALL count as part of the task, not as a module touched against the request; otherwise a plan SHALL offer that fix as an approach, and a build SHALL report it before building on it, a child to its parent, with the choice the «Defects are fixed at their cause» requirement gives for a cause, the base in its place; building on the base once the user accepts SHALL be marked, reported and proposed as a ticket as that requirement sets for a workaround. The code review SHALL check for new code that copies or extends a base that is not sound, taking that base's spread as the failure scenario, and SHALL NOT rate a fix of the base the rule has the author make first as a change the task did not ask for. The rule SHALL live in the `hive-core` rule leaf `reuse-before-writing`, and no other file SHALL restate its conditions.

#### Scenario: Copying an existing workaround

- **GIVEN** the surrounding code repeats a retry that a comment marks as a temporary workaround
- **WHEN** a task writes new code in that area
- **THEN** the new code does not copy the retry and the report names it

#### Scenario: A helper that needs a flag for the new caller

- **GIVEN** a formatting helper used by three views would need a boolean `isExport` branch to serve a new export view
- **WHEN** a build task without a plan writes the export view and fixing the helper is a small understood change
- **THEN** the agent first restructures the helper, with its output for the three views pinned by a characterization test, and the export view uses the restructured helper
- **AND** no `isExport` flag is added to the helper

#### Scenario: Duplicated logic that diverges, too large to fix in passing

- **GIVEN** two modules compute a price total in different ways and a plan's new feature needs that total
- **WHEN** the plan weighs approaches
- **THEN** unifying the total first is one of the approaches the user chooses between
- **AND** the plan does not silently copy either version

#### Scenario: A base that breaks the project's standards

- **GIVEN** `CODING_STANDARDS.md` forbids raw SQL in views and an existing view builds a query by string concatenation
- **WHEN** a task writes a new view that needs the same data
- **THEN** the new view does not copy the concatenated query, and the report names the existing view and the standard it breaks
