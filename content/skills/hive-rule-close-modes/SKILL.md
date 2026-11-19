---
name: hive-rule-close-modes
description: "Hive rule leaf: read when a close starts. One table sets which steps of flow-close run now, which wait, and whether the close ends on the footer or on the report."
disable-model-invocation: true
---

# Close modes

Read when a close starts, before step 1 of [flow-close](../flow-close/SKILL.md). Find the row that fits, then run the steps as it says. The close also runs when an operational task ends, as the global guidance's «Proportionality» defines it, and when the user abandons the work; a step with nothing to act on, such as the ticket move of work that was not integrated, is skipped.

| Situation | Runs now | Waits | Ends with |
| --- | --- | --- | --- |
| Completion point reached, or an operational task ends | Steps 1, 3, 6 and 7 | Steps 2, 4 and 5, recorded as pending | Step 8's footer |
| That case, when step 2's inventory lists no item to act on, step 4 has no change folder and step 5 has nothing of this work to remove or ask about | Steps 1, 3, 6 and 7, and steps 2, 4 and 5 at once, before step 6 | Nothing | Step 8's footer |
| Follow-up close of a pull request that waited at its completion point, abandoned work, a deploy's cleanup close, a hygiene request, the session's end, or any close under [unattended delegation](../unattended-delegation/SKILL.md) | Every step, step 2's question included | Nothing | The report, without the footer |
| Work delivered under `hold`, before it is delivered | Step 1 pauses its viewer session; step 2 lists its support material without asking or acting on it; step 5 stops the processes and containers it started and keeps everything else; step 6 reports what is pending; step 7 records it | The full close, until the work is delivered and reaches its completion point | The report, without the footer |
| Work paused outside `hold`, such as a plan paused before its build | As the `hold` row, and also commits its change folder as [git-workflow step 4](../git-workflow/SKILL.md) sets for a pause | The full close | The report, without the footer |

When steps wait:

- Retain step 1's readback and receipts and record their location for later input cleanup; do not close an already closed viewer again.
- Run the pending steps in order when the user chooses to close (the footer's **Close** reply, which exists only while a step is pending), starts another work item or ends the session. Report what they did and any viewer limitation, and update memory.
- When another work starts, step 2's question joins its first round; steps 4 and 5 wait for the answer.
- If the session stops without a reply, a later close or hygiene request resumes the pending steps from memory.

Close a work item once. The follow-up close of a pull request that waited at its completion point, and the full close of `hold` work after its pause close, are part of that close, not a repeat. After the footer's reply, act on it; do not run the close or its final report again.
