# Playbook: pr-review-deploy

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work is to review a pull request, to take one through its reviews and merge, or to promote what is already integrated.

**Stages.** `hive-mode`, then a review of the change, then, when the review finds blocking defects the work fixes, [flow-build](../../flow-build/SKILL.md) for those fixes only, then the merge gate, then [flow-deploy](../../flow-deploy/SKILL.md) when the user asks to promote, then [flow-close](../../flow-close/SKILL.md). Promotion of an integrated candidate with no review needs only `flow-deploy`.

**Specific to reviewing and promoting**
- Review the diff against its request and acceptance criteria with the `hive-review-code` role or the host's native review, briefed with the artifact, the requirements and the check output, never the author's case ([reviewer brief](../../hive-rule-reviewer-brief/SKILL.md)). A finding is confirmed by evidence, not by a reviewer's assertion ([review evidence](../../hive-rule-review-evidence/SKILL.md)).
- Inventory added suppressions, skipped tests and exclusions that weaken checks ([weakened checks](../../hive-rule-suppression-review/SKILL.md)).
- Read the pull request's checks and reviews once and answer each comment ([PR reviews](../../hive-rule-pr-reviews/SKILL.md)); the merge waits on the CI result, the current-base check and any human reviewer ([merge gate](../../hive-rule-merge-gate/SKILL.md)).
- Merging a pull request this work did not write needs the user's instruction in every case; the delivery mode in force covers only the fixes the work commits itself, and this work never merges its own pull request when that mode stops it open for the user.
- Name commits on the branch that are not this work's before the validation or the push ([foreign commits](../../hive-rule-foreign-commits/SKILL.md)).
- Promotion follows [flow-deploy](../../flow-deploy/SKILL.md): its authorization per environment and the served-identity checks. Environment branches change only through the declared path ([promotion path](../../hive-rule-promotion-path/SKILL.md)).
- A check that signs in reads the account index first ([test accounts](../../hive-rule-test-accounts/SKILL.md)).

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-rule-close-report/SKILL.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-rule-report-endings/SKILL.md) that [close modes](../../hive-rule-close-modes/SKILL.md) sets; a promotion's cleanup close uses the deployment row of [close report](../../hive-rule-close-report/SKILL.md#variants).

**Handoffs.** A defect the review confirms and the user wants fixed in this pull request is a correction in the same work; one that needs its own change is [bugfix](bugfix.md) or a separate topic, as [scope changes](../../hive-rule-scope-changes/SKILL.md) sets.

**Leaves.** [merge gate](../../hive-rule-merge-gate/SKILL.md), [PR reviews](../../hive-rule-pr-reviews/SKILL.md), [review evidence](../../hive-rule-review-evidence/SKILL.md), [reviewer brief](../../hive-rule-reviewer-brief/SKILL.md), [weakened checks](../../hive-rule-suppression-review/SKILL.md), [foreign commits](../../hive-rule-foreign-commits/SKILL.md), [promotion path](../../hive-rule-promotion-path/SKILL.md).
