## MODIFIED Requirements

### Requirement: Walkthrough ingestion and grouping
The server SHALL copy valid `walkthrough.json` files found under `_support/workspace/*/` of registered roots, with their referenced files, into its own store. Physical copies SHALL retain their folder/session-derived keys and per-copy resources. The viewer SHALL group normal visible copies by exact project key, declared host and `session.id`, across tickets, and SHALL present each copy as a chronological segment of that session. Different tuple values SHALL remain separate; titles and tickets SHALL NOT establish identity. It SHALL NOT ingest a manifest confirmed as tracked by a complete check of its owning Git index. Tracking SHALL be checked before unchanged-source shortcuts, including ancestor repositories and worktrees. A confirmed non-Git workspace SHALL remain supported. An unavailable or incomplete tracking check SHALL produce unknown status and a warning, never confirmed untracked status; ingestion and access in that unknown state SHALL NOT claim tracking exclusion or verified authorship. The project key SHALL be the normalized Git remote, or the repository root without one; the project label SHALL come from the root's `## Hive` `Project:` value, then `.engram/config.json` `project_name`, then the remote's last segment or the folder name.

#### Scenario: Multiple tickets in one session
- **WHEN** two visible normal copies share project, host and session ID but carry different tickets
- **THEN** the sidebar shows one session and its content includes both chronological segments with their own headings and resources.

#### Scenario: Same ticket in different sessions
- **WHEN** copies share a ticket but differ in project, host or session ID
- **THEN** they remain separate sessions.

#### Scenario: Invalid manifest
- **WHEN** a manifest has an unknown field or a path outside its folder
- **THEN** it is not copied, the error is listed in the index, and the server keeps serving.

#### Scenario: Force-added manifest
- **WHEN** a complete owning-index check confirms that a valid manifest is tracked even under an ignored workspace directory
- **THEN** no new stored copy is created from it.

### Requirement: Hierarchical React viewer navigation
The React shell SHALL list logical sessions directly under projects, with visible-session counts and compact session title, ticket summary, segment count and date presentation, without a ticket navigation level or connector branches. Projects SHALL initially open only when they contain a visible explicitly in-progress segment, with manual expansion taking precedence. Sidebar visibility, project expansion, archive filter and selection by physical segment key SHALL survive refresh in sessionStorage within the same tab and origin. Searching a segment ticket SHALL find its session and allow navigation to that segment without hiding its siblings. The sidebar brand header SHALL show a route icon. The session content SHALL stack visible segments chronologically with separators and individual headings. Each segment SHALL expose its informational work status, archive/restore, pin/unpin and delete actions; these SHALL target only its physical key. Lifecycle SHALL remain separate from source provenance and SHALL NOT be editable in the viewer. Explicit segment states SHALL be presented as in progress or closed; absent status SHALL be presented as unreported, never inferred as conversation termination. Archived segments SHALL be filtered before visible grouping; a session SHALL remain visible while it has a visible segment. The shell SHALL retain keyboard and mobile navigation and SHALL consume the existing Go API and shared walkthrough renderer. Simultaneous document controls, anchors, ARIA references, printing and cleanup SHALL remain scoped to their owning segment.

#### Scenario: Independent segment actions
- **WHEN** a user archives, pins or deletes one of two segments in a session
- **THEN** only that copy changes and its sibling retains its content and metadata; selection falls back within the session before another session.

#### Scenario: Different lifecycle and provenance
- **WHEN** a segment is closed while its source classification is unknown
- **THEN** the lifecycle indicator remains closed and the source uncertainty remains independently visible.

#### Scenario: Refresh after manual navigation
- **WHEN** the user changes project expansion, sidebar visibility, archive filter or segment selection and refreshes the same tab
- **THEN** the shell restores those choices instead of replacing them with initial defaults.

#### Scenario: Two walkthroughs with identical flow indexes
- **WHEN** two segments with their own capture tabs and carousel are mounted together
- **THEN** IDs and ARIA references remain unique, each interaction controls its own document, and unmounting retires its listeners.


### Requirement: Explicit walkthrough session lifecycle
The walkthrough manifest SHALL accept an optional `session.status` with values `in_progress` or `closed`. A missing status SHALL remain valid and SHALL be exposed by the viewer as `unknown`, never inferred from the manifest kind, age, source classification, or host process. Active walkthrough producers SHALL declare `in_progress`. Closing the current work through flow-close SHALL declare `closed`; this describes closure of that work, not approval or termination of the host conversation. Explicitly resuming it SHALL declare `in_progress` again.

