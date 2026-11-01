# walkthrough-viewer delta

## ADDED Requirements

### Requirement: Required native session registration
New timeline sessions SHALL require the resolved registered project, canonical host and actual native session ID. The server SHALL assign a distinct viewer ID and enforce exact tuple uniqueness across concurrent requests, retries and restarts. Missing identity SHALL be rejected rather than synthesized. Native IDs SHALL remain opaque and separate from viewer storage IDs. Existing registration SHALL not overwrite metadata or reopen closed/deleted work. API registration SHALL recover persisted root/project bindings before hashing identity; temporary metadata failure SHALL not replace a known project key with a path fallback, and unresolved/conflicting identity SHALL fail explicitly.

#### Scenario: Resume or concurrent creation
- WHEN two clients register the same resolved project/host/native-session tuple
- THEN both receive the same viewer ID and existing session content remains intact.

#### Scenario: Separation
- WHEN the native ID is equal but project or host differs
- THEN the server creates separate viewer sessions.

### Requirement: Versioned event and immutable artifact publication
The local server SHALL receive strictly typed, bounded milestone events and image artifacts via authenticated session-scoped API/CLI. It SHALL assign durable IDs, versions and event ordering. Operation-key retries, including successful DELETE after restart, SHALL recover the same result; tombstones SHALL retain operation receipts and replay SHALL precede stale-version/lifecycle rejection; changed content under a reused key SHALL conflict. Stale mutations and cross-session references SHALL fail without overwrite. Image bytes and origin SHALL remain immutable once published; substantive evidence changes SHALL create an iteration linked to its prior view. An immutable evidence_id SHALL identify each validation payload independently of mutable presentation versions, keeping referenced instructions/images resolvable after copy edits and restart.

#### Scenario: Lost response
- WHEN a committed image/event publication is retried after a lost response or server restart
- THEN it returns its original IDs without duplicate nodes/artifacts.

#### Scenario: Conditional text edit
- WHEN the caller edits a node with its current version
- THEN allowed copy changes commit without moving its event order; an obsolete version fails unchanged.

### Requirement: Compact session content timeline
The content pane SHALL present project, inline-editable title, date/walkthrough count, compact status/tickets/Git/archive/overflow controls, a short objective/scope summary and a newest-first milestone timeline. Per-event timestamps SHALL carry chronology. It SHALL omit redundant session cards, index, checks, environment and provenance presentation sections. Removing those sections SHALL not remove build/review requirements. Open PRs and agent-reported chat approvals SHALL not masquerade as integrated work or authenticated browser actions.

#### Scenario: Edit and navigate
- WHEN the user renames the session or opens ticket/Git lists
- THEN the header preserves context, supports keyboard/mobile use and shows real links without repeating a deliveries card.

### Requirement: Ticket-scoped evidence views
Each validation/iteration SHALL associate evidence with explicit views and related tickets. Shared views SHALL not duplicate their image collection for each ticket; distinct screens SHALL remain separately labeled within one timeline. Each view SHALL open its event/evidence-ID-bound instructions in a Dialog and its image collection in a carousel. Later iterations SHALL identify only affected views and link superseded evidence without changing images in an already-open historical dialog.

#### Scenario: Two tickets, one view
- WHEN COM-1 and COM-2 are verified by the same Chat screen
- THEN one view/image collection displays both ticket links.

#### Scenario: Focused later iteration
- WHEN a validation contains Profile and Chat but only Chat changes
- THEN the new iteration contains Chat evidence and links its predecessor while Profile retains its earlier evidence.

### Requirement: Proportional identity-aware publication procedure
Shared walkthrough instructions SHALL name host capabilities rather than concrete host tool identifiers. Host-specific identity acquisition SHALL have one conditional reference verified against current official documentation and installed versions; unsupported discovery SHALL not be represented as a native ID. The canonical walkthrough procedure SHALL obtain host/native session identity, register or resume the server session, preserve returned IDs in the work checkpoint, upload immutable new captures and publish typed meaningful milestones. It SHALL verify committed readback before claiming a visible update. Initial affected-layout verification SHALL cover relevant desktop/mobile widths and supported themes; correction rounds SHALL check affected regions/neighbors, relevant states/widths and theme-sensitive behavior only. Navigation changes SHALL exercise their affected sequence. Publication consumers SHALL follow flow-walkthrough; review/capture consumers SHALL follow the single verify-owned conditional UI policy without a conflicting full-matrix iteration instruction. Human responses SHALL continue through native host questions in this delivery.

## MODIFIED Requirements

