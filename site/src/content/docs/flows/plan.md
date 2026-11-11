---
title: Plan
description: Turn a larger change into settled decisions and small tasks that each come with a way to check them.
---

Plan settles the how. [Research](/flows/research/) has already settled what you need; Plan sorts out the remaining choices with you first, so whoever builds it does not have to guess.

## When it starts

Research sends a request here only when its how has a decision that is yours, or when the work cannot be finished from your request and the `git diff` alone, such as work handed to another session or agent. You can also ask for a plan directly. A change whose how leaves you no decision needs no plan; Research continues it into [Build](/flows/build/).

Example requests:

```text wrap
Plan the API change, including compatibility and acceptance checks.
Plan moving our uploads to the new storage service. Ask me before choosing anything that affects existing data.
Break this feature into tasks I can review one at a time.
```

## What it asks you

The agent reads the code, your project settings and the existing specs first, then asks only for the decisions that are yours. Facts it can find, it looks up itself.

- When the how has more than one sensible way, two or three approaches in a message of their own before the tasks, each with its tradeoff and one recommended, preferring what the codebase, the standard library or an installed dependency already does over new code or a new dependency; you choose, and they are recorded in `design.md`. When only one way is sensible, the agent says so and why.
- Choices that change the outcome, asked in rounds and in the shared question format: each decision carries a `Situation:` line, an `If you choose it:` line under every option, and an `I recommend` line, and the round says how many decisions are still to come. A change to what an interface sends or receives counts as a contract change, and you decide whether it stays compatible.
- For a changed screen, the layout and its states, shown as before-and-after sketches you can choose between; when the Figma design stage left an approved screen example, the plan records it in place of the after sketch, adds the task that writes each convention you agreed to record in `DESIGN.md`, instead of drawing that region again, and does not settle that composition with a design helper. The business rules you approved in that stage become decisions of the plan, and a rule still waiting for your answer is asked before the plan is ready. Before that, the agent loads the project's design system through the [UI criteria skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-ui-criteria/SKILL.md): the library, its installed version and its customization layer, recorded in a `DESIGN.md` at the root of the app that owns the screens. When that file is missing or out of date, the plan includes a task that writes or refreshes it, delivered with the change and shown as a diff; outside a plan, the agent writes it only when you ask.
- When the chosen approach reads or changes a system outside the repository, such as a cloud account, an organization or a third-party service, the facts the plan depends on, such as what an account allows or what a branch holds, before the tasks are written. Each external fact is measured with a read-only command and recorded with its source; with two or more systems, one read-only helper per system measures in parallel, and the agent accepts a fact only with the command and an excerpt of its output, checked against the provider's current documentation when it is about a capability, a limit or a permission. A fact that cannot be checked that way is listed as not measured, with the task that meets it first.
- Once the plan is written, how the work is delivered and which code review runs, unless you already answered in this session and repository and the new work does not call for a validation stop or a review your answer lacks. The recommended delivery follows the kind of work: automatic when its checks, review and verification can judge it, such as infrastructure or CI; interactive when it has a screen, copy or flow only you can judge. A one-off operation that leaves nothing in the repository, such as a database dump, gets neither question.
- In the same round, one decision per class of effect the tasks foresee outside the repository: a kind of action on a named target, such as changing access policies in an account, with its spending ceiling. A class you approve covers each effect of that class that the plan's tasks take, within the ceiling, and does not carry over to other work; a class you refuse is not taken and not asked again. Deploying to production and anything irreversible or that interrupts a service people use keep their own question. A plan with no such effects adds no question.

The rules for questions and for delivery are in the [shared guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md); the measurement of external facts and the question per class of effect are in the [outside the repository reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/outside-the-repository.md).

## What it produces

Plan publishes nothing to the [viewer](/viewer/); the plan documents stay in their change folder in the repository, and you review them there or in the message.

A plan stays in the message only when the work fits one session on this machine: it spans no other session, waits on nothing external and is not handed to helpers. Such a plan carries its decisions, criteria, prerequisites and tasks in the message, and the agent keeps a checkpoint file in the host's temporary folder with the decisions, the pending tasks and anything it created outside the repository, so a compacted session can pick up where it left off. The file is removed at Close. The agent writes a change folder in your specs directory when the plan fails that test, such as work that spans sessions or is handed to helpers, when you ask for one, when its deliveries need coordination, or when it changes a contract, stored data, security, money, concurrency or a dependency. A one-off operation always keeps its plan in the message; when one of those conditions applies, or you agree to add files to the repository for it, it is handled as a regular change. A plan in the message that changes a behavior your specs describe includes a task that updates them. Without a ticket, the folder is named by a short slug; a plan never waits for a ticket to exist.

