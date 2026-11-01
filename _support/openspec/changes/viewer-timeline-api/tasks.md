# Tasks

Execute on refactor/flow-ceremony, planning HEAD b5e8ad704159057d0f332ec053cec24076db15b7; integration development (observed bb9604c56faadf60b92f7510d1b7206635ee1491). Reconcile current state before edits, preserve existing branch changes and do not create a worktree. Main thread owns coupled contract/storage/UI work. One independent plan review now; one independent UX reviewer and native Codex /review at verification. Task/criterion numbering continues from the replaced plan.

### T7 — Set the session/event/artifact contracts

- [x] Add strict typed contracts, fixtures and API-version discriminants for required native identity, operations, evidence scopes, index modes and errors.

**Closes:** AC9, AC10, AC11, AC12, AC15, AC17

**Depends on:** None

**Locations:** tooling/viewer typed timeline contract and tests; frontend typed API/client; CLI input/output schema; source documentation reference.

**Execution:** Main thread; interfaces cross the same files/languages.

**Test approach:** TDD, synthetic fixtures.

**Changes:** API source tuple distinct from native runtime ID, required values, server-assigned IDs/order/version, safe ticket URLs, view/evidence references, operation receipts, source-mode union. No implementation of future feedback or manifest v2.

**Verification:** Unit cases reject missing project/host/native ID, unknown keys/kinds, duplicate JSON keys/trailing data, wrong reference scopes and sizes; evidence_id versus event.version semantics; positive fixtures cover two tickets/shared view, separate views, focused iteration, duplicate operations and legacy index entry. Prove chosen negative observations against inspected development sources; ordinary tests only.

### T8 — Build the content layout against typed stubs

- [x] Implement compact header, scope summary, reverse timeline, view thumbnails, carousel and instruction Dialog using T7 realistic typed fixtures.

**Closes:** AC13, AC14, AC15, AC16

**Depends on:** T7

**Locations:** tooling/viewer/frontend shell/application/controller, existing component imports, new timeline/evidence components and tests; missing ReUI registry source files only.

**Execution:** Main thread; keep session toolbar and content state together.

**Test approach:** Component tests and first rendered browser view.

**Changes:** Reuse timeline-3/c-button-group-5 composition and installed base-nova variants. Fetch/read actual missing components through ReUI MCP/authorized registry, inspect expected publisher/advisories before registry installs and preserve package versions. Avoid installing unused avatar/rating dependencies. Print the seven-line ReUI implementation plan before UI code. Remove duplicate presentation sections, not test execution or stored evidence. Use Popover links, DropdownMenu overflow and AlertDialog deletion. No human validation buttons.

**Verification:** pnpm --dir tooling/viewer/frontend check, lint, test; browser fixtures at 1280/390/320, both themes, long title, zero tickets/PRs, failed load/save, version conflict, keyboard and open-dialog refresh. ReUI usage/audit/self-check, at most two fix rounds. Show the first rendered view with flow-walkthrough before continuing backend integration; no actual CAM writes.

### T14 — Extend contracts for stage lifecycle, documents and introspection

- [x] Extend T7 Go/TS contracts and shared fixtures for stage state, foreground pointer, immutable document revisions, document approval targets, strict limits and CLI contract output.

**Closes:** AC20, AC21, AC22, AC23

**Depends on:** T7 (completed)

**Locations:** Existing timeline contracts/tests/shared fixtures, new self-describing contract definitions.

**Execution:** Settled-contract backend builder; main integrates; one independent task verifier.

**Test approach:** Exact Go/TS parity and schema/validator fixtures.

**Verification:** Start/complete/resume/failed/re-entry sequences; stale transitions; one foreground pointer; terminal closure handling; stable event order independent of reported occurred_at; foreground withdrawal conflict; awaiting-validation atomic pause/resume; immutable snapshots; foreign/missing document/decision references; UTF-8/byte/count bounds; offline contract examples parse with production validators; legacy fixtures remain valid. No API1 compatibility claim for unshipped draft records.

### T15 — Render stage progress and document Dialog against typed stubs

