---
name: flow-research
description: Settle what the user needs before any work, then send it to flow-build or flow-plan, or answer a question, a reported failure or a ticket assessment. Use for every request to do work without a ready plan, an investigation one bounded lookup cannot answer, a failure reported without a fix request, a question whether or how to do work, and assessing a ticket; not for the project's status or backlog (flow-status) or a simple lookup.
---

# Research what the user needs

Research settles what the work is; `flow-plan` settles how. It publishes nothing to the viewer, and changes nothing in the project, tracker or shared data except through the work it hands to `flow-build` and a single action the user asks for. Follow-up questions keep it open and authorize nothing.

0. **Complete request.** When the request already says everything that changes the result, such as "add this user to the repository", keep the next steps short: read what it touches (step 1), check that nothing blocks it, such as missing access or a false premise, and skip step 3's questions. An action that changes no repository file is then done and reported, under the global guidance's effect rules; a change to repository files takes step 5's route. When one bounded lookup (as «Proportionality» defines it) answers a question, report the answer with its source and uncertainty, and stop.
1. **Read before asking.** Before the first read, unless the reads fit one bounded lookup, write in the message the aspects the answer needs, the sources each takes and which go to a child under step 4. Then read the code, ticket, configuration and verified memory the request touches; these reads count toward step 4's limits. Never ask what they answer.
2. **Play the request back.** State in your words what the user wants, for whom and why, with what you found: a conflict, a false premise, a cheaper path, or that the request is really several.
3. **Ask what is missing.** When something that changes what gets built is still open, such as why now, what success looks like, what must not happen or what they already tried, ask one open question per message, in the text, with why you ask and no menu of example answers, and add that they can tell you to just start. When they say to just guess, propose something concrete. Choices between options go through the question tool, as «Questions and routes» sets.
4. **Investigate what is unknown.**
   - Delegate to a `hive-research` child each aspect of step 1's list that needs more than 5 large reads, as the global guidance's «Proportionality» defines them, before reading it, and do not read in the main thread what a child covers; filtered reads do not count. When the main thread reaches its sixth large read on one aspect without its answer, stop and delegate what remains of it. Add an aspect found later to the list in the message before its first read. Launch independent children together, briefed and awaited as [delegation](../flow-build/references/delegation.md) sets. A host's built-in explorer only locates files; verify what it returns.
   - Run a tracker query that lists tickets in the main thread, capped at 50 tickets and short fields; use `hive-read-state` children only for several independent collections at once. Search the tracker for an existing ticket before proposing one.
   - For a failure whose cause is not yet shown, list the plausible causes and check each one, in parallel when independent; report the ones ruled out with their evidence.
   - Name each system the decision touches and whether you searched it.
   - Open the source of each claim that decides the answer instead of recalling it. Resolve conflicting sources by date, version, scope and direct observation; missing delivery evidence means delivery is not verified, not that it did not happen.
5. **Close the what.** Report the answer, its evidence and what remains uncertain, and play back what you understood, marking what the user said apart from what you propose. Then:
   - When the user asked to analyze, explain, review or diagnose without asking for a change, end at the findings with the block of «Questions and routes», recommending no flow; when the findings confirm a defect in the current project, list the reply that fixes it among its options. When the user picked a ticket for research without asking for a change, and the findings confirm work it asks for, end instead with the next-step block toward `flow-plan`, or toward `flow-build` when that work is a small understood change (the global guidance's «Proportionality»), and never continue into build in the same turn; when they confirm no work, end at the findings with that block.
   - When the user asked for a change, not whether or how to make one, and its how leaves no decision for them and could be finished from the request and `git diff` alone, continue into `flow-build` in the same turn.
   - Otherwise, end with the route to `flow-plan` and its reason.

   Creating a ticket is a single action, offered as a reply or asked, never part of a route. The user's reply authorizes only the flow it names, or the single action it answers.

## Assess a chosen ticket

When the user picks a ticket, add to the report its **impact** (what uses what it changes, found by searching its literal values and direct callers, and whether it touches production), **complexity** (the files, modules and contracts it changes), **readiness** (settled criteria and decisions, and the access or answers missing) and **dependencies** (the tickets it waits on or blocks). A ticket whose criteria are missing, that names more than one change, or that the specs contradict is not ready: settle its scope here, in rounds, before step 5.
