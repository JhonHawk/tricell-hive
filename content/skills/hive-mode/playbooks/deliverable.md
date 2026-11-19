# Playbook: deliverable

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the product the user asks for is a document, a report, a requirement document or a design, and no code or configuration change is the point. The deliverable picks the skill that makes it.

**Stages.** `hive-mode` settles what the deliverable is, for whom and why, then continues in the same turn into the skill for its kind:
- A requirement document (PRD, functional requirements, business rules of a module or feature, or the validation of one someone else wrote): [create-prd](../../create-prd/SKILL.md), which reads the source and asks its own questions.
- A design in Figma, such as replicating a screen there: [flow-figma-design](../../flow-figma-design/SKILL.md), after the connection check of [design stage](../../hive-rule-design-stage/SKILL.md); without a connection, report it as the blocker. The stage does not report the design done before the user approves it.
- A visual HTML report, a printable paper, a comparison, an explainer, a deck, or a documentation site: [create-report](../../create-report/SKILL.md).
- A ticket or a breakdown of work into tickets: [create-issue](../../create-issue/SKILL.md).

**Specific to a deliverable**
- The deliverable is the result: no `flow-plan`, `flow-build` or `flow-verify` runs unless the user also asks for a repository change; that part is a separate piece of work with the playbook of its kind.
- Put a support artifact where the project's layout says, and read [layout and retention](../../project-layout/references/layout.md) before writing it. Findings stay in the conversation unless the user asks for a saved file.
- A business rule the deliverable decides is recorded as [business rules](../../hive-rule-business-rules/SKILL.md) sets.
- Publishing or sharing the deliverable is the user's choice; this playbook authorizes none.

**Handoffs.** A requirement document that is approved and should become work goes to `flow-plan` or the playbook of its kind; the user's reply authorizes only the flow it names ([scope changes](../../hive-rule-scope-changes/SKILL.md)).

**Leaves.** [design stage](../../hive-rule-design-stage/SKILL.md), [business rules](../../hive-rule-business-rules/SKILL.md), [scope changes](../../hive-rule-scope-changes/SKILL.md).
