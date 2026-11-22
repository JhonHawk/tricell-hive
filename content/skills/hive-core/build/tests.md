# Tests

Read before writing, changing or deleting a test, or when a check fails. Also read when choosing a task's test approach in a plan. In a repository without a test runner, before the first edit of a defect fix or consequential logic, also read [no test runner](tests-no-runner.md).

**When a check fails**, first decide whether the code, the criterion or the test is wrong, and fix the code, as [defect fix](defect-fix.md) sets, or the test: change, skip or delete an existing test only when the test itself is wrong, and say so in the task's line or your report. A wrong criterion is the user's decision: report it, or a child reports it to its parent, instead of changing it. Once the user decides, revise the plan as [Revise a plan in progress](../plan/change-folder.md#revise-a-plan-in-progress) sets.

**When a test must fail first (`tdd`).** Write the test, run it and see it fail for the intended reason, then change the code until it passes, one behavior at a time. This applies only to:

- A defect you fix: one regression test, in the test that owns the broken behavior as point 2 of the filter below sets, seen failing on the code before the fix. Do not repeat it at every layer the defect crosses. When the failure cannot be reproduced before the fix, say so.
- Consequential logic: permissions, money, data integrity, or a public or persisted contract.

Verify any other behavior change with a direct check, or by extending its owning test (`check`). A refactor that keeps behavior uses the existing tests, or first a test that pins the current behavior (`characterization`). A mechanical change adds no test.

**Before writing a test**, apply this filter; do not write a test that fails any point:

1. Name the realistic change to the code that would make it fail. If none exists, do not write it.
2. Find the test that already owns that behavior. Each contract has one owner, at the strongest boundary the repository's harness runs cheaply, such as an integration test through HTTP and the database rather than a unit test with mocked collaborators. Extend the owner, its table or its fixtures, instead of adding a sibling test or the same assertion at another layer. Keep a lower-layer test only for a branch the owner cannot reach cheaply.
3. Replace with a test double only what crosses the system boundary: an external service, time, randomness, or, in a client's tests, the project's own server API through stub data typed by its contract. Do not mock the project's own modules in the same process, nor its own database when the repository has an integration harness: the test then stops covering them.
4. Keep a test that fails when the behavior its users observe breaks and still passes when only the internals change. Reject one that asserts how the code is written instead of what it does, such as which internal functions were called or in what order; that uses a double to supply the behavior it claims to check; that reads source code as text to check how it is written, unless the file's text is itself what the project delivers to its readers, such as guidance or documentation, never code or build configuration; that copies a list or inventory from the code; that computes its expected value with the code under test; that asserts something always true; or that needs an export, flag or wrapper only tests use.
5. Wait on an observable condition, never a fixed sleep, and isolate state shared between tests.

When the change handles a failure, a boundary value, a concurrent actor or a transaction, its verification covers that case too: the negative case, the limit, the ordering, or the rollback, in the owner test when there is one, otherwise in the direct check.
