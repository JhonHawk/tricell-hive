# walkthrough-viewer

## Purpose

A local server, `hive view`, that the user starts and stops to keep and browse the UI walkthroughs agents write, grouped by project and session with ticket metadata.

## Requirements

### Requirement: User-controlled walkthrough server
The manager SHALL provide `hive view start`, `stop`, `status`, `open`, `serve` and `register`. The server SHALL bind only to 127.0.0.1, SHALL run until the user stops it, and SHALL never be started by `register`, `status` or `open`.

#### Scenario: Second start
- **WHEN** a live server is recorded and the user runs `hive view start`
- **THEN** the command reports the running server and starts no second process.

#### Scenario: Workspace root that is not a Git repository
- **WHEN** `hive view register --dir` names a folder inside `<root>/_support/workspace/` and `<root>` is not a Git repository
- **THEN** the registry records `<root>`, and its walkthroughs are grouped under the project named by the root's `## Hive` `Project:` value.

#### Scenario: Idempotent registration
- **WHEN** a repository is already registered and `hive view register` runs in it
- **THEN** the command exits successfully without writing the registry.

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

### Requirement: Store retention and deletion
The server SHALL purge copies older than the configured age, 14 days by default, except pinned ones, and SHALL delete a copy on a request carrying its token and its own origin. Deleting SHALL never modify files in a repository. Every API read SHALL require the server token, and copied files SHALL be served only under a per-session capability derived from it.

#### Scenario: Read without token
- **WHEN** a request for the index, a session or a copied file lacks the token or the matching capability
- **THEN** the server answers 403 and returns no data.

#### Scenario: Delete without token
- **WHEN** a delete request lacks the server token or carries a foreign `Origin` or `Host`
- **THEN** the server answers 403 and the copy remains.

### Requirement: Single-use viewer access handoff
The CLI SHALL obtain a cryptographically random access code through a token-authenticated same-origin POST and pass only that code, never the server token, in the system opener URL fragment. Codes SHALL expire after 60 seconds, SHALL be consumed atomically once, and SHALL exist only in server memory. The exchange POST SHALL require a valid code and the server's own Host and Origin instead of the server token; this exception SHALL NOT relax authentication on any other existing API route. Exchange responses SHALL be non-cacheable. Tokens and codes SHALL NOT appear in CLI output, errors or logs.

#### Scenario: Open and reload
- **WHEN** the user opens the running viewer through the CLI
- **THEN** the page removes the code fragment before awaiting its exchange, receives the server token, and preserves authenticated access across a normal tab reload.

#### Scenario: Reuse or expiry
- **WHEN** a code is exchanged twice, concurrently, or at least 60 seconds after issuance
- **THEN** at most one exchange before expiry succeeds and all others grant no access.

#### Scenario: Invalid origin
- **WHEN** exchange has an absent, null or foreign Origin, or a disallowed Host
- **THEN** it returns 403 and grants no access.

#### Scenario: Overlapping opens
- **WHEN** an older exchange finishes after a newer open begins
- **THEN** the older result does not overwrite the newer access state; a failed current exchange offers reopening through the CLI.

### Requirement: Viewer command group help
The manager SHALL display the six viewer subcommands and exit successfully for `hive view --help`, `hive view -h` and `hive view help`, without resolving or writing server state or launching processes.

#### Scenario: Help without a running server
- **WHEN** the user requests viewer group help without a server or state directory
- **THEN** the command lists start, stop, status, open, register and serve and exits successfully without creating state.

### Requirement: Current source classification
The server SHALL derive `tracked`, `local` or `unknown` source status for every stored copy before publishing its initial or reconstructed view and on each complete scan. It SHALL place tracked copies outside the normal project/session list, preserve their files, keys, ingestion time, pins, archive state and existing retention behavior, and show an explanatory entry without rendering the document's instructions. Authenticated normal reads of a tracked copy's manifest or files SHALL return 409 without document bytes; authenticated archive/restore, pin and delete operations SHALL remain available. Missing sources without positive tracking evidence and failed classification SHALL be shown as unknown with a warning. Local status SHALL NOT claim verified agent authorship. The list and open document SHALL update source warnings when source status or reason changes, even when content hashes remain unchanged. Classification SHALL NOT require changes to stored per-copy metadata.

#### Scenario: Tracking changes without content changes
- **WHEN** an existing source becomes tracked or stops being tracked while its bytes remain unchanged and its file remains present
- **THEN** the next complete scan changes its placement without resetting its ingestion time or pin.

