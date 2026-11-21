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

## ADDED Requirements

### Requirement: A flow reads its stage packet and records it in the checkpoint
A flow that has a stage packet SHALL tell the agent, at its entry, to read that packet whole and not to read again the leaves it holds, to read the linked leaves when the packet is missing, and SHALL keep its link to each leaf for the moments outside the packet. The work's checkpoint SHALL carry a line `- Rules: <packet path relative to the skills store> · extra: <conditional leaves read>`, such as `- Rules: flow-build/stage.md · extra: defect-fix`, written by `tasks.py <change-folder> checkpoint --rules` for a change folder, kept by later checkpoint calls that omit `--rules`, accepted by `check_change.py` when well formed and rejected otherwise, and written by hand in the checkpoint file of a plan kept in the conversation. After a compaction the agent SHALL read, in this order, the checkpoint, the local instruction files, the packet its `Rules:` line names, or the leaves its flow links when that packet is missing, the extra leaves, and then persistent memory, before its next step; the global core SHALL state this order in one line, and no other file under `content/` SHALL state a different one.

#### Scenario: The checkpoint keeps the rules line
- **WHEN** `tasks.py <folder> checkpoint --rules "flow-build/stage.md · extra: defect-fix"` runs and a later checkpoint call omits `--rules`
- **THEN** `## Checkpoint` still holds the same `- Rules:` line and `check_change.py` exits 0.

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
