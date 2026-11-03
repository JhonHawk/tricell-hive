# Plan handoff

Read right before returning a ready plan, at `flow-plan` step 6. Reading this reference does not invoke the planning workflow.

## Write the summary first

A saved change folder holds the detail and the user may not open it, so summarize the plan in the message text, in the session language. A message that only points at the change folder, or a question card in place of the summary, is a failure of this step. Do not paste the whole plan. A plan kept in the conversation, as [flow-plan](../SKILL.md#where-the-plan-lives) sets, has no folder behind it: its summary also carries each task's `Closes:` line and verification. The summary has, in this order:

1. The objective in one or two sentences.
2. The text of each acceptance criterion `AC<n>`, and the name of each requirement the delta adds, modifies or removes: they are the contract the user approves.
3. The decisions in force, and the minor calls (as «Questions and routes» in the global guidance defines them) taken without asking with their reasons.
4. For a change to a screen, the sketch of each changed region before and after from the `UI` section of `design.md`, with its components and states. Check that each sketch matches the text that describes it, such as two actions drawn side by side when the text says so.
5. Each task on one line, with what it delivers and who executes it.
6. The prerequisites still unmet, with who provides each.
7. The verification: the tests, the UI walk and review, and the code review.
8. The delivery mode, with the validation stop and the endpoint after it.
9. What remains open or unauthorized, and the change folder path or that the plan lives in the conversation.

## Then end with the build route

End the summary with the next-step block that the global guidance's «Questions and routes» section sets, with its compaction reply for a build in the same session, never with a question card; when the plan has a viewer session, that block carries the open-viewer reply line that [viewer publication](../../flow-walkthrough/references/viewer-publication.md#open-the-viewer) sets. The build route is a route, not a request for leave to proceed: a plan does not authorize implementation, and only the user's reply starts `flow-build`.

- Recommend building unless something blocks it, and say in the block that typing the build command approves the criteria shown, by their IDs with a gloss, such as "approves AC1–AC4 (export, permissions and audit)": `/flow-build <change folder>` in the host's syntax, or, for a plan kept in the conversation, `/flow-build` with the ticket key or a short description of the work.
- Recommend stopping with the plan ready only for a named blocker, such as an open decision, an unmet prerequisite or the user's statement not to implement now, and say which; that the user asked for a plan is not a blocker.

A correction to a criterion or a sketch is not approval: revise the plan, show the revised criteria or sketch, and end with the build route again.

When execution is already explicitly authorized, show the summary the same way and continue to `flow-build`.

Leave a change folder uncommitted: `flow-close` versions it once at close.
