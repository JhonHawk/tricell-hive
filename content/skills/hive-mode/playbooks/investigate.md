# Playbook: investigate

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work is a question, a diagnosis of a failure the user reports without asking for the fix, a comparison, an assessment of a ticket, or an audit of a codebase for what to improve.

**Stages.** `hive-mode` alone: steps 1 to 5 of the [router](../SKILL.md) are the procedure, and the work ends at its findings. No stage after it runs unless the user asks for a change; a confirmed defect or a requested change is a handoff.

**Specific to an investigation**
- Read and investigate as steps 1 to 4 of the [router](../SKILL.md) set, children and the evidence of a claim included.
- An audit of a codebase for what to improve follows [code audit](../references/code-audit.md): its aspects, the brief of its children, its report and the rejected findings it skips.
- A chosen ticket is assessed for impact, complexity, readiness and dependencies as the router's «Assess a chosen ticket» sets.
- The investigation changes no repository file, tracker or shared data. A ticket for a finding is a single action offered at the end, never part of a route ([create-issue](../../create-issue/SKILL.md) writes it).

**Reporting.** The findings end in the analysis row of [report endings](../../hive-core/messages/report-endings.md); if the work later closes, its final report uses the row of [close report](../../hive-core/close/close-report.md#variants) that the delivery mode in force picks.

**Handoffs.** When the findings confirm a defect and the user asks for its fix, hand off to [bugfix](bugfix.md) as «Changing playbook» of the [router](../SKILL.md#pick-the-playbook) sets. Other confirmed work goes to the playbook of its kind, or to `flow-plan` when its how has a decision for the user.

**Leaves.** [delegation](../../hive-core/delegation/delegation.md), [child brief](../../hive-core/delegation/child-brief.md), [report endings](../../hive-core/messages/report-endings.md), [scope changes](../../hive-core/plan/scope-changes.md), [external facts](../../hive-core/plan/external-facts.md) when the answer depends on an outside system.
