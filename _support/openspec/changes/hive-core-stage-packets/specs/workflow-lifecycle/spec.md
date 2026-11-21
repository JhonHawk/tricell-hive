## MODIFIED Requirements

### Requirement: Checks with saved output run through a capped helper

A check whose output the agent saves to a file SHALL run through `hive-core/scripts/checks/run_check.py`, which SHALL take the command as arguments without a shell, write its whole output to a log file under a size cap, stop the command's process group when the cap is passed, close the log before printing its last lines or the lines that match a pattern, and exit with the command's own status, or with 125 when the cap stopped it. The agent SHALL NOT write a shell wrapper that redirects a block's output to a file and reads that file in the same command. The session auditor SHALL fail a shell command in which a compound block or `exec` redirects output to a file that a trap or a command inside that block reads, including a Codex command whose text is interpolated, and SHALL NOT fail a simple command's redirection.

#### Scenario: A failing check under the helper

- **GIVEN** a test run that fails after printing a thousand lines
- **WHEN** the agent runs it through the helper
- **THEN** the helper prints the log's path, then the last 40 lines after the run ends, and exits with the test run's status
- **AND** the log holds the thousand lines and no copy of its own end

#### Scenario: A runaway output

- **GIVEN** a command that writes without end
- **WHEN** it runs through the helper with a 1 MiB cap
- **THEN** the helper stops its process group, the log stays near 1 MiB, and the helper exits with 125 naming the log and the cap

### Requirement: The session audit flags an analysis that ends without the reply block

`go run ./tests/pilot --audit-session` SHALL report `analysis_reply_block` for Claude, Codex and Grok sessions. It SHALL fail a main-session turn whose last assistant text ends with two or more list items labeled in italics by something other than a single letter, or with lettered items and no line above them holding `▶` or a question in bold. It SHALL also fail a main-session turn that loaded `hive-mode`, through the skill call or a read of its installed `SKILL.md`, whose last assistant text offers no replies as a list, has six or more non-blank lines and holds no `▶`, no question in bold among its last three non-blank lines and no flow command, when that turn changed no file outside the host's temporary location and delivered no work. A file change counts through a write or delete tool the host did not refuse, an output redirection, an in-place `sed` or `perl` edit, `find -delete` or a writer run by `find -exec`, or the files `tee`, `mv`, `rm`, `rmdir`, `mkdir`, `touch` and `truncate` name and the destination of `cp`, `ln` and `install`, including behind `bash -c`, `env`, `sudo`, `nohup` or `xargs`. The shell line is read as tokens, so text inside quotes and the body of a here-document fed to a program other than a shell are not commands. A relative path resolves against an earlier `cd` in the same line, and a path it cannot resolve counts as outside. Delivering work means a git subcommand that changes the working tree, the history or a remote, `gh pr create`, `gh pr merge` or `git_state.py commit`. A shell command whose text the trace withholds entirely, or that cannot be tokenized, counts as work; one the trace keeps as a template is read with each interpolated value unknown. Its evidence SHALL say whether the installed rule leaf `report-endings` was read earlier in the session, at its former path, at its `hive-core` path or in a stage packet that holds it. It SHALL NOT judge a turn that ends with a question card, a flow command or a decision round written in text, nor a child session's turns, and SHALL change no other criterion's result.

#### Scenario: Replies offered as an unlettered list
- **WHEN** an analysis ends with "Puedes responder:" followed by two bullets labeled in italics
- **THEN** `analysis_reply_block` fails at that line and says whether the report-endings rule was read

#### Scenario: The reply block in its format
- **WHEN** an analysis ends with a bold heading holding `▶` and lettered options with the recommended one starred
- **THEN** `analysis_reply_block` does not fail

#### Scenario: An analysis that offers no replies
- **WHEN** a turn that loaded `hive-mode` and edited no file ends a report of six or more lines with "Queda pendiente tu visto bueno" and no block
- **THEN** `analysis_reply_block` fails at that line, says the analysis ended without a reply block and whether the report-endings rule was read

#### Scenario: A short answer or a work turn
- **WHEN** a turn ends with fewer than six non-blank lines, changed a file outside the host's temporary location before its last text, for example with `sed -i`, or did not load `hive-mode`
- **THEN** `analysis_reply_block` does not judge that turn for a missing block

### Requirement: A background wait ends on every outcome and has a deadline

