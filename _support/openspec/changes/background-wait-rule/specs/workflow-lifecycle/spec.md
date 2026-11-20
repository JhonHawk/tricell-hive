## ADDED Requirements

### Requirement: A background wait ends on every outcome and has a deadline

Before starting a polling loop or a background command that waits on a process or an external job, the agent SHALL read the rule leaf `hive-rule-background-wait`. CI waits SHALL keep following the merge gate, and child waits the child brief. The agent SHALL prefer a command that blocks until the job reaches a final state and exits with its status. A hand-built wait SHALL end on success, on failure, on a target that does not exist yet, on a process that hangs and on output it does not recognize. Unrecognized output SHALL end the wait with a non-zero status and print what was seen. The wait SHALL compare a structured field or an exit status rather than free text, or SHALL first check a text pattern against the query's actual output. A watch that exits reporting no target SHALL be rerun with the same command. Every such wait SHALL have a deadline that bounds the whole loop. When the host wakes the agent only when the command exits, each wait stretch SHALL last 5 to 10 minutes. On waking, the agent SHALL check progress. While the job advances, the agent SHALL rerun the same command. After two stretches in a row without progress, the agent SHALL stop and report a blocker with the last state seen.

#### Scenario: A deploy status query prints an extra line

- **GIVEN** a cloud build whose status query with `--output text` prints `SUCCEED` followed by a `None` pagination line
- **WHEN** the agent waits in the background for the build to finish
- **THEN** the wait ends within its deadline, on the final state or on the unrecognized output, and the agent reports what it saw

### Requirement: The session audit flags a background wait without a deadline

`go run ./tests/pilot --audit-session` SHALL report `unbounded_background_wait` for Claude Code and Grok sessions. A background shell command is one launched with `run_in_background` on Claude Code or with `block_until_ms` set to 0 on Grok. The criterion SHALL fail when such a command holds a `while`, `until` or `for` loop with `sleep` and has no deadline. A deadline SHALL be one of these:
- `timeout` or `gtimeout` as the first command of the segment that holds the loop; a `timeout` inside the loop body SHALL NOT count;
- a `for` loop over a finite list, or a C-style `for` whose condition compares a counter or the clock to a limit;
- `$SECONDS`, `date +%s` or a counter compared to a limit in the loop's own condition, or in a test inside the loop that leads to `exit` or `break`.

A loop SHALL be recognized wherever it starts a command, including after `then`, `do`, `else`, `elif`, `!`, `(` or `{`. The criterion SHALL report pass when every such loop has a deadline, and not_observed when there is none. It SHALL report not_observed for Codex sessions.

#### Scenario: The Amplify wait in a Grok session

- **GIVEN** a Grok session that launched with `block_until_ms: 0` a `while :` loop polling `aws amplify list-jobs` every 20 seconds, with no timeout
- **WHEN** the session audit reads it
- **THEN** `unbounded_background_wait` fails with that call's line as evidence

### Requirement: The session audit fails a merged branch left only after a close

In `go run ./tests/pilot --audit-session`, `merged_branch_deleted` SHALL fail only when an assistant close report follows the session's last merge and no branch deletion follows that merge. When a merge has no later deletion and no close report follows it, the criterion SHALL report review and SHALL name that the session may still be open or the merge may be a promotion. The pilot criterion of the same name SHALL keep its behavior.

#### Scenario: A session still promoting

- **GIVEN** a real session whose last merge is the promotion pull request from `development` to `qa`, with no close report after it
- **WHEN** the session audit reads it
- **THEN** `merged_branch_deleted` reports review, not fail
