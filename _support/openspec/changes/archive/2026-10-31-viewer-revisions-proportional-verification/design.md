# Design

Superseded on 2026-10-31 by the viewer-timeline-api change. Historical only; do not execute this contract or checkpoint.
## Decisions

- User decisions: initial full affected-UI walk, focused iterative verification, mobile automatic for responsive impact, interactive delivery, native Codex /review, scoped viewer cleanup, current branch.
- Use an explicit schema v2 rather than ambiguous optional copies of current captures plus history. New readers keep v1 support without converting or guessing historical relationships. New producers require the new binary before content rollout. Older readers will reject v2; this is deliberate, visible incompatibility of new output with older binaries.
- Screen identity is explicit and scoped by the existing exact project/host/session tuple. Titles, ticket IDs, dates, closure status and filesystem paths never establish supersession. A continuation reuses the same screen ID only when it deliberately revises that screen.
- A round is a publication batch, not a viewport or a user decision. No per-viewport approval, automatic opening of new viewer tabs, or repeated whole-app navigation for a local visual correction.
- No new persistent database or automatic snapshot service. Source manifests carry bounded history and remain the evidence authority; the store copies all referenced resources as it does today.

## Proposed shape

### V2 source contract

Keep top-level lang/kind/title/ticket/session/candidate/created_at/result/checks/action_required/environment/provenance/delivery semantics. Set schema_version to 2. Add publication {number, published_at, summary}: number is a positive, strictly increasing integer within the logical session's new evidence stream. The main producer owns assigning it after inspecting that session's current manifest(s); children return evidence and do not allocate or publish rounds. Publication and revision numbers are integers in 1..2147483647. Every revision must be at most its enclosing publication number; a revision in that publication must use its exact published_at. Both validators reject impossible future revisions and inconsistent current-round timestamps. Equal publication numbers across documents require identical publication metadata, including summary.

Each flow has id, screen_id, title, revisions; remove the v1 current steps/captures/recording fields from v2 flows. id remains local navigation identity. screen_id uses the existing flow-ID syntax and is unique in a manifest. Each revision has number (its publication number), captured_candidate (branch/commit/base), published_at, summary, steps when applicable, and captures. Revisions are ordered by increasing number; the greatest number is current for that screen. Publication summary describes the whole round, including retained screens.

Each capture retains path/caption/alt/viewport and adds origin_publication, captured_at, captured_candidate and sha256 (64 lowercase hex, checked against bytes). origin_publication is an integer in 1..revision.number and stays unchanged when copied or reused. Derive New versus Retained from equality with the containing revision number, never from dates, hashes or path parsing. Use immutable images/r<publication-number>/<name> paths for newly captured bytes. Reusing a capture is an exact reference to the prior path/hash/date/candidate, not a new capture with a newer candidate. A mixed revision may reuse unaffected state captures and add new affected states. Store and display that distinction per capture. Replacing old bytes behind an existing hash is a validation failure. Retained history resources must still exist under the source folder; no cross-folder filesystem traversal is allowed.

One manifest remains authoritative for a work item; update it in place, append revisions only for changed screens, and keep unaffected revisions. For an explicitly related continuation in another work folder, the producer copies prior revisions and referenced immutable resources into the new folder (under safe relative paths) before appending the new revision, preserving provenance. It does not edit old manifests or point to their filesystem paths. Thus standalone HTML remains complete for that continuation.

The installed viewer projects all ready manifests in an exact logical session into a single screen list. Group only v2 screen_id values. Choose the highest revision number, deduplicate equal revision numbers only when normalized revision metadata and referenced content hashes agree (resource base URLs may differ); a conflict is a visible error, not a chronological tiebreak. Maintain the owning document's capability base separately from source JSON; never accept arbitrary resource URLs from the manifest. Legacy v1 segments remain visibly separate, with 'Revision information unavailable'. Different logical sessions/projects/hosts never collapse together. Session links/deliveries continue to aggregate through the existing mechanism.

