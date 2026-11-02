[canonical-references]: ../../canonical-references.md

## MODIFIED Requirements

### Requirement: Versioned event and immutable artifact publication
The local server SHALL accept typed, bounded protocol v3 events, images and Markdown documents under work-scoped authentication, assign durable IDs/versions/sequences and preserve immutable bytes. Native publication mutations SHALL carry an accepted producer and operation ID. Local viewer management mutations SHALL use the separate local_viewer actor and SHALL NOT impersonate a native producer. Exact replay SHALL return its durable receipt after restart; changed operation payloads, stale mutations and cross-work references SHALL fail without overwrite.

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

### Requirement: Walkthrough ingestion and grouping
The protocol v3 runtime SHALL serve API-owned works from a new v3 state directory exclusively. It SHALL NOT scan or import legacy manifests, API v1 or provisional v2 sessions or copy preferences. Project roots SHALL be explicitly registered in v3. Existing state from an earlier binary SHALL cause cutoff_required before any state mutation rather than migration or silent empty-state fallback. Unknown or corrupt state SHALL fail without cleanup. Legacy standalone HTML SHALL remain independent and SHALL NOT be ingested into v3. API commits SHALL finalize blobs and synchronize files/directories before success and SHALL reconcile outcome_unknown after uncertain durable commits.

#### Scenario: Clean start
- WHEN v3 starts on a new state directory and roots are registered
- THEN its index is initially empty and old project walkthrough.json files do not appear.

#### Scenario: Old directory selected
- WHEN v3 is pointed at earlier viewer state
- THEN startup fails actionably before creating or altering files, preserving that state.

### Requirement: Self-describing protocol and confirmed-input hygiene
Offline contract usage/schemas/limits SHALL describe protocol/storage v3 exclusively and the clean-cutoff policy. New CLI and viewer SHALL require compatible v3 peers for publication. v1/v2 writes/replays SHALL fail before mutation with safe update guidance. No v1/v2 reader, migration or imported evidence reconstruction SHALL be part of the v3 runtime. Cutover SHALL coordinate old writers and service/client/content versions and use a new state directory without modifying the earlier one. v3 receipts, immutable blobs and atomic state transitions SHALL remain durable after restart or uncertain commit.

One canonical publication procedure SHALL own collaboration/native identity, meaningful snapshots, states before user stops, explicit resumption and readable summaries. Consumers SHALL link to it. Children SHALL return evidence to the main producer. Viewer identity SHALL NOT change Engram runtime identity. Hygiene SHALL retain authoritative plans, current review inputs, uncertain uploads and unique evidence.

#### Scenario: v3 restart
- WHEN a v3 publication commits and the service restarts
- THEN IDs, hashes, actor audit and operation receipts remain resolvable without a legacy adapter.

#### Scenario: Old writer
- WHEN a v1 or provisional v2 publisher sends a mutation to v3
- THEN it receives an update diagnostic before mutation.

### Requirement: Store retention and deletion
All v3 work content SHALL use manual retention without age-based expiry. Archive/pin SHALL affect visibility/preferences only. Authenticated same-origin, session-scoped, version-conditional deletion SHALL remove the selected work and its owned stored blobs without editing project files or other works. A durable tombstone and operation receipt SHALL precede cleanup; partial failure SHALL report remaining work and allow scoped retry without resurrection. Reads SHALL require the existing token/capability protections. Restart recovery SHALL complete pending deletion without treating a failed write as success.

#### Scenario: Delete and retry
- WHEN deletion commits but its response is lost
- THEN exact retry after restart returns the receipt and no deleted work reappears.

### Requirement: Safe and deterministic Git tracking lookup
Any retained Git source/identity lookup SHALL disable repository-configured fsmonitor callbacks for that process without changing configuration, neutralize inherited pathspec interpretation, preserve bounded-output handling and avoid conflating distinct filesystem objects. Removing the legacy scanner SHALL NOT weaken these controls in remaining lookups.

#### Scenario: Retained Git lookup
- WHEN a v3 project/source operation needs a Git lookup
- THEN repository callbacks do not execute and inherited pathspec variables do not change the selected path.

### Requirement: Open-file hardlink rejection
Retained file-based upload and standalone rendering operations SHALL inspect the opened descriptor before hashing/reading referenced local resources, rejecting regular files with multiple hard links or unavailable required link-count metadata. Rejection SHALL not publish partial content or replace valid content. API endpoints SHALL receive bounded bytes, never follow publisher-supplied filesystem paths.

#### Scenario: Resource replacement
- WHEN a local referenced resource is replaced by a hard-linked regular file before reading
- THEN descriptor checks reject the read and no new publication is committed.

### Requirement: Explicit walkthrough session lifecycle
The independent standalone HTML fallback MAY keep its existing optional manifest status contract and SHALL NOT be ingested into v3. v3 lifecycle SHALL use the typed work/stage API exclusively; neither age, document completion, host process nor fallback manifest kind SHALL imply approval or conversation termination.

