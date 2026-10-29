# Plan criteria and task markers

## When to read

Read when writing, reviewing, verifying or executing a retained plan's acceptance criteria, checkpoints or task markers. The rest of a retained plan's format is in [the plan format](plan-format.md). Reading this reference does not invoke the planning workflow.

## Acceptance criteria and task markers

Number acceptance criteria `AC1`, `AC2`, and so on, one observable behavior each. For each criterion, name the observation that would show it false, and make sure that observation is false at the change's base commit, the commit the build starts from; while planning, check it against the current head of the base branch. A criterion that is already true there, that only another change can satisfy, or that restates the request grades nothing: rewrite it or drop it, and tell the user when a dropped criterion came from their request or ticket. Conditions that must stay true, such as "the existing suite still passes", are constraints: list them apart from the criteria.

Every criterion appears in the `Closes:` line of at least one task, and every task that satisfies a criterion carries a `**Closes:** AC<n>, …` line, and its **Verification** covers each criterion it closes. Name in a task's **Verification** only the tests that exercise those criteria, such as one package's tests or a test filter. A whole-repository suite, a slower variant of it such as a race-detector run, or a build of everything is a shared gate: list it once in the plan's verification section, to run on the final candidate, not in each task. Every task whose test approach is `tdd` or `check` carries `Closes:`; if such a task closes nothing, a criterion is missing or the task is unnecessary. Keep `AC<n>`, `T<n>`, the `Closes:` and `Checkpoint:` labels, and the status lines in the marker table below untranslated in any session language, because the build and the task verifier search for them.

**Checkpoints.** In a plan with numbered criteria, every `tdd` or `check` task carries a `**Checkpoint:** C<n>` line. A checkpoint groups tasks that `flow-build` verifies together, as [plan execution](../../flow-build/references/plan-execution.md) sets. Number the checkpoints in plan order, and let the last `tdd` or `check` task close the last one. The plan's verification section lists the checkpoints in `tasks.md` with their tasks and the reason for each one's place. A task is the last of its group and closes the checkpoint when either condition holds:

1. two or more later tasks depend on it, directly or through other tasks, so a defect in it would spread; or
2. it changes a public or persisted contract, stored data, permissions, a security boundary, or concurrency.

Group every other task into the next checkpoint. A task depends only on earlier tasks of the plan that exist, never on itself. Number the checkpoints C1, C2 and so on without gaps, so that no task carries a lower number than an earlier task and the last `tdd` or `check` task carries the highest. An older plan whose tasks omit the field is verified task by task: without the `Checkpoint` field, each task is its own checkpoint. This rule and the risk criterion live only here; other files point to them.

Mark each task with one of these markers:

| Marker | Meaning |
| --- | --- |
| `[ ]` | Pending. |
| `[/]` | In progress. When the fix of an earlier task reopened it, a line `reopened by the fix of T<n>` follows it. When a prerequisite stops it before it reports, a line `blocked: <prerequisite>` follows it, and no verifier runs for it. |
| `[?]` | Done by the implementer, waiting for verification. When verification is blocked, a line `blocked: <prerequisite>` follows it. When all its criteria are met but a dependency in the same checkpoint is not yet `[x]`, a line `waiting: T<n>` follows it. |
| `[x]` | Verified. When the user accepted the task without verification, a line `accepted unverified by the user: <reason>` follows it. |
| `[!]` | Rejected by verification. A line with the round, the `AC<n>`, and the reason follows it. |

The `<prerequisite>` of a `blocked:` line is one that [runtime and UI gates](../../flow-verify/references/runtime-gates.md#runtime-and-ui-gates) leaves as a reported prerequisite, named with its cause; the local setup it lists is resolved first.

`flow-build` and `flow-verify` own the transitions, each in its own stage, as [plan execution](../../flow-build/references/plan-execution.md) and [flow-verify](../../flow-verify/SKILL.md) set. A plan without `AC<n>` criteria, such as an older `<topic>.plan.md`, uses only `[ ]` and `[x]`.
