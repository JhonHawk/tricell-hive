---
title: Research
description: Investigate a question and get an answer backed by evidence, without changing your project.
---

Research is for the moments when you need to know something before you decide anything. The agent reads your code, documents and project state, then reports what it found and how sure it is.

## When to use

Use Research when the next step depends on facts you do not have yet.

- Something is failing and nobody knows why.
- You want to compare two ways of solving a problem.
- You need the real state of a project: what is done, what is open, and where the records disagree.
- You are asking whether something can be done, such as "can you fix this record?". That is a question, so the agent looks first and waits for your choice.

Example requests:

```text
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

For large questions the agent may hand reading work to helper agents, then check their claims itself before reporting. If the findings point to work that no ticket covers, it asks whether you want a ticket. It opens one only after you agree.

When the findings suggest work you could start, the agent names the route it recommends, Plan for larger work or Build for a small, well understood change.

## What it does not do

- It does not change code, tickets, documents or deployments, unless you separately ask for that change.
- It does not save a research document unless you ask for one.
- It does not treat a documented feature as proof that your setup has it, and it does not treat "no search results" as proof that something is missing.

## Related flows

- Continue to [Plan](/flows/plan/) when the answer shows a change that spans several steps or needs decisions.
- Continue to [Build](/flows/build/) when the answer shows a small change you are ready to approve.
- You can also stop here. Research does not require anything afterward.

## FAQ

**Does research give the agent permission to implement the fix?**
No. A question, an investigation or a recommendation is not a go-ahead. You still choose the route and approve the work.

**Will it write a report file?**
Only if you ask for one or accept a proposal to keep it. By default the findings stay in the chat.

**Why does the agent sometimes use helper agents?**
Reading long logs, installed packages or many documents fills the main conversation. Helpers read that material and return only what matters. The main agent still verifies the important claims.

**What if two sources disagree?**
The agent compares dates, versions and what each source actually covers. If it cannot settle the conflict, it tells you what evidence would.

**Where is the exact rule?**
See the source linked below for the full text.

Source: [content/skills/flow-research/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/SKILL.md)
