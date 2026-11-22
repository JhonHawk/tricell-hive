---
name: hive-mode
description: Settle what the user needs before any work and pick the playbook for its kind, then send it to flow-build or flow-plan, or answer a question, a reported failure or a ticket assessment. Use for every request to do work with no ready plan, an investigation one bounded lookup cannot answer, a failure reported without a fix request, a question whether or how to do work, assessing a ticket, auditing a codebase for what to improve, the project's status, backlog or what to work on next, and whether a published security advisory affects your repositories; not for a simple lookup.
---

# Settle what the user needs and pick the playbook

At entry, read `stage.md` in this skill's folder whole: it holds the leaves questions, report-endings, so a step that links one of them needs no new read until a compaction, after which read the packet again. If it is missing, read the linked leaves instead; if its last line is not its packet end line, read the rest; if a `## ` heading the end line names is missing, read that leaf; and if the output says it was cut, read the linked leaves.

For an explicitly delegated E2E objective, read [coordination](../flow-build/references/e2e-coordination.md) before routing covered stages; a request here alone grants no build authority.

This skill settles what the work is and which playbook it follows; `flow-plan` settles how. It publishes nothing to the viewer, edits no repository file and changes nothing in the tracker or shared data except through the work it hands to `flow-build`, a single action the user asks for, and [local write](../hive-core/local/local-write.md)'s writes. Follow-up questions keep it open and authorize nothing.

Before step 1, open [investigate](playbooks/investigate.md) when the user asks to audit a codebase or find what to improve in it without one question to answer, [ticket status](playbooks/ticket-status.md) when they ask for the project's status, its backlog, unfinished work or what to work on next, and [dependency security](playbooks/dependency-security.md) when they ask whether a published advisory, CVE or package compromise affects their repositories. The last two replace steps 0 to 5 and end as they set. A walkthrough or demo of a built interface follows flow-verify's [walkthrough](../hive-core/ui/walkthrough.md) alone.

