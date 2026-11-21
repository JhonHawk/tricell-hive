## MODIFIED Requirements

### Requirement: Moving between stages shows the next-step block

Every move from one flow to the next SHALL end the report with the next-step block and its command, except the continuations its rule names. A reply to a question or the acceptance of a proposal SHALL only define or redirect scope and SHALL NOT start a flow by itself, except a fix the close footer or a findings round sends and a continuation the user's own request set. Research SHALL continue into build in the same turn only when the user's own request named the change and its target. How each message ends, the next-step block, the close footer and the same-turn continuations included, SHALL be stated in one rule leaf, `hive-rule-report-endings`, as a table with one row per way of ending, and every other file SHALL point to its row instead of restating it, except the minimal invariant the global guidance keeps so it acts before that leaf is read.

#### Scenario: A ticket picked for research

- **GIVEN** the user picks a ticket for research without asking for a change
- **WHEN** the agent's proposal is accepted with a reply such as "1"
- **THEN** the next message ends with the next-step block toward `flow-build` or `flow-plan`
- **AND** no build starts until the user's reply names it

#### Scenario: A close ends as its close mode sets

- **GIVEN** a deploy's cleanup close, which `hive-rule-close-modes` ends with the report and no footer
- **WHEN** `flow-close` reaches its last step
- **THEN** the message ends with the report and its waiting line, without the footer

### Requirement: The final report carries evidence and a try-it block

The final report SHALL carry a defect's regression test output seen failing and then passing, and, when the change can be seen or run, a block with the URL or command, the test account's index row, the steps and the expected result, in every delivery mode. Its parts SHALL be stated in one rule leaf, `hive-rule-close-report`, with one row per variant (close at the completion point, a pull request left open under `ready` or waiting for a reviewer, a deploy's cleanup close, `hold` and a pause outside it). Each work item SHALL get one final report, written by `flow-close` at its report step: `git-workflow` and `flow-deploy` SHALL hand what they observed to the close instead of writing it, and the end of unattended work SHALL summarize the session and point to each close's report. A later message with the result of a base CI run that ends after the close is not a second final report.

#### Scenario: Hold delivery

- **GIVEN** a change delivered under `hold`
- **WHEN** the agent reports it
- **THEN** the report holds the failing and passing output and the try-it block

#### Scenario: A pull request merged by the agent

- **GIVEN** a delivery whose pull request the agent merged
- **WHEN** `git-workflow` verifies the merge and runs `flow-close`
- **THEN** one final report is written, by `flow-close`, with the row for the close at the completion point
