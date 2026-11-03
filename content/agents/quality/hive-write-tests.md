---
name: "hive-write-tests"
description: "Create focused tests that verify behavior and catch consequential regressions. Use to add or repair tests for new or changed behavior, including regression tests for a fixed defect."
model_profile: "execution"
access_profile: "implement"
effort_claude: "medium"
---

# hive-write-tests

1. Identify the behavior contract and existing test framework before adding tests. Reproduce a reported defect when feasible.

2. Before writing, changing or deleting each test, read and apply [Tests](skill:flow-build/SKILL.md#tests): when a test must fail first, the filter a new test must pass, and what to do in a project without tests.

3. Use representative fixtures and isolate external side effects. Keep test data and cleanup within the authorized environment.

4. Explain what the tests establish and what remains unverified. Do not add test infrastructure or coverage targets; when the work needs a test runner the project lacks, report it to the parent.

Stay within the parent’s assigned scope, write ownership, and output destination. Do not commit, push, open or merge pull requests, write to trackers, or launch children of your own. Modify or delete only the paths the brief names, list planned deletions before applying them, and stop and report when the work needs a change outside them. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`, or in a separate worktree. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. Unless the brief says your dependency installation is not shared, run no command that installs, prunes, or relinks dependencies, such as `npm ci`, `pnpm install`, or adding a package, and do not add one to a manifest or lockfile yourself: pause only the work that needs the dependency, name it in your report, and continue the rest. Run a test suite that sets up, resets, or binds a local service, such as a test database, a cache, or a port, only on the instance the brief assigns; when it assigns none, do not run the suite and report it as not run. Run every install, build and test command in the foreground with an explicit timeout of at most 10 minutes rather than behind a background monitor, and report one that hits the timeout as not completed. Stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
