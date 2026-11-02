# Viewer revisions and proportional verification

| Field | Value |
| --- | --- |
| Status | Superseded — not built; archived 2026-11-01, replaced by ../2026-10-31-viewer-timeline-api/ |
| Tracker | Untracked user-requested change; no tracker write authorized. Related issue #281 remains separate. |
| Git | tricell-hive; integration base development; user-selected execution branch refactor/flow-ceremony; planning HEAD 1acc797547a1acc96153f00e76a31b31fd1b84ea; observed development bb9604c56faadf60b92f7510d1b7206635ee1491 |
| Decisions | 2026-10-31: proportional mobile verification accepted; D2 A interactive delivery; D3 native Codex /review; D4 A clear viewer copies and suppress unchanged reingestion, preserve project sources; user explicitly selected current branch. |

## Request

> "debia actualizar las imagenes en el visor, no lo hizo para todas las tareas, y cuando si lo hacia no distinguia las nuevas de las anteriores"

> "fue un poco rebuscado que hiciera varios recorridos para la versión mobile"

The user accepted an initial full walk of affected UI, followed by focused checks and captures of affected regions and states. Mobile remains automatic when responsive behavior is affected. Desktop/mobile belong to one walkthrough; iterations are revisions rather than parallel walkthroughs. The user then invoked flow-plan and selected the existing branch. These requests are the acceptance source; there is no ticket with additional criteria.

## Objective

Make the current evidence identifiable per screen and publication round, retain earlier revisions behind history, and confirm exact ingestion before claiming that a round is visible. Reduce repeated navigation and capture work without losing coverage of responsive changes.

## Scope and acceptance

- AC1 — A three-round, multi-screen fixture shows only each screen's current revision by default, marks updated versus retained screens for the current round, and exposes earlier revisions under History. False today: the manifest has neither screen revision history nor round-level freshness; captions alone carry this information.
- AC2 — Republishing a screen with changed captures preserves the earlier capture bytes and their candidate/date; retained captures preserve their original provenance and are labeled retained. False today: captures can be overwritten at the same path with no declared revision or capture hash.
- AC3 — Explicitly related screen revisions in two work folders of the same logical session appear as one screen with history; an equal title/ticket in another identity never supersedes it, and ambiguous/conflicting revisions produce an inline error rather than a guessed winner. False today: the application renders every folder as an independent chronological segment.
- AC4 — Active publication confirmation requires exact storage and accessible, conflict-free current screen revisions; excluded, stale, superseded or conflicting evidence cannot confirm a handoff. Closed cleanup keeps its existing exact-storage guarantee. False today: confirmation is restricted to closed and found includes excluded copies without projection validation.
- AC5 — A revisioned standalone HTML output shows the same current/history, provenance and desktop/mobile selection as the installed viewer for all history contained in its source manifest. False today: neither renderer exposes revision history.
- AC6 — Guidance gives one canonical proportional-verification procedure: initial coverage of affected layout at desktop/390/320 in supported themes; subsequent checks and recaptures only for affected regions, states, widths and theme-sensitive behavior, with full navigation only when its sequence changed. No added human question per viewport or recapture. False today: walkthrough step 2 and UX role prescribe the full width/theme matrix without an iteration exception, while correction handling only says affected screens.
- AC7 — Go and Python accept legacy v1 without rewriting it, accept the new versioned contract with identical validation outcomes, and reject malformed/conflicting history, unsafe resources and changed bytes behind retained capture hashes. False today: both validators reject schema_version other than 1 and there is no history/hash contract.
- AC8 — Cleanup deletes only groups whose member keys/hashes atomically match the authorized historical inventory, suppresses unchanged return and reports partial failures; mismatched/new groups and project sources/settings remain intact. False today: deletion lacks this atomic precondition and no inventory has been deleted. Authenticated read-only planning inventory: 32 copies/27 groups, 29 archived copies, zero pinned/excluded copies, one ingest error outside deletion scope.

Out of scope: implementing issue #281's before/base-versus-after comparator; Jira attachments or CAM changes; reconstructing old revision relations; public hosting; model pilots; new packages; global rule expansion; a new bulk-delete product UI; automatic acceptance inferred from closed status; changes to the existing TTL.

## Prerequisites

None external for the local candidate: use synthetic data and isolated viewer state. The live cleanup uses the fixed planning-time inventory and an atomic expected-members guard; never expand it at build entry or display access material. If membership cannot be resolved, skip destructive work and report the exact blocker. Native Codex /review must be available for the selected review gate; report unavailability rather than substituting another mechanism without a decision.

## Verified context

2026-10-31: inspected current manifest/publisher, Go validator/store/delete/status, shared renderer, React document/application/controller, installed base-nova components, and walkthrough-viewer spec. Compared relevant sources on development: revision/history fields and in-progress confirmation are absent there as well. Current branch has unrelated pre-existing changes relative to development; do not revert, rebase, duplicate or independently integrate them. User selected this branch, superseding the earlier suggestion to wait for its integration.

Observed CAM Codex session: 01a19339-c0d8-7ca1-b7a7-ecc165466e7a, 2026-10-31. It republished corrections under reused capture names, then retained two CAM-4 walkthrough segments without explicit supersession. The user asked to open Finder because filter captures were hard to distinguish. Preserve this failure attribution in the eventual guidance-fix commit. This is one observed session; the user's acceptance of the proposal and subsequent build approval authorize adoption without claiming recurrence or measured time savings.

## Delivery

D2 A: keep work on refactor/flow-ceremony; native Codex /review plus the required own browser walk and independent UX review. Stop before push for the user to validate the rendered candidate. Then push and use the branch's existing PR if one exists, otherwise open one, and merge to development only when its required checks pass. Preserve pre-existing branch scope; describe this work's commits/diff separately and never claim the rest of the branch was reviewed by this change's review. No QA/master promotion or release.

At build start inspect existing PR/delivery state read-only. If the existing branch has a conflicting delivery stop or another active owner, resolve that conflict before publishing rather than creating a second competing PR. Do not create another branch/worktree for this work.

When the maintainer marker exists, integrated viewer/content changes trigger the repository's declared refresh: binary first, content next, and restart an already-running viewer with its original port/state directory/TTL. Do not refresh uncommitted planning files. T6's cleanup is separately authorized by D4 A and occurs only after candidate validation, before generating this work's final retained walkthrough, so new evidence is not erased.

## Review status and progress

2026-10-31 — User requested a content redesign and accepted flow-plan for the timeline/API proposal, adding required project/host/native session identity. Superseded by ../viewer-timeline-api/. Retain this record as history and its fixed cleanup inventory as authorized evidence; AC1–AC8/T1–T6 are not the current build contract. No implementation occurred.

2026-10-31 — Research complete; D1's migration choice superseded by D4 A's scoped cleanup. Reading legacy v1 remains a low-cost compatibility measure, without historical conversion. One independent contract reviewer found three blocking gaps (stored versus presentable evidence, cleanup race, future-round revisions) and one ambiguity (capture origin); all were resolved in the saved design/tasks/delta. Implementation has not started.
