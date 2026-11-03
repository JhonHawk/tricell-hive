# Tasks

Base inspected: `development` at `158638346b1fae8fc58ff1172273538a67ac0375`; planning checkout `175a0263b0f15e14700e4432a84529ae8de71259`. Base reconciled at build entry: remote development remains 15863834. User started flow-build on 2026-11-03. All six tasks completed. Main thread owns integration, tracker/Git effects and verification. Implementation delegation is selected at build entry under flow-build; the plan does not authorize a child to publish or spawn children.

### T1 — Define the additive command and semantic input contracts

- [x] Add strict `work` parser/help/offline contract and typed semantic inputs/results, preserving the low-level commands.

**Closes:** AC1, AC6, AC8, AC9

**Depends on:** approved plan and delivery decisions

**Locations:** `tooling/cli/view_work*.go`, CLI routing/help and focused contract tests

**Execution:** main thread

**Test approach:** check for parser/interface cases; failure-first for consequential validation paths

**Changes:** Define all commands in design.md, explicit native source and no invisible service effects; bound successful mutation output and require all v3 semantic prerequisites without exposing internal versions.

**Verification:** Compile and exercise parser/help offline. Reject mixed creation/update selectors, unsupported states, incomplete blocked/closure input, internal ID/version injection and malformed source without state mutation. Existing low-level command tests still pass. New contract scenarios fail against the baseline because `work` is unsupported.

### T2 — Persist work bindings and recoverable publication intents

- [x] Implement safe publisher state, lock, work resolution/list/show/bind and exact-request replay.

**Closes:** AC2, AC4, AC7, AC8

**Depends on:** T1

**Locations:** CLI publisher module/package, `tooling/cli/view_work*.go`, focused persistence/subprocess tests

**Execution:** main thread or `hive-build-backend` if delegation pays for its brief

**Test approach:** failure-first for journal crashes, identity and concurrent writes

**Changes:** Separate versioned client journal outside viewer inventory; preserve explicit directory/source/work key binding; use index + GET, not repeated fresh create; pending reads report unavailability honestly. Read-only list/show must not mutate remote state. Scope-only rebind is explicit and refuses unresolved pending work. Distinguish unknown outcomes from definitively rejected intents and provide atomic retained archival through explicit `work reconcile`.

**Verification:** Two folders/keys in one native session remain independent through list/reopen; cwd changes are harmless; mismatched paths/sources fail; explicit accepted-producer rebind succeeds under its constraints. Crash before/after create acknowledgment recovers the same work. Two subprocesses contend safely; corrupt journal/unknown version remains untouched. Reconciliation retains bytes/receipts, refuses unknown outcomes, and crash recovery leaves exactly one active-or-archived classification. Start an unmodified v3 server with publisher state as a sibling and verify existing inventory validation still passes.

### T3 — Scan, freeze and publish document revisions under demand

- [x] Implement preflight, semantic no-op, upload/snapshot sequence and verified readback.

**Closes:** AC1, AC3, AC4, AC5, AC7, AC8

**Depends on:** T1, T2 for orchestration. The pure scanner can start independently with its settled document-set interface.

**Locations:** publisher scanner/sequencer, CLI adapters, real-server integration tests

**Execution:** main thread or `hive-build-backend`

**Test approach:** failure-first for bounded input, uncertain commit and conflict cases

**Changes:** Default plan documents plus explicit-file mode; deterministic order, limits and labels; immutable frozen bytes. Resume exact attempted requests and preserve partial uploads without falsely committing a review snapshot. Use remote current chain for no-op, never only local fingerprints.

**Verification:** One command publishes exact bytes and the expected tab list; second unchanged publication creates no remote event/receipt; changing a single document creates one relevant revision. Exercise missing required file, optional removal, >8 tabs, >256 KiB, invalid UTF-8, path escape, symlink ancestors, shared inode, changing source, stopped server, lost responses, crash at journal boundaries, unknown outcome, external stage/chain change and work deletion. Confirm frozen-byte retry after editing source. A definite predecessor conflict followed by explicit reconciliation must allow a fresh valid publication in the same active work while retaining the externally published revision. Check compact stdout and compare downloaded bytes/hashes with sources.

### T4 — Resolve structured task and lifecycle updates in the client

- [x] Add task registration/metadata transitions and explicit stage/close commands over existing v3 operations.

**Closes:** AC6, AC7, AC8

**Depends on:** T1, T2; T3 for plan-revision integration cases only

**Locations:** CLI semantic task/stage adapter and targeted real-server tests

**Execution:** main thread or `hive-build-backend`

**Test approach:** failure-first for transitions, authority and partial-batch recovery

**Changes:** Resolve keys and selected stage occurrences, retain segment order/assignee/plan references, carry exact receipts and versions internally. No Markdown status inference; no cascade to approval or whole-work completion. Advanced CI/artifact/membership operations stay on the existing API.

**Verification:** Register T1/T2 once, replay without reset/duplicate, resume a partial registration, update build/verify segments independently, reject absent membership or wrong assignment, require complete blocked/resolution data, preserve metadata reference rules, reenter stages explicitly, keep `in_review` distinct from closure. Publish edited task Markdown and confirm live registry states are untouched.

### T5 — Adopt the concise publication workflow in maintained guidance

- [x] Update the canonical procedure, public documentation and affected consumers without duplicating rules.

**Closes:** AC9, AC8

**Depends on:** T1 for settled command contract; T3 and T4 for verified behavior before final copy