The newest publication in the loaded logical session is the comparison round for updated/retained labels; highest equal publication numbers with incompatible metadata are conflicts. Until all required documents load, show loading or an error; do not label a screen current using a partial group. A screen is 'Updated in this round' when its current revision number matches that round; otherwise 'Retained from R<n>'. Do not infer correctness merely because a screen was retained. Current producer must have checked whether it was affected.

Limits: keep 1 MiB manifest, 50 screens/flows and 20 MiB per capture. Permit at most 20 retained revisions per screen and 100 capture entries total per screen across its revisions. Duplicate paths with identical metadata may be reused; paths with contradictory hashes/provenance fail. At the limit fail with a concrete retention instruction; never silently discard old evidence. A user-selected archival/retention decision is needed before pruning unique history. No recordings in newly produced v2; v1 compatibility remains.

### Publication confirmation

Extend --confirm-viewer to accept either explicit --session-status in_progress or closed. Expose stored_exact (the existing complete source fingerprint, including resources/status, which can include excluded copies) separately from presentable_current (accessible current evidence in a conflict-free session projection). Active confirmation requires both. Compute the latter from all relevant ingested group manifests using the same normalized projection invariants as the frontend; use shared Go/frontend cases. The target must not be excluded, its declared latest screen revisions must still be current, and unreadable/unknown/conflicting relevant members fail confirmation. This proves server-side presentability, not browser rendering, which requires the actual browser walk. Existing found/closed retention confirmation remains compatible, including exact excluded copies.

Reuse hive view status --dir with additive status facts. Keep exit 1 for validation/publication failure and exit 2 for successful publication with failed live confirmation. Bound polling; do not automatically start a viewer or switch to HTML because it is unavailable. Fallback without a binary publishes HTML and reports no live confirmation. Preserve the old closed-cleanup guarantee.

Producer handoff says 'updated in the viewer' only after successful live confirmation. Otherwise preserve evidence, report pending publication, and do not ask approval of a claimed visible revision. A stale confirmation or an unrelated document must fail. Existing authentication, same-origin checks and capability restrictions remain.

## UI

Surface: operational. Archetype: Record detail in the existing sidebar/detail shell. Container: existing Frame-led shell; no new nested Card. Density: dense, Button size sm. Regions: session delivery links, current round summary, one section per screen, desktop/mobile capture strip, collapsed History below that screen. Motion: no new entrance; use installed disclosure behavior and current capture dialog.

Follow application.tsx's current section headings and session shell, and walkthrough.mjs's existing viewport strip/image enlargement. In the React shell use installed ReUI Badge default size, info-light for updated, outline for retained, warning-light for unavailable revision data; Button outline/ghost size sm; Collapsible for History; Skeleton and destructive Alert with Retry for data states. The framework-free renderer uses its existing semantic button/details/dialog primitives and shared CSS tokens; do not install React/ReUI into standalone HTML. Match behavior/copy rather than introducing a second framework.

Primary action in each evidence region is opening a capture. History is a secondary disclosure; no new approve button or approval API. Retain original user approval in the conversation and identify the round in the question. Closed lifecycle does not mean accepted. Desktop/mobile switches remain in one screen; do not create a separate flow per viewport.

Before:

```text
CAM-4 — filters
[Desktop] [Mobile]
[capture strip]
CAM-4 — follow-up
[another capture strip]
```

After (session shell, 3 screens):

```text
Round 4 — compact mobile filter
Updated: Chat · Retained: Profile, invitation

CAM-4 — Chat                  [Updated · R4]
Filter shows only its icon on mobile
[Desktop] [Mobile]
[current captures; date and candidate]
[History · 3 revisions]

CAM-1 — Profile          [Retained from R1]
[Desktop] [Mobile]
[current captures]
```

After (expanded history and mixed capture provenance):

```text
CAM-4 — Chat                  [Updated · R4]
[Mobile]
Filter · R4 · 390px · dark      [New]
Header · R3 · 390px · dark      [Retained]
[- History · 3 revisions]
  R3 — search and filter in separate rows
  [Open revision R3]
  R2 — mobile search overlap corrected
  [Open revision R2]
  R1 — initial filter panel
  [Open revision R1]
```