- [x] Extend approved R2 composition with declared stage indicators and read-only native Tabs/Markdown snapshot Dialog.

**Closes:** AC21, AC22

**Depends on:** T14

**Locations:** Timeline components/copy/dev fixtures/tests; native Tabs/Select source if missing; Markdown dependencies/lock; generated bundle.

**Execution:** Frontend builder; main checks rendered result.

**Verification:** Typed current plan revision fixture, draft/completed/paused/error states, snapshot isolation after refresh, historic selection, keyboard/focus/Tabs, headings/lists/tables/code and safe links, hostile raw HTML/URLs and remote images remain inert. Check expected publisher/advisories before installs. Focused desktop/narrow layout checks only; reuse R2 evidence outside changed regions. Independent UX and show only new stage/Dialog composition before T9.

### T9 — Implement server-owned publishing and persistence

- [x] Add authenticated creation/reads/versioned mutation, immutable image upload, operation receipts, restart recovery, API/legacy separation and conditional cleanup guard.

**Closes:** AC9, AC10, AC11, AC12, AC14, AC15, AC17, AC19, AC20, AC21, AC22

**Depends on:** T8 (approved D6 A) and T15 (new stage/document UI before backend)

**Locations:** tooling/viewer server/store/index/scan/session deletion and new timeline persistence/API files with focused tests.

**Execution:** Main thread.

**Test approach:** TDD and deterministic concurrency/crash tests.

**Changes:** Reuse loopback/auth/capability/mutex/atomic-write mechanisms. Get-or-create tuple uniqueness across concurrent requests/restart, persisted request digests, conditional edit/withdraw/delete, immutable artifact ownership and evidence identity, durable root/project bindings, bounded input/storage with reserved terminal capacity. Extend only API writer durability to file+directory sync with artifact-before-record ordering and explicit uncertain-outcome recovery; do not assume legacy atomic rename is a power-loss guarantee. Keep DELETE receipts/recovery in tombstones and replay before version/lifecycle checks. Scanner cannot replace API tuples; new source mode does not inherit fake legacy history. Add expected_members to existing legacy DELETE and check it atomically. No paid services or new DB.

**Verification:** go test ./tooling/viewer, go test -race ./tooling/viewer; lost-response replay across restart, same operation/different body conflict, concurrent tuple creation, temporary unreadable remote with API-only stored bindings after restart, equivalent worktree root, conflicting authoritative project identity, stale patch, event copy edit preserves sequence and evidence_id, referenced original instructions/images resolve after the edit and restart, new iteration preserves old evidence, cross-session refs denied, spoofed actor fields rejected, closed/deleted session append denied, identity tombstone and lost-response DELETE replay, changed DELETE body conflict, unknown operation on deleted identity, replay before stale-version validation, quota-full lifecycle/deletion, truncated/corrupt records, interrupted artifact upload, failure injection before/after artifact fsync/rename/parent-sync and session fsync/rename/parent-sync, outcome_unknown reconciliation without duplicate mutation, failed partial deletion, scanner pass over equivalent tuple, pin/archive/no-expiry, wrong token/origin/host, unsafe upload/URL and oversized payload. Test legacy cleanup membership changing after inventory but before deletion lock; nothing is removed on mismatch.

### T10 — Connect the agent CLI and real UI

- [x] Add machine-readable session/event/artifact commands, optional open target, live typed UI integration and publication receipts without token exposure.

**Closes:** AC9, AC10, AC11, AC12, AC13, AC14, AC15, AC16, AC17, AC21, AC22, AC23

**Depends on:** T14/T15 for independent client code; T9 for real integration and verification

**Locations:** tooling/cli view command/client/help/tests; frontend API/controller/timeline integration and generated bundle.

**Execution:** Main thread.

**Test approach:** CLI/API integration on isolated temporary state.

**Changes:** JSON via file/stdin, local image bytes upload, precise committed IDs/version/sequence, nonsecret session/event deep link carried through one-use opener. Source identity remains required at session registration; a resumed existing receipt must be checked against the native tuple. No credentials in args/output, no automatic viewer startup, no curl snippets for the agent. Missing-binary legacy static HTML remains separate. Reads do not allocate a new session.

