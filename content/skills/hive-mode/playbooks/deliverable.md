# Playbook: deliverable

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the product the user asks for is a document, a report, a requirement document or a design, and no code or configuration change is the point. The deliverable picks the skill that makes it.

**Stages.** `hive-mode` settles what the deliverable is, for whom and why, then continues in the same turn into the skill for its kind:
- A requirement document (PRD, functional requirements, business rules of a module or feature, or the validation of one someone else wrote): [create-prd](../../create-prd/SKILL.md), which reads the source and asks its own questions.
- A design in Figma, such as replicating a screen there: [create-design](../../create-design/SKILL.md), after the connection check of [design stage](../../hive-core/ui/design-stage.md); without a connection, report it as the blocker. The stage does not report the design done before the user approves it.
- A review of a design someone already drew, given as a Figma link or an image: the [review entry](../../create-design/references/review.md) of `create-design`, which draws nothing. An image skips the connection check; a link needs a connection that reads Figma.
- A visual HTML report, a printable paper, a comparison, an explainer, a deck, or a documentation site: [create-report](../../create-report/SKILL.md).
- A ticket or a breakdown of work into tickets: [create-issue](../../create-issue/SKILL.md).

**Specific to a deliverable**
- The deliverable is the result: no `flow-plan`, `flow-build` or `flow-verify` runs unless the user also asks for a repository change; that part is a separate piece of work with the playbook of its kind.
- Put a support artifact where the project's layout says, and read [layout and retention](../../project-layout/references/layout.md) before writing it. Findings stay in the conversation unless the user asks for a saved file.
- A business rule a requirement document decides is written in that document, as [create-prd](../../create-prd/SKILL.md) sets; any other deliverable's rule is recorded as [business rules](../../hive-core/business/business-rules.md) sets.
- Publishing or sharing the deliverable is the user's choice; this playbook authorizes none.

**Reporting.** When the work closes, `flow-close` writes its final report with the row of [close report](../../hive-core/close/close-report.md#variants) that the delivery mode in force picks, and the message ends in the close row of [report endings](../../hive-core/messages/report-endings.md) that [close modes](../../hive-core/close/close-modes.md) sets.

**Handoffs.** A requirement document that is approved and should become work goes to `flow-plan` or the playbook of its kind; the user's reply authorizes only the flow it names ([scope changes](../../hive-core/plan/scope-changes.md)).

**Leaves.** [design stage](../../hive-core/ui/design-stage.md), [business rules](../../hive-core/business/business-rules.md), [scope changes](../../hive-core/plan/scope-changes.md).