### Requirement: Walkthrough ingestion and grouping
For legacy inputs, the server SHALL copy valid `walkthrough.json` files found under `_support/workspace/*/` of registered roots, with their referenced files, into its own store. Physical copies SHALL retain their folder/session-derived keys and per-copy resources. The viewer SHALL group normal visible copies by exact project key, declared host and `session.id`, across tickets, and SHALL present each copy as a chronological segment of that session. Different tuple values SHALL remain separate; titles and tickets SHALL NOT establish identity. It SHALL NOT ingest a manifest confirmed as tracked by a complete check of its owning Git index. Tracking SHALL be checked before unchanged-source shortcuts, including ancestor repositories and worktrees. A confirmed non-Git workspace SHALL remain supported. An unavailable or incomplete tracking check SHALL produce unknown status and a warning, never confirmed untracked status; ingestion and access in that unknown state SHALL NOT claim tracking exclusion or verified authorship. The project key SHALL be the normalized Git remote, or the repository root without one; the project label SHALL come from the root's `## Hive` `Project:` value, then `.engram/config.json` `project_name`, then the remote's last segment or the folder name.

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

New API-owned sessions SHALL use the existing project/host/native-session identity principle and SHALL have exclusive server-side content authority. Scanner passes SHALL not overwrite their events or expose duplicate legacy content of the same tuple. Legacy manifests and existing standalone HTML fallback SHALL remain readable without invented event reconstruction. An installed CLI with an unavailable server SHALL report failed publication and retain local evidence without automatically starting a daemon or creating an alternate authority. API records SHALL use explicitly versioned storage and preserve readable committed state after restart or failed writes. API commits SHALL durably finalize artifacts before records/receipts reference them, synchronize files and parent directories before success, and report/reconcile outcome_unknown after an uncertain post-rename failure rather than claim a rollback. Content limits SHALL reserve lifecycle capacity and SHALL not prevent terminal deletion.

#### Scenario: Legacy producer and new API session
- WHEN a source manifest changes for a tuple already owned by an API session
- THEN it cannot replace the API session's content or produce a duplicate visible session.

#### Scenario: Server unavailable
- WHEN a CLI attempts API publication while the viewer is stopped
- THEN it fails actionably, preserves captures and neither starts the viewer nor claims publication success.

### Requirement: Store retention and deletion
Neither legacy nor API content SHALL expire or be excluded from ingestion due to age; pinning and archiving SHALL change preferences or visibility only. Explicit deletion through a session member's authenticated same-origin management route SHALL delete all currently stored copies with the same exact project key, declared host and session ID, including archived and source-excluded members. Deleting SHALL never modify repository files. Per-copy tombstones SHALL be persisted before removing copy directories and SHALL prevent unchanged deleted content from returning; new or modified source content MAY be ingested again. Every API read SHALL require the server token, and copied files SHALL be served only under a per-copy capability derived from it. Partial deletion SHALL be reported as a failure with remaining-copy counts and a retry path, never as complete success. Restart recovery SHALL reconstruct actual surviving copies and complete pending logical-record cleanup when no member remains.

#### Scenario: Read without token
- **WHEN** a request for the index, a session or a copied file lacks the token or matching capability
- **THEN** the server answers 403 and returns no data.

#### Scenario: Delete without token or same origin
- **WHEN** deletion lacks the server token or carries a foreign Origin or Host
- **THEN** the server answers 403 and all copies remain.

#### Scenario: Delete a logical session
- **WHEN** the user confirms deletion of a session containing several tickets and archived members
- **THEN** every current member is deleted from the viewer, its source is untouched, and new or modified source content can appear later.

#### Scenario: Preserve a session independently of pinning
- **WHEN** a session receives a new walkthrough and its members age while pinned, unpinned or archived
- **THEN** all its members remain stored until explicit deletion.

#### Scenario: Partial storage failure during deletion
- **WHEN** a member cannot be deleted after other members were removed
- **THEN** the operation reports failure, publishes the surviving members and allows retry without resurrecting successfully deleted unchanged content.

#### Scenario: Pending record cleanup without restart
- **WHEN** the last copy was removed but logical-record cleanup failed
- **THEN** the periodic scan retries cleanup and any new ingestion for that identity waits until cleanup succeeds, without inheriting stale name or preference values; the UI distinguishes pending cleanup from a completed delete and lets the user check recovery without presenting a read as a repair.

This scenario was approved under D10-B (final cleanup recovery and implementation).

Existing local-server authentication, capability serving, pin/archive preferences and source-preserving explicit deletion SHALL extend to API sessions. API identity deletion SHALL retain a tombstone against implicit recreation. Legacy historical cleanup SHALL require the exact approved member-key/hash set inside the deletion lock; a mismatch SHALL preserve the whole group. It SHALL not expand the fixed inventory or remove source files, new sessions, registry, authentication or service settings. Partial deletion SHALL report survivors and preserve recovery state.

#### Scenario: Historical group changes before cleanup
- WHEN a member is added or changed after the authorized inventory but before deletion
- THEN the conditional deletion rejects that whole group without tombstones or removed copies.

