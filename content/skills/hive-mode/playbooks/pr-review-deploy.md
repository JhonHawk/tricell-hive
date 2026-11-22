# Playbook: pr-review-deploy

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work is to review a pull request, to take one through its reviews and merge, or to promote what is already integrated.

**Stages.** `hive-mode`, then a review of the change, then, when the review finds blocking defects the work fixes, [flow-build](../../flow-build/SKILL.md) for those fixes only, then the merge gate, then [flow-deploy](../../flow-deploy/SKILL.md) when the user asks to promote, then [flow-close](../../flow-close/SKILL.md). Promotion of an integrated candidate with no review needs only `flow-deploy`.

**Specific to reviewing and promoting**
- Review the diff against its request and acceptance criteria with the `hive-review-code` role or the host's native review, briefed with the artifact, the requirements and the check output, never the author's case ([reviewer brief](../../hive-core/delegation/reviewer-brief.md)). A finding is confirmed by evidence, not by a reviewer's assertion ([review evidence](../../hive-core/review/review-evidence.md)).
- Inventory added suppressions, skipped tests and exclusions that weaken checks ([weakened checks](../../hive-core/review/suppression-review.md)).
- Read the pull request's checks and reviews once and answer each comment ([PR reviews](../../hive-core/git/pr-reviews.md)); the merge waits on the CI result, the current-base check and any human reviewer ([merge gate](../../hive-core/git/merge-gate.md)).
- Merging a pull request this work did not write needs the user's instruction in every case; the delivery mode in force covers only the fixes the work commits itself, and this work never merges its own pull request when that mode stops it open for the user.
- Commits that are not this work's are named and asked about when [foreign commits](../../hive-core/git/foreign-commits.md) says.
- Promotion follows [flow-deploy](../../flow-deploy/SKILL.md): its authorization per environment and the served-identity checks. Environment branches change only through the declared path ([promotion path](../../hive-core/git/promotion-path.md)).
- A check that signs in reads the account index first ([test accounts](../../hive-core/local/test-accounts.md)).

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-core/close/close-report.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-core/messages/report-endings.md) that [close modes](../../hive-core/close/close-modes.md) sets; a promotion's cleanup close uses the deployment row of [close report](../../hive-core/close/close-report.md#variants).

**Handoffs.** A defect the review confirms and the user wants fixed in this pull request is a correction in the same work; one that needs its own change is [bugfix](bugfix.md) or a separate topic, as [scope changes](../../hive-core/plan/scope-changes.md) sets.

**Leaves.** [merge gate](../../hive-core/git/merge-gate.md), [PR reviews](../../hive-core/git/pr-reviews.md), [review evidence](../../hive-core/review/review-evidence.md), [reviewer brief](../../hive-core/delegation/reviewer-brief.md), [weakened checks](../../hive-core/review/suppression-review.md), [foreign commits](../../hive-core/git/foreign-commits.md), [promotion path](../../hive-core/git/promotion-path.md).
