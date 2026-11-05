## MODIFIED Requirements

### Requirement: Hierarchical React viewer navigation

The shell SHALL preserve the existing project/workspace groups, source metadata, title/ticket search, counts, selection, expansion, keyboard/mobile navigation and full accessible labels. Works within each group SHALL be ordered by last substantive native work activity descending with stable ID ties, and show that activity's local date/time. Creation SHALL supply the initial activity timestamp. Administrative rename/pin/archive/restore, reads, unchanged replay and automatic maintenance SHALL NOT advance activity. Public events and private native operation receipts SHALL support truthful reconstruction of older activity without using source mtimes or localized labels as sort keys.

Every row SHALL show its actual active, in_review, awaiting_validation, paused, closed or failed state visibly and accessibly through the installed semantic Badge variants, not color alone. Existing internal timeline ordering SHALL be preserved. Loading, empty, error and authentication states SHALL remain actionable. A selected work archived by maintenance SHALL remain readable in its current view until navigation, with an archived explanation and Restore action, without selecting an arbitrary other work or a persistent application-wide banner.

#### Scenario: Work resumes in an old session
- WHEN an older work receives a new substantive publication
- THEN it precedes less recently active works without moving its internal timeline groups.

#### Scenario: Administrative changes
- WHEN a work is renamed, fixed, archived, restored or automatically maintained
- THEN its last-work-activity timestamp and ordering remain unchanged.

### Requirement: Reversible copy archiving

Archival SHALL affect visibility only, preserving work content, source identity, lifecycle and pin. Manual archive/restore SHALL remain scoped, authenticated, atomic, version-conditional and persistent. Autoarchive SHALL apply only to explicitly closed, unpinned works at least two hours after their current closure episode and SHALL run at startup and within a 60-second runtime sweep. It SHALL NOT infer completion from inactivity or archive active, paused, review, validation or failed works. Background archival SHALL NOT produce a toast per work; existing temporary Undo for manual archival SHALL remain work-specific.

A manual restore SHALL durably exempt that closure episode. Pin SHALL prevent autoarchive; unpin SHALL reevaluate the original deadline without clearing a restore exemption. Explicit reopen SHALL clear the prior episode/exemption and atomically restore automatically archived work, while preserving a manual archive. A subsequent explicit close SHALL start a new episode and deadline.

#### Scenario: Deadline and restore
- WHEN a closed eligible work reaches 7,200 seconds and a sweep succeeds
- THEN it becomes archived without losing content; a subsequent manual restore keeps it visible through later sweeps and restarts until a new closure.

#### Scenario: Concurrent user action
- WHEN a sweep races a versioned restore, pin or reopen
- THEN storage serializes them, retains the winning current intent and returns a safe conflict for stale requests rather than overwriting the newer state.

### Requirement: Manual viewer retention

API content SHALL NOT be purged due to age. Automatic two-hour archival SHALL change visibility only, and pin SHALL remain unnecessary for content preservation. Explicit deletion SHALL remain the only removal of completed work content. No configurable age/TTL deletion option SHALL be introduced.

#### Scenario: Archived content remains available
- WHEN an automatically archived, unpinned work ages beyond fourteen days and the server restarts
- THEN its documents and images remain readable until an explicit authorized deletion.

### Requirement: Compatible opt-in navigation projection

The authenticated revision-2 index SHALL expose version-1 navigation metadata only when navigation=grouped is requested. Requests without that projection SHALL retain the revision-2 canonical grouping shape. The projection SHALL reference each indexed work exactly once, preserve stable namespaced grouping keys and original-root origin metadata, and observe each source-root configuration at most once per index. The revision-2 frontend SHALL fall back to canonical groups with a diagnostic when optional projection metadata is invalid, without weakening session validation. Cross-revision client/server operation SHALL be replaced by the coordinated revision gate; schema/storage version 3 identities SHALL remain unchanged.

#### Scenario: Matching revisions with invalid optional grouping
- WHEN matching revision-2 clients and server exchange an index with invalid optional navigation metadata
- THEN the frontend renders canonical groups and a diagnostic without losing or weakening validation of works.

#### Scenario: Mixed client and server generations
- WHEN the client and server timeline revisions differ or the required revision is missing
- THEN timeline/index access requests update or reload and no timeline mutation is accepted.

### Requirement: On-demand high-level publication

The manager SHALL provide the existing additive hive view work semantic interface, reading explicitly bounded local sources, freezing selected Markdown and publishing through authenticated protocol v3 operations at matching timeline contract revision 2. It SHALL return compact confirmed results without callers authoring transport IDs, versions, hashes or request bodies. Existing low-level operation meanings, native payload identity and Markdown rendering SHALL be preserved after the coordinated update; operation against unmodified older server revisions SHALL not be promised. Publication SHALL NOT start/restart the viewer or open a browser.

#### Scenario: Publish and revise a plan
- WHEN a matching-revision caller publishes a valid plan folder and later its edited contents to the same work
- THEN two immutable revisions retain original bytes and the CLI confirms their revision IDs.

