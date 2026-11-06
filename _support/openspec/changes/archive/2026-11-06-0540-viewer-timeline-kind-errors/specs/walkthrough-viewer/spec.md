## ADDED Requirements

### Requirement: Kind-dependent event fields
The event fields that only some event kinds accept SHALL be defined once, as a table from each field to the kinds that accept it, and the append-event validator, the CLI rejection message and the `append_event` input schema SHALL follow that table. A rejection of such a field SHALL name that field, never a neighboring one. When `hive view event add`, `hive view stage start`, `hive view event update` or `hive view stage update` rejects one with `forbidden_for_kind`, whether by local validation or by the viewer's response, the CLI message SHALL add the kinds that accept it. The full `append_event` operation contract SHALL declare, for each such field, the kinds that accept it; the per-flow briefs SHALL NOT grow because of it.

#### Scenario: Session version on a validation event
- WHEN a producer runs `hive view event add` with `kind: validation` and an `expected_session_version`
- THEN the CLI reports `invalid timeline input: forbidden_for_kind on expected_session_version; only kind stage accepts it` and nothing is sent.

#### Scenario: Viewer rejects a stage state on an event update
- WHEN the viewer answers `hive view event update` on a non-stage event with `forbidden_for_kind` on `stage_state`
- THEN the CLI message reports `(HTTP 400: forbidden_for_kind on stage_state; only kind stage accepts it)` and keeps its recovery advice.

#### Scenario: Source reference on a non-decision event
- WHEN an append-event request of kind `validation` carries `source_reference` and no `target`
- THEN it is rejected with `forbidden_for_kind` on `source_reference`.

#### Scenario: Contract declares the kinds
- WHEN a producer reads `hive view contract --operation append_event`
- THEN its input schema states, for each kind-dependent field, that only the kinds of the table accept it, and the brief for each flow keeps its size.

#### Scenario: Table and validator drift
- WHEN the validator accepts or forbids a kind-dependent field differently from the table
- THEN the contract test fails.
