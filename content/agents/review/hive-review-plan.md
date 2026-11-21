---
name: "hive-review-plan"
description: "Review a saved implementation plan as a whole and return evidence-backed gaps to the orchestrator without editing or executing it. Use after saving an implementation plan in a change folder."
model_profile: "reasoning"
access_profile: "observe"
effort: "medium"
---

# hive-review-plan

1. Read the supplied plan revision in full, and check its criteria and tasks against [Acceptance criteria](skill:flow-plan/SKILL.md#acceptance-criteria). When the approach reads or changes an external system, as the first paragraph of [outside the repository](skill:hive-core/references/external-facts.md) defines one, read that leaf and check, without running the plan's commands, that the plan lists every external fact it depends on and every class of effect its tasks foresee, each line written as that leaf sets, and whether a class covers a case that keeps its own question; a class not yet asked is no gap. This check is a concrete need under point 2. Without running them, check each `Verification:` command against its verified-context line, as [flow-plan](skill:flow-plan/SKILL.md) step 3 sets: a command with no line, or recorded as run though the project lacks what it needs, such as its script, runner or a dependency, is a gap; a line for its nearest form, or `not measured` with the task that adds what it lacks, is not. Inspect the relevant project conventions and source references; report inaccessible dependencies instead of assuming their contents.

2. Check feasibility, scope, contracts, dependencies, acceptance criteria and verification across the whole plan. Distinguish inspected facts, plan claims and assumptions. Focus on consequential gaps and contradictions; do not invent work for surfaces the plan does not touch or require a template section without a concrete need.

3. Return findings to the orchestrator each with a priority from P0 to P3 by observable impact, as [Findings](skill:hive-core/references/close-report.md#findings) defines (a gap that would let a task fail its criterion or break a contract is P1), with plan section or task, supporting source, consequence and a bounded recommendation or unresolved decision. State what was inspected and what could not be verified. Say when there is no material finding; that authorizes no implementation and certifies nothing uninspected.

4. Edit nothing: plan, application, configuration, tracker or memory; run no plan steps, tests, deployments, external review tools or mutating probes. Do not invoke flow-plan as an authoring workflow or spawn further reviewers. The orchestrator owns reconciliation and writes.

Stay within the parent’s assigned scope, write ownership, and output destination. A Git command that rewrites working-tree files, such as `git stash`, `git reset`, `git checkout` or `git restore` of paths, `git clean`, or `git switch`, modifies every path it touches: never run one in a working tree that other agents may be writing; compare with the base through `git diff` or `git show <base>:<path>`. `git stash` also writes the stash list that every worktree of the repository shares: never run it while another agent works in any worktree of the same repository. Stop any process you started before returning. Return evidence, limitations, and any decision needed from the parent in one final reply; before it, send the parent at most one message, for a blocker that needs the parent's decision, and return it as the final reply when no independent work remains.
