# Workspace behavior screening

These declarative fixtures are exercised by `go run ./tests/pilot`. The historical project-delivery screen compared two guidance configurations on Codex and Claude. The deployed-global screen uses the five installed hosts and their real global guidance. Both are observational: inherited instructions, plugins, tools, and memory prevent attributing every difference to skill packaging. See the implementation and pilot report (historical evidence omitted from public history) for recorded results and limitations.

## Deployed-global screen

Use `--delivery deployed-global` after deployment has been independently verified. This mode creates only the declared fixture files and disposable Git repositories; it does not call the deployment manager, inject project instructions or skills, select an A/B arm, or pass `--skill`. Daily authentication, plugins, MCP servers, and external memory remain available. Fixture roots stay outside the checkout's instruction ancestry. Record any inherited instructions outside the fixture; do not describe the environment as isolated.

Both `--model` and `--configured-model` are mandatory in this mode. The coordinator obtains configuration before the batch, records its provenance, and freezes those values for each launch. `ModelConfigured` and `ModelRequested` differ for an intentional override such as Codex Terra. No provider/model fallback occurs. `--provider` is Pi-only. `--effort` maps to native Codex configuration, Claude effort, Grok reasoning effort, or Pi thinking; omit it to retain a host default. OpenCode uses its native `provider/model#variant` model syntax instead. Observed response model and startup model remain separate; absence is `not_observed`.

```sh
go run ./tests/pilot --delivery deployed-global --host codex --case scratch --model gpt-5.6-terra --configured-model gpt-6-astra --effort medium --allow-native-trust --out /absolute/new-run-directory --timeout 180s
go run ./tests/pilot --assess /absolute/new-run-directory --source /absolute/hive-checkout
```

`--allow-native-trust` still requires explicit authorization and accepts only the exact native Codex trust insertion described below. It does not write configuration. The assessment command creates `criterion-assessment.json` with exclusive creation; it never replaces an assessment or rewrites `run.json`. Save any human adjudication separately with trace/path evidence.

Native launch interfaces are intentionally small:

| Host | Invocation-local behavior |
| --- | --- |
| Codex | `-a never exec --json --ephemeral -m MODEL -s workspace-write -C FIXTURE -`, optional `-c model_reasoning_effort=...`; prompt on stdin |
| Claude | `-p --model MODEL --output-format stream-json --verbose --no-session-persistence --permission-mode acceptEdits --allowedTools Read,Write,Edit,Glob,Grep,Bash`, optional `--effort`; prompt on stdin |
| Grok | `--cwd FIXTURE -p PROMPT --model MODEL --output-format streaming-messages-json --permission-mode dontAsk --allow Edit(./**) --allow Write(./**)`, optional `--reasoning-effort` |
| Pi | `-p --mode json --model MODEL --session-dir RUN/native-sessions`, optional `--provider` and `--thinking`; prompt on stdin |
| OpenCode | `run --standalone --format json --agent build --model MODEL PROMPT` |

These options do not disable daily extensions or MCP. Invocation-local approvals retain native deny rules; a denial is recorded as `blocked_permission`, not bypassed. Pi has no native path sandbox. Cwd, prompt scope, and tool selection are not OS-level confinement. Permission intervention, timeout, empty output with exit zero, malformed/truncated JSONL, host errors, and missing terminal events cannot become completed passes. OpenCode intermediate `step_finish` events with `tool-calls` are not terminal completion.

Run one unscored `--case smoke` per host before scoring. In deployed-global mode its distinct prompt asks for skill availability, explicitly requests consultation of the skill body, and asks for the temporary-placement rule without file changes. This is an explicit loading check, not spontaneous skill selection. The historical project-mode smoke prompt remains unchanged. Native Claude/Grok `system/init.skills` can prove advertised discovery; successful read events can prove body access; the model's answer alone proves neither. Global fingerprints establish presence/integrity, not that a host loaded the body. Keep source loading, discovery, successful body/reference reads, and self-report separate. The archive case requires successful skill reading before the first possible mutation; unavailable evidence requires human review.

