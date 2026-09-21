---
name: flow-research
description: Investigate a technical question or project and delivery state by inspecting code, documentation, and evidence, comparing alternatives or resolving conflicting claims. Use for substantive research and diagnosis, not a simple lookup or implementation already understood.
---

# Research a question

Turn uncertainty into a supported answer that the next decision can use. This procedure can stand alone or support an already authorized task; it does not require a planning or implementation phase afterward.

## Frame the investigation

Identify the question, the decision it informs, relevant constraints, and what evidence would change the answer. Use context already supplied. Ask only for missing intent that affects the conclusion; discover repository facts directly.

Inspect the relevant entry points, callers, tests, configuration, and project guidance before extending the search. When the question concerns project state or delivery, also reconcile the relevant backlog, documentation, code, pull requests, CI, and deployment evidence. Follow references that bear on the question rather than inventorying the entire repository. Let uncertainty and consequences determine depth, not a fixed file or source count.

When useful and authorized, delegate independent questions or source families under the shared delegation contract. Ask for source-backed findings, counterevidence, and access limits. The main thread checks consequential claims against their sources and resolves conflicting conclusions; do not treat agreement among subagents as proof.

## Establish evidence

- Distinguish current observed behavior, documented intent, inference, and recommendation. A test's presence is not evidence that it passed; a documented capability is not evidence that this installation exposes it.
- For project status, distinguish declared state from observed evidence and unresolved discrepancies. Report discrepancies without changing tickets, documentation, code, or delivery state unless those effects are separately authorized.
- For version-dependent external behavior, consult current primary documentation and the installed version. Use an available documentation skill or tool when applicable; record meaningful access or freshness limits rather than inventing verification.
- Resolve conflicting sources against their dates, versions, scope, and direct observations. Neither backlog nor documentation takes automatic precedence. Missing delivery evidence means delivery is not verified, not that it did not happen. Preserve an unresolved contradiction with the evidence needed to settle it.
- Use a bounded check when reading alone cannot answer a consequential question and the check is within the task's authorized effects. Do not turn a read-only investigation into an implementation or deployment.

## Deliver the finding

State the answer, supporting source locations or URLs, material alternatives and tradeoffs, remaining uncertainty, and the decision or next step it enables. Do not present a recommendation as an implemented result.

Keep ordinary answers in the conversation. Retain research when it is requested, supports work that must resume or be handed off, preserves a consequential decision, or would be costly to reconstruct. Before writing a retained artifact, resolve its destination through the applicable global artifact-placement instructions and existing work identifier. Choose the record type and filename using the global naming convention. Keep it concise: question, inspected evidence, findings, decisions or recommendation, and unresolved items. Link selected evidence rather than copying raw output.

For an investigation-only request, stop at the findings. Within a broader authorized task, carry the findings into that task without treating this skill as another approval gate. Preserve established records named either `<topic>.research.md` or legacy `<topic>-research.md`; do not rename records merely to normalize the convention.
