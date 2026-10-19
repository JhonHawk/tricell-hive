---
title: Research
description: Investigate a question and get an answer backed by evidence, without changing your project.
---

Research is for the moments when you need to know something before you decide anything. The agent reads your code, documents and project state, then reports what it found and how sure it is.

## When to use

Use Research when the next step depends on facts you do not have yet.

- Something is failing and nobody knows why, or you report a bug without asking for the fix yet: the agent confirms the cause first, then asks in one round whether to track it and how to continue.
- You want to compare two ways of solving a problem.
- You need the real state of a project: what is done, what is open, and where the records disagree.
- You are asking whether something can be done, such as "can you fix this record?". That is a question, so the agent looks first and waits for your choice.

Example requests:

```text wrap
Investigate why this endpoint times out. Read only; report evidence and options.
Compare the two queue libraries we could use and tell me which fits this project.
What is the real delivery state of the billing work? Check tickets, code and pull requests.
```

## What it produces

An answer in the conversation. It normally tells you:

- what the agent concluded, and where it looked (file paths or links);
- what it observed directly, and what it only inferred;
- the options it sees and what each one costs;
- what is still uncertain;
- what decision or next step the answer makes possible.

For large questions the agent hands reading work to helper agents. Each helper says where every key claim comes from and what it did not check; the agent opens the cited lines of the claims that decide the answer before reporting. Long ticket, CI or deployment status is collected by a helper too. If the findings point to work that no ticket covers, it asks whether you want a ticket. It opens one only after you agree.

The global core keeps essential boundaries; Research links directly to the
canonical questions, reporting and delegation procedures when needed. Reporting
applies to findings, relayed results, explanations and messages with external
links or IDs; greetings and short direct answers need no reference read.
Optional context delivery covers only the
explicit OpenCode plan-review and harness-review roles, not intent detection or
automatic loading for every research task.

When the findings concern work you could start, the same answer ends with the route question. Its options are the ways to act on the findings, each naming the flow it starts (Plan for larger work, Build for a small, well understood change), plus keeping the investigation going and stopping at the findings. The recommended option comes first, with its reason. The message ends with a fixed last line, "Waiting for your decision:" (in Spanish sessions, «Esperando tu decisión:»). A ticket offer can join the same round as its own numbered question. If the findings fully answer the question and there is no work to run, there is no route question.

When the next steps would alternate investigating and fixing toward one measurable goal, such as measuring a tool, correcting it and measuring again, the route question can also offer an iterative route. Its option tells you the goal and how it is measured, how many rounds the agent proposes, and when it stops. If you choose it, the agent reports each round in one line and keeps going, as long as every step stays local to your project. It stops and asks when the goal is met, the rounds run out, a round brings no measurable improvement, a step needs something the route does not cover (such as a commit or an installation), a round grows into work that needs a plan, the scope changes, or a decision is yours. The exact conditions are in the [skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/SKILL.md).

The message opens with a bold context line, stage · ticket · progress, so you can tell which piece of work is asking, and a question card opens its question with the stage and ticket. If your tool shows a question card that waits for your answer, the message keeps the explanation plus one summary line per question with the recommended option, and the options are in the card. Without such a card, or after you dismiss it or it times out, the options are written in the message. A dismissed card is not an answer: the agent asks the question again in text.

After each answer, the agent says in one line whether it settles the question, opens the next round, or returns the work to an earlier flow, and what that invalidates.

## What it does not do

- It does not change code, tickets, documents or deployments, unless you separately ask for that change.
- It does not save a research document unless you ask for one.
- It does not treat a documented feature as proof that your setup has it, and it does not treat "no search results" as proof that something is missing.

## Related flows

- Continue to [Plan](/flows/plan/) when the answer shows a change that spans several steps or needs decisions.
- Continue to [Build](/flows/build/) when the answer shows a small change you are ready to approve.
- You can also stop here. Research does not require anything afterward.

## FAQ

### Does research give the agent permission to implement the fix?
No. A question, an investigation or a recommendation is not a go-ahead. You still choose the route when the agent asks, and approve the work.

### Will it write a report file?
Only if you ask for one or accept a proposal to keep it. By default the findings stay in the chat.

### Why does the agent sometimes use helper agents?
Reading long logs, installed packages or many documents fills the main conversation. Helpers read that material and return only what matters. The main agent still verifies the important claims.

### What if two sources disagree?
The agent compares dates, versions and what each source actually covers. If it cannot settle the conflict, it tells you what evidence would.

### Where is the exact rule?
See the source linked below for the full text.

Source: [content/skills/flow-research/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/SKILL.md)
