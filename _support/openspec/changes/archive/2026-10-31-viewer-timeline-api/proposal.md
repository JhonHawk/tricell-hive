# Viewer content timeline and session-bound publishing API

| Field | Value |
| --- | --- |
| Status | Completed — PR #304 integrated into development at0b6f2380; final candidate accepted D10 A |
| Tracker | User-requested, untracked; no ticket creation authorized. Issue #281 remains a separate base/candidate comparator. |
| Git | tricell-hive; current execution branch refactor/flow-ceremony, HEAD b5e8ad704159057d0f332ec053cec24076db15b7; integration base development, observed bb9604c56faadf60b92f7510d1b7206635ee1491 |
| Decisions | D2 A interactive delivery; D3 native Codex /review; D4 A historical viewer-copy cleanup only; current branch explicitly chosen. User accepted the timeline/API proposal by invoking flow-plan and required host/native session identity at creation. |

## Request

> "sera momento de rediseñar todo el visor, por lo menos la parte del contenido"

> "Las imagenes podrian mostrarse como miniaturas, y al dar clic en una se abre el carrusel"

> "un boton para ver las instrucciones y se abriria un modal con ellas"

> "La intención es tener menos info repetida"

> "para crear la sesión del visor pidamos el id de la sesión, el host, etc, para confirmar que sea unica"

The user selected the preceding proposal: reverse-chronological milestones, compact project/title/date/count header and action group, ticket-scoped view evidence, and a local API returning IDs. Future in-viewer approval/comments are a later delivery, not part of this candidate. This plan replaces viewer-revisions-proportional-verification; it does not execute its schema-v2/history-projection design. Criterion/task numbering continues rather than reusing AC1–AC8/T1–T6.

## Objective

Present a session as a compact evidence timeline. Let the agent publish bounded, typed events and images to the existing local server through Hive CLI, recovering the same server session from its verified project/declared host/native session ID. Separate storage IDs from native runtime identity and preserve explicit evidence-to-ticket relationships.

## Scope and acceptance

- AC9 — Creating or resuming the same resolved project/host/native-session tuple, including concurrent requests or restart, returns exactly one viewer session ID; incomplete identity is rejected. False on development: no API session creation exists and stored-copy keys derive from folder/native ID.
- AC10 — Retrying any completed publishing/mutation operation, including a deletion after restart, returns the same assigned IDs without duplicate content; reusing its operation key with different content fails. False on development: no publication operation receipts exist.
- AC11 — Stale event/session edits or withdrawals are rejected without changing the record; a permitted text edit preserves its timeline order, and changed published evidence requires a new iteration. False on development: there are no versioned event mutations.
- AC12 — An uploaded image becomes a session-owned immutable artifact; later reuse preserves its origin and another session cannot reference it. False on development: images are source paths copied/replaced as part of a manifest, not immutable API artifacts.
- AC13 — The content header offers inline title editing, one compact status/tickets/Git/archive/overflow group and a short scope summary, without the redundant session card/index/checks/environment/provenance sections. False on development: shell, deliveries and walkthrough renderer present those separate regions.
- AC14 — The timeline shows newest milestone first with server ordering and per-event date/time; editing copy does not promote it and an open PR does not close the session. False on development: chronological document segments have neither typed milestones nor server event sequence.
- AC15 — A validation can show one shared view for two tickets or separate view groups for different screens; its next iteration updates only the affected view and links to the prior evidence. False on development: tickets belong to document segments, not individual evidence views.
- AC16 — Each evidence view opens its own immutable image collection in a carousel and its immutable-evidence instructions in a Dialog; keyboard and mobile use preserve the event/view context. False on development: steps appear inline and there is no event/version-bound instruction dialog.
- AC17 — API-owned sessions survive scanner passes/restart without source-manifest overwrites; legacy inputs remain readable separately and publishing failure does not silently select another authority. False on development: all normal content comes from source manifests and replacement ingestion.
- AC18 — The publication procedure uses capability-based instructions with verified host-specific identity discovery when necessary, server receipts and proportional verification, with changed/retained evidence identified in each handoff. False on development: procedures publish complete manifests with synthesized-ID fallback and prescribe the full viewport/theme matrix without a proportional iteration rule.
- AC19 — Historical cleanup acts only when an atomic deletion precondition matches the fixed approved member-key/hash inventory; it suppresses unchanged reingestion and preserves new evidence, sources and settings. False on development: unconditional group deletion has no inventory precondition and the approved inventory remains untouched.

- AC20 — New and existing API/legacy viewer content does not expire, get purged or get skipped because of age, whether pinned or archived; restart preserves it. False on development: the store defaults to fourteen days, periodically purges and skips old manifests.
- AC21 — A substantive research/plan/build/verify stage can start before screenshots, display explicit in-progress/completed/paused/failed state and resume without duplicate nodes; status updates preserve original order. A later re-entry creates a linked new occurrence. False on development: manifests have no typed stage lifecycle.
- AC22 — A stage opens immutable Markdown document snapshots in a native tabbed Dialog; a plan revision remains separately identifiable, and an already-open historical snapshot cannot change underneath the reader. False on development: no stage document storage or viewer exists.
- AC23 — An offline CLI contract command returns versioned usage, schemas, limits and examples that agree with actual validators, without exposing credentials or starting the server. False on development: no contract command exists.
- AC24 — Each participating flow uses one short pointer to the canonical viewer publication skill, while retaining its own review/capture/approval policy; hygiene deletes only task-owned disposable upload inputs after exact committed document/image readback. Failed or uncertain publication preserves inputs. False on development: walkthrough-specific publication does not cover research/plan stages or document receipts.

