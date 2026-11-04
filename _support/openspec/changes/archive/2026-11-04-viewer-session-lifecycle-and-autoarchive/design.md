# Design

## Approaches

The user chose to finalize the viewer record at the authorized delivery's completion point and keep optional hygiene separate (D1 A). The alternative kept the deferred close and introduced a new pending-close lifecycle state; it retained an extra user step and expanded the state machine without completing the record. Existing in_review, awaiting_validation and paused states remain meaningful and are not auto-completed. The two-hour policy was chosen over one hour (D2 A).

The user chose last substantive activity over creation time (D3 A), persistent manual restore over another two-hour countdown (D4 A), a coordinated contract update over simultaneous backward-client support (D5 B), and recorded closure times over a universal extra grace period (D6 A).

## Proposed shape

### Work closure and hygiene

Keep `flow-close` as the canonical owner of the completion sequence. Its initial completion phase includes step 1's viewer transition/readback, step 3's tracker/specs commit, the final report and memory/pending-work recording. Its footer defers only the support decision, change-record archive and resource/branch cleanup (steps 2, 4 and 5). On a ready/human-review stop, step 1 sets in_review; on a real pause it sets paused; only integrated/abandoned terminal work becomes closed. QA promotion stays a separate operation, consistent with the project completion setting when one exists.

The subsequent hygiene phase must not close an already closed viewer again. Retain any readback/receipts needed to safely clean its upload inputs. If viewer publication is unavailable or authority is missing, keep the current limitation path: preserve inputs, report the recorded state honestly and continue independent hygiene. Do not infer closure or recover authority silently. Preserve PR #352's new specs commit and pending-work semantics, including its distinctions for operational work and declared promotion paths.

Update all direct consumers of this sequence: `flow-walkthrough` publication reference, site close/viewer pages and workspace/deployment architecture where affected. Search the global guidance, role definitions and pilot detectors for dependent statements; make changes only where the behavior is stated. Global guidance needs no new user ceremony; record its before/after byte count in the eventual PR, including zero delta. Keep existing regression fixtures unchanged and do not run model pilots.

### Contract and client coordination

Keep timeline schema/storage version 3 and its existing source identities, URLs and request DTO meanings. Introduce a separate timeline contract revision constant, initially 2, for this coordinated API extension. Advertise it in HTTP response headers and the offline protocol/catalog metadata, and require matching `X-Hive-Timeline-Revision: 2` on index/timeline metadata requests and every timeline mutation. CLI and UI send it; health/bootstrap/authentication and capability-protected image/file reads retain their existing access contract. Return HTTP 409 with `code: reload_required` and a safe update/reload instruction before parsing/committing mismatched timeline writes. A new client also refuses mutation against an old server that does not acknowledge the revision. Do not enable CORS or weaken same-origin/action-scope checks.

Extend the strict shared Go/TS session and index projections with:

- `last_activity_at: string` — required valid server timestamp, no later than updated_at;
- `closed_at: string | null` — current episode's proved closure time, null when unknown/not closed;
- `autoarchive_at: string | null` — effective deadline for eligible closed work, null when active, pinned or manually exempt;
- `archive_origin: "manual" | "automatic" | null` — how current archival happened; null when not archived.

No configurable TTL flag, policy settings screen, compatibility projection, new service, package or journal is introduced. The protocol catalog documents two-hour automatic visibility archival and manual content deletion distinctly; `retention: manual` continues to mean content preservation, not permanent default-list visibility. Update the CLI's response validation and error handling, frontend decoders/fixtures and generated embedded bundle together. Replace the canonical requirements Compatible opt-in navigation projection, On-demand high-level publication and Bounded publication interface and compatibility so their grouping/output/authority guarantees remain, but they no longer promise cross-revision operation on unmodified servers. Pending native requests retain their operation IDs and bytes; a contract mismatch never replays or deletes them automatically. After upgrading a client, explicit resume uses the original payload/operation ID; do not manufacture requests under the new revision to clear errors.

### Stored policy metadata and closure episodes

Add a bounded optional private maintenance structure to the existing timeline record. It holds policy version, current closure-episode identity, proved closure time or null, persisted fallback eligibility start, manual-restore exemption for that episode, current archive origin, latest automatic action provenance and last substantive activity. Missing metadata is valid only on the stored-read path; normalize it before validating/projection and persist the first eligibility decision once. New public decoders stay strict. Public derived fields are projections of this canonical stored metadata, not a second independently editable stored state.

