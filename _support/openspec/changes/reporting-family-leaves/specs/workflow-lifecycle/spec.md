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

## ADDED Requirements

### Requirement: Playbooks and the close hygiene leaf name their own row

Each of the 13 playbooks of `hive-mode` SHALL carry one reporting line that links its row of `hive-rule-report-endings` and the row of `hive-rule-close-report` that the delivery mode in force picks, without naming a delivery mode; `pr-review-deploy` SHALL also link the deployment row. A playbook SHALL NOT copy text from those leaves. The `hive-rule-close-hygiene` leaf and its line in the global guidance SHALL cover steps 4, 5 and 7 of `flow-close` and SHALL NOT claim step 6, which belongs to `hive-rule-close-report`. Aligning each line of the global index with the first «Read» line of its leaf, and the test for it, SHALL belong to the change `hive-core-stage-packets`, not to this one.

#### Scenario: A playbook reports through its rows

- **GIVEN** the `bugfix` playbook
- **WHEN** the work closes
- **THEN** the playbook's reporting line leads to the close row of `hive-rule-report-endings` and to the row of `hive-rule-close-report` that the delivery mode in force picks

#### Scenario: Close hygiene stays out of the report step

- **GIVEN** the index line and the leaf of `hive-rule-close-hygiene`
- **WHEN** a close reaches step 6
- **THEN** neither the line nor the leaf asks to be read for the final report

### Requirement: The two reporting homes have bounded size and reading cost

`hive-rule-report-endings` SHALL have its own size limit of 9,000 bytes and `hive-rule-close-report` SHALL stay in the 7,000-byte exception; every other leaf keeps its limit and no text SHALL be cut to fit. The reading-footprint test for the plan and build route SHALL accept the growth that comes from reading those two leaves whole at each stage's end, as a second allowance of 8,880 bytes beside the earlier 182 bytes. The global guidance SHALL NOT grow past its size at the base of the build (28,593 bytes at `7f6277e0`).

#### Scenario: The leaves fit their limits

- **GIVEN** the candidate
- **WHEN** the content tests run
- **THEN** `hive-rule-report-endings` is at most 9,000 bytes, `hive-rule-close-report` at most 7,000 bytes, and the global guidance at most 28,593 bytes
