# Playbook: hotfix

Read from [hive-mode](../SKILL.md#pick-the-playbook) when a production defect cannot wait for the next scheduled promotion, because the user says so or because the defect is P0 as [close report](../../hive-rule-close-report/SKILL.md#findings) defines it. [urgent fixes](../../flow-deploy/references/urgent-fixes.md) owns the definition, the promotion route, the hotfix branch, the mitigation rules and the postmortem; this page does not restate them.

**Stages.** `hive-mode`, then [flow-build](../../flow-build/SKILL.md) for the fix, which integrates into the base branch as the delivery mode in force sets, then [flow-verify](../../flow-verify/SKILL.md), then [flow-deploy](../../flow-deploy/SKILL.md) through the urgent-fixes route, then [flow-close](../../flow-close/SKILL.md).

**Specific to a hotfix**
- Reduce the incident's ongoing impact first, within existing authority, and keep the evidence; tell a temporary mitigation from a verified fix.
- The fix is still a defect fix: the cause is named and fixed at its source ([defect fix](../../hive-rule-defect-fix/SKILL.md)), with the smallest change that removes it and a regression check.
- Each stop, push, pull request and merge happens under the delivery mode in force. The promotion is never covered by it: [flow-deploy](../../flow-deploy/SKILL.md) asks its own authorization for each environment, and urgent-fixes sets what is reported when the mode in force publishes nothing.
- Before each promotion, list the other commits it would carry and ask about them, as urgent-fixes sets.
- Environment branches change only through the declared path: [promotion path](../../hive-rule-promotion-path/SKILL.md).

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-rule-close-report/SKILL.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-rule-report-endings/SKILL.md) that [close modes](../../hive-rule-close-modes/SKILL.md) sets.

**Handoffs.** Work the incident exposes beyond the fix, such as a missing guard or a refactor, is separate work, as [scope changes](../../hive-rule-scope-changes/SKILL.md) sets.

**Leaves.** [verification floor](../../hive-rule-verification-floor/SKILL.md), [urgent fixes](../../flow-deploy/references/urgent-fixes.md), [defect fix](../../hive-rule-defect-fix/SKILL.md), [promotion path](../../hive-rule-promotion-path/SKILL.md), [foreign commits](../../hive-rule-foreign-commits/SKILL.md), [merge gate](../../hive-rule-merge-gate/SKILL.md).
