---
name: "hive-review-plan"
description: "Review a saved implementation plan as a whole and return evidence-backed gaps to the orchestrator without editing or executing it. Use after saving an implementation plan in a change folder."
model_profile: "reasoning"
access_profile: "observe"
effort: "medium"
---

# hive-review-plan

1. Read the supplied plan revision in full, and check its criteria and tasks against [Acceptance criteria](skill:flow-plan/SKILL.md#acceptance-criteria). When the plan's approach reads or changes an external system, also check it against [outside the repository](skill:flow-plan/references/outside-the-repository.md): each external fact carries its command or source, its result and its date, or is written as not measured with the task that meets it first; and each class of effect names its target, its ceiling, its tasks and its answer, and covers none of the cases that keep their own question. Inspect the relevant project conventions and source references; report inaccessible dependencies instead of assuming their contents.

2. Check feasibility, scope, contracts, dependencies, acceptance criteria and verification across the whole plan. Distinguish inspected facts, plan claims and assumptions. Focus on consequential gaps and contradictions; do not invent work for surfaces the plan does not touch or require a template section without a concrete need.

3. Return findings to the orchestrator each with a priority from P0 to P3 by observable impact, as [Findings](skill:flow-close/SKILL.md#findings) defines (a gap that would let a task fail its criterion or break a contract is P1), with plan section or task, supporting source, consequence and a bounded recommendation or unresolved decision. State what was inspected and what could not be verified. Explicitly report no material findings when appropriate; that does not authorize implementation or certify uninspected areas.

4. Do not edit the plan, application, configuration, tracker or memory; do not execute plan steps, tests, deployment, external review tools or mutating probes. Do not invoke flow-plan as an authoring workflow or spawn further reviewers. The orchestrator owns reconciliation and writes.

Stay within the parent’s assigned scope, write ownership, and output destination. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`, or in a separate worktree. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. Stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
