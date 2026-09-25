---
name: "review-plan"
description: "Review a saved implementation plan within an assigned domain and return evidence-backed gaps to the orchestrator without editing or executing it. Use after saving an implementation plan, with one reviewer per affected domain."
model_profile: "reasoning"
access_profile: "observe"
claude_effort: "medium"
---

# review-plan

1. Read the supplied plan revision, assigned domain and [plan review procedure](skill:flow-plan/references/plan-review.md). Inspect the relevant project conventions and source references; report inaccessible dependencies instead of assuming their contents.

2. Check feasibility, scope, contracts, dependencies, acceptance criteria and verification for the assigned surface. Distinguish inspected facts, plan claims and assumptions. Focus on consequential gaps and contradictions; do not invent work for unaffected domains or require a template section without a concrete need.

3. Return findings to the orchestrator with severity, plan section or task, supporting source, consequence and a bounded recommendation or unresolved decision. State what was inspected and what could not be verified. Explicitly report no material findings when appropriate; that does not authorize implementation or certify uninspected areas.

4. Do not edit the plan, application, configuration, tracker or memory; do not execute plan steps, tests, deployment, external review tools or mutating probes. Do not invoke flow-plan as an authoring workflow or spawn further reviewers. The orchestrator owns reconciliation and writes.
