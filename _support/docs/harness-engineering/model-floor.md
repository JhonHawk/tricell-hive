# Model support floor

This page is the canonical list of the models Hive's distributed guidance is written for and of the candidates whose observed failures count toward a correction. Root `AGENTS.md` «Guidance structure» points here for the writing target, and «Measurement» points here for floor membership; neither restates the list. `tests/content/model_floor_test.go` keeps the tables below aligned with `integrations/agent-profiles.json` and with the flash-class pilot list in `tests/pilot/flash_class.go`.

## Two sets

- **Writing target.** Distributed content is written clearly for the named models of the `execution`, `focused`, `reasoning` and `verifier` profiles in `integrations/agent-profiles.json`: the models that run Hive's delegated roles, from execution to reasoning and verification, which include the frontier models. The profiles do not set a host's main-session model: the user picks it, and only a floor model counts there. A rule aimed at a model outside this set needs the recurrence that root `AGENTS.md` «Measurement» sets; a plausible misreading by a smaller model is not enough.
- **Support floor.** The writing target plus the Haiku and flash-class models the user kept on 2026-10-11 and confirmed on 2026-11-08. A failure observed on any floor model, provisional or qualified, counts toward a correction as «Measurement» sets; a failure seen only on a model outside the floor does not.

On hosts that inherit the session model (Pi, Cursor, Antigravity) or declare none (OpenCode), delegated roles run whatever model the session runs. Hive's guidance is not written for that open set (decision of 2026-11-08: write for the role-profile models and correct on observed failures instead); only floor models count there.

## Membership and normalization

A model belongs to the floor by its identity, whatever host runs it:

1. Remove the provider prefix and the variant from both the session's model and the table's model before comparing them: `opencode-go/claude-haiku-5-5#high` and the row `opencode-go/claude-haiku-5-5` both become `claude-haiku-5-5`, so a session on `claude-haiku-5-5` through another provider matches that row.
2. Resolve an alias to the full IDs listed under «Aliases»; membership means being one of those IDs. An alias that resolves to an older model on another provider, such as `sonnet` on Amazon Bedrock, is not a floor model there.
3. When a session's served model cannot be seen (the Codex and OpenCode traces do not record it yet), use its configured model and name it as configured.

## Floor

| Host | Model | Efforts | Origin | Writing target | Status | Verified | Reference |
| --- | --- | --- | --- | --- | --- | --- | --- |
| claude | `opus` | `high` | `reasoning`, `verifier` | yes | provisional | none | — |
| claude | `sonnet` | `high` | `execution`, `focused` | yes | provisional | none | — |
| codex | `gpt-6.1-sol` | `medium`, `high` | `execution`, `reasoning`, `verifier` | yes | provisional | none | — |
| codex | `gpt-6.1-terra` | `medium` | `focused` | yes | provisional | none | — |
| grok | `grok-4.6` | — | `verifier` | yes | provisional | none | — |
| opencode | `opencode-go/claude-haiku-5-5` | `high` | floor of 2026-10-11 (`flashModels`) | no | provisional | none | — |
| opencode | `opencode-go/deepseek-v4.1-flash` | `high` | floor of 2026-10-11 (`flashModels`) | no | provisional | none | — |
| opencode | `opencode-go/qwen3.8-flash` | `medium` | floor of 2026-10-11 (`flashModels`) | no | provisional | none | — |
| opencode | `opencode-go/glm-5.3-flash` | `high` | floor of 2026-10-11 (`flashModels`) | no | provisional | none | — |

A row is **provisional** until a qualification run passes; **Verified** then records its date and **Reference** the reference model of that run.

## Aliases

| Alias | Resolves to | Source | Verified |
| --- | --- | --- | --- |
| `opus` | `claude-opus-5-5` (Opus 5.5, on the Anthropic API) | [Claude Code model configuration](https://code.claude.com/docs/en/model-config) and [Claude models overview](https://platform.claude.com/docs/en/about-claude/models/overview) | 2026-11-08 |
| `sonnet` | `claude-sonnet-5-5` (Sonnet 5.5, on the Anthropic API) | [Claude Code model configuration](https://code.claude.com/docs/en/model-config) and [Claude models overview](https://platform.claude.com/docs/en/about-claude/models/overview) | 2026-11-08 |

Aliases move to newer models over time: when the documentation names a new model for an alias, update its row and date in the same change.

## Qualification

A model becomes **qualified** when, on held-back tasks of the qualification suite with Hive guidance:

1. it passes at least 90% of its runs;
2. no task is in MISS, meaning the reference model passed every run and the candidate none; the reference model is given for each run with `--reference`;
3. on the tasks marked `compare_none`, the guidance arm is not worse than the no-guidance arm by more than two standard errors.

A host without a no-guidance arm can reach only the first two conditions. The 90% bar is the maintainer's choice, not an external standard, and the statistics are open for revision when the tool is planned.

The qualification tool is not built yet: it is planned in `_support/openspec/changes/model-qualification-tool/`. Running it will be a model pilot, which root `AGENTS.md` «Measurement» keeps paused: it needs the user's authorization and a spending cap, and follows [the pilot procedure](pilot-procedure.md).

## Known limits

- Claude, Grok and Pi have no no-guidance arm: the pilot runner gives Claude and Pi no isolated guidance home, and Grok's isolated home keeps the real `GROK_HOME`, where Hive installs its roles and hooks.
- The Codex and OpenCode traces do not record the served model yet, so their sessions count under the configured model.
- Pi, Cursor and Antigravity inherit the session model: only floor models running there count.
