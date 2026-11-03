## MODIFIED Purpose

A user-controlled local server for shared work, native producers, declared stages, canonical tasks and references, immutable documents and view-scoped evidence. Publication uses only the protocol v3 API; there is no standalone HTML fallback.

## REMOVED Requirements

### Requirement: Explicit walkthrough session lifecycle
**Reason:** the standalone HTML fallback and its manifest status contract are removed (#319, D10). v3 lifecycle is covered by «Declared flow stages and immutable documents».

### Requirement: Full-width desktop screen capture preview
**Reason:** it governs the standalone walkthrough page and the byte-identical skill asset, both removed (#319, D10). The React viewer's capture sizing is governed by «Ticket-scoped evidence views».

## MODIFIED Requirements

### Requirement: Self-describing protocol and confirmed-input hygiene
Offline contract usage/schemas/limits SHALL describe protocol/storage v3 exclusively. New CLI and viewer SHALL require compatible v3 peers for publication. No v1/v2 reader, endpoint, migration or imported evidence reconstruction SHALL be part of the v3 runtime. v3 receipts, immutable blobs and atomic state transitions SHALL remain durable after restart or uncertain commit.

One canonical publication procedure SHALL own collaboration/native identity, meaningful snapshots, states before user stops, explicit resumption and readable summaries. Consumers SHALL link to it. Children SHALL return evidence to the main producer. Viewer identity SHALL NOT change Engram runtime identity. Hygiene SHALL retain authoritative plans, current review inputs, uncertain uploads and unique evidence.

#### Scenario: v3 restart
- WHEN a v3 publication commits and the service restarts
- THEN IDs, hashes, actor audit and operation receipts remain resolvable without a legacy adapter.

#### Scenario: Old writer
- WHEN a v1 or provisional v2 publisher sends a request to a v1 endpoint of v3
- THEN the endpoint does not exist and nothing is mutated.

### Requirement: Walkthrough ingestion and grouping
The protocol v3 runtime SHALL serve API-owned works from a v3 state directory exclusively. It SHALL NOT scan or import manifests, API v1 or provisional v2 sessions or copy preferences. Any name outside the v3 inventory SHALL make startup fail as unknown state. Project roots SHALL be explicitly registered in v3. Unknown or corrupt v3 state SHALL fail without cleanup. API commits SHALL finalize blobs and synchronize files/directories before success and SHALL reconcile outcome_unknown after uncertain durable commits.

#### Scenario: Clean start
- WHEN v3 starts on a new state directory and roots are registered
- THEN its index is initially empty.

#### Scenario: Earlier state selected
- WHEN v3 is pointed at a state directory that holds an earlier viewer's files or lacks the v3 storage marker
- THEN startup fails with invalid_storage before creating or altering files, preserving that state.

### Requirement: Open-file hardlink rejection
Retained file-based upload operations SHALL inspect the opened descriptor before hashing/reading referenced local resources, rejecting regular files with multiple hard links or unavailable required link-count metadata. Rejection SHALL not publish partial content or replace valid content. API endpoints SHALL receive bounded bytes, never follow publisher-supplied filesystem paths.

#### Scenario: Resource replacement
- WHEN a local referenced resource is replaced by a hard-linked regular file before reading
- THEN descriptor checks reject the read and no new publication is committed.

### Requirement: Confirmed lifecycle ingestion before artifact cleanup
For v3, closure cleanup SHALL confirm the exact committed work/version, state, revision/artifact IDs, hashes and sizes through authenticated API readback before removing owned disposable upload inputs. An unavailable runtime, mismatch or uncertain commit SHALL preserve inputs and report the limit without starting/restarting the viewer. Tokens/capabilities SHALL never appear in evidence.

#### Scenario: Uncertain publication
- WHEN exact committed readback cannot be established
- THEN review inputs remain available and no successful cleanup/publication is claimed.

### Requirement: Explicit walkthrough delivery links
v3 API delivery events SHALL accept only verified PR/commit links, distinguish candidate metadata from delivery and never infer integration. URLs SHALL be absolute HTTPS with valid ASCII hosts and no credentials, spaces or controls; labels SHALL render as text, destinations SHALL be validated on rendering, and new contexts SHALL use noopener noreferrer. The v3 compact header SHALL expose deduplicated canonical work delivery entities by ID in its accessible delivery popover; no links SHALL be fabricated from candidate text.

#### Scenario: Unsafe URL
- WHEN a delivery contains an executable, relative or credential-bearing URL
- THEN validation rejects it and no unsafe browser link is created.

### Requirement: Viewer-first walkthrough publication
Substantive publication SHALL use only the compatible v3 API and confirmed readback. A missing executable or a stopped/incompatible viewer SHALL preserve inputs and report publication unavailable without starting a daemon or producing any other format. Previously generated HTML pages and user-selected artifact units SHALL not be deleted or rewritten by the removal of the renderer. The agent SHALL NOT open the viewer on its own: at each stop where the viewer holds something for the user to review (a plan at its approval stop, a research report, a UI walkthrough, including the build's first rendered view) it SHALL ask whether to open it, until the user accepts once in the conversation; after acceptance it SHALL open the work once and not ask again unless the user asks.

#### Scenario: Binary present, server incompatible
- WHEN publication finds a v1 server
- THEN it requests the coordinated v3 update and preserves inputs without HTML substitution.

#### Scenario: User declines
- WHEN the user declines to open the viewer at a review stop
- THEN the viewer stays closed and the question is asked again at the next review stop.

### Requirement: Fixed session toolbar
The selected v3 work title and management actions SHALL stay visible outside its scrolling timeline/document canvas. The header SHALL show the title with its edit action on its first row, and the work summary, clamped to two lines with an expand control, beside a right-aligned toolbar on its second row; it SHALL NOT show the project name or the creation date. The toolbar SHALL hold the state, the validation count and the ticket, delivery and participant counters, and the archive and more actions. Deep-linked stage/event headings SHALL remain visible below the toolbar. Keyboard actions, unbroken long names through bounded local scrolling, narrow layouts without page overflow (the toolbar wrapping below the summary) and print omission of management controls SHALL remain supported.

#### Scenario: Narrow long title
- WHEN a valid unbroken title exceeds a narrow viewport
- THEN local keyboard-accessible title scrolling exposes it without page-level overflow or hiding actions.

#### Scenario: Long summary
- WHEN the work summary exceeds two lines
- THEN the header shows two lines with an expand control, and the canvas does not repeat the summary.

### Requirement: Required native session registration
New timeline sessions SHALL require the resolved registered project, canonical host and actual native session ID, and MAY carry a work key. The server SHALL derive a distinct viewer ID from the project, host, native session ID and the effective work key: the given work key, or, when absent, the creating operation ID. It SHALL enforce exact uniqueness across concurrent requests, retries and restarts. Sessions stored before work keys existed SHALL keep their IDs. Missing identity SHALL be rejected rather than synthesized. Native IDs SHALL remain opaque and separate from viewer storage IDs. Existing registration SHALL not overwrite metadata or reopen closed/deleted work. API registration SHALL recover persisted root/project bindings before hashing identity; temporary metadata failure SHALL not replace a known project key with a path fallback, and unresolved/conflicting identity SHALL fail explicitly.

#### Scenario: Resume or concurrent creation
- WHEN two clients from the same native session register the same project/host/native-session tuple with the same work key
- THEN both receive the same viewer ID and existing session content remains intact.

#### Scenario: Two works in one native session
- WHEN one native session registers two different work keys, or two creations without a work key with different operation IDs
- THEN the server creates two viewer sessions.

#### Scenario: Separation
- WHEN the native ID is equal but project or host differs
- THEN the server creates separate viewer sessions.

### Requirement: Manual viewer retention
API content SHALL NOT be purged due to age. Archive SHALL change visibility only; pin SHALL not be necessary for preservation. The viewer commands SHALL NOT accept an age or TTL option. Explicit deletion and bounded incomplete-staging recovery SHALL remain separate from age retention.

#### Scenario: Inactive session survives
- WHEN an unpinned archived session ages beyond fourteen days and the server restarts
- THEN its documents and images remain readable until explicit authorized deletion.

### Requirement: Shared work producers and lead authority
Protocol v3 SHALL preserve the original source tuple, its work key and the opaque viewer_session_id as work identity, accepting additional real native producers only by explicit join request and lead acceptance within the same registered project. A request SHALL NOT confer membership. Native lead transfer SHALL be version-conditional to an accepted member. Explicit local recovery SHALL atomically accept a chosen pending same-project request and reassign lead under expected version, durable operation receipt, reason and audit. Recovery SHALL NOT infer process death, stop processes, reopen terminal works, change original source or alter stage/evidence state. The previous lead SHALL remain a collaborator without lead privileges. Creation with the same tuple and work key SHALL remain idempotent. Joining SHALL NOT merge historical works or reassign IDs.

Pending, foreign-project and unauthorized publishers SHALL fail unchanged. Producers SHALL use their actual native IDs, never another session's or invented IDs. Provenance SHALL be declared rather than represented as cryptographic host authentication; local capability remains the authentication boundary. No historical actor or content SHALL be synthesized.


#### Scenario: Continue elsewhere
- WHEN the user authorizes continuation of an explicit work in another native session
- THEN the session joins under its own identity, the lead accepts/transfers authority, and future content stays in one timeline with correct attribution.

#### Scenario: Unauthorized transfer
- WHEN a pending member or stale transfer attempts control
- THEN authority and content stay unchanged.


#### Scenario: Original lead unavailable
- WHEN the user selects a pending same-project request in the authenticated local viewer and confirms Reassign with a reason
- THEN that producer is accepted and becomes lead in one durable commit, with immutable audit and no inferred abandonment or changed evidence.
- AND a stale competing recovery fails, and exact replay after restart returns the original result.