#### Scenario: Standalone fallback
- WHEN the binary is absent and a standalone walkthrough is rendered
- THEN its existing manifest/HTML behavior remains usable without a v3 server or import path.

### Requirement: Confirmed lifecycle ingestion before artifact cleanup
For v3, closure cleanup SHALL confirm the exact committed work/version, state, revision/artifact IDs, hashes and sizes through authenticated API readback before removing owned disposable upload inputs. An unavailable runtime, mismatch or uncertain commit SHALL preserve inputs and report the limit without starting/restarting the viewer. This is API confirmation, not manifest ingestion. Standalone fallback retention SHALL remain an independent artifact unit. Tokens/capabilities SHALL never appear in evidence.

#### Scenario: Uncertain publication
- WHEN exact committed readback cannot be established
- THEN review inputs remain available and no successful cleanup/publication is claimed.

### Requirement: Hierarchical React viewer navigation
The React shell SHALL list API v3 works under registered projects, with title, ticket summary, date and visible count. Archive filtering, sidebar visibility, manual expansion and selection by work ID SHALL survive refresh in the same tab/origin. Search SHALL find work titles and tickets without splitting shared work by producer. Keyboard/mobile navigation SHALL remain accessible; controls, anchors and ARIA IDs SHALL be scoped. Declared lifecycle SHALL remain separate from process liveness; local management actions SHALL occur once in the toolbar. There SHALL be no legacy physical-member selection or manifest-renderer branch in v3.

#### Scenario: Refresh navigation
- WHEN the user changes selection/expansion/filter and reloads
- THEN the same work and manual choices remain selected with its own scroll context.

### Requirement: Reversible copy archiving
The retained archive behavior SHALL apply to a v3 work rather than legacy copies. Authenticated same-origin scoped local-viewer writes SHALL atomically update archived under expected version; archive SHALL not modify content, source identity, lifecycle or pin and SHALL survive restart. Archived works SHALL be hidden by default with the accessible include-archived toggle. Undo SHALL target the originally archived work even after selection changes. No archive SHALL enable TTL deletion.

#### Scenario: Undo elsewhere
- WHEN archiving hides work A and the user selects B before Undo
- THEN Undo restores A without changing B.

### Requirement: Explicit walkthrough delivery links
v3 API delivery events and independent standalone pages SHALL accept only verified PR/commit links, distinguish candidate metadata from delivery and never infer integration. URLs SHALL be absolute HTTPS with valid ASCII hosts and no credentials, spaces or controls; labels SHALL render as text, destinations SHALL be validated on rendering, and new contexts SHALL use noopener noreferrer. The v3 compact header SHALL expose deduplicated canonical work delivery entities by ID in its accessible delivery popover; no links SHALL be fabricated from candidate text. Standalone manifests/HTML may retain their existing validated link rendering independently, without import or backfill.

#### Scenario: Unsafe URL
- WHEN a delivery contains an executable, relative or credential-bearing URL
- THEN validation rejects it and no unsafe browser link is created.

### Requirement: Viewer-first walkthrough publication
With the binary present, substantive publication SHALL use only the compatible v3 API and confirmed readback. A stopped/incompatible viewer SHALL preserve inputs and report publication unavailable without starting a daemon, emitting a legacy viewer manifest or switching to HTML. Only absence of the executable SHALL select the independent existing standalone HTML fallback. Previously generated HTML and selected artifact units SHALL not be deleted or rewritten by this cutoff.

#### Scenario: Binary present, server incompatible
- WHEN publication finds a v1 server
- THEN it requests the coordinated v3 update and preserves inputs without manifest import or HTML substitution.

### Requirement: Fixed session toolbar
The selected v3 work title and management actions SHALL stay visible outside its scrolling timeline/document canvas. Deep-linked stage/event headings SHALL remain visible below the toolbar. Keyboard actions, unbroken long names through bounded local scrolling, narrow layouts without page overflow and print omission of management controls SHALL remain supported. No physical walkthrough segment controls SHALL be required.

#### Scenario: Narrow long title
- WHEN a valid unbroken title exceeds a narrow viewport
- THEN local keyboard-accessible title scrolling exposes it without page-level overflow or hiding actions.

### Requirement: Local session name and preferences
v3 title/archive/pin management SHALL use work-local typed API state and local_viewer provenance, not imported logical-copy preferences. Names SHALL be trimmed, nonempty, bounded to 200 Unicode code points, reject controls and render as text. A canceled edit SHALL write nothing; a failed save SHALL retain input and a retryable error. Renaming SHALL preserve source/producers, stage titles and content, remain searchable and survive restart. Corrupt persisted state SHALL fail explicitly rather than enable defaults that delete content.

#### Scenario: Rename and restart
- WHEN the user saves a title and restarts v3
- THEN the title is searchable and source, producers and stages remain unchanged.