**Verification:** go test ./tooling/cli; isolated running viewer integration creates request/work_started/two-ticket validation/focused iteration/reported decision/delivery/closure, retries lost operations, restarts server, and reopens exact session/event. Inspect persisted content and actual browser presentation. Stopped/unavailable server returns error with source images preserved and no fallback/session duplication. Legacy render/status tests stay passing; old plain open and title/archive/pin/delete behaviors work in their mode.

### T11 — Replace the publication procedure and document proportional checks

- [x] Update walkthrough authoring/close handoff to the CLI/API and native identity, keep legacy fallback explicit, and align all dependent guidance/docs.

**Closes:** AC18, AC24

**Depends on:** T10

**Locations:** content/skills/flow-walkthrough/SKILL.md plus CLI contract pointer; verify-owned conditional UI capture reference; minimal pointers in flow-research/plan/build/verify/close, hive-review-ux; site flow pages, README, llms, architecture/deployment documentation and affected ordinary detectors.

**Execution:** Main thread.

**Test approach:** Existing content/link/size tests and static scenario validation; no model pilot.

**Changes:** Shared guidance names capabilities (native session identity, terminal/process, files, browser/capture, native questions), not host tool names. Keep one conditional host-identity reference; verify official documentation and installed versions before recording exceptions, separating documentation, observation and inference. Treat the supplied cross-host screenshot as a research lead, not verified availability. Missing verified identity stops registration without inventing a fallback; independent work continues. Preserve the selected native Codex /review delivery mechanism. Main producer obtains runtime ID/host, registers required source, retains returned viewer/event/view/artifact IDs in task checkpoint, uploads new images, appends meaningful milestones only, verifies committed readback before a visible-update claim. First affected-layout verification covers relevant desktop/390/320 and supported themes; follow-up checks/captures cover only changed regions, affected neighbors, states/widths and theme-sensitive cases. Full mobile sequence only for navigation changes. Identity/data gaps are reported rather than invented. Existing native human question tools remain the approval channel in this release. No duplicate proportional rules across skills/roles. Guidance-fix commit names the observed CAM session/failure.

**Verification:** go test ./tests/content ./tooling/management; Python legacy walkthrough tests; inspect all global/skill/role/site/architecture/pilot references and record exact checked places/global byte delta in PR. Test static examples for mobile filter-width correction, theme color change, list/thread navigation, desktop-only change and missing native session ID, hosts exposing identity through different documented capabilities, and an unverified/unavailable identity mechanism. Reject any common instruction that assumes one named tool or variable works on all hosts. No new tests/fixtures/regression cases or model runs. Keep each distributed skill/reference <=16000 bytes.

### T12 — Verify and validate the full candidate

- [ ] Complete acceptance evidence, full existing suites, native Codex /review, own browser walk, independent UX review and interactive user validation.

**Closes:** AC9–AC18, AC20–AC24

**Depends on:** T11

**Locations:** Candidate code, this plan's evidence/checkpoint, isolated server state.

**Execution:** Main thread; one read-only hive-review-ux child; code mechanism native Codex /review.

**Test approach:** Full suites and end-to-end isolated workflow.

**Verification:** Inspect actual workflows and run their applicable local jobs, including go test ./..., go vet ./..., focused race tests, Python legacy tests, frontend check/lint/test/build/verify:dist and site build for changed docs. Own browser walk uses production-style candidate assets, two tickets and two screens; verify source identity uniqueness, chronological order, focused iteration, instructions/carousel context, stale mutation and offline/legacy states. Initial 1280/390/320 both themes; corrections reuse unaffected evidence. UX measures/interacts without duplicating captures unnecessarily. Native code review covers this work's diff, not unrelated branch history. Show exact final candidate and stop before push. No global install or historical cleanup during candidate verification.

### T13 — Integrate, refresh and perform the authorized cleanup

- [ ] After user validation, deliver through the selected branch PR, refresh the maintainer installation and clear only the fixed historical inventory through guarded supported deletion.

**Closes:** AC19

**Depends on:** T12 and interactive acceptance

