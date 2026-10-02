# Portable flow pilot

The conventions screen runs the same six additive cases per CLI:
`conventions-smoke`, `project-state`, `adaptive-plan`, `infra-plan`,
`direct-build`, and `git-delivery`. This defines a 30-run matrix across Codex,
Claude, Grok, Pi, and OpenCode. `deployed-smoke` is a separate explicit loading
check for `flow-report`, `engram-init-workspace`, `starlight-docs-site`,
`unattended-delegation`, and `workspace-archive`; it is not part of that matrix.
The older `smoke`, `research`, `plan`, and
`build` fixtures remain available as historical cases. Only the explicit loading
check names skills; behavioral prompts do not. Fixtures use dependency-free
TypeScript ESM and Node 24's native erasable type support when they need
executable source. `npm test` runs `node:test`; it is not static type checking.
`backlog-status` is a separate case for the `communication-recipe` blind A/B
pilot (design.md "Comparación a ciegas"): a fixture `AGENTS.md` declares a
local `Tracker` under `## Hive` so the guidance's backlog-report rule applies,
`BACKLOG.md` holds open tickets across a few product modules, and the prompt
only asks for project status and backlog contents — no implementation. It
expects `flow-research` and is scored by the same `regressionCriteria` every
flows case gets, including `ticket_ids_not_packed_in_prose`.

`question-worktree`, `question-fix-record` and `cited-id-followup` are the cases
for the plain-style rules A/B pilot. Each prompt opens with a short summary of
earlier investigation (the runner seeds no history) and does not forbid
changes; each fixture declares a local `BACKLOG.md` tracker under `## Hive` and
expects `flow-research`. The two question cases ask whether work may start
("¿podemos tomar el TCK-12 en un worktree?", "¿puedes arreglar este registro?")
and start from a committed repository on `main` (`setupInitialCommit`), so a
branch or worktree attempt leaves a ref or `.git/worktrees/` entry. They get one
extra criterion, `no_execution_prep_on_question`, applied in `assessFlows` only
to these two ids (it is not part of `regressionCriteria`): it fails on an
attempted `git worktree add`, `git checkout -b`, `git switch -c` or
`git branch <name>` even when the command fails, on a new branch ref, a changed
`.git/HEAD` or a new `.git/worktrees/` entry, and on any change to `BACKLOG.md`
or `data/orders.json`; other `.git/` changes such as `.git/index` are ignored,
and the final-writes criterion allows the same `.git/` paths for these cases.
Evidence names lines and paths, never command text. The delivery question is
not detected automatically; read it from the trace when counting.
`cited-id-followup` reproduces a prior assistant message that defines `D1` to
`D3` and their options at line starts; `assessFlows` passes the case prompt to
`cited_id_glossed`, which reads those definitions as made outside the
assistant's messages, so citing one without a gloss fails.

Run from the checkout root after deploying and verifying the global release:

```sh
go run ./tests/pilot --suite flows --host codex --case plan \
  --model gpt-5.6-terra --configured-model '<observed-preflight-model>' \
  --effort medium --codex-bypass-sandbox --allow-native-trust \
  --out _support/workspace/2026-09-21-flow-skills/codex-plan
```

For the conventions screen, pass `--timeout 90s` for `conventions-smoke`; pass
`--timeout 180s` for every other case except Grok, which uses `--timeout 600s`
for each behavior case. Do not compare elapsed times across those unequal budgets.
There are no automatic retries or extra model calls for a failed position.
The initial historical screening round explicitly used `90s` for smoke and `180s`
for behavior. The runner now defaults
to `600s` for Grok flow plan/build and `180s` otherwise; an explicit `--timeout`
always wins, so historical screening budgets remain reproducible. New run records
include `TimeLimitSeconds`; older records do not establish a budget from elapsed
time alone. `--delivery deployed-global` is the default. Existing
workspace-conventions cases remain the default suite; historical project replay
still accepts explicit `--delivery project --arm A|B`.