#### Scenario: Equal content and context
- WHEN current remote content and semantic publication context match the requested set
- THEN the command returns unchanged without remote mutation.

### Requirement: Bounded publication interface and compatibility

Successful mutation responses SHALL fit 4096 UTF-8 bytes without full source documents or credentials. The work list SHALL retain bounded pages and honest missing-binding state. High- and low-level clients SHALL share protocol v3 operation meanings at matching timeline contract revision 2; a coordinated update SHALL replace the prior requirement to operate against an unmodified older server. Publisher files SHALL remain outside viewer storage. The compact contract and canonical procedure SHALL preserve publication timing, producer authority, human validation and evidence checks while the binary owns bookkeeping. Native pending-operation identities/bytes and retained publication receipts SHALL survive the update for explicit resume.

#### Scenario: Coordinated server and clients
- WHEN high- and low-level clients use the matching upgraded server
- THEN they operate on preserved work histories, keep publisher journals outside viewer storage and retain explicit interfaces for unsupported advanced operations.

### Requirement: Local viewer management actor

The actor union SHALL keep accepted native_producer separate from local_viewer. Browser rename, archive/restore, pin/unpin and confirmed deletion SHALL remain authenticated, scoped and available without native identity, and SHALL retain actual local_viewer provenance rather than native publication or human-approval claims. Revision-2 archive/restore/pin-only operations MAY use bounded private administrative state/receipt slots instead of adding public local_action events or publication receipts, including on saturated legacy records. Exact retained-slot retry SHALL replay; older stale requests SHALL conflict without mutation. Existing public history and native publication receipts SHALL NOT be pruned. local_viewer SHALL NOT publish content/evidence/approval or change stage lifecycle; explicit recover_lead and confirmed deletion SHALL retain their existing authority and durable tombstone/receipt guarantees. Mixed actor branches SHALL fail.

#### Scenario: Restore at public capacity
- WHEN autoarchive hides a valid record whose public events or publication receipts are full
- THEN an authenticated versioned Restore commits its exemption with actual local_viewer provenance, exact retry replays and restart preserves visibility, without deleting history.

#### Scenario: Confirmed browser deletion
- WHEN the user confirms deletion through the existing scoped browser flow
- THEN deletion retains its durable operation result without native impersonation or content resurrection.

## ADDED Requirements

### Requirement: Durable closure-episode maintenance

The store SHALL maintain bounded private policy metadata and truthful system provenance using its existing serialized durable-write/reconciliation path, without fabricating native or local_viewer actors or consuming an unbounded public event/receipt stream. Both automatic archival and its manual reversal SHALL have bounded capacity independent of saturated public event/publication-receipt budgets. At most 16 KiB additional private per-record maintenance capacity SHALL be reserved, inaccessible to content, including for valid legacy records already at former serialized limits. Public derived fields SHALL project canonical metadata rather than independently persist editable copies. Physical write failure or numeric version exhaustion SHALL leave the work visible and report maintenance failure instead of hiding work without safe reversal. Successful archive SHALL bump the optimistic work version once; repeated sweeps SHALL be no-ops. Uncertain persistence SHALL NOT be reported confirmed and SHALL reconcile before further maintenance.

Existing records SHALL recover the current explicit closure time from server-recorded transitions or closure receipts after any reopen. An unprovable time SHALL remain unknown and receive a full two-hour window from the first durably recorded eligibility evaluation; restarts SHALL NOT extend that window. Prior manual restore/pin choices SHALL be honored. Malformed records SHALL fail explicitly without destructive defaulting.

#### Scenario: Unknown old closure date
- WHEN an existing closed work has no provable closure timestamp
- THEN closed_at remains null, eligibility starts once on durable initialization, and restarts retain that same deadline.

### Requirement: Coordinated timeline contract revision

The upgraded server, CLI and UI SHALL require timeline contract revision 2 while preserving schema/storage version 3 identities and all existing evidence. Matching revision SHALL be acknowledged before upgraded clients write; old or missing revisions SHALL receive a safe actionable update/reload error before timeline mutations. No simultaneous old-client projection SHALL be required. Bootstrap/authentication and existing file-capability reads SHALL retain their access protections.

The shared session/index projection SHALL expose required last_activity_at, nullable closed_at and autoarchive_at, and nullable archive_origin limited to manual/automatic. Stored-read normalization SHALL accept valid pre-policy records without weakening strict public validation. Contract mismatch SHALL preserve publisher journal inputs, operation identity and stored receipts; it SHALL NOT trigger automatic replay, request replacement or deletion.

#### Scenario: Old tab or CLI
- WHEN an outdated client attempts a timeline mutation
- THEN the server requests update/reload before writing, and the pending operation/input remains available for explicit recovery with the upgraded client.

#### Scenario: Upgraded client against old server
- WHEN the server does not acknowledge the new contract revision
- THEN the upgraded client refuses mutation before sending a write.
