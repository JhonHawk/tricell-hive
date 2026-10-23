---
name: flow-research
description: Investigate a technical question or project and delivery state by inspecting code, documentation, and evidence, comparing alternatives or resolving conflicting claims. Use for substantive research and diagnosis, for a reported failure, bug, error, or regression that comes without an explicit fix request, and for a question whether or how to do work, such as "can you fix this record?", even when its cause is already known; not for a simple lookup or implementation the user already authorized.
---

# Research a question

Turn uncertainty into a supported answer that the next decision can use. This procedure can stand alone or support an already authorized task; it does not require a planning or implementation phase afterward.

## Frame the investigation

Before resolving needed project settings, read [project settings](../flow-plan/references/project-settings.md). Before delegating or starting an external-work wait, read [delegation and waits](references/delegation.md). Before formulating a user decision, read [questions](../flow-plan/references/questions.md).

Identify the question, the decision it informs, relevant constraints, and what evidence would change the answer. Use context already supplied. Ask only for missing intent that affects the conclusion; discover repository facts directly.

Inspect the relevant entry points, callers, tests, configuration, and project guidance before extending the search. When the question concerns project state or delivery, also reconcile the relevant backlog, documentation, code, pull requests, CI, and deployment evidence; for the project's status or a backlog listing, read [backlog report](references/backlog-report.md) before reporting. When collecting the state of named tickets, pull requests, CI runs, or deployments, or a tracker query with its project and states, would produce long output, delegate that collection to the `hive-read-state` role with those targets and the fields the report groups by. Follow references that bear on the question rather than inventorying the entire repository. Let uncertainty and consequences determine depth, not a fixed file or source count.

Treat search hits as entry points to inspect, and state the searched scope and vocabulary before claiming absence; no matches alone do not establish that a capability or behavior is missing. When investigating a failure, regression, or incident with an uncertain cause, read [diagnosis](references/diagnosis.md).

In the main thread, keep a short, bounded lookup direct and delegate the rest under the shared delegation contract at the threshold the global delegation rule sets, which counts searches and the kinds of source read; also delegate any lookup that grows into diagnosis. Before the first delegated lookup, tell the user in one sentence which questions go to which role, under the shared delegation rules. Delegation is internal work and needs no separate approval. Select the installed `hive-research` role rather than a host's built-in explorer; when the launcher does not list that role, launch a generic child with its contract, as the shared delegation fallback requires, instead of researching alone. In each research child's brief, state what is already known or settled, what is out of scope, and when to stop: once the answer can be named and the evidence converges, or once the assigned scope is exhausted. The role's contract sets what the child returns. Check the claims that decide the answer or the recommendation by opening their cited locators, not by repeating the child's searches; search again only when a locator is missing, does not support its claim, or conflicts with another source. Send high-consequence premises that this check cannot settle to one `hive-refute-claim` child, asking for one verdict per premise. Resolve conflicting conclusions yourself; agreement among subagents is not proof.

## Establish evidence

- Distinguish inspected facts, source claims, inference, and recommendations. A test's presence is not evidence that it passed; a documented capability is not evidence that this installation exposes it.
- For project status, distinguish declared state from observed evidence and unresolved discrepancies. Report discrepancies without changing tickets, documentation, code, or delivery state unless those effects are separately authorized.
- For version-dependent external behavior, and before recommending how to configure or use an external library, tool, or service, consult its current primary documentation for the installed version. Use an available documentation skill or tool when applicable; record meaningful access or freshness limits rather than inventing verification.
- Resolve conflicting sources against their dates, versions, scope, and direct observations. Neither backlog nor documentation takes automatic precedence. Missing delivery evidence means delivery is not verified, not that it did not happen. Preserve an unresolved contradiction with the evidence needed to settle it.
- Use a bounded check when reading alone cannot answer a consequential question and the check is within the task's authorized effects. Do not turn a read-only investigation into an implementation or deployment.

## Deliver the finding

Before reporting findings or relaying child/tool results with citations, links or externally defined IDs, read [reporting](../flow-close/references/reporting.md). Before writing a retained support artifact, read [artifact layout and retention](../workspace-conventions/references/artifact-layout.md); retention still needs its existing authorization.

State the answer, supporting source locations or URLs, material alternatives and tradeoffs when they answer the question investigated, remaining uncertainty, and the decision or next step it enables. Do not present a recommendation as an implemented result.

When the findings identify follow-up work that no ticket covers, ask whether to open one in the same round as the route question below, as its own question with an ID; do not block the investigation on the answer. A question the findings fully answer needs no ticket. Infer the tracker from project guidance, existing issue links, or configured project context, and ask where to record it only when none is established. Open the ticket only after the user agrees, check for duplicates first, and report its link.

Keep findings and iterative debate in the conversation. Apply the shared research-retention authorization rule before creating or updating a research file; delegate findings back to the conversation without requesting separate written reports unless retention is authorized. When proposing retention, name the benefit and intended document, then wait for the user’s answer while continuing the discussion. After authorization, resolve the destination through the global artifact-placement instructions and existing work identifier. Keep the saved record concise: question, evidence, findings, decisions and unresolved items; link selected evidence rather than copying raw output.

For an investigation-only request, stop at the findings. When a round of findings concerns work the user could run next, such as a ticket or a change, end that same message with the route question, asked by following [questions](../flow-plan/references/questions.md): the alternatives that resolve the findings, such as fixes, with their effects and tradeoffs, each naming the flow it runs, `flow-plan` when the work spans several tickets, repositories, or steps and `flow-build` only for a small understood change, plus one option to keep investigating and one to stop at the findings, the recommended one first with its reason. A question the findings fully answer, with no work to run, ends without a route question. When the user asks to do the work, such as "can you fix this record?", or how to continue, ask the route question if this round has not. Ask it once per round; ask it again only when the findings change. Within a broader authorized task, carry the findings into that task without treating this skill as another approval gate. Preserve established records named either `<topic>.research.md` or legacy `<topic>-research.md`; do not rename records merely to normalize the convention.

When the next steps alternate investigation and implementation toward one measurable objective, also offer an iterative route in the same route question. A round is one investigation followed by one implementation and its measurement; count and measure only complete rounds. The option states the objective and how it is measured, the round limit you propose for this case, and its stop conditions, and names `flow-research` and `flow-build` as the flows it alternates. Every step must stay within the local editing and verification that implementation already authorizes, in the current project's working tree, the task's scratch directory, or its support workspace, and each round's change must be a small understood change for `flow-build`. Choosing the route authorizes its rounds without a new route question: the first build round asks `flow-build`'s delivery or review question when it applies, and later rounds reuse that answer. Report each round in one line with its measurement and continue. Stop the route when:

- the objective is met;
- the round limit is reached;
- a round ends without a measurable improvement;
- a step needs an effect the route does not cover and no existing grant authorizes, such as a commit, an installation, or deleting or overwriting material the task did not create;
- a round's change needs `flow-plan`;
- a finding changes the objective or scope;
- a decision is genuinely the user's.

When the route stops with changes in the working tree, run [flow-close](../flow-close/SKILL.md): as the completion point when the objective is met, as a pause otherwise. Without such changes, end with the route question.
