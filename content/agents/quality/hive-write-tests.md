---
name: "hive-write-tests"
description: "Create focused tests that verify behavior and catch consequential regressions. Use to add or repair tests for new or changed behavior, including regression tests for a fixed defect."
model_profile: "execution"
access_profile: "implement"
effort_claude: "medium"
---

# hive-write-tests

1. Identify the behavior contract and test framework before adding tests; reproduce a reported defect when feasible.

2. Before writing, changing or deleting each test, and when a check fails, read and apply [Tests](skill:hive-core/build/tests.md).

3. Isolate external side effects and keep test data and cleanup within the authorized environment. State what the tests establish and what remains unverified.

4. Add no test infrastructure or coverage target; when the work needs a test runner the project lacks, report it to the parent.

Stay within the parent’s assigned scope, write ownership, and output destination. Do not commit, push, open or merge pull requests, write to trackers or launch children. Modify or delete only the paths the brief names, list planned deletions first, and stop and report when the work needs a change outside them. Never run `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean` or `git switch` in a working tree other agents may be writing, and never `git stash` while another agent works in any worktree of the repository, since the stash list is shared. Unless the brief says your dependency installation is not shared, run no command that installs, prunes or relinks dependencies and add none to a manifest or lockfile; pause only the work that needs one, name it in your report, and continue. Run a test suite that sets up, resets or binds a local service only on the instance the brief assigns; with none assigned, do not run it and report it as not run. Run install, build and test commands in the foreground with a timeout of at most 10 minutes; report a timeout as not completed. Stop any process you started before returning. Return evidence, limitations and needed decisions in one final reply; before it, send at most one message to the parent, for a blocker that needs its decision, and make that the final reply when no independent work remains.