#### Scenario: Historical final report
- **WHEN** an existing final report has no explicit session status
- **THEN** it remains valid and its status indicator states that no status was reported.

#### Scenario: Closing the current work
- **WHEN** flow-close completes, abandons or pauses a work item with a walkthrough
- **THEN** it updates only that work's manifest to closed, publishes JSON with HTML only when the executable is absent, and confirms the current revision before deleting its worktree or review artifacts.

#### Scenario: Cleanup without closure
- **WHEN** hygiene runs without closing the work item
- **THEN** it does not mark the walkthrough closed.

## ADDED Requirements

### Requirement: Explicit walkthrough delivery links
The manifest SHALL accept optional `delivery` entries containing `kind` (`pull_request` or `commit`), nonempty text `label`, and an absolute HTTPS `url` with an ASCII host (including punycode names or IP literals) and without credentials, literal spaces or control characters. Unicode host names SHALL be rejected without normalization; labels and paths MAY contain Unicode. Named hosts SHALL use ASCII letters, digits, dots, hyphens or underscores; bracketed hosts SHALL be valid IPv6 literals. Python and Go validators SHALL apply the same contract; existing manifests without the field SHALL remain valid in updated readers. Delivery SHALL identify known work links separately from `candidate`, and SHALL NOT infer integration or remote check status. New readers SHALL be distributed before producing manifests with the new field; older strict readers may reject it. Producers SHALL add only verified links available for the current work; closure SHALL update only the current artifact unit while preserving the shown candidate. No historical backfill SHALL run automatically.

The viewer header SHALL aggregate delivery links from visible accessible segments, deduplicated by exact URL, with two direct links and access to remaining links through an accessible dialog. Each segment and its standalone page SHALL expose its own links. An absence SHALL be shown as unregistered delivery rather than fabricated URLs; pending or failed document reads SHALL be distinguished from absence. Link rendering SHALL validate its destination again, treat labels as text, and isolate a new browsing context with `noopener noreferrer`.

#### Scenario: Session with several deliveries
- **WHEN** visible segments provide more than two distinct verified delivery URLs
- **THEN** the header shows two links and an accessible way to inspect all of them, while each segment retains its own association.

#### Scenario: Historical candidate without delivery
- **WHEN** an older manifest contains only `candidate.commit`, including `working tree`
- **THEN** the updated reader accepts it and shows no registered delivery without inventing a PR or commit URL.

#### Scenario: Unsafe delivery URL
- **WHEN** a delivery entry has a relative, executable, credential-bearing or otherwise invalid URL
- **THEN** both validators reject it without replacing the last valid copy, and a renderer never creates an unsafe external link.

#### Scenario: Closing a current work with a known PR
- **WHEN** closure has a verified implementation PR for the current walkthrough
- **THEN** it can record that delivery before rendering and confirmation while preserving the candidate and leaving other session segments untouched.

### Requirement: Viewer-first walkthrough publication
The walkthrough publisher SHALL determine whether the `hive` executable is available on PATH. When available, it SHALL validate and publish only the JSON manifest and its explicit lifecycle update, SHALL NOT build or write an HTML page, and SHALL leave any existing HTML unchanged. Producers SHALL register the artifact for the viewer and identify the JSON/visual evidence in the handoff. When the executable is absent, the publisher SHALL generate the standalone HTML fallback, preserving the existing preparation and recovery behavior. Server availability, registration errors and failed confirmation SHALL NOT enable HTML generation when the executable is present. Closure SHALL still confirm the current JSON/resources or retain the unit when confirmation fails. The artifact unit SHALL include HTML only when present; selecting the JSON manifest or an existing HTML page for retention SHALL retain its referenced captures and recordings. Existing HTML SHALL NOT be deleted automatically.

#### Scenario: Installed binary with stopped server
- **WHEN** `hive` is available but the viewer is stopped
- **THEN** publication validates or updates JSON without building or writing HTML, and failed closure confirmation retains evidence rather than switching output modes.

#### Scenario: Previous standalone page
- **WHEN** an HTML fallback already exists and a later publication finds `hive`
- **THEN** it leaves that HTML byte-identical and reports JSON as the current output.

#### Scenario: Binary absent
- **WHEN** no `hive` executable is available
- **THEN** publication generates the standalone HTML fallback from the validated JSON and keeps its existing atomic replacement and recovery checks.

#### Scenario: Retaining the viewer manifest
- **WHEN** the user selects walkthrough.json for retention
- **THEN** cleanup retains its referenced captures and recordings together with it, even without an HTML page.
