# Plan handoff

Read right before returning a ready plan, at `flow-plan` step 6.

## Write the summary first

Ask outstanding decisions with just the facts needed to choose. Then give the full summary once. Summarize even a saved plan in the message, in the session language: the reader may skip the file. A link, a question card or the entire plan pasted here does not serve as its summary. For a [conversation plan](../SKILL.md#where-the-plan-lives), include each task's `Closes:` and verification, each [external measurement](../../hive-core/plan/external-facts.md), and each verification command, result and date. Use this order:

1. The objective in one or two sentences, and the files and areas of the scope line [hive-mode](../../hive-mode/SKILL.md) step 5 sets, recounted from the plan; its stages and checks go in items 7 and 8. When research gave that line in this conversation and the file count differs by more than a quarter, or an area was added or removed, say what changed and why, such as "~42 files, not ~30: the retired contract's tests are removed"; without that line, give the count alone.
2. The text of each acceptance criterion `AC<n>`: they are the contract the user approves.
3. The decisions in force and those taken without asking, as [questions](../../hive-core/messages/questions.md) «Decide and list» sets.
4. For a change to a screen, the sketch of each changed region before and after from the `UI` part of the plan's `## Design`, with its components and states; for a region with an approved example, the example's record in place of its after sketch. Check that each sketch matches the text that describes it.
5. Each task on one line: the change it delivers and who executes it.
6. The prerequisites still unmet, with who provides each, and each external fact written as not measured, with the task that meets it first.
7. Group a shared check once by the task IDs it covers. Name each exception with its task. The verification: the tests, the UI walk and review, and the code review, and each verification command not measured.
8. The delivery mode, with its validation stop when the work has a surface, as the global guidance's «Delivery» defines it, and the endpoint after it.
9. Each class of effect with its answer as the [effects round](../../hive-core/plan/external-effects.md) records it, what remains open or unauthorized, and the plan file's path or that the plan lives in the conversation.

## Then end with the build route

End the summary with the [next-step block](../../hive-core/messages/report-endings.md#the-next-step-block), with its compaction reply for a build in the same session, and not with a question card, since the route is not a decision to ask. A plan does not authorize implementation: the next-step block sets what starts `flow-build`.

- Recommend building unless something blocks it, and say in the block that typing the build command approves the criteria shown, by their IDs with a gloss, such as "approves AC1–AC4 (export, permissions and audit)": the `flow-build` command with the plan file, or, for a plan kept in the conversation, with the ticket key or a short description of the work, in the host's syntax that [the next-step block](../../hive-core/messages/report-endings.md#the-next-step-block) sets.
- Recommend stopping with the plan ready only for a named blocker, such as an open decision, an unmet prerequisite or the user's statement not to implement now, and say which; that the user asked for a plan is not a blocker.

A correction to a criterion or a sketch is not approval: [revise the plan](../../hive-core/plan/plan-record.md#revise-a-plan-in-progress), show the revised criteria or sketch, and end with the build route again. Continue to `flow-build` in the same turn when [same-turn continuations](../../hive-core/messages/report-endings.md#same-turn-continuations) lets the plan continue.

The plan file is never committed: it stays local, and the pull request body carries the public record, as [git-workflow](../../git-workflow/SKILL.md) step 6 sets.
