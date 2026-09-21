---
name: flow-build
description: Implement or resume an authorized code or configuration change, reconcile a supplied plan with current state, verify behavior, and close with evidence. Use for execution rather than research-only or planning-only requests; a formal plan is not a prerequisite for a small understood change.
---

# Implement and verify authorized work

Use the requested outcome or supplied plan as the execution brief. This skill does not require a formal plan or a prior research skill invocation, and does not grant effects beyond the user's instructions.

## Reconcile before editing

Read the relevant project guidance, current source and tests, and the supplied plan and references. Inspect existing changes and completed work. Treat the record as a claim to reconcile with current evidence, not a script to replay blindly.

Identify the next unmet acceptance criterion and its dependencies. Preserve valid completed work. If the current state invalidates an important decision, investigate the mismatch and update the affected plan or ask for the missing decision before dependent edits. Resolve ordinary implementation details within scope; do not stop at every task boundary merely to ask again.

## Execute in verifiable increments

Choose a coherent next result and implement it in the existing source layout. Use native tools and established project capabilities. Delegate independent bounded work when available and authorized, with clear ownership, relevant context, and expected evidence; dependent edits remain ordered.

Use the project's applicable testing approach. For a defect, establish the failing behavior when feasible; after changing it, run the relevant checks and observe the result. Do not add a new testing framework or impose TDD merely because a plan exists. Preserve meaningful failing checks rather than weakening them to make the task appear complete.

When a retained plan exists, keep its task list and current status consistent with the work. Mark a task complete only when its expected result is evidenced. Record changed decisions, failed or omitted checks, and why a completed task was reopened. Keep raw output separate and link the useful evidence. Native task tools may reflect this progress without replacing the canonical record.

## Verify and close

Compare the final changes and observed behavior against the acceptance criteria and authorized scope. Run the checks appropriate to the changes; distinguish executed checks from suggested commands. Check that the handoff's paths and next step still match the final state.

Report what changed, what was verified, unresolved failures or uncertainty, and the next step. Use a separate `<topic>.report.md` only when a retained deliverable needs it; a concise update to the plan can be enough. A successful tool exit, a checked box, or an agent's assertion alone does not establish completion.

Complete delivery actions already explicitly authorized, subject to their conditions. Otherwise stop at the authorized result with remaining delivery steps identified. Apply the existing artifact and memory hygiene rules without deleting prior or uncertain material.
