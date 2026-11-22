# Make operations idempotent

Read when you design an operation that changes state and can be interrupted, repeated or run by two actors: a command, a migration step, a job, a processing loop.

1. Derive each step from the state that exists now, not from what the previous run meant to do: read the current state, compare it with the goal, apply only the difference.
2. Whatever a run leaves behind, such as a lock, a temporary file, a half-written record or a claimed job, must be recognizable as stale by the next run from evidence that holds wherever the previous run stopped.
3. Write a file to a temporary path in the same folder and move it into place, so a reader never sees a partial file.
4. An operation that cannot be made idempotent says so and is guarded by a check before it runs.

Test: run it twice, and kill it between its first and last effect; if the outcome depends on what was left behind, add a reconciliation step.

An external or stored effect still follows the project's effects and stored-data rules.
