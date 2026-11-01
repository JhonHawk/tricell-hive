# Design

## Decisions

Use the existing Go loopback server, authenticated API, React/base-nova shell and local JSON persistence. New sessions are API-owned. Legacy source manifests are compatibility inputs only and cannot overwrite an API session. Preserve the sidebar; redesign its content pane. Native host identity is metadata for uniqueness, not cryptographic attestation of which process sent a request.

The accepted UI composition follows https://reui.io/preview/base/timeline-4?ref=mcp (primary feed) with action/status organization from https://reui.io/preview/base/timeline-3?ref=mcp and https://reui.io/preview/base/components/c-button-group-5?ref=mcp. Use type icons instead of actor avatars; this is a dated feed, not a completed-step wizard. No nested milestone cards, new entrance motion or decorative progress states. No new database or event-sourcing framework.

## Proposed shape

### Identity and session authority

Shared instructions name capabilities, not host tool identifiers: the host's native session identity, terminal/process execution, file access, browser/capture and native question capabilities. They must not prescribe a universal tool name or environment variable. The API field names and canonical host values are data-contract identifiers, not a per-host tool procedure.

Keep acquisition in one directly linked, conditionally read host-identity reference. Prefer a host-documented identity already exposed to the current runtime; consult that host's verified discovery mechanism only when needed. Before documenting a host-specific mechanism, verify current official documentation and installed version, separating documentary support, runtime observation and inference. A table or screenshot from another session is a lead, not current proof. Read only necessary identity values, never dump the full environment. Do not guess aliases, synthesize a timestamp or claim uniform support across all hosts. If no verified native ID is available, explain that the required capability is unavailable and stop API registration while independent work continues. No model pilot is needed or authorized to establish this mapping. The user-selected Codex /review mechanism remains specific because its behavior and prior authorization are specific to this session.

Create requires schema_version:1, source {project_root, host, native_session_id}, title, summary, tickets, operation_id. Root must resolve to an existing registered repository/workspace using the current resolver and ProjectKey rules. Server derives project_key; reject a caller-supplied project key or label override. Persist each canonical root association and bound project key in the API session record (including aliases learned from equivalent worktrees). On registration/resume, recover this durable association before hashing, including after restart with only API records. A temporary remote-read failure must reuse a known binding; without a binding, fail explicitly when identity resolution is uncertain rather than adopt a path fallback accidentally. A deliberately remote-less registered workspace may use its verified canonical path. If authoritative metadata contradicts an existing binding, return project_identity_changed and require explicit resolution outside this flow; never silently split/reassign a native session. Index labels may refresh without changing bound identity. Host uses canonical values codex, claude-code, grok, pi, opencode, cursor; reject unknown aliases rather than splitting identity. native_session_id is an opaque nonempty UTF-8 string of at most 256 characters with no whitespace/control characters; do not require UUID syntax, case-fold it or generate it. The agent obtains it from confirmed runtime context; missing facts block creation only.

Uniqueness is the exact tuple (resolved project_key, canonical host, native_session_id). Tickets/title/branch/root aliases are not identity. Worktrees resolving to the same project key reuse the tuple. Viewer ID is assigned by the server as a namespaced SHA-256 of the canonical tuple; reuse the existing 64-hex shape and JSON-tuple canonicalization, with a timeline-api-v1 namespace to avoid legacy-copy collisions. Expose viewer_session_id separately from native_session_id. GET-or-create is serialized and durable; repeated/concurrent creates return the same ID and existing content, not overwrite title/summary/tickets. Return created:false on an existing tuple; explicit conditional update changes metadata. Closing/archiving does not create a new identity or silently reopen it. Deletion preserves a bounded identity tombstone and registration returns 410 rather than resurrecting it implicitly.

### Records

