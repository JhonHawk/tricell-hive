---
title: Research
description: Investigate a question and get an answer backed by evidence, without changing your project.
---

Research is for the moments when you need to know something before you decide anything. The agent reads your code, documents and project state, then reports what it found and how sure it is.

## When to use

Use Research when the next step depends on facts you do not have yet.

- Something is failing and nobody knows why, or you report a bug without asking for the fix yet: the agent confirms the cause first, then asks in one round whether to track it and how to continue.
- You want to compare two ways of solving a problem.
- A single quick lookup answers it: the agent reports the answer with its source and any doubt, with no helper agents except one that lists tickets from the tracker, so that listing stays out of the conversation, no offer to save a document and no route question.
- You need the real state of a project: what is done, what is open, and where the records disagree. The status report also counts how much of the work integrated since the last promotion, or in the last 30 days when there is none, had to be redone: reverts of that work, its tickets reopened since, and bug tickets that point to it. These are plain counts, each item listed with its link, not rates or targets; the rule is in the [backlog report reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/references/backlog-report.md).
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

When the answer recommends removing, closing, disabling, migrating or renaming something that runs in production, moving it to another server or process, or changing untested production behavior, the agent first checks what depends on it, only as far as what it finds calls for. Anything it cannot see from the code it marks as unverified, with how to check it and how to undo the change. The steps are in the [change impact reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/references/change-impact.md).

For large questions the agent hands reading work to helper agents. Each helper says where every key claim comes from and what it did not check; the agent opens the cited lines of the claims that decide the answer before reporting. Each helper is told which decision its findings feed and every angle you asked about, and it stops when it has enough for that decision. Before recommending a way forward, the agent checks that the findings answer every part of your question, and it confirms its own guesses that decide the recommendation with a quick local check when one exists; when it cannot, it recommends investigating further instead. Long ticket, CI or deployment status is collected by a helper too. It opens a ticket only after you choose an option that names one.

When you ask it to review a pull request outside the current work, such as a colleague's, Research reads it without changing it, reports its findings by priority with their evidence and a recommended verdict for each pull request, and posts nothing on it until you answer. The steps are in the [pull request review reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/references/pr-review.md).

The global core keeps essential boundaries; Research links directly to the
canonical questions, reporting and delegation procedures when needed. Reporting
applies to findings, relayed results, explanations and messages with external
links or IDs; greetings and short direct answers need no reference read.
Optional context delivery covers only the
explicit OpenCode plan-review and harness-review roles, not intent detection or
automatic loading for every research task.

When the findings concern work you could start, the same answer ends with the route question. Its options are the ways to act on the findings, each naming the flow it starts (Plan for larger work, Build for a small, well understood change), plus keeping the investigation going and stopping at the findings. Every route question, including one asked again after new findings, offers at least one way to act on them: opening a ticket and stopping is never the only one. The recommended option comes first, with its reason. When the ways to act would build different amounts, one of them is the smallest that solves the problem you stated, and a larger recommended one says what the smallest leaves unsolved. A ticket appears in the options only when it helps: an option that acts on the findings can open one when your project tracks its work in a tracker, and opening a ticket and stopping is offered when the work will wait, because it is large or belongs to someone else, or when you ask for one. There is no separate ticket question; choosing an option that names a ticket authorizes opening it. If the findings fully answer the question and there is no work to run, there is no route question. When you ask what to take next from the backlog, each option is a ticket or group instead, starting the flow its readiness sets: a ticket whose scope does not fit what you asked for, such as one session, starts research to cut it, and two options never differ only in the flow they start for the same ticket. The rule is in the [backlog report reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/references/backlog-report.md).

When the next steps would alternate investigating and fixing toward one measurable goal, such as measuring a tool, correcting it and measuring again, the route question can also offer an iterative route. Its option tells you the goal and how it is measured, how many rounds the agent proposes, and when it stops. If you choose it, the agent runs it as a loop: it keeps the goal, its measurement, the round limit and each round's result in a loop state file on disk, so a compaction or a new session can resume it, keeps going round after round without reporting each one, as long as every step stays local to your project. It stops and asks when the goal is met, the rounds run out, two rounds in a row bring no measurable improvement and one change of approach does not help either, a step needs something the route does not cover (such as a commit or an installation), a round grows into work that needs a plan, the scope changes, or a decision is yours; when you delegated the work unattended, it takes that decision itself and reports it. Any long-running work that repeats cycles follows the same contract, in the [loops reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/loops.md). The exact conditions are in the [skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/SKILL.md).

A new change you ask for while the research runs joins the route question as a route of its own. When you add topics while the research runs, the agent keeps a list of the open topics, tells you in one line how it classifies each new one, closes none without a destination, and names them all, with a proposed destination, in its findings, its completion report and the route question. The rule is in the [guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).

Every question follows the [questions reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/questions.md): the message opens with a bold context line that says which piece of work is asking, ends with a line that marks it as waiting for you, and puts the options in a question card when your tool shows one that waits for your answer, or in the message otherwise. A plain yes-or-no confirmation of an action the message describes can be that last line itself, without options or a card. A dismissed card is not an answer, so the agent asks again in text. A question whose options depend on the answer to another one is asked in a later turn, after that answer. When an answer is free text, invalidates something or sends the work back, the agent says in one line what it changes before it acts; only picking an offered option needs no such line. When your answer leaves a selection to the agent, such as which findings to fix now, it lists what goes in and what stays out and asks you to confirm first.

## What it does not do

- It does not change code, tickets, documents or deployments, unless you separately ask for that change.
- It does not save a research document unless you ask for one.
- It does not treat a documented feature as proof that your setup has it, and it does not treat "no search results" as proof that something is missing.

## Related flows

- Continue to [Plan](/flows/plan/) when the answer shows a change that spans several steps or needs decisions.
- Continue to [Build](/flows/build/) when the answer shows a small change you are ready to approve.
- If a finding changes what you asked, the iterative route stops; see [changes and loops](/flows/overview/).
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
