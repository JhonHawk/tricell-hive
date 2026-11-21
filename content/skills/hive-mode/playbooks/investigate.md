# Playbook: investigate

Read from [hive-mode](../SKILL.md#pick-the-playbook) when the work is a question, a diagnosis of a failure the user reports without asking for the fix, a comparison, an assessment of a ticket, or an audit of a codebase for what to improve.

**Stages.** `hive-mode` alone: steps 1 to 5 of the [router](../SKILL.md) are the procedure, and the work ends at its findings. No stage after it runs unless the user asks for a change; a confirmed defect or a requested change is a handoff.

**Specific to an investigation**
- Read before asking, and send to a `hive-research` child each aspect that needs more than five large reads, briefed and awaited as [delegation](../../hive-core/references/delegation.md) and [child brief](../../hive-core/references/child-brief.md) set. Verify what a child returns.
- For a failure whose cause is not shown, list the plausible causes and check each one, in parallel when independent; report the ones ruled out with their evidence.
- Open the source of each claim that decides the answer; resolve conflicting sources by date, version, scope and direct observation, and mark what you observed apart from what a document claims and what you infer.
- An audit of a codebase for what to improve follows [code audit](../references/code-audit.md): its aspects, the brief of its children, its report and the rejected findings it skips.
- A chosen ticket is assessed for impact, complexity, readiness and dependencies as the router's «Assess a chosen ticket» sets.
- The investigation changes no repository file, tracker or shared data. A ticket for a finding is a single action offered at the end, never part of a route ([create-issue](../../create-issue/SKILL.md) writes it).
- The findings end as [report endings](../../hive-core/references/report-endings.md) sets, with no flow recommended that the user did not ask for.

**Reporting.** The findings end in the analysis row of [report endings](../../hive-core/references/report-endings.md); if the work later closes, its final report uses the row of [close report](../../hive-core/references/close-report.md#variants) that the delivery mode in force picks.

**Handoffs.** When the findings confirm a defect and the user asks for its fix, hand off to [bugfix](bugfix.md) as a scope change, state its class and destination in one line, and record the new playbook ([scope changes](../../hive-core/references/scope-changes.md)). Other confirmed work goes to the playbook of its kind, or to `flow-plan` when its how has a decision for the user.

**Leaves.** [delegation](../../hive-core/references/delegation.md), [child brief](../../hive-core/references/child-brief.md), [report endings](../../hive-core/references/report-endings.md), [scope changes](../../hive-core/references/scope-changes.md), [external facts](../../hive-core/references/external-facts.md) when the answer depends on an outside system.