Session: storage schema_version, immutable source identity, server ID/created_at, version, title, summary, state (active|awaiting_validation|paused|closed), archived/pinned, tickets, events, artifacts and bounded operation receipts. State is explicitly updated by the producer; the UI does not infer an active model process or approval from it. Creation defaults to active, unarchived and unpinned. Read/index results include mode:timeline_api|legacy and validation_count. There is no automatic age-based retention. Neither reads nor inactivity trigger deletion or age-based ingestion exclusion. Pin remains a user preference, not an exemption required to preserve evidence. Archive changes visibility only. See R3 retention compatibility below.

Ticket: key, optional title, verified HTTPS URL when known. Use URL plus key as identity; missing URLs render unlinked keys rather than guessed links. Session tickets are a deduplicated union and cannot be removed while referenced. Treat strings as text and reuse existing safe URL validation. No tracker polling implied.

Event: server ID, server sequence, created_at/updated_at, version, kind, title, summary, ticket_refs[], optional related_event_id and candidate {branch,commit,base}. Kinds: request, work_started (legacy API draft compatibility only), stage, validation, changes_reported, iteration, approval_reported, delivery, closure. New producers use stage for lifecycle progress. The server sets origin:agent_reported for producer-created records; clients cannot claim to be a human UI author. Reported user decisions require a reference to an existing event/evidence_id and a short source reference from the conversation. They are not authenticated browser approvals and are labeled accordingly. Delivery carries safe PR/commit links and declared open|merged|local status; closure carries completed|paused|abandoned reason and explicit result. An open PR never implies closed work.

Validation/iteration includes views[]. Each view has a server-assigned view_id on first introduction (reuse returned ID thereafter), title, ticket_refs[], ordered steps, expected_result, image_refs[], and optional supersedes {event_id,evidence_id,view_id}. One view may cover several tickets; distinct views remain separate inside one event. An iteration lists only affected views and refers to prior evidence; unmentioned views remain on earlier nodes. Every validation/iteration receives a server-assigned immutable evidence_id, distinct from its mutable presentation version. The evidence payload (views, steps, candidate, artifact references) is never patched. Decisions and supersession target event_id + evidence_id + optional view IDs, not event.version. A copy edit increments event.version only; all evidence references remain resolvable after edit/restart without storing historical presentation snapshots. All referenced tickets/views/events/artifacts must belong to this session. Server-assigned IDs for events/views/artifacts are opaque random identifiers; do not derive a view identity from its label. The top-level event ticket union must include its views' tickets; the server computes this union rather than trusting inconsistent duplicates.

Artifact: immutable server ID, session ID, server-computed sha256/size/media type, received_at, reported captured_at/candidate/viewport/theme metadata. Accept PNG/JPEG/WebP only, at most the current 20 MiB limit per file; validate declared type against supported signatures and do not accept arbitrary URL/local-path references in HTTP payloads. CLI reads the local file and uploads bytes. Uploads are staged and finalized atomically before attachment; unreferenced staged files are bounded and cleaned after interruption. Artifact reuse within the session returns the same ID/origin; no cross-session artifact references or silent byte replacement. UI distinguishes original capture metadata from server receipt time.

### API and CLI surface

Proposed authenticated routes (not existing commands):

| Route | Contract |
| --- | --- |
| POST /api/timeline/sessions | Validate required source, get/create; operation_id |
| GET /api/timeline/sessions/{sid} | Full typed session, version; no mutations |
| PATCH /api/timeline/sessions/{sid} | expected_version, operation_id and allowlisted title/summary/state/tickets/archive/pin changes |
| DELETE /api/timeline/sessions/{sid} | expected_version and operation_id; durable tombstone, then complete-session cleanup |
| POST /api/timeline/sessions/{sid}/artifacts | Bounded multipart metadata+image, operation_id |
| POST /api/timeline/sessions/{sid}/documents | Bounded UTF-8 Markdown bytes/metadata, operation_id; returns document_id/hash/size/session_version |
| GET /api/timeline/sessions/{sid}/documents/{did} | Authenticated immutable document read; no mutation |
| POST .../{sid}/events/{eid}/document-snapshots | operation_id, expected_version, tabs, draft-or-complete; returns document_revision_id/event_version/session_version |
| POST /api/timeline/sessions/{sid}/events | Strict kind-dependent fields, operation_id; server assigns order/IDs |
| PATCH or DELETE .../{sid}/events/{eid} | expected_version and operation_id; presentation edits or withdrawal; stage transitions additionally compare expected_session_version |

