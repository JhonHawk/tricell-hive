---
title: Build
description: Make an approved change in your project and check that it works before calling it done.
---

Build is where the work gets done. The agent changes code or configuration, runs the checks your project has, and reports what it verified and what it could not.

## When to use

Use Build when you have approved a change and the outcome is clear. That can be a saved plan, or just a sentence.

Example requests:

```text wrap
Implement the agreed change locally and verify it. Do not commit or publish.
Continue the plan in the change folder from the next unfinished task.
Fix the date parsing bug and add a test that would have caught it.
```

You do not need a plan first for a small change you understand.

The global core still controls scope, authorization, preservation and secrets.
Build links to canonical procedures when a decision, delegation, artifact or
verification needs them. Reporting applies to findings, relayed results,
explanations and messages with external links or IDs; greetings and short
direct answers need no reference read. Neither an installed reference nor an
optional loader proves that a helper followed it.

## What it produces

A changed working tree, plus evidence:

- The agent checks the current state before editing. If a plan exists, it treats the plan as a claim to confirm, not a script to replay.
- It works in small steps, often with helper agents for pieces that do not overlap, and tells you how it split the work. A helper reads its task, your quoted request and the criteria from the plan itself instead of a summary; without a saved plan, it gets your request word for word.
- For behavior that would hurt if it broke, such as calculations, stored data, permissions or public contracts, it writes the test first and watches it fail, when the project's tests can exercise that behavior. It takes one behavior at a time, and it fakes only what lies outside the project, such as an external service or the clock, never the project's own code.
- Before it writes a test, it asks which realistic change to the code would make that test fail, and skips the test when there is none. It extends the test that already covers the behavior instead of writing a duplicate, and skips an example test that computes its expected value with the code it tests, and it never adds an entry point to the code only so tests can reach it. For pure logic that has a check independent of the code, it can use a property-based test instead of examples, as [test-first development](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/references/tdd.md) sets.
- When a change replaces or migrates a writer, parser, query or system and the old path can still run, it runs the old and the new path on the same inputs, with no real external effects, and treats each difference as a defect or a change the plan documents. This old-against-new comparison is part of the [verification reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/references/verification.md).
- It runs the project's own tests, builds and linters and reports the results, including any that failed or could not run.
- For visible interface changes, independent reviewers look at the running result. The agent's own look does not count for that. The reviewers report problems that were already there only on the screens the change touches, and a review of a deployed environment whose code matches what was already reviewed locally starts from that local report instead of repeating it. The agent ranks those earlier problems from P0 to P3 by their impact before deciding whether to propose a ticket; breakage the change introduced or made worse is fixed as part of the work. The rules live in the [UI gate](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/references/ui-gate.md).
- Every progress or decision message opens with a bold context line, stage · ticket · progress, so you can tell which piece of work it belongs to when you follow several at once, as [reporting](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-close/references/reporting.md) sets. Between tasks it reports progress in a line and keeps going; the completion report and the closing question come when the scope you authorized is done, not after every task. On an iterative route you chose in Research, each finished round is one more progress line; the report and the closing question come when the route stops.
- If you ask how it is going, it answers in a line and keeps working. When a helper or a background process reports back and nothing changes for you, the agent answers in one line too, without listing again what is still running or repeating a question that waits for you. When the notice only says that a process the agent stopped on purpose has exited, such as a server it shut down while cleaning up, the agent writes nothing. A helper that stops while it still waits on work of its own, such as a watcher, a wait or a background command, counts as stopped, not as in progress: the agent checks at once whether the helper and that work are alive and resumes it, as [delegation](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-research/references/delegation.md#check-returned-work) sets. When work is still running at the end of a message, it tells you what will wake it and the duration or deadline of its next wait. Builds, deployments and CI runs it starts are watched with your CLI's monitor tool when it has one, until they finish, with a time limit so a run that hangs still wakes the agent, and a failure does not go unnoticed.
- Topics you add during the build go on a list of open topics. The agent tells you in one line how it classifies each, closes none without a destination, and names them all in its findings and completion reports and in the route and close questions. The rule is in the [guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md).
- When your CLI has a task list, the agent keeps the plan's tasks there, also when it continues a plan after a close, so you can see what is left.
- When a plan has numbered criteria, a separate verifier checks each task that is proven by a test or a named check. It starts with a clean context, reads the plan itself and reruns the checks, and never sees the implementer's reasoning or claims; where your CLI offers it, it runs on a different model, but the clean context is what makes its verdict independent. Other tasks are marked done on the evidence of their own check.
- A task added to the plan during the build or after you validated the result is planned and checked like any other: it gets its criteria and a way to check it, and when that is a test written first or a named check, the separate verifier runs before its commit. If it arrives after the code review ran, the agent tells you and asks whether to review it, as the [Plan flow's revision rules](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-plan/references/plan-revision.md) set.
- When it opens a pull request you authorized, the description says what stayed unverified and which assumptions the result relies on. When the work has numbered criteria, it lists the evidence recorded for each criterion and names anything you accepted without verification, as the [Git workflow skill](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/git-workflow/SKILL.md) sets.

If Git delivery was not settled earlier, the agent asks once, before the first edit, how you want it delivered and reviewed. When the repository's `## Hive` section declares `Delivery` and `Review`, the agent states them in one line before the first edit instead of asking, together with `UI verification`, which sets how a user interface change is checked before you validate it: `full` means a UX review and a functional walk by independent agents, `functional` only the functional walk, and `none` neither, only the agent's own checks and your validation. Without that key, and when the change has a visible interface effect beyond text or style, the agent asks once per repository and session, in that same round; a narrower instruction in your request still wins. When the work will not reach Git, it asks how it should be reviewed only if that choice is still open, so the review does not stop the work later. In that same round, or at that same point when your settings, an earlier answer or your request already settled delivery and review, it asks for the authorizations the plan already foresees for later steps, such as a costly check or preparing a shared environment, each named with its scope, so the work does not stop for them while you are away; one you leave unanswered stays unauthorized. A check counts as costly when it spends money or a third-party service's quota, changes shared or remote resources, or should take longer than 30 minutes, and ordinary reads such as fetching dependencies do not count; a local check of a few minutes with nothing irreversible is routine and the agent just runs it. The effects beyond your machine that the work and its verification need are classified by where they land and whether they can be undone: local ones take no question, the test data of a non-production environment takes one question for the whole set, which offers to look first for existing accounts or records, reversible production effects take one question per kind, and each critical effect (irreversible, spending real money, destroying or replacing infrastructure, or changing data the task did not create) takes its own question, as the [verification reference](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/references/verification.md#gate-selection-and-timing) sets. When the work later reaches an environment the plan did not cover, the agent asks for everything that environment needs in one round when the work there starts. The [global guidance](https://github.com/JhonHawk/tricell-hive/blob/master/content/guidance/global.md) defines the costly check.

Findings from that review are fixed as part of the work, by severity, as Build sets. After one follow-up review, a new finding with a small fix inside the change that changes no contract is fixed and covered by a test without another review round, and the report says so; anything larger or changing a contract comes to you as a decision. A follow-up pull request that fixes the same ticket's change, such as a CI or defect fix after its merge, is covered by the delivery you already chose, without a new question. A condition you attach to a grant, such as "no merge until I return", lasts only for the scope you stated; when that scope ends, the delivery mode applies again, and a handoff records the condition with its scope. When a human reviewer is assigned or requested on the pull request, the agent reads each of their comments against the code, asks in the thread when one is unclear, applies them blocking-first and replies in each thread, as the Git workflow skill sets.

## What it does not do

- It does not commit, push, open pull requests, merge or deploy unless you asked for that effect.
- It does not mark a check as passed when it could not run it.
- It does not edit a test you supplied as the acceptance check just to make it pass.
- It does not treat a successful build or a ticked box as proof by itself.

## Related flows

- [Plan](/flows/plan/) comes first for larger work and gives Build its tasks and checks.
- [Research](/flows/research/) helps when the cause of a failure is still unknown.
- [Close](/flows/close/) follows when the work is finished, abandoned or paused.

## FAQ

### Do I need a plan for a small change?
No. Give Build the outcome and its limits.

### Who commits and merges?
You decide. Say so in your request, declare it once in the repository's `## Hive` section, or answer the delivery question the agent asks. Without that, the agent stops at a verified local result. When you chose a delivery that commits, the agent makes one commit per task once that task is verified, with the task ID in the message, so the branch history shows each step. A task with a visible interface effect is committed later: after the first rendered view of its change was accepted, when one was shown, and the UI review of the change passed. A fix that the full suite, the security review or your validation needs on a task already committed goes in an additional commit that names the task; history is never rewritten. Commits and pushes run the repository's hooks. The agent skips one only with your approval, which lasts for the session in that repository. When a hook cannot run its checks in this environment, the agent runs them another way or tells you which did not run and which tools are missing, before pushing; when you cannot answer, it pushes only if the repository's CI runs those checks. Before merging, the agent makes sure the checks ran on the current base branch, updating the branch and waiting for them again when the base moved, or using the repository's merge queue when it has one.

### What about an urgent fix for production?
The fix enters the base branch and is promoted at once through each environment, in order. If a promotion would also carry work you are holding back, the agent stops and offers a hotfix branch from production, only where your repository's branch policy accepts it, and asks you before that push. Afterwards it offers a short postmortem, written only if you want it.

### What if the agent finds that the plan no longer matches the code?
It looks into the mismatch and updates the plan or asks you for the missing decision before it keeps editing. The update follows the Plan flow's rules for revising a plan in progress: a changed criterion, decision or contract goes back to the plan reviewers once, and a change of intent becomes a new change.

### What happens to a user interface change?
When a change adds or rearranges a layout, the agent first shows you its first rendered version through a small `walkthrough.json` and selected captures, gives you the artifact's path in one line and then asks whether the composition is right, while independent work keeps running. The agent leaves the running result available for you to check, with steps to follow, and keeps its screen captures in the work's support folder until the session ends, then cleans up the ones you did not ask to keep. With Hive installed, the viewer renders the JSON; only an absent binary selects a standalone HTML fallback. A stopped server or failed registration does not generate HTML instead. When you validate a change, the manifest carries the checks table, captures and, when the skill calls for one, a short video of a flow, as [flow-walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md) sets. Before it asks you to validate, in any mode that stops for your validation, it runs in order the UI review, the full test suite of each affected application locally, even when CI will run it later, and the security review when the change crosses a security boundary. Both questions let you have the agent open the viewer with `hive view open`, approve, or write your corrections in the answer's free-text field so they arrive with their detail; the agent opens nothing any other way, as [flow-walkthrough](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/SKILL.md) sets. It keeps these files until the session ends too. To keep walkthroughs after the session and browse them by project and session, with ticket details on each session, you can run the local viewer `hive view`, which the agent never starts; the [README](https://github.com/JhonHawk/tricell-hive#review-ui-walkthroughs-in-one-place) shows how.

Approval applies to the result you saw. If a later fix, including one from code review, affects verified views or journeys, the agent refreshes that evidence before delivery. It asks for approval again when the visible result you accepted changes, as [the UI gate](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-walkthrough/references/ui-gate.md#changes-after-a-gate-or-approval) sets.

The local viewer excludes new manifests confirmed as tracked in Git. If a stored
copy's source becomes tracked, it moves to a separate explanatory section; its files,
pin and normal retention remain. When Git or the source is unavailable, the viewer
warns that its origin is unknown, rather than claiming agent authorship. It also
rejects hard-linked manifests and referenced files during ingestion.

The viewer lists one entry per project, host and declared session ID, with session
counts and compact title, ticket summary, segment count and date rows. Tickets from
the same conversation remain inside that session as chronological segments with
separators. Ticket search can jump to a matching segment without hiding its siblings. Projects initially open only when they contain work in
progress; your expansion choices, sidebar visibility, archive filter and selection
survive refresh in the same tab. The sidebar brand header shows a route icon.
Each segment has a toolbar that shows work status as information and offers
archive/restore, pin/unpin and deletion of only that stored copy. The session header
collects known PR or commit delivery links; each segment keeps its own links too.
No registered delivery is shown when metadata is absent, and the captured candidate
is kept separate from the work delivery. Older strict readers need the new binary
and an explicitly restarted viewer before producers use the optional delivery field.
Archived copies are hidden by default; the icon beside Projects reveals them.
Archiving never changes source files, work status or pinning, and an unpinned
archived copy still expires under normal retention.

New walkthroughs explicitly mark the work as in progress. The viewer shows that
state separately from Git provenance; a final report can still describe work in
progress. [Close](/flows/close/) marks this work closed when it completes, is
abandoned or is paused, and explicitly resuming it restores in progress. Older
walkthroughs without a state remain valid and show that no state was reported.
This does not detect whether your CLI process is alive.

On screen, desktop capture previews use one strip width at a time. Additional
captures remain accessible horizontally, narrow strips keep a peek of the next
capture, and the carousel and print layout remain available. Previously generated
standalone pages keep their embedded stylesheet until regenerated.

### Where is the exact rule?
See the source linked below.

Source: [content/skills/flow-build/SKILL.md](https://github.com/JhonHawk/tricell-hive/blob/master/content/skills/flow-build/SKILL.md)
