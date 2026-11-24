# Playbook: ui-change

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work adds a screen or changes what one shows, how it is composed or how it behaves. A change that keeps the layout, such as a copy or a field, is a `feature`.

**Stages.** `hive-mode`, then [create-design](../../create-design/SKILL.md) when [design stage](../../hive-core/ui/design-stage.md) offers it and the user chooses it, then `flow-plan`, `flow-build`, `flow-verify` and `flow-close`. Without a Figma connection or file, the plan settles the UI.

**Specific to a screen change**
- The project's design system governs: the [UI criteria](../../create-design/references/ui-criteria.md) detect it, keep its `DESIGN.md` verified and give observable criteria that never override it.
- Plan the screen as [UI decisions](../../hive-core/ui/ui-plan.md) sets; an approved Figma example in the work folder is the plan's decision ([approved example](../../hive-core/ui/ui-example.md)), and a pending one settles nothing.
- Capture an existing screen unchanged before the first edit ([before captures](../../hive-core/ui/before-captures.md)), and send the first rendered view to the user before the build continues ([first view](../../flow-build/references/first-view.md)).
- Verification captures and reviews the screen as [UI verification](../../hive-core/ui/ui-verification.md) sets, audits accessibility as [accessibility audit](../../hive-core/ui/a11y-audit.md) sets and shows the result with the [walkthrough](../../hive-core/ui/walkthrough.md); the user's validation stop is part of the delivery mode in force.

**Handoffs.** Behavior with no screen change is a `feature`; a broken screen fixed at its cause is a `bugfix`.
