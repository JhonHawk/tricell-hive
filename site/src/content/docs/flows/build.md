---
title: Build
description: Make an approved change in your project and hand over a built candidate with the evidence of its own checks.
---

Build is where the work gets done. The agent changes code or configuration, runs the checks your project has, and tracks what it verified and what it could not. It ends at a built candidate and hands it straight to [Verify](/flows/verify/), without stopping; it writes a full report only when Verify is skipped, and then inside the completion report. Verify decides the level of independent checking, integrates the result and leaves the close and any deployment to their own steps.

## When to use

Use Build when you have approved a change and the outcome is clear. That can be a saved plan, or just a sentence.

Example requests:

```text wrap
Implement the agreed change locally and verify it. Do not commit or publish.
Continue the plan in the change folder from the next unfinished task.
Fix the date parsing bug and add a test that would have caught it.
```

You do not need a plan first for a small change you understand: one result in one repository whose approach is already settled. When you ask for anything bigger, such as a ticket with several steps, the agent goes to [Plan](/flows/plan/) first, or to [Research](/flows/research/) when what to change is not yet clear, unless you explicitly chose to skip the plan. When you let it pick the ticket, the completion report names the one it chose and why. A ticket another session already has in progress is listed apart, with a note that you can resume it if it is yours, and never recommended or chosen as the next one, as the [guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md) sets.

The global core still controls scope, authorization, preservation and secrets.
Build links to canonical procedures when a decision, delegation, artifact or
verification needs them. Reporting applies to findings, relayed results,
explanations and messages with external links or IDs; greetings and short
direct answers need no reference read. Neither an installed reference nor an
optional loader proves that a helper followed it.

## What it produces

A changed working tree, plus evidence:

