---
title: Plan
description: Turn a larger change into settled decisions and small tasks that each come with a way to check them.
---

Plan is for work that is too big or too uncertain to start typing. The agent sorts out the choices with you first, so whoever builds it does not have to guess.

## When to use

Use Plan when a change touches several places, has trade-offs you should decide, or will be handed to another session or agent.

Example requests:

```text
Plan the API change, including compatibility and acceptance checks.
Plan moving our uploads to the new storage service. Ask me before choosing anything that affects existing data.
Break this feature into tasks I can review one at a time.
```

A small, clear edit does not need a plan. Ask for it directly through [Build](/flows/build/).

## What it produces

A plan that someone else could follow. It covers:

- the outcome you want and what is in and out of scope;
- the decisions you made, with the alternatives that mattered;
- what changes for the pieces that talk to each other, such as inputs, outputs and data;
- tasks grouped into results you can check separately, each with the command or observation that proves it and what you should see;
- acceptance criteria that are numbered and can be tested.

When the plan is saved, independent reviewer agents read it before the agent calls it ready, one per affected area. It also asks how you want the work delivered through Git and how it should be reviewed.

A short plan can live in the conversation. When work will span sessions, or you ask for a saved plan, the agent writes it as a change folder in your project's specs directory. That folder stays uncommitted while the work runs. [Close](/flows/close/) versions it at the end.

## What it does not do

- It does not write the implementation.
- It does not approve commits, pushes, pull requests, merges or deployments. Those stay your call.
- It does not hide open technical doubts inside later tasks. A doubt that could sink the approach is investigated first or marked as blocking.
- It does not ask you to keep two versions of an interface running unless evidence shows a real need.

## Related flows

- Start from [Research](/flows/research/) when you still need facts before planning. The agent reuses sound findings and rechecks the ones that may have changed.
- Hand the finished plan to [Build](/flows/build/), pointing it at the change folder or plan path.
- [Close](/flows/close/) archives the change folder when the work is integrated.

## FAQ

**Do I need a plan for a small change?**
No. A small change you already understand can go straight to Build. Keep its checks in the conversation.

**Does a ready plan mean the agent may start building?**
No. The plan states the next step, but starting the work still needs your instruction, unless you already gave it.

**Will the agent ask me many questions?**
It asks about choices that change the outcome, and groups independent questions into one exchange. Facts it can find in the code, it looks up itself.

**Why are the criteria numbered?**
Each one is checked later, so every criterion must be testable and covered by at least one task.

**Where is the exact rule?**
See the source linked below.

Source: [content/skills/flow-plan/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md)
