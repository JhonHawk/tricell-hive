# Portable flow pilot

Four cases per CLI: explicit loading of three skills, research, persisted planning,
and implementation of another CLI's plan. Behavioral prompts never name a skill.
Fixtures use dependency-free TypeScript ESM and Node 24's native erasable type
support. `npm test` runs `node:test`; it is not static type checking.

Run from the checkout root after deploying and verifying the global release:

```sh
go run ./tests/pilot --suite flows --host codex --case plan \
  --model gpt-5.6-terra --configured-model '<observed-preflight-model>' \
  --effort medium --codex-bypass-sandbox --allow-native-trust \
  --out _support/workspace/2026-09-21-flow-skills/codex-plan
```

Use cases `smoke`, `research`, `plan`, and `build`. The initial screening round
explicitly used `90s` for smoke and `180s` for behavior. The runner now defaults
to `600s` for Grok flow plan/build and `180s` otherwise; an explicit `--timeout`
always wins, so historical screening budgets remain reproducible. New run records
include `TimeLimitSeconds`; older records do not establish a budget from elapsed
time alone. `--delivery deployed-global` is the default. Existing
workspace-conventions cases remain the default suite; historical project replay
still accepts explicit `--delivery project --arm A|B`.

Build requires `--handoff-from <producer-plan-run-directory>`. The producer must
have completed and retained exactly one `.plan.md` below `_support/sessions/`.
The importer verifies snapshot hashes, copies that plan and recursively linked
Markdown records from `_support/sessions/` or `_support/docs/`, and records
provenance in `handoff.json` and `run.json`. It rejects symlinks, path escapes,
missing local references, collisions and instruction files. Source links must
exist in the receiving fixture and never copy source from the producer. Use
relative Markdown links for companion work records. External links stay links.
A failed import stops before launching a model; report that matrix position as
blocked, do not synthesize a replacement plan.

```sh
go run ./tests/pilot --assess <run-directory>
# Preserve original assessments when updating observational parsers:
go run ./tests/pilot --assess <run-directory> --assessment-file criterion-assessment.v3.json
```

Assessment emits per-skill native advertisement and successful source-read
observations separately from task criteria. Plan assessment also records successful
`flow-plan/references/plan-format.md` reads and, when trace writes are observable,
whether the read preceded the first observed plan write. Missing reference reads
remain `not_observed`; unobservable ordering remains `not_verified`. These fields
do not substitute for plan quality or task outcome. Build additionally executes a bounded,
independent Node behavioral contract against the final snapshot. Semantic quality,
Spanish language, grounding, complete decisions and evidence reconciliation
require human review and remain `not_verified` automatically. These cases do not
establish reliability or comparative improvement. Assessment executes fixture
code; run only the authorized disposable pilot snapshots.

Each run creates its fixture under its new output directory in repository
`_support/workspace/`, uses the existing isolated Engram lifecycle, and monitors
all four managed skills. Raw output remains ignored. The batch coordinator must
wait for all writers and audit exact fixture leakage in everyday Engram after
the round; local per-run cleanup does not replace that audit.

## Scope correction after the first round

The first 20-position matrix used the original prompt without an explicit fixture
root boundary. Native agents could follow inherited parent-repository context;
several cases read sibling/parent material and one wrote a research record in the
real parent session. That round remains evidence of this limitation; it was not
rerun or relabeled after the correction.

New flows prompts explicitly identify the freshly created fixture as the sole
task repository, exclude parent/sibling/evaluator artifacts, and permit intended
global skill and reference reads. This is an instruction boundary, not OS-level
isolation, and does not grant or change permissions. A prompt-construction test
checks the boundary; live effectiveness remains unverified until another
explicitly authorized round.

Native source observations also recognize successfully delivered skill bodies:
OpenCode's matched `skill_content` result and Claude's synthetic source message
after a successful matching Skill invocation. A launch acknowledgment alone does
not count. The plan reference observer recognizes literal successful `cat`/`sed`
reads with corresponding reference source output; variable loops and wrappers
remain opaque and require trace review. Versioned offline assessments preserve
all original run evidence and do not make additional model calls.

## Grok completion budget

Grok 1.0.40 with `grok-4.6` / high effort remained active at the old 180-second
cutoff: native timing showed fast permission decisions and tool completion,
followed by ongoing model generation. One bounded diagnostic with a 600-second
budget completed planning in 411.1 seconds, including a retained plan, native
review and terminal success. The plan was in the fixture's `_support/sessions/`.
See the diagnosis (historical evidence omitted from public history).

The larger default is a pilot budget, not a change to Grok's everyday settings,
model, reasoning effort, permissions, or MCP configuration. It covers plan/build
completion checks; it does not guarantee completion in ten minutes or claim
faster execution. A build check using the new default (without an explicit timeout flag) completed
in 182.0 seconds and passed its independent behavioral contract. These two
observations establish completion for those runs, not general reliability.
For matched latency comparisons, set the same explicit budget and disclose any
censored runs. Never relabel the original 180-second timeouts as successful.