- The agent checks the current state before editing. If a plan exists, it treats the plan as a claim to confirm, not a script to replay.
- It works in small steps, often with helper agents for pieces that do not overlap. A helper reads its task, your quoted request and the criteria from the plan itself instead of a summary; without a saved plan, it gets your request word for word.
- For behavior that would hurt if it broke, such as calculations, stored data, permissions or public contracts, it writes the test first and watches it fail, when the project's tests can exercise that behavior. It takes one behavior at a time, and it fakes only what lies outside the project, such as an external service or the clock, never the project's own code.
- Before it writes a test, it asks which realistic change to the code would make that test fail, and skips the test when there is none. It extends the test that already covers the behavior instead of writing a duplicate, and skips an example test that computes its expected value with the code it tests, and it never adds an entry point to the code only so tests can reach it. For pure logic that has a check independent of the code, it can use a property-based test instead of examples, as [test-first development](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/references/tdd.md) sets.
- When a change replaces or migrates a writer, parser, query or system and the old path can still run, it runs the old and the new path on the same inputs, with no real external effects, and treats each difference as a defect or a change the plan documents. This old-against-new comparison is part of the [verification reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/references/verification.md).
- Before it chooses which tests to run for any build, it reads the [verification reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/references/verification.md). It runs the project's own tests, builds and linters and records the results, including any that failed or could not run, for the hand-off or the report.
- A defect that blocks or breaks what you authorized, such as a failing check, a review finding or a failed verification round, is fixed as part of the work without asking and named in the final report. That includes a CI run the change breaks: the fix goes in a new commit on the same pull request, or in a follow-up pull request under the same delivery once it is merged. A failure the base branch already had before this work, and that does not block it, does not hold the work open: the agent reports it apart as a side finding. The agent asks only when the fix needs something you have not authorized or a choice that is yours. The rule is in the [guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).
- Other side findings are reported, never silently absorbed or dropped; a small, low-risk one may still be fixed in the work and reported apart. That includes a command or tool that failed and that the agent worked around, such as by retrying with another path: the report names the failure and the workaround, and the agent proposes a fix or a ticket for it only when it lies in this project, reporting one in another project for that project's own session. Before proposing a finding as a fix or a ticket, during the build or at the close, the agent searches both the tracker and its memory for duplicates and for findings you declined or asked not to hear about, as [incidental findings](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/references/incidental-findings.md) sets.
- For visible interface changes, the agent first inspects the running result itself and, before reporting a task, reads the scope and inspection rules in the [UI gate](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/references/ui-gate.md#the-implementers-inspection). That look is not the gate: [Verify](/flows/verify/) has independent reviewers look at the running result, and the agent's own look does not count for that. The reviewers report problems that were already there only on the screens the change touches, and a review of a deployed environment whose code matches what was already reviewed locally starts from that local report instead of repeating it. The agent ranks those earlier problems from P0 to P3 by their impact before deciding whether to propose a ticket; breakage the change introduced or made worse is fixed as part of the work, except a fix that would customize the UI library's theme or installed components, which needs your decision or one the project already recorded. The rules live in the [UI gate](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/references/ui-gate.md).
- Every result, blocker, closing message or question opens with a bold context line, stage · ticket · progress, where the stage is the running flow's name, untranslated, such as Build or Research, so you can tell which piece of work it belongs to when you follow several at once, as [reporting](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/references/reporting.md) sets. Between tasks it owes you no progress message and keeps going; anything it does write goes together with its next action, because some CLIs, such as Grok Build, end the turn on a reply without one. While work it was authorized to do remains, it stops only to ask you something, at a validation stop, to wait for a result with nothing else to do, at a blocker, or when you or another rule stop the work, as [work-item boundary](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/references/work-item-boundary.md) sets. The completion report and the closing question come when the scope you authorized is done, not after every task; on an iterative route you chose in Research, they come when the route stops.
- If you ask how it is going, it answers in a line and keeps working. When a helper or a background process reports back and nothing changes for you, the agent writes no report and does not repeat a question that waits for you. When the notice only says that a process the agent stopped on purpose has exited, such as a server it shut down while cleaning up, the agent writes nothing. A helper that stops while it still waits on work of its own, such as a watcher, a wait or a background command, counts as stopped, not as in progress: the agent checks at once whether the helper and that work are alive and resumes it, as [delegation](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/references/delegation.md#check-returned-work) sets. When its turn ends to wait on work still running, it says in one sentence what is still running and what its result decides. Builds, deployments and CI runs it starts are watched with your CLI's monitor tool when it has one, until they finish, with a time limit so a run that hangs still wakes the agent, and a failure does not go unnoticed.
- Topics you add during the build go on a list of open topics. The agent tells you in one line how it classifies each, closes none without a destination, and names them all in its findings and completion reports and in the route and close questions. The rule is in the [guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).
- When your CLI has a task list, the agent keeps the plan's tasks there, also when it continues a plan after a close, so you can see what is left.
- When a plan has numbered criteria, a separate verifier checks each task that is proven by a test or a named check, at the checkpoint the plan assigns to that task rather than after every task. Tasks of one checkpoint are verified together, those that share a file one after another, and a task that depends on a task of its own checkpoint starts without waiting for a verdict, while a task that depends on an earlier checkpoint starts once those tasks are verified. A plan that sets no checkpoints is verified task by task. The verifier starts with a clean context, reads the plan itself and reruns the checks, and never sees the implementer's reasoning or claims; where your CLI offers it, it runs on a different model, but the clean context is what makes its verdict independent. Other tasks are marked done on the evidence of their own check.
- A task added to the plan during the build or after you validated the result is planned and checked like any other: it gets its criteria and a way to check it, and when that is a test written first or a named check, the separate verifier runs before the work is published for merge. If it arrives after the code review ran and before the one re-review that review allows, the agent reviews what changed since the reviewed point when an earlier review answer covers it, and otherwise asks you; after that re-review, or after the review of the corrections you ask for when you validate, it reports a small edit inside the change's scope that changes no contract as unreviewed and asks you about anything else, as the [Plan flow's revision rules](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/plan-revision.md) set.
- When it opens a pull request you authorized, the description says what stayed unverified and which assumptions the result relies on. When the work has numbered criteria, it lists the evidence recorded for each criterion and names anything you accepted without verification, as the [Git workflow skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/git-workflow/SKILL.md) sets.

Before the first edit, the agent moves a tracked ticket to its in-progress
state. Once per work item it asks for any delivery and code-review decisions
that remain open, reusing decisions you already gave; a mechanical change
needs only the delivery decision. A substantive visible interface change receives its own
walk and independent UX review in [Verify](/flows/verify/). A security review
runs only when you request it. The current delivery rules live in the
[shared guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).

The selected code review runs on the candidate before integration. When the
delivery stops for your validation, Verify can show the candidate while the
reviews run and says that acceptance is subject to their findings. It handles
confirmed defects and your corrections as [Verify](/flows/verify/) sets.

## What it does not do

- It does not commit, push, open pull requests or merge unless you asked for that effect.
- It does not run the independent verification, the dedicated code review, the close or a deployment; [Verify](/flows/verify/), [Close](/flows/close/) and [Deploy](/flows/deploy/) do.
- It does not mark a check as passed when it could not run it.
- It does not edit a test you supplied as the acceptance check just to make it pass.
- It does not treat a successful build or a ticked box as proof by itself.

## Related flows

- [Plan](/flows/plan/) comes first for larger work and gives Build its tasks and checks.
- [Research](/flows/research/) helps when the cause of a failure is still unknown.
- [Verify](/flows/verify/) follows with the independent checks, the review and integration; [Security](/flows/security/) runs inside it when you ask.
- [Close](/flows/close/) follows when the work is finished, abandoned or paused.
- [Deploy](/flows/deploy/) promotes an integrated change through your environments.

A running viewer can show the session's declared milestones as work progresses,
including a Build stage before screenshots exist. The timeline keeps ticket-scoped
visual evidence and prior document revisions with the stage that produced them.
The agent reports a state from the conversation; the indicator does not prove a
host process is running or that you approved from the viewer. The
[publication skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md)
owns the viewer procedure.

## FAQ

### Do I need a plan for a small change?
No. Give Build the outcome and its limits. For a change that is not small, the agent plans first unless you say to skip the plan.

### Who commits and merges?
You decide through the work item's delivery answer or an explicit request.
Under `interactive`, the agent stops before the push for your validation;
under `hold`, the verified working tree is the result. When the chosen delivery
includes commits, Build checks and commits each task, then Verify checks the
candidate before integration. The answer applies to this work item, as the
[shared guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md)
sets.

### What about an urgent fix for production?
[Deploy](/flows/deploy/) handles it: the fix enters the base branch and is promoted at once through each environment, in order. If a promotion would also carry work you are holding back, the agent stops and offers a hotfix branch from production, only where your repository's branch policy accepts it, and asks you before that push. Afterwards it offers a short postmortem, written only if you want it. The procedure is in the [urgent fixes reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-deploy/references/urgent-fixes.md).

### What if the agent finds that the plan no longer matches the code?
It looks into the mismatch and updates the plan. A missing decision that is yours does not stop the build: when undoing it only touches this work's unpublished code on your machine, the agent applies the recommended option as a recorded assumption and keeps working; otherwise it finishes the rest of the work first and then asks, and it asks at once only when the decision blocks everything left. You confirm the assumptions at the end. A request that adds a task or changes the outcome is still classified first, as [Changes and loops](/flows/overview/) describes. The update follows the Plan flow's rules for revising a plan in progress: a changed criterion, decision or contract goes back to the plan reviewers once, and a change of intent becomes a new change.

### What if I ask for something extra during the build?
The agent classifies the request before acting. A small understood addition stays in Build; one that adds another ticket, another repository, a contract or stored data goes back to [Plan](/flows/plan/). See [changes and loops](/flows/overview/) and the [scope changes reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/hive-core/references/scope-changes.md).

### What happens to a user interface change?
When a change adds or rearranges a layout, Build shows its first rendered view
with realistic data before it connects the backend, then adjusts the interface
in response to your feedback. For a final candidate, [Verify](/flows/verify/)
walks the changed behavior, runs its checks and independent reviews, and shows
the relevant result when your delivery stops for validation. Captures are
proportional to the change: a correction refreshes the affected views and
states, retaining identified evidence for what did not change. The
[Verify skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/SKILL.md)
owns verification and the
[publication skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md)
owns the viewer handoff. With no Hive binary, a UI walkthrough can use a
standalone HTML fallback. An installed but stopped viewer does not select that
fallback or start itself.

The UI review keeps the design system of your UI library as its baseline: a value the library sets by default, or one your project recorded, is not reported as a contrast, color or sizing problem, though broken layout, missing states, keyboard order, a missing focus indicator or a missing accessible name still are, and a fix that would customize the library's theme or components is applied only with your decision or the project's. When a problem comes from a shared source, such as a theme token, it is reported once and recorded in the project's UI guide, so it does not come back for every screen you touch. The rules live in [UI review criteria](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-verify/references/ui-review-criteria.md#design-baseline).

Approval applies to the result you saw. When a later change affects a verified
view or journey, the agent refreshes that evidence and asks again when the
visible result you accepted changes, as [Verify](/flows/verify/) sets.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-build/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md)