Use existing API index/access handoff and capability-backed image serving, extended with typed API-session entries. No credentials in JSON output, URLs, errors or CLI arguments. Every mutation retains token, Host/Origin checks and session scope; reject cross-session references. Bound bodies/field lengths before allocation; reject duplicate/unknown JSON keys and trailing data. Limits: 2 MiB session JSON, 500 events, 50 views per event, 100 image refs per view, 1000 operation receipts. Reject content additions above 1.75 MiB or 992 receipts, reserving capacity for lifecycle/pin/archive/closure operations. Terminal DELETE receipts live in the tombstone and are never blocked by document/content quotas; fail visibly at a limit, never silently prune accepted evidence or idempotency receipts. No HTML/script execution or remote-media fetching from event text.

CLI adds hive view session create|get|update|delete, hive view event create|update|withdraw, hive view artifact add. JSON bodies use --json FILE or stdin; file upload uses --file with the source path held locally. Auth/port/state discovery stays in CLI. Output is machine-readable IDs, versions, committed event sequence and operation result. The agent retains those receipts in its current work checkpoint, not per-event persistent memory. Open gains optional --session-id/--event-id, validated and carried as a nonsecret target alongside the existing one-use access-code handoff. Plain open remains compatible and is not repeated on every update. No automatic server startup.

All mutation operations persist operation_id and canonical request digest with their result before success. Replay lookup precedes expected-version and lifecycle checks: an already completed operation recovers its original result even when the current record version advanced. Tombstones retain completed operation receipts plus the terminal DELETE digest/result, independently of removed session content. After a lost DELETE response/restart, the same operation returns the completed deletion result; a changed body under the same key conflicts. Unknown operations against deleted identity return 410. An interrupted deletion persists its recovery intent and resumes on replay; only completed operations promise an identical success result, while pending/partial failures report current remaining cleanup. Replaying the exact operation returns its original IDs/result even after restart; same ID/different request is 409. Compare mutation versions under the same store lock used to commit. Event append does not need an obsolete session version because it does not replace siblings, but it must reject closed/deleted sessions and serialize sequence assignment. Editing title/summary changes updated_at/version, not sequence/created_at. PATCH cannot change event kind, scope, steps, images or candidate evidence; substantive changes create an iteration. DELETE marks an event withdrawn; referenced decision/evidence records remain as visible placeholders with their relationships, not hard-deleted content disguised as a fresh timeline. Whole-session deletion is the explicit irreversible cleanup route.

Persist one atomic authoritative session JSON and its private artifact directory in a separate API-owned namespace. Deriving the session key from source identity avoids a second uniqueness-index transaction. Operation receipts and event mutations commit in that same JSON. For the new API namespace, durable publication orders: stage and fsync artifact bytes, rename final artifact, fsync its directory, then stage and fsync session JSON containing artifact metadata/receipt, rename it, and fsync its directory before success. Use existing serialized access plus a narrowly scoped durable writer; the current legacy helper alone does not fsync the directory and is insufficient for this guarantee. Failures before rename preserve the previous record. A failure after rename/directory-sync yields outcome_unknown rather than a false rollback claim: reconcile the on-disk record and operation receipt before accepting further mutations, then recover through the same operation ID. Directory-sync failure cannot produce a reported successful durable commit. Do not expose a session reference before its artifact is durable. Reuse existing crash-recovery conventions without changing unrelated manager transactions. Restart validates records and rebuilds the readable index without moving/replacing corrupt evidence or weakening identity. Partial artifact/session deletion uses recovery markers and exact remaining counts. Keep API tombstones so neither create retry nor scanner resurrects a deleted identity.

### Legacy and unavailable-service behavior

