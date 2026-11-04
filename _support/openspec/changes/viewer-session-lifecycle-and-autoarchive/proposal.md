# viewer-session-lifecycle-and-autoarchive — Clear completed work from the viewer

| Field | Value |
| --- | --- |
| Status | In verification · T1–T5 built; real-path acceptance and reviews underway |
| Tracker | None; creating a ticket is outside this plan |
| Git | tricell-hive; base `development` at observed `2765624b`; prerequisite candidate PR #352 at `8a04118d` |
| Viewer | Work `d175931f5bfa5078e1c87da68b2ee6c74fcc6148142a4d47e606cf175131e6b6`; native Codex session `01a1a55e-9a93-77e0-9237-dc4758e6e4da` |

## Decisions

2026-11-04, answered through the native question tool:

- When does a completed work leave the active lifecycle? Close its viewer record before asking about hygiene; keep review, validation and pause states honest. Cleanup remains a separate user decision. (D1 A)
- How long does completed work remain in the normal list? Two hours after the current explicit closure. (D2 A)
- What determines recency? Last substantive work activity, excluding rename, pin, archive and restore. Keep the internal timeline's existing reading order. (D3 A)
- What overrides automation? Pin prevents autoarchive; manual restore exempts that closure episode until explicit reopen and a new closure; unpin reuses the original deadline; reopen restores automatically archived work but preserves manual archive. (D4 A)
- Do older clients stay compatible? No simultaneous old-client support: update CLI, server and UI together, safely reject outdated clients before mutation, and require page reload. (D5 B)
- What happens to existing closed work? Use the recorded closure time, immediately processing overdue eligible work; grant a full two hours from persisted first evaluation only when no closure time can be proved. Honor prior manual restores and pins. (D6 A)
- How is the built change delivered? Ready: publish the reviewed candidate as a non-draft PR, wait for checks and answer reviews, then leave merging to the user. (D7 B)
- Which candidate code review runs? Native: Codex /review, in an independent context; the project's Claude Code preference is unavailable as a native mechanism on this host. (D8 A)

- May implementation use the still-open prerequisite candidate? Yes: build on `8a04118d`; preserve #352 and resolve its foreign commits before publication if it remains unmerged. (D9 B)

## Request

The user reported: “Muchas sesiones sin cerrar, ya estan generando demasiado ruido, porque no completan su ciclo interno” and proposed “Sesiones Finalizadas, en 1 o 2 horas se archivan automaticamente” with “la más reciente arriba”. The user then asked to inspect the concurrent specs/closure implementation before planning this change and explicitly invoked `$flow-plan`.

## Objective

Completed authorized deliveries stop appearing active while optional hygiene waits. Explicitly closed works disappear from default navigation after two hours without losing content; manual choices persist. The list shows recent work first and distinguishes active work from review, validation, pause and failure.

## Scope and acceptance

The build changes distributed closure guidance, the local viewer's contract/storage/runtime, its existing navigation, and corresponding documentation/specs. It does not infer completion from elapsed inactivity, merge PR #352, promote to QA/production, fix other projects' old specs, repair publisher membership, delete viewer content, or run model behavior pilots. A dedicated code review and UX review are part of verification; a security review is not requested.

Each criterion describes a new observable result. Its baseline is the observed `development` reference, plus the explicitly named prerequisite where relevant; existing preservation guarantees are test constraints, not new acceptance criteria.

