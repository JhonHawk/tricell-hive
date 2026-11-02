---
title: Research
description: Investigate a question and get an answer backed by evidence, without changing your project.
---

Research is for the moments when you need to know something before you decide anything. The agent reads your code, documents and project state, then reports what it found and how sure it is.

## When it starts

Use Research when the next step depends on facts you do not have yet.

- Something fails and the cause is unknown, or a bug is reported without a request to fix it.
- You want to compare two ways of solving a problem, or settle two sources that disagree.
- You ask whether or how something could be done, such as "can you fix this record?". That is a question, so the agent looks first and waits for your choice.
- You pick a ticket to work on and want it assessed first.

A single lookup that answers the question needs no investigation: the agent reports the answer with its source and any doubt. For the project's status or backlog, use [Status](/flows/status/).

Example requests:

```text wrap
Investigate why this endpoint times out. Read only; report evidence and options.
Compare the two queue libraries we could use and tell me which fits this project.
Assess COM-833 before I start it.
```

## What it asks you

Very little. It asks only about intent that would change the conclusion. Questions follow the format in the [shared guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md): a bold context line, the options with the recommended one first, and a line saying what the agent waits for.

When a ticket you picked has open scope, such as missing criteria or more than one change in it, the agent settles those scope decisions with you during research, before it recommends a route.

## What it does

- Splits large questions across helper agents, then opens the cited sources of every claim that decides the answer before reporting.
- For a failure, lists the plausible causes and checks each one, and tells you which it ruled out and with what evidence.
- Resolves conflicting sources by date, version, scope and direct observation. Missing delivery evidence means delivery is not verified, not that it did not happen.
- For a chosen ticket, writes an assessment of its impact, complexity, readiness and dependencies.
- Publishes the stage to a running local viewer when one applies; if the viewer is unavailable, the investigation continues and the gap is reported. See the [walkthrough skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md).

It does not change code, tickets, documents or deployments, and it does not treat "no search results" as proof that something is missing.

## How it ends

The answer comes in the conversation: what was concluded, the evidence, and what is still uncertain. When a long or shareable result would help, a report page is offered as one option.

When the findings leave work you could run next, the report ends with a route: [Build](/flows/build/) for a small change that is understood, [Plan](/flows/plan/) for larger work, with the reason, plus how to keep investigating or stop. Your reply starts only the flow it names; research never authorizes the fix.

## Related flows

- [Status](/flows/status/) reports the backlog; Research takes the one ticket you pick from it.
- [Plan](/flows/plan/) and [Build](/flows/build/) are the two routes out of the findings.

## FAQ

### Does research give the agent permission to implement the fix?
No. A question, an investigation or a recommendation is not a go-ahead.

### Will it write a report file?
Only if you ask for one. By default the findings stay in the chat.

### Where is the exact rule?
Source: [content/skills/flow-research/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/SKILL.md)
