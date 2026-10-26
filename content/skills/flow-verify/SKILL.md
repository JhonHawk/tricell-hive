---
name: flow-verify
description: "Verify a built candidate before it is accepted: affected tests, runtime checks, the UI gate, the full suite and the dedicated review, at the level the delivery, review and UI answers and the request set. Use when a candidate exists and needs evidence; not for writing code, deploying, or checking a deployed environment."
---

# Verify a built candidate

## Entry and exit

Enter with an identified candidate (commits, or the working tree under `hold` delivery), the tasks verified at their checkpoints, the criteria it must meet, and the review values resolved; when these exist, enter directly, with no plan or build before it. Exit with the level that ran and each check not met named, and with the accepted candidate integrated as the delivery mode sets, through [git-workflow](../git-workflow/SKILL.md), or the verified working tree under `hold`. This skill does not grant effects beyond the user's instructions.

## Prepare

Before resolving needed project settings, read [project settings](../flow-plan/references/project-settings.md). Before a user decision or manual validation request, read [questions](../flow-plan/references/questions.md). Before delegating or starting an external-work wait, read [delegation and waits](../flow-research/references/delegation.md). Before progress reports, completion reports or messages containing citations, links or externally defined IDs, read [reporting](../flow-close/references/reporting.md). Before writing or cleaning support evidence and UI images, read [artifact layout and retention](../workspace-conventions/references/artifact-layout.md). To select evidence for the changed surface, read [verification by risk and surface](../flow-build/references/verification.md). Before the dedicated code review, the local runtime checks or the human UI handoff, read [review and gates](references/review-and-gates.md). Before any gate on a change with a rendered UI effect, read [the UI gate](references/ui-gate.md); the reviewer's brief is [UI review criteria](references/ui-review-criteria.md).

## Decide whether to run and at what level

At entry, say in one line whether verify runs or is skipped, and at which level.

Verify is skipped only for a mechanical change, as the global guidance defines it, or when the user asks; a skip asked for after this stage started runs it at no level, and this stage still integrates. Name what was skipped in the report. In a mechanical change, report the checks that already cover it. A dedicated code review is not needed there.

The level combines values that already exist; no new setting selects it. Resolve the questions behind them where [delivery decisions](../flow-plan/references/delivery-decisions.md) and [review decisions](../flow-plan/references/review-decisions.md) set: at the plan's close, or at the start of `flow-build` when there is no plan, or reused from earlier in the session. This skill adds no question: it reads those answers, announces the level and runs it. Only when the work enters here directly, with no plan and no build before it, run that same round once, for the missing values.