Scope includes typed API/CLI, JSON persistence, required identity, header/timeline/evidence UI, legacy reading and existing offline fallback, updated flow documentation, and the already authorized bounded cleanup. Excludes human approval/comment endpoints/UI, feedback polling/SSE, automatic host wakeup, cloud hosting, a second daemon/database/framework, base-versus-candidate comparison, historical event reconstruction, CAM/Jira modifications, dependency upgrades and model pilots.

## Prerequisites

None external for synthetic local development. ReUI MCP returned the required block metadata; existing Base UI/React dependencies cover the selected registry components. At T8 fetch authorized registry source using the existing project ReUI connection without displaying license values; if unavailable, use the free Timeline primitive and installed components to implement the approved composition, not a different design. No license purchase or global installer update is authorized.

Real publication requires a user-started running viewer and an actual source session ID. The producer discovers facts from its runtime rather than asking the person each round; a missing ID prevents API-session creation and is reported, never replaced with a timestamp. The current session's confirmed ID is available for eventual self-publication. Native Codex /review availability is checked at its gate; do not silently replace it.

## Verified context

2026-10-31: reviewed the existing HTTP server/mutation auth, scanner and replacement storage, identity/group preferences, CLI commands, shell/controller/document renderer, manifest publisher and canonical walkthrough-viewer spec. Development lacks the new creation/event/artifact routes and timeline UI. Current branch includes other active guidance changes; preserve them and use its existing PR if present. No new branch/worktree.

ReUI MCP consulted Timeline API, timeline-3 block, c-button-group-5, Popover/Dropdown metadata; installed Button/ButtonGroup/Dialog/Badge/Frame implementations were read. Missing registry component files are an implementation prerequisite handled in T8, not permission to guess their API. The existing composition fits Record detail; no bespoke design pattern or theme deviation is proposed.

The motivating CAM Codex session is 01a19339-c0d8-7ca1-b7a7-ecc165466e7a, 2026-10-31: reused image names and separate follow-up walkthroughs obscured current evidence, and the user requested Finder to distinguish it. One observed session; user-approved adoption replaces a recurrence claim. The eventual guidance-fix commit names this session/failure. No measured time savings or model reliability claim.

## Delivery

Interactive D2 A on refactor/flow-ceremony. Native Codex /review, own browser walk and one independent UX reviewer. Show the first implemented layout at T8 and stop before push for final candidate validation. After acceptance, push/use the branch PR (or create it if absent) and merge to development only after required checks pass. Identify this work's changes separately from pre-existing branch work; reconcile any conflicting existing delivery stop before publication.

Maintainer marker is present. After integration refresh the binary before content; restart only an already-running viewer with its captured port/state directory and age retention disabled (old TTL recorded, not reactivated), then verify exact active revision and new API/UI. No master promotion/release. This planning invocation authorizes plan edits only; no implementation or cleanup occurs now.

## Review status and progress

2026-10-31 — Draft saved. D2/D3/D4 and branch decisions reused; no repeated delivery question. Existing historical inventory is retained as the cleanup allowlist. One independent contract review completed. Resolved four blocking gaps: immutable evidence_id separate from mutable node version; DELETE receipts/recovery in tombstones; durable project/root bindings before identity hashing; artifact-before-record commits with file/parent-directory sync and explicit uncertain outcomes. Reserved terminal-operation capacity. Ready for build approval; no implementation performed.

2026-10-31 — User correction: name host tools by capability in shared guidance; isolate and verify host-specific identity acquisition rather than assuming one environment variable or tool name. D3 native Codex /review and canonical host enum values remain specific session/contract data.

2026-10-31 — User invoked flow-build. Main reconciled branch b5e8ad70 and empty index; no open branch PR. T7 delegated within new contract/type/test files only; main retains integration. Existing everyday viewer PID44408/port4483/TTL336h remains untouched.

## R3 scope expansion — 2026-10-31

User approved R2 (D6 A) and requested no default retention, research/plan milestones with live declared stage state, tabbed Markdown plan snapshots, and minimal flow-skill publication pointers. D7 is superseded by this explicit refinement, not accepted as originally worded. Incorporate the CLI contract and confirmed-ingestion hygiene proposals without the fourteen-day policy or build-only session trigger. This is the same work item; current branch, interactive delivery and native Codex /review remain.

The publication owner remains flow-walkthrough (broaden its description, no new skill or rename). Review policy remains in flow-verify; capture policy has one conditional UI-verification reference owned there, with direct pointers from consumers. Research/plan/build/verify/close add only their relevant trigger and a direct pointer, no duplicated protocol. No universal hook, tracker polling, automatic startup or model pilot.

Additional scope: typed stage/document contracts, minimal Markdown/GFM renderer, read-only document tabs, removal of age-based retention in legacy and API modes, contract introspection and precise source-temporary hygiene. Future bulk-cleanup UI, authenticated feedback and automatic host execution remain excluded. Explicit deletion, approved historical cleanup and crash-owned incomplete staging cleanup remain supported.
