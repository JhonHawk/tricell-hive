---
name: flow-build
description: Implement or resume an authorized code or configuration change, reconcile a supplied plan with current state, verify behavior, and close with evidence. Use for execution rather than research-only or planning-only requests; a formal plan is not a prerequisite for a small understood change.
---

# Implement and verify authorized work

Use the requested outcome or supplied plan as the execution brief. This skill does not require a formal plan or a prior research skill invocation, and does not grant effects beyond the user's instructions.

## Reconcile before editing

Read the relevant project guidance, current source and tests, and the supplied plan and references. Inspect existing changes and completed work. Treat the record as a claim to reconcile with current evidence, not a script to replay blindly.

Identify the next unmet acceptance criterion and its dependencies. Preserve valid completed work. If the current state invalidates an important decision, investigate the mismatch and update the affected plan or ask for the missing decision before dependent edits. Resolve ordinary implementation details within scope; do not stop at every task boundary merely to ask again.

Identify the available test runner, actual project check/build commands, relevant baseline failures, and the evidence needed for acceptance. For changes involving integration, runtime, UI, performance, hardware, or costly verification, read [verification by risk and surface](references/verification.md) before editing to select checks, prerequisites, and their timing. Reconcile a supplied verification plan rather than silently weakening it.

Before installing a missing tool, check the project-local and available system installations and the required version. Prefer the project’s existing setup. Distinguish task-local dependency setup from system-wide installs or shell/global configuration changes; the latter require explicit authorization, which may already be present. Verify the actual OS and shell before using platform-specific commands rather than assuming them from the harness name.

## Execute in verifiable increments

Choose a coherent next result and implement it in the existing source layout. Look for an existing solution or pattern before adding a mechanism or abstraction, and justify a new one by the need it serves. Confirm that reused tools or patterns meet this project's requirements. Use the project's declared dependency manager and preserve its lockfile conventions. When delegation is useful and authorized, assign independently implementable units with settled interfaces under the shared delegation contract. Select technical references from the actual project stack and affected behavior. Keep dependent edits ordered. Assign independent verification or a specialist review when the acceptance criteria or risk warrant it; the implementer still verifies its own change.

Before adding a dependency, assess whether an existing project dependency, standard library, or native platform capability meets the requirement, including established design-system conventions. Choose based on compatibility, maintainability, and fit rather than a universal library preference. Explain a new dependency’s purpose and any overlap; consolidating existing usages remains separate unless authorized.

When choosing an infrastructure resource name, read [the infrastructure naming reference](../flow-plan/references/infra-naming.md), even if no retained plan exists. Preserve an existing name unless its authorized change includes the necessary impact and migration work.

Use red-green-refactor TDD by default for new or changed automatically testable behavior when the repository has usable test infrastructure: write the behavioral test, observe it fail for the intended reason, implement the change, observe it pass, then refactor while preserving passing checks. A missing test for a new behavior is a reason to write one, not to skip TDD. For behavior-preserving refactors, use existing tests or characterization; trivial presentation or passive documentation changes need no artificial failing test. If no usable test infrastructure exists, choose an available validator or direct check and report the limitation; do not introduce a framework without an authorized need.

Preserve implementation that already exists before RED. Verify the regression check against an isolated pre-change baseline or controlled reproduction when feasible. Report fail-to-pass only when that comparison was observed; otherwise disclose the missing evidence, without claiming historical TDD. A dependency/setup error is not behavioral RED. Review what assertions actually prove; do not weaken a valid check to make the task appear complete. When the cause remains uncertain, compare hypotheses against evidence before accumulating fixes. Remove obsolete code, configuration, or temporary residue created by the change when that cleanup is within scope.

Before writing a retained supporting artifact, resolve its destination through the applicable global artifact-placement instructions and existing work identifier. When a retained plan exists, keep its task list and current status consistent with the work. Preserve either established `<topic>-plan.md` or `<topic>.plan.md` names; do not rename a valid retained handoff merely to normalize its filename. Mark a task complete only when its expected result is evidenced. Record changed decisions, failed or omitted checks, and why a completed task was reopened. Keep raw output separate and link the useful evidence. Native task tools may reflect this progress without replacing the canonical record.

Keep task-started processes identifiable. Silence alone does not mean a command is hung: inspect available liveness and progress signals before terminating it. Use the host’s supported long-running execution controls when needed and continue independent work. Do not stop unrelated processes.

## Verify and close

Compare the final changes and observed behavior against the acceptance criteria and authorized scope. Run affected tests and applicable build, typecheck, and lint commands that the project provides. Exercise changed runtime behavior through the real integration path; inspect changed user-facing surfaces rendered and interactive, including relevant states and viewports. Tests passing or a successful build do not replace those observations. Apply the shared verification reference for domain-specific depth and timing.

Use existing evidence only while its relevant code, configuration, and environment remain valid. Close regression coverage through the project's actual CI gate or run the existing full suite locally when no downstream coverage exists. After a fix, rerun the failing check and affected scope; broaden when the impact warrants it. Do not mark a required but unavailable check as passed or silently omit it. Distinguish failed, not verified/blocked, and not applicable with a reason; resolve authorized local prerequisites and report any remaining delivery constraint. Check that the handoff's paths and next step still match the final state.

Stop task-started temporary servers and browser sessions when they are no longer needed. If the user needs one for ongoing verification, follow an existing instruction to leave it running or resolve that handoff with the user; report its address, purpose, and how to stop it. Preserve pre-existing services and sessions.

Report what changed, what was verified, unresolved failures or uncertainty, and the next step. Use a separate `<topic>.report.md` only when a retained deliverable needs it; a concise update to the plan can be enough. A successful tool exit, a checked box, or an agent's assertion alone does not establish completion.

Complete delivery actions already explicitly authorized, subject to their conditions. Otherwise stop at the authorized result with remaining delivery steps identified. Apply the existing artifact and memory hygiene rules without deleting prior or uncertain material.