Do not migrate old documents into guessed events. Preserve legacy reading and existing standalone HTML renderer; legacy content is labeled historical and stays outside event editing. Scanner never writes an API-owned record and suppresses legacy projection of the same exact native tuple when an API record owns it, without deleting source files. An API tombstone retains that suppression. Existing legacy rules for tracked sources/unknown classification and tombstones remain. The session index supports both modes without rebuilding sidebar navigation.

With Hive installed but server stopped/unreachable, API publication fails with actionable startup guidance, leaves local source captures intact and never says published; it neither starts the server nor emits an alternate session authority. With no Hive executable, keep the existing legacy static HTML fallback, clearly a standalone document with no server session/event IDs or interaction. Do not silently import that document later into an already API-owned tuple. New API sessions have no automatic HTML export in this delivery. Flow-close confirms API source identity, committed session version/event sequence and retained artifacts; legacy close keeps its existing exact-copy confirmation.

## UI

Surface operational; archetype Record detail; existing Frame-led shell; dense controls; regions header/action group, short scope summary, timeline; no new entrance motion. Preserve sidebar/project grouping/theme/navigation preferences. Follow installed Button/Badge/Frame/Input/Dialog and ReUI timeline-3/c-button-group-5 compositions; add only missing registry source files after inspecting them, not new package versions. Timeline step numbers are rendering positions, never authoritative progress/approval state; control default completion so reverse ordering does not mark all prior nodes approved.

Header uses ButtonGroup plus ButtonGroupText for status, default-size Badge outline for readable state text on the existing dark-only surface; Button outline size sm and icon-sm, Tooltip for icon-only labels. Pencil changes title to installed Input; Enter saves, Escape restores, failed save keeps the edit with inline Retry, and concurrent version conflict asks reload without overwriting. Archive remains direct; overflow uses DropdownMenu (Pin/Unpin and Delete); Delete opens AlertDialog, disabled while pending. Tickets/Git use Popover containing native links with accessible text; missing links show an explicit unavailable value, not a fake anchor. No avatars, repeated delivery card or extra toolbar index. At 320px the action strip wraps into reachable groups; nothing is dropped.

Before header:

```text
Project / ticket
Session title
Date · 2 walkthroughs         [many actions]
[Session / deliveries card]
[Index of walkthroughs]
```

After header (ButtonGroup aligned far right on desktop):

```text
CAM
Profile, invitations and filters                 [Edit]
31 Oct 2026 · 2 walkthroughs
              [dot Awaiting validation][2][1][Archive][…]
Improve profile spacing and chat filter readability.
```

Before content:

```text
Result / checks / metadata
Walkthrough 1: steps and captures
Walkthrough 2: steps and captures
Environment / provenance
```

After content (Timeline-4 feed, 512px maximum, centered in canvas):

```text
19:10  Iteration 2 — compact filter          COM-2
  Chat mobile · supersedes validation 1
  [thumb 320][thumb 390] [Instructions]
18:55  Changes requested · reported from chat
  Show only the filter icon on mobile.
18:30  Validation 1                      COM-1 COM-2
  Profile · COM-1
  [thumb desktop][thumb mobile] [Instructions]
  Chat · COM-2
  [thumb desktop][thumb mobile] [Instructions]
17:30  Work started · fix/profile-chat
17:20  Request                          COM-1 COM-2
```

Shared screen variant: one Chat view labeled COM-1 COM-2, one thumbnail collection, one instruction dialog. Different screen variant above uses two view groups. Session header links are the union, never a separate timeline per ticket. The walkthrough count counts nonwithdrawn validation/iteration publications, not arbitrary event nodes. An older view says Superseded with a link to the new event; if some views in that event remain current, label only the superseded view. Later copy edits do not reorder events.

Instructions/carousel before: inline step list and existing capture enlargement. After:

```text
Dialog: Chat mobile · Iteration 2             [Close]
Desktop width up to640px; mobile within viewport
1. Open Conversations at 390px.
2. Activate a filter and close the panel.
Expected: icon stays beside the search field.
```