The scored matrix is four cases (`placement`, `continuation`, `archive`, `scratch`) by five hosts, once each: **20 runs**. The coordinator may run at most three host workers concurrently; each worker owns one host and executes its four cases serially in fresh sessions/fixtures. Retain failures before any separately authorized diagnostic rerun. Smoke and corrective diagnostics are not scored repetitions.

Before launching workers, the coordinator captures a batch-wide configuration/guidance baseline. Each worker fingerprints shared immutable global instructions and the complete workspace-conventions directories (including references), but monitors only its own mutable configuration. This prevents a Codex trust insertion from falsely becoming a Claude/Grok mutation. Fingerprints include symlink identity and followed payload; no configuration content is saved. At batch end, reconcile all host-specific deltas and authorized native registrations against the batch baseline. Shared mutations are observed changes requiring attribution; a worker's observation alone does not identify the writer. Native histories, caches, credentials, Engram data, and all filesystem writes are outside this bounded fingerprint audit.

Each run retains `fixture.json`, initial/final inventories, protected fingerprints, native stdout/stderr, `events.json`, `run.json`, and regular final fixture files. `events.json` normalizes visible reads, writes, shell invocations, results, memory tools, discovery, models, terminal state, and native usage without conflating their evidence. Usage entries identify their event source; overlapping message and result totals must not be summed.

The assessor loads the fixture's evaluator-only criteria and compares file hashes, paths, links, and the normalized timeline. It returns criterion-level `pass`, `fail`, or `not_observed`, with overall `pass`, `fail`, or `manual_review`; smoke is `unscored`. Language/semantic judgments require manual review. Shell commands are not generally interpretable: recognized misplaced transient writes fail even if removed, while opaque operations require review. A clean final scratch inventory never proves correct temporary placement or cleanup. Record creation, exact contents/use and result 55, deletion, and preservation of both preexisting scratch files before adjudicating scratch as passed.

The sections below describe the retained historical A/B project mode. Their sixteen-run design and project injection do not apply to the deployed-global screen.

## Inputs

- **A — inline baseline:** `baseline/global.md`, frozen from the authored source before the split. SHA-256: `141b4baf16baf0c7222ca45d428ec684fbbb0942c0e9d90b8aaf227eb1be2e63`. Do not edit or distribute this historical fixture as current policy.
- **B — split candidate:** current [`global.md`](../../../content/guidance/global.md) plus [`workspace-conventions/SKILL.md`](../../../content/skills/workspace-conventions/SKILL.md). Record their exact hashes before running and retain the tested versions with the run evidence if they are not committed.
- `cases.json`: four independent setups, Spanish task prompts, and evaluator-only observable criteria. `files` maps relative paths to exact UTF-8 content; `cwd` selects the task directory. No real credentials or external services are involved.

## Prepare and load

Use installed native CLIs and existing authentication. Record host versions, model selection, permissions, and inherited environment. The selected models are Codex `gpt-5.6-terra` with `medium`, and Claude's resolved `claude-opus-5[1m]` with `high`. Do not silently substitute a model. Model selection and independently observed response identity are separate fields; record `not_observed` where events omit the latter.

Each run has a fresh temporary directory outside this checkout's instruction ancestry. Materialize only the fixture files and the selected instruction arm; initialize the specified disposable Git repositories before the run. A installs the frozen baseline body with the same delimiters but no Hive skill. B uses the manager's project-scoped installation, including native skill discovery. The evaluator's criteria and previous transcripts are not supplied to the task agent, but the daily configuration does not make external memory or filesystem reads impossible.

The pilot uses project instruction loading, not real user-global installation. Deterministic manager tests exercise user-global destinations in synthetic homes. Passing this pilot does not verify actual user-global loading.

