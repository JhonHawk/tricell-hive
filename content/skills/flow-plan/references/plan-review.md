# Review a saved plan

After the orchestrator saves an implementation plan, review it before declaring it ready. This bounded, read-only feedback step is part of planning; it is not an implementation code review, deployment gate or grant to execute the plan.

## Choose the useful coverage

Select reviewers by affected surfaces, risk and unresolved contracts, not a fixed file-count threshold. Use one reviewer for a small cohesive plan. Split independent domains for broader work and run them in parallel when supported. Skip absent domains: a backend-only plan needs no UI reviewer.

Use the `review-plan` role for each bounded assignment. Domain labels describe the brief, not separate agent implementations:

| Assigned domain | Questions to examine |
| --- | --- |
| Backend / data / integration | API and event contracts, producers/consumers, data ownership, authorization, migrations, compatibility and failure behavior; concrete tests and rollout dependencies. |
| Frontend / UI | Application/area, accepted shell and patterns, reuse and justified differences, states, accessibility, local interaction/visual verification and matching backend contracts. Read the applicable project UI guide and [UI planning](ui-planning.md). |
| Infrastructure / delivery | Environments, provider constraints, identities, secrets handling, deployment order, migrations, observability, recovery and the authorization boundaries of operational gates. |
| Other affected surface | Assign the actual responsibility, such as mobile lifecycle, firmware/hardware or concurrency/performance; supply its relevant source and project guidance rather than forcing it into a web/backend checklist. |

Include cross-domain interfaces in the briefs where needed. The orchestrator reconciles those findings; a separate generic reviewer is not mandatory on top of domain reviewers. More reviewers or repeated agreement do not establish stronger evidence by themselves.

## Give a bounded assignment

Supply the saved plan path and revision identifier (commit, content hash, or timestamped snapshot), objective, assigned surface, relevant source/project guidance, accessible skill/resource paths and expected feedback. Keep the candidate stable while that round runs. A reviewer must inspect relevant references, not inherit the parent's interpretation as fact.

Require read-only analysis and feedback to the parent. Do not delegate edits, test execution, hosted reviews, tracker writes or additional agents. Use native read-only controls where available; a prompt contract alone is not proof of technical isolation. If the role is unavailable but native subagents exist, supply its contract to a read-only bounded child. If delegation is unavailable, disclose the missing independent review and offer manual review or a later supported session; a labeled self-check is not an equivalent pass.

## Reconcile and close

Reviewers return material findings with plan location, source evidence, consequence and suggested correction or decision. Separate blocking gaps, nonblocking suggestions and access limits. They do not approve execution or claim runtime behavior was tested.

The orchestrator checks findings against evidence, resolves contradictions, and alone updates the plan. Record accepted fixes, justified dismissals and unresolved decisions in the existing plan's progress section. Ask the user only for consequential choices not already settled. A plan with a material unresolved gap is not ready.

Re-review materially changed criteria, interfaces or assumptions with the affected reviewer; do not repeat unchanged domains for cosmetic edits. Avoid unbounded loops: recurring uncertainty or expansion needs a concrete decision, not more identical reviews. End with the reviewed revision, coverage, outcome and remaining limits. This review does not authorize build, code review of implementation, publication or deployment.
