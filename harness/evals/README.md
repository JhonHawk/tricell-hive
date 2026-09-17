# Skill evaluation pilot

This directory contains the first reproducible pilot for Hive skills. It applies the
method from [Testing Agent Skills Systematically with Evals](https://developers.openai.com/blog/eval-skills)
to three skills:

- `language-rules`
- `flow-plan`
- `flow-report`

The dataset has 30 cases: 10 per skill, with explicit, implicit, contextual and
negative controls. The pilot runs each selected case twice, using the same model and
reasoning setting for the comparison:

| Harness | Version | Model | Effort | Repeats |
|---|---|---|---|---:|
| Codex | 0.154.0 | `gpt-5.6-luna` | `high` | 2 |
| Pi | 0.85.1 | `openai-codex/gpt-5.6-luna` | `high` | 2 |

The harness, version, requested model, resolved model when exposed by the trace,
effort, case and repetition are one experimental unit. If the trace does not expose
a resolved model, it remains unknown; the runner does not infer it from the request.
Results from different harnesses or models are reported side by side; they are not
interchangeable scores.

## Run the pilot

Run from the repository root. The runner is standard-library Python and does not
install packages. Check its current interface first:

```bash
uv run --no-project python harness/evals/runner.py --help
```

The canonical commands are:

```bash
uv run --no-project python harness/evals/runner.py \
  --manifest harness/evals/cases.json \
  --harness codex \
  --model gpt-5.6-luna \
  --effort high \
  --timeout 180 \
  --repeat 2 \
  --run \
  --output _support/workspace/YYYY-MM-DD-skill-evals-pilot/codex.json

uv run --no-project python harness/evals/runner.py \
  --manifest harness/evals/cases.json \
  --harness pi \
  --model openai-codex/gpt-5.6-luna \
  --effort high \
  --timeout 180 \
  --repeat 2 \
  --run \
  --output _support/workspace/YYYY-MM-DD-skill-evals-pilot/pi.json
```

Replace `YYYY-MM-DD` with the local run date. Omit `--run` to inspect the planned
argv and selected cases without starting a harness. Use `--list` to print the
manifest's case ids, and repeat `--case <id>` to run a bounded subset.

Use a fresh invocation for each harness. The runner creates a disposable workspace
per case, with fixture files committed as a clean local Git baseline. This commit
is confined to the disposable fixture and never publishes repository work. Save each result JSON and its retained case artifacts under the ignored run
directory `_support/workspace/YYYY-MM-DD-skill-evals-pilot/` with `--output`; do not
copy raw traces into a commit. The result JSON contains the normalized trace and the
raw stdout/stderr captured by the runner. Keep that result manifest with the run so
every case and repetition remains auditable.

The isolated process must not load the user's deployed Hive hooks, MCP servers,
Engram memory or global project instructions. The runner gives each case a fresh
temporary workspace, stages the generated skills there, and sets isolated `HOME`,
`CODEX_HOME` and `PI_CODING_AGENT_DIR` values. The isolated
`HOME/.agents/skills` symlinks to the staged tree so generated cross-skill references
resolve inside the fixture. Authentication, when present, is
exposed by a symlink into the isolated home; it is never copied or printed. Do not
run `/deploy-global`, install packages, or print existing authentication.

Codex is invoked with `--ignore-user-config`, `--ignore-rules`, `--ephemeral`,
`--sandbox workspace-write`, `--json` and `--skip-git-repo-check`. Pi is invoked
with `--no-extensions`, `--no-context-files`, `--no-prompt-templates`,
`--no-themes`, `--no-session`, `--mode json` and `--no-approve`. Explicit Pi cases
also receive `--skill` with the staged target skill path; other categories exercise
the harness's skill discovery behavior against the same staged directory. Explicit
Codex cases append an instruction to read and apply the staged target path in the
prompt. Pi `--skill` registers the path; it does not inject the body. In Pi 0.85.1,
`disable-model-invocation` hides gated skills from the automatic catalog, while a
user `/skill:name` command expands the body. This adapter does not send that command.
Explicit delivery is asymmetric, and natural planning requests without an observed
read remain unverified for activation. These cases do not establish native invocation
or full Hive routing equivalence between harnesses. Every
Pi case is additionally wrapped with macOS
`/usr/bin/sandbox-exec`; the runner fails closed when that executable is absent and
the generated profile limits writes to the disposable run root and case workspace
(plus `/dev/null`).
The model still provides the provider service, so this is an isolated adapter run,
not an offline model test.

Provider built-in `.system` skills remain outside the repository snapshot even when
the selected skill and host extensions are disabled. Record that boundary in any
interpretation; do not silently infer production parity.

This isolation is deliberate, but it has a known boundary: the provider's model
service remains outside the repository snapshot. The pilot therefore measures the
staged skill package and adapter behavior. It does not prove full production harness
parity, global hook parity, MCP behavior, memory behavior or portability to a
harness/model combination that was not run.

## Case contract

`cases.json` is versioned test data. Each case contains:

| Field | Meaning |
|---|---|
| `id` | Stable case identifier; unique within the manifest. |
| `skill` | Target skill under test. |
| `category` | `explicit`, `implicit`, `contextual` or `negative`. |
| `prompt` | User request sent to the isolated harness. |
| `fixtures` | Relative files created before the case. Paths cannot escape the case workspace. |
| `protected_files` | Fixture files that must remain byte-identical. |
| `expected_activation` | Whether the target skill is expected to load. |
| `required_reads` | Skill/reference paths whose completed reads must be observed. |
| `allowed_writes` | Relative paths the case permits the model to write. An empty list is read-only. |
| `required_files` | Files that must exist after the run. |
| `file_contains` | Required strings per file, where the case needs content checks. |
| `output_contains` | Required strings in visible harness output; matching may include intermediate assistant text, not only the final message. |

An existing required file does not count as output from the case: the runner checks
the before and after snapshots. A started tool call or final prose claim does not
count as completed evidence. A terminal event is required to treat a process as
finished. Read evidence recognizes direct `cat` and simple `sed -n` commands,
including simple `sh`, `bash` or `zsh` wrappers. Whitelisted read-only `&&` chains
credit every successful read; semicolon/newline sequences credit only the final
command or final `&&` suffix. Every segment must be a supported read. Other shell
programs remain unknown; `cat --help` and `cat --version` never prove a read. The evaluator does not execute canonical semantic validators for skill
quality, plan correctness or report correctness.

## Result statuses

Every case and repetition records independent dimensions for activation, process and
outcome, plus the raw normalized trace and hashes:

| Pilot status | Use |
|---|---|
| `pass` | The observed evidence satisfies the case contract. |
| `fail` | The harness completed, and the skill or its output violated the contract. |
| `blocked` | A declared prerequisite or authorization prevents the case from running. |
| `not_verified` | The available evidence cannot establish the criterion. |

The runner's raw JSON currently uses `passed`, `failed` and `unknown` for its
dimension checks. When curating the pilot report, map `passed` to `pass`, terminal
contract violations to `fail`, unavailable prerequisites to `blocked`, and
insufficient evidence (including `unknown` process state) to `not_verified`. Keep
the raw value in the evidence; never hide the normalization.

The runner also separates a skill failure from an infrastructure failure. A missing
required read, unexpected write or wrong output after a terminal run is a skill or
case-contract failure. A missing executable, timeout, unavailable provider, malformed
stream or absent terminal event is infrastructure or incomplete evidence; report it
as `blocked` or `not_verified` according to the evidence instead of blaming the
skill. Never turn an exit code alone into a pass.

The initial pilot has no LLM judge. Style, semantic quality, plan quality, report
quality and visual quality remain `not_verified` until a human labels them using
[`RUBRIC.md`](RUBRIC.md). A deterministic file check cannot establish that an answer
was clear or that an HTML report is visually readable.

## Provenance and review

Each result records the source snapshot and artifact hashes. Review the source and
artifact hash pairs before interpreting a result; a hash proves identity, not
correctness. The current result keys are:

| Hash | What it identifies |
|---|---|
| `runner` | Runner source captured at process initialization. |
| `source_skill` | Canonical skill tree under `global/skills/`. |
| `generated_skill` | Generated universal skill tree under `harness/agents-skills/`. |
| `staged_skills` | Staged skill tree at the end of the case. |
| `config` | Harness, model, effort, argv and isolated-run configuration. |
| `model` | Requested model, resolved model from the trace when available, and CLI version observed by the runner. |

Preserve the raw trace long enough to reproduce the finding, then keep only curated
evidence according to the workspace support-artifact rules. This pilot has no
recorded model results until the two harness runs are explicitly executed; the
README does not predeclare pass or fail counts.

For a baseline comparison, run the same manifest, harness, version, model, effort
and repeat count against the exact source snapshot. Change one input at a time. Keep
case-level results instead of averaging away a negative control or a failed
repetition. A two-repeat pilot reveals variation; it does not establish a stable
success rate.

The first diagnostic run and its limitations are recorded in the
dated pilot report (historical evidence omitted from public history).

## Replay read observability offline

Run from the repository root, with the original ignored raw reports still present:

```bash
uv run --no-project python harness/evals/replay.py \
  --baseline _support/archive/audits/2026-09-16-skill-evals-pilot.json \
  --output _support/workspace/2026-09-16-skill-evals-pilot/new-read-replay.json
```

The output must be a new file. Replay verifies manifest and raw-report SHA-256 hashes
against the curated baseline, then applies the current read parser. It preserves
recorded file checks, protected-file checks, curated unexpected writes and process
failures. It does not rerun models or inspect retained output artifacts. Timeouts and
missing terminal events remain `not_verified`. The comparison records its own source
hashes and never rewrites the baseline or raw traces. Exit zero means the replay
completed, not that every case passed.

The observability follow-up (historical evidence omitted from public history)
records the actual offline comparison.