Use installed Dialog size default, scroll content without changing component padding/height/radius/motion. Carousel is scoped to event ID/evidence_id/view ID, shows viewport/theme/capture origin and supports arrow keys/Escape; reuse existing image behavior with typed immutable artifact URLs. Live changes must not silently swap the images/instructions under an open dialog: keep its snapshot and show an update notice. On close return focus to the triggering view control. Long names have ellipsis plus full accessible text, with no horizontal page scroll.

States: each data region has Skeleton loading, empty text with Retry/Reload, inline error with Retry and preserved prior content; rejected auth uses current access reopening flow. No tickets/deliveries popovers say No tickets/No deliveries rather than disappear unexpectedly. Timeline empty says No milestones yet and Reload. Mutation errors remain adjacent to their control. Summary is the short objective/scope of the plan, or request when no plan exists; not the latest event repeated. Checks/environment/provenance remain in stored evidence or source reports where relevant but are not rendered as content sections. No approval/comment buttons in this release; reported chat decisions remain visibly agent-reported.

## Copy

Use existing Spanish/English language selection. Add paired text: En curso/Active; Esperando validación/Awaiting validation; Pausada/Paused; Cerrada/Closed; Recorridos/Walkthroughs; Solicitud/Request; Trabajo iniciado/Work started; Validación/Validation; Iteración/Iteration; Cambios solicitados — registrados del chat/Changes requested — reported from chat; Aprobación comunicada desde el chat/Approval reported from chat; Entrega/Delivery; Cierre/Closure; Instrucciones/Instructions; Sustituida por/Superseded by; Actualización disponible/Update available; Evento retirado/Withdrawn event. Icon actions all have names and tooltips. Status is never Live without an actual liveness contract.

## Impact

Code: tooling/viewer server/store/scan/group/index and new typed timeline files; tooling/cli view commands/client/help; frontend api/controller/shell/application plus timeline/evidence/dialog components and tests. Existing manifest validator/renderer remains legacy compatibility, not a new v2 authoring route. No DB dependency.

Publication procedure stays in flow-walkthrough; R3 places review/capture policy canonically in flow-verify and retrieves protocol syntax from hive view contract. Reconcile flow-research/plan/build/verify/close, flow-walkthrough and hive-review-ux so only the verify-owned conditional reference states proportional capture rules. Check global guidance, all role/skill consumers, site flows, architecture docs and pilot detectors for conflicting publication/status assumptions. Update README and llms for new commands/required API publishing behavior. Target global guidance byte delta zero; record actual delta and checked surfaces in eventual PR. No pilot or synthetic regression fixture additions. Native IDs remain host-owned; do not add a host authentication/model adapter.

## Recovery

Version API storage explicitly. Old binaries cannot interpret API records; coordinate binary/content rollback and retain API files/artifacts for later recovery rather than deleting or downgrading them. Legacy files stay readable and no startup conversion removes data. Server shutdown is user-controlled; integration refresh restarts an already-active maintainer viewer only with its recorded settings.

Historical cleanup keeps the previously approved fixed inventory at _support/workspace/2026-10-31-viewer-revisions-proportional-verification/cleanup-inventory.json (32 copies/27 groups; 29 archived, no pinned/excluded at inventory time). Add an optional expected_members key/hash precondition to existing legacy group deletion, checked under its Store lock before tombstones/removal. Changed/new groups fail without deletion. Do not expand the inventory, delete source files or treat the one existing ingest error as disposable evidence. New API sessions are never in that allowlist. Preserve registry/auth/settings/tombstones and report partial failures. No new bulk-delete UI.

2026-10-31 D5 correction: current first view is not accepted. Align the action strip to the far right of the header; center a bounded feed in the content canvas; use the larger node/icon scale explicitly requested from timeline-4. Repair missing native preflight for portalled Popover/Dropdown content without editing installed component internals or retiming motion. Instructions use the existing Dialog with wider desktop bounds and spaced semantic steps; source padding/radius/motion remain. Update sketches to these corrections before closing T8.