### Requirement: Explicit session mutation scope
Browser title/archive/pin/deletion and recovery SHALL require existing authentication, same-origin/Host checks, explicit session action scope and the v3 local_viewer actor. Obsolete clients SHALL receive a safe reload/update rejection before mutation; the UI SHALL not replay a rejected write automatically. Native publisher operations SHALL follow the separately typed producer authority rules, not claim a browser actor.

#### Scenario: Old tab
- WHEN an old browser tab sends an obsolete management write
- THEN v3 rejects it without changing work and offers reload instead of automatic replay.


## ADDED Requirements

### Requirement: Local viewer management actor
The v3 actor union SHALL distinguish accepted native_producer from local_viewer. Browser rename, archive/restore, pin/unpin and confirmed deletion SHALL remain available without native identity and SHALL be audited as local viewer actions, not native publication or authenticated human approvals. local_viewer SHALL NOT publish documents, evidence or approval events or operate stage lifecycle. An explicit recover_lead management operation MAY accept a pending same-project join and replace the lead atomically without changing declared stage/work state. Mixed actor branches SHALL fail. Deletion SHALL preserve an auditable receipt/tombstone without retaining or resurrecting deleted content.

#### Scenario: Browser management without native identity
- WHEN the authenticated local viewer renames, archives or pins a work
- THEN the action commits with local_viewer provenance and no fabricated producer ID.

#### Scenario: Confirmed browser deletion
- WHEN the user confirms deletion through the existing scoped browser flow
- THEN deletion retains its durable operation result for retry without a native actor or content resurrection.

### Requirement: Shared work producers and lead authority
Protocol v3 SHALL preserve the original source tuple and opaque viewer_session_id as work identity, accepting additional real native producers only by explicit join request and lead acceptance within the same registered project. A request SHALL NOT confer membership. Native lead transfer SHALL be version-conditional to an accepted member. Explicit local recovery SHALL atomically accept a chosen pending same-project request and reassign lead under expected version, durable operation receipt, reason and audit. Recovery SHALL NOT infer process death, stop processes, reopen terminal works, change original source or alter stage/evidence state. The previous lead SHALL remain a collaborator without lead privileges. Original tuple creation SHALL remain idempotent. Joining SHALL NOT merge historical works or reassign IDs.

Pending, foreign-project and unauthorized publishers SHALL fail unchanged. Producers SHALL use their actual native IDs, never another session's or invented IDs. Provenance SHALL be declared rather than represented as cryptographic host authentication; local capability remains the authentication boundary. No historical actor or content SHALL be synthesized during cutoff.

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

### Requirement: Independent task progress and task detail
Tasks SHALL belong to the work with stable task_id, unique task_key, short title, plan reference when available and explicit required stage segments. The lead SHALL declare the full plan including future tasks without needing fake stage occurrences. Each segment SHALL have planned_stage, local state, accepted assignee, order and an explicit binding to a real matching stage occurrence when available. Stage panes SHALL display only their bound segments; future-stage tasks SHALL remain accessible in the plan and global count, not inside Build.

A task may span verify and delivery without becoming multiple tasks. Each stage projection SHALL show only its segment state; the task Sheet SHALL show aggregate state and the complete stage-labeled history. Global count SHALL deduplicate task_id and count completed only when all its required active segments are complete. Stage counts SHALL use only their assigned segments. No foreground change SHALL automatically move, clone or complete a task/segment. Assignment changes SHALL be explicit, versioned and audited.

Segment states SHALL be not_started, in_progress, in_review, blocked and completed. Aggregate task state SHALL be blocked if any active required segment is blocked; otherwise in_review if any is in review; completed if all are complete; not_started if all are not started; in_progress otherwise. State SHALL be displayed only in the row indicator with empty circle, loader, pause, alert or check and accessible text/tooltip. Indicators SHALL use semantic context colors alongside their distinct shapes and accessible labels; the icon stroke SHALL NOT receive a second decorative container border. Neither index, commits, checklist text nor process observation SHALL infer completion. No heuristic percentage SHALL be displayed.

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

## REMOVED Requirements

### Requirement: Import existing copy preferences
**Reason:** The user selected a clean cutoff, with final v3 replacing the provisional v2 without legacy preservation or migration. Old viewer state remains outside the v3 runtime and is not imported.
**Migration:** None. Start v3 on a new state directory; do not modify or silently consume the previous directory.

### Requirement: Current source classification
**Reason:** v3 does not ingest or serve source-manifest copies; tracked/local/unknown placement and scanner warnings have no v3 input. This does not remove safe retained Git/file checks or actor provenance.
**Migration:** None; no copy or classification state is imported.

## Operational constraint

Building and testing this candidate SHALL use an isolated output binary, port and v3 state directory. While CAM depends on the installed binary/viewer, this work SHALL NOT replace that binary, run the installed-content refresh, restart the habitual viewer or change its configuration/state. The maintainer auto-refresh exception SHALL NOT override this user constraint. A later real cutover SHALL require a separate concrete instruction.


## ADDED Requirements

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

## ADDED Requirements

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