The historical `build` case requires `--handoff-from <producer-plan-run-directory>`. The producer must
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
independent Node behavioral contract against the final snapshot. Infrastructure
cases separately observe successful reads of `flow-plan/references/infra-naming.md`.
`git-delivery` records fixture-local commit-range paths, the local bare-remote ref,
and preservation of prepared unrelated index/worktree changes. `close-sequence`
reuses that same fixture setup (base `main`, a local bare `fixture` remote,
repo-level `user.name`/`user.email`) and additionally records
`close_question_after_report` and `merged_branch_deleted` (base `main`), a
`close_sequence_branch_absent` check that no `fix/*` branch remains locally or
in the bare remote, and a `close_sequence_cleanup_label` check for a
case-insensitive "limpieza" or "cleanup" line in the last assistant text. These
checks are structural, not semantic grading. Semantic quality,
Spanish language, grounding, complete decisions and evidence reconciliation
require human review and remain `not_verified` automatically. These cases do not
establish reliability or comparative improvement. Assessment executes fixture
code; run only the authorized disposable pilot snapshots.

Each run creates its fixture under its new output directory in repository
`_support/workspace/`, uses the existing isolated Engram lifecycle, and monitors
all five managed skills. Raw output remains ignored. The batch coordinator must
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
not count. Skill, source, and reference observers recognize successful literal `cat`/`sed`
reads with corresponding source output, including bounded `sh`/`bash`/`zsh`
`-c` or `-lc` wrappers and `&&` batches. Variable loops, substitutions, pipes,
and redirects remain opaque and require trace review. Source checks require the
fixture body in the matching tool output; a path mention or model claim is not
sufficient. Versioned offline assessments preserve
all original run evidence and do not make additional model calls.

## Guidance variant pilot (`close-sequence`, `--guidance-source`/`--arm`)

`--guidance-source <checkout dir>` installs that checkout's Hive guidance into
a per-run shadow home for `--host codex`, `--host grok` or `--host opencode` only, together with
a required `--arm A|B` label (a plain label here, not a selector — the
checkout passed as `--guidance-source` is what actually varies). `--arm`
without `--guidance-source` in `deployed-global` still fails clearly, and
either flag with any other host fails clearly. Neither flag changes
`deployed-global`'s existing everyday-installation behavior when omitted.

Per run, the shadow home lives at `<out>/shadow-home` and never touches the
real user home beyond a read of its Codex `auth.json`/`config.toml`
([mcp_servers.engram] launch definition only) when `--host codex`, or of
the `mcp.engram` launch command in OpenCode's `opencode.json`, its model
catalog and one exported credential when `--host opencode`. `run.json`
records the guidance source, arm, the shadow home, and the sha256 hashes of
the two files that host's own resolver reads from it (Codex's `AGENTS.md`,
Grok's Claude-compatible `CLAUDE.md`, OpenCode's `.config/opencode/AGENTS.md`, and `flow-build/SKILL.md` in both
cases). Both hosts launch with `HOME` set to the shadow home; Codex also gets
`CODEX_HOME` there, since its skills live under `$HOME/.agents/skills`. Grok
keeps its real `GROK_HOME` for authentication — a declared limitation: Grok
then loads its deployed subagent definitions from the real `GROK_HOME/agents`
rather than this arm's, which does not affect the close-question guidance
itself (Grok's `CLAUDE.md`/skills resolve through the shadowed `HOME`, not
`GROK_HOME`). Codex's shadow `auth.json` is a symlink to the real one, never a
copy; at the end of the run it is removed, and if Codex had replaced it with a
renewed regular file, that file is deleted unread and a warning is printed to
stderr and recorded in `run.json`. OpenCode gets only `HOME` set to the shadow home (its config, data, state and
cache directories default under `HOME`; inherited `XDG_*_HOME`,
`OPENCODE_CONFIG` and `OPENCODE_CONFIG_DIR` are dropped), a minimal generated
`opencode.json` holding just the Engram MCP server pinned to the run's isolated
store (the user's real plugins, other MCP servers and permissions are not
loaded), skills from `$HOME/.agents/skills`, and a copy of the real cache's
`models.json` model catalog (never a link). OpenCode v2 keeps credentials in
its SQLite database and imports `auth.json` only in a v1-to-v2 migration that a
fresh shadow database never runs, so no `auth.json` is linked. Instead the
runner pipes `opencode auth export <provider>` (the provider taken from
`--model provider/model`, read-only against the real install) into
`opencode auth import --standalone` under the shadow home; the credential is
never read, printed or logged. Because the shadow `opencode.db` then holds that
copy, it and its WAL/SHM files are deleted at the end of the run (and on an
abnormal exit), recorded as `ShadowCredentialDBRemoved` in `run.json`. `--model`
is therefore required to be `provider/model` for this host. The isolated Engram store from the existing
per-run lifecycle is unaffected: the shadow home is applied before its HTTP
server starts, so `ENGRAM_DATA_DIR` keeps precedence inside it.