**Locations:** Current branch/PR; maintainer installation/viewer own store; previously approved ignored cleanup inventory.

**Execution:** Main thread only.

**Test approach:** Exact live identity/health/readback and bounded cleanup observation.

**Changes:** Inspect current branch PR and prior delivery stops; do not create a competing branch/PR. Push then merge development only with required green checks. Capture viewer port/state/prior TTL before maintainer refresh; rebuild binary before content and restart only the already-active viewer per repository rules, preserving port/state with age retention disabled. Keep registry/auth/settings/sources. Use inventory _support/workspace/2026-10-31-viewer-revisions-proportional-verification/cleanup-inventory.json only, never expand it; 32 copies/27 groups including 29 archived, no pinned/excluded at capture. Conditional delete expected member hashes uses T9 guard; mismatch preserves the entire group. Preserve new/API sessions, source files and the existing unidentified ingest-error source. Partial failures retain recovery records and exact remaining counts. No raw directory deletion or new bulk-delete UI.

**Verification:** Running build/health reports exact integrated revision and original service settings; create/read an isolated self-test session through installed CLI only with task-owned evidence, verify no duplicate on retry, then delete that self-test only. Before/after historical counts reconcile against inventory; scanner/restart does not resurrect unchanged targeted content; later evidence and original source files survive. Confirm current work's API evidence is retained before source temporaries are cleaned. Report deployed revision, active runtime, cleanup counts/remaining groups; no master/release. Cleanup failure does not erase a verified code delivery or claim full completion.

## Verification and human review

| Check | When | Evidence | Authorization |
| --- | --- | --- | --- |
| Contract/unit/concurrency | T7/T9/T10 | Focused local fixtures and failure injection | Build approval |
| First layout | T8 | Typed-stub rendered header/timeline/views/dialog | Build approval, flow-build |
| Full suite | T12 | Actual CI-local jobs above | Build approval |
| Own walk and independent UX | T12 | Production-style isolated viewer, desktop/mobile and themes | Required UI verification |
| Native Codex /review | T12, before final validation | Findings and bounded fixes | D3 |
| Human acceptance | Before push | Exact session/event/candidate shown | D2 A |
| Historical cleanup | T13 after new runtime supports guard | Fixed inventory and exact remaining counts | D4 A |
| Refresh | After development integration | Exact active revision/service | Maintainer marker rule |

No real account or third-party data is needed. The synthetic walkthrough includes a shared Chat view for two tickets, separate Profile/Chat views, a later Chat-only iteration, instructions and carousel open during a refresh, title save/conflict, delivery open versus merged, and paused/closed states. Missing prerequisites or review tools are disclosed at their actual gate without substituting authorization.

## Review status and progress

2026-10-31 — Planning only; T7–T13 pending. One independent contract review completed and main resolved evidence-version references, deleted receipts, project-key fallback and durable commit ordering plus terminal capacity. Checkpoint flow-plan step 6: ready handoff. New change supersedes prior draft; native identity is mandatory. No application code, runtime, tracker, Git delivery or cleanup changed.

2026-10-31 — Planning correction applied to T11/AC18: capability-first shared procedure and conditional, source-verified host discovery. No contract shape change or second plan review; no application edits or host experiments.

Build checkpoint — flow-build step2, T7 in progress. Starting HEAD b5e8ad70, pre-existing index empty. Native task-list tool unavailable; this file governs tasks. T7 contract builder owns only new Go/TS contract and tests, no server/UI integration. T8 pending contract verification, T9–T13 pending. No commits or global/runtime mutations yet.

T8 browser observations: existing viewer surface is dark-only (styles.css fixes background/color-scheme on :root), so theme-class toggles are not evidence of a light viewer theme. Initial viewport observations cover1280/390/320 on the actual supported dark surface; light/dark captures inside nodes are synthetic evidence, not viewer theme support. Fixed native figure/heading margins and thumbnail full width. Axe observed status info-light contrast2.96:1; switched only that new status control to installed outline variant and added a proper labeled timeline region/heading. No global theme changes.