Measured reference2026-10-31: public timeline-4 feed512px wide; indicators24px, vertical item start36px, marker left offset28px. Keep the24px indicator/16px icon and axis composition values. The later user width correction below supersedes the512px feed cap; do not resize component internals.

2026-10-31 user correction: keep the shell and component composition as native to installed ReUI as possible. Reuse Sidebar/SidebarInset, component variants and reference composition; remove compensatory margins and unnecessary width overrides. Preserve installed padding, radius, shadows and motion. Only explicit layout requirements justify composition constraints: centered timeline-4 feed, right-aligned actions, readable instructions and reference marker scale. Portal reset support restores the native component baseline instead of restyling individual overlays.

## R3 — Stage lifecycle, document snapshots and compact publication ownership

### Native composition and sketches

Keep the approved timeline-4 centered feed and compact header. Use timeline-1 only for declared active/completed indicator semantics; no owner/avatar card duplication. Reference: https://reui.io/preview/base/timeline-1?ref=mcp. Reuse installed Timeline, Badge outline, Button outline size sm, Dialog/Header/Footer/Cancel. Add native Tabs from the same base-nova registry (no local restyling). Use an installed Spinner if available, otherwise the selected icon library's loading indicator with accessible text and reduced-motion support; stage state must never depend on animation alone. Do not infer process liveness from a spinner.

Before (R2):
```text
  [iteration] Filter correction
              [thumbnails] [Instructions]
  [request]   Original request
```
After (R3, newest first):
```text
  [spinner] Plan revision 2       [In progress]
            Add stages and retained documents
            [View plan]  21:40
  [check]   Build                 [Completed]
            [Evidence]
  [check]   Plan revision 1       [Completed]
            [View plan]  20:18
  [check]   Research              [Completed]
            [View research]
```
Document dialog:
```text
  Plan revision 2                 [Draft]
  [Proposal][Design][Tasks][Criteria]  (Tabs)
  --------------------------------------------------
  Read-only Markdown, scroll inside the dialog
  Headings / lists / tables / code / verified links
  Snapshot date + revision                 [Close]
```
Tabs wrap or scroll within their own region on narrow screens; the page must not overflow. Loading uses Skeleton; empty documents offer Reload; fetch errors stay inline with Retry; unavailable/auth-denied reads retain context and use existing access recovery. Dialog size follows the approved readable instruction dialog; use native spacing/motion. Latest snapshot is default; a native revision Select permits an explicit historical selection. Opening captures an immutable snapshot; newer publication shows Update available without replacing open text. No inline editing or browser feedback controls.

### Stage and revision contract

A stage event adds stage: research|plan|build|verify and stage_state: in_progress|completed|paused|failed, started_at, optional finished_at, and document_snapshots[]. Times are server-owned. Optional immutable occurred_at is explicitly agent-reported source provenance, never a server start/completion time and never a sorting key. Validate its timestamp syntax, label it separately, and omit it when the source time is unknown. One event is one occurrence; completion PATCH updates state/version/time without changing sequence/created_at or creating a second completion node. Re-entering a finished stage creates a new related occurrence (e.g. plan revision 2); semantic publication, not every file write or command, is a milestone. A paused occurrence may resume with conditional PATCH; completed/failed occurrences are terminal. A single primary producer owns stage transitions; children report evidence to it and do not publish independent stage sessions. Concurrent reviews do not implicitly change the foreground stage. Expose current_stage_event_id as an explicit nullable session field, updated under the session lock; entering a new foreground stage pauses a previous in-progress stage atomically rather than leaving contradictory active indicators. Stage-start requests include expected_session_version so concurrent foreground transitions reject stale intent instead of silently pausing a newer stage; only this kind-dependent event append requires that session precondition. Completing/pausing/failing clears the pointer when it targets that event. Stage status mutations compare expected_version (event) plus expected_session_version and atomically return event_version, session_version and current_stage_event_id. Withdraw of the foreground stage is rejected with stage_active until a version-conditional pause/finalize clears it. Setting session state awaiting_validation atomically pauses its foreground occurrence and clears the pointer; resumption explicitly updates that occurrence and the session state to active under both versions. Waiting for a human is awaiting_validation, not an assertion the agent is actively running. Closure/direct closed rejects an active foreground stage. The producer first conditionally completes or pauses that occurrence, then closes the session with its current version; no extra atomic-finalization payload is introduced.