Perform unscored loading checks first. For Claude, native init events expose skills and the initial resolved model; an invoked skill's delivered body can appear in the trace. For Codex, keep source-loading evidence distinct from the model's self-report. Never infer reading from a file listing or correct final outcome.

The runner fingerprints selected global instruction/configuration files before and after each run. This is a bounded monitor, not an audit of all writes by the host and its plugins. Native Engram startup may create `.git/engram-project-identity.json` in the disposable repo even when the model runs no tools; preserve and report it separately from agent task writes.

Codex 0.155.1 was observed persisting a trusted-project entry during native startup, even with a process-local trust override. The user explicitly authorized this narrow effect for the scored pilot. `--allow-native-trust` accepts only an exact new trusted-project table for that run's directory, including its table separator, with remaining configuration bytes unchanged. The runner never edits the real configuration itself. Without this explicit exception, or on any other monitored change, it stops the batch. The failed attempt to avoid persistence is retained as an unscored diagnostic, not a successful isolation mechanism.

## Running a case

From the checkout root, choose a new output directory under ignored scratch:

```sh
go run ./tests/pilot --host claude --case archive --arm B --model 'claude-opus-5[1m]' --out /absolute/new-output-directory --timeout 180s
go run ./tests/pilot --host codex --case placement --arm A --allow-native-trust --out /absolute/new-output-directory --timeout 180s
```

The Codex exception requires the user's explicit authorization; it is not a default permission inferred from these examples. Run `--case smoke --arm B` for the unscored loading check. JSONL traces, run metadata, inventories, and copied final fixture files stay in the private raw output directory. Temporary repo paths are retained in `run.json`; the runner does not purge them or native memory identities.

## Execute and inspect

Run each case once per arm and host in a fresh session, alternating A/B and B/A order across cases (sixteen scored runs total). Keep the host/model selection, permissions, prompt, and initial fixture files fixed within each pair. Record native model/tool visibility and memory interactions rather than claiming the everyday environment is isolated or immutable. Use the same 180-second deadline for each run. Native usage is recorded when available; no separate monetary cap or billing claim is inferred. A timeout, intervention, or missing terminal state is not a completed pass. Do not silently retry or repair a scored run; retain the failure before considering a separate diagnostic rerun.

For each run retain:

- Arm, case ID, date, host/version, resolved model, instruction and skill hashes, launch configuration, and effective instruction sources.
- Skill discovery visibility, actual instruction/reference reads with ordering relative to mutations, or `not_observed` where the host cannot expose them.
- Initial/final file inventories and hashes, writes/deletions, relevant tool events, human intervention, and terminal state.
- Each criterion as `pass`, `fail`, or `not_observed`, supported by paths, diffs, or trace events. Report discovery, reading, and outcome separately: a correct archive does not prove the skill was read.
- Native duration/usage metrics only if reported. Do not infer token savings or billing from word counts or wall time.

The archive case expects a candidate skill read before mutation. The other cases do not require a skill read; incidental reads are a cost/discovery observation, not automatically an outcome failure. Check file outcomes directly and resolve Markdown links; do not grade prose quality as part of this screen.

Retain raw transcripts in ignored scratch. Curate non-sensitive evidence and an interpreted report under the existing global-guidance work identifier. Do not publish raw host configuration or credentials.

## Decision boundary

Any unauthorized mutation, loss of prior/unique evidence, or overwrite blocks accepting the candidate. If basic placement or continuity regresses, correct the narrow failure before expanding. If skill loading or reading is not observable, report that dimension as inconclusive. Even sixteen completed passes demonstrate only these cases on the recorded host/model/configuration combinations; they do not establish reliability, superiority, or portability.

The first screen does not cover every scope convention, curation/promotion, destination collisions, sensitive evidence, ambiguous authorization, or installation/removal. Add discriminating cases or repetitions based on observed gaps before broader claims. Subjective quality comparisons require separately scoped blinded human review.