✓ T7 — contracts verified independently after fixes; commits d9ebab77 and0ae93535 (all nine shared fixture files retained).63 shared Go/TS cases,83 TS contract tests; unknown-key casing, private receipts/public results, legacy index union and optional-text parity corrected. This verifies contracts only, not T9 API/persistence/restart behavior.
T8 implementation ready against typed stubs; viewport checks1280/390/320 supported dark surface, no horizontal overflow/missing images; instructions/carousel/popovers and5 UI tests pass. Current header uses accessible outline state variant. First-view artifacts under _support/workspace/2026-10-31-viewer-timeline-api/. Task-owned Vite session65902/port5192; browser session hive-timeline-build-01a19479. Next: final local T8 checks/commit, publish R1 and ask D5 before T9. Everyday viewer PID44408/4483 remains unchanged.

✓ T8 — local commit cbc92592, typed stub composition complete.142 frontend tests +3 bundle tests,TypeScript/Biome,production build/verify:dist and go test ./tooling/viewer pass. ReUI audit/self-checks have no application-code hits; axe has0 confirmed violations with offscreen contrast checks still incomplete. No supported light surface is claimed. R1 uses six current captures (1280/390/320, instructions, carousel, tickets) and exact native ID verified against the current runtime.
Checkpoint — flow-build step4/flow-walkthrough, D5 first rendered composition pending. HEAD cbc92592; T7/T8 done, T9–T13 pending. No push, real API, installation refresh or historical cleanup. Keep Vite process65902/port5192 for review; everyday viewer PID44408/4483 unchanged. Next after D5 A: T9 server/persistence then T10 CLI/UI integration. D5 B: adjust only requested UI and republish affected screens.

D5 B — user correction2026-10-31: header actions must align far right; portal dropdowns need repair; instructions need greater width and readable step formatting; node icons must match reference scale; timeline-4 is primary composition and timeline-3 guides actions/status; feed must be centered in viewer content container. Reopen T8 for R2 only, no T9/backend. Main owns portal CSS resets and rendering evidence; component builder owns its prior header/content/copy/tests files. R1 capture paths preserved; R2 will replace manifest references using distinct paths.

T8 correction refinement2026-10-31: user requires native ReUI shells and components with minimal custom spacing. Audit composition against installed defaults, remove unnecessary overrides, preserve explicit centering/right alignment/readability/reference marker requirements. No expansion of backend scope.

✓ T8 R2 — local commit4320d016; native ReUI defaults restored, central scoped baseline fixed, Separator attribute wiring repaired, native dialog footer used. TypeScript/Biome,142 frontend+3 bundle tests,focused5 interaction tests,build/verify:dist,Go viewer tests and gitleaks pass. Independent UX R2 passes after H1 menu clipping fix; contrast remains partly incomplete. Distinct R2 captures replace manifest references; R1 files preserved.
Checkpoint — flow-build step4/flow-walkthrough: D6 composition validation pending. T9–T13 remain pending; no push/runtime restart/historical cleanup. Keep Vite5192 for review.

## R3 revised execution checkpoint — 2026-10-31

D6 A accepted by "Me gusta". User then requested plan expansion; no T9 implementation has begun. Completed T7/T8 remain valid for their old scope; T14/T15 add the new contract/UI before backend. T9 additionally implements document durability/readback, stage transactions and complete removal of automatic age purge/scan filters, including deprecated TTL no-op compatibility. Verify old legacy copies and API sessions survive simulated years, restart and archive without a pin; explicit guarded deletion still works. T10 adds offline hive view contract [--operation], stage/document commands and real self-publication of this plan revision without fabricated IDs. T11 broadens publication owner flow-walkthrough, uses short pointers in research/plan/build/verify/close, places review/capture policies canonically in verify, and reconciles workspace retention and source-input hygiene. T12 covers Markdown rendering security, all stage/document criteria and no-expiry legacy regression. T13 preserves service port/state/registry but records effective retention manual rather than reviving old TTL. Existing fixed historical cleanup allowlist is unchanged.

Checkpoint: flow-plan revision R3; T14/T15/T9–T13 pending. Only plan/support publication edits authorized in this turn. Last application commit4320d016; no push/restart/package install. Plan reviewer checks this expansion before handoff. Next implementation entry is T14, not T9.