Find the current episode from the final explicit server-recorded state transition into closed after any reopen, or its corresponding closure event; use server receipt timestamps, never occurred_at, mtime, updated_at or an idle timer. If closure time is unprovable, keep closed_at null and start a full two-hour eligibility window at the first persisted evaluation. A later restart uses the saved window. Existing local archive/restore events after the current closure establish manual origin/exemption, so updating the viewer does not undo a historical manual restore. Archived records with ambiguous provenance remain manually archived; never guess automatic origin.

Maintain activity from successful substantive native publication/lifecycle mutations: stages, evidence, document/snapshot publication, tasks, ticket/delivery/candidate catalogs and CI waits. Creation supplies the initial value. Exclude native/local rename, pin/archive/restore, join/recover/foreground-only administration, reads, replay/unchanged results and automatic maintenance. Reconstruct existing activity from accepted native public events and private operation audits using server receipt times; use creation only when no work activity can be proved. `updated_at` remains general modification time. Backend and UI sort numeric timestamps, then ID for ties, and do not sort localized labels.

### Automatic archive transaction

Add a startup catch-up and a 60-second context-cancelable tick to Store.Run. Use the existing mutex, uncertainty reconciliation, clone/validate/atomic durable write and index publication path. Select/recheck eligibility under that lock: explicit closed state, not already archived, not pinned, no current-episode restore exemption and deadline reached. A failed metadata initialization never makes an unknown-time record eligible.

Persist `archived=true`, archive origin automatic and bounded system-action metadata without fabricating native/local_viewer actors or adding an unbounded public event/receipt stream. Keep normal optimistic versions coherent: one successful archive bumps the work version once; later sweeps are no-ops. Metadata initialization must also invalidate any version whose projected contract fields changed. Preserve native pending-request receipts so a replay remains idempotent; stale writes return version_conflict rather than overwriting policy or retrying with invented intent. Failed/uncertain persistence follows existing reconciliation and never returns a confirmed archive prematurely. Report sweep errors through the existing RunOptions callback and continue unrelated works where safe.

Manual restore stores an episode exemption atomically, including an explicit restore of already visible closed work when the API operation expresses that intent. Reopen clears the old episode/exemption and automatically unarchives only automatic-origin records; a manually archived work stays manually archived. Closing again creates a new deadline. Pin disables eligibility while set; unpin reevaluates the original deadline, unless restored-exempt. No policy action changes lifecycle, documents, artifacts, source identity, producer authority or native process state.

Restore capacity is a requirement, not an assumption about spare public events. For revision-2 local_viewer archive/restore/pin-only patches, persist bounded auditable administrative state/receipt slots in maintenance metadata; do not require appending a public local_action or consuming a native-publication receipt when those budgets are full. Preserve all existing public history and native receipts. Each slot records the actual local_viewer provenance, operation ID, request digest, expected/committed version, timestamp and compact result; exact retry of the retained operation replays its result, and an older stale operation fails version_conflict without mutation. Do not claim unbounded replay history for replaced administrative slots. Existing unrelated rename/deletion/publication paths keep their existing audit/receipt guarantees, and no native producer can use the local administrative branch.

Reserve at most 16 KiB of additional per-record maintenance storage outside the unchanged content/public event/receipt budgets, including a legacy record already at its former serialized limit. Define strict per-field and aggregate limits in T1; content cannot consume that allowance. Administrative-only growth of Version/UpdatedAt/archived/pinned is accounted from a bounded private baseline inside that same allowance; ordinary publications retain strict original admission. Response read/decoder limits add only the same 16 KiB for projection/envelope overhead, while request admission remains unchanged. Storage encoding keeps derived public fields out of the content-budget calculation. Test legacy full-size records as well as full event/receipt ledgers through initialize → autoarchive → restore → exact retry → stale retry → restart. If physical storage or numeric version exhaustion prevents safe autoarchive and reversal, leave the work visible and report maintenance failure. Never prune evidence or use deletion as recovery.

### Upgrade and recovery

Use isolated copied fixtures for migration tests, not the user's habitual viewer state. Upgrade adds optional record metadata through the existing atomic persistence path, without moving or deleting old records/blobs or synthesizing timeline events. Preserve malformed records as errors, not as defaults that archive them.

The eventual coordinated install requires a pre-upgrade private state snapshot before the first new maintenance write and preserves the publisher journal, registry, token access semantics and stored copies. A snapshot is unique evidence, must be named and retained, and its removal requires a later explicit decision. Downgrade-in-place is unsupported if the old reader rejects new metadata; rollback uses a separate restored pre-upgrade state after stopping only the authorized viewer, retaining new state and post-upgrade records for reconciliation. Never replace current state with the snapshot automatically. Rebuild/install/restart occur only through the authorized delivery and maintainer-refresh conditions; planning starts or stops no service.

