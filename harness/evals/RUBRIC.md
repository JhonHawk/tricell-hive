# Skill pilot grading

This pilot evaluates the isolated skill package and its staged dependencies, not a
full Hive installation. A run measures one exact source snapshot, harness version,
resolved model and reasoning setting. It does not establish portability to an
unexecuted combination.

## Deterministic dimensions

| Dimension | Evidence | Passing condition |
|---|---|---|
| Activation | Completed tool reads of the target skill | Positive cases read it; negative cases do not. Catalog visibility alone is not activation. |
| Process | Completed reads, tool results and workspace changes | Required references are read successfully; observed writes remain within the case allowance. |
| Outcome | Before/after file snapshots and contract checks | Required files are produced or changed by this run and satisfy the case checks; protected fixtures remain unchanged. |
| Efficiency | Duration, available token counters and tool counts | Report descriptively for the same harness/model setting; no initial hard threshold. |

A started tool call is not evidence of successful execution. Final-answer claims
do not prove reads, verification or file writes. A required file already present
before the run does not prove that this run produced it. Negative cases include
existing artifacts precisely to protect against this false positive.

The activation check is an operational proxy for loading a skill, not proof that
the model reasoned from it. Process checks cover the explicit manifest assertions,
not every ordering constraint in Hive. Do not describe a passing filesystem check
as proof that no external action occurred; report the isolation boundary and trace
coverage separately. Missing or unsupported evidence is **not verified**.

## Human qualitative assessment

The initial pilot has no LLM judge. Until a human inspects the output, report style
and semantic quality as **not verified**, independently of deterministic results.
Use the following pass/fail rubric after inspection. Record the case, harness,
model, repetition, reviewer, result and an exact excerpt or artifact location for
each assessment. A failure names the observed defect, not a stylistic preference.

| Criterion | Pass | Fail |
|---|---|---|
| Task fidelity | Addresses the actual request and respects planning-only or read-only scope. | Answers a different question, manufactures authority or executes excluded work. |
| Evidence fidelity | Distinguishes supplied facts, inference and unverified results. | Invents measurements, sources, checks, approvals or completion. |
| Language and clarity | Clear Mexican Spanish, appropriate detail, explained technical terms. | Ambiguous instructions or omissions prevent the intended reader from using the result. |
| Plan quality | Bounded tasks identify files, verification and exclusions; grants match the request. | Generic tasks, invented approvals or a scope assent treated as implementation consent. |
| Report quality | Shareable report uses the supplied facts; table and second information kind help understanding. | Word-count padding, invented project facts or decorative structure without information. |
| Visual contract | Report remains readable with its specified palettes, narrow viewport and print layout. | Clipping, unreadable contrast, hidden essential content or broken print presentation. |

For report visuals, inspect the rendered artifact; source HTML alone cannot earn
a visual pass. For read-only answers, assess semantics rather than matching golden
wording. An accurate answer may differ in phrasing and tool choice.

## Interpreting the dataset

- `explicit`: the user names the skill or directly requests its procedure.
- `implicit`: the task meets a model-invoked skill's description without naming it.
- `contextual`: the fixture or task boundary supplies the decisive signal.
- `negative`: the target skill should remain inactive; other appropriate skills
  may still activate.

`flow-plan` accepts an explicit equivalent such as “planeemos”; a slash command is
not its only valid entry point. A question about an approach and agreement with a
scope are negative controls when no plan is requested. All positive plan cases
request drafts only: no implementation or publication grants are implied.

`flow-report` positives request a kept/shared artifact with at least 300 words, a
table and another information kind. Length or a comparison alone does not trigger
it. A rich conversational answer remains a negative when no artifact is wanted.

`language-rules` positives include code review and generation as well as edits.
Mentioning a language in ordinary translation or copy editing is not a code task.

Two repetitions can reveal variation; they do not establish a reliable success
rate. Keep results by case and repetition rather than hiding a failed run in an
average. Hard scope/authorization failures cannot be compensated by better style
or fewer tokens. Improvements require a new source snapshot and fresh runs, not
weakened assertions against an existing failure.
