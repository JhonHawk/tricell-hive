---
name: flow-plan
description: Prepare an implementation plan by resolving consequential decisions, defining affected contracts, and decomposing work into verifiable tasks. Use when planning is requested or needed before a multi-step change; small understood edits need no formal plan.
---

# Prepare an executable plan

Produce enough shared understanding for an implementer to act without inventing product, scope, compatibility, or architectural decisions. Use the CLI's available planning capabilities and respect its active mode. A native task list is a working view, not proof of a complete or portable plan.

## Entry and exit

Enter with a goal that has a scope; when it does, enter directly, with no research before it. Exit with a ready plan, in a change folder when retained, and the delivery decisions resolved or stated, for [flow-build](../flow-build/SKILL.md).

## Ground the decisions

Before resolving needed project settings, read [project settings](references/project-settings.md). Before formulating a user decision, read [questions](references/questions.md). Before delegating or waiting for external work, read [delegation and waits](../flow-research/references/delegation.md).

Establish the intended outcome, scope, constraints, authorized effects, and observable acceptance criteria from the conversation. Inspect the relevant implementation, callers, tests, configuration, and project guidance, delegating exploration under the delegation criterion of [flow-research](../flow-research/SKILL.md). Reuse sound research but recheck facts that may have changed. While delegated exploration runs, draft the parts of the plan that settled decisions already determine, such as scope, delivery, and tasks whose approach is known, instead of waiting idle.

Resolve missing consequential choices with the user, offering meaningful alternatives and their tradeoffs. Investigate discoverable facts directly. Handle open topics under the open-topics rule in the global guidance (Communication), and record each one's destination in the plan. If a technical uncertainty can invalidate the approach, resolve it with authorized investigation or mark the affected work blocked; do not hide it in an implementation task. Routine local coding choices need not be frozen in advance. A nearby fix proposed as an addition to the plan, or a task added during the build, that creates a new interaction surface, such as a focus stop, a control or a route, carries its own acceptance criterion and enters the UI checks that the UI verification value selects, or, when that value is none or not yet resolved, the implementer's rendered check that the UI gate sets, or stays out of the plan.

