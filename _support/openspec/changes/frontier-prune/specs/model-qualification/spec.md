## MODIFIED Requirements

### Requirement: Writing target and support floor
The repository guidance SHALL direct that distributed content is written clearly for the writing target: Claude Opus 5.5 and OpenAI gpt-6.2-sol, listed in `_support/docs/harness-engineering/model-floor.md`. It SHALL direct that distributed text written or edited uses «never» and «always» only for invariants (authorization, secrets, irreversible effects and preservation) and states scope and reason for every other rule; existing text SHALL change only when an edit touches it. A rule aimed at a model outside the writing target SHALL need the recurrence the Measurement section sets; the guidance SHALL keep a single home for that threshold and for the evidence needed to remove a rule. The support floor SHALL hold the writing target, the models one tier above it (Claude Fable 5.2, gpt-6.2-astra) and one tier below it (Claude Sonnet 5.5, gpt-6.1-terra), and Grok 4.7 with its variant `grok-4.7-build-fast`, which the user runs in main sessions and kept by decision although benchmarks place Grok 4.7 below Sonnet 5.5, each with its host, model, efforts, origin, whether it is a writing target, status (provisional or qualified), the date it was last verified or none, and the reference model of that verification; until qualified, each SHALL count as a provisional candidate. Every model a role profile of `integrations/agent-profiles.json` names SHALL have a floor row. Haiku, Luna, flash-class models and superseded versions SHALL NOT be floor candidates; another main-session model SHALL be added only by the user's decision, never when it is a Haiku, Luna or flash-class model. The «Two sets» section SHALL list the writing-target models as the backquoted Model values of its «Floor» rows. Floor membership SHALL be decided by model identity regardless of host, after removing the provider prefix and variant; each alias SHALL be listed with the full model IDs it resolves to and the date that was verified, and membership SHALL mean one of those IDs.

#### Scenario: A preventive rule for a smaller model
- **WHEN** an author proposes a new rule only because a model outside the writing target might misread an existing one, with no observed failure
- **THEN** the guidance does not support adding it.

#### Scenario: A floor model on an inheriting host
- **WHEN** a Pi session runs `opus` and fails a rule
- **THEN** that session counts as an observation on a floor candidate.

#### Scenario: An absolute in a rule that is not an invariant
- **WHEN** an edit to distributed guidance writes "never" in a rule about how to word a report, which is not an invariant
- **THEN** the repository guidance directs that the rule state its scope and reason instead.

#### Scenario: A failure on a small model
- **WHEN** a real session on `claude-haiku-5-5` fails a rule
- **THEN** that session does not count toward a correction.

### Requirement: Recurrence evidence from floor candidates
A correction SHALL need the failure observed twice, independently, on floor candidates, at least once in a real session; the other observation MAY be an authorized qualification or pilot run that ran the current version of the governing rule, never the no-guidance arm, and a run directory or batch SHALL count as at most one observation. A failure observed only outside the floor SHALL NOT count. A session whose served model cannot be seen SHALL count under its configured model, named as such. A watch item SHALL record host, model, session or run directory, and date; authorized runs that exercise the rule SHALL count among the episodes that close a watch item; a correction's commit message SHALL name the session or run with host, model and date. The user's approval SHALL still adopt a correction at once. Only sessions and runs that ran the current version of the governing rule SHALL count. Relaxing or removing a rule SHALL remain the user's decision, resting on a comparison of real sessions before and after the change with `go run ./tests/pilot --audit-session`, on the rule's own stated reason verified as no longer holding in current documentation or the installed version, or on the current prompting guidance of a writing-target model's vendor stating the behavior as a default or the rule as scaffolding to remove; authorized runs MAY add to that evidence but SHALL NOT suffice alone. Removing a restatement, a dangling pointer or a circular reference SHALL NOT count as relaxing a rule, since the rule stays in its home.

#### Scenario: A failure seen in a real session and reproduced in a run
- **WHEN** a real session on `gpt-6.1-terra` fails a rule and an authorized qualification run on the deployed arm reproduces it
- **THEN** the recurrence a correction needs is met.

#### Scenario: Failures seen only in runs
- **WHEN** two authorized runs fail a rule and no real session has shown the failure
- **THEN** the recurrence is not met and the failure stays a watch item.

#### Scenario: Repetitions of one batch
- **WHEN** two repetitions in the same run batch fail the same rule
- **THEN** they count as one observation.

#### Scenario: A caution whose reason no longer holds
- **WHEN** a rule states it exists because a host lacks a capability, and the host's current documentation and installed version show the capability exists
- **THEN** the user may decide to remove the rule on that evidence alone.

#### Scenario: A restatement removed
- **WHEN** a skill repeats part of a rule leaf it also links, and the repetition is replaced by the link
- **THEN** no recurrence or session evidence is needed, since the rule still acts from its leaf.

#### Scenario: A removal backed by vendor guidance
- **WHEN** the user decides to remove an instruction to re-check finished work, and the Opus 5 prompting guide says such instructions cause over-verification
- **THEN** the removal meets the evidence the Measurement section sets.
