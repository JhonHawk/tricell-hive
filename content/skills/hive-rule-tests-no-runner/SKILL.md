---
name: hive-rule-tests-no-runner
description: "Hive rule leaf: read in a repository without a test runner, or a package without tests, before editing a defect fix or consequential logic. Direct checks, caller pinning, test infrastructure."
disable-model-invocation: true
---

# Project or area without tests

Read in a repository without a test runner, before the first edit of a defect fix or consequential logic, and when a package the change touches has no tests.

When the repository has a test runner but the package the change touches has no tests, the rules of [Tests](../hive-rule-tests/SKILL.md) apply, and the first test of a contract becomes its owner. When the repository has no test runner:

- Verify a change without consequences with a direct check, such as the build, the type checker, or running the changed path with a request, a script or the browser, and report "no tests; verified with <check>".
- For a defect or consequential logic, before editing, find every caller of the behavior by searching its literal values, such as names, routes, roles and hosts, and pin its current behavior for those callers with a check through the real path, adding no dependency. Keep that check where the project already keeps checks, or as evidence of the change when it has none.
- Adding test infrastructure is the user's decision: the main thread proposes it in the plan, or before the first edit when the work has no plan, the first time a defect fix or consequential logic reaches the project, with the package check of the global guidance, and saves the answer in persistent memory as a project decision so it is not asked again. Never install it without that answer, which [unattended delegation](../unattended-delegation/SKILL.md) step 5 gives while it is active; a child that needs a runner the project lacks reports it to its parent instead.