0. **Complete request.** When the request already says everything that changes the result, such as "add this user to the repository", keep the next steps short: read what it touches (step 1), check that nothing blocks it, such as missing access or a false premise, and skip step 3's questions. An action that changes no repository file is then done and reported, under the global guidance's effect rules; a change to repository files takes step 5's route. A request whose deliverable is a design in Figma, such as replicating a screen, is not drawn here: check the connection as [design stage](../hive-core/ui/design-stage.md) sets; with it, continue into [create-design](../create-design/SKILL.md) in the same turn, which draws it and does not report it as done before the user approves it, and without it report that missing connection as the blocker. A request to review a design someone already drew goes to its [review entry](../create-design/references/review.md) in the same turn. A request whose deliverable is a requirement document, such as the PRD, the functional requirements or the business rules of a module or feature, or the validation of one someone else wrote, is not written here either: follow [create-prd](../create-prd/SKILL.md) in the same turn, which reads the source and asks its own questions. When one bounded lookup (as «Proportionality» defines it) answers a question, report the answer with its source and uncertainty, and stop.
1. **Read before asking.** Before the first read, unless the reads fit one bounded lookup, write in the message the aspects the answer needs, the sources each takes and which go to a child under step 4. Then read the code, the ticket with its comments, also when a child reads it for you, the configuration and verified memory the request touches; these reads count toward step 4's limits. Never ask what they answer.
2. **Play the request back.** State in your words what the user wants, for whom and why, with what you found: a conflict, a false premise, a cheaper path, or that the request is really several. Name the playbook you pick, or that none fits, ([Pick the playbook](#pick-the-playbook)).
3. **Ask what is missing.** When something that changes what gets built is still open, such as why now, what success looks like, what must not happen or what they already tried, ask one open question per message, in the text, with why you ask and no menu of example answers, and add that they can tell you to just start. When they say to just guess, propose something concrete. Choices between options go through the question tool, as [asking a round](../hive-core/messages/questions.md) sets.
4. **Investigate what is unknown.**
   - Delegate to a `hive-research` child each aspect of step 1's list that needs more than 5 large reads, as the global guidance's «Proportionality» defines them, before reading it, and do not read in the main thread what a child covers; filtered reads do not count. When the main thread reaches its sixth large read on one aspect without its answer, stop and delegate what remains of it. Add an aspect found later to the list before its first read. Launch independent children side by side, briefed and awaited as [delegation](../hive-core/delegation/delegation.md) sets. A host's built-in explorer only locates files; verify its output.
   - Run a tracker query that lists tickets in the main thread, capped at 50 tickets and short fields; use `hive-read-state` children only for several independent collections at once. Before proposing a ticket, check for an existing one as [create-issue](../create-issue/SKILL.md) step 2 sets.
   - For a failure whose cause is not yet shown, list the plausible causes and check each one, in parallel when independent; report the ones ruled out with their evidence.
   - Name each system the decision touches and whether you searched it.
   - Open the source of each claim that decides the answer instead of recalling it. Resolve conflicting sources by date, version, scope and direct observation; missing delivery evidence means delivery is not verified, not that it did not happen.
5. **Close the what.** Report the answer, its evidence and what remains uncertain, and play back what you understood, marking what the user said apart from what you propose. Then:
   - When the user picked a ticket for research without asking for a change, and the findings confirm work it asks for, end, after «Assess a chosen ticket», with the next-step block toward `flow-plan`, or toward `flow-build` when that work is a small understood change (the global guidance's «Proportionality»), unless the design stage below sends it to its own next-step block, and never continue into build in the same turn. Every other analysis, explanation, review, diagnosis or assessment without a requested change ends as [report endings](../hive-core/messages/report-endings.md) sets.
   - When the findings meet the research-into-build continuation of [same-turn continuations](../hive-core/messages/report-endings.md#same-turn-continuations), continue into `flow-build` as it sets, except for a new screen that the design stage below sends to its next-step block. A reply to your question, or the acceptance of a proposal you made, never starts it: close the what again and end with the next-step block, unless the user's own request already set one of this step's continuations, such as a design in Figma.
   - When the request's deliverable is a design in Figma, continue into `create-design` or report the blocker, or into its review entry; when it is a requirement document, follow `create-prd`; both as step 0 sets, whether or not step 3 asked questions.
   - Otherwise, end with the route to `flow-plan` and its reason, unless the design stage below sends the work to its next-step block.

   When this step ends with a route to `flow-plan`, `create-design` or `flow-build`, or continues into `flow-build`, write one line with the scope estimated from the reads, before the next-step block when there is one, so the user can narrow it: about how many files and which areas (for a chosen ticket, its complexity from «Assess a chosen ticket»), the playbook, and the stages and checks the guidance sends such work through, such as plan, per-task verification, a review, CI, and a promotion only when the request asks for one. Mark it an estimate, such as "Scope (estimate): ~42 files (contracts, runtime, tests, manual) · plan + per-task verification + review + UX review · PR and CI". A reply that narrows the scope without naming a flow gets the revised line and the same ending again.

   **A screen to design.** For a change the user asked for or a ticket they picked that adds a screen or recomposes one, decide whether to offer [create-design](../create-design/SKILL.md) as [design stage](../hive-core/ui/design-stage.md) sets, and name that stage in the scope line when you offer it.

   Creating a ticket is a single action, offered as a reply or asked, never part of a route; once answered, it follows [create-issue](../create-issue/SKILL.md). The user's reply authorizes only the flow it names, or the single action it answers.

## Assess a chosen ticket

For a chosen ticket, write the report in these blocks, in order, headed in the session language:

1. **Summary:** **impact** (what uses what it changes, found by searching its literal values and direct callers, and whether it touches production), **complexity** (the files, modules and contracts it changes), **dependencies** (the tickets it waits on or blocks).
2. **Verdict:** ready or not, naming by number each check of the ready list of [ticket types](../create-issue/references/ticket-types.md) it fails with what is missing, each contradicting spec, and missing access or answers. A failed check or a contradicting spec means not ready.
3. **Proposed criteria**, when it has none.
4. **Decided**, as [questions](../hive-core/messages/questions.md) «Decide and list» sets; then the card with what a–f send.

Settle a not-ready ticket's scope here: the plan settles how, never what enters. Step 5's ending closes this message when nothing is asked, else the one after the answers.

## Pick the playbook

A playbook, `playbooks/<kind>.md`, orders the stages of a kind of work and names what is specific to it. Read the one you pick before the route; the stages stay commands that read it.

| Playbook | Work |
| --- | --- |
| `feature` | new behavior, no screen |
| `ui-change` | a screen added or changed |
| `bugfix` | a defect to fix |
| `hotfix` | a production defect that cannot wait |
| `data-change` | stored data or a migration |
| `dependency-security` | a dependency or a published advisory |
| `infra-ci` | infrastructure, CI or tooling |
| `investigate` | a question, diagnosis or audit |
| `ticket-status` | status, backlog or what to take next |
| `deliverable` | a document or design as the product |
| `pr-review-deploy` | a pull request review or a promotion |
| `refactor` | restructuring that keeps behavior |
| `perf` | a measured speed or size goal |

- **One playbook.** Open at most one. When none fits, or the work is large and crosses several kinds, follow steps 0 to 5 and open none.
- **Changing playbook.** A switch, such as `investigate` to `bugfix` once the findings confirm a defect and the user asks for its fix, is an explicit handoff, never a silent jump: state it as a scope change with its class and destination in one line, as [scope changes](../hive-core/plan/scope-changes.md) sets, and record the new playbook.
- **Record.** With a change folder, the flow that writes it records its name in the `Playbook` field of its `proposal.md`. Without a change folder, `flow-build` writes a `- Playbook: <kind>` line by hand in the work's checkpoint file ([checkpoint](../hive-core/plan/checkpoint.md)). Each later stage reads it from there.

Reply: ends as the row of [report endings](../hive-core/messages/report-endings.md) that fits the request: the analysis row, the ticket-assessment row, or [the next-step block](../hive-core/messages/report-endings.md#the-next-step-block) toward `flow-plan`, `create-design` or `flow-build`, with step 5's scope line before it.
