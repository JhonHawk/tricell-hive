---
name: hive-rule-close-report
description: "Hive rule leaf: read when you write a final report, in any variant, or classify a finding. Sets the report's parts and variants, the findings at the close, priorities P0 to P3, candidates."
disable-model-invocation: true
---

# Final report, findings and candidates

Read when you write a work item's final report, in any variant below, or classify a finding.

## Final report

Report the whole work item in one message, each part only when it has content:

- **Integrated:** what changed and where: the pull request, the commits, the ticket's state.
- **Resolved:** each acceptance criterion with how it was shown, the corrections the user asked for, and the defect's proof the [verification floor](../hive-rule-verification-floor/SKILL.md) sets.
- **Try it:** the block the verification floor sets, when the change can be seen or run.
- **Omitted:** each check that was skipped or could not run, with its reason, including the CI-only jobs.
- **Repaired:** the local environment repairs made, such as reinstalled dependencies or seeded data.
- **Support material:** each item with its path and the action applied, or listed as pending, as [support material](../flow-close/references/support-material.md) sets.
- **Findings:** the defects not fixed, as [Findings](#findings) sets.
- **Candidates:** the work's repetitions, as [Candidates](#candidates) sets.
- **Next:** what remains, with who owns it, and what comes next, such as the merge left to the user, review bots or base CI still pending, and the pending archives with the hygiene reply.

When the user asks to keep or share this report, or it exceeds what fits on one screen, offer a `create-report` page as one reply of the close's reply block in [report-endings](../hive-rule-report-endings/SKILL.md).

## Variants

`flow-close` writes the report at its step 6; the delivery mode in force, or the flow that called the close, picks the row. Closes with no row of their own (an abandon, hygiene, the end of a session, an unattended close) use the first row.

| Variant | Parts | Adds |
| --- | --- | --- |
| Close at the completion point (merge done) | every part with content | none |
| Pull request left open in `ready` or waiting on a reviewer | every part | in Integrated, the head commit, the base commit its current-base check passed on and the CI result; in Next, the pending merge and on whom, and the base CI still pending, as [merge gate](../hive-rule-merge-gate/SKILL.md) sets |
| Deployment (a cleanup close that `flow-deploy` calls) | every part | the run; per environment the identity served, the health and the functional acceptance, each as met, failed, blocked or "did not run"; the triggers decided; the CI failures from before that step 5 of `flow-deploy` let pass; the tickets waiting on a check |
| `hold` | what is pending, the defect's proof and Try it | none |
| Pause outside `hold` | as `hold` | the commit of the change folder |

## Findings at the close

**See the findings** shows one table, after checking each finding for an existing ticket as [create-issue](../create-issue/SKILL.md) step 2 sets:

| ID | Finding | Detail and impact | Recommendation |
| --- | --- | --- | --- |
| H1 | <what, with its severity> | <where, how it fails, who it affects, the evidence> | <open a ticket, ignore, or fix now, with its reason> |

Recommend fixing now or a ticket as the ticket rule of [scope changes](../hive-rule-scope-changes/SKILL.md) sets; recommend ignoring when the finding did not recur or has no impact. After the table, ask the findings as a round, as [asking a round](../hive-rule-questions/SKILL.md) sets: each finding is one question with the options fix now, ticket and ignore, its recommendation starred; the user may answer all at once or a few at a time, such as "ok" or "H1 ticket, H2 ignore". A ticket answer creates it as that skill sets; a fix-now answer goes as **Resolve findings now** in [report-endings](../hive-rule-report-endings/SKILL.md) sets; an ignore answer is noted in the report. When every finding has an answer, end again with the close's reply block, without the findings replies.

**Find the next ticket** runs research's [status report](../hive-mode/playbooks/ticket-status.md#find-the-next-ticket) as it sets for that reply.

## Findings

A finding is a defect the work found but did not fix: a review finding below P1, a UI defect that was already there, a test that already failed or fails intermittently (keep its error output), or anything else noticed on the way. Every finding carries one priority by its observable impact, whatever scale its source used: P0 blocks the release, an operation or a core flow for every input; P1 is blocking, as [review findings](../hive-rule-review-findings/SKILL.md) defines it, and must be fixed before merge; P2 should be fixed but does not block; P3 is minor. A finding of P2 or higher with a concrete failure scenario is proposed; the rest get one line each in the report. P2 findings that share an area or a cause are proposed as one group that lists each member with its evidence; findings of P1 or higher are proposed one by one. Never open a ticket without the user's answer, which unattended delegation gives at its step 5.

## Candidates

A candidate is a repetition the work did not turn into a mechanism under [defect fix](../hive-rule-defect-fix/SKILL.md) point 5: an inline script or command chain the work ran two or more times for the same purpose, such as the same link check or hash comparison, or a correction the user gave twice in the work. The report lists each one on one line, with what repeated, how many times and the mechanism that point would choose, and asks nothing about it; a work with none has no such part. [Script lifecycle](../project-layout/references/script-lifecycle.md) sets where a script goes at each level.

At step 7, search the project's persistent memory for each candidate by the terms of its purpose, then create or update one record under the topic `candidate/<tool>-<target>` in kebab case, such as `candidate/link-check-docs`, with the sessions that saw it, their dates and its count. A candidate seen in two sessions or more is offered as the **Turn a candidate into a mechanism** reply of the close row of [report-endings](../hive-rule-report-endings/SKILL.md); it is not a finding, since it is no defect.