R3 independent review resolved H1 (explicit document/snapshot API and stage version preconditions, foreground withdrawal and human-wait behavior), H2 (single verify-owned capture/review policy; walkthrough only publication), H3 (server times separated from optional agent-reported occurred_at; current-plan upload labeled later import). Review ran read-only; main reconciled edits made concurrently with its read. Ready plan handoff; no application implementation in this revision turn.

Build resumed on explicit user command2026-10-31: R3 criteria approved. flow-build step2/T14 in progress, last application commit4320d016, pre-existing index empty; only retained support/plan folders untracked. Contract builder owns existing Go/TS contract+fixtures and new protocol definitions, frontend builder performs read-only T15 dependency/component readiness until interface stabilization. Main owns integration/checkpoint. T15 and T9–T13 pending; no runtime restart, cleanup or publication.

T14 minor contract settlement: close is sequential and explicit—reject while a foreground stage is active; complete/pause it first, then close with current session version. No extra finalization envelope. Stage/document public shapes settled with builder; TS stabilization pending before T15 edits.

✓ T14 — local commit cbacddfd; independent contract verifier passes after UTF8 BOM parity fix.104 shared payload cases plus a dedicated shared BOM fixture,25 transition cases,152 TS contract tests. Go contract/protocol tests,TypeScript and Biome pass. Description only: actual no-expiry/persistence/atomic transitions/CLI remain T9/T10.
✓ T15 — stage progress + native Tabs/Select/Spinner document Dialog implemented.216 frontend tests+3 bundle tests,Go viewer package,build/verify:dist,gitleaks and native ReUI checks pass. Read-only independent UX passes1280/390/320, no page overflow; axe0 confirmed violations/1 incomplete check. Exact reviewed Markdown dependencies and103 additional license notices included; prod audit clean. Full audit has8 pre-existing undici8.11.2 dev advisories (4high/4moderate), outside renderer graph; not remediated. Semantic GFM table is the one justified C4 hit.
Checkpoint — flow-build step4/flow-walkthrough: T14/T15 complete locally; D8 stage/document first-view validation pending. T9–T13 not started. Preserve Vite5192 and current viewer44408/4483 (old TTL unchanged until integration). No push/global refresh/cleanup. Next after D8 A: implement T9 backend against settled contract.

T15 commit13ecd180. R3 first-view manifest reuses same source/session and marks R2 captures retained; seven distinct r3-* paths replace only changed evidence. Publication/readback pending below; native typed stage self-publication still T10.

R3 exact-source viewer readback found:true/session_status:in_progress; existing ingest_errors:1 unchanged. T14/T15 working tree committed13ecd180 and index empty; only previously retained support/plan folders untracked. D8 pending; Vite/runtime kept for human review.

D8 B — user accepted general composition, requested native semantic badge colors and removal of thumbnail bottom strip. Fixed in local commit6fa74f3f: warning-light/success-light/outline/destructive-light for stage states; img display:block removes5px baseline gap. Narrow UX verification passes contrast8.11/6.57/7.25/5.79 and81px image+2px border, no extra strip.10 interaction tests,TypeScript/Biome,build/verify:dist,Go bundle test,gitleaks pass. R4 changed-feed captures replace R3 feed; unchanged R3 document evidence retained explicitly. Checkpoint step4: D9 confirmation pending; T9 remains unstarted.

D9 A accepted explicitly. Build checkpoint step2/T9 in progress at6fa74f3f. Settled backend implementation delegated to build_t9_backend in tooling/viewer Go source/tests and CLI TTL-only compatibility; main retains integration/commits and independent verifier. Current everyday server remains44408/4483,33 copies/1pre-existing ingest error; no global write/restart. T10–T13 pending.

T9 pre-write inventory verified read-only at default local state:33 physical legacy copy directories, no timeline namespace directories, same1 pre-existing ingest error. No migration of historical events; new API namespace starts empty. Additive index wire shape documented in design.

