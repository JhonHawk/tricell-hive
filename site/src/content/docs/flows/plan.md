---
title: Plan
description: Turn a larger change into settled decisions and small tasks that each come with a way to check them.
---

Plan is for work that is too big or too uncertain to start typing. The agent sorts out the choices with you first, so whoever builds it does not have to guess.

## When to use

Use Plan when a change touches several places, has trade-offs you should decide, or will be handed to another session or agent.

Example requests:

```text wrap
Plan the API change, including compatibility and acceptance checks.
Plan moving our uploads to the new storage service. Ask me before choosing anything that affects existing data.
Break this feature into tasks I can review one at a time.
```

A small, clear edit does not need a plan. Ask for it directly through [Build](/flows/build/).

## What it produces

A plan that someone else could follow. It covers:

- the outcome you want and what is in and out of scope;
- the decisions you made, with the alternatives that mattered;
- in-scope questions that cannot be stated precisely yet, kept apart from the decisions and the exclusions, with no task built on them until they can;
- what changes for the pieces that talk to each other, such as inputs, outputs and data;
- when the work removes, closes, disables, migrates, renames or moves something that runs in production, or changes untested production behavior, what depends on it today and what is still unverified, checked only as far as the situation calls for, as the [change impact reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/references/change-impact.md) sets;
- tasks grouped into results you can check separately, each with the command or observation that proves it and what you should see;
- when the work replaces or migrates a writer, parser, query or system and the old path can still run, the old-against-new comparison that the [verification reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/references/verification.md) sets, so the plan checks the new path against the old one;
- acceptance criteria that are numbered and can be tested;
- when the criteria are numbered, the checkpoints at which Build verifies the tasks: the plan groups tasks into checkpoints and closes a group after each task that is risky to get wrong, chosen by the risk criterion in the [plan criteria and markers](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/plan-criteria-and-markers.md), so a long plan is not verified after every task.

When the plan is saved, independent reviewer agents read it before the agent calls it ready, one per affected area. A reviewer finding whose fix has a conventional default and changes no scope, contract, criterion or decision of yours is applied without asking and noted in the plan's progress section with its reason. A plan with a screen where you enter or choose data also gets a reviewer that walks your task looking for data you would have to enter twice or steps you could skip. A plan that edits agent guidance also gets a guidance-consistency reviewer, which searches the guidance for sentences that would still contradict a rule the plan changes and flags an option or task that contradicts a standing rule; a decision you recorded never counts as a violation, as the [plan review reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/plan-review.md) sets. It also asks how you want the work delivered through Git and how it should be reviewed, together with the last design questions that do not depend on those answers.

The skill links directly to the procedures for questions, project settings,
change records and plan review. Reading one reference does not start its whole
owning flow. The optional OpenCode loader can supply a generated plan-review
snapshot for the explicit reviewer role without runtime reference-file reads;
refresh requires manager regeneration and plugin reload. It does not prepare the
plan or choose its scope.

A short plan can live in the conversation. When work will span sessions, or you ask for a saved plan, the agent writes it as a change folder in your project's specs directory, starting with your request quoted word for word, plus the acceptance criteria of any ticket it names. Before it decides which current requirements the change alters, it checks the specs for any requirement the change could contradict and notes that check in the folder. That folder stays uncommitted while the work runs. [Close](/flows/close/) versions it at the end.

## What it does not do

- It does not write the implementation.
- It does not approve commits, pushes, pull requests, merges or deployments. Those stay your call.
- It does not hide open technical doubts inside later tasks. A doubt that could sink the approach is investigated first or marked as blocking.
- It does not ask you to keep two versions of an interface running unless evidence shows a real need.

## Related flows

- Work that [Build](/flows/build/) sends back to plan keeps its commits on the work branch and the ticket in progress; the plan records that work as done.
- Start from [Research](/flows/research/) when you still need facts before planning. The agent reuses sound findings and rechecks the ones that may have changed.
- Hand the finished plan to [Build](/flows/build/), pointing it at the change folder or plan path.
- [Close](/flows/close/) archives the change folder when the work is integrated.