- AC1 — The initial normal completion report observes viewer close/readback and offers one subsequent hygiene footer, with support decisions and cleanup still deferred. Falsifier: the report waits for support decisions or its footer leaves viewer close deferred. The observed base asks the support decision before its completion report; the actual build prerequisite PR #352 moves that decision after the report but also defers viewer close. Thus this criterion grades the combined sequence against the accepted dependency, not a claim that base lacks all closure behavior. Verify with the canonical procedure and all consumers, not a synthetic model pilot.
- AC2 — An unpinned, nonexempt explicitly closed work is automatically archived once 7,200 seconds have elapsed, at startup or within the next 60-second runtime sweep. Falsifier: it archives before the deadline or remains visible after the sweep. Base has no sweep and only manual archival.
- AC3 — A manual restore persistently exempts the current closure; explicit reopen/new close resets the exemption, unpin reevaluates the original deadline, and reopening automatically archived work restores it atomically. Falsifier: restore is immediately undone, the exemption is lost at restart, or reopened automatically archived work stays hidden. Base has no archive policy or episode exemption.
- AC4 — Navigation orders works within each existing group by last substantive work activity descending, with a stable ID tie-breaker, and displays that activity's local date/time. Falsifier: hash order, formatted-date sorting, administrative edits or automatic archival determine position. Base sorts IDs and displays creation time.
- AC5 — Each row exposes its actual state visibly and accessibly: En curso, En revisión, Esperando validación, Pausada, Finalizada or Fallida, with equivalent English copy. Falsifier: review, pause or validation is collapsed to En curso, or state is available only through color/hover. Base collapses those states.
- AC6 — Existing closed records recover their current closure episode from server-recorded transitions; records lacking a provable closure timestamp get a persisted two-hour initial window without inventing a historical closure. Falsifier: mtime/updated_at is treated as closure, restarts reset the window, or prior manual restores are ignored. Base has no such policy.
- AC7 — The upgraded CLI/UI/server share an explicit contract revision; missing/old revisions fail with an actionable update/reload response before any timeline/management mutation. Existing stored work, documents, artifacts and publisher replay identities are preserved. Falsifier: an old client mutates the new store, a mismatch causes input deletion, or the new reader silently drops old history. Base has no revision gate for this changed contract.
- AC8 — Automatic archival and its manual reversal have bounded durable capacity independent of public event/receipt exhaustion, serialize against user actions and recover after write failures without impersonating actors. Falsifier: a race overwrites a newer choice, an uncertain write is reported successful, or a saturated public event/receipt budget lets a work be hidden without a working Restore. Physical I/O failure or numeric version exhaustion must leave work visible and report the failure rather than hide it irreversibly. Base has no system archival action or restore capacity independent of public budgets.

## Prerequisites

- [x] The user explicitly accepted PR #352 candidate `8a04118d` as the dependency base (D9 B). T4 may proceed. The PR remains open; its merge is not authorized by this answer. Before publishing, compare the foreign #352 commits with the current base and resolve their inclusion as Delivery requires.

No third-party account, paid quota, new package, real test account or external data write is needed for local verification. Use the existing installation and isolated temporary viewer state with synthetic works.

## Verified context

- PR #352 was OPEN at head `8a04118d`, base `2765624b`, with no CI check runs when inspected; this is not an integrated prerequisite.
- The checkout was on PR #352's branch with an unrelated edited `tasks.md` and unrelated untracked support folders. Preserve them; implementation must use its own worktree/branch after the dependency decision.
- The running habitual viewer reported revision `b35d7544` on port 4483. The installed closure/plan/git/status skills differed from PR #352. This plan does not update or restart them.
- `timeline_store.go:timelinePublish` sorts by ID; `api.ts:mapTimelineIndex` displays creation time and collapses four live states; `scan.go:Run` waits after startup without a retention ticker.
- Current closure history contains server `state_transition` events even for inspected closed works without a standalone closure event. Unknown timestamp handling remains necessary for other valid legacy records.
- Root AGENTS.md governs this repo: Project tricell-hive; base development; environments development → master; tracker GitHub Issues JhonHawk/tricell-hive; Specs `_support/openspec`. Maintainer refresh applies only when the marker and an integrated delivery meet its conditions.
- The root Measurement rule requires user approval for relaxing a governing rule; D1 A is that explicit product decision. The observed cases do not establish two failures on the same deployed rule, and this plan makes no such claim.

## Delivery

Delivery mode: ready (D7 B). After the user starts the build and the candidate passes checks and Native: Codex /review (D8 A), commit/push the work and open a non-draft PR, wait for applicable checks and answer reviews, then stop with the PR open for the user. Never merge, enable auto-merge or add a merge queue entry. UX review and the producer's browser walk still run; this mode adds no pre-push human validation stop. Current invocation authorizes the plan and its local viewer publication, not implementation. Maintainer install/restart is deferred until a later authorized integration meets AGENTS.md.

## Review status and progress

- 2026-11-04 — D1–D6 answered; backend feasibility researched read-only in a fresh child. This is design evidence, not implementation verification.
- 2026-11-04 — D7 B ready and D8 A native Codex review selected while one independent plan review runs.
- 2026-11-04 — One hive-review-plan review found two material gaps. Source inspection confirmed saturated public event/receipt budgets can block TimelinePatch restore, and three canonical compatibility requirements contradicted D5 B. The plan now provides bounded auditable visibility-control recovery and deltas for every affected compatibility requirement. AC1's prerequisite-relative baseline is explicit. No second plan-review round or implementation was run.
- 2026-11-04 — Planning checks passed: five complete Markdown files within publication limits; every AC has an explicit task mapping; ASCII sketches are at most 60 characters wide; all four affected management/compatibility requirements have replacement deltas; git diff whitespace check passed. PR #352 remains OPEN at 8a04118d. These checks do not claim candidate behavior or model compliance.
