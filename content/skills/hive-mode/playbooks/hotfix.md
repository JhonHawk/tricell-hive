# Playbook: hotfix

Read from [hive-mode](../SKILL.md#pick-the-playbook) when a production defect cannot wait for the next scheduled promotion, because the user says so or because the defect is P0 as [close report](../../hive-core/close/close-report.md#findings) defines it. [urgent fixes](../../flow-deploy/references/urgent-fixes.md) owns the definition, the promotion route, the hotfix branch, the mitigation rules and the postmortem; this page does not restate them.

**Stages.** `hive-mode`, then [flow-build](../../flow-build/SKILL.md) for the fix, which integrates into the base branch as the delivery mode in force sets, then [flow-verify](../../flow-verify/SKILL.md), then [flow-deploy](../../flow-deploy/SKILL.md) through the urgent-fixes route, then [flow-close](../../flow-close/SKILL.md).

**Specific to a hotfix**
- While the incident lasts, mitigate as [urgent fixes](../../flow-deploy/references/urgent-fixes.md) sets; the fix itself is a defect fix, as [defect fix](../../hive-core/build/defect-fix.md) sets.
- Each stop, push, pull request and merge happens under the delivery mode in force. The promotion is never covered by it: [flow-deploy](../../flow-deploy/SKILL.md) asks its own authorization for each environment, and urgent-fixes sets what is reported when the mode in force publishes nothing.
- Before each promotion, list the other commits it would carry and ask about them, as urgent-fixes sets.
- Environment branches change only through the declared path: [promotion path](../../hive-core/git/promotion-path.md).

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-core/close/close-report.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-core/messages/report-endings.md) that [close modes](../../hive-core/close/close-modes.md) sets; a promotion's cleanup close uses the deployment row of [close report](../../hive-core/close/close-report.md#variants).

**Handoffs.** Work the incident exposes beyond the fix, such as a missing guard or a refactor, is separate work, as [scope changes](../../hive-core/plan/scope-changes.md) sets.

**Leaves.** [verification floor](../../hive-core/build/verification-floor.md), [urgent fixes](../../flow-deploy/references/urgent-fixes.md), [defect fix](../../hive-core/build/defect-fix.md), [promotion path](../../hive-core/git/promotion-path.md), [foreign commits](../../hive-core/git/foreign-commits.md), [merge gate](../../hive-core/git/merge-gate.md).
