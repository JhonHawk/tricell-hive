## ADDED Requirements

### Requirement: Checks with saved output run through a capped helper

A check whose output the agent saves to a file SHALL run through `hive-rule-checks/scripts/run_check.py`, which SHALL take the command as arguments without a shell, write its whole output to a log file under a size cap, stop the command's process group when the cap is passed, close the log before printing its last lines or the lines that match a pattern, and exit with the command's own status, or with 125 when the cap stopped it. The agent SHALL NOT write a shell wrapper that redirects a block's output to a file and reads that file in the same command. The session auditor SHALL fail a shell command in which a compound block or `exec` redirects output to a file that a trap or a command inside that block reads, including a Codex command whose text is interpolated, and SHALL NOT fail a simple command's redirection.

#### Scenario: A failing check under the helper

- **GIVEN** a test run that fails after printing a thousand lines
- **WHEN** the agent runs it through the helper
- **THEN** the helper prints the log's path, then the last 40 lines after the run ends, and exits with the test run's status
- **AND** the log holds the thousand lines and no copy of its own end

#### Scenario: A runaway output

- **GIVEN** a command that writes without end
- **WHEN** it runs through the helper with a 1 MiB cap
- **THEN** the helper stops its process group, the log stays near 1 MiB, and the helper exits with 125 naming the log and the cap
