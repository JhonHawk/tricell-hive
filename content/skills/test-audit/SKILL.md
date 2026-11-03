---
name: test-audit
description: Audit one area's tests and remove duplicate, implementation-coupled or vacuous proof without leaving any contract without its only proof, through a baseline, a read-only ledger per test, a layer plan, keepers before deletions, an independent preservation review and targeted mutations. Use when the user asks to audit, prune, slim down or deduplicate the tests of an area, module or subsystem; not for writing the tests of a change, which follows flow-build's Tests section.
---

# Test audit

> Adapted from the `test-audit` skill of [openclaw/openclaw](https://github.com/openclaw/openclaw/tree/main/.agents/skills/test-audit) (MIT, commit `3195fba90643`). The complete notice is in [OpenClaw license](references/license-openclaw.md).

Prune an area's tests to one owner per contract, proven by mutation, never by a deletion count. The [Tests](../flow-build/SKILL.md#tests) section of flow-build is the criterion for every mark: read it before step 4. Steps 1–5 only read and change no repository file. Step 6 edits the repository and starts only after the user answers, for this audit, the delivery and code-review questions of the global guidance's «Delivery».

## Scope

1. **Name one area.** The area is the code the user names, such as a module, a service or a package, with every test that exercises it at any layer. When the user names none, propose up to three areas ranked by measurable signals and ask which one: test lines per production line, the share of assertions on calls (such as `toHaveBeenCalled`, `verify(` or `assert_called`), and test files named after tickets. One area is one work item and one pull request.
2. **Record a baseline** at a pinned commit:
   - the area's test lines and production lines, counted separately;
   - each test file's pass or fail state; put a failing test in its own list as a possible product defect, never as a deletion candidate;
   - line and branch coverage of the area only when the repository already measures it, adding no dependency, with unit and integration tests together when both exist;
   - the wall time of each suite that runs the area's tests.

   When the repository has no test runner, say so and stop: there is nothing to audit, and the Tests section sets what applies there.
3. **Split into lanes** along the production code that owns the behavior, not along file names. Each test file belongs to exactly one lane.

## Ledger

4. **Mark every test.** For each lane, a `hive-research` child reads every test in full, including parameter tables and shared fixtures, and its production owner, callers, history and CI routing, and returns one row per test declaration; a parameterized test is one row unless its cases need different marks. Delegate as [flow-research](../flow-research/SKILL.md) step 2 sets, and keep a small lane in the main thread. Brief each child with the lane, this section and the Tests section; the child edits nothing and runs no test suite that shares a local database or service with other work. The marks:
   - `R` retain: name the contract and the realistic change that makes the test fail;
   - `F` repair: the contract stays but the assertion needs a fix, such as a negative case that passes on an unrelated guard; say what to repair;
   - `C` consolidate: name the keeper that absorbs it and cite the keeper's `file:line` that already asserts it, or the exact assertion to carry into it;
   - `D` delete: cite the remaining proof as `file:line`, or say why no contract exists.

   Judge a test by its assertions, not its name. A row without that evidence is `R`.

   **Retention bar.** Whatever the other marks say, keep at least one owner test for every permission, security rule, public or persisted contract, migration, storage behavior, configuration default, concurrency rule and observable ordering. A slow or static test is not a reason to delete it, and a baseline failure is never deleted. The child also lists test-only production seams (exports, flags, parameters or wrappers with no production caller), vacuous or misnamed tests, and contracts whose only proof today is a test marked `C` or `D`.
5. **Plan the layers** in the main thread, from the ledger: for each contract its keeper test, the assertions to carry, the files retired and the seams the lane removes. Correct the ledger errors this pass finds. Then report the ledger summary and this layer plan to the user, and ask the delivery and code-review questions unless the user already answered them for this audit. A plan kept in a change folder holds the ledger as `ledger.md`.

## Cutover

6. **Keepers before deletions.** In each lane, add or extend the keeper tests with their carried assertions, run them and see them pass, and only then delete the tests they absorb, so no contract is ever left without proof. Remove the test-only seams the lane unlocks in the same lane. Change a shared fixture in one lane at a time. Delegate a lane to `hive-write-tests` as flow-build's [Delegation](../flow-build/SKILL.md#delegation) section sets. Never edit files while a test run reads the same checkout.
7. **Review preservation.** One read-only `hive-review-code` child per group of lanes that share a boundary, such as HTTP or the database, compares every deleted or trimmed test, as it was at the baseline commit, with the keeper the ledger names, reading the keeper's current code. It reports each contract that lost its only proof, with the deleted test, the production line it guarded and a change to that line that every remaining test would now miss, and each new or kept assertion that cannot fail. Restore each confirmed gap in its keeper.
8. **Prove with mutations.** For each contract that lost one of its tests, and each restored gap: copy the production file aside, change the one line the contract depends on, run the keeper and see it fail, then restore the file and confirm `git diff --quiet -- <file>` exits 0. Record each mutation and the test that caught it. A keeper that stays green is fixed and mutated again. Never commit a mutated file.
9. **Route product defects.** A baseline failure that survives into a keeper, or a defect a keeper exposes, goes to the user as a decision: fix it in this work, as its own commit, or open a ticket. A fix carries a control: its test fails without the fix and passes with it on the same harness. Record any other product discrepancy as a finding, not a fix.

## Report

10. Report against the baseline, in flow-close's final report:
    - test lines and production lines, before and after, separately, and the number of tests;
    - coverage before and after, when the repository measures it;
    - each suite's wall time before and after: moving proof to a stronger boundary can make a suite slower;
    - the lanes, the keeper of each contract, the retired files and the removed seams;
    - the preservation gaps found and the mutation that proves each contract that lost a test;
    - the product defects with their control, and the follow-ups.

    Name each test kept for lack of evidence. Never set or report a deletion target.
