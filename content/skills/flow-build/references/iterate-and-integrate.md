# Iterate on checks and integrate without verify

## When to read

Read in [flow-build](../SKILL.md) before running the checks of your change, while iterating on those checks, and before integrating when verify is skipped. Reading this reference does not invoke another workflow.

## Iterate on the checks

While iterating, run the project's affected-test selection for the changed code, widened and bounded as [the verification reference](verification.md#frequency-and-coverage) sets. For a mechanical rename or move, find every reference to the old name or path, including strings and paths the compiler cannot see, and run the tests of the packages that hold them. When the verification the change needs is much larger than the change itself, say so in the progress update with the reason. After a fix, rerun the failing check and affected scope; broaden to affected callers and consumers when the fix touches shared code or new failure evidence appears. Report each check's status as [verification](verification.md#report-check-status) sets; resolve authorized local prerequisites and report any remaining delivery constraint. Check that the handoff's paths and next step still match the final state.

## Integrate when verify is skipped

When `flow-verify` is skipped, flow-build integrates the candidate: a mechanical change reports the checks that already cover it, and a skip the user asked for names each skipped check as accepted without verification, as [flow-verify](../../flow-verify/SKILL.md#decide-whether-to-run-and-at-what-level) sets. Integration steps stay in [git-workflow](../../git-workflow/SKILL.md) and [delivery decisions](../../flow-plan/references/delivery-decisions.md); before a branch, worktree, commit, push, pull request, or merge, read git-workflow first. Under `interactive`, stop before the push with the diff and the checks that cover it for the user's validation, as [delivery decisions](../../flow-plan/references/delivery-decisions.md) sets, even though verify was skipped. Complete only the delivery actions already explicitly authorized, subject to their conditions; otherwise stop at the authorized result with remaining delivery steps identified.

When verify is skipped, report what changed, what was verified, unresolved failures or uncertainty, and the next step, in the validation request or completion report when one follows.
