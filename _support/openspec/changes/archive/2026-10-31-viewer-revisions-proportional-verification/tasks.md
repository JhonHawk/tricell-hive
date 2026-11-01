# Tasks

Superseded on 2026-10-31 by the viewer-timeline-api change. Historical only; do not execute this contract or checkpoint.
Planning HEAD: 1acc797547a1acc96153f00e76a31b31fd1b84ea on user-selected refactor/flow-ceremony. Integration base: development (observed bb9604c56faadf60b92f7510d1b7206635ee1491). Reconcile changed HEAD and existing branch PR at build entry; preserve other work and its delivery decisions. No new branch or worktree. Execute coupled implementation in the main thread; one plan reviewer now, one UX reviewer at verification, native Codex /review for code.

### T1 — Define the versioned revision contract

- [ ] Add typed v2 fixture/type shape, matching strict Go/Python validators, resource enumeration and negative cases; retain unchanged v1 coverage.

**Closes:** AC2, AC7

**Depends on:** None

**Locations:** tooling/viewer/manifest.go and tests; frontend typed data normalization; content/skills/flow-walkthrough/scripts/render_walkthrough.py; tests/skills/walkthrough_render_test.py; manifest reference.

**Execution:** Main thread (shared contract across languages).

**Test approach:** TDD with synthetic fixtures.

**Changes:** V2 publications/screen revisions/capture hashes/provenance, bounded history, exact resource retention; no legacy backfill.

**Verification:** Go/Python shared cases agree for v1/v2, first view/final, missing resource, mutated hash, duplicate/conflicting revision, out-of-order history, future revisions, noninteger/out-of-range counters, inconsistent publication timestamps, invalid capture origin, unsafe/symlink path, limit exhaustion and retained capture. Run go test ./tooling/viewer and python3 -m unittest discover -s tests/skills -p 'walkthrough*_test.py'.

### T2 — Render current evidence and history with typed stub data

- [ ] Implement one current section per screen and explicit retained/new provenance, history, conflict/loading/error states in installed and fallback surfaces using T1 fixtures.

**Closes:** AC1, AC3, AC5

**Depends on:** T1

**Locations:** tooling/viewer/frontend/src/{application,controller,document}.tsx or .ts as applicable, existing component imports/tests; content/skills/flow-walkthrough/assets/{walkthrough.mjs,walkthrough.css}; tooling/viewer/static identical copies.

**Execution:** Main thread, keeping the shared renderer and session projection coupled.

**Test approach:** Component/renderer characterization plus browser observation.

**Changes:** Group explicit screen IDs only within exact logical session, owner-specific resource resolution, current first, collapsed history; existing viewport and capture enlargement behavior. Match approved design, no new package or framework in fallback. Print the ReUI seven-line implementation plan before UI code.

**Verification:** Frontend tests for three rounds/three screens, same title different session, cross-folder duplicate agreement/conflict, failed document, keyboard/history focus and refresh. Browser desktop/390/320, light/dark, long labels, same-folder HTML without hive; no duplicate viewport walks, old/current labels always clear. ReUI self-check at most two fix rounds. Show first rendered view under flow-build before backend integration, since later tasks remain.

### T3 — Integrate ingestion and exact active-publication confirmation

- [ ] Connect revision ingestion, active presentability confirmation and atomic conditional deletion, preserving closed-copy retention confirmation.

**Closes:** AC2, AC3, AC4, AC5, AC7, AC8

**Depends on:** T2

**Locations:** tooling/viewer store/scan/projection/session_delete/session API and tests, tooling/cli/view_command.go status, publisher confirmation and tests/skills/walkthrough_status_test.py.

**Execution:** Main thread.

**Test approach:** TDD for confirmation/error states; isolated live viewer integration.

**Changes:** Reuse existing source fingerprint/status/capability boundaries, hash every referenced history resource; no second history store or invented archive service. Add stored_exact/presentable_current using shared Go/frontend projection fixtures: active confirmation requires both; closed retention keeps its existing exact-storage rule. Add optional expected_members to session DELETE, compare exact keys/hashes under the same Store lock before mutation and reject mismatch with bounded 409. Preserve ordinary explicit-user deletion behavior. Exit 2 remains publication-success/confirmation-failure. Do not start the everyday viewer or change runtime configuration during local tests.

**Verification:** In temporary registered fixtures and isolated server state, publish R1/R2/R3, confirm current in_progress, replace an image without matching metadata and reject it, confirm closed, reject stale fingerprints and nonmatching resources/status, reject excluded active sources, superseded target revisions, cross-source conflicts and unreadable group documents; test offline fallback and unavailable daemon. Deterministically add/change a member after inventory but before deletion lock: precondition rejects the whole group without tombstones or removal. Test matching deletion, malformed/oversized payload, denied auth/origin and existing partial-failure recovery. Existing auth/same-origin/retention tests remain passing. Stop only test-owned processes. No model pilot.

### T4 — Publish the proportional verification procedure and documentation

- [ ] Align the walkthrough procedure and all consumers with initial coverage followed by focused corrections; describe revision publishing and exact confirmation.

**Closes:** AC6

**Depends on:** T3

**Locations:** flow-walkthrough SKILL.md/reference; flow-verify and hive-review-ux only to replace competing copies with pointers; applicable site flow pages; _support/docs/architecture/deployment-manager.md and instruction-resource references when their actual descriptions change. Read README/llms and update only if a stated user-facing command, requirement or structure changes.

**Execution:** Main thread.

