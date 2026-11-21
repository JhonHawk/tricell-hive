# Playbook: ui-change

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work adds a screen or changes what one shows, how it is composed or how it behaves. A change that keeps the layout, such as a copy or a field, can be a [feature](feature.md).

**Stages.** `hive-mode`, then [create-design](../../create-design/SKILL.md) when [design stage](../../hive-core/references/design-stage.md) offers it and the user chooses it, then [flow-plan](../../flow-plan/SKILL.md), [flow-build](../../flow-build/SKILL.md), [flow-verify](../../flow-verify/SKILL.md) and [flow-close](../../flow-close/SKILL.md). Without a Figma connection or file, the plan settles the UI.

**Specific to a screen change**
- The project's design system governs: the [UI criteria](../../create-design/references/ui-criteria.md) detect it, keep its `DESIGN.md` verified and give observable criteria that never override it.
- Plan the screen as [UI decisions](../../hive-core/references/ui-plan.md) sets; when the work folder holds an approved Figma example, [approved example](../../hive-core/references/ui-example.md) makes it the plan's decision, and an example still pending settles nothing.
- Before the first edit of an existing screen, capture it unchanged: [before captures](../../hive-core/references/before-captures.md).
- The build shows the user the first rendered view and adjusts only the UI before continuing ([first view](../../flow-build/references/first-view.md), from flow-build step 4).
- Verification captures and reviews the screen as [UI verification](../../hive-core/references/ui-verification.md) sets, audits accessibility as [accessibility audit](../../hive-core/references/a11y-audit.md) sets, and shows the result with the [walkthrough](../../hive-core/references/walkthrough.md); the user's validation stop is part of the delivery mode in force.

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-core/references/close-report.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-core/references/report-endings.md) that [close modes](../../hive-core/references/close-modes.md) sets.

**Handoffs.** Behavior with no screen change is [feature](feature.md); a broken screen the user wants fixed at its cause is [bugfix](bugfix.md). Either is a scope change, as [scope changes](../../hive-core/references/scope-changes.md) sets.

**Leaves.** [verification floor](../../hive-core/references/verification-floor.md), [viewer](../../hive-core/references/viewer.md), [open the viewer](../../hive-core/references/viewer-open.md), [close the viewer](../../hive-core/references/viewer-close.md), [tests](../../hive-core/references/tests.md), [checks](../../hive-core/references/checks.md).