| Check | Runs when | Set by |
| --- | --- | --- |
| Affected tests and evidence | whenever verify runs | none |
| Full local suite | before a validation stop, or once on the final candidate when no CI gate runs it, in a change that is not mechanical | the full-suite rule below |
| Local in-vivo run | the changed behavior calls for it; when it does not, the report says why | the rule in [review and gates](references/review-and-gates.md) |
| UI gate | the change has a rendered UI effect | the `UI verification` answer (`full`, `functional`, `none`) |
| Dedicated code review | on the local candidate, as [Dedicated review timing](../flow-plan/references/review-decisions.md#dedicated-review-timing) sets | the answer to the review question |
| Security review | only when the user asked for it | [flow-security](../flow-security/SKILL.md) |

The user can raise or lower the level for the work with a request such as "full verify", "no UI review" or "skip verify"; the request holds for that work. Skipping verify or lowering its level at the user's request counts as accepting without verification each check it omits: name them one by one at entry and in the report, and treat a check accepted this way as no pending gate. Only the user's own message, before the check runs, asks for it; declining a gate when asked is not a skip; unattended delegation can neither skip verify nor lower its level. This acceptance never waives CI, branch protection, or a review the project requires. The chosen review mechanism is the default and not a demand: when the user skips verify, the dedicated review counts as accepted without verification, unless the repository requires it through branch protection or a required review.

## Verify the candidate

Compare the final changes and observed behavior against the acceptance criteria and authorized scope. Run the affected tests and the applicable build, typecheck and lint commands, and exercise changed runtime behavior through the real integration path. Tests passing or a successful build do not replace those observations. Before reporting a UI change complete, obtain the results of the independent UI review and in-vivo children that [the UI gate](references/ui-gate.md) selects by the `UI verification` answer; the implementer's own walk satisfies neither.

In any delivery mode that stops for validation, run these steps in order on the built candidate before asking the user to validate a result; they do not run before the first rendered view that [flow-walkthrough](../flow-walkthrough/SKILL.md) shows:

1. The UI gate that the `UI verification` answer selects.
2. The full suite of each affected application, locally, even when CI will run it later, since under `interactive` CI runs after the push that follows validation.
3. The dedicated code review, as [Dedicated review timing](../flow-plan/references/review-decisions.md#dedicated-review-timing) sets, with its findings fixed.
4. The security review, run as [flow-security](../flow-security/SKILL.md) sets, when the user asked for it; otherwise the validation message offers it for any boundary no earlier offer named, as [Offer the security review](../flow-build/references/security-boundaries.md#offer-the-security-review) sets. Both also apply without a validation stop.
5. The delivery of [flow-walkthrough](../flow-walkthrough/SKILL.md), which gives the page's path and asks the validation question it sets.

These steps run before the first validation request. Before delivering any later fix with a UI effect, follow [changes after a gate or approval](references/ui-gate.md#changes-after-a-gate-or-approval). Without a validation stop, run the dedicated code review at the moment [Dedicated review timing](../flow-plan/references/review-decisions.md#dedicated-review-timing) sets; [review and gates](references/review-and-gates.md) sets how its findings are handled.

Use existing evidence only while its relevant code, configuration, and environment remain valid. Close regression coverage through the project's actual CI gate, except that the full suite also runs locally before asking the user to validate, as the sequence above sets. When no CI gate will run it, run the full suite once on the final candidate of any change that is not mechanical, and rerun it only when a later edit could affect what it covers. Do not rerun a passing suite for confidence. Report each check's status as [verification](../flow-build/references/verification.md#report-check-status) sets.

## Validation by the user

For UI work awaiting human validation, keep the verified stack and the handoff as [flow-walkthrough](../flow-walkthrough/SKILL.md) sets: the production-style build of the candidate, with local frontend, backend, database and dependencies, left running.

While the user requests corrections during a validation pause, whether to text, behavior, or layout, run one correction round per request: apply the correction; write its test in the same round when it alters behavior of a kind [test-first development](../flow-build/references/tdd.md) covers; run the affected tests; rebuild; capture the affected views; and ask again right away in the same way. Between rounds, run no `hive-review-ux`, `hive-verify-change`, `hive-verify-task`, or dedicated code review; the corrections get the scoped review that [Dedicated review timing](../flow-plan/references/review-decisions.md#dedicated-review-timing) sets, at the accepted-candidate gate. When you offer alternatives, sketch them as [UI planning](../flow-plan/references/ui-planning.md) describes.

When the user accepts, run the accepted-candidate gate as [the UI gate](references/ui-gate.md) sets before any further step; at that gate, rerun the full suite only when the corrections could affect what it covers, and offer the security review for the corrections as [Offer the security review](../flow-build/references/security-boundaries.md#offer-the-security-review) sets, or rerun it as [flow-security](../flow-security/SKILL.md) sets. A local run that makes no progress stops as [verification](../flow-build/references/verification.md) sets, and the validation message names it as not run.

Markers: when a correction changes the files of a task marked `[x]`, set that task back to `[?]` with a line `reopened by correction round <n>`, and verify it at the accepted-candidate gate. Record a correction outside the plan's tasks in the progress section; `hive-verify-change` covers it at that gate when the value runs it. A correction that changes the retained plan follows [Revise a plan in progress](../flow-plan/references/plan-revision.md); the plan re-review it may make due runs with the accepted-candidate gate above, not between rounds.

## Integrate and close

Integration is the last stage that runs before `flow-close`: this skill in the normal flow, and `flow-build` when verify is skipped. Its steps, their order and its exceptions stay in [git-workflow](../git-workflow/SKILL.md) and [delivery decisions](../flow-plan/references/delivery-decisions.md); before a branch, worktree, commit, push, pull request or merge, read git-workflow. Do not restate them here. The dedicated code review has run on the local candidate, or runs on the pull request for a mechanism that only works there, as [Dedicated review timing](../flow-plan/references/review-decisions.md#dedicated-review-timing) sets. Otherwise stop at the authorized result with the remaining delivery steps identified.

Report what changed, what was verified and at which level, each check not met or accepted without verification, unresolved failures or uncertainty, and the next step. A successful tool exit, a checked box, or an agent's assertion alone does not establish completion. Run [flow-close](../flow-close/SKILL.md) as its [When to run it](../flow-close/SKILL.md#when-to-run-it) sets; inside an iterative route, as [flow-research](../flow-research/SKILL.md) sets.
