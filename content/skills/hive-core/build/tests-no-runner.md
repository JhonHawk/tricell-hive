# Project or area without tests

Read in a repository without a test runner, before a defect fix or consequential logic. The point is the first edit of either. Also read when a package the change touches has no tests.

When the repository has a test runner but the package has no tests, [Tests](tests.md) applies and the first test of a contract becomes its owner. When the repository has no test runner:

- For a defect or consequential logic, before editing, find every caller of the behavior by searching its literal values, such as names, routes, roles and hosts, and pin its current behavior for those callers with a check through the real path, adding no dependency. Keep that check where the project keeps checks, or as evidence of the change when it has none.
- Adding test infrastructure is the user's decision: propose it in the plan, or before the first edit when the work has no plan, the first time a defect fix or consequential logic reaches the project, with the package check of the global guidance, and save the answer in persistent memory as a project decision. Never install it without that answer, which [unattended delegation](../../unattended-delegation/SKILL.md) step 5 gives while it is active; a child that needs a runner the project lacks reports it to its parent.
