# Tasks

Observed base: `development` at `2765624b`; prerequisite PR #352 candidate `8a04118d`, not yet merged. Do not implement on its occupied branch. Use a separate authorized branch/worktree after reconciling the prerequisite and move this change folder as the current git-workflow sets. Planning creates no worktree, branch or code changes.

The main thread owns coordination, source adjudication, checks, viewer publication and authorized Git delivery. Children start fresh, stay in assigned paths, never launch children or perform Git/tracker/memory writes, and preserve others' changes. T1 freezes a typed contract; only then may T2 and T3 run in parallel in disjoint Go-store and UI files. T4 is independent of them after the #352 dependency is settled. Run one plan review. Delivery is ready (D7 B), with Native: Codex /review (D8 A) in an independent context; leave the delivered PR open for the user's merge.

### T1 — Define the coordinated contract and stored lifecycle metadata

- [ ] Publish typed Go/TS projections, contract revision/error handshake, strict normal decoders and backward stored-read initialization, preserving existing history and request replay identity.

**Closes:** AC7, AC6 (contract and stored-read portion)

**Depends on:** none; confirm integration base before editing.

**Locations:** `tooling/viewer/timeline_contract*.go`, `timeline_protocol.go`, `server.go`, `timeline_server.go`, `timeline_store.go` loading path; `tooling/cli/view_timeline.go`, HTTP client and protocol validation; `tooling/viewer/frontend/src/timeline-contract*.ts`, `api.ts`; typed tests/fixtures.

**Execution:** hive-build-backend, because the public/stored contract crosses more than three files. Main thread checks Go/TS parity before allowing dependent work.

**Test approach:** tdd, existing old/new client and persisted-record fixtures.

**Changes:** Matching timeline revision is required before writes; old/new mismatches preserve native requests and return actionable errors. Define storage metadata, bounded local administrative receipt slots, a maximum 16 KiB maintenance allowance independent of existing content budgets, and projection invariants from design.md. Replace every canonical mixed-generation compatibility promise in the delta, preserving its other guarantees. Freeze the exact typed contract and file ownership before parallel builders start.

**Verification:** Focused Go contract/protocol/server/CLI tests and frontend decoder tests prove old-client writes make zero store changes, new clients reject old-server mutations, corrupt metadata fails explicitly, valid legacy records load, documents/history remain byte-identical and original pending operation IDs survive update/resume. Compare the replacement deltas with the canonical Compatible opt-in navigation projection, On-demand high-level publication and Bounded publication interface and compatibility blocks and their scenarios; no old-server interoperability promise may conflict with revision 2. Verify exact administrative-slot replay versus rejected older stale replay and that content cannot consume its maintenance allowance.

### T2 — Implement episode-aware automatic archival and activity projection

- [ ] Persist episode dates, historical restore exemption, archive origin, activity and bounded system provenance; run deadline-aware startup/ticks and serialize policy/manual mutations with existing durable storage.

**Closes:** AC2, AC3, AC6, AC8; backend portion of AC4

**Depends on:** T1's frozen types and protocol tests.

**Locations:** new `tooling/viewer/timeline_maintenance.go` and tests; store/lifecycle integration; `scan.go`, `timeline_capacity.go` and focused runtime tests. Do not edit frontend or CLI contract files while T3 runs.

**Execution:** hive-build-backend; use deterministic injected clocks, not two-hour sleeps. A fresh hive-verify-task checks this task against AC2/AC3/AC6/AC8 before it is marked verified because it changes stored data and concurrency.

**Test approach:** tdd, advancing an injected clock around 7,199/7,200 seconds and one 60-second sweep window.

**Changes:** Archive only explicitly closed, unpinned, nonexempt work. No purge, idle completion or synthetic user actor. Recover proved dates and original manual choices; unknown dates get one persisted window. Preserve public content budgets and handle uncertain writes through existing reconciliation.

**Verification:** `go test ./tooling/viewer ./tooling/cli` with focused tests for deadline boundaries, offline/startup catch-up, repeated restart, restore/unpin/reopen/new-close, manual-origin archive preservation, native replay/stale version behavior, simultaneous archive/restore/reopen/pin and write-failure recovery. At full legacy public event, receipt, body and serialized-record budgets run initialize → autoarchive → Restore → exact retry → stale retry → restart, verifying exemption and local provenance without pruning history. Inject numeric version exhaustion and physical write failure and prove visibility is preserved. Run `go test -race ./tooling/viewer` on maintenance/concurrency tests. Hash document/artifact fixtures before/after. The verifier also checks that ordinary reads, administrative changes and unchanged replay do not advance last_activity_at.

