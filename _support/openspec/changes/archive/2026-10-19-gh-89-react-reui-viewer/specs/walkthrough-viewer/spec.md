# walkthrough-viewer

## ADDED Requirements

### Requirement: Explicit walkthrough session lifecycle
The walkthrough manifest SHALL accept an optional `session.status` with values `in_progress` or `closed`. A missing status SHALL remain valid and SHALL be exposed by the viewer as `unknown`, never inferred from the manifest kind, age, source classification, or host process. Active walkthrough producers SHALL declare `in_progress`. Closing the current work through flow-close SHALL declare `closed`; this describes closure of that work, not approval or termination of the host conversation. Explicitly resuming it SHALL declare `in_progress` again.

#### Scenario: Historical final report
- **WHEN** an existing final report has no explicit session status
- **THEN** it remains valid and its sidebar badge indicates that no status was reported.

#### Scenario: Closing the current work
- **WHEN** flow-close completes, abandons or pauses a work item with a walkthrough
- **THEN** it updates only that work's manifest to closed and renders the page before deleting its worktree or review artifacts.

#### Scenario: Cleanup without closure
- **WHEN** hygiene runs without closing the work item
- **THEN** it does not mark the walkthrough closed.

### Requirement: Confirmed lifecycle ingestion before artifact cleanup
The manager SHALL support a directory-targeted viewer status query through its existing authenticated API access. The closing flow SHALL confirm that the exact source/session's current closed revision was ingested before removing its review artifact unit. An unavailable viewer, unsupported runtime, ingestion error or bounded confirmation timeout SHALL retain that unit and report the unresolved confirmation, without starting or restarting the viewer. Queries SHALL NOT expose authentication tokens or file capabilities.

#### Scenario: Reopened walkthrough has an old closed copy
- **WHEN** an earlier closed copy exists but the source is reopened and later closed with changed content
- **THEN** the earlier closed copy does not count as confirmation of the current revision.

### Requirement: Hierarchical React viewer navigation
The React shell SHALL distinguish project, ticket and session levels with visible nesting and connector branches, and SHALL display session lifecycle separately from source provenance. Explicit states SHALL be presented as in progress or closed; absent status SHALL be presented as unreported. The shell SHALL retain keyboard and mobile navigation, and SHALL consume the same Go server and shared walkthrough renderer.

#### Scenario: Different lifecycle and provenance
- **WHEN** a session is closed while its source classification is unknown
- **THEN** the lifecycle badge remains closed and the source uncertainty remains independently visible.
