# Checks

Read before running or repeating a check, after a build's last task, and before the push.

## Check reuse

The full local suite must have passing evidence for the candidate. Reuse a passing check from this work only when its command and output are available and its covered code, dependencies, tests, configuration and relevant environment are unchanged since that run, comparing the candidate, uncommitted changes included, with the checked state; a stage change alone does not invalidate a result. Run checks with missing evidence, changed inputs or inputs whose equivalence you cannot establish: only the affected checks when their scope is known, otherwise the full check. Record reused and rerun checks in the existing checkpoint. Reuse never replaces the real-path acceptance walk or the reruns a required independent verifier performs.

Read each check's result from the exit status of the check command itself. When you filter its output, keep its own status: bash's `${PIPESTATUS[0]}` for a pipe, or [run_check](checks/run-check.md), which saves the whole output; never read a file in the command that writes it. Read an unclear result from the saved log, and run the check again only when neither the status nor the output was kept.

A round of fixes is every fix made for the failures, corrections and findings known when the round starts, such as the round of [review findings](../review/review-findings.md). In any flow, after each fix of a round, run the tests of the edited files' packages or folders and each failing test or check they do not include, such as a lint or a build. Once the round's last fix is in, run the checks the first paragraph requires, the full suite when the fixes' scope is not known; never the full suite after each fix.

## Long checks

A check that spends money or third-party quota, changes shared or remote resources, or runs over 30 minutes needs the user's confirmation first. Estimate a check's duration from evidence, such as a previous run, the same job's time in CI or a timed partial run, never from an unmeasured guess. With no evidence, a check that needs no other confirmation starts in the background, and you ask whether to continue once it has run 30 minutes (a child reports it to its parent instead, and unattended work follows its delegation's ceiling).

## Local CI

Read the repository's CI workflow files and run locally every job that needs no external service, secret or deployment, for the packages the change touches and those that depend on them; report the jobs that need one by name as CI-only. A build does this when every task is done, recording in the checkpoint the commands, relevant output, checked commit, uncommitted changes and the environment the checks depend on. Before the push, `flow-verify` ensures passing evidence covers every such job, running the missing or invalidated ones as Check reuse sets.