When a decision changes the design of the plan or of the research it rests on, present it as soon as it appears, without waiting for the rest of the plan: give the explanation, ask it by following [questions](references/questions.md), which sets [the pause line](references/questions.md#pause-and-handle-the-answer) that ends the message, and stop the work that depends on it. When an answer returns the work to research or invalidates tasks already written, say so as [questions](references/questions.md) requires for an answer and update the plan as [Revise a plan in progress](references/plan-revision.md) sets. Ask a consequential design question before writing dependent tasks; group the others into rounds as [questions](references/questions.md) sets, asking in a later turn a question whose options depend on another's answer, as [questions](references/questions.md#decide-whether-a-question-is-needed) sets. At the planning handoff, read [delivery decisions](references/delivery-decisions.md) to resolve the Git mode, [review decisions](references/review-decisions.md) for the applicable code-review mechanism and, for a rendered UI effect, [the UI verification decision](references/ui-verification-decision.md) for the `UI verification` level: stated in one line when an earlier answer in the session settles them, otherwise asked in the same round as the last design questions that do not depend on them. Work implemented without a plan resolves it at the start of `flow-build` instead. If no design questions remain, briefly state which existing decisions or conventions settled them.

Before proposing a new mechanism or abstraction, inspect relevant existing solutions and patterns, including those in other projects only when their destination requirements are compatible. Define only the contracts affected by the change: inputs, outputs, error behavior, data or compatibility rules, and migration or recovery requirements where applicable. Record why the selected approach fits the evidence and constraints. Mention meaningful rejected alternatives only when they explain a decision.

When a change alters an interface that other code, services, or data depend on, read [interface changes](references/interface-changes.md) before planning work that depends on it.

For changes to untrusted inputs, permissions, process execution, public data exposure, or dependencies, read [security boundaries](../flow-build/references/security-boundaries.md); a security review the user asks for runs as [flow-security](../flow-security/SKILL.md) sets. For stored schema, representation, migration, or backfill changes, read [persistent data changes](../flow-build/references/data-changes.md). Before planning a change that removes, closes, disables, migrates, or renames something that runs in production, moves it to another host or process, or changes untested production behavior, read [change impact](../flow-research/references/change-impact.md). Apply only the relevant criteria and pass the reference paths to delegated implementers or reviewers.

When replacing or migrating an existing system, state which legacy paths are removed, retained temporarily for recovery, or remain operational alongside the replacement. Name the retirement condition for temporary retention and the maintenance and verification cost of coexistence. Reuse decisions already settled by the request; resolve only consequential ambiguity before dependent implementation. Do not introduce parallel implementations or fallback layers merely as an assumed precaution. When the old path can still run on the same inputs, include in the plan's verification the old-against-new comparison that the "Replacement or migration" row of [verification](../flow-build/references/verification.md) sets; otherwise the plan records why it cannot run. Do not keep a legacy path alive only for the comparison.

For new or changed screens, shells, layouts, reusable UI patterns, or a dialog or form where the end user enters or chooses data, read [UI planning](references/ui-planning.md). Resolve the affected application and pattern, applicable project conventions, reuse, and intentional differences before calling the visual work ready.

When the plan chooses infrastructure resource names, or creates resources whose names existing infrastructure code generates, read [the infrastructure naming reference](references/infra-naming.md). Apply project and provider constraints before its defaults, and record any migration impact for an existing name.

## Make the work verifiable

Group tasks into coherent results that can be checked independently. Prefer usable end-to-end increments when the change crosses layers. Include dependencies, concrete source locations, the expected behavior, and the verification command or observation with its expected result. Look up actual project commands rather than inventing them.

Select the verification approach and timing for each coherent result, including prerequisites and observable pass criteria. For integration, runtime, UI, performance, hardware, test replacement/removal, or costly checks, read [verification by risk and surface](../flow-build/references/verification.md); for promotion or deployment, read [deployment checks](../flow-deploy/references/deployment-checks.md). For substantial gates, identify scope/candidate, mechanism, environment, relevant effects and expected evidence; record whether execution authorization is pending or already sufficient under the shared rule, and ask the pending ones in the delivery and review round, as [verification](../flow-build/references/verification.md#gate-selection-and-timing) sets. Plan the local in-vivo and UI checks that [flow-verify](../flow-verify/SKILL.md) runs on the built candidate, and offer deployed-environment walks separately. Use the project's commands, environments, and delivery gates; name unavailable checks and where regression coverage will run. Keep this in the existing task or conversation rather than creating a separate verification document. Give each task that changes behavior a test approach under [the TDD criterion](../flow-build/references/tdd.md), which you read when you choose a task's test approach: `tdd`, `characterization` for behavior-preserving refactors, or `check` for a change outside the TDD criterion, naming the check that verifies it.

Use checkbox tasks. Keep setup and documentation with the result they support; split tasks when dependencies, review, or delivery boundaries justify it. Do not require full implementation code in the plan or replace a behavioral check with a placeholder that names no command, observation, or expected result, such as "add appropriate tests" or "verify it works".

Before calling the plan ready, check:

- Important claims are grounded in inspected sources; facts and assumptions are distinguishable.
- Decisions that could change the outcome, contracts, or compatibility are resolved; remaining uncertainties are bounded.
- Every acceptance criterion, task marker, dependency and checkpoint meets the rules in [plan criteria and markers](references/plan-criteria-and-markers.md#acceptance-criteria-and-task-markers); a break of any of them blocks readiness.
- A change to stored data records the affected-record count per environment, or why one could not be taken, and its technique and recovery match that count, as [persistent data changes](../flow-build/references/data-changes.md) describes.
- Task interfaces and dependencies agree; no task relies on an undefined contract.
- When the plan lists work as **Not yet specified**, it follows that rule in [the plan format](references/plan-format.md#what-the-document-must-carry).
- Each task states its execution, consistent with its dependencies and write ownership.
- A change with a rendered UI effect plans the independent `hive-review-ux` and `hive-verify-change` children that the `UI verification` value selects, as [the UI gate](../flow-verify/references/ui-gate.md) sets: the candidate identity, the base revision for comparison, viewports and themes, isolated data and browser sessions, and the browser tool selected by its [browser automation](../flow-build/references/browser-automation.md) rules. The implementer's walk does not replace them.
- An implementer with this plan and its linked sources can identify what to change, what to preserve, and how to verify it.

Assign each task's execution: the main thread or a delegated child, with the reason. Choose by the criterion in [delegation and waits](../flow-research/references/delegation.md#split-and-announce-work). For delegated work, identify separable results and their interfaces, dependencies, and ownership, without assigning a permanent specialist for every framework. Link the project conventions and technical references each unit needs, without freezing runtime-specific agent names or repeating the shared delegation contract.

After saving an implementation plan, read [plan review](references/plan-review.md) and dispatch bounded read-only subagents before declaring it ready. Dispatch one reviewer per affected domain in parallel, even for a small plan; use a single reviewer only when one domain is affected, and omit unaffected areas, including UI when absent. Reviewers return feedback to the orchestrator, who validates findings and alone revises the plan. Disclose unavailable delegation and unresolved coverage.

## Revise a plan in progress

When a retained plan changes after it was written, during planning or during the build, read and follow [plan revision](references/plan-revision.md) before editing the change folder.

## Preserve a useful handoff

Before writing a support artifact, read [artifact layout and retention](../workspace-conventions/references/artifact-layout.md). Before a plan-handoff report or any message containing citations, links or externally defined IDs, read [reporting](../flow-close/references/reporting.md).

Keep a bounded, understood plan in the conversation unless retention helps; plan review applies to a saved plan, and a plan kept in the conversation meets the readiness check above without reviewers. Retain a plan when requested, when work must resume across sessions or agents, when deliveries need coordination, or when consequential decisions or expensive investigation need a durable record. Before writing a retained artifact, resolve its destination through the applicable global artifact-placement instructions and existing work identifier.

When creating or reviewing a retained plan, read [the plan format and adaptable template](references/plan-format.md) and [change records](references/change-records.md). They define the handoff in the change folder, not a mandatory document set. Keep the readiness procedure above here; use the reference to make its results resumable. Preserve an established `<topic>-plan.md` or `<topic>.plan.md` path without renaming it merely to normalize the convention.

If the active CLI mode prohibits writing the artifact, present its contents and intended destination; do not bypass the mode. For a planning-only request, stop with the plan or unresolved decisions. When implementation is already authorized and the active mode permits it, follow that authorization without introducing another approval requirement.

When returning a ready plan without implementing, follow [Hand off a ready plan](references/plan-handoff.md#hand-off-a-ready-plan).
