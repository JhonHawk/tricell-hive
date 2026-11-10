---
title: Research
description: The entry for every request to do work, which settles what you need and then answers, builds or plans; also investigates a question with evidence.
---

Research settles what the work is; [Plan](/flows/plan/) settles how. Every request to do work enters here, so the agent understands what you need before it builds anything, and it is also where you ask a question and get an answer backed by evidence. It reads your code, documents and project state, then reports what it found and how sure it is.

## When it starts

It starts on any request to do work, with no command to type, and when the next step depends on facts you do not have yet.

- A request that already says everything that matters, such as adding a user to the repository, is checked for blockers without questions: an action that changes no repository file is done and reported, and a change to the repository continues into [Build](/flows/build/) in the same turn when its how leaves you no decision, or goes to [Plan](/flows/plan/) otherwise.

- Something fails and the cause is unknown, or a bug is reported without a request to fix it.
- You want to compare two ways of solving a problem, or settle two sources that disagree.
- You ask whether or how something could be done, such as "can we fix this record?" or "how would we fix it?". That is a question, so the agent looks first and waits for your choice. A polite request that names the action and its target, such as "can you add this user to the repository?", counts as a request for that change.
- You pick a ticket to work on and want it assessed first.

A single lookup that answers the question needs no investigation: the agent reports the answer with its source and any doubt. For the project's status or backlog, use [Status](/flows/status/).

Example requests:

```text wrap
Investigate why this endpoint times out. Read only; report evidence and options.
Compare the two queue libraries we could use and tell me which fits this project.
Assess COM-833 before I start it.
```

## What it asks you

Very little, and only after it has read what the project already answers. It plays your request back: what you want, for whom and why, with any conflict, false premise or cheaper path it found. When something that changes the result is still open, it asks one open question per message, in the text, with the reason and the option to tell it to just start. Choices between options use the cards of the [shared guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md): a bold context line, a `Situation:` line, the options with the recommended one first, each with an `If you choose it:` line, an `I recommend` line, and a line saying what the agent waits for.

When a ticket you picked has open scope, such as missing criteria or more than one change in it, the agent settles those scope decisions with you during research, before it closes the what.

## What it does

- Writes the aspects of the question, their sources and which go to a helper agent in its message before the first read, splits across helper agents each aspect that needs more than five large reads (whole files, documentation pages, logs or unfiltered listings; filtered queries stay in the main conversation), and hands over the rest of an aspect when its sixth large read in the main conversation still has no answer; uses the host's built-in explorer only to locate files and checks what it returns, names each system the decision touches and whether it searched it, then opens the cited sources of every claim that decides the answer before reporting.
- For a failure, lists the plausible causes and checks each one, and tells you which it ruled out and with what evidence.
- Resolves conflicting sources by date, version, scope and direct observation. Missing delivery evidence means delivery is not verified, not that it did not happen.
- For a chosen ticket, writes an assessment of its impact, complexity, readiness and dependencies.
- When the work brings a screen that is new, or one that adds, removes or moves a region, and nothing fixes its composition, and the host has a connection that can write to Figma, recommends [Figma design](/flows/figma-design/) in its next-step block, also after assessing a ticket. It draws nothing itself. It makes no recommendation for a change that keeps the screen's layout, such as a button, a copy, a field or a column, nor while work runs [unattended](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/unattended-delegation/SKILL.md). A request whose deliverable is the design itself, such as replicating a screen in a design file, is sent to that stage with the same next-step block.
- Searches for an existing ticket before proposing one. It publishes nothing to the viewer: the [viewer](/viewer/) session starts at the first UI capture of a Build or Verify, with a one-paragraph summary of the work.

It does not change code, tickets, documents or deployments and it does not treat "no search results" as proof that something is missing.

## How it ends

The answer comes in the conversation: what was concluded, the evidence, and what is still uncertain, with what you said kept apart from what the agent proposes.

- If you asked to analyze, explain or find out why, it ends at the findings with a footer that asks how to continue and lists the replies you can give, each with a letter and the best one starred, with no command to type and never a flow you did not ask for. When it confirmed a defect in your project, fixing it is one of those replies.
- If you picked a ticket and the findings confirm work it asks for, it ends with the command to start [Plan](/flows/plan/), or [Build](/flows/build/) when the fix leaves you no decision, and waits for your reply before starting either.
- If you asked for a change whose how leaves you no decision and can be finished from your request and the `git diff`, it continues into Build in the same turn.
- Otherwise it ends with the command to start [Plan](/flows/plan/) and the reason, because the how has a decision for you or the work could not be resumed from the request and the diff.

Whenever it ends with the command to start Plan or Build, or continues into Build, one line estimates the scope: about how many files and which areas the work touches, and which stages and checks it will pass through, such as the plan, per-task verification, a review, CI, and a promotion only when you asked for one. You can narrow it in your reply; the agent then gives the revised line and the same ending again.

Your reply starts only the flow it names. No flow needs a ticket. A single action that needs no flow, such as opening a ticket for work that will be done later or by someone else, is asked directly ("Do you want me to create the ticket?") instead of being sent through a flow command.

## Related flows

- [Status](/flows/status/) reports the backlog; Research takes the one ticket you pick from it.
- [Figma design](/flows/figma-design/) is the optional step it recommends before Plan for a new or recomposed screen.
- [Build](/flows/build/) follows in the same turn when the how is settled; [Plan](/flows/plan/) follows when it has a decision for you.

## FAQ

### Does research give the agent permission to implement the fix?
No for a question, an investigation or a recommendation. A request to do work, once its what is clear and its how leaves you no decision, does continue into Build.

### Will it write a report file?
Only if you ask for one. By default the findings stay in the chat, and nothing goes to the viewer.

### Where is the exact rule?
Source: [content/skills/flow-research/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/SKILL.md)
