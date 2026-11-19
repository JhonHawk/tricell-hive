# Playbook: ui-change

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work adds a screen or changes what one shows, how it is composed or how it behaves. A change that keeps the layout, such as a copy or a field, can be a [feature](feature.md).

**Stages.** `hive-mode`, then [flow-figma-design](../../flow-figma-design/SKILL.md) when [design stage](../../hive-rule-design-stage/SKILL.md) offers it and the user chooses it, then [flow-plan](../../flow-plan/SKILL.md), [flow-build](../../flow-build/SKILL.md), [flow-verify](../../flow-verify/SKILL.md) and [flow-close](../../flow-close/SKILL.md). Without a Figma connection or file, the plan settles the UI.

**Specific to a screen change**
- The project's design system governs: [hive-ui-criteria](../../hive-ui-criteria/SKILL.md) detects it, keeps its `DESIGN.md` verified and gives observable criteria that never override it.
- Plan the screen as [UI decisions](../../hive-rule-ui-plan/SKILL.md) sets; when the work folder holds an approved Figma example, [approved example](../../hive-rule-ui-example/SKILL.md) makes it the plan's decision, and an example still pending settles nothing.
- Before the first edit of an existing screen, capture it unchanged: [before captures](../../hive-rule-before-captures/SKILL.md).
- The build shows the user the first rendered view and adjusts only the UI before continuing ([flow-build](../../flow-build/SKILL.md) step 4).
- Verification captures and reviews the screen as [UI verification](../../hive-rule-ui-verification/SKILL.md) sets, audits accessibility as [accessibility audit](../../hive-rule-a11y-audit/SKILL.md) sets, and shows the result with the [walkthrough](../../hive-rule-walkthrough/SKILL.md); the user's validation stop is part of the delivery mode in force.

**Handoffs.** Behavior with no screen change is [feature](feature.md); a broken screen the user wants fixed at its cause is [bugfix](bugfix.md). Either is a scope change, as [scope changes](../../hive-rule-scope-changes/SKILL.md) sets.

**Leaves.** [verification floor](../../hive-rule-verification-floor/SKILL.md), [viewer](../../hive-rule-viewer/SKILL.md), [open the viewer](../../hive-rule-viewer-open/SKILL.md), [close the viewer](../../hive-rule-viewer-close/SKILL.md), [tests](../../hive-rule-tests/SKILL.md), [checks](../../hive-rule-checks/SKILL.md).
