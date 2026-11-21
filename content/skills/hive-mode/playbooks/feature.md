# Playbook: feature

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work adds or extends behavior and changes no screen's layout. A new or recomposed screen is [ui-change](ui-change.md).

**Stages.** `hive-mode` settles the what, then [flow-plan](../../flow-plan/SKILL.md), or straight to [flow-build](../../flow-build/SKILL.md) when the change is a small understood change (the global guidance's «Proportionality»), then [flow-verify](../../flow-verify/SKILL.md) and [flow-close](../../flow-close/SKILL.md). [flow-deploy](../../flow-deploy/SKILL.md) runs only when the user asks to promote.

**Specific to a feature**
- Write the acceptance criteria first, one observable behavior each, with the place where its real effect shows, as [acceptance criteria](../../flow-plan/SKILL.md#acceptance-criteria) sets. A criterion that is already true on the base branch grades nothing.
- Settle each choice the request leaves open in the plan, never in the code. When a decision or an assumption adds a business rule, record it as [business rules](../../hive-core/references/business-rules.md) sets.
- Weigh the approaches against what the project already has before writing anything new: [reuse before writing](../../hive-core/references/reuse-before-writing.md).
- When the approach reads or changes a system outside the repository, list and measure its facts first ([external facts](../../hive-core/references/external-facts.md)) and plan its effects as [effects round](../../hive-core/references/external-effects.md) sets.
- Tests follow [tests](../../hive-core/references/tests.md): the ones that pin the new behavior, and none that only restate the code.
- Each stop, push, pull request and merge happens under the delivery mode in force.

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-core/references/close-report.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-core/references/report-endings.md) that [close modes](../../hive-core/references/close-modes.md) sets; a promotion's cleanup close uses the deployment row of [close report](../../hive-core/references/close-report.md#variants).

**Handoffs.** A request that turns out to need a new or recomposed screen goes to `ui-change`; a defect found along the way is a separate topic, or `bugfix` when the user asks for its fix. Either is a scope change, as [scope changes](../../hive-core/references/scope-changes.md) sets.

**Leaves.** [verification floor](../../hive-core/references/verification-floor.md), [change folder](../../hive-core/references/change-folder.md), [checks](../../hive-core/references/checks.md), [project settings](../../hive-core/references/project-settings.md) when a `## Hive` key is needed.