The change folder is copied from the [templates](https://github.com/JhonHawk/tricell-hive/tree/master/content/skills/flow-plan/assets/change-folder):

- `proposal.md`: your request quoted, your decisions in plain words, the scope, numbered acceptance criteria, and what the build needs from outside the repository, including the measured external facts and one line per class of effect with your answer, under `## Effects`.
- `tasks.md`: small tasks, each saying which criteria it closes, what it depends on and how to check it.
- `design.md` and spec deltas, only when the change has design decisions or alters a behavior, contract, permission or stored data.

Every criterion is one observable behavior, with the observation that would show it false today and an `Observed in:` place where its real effect shows: for a screen or an API the project serves, the candidate served locally with the real build and seeded data, never a mock or emulation of a service the project does not own. Until a walk observes a criterion in that place it is reported as not verified: Verify's walk closes the ones it reaches, and a criterion that names a deployed environment waits for the functional walk of [Deploy](/flows/deploy/), which naming the environment does not ask for. A criterion that is already true, or that only restates your request, is rewritten or dropped. When the plan has a change folder, the agent first runs the [coherence check](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/scripts/check_change.py), which reports criteria no task closes and `Closes:` lines naming criteria that do not exist, and then one independent reviewer checks the folder before it is ready.

When the ticket or your request asks for the change in a named environment, the plan includes the promotion to that environment as a task run through [Deploy](/flows/deploy/). Only your own request in the conversation counts as the instruction to deploy there; when only the ticket asks for it, the task asks you first, as Deploy does. Otherwise the plan has no promotion task, and promotions are grouped and run when you ask.

A change folder stays on disk while the work runs and is committed with the code it plans: on the work's branch before the push when the specs live in the same repository, or in the specs repository when the work reaches its completion point. When that repository declares `Delivery: direct-base`, the folder is edited in its main checkout, with no branch or worktree, and a paused plan is committed there. Tasks still open after the merge do not hold it back; [Close](/flows/close/) records them and archives the folder.

## Revising a plan in progress

When a written plan changes during planning or the build, the agent updates in one pass every place the change folder states the change: proposal, design, the tasks' text, `Closes:` and verification, and the spec delta. It marks the new item with `Replaces:`, or a dropped criterion with `Retired:`, reopens the tasks the revision invalidates, and runs the check, which also reports a criterion that names no `Observed in:` place, until it is clean. When a criterion or the delta changed, the plan reviewer (`hive-review-plan`) reviews it once more, one round per revision. The revision is recorded in the progress section and the agent tells you in one line. The steps are in the [flow-plan skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md).

## How it ends

The agent summarizes the plan in the message: the objective with the files and areas recounted, and what changed since Research when the count moved by more than a quarter, each criterion, your decisions, the tasks on one line each, what is still missing, the checks and the delivery mode. It then gives the Build command as the next step, and offers to compact the session first, with what to keep, so the build starts without the research and planning history; after you compact, it shows the same next step again. A plan does not authorize implementation: typing that command is what approves the criteria shown and starts [Build](/flows/build/). A correction to a criterion is not approval; the agent revises and shows it again. The summary also names each external fact not measured and each class of effect you answered, with the effects left unauthorized. The steps are in the [plan handoff reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/plan-handoff.md).

## What it does not do

- It does not write the implementation, or approve commits, pushes, pull requests, merges or deployments.
- It does not hide open technical doubts inside later tasks.

## Related flows

- Start from [Research](/flows/research/) when you still need facts before planning.
- [Figma design](/flows/figma-design/) comes first when a new screen needs an approved example.
- Hand the finished plan to [Build](/flows/build/) by pointing it at the change folder.
- [Close](/flows/close/) archives the change folder when the work is integrated.
- If something new comes up while the plan is open, the agent classifies it first; see [changes and loops](/flows/overview/).

## FAQ

### Do I need a plan for a small change?
No. A change whose how leaves you no decision goes from Research to Build in the same turn.

### Does a ready plan mean the agent may start building?
No. Your reply starts Build, unless you asked beforehand to build after the plan without stopping, or unattended delegation is active.

### Where is the exact rule?
Source: [content/skills/flow-plan/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/SKILL.md)
