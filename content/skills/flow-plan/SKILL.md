---
name: flow-plan
description: Prepare an implementation plan by resolving consequential decisions, defining affected contracts, and decomposing work into verifiable tasks. Use when planning is requested or needed before a multi-step change; small understood edits need no formal plan.
---

# Prepare an executable plan

Produce enough shared understanding for an implementer to act without inventing product, scope, compatibility, or architectural decisions. Use the CLI's available planning capabilities and respect its active mode. A native task list is a working view, not proof of a complete or portable plan.

## Ground the decisions

Establish the intended outcome, scope, constraints, authorized effects, and observable acceptance criteria from the conversation. Inspect the relevant implementation, callers, tests, configuration, and project guidance. Reuse sound research but recheck facts that may have changed.

Resolve missing consequential choices with the user, offering meaningful alternatives and their tradeoffs. Investigate discoverable facts directly. Treat information that clarifies the requested outcome as an update to the plan; when it proposes a material expansion beyond that outcome, resolve the scope before planning dependent work. If a technical uncertainty can invalidate the approach, resolve it with authorized investigation or mark the affected work blocked; do not hide it in an implementation task. Routine local coding choices need not be frozen in advance.

Before proposing a new mechanism or abstraction, inspect relevant existing solutions and patterns, including those in other projects only when their destination requirements are compatible. Define only the contracts affected by the change: inputs, outputs, error behavior, data or compatibility rules, and migration or recovery requirements where applicable. Record why the selected approach fits the evidence and constraints. Mention meaningful rejected alternatives only when they explain a decision.

When the plan chooses infrastructure resource names, read [the infrastructure naming reference](references/infra-naming.md). Apply project and provider constraints before its defaults, and record any migration impact for an existing name.

## Make the work verifiable

Group tasks into coherent results that can be checked independently. Prefer usable end-to-end increments when the change crosses layers. Include dependencies, concrete source locations, the expected behavior, and the verification command or observation with its expected result. Look up actual project commands rather than inventing them.

Use checkbox tasks. Keep setup and documentation with the result they support; split tasks when dependencies, review, or delivery boundaries justify it. Do not require full implementation code in the plan or replace a behavioral check with a heading such as "add appropriate tests".

Before calling the plan ready, check:

- Important claims are grounded in inspected sources; facts and assumptions are distinguishable.
- Decisions that could change the outcome, contracts, or compatibility are resolved; remaining uncertainties are bounded.
- Every acceptance criterion has an implementing task and a concrete verification.
- Task interfaces and dependencies agree; no task relies on an undefined contract.
- An implementer with this plan and its linked sources can identify what to change, what to preserve, and how to verify it.

For consequential or delegated work, use an available authorized fresh-context review when it can expose missing decisions. Reviewers should identify gaps, not simply certify that sections exist. Fix material gaps or report why the plan remains incomplete. Readiness does not grant execution permission.

## Preserve a useful handoff

Keep a bounded, understood plan in the conversation unless retention helps. Retain a plan when requested, when work must resume across sessions or agents, when deliveries need coordination, or when consequential decisions or expensive investigation need a durable record. Before writing a retained artifact, resolve its destination through the applicable global artifact-placement instructions and existing work identifier.

When creating or reviewing a retained plan, read [the plan format and adaptable template](references/plan-format.md). It defines the handoff document, not a mandatory document set. Keep the readiness procedure above here; use the reference to make its results resumable.

If the active CLI mode prohibits writing the artifact, present its contents and intended destination; do not bypass the mode. For a planning-only request, stop with the plan or unresolved decisions. When implementation is already authorized and the active mode permits it, follow that authorization without introducing another approval requirement.
