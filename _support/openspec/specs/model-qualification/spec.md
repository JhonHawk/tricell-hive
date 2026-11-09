# model-qualification

## Purpose

TBD: created by archiving model-support-floor; update the purpose.

## Requirements

### Requirement: Writing target and support floor
The repository guidance SHALL direct that distributed content is written clearly for the writing target: the named models of the `execution`, `focused`, `reasoning` and `verifier` profiles of `integrations/agent-profiles.json`, listed in `_support/docs/harness-engineering/model-floor.md`. A rule aimed at a model outside the writing target SHALL need the recurrence the Measurement section sets; the guidance SHALL keep a single home for that threshold and for the evidence needed to remove a rule. The support floor SHALL hold the writing target plus the Haiku and flash-class models of `tests/pilot/flash_class.go`, each with its host, model, efforts, origin, whether it is a writing target, status (provisional or qualified), the date it was last verified or none, and the reference model of that verification; until qualified, each SHALL count as a provisional candidate. Floor membership SHALL be decided by model identity regardless of host, after removing the provider prefix and variant; each alias SHALL be listed with the full model IDs it resolves to and the date that was verified, and membership SHALL mean one of those IDs.

#### Scenario: A preventive rule for a smaller model
- **WHEN** an author proposes a new rule only because a model outside the writing target might misread an existing one, with no observed failure
- **THEN** the guidance does not support adding it.

#### Scenario: A floor model on an inheriting host
- **WHEN** a Pi session runs `opus` and fails a rule
- **THEN** that session counts as an observation on a floor candidate.

### Requirement: Recurrence evidence from floor candidates
A correction SHALL need the failure observed twice, independently, on floor candidates, at least once in a real session; the other observation MAY be an authorized qualification or pilot run that ran the current version of the governing rule, never the no-guidance arm, and a run directory or batch SHALL count as at most one observation. A failure observed only outside the floor SHALL NOT count. A session whose served model cannot be seen SHALL count under its configured model, named as such. A watch item SHALL record host, model, session or run directory, and date; authorized runs that exercise the rule SHALL count among the episodes that close a watch item; a correction's commit message SHALL name the session or run with host, model and date. The user's approval SHALL still adopt a correction at once. Only sessions and runs that ran the current version of the governing rule SHALL count. Relaxing or removing a rule SHALL remain the user's decision, resting on real-session evidence that the floor's candidates behave correctly without it; authorized runs MAY add to that evidence but SHALL NOT suffice alone.

#### Scenario: A failure seen in a real session and reproduced in a run
- **WHEN** a real session on `qwen3.8-flash` fails a rule and an authorized qualification run on the deployed arm reproduces it
- **THEN** the recurrence a correction needs is met.

#### Scenario: Failures seen only in runs
- **WHEN** two authorized runs fail a rule and no real session has shown the failure
- **THEN** the recurrence is not met and the failure stays a watch item.

#### Scenario: Repetitions of one batch
- **WHEN** two repetitions in the same run batch fail the same rule
- **THEN** they count as one observation.