Before starting a polling loop or a background command that waits on a process or an external job, the agent SHALL read the rule leaf `background-wait`. CI waits SHALL keep following the merge gate, and child waits the child brief. The agent SHALL prefer a command that blocks until the job reaches a final state and exits with its status. A hand-built wait SHALL end on success, on failure, on a target that does not exist yet, on a process that hangs and on output it does not recognize. Unrecognized output SHALL end the wait with a non-zero status and print what was seen. The wait SHALL compare a structured field or an exit status rather than free text, or SHALL first check a text pattern against the query's actual output. A query that reports the job does not exist yet SHALL keep the wait on the same command. Every run of such a wait SHALL have a deadline that bounds the whole loop, not one iteration. When the host wakes the agent only when the command exits, that deadline SHALL be 5 to 10 minutes. On waking, the agent SHALL check progress. While the job advances, the agent SHALL rerun the same command for another stretch. After two stretches in a row without progress, the agent SHALL stop and report a blocker with the last state seen.

#### Scenario: A deploy status query prints an extra line

- **GIVEN** a cloud build whose status query with `--output text` prints `SUCCEED` followed by a `None` pagination line
- **WHEN** the agent waits in the background for the build to finish
- **THEN** the wait ends within its deadline, on the final state or on the unrecognized output, and the agent reports what it saw

### Requirement: Moving between stages shows the next-step block

Every move from one flow to the next SHALL end the report with the next-step block and its command, except the continuations its rule names. A reply to a question or the acceptance of a proposal SHALL only define or redirect scope and SHALL NOT start a flow by itself, except a fix the close footer or a findings round sends and a continuation the user's own request set. Research SHALL continue into build in the same turn only when the user's own request named the change and its target. How each message ends, the next-step block, the close footer and the same-turn continuations included, SHALL be stated in one rule leaf, `report-endings`, as a table with one row per way of ending, and every other file SHALL point to its row instead of restating it, except the minimal invariant the global guidance keeps so it acts before that leaf is read.

#### Scenario: A ticket picked for research

- **GIVEN** the user picks a ticket for research without asking for a change
- **WHEN** the agent's proposal is accepted with a reply such as "1"
- **THEN** the next message ends with the next-step block toward `flow-build` or `flow-plan`
- **AND** no build starts until the user's reply names it

#### Scenario: A close ends as its close mode sets

- **GIVEN** a deploy's cleanup close, which `close-modes` ends with the report and no footer
- **WHEN** `flow-close` reaches its last step
- **THEN** the message ends with the report and its waiting line, without the footer

### Requirement: The final report carries evidence and a try-it block

