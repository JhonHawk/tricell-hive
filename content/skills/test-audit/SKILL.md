---
name: test-audit
description: Audit one area's tests and remove duplicate, implementation-coupled or vacuous proof without leaving any contract without its only proof, through a baseline, a read-only ledger per test, a layer plan, keepers before deletions, an independent preservation review and targeted mutations. Use when the user asks to audit, prune, slim down or deduplicate the tests of an area, module or subsystem; not for writing the tests of a change, which follows flow-build's Tests section.
---

# Test audit

> Adapted from the `test-audit` skill of [openclaw/openclaw](https://github.com/openclaw/openclaw/tree/main/.agents/skills/test-audit) (MIT, commit `3195fba90643`). The complete notice is in [OpenClaw license](references/license-openclaw.md).

Prune an area's tests to one owner per contract, proven by mutation, never by a deletion count. The [Tests](../flow-build/SKILL.md#tests) section of flow-build is the criterion for every mark: read it before step 4. The audit runs through the normal flows: steps 1–4 are research and change no repository file; step 5 is the plan, written as [flow-plan](../flow-plan/SKILL.md) sets; steps 6–9 run inside [flow-build](../flow-build/SKILL.md), after the user answers the delivery and code-review questions for this audit; then [flow-verify](../flow-verify/SKILL.md) runs the full suite, the chosen review and the delivery, and flow-close reports as step 10 sets.

## Scope

1. **Name one area.** The area is the code the user names, such as a module, a service or a package, with every test that exercises it at any layer. When the user names none, propose up to three areas ranked by measurable signals and ask which one: test lines per production line, the share of assertions on calls (such as `toHaveBeenCalled`, `verify(` or `assert_called`), and test files named after tickets. One area is one work item and one pull request.
2. **Record a baseline** at a pinned commit:
   - the area's test lines and production lines, counted separately;
   - each test file's pass or fail state; put each failing test in a baseline-failure list, as a possible product defect, never as a deletion candidate;
   - line and branch coverage of the area only when the repository already measures it, adding no dependency, with unit and integration tests together when both exist;
   - the wall time of each suite that runs the area's tests.

   When the repository has no test runner, say so and stop: there is nothing to audit, and the Tests section sets what applies there.
3. **Split into lanes** along the production code that owns the behavior, not along file names. Each test file belongs to exactly one lane. Split a lane with more than 40 test declarations into parts of at most 40.

## Ledger

4. **Mark every test.** Each lane or part with more than 10 test declarations goes to one `hive-research` child; keep the others in the main thread. The child reads every test in full, including parameter tables and shared fixtures, and its production owner, callers, history and CI routing, and returns one row per test declaration as `mark · file:line · contract · evidence`; a parameterized test is one row unless its cases need different marks. Brief each child with the lane, this section and the Tests section; the child edits nothing and runs no test suite that shares a local database or service with other work. The marks:
   - `R` retain: name the contract and the realistic change that makes the test fail;
   - `F` repair: the contract stays but the assertion needs a fix, such as a negative case that passes on an unrelated guard; say what to repair;
   - `C` consolidate: name the keeper that absorbs it and cite the keeper's `file:line` that already asserts it, or the exact assertion to carry into it;
   - `D` delete: cite the remaining proof as `file:line`, or say why no contract exists.

   Judge a test by its assertions, not its name. A row without that evidence is `R`, and the report names it.

   **Retention bar.** Whatever the other marks say, keep at least one owner test for every permission, security rule, public or persisted contract, migration, storage behavior, configuration default, concurrency rule and observable ordering. A slow or static test is not a reason to delete it, and a baseline failure is never deleted. The child also lists test-only production seams: exports, flags, parameters or wrappers with no caller in production code and not part of a published interface, such as a package's public exports or a public endpoint. It lists vacuous or misnamed tests, and contracts whose only proof today is a test marked `C` or `D`.

## Plan

5. **Plan the layers** in the main thread, from the ledger: for each contract its keeper test, the assertions to carry, the `F` repairs, the files retired and the seams removed. Correct the ledger errors this pass finds. Write the plan in a change folder as flow-plan sets, because its tasks are delegated: one task per lane, with the ledger as `ledger.md` beside `tasks.md`, and the baseline-failure list in `proposal.md`. Then ask the delivery and code-review questions, and each baseline failure as a decision: investigate it as a possible product defect in this work, open a ticket, or leave it listed.

## Cutover

Steps 6–9 run inside flow-build.

6. **Keepers before deletions.** In each lane, add or extend the keeper tests with their carried assertions and apply the `F` repairs, run them and see them pass, and only then delete the tests they absorb, so no contract is ever left without proof. Remove the test-only seams the lane unlocks in the same lane. Lanes that share one checkout run one after another; run lanes in parallel only in separate worktrees, each with its own test database or service, as [git-workflow](../git-workflow/SKILL.md) sets for worktrees. Delegate a lane to `hive-write-tests` as flow-build's [Delegation](../flow-build/SKILL.md#delegation) section sets; its brief names the change folder, the lane's task, and the local test instance the child may run. Never edit files while a test run reads the same checkout. Commit each lane when the delivery makes commits.
7. **Review preservation.** One read-only `hive-review-code` child per group of lanes that share a boundary, such as HTTP or the database, compares every deleted or trimmed test, as it was at the baseline commit, with the keeper the ledger names, reading the keeper's current code. It reports each contract that lost its only proof, with the deleted test, the production line it guarded and a change to that line that every remaining test would now miss, and each new, kept or repaired assertion that cannot fail. Restore each confirmed gap in its keeper.
8. **Prove with mutations.** For each contract that lost one of its tests, each `F` repair and each restored gap: copy the production file aside, change the one line the contract depends on, run the keeper and see it fail, then copy the saved file back and confirm with `cmp <saved copy> <file>` that it is identical. Record each mutation and the test that caught it in `ledger.md`. A keeper that stays green is fixed and mutated again. Never commit a mutated file.
9. **Route product defects.** A defect a keeper or a repaired test exposes, and each baseline failure the user chose to investigate in step 5, goes to the user as a decision: fix it in this work, as its own commit, or open a ticket. A fix carries a control: its test fails without the fix and passes with it on the same harness. Record any other product discrepancy as a finding, not a fix.

Then hand the candidate to flow-verify.

## Report

10. Report against the baseline, in flow-close's final report:
    - test lines and production lines, before and after, separately, and the number of tests;
    - coverage before and after, when the repository measures it;
    - each suite's wall time before and after: moving proof to a stronger boundary can make a suite slower;
    - the lanes, the keeper of each contract, the repaired tests, the retired files and the removed seams;
    - the preservation gaps found and the mutation that proves each contract that lost or repaired a test;
    - every baseline failure with what was decided for it, the product defects with their control, and the follow-ups.

    Name each test kept for lack of evidence. Never set or report a deletion target.
