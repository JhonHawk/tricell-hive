## MODIFIED Requirements

### Requirement: Declared source footprint without token claims

The size check SHALL report the global core separately from explicit simple-query, question, retained plan/build and E2E scenarios. Each scenario SHALL name its read conditions and source files, include the entrypoints and mandatory procedures selected by those conditions, and count each source once. Missing required sources SHALL fail measurement. The change report SHALL distinguish prose, references and executable code, and SHALL retain a before/after comparison using the same source set so a scenario-definition change is not misreported as savings. Existing scenario measurements SHALL remain identifiable. Source bytes and characters SHALL NOT be presented as actual tokens, observed native delivery or model reliability. Native metadata/rendering, repeated reads, project inputs, history and host policy SHALL remain explicit exclusions. Each skill entrypoint and reference SHALL retain the existing 16,000-byte limit; no new aggregate quota is implied.

#### Scenario: Coordination reference is added
- **WHEN** E2E requires a new situational reference
- **THEN** the E2E scenario includes it, and the report does not count moving prose from an entrypoint into that reference as a reduction in the required source union.

#### Scenario: Comparable baseline
- **WHEN** the measured scenario's source list changes
- **THEN** the report separates the methodology change from the before/after measurement on matching paths and reports new sources explicitly.

## ADDED Requirements

### Requirement: Bounded native coordination observation

A behavior observation SHALL identify the exact candidate installed through the existing delivery mechanism, the host and model identities, evidence of selected source reading, owner/verifier assignments, human interventions, artifacts and terminal outcome. The newly authorized E2E case SHALL be one native run bounded by a total 20-minute deadline including setup and process termination, with no automatic retry or second host run. It SHALL exercise two independent units, a justified verifier correction, a controlled owner replacement and verified local integration in a disposable repository, without remote publication. Its launcher SHALL explicitly configure and test the native skills/delegation/communication permissions and session persistence the case needs, without inheriting assumptions from another case name or expanding every pilot's permissions. Injected fixture faults SHALL be labeled and SHALL invalidate affected evidence. A single run SHALL NOT establish reliability, qualification or cross-host support.

Temporary installation SHALL require its own covered effect, preserve the current trial and unrelated state, and use the manager's conflict-aware plans for installation and restoration. The pilot SHALL preserve authentication and native histories, isolate fictitious memory and verify cleanup of only its own resources. It SHALL NOT silently overwrite later changes to meet the deadline. If restoration or behavior evidence is incomplete, its status SHALL remain explicit and the affected acceptance criterion SHALL remain unverified.

#### Scenario: Catalog without source reading
- **WHEN** the host lists a skill or role but the trace contains no supported evidence that its relevant instructions were read
- **THEN** the observation does not report that instruction as exercised.

#### Scenario: Deadline or incomplete return
- **WHEN** the deadline is reached or a required behavior is not observed
- **THEN** the case stops its model work and known process groups, preserves necessary evidence, reports failed/blocked/inconclusive as supported, and launches no second attempt without renewed authorization.