Historical revision view is explicitly labeled 'Superseded by R4' with its own date, candidate, steps and viewport strip. It never replaces the current label. Keyboard operation and focus remain valid after live refresh; expansion state remains for unchanged screen IDs. Narrow headers wrap, controls stay reachable and the document never horizontally scrolls at 320/390px.

Loading: Skeleton matching heading/strip; no provisional current winner. Empty: 'No captures in this session' with Retry (reload index), not fabricated evidence. Error: inline Alert with Retry scoped to failed load; conflicts say 'Conflicting revision R<n>' and require republishing corrected metadata, not a retry loop that invents a winner. Permission denied: existing authenticated reopening flow; no new permissions or source paths in errors. Legacy: display original captures and a neutral revision-unavailable label.

## Copy

Ship Spanish and English: 'Actualizada en esta ronda' / 'Updated in this round'; 'Conservada de R<n>' / 'Retained from R<n>'; 'Historial · <n> revisiones' / 'History · <n> revisions'; 'Sustituida por R<n>' / 'Superseded by R<n>'; 'Nueva' / 'New'; 'Información de revisión no disponible' / 'Revision information unavailable'; 'Publicación pendiente de confirmar' / 'Publication confirmation pending'; 'Revisión R<n> en conflicto' / 'Conflicting revision R<n>'. Put candidate/time adjacent to capture details and full metadata in enlargement. Do not use color alone to encode state.

## Impact

Code: tooling/viewer/manifest.go, store resource enumeration, shared skill assets plus identical embedded copies, frontend application/controller/document/types/tests, tooling/cli/view_command.go only if status needs typed v2 exposure, Python validator/publisher/tests. Existing fingerprint matching already checks exact content, so avoid a second revision database or a new API if normalization and existing status suffice.

Guidance canonical home: flow-walkthrough for proportional capture/publication procedure; manifest reference for contract; flow-verify and hive-review-ux point there instead of retaining contradictory full-matrix iteration instructions. Check flow-build, flow-close, global guidance, all skills/roles, site flow pages, architecture docs and tests/pilot detectors for consumers. Global guidance should need no new invariant (target size delta zero); record actual before/after byte size and every checked surface in PR text. Do not change pilot fixtures or resume pilots.

Related #281 is context only: base/candidate comparison and unchanged-control evidence are different from retention across rounds. Current-phase retained means unchanged by this round, not proof of equality with the base.

## Recovery

Legacy v1 remains readable; no migration of source files. Roll back binary and content together; older binaries cannot read v2, so preserve its source evidence before coordinated rollback. No automatic rewriting/downgrade of v2. Existing store TTL and pin rules apply to complete source retention units.

D4 cleanup is irreversible for viewer-only copies and their per-session preferences. Preserve registry, authentication, server settings and tombstones. Repositories remain intact but unchanged sources deliberately do not reappear; genuinely changed/new evidence can appear through the normal scanner. Do not retain a recovery archive of historical captures the user explicitly asked to clear. New task evidence is outside the cleanup inventory.


### Atomic historical cleanup guard

The fixed planning-time allowlist is _support/workspace/2026-10-31-viewer-revisions-proportional-verification/cleanup-inventory.json: 32 copies in 27 groups, 29 archived copies, zero pinned/excluded copies. One ingest error is outside the deletion inventory. Do not expand the allowlist at build entry.

Add optional expected_members to existing authenticated session DELETE: an exact set of member keys/content hashes. Check it inside the same Store mutex that determines and deletes group membership, before tombstones or removal. Mismatch rejects the entire group without mutation (bounded HTTP 409). Validate types, unique keys and payload size; preserve authentication, same-origin requirements and existing unconditional explicit-user deletion callers. Cleanup uses only the conditional form. Deterministically test a new/changed member arriving between inventory and lock acquisition. Partial I/O failures after a successful precondition keep existing recovery. No bulk-delete UI, external producer locks or raw filesystem deletion.