### Requirement: Declared flow stages and immutable documents
The producer SHALL create/resume a session at substantive flow entry before captures, retain one stage occurrence through conditional progress/completion transitions without reordering, and publish a linked new occurrence for later re-entry. The server SHALL atomically maintain at most one foreground in-progress stage and distinguish human waiting from process liveness. Immutable document revisions SHALL belong to the same session, remain resolvable after restart, and render as read-only Markdown in native tabbed Dialogs. Approval targets SHALL distinguish document revisions from visual evidence. Open snapshots SHALL not change on refresh.

#### Scenario: Revise an existing plan
- WHEN the person expands a completed plan
- THEN the same session receives a linked new plan occurrence, separate immutable revision and truthful declared stage state, without replacing the earlier approved plan.

### Requirement: Manual viewer retention
Neither legacy nor API content SHALL be purged or excluded from ingestion due to age. Archive SHALL change visibility only; pin SHALL not be necessary for preservation. Previously accepted TTL flags SHALL not silently activate expiry. Explicit deletion and bounded incomplete-staging recovery SHALL remain separate from age retention.

#### Scenario: Inactive session survives
- WHEN an unpinned archived session ages beyond fourteen days and the server restarts
- THEN its documents and images remain readable until explicit authorized deletion.

### Requirement: Self-describing protocol and confirmed-input hygiene
The CLI SHALL expose offline versioned usage/schemas/limits/examples derived from authoritative contracts. One publication skill SHALL own the protocol, with short flow-specific pointers and no duplicated review policy. Hygiene SHALL delete only owned disposable source inputs after exact committed session/revision/artifact readback, preserving failed/uncertain uploads, paused work, selected evidence and authoritative source plans.

#### Scenario: Lost publication response
- WHEN the client cannot establish durable storage of every referenced document/image
- THEN source inputs remain and the handoff does not claim publication or successful cleanup.

### Requirement: Viewer command group help
The manager SHALL display the viewer management and publishing subcommands and exit successfully for `hive view --help`, `hive view -h` and `hive view help`, without resolving or writing server state or launching processes.

#### Scenario: Help without a running server
- **WHEN** the user requests viewer group help without a server or state directory
- **THEN** the command lists start, stop, status, open, register, serve, contract, session, event, stage, artifact and document and exits successfully without creating state.

### Requirement: Viewer-first walkthrough publication
For legacy manifest publication, the walkthrough publisher SHALL determine whether the `hive` executable is available on PATH. When available, it SHALL validate and publish only the JSON manifest and its explicit lifecycle update, SHALL NOT build or write an HTML page, and SHALL leave any existing HTML unchanged. Producers SHALL register the artifact for the viewer and identify the JSON/visual evidence in the handoff. When the executable is absent, the publisher SHALL generate the standalone HTML fallback, preserving the existing preparation and recovery behavior. Server availability, registration errors and failed confirmation SHALL NOT enable HTML generation when the executable is present. Closure SHALL still confirm the current JSON/resources or retain the unit when confirmation fails. The artifact unit SHALL include HTML only when present; selecting the JSON manifest or an existing HTML page for retention SHALL retain its referenced captures and recordings. Existing HTML SHALL NOT be deleted automatically.

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

### Requirement: Local session name and preferences
For legacy sessions, the viewer SHALL allow the user to edit one local name per exact project/host/session identity, using a prefilled single-field dialog without asking again for known ticket, project or date values. The name SHALL be trimmed, contain at most 200 Unicode code points, reject control characters and be rendered as text. An empty name or reset action SHALL remove the override and recover the existing derived title. The name SHALL appear in the session toolbar and sidebar, participate in search alongside individual walkthrough titles and tickets, and survive reload, restart and index reconstruction. Naming SHALL NOT modify identity, source manifests or individual walkthrough titles. The server SHALL persist name, archive and pin preferences in a versioned logical-session record independent of the reconstructible index. Corrupt preferences SHALL produce an explicit error rather than silently allowing destructive retention under default flags.

#### Scenario: Save a session name
- **WHEN** the user saves a valid name and reloads or the viewer rebuilds its index
- **THEN** the same session retains that name and all member source titles remain unchanged.

#### Scenario: Cancel or failed save
- **WHEN** the user cancels editing or saving fails
- **THEN** cancellation writes nothing, and a failed save preserves the entered name with an inline retryable error.

#### Scenario: Reset the local name
- **WHEN** the user resets the session name
- **THEN** the stored override is removed and the derived title appears without renaming any walkthrough.

#### Scenario: Find the edited name
- **WHEN** the user renames a session, reloads and searches for the saved name
- **THEN** the complete session is found with all its accessible siblings, while individual ticket and title searches still work.
