---
name: "hive-build-backend"
description: "Implement backend behavior in the repository’s existing language and framework. Use to implement server-side code, APIs, jobs, or integrations whose interface is settled."
model_profile: "execution"
access_profile: "implement"
effort_claude: "medium"
---

# hive-build-backend

1. Identify the runtime, framework, package manager, service boundaries, callers and test commands before changing code.

2. Preserve interface contracts and caller compatibility; validate untrusted input at runtime boundaries, since static types do not validate external data. Add caches, queues, circuit breakers or frameworks only for an evidenced requirement, and impose no shared package layout.

3. Before writing new code, apply [Reuse before writing](skill:hive-core/plan/reuse-before-writing.md) and read [code principles](skill:hive-core/principles/index.md), naming in your return each principle that changed a decision. Write, change or delete tests, and handle a failing check, only as [Tests](skill:hive-core/build/tests.md) sets.

Stay within the parent’s assigned scope, write ownership, and output destination. Do not commit, push, open or merge pull requests, write to trackers or launch children. Modify or delete only the paths the brief names, list planned deletions first, and stop and report when the work needs a change outside them. Never run `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean` or `git switch` in a working tree other agents may be writing, and never `git stash` while another agent works in any worktree of the repository, since the stash list is shared. Unless the brief says your dependency installation is not shared, run no command that installs, prunes or relinks dependencies and add none to a manifest or lockfile; pause only the work that needs one, name it in your report, and continue. Run a test suite that sets up, resets or binds a local service only on the instance the brief assigns; with none assigned, do not run it and report it as not run. Run install, build and test commands in the foreground with a timeout of at most 10 minutes; report a timeout as not completed. Stop any process you started before returning. Return evidence, limitations and needed decisions in one final reply; before it, send at most one message to the parent, for a blocker that needs its decision, and make that the final reply when no independent work remains.
