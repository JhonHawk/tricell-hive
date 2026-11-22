# Model support floor

This page is the canonical list of the models Hive's distributed guidance is written for and of the candidates whose observed failures count toward a correction. Root `AGENTS.md` «Guidance structure» points here for the writing target, and «Measurement» points here for floor membership; neither restates the list. `tests/content/model_floor_test.go` checks that every model a role profile of `integrations/agent-profiles.json` names has a row, and that the rows marked as writing target are exactly the ones «Two sets» lists.

## Two sets

- **Writing target.** `opus` `gpt-6.2-sol`
- **Support floor.** The writing target plus one tier above it (Claude Fable 5.2, gpt-6.2-astra), one tier below it (Claude Sonnet 5.5, gpt-6.1-terra), and Grok 4.7 with its variant `grok-4.7-build-fast`, which the user runs in main sessions. A failure observed on any floor model, provisional or qualified, counts toward a correction as «Measurement» sets; a failure seen only on a model outside the floor does not. Haiku, Luna, flash-class models and superseded versions are not on the floor. When a reviewed session runs a model no row lists, ask the user whether it counts, unless it is a Haiku, Luna or flash-class model; a yes adds its row, marked as a main-session model and not a writing target, and the session then counts.

Distributed content is written clearly for the writing target: Claude Opus 5.5 and OpenAI gpt-6.2-sol (decision of 2026-11-22). The role profiles also run models one tier below, Sonnet 5.5 and gpt-6.1-terra, which the user kept: agentic benchmarks place them close to the target, and their documented risks (stopping early or skipping a check at low or medium effort) depend on effort, which the profiles set. A rule aimed at a model outside the writing target needs the recurrence that root `AGENTS.md` «Measurement» sets; a plausible misreading by a smaller model is not enough.

On hosts that inherit the session model (Pi, Cursor, Antigravity) or declare none (OpenCode), delegated roles run whatever model the session runs; only floor models count there.

## Membership and normalization

A model belongs to the floor by its identity, whatever host runs it:

1. Remove the provider prefix and the variant from both the session's model and the table's model before comparing them: `opencode-go/claude-sonnet-5-5#high` becomes `claude-sonnet-5-5`, which the `sonnet` row's alias resolves to, so a session on Sonnet 5.5 through another provider matches that row.
2. Resolve an alias to the full IDs listed under «Aliases»; membership means being one of those IDs. An alias that resolves to an older model on another provider, such as `sonnet` on Amazon Bedrock, is not a floor model there.
3. When a session's served model cannot be seen (Codex never exposes it, and an OpenCode session's trace does not carry it, although qualification runs read it from the session export), use its configured model and name it as configured.

## Floor

| Host | Model | Efforts | Origin | Writing target | Status | Verified | Reference |
| --- | --- | --- | --- | --- | --- | --- | --- |
| claude | `opus` | `high` | `reasoning`, `verifier` | yes | provisional | none | — |
| codex | `gpt-6.2-sol` | `medium`, `high`, `max` | `execution`, `reasoning`, `verifier`; user's main session, 2026-11-17 | yes | provisional | none | — |
| claude | `claude-fable-5-2` | — | tier above, 2026-11-22 | no | provisional | none | — |
| codex | `gpt-6.2-astra` | — | tier above, 2026-11-22 | no | provisional | none | — |
| claude | `sonnet` | `high` | `execution`, `focused`; tier below | no | provisional | none | — |
| codex | `gpt-6.1-terra` | `medium` | `focused`; tier below | no | provisional | none | — |
| grok | `grok-4.7` | — | `verifier` | no | provisional | none | — |
| grok | `grok-4.7-build-fast` | `high` | user's main session, 2026-11-17 | no | provisional | none | — |

Row changes: on 2026-11-22 the writing target narrowed to `opus` and `gpt-6.2-sol`, the Fable and Astra rows joined, and the Haiku, flash-class and superseded rows left the floor; observations already counted on them stay in their watch items.

A row is **provisional** until a qualification run passes; **Verified** then records its date and **Reference** the reference model of that run.

## Aliases

