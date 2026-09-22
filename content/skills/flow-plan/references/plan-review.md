# Review a saved plan

After the orchestrator saves an implementation plan, review it before declaring it ready. This bounded, read-only feedback step is part of planning; it is not an implementation code review, deployment gate or grant to execute the plan.

## Choose the useful coverage

Select reviewers by affected domain, not plan size or a file-count threshold. Assign one reviewer to each affected domain in the table below and run them in parallel when supported, even when the plan is small or its domains are coupled; give each the shared contracts it depends on. Use a single reviewer only when one domain is affected. Skip absent domains: a backend-only plan needs no UI reviewer.

Resolve and read the canonical `review-plan` contract through the native role catalog or an explicit installed/source path before dispatch. Use the `review-plan` role for each bounded assignment under the shared native-selection and fallback contract. Domain labels describe the brief, not separate agent implementations:

| Assigned domain | Questions to examine |
| --- | --- |
| Backend / data / integration | API and event contracts, producers/consumers, data ownership, authorization, migrations, compatibility and failure behavior; concrete tests and rollout dependencies. |
| Frontend / UI | Application/area, accepted shell and patterns, reuse and justified differences, states, accessibility, local interaction/visual verification including the required independent UI review and in-vivo children, and matching backend contracts. Read the applicable project UI guide and [UI planning](ui-planning.md). |
| Infrastructure / delivery | Environments, provider constraints, identities, secrets handling, deployment order, migrations, observability, recovery and the authorization boundaries of operational gates. |
| Other affected surface | Assign the actual responsibility, such as mobile lifecycle, firmware/hardware or concurrency/performance; supply its relevant source and project guidance rather than forcing it into a web/backend checklist. |

Include cross-domain interfaces in the briefs where needed. The orchestrator reconciles those findings; a separate generic reviewer is not mandatory on top of domain reviewers. More reviewers or repeated agreement do not establish stronger evidence by themselves.

## Give a bounded assignment

Supply the saved plan path and revision identifier (commit, content hash, or timestamped snapshot), objective, assigned surface, relevant source/project guidance, accessible skill/resource paths and expected feedback. Keep the candidate stable while that round runs. A reviewer must inspect relevant references, not inherit the parent's interpretation as fact. Frame uncertain technical claims as questions to verify against sources, not facts to confirm. Separate explicit user decisions from technical assumptions; provide the source and uncertainty rather than an expected verdict. For example, ask whether the installed migration runner wraps the file in a transaction and what partial failure leaves behind, instead of asserting that it does.

Require read-only analysis and feedback to the parent. Do not delegate edits, test execution, hosted reviews, tracker writes or additional agents. Use native read-only controls where available; a prompt contract alone is not proof of technical isolation. Apply the shared dispatch evidence requirements in the existing plan review record, alongside the candidate revision and domain. Require the child to read assigned resources. If delegation is unavailable, disclose the missing independent review and offer manual review or a later supported session; a labeled self-check is not an equivalent pass.

## Reconcile and close

Reviewers return material findings with plan location, source evidence, consequence and suggested correction or decision. Separate blocking gaps, nonblocking suggestions and access limits. They do not approve execution or claim runtime behavior was tested.

The orchestrator checks findings against evidence, resolves contradictions, and alone updates the plan. Record accepted fixes, justified dismissals and unresolved decisions in the existing plan's progress section. Ask the user only for consequential choices not already settled. A plan with a material unresolved gap is not ready.

Re-review materially changed criteria, interfaces or assumptions with the affected reviewer, including corrections introduced in response to a review. Supply the revised candidate and changed scope; check the correction and its consequences rather than asking only for confirmation that the requested wording was added. Do not repeat unchanged domains for cosmetic edits. Before declaring the plan ready, compare its final content with the reviewed revisions and account for every material delta in the existing progress record: affected domain, reviewed candidate, outcome and remaining limit. A material correction not yet reviewed leaves that coverage pending; an earlier verdict does not cover it. Cosmetic edits may be reconciled by the orchestrator without another round. No new findings means none found within the stated coverage, not proof that the entire plan is defect-free. Avoid unbounded loops: recurring uncertainty or expansion needs a concrete decision, not more identical reviews. End with the reviewed revision, coverage, outcome and remaining limits. This review does not authorize build, code review of implementation, publication or deployment.
