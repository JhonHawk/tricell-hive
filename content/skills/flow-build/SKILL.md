---
name: flow-build
description: Implement or resume authorized code, configuration, promotion, or deployment work; reconcile current state, verify results, and close with evidence. Use for execution rather than research-only or planning-only requests; a formal plan is not a prerequisite for a small understood change.
---

# Implement and verify authorized work

Use the requested outcome or supplied plan as the execution brief. This skill does not require a formal plan or a prior research skill invocation, and does not grant effects beyond the user's instructions.

## Reconcile before execution

Read the relevant project guidance, current source and tests, and the supplied plan and references. Inspect existing changes and completed work. Treat the record as a claim to reconcile with current evidence, not a script to replay blindly.

Identify the candidate, affected applications/services and environments, and the checks needed to close each surface. A supplied plan is optional; keep a small task’s criteria in the conversation. Identify the next unmet acceptance criterion and its dependencies. Preserve valid completed work. If the current state invalidates a decision that sets scope, a contract, an acceptance criterion, or delivery, investigate the mismatch and update the affected plan or ask for the missing decision before dependent edits. Resolve ordinary implementation details within scope; do not stop at every task boundary merely to ask again.

Identify the available test runner, actual project check/build commands, relevant baseline failures, and the evidence needed for acceptance. For integration, runtime, UI, performance, hardware, test replacement/removal, promotion, deployment, or costly verification, read [verification by risk and surface](references/verification.md) before execution to select checks, prerequisites, and their timing. Reconcile a supplied verification plan rather than silently weakening it.

When the work will end in a Git delivery and neither a plan nor an earlier answer settled it, read [delivery decisions](../flow-plan/references/delivery-decisions.md) and ask its delivery-mode and review question once, before the first edit. The selected mode then carries the later gates, including waiting for the CI its merge depends on, without asking again for each operation.

Before dependent changes, including deployment of existing code, read the applicable situational reference: [security boundaries](references/security-boundaries.md) for untrusted inputs, permissions, process execution, public data exposure, or dependencies; [persistent data changes](references/data-changes.md) for stored schema, representation, migration, or backfill changes; and [diagnosis](../flow-research/references/diagnosis.md) when a failure’s cause remains uncertain. These apply even without a retained plan or specialist assignment.

For new or changed screens, shells, layouts, or reusable UI patterns, read [UI planning](../flow-plan/references/ui-planning.md) and any existing applicable project UI guide before dependent edits. Reconcile the supplied view decisions; if no plan exists, resolve the same decisions in a bounded implementation brief rather than inventing a new composition.

Before installing a missing tool, check the project-local and available system installations and the required version. Prefer the project’s existing setup. Distinguish task-local dependency setup from system-wide installs or shell/global configuration changes; the latter require explicit authorization, which may already be present. Verify the actual OS and shell before using platform-specific commands rather than assuming them from the harness name.

## Execute in verifiable increments

Choose a coherent next result and implement it in the existing source layout. Look for an existing solution or pattern before adding a mechanism or abstraction, and justify a new one by the need it serves. Confirm that reused tools or patterns meet this project's requirements. Use the project's declared dependency manager and preserve its lockfile conventions. Select technical references from the actual project stack and affected behavior. Keep dependent edits ordered.

Delegation is internal work and needs no separate approval; prefer it. Announce the split before the first edit under the shared delegation rules, following a plan's execution assignments unless current state invalidates them. Delegate each unit whose interface is settled and whose writes do not overlap concurrent work; keep a unit in the main thread only when it is coupled to work in progress or its brief would be longer than the change itself. A change spanning several files with its tests, or a diagnosis likely to produce long output, goes to a child by default. A plan task's text already serves as most of its brief; the main thread keeps coordination, integration, and user decisions. Work outside the plan's tasks, such as fixes after review or human validation, gets its own execution decision, announced before its first edit.

For changes with a rendered UI effect, the independent UI review and in-vivo children in [verification](references/verification.md) are required, not proposed. Propose other independent verification or specialist review when the acceptance criteria or risk warrant it, and dispatch within the agreed gate authority. Independent surface checks may run in parallel; keep one owner for shared mutations and wait for the required deployed version before dependent checks. The implementer still verifies its own change.

Before adding a dependency, assess whether an existing project dependency, standard library, or native platform capability meets the requirement, including established design-system conventions. Choose based on compatibility, maintainability, and fit rather than a universal library preference. Explain a new dependency’s purpose and any overlap; consolidating existing usages remains separate unless authorized.

When choosing an infrastructure resource name, read [the infrastructure naming reference](../flow-plan/references/infra-naming.md), even if no retained plan exists. Preserve an existing name unless its authorized change includes the necessary impact and migration work.

For a changed shared interface, reconcile the actual producers, consumers, coexisting versions, and delivery order from the plan or current project evidence before implementation. Verify affected compatibility, including partial rollout when relevant, using the existing integration or contract checks.

