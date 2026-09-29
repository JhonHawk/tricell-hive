---
name: "hive-write-tests"
description: "Create focused tests that verify behavior and catch consequential regressions. Use to add or repair tests for new or changed behavior, including regression tests for a fixed defect."
model_profile: "execution"
access_profile: "implement"
---

# hive-write-tests

1. Identify the behavior contract and existing test framework before adding tests. Reproduce a reported defect when feasible.

2. Test meaningful boundaries, negative cases, asynchronous ordering, and transactions where they affect the change. Before writing each test, name the realistic change to the code under test that would make it fail; do not write a test that no such change can fail or that only restates the current structure.

3. Use representative fixtures and isolate external side effects. Keep test data and cleanup within the authorized environment.

4. Explain what the tests establish and what remains unverified. Do not introduce infrastructure or arbitrary coverage targets without a concrete need.

Stay within the parent’s assigned scope, write ownership, and output destination. Commit only in your assigned working tree and branch, and only when the brief allows it; never push, open or merge pull requests, or write to trackers. Modify or delete only the paths the brief names, list planned deletions before applying them, and stop and report when the work needs a change outside them. Run installs, builds, and other long commands in the foreground with a bounded timeout rather than behind a background monitor, and stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent.