#### Scenario: Source missing or tracking unavailable
- **WHEN** the source is missing without positive tracking evidence or its tracking cannot be determined
- **THEN** its retained copy remains accessible with an explicit unknown-origin warning rather than a verified-authorship claim.

### Requirement: Safe and deterministic Git tracking lookup
The source classifier SHALL explicitly disable repository-configured `core.fsmonitor` callbacks for its Git read processes without changing stored Git configuration. Inherited pathspec environment variables SHALL NOT alter the classifier's selected pathspec interpretation. Index paths resolving to the same filesystem object as a manifest SHALL be recognized as tracked despite case differences; distinct files on case-sensitive filesystems SHALL NOT be conflated. Prefix selection SHALL account for case aliases in parent directories. These controls SHALL preserve the existing bounded-output and unknown-status policies.

#### Scenario: Repository filesystem monitor
- **WHEN** Git accepts a repository configured with a filesystem-monitor callback and the classifier reads its index
- **THEN** the callback is not executed and the index still determines tracking status.

#### Scenario: Inherited literal pathspec semantics
- **WHEN** the parent environment sets `GIT_LITERAL_PATHSPECS=1` and a manifest is tracked
- **THEN** the complete query still identifies that manifest as tracked rather than producing a false local result.

### Requirement: Open-file hardlink rejection
Before reading an ingestion manifest or referenced resource, including each hashing and copying read, the server SHALL inspect the opened file descriptor and reject a regular file with more than one hard link. It SHALL reject unavailable link-count metadata explicitly rather than claiming the check passed. This policy SHALL NOT require the file owner to equal the server's effective user. Rejection SHALL NOT install a partial copy or replace the last valid stored copy.

#### Scenario: Resource replaced after validation
- **WHEN** a validated resource is replaced with a hard-linked regular file before a hashing or copying open
- **THEN** descriptor inspection rejects it before reading its bytes and leaves the last valid stored copy intact.

### Requirement: Full-width desktop screen capture preview
On screen, a desktop capture preview in a strip wider than 520 CSS pixels SHALL occupy the strip's usable width. Narrow strips SHALL retain the existing next-capture peek, and additional captures SHALL remain horizontally accessible. Mobile capture sizing, carousel access and the existing print arrangement SHALL remain unchanged. The skill asset and embedded viewer stylesheet SHALL remain byte-identical. Previously generated standalone pages SHALL NOT be rewritten automatically.

#### Scenario: Wide screen strip
- **WHEN** a newly generated standalone walkthrough or the current viewer displays desktop captures in a wide strip
- **THEN** one capture occupies its usable width rather than half, without adding page-level horizontal overflow.

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

### Requirement: Confirmed lifecycle ingestion before artifact cleanup
The manager SHALL support a directory-targeted viewer status query through its existing authenticated API access. The closing flow SHALL confirm that the exact source/session's current closed revision was ingested before removing its review artifact unit. An unavailable viewer, unsupported runtime, ingestion error or bounded confirmation timeout SHALL retain that unit and report the unresolved confirmation, without starting or restarting the viewer. Queries SHALL NOT expose authentication tokens or file capabilities.

#### Scenario: Reopened walkthrough has an old closed copy
- **WHEN** an earlier closed copy exists but the source is reopened and later closed with changed content
- **THEN** the earlier closed copy does not count as confirmation of the current revision.

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

### Requirement: Reversible copy archiving
The server SHALL support authenticated, same-origin POST and DELETE requests to `/api/sessions/<key>/archive` to archive and restore stored copies, including excluded tracked-source copies, with 204 on success. Archive state SHALL be optional per-copy `archived` metadata, defaulting to false, and SHALL appear in the index for normal and excluded entries. It SHALL survive restarts, index reconstruction and source-content replacement, independently of pinning, ingestion time, TTL and manifest session status. Archiving SHALL NOT edit source files or exempt an unpinned copy from normal TTL deletion. Archived copies SHALL be hidden by default; an icon-only toggle beside Projects SHALL show or hide them with a tooltip and accessible label.

#### Scenario: Archive and restore a pinned copy
- **WHEN** the user archives and later restores a pinned copy
- **THEN** it disappears from and returns to the default list, remains pinned throughout, and its source manifest and work status remain unchanged.

#### Scenario: Archived unpinned copy expires
- **WHEN** an archived, unpinned copy exceeds the configured retention age
- **THEN** normal TTL purging removes it.

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