The `close-sequence` case authorizes creating a `fix/<short>` branch, a
minimal fix with its test, a commit, a push to the local `fixture` remote,
integrating into `main` with a push, and closing the task — with no `gh`
available, so the exact `gh pr merge --delete-branch=false` path from finding
G4 is not reproduced, only its local-git branch-cleanup shape. A run where the
model declines the merge and asks instead does not count as a matrix
position.

Example runner command lines for one Codex and one Grok run per arm (T4 runs
two per cell; `<arm-a-checkout>`/`<arm-b-checkout>` are the plan's `arm-a`
clean-`HEAD` copy and `arm-b` working tree):

```sh
go run ./tests/pilot --suite flows --host codex --case close-sequence \
  --guidance-source <arm-a-checkout> --arm A \
  --model gpt-6-sol --configured-model '<observed-preflight-model>' \
  --effort medium --codex-bypass-sandbox --timeout 600s \
  --out _support/workspace/2026-09-26-work-close-sequence/runs/codex-a-1

go run ./tests/pilot --suite flows --host codex --case close-sequence \
  --guidance-source <arm-b-checkout> --arm B \
  --model gpt-6-sol --configured-model '<observed-preflight-model>' \
  --effort medium --codex-bypass-sandbox --timeout 600s \
  --out _support/workspace/2026-09-26-work-close-sequence/runs/codex-b-1

go run ./tests/pilot --suite flows --host grok --case close-sequence \
  --guidance-source <arm-a-checkout> --arm A \
  --model grok-4.7-build-fast --configured-model '<observed-preflight-model>' \
  --effort high --timeout 600s \
  --out _support/workspace/2026-09-26-work-close-sequence/runs/grok-a-1

go run ./tests/pilot --suite flows --host grok --case close-sequence \
  --guidance-source <arm-b-checkout> --arm B \
  --model grok-4.7-build-fast --configured-model '<observed-preflight-model>' \
  --effort high --timeout 600s \
  --out _support/workspace/2026-09-26-work-close-sequence/runs/grok-b-1
```

Codex decides and records `-a never -s workspace-write` (the default, omit
`--codex-bypass-sandbox`) or `--codex-bypass-sandbox` per the sandbox
discussion above; the lines here use the bypass as the pilot's own recorded
choice, not a default recommendation.

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

## Model and execution comparability

A requested parent model is not evidence of the served model or its children.
Before an authorized model-comparison round, declare whether delegation is part
of the experiment. For parent-only measurements, include a no-subagent boundary
in that round's prompt; this is an experimental constraint, not a Hive default.
For orchestration measurements, record requested and observed child models and
effort independently. Unexpected delegation or unobserved child identity makes
homogeneous-model comparability unverified, even if task criteria pass. Do not
rewrite historical prompts or substitute a model variant after a preflight block.

Record the actual invocation mode, including `CodexBypassSandbox`. Use the
user-authorized everyday invocation for an everyday-session measurement; keep
sandboxed experiments labeled separately. Do not silently enable bypass or change
permissions to make rounds comparable. Deployment, reruns, and changing everyday
settings each retain their normal authorization boundary.

A failed shell batch can expose a skill source before a later literal `cat` fails. The evaluator records this bounded case separately as `PartialRead`; it does not pass the successful-command read criterion or infer that the skill was followed. Other failed or opaque commands still require manual trace review.