| Alias | Resolves to | Source | Verified |
| --- | --- | --- | --- |
| `opus` | `claude-opus-5-5` (Opus 5.5, on the Anthropic API) | [Claude Code model configuration](https://code.claude.com/docs/en/model-config) and [Claude models overview](https://platform.claude.com/docs/en/about-claude/models/overview) | 2026-11-08 |
| `sonnet` | `claude-sonnet-5-5` (Sonnet 5.5, on the Anthropic API) | [Claude Code model configuration](https://code.claude.com/docs/en/model-config) and [Claude models overview](https://platform.claude.com/docs/en/about-claude/models/overview) | 2026-11-08 |

Aliases move to newer models over time: when the documentation names a new model for an alias, update its row and date in the same change.

## Qualification

A model becomes **qualified** when, on the held-back tasks of the qualification suite (`tests/fixtures/qualification/suite.json`) with Hive guidance deployed:

1. its pass rate is at least 0.90, where the rate is the mean of its per-task pass rates and the report shows its standard error across tasks;
2. no task is in MISS, meaning the reference model passed every usable run of the task and the candidate a third or fewer of its own; the reference model is given with `--reference`;
3. on the tasks marked `compare_none`, the difference between the guidance arm and the no-guidance arm, paired by task, is not below minus two paired standard errors.

The verdict is `candidate` when all three hold, `candidate-partial` when the first two hold on a host without a no-guidance arm (Claude, Grok, Pi), and `not-candidate` otherwise, naming the failed condition. Tuning tasks never change the verdict. The 0.90 bar is the maintainer's choice, not an external standard. Per-task statistics follow Miller, "Adding Error Bars to Evals" (arXiv 2411.00640).

An attempt is **usable** when it ended `completed`, `blocked_permission` or `background_unverified` (the last two count as failures), its guidance hash and effort match the batch, and its served model is the requested one: both sides drop the provider prefix, the variant and a bracketed suffix such as `[1m]`, aliases resolve through «Aliases», and on OpenCode the variant must also match. A missing served model makes the attempt unusable on Claude, Grok, Pi and OpenCode; Codex never exposes it, so its requested model is used and reported as configured. Unusable attempts do not count, are reported by cause, and are launched again by the next `qualify run`.

Running a qualification is a model pilot, which root `AGENTS.md` «Measurement» keeps paused: it needs the user's authorization and a spending cap, and follows [the pilot procedure](pilot-procedure.md). With that authorization, from the repository root, build the runner outside the temporary directory, since installing arm A's guidance refuses a `go run` binary: `go build -o _support/workspace/bin/pilot ./tests/pilot`. Then:

1. `_support/workspace/bin/pilot qualify run --host <host> --models <candidates> --reference <model> --source . --guidance-source . --configured-model <value> --split heldback --runs 9 --out _support/workspace/<dir>` prints the sessions it would launch and launches none.
2. The same command with `--confirm-sessions <N>`, where N is the printed count, launches them. It records the batch in `<dir>/batch.json` and refuses, without launching, when the host, models, reference, efforts or the deployed or arm A guidance hash differ from that record; a redeploy or another checkout starts a new `<dir>`. Each model's effort comes from «Floor» when the row gives exactly one, none is passed for `—`, and otherwise `--effort <model>=<value>` is required.
3. `_support/workspace/bin/pilot qualify score --out _support/workspace/<dir> --reference <model> --source .` prints the report and the verdict, or exits 1 without a verdict when the batch is incomplete.

`--guidance-source` is needed only when arm A is scheduled (Codex or OpenCode with a `compare_none` task in the split). A launch that fails leaves `qualify run` with exit 1; calling it again relaunches that repetition.

## Known limits

- Claude, Grok and Pi have no no-guidance arm: the pilot runner gives Claude and Pi no isolated guidance home, and Grok's isolated home keeps the real `GROK_HOME`, where Hive installs its roles and hooks.
- Codex never exposes the served model, and an OpenCode session's trace does not carry it, so their sessions count under the configured model; qualification runs on OpenCode read it from the session export.
- Pi, Cursor and Antigravity inherit the session model: only floor models running there count.
