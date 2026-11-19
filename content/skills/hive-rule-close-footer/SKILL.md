---
name: hive-rule-close-footer
description: "Hive rule leaf: read when you end a final report with the close footer. Sets the footer's heading, its replies and the findings table."
disable-model-invocation: true
---

# Close footer

Read at step 8 of [flow-close](../flow-close/SKILL.md), before the footer.

End the final report with [the next-step block](../hive-rule-next-step-block/SKILL.md), never with the question tool. Its heading recommends **close** with its reason, unless a proposed finding is P1 or higher, when it recommends seeing the findings, or a plan is in progress, when it recommends continuing it; with no step pending and neither case, its heading is **▶ Closed**, with no recommendation. Its replies come from this work's context, each only when it applies:

- **Close:** runs the pending steps, as [close modes](../hive-rule-close-modes/SKILL.md) sets; only while one is pending.
- **Resolve findings now**, naming them, such as "resolve H2 and H3 in this branch": they enter `flow-build` on the work's branch, with the code review of the fix and its checks, or `flow-plan` when the fix is not a small understood change.
- **See the findings**, with their count and highest severity, such as "see the 3 findings (1 P1)".
- **Explain a finding or a result**, naming it.
- **Turn a candidate into a mechanism**, naming it: only for a candidate whose record counts two sessions or more; it enters `flow-build` as **Resolve findings now** does.
- **Find the next ticket**, as **Find the next ticket** below sets, run only when the user picks it.
- **Continue** with what remains of the work, naming it, such as the next phase of a plan in progress.

**See the findings** shows one table, after checking each finding for an existing ticket as [create-issue](../create-issue/SKILL.md) step 2 sets:

| ID | Finding | Detail and impact | Recommendation |
| --- | --- | --- | --- |
| H1 | <what, with its severity> | <where, how it fails, who it affects, the evidence> | <open a ticket, ignore, or fix now, with its reason> |

Recommend fixing now or a ticket as the ticket rule of [scope changes](../hive-rule-scope-changes/SKILL.md) sets; recommend ignoring when the finding did not recur or has no impact. After the table, ask the findings as a round, as [asking a round](../hive-rule-question-channel/SKILL.md) sets: each finding is one question with the options fix now, ticket and ignore, its recommendation starred; the user may answer all at once or a few at a time, such as "ok" or "H1 ticket, H2 ignore". A ticket answer creates it as that skill sets; a fix-now answer goes as **Resolve findings now** sets; an ignore answer is noted in the report. When every finding has an answer, end again with the close footer, without the findings replies.

**Find the next ticket** runs research's [status report](../flow-research/references/status.md#find-the-next-ticket) as it sets for that reply.