Document uploads are session-owned immutable text/markdown UTF-8 blobs, separate from image artifacts. Server assigns document_id, computes sha256/size and receipt time. Snapshot publication appends an immutable document_revision_id and ordered tabs [{label, document_id}], draft|complete label and server timestamp to its stage occurrence, with operation_id and expected_version. It may not replace an existing revision. A plan occurrence can publish a draft then a complete snapshot; a later requested scope revision creates a new linked plan occurrence. Approval references event_id plus document_revision_id for documents, or event_id plus evidence_id for UI evidence, as a strict discriminated target. An immutable revision remains readable after stage copy/status changes and restart; withdrawal leaves referenced placeholders. Snapshot publication is an atomic mutation, not a broad PATCH accepting arbitrary history. Add POST /api/timeline/sessions/{sid}/documents for bytes/metadata, GET .../documents/{did} for scoped reads, and POST .../events/{eid}/document-snapshots for operation_id, expected_version and immutable tabs. Existing event PATCH accepts stage_state only for stage events with valid transition preconditions; normal evidence PATCH remains presentation-only. CLI mirrors document add and stage start|update|snapshot without exposing transport credentials. Session/index contracts include current_stage_event_id; get/contract distinguishes document and image payloads.

Limits: each document256KiB, at most8 tabs per snapshot,64 unique documents and50 snapshots per session. Store only metadata/references in the bounded session JSON; document bytes have a16MiB aggregate cap independent of the existing bounded image policy. Reject additions at capacity without pruning old content; reserved lifecycle/delete capacity still applies. Same operation/different bytes conflicts; same-session content deduplication must return original metadata, never cross-session bytes. No arbitrary filesystem paths, remote Markdown fetching or iframe content. Use existing authenticated artifact transport/capability boundaries for document reads, with distinct text MIME, nosniff and no executable response rendering.

Use a standard read-only React Markdown renderer with GFM tables (react-markdown plus remark-gfm), with raw HTML disabled, no rehype-raw, no remote image rendering, and the existing strict safe-link validator. Render code as text; links require explicit safe HTTPS and noopener/noreferrer. Check actual publisher, maintenance/advisories and installed compatibility before adding these packages; do not invent a Markdown parser. This is a justified document-rendering dependency, not a new app framework. Record package lock changes and review the boundary. Source plan files stay authoritative in openspec; uploads are immutable snapshots, not an alternate editable plan.

### Self-describing CLI and skills

hive view contract runs offline and emits a versioned JSON envelope with protocol_version, supported capabilities, commands/usage, input/output schemas, limits and minimal examples. Optional --operation narrows output so the agent reads only the needed contract. Generate/check against authoritative Go contracts and shared fixtures; do not hand-maintain a second contradictory schema. Include cross-reference and lifecycle invariants which JSON Schema alone cannot enforce. Runtime feature/version negotiation rejects unsupported operations and never retries through a guessed legacy authority. This is a new unshipped API draft: revise Go/TS fixtures together, keep legacy manifest schema1 readable, no migration of fabricated old stages.

flow-walkthrough owns session discovery, registration/resume, publication, receipts, readback and publication failure handling; its short entry reads the CLI contract instead of embedding full schema examples. It does not own code-review, UI-capture breadth or approval gates. flow-verify owns those policies (one directly linked conditional UI reference), while flow-build retains the first-view trigger and existing plan owns its review gate. Each research/plan/build/verify/close skill gets at most one short stage-specific pointer paragraph; no new global rule, extra approval question, hook or per-task memory write. Static content/link tests prove the pointers and single owners; no model-reliability claim.