## UI

Surface: operational. Archetype: existing Record detail with Sidebar. Container family: existing Frame-led shell. Density: existing dense controls. Regions: project sidebar, selected-work header and existing timeline. States: preserve current skeleton, actionable empty, inline error/Retry and access-error handling; add actionable update/reload state for mismatched revision. Motion: none added; preserve installed transitions and reduced-motion handling.

Reuse `tooling/viewer/frontend/src/shell.tsx`, the shared state catalog and `timeline-header.tsx`; no new composition requires a design agent. Existing installed components are Sidebar, Collapsible, Input, Tooltip, Skeleton, ReUI Badge, Alert and Button. Row state uses Badge at default size: active info-light, review focus-light, awaiting validation warning-light, paused secondary, closed success-light, failed destructive-light. Do not style Badge height/padding/radius/motion. Lucide remains the icon library.

Use a title line followed by one compact state/date line, plus the existing ticket/origin metadata when present. Keep source metadata, full accessible titles, selected-ID and group expansion persistence, filtering and existing archive/restore/pin controls. The session header keeps its existing actions and accurate state. Archiving a selected work must not redirect to an arbitrary next work or flash an empty data failure: retain that selected record in the current view until navigation, with an existing-pattern archived explanation and Restore action; do not add a persistent application-wide banner. Do not show a toast for every background archive. Manual restore remains available through Include archived and the existing action; keep existing temporary manual-archive Undo behavior.

Illustrative fixture, not a fresh measurement of live sessions; 3 works in the same group before and after:

```text
Before: Sidebar / Collapsible — cam (3)
  Portal choice borders
    COM-915 · 3 nov, 23:20       [dot: En curso]
  Meta Graph API
    COM-910 · 3 nov, 03:19       [dot: Cerrada]
  Transactional
    COM-920 · 3 nov, 20:15       [dot: En curso]

After: Sidebar / Collapsible — cam (3)
  Portal choice borders
    [Badge: Finalizada]       4 nov, 01:44
    COM-915
  Transactional
    [Badge: Esperando validación] 3 nov, 21:33
    COM-920
  Meta Graph API
    [Badge: Finalizada]       3 nov, 11:28
    COM-910

After overdue eligible archives: cam (2)
  Portal choice borders
    [Badge: Finalizada]       4 nov, 01:44
    COM-915
  Transactional
    [Badge: Esperando validación] 3 nov, 21:33
    COM-920
  [Include archived toggle: off]
```

At mobile 440 × 956, keep state/date wrapping within the existing Sidebar Sheet, no horizontal page overflow or tiny text, and preserve focus/selection during refresh/reordering. Match the displayed date to last_activity_at; full timestamp remains accessible. Test keyboard navigation and both supported themes with realistic long titles and at least 20 works across existing groups. The retained selected autoarchived work remains separately counted from the default navigation rows; it must not silently add itself back into that count.

## Copy

| Meaning | Spanish | English |
| --- | --- | --- |
| active | En curso | In progress |
| in_review | En revisión | In review |
| awaiting_validation | Esperando validación | Awaiting validation |
| paused | Pausada | Paused |
| closed | Finalizada | Finished |
| failed | Fallida | Failed |
| activity date accessible label | Última actividad | Last activity |
| automatic archive explanation | Archivada automáticamente dos horas después de finalizar. | Automatically archived two hours after completion. |
| selected archived action | Restaurar | Restore |
| revision mismatch | Actualiza Hive y recarga el visor para continuar. | Update Hive and reload the viewer to continue. |

## Impact

Consumers include the embedded frontend, CLI low-level timeline client and semantic publisher, offline protocol description, persisted timeline records, unit/runtime/capacity tests and the viewer/manual workflow docs. Existing native sessions must use the upgraded CLI after the coordinated update. This affects the local viewer and distributed guidance, not a hosted product environment or another repository's data. The closure procedure must retain PR #352's specs/pending-work work rather than reintroducing its old coupling.

## Recovery

Local implementation remains in its own worktree until delivery. Automatic archive is reversible via Restore/Include archived; evidence is never removed. Policy exemptions and archive origin survive process restart. A new explicit closure reenables the policy. Contract rollback and preservation of post-upgrade state follow the upgrade section; a Git revert alone is not claimed to downgrade a changed store safely.