### T3 — Show recent works first with actual states

- [ ] Consume last_activity_at and true lifecycle states in the existing navigation; keep selected-work context usable when background autoarchive occurs.

**Closes:** AC4, AC5; UI portion of AC3/AC7

**Depends on:** T1's frozen typed contract, not T2 or a first-view approval.

**Locations:** `tooling/viewer/frontend/src/shell.tsx`, `application.tsx`, `controller.ts`, shared catalog/messages, archive notification/header as needed, typed test/dev fixtures and generated embedded assets. `api.ts` contract/mapping changes follow T1 sequentially and must not overlap another writer.

**Execution:** hive-build-frontend; use installed ReUI components. Main thread performs browser verification; one hive-review-ux child reviews rendered changes and containing layouts.

**Test approach:** tdd for order/state/focus/selection/restore behavior; existing UI checks for visual structure.

**Changes:** Sort actual timestamps descending with stable ID ties, preserve existing grouping/search/counts and internal timeline order, show default-size semantic Badges, display last activity, keep archived selected work until navigation without a persistent global banner or background-toast flood. Provide update/reload state without automatic mutation replay.

**Verification:** Frontend `check`, `lint`, `test`, `build`, `verify:dist`; isolated browser with at least 20 realistically named works, both themes, desktop and 440 × 956. Exercise search, include archived, expiry of selected work, restore, pin/unpin, keyboard focus and mobile Sheet. Capture affected regions and run the installed ReUI self-check and accessibility audit under the UI-verification procedure; fix measured defects, do not install tools.

### T4 — Finalize viewer state before the hygiene footer

- [x] Move viewer close/readback into the initial completion phase and keep support decisions/archive/cleanup deferred; align all guidance consumers without undoing #352's specs/pending-work behavior.

**Closes:** AC1

**Depends on:** PR #352 dependency resolution; independent of T1–T3.

**Locations:** `content/skills/flow-close/SKILL.md`, `content/skills/flow-walkthrough/references/viewer-publication.md`, site close page, architecture workspace/deployment explanations when their text states the sequence; global guidance/roles/pilot detectors only when the behavior search finds a conflicting dependency.

**Execution:** main thread if scope is at most three edited files/five large reads; otherwise hive-write-docs with these exact paths assigned. Canonical policy stays in flow-close and consumers point to it; split a reference by responsibility if the 16,000-byte skill limit would be exceeded.

**Test approach:** check, existing content/package/link/size tests plus a source-backed consistency ledger; no synthetic new regression fixture or model pilot.

**Changes:** Follow design.md's state matrix, missing-runtime/authority fallback and readback-before-input-cleanup requirements. State user approval D1 rather than claiming recurrence proof. The eventual guidance commit names the observed session and date under root Measurement rules.

**Verification:** `go test ./tests/content/...`, instruction-resource validation and a behavior search across global, skills/references, agents, site, architecture and detectors. Record every checked location and global.md byte delta for the PR. Confirm no path keeps viewer close pending in the initial normal completion phase and that ready/paused paths never claim integrated completion. Ordinary static checks prove instruction consistency, not model compliance; real-session follow-up remains separate.

### T5 — Integrate the policy, document the upgrade and verify the candidate

- [ ] Deliver aligned viewer/protocol docs and spec deltas; exercise the compiled CLI, API and browser against isolated persisted state with the full policy and closure procedures.

**Closes:** AC1–AC8 (cross-boundary observations; T1–T4 retain their own proof)

**Depends on:** T2, T3, T4.

**Locations:** `site/src/content/docs/viewer.md`, architecture viewer/protocol docs, `CHANGELOG.md`, this change's deltas and progress; README/llms only where the changed contract/required update alters a stated claim. Include the three replaced compatibility requirements and narrowed Local viewer management actor requirement when reconciling deltas at integration. Do not edit other projects or the occupied #311 record.

**Execution:** main thread for integration and verification; hive-write-docs only for a bounded settled docs surface if delegation thresholds require it. Use the chosen candidate code review; one UX reviewer, one T2 task verifier, no extra blanket verification agent.

**Test approach:** end-to-end integration using existing harnesses and a temporary compiled binary/state, not host/model behavior pilots.

