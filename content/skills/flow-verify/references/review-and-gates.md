# Review and gates for a candidate

Read by [flow-verify](../SKILL.md) when it runs the dedicated code review, the local runtime checks or the human UI handoff on a candidate. [Verification by risk and surface](../../flow-build/references/verification.md) keeps the evidence selection, the gate authorization rule and the frequency rules; this reference sets what a candidate goes through before it is accepted and integrated. Verification does not grant deployment, publication, destructive effects, or external access.

## Local candidate and publication

| Boundary | Required decision or evidence |
| --- | --- |
| Local candidate | Complete applicable local in-vivo and UI verification, using the built artifact for final UI/runtime validation where the stack has a build step; a rendered UI effect adds the independent children that [the UI gate](ui-gate.md) selects, then deliver it with [flow-walkthrough](../../flow-walkthrough/SKILL.md), after the sequence that [flow-verify](../SKILL.md) sets before asking the user to validate. Run the dedicated code review that the review question settled, at the moment [Dedicated review timing](#dedicated-review-timing) sets; propose it here only when it was not settled. |
| Publication / integration | Inspect actual CI coverage; the selected review mechanism ran on the local candidate, or runs on the pull request for a mechanism that only works there, as [Dedicated review timing](#dedicated-review-timing) sets. Account for automatic review/deployment effects before the push or PR event that triggers them. |

## Dedicated code review

When the selected mechanism is `hive-review-code`, or any review that does not verify its own findings, send each P0 or P1 finding, and any finding that would block the merge, to a `hive-refute-claim` child before acting on it. Fix confirmed and plausible findings under the normal rules, drop refuted ones with the refuter's evidence, and report both. Native reviews that already verify each finding, such as Claude Code's `/code-review` at medium effort or above, need no second pass.

Apply these rules to the findings of a code review:

- **Severity.** When a mechanism reports its own severity levels, map each finding to P0–P3 as the severity definitions in [incidental findings](../../flow-close/references/incidental-findings.md#which-findings-are-listed-and-how) set, before applying the rules below.
- **P0 and P1.** Fix the ones that are confirmed or plausible after refutation, or verified by a mechanism that checks its own findings.
- **P2.** Fix one the change introduced when the fix is local and inside the change's scope; otherwise handle it under the shared incidental-findings rule.
- **P3.** Report without fixing or recording elsewhere; a P3 finding is not a defect under the shared incidental-findings rule. The one exception is a P3 global finding of the UI review, which the [UI gate](ui-gate.md#the-independent-children) records once in the project UI guide without a question.
- **Pre-existing findings.** Dispose of them under the shared incidental-findings rule. A pre-existing defect in a path the change makes reachable, or whose access it widens, counts as introduced by the change, whatever label the reviewer gave it; check that against the diff for each finding labeled pre-existing.
- **Review count.** Review each candidate in full once, at the moment [Dedicated review timing](#dedicated-review-timing) sets, with one code reviewer per candidate.
- **Re-review.** Re-review at most once, limited to the diff of the fixes against the reviewed commit, or under `hold` the reviewed candidate identity, and only when a P0 or P1 fix changed logic or a contract. After that round, fix a finding still open or newly reported, without another review, when it would be fixed under these rules, its fix is local and inside the change's scope, and it changes no contract; cover it with a test and report it as fixed without review. Any other finding still open after that round goes to the user as a decision.
- **Later edits.** Later edits to the same candidate, including added scope, never start another full review. When an earlier review answer covers them, review the diff since the reviewed commit, or under `hold` since the reviewed candidate identity that [the UI gate](ui-gate.md#the-independent-children) defines, with the same mechanism; otherwise report what changed since that point and ask the user whether to review it.

A technical verdict does not authorize delivery. A required gate declined or unavailable remains an unmet transition condition; an optional gate may be deferred explicitly. Interpret actual findings and terminal status rather than treating absent comments, a queued run, or a non-failing check as approval.

## Security review findings

Offer `hive-review-security` only as [Offer the security review](../../flow-build/references/security-boundaries.md#offer-the-security-review) sets, and run it only as [flow-security](../../flow-security/SKILL.md) sets. That stage reports and does not fix; when the candidate is not yet integrated, the stage that asked for the review then handles its findings with the rules above:

- Map each finding to P0–P3 and refute it as [Dedicated code review](#dedicated-code-review) sets for a review that does not verify its own findings, which the security reviewer does not.
- Fix, drop and report as the code-review rules set, and re-review at most once under the same limits, with `hive-review-security` limited to the diff of the fixes.
- When both reviews are asked before the code review starts, run them in parallel, not one after the other; a security review asked later, such as at the validation or beside a code review that only runs on the pull request, runs when asked. Handle both reports before the re-review round.
- A finding on a merged commit is not handled here: its fix goes in a follow-up pull request, as [delivery decisions](../../flow-plan/references/delivery-decisions.md) sets.

## Dedicated review timing

The dedicated code review, whichever mechanism the answer selects, runs once on the local candidate, after the affected tests and the full suite when it runs locally, and before the next step that publishes it or asks for its validation: under `interactive`, before the validation request; under `automatic`, and under `direct-base` chosen for a code repository, before the push; under `hold`, on the verified working tree. Fix its findings before that step, so the user validates, and the pull request carries, a reviewed candidate. A mechanism that only works on a pull request, such as a pull-request bot or a human reviewer, runs on the pull request right after the push instead, and the merge waits for its findings; under `interactive` that push follows the validation, so the validation message says the review is still due, and under `hold` or a chosen `direct-base`, which open no pull request, ask for a local mechanism or report the review as not run. Record in the pull request body which mechanism reviewed which commit, so a later promotion can tell which commits were reviewed. The corrections the user's validation asks for get one more review, limited to the diff between the reviewed candidate and the accepted one, alongside the accepted-candidate gate of [the UI gate](ui-gate.md); no correction round runs a full review. A later edit to the same candidate follows [Dedicated code review](#dedicated-code-review). A follow-up pull request of the same ticket gets its review on its own local candidate in the same way, and a promotion pull request reviews only the commits not reviewed before their integration, as the promotion rule in [delivery decisions](../../flow-plan/references/delivery-decisions.md) sets.

## Runtime and UI gates

Decide local in-vivo applicability from the actual changed behavior. When needed, include and execute the relevant walk within authorized implementation without asking whether to run it. When judged unnecessary, say why in one line of the report and do not run it; the user can ask for it. That judgment never waives a required gate, overrides paused model pilots, or grants external effects. Deployed-environment walks retain the explicit authorization boundary in [gate selection and timing](../../flow-build/references/verification.md#gate-selection-and-timing).

For local implementation, always exercise the applicable changed path and relevant failure or permission states. Perform the corresponding deployed-environment exercise only within its explicit gate authorization. A mocked integration or source inspection does not establish behavior in the running system. Resolve authorized local setup, accounts, roles and seed data; a failure while setting one up is a defect that blocks the work, fixed within the stage's fix-round limits; access only the user can supply, or local setup still failing after those rounds, remains a reported prerequisite, named with its cause. Before any request for the user's validation, open each target the request names as the user will, such as a page with its seeded record, rendered under the browser rules below, or a running service: a target that fails is such a setup failure, fixed before the request, and one still failing after those rounds is named in the request with its cause.

Use dev/HMR for iteration. For the final in-vivo or UI handoff, build and serve the compiled candidate using the project's production-style start or preview command where the stack distinguishes build from dev. Services without a build step use their normal runtime. Build only the affected applications and required dependencies, not every app per edit.

Default the handoff stack to local frontend, backend, database and necessary dependencies unless a different target was explicitly specified. Check effective connections, including build-time frontend URLs and backend persistence targets, without exposing credentials; localhost in the browser does not prove the whole path is local. Prepare authorized local accounts, roles and seed data before the walk, through the project's seed scripts, API or database, and pass each child the identifiers of the records it needs; a child creates data through the interface only when that creation is a journey under test. Report missing dependencies and any uncovered integration; never silently substitute a remote environment. External authentication, messaging, payments and similar effects need their own authorized boundary.

For any change with a rendered UI effect, read and follow [the UI gate](ui-gate.md): it owns the implementer's rendered inspection, the independent children that the `UI verification` answer selects with their briefs and stops, and the accepted-candidate gate, in every delivery mode. Before delivering a later fix, including a code-review fix after approval, apply its [changes after a gate or approval](ui-gate.md#changes-after-a-gate-or-approval) rule.

Before driving a browser, read [browser automation](../../flow-build/references/browser-automation.md) and pass it to children that drive one.

## Human UI handoff

Before showing a user interface to the user for validation, read and follow [flow-walkthrough](../../flow-walkthrough/SKILL.md): it owns the early composition check on the dev server, the captures of relevant states, and the handoff message and page. [flow-verify](../SKILL.md) owns the validation loop: the correction rounds, their task markers and the accepted-candidate gate.