### Requirement: Qualification suite and arms
The pilot runner SHALL run the tasks of `tests/fixtures/qualification/suite.json` with `--suite qualification` under `--arm deployed` (the user's deployed guidance) on every runner host, `--arm A` with `--guidance-source` on Codex, Grok and OpenCode, and `--arm none` on Codex and OpenCode, which prepares the same isolated home as arm A (credentials, memory isolation, protected-path fingerprints) without installing any Hive-managed file. It SHALL reject `--arm none` on Claude, Grok and Pi before launching, naming the limitation. Each qualification run SHALL record its task, split, arm, requested effort and a hash of the guidance it ran with, or `none`. A task marked `compare_none` SHALL use only checks that do not depend on Hive's own formats or files.

#### Scenario: The none arm on an unsupported host
- **WHEN** `--suite qualification --arm none` is passed with `--host grok`
- **THEN** the runner exits with an error naming the limitation and launches nothing.

#### Scenario: A clean none home
- **WHEN** a Codex qualification run uses `--arm none`
- **THEN** no file in its isolated home carries Hive's managed block marker or a Hive skill or agent.

### Requirement: Qualification scoring and verdict
`qualify score` SHALL count only usable attempts: those that ended `completed`, `blocked_permission` or `background_unverified`, whose guidance hash and effort match the batch record `qualify run` wrote, and whose served model, after removing on both sides the provider prefix, the variant and a bracketed suffix and resolving aliases as the floor document sets, is the requested one, with the same variant on OpenCode. A missing served model SHALL make an attempt unusable on Claude, Grok, Pi and OpenCode; on Codex the scorer SHALL use the requested model and report it as configured. The report SHALL count unusable attempts by cause. It SHALL fail a usable run that changed files outside its fixture. It SHALL compute each model and arm's pass rate as the mean of per-task rates with the standard error across tasks, and the difference between arm A and `none` on `compare_none` tasks as a paired per-task difference with its paired standard error. A held-back task SHALL be in MISS when the reference model's deployed arm passed all its valid runs and the candidate passed a third or fewer. The verdict SHALL use only held-back tasks: `candidate` when the deployed rate is at least 0.90, no task is in MISS, and on `compare_none` tasks the paired difference is not below minus two standard errors; `candidate-partial` when the first two hold and the host has no `none` arm; `not-candidate` otherwise, naming the failed condition. A usable run SHALL fail when a criterion returns a state the task does not list as accepted. It SHALL exit non-zero without a verdict when the batch record or the reference is missing, the reference differs from the batch record's, compared arms have unequal usable repetitions of a task, or a scheduled model, arm and held-back task has no usable repetition.

#### Scenario: Unequal runs
- **WHEN** arm A has 9 valid repetitions of a held-back task and the `none` arm has 8
- **THEN** `qualify score` exits non-zero and prints no verdict.

#### Scenario: A weak candidate on one task
- **WHEN** the reference passes 9 of 9 runs of a held-back task and the candidate passes 3 of 9
- **THEN** that task is in MISS and the verdict is `not-candidate`.

#### Scenario: The candidate beats the reference
- **WHEN** the reference fails every run of a held-back task and the candidate passes every run
- **THEN** that task is not in MISS.

#### Scenario: Host without a none arm
- **WHEN** a Claude model's deployed rate is 0.95 on held-back tasks with no MISS
- **THEN** the verdict is `candidate-partial`.

### Requirement: Served model in OpenCode qualification runs
A qualification run on OpenCode SHALL record the served model from the session's native export, read before the isolated home is cleaned up and with its own time limit, as `provider/model#variant`, and SHALL leave it `not_observed`, never inferred, when the export lacks it. A Codex run SHALL leave it `not_observed`.

#### Scenario: An export that names its model
- **WHEN** the OpenCode export of a qualification run names provider `opencode-go`, model `deepseek-v4.1-flash` and variant `high`
- **THEN** the run records `opencode-go/deepseek-v4.1-flash#high` as its served model.

### Requirement: Session confirmation before a qualification run
`qualify run` SHALL print the number of sessions this call would launch, including the reference model's and excluding repetitions that already have a usable attempt, with its breakdown, and SHALL launch none unless `--confirm-sessions` equals that number. Before launching it SHALL compare the host, models, reference, each model's resolved effort and the current deployed and arm A guidance hashes with the batch record and, when any differs, exit without launching and name both values. Guidance hashes SHALL not depend on the home directory the guidance is installed in. It SHALL take each model's effort from the floor document when the document gives exactly one, pass none when it gives `—`, and otherwise require it explicitly. A repetition with a usable attempt SHALL NOT be launched again, and an attempt that is not usable SHALL be kept as evidence while the next one gets a new attempt directory.

#### Scenario: Missing confirmation
- **WHEN** `qualify run` is called without `--confirm-sessions`
- **THEN** it prints the session count and launches no host.

#### Scenario: A provider fallback
- **WHEN** a run ends `completed` but its served model differs from the requested one
- **THEN** the scorer does not count it and the next `qualify run` launches that repetition again.

#### Scenario: Resuming a partial batch
- **WHEN** 5 of 18 scheduled sessions already have a valid attempt
- **THEN** `qualify run` asks for 13 and launches only those.