The final report SHALL carry a defect's regression test output seen failing and then passing, and, when the change can be seen or run, a block with the URL or command, the test account's index row, the steps and the expected result, in every delivery mode. Its parts SHALL be stated in one rule leaf, `close-report`, with one row per variant (close at the completion point, a pull request left open under `ready` or waiting for a reviewer, a deploy's cleanup close, `hold` and a pause outside it). Each work item SHALL get one final report, written by `flow-close` at its report step: `git-workflow` and `flow-deploy` SHALL hand what they observed to the close instead of writing it, `flow-build` SHALL hand it the line of code principles that changed a decision, which the report carries when it was handed over in that work item, and the end of unattended work SHALL summarize the session and point to each close's report. A later message with the result of a base CI run that ends after the close is not a second final report.

#### Scenario: Hold delivery

- **GIVEN** a change delivered under `hold`
- **WHEN** the agent reports it
- **THEN** the report holds the failing and passing output and the try-it block

#### Scenario: A pull request merged by the agent

- **GIVEN** a delivery whose pull request the agent merged
- **WHEN** `git-workflow` verifies the merge and runs `flow-close`
- **THEN** one final report is written, by `flow-close`, with the row for the close at the completion point


### Requirement: Playbooks and the close hygiene leaf name their own row

Each of the 13 playbooks of `hive-mode` SHALL carry one reporting line that links its row of `report-endings` and the row of `close-report` that the delivery mode in force picks, without naming a delivery mode; `pr-review-deploy` SHALL also link the deployment row. A playbook SHALL NOT copy text from those leaves. The `close-hygiene` leaf and its line in the global guidance SHALL cover steps 4, 5 and 7 of `flow-close` and SHALL NOT claim step 6, which belongs to `close-report`. Each trigger of the global index SHALL quote the first sentence of its leaf after «Read», and a content test SHALL check it.

#### Scenario: A playbook reports through its rows

- **GIVEN** the `bugfix` playbook
- **WHEN** the work closes
- **THEN** the playbook's reporting line leads to the close row of `report-endings` and to the row of `close-report` that the delivery mode in force picks

#### Scenario: Close hygiene stays out of the report step

- **GIVEN** the index line and the leaf of `close-hygiene`
- **WHEN** a close reaches step 6
- **THEN** neither the line nor the leaf asks to be read for the final report

### Requirement: The two reporting homes have bounded size and reading cost

`report-endings` SHALL have its own size limit of 9,000 bytes and `close-report` SHALL stay in the 7,000-byte exception; every other leaf keeps its limit and no text SHALL be cut to fit. The reading-footprint test for the plan and build route SHALL accept the growth that comes from reading those two leaves whole at each stage's end, as a second allowance of 9,230 bytes beside the earlier 182 bytes. The global guidance SHALL stay at or under 30,000 bytes.

#### Scenario: The leaves fit their limits

- **GIVEN** the candidate
- **WHEN** the content tests run
- **THEN** `report-endings` is at most 9,000 bytes, `close-report` at most 7,000 bytes, and the global guidance at most 30,000 bytes

## ADDED Requirements

### Requirement: A flow reads its stage packet and records it in the checkpoint
A flow that has a stage packet SHALL tell the agent, at its entry, to read that packet whole and not to read again the leaves it holds, to read the linked leaves when the packet is missing, and SHALL keep its link to each leaf for the moments outside the packet. The work's checkpoint SHALL carry a line `- Rules: <packet paths relative to the skills store, joined by ', '> · extra: <conditional leaves read>`, such as `- Rules: flow-build/stage.md · extra: defect-fix` or `- Rules: flow-verify/stage-entry.md, flow-verify/stage-exit.md`, joining each packet and leaf of the same flow as the flow reads it and starting again for another flow, written by `tasks.py <change-folder> checkpoint --rules` for a change folder, kept by later checkpoint calls that omit `--rules`, accepted by `check_change.py` when well formed and rejected otherwise, and written by hand in the checkpoint file of a plan kept in the conversation. The checkpoint SHALL also carry a line `- Principles: <value>`, where the value is `none` or one or more principle names matching `[a-z0-9-]+` separated by `, `, so the line `flow-build` hands to `flow-close` survives a compaction or a new session. `tasks.py <change-folder> checkpoint --principles` SHALL add its names to those the line holds, without duplicates, SHALL write `none` only when the line holds no name, SHALL replace `none` with the first name added after it, and SHALL never remove the line; a later call without `--principles` SHALL keep it, and `check_change.py` SHALL reject an empty or malformed value, including `none` together with a name. `flow-build` SHALL record a principle when it changes a decision and SHALL record `none` at its last step when the line is absent; `flow-close` SHALL read the line before its cleanup removes the checkpoint, and the **Principles** part of the final report SHALL come from it. After a compaction the agent SHALL read, in this order, the checkpoint, the local instruction files, the packets its `Rules:` line names, or the leaves its flow links when a packet is missing, the extra leaves, and then persistent memory, before its next step; the global core SHALL state this order in one line, and no other file under `content/` SHALL state a different one.

#### Scenario: The checkpoint keeps the rules line
- **WHEN** `tasks.py <folder> checkpoint --rules "flow-build/stage.md · extra: defect-fix"` runs and a later checkpoint call omits `--rules`
- **THEN** `## Checkpoint` still holds the same `- Rules:` line and `check_change.py` exits 0.

#### Scenario: The checkpoint keeps the principles line
- **WHEN** `tasks.py <folder> checkpoint --principles "boundary-discipline"` runs and a later checkpoint call omits `--principles`
- **THEN** `## Checkpoint` still holds `- Principles: boundary-discipline` and `check_change.py` exits 0.

#### Scenario: Principles accumulate
- **WHEN** `--principles "boundary-discipline"` runs, then `--principles "model-the-domain, boundary-discipline"`, then `--principles none`
- **THEN** `## Checkpoint` holds `- Principles: boundary-discipline, model-the-domain`.

#### Scenario: A name replaces none
- **WHEN** `--principles none` runs, then `--principles "boundary-discipline"`
- **THEN** `## Checkpoint` holds `- Principles: boundary-discipline`.

#### Scenario: A malformed principles line
- **WHEN** `## Checkpoint` holds `- Principles:` with an empty value
- **THEN** `check_change.py` prints the line and exits non-zero.

#### Scenario: A malformed rules line
- **WHEN** `## Checkpoint` holds `- Rules:` with no packet path
- **THEN** `check_change.py` prints the line and exits non-zero.

#### Scenario: Resuming after a compaction
- **WHEN** a build resumes after a compaction with a checkpoint whose `Rules:` line names `flow-build/stage.md`
- **THEN** the agent reads the checkpoint, the local instruction files, that packet and its extra leaves before memory, and does not read again each leaf the packet holds.

### Requirement: Each flow names its reply
`hive-mode`, `flow-plan`, `flow-build`, `flow-verify`, `flow-close`, `flow-deploy` and `git-workflow` SHALL each carry one line starting with `Reply:` that names the row of the report rule leaf that ends the flow and only what the flow adds that the row does not hold, such as the change folder's path and the build route of a plan, without restating that row. A content test SHALL fail when one of these flows lacks the line.

#### Scenario: A flow without its reply line
- **WHEN** the `Reply:` line is removed from `flow-build/SKILL.md`
- **THEN** `go test ./tests/content/...` fails and names `flow-build`.
