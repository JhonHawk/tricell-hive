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
New API-owned sessions SHALL use the existing project/host/native-session identity principle and SHALL have exclusive server-side content authority. Scanner passes SHALL not overwrite their events or expose duplicate legacy content of the same tuple. Legacy manifests and existing standalone HTML fallback SHALL remain readable without invented event reconstruction. An installed CLI with an unavailable server SHALL report failed publication and retain local evidence without automatically starting a daemon or creating an alternate authority. API records SHALL use explicitly versioned storage and preserve readable committed state after restart or failed writes. API commits SHALL durably finalize artifacts before records/receipts reference them, synchronize files and parent directories before success, and report/reconcile outcome_unknown after an uncertain post-rename failure rather than claim a rollback. Content limits SHALL reserve lifecycle capacity and SHALL not prevent terminal deletion.

#### Scenario: Legacy producer and new API session
- WHEN a source manifest changes for a tuple already owned by an API session
- THEN it cannot replace the API session's content or produce a duplicate visible session.

#### Scenario: Server unavailable
- WHEN a CLI attempts API publication while the viewer is stopped
- THEN it fails actionably, preserves captures and neither starts the viewer nor claims publication success.

### Requirement: Store retention and deletion
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