Create/resume at the beginning of an explicitly invoked substantive flow, including research or plan, after resolving project/host/native identity; simple factual replies do not allocate a session. Reuse the same tuple across stages, including this plan expansion. If a viewer is stopped/unreachable, record unpublished state in the existing work checkpoint and continue authorized research/planning/build; no automatic server startup or extra authorization question. Preserve publication inputs and report availability at the handoff. First rendering of a deliverable may open the viewer once; stage updates do not reopen it. Never claim a completed upload or live model activity without evidence.

### R3 retention and hygiene compatibility

Remove age-based Purge scheduling and old-source age rejection for both legacy and API modes. Previously unpinned/archived content remains across clock advances and restarts. Legacy --ttl is accepted only as a deprecated no-op with a clear warning; it must not silently re-enable expiry after the maintainer restarts with recorded service arguments. Report retention: manual, ttl_seconds:0 and compatibility ttl:0s. Preserve port/state/registry/auth/preferences and do not restore already deleted content. Explicit Delete, inventory-guarded cleanup, and narrowly identified uncommitted staging recovery remain; no new bulk/manual cleanup feature is added now. Existing pinned preference remains available without inventing a new retention meaning.

Once a publish is durably committed, confirm the exact source tuple/session, document revision/event IDs, version/sequence and artifact hashes/sizes through readback before considering source upload inputs disposable. Keep local source inputs for the current session; flow-close/session-end hygiene removes only its owned disposable unit under the existing retention/pause/selection rules. A publish receipt alone or generic server health is insufficient. Failed/partial/uncertain publication preserves inputs. Never delete openspec plans, source edits, selected evidence, user files or server-owned copies. Pause preserves inputs/checkpoint for resume. No broad folder sweeps.

### Current revision self-publication

This plan expansion is a real current-session planning occurrence. Until the new stage API is built, publish a clearly labeled text-only legacy entry in the existing walkthrough and retain the immutable R3 plan snapshot locally. Do not present a stub node as a committed API event or fabricate server IDs. T10 self-host verification publishes this actual planning revision with truthful source timestamps and server receipt timestamps, records returned IDs, and verifies it in the real timeline. This is labeled a later import: server registration/receipt time stays truthful, optional occurred_at carries only a locally recorded source date, and no unknown original start time is invented. This temporary compatibility bridge does not implement or alter the old schema.

T9 additive index wire settlement: /api/index keeps its legacy projects/tickets/sessions/excluded_sessions fields intact and adds timeline_sessions:[]TimelineAPIIndexEntry (empty array, never null). New clients merge these into the sidebar using explicit mode/source identity; API records are never coerced into legacy paths/fileBase. Legacy-only clients ignore the added field. The typed standalone index-entry union remains canonical for its member contracts.

T10 opener target wire: optional validated session_id/event_id query parameters travel with the existing #code one-use fragment. event_id requires session_id. These IDs are not credentials; keep the target query when removing the code. Explicit target wins over saved selection; missing/forbidden targets fail visibly without another-session fallback; archived targets may be shown temporarily without changing stored archive preference.

User width correction2026-10-31 during T9: keep w-full and centered mx-auto, increase the feed cap from max-w-lg512px to native max-w-2xl672px. No custom margins or nowrap forced on prose. Measured at1280:672px feed,center delta0, requested filter sentence has one line rectangle. At320:288px feed with neither page nor inner-canvas horizontal overflow. This changes only the container cap and stays in the current UI correction scope while backend work continues.

Producer integration correction: closing work does not delete its native conversation identity. Explicit user-requested new/resumed work may conditionally PATCH the existing closed session to active with a new operation ID/current version, preserving prior closure/history and the same viewer ID. Reads, get/create and appends never implicitly reopen it; deleted identity remains410. This realizes the existing public state PATCH contract and the prohibition on silent reopening, without a new command/schema.
