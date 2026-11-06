# walkthrough-viewer

[canonical-references]: canonical-references.md

## Purpose

A user-controlled local server for shared work, native producers, declared stages, canonical tasks and references, immutable documents and view-scoped evidence. Publication uses only the protocol v3 API; there is no standalone HTML fallback.

## Requirements

### Requirement: User-controlled walkthrough server

The manager SHALL provide `hive view start`, `stop`, `status`, `open`, `serve` and `register`. The server SHALL bind only to 127.0.0.1, run until the user stops it, and never be started by register, status or open. The manager SHALL retain existing Git-root and `_support/workspace` registration resolution and SHALL additionally accept the exact non-Git input directory as a root when it contains valid explicit project configuration. It SHALL NOT infer parent membership or enroll children. Repeated registration of that root SHALL be idempotent and preserve added_at. Invalid explicit configuration SHALL provide an actionable failure for this new registration path without writing. Existing registrations SHALL not be removed when configuration later becomes invalid or absent.

#### Scenario: Second start
- WHEN a live server is recorded and hive view start runs
- THEN it reports the running server without creating another process.

#### Scenario: Legacy workspace entry
- WHEN register names a folder inside <root>/_support/workspace and root is not Git
- THEN it records root and retains the existing project name fallback; explicit presentation grouping applies only when configured.

#### Scenario: Idempotent registration
- WHEN an already registered repository is registered again
- THEN the command exits successfully without rewriting the registry.

#### Scenario: Register an explicitly configured non-Git workspace directly
- WHEN register names its exact root with valid configuration
- THEN that root is registered and reported.
- WHEN the same root is registered again
- THEN no duplicate entry or added_at change occurs.

### Requirement: Walkthrough ingestion and grouping
The protocol v3 runtime SHALL serve API-owned works from a v3 state directory exclusively. It SHALL NOT scan or import manifests, API v1 or provisional v2 sessions or copy preferences. Any name outside the v3 inventory SHALL make startup fail as unknown state. Project roots SHALL be explicitly registered in v3. Unknown or corrupt v3 state SHALL fail without cleanup. API commits SHALL finalize blobs and synchronize files/directories before success and SHALL reconcile outcome_unknown after uncertain durable commits.

#### Scenario: Clean start
- WHEN v3 starts on a new state directory and roots are registered
- THEN its index is initially empty.

#### Scenario: Earlier state selected
- WHEN v3 is pointed at a state directory that holds an earlier viewer's files or lacks the v3 storage marker
- THEN startup fails with invalid_storage before creating or altering files, preserving that state.

### Requirement: Store retention and deletion
All v3 work content SHALL use manual retention without age-based expiry. Archive/pin SHALL affect visibility/preferences only. Authenticated same-origin, session-scoped, version-conditional deletion SHALL remove the selected work and its owned stored blobs without editing project files or other works. A durable tombstone and operation receipt SHALL precede cleanup; partial failure SHALL report remaining work and allow scoped retry without resurrection. Reads SHALL require the existing token/capability protections. Restart recovery SHALL complete pending deletion without treating a failed write as success.

#### Scenario: Delete and retry
- WHEN deletion commits but its response is lost
- THEN exact retry after restart returns the receipt and no deleted work reappears.

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
The manager SHALL display the viewer management and publishing subcommands and exit successfully for `hive view --help`, `hive view -h` and `hive view help`, without resolving or writing server state or launching processes.

#### Scenario: Help without a running server
- **WHEN** the user requests viewer group help without a server or state directory
- **THEN** the command lists start, stop, status, open, register, serve, contract, session, event, stage, artifact and document and exits successfully without creating state.

### Requirement: Safe and deterministic Git tracking lookup
Any retained Git source/identity lookup SHALL disable repository-configured fsmonitor callbacks for that process without changing configuration, neutralize inherited pathspec interpretation, preserve bounded-output handling and avoid conflating distinct filesystem objects. Removing the legacy scanner SHALL NOT weaken these controls in remaining lookups.

#### Scenario: Retained Git lookup
- WHEN a v3 project/source operation needs a Git lookup
- THEN repository callbacks do not execute and inherited pathspec variables do not change the selected path.

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

### Requirement: Hierarchical React viewer navigation

The shell SHALL preserve the existing project/workspace groups, source metadata, title/ticket search, counts, selection, expansion, keyboard/mobile navigation and full accessible labels. Works within each group SHALL be ordered by last substantive native work activity descending with stable ID ties, and show that activity's local date/time. Creation SHALL supply the initial activity timestamp. Administrative rename/pin/archive/restore, reads, unchanged replay and automatic maintenance SHALL NOT advance activity. Public events and private native operation receipts SHALL support truthful reconstruction of older activity without using source mtimes or localized labels as sort keys.

Every row SHALL show its actual active, in_review, awaiting_validation, paused, closed or failed state visibly and accessibly through the installed semantic Badge variants, not color alone. Existing internal timeline ordering SHALL be preserved. Loading, empty, error and authentication states SHALL remain actionable. A selected work archived by maintenance SHALL remain readable in its current view until navigation, with an archived explanation and Restore action, without selecting an arbitrary other work or a persistent application-wide banner.

#### Scenario: Work resumes in an old session
- WHEN an older work receives a new substantive publication
- THEN it precedes less recently active works without moving its internal timeline groups.

#### Scenario: Administrative changes
- WHEN a work is renamed, fixed, archived, restored or automatically maintained
- THEN its last-work-activity timestamp and ordering remain unchanged.

### Requirement: Reversible copy archiving

Archival SHALL affect visibility only, preserving work content, source identity, lifecycle and pin. Manual archive/restore SHALL remain scoped, authenticated, atomic, version-conditional and persistent. Autoarchive SHALL apply only to explicitly closed, unpinned works at least two hours after their current closure episode and SHALL run at startup and within a 60-second runtime sweep. It SHALL NOT infer completion from inactivity or archive active, paused, review, validation or failed works. Background archival SHALL NOT produce a toast per work; existing temporary Undo for manual archival SHALL remain work-specific.

A manual restore SHALL durably exempt that closure episode. Pin SHALL prevent autoarchive; unpin SHALL reevaluate the original deadline without clearing a restore exemption. Explicit reopen SHALL clear the prior episode/exemption and atomically restore automatically archived work, while preserving a manual archive. A subsequent explicit close SHALL start a new episode and deadline.

#### Scenario: Deadline and restore
- WHEN a closed eligible work reaches 7,200 seconds and a sweep succeeds
- THEN it becomes archived without losing content; a subsequent manual restore keeps it visible through later sweeps and restarts until a new closure.

#### Scenario: Concurrent user action
- WHEN a sweep races a versioned restore, pin or reopen
- THEN storage serializes them, retains the winning current intent and returns a safe conflict for stale requests rather than overwriting the newer state.