**Test approach:** Existing content/link/size checks and static scenario review, not model pilots.

**Changes:** Initial affected layout at desktop/390/320 and supported themes. Later review only changed regions plus displaced/resized/reflowed neighbors, relevant states and widths. Theme-neutral geometry can reuse earlier opposite-theme evidence after checking the change has no theme-dependent impact; color/theme-sensitive corrections check affected themes. Responsive/navigation change checks relevant mobile sequence; nonresponsive changes do not automatically rerun mobile. UX reviewer consumes this scope and existing captures, measures/interacts independently and avoids duplicating captures without need. Each handoff identifies changed versus retained screens and round; validation applies to that round and changes reopen affected-screen acceptance only. No extra viewport question. Record CAM session failure attribution in the guidance-fix commit.

**Verification:** Review example scopes (filter width, theme color, mobile list/thread navigation, desktop-only text) against the canonical procedure; go test ./tests/content ./tooling/management and go vet ./...; inspect all instruction/site/pilot consumers without changing synthetic regression fixtures. Record global.md byte delta and checked locations for PR.

### T5 — Verify the whole candidate and obtain interactive validation

- [ ] Complete own browser walk, independent UX review, native Codex /review, and user validation of current/history/mobile presentation before push.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6, AC7

**Depends on:** T4

**Locations:** Candidate and this change's evidence/checkpoint; no application scope expansion.

**Execution:** Main thread; one hive-review-ux child read-only; code review through native Codex /review.

**Test approach:** Full existing suites and isolated live verification.

**Verification:** go test ./...; go vet ./...; Python walkthrough tests; pnpm --dir tooling/viewer/frontend check, lint, test, build, verify:dist; site build if pages change. Inspect actual workflows at execution and add any required local jobs not listed here. Validate latest/history, retained provenance, cross-folder ownership, conflict and failed-publication UI, legacy, HTML fallback, update while history/capture dialog is open, keyboard and 320/390 layout. Use bounded synthetic fixtures, no CAM mutation or paid quota. After corrections reuse unaffected captures and show only changed evidence. Stop before push for D2 A validation.

### T6 — Clear the authorized historical inventory and deliver

- [ ] Inventory and remove only authorized historical viewer sessions using supported deletion/tombstones, verify non-reingestion and preserved sources/settings, then complete the approved Git delivery and maintainer refresh.

**Closes:** AC8

**Depends on:** T5 and interactive acceptance

**Locations:** Current viewer's own store through supported authenticated management; selected branch/PR; installation only under maintainer refresh rules.

**Execution:** Main thread only.

**Test approach:** Deletion/failure/restart behavior already covered by existing Store/Delete tests plus bounded live observation.

**Changes:** Use only the fixed planning-time allowlist _support/workspace/2026-10-31-viewer-revisions-proportional-verification/cleanup-inventory.json (32 copies/27 groups, 29 archived, zero pinned/excluded). Reconcile at build entry but never add later members. No tokens/capability URLs in inventory. Preserve this task and subsequently created/changed evidence. Delete through T3's atomic expected_members guard, not separate precheck/unconditional deletion. Mismatched or unreadable groups remain untouched and are reported. Classify the existing ingest error only to understand ownership; do not delete unknown sources or erase diagnostics. Keep existing partial-failure recovery and exact remaining counts. Preserve sources, registry, authentication and settings. No raw rm of viewer state or repository-source deletion.

**Verification:** Inventory targeted group counts before/after, verify unchanged source files, scanner does not resurrect targeted hashes, unrelated/new groups survive, and pin/archive do not conceal survivors. Verify after scan/restart when maintainer restart is authorized. If live active runtime cannot safely perform supported deletion, stop that operation and state the smallest required action; do not fake completion. Then push/use branch PR after accepted D2 A validation, require green checks before merge, refresh binary then content under marker rules and preserve active viewer port/state/TTL. Report actual active revision/service state. Close only this work and its own temporaries; preserve branch until merged and its work is clean.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Contract and publication tests | T1/T3 isolated local fixtures | Go/Python agree, stale publication rejected | Build command |
| First rendered view | T2 before backend integration | Current/history sketches rendered | Build command and flow-build |
| Full suites | T5 local | Commands above and actual workflow coverage | Build command |
| Own UI walk and UX | T5 isolated viewer, 1280/390/320, both themes initially | Captures/measurements; focused corrections | Build command; required UX child |
| Native Codex /review | T5 candidate diff limited to this work | Verdict/findings and repairs | D3 |
| Human validation | Before push | Selected revision and current/history/mobile view accepted | D2 A |
| Historical cleanup | T6 actual viewer inventory | Allowlist, deleted/remaining counts, no unchanged return | D4 A; sources preserved |
| Git integration and maintainer refresh | After validation | Exact PR/merge/checks and active runtime revision | D2 A plus repository marker rules |

Human walkthrough: in isolated local viewer state, open a synthetic session containing Profile, invitation and Chat. Verify R1 initial, R2 mobile correction, R3 retained Profile/email, History and desktop/mobile switch. Open prior revision and confirm a visible superseded label; publish R4 while viewer is open and see only affected evidence change. A separate synthetic session with the same title must remain separate. Authentication uses task-owned disposable state; no real account is needed.

## Review status and progress

2026-10-31 — Planning only; all task checkboxes pending. One independent review completed: stored/presentable confirmation, atomic deletion precondition, bounded round invariants and capture-origin identity corrected. Checkpoint: flow-plan step 6, ready handoff; no build, cleanup, tracker write, commit or deployment performed.
