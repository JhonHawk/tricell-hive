# Playbook: feature

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work adds or extends behavior and changes no screen's layout. A new or recomposed screen is [ui-change](ui-change.md).

**Stages.** `hive-mode` settles the what, then [flow-plan](../../flow-plan/SKILL.md), or straight to [flow-build](../../flow-build/SKILL.md) when the change is a small understood change (the global guidance's «Proportionality»), then [flow-verify](../../flow-verify/SKILL.md) and [flow-close](../../flow-close/SKILL.md). [flow-deploy](../../flow-deploy/SKILL.md) runs only when the user asks to promote.

**Specific to a feature**
- Write the acceptance criteria first, one observable behavior each, with the place where its real effect shows, as [acceptance criteria](../../flow-plan/SKILL.md#acceptance-criteria) sets. A criterion that is already true on the base branch grades nothing.
- Settle each choice the request leaves open in the plan, never in the code. When a decision or an assumption adds a business rule, record it as [business rules](../../hive-rule-business-rules/SKILL.md) sets.
- Weigh the approaches against what the project already has before writing anything new: [reuse before writing](../../hive-rule-reuse-before-writing/SKILL.md).
- When the approach reads or changes a system outside the repository, list and measure its facts first ([external facts](../../hive-rule-external-facts/SKILL.md)) and plan its effects as [effects round](../../hive-rule-external-effects/SKILL.md) sets.
- Tests follow [tests](../../hive-rule-tests/SKILL.md): the ones that pin the new behavior, and none that only restate the code.
- Each stop, push, pull request and merge happens under the delivery mode in force.

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-rule-close-report/SKILL.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-rule-report-endings/SKILL.md) that [close modes](../../hive-rule-close-modes/SKILL.md) sets.

**Handoffs.** A request that turns out to need a new or recomposed screen goes to `ui-change`; a defect found along the way is a separate topic, or `bugfix` when the user asks for its fix. Either is a scope change, as [scope changes](../../hive-rule-scope-changes/SKILL.md) sets.

**Leaves.** [verification floor](../../hive-rule-verification-floor/SKILL.md), [change folder](../../hive-rule-change-folder/SKILL.md), [checks](../../hive-rule-checks/SKILL.md), [project settings](../../hive-rule-project-settings/SKILL.md) when a `## Hive` key is needed.