### Requirement: Explicit walkthrough delivery links
v3 API delivery events SHALL accept only verified PR/commit links, distinguish candidate metadata from delivery and never infer integration. URLs SHALL be absolute HTTPS with valid ASCII hosts and no credentials, spaces or controls; labels SHALL render as text, destinations SHALL be validated on rendering, and new contexts SHALL use noopener noreferrer. The v3 compact header SHALL expose deduplicated canonical work delivery entities by ID in its accessible delivery popover; no links SHALL be fabricated from candidate text.

#### Scenario: Unsafe URL
- WHEN a delivery contains an executable, relative or credential-bearing URL
- THEN validation rejects it and no unsafe browser link is created.

### Requirement: Viewer-first walkthrough publication
Substantive publication SHALL use only the compatible v3 API and confirmed readback. A missing executable or an incompatible viewer SHALL preserve inputs and report publication unavailable without starting a daemon or producing any other format; a stopped viewer SHALL do the same, except that semantic open, stage and evidence declarations SHALL be queued as «Queued semantic publication while the viewer is unavailable» sets. Previously generated HTML pages and user-selected artifact units SHALL not be deleted or rewritten by the removal of the renderer. Research SHALL NOT publish to the viewer; a work's viewer session SHALL start at its first plan or build flow that was explicitly invoked (its command typed, its route accepted, or research continuing into build at the user's request), created with a summary of about one paragraph on what the work is about. The agent SHALL NOT open the viewer on its own: at each stop where the viewer holds something for the user to review (a plan at its approval stop, a UI walkthrough, including the build's first rendered view) it SHALL offer to open it inside what that stop already asks, until the user accepts once in the conversation. When the stop asks through the host's question tool, the offer SHALL be one more option of the stop's own approval, acceptance or validation question, never a separate question, a question with its own decision ID or its own question-tool call; choosing it SHALL open the work once and ask that same question again, unchanged, with the same ID and without the open option, keeping the answers to the round's other questions. The open option SHALL NOT be the first option and SHALL NOT be starred. Otherwise, or when the question already holds the most options the tool accepts, the offer SHALL be a reply line of the stop's waiting line or next-step block, and using it SHALL open the work while the stop keeps waiting for its answer. After acceptance the agent SHALL NOT offer or open it again unless the user asks.

#### Scenario: Binary present, server incompatible
- **WHEN** publication finds a v1 server
- **THEN** it requests the coordinated v3 update and preserves inputs without HTML substitution

#### Scenario: User opens the viewer before deciding
- **WHEN** a validation stop asks its approval question through the host's question tool and the user picks the open-viewer option
- **THEN** the work opens in the viewer once
- **AND** the same approval question is asked again with the same ID and without the open-viewer option, and the round's other answers stand

#### Scenario: User declines
- **WHEN** the user answers a review stop's question without choosing the open-viewer option
- **THEN** the viewer stays closed and the offer is repeated at the next review stop

#### Scenario: Research report
- **WHEN** a research report ends, with or without a route to plan or build
- **THEN** nothing is published to the viewer and no open-viewer offer is made
- **AND** the first plan or build flow of that work creates its viewer session with a one-paragraph summary

### Requirement: Fixed session toolbar
The selected v3 work title and management actions SHALL stay visible outside its scrolling timeline/document canvas. The header SHALL show the title with its edit action on its first row, and the work summary, clamped to two lines with an expand control, beside a right-aligned toolbar on its second row; it SHALL NOT show the project name or the creation date. The toolbar SHALL hold the state, the validation count and the ticket, delivery and participant counters, and the archive and more actions. Deep-linked stage/event headings SHALL remain visible below the toolbar. Keyboard actions, unbroken long names through bounded local scrolling, narrow layouts without page overflow (the toolbar wrapping below the summary) and print omission of management controls SHALL remain supported.

#### Scenario: Narrow long title
- WHEN a valid unbroken title exceeds a narrow viewport
- THEN local keyboard-accessible title scrolling exposes it without page-level overflow or hiding actions.

#### Scenario: Long summary
- WHEN the work summary exceeds two lines
- THEN the header shows two lines with an expand control, and the canvas does not repeat the summary.

### Requirement: Local session name and preferences
v3 title/archive/pin management SHALL use work-local typed API state and local_viewer provenance, not imported logical-copy preferences. Names SHALL be trimmed, nonempty, bounded to 200 Unicode code points, reject controls and render as text. A canceled edit SHALL write nothing; a failed save SHALL retain input and a retryable error. Renaming SHALL preserve source/producers, stage titles and content, remain searchable and survive restart. Corrupt persisted state SHALL fail explicitly rather than enable defaults that delete content.

#### Scenario: Rename and restart
- WHEN the user saves a title and restarts v3
- THEN the title is searchable and source, producers and stages remain unchanged.

### Requirement: Explicit session mutation scope
Browser and CLI title/archive/pin/deletion and recovery SHALL require existing authentication, explicit session action scope and the v3 local_viewer actor, and browser requests SHALL also pass same-origin/Host checks. Obsolete clients SHALL receive a safe reload/update rejection before mutation; the UI SHALL not replay a rejected write automatically. Native publisher operations SHALL follow the separately typed producer authority rules, not claim a local_viewer actor.

#### Scenario: Old tab
- WHEN an old browser tab sends an obsolete management write
- THEN v3 rejects it without changing work and offers reload instead of automatic replay.

### Requirement: Required native session registration
New timeline sessions SHALL require the resolved registered project, canonical host and actual native session ID, and MAY carry a work key. An identity resolved by the manager from a verified host source, as «Resolved native identity and current work» sets, is the actual native session ID, not a synthesized one. The server SHALL derive a distinct viewer ID from the project, host, native session ID and the effective work key: the given work key, or, when absent, the creating operation ID. It SHALL enforce exact uniqueness across concurrent requests, retries and restarts. Sessions stored before work keys existed SHALL keep their IDs. Missing identity SHALL be rejected rather than synthesized. Native IDs SHALL remain opaque and separate from viewer storage IDs. Existing registration SHALL not overwrite metadata or reopen closed/deleted work. API registration SHALL recover persisted root/project bindings before hashing identity; temporary metadata failure SHALL not replace a known project key with a path fallback, and unresolved/conflicting identity SHALL fail explicitly.

#### Scenario: Resume or concurrent creation
- WHEN two clients from the same native session register the same project/host/native-session tuple with the same work key
- THEN both receive the same viewer ID and existing session content remains intact.

#### Scenario: Two works in one native session
- WHEN one native session registers two different work keys, or two creations without a work key with different operation IDs
- THEN the server creates two viewer sessions.

#### Scenario: Separation
- WHEN the native ID is equal but project or host differs
- THEN the server creates separate viewer sessions.

### Requirement: Versioned event and immutable artifact publication
The local server SHALL accept typed, bounded protocol v3 events, images and Markdown documents under work-scoped authentication, assign durable IDs/versions/sequences and preserve immutable bytes. Native publication mutations SHALL carry an accepted producer and operation ID; host-observed events SHALL carry the accepted producer of their reporting session and an operation ID derived from the host event, as «Host-observed events» sets. Local viewer management mutations SHALL use the separate local_viewer actor and SHALL NOT impersonate a native producer. Exact replay SHALL return its durable receipt after restart; changed operation payloads, stale mutations and cross-work references SHALL fail without overwrite.

Server-side durable receipts SHALL preserve successful DELETE replay after restart, with replay preceding stale-version or lifecycle rejection for v3 operations. Image bytes and origin SHALL remain immutable; substantive evidence changes SHALL create an iteration linked to the prior view. Each validation payload SHALL retain its immutable evidence_id independently of copy edits, preserving instructions/images after restart.

Stage-associated events SHALL receive server-owned stage_event_id from a validated target_stage_event_id or the lead's foreground/latest preceding non-withdrawn stage. Non-lead publishers SHALL provide explicit stage targets. Clients SHALL NOT set the output field. Stages, closure and global membership/handoff events SHALL carry none. Protocol v1 events SHALL NOT be imported or reconstructed.

#### Scenario: Concurrent producers
- WHEN two publishers use the same expected version
- THEN one commits and one conflicts without overwriting; exact retry preserves receipt identity.

#### Scenario: Foreign stage
- WHEN a publisher targets a foreign or withdrawn stage
- THEN publication fails without content or state mutation.

### Requirement: Compact session content timeline
The content pane SHALL preserve project, inline-editable title, date, compact status/archive/overflow controls and honest agent-reported chat approvals without presenting them as browser-authenticated approval or integrated work. It SHALL omit redundant session cards, index, checks, environment and provenance presentation sections while retaining required build/review verification.

The content pane SHALL show compact icon-and-count ticket/delivery controls with full accessible labels and hover help, hide zero validation count, clamp expandable summaries and render Markdown lists. It SHALL show pre-stage events first, stages by ascending start sequence, only explicitly assigned task-segment projections in declared plan order and ascending activity within each task or the stage general-activity section, and closure last. Content associated with a task SHALL NOT be duplicated as a sibling stage event. Current, paused, review and validation stages and stages with in-progress/review/blocked tasks SHALL stay expanded; finished/failed stages SHALL initially collapse but remain expandable. Withdrawn stages SHALL retain their content under a withdrawn label.

Resumed old occurrences SHALL retain their original group position; timestamps SHALL not imply global chronological ordering across nested groups. Captions SHALL show non-empty distinctive text with full accessible names. A compact read-only session popover SHALL identify contributors and lead. New content SHALL preserve the viewport; Go to current SHALL navigate to the foreground or latest visible event. Deep links and per-work scroll SHALL preserve navigation context.

#### Scenario: Read work order
- WHEN research, plan, build and verification publish content
- THEN reading downward follows occurrence starts and their events, with closure at the bottom.

#### Scenario: Live arrival
- WHEN new content arrives while the user reads earlier content
- THEN the viewport stays put and Go to current remains available.

### Requirement: Ticket-scoped evidence views
Each validation/iteration SHALL associate evidence with explicit views and related tickets. Shared views SHALL not duplicate their image collection for each ticket; distinct screens SHALL remain separately labeled within one timeline. Each view SHALL open its event/evidence-ID-bound instructions in a Dialog and its image collection in a carousel. Later iterations SHALL identify only affected views and link superseded evidence without changing images in an already-open historical dialog.

#### Scenario: Two tickets, one view
- WHEN COM-1 and COM-2 are verified by the same Chat screen
- THEN one view/image collection displays both ticket links.

#### Scenario: Focused later iteration
- WHEN a validation contains Profile and Chat but only Chat changes
- THEN the new iteration contains Chat evidence and links its predecessor while Profile retains its earlier evidence.

### Requirement: Proportional identity-aware publication procedure
Shared walkthrough instructions SHALL name host capabilities rather than concrete host tool identifiers. Identity acquisition SHALL belong to the manager, which resolves it from verified host sources; on the semantic path the procedure SHALL NOT ask the agent to discover, guess or pass its own native ID, except as each host's one conditional identity reference, verified against current official documentation and installed versions, sets when the manager reports identity unresolved; unsupported discovery SHALL not be represented as a native ID. The canonical walkthrough procedure SHALL open or resume the work, keep its work key in the work checkpoint (the manager keeps the returned IDs), publish each stage change with one declaration, upload immutable new captures with one evidence command and publish typed meaningful milestones; tasks the host already reports SHALL NOT be registered again by hand. A mutation's returned receipt (IDs and versions) SHALL confirm that write; the procedure SHALL read the committed session back before claiming to the user that a publication is visible, at close, and when resuming after an interruption or a version conflict, and SHALL NOT require a session read after every mutation; a close declaration that reads the work back satisfies the read at close. The procedure SHALL be read once per session and again after a compaction, not at each stage change. Initial affected-layout verification SHALL cover relevant desktop/mobile widths and supported themes; correction rounds SHALL check affected regions/neighbors, relevant states/widths and theme-sensitive behavior only. Navigation changes SHALL exercise their affected sequence. Publication consumers SHALL follow flow-walkthrough; review/capture consumers SHALL follow the single verify-owned conditional UI policy without a conflicting full-matrix iteration instruction. Human responses SHALL continue through native host questions in this delivery.

#### Scenario: Publication without per-write reads
- WHEN a producer starts a stage and later updates it in the same session
- THEN it uses the receipt returned by the first declaration for the update, without reading the session in between
- AND it reads the session back before telling the user the stage is visible.

### Requirement: Declared flow stages and immutable documents
Stages SHALL include research, plan, build, verify and delivery. Stage states SHALL include in_progress, in_review, awaiting_validation, paused, completed and failed. Work states SHALL distinguish active, in_review, awaiting_validation, paused, closed and failed. Labels SHALL be short, reasons separate. One accepted lead SHALL control at most one foreground; global/foreground review or validation changes SHALL be atomic. Parallel review SHALL NOT change that foreground. completed/failed occurrences SHALL be terminal; re-entry SHALL create a linked occurrence.

Distinct transitions SHALL append immutable server-timestamped audit events with declared producer, reason and previous/next states. Membership/handoff and all other new mutations SHALL retain actor provenance and audit; local browser renaming, archiving, pinning and confirmed deletion SHALL use local_viewer actor provenance with existing scope/origin/version checks; audit events SHALL NOT be editable or withdrawable. User questions SHALL publish awaiting_validation first. Responses SHALL reference concrete targets when known, never be inferred from silence, and resumption SHALL be explicit. Closure SHALL require no foreground.

Documents from all five stage types SHALL use one read-only Markdown Sheet, full width mobile/two-thirds desktop, fixed scrollable tabs and one close action/publication indicator. Only structured plan deltas SHALL group under Specs. Work-scoped document_key SHALL separate independent revision chains. Successors SHALL identify their predecessor and cross-occurrence continuation SHALL require linked same-type stages. Open content SHALL not change until user action. Only a later revision in that chain SHALL trigger an update notice; switching SHALL preserve the selected label or announce its removal. Comparisons SHALL identify added/changed/removed tabs by label/document identity; the first revision SHALL be labeled initial.

#### Scenario: Wait for a plan choice
- WHEN the lead presents a plan for user input
- THEN work/foreground show Validation with a separate reason and immutable transition before asking.
- AND a complete published document does not imply approval or completed work.

#### Scenario: Parallel review
- WHEN a collaborator posts review progress or findings during build
- THEN its declared producer and stage target are visible while build stays In progress.

#### Scenario: Independent documents
- WHEN separate document keys use identical tab labels
- THEN comparisons and update notices stay separate, including across linked same-type occurrences.

### Requirement: Manual viewer retention

API content SHALL NOT be purged due to age. Automatic two-hour archival SHALL change visibility only, and pin SHALL remain unnecessary for content preservation. Explicit deletion SHALL remain the only removal of completed work content. No configurable age/TTL deletion option SHALL be introduced.

#### Scenario: Archived content remains available
- WHEN an automatically archived, unpinned work ages beyond fourteen days and the server restarts
- THEN its documents and images remain readable until an explicit authorized deletion.

### Requirement: Self-describing protocol and confirmed-input hygiene
Offline contract usage/schemas/limits SHALL describe protocol/storage v3 exclusively. New CLI and viewer SHALL require compatible v3 peers for publication. No v1/v2 reader, endpoint, migration or imported evidence reconstruction SHALL be part of the v3 runtime. v3 receipts, immutable blobs and atomic state transitions SHALL remain durable after restart or uncertain commit.

The offline contract SHALL be readable by part: the full contract; one operation in full; a brief per flow (research, plan, build, verify, delivery, close, collab, local) with each operation's usage, required and optional input fields, enum values and example, without output schemas; a list of every operation with its flows and usage; and the shared rules block. Every operation SHALL belong to at least one flow. Contract and command output SHALL be compact JSON with the same content as the indented form. A rejected input SHALL name its field and code.

One canonical publication procedure SHALL own collaboration/native identity, meaningful snapshots, states before user stops, explicit resumption and readable summaries. Consumers SHALL link to it. Children SHALL return evidence to the main producer. Viewer identity SHALL NOT change Engram runtime identity. Hygiene SHALL retain authoritative plans, current review inputs, uncertain uploads and unique evidence.

#### Scenario: v3 restart
- WHEN a v3 publication commits and the service restarts
- THEN IDs, hashes, actor audit and operation receipts remain resolvable without a legacy adapter.

#### Scenario: Old writer
- WHEN a v1 or provisional v2 publisher sends a request to a v1 endpoint of v3
- THEN the endpoint does not exist and nothing is mutated.

#### Scenario: Brief contract for one flow
- WHEN a producer runs `hive view contract --for build`
- THEN it receives only the build operations, each with usage, required and optional fields, enum values and an example, plus the common session, stage and ticket operations, in under 10,000 bytes of compact JSON.

#### Scenario: Operation without a flow
- WHEN a new operation is added without a flow
- THEN the contract test fails.

#### Scenario: Rejected input
- WHEN a stage start omits its reason
- THEN the CLI reports `invalid timeline input: invalid_text on reason` and nothing is sent.

### Requirement: Local viewer management actor

The actor union SHALL keep accepted native_producer separate from local_viewer. Rename, archive/restore, pin/unpin and confirmed deletion, from the browser or the CLI, SHALL remain authenticated, scoped and available without native identity, and SHALL retain actual local_viewer provenance rather than native publication or human-approval claims. The local_viewer actor SHALL accept an optional declared `origin` of `browser` or `cli`, default `browser`; it is declared provenance, not authentication, and SHALL be kept in each action's public event, administrative receipt or private deletion audit. Revision-2 archive/restore/pin-only operations MAY use bounded private administrative state/receipt slots instead of adding public local_action events or publication receipts, including on saturated legacy records. Exact retained-slot retry SHALL replay; older stale requests SHALL conflict without mutation. Existing public history and native publication receipts SHALL NOT be pruned. local_viewer SHALL NOT publish content/evidence/approval or change stage lifecycle; explicit recover_lead and confirmed deletion SHALL retain their existing authority and durable tombstone/receipt guarantees. Mixed actor branches SHALL fail.

#### Scenario: Restore at public capacity
- WHEN autoarchive hides a valid record whose public events or publication receipts are full
- THEN an authenticated versioned Restore commits its exemption with actual local_viewer provenance, exact retry replays and restart preserves visibility, without deleting history.

#### Scenario: Confirmed browser deletion
- WHEN the user confirms deletion through the existing scoped browser flow
- THEN deletion retains its durable operation result without native impersonation or content resurrection.

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

#### Scenario: Lead recovered from the CLI
- WHEN the user asks and `hive view work recover --work-id W --request-id R --reason T` runs for a pending same-project request
- THEN the result equals the browser recovery, with the audit recording `local_viewer`, `origin: cli` and reason T.

### Requirement: Independent task progress and task detail
Tasks SHALL belong to the work with stable task_id, unique task_key, short title, plan reference when available and explicit required stage segments. The lead SHALL declare the full plan including future tasks without needing fake stage occurrences. Each segment SHALL have planned_stage, local state, accepted assignee, order and an explicit binding to a real matching stage occurrence when available. Stage panes SHALL display only their bound segments; future-stage tasks SHALL remain accessible in the plan and global count, not inside Build.

A task may span verify and delivery without becoming multiple tasks. Each stage projection SHALL show only its segment state; the task Sheet SHALL show aggregate state and the complete stage-labeled history. Global count SHALL deduplicate task_id and count completed only when all its required active segments are complete. Stage counts SHALL use only their assigned segments. No foreground change SHALL automatically move, clone or complete a task/segment. Assignment changes SHALL be explicit, versioned and audited.

Segment states SHALL be not_started, in_progress, in_review, blocked and completed. Aggregate task state SHALL be blocked if any active required segment is blocked; otherwise in_review if any is in review; completed if all are complete; not_started if all are not started; in_progress otherwise. State SHALL be displayed only in the row indicator with empty circle, loader, pause, alert or check and accessible text/tooltip. Indicators SHALL use semantic context colors alongside their distinct shapes and accessible labels; the icon stroke SHALL NOT receive a second decorative container border. Neither index, commits, checklist text nor process observation SHALL infer completion; a task state declared by the host's native task list and reported under «Host-observed events» is that session producer's declaration, not an inference. No heuristic percentage SHALL be displayed.

Each row SHALL expose available review content before expansion through compact evidence/document/wait/history affordances. Contentful rows SHALL provide independent disclosure and title actions: disclosure expands concise summary/optional relevant evidence; title opens the common-width task Sheet with Summary/Evidence/History and the selected segment in context. A task containing only a description and assignment/stage metadata, without linked evidence, documents, waits, blocking details or meaningful transition history, SHALL use inline disclosure rather than open a mostly empty Sheet. Refresh, state changes and return from Sheet SHALL preserve expansion/focus/scroll keyed by work/task/segment. Open task snapshots SHALL update only on explicit action and only a newer task_version SHALL trigger the notice. Any mutation changing visible task snapshot content, including linked evidence, document revisions and CI observations/attempt selection, SHALL atomically advance each affected task_version with the durable mutation/receipt. Unlinked uploads, unrelated tasks, no-ops and exact retries SHALL NOT advance it. Nested documents SHALL have a return path without stacked Sheets.

Blocked transitions SHALL require reason, concrete steps, owner and resume condition; incomplete payload SHALL fail unchanged. Blocking a segment SHALL NOT automatically block others or pause work. The lead or segment-assigned accepted producer SHALL perform conditional transitions and retain immutable actor audit and receipts. completed segments may reopen explicitly with a reason; terminal-stage re-entry SHALL use a linked new occurrence and explicit segment binding. Old evidence/history SHALL remain. Local viewer SHALL not approve/complete tasks. Withdrawal SHALL preserve history and exclude the withdrawn requirement from current counts.

Task content SHALL identify task_id/segment_id and its actual stage occurrence; references SHALL belong to that work, segment and origin stage. Activity from multiple stages SHALL stay attributable without duplicating task entities or blobs. Per-stage expansions SHALL show only relevant content, while the Sheet aggregates it. Decisions SHALL retain exact evidence/revision targets. Images SHALL be optional. Bounds SHALL fail atomically without preventing terminal cleanup.

#### Scenario: CAM task crossing stages
- WHEN Build tasks T1–T6 are complete and T7 has a completed verify segment and pending delivery segment
- THEN Build shows 6/6, Verify 1/1 and Delivery 0/1 when that stage exists, while the work remains 6/7 unique tasks.
- AND T7 does not appear within Build or create a fake future Delivery occurrence.

#### Scenario: Independent indicator states
- WHEN T1 and T3 segments are completed, T2 blocked, T4 in review and T5 not started
- THEN each indicator uses its explicit local state and the stage count is 2/5 regardless of index.

#### Scenario: Block and correct
- WHEN incomplete unblock instructions are submitted
- THEN transition fails unchanged.
- WHEN a completed segment reopens explicitly or a blocked one resolves
- THEN it retains IDs, evidence and immutable history across retry/restart.

### Requirement: Delivery stage and CI wait node
Delivery SHALL be a distinct declared stage after candidate verification, preserving existing foreground/lifecycle rules. It SHALL expose a typed stable CI wait node, optionally tied to a task's delivery segment. CI waiting SHALL NOT be represented as awaiting a human response or as task approval. The row SHALL use a pause indicator for waiting; its Sheet SHALL show candidate/PR, required-check authority/source/time, observed results, attempts and safe real links. The node SHALL not initiate reruns, modify provider policy or merge.

Wait observations SHALL be published by authenticated provider reads through existing authorized tools, not anonymous private-repository queries. State SHALL be waiting, succeeded, failed, cancelled or unknown; lack of access/confirmed requirements SHALL remain unknown, never success from an empty list. succeeded SHALL require observed outcomes satisfying all confirmed required checks for the actual candidate and policy. CI success SHALL not imply merge, integration or acceptance.

Wait IDs/versions and operation receipts SHALL be durable. Same candidate/check policy reruns SHALL preserve the node and append immutable attempts/observations; each required check SHALL have an explicit applicable_attempt_id selected by a versioned attempt-start/promotion operation with observed current-provider evidence. Only observations matching that applicable attempt SHALL affect current outcome. Earlier attempts may extend history but SHALL NOT overwrite current results, even using the latest wait version. Unknown attempts SHALL NOT self-promote, and uncertain applicability SHALL remain unknown. Attempt mapping/generations SHALL persist across restart; a new commit or changed required set SHALL create a linked new wait without reusing prior success. Foreign stage/task/candidate references and stale mutations SHALL fail unchanged. Reading snapshots SHALL not change them automatically. Failed CI SHALL not fabricate task-blocked instructions: a separate explicit blocked transition SHALL include its full unblock payload.

#### Scenario: Wait during delivery
- WHEN verification finishes and delivery starts awaiting PR CI
- THEN Delivery contains the wait node and its actual task segment, while Build stays complete and no user-validation state is fabricated.

#### Scenario: Wrong candidate or unconfirmed policy
- WHEN a previous commit is green or required checks cannot be confirmed
- THEN the current wait is not succeeded and no integration is claimed.

#### Scenario: Late successful prior attempt
- WHEN a rerun is waiting or failed and an older successful attempt is published later using the current expected version
- THEN the old result is retained only as history and does not restore succeeded, including after restart.

#### Scenario: Task snapshot invalidation without state change
- WHEN linked task evidence, a document revision or CI observation changes without a task state transition
- THEN its task_version advances atomically and the open Sheet offers an update while preserving its snapshot.
- AND a mutation of an unrelated task does not trigger that notice.

#### Scenario: Retry and rerun
- WHEN an exact publication retries after restart or CI reruns on the same commit/policy
- THEN receipts avoid duplicate nodes and attempt history remains readable.

### Requirement: One canonical origin for referenced entities

The final protocol/schema/storage SHALL be 3, replacing v1 and provisional v2 without a compatibility reader or migration. The detailed closed contract is [canonical references][canonical-references]. Work-owned task, ticket, delivery, candidate, producer, artifact, document and wait IDs SHALL be the origin of current projections. Stages SHALL NOT own redeclared task entities. The live task catalog SHALL be the toolbar Popover, ordered by explicit unique plan_order; stage segment order SHALL remain independent.

Native ticket/delivery writers SHALL update one work-local entity under lead authority and CAS. Candidate records SHALL be immutable and creatable by accepted native producers. Typed natural identity SHALL deduplicate repeated external-resource creation without using editable labels/URLs as identity. Projection inputs SHALL accept IDs, not redeclared entity values or caller-generated historical snapshots. Unknown, foreign, mixed-type, duplicate or retired inline references SHALL fail without writes. Entity/catalog mutations SHALL preserve audit, receipts, restart/replay, failure atomicity and all terminal capacity reserves.

A visible ticket/delivery patch SHALL update current labels/links/status across all resolving views after one read, advancing each affected task_version and wait_version exactly once where their visible snapshots change. No-op, exact retry and unrelated changes SHALL NOT advance these reader versions. Published reference context SHALL be captured by the server and remain immutable alongside documents, artifacts, audit and provider observations. Explicit current/published labels SHALL prevent attributing a current status to an earlier publication. An open reader SHALL preserve its snapshot until manual update. New candidate selection for a publication SHALL NOT retarget older evidence/waits or imply task/stage completion.

#### Scenario: One update, several projections
- WHEN a lead updates a ticket URL/key or a delivery label/status by stable ID
- THEN every current reference resolves the new values without republication of the referencing events/views.
- AND original publication snapshots and stored document/image hashes remain unchanged.

#### Scenario: Coherent seven-task plan
- WHEN the complete example registers T1–T7 from one typed definition
- THEN the Popover lists seven unique tasks, Build projects six, and Verify/Delivery share T7 without inventing tasks for Research/Plan.
- AND a task transition changes its current projections and derived counts after one canonical mutation; the published Markdown task snapshot remains historical.

### Requirement: Clear origin and compact task toolbar

The timeline SHALL start with a clear Inicio marker derived from work.created_at and its original Source, without publishing a fictitious API event or incrementing activity counts. A horizontal Tareas/Ir al actual toolbar SHALL replace the standalone global task-count column. Tareas SHALL open the installed compact Popover with one row per canonical task, state by point, truncated visual title with full accessible text and segment progress. The global total SHALL live in that Popover. The list SHALL remain viewport-bounded and keyboard-scrollable; Escape SHALL restore focus. Existing task disclosures, content affordances, navigation, polling scroll and immutable Sheets SHALL remain intact.

### Requirement: Reserved labels and browser-selected interface language
The viewer SHALL support English as default/fallback and Spanish for es/es-* browser primary language. Production and previews SHALL share that selection for system copy, accessible labels, dates and document language. Producer text and immutable content SHALL remain verbatim. The reserved-label contract is defined in [canonical references][canonical-references].

The existing stage enum SHALL select its translated system heading once; optional title context SHALL not redeclare that heading. Other nodes SHALL retain free titles. The schema/catalog SHALL expose closed reserved-key/language metadata. Empty stage context SHALL be allowed; empty non-stage titles and unknown stage values SHALL fail without mutations. Context updates/clearing SHALL obey existing CAS, authority, replay and persistence.

#### Scenario: English browser and Spanish browser
- WHEN the same work opens under en-US, es-MX or an unsupported browser locale
- THEN system labels use English, Spanish or English respectively without rewriting producer content.

#### Scenario: Custom node and standard stage
- WHEN a stage has empty or nonempty producer title context, or a custom non-stage node has a free title
- THEN the stage heading renders its localized name once with optional verbatim context, and the custom node renders its free title.

### Requirement: Clear reading bounds
The task Popover SHALL avoid residual scrolling for seven rows where space permits, while remaining viewport-bounded and keyboard-scrollable for longer lists. The timeline scroll canvas SHALL retain visible space after its true last footer without a fictitious node or event.

### Requirement: Explicit presentation grouping independent of project authority

The viewer SHALL derive a read-only navigation group from an exact registered original source root's valid `.engram/config.json` `project_name`. Equal trimmed case-sensitive explicit names SHALL share a presentation group across different canonical project keys. Missing configuration SHALL retain canonical-project grouping. Invalid or unavailable configuration SHALL retain canonical-project grouping and expose a bounded recoverable diagnostic without file content. An absent registered source root MAY use the unambiguous current explicit name of another available registered source represented in the index with the exact same persisted canonical project key. This exception SHALL preserve original origin metadata and SHALL NOT infer identity from equal display names; missing witnesses or conflicting explicit names SHALL keep the fallback and diagnostic. Configuration lookup SHALL NOT ascend, scan child repositories, invoke Engram or follow configuration symlinks. Reading and grouping SHALL NOT rewrite stored works or change identities, provenance, collaboration eligibility, memberships or receipts.

#### Scenario: Existing works in separate projects share a declared name
- WHEN two stored works have distinct project keys and their original registered roots declare the same valid name
- THEN navigation shows one named group with both works and the correct visible count
- AND their IDs, project keys, sources and join boundaries remain unchanged.

#### Scenario: Configuration changes after publication
- WHEN the explicit name changes, is removed or becomes invalid
- THEN the next completed index refresh changes only the presentation, using canonical fallback on absence or invalidity
- AND invalidity is reported without hiding the works or changing their stored bytes.

### Requirement: Compatible opt-in navigation projection

The authenticated revision-2 index SHALL expose version-1 navigation metadata only when navigation=grouped is requested. Requests without that projection SHALL retain the revision-2 canonical grouping shape. The projection SHALL reference each indexed work exactly once, preserve stable namespaced grouping keys and original-root origin metadata, and observe each source-root configuration at most once per index. The revision-2 frontend SHALL fall back to canonical groups with a diagnostic when optional projection metadata is invalid, without weakening session validation. Cross-revision client/server operation SHALL be replaced by the coordinated revision gate; schema/storage version 3 identities SHALL remain unchanged.

#### Scenario: Matching revisions with invalid optional grouping
- WHEN matching revision-2 clients and server exchange an index with invalid optional navigation metadata
- THEN the frontend renders canonical groups and a diagnostic without losing or weakening validation of works.

#### Scenario: Mixed client and server generations
- WHEN the client and server timeline revisions differ or the required revision is missing
- THEN timeline/index access requests update or reload and no timeline mutation is accepted.

### Requirement: On-demand high-level publication

The manager SHALL provide the existing additive hive view work semantic interface, reading explicitly bounded local sources, freezing selected Markdown and publishing through authenticated protocol v3 operations at matching timeline contract revision 3. It SHALL return compact confirmed results without callers authoring transport IDs, versions, hashes or request bodies. Existing low-level operation meanings, native payload identity and Markdown rendering SHALL be preserved after the coordinated update; operation against unmodified older server revisions SHALL not be promised. Publication SHALL NOT start/restart the viewer or open a browser.

#### Scenario: Publish and revise a plan
- WHEN a matching-revision caller publishes a valid plan folder and later its edited contents to the same work
- THEN two immutable revisions retain original bytes and the CLI confirms their revision IDs.

#### Scenario: Equal content and context
- WHEN current remote content and semantic publication context match the requested set
- THEN the command returns unchanged without remote mutation.

### Requirement: Explicit work and directory bindings
The manager SHALL bind canonical source directories to explicit work keys/IDs and actual native producers, without treating cwd, title or content hash as work identity. Distinct works in one native session SHALL remain separate. Listing/showing SHALL use existing read APIs and SHALL NOT create, reopen or mutate works. A changed binding SHALL require explicit bind with a reason, matching project/authority and no pending publication. Closed/deleted works SHALL not reopen through discovery, publication or technical recovery.

#### Scenario: Two changes in one conversation
- WHEN one native session opens change A and change B with distinct keys and publishes updates to each
- THEN list returns both identities, updates remain in their own histories, and changing cwd changes neither binding.

#### Scenario: Mismatched folder or producer
- WHEN a caller supplies a different directory for a bound work or lacks accepted membership
- THEN mutation is rejected without selecting another work or impersonating its lead.

### Requirement: Safe bounded document collection
The manager SHALL validate all selected files and their containment/types/limits before any publication mutation. Default plan collection SHALL include required proposal/tasks, optional design and deterministically ordered spec deltas. Explicit Markdown selection SHALL support retained documents outside that default shape without implicit linked-file traversal. It SHALL reject escapes, symlink components beneath the canonical root, shared inodes, nonregular files, unstable reads, invalid UTF-8 and exceeded existing limits. No document SHALL be silently truncated or omitted to fit a limit. Frozen bytes SHALL not change when source files are later edited.

#### Scenario: Invalid member of a set
- WHEN any selected member fails preflight
- THEN no review snapshot is published and the error names the bounded cause and corrective action.

### Requirement: Durable client publication recovery
The manager SHALL persist versioned, permission-restricted publisher bookkeeping outside the server-owned viewer inventory. Every attempted mutating request SHALL have its exact bytes and operation identity durably recorded before transmission. A semantic declaration queued while the viewer is unavailable SHALL have its semantic input and operation identity durably recorded before the command returns, and the requests built when it is sent SHALL follow the same rule. Uncertain outcomes SHALL replay unchanged requests before accepting replacement work. Successful publication SHALL require exact revision/document readback; partial uploads SHALL not be described as a published revision. Unconfirmed inputs, queued declarations included, SHALL survive process exit and SHALL not expire by age. The local management commands (`hive view work rename`, `archive`, `restore`, `pin`, `unpin`, `recover` and `delete`) are exempt from this bookkeeping: after an uncertain outcome they SHALL resend the identical request once, which the viewer answers from its stored receipt, and only when that is uncertain too re-read the work and report whether the action holds; their expected version or a deleted work SHALL keep a retry from applying twice.

#### Scenario: Response lost after server commit
- WHEN the CLI exits after a committed snapshot but before recording its acknowledgment
- THEN resume replays the original operation and confirms the same revision without a duplicate.

#### Scenario: Competing publisher or external change
- WHEN a second publisher holds the client lock, or another actor changes the selected stage/chain incompatibly
- THEN the operation reports busy/conflict without silently overwriting, closing or reopening work.

#### Scenario: Recover from a definitive rejection
- WHEN the server has definitively rejected the exact pending request, every earlier attempted outcome is resolved, and an authorized caller explicitly reconciles that pending intent with a reason
- THEN the client atomically archives its bytes/receipts/rejection evidence locally, leaves remote content untouched and permits a new explicit publication against the current chain.
- WHEN any attempted outcome remains unknown
- THEN archival is refused and exact replay remains the recovery path.
- WHEN the work was closed or deleted
- THEN reconciliation never reopens or recreates it.

### Requirement: Semantic task and lifecycle operations
The manager SHALL accept structured task keys, titles, orders and stage segments and resolve their internal identities/versions for registration and updates. Registration SHALL prevalidate the entire input, preserve partial-batch receipts and skip identical entries without resetting progress. Metadata changes SHALL be explicit. Markdown publication SHALL not infer state, approval, assignment or completion. Stage/task/closure authority and semantic preconditions SHALL remain those of protocol v3.

#### Scenario: Markdown refresh after progress
- WHEN a plan's tasks document is republished after a task starts
- THEN the document revision changes while the live task state remains unchanged.

#### Scenario: Compact transition and separate closure
- WHEN an authorized producer updates task T1 in build using semantic fields and later closes work A
- THEN the correct segment changes without requiring technical version fields, and work B in the same native session remains unaffected.

### Requirement: Bounded publication interface and compatibility

Successful mutation responses SHALL fit 4096 UTF-8 bytes without full source documents or credentials. The work list SHALL retain bounded pages and honest missing-binding state. High- and low-level clients SHALL share protocol v3 operation meanings at matching timeline contract revision 3; a coordinated update SHALL replace the prior requirement to operate against an unmodified older server. Publisher files SHALL remain outside viewer storage. The compact contract and canonical procedure SHALL preserve publication timing, producer authority, human validation and evidence checks while the binary owns bookkeeping. Native pending-operation identities/bytes and retained publication receipts SHALL survive the update for explicit resume.

#### Scenario: Coordinated server and clients
- WHEN high- and low-level clients use the matching upgraded server
- THEN they operate on preserved work histories, keep publisher journals outside viewer storage and retain explicit interfaces for unsupported advanced operations.

### Requirement: Durable closure-episode maintenance

The store SHALL maintain bounded private policy metadata and truthful system provenance using its existing serialized durable-write/reconciliation path, without fabricating native or local_viewer actors or consuming an unbounded public event/receipt stream. Both automatic archival and its manual reversal SHALL have bounded capacity independent of saturated public event/publication-receipt budgets. At most 16 KiB additional private per-record maintenance capacity SHALL be reserved, inaccessible to content, including for valid legacy records already at former serialized limits. Public derived fields SHALL project canonical metadata rather than independently persist editable copies. Physical write failure or numeric version exhaustion SHALL leave the work visible and report maintenance failure instead of hiding work without safe reversal. Successful archive SHALL bump the optimistic work version once; repeated sweeps SHALL be no-ops. Uncertain persistence SHALL NOT be reported confirmed and SHALL reconcile before further maintenance.

Existing records SHALL recover the current explicit closure time from server-recorded transitions or closure receipts after any reopen. An unprovable time SHALL remain unknown and receive a full two-hour window from the first durably recorded eligibility evaluation; restarts SHALL NOT extend that window. Prior manual restore/pin choices SHALL be honored. Malformed records SHALL fail explicitly without destructive defaulting.

#### Scenario: Unknown old closure date
- WHEN an existing closed work has no provable closure timestamp
- THEN closed_at remains null, eligibility starts once on durable initialization, and restarts retain that same deadline.

### Requirement: Coordinated timeline contract revision

The upgraded server, CLI and UI SHALL require timeline contract revision 3 while preserving schema/storage version 3 identities and all existing evidence. Matching revision SHALL be acknowledged before upgraded clients write; old or missing revisions SHALL receive a safe actionable update/reload error before timeline mutations. No simultaneous old-client projection SHALL be required. Bootstrap/authentication and existing file-capability reads SHALL retain their access protections.

The shared session/index projection SHALL expose required last_activity_at, nullable closed_at and autoarchive_at, and nullable archive_origin limited to manual/automatic. Stored-read normalization SHALL accept valid pre-policy records without weakening strict public validation. Contract mismatch SHALL preserve publisher journal inputs, operation identity and stored receipts; it SHALL NOT trigger automatic replay, request replacement or deletion.

#### Scenario: Old tab or CLI
- WHEN an outdated client attempts a timeline mutation
- THEN the server requests update/reload before writing, and the pending operation/input remains available for explicit recovery with the upgraded client.

#### Scenario: Upgraded client against old server
- WHEN the server does not acknowledge the new contract revision
- THEN the upgraded client refuses mutation before sending a write.

### Requirement: Local management actions from the CLI
The manager SHALL offer `hive view work rename`, `archive`, `restore`, `pin`, `unpin`, `recover` and `delete`, each with the same effect, expected version and result as the browser action it mirrors, sent as the `local_viewer` actor with `origin: cli` and without a native producer identity. Only `recover` SHALL take a reason. `delete` SHALL send nothing unless `--confirm` equals the work ID. `recover` SHALL accept the named pending same-project request and make it lead in one commit without inferring that the previous lead's session ended. `/api/health` SHALL announce the capability `local_actions_cli`; when the running viewer does not announce it, the CLI SHALL send nothing and ask to update the viewer. The native publication commands SHALL keep refusing a `local_viewer` actor.

#### Scenario: Delete without confirmation
- **WHEN** `hive view work delete --work-id W` runs without `--confirm W`
- **THEN** nothing is sent and the work is unchanged.

#### Scenario: Older running viewer
- **WHEN** the running viewer does not announce `local_actions_cli`
- **THEN** the CLI sends nothing and names the viewer update as the next step.

#### Scenario: Browser unchanged
- **WHEN** the browser sends a local action without `origin`
- **THEN** it is accepted as before and recorded with `origin: browser`.

### Requirement: Host-observed events
The viewer SHALL accept host events reported by the local `hive hook` process through a capability-gated endpoint authenticated by the local viewer token, without a browser origin. A host event SHALL carry the host, its native session ID, the project directory, an event type, bounded data and an operation ID derived from the host event so a resend applies once; the hook SHALL resolve the session's current work and that work's active stage from the manager's current-work record and send them. Host events SHALL act as the accepted native producer of that session, SHALL be stored with the provenance `host_observed`, distinct from `agent_reported` and `server`, and SHALL be rejected for a session that is not an accepted producer of the work.

Every host event SHALL update the last activity of the works of its native session as maintenance metadata, not as a timeline event, and SHALL NOT invalidate the expected versions of agent mutations. When the current work has no agent-registered task, native task-list changes SHALL become the work's tasks, keyed `host:<native task ID>`, assigned to that session's producer and bound to the active stage, up to 50 tasks; when the work has agent-registered tasks, native task-list changes SHALL add no task and only update last activity, except that a host task already mirrored SHALL keep following the host's state and title, and the agent's own plan and segment numbering SHALL ignore the `host:` tasks. Without a current work or an active stage, a host event SHALL only update last activity. Host events SHALL NOT start, complete or close a stage, change producers or authority, record approvals or validations, or close, pause or fail a work; a session end SHALL update last activity only and SHALL NOT be read as process death. The UI SHALL show host-observed provenance apart from agent-published content. Mirrored tasks SHALL carry their host provenance through the `host:` task key, and the work SHALL expose the time of its last host activity, set only by host events, apart from its last activity.

#### Scenario: Native task completed in a build without a plan
- WHEN the host reports a native task created and then completed in a session whose current work has an active build stage and no agent-registered task
- THEN that work shows the task with host-observed provenance and its completion, and no stage changes state.

#### Scenario: Plan tasks already registered
- WHEN the host reports native task changes for a work whose plan tasks the agent registered
- THEN no task is added, no agent task is changed and only last activity moves, while a host task mirrored earlier still follows the host.

#### Scenario: Event attempts a lifecycle change
- WHEN a host event carries data that would complete a stage or close the work
- THEN the viewer rejects that effect and stores no change to stages, authority or work state.

#### Scenario: Agent write after host activity
- WHEN host events arrive between an agent's stage declaration and its next conditional mutation
- THEN the agent's mutation commits with the versions from its last receipt.

### Requirement: Resolved native identity and current work
The manager SHALL resolve the producer identity of a semantic command, when no explicit identity is given, from the first available source: a host environment variable verified against the host's official documentation and installed version, then a process-to-session map written by the host's session-start hook and found through the process ancestry. A map entry SHALL be used only when it names exactly one live session for that ancestor and is newer than the host's last session change for that process; otherwise identity is unresolved. A resolved identity is the host's actual native session ID; with no source, or an ambiguous one, the command SHALL be rejected as missing identity, and the host's conditional identity reference SHALL say how to pass it explicitly. The manager SHALL keep the current work of each (host, native session) so semantic commands may omit the work ID while exactly one work is current, and SHALL require the work ID otherwise. Opening a work SHALL register its project root when it is not registered.

#### Scenario: Claude Code without flags
- WHEN a semantic command runs in a Claude Code shell that exposes the session ID and omits identity flags
- THEN it publishes as that native session.

#### Scenario: Two sessions under one process
- WHEN the session map holds two live sessions for the nearest host ancestor
- THEN the command fails as missing identity and writes nothing.

#### Scenario: No identity source
- WHEN no flag, verified variable or session map applies
- THEN the command fails as missing identity and writes nothing.

### Requirement: Declarative stage publication without a change folder
The manager SHALL open a work without a change folder or plan revision and SHALL apply, in one command, a stage state with its reason, the registration of missing tasks with their stage segments bound to that stage, task state changes and an optional close that first completes the active stage and then reads the committed work back. Repeating the same declaration SHALL return `unchanged`. Its compact receipt SHALL include task and segment IDs. One evidence command SHALL upload captures, register the candidate when needed and append the event with its view, including a view that supersedes an earlier one. Low-level operations SHALL remain available and unchanged.

#### Scenario: Build without plan
- WHEN a producer opens a work without a change folder, declares a build stage with three tasks and their states, and closes it
- THEN the viewer shows the stage, the three tasks bound to it and the closed work, published by at most three commands, with no separate session read and no binding error.

#### Scenario: Captures for validation
- WHEN a producer publishes two captures with a validation view that supersedes an earlier view in one evidence command
- THEN the viewer shows the new view with both images, linked to the candidate, and the earlier view as superseded.

### Requirement: Queued semantic publication while the viewer is unavailable
When the viewer is not running or does not answer, a semantic open, stage declaration or evidence command SHALL record its semantic input and operation identity durably, report it as queued, and SHALL NOT start the viewer. A queued open SHALL return a local handle derived from the project, host, native session and work key, which the manager maps to the viewer ID once the open is sent. The next semantic command on that work, or the next hook run for its session, with a running viewer SHALL send that work's queued declarations in order before its own, building their requests against the current remote state; a declaration that no longer applies, such as a stage already completed, SHALL be reported as rejected with its reason and SHALL NOT overwrite remote state. Locks SHALL be held per work; commands on different works SHALL NOT block each other.

#### Scenario: Viewer stopped at session start
- WHEN a producer opens a work and declares a build stage while the viewer is stopped, and the viewer later runs again
- THEN both are reported queued, and the next command creates the work and the stage once, in order.

#### Scenario: Remote state moved on
- WHEN a queued declaration would complete a stage that another producer already completed
- THEN that part is reported rejected and the stage keeps its recorded state.

### Requirement: Hook registration per host
The manager SHALL install, update and remove only its own hook entries: entries whose command runs `hive hook` in Claude Code's user settings and Codex's and Cursor's hooks files, its own native hook file in Grok, its own event plugin in OpenCode's plugin directory and its own extension in Pi, leaving every other hook and setting byte-identical and keeping a backup of each edited JSON file. It SHALL NOT write Codex hook trust; its report SHALL tell the user to approve the hook in Codex. Hook processes SHALL print nothing, finish within the host's time limit and exit successfully on any error. Host support not verified against current documentation and the installed version SHALL be reported as unverified.

#### Scenario: Existing third-party hooks
- WHEN Claude Code's settings already hold other tools' hooks and Hive installs and later removes its hooks
- THEN the other hooks and settings stay byte-identical after both operations.

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
