## ADDED Requirements

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
