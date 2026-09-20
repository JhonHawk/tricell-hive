# Skill evaluation pilot

This directory contains the first reproducible pilot for Hive skills. It applies the
method from [Testing Agent Skills Systematically with Evals](https://developers.openai.com/blog/eval-skills)
to three skills, plus three cases for a fourth:

- `language-rules`
- `flow-plan`
- `flow-report`
- `task-routing` (direct-route task record only)

The dataset has 33 cases: 10 per skill for the three pilot skills (explicit, implicit,
contextual and negative controls) plus 3 for `task-routing` covering only the
direct-route task record (explicit, implicit, negative; no contextual case). The pilot runs each selected case twice, using the same model and
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
Engram memory or global project instructions. This original pilot runner gives each
case a fresh temporary workspace, stages generated skills there, and sets isolated
`HOME`, `CODEX_HOME` and `PI_CODING_AGENT_DIR` values. The isolated
`HOME/.agents/skills` symlinks to the staged tree so generated cross-skill references
resolve inside the fixture. In this historical adapter, Codex auth is symlinked and Pi
auth was exposed by symlink; the current activity-screen adapters have a different,
explicit auth policy documented below. Do not run `/deploy-global`, install packages,
or print existing authentication.

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

## Activity-skills screen

`activity-skills-screen.json` is a separate two-case screen for the activity-scoped
`flow-research` candidate. It does not replace or rewrite the historical 33-case pilot
above. `implicit-research` maps to `task-routing` in the baseline arm and `flow-research`
in the candidate arm. `standalone-save` requests automatic `workspace-conventions`
selection and a retained session finding in both arms. The prompt does not prescribe the
artifact path or slug.

Run each arm and harness as a separate invocation, preserving the source root, model,
effort and case identifiers in the output. The source root must contain the generated
core and skill trees for that arm. To prepare the baseline used by the initial screen:

```bash
BASELINE_ROOT=_support/workspace/YYYY-MM-DD-activity-skills-pilot/baseline
mkdir -p "$BASELINE_ROOT"
git archive 31d00c03c48b2ce8ad4230059634bbb77d5ddb51 | tar -x -C "$BASELINE_ROOT"
```

Replace the folder date with the local run date. Then a Codex baseline run is:

```bash
uv run --no-project python harness/evals/runner.py \
  --manifest harness/evals/activity-skills-screen.json \
  --source-root "$BASELINE_ROOT" \
  --arm baseline \
  --harness codex \
  --model gpt-6-astra \
  --effort medium \
  --case implicit-research \
  --case standalone-save \
  --run \
  --output _support/workspace/YYYY-MM-DD-activity-skills-pilot/codex-baseline.json
```

For the candidate, use the candidate repository root as `--source-root` and set
`--arm candidate`; use a separate output file. For each harness, keep its actual model
and effort fixed between arms. The activity runner accepts `codex`, `claude`, `pi`,
`grok`, and `opencode`. These are five native CLI adapters with isolated configuration;
this does not prove identical invocation semantics or production parity. Do not count
deployment fixture tests as in-harness behavior runs.

The selected source root supplies both the generated `harness/agents-skills/` tree and
its generated core files: `harness/AGENTS.md` for Codex and Pi, `global/CLAUDE.md` for
Claude Code and Grok, and a workspace `AGENTS.md` for OpenCode. Each invocation stages
the selected core and generated skills in fresh harness-specific configuration roots.
The case workspace starts as a clean local Git repository. User hooks, MCP servers,
Engram memory, and deployed global project instructions are isolated; the provider's
model service remains external. Compare source and generated skill hashes, core hashes,
staged-tree hash, configuration hash and model metadata in each JSON result. A requested
model name is not a resolved model unless the trace exposes that identity.

Authentication remains outside the staged harness configuration. Codex and Grok use
the existing `auth.json` through a symlink into their isolated roots. OpenCode's native
auth is held in its SQLite database in the isolated data home, not through the former
`auth.json` symlink ([v2 storage guidance](https://opencode.ai/v2/docs/troubleshooting),
[upstream database implementation](https://github.com/anomalyco/opencode/blob/dev/packages/core/src/database/database.ts)).
Pi copies its auth file into the isolated root with mode `0600` and excludes it from
retained artifacts. Claude Code uses `CLAUDE_CODE_OAUTH_TOKEN` from the environment or
reads the existing default macOS Keychain access token and supplies it only in the child
process environment; it is never written to the temporary `HOME`, logs, or retained case
artifacts. A custom `CLAUDE_CONFIG_DIR` or `CLAUDE_SECURESTORAGE_CONFIG_DIR` requires an
explicit environment token. If no supported credential is available, classify the run
as blocked before behavior evaluation. Never print or persist a credential.

Claude Code runs headlessly with `stream-json`, no session persistence, an empty strict
MCP configuration, `acceptEdits`, no permission prompts, and WebSearch, WebFetch and
Task disallowed. `HOME`, `TMPDIR` and `CLAUDE_CODE_TMPDIR` point to the disposable case
root. This configuration can still produce a terminal run with a behavior failure or
an unverified artifact; terminal completion alone is not a pass. The Claude skill must
be reachable in the actual isolated `$HOME/.claude/skills` tree. Align `--add-dir` and
the exact-tree `Edit` restriction with that runtime path; targeting a second staged
skills directory can leave the loaded skill and permitted artifact tree mismatched.

All non-Codex adapters are launched under the macOS sandbox. The OpenCode adapter is
designed to start a native local IPC listener after provider access. In the 2026-09-20
pilot, the v3 and v4 real-provider attempts stopped at routing before listener startup;
neither model execution nor listener binding was observed. A separate isolated
fake-provider probe loaded config but initially returned an empty plugin list and empty
standalone model list. Polls against the same private server showed plugins `0 → 86`
and providers `3 → 4` by two seconds; the exact fake model appeared later. This suggests
startup catalog population can be delayed, but the exact requested route was not sampled
through the initial transition, so this does not prove a startup race is the sole cause
of `no-route`. A health response alone is not route-readiness evidence. The proposed
private-server startup/readiness redesign was abandoned before implementation. The
user's preferred next validation is through real harnesses after global deployment; that
deployment has since occurred, but this diagnostic produced no model-reaching OpenCode
result. No global config was changed during this diagnostic. The profile's SBPL
`localhost` filtering does not prove that a listener is loopback-only. Treat network
containment as `not_verified` until the live listener is inspected and confirmed on a
loopback address; stop the run if it listens on a non-loopback interface.

Use [`activity-skills-screen-rubric.md`](activity-skills-screen-rubric.md) for blind
research-answer review and standalone-save placement review. Freeze the rubric before
opening arm-labeled outputs; record the rubric version/hash and reveal the arm mapping
only after scores are fixed. Activation, answer quality, and placement are separate
measures. Equal scores in a small fixture do not prove equivalence or improvement.
The 2026-09-20 exploratory review-refuter model-agent scores were Codex 10/10 for both
arms, Claude 7/10 for both, Pi 7/10 for both, and Grok 7/10 baseline versus 8/10
candidate. This was a model agent, not the human review required by the frozen rubric,
so the scoring procedure deviated and no human validation occurred. The non-Codex
answers shared errors or unsupported reasoning in percentage-point interpretation,
candidate attribution/counterfactual wording, and the independent-review recommendation.
Grok's candidate gained one source-grounding point in one pair only; this does not
establish a causal or robust benefit. No quality improvement has been demonstrated.

The initial 2026-09-20 Codex JSON reports remain `unknown` where the original parser did
not recognize `nl -ba` reads or `file_change` writes. The manual trace audit (historical evidence omitted from public history)
records how those original traces were interpreted. Later adapter versions produce new
reports; do not rewrite the initial outcomes. Preserve each run's machine result and
report manual observations separately. A final answer or a started tool call alone is
not proof of a completed read or write.

The pilot measurement record (historical evidence omitted from public history)
keeps the initial Codex checkpoint separate from follow-up observations. Manual
inspection verified four Codex cells and four each for Pi and Grok across research and
standalone-save. Claude research reached the required source reads in both arms. An
independent trace audit verified both corrected native Claude standalone-save cells:
the skill and three references were read before the write, and both findings used the
correct session date/slug and plural filename. Earlier failed Claude attempts remain
historical. The correction was observed red→green in its regression, 57 runner unit
tests passed, and review approved the exact flags. OpenCode v3's historical route and v4's
current-catalog route each failed all four arm/case attempts with provider
`no-route`/model-unavailable responses before inference; no listener was observed. The
v4 reports carried `high` effort metadata, but the CLI effort flag was not sent, so effort
was not controlled for those attempts. The original bounded screen remains unqualified.
Follow-up testing in real harnesses after global deployment has now completed; the
outcomes are summarized below and do not qualify the candidate for a full-migration design.
The original Codex parser outcomes remain `unknown`; curated
manual observations are separate from machine-reported evaluator statuses. The local
follow-up evidence index is retained at
[`evidence-index.json`](../../_support/backup/2026-09-20-activity-skills-pilot-evidence/followup/evidence-index.json).

The requested global deployment completed on 2026-09-20 at 13:35:36 local across
Claude Code, Codex, OpenCode, Grok Build and Pi (`deploy-global.sh --apply`, exit 0).
Generated-tree checks and targeted SHA-256 checks passed; the [deployment record](../../_support/backup/2026-09-20-activity-skills-pilot-evidence/native/deploy/deploy-findings.md)
lists their scope. This confirms delivery, not harness behavior. Bounded native
execution is complete, but the results do not qualify the candidate:

- **Codex:** research and standalone-save traces passed independent review for the
  observed cells.
- **Claude Code:** research did not load the actual `flow-research` skill. Its table
  arithmetic is 2/4 as written, or 2/3 excluding C4 from the denominator; the separate
  “5/5 effects” claim is uncalibrated. The corrected standalone-save rerun read the two
  refs that held its first write, then produced the artifact; independent review verified
  the refs preceded the effective write.
- **Pi:** research completed but made an absence claim without `code-search.md`, so that
  claim is not verified. Standalone-save read the intended skill and three refs but also
  read a sibling Claude finding before writing, so content and placement are not
  independently attributable.
- **Grok Build:** research loaded its skill and fixture sources, but shared Engram
  retrieved Codex output. Standalone-save reached the correct dated output after the
  hook-required `.claude` refs were read, but also read Codex/Claude findings and Engram
  memories; neither result is an independent quality measurement. A duplicate lost its
  process stdout log to `O_EXCL`; its native history was later archived without
  re-auditing it, so the duplicate remains `not_verified`. No further rerun is
  planned in this milestone.
- **OpenCode:** research completed with a sanitized export but lacked `code-search.md`
  for its absence claim and repeated the denominator error and a causal overclaim.
  Standalone-save's final stop/idle, skill/reference reads, write and output passed
  independent review. Earlier inherited-working-directory and headless-permission
  launch failures remain historical.

Scoped cleanup completed: 72 native CLI-history files with hashes and three sanitized
OpenCode exports were preserved under
[`native/history`](../../_support/backup/2026-09-20-activity-skills-pilot-evidence/native/history/).
Owned Claude/Pi project directories, two Codex rollout files, four Grok sessions and two
prompt-history files, and three OpenCode sessions were removed through their native
interfaces. OpenCode export checks returned `Session not found` for all three removed
sessions. Engram relation-audit metadata 419/420 was automatically orphaned and left
untouched; global history indexes and SQLite stores were not manually edited, so the
cleanup does not claim every internal index was erased.

Rollback is incomplete because previous Codex/OpenCode `AGENTS.md` files and harness
agent/command trees were not backed up; their exact prior bytes/hash proofs are
unavailable. The pilot therefore remains unqualified.
