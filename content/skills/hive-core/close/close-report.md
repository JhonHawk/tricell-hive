# Final report, findings and candidates

Read before a final report or classifying a finding.

## Final report

One message for the whole work item, each part only when it has content:

- **Integrated:** the pull request, the commits and the ticket's state.
- **Resolved:** each acceptance criterion with how it was shown, the user's corrections, and the defect's proof the [verification floor](../build/verification-floor.md) sets.
- **Try it:** the block the [verification floor](../build/verification-floor.md) sets.
- **Omitted:** each check skipped or unable to run, with its reason, CI-only jobs included.
- **Repaired:** local environment repairs, such as reinstalled dependencies or seeded data.
- **Support material:** each item with its path and the action applied or pending, as [support material](../../flow-close/references/support-material.md) sets.
- **Principles:** when `flow-build` handed it over, or the checkpoint holds its `Principles:` line, one line naming each code principle read that changed a decision, or none.
- **Findings:** the defects not fixed, as [Findings](#findings) sets.
- **Candidates:** the work's repetitions, as [Candidates](#candidates) sets.
- **Next:** what remains and who owns it (the merge left to the user, review bots or base CI pending, what [support material](../../flow-close/references/support-material.md) sets to name there).

When the user asks to keep or share the report, or it exceeds one screen, offer a `create-report` page as one reply of the close's reply block in [report-endings](../messages/report-endings.md#the-close-footer).

## Variants

`flow-close` writes the report at its step 6; the delivery mode in force, or the flow that called the close, picks the row. A close with no row of its own (an abandon, an unattended close, a hygiene request or the session's end with no earlier close) uses the first.

| Variant | Parts | Adds |
| --- | --- | --- |
| Close at the completion point (merge done) | every part with content | in Next, the base CI still running, as [merge gate](../git/merge-gate.md) sets |
| Close pending | Integrated, Resolved, Try it, Omitted, Findings and Next, first line not a bold «Close» | the «Adds» of its later close's row; in Next, what waits, the branch, worktree and processes left and the close command |
| Pull request left open in `ready` or waiting on a reviewer | every part | in Integrated, the head commit, the base commit its current-base check passed on and the CI result; in Next, the pending merge and on whom |
| Deployment (a cleanup close that `flow-deploy` calls) | every part | the run named in step 1 of `flow-deploy`; per environment the identity served, the health and the functional acceptance, each as met, failed, blocked or "not run"; the triggers decided; the CI failures from before that step 5 of `flow-deploy` let pass; each ticket that stayed out of its `Completion` state, or moved without a walk, as [ticket state](../tracker/ticket-state.md) sets |
| `hold` | what is pending, the defect's proof, Try it and Principles | none |
| Pause outside `hold` | as `hold` | none |
| Hygiene request or session's end after the work item's close | what ran of the pending items, and what stays pending | none |

## Findings at the close

**See the findings** shows one table, after checking each finding for an existing ticket as [create-issue](../../create-issue/SKILL.md) step 2 sets:

| ID | Finding | Detail and impact | Decision |
| --- | --- | --- | --- |
| H1 | <what, with its severity> | <where, how it fails, who it affects, the evidence> | <listed, in the ticket proposal, or asked: fix now, ticket or ignore recommended, with its reason> |

Each decision is the one the sorting of [review findings](../review/review-findings.md) gives it. Then ask as a round ([asking a round](../messages/questions.md)) only what that sorting leaves to the user and has no answer yet:

- the ticket proposal;
- each finding asked on its own, with the options fix now, ticket and ignore and its recommendation starred: fixing now or a ticket as the ticket rule of [scope changes](../plan/scope-changes.md) sets, ignoring when the finding did not recur or has no impact.

A listed finding is acted on only when the user names it. The user may answer all at once or a few at a time, such as "H1 ticket, H2 ignore". A ticket answer creates it as create-issue sets; a fix-now answer runs **Resolve findings now** as [report-endings](../messages/report-endings.md#the-close-footer) sets; an ignore answer is noted in the report. When every finding has an answer, end again with the close's reply block, without the findings replies.

**Find the next ticket** runs research's [status report](../../hive-mode/playbooks/ticket-status.md#find-the-next-ticket) for that reply.

## Findings

A finding is a defect the work found but did not fix: a review finding below P1, a UI defect that was already there, a test that already failed or fails intermittently (keep its error output), or anything else noticed on the way. Each carries one priority by its observable impact, whatever scale its source used:

- P0 blocks the release, an operation or a core flow for every input.
- P1 is blocking, one that breaks a criterion, a contract, data, security or the build, and must be fixed before merge.
- P2 should be fixed but does not block.
- P3 is cosmetic or internal: it breaks no criterion, contract, data or flow.

A finding against code that classifies free text by patterns that fails only on text the reviewer wrote, not on a file or record that predates the review, is at most P2 unless it breaks a contract, data or the build, or sits at a [security boundary](../build/security-boundaries.md). Each finding takes the decision the sorting of [review findings](../review/review-findings.md) gives it, and one no review sorted, such as one noticed on the way, is sorted the same way at the close. The report gives each finding one line with its decision. Never open a ticket without the user's answer, which unattended delegation gives at its step 5.

## Candidates

A candidate is a repetition the work did not turn into a mechanism under [defect fix](../build/defect-fix.md) point 5: an inline script or command chain the work ran two or more times for the same purpose, or a correction the user gave twice in the work. List each on one line, with what repeated, how many times and the mechanism that point would choose, and ask nothing about it; a work with none has no such part. [Script lifecycle](../../project-layout/references/script-lifecycle.md) sets where a script goes.

At step 7, search the project's persistent memory for each candidate by the terms of its purpose, then create or update one record under the topic `candidate/<tool>-<target>` in kebab case, such as `candidate/link-check-docs`, with the sessions that saw it, their dates and its count. A candidate seen in two sessions or more is offered as the **Turn a candidate into a mechanism** reply of the close row of [report-endings](../messages/report-endings.md#the-close-footer); it is not a finding, since it is no defect.
