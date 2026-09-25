---
name: flow-plan
description: Prepare an implementation plan by resolving consequential decisions, defining affected contracts, and decomposing work into verifiable tasks. Use when planning is requested or needed before a multi-step change; small understood edits need no formal plan.
---

# Prepare an executable plan

Produce enough shared understanding for an implementer to act without inventing product, scope, compatibility, or architectural decisions. Use the CLI's available planning capabilities and respect its active mode. A native task list is a working view, not proof of a complete or portable plan.

## Ground the decisions

Establish the intended outcome, scope, constraints, authorized effects, and observable acceptance criteria from the conversation. Inspect the relevant implementation, callers, tests, configuration, and project guidance. Reuse sound research but recheck facts that may have changed. While delegated exploration runs, draft the parts of the plan that settled decisions already determine, such as scope, delivery, and tasks whose approach is known, instead of waiting idle.

Resolve missing consequential choices with the user, offering meaningful alternatives and their tradeoffs. Investigate discoverable facts directly. Treat information that clarifies the requested outcome as an update to the plan; when it proposes a material expansion beyond that outcome, resolve the scope before planning dependent work. If a technical uncertainty can invalidate the approach, resolve it with authorized investigation or mark the affected work blocked; do not hide it in an implementation task. Routine local coding choices need not be frozen in advance.

Ask a consequential design question before writing dependent tasks; ask independent design questions together in one exchange, since each extra round adds a round-trip without a new decision. Ask one separately only when its answer determines which other questions apply. At the planning handoff, read [delivery decisions](references/delivery-decisions.md) to resolve the Git mode and applicable code-review mechanism. Work implemented without a plan resolves it at the start of `flow-build` instead. If no design questions remain, briefly state which existing decisions or conventions settled them.

Before proposing a new mechanism or abstraction, inspect relevant existing solutions and patterns, including those in other projects only when their destination requirements are compatible. Define only the contracts affected by the change: inputs, outputs, error behavior, data or compatibility rules, and migration or recovery requirements where applicable. Record why the selected approach fits the evidence and constraints. Mention meaningful rejected alternatives only when they explain a decision.

When an interface changes, identify its actual producers and consumers, the versions that may coexist, and the compatible delivery order. Use the project’s existing contract and distribution mechanism; a separate specs repository, published package, or prerelease channel is not required by this workflow. Define evidence for the affected integration and any partial rollout.

For changes to untrusted inputs, permissions, process execution, public data exposure, or dependencies, read [security boundaries](../flow-build/references/security-boundaries.md). For stored schema, representation, migration, or backfill changes, read [persistent data changes](../flow-build/references/data-changes.md). Apply only the relevant criteria and pass the reference paths to delegated implementers or reviewers.

When replacing or migrating an existing system, state which legacy paths are removed, retained temporarily for recovery, or remain operational alongside the replacement. Name the retirement condition for temporary retention and the maintenance and verification cost of coexistence. Reuse decisions already settled by the request; resolve only consequential ambiguity before dependent implementation. Do not introduce parallel implementations or fallback layers merely as an assumed precaution.

For new or changed screens, shells, layouts, or reusable UI patterns, read [UI planning](references/ui-planning.md). Resolve the affected application and pattern, applicable project conventions, reuse, and intentional differences before calling the visual work ready.

When the plan chooses infrastructure resource names, read [the infrastructure naming reference](references/infra-naming.md). Apply project and provider constraints before its defaults, and record any migration impact for an existing name.

## Make the work verifiable

Group tasks into coherent results that can be checked independently. Prefer usable end-to-end increments when the change crosses layers. Include dependencies, concrete source locations, the expected behavior, and the verification command or observation with its expected result. Look up actual project commands rather than inventing them.

Select the verification approach and timing for each coherent result, including prerequisites and observable pass criteria. For integration, runtime, UI, performance, hardware, test replacement/removal, promotion, deployment, or costly checks, read [verification by risk and surface](../flow-build/references/verification.md). For substantial gates, identify scope/candidate, mechanism, environment, relevant effects and expected evidence; record whether execution authorization is pending or already sufficient under the shared rule. Include applicable local in-vivo/UI verification in build completion and offer deployed-environment walks separately. Use the project's commands, environments, and delivery gates; name unavailable checks and where regression coverage will run. Keep this in the existing task or conversation rather than creating a separate verification document.

Use checkbox tasks. Keep setup and documentation with the result they support; split tasks when dependencies, review, or delivery boundaries justify it. Do not require full implementation code in the plan or replace a behavioral check with a placeholder that names no command, observation, or expected result, such as "add appropriate tests" or "verify it works".

Before calling the plan ready, check:

- Important claims are grounded in inspected sources; facts and assumptions are distinguishable.
- Decisions that could change the outcome, contracts, or compatibility are resolved; remaining uncertainties are bounded.
- Every acceptance criterion has an implementing task and a concrete verification.
- Task interfaces and dependencies agree; no task relies on an undefined contract.
- Each task states its execution, consistent with its dependencies and write ownership.
- A change with a rendered UI effect plans the required independent `review-ux` and `sdd-verify` children from the verification reference: the base revision for comparison, viewports and themes, isolated data and browser sessions, and the browser tool selected by its [browser automation](../flow-build/references/browser-automation.md) rules. The implementer's walk does not replace them.
- An implementer with this plan and its linked sources can identify what to change, what to preserve, and how to verify it.

Assign each task's execution: the main thread or a delegated child, with the reason. Prefer delegation for units with settled interfaces and no overlapping writes; keep a unit in the main thread only when it is coupled to work in progress or its brief would be longer than the change itself. For delegated work, identify separable results and their interfaces, dependencies, and ownership, without assigning a permanent specialist for every framework. Link the project conventions and technical references each unit needs, without freezing runtime-specific agent names or repeating the shared delegation contract.

After saving an implementation plan, read [plan review](references/plan-review.md) and dispatch bounded read-only subagents before declaring it ready. Dispatch one reviewer per affected domain in parallel, even for a small plan; use a single reviewer only when one domain is affected, and omit unaffected areas, including UI when absent. Reviewers return feedback to the orchestrator, who validates findings and alone revises the plan. Disclose unavailable delegation and unresolved coverage.

## Preserve a useful handoff

Keep a bounded, understood plan in the conversation unless retention helps. Retain a plan when requested, when work must resume across sessions or agents, when deliveries need coordination, or when consequential decisions or expensive investigation need a durable record. Before writing a retained artifact, resolve its destination through the applicable global artifact-placement instructions and existing work identifier.

When creating or reviewing a retained plan, read [the plan format and adaptable template](references/plan-format.md) and [change records](references/change-records.md). They define the handoff in the change folder, not a mandatory document set. Keep the readiness procedure above here; use the reference to make its results resumable. Preserve an established `<topic>-plan.md` or `<topic>.plan.md` path without renaming it merely to normalize the convention.

If the active CLI mode prohibits writing the artifact, present its contents and intended destination; do not bypass the mode. For a planning-only request, stop with the plan or unresolved decisions. When implementation is already authorized and the active mode permits it, follow that authorization without introducing another approval requirement.

When returning a ready plan without implementing, state the concrete next step: `flow-build` with the change folder or canonical plan path (or direct execution for a small change). State the first execution scope, the human-validation stop, the agreed endpoint after validation and any excluded promotion. A saved change folder that is not yet versioned is a pending item: deliver it under the specs repository's Git rules when that is authorized, and otherwise report it as pending. Avoid a vague "when you want to build" close or leaving PR versus merge implicit. This recommendation does not execute or authorize the work; when execution is already explicitly authorized, continue as permitted instead of requiring another invocation.