**Locations:** `content/skills/flow-walkthrough/`, affected flow references/callers, `README.md`, `llms.txt`, `site/src/content/docs/viewer.md`, affected flow pages, `_support/docs/architecture/deployment-manager.md`, instruction-resource pointers as necessary

**Execution:** main thread; `hive-write-docs` only with verified sources and explicit file ownership

**Test approach:** check existing content/link/size tests; no tests mirroring prose implementation

**Changes:** Move technical protocol bookkeeping to the new client contract while preserving semantic duties and publication stops. Audit global/skills/roles/site/architecture/detector consumers. Keep the separately evolving research rule intact; no model pilots or synthetic regression additions.

**Verification:** `go test ./...`, `go vet ./...`, Python skill suite; inspect links and skill/reference limits. Run `pnpm --dir site check`, `pnpm --dir site lint`, `pnpm --dir site build`, then a local site preview and inspect affected pages, navigation/search, keyboard access, mobile and both themes per CONTRIBUTING.md. Reuse installed dependencies; if missing, follow the repository package provenance rules before a frozen-lockfile install. Measure `content/guidance/global.md` byte delta (expected 0) and list checked consumers for delivery. Manual comparison of before/after instructions against each preserved behavior; update obsolete detectors only if their actual parsing contract changes, retaining existing fixture cases.

### T6 — Verify the integrated user workflow and compatibility

- [x] Run the two-ticket scenario against an isolated real v3 server and produce bounded evidence of command/output reduction.

**Closes:** AC1, AC2, AC3, AC4, AC5, AC6, AC7, AC8, AC9

**Depends on:** T3, T4, T5

**Locations:** focused integration tests and task-owned temporary runtime/evidence

**Execution:** main thread; per-task `hive-verify-task` at contract/state/concurrency checkpoints as flow-build requires

**Test approach:** check + targeted fault injection; reuse unit/integration coverage rather than duplicate each assertion

**Changes:** No production changes. Verify with an old/unmodified v3 server binary and the candidate CLI, plus current low-level commands. Render the resulting documents in the unchanged viewer to confirm readable revisions and task separation.

**Verification:** Start an isolated temporary viewer with its own port/state, open work A, publish plan, revise/reload A, update tasks, open B in the same native session, publish B, list both, close A without affecting B, restart/recover interrupted publication. Read server state and bytes, then inspect rendered document tabs/history. Record high-level invocations and UTF-8 stdout bytes; one publication invocation per revision and <=4096-byte successful mutation responses. No token-saving percentage or model behavior claim. Stop only task-owned runtime and remove reproducible temp assets after readback.

## Verification and human review

| Check | When and where | Evidence | Authorization |
| --- | --- | --- | --- |
| Focused contract/state/concurrency tests | T1–T4, locally | Existing and new behavioral tests | Build reply required |
| Full Go suite and vet | Final candidate, locally | `go test ./...`; `go vet ./...` under umask 022 | Build reply required |
| Race suite | Final candidate, locally | `go test -race ./...` | Build reply required; confirm beforehand if expected to exceed 30 minutes |
| Skill tests | T5/final | `python3 -m unittest discover -s tests/skills -p '*_test.py'` | Build reply required |
| Site documentation | T5/final | `pnpm --dir site check`, `lint`, `build`; `pnpm --dir site preview --host 127.0.0.1`, affected-page/navigation/search/keyboard/mobile/theme checks | Build reply required; no publication |
| Real path and browser readback | T6 isolated server | Exact bytes, revision IDs, separate task histories, rendered docs | Build reply required; no habitual viewer restart |
| Plan review | Before handoff | `hive-review-plan` findings and resolution | Authorized by flow-plan |
| Code review | Candidate, before push | Native Codex `/review`; resolve findings before automatic delivery | D4 A; build reply still required |

No UI source change is planned, so no new UI composition/sketch or dedicated UX review is required. The existing rendered path is still checked. If implementation needs a UI/HTTP change, return that scope decision to the user before proceeding. Never broaden into model pilots.

## Final verification and delivery

- Runtime implementation: a5726c58; documentation: 6b722b35; review fixes: b25dc99d.
- Integrated PR #353 as c22ad325 on development. No development test CI configured; validation was local.
- Full Go suite and vet passed on final candidate. Full race suite passed before review fixes; affected publisher race suite passed again after them (35.030s).
- All 23 Python skill tests passed. Site check/lint/build used existing installed Astro/Biome binaries; browser verified navigation, search, keyboard, mobile and both themes.
- Independent task verification accepted T1–T4. Native Codex review covered6b722b35; focused re-review of b25dc99d found no actionable regressions.
- Real two-ticket scenario against unmodified API v3 verified exact Markdown, stable separate work IDs, cwd independence, unchanged no-op, one-document revision, history, explicit task transitions, separate closure and recovery. Successful publication receipts measured267 UTF-8 bytes. No model pilots or token-savings percentage.
- Review fixes covered authority before replay, bounded journal and semantic validation, frozen publication recovery, retained closure replay, required scope references and lead-owned collaborator segment binding.
- Global guidance byte delta0; canonical consumers inspected across skills, roles, site, architecture and pilot detectors.
- User requested close. Viewer record closed and read back; existing service kept running. Test-owned runtimes stopped.
- Local binary updated to c22ad325. Full content refresh deferred because installed concurrent instructions from56704a0a are not yet in development; refresh only from the combined integrated source.