**Verification:** Full `go test ./...` and `go vet ./...`; frontend checks above, site build and focused runtime race tests. Start a task-owned isolated server on an available port and exercise compiled CLI → store → index → browser with injected-time test control or a test fixture clock (never change machine time or the habitual runtime's clock). Prove revision mismatch blocks writes, expiry and restore survive restart, stored bytes remain, native operations recover after version conflicts, and cleanup pending does not mislabel a completed viewer record. Maintain a per-criterion evidence ledger with commands/results; distinguish source consistency from real model evidence. Preserve a test-state export only if needed as unique evidence and ask before removal; remove reproducible task-only fixtures/processes by ownership.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Plan review | Before ready-plan handoff | One fresh hive-review-plan report and adjudication | flow-plan invocation |
| T2 task verification | T2 checkpoint | Fresh hive-verify-task AC2/3/6/8 verdicts | Build after plan approval |
| Candidate checks | Local isolated candidate | Go/vet/race/frontend/site output and AC ledger | Build after plan approval |
| Browser/UX | Desktop and 440 × 956, light/dark | Focused captures, keyboard/restore flow and accessibility results | Build after plan approval |
| Candidate code review | Before push | Native: Codex /review in independent context | D8 A |
| Human acceptance | User reviews the open ready PR | Ordered rows, states and reversible archival evidence | D7 B; user owns merge |

No install/restart or update of the habitual viewer is a test. A later integrated maintainer refresh follows AGENTS.md, records original service settings, preserves state, takes/retains the pre-upgrade snapshot, reports new revision/service state and token invalidation, and offers opening the viewer without a separate decision ID. Do not restart before the authorized delivery.

## Review status and progress

- 2026-11-04 — D1–D6 settled. This is a plan; no task implemented, no tests run for a candidate, no change folder committed or published to Git.
- 2026-11-04 — D7 B ready and D8 A Native: Codex /review settled. Independent plan review running; no code build started.
- 2026-11-04 — One plan review completed. Confirmed gaps in restore-at-capacity and retained mixed-generation compatibility promises were corrected in the plan and spec deltas; AC1's accepted-prerequisite baseline was clarified. No implementation checks claimed.
- 2026-11-04 — Ready-plan checkpoint: all product/delivery/review decisions answered. Five documents pass planning shape/mapping/width/publication-size checks. PR #352 remains OPEN at 8a04118d; the user owns its merge. Next: publish the plan snapshot to the retained viewer work, present its summary and wait for the build route after dependency resolution. No source implementation, Git delivery, install or runtime restart has occurred.
- 2026-11-04 — Plan snapshot confirmed: viewer work d175931f5bfa5078e1c87da68b2ee6c74fcc6148142a4d47e606cf175131e6b6, revision revision-5ac1a8ce560c1469702aa75bc0ab2d1d, session version 11, plan occurrence event-17a9ce89f69c0c9f31e05b648819ebf0, five documents. Next action is the user's build route after dependency resolution, not another publication for this bookkeeping line.

- 2026-11-04 — Build started by explicit user command. Worktree `/Users/jmartinez/.codex/worktrees/viewer-session-autoarchive/tricell-hive`, branch `feat/viewer-session-autoarchive`, base `2765624b`. The task-owned change folder was moved here; original checkout/other work preserved. D9 prerequisite choice requested asynchronously; T1–T3 can proceed. Last commit: base 2765624b; next: T1.

- 2026-11-04 — Worktree location corrected before children edited: `/Users/jmartinez/Development/projects/tricell/tricell-hive/.codex/worktrees/viewer-session-autoarchive`, ignored through `.git/info/exclude`. Publisher bind rejects directories outside the registered project; moving this owned worktree preserves containment without bypasses or duplicate plans. Local read-only inventory: 30 work records; no production/QA state in this task; isolated verification state starts empty. T1 delegated to build_t1_contract, API/storage contract not yet frozen.

- 2026-11-04 — D9 B: user accepted candidate8a04118d as dependency base. Own branch fast-forwarded from2765624b to8a04118d after proving no tracked-path overlap with T1 edits; no history rewrite or change to #352. T4 unblocked. Publication still requires current-base/foreign-commit check.

- 2026-11-04 — T4 implemented by build_t4_closure. Main-thread checks: go test ./tests/content/... PASS (0.498s), management instruction/link/anchor checks PASS; code diff inspected. Canonical phase now1,3,6,7beforefooter; deferred2,4,5. Four files; flow-close15804bytes, global34029→34029(0delta). Commit pending; no modelpilot run.