T10 implementation ordering refinement: frozen T14/T15 wire permits independent frontend/client code while T9 persistence is built; actual integration/acceptance still waits T9. Main owns NEW CLI timeline client files only until backend releases existing TTL CLI files; frontend builder owns api/controller/application/main and corresponding frontend tests. No duplicated controller or polling loop. Canonical JSON-input flag follows compiled protocol --json FILE|- (replace old plan --input wording). Nonsecret opener targets use session_id/event_id query plus unchanged one-use code fragment.

T9 concurrent UI correction requested by user: timeline-content.tsx cap max-w-lg→max-w-2xl. Main measured672px centered feed/reference sentence one line at1280;320 feed288px/no page or inner overflow. R5 capture distinct from R4; remains uncommitted with ongoing T10 client work until bundle regenerated. Backend continues; no new scope or gate.

✓ T9 backend verified independently; fixed H1 retry of directory/ancestor fsync and H2 private4MiB decode vs public2MiB limit. Full viewer race passed; targeted fault/capacity reverification passed, no remaining backend blockers. CLI TTL compatibility/provider resolution delivered with T10 client commit. No livecleanup/defaultserverrestart.
T10 integration in progress: frontend226tests+3bundle pass; main realCLI isolatedserver tests coverdocuments/BOM,stage transitions,openertarget,restart/delete replay,two-ticket separateviews,focusediteration,copyeditimmutability,reporteddecision,openPR versusclosure. Productioncandidate runs taskowned60643/state under native temp; actual current session+R3doc snapshot registered,IDs in candidate-runtime.json. Defaultserver44408/4483 unchanged. T11 documentation worker editing canonicalcontent; main owns remaining tests/finalintegration.

T11 main consumer reconciliation: flow-report routes viewer/fallback output to the publication owner; unattended-delegation owns its existing first-view-question exception; legacy retention keeps explicit delete-at-session-end semantics after confirmation; research summaries do not create unauthorized retained reports. Capture scenarios manually checked: mobilefilterwidth→affectedregion/neighbors/widths only; themecolor→affectedthemes; navigation→fullaffectedmobilejourney; desktoponly→no irrelevantmobilematrix; missing/unverifiednativeID→no API publication/no fabrication/independentworkcontinues. Wholeguidance/site/architecture/role/pilot-detector consumers inspected; global.md byte delta0. Content/management/pilot statictests and64 Python tests pass. Sitecheck/build pass with existingi18n/404 warnings.

✓ T10 — local commit f3d779e0; independent verifier passes CLI/client lifecycle and real isolated-server integration. T11 — local commit9f8db1df; canonical publication owner and proportional verification policy reconciled across dependent flows, roles, site and architecture; global.md unchanged19417bytes (delta0). Full Go suite/vet, frontend226+3tests,64Python checks,sitecheck/build pass.
D9 A and R5 width accepted by user. T12 in progress: native review of b5e8ad70..9f8db1df found P2 deletion recovery and carousel failure-state reset; both confirmed and assigned to one fix round. Main found missing artifact enums in offline schema; fixing schema/validator shared values with regression. Full race suite running. Real candidate9f8db1df at60643 preserves session identity through restart; actual plan/artifacts read back and exact hashes/sizes match. Independent final UX confirms new interactions and672px centered/nooverflow390/320; pre-existing sidebar landmark issue outside changed layout remains nonblocking, partial contrast inconclusive. No push/default-runtime refresh/cleanup.

T12 final checkpoint5af2416f: full Go race PASS (CLI336s), native review9f8db1df findings fixed ina378cb9e; bounded second review found terminal deletion recovery, fixed5af2416f with regression and required Unreviewed line.233frontend+3bundle,TS/Biome/build/dist and focused Go bundle/protocol/race/vet pass. Main actual-browser image404→valid recovery and abortedDELETE→Reintentar visible; no session deleted. Final real API evidence stored/readback; state awaiting_validation. D10 final acceptance pending beforepush; no third review. T13 pending. Concurrent externalcommitbc3fdcc7 preserved: global.md now19634bytes (+217frombaseline), viewerchange itself0; includes separatecheckpoint-guidance change not authored here. Everydayruntime44408/4483 unchanged; candidate39210/60643 active.