## FAQ

### Do I need a plan for a small change?
No. A small change you already understand can go straight to Build. Keep its checks in the conversation.

### Does a ready plan mean the agent may start building?
No. When the plan is ready, the agent summarizes it in the conversation: the objective, the text of each acceptance criterion and the requirements that change, the decisions in force and the minor calls it took without asking with their reasons, each task in one line, the checks, how it will be delivered and reviewed, the `UI verification` value for a visible interface change, what is still open, and where the full plan is saved. It then asks whether to start Build now or stop with the plan ready, giving the plan's estimated size on the Build option, and recommending Build unless something concrete blocks it, which it names; asking for a plan is not such a block, and it never suggests building the plan in a later session unless you said so. Choosing Build approves those criteria as the contract the work is checked against; if you correct one in your own words instead, the agent revises the plan and asks again with the revised criteria. Starting the work still needs your answer, unless you already gave the instruction. When the session already holds more than about half of the model's context, that message adds one line suggesting you run your CLI's compaction command, such as `/compact`, before answering, since the plan is on disk and compacting while the prompt cache is warm costs less; it is a suggestion, not a question, as the [context reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/context.md) sets.

### Will the agent ask me many questions?
It asks about choices that change the outcome, and groups independent questions into one exchange. Facts it can find in the code, it looks up itself. A question whose options depend on the answer to another one, such as one that changes the course of the plan or assumes a recommended option was chosen, is asked in a later turn, after that answer; the same round carries only questions whose options hold with any answer to the others. The planning round also settles how the work is delivered and reviewed and, for a visible interface change, the `UI verification` level that sets which independent checks run on the built result before you validate it, unless you already answered them earlier in the session for that repository. For each criterion checked outside your machine, such as in QA, the plan names the environment and the test data, access and other effects that check needs there, so they are known up front. Authorizations the plan already foresees for later steps are asked in that same round, or on their own at that point when an earlier answer or your request already settled delivery and review, so the build does not stop for them later; one you leave unanswered stays unauthorized. When its options would build different amounts for what you asked, one of them is the smallest that solves the problem you stated, and a larger recommended option says what that smallest one leaves unsolved. Options that start large or chained work give their size as the number of tasks and what that count rests on, with no time estimate. Those rules are in the [option scope and size reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/option-scope.md); the question format and the grouping rules are in the [questions reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/questions.md); the classes of effects beyond your machine are in the [effect authorization reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/effect-authorization.md).

Topics you add while the plan is written go on a list of open topics. The agent tells you in one line how it classifies each, closes none without a destination, and names them all in its findings and completion reports and in the route and close questions. The rule is in the [guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).

### What happens when a decision changes the design?
The agent does not save it for later. It shows the decision at once, under the context line that opens every decision, result or closing message, with an explanation and the recommended option, asked like any other question. The message ends with a line marking that it waits for you, and work that depends on the decision waits for your answer: a consequential design question is asked before any task that depends on it is written. When your answer is free text, invalidates something or sends the work back to research, the agent says in one line what it changes, and if tasks already written become invalid it updates the plan; only picking an offered option needs no such line. Independent questions are still grouped. The exact format is in the [questions reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/questions.md#pause-and-handle-the-answer).

### What if the plan has to change while the work is being built?
The agent drafts the change in the conversation, names what it touches (your request, a criterion, a decision, a task or a requirement delta) and then edits the saved plan itself, keeping those parts consistent with each other. A change that alters a criterion, a decision or a contract goes back to the plan reviewers once; a smaller one the agent reconciles and records. If the change alters what you asked for, it becomes a new change rather than an edit of this one.

### What if I ask for something new while the plan is open?
The agent first classifies the request as a correction, an added task, a new change or a separate topic, and says where the work goes. [Changes and loops](/flows/overview/) lists where each flow returns the work, and the classification lives in the [scope changes reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/scope-changes.md).

### Why are the criteria numbered?
Each one is checked later, so every criterion must be testable and covered by at least one task.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-plan/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md)
