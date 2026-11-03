---
name: test-audit
description: Audit one area's tests and remove duplicate, implementation-coupled or vacuous proof without leaving any contract without its only proof, through a baseline, a read-only ledger per test, a layer plan, keepers before deletions, an independent preservation review and targeted mutations. Use when the user asks to audit, prune, slim down or deduplicate the tests of an area, module or subsystem; not for writing the tests of a change, which follows flow-build's Tests section.
---

# Test audit

> Adapted from the `test-audit` skill of [openclaw/openclaw](https://github.com/openclaw/openclaw/tree/main/.agents/skills/test-audit) (MIT, commit `3195fba90643`). The complete notice is in [OpenClaw license](references/license-openclaw.md).

Prune an area's tests to one owner per contract, proven by mutation, never by a deletion count. The [Tests](../flow-build/SKILL.md#tests) section of flow-build is the criterion for every mark: read it before step 4. The audit adds to the normal flows and replaces none of their steps: steps 1–4 are research and change no repository file; step 5 adds to [flow-plan](../flow-plan/SKILL.md); steps 6–9 add to [flow-build](../flow-build/SKILL.md); [flow-verify](../flow-verify/SKILL.md) and flow-close follow, with the exceptions steps 9 and 10 name. When you route from one flow to the next, name the audit in the command's argument, such as `/flow-plan test-audit <area>`, so this skill keeps applying.

## Scope

1. **Name one area.** The area is the code the user names, such as a module, a service or a package, with every test that exercises it at any layer. When the user names none, propose up to three areas ranked by measurable signals and ask which one: test lines per production line, the share of assertions on calls (such as `toHaveBeenCalled`, `verify(` or `assert_called`), and test files named after tickets. One area is one work item and one pull request.
2. **Record a baseline** at a pinned commit:
   - the area's test lines and production lines, counted separately;
   - each test file's pass or fail state; put each failing test in a baseline-failure list, as a possible product defect, never as a duplicate to delete;
   - line and branch coverage of the area only when the repository already measures it, adding no dependency, with unit and integration tests together when both exist;
   - the wall time of each suite that runs the area's tests.

   When the repository has no test runner, say so and stop: there is nothing to audit, and the Tests section sets what applies there.
3. **Split into lanes** along the production code that owns the behavior, not along file names. Each test file belongs to exactly one lane, and tests that may assert the same contract stay in the same lane.

## Ledger

4. **Mark every test.** When the whole area has at most 10 test declarations, mark them in the main thread. Otherwise give each lane to one `hive-research` child, briefed as [flow-research](../flow-research/SKILL.md) step 4 sets, with this section, the Tests section, and a ledger file in the host's temporary location as its output destination. The child reads every test in full, including parameter tables and shared fixtures, and its production owner, callers, history and CI routing; writes one row per test declaration to that file as `mark · file:line · contract · evidence`, a parameterized test being one row unless its cases need different marks; and returns a short reply with its counts per mark and the lists below. It edits nothing else and runs no test suite that shares a local database or service with other work. The marks:
   - `R` retain: name the contract and the realistic change that makes the test fail;
   - `F` repair: the contract stays but the assertion needs a fix, such as a negative case that passes on an unrelated guard; say what to repair;
   - `C` consolidate: name the keeper that absorbs it and cite the keeper's `file:line` that already asserts it, or the exact assertion to carry into it;
   - `D` delete: cite the remaining proof as `file:line`, or say why no contract exists.

   Judge a test by its assertions, not its name. A row without that evidence is `R`, and the report names it.

   **Retention bar.** Whatever the other marks say, keep at least one owner test for every permission, security rule, public or persisted contract, migration, storage behavior, configuration default, concurrency rule and observable ordering. A slow or static test is not a reason to delete it, and a baseline failure is never deleted as a duplicate. The child also lists test-only production seams: exports, flags, parameters or wrappers with no caller in production code and not part of a published interface, such as a package's public exports or a public endpoint. It lists vacuous or misnamed tests, and contracts whose only proof today is a test marked `C` or `D`.

## Plan

5. Plan as flow-plan sets, adding: from the ledger, each contract's keeper test, the assertions to carry, the `F` repairs, the files retired and the seams removed, correcting the ledger errors this pass finds; a change folder, since the tasks are delegated, with the ledger as `ledger.md`, one task per lane and the baseline-failure list in `proposal.md`; and one decision per baseline failure: investigate it as a possible product defect in this work, open a ticket, or leave it failing.

## Cutover

Steps 6–9 add to flow-build.

6. **Keepers before deletions.** Lanes run one after another in the work branch's checkout. In each lane, add or extend the keeper tests with their carried assertions and apply the `F` repairs, run them and see them pass, and only then delete the tests they absorb, so no contract is ever left without proof. Then remove the test-only seams the lane unlocks, in the main thread or through the build role for that code's surface. Change a shared fixture only in the lane that owns it, and run every lane that uses it after the change. Delegate a lane's tests to `hive-write-tests` as flow-build's [Delegation](../flow-build/SKILL.md#delegation) section sets; its brief names the change folder, the lane's task, the paths it may change (the lane's test files and fixtures) and the local test instance it may run. Never edit files while a test run reads the same checkout.
7. **Review preservation.** One read-only `hive-review-code` child per group of lanes that share a boundary, such as HTTP or the database, compares every deleted or trimmed test, as it was at the baseline commit, with the keeper the ledger names, reading the keeper's current code. It reports each contract that lost its only proof, with the deleted test, the production line it guarded and a change to that line that every remaining test would now miss, and each new, kept or repaired assertion that cannot fail. Restore each confirmed gap in its keeper.
8. **Prove with mutations.** For each contract that lost one of its tests, each `F` repair and each restored gap: copy the production file to the host's temporary location, change the one line the contract depends on, run the keeper, and see it fail at the assertion on that contract; a failure to build or import, or one at another assertion, proves nothing. Then copy the saved file back and confirm with `cmp <saved copy> <file>` that it is identical. Record each mutation and the assertion that caught it in `ledger.md`. A keeper that does not fail that way is fixed and mutated again. Never commit a mutated file.
9. **Route product defects.** A defect a keeper or a repaired test exposes, and each baseline failure the user chose to investigate, goes to the user as a decision: fix it in this work, as its own commit, or open a ticket. A fix carries a control: its test fails without the fix and passes with it on the same harness. When the investigation shows the test itself is wrong, the Tests section governs changing or deleting it. A baseline failure the user chose to leave failing is pre-existing: flow-verify does not count it as a failure of this change when it fails the same way as at the baseline commit, and the pull request lists it. Record any other product discrepancy as a finding, not a fix.

## Report

10. Add to flow-close's final report, against the baseline:
    - test lines and production lines, before and after, separately, and the number of tests;
    - coverage before and after, when the repository measures it;
    - each suite's wall time before and after: moving proof to a stronger boundary can make a suite slower;
    - the lanes, the keeper of each contract, the repaired tests, the retired files and the removed seams;
    - the preservation gaps found and the mutation that proves each contract that lost or repaired a test;
    - every baseline failure with what was decided for it, the product defects with their control, and the follow-ups.

    Name each test kept for lack of evidence. Never set or report a deletion target.
