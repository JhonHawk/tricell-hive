---
name: flow-research
description: Investigate a technical question or project and delivery state by inspecting code, documentation, and evidence, comparing alternatives or resolving conflicting claims. Use for substantive research and diagnosis, and for a question whether or how to do work, such as "can you fix this record?", even when its cause is already known; not for a simple lookup or implementation the user already authorized.
---

# Research a question

Turn uncertainty into a supported answer that the next decision can use. This procedure can stand alone or support an already authorized task; it does not require a planning or implementation phase afterward.

## Frame the investigation

Identify the question, the decision it informs, relevant constraints, and what evidence would change the answer. Use context already supplied. Ask only for missing intent that affects the conclusion; discover repository facts directly.

Inspect the relevant entry points, callers, tests, configuration, and project guidance before extending the search. When the question concerns project state or delivery, also reconcile the relevant backlog, documentation, code, pull requests, CI, and deployment evidence; for the project's status or a backlog listing, read [backlog report](references/backlog-report.md) before reporting. Follow references that bear on the question rather than inventorying the entire repository. Let uncertainty and consequences determine depth, not a fixed file or source count.

Treat search hits as entry points to inspect, and state the searched scope and vocabulary before claiming absence; no matches alone do not establish that a capability or behavior is missing. When investigating a failure, regression, or incident with an uncertain cause, read [diagnosis](references/diagnosis.md).

In the main thread, delegate under the shared delegation contract any question whose investigation reads installed packages, built or minified code, long documentation, or logs, or needs broad searches or several independent source families: that output stays in the main thread's context for the rest of the session. Keep a short, bounded lookup in the main thread; at the sixth search or read without an answer, not counting project guidance, skills, and their references, stop and delegate the rest, and delegate any lookup that grows into diagnosis. Before the first delegated lookup, tell the user in one sentence which questions go to which role, under the shared delegation rules. Delegation is internal work and needs no separate approval. Select the installed `hive-research` role rather than a host's built-in explorer; when the launcher does not list that role, launch a generic child with its contract, as the shared delegation fallback requires, instead of researching alone. Ask for source-backed findings, counterevidence, and access limits. The main thread checks consequential claims against their sources and resolves conflicting conclusions; do not treat agreement among subagents as proof.

## Establish evidence

- Distinguish current observed behavior, documented intent, inference, and recommendation. A test's presence is not evidence that it passed; a documented capability is not evidence that this installation exposes it.
- For project status, distinguish declared state from observed evidence and unresolved discrepancies. Report discrepancies without changing tickets, documentation, code, or delivery state unless those effects are separately authorized.
- For version-dependent external behavior, and before recommending how to configure or use an external library, tool, or service, consult its current primary documentation for the installed version. Use an available documentation skill or tool when applicable; record meaningful access or freshness limits rather than inventing verification.
- Resolve conflicting sources against their dates, versions, scope, and direct observations. Neither backlog nor documentation takes automatic precedence. Missing delivery evidence means delivery is not verified, not that it did not happen. Preserve an unresolved contradiction with the evidence needed to settle it.
- Use a bounded check when reading alone cannot answer a consequential question and the check is within the task's authorized effects. Do not turn a read-only investigation into an implementation or deployment.

## Deliver the finding

State the answer, supporting source locations or URLs, material alternatives and tradeoffs, remaining uncertainty, and the decision or next step it enables. Do not present a recommendation as an implemented result.

When the findings identify follow-up work that no ticket covers, ask the user whether to open one, together with the findings or earlier when other questions are pending; do not block the investigation on the answer. A question the findings fully answer needs no ticket. Infer the tracker from project guidance, existing issue links, or configured project context, and ask where to record it only when none is established. Open the ticket only after the user agrees, check for duplicates first, and report its link.

Keep findings and iterative debate in the conversation. Apply the shared research-retention authorization rule before creating or updating a research file; delegate findings back to the conversation without requesting separate written reports unless retention is authorized. When proposing retention, name the benefit and intended document, then wait for the user’s answer while continuing the discussion. After authorization, resolve the destination through the global artifact-placement instructions and existing work identifier. Keep the saved record concise: question, evidence, findings, decisions and unresolved items; link selected evidence rather than copying raw output.

For an investigation-only request, stop at the findings. When the first findings on a topic concern work the user could run next, such as a ticket or a change, end them with one line of plain text, not the native question tool, naming the routes under the global route criterion, `flow-plan` when the work spans several tickets, repositories, or steps and `flow-build` only for a small understood change, and the one you recommend with its reason. Do not repeat that line in answers to follow-up questions, adjustments, or reframings of the same topic; when the findings change the recommended route, say so in one line. Ask the route through the native question tool, with one option per flow and one to stop at the findings, only when the user asks to do the work, such as "can you fix this record?", or asks how to continue. Without a native question tool, ask the same route in text and end the turn. Start the chosen flow without a new command. Within a broader authorized task, carry the findings into that task without treating this skill as another approval gate. Preserve established records named either `<topic>.research.md` or legacy `<topic>-research.md`; do not rename records merely to normalize the convention.
