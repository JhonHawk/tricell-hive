---
name: "hive-review-plan"
description: "Review a saved implementation plan as a whole and return evidence-backed gaps to the orchestrator without editing or executing it. Use after saving an implementation plan in a plan file."
model_profile: "reasoning"
access_profile: "observe"
effort: "medium"
---

# hive-review-plan

1. Read the supplied plan revision in full, and check its criteria and tasks against [Acceptance criteria](skill:flow-plan/SKILL.md#acceptance-criteria). When the approach reads or changes an external system, as the first paragraph of [outside the repository](skill:hive-core/plan/external-facts.md) defines one, read that leaf and check, without running the plan's commands, that the plan lists every external fact it depends on and every class of effect its tasks foresee, each line written as that leaf sets, and whether a class covers a case that keeps its own question; a class not yet asked is no gap. This check is a concrete need under point 2. Check each `Verification:` command against its verified-context line without running it, as [flow-plan](skill:flow-plan/SKILL.md) step 3 sets: a command with no line, or recorded as run though the project lacks what it needs (its script, runner or a dependency), is a gap; a line for its nearest form, or `not measured` with the task that adds what it lacks, is not. Inspect the relevant project conventions and source references; report inaccessible dependencies instead of assuming their contents.

2. Check feasibility, scope, contracts, dependencies, acceptance criteria and verification across the whole plan, distinguishing inspected facts, plan claims and assumptions. Report consequential gaps and contradictions; invent no work for surfaces the plan does not touch and require no template section without a concrete need.

3. Return findings to the orchestrator each with a priority from P0 to P3 by observable impact, as [Findings](skill:hive-core/close/close-report.md#findings) defines (a gap that would let a task fail its criterion or break a contract is P1), with the plan section or task, supporting source, consequence and a bounded recommendation or unresolved decision. State what was inspected and what could not be verified. Say when there is no material finding; that authorizes no implementation and certifies nothing uninspected.

4. Edit nothing (plan, application, configuration, tracker, memory); run no plan steps, tests, deployments, external review tools or mutating probes. Do not invoke flow-plan as an authoring workflow or spawn further reviewers. The orchestrator owns reconciliation and writes.

Stay within the parent’s assigned scope, write ownership, and output destination. Never run `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean` or `git switch` in a working tree other agents may be writing, and never `git stash` while another agent works in any worktree of the repository, since the stash list is shared. Stop any process you started before returning. Return evidence, limitations and needed decisions in one final reply; before it, send at most one message to the parent, for a blocker that needs its decision, and make that the final reply when no independent work remains.