Use red-green-refactor TDD by default for new or changed automatically testable behavior when the repository has usable test infrastructure: write the behavioral test, observe it fail for the intended reason, implement the change, observe it pass, then refactor while preserving passing checks. A missing test for a new behavior is a reason to write one, not to skip TDD. For behavior-preserving refactors, use existing tests or characterization; trivial presentation or passive documentation changes need no artificial failing test. If no usable test infrastructure exists, choose an available validator or direct check and report the limitation; do not introduce a framework without an authorized need.

Preserve implementation that already exists before RED. Verify the regression check against an isolated pre-change baseline or controlled reproduction when feasible. Report fail-to-pass only when that comparison was observed; otherwise disclose the missing evidence, without claiming historical TDD. A dependency/setup error is not behavioral RED. Review what assertions actually prove; do not weaken a valid check to make the task appear complete. Remove obsolete code, configuration, or temporary residue created by the change when that cleanup is within scope.

Before writing a retained supporting artifact, resolve its destination through the applicable global artifact-placement instructions and existing work identifier. When a retained plan exists, keep its task list and current status consistent with the work: `tasks.md` and the status in `proposal.md` for a change folder. Update that status before reporting a delivery milestone, so the report and the record agree, and version the update at the points the delivery decision names. Preserve either established `<topic>-plan.md` or `<topic>.plan.md` names; do not rename a valid retained handoff merely to normalize its filename. Mark a task complete only when its expected result is evidenced. Record changed decisions, failed or omitted checks, and why a completed task was reopened. Keep raw output separate and link the useful evidence.

For any multi-step work, with or without a retained plan, when the host exposes a native task list, list the tasks or tickets in it at the start and mark each one in progress when it starts and completed when it completes, including tickets added during the session, so the user sees progress. A retained plan's record, when it exists, still governs.

Keep task-started processes identifiable. Before telling the user that a child is still in progress after a silence longer than its expected duration, check its liveness through its tool activity or processes; when a child returns, confirm it left no running processes. Silence alone does not mean a command is hung: inspect available liveness and progress signals before terminating it. Use the host’s supported long-running execution controls when needed and continue independent work. Do not stop unrelated processes.

## Verify and close

Compare the final changes and observed behavior against the acceptance criteria and authorized scope. Local in-vivo verification and UI review are required when applicable, without a separate approval request; unavailable prerequisites leave the affected criterion unverified. Run affected tests and applicable build, typecheck, and lint commands that the project provides. Exercise changed runtime behavior through the real integration path; inspect changed user-facing surfaces rendered and interactive, including relevant states and viewports. In deployed environments, apply the shared gate authorization boundary before the functional walk. Tests passing or a successful build do not replace those observations. Before reporting a UI change complete, obtain the independent UI review and in-vivo results required by the verification reference; the implementer's own walk satisfies neither. Apply the shared verification reference for domain-specific depth and timing.

Use existing evidence only while its relevant code, configuration, and environment remain valid. Close regression coverage through the project's actual CI gate or run the existing full suite locally when no downstream coverage exists. After a fix, rerun the failing check and affected scope; broaden to affected callers and consumers when the fix touches shared code or new failure evidence appears. Do not mark a required but unavailable check as passed or silently omit it. Distinguish failed, not verified/blocked, and not applicable with a reason; resolve authorized local prerequisites and report any remaining delivery constraint. Check that the handoff's paths and next step still match the final state.

For UI work awaiting human validation, leave the verified built local stack available and deliver the walkthrough and temporary captures defined in [verification](references/verification.md); do not ask again merely to keep this handoff available. Pending human review means these resources are still needed. Once no longer needed, stop only task-started temporary processes and apply the global UI-image hygiene rule, preserving user-selected captures and pre-existing services, sessions and evidence. Report remaining resources and how to stop them, and respect an explicit shutdown request.

Report what changed, what was verified, unresolved failures or uncertainty, and the next step. Use a separate report only when a retained deliverable needs it; a concise update to the plan can be enough. A successful tool exit, a checked box, or an agent's assertion alone does not establish completion.

Complete delivery actions already explicitly authorized, subject to their conditions; for a branch, commit, push, pull request, or merge, read [git-workflow](../git-workflow/SKILL.md) first. Otherwise stop at the authorized result with remaining delivery steps identified. Once a change folder's work is integrated into its base branch, or the change is abandoned, close it as [change records](../flow-plan/references/change-records.md) describes; its versioned delivery follows the same authorization.

When the work item completes, close it in this order; the last step ends the turn:

1. Run the global task-close cleanup and, when you merged, the post-merge cleanup in [git-workflow](../git-workflow/SKILL.md). Do not delete prior or uncertain material.
2. Write the completion report with its cleanup line.
3. Ask the global close question.
