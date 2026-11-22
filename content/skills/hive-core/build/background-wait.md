# Background waits

Read before a polling loop or a background wait on a process or job. A background wait is a command that waits on a process or an external job, such as a deploy, a cloud build or a long local run. A CI wait follows [merge gate](../git/merge-gate.md); a wait on a child follows [child brief](../delegation/child-brief.md).

1. **Prefer a blocking command** that waits for the job's final state and exits with its status, such as a watch mode with an exit-status flag, before writing a loop.
2. **End on every outcome.** A loop you write ends on success, failure, a hang and output it does not recognize (non-zero exit, printing what it saw). Compare a structured field or an exit status, not free text; when you must match text, check the pattern against the query's real output first. A job that does not exist yet has not finished: keep waiting.
3. **Give each run of the wait a deadline.** Wrap the loop in `timeout <seconds>`, or count iterations or elapsed seconds and exit with the last state seen at the limit. A `timeout` on one command inside the loop bounds one iteration, not the wait.
4. **Wake in stretches.** When the host wakes you only when the command exits, that deadline is what wakes you: set it to 5 to 10 minutes, never a 30- or 60-second slice. On each wake, read the job's state; while it advances, rerun the same command for another stretch. After two stretches in a row with no progress, stop and report a blocker with the last state seen.
