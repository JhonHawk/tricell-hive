## ADDED Requirements

### Requirement: Local management actions from the CLI
The manager SHALL offer `hive view work rename`, `archive`, `restore`, `pin`, `unpin`, `recover` and `delete`, each with the same effect, expected version and result as the browser action it mirrors, sent as the `local_viewer` actor with `origin: cli` and without a native producer identity. Only `recover` SHALL take a reason. `delete` SHALL send nothing unless `--confirm` equals the work ID. `recover` SHALL accept the named pending same-project request and make it lead in one commit without inferring that the previous lead's session ended. `/api/health` SHALL announce the capability `local_actions_cli`; when the running viewer does not announce it, the CLI SHALL send nothing and ask to update the viewer. The native publication commands SHALL keep refusing a `local_viewer` actor.

#### Scenario: Delete without confirmation
- **WHEN** `hive view work delete --work-id W` runs without `--confirm W`
- **THEN** nothing is sent and the work is unchanged.

#### Scenario: Older running viewer
- **WHEN** the running viewer does not announce `local_actions_cli`
- **THEN** the CLI sends nothing and names the viewer update as the next step.

#### Scenario: Browser unchanged
- **WHEN** the browser sends a local action without `origin`
- **THEN** it is accepted as before and recorded with `origin: browser`.

## MODIFIED Requirements

### Requirement: Explicit session mutation scope
Browser and CLI title/archive/pin/deletion and recovery SHALL require existing authentication, explicit session action scope and the v3 local_viewer actor, and browser requests SHALL also pass same-origin/Host checks. Obsolete clients SHALL receive a safe reload/update rejection before mutation; the UI SHALL not replay a rejected write automatically. Native publisher operations SHALL follow the separately typed producer authority rules, not claim a local_viewer actor.

#### Scenario: Old tab
- WHEN an old browser tab sends an obsolete management write
- THEN v3 rejects it without changing work and offers reload instead of automatic replay.

### Requirement: Local viewer management actor

The actor union SHALL keep accepted native_producer separate from local_viewer. Rename, archive/restore, pin/unpin and confirmed deletion, from the browser or the CLI, SHALL remain authenticated, scoped and available without native identity, and SHALL retain actual local_viewer provenance rather than native publication or human-approval claims. The local_viewer actor SHALL accept an optional declared `origin` of `browser` or `cli`, default `browser`; it is declared provenance, not authentication, and SHALL be kept in each action's public event, administrative receipt or private deletion audit. Revision-2 archive/restore/pin-only operations MAY use bounded private administrative state/receipt slots instead of adding public local_action events or publication receipts, including on saturated legacy records. Exact retained-slot retry SHALL replay; older stale requests SHALL conflict without mutation. Existing public history and native publication receipts SHALL NOT be pruned. local_viewer SHALL NOT publish content/evidence/approval or change stage lifecycle; explicit recover_lead and confirmed deletion SHALL retain their existing authority and durable tombstone/receipt guarantees. Mixed actor branches SHALL fail.

#### Scenario: Restore at public capacity
- WHEN autoarchive hides a valid record whose public events or publication receipts are full
- THEN an authenticated versioned Restore commits its exemption with actual local_viewer provenance, exact retry replays and restart preserves visibility, without deleting history.

#### Scenario: Confirmed browser deletion
- WHEN the user confirms deletion through the existing scoped browser flow
- THEN deletion retains its durable operation result without native impersonation or content resurrection.

### Requirement: Shared work producers and lead authority
Protocol v3 SHALL preserve the original source tuple, its work key and the opaque viewer_session_id as work identity, accepting additional real native producers only by explicit join request and lead acceptance within the same registered project. A request SHALL NOT confer membership. Native lead transfer SHALL be version-conditional to an accepted member. Explicit local recovery SHALL atomically accept a chosen pending same-project request and reassign lead under expected version, durable operation receipt, reason and audit. Recovery SHALL NOT infer process death, stop processes, reopen terminal works, change original source or alter stage/evidence state. The previous lead SHALL remain a collaborator without lead privileges. Creation with the same tuple and work key SHALL remain idempotent. Joining SHALL NOT merge historical works or reassign IDs.

Pending, foreign-project and unauthorized publishers SHALL fail unchanged. Producers SHALL use their actual native IDs, never another session's or invented IDs. Provenance SHALL be declared rather than represented as cryptographic host authentication; local capability remains the authentication boundary. No historical actor or content SHALL be synthesized.

#### Scenario: Continue elsewhere
- WHEN the user authorizes continuation of an explicit work in another native session
- THEN the session joins under its own identity, the lead accepts/transfers authority, and future content stays in one timeline with correct attribution.

#### Scenario: Unauthorized transfer
- WHEN a pending member or stale transfer attempts control
- THEN authority and content stay unchanged.

#### Scenario: Original lead unavailable
- WHEN the user selects a pending same-project request in the authenticated local viewer and confirms Reassign with a reason
- THEN that producer is accepted and becomes lead in one durable commit, with immutable audit and no inferred abandonment or changed evidence.
- AND a stale competing recovery fails, and exact replay after restart returns the original result.

#### Scenario: Lead recovered from the CLI
- WHEN the user asks and `hive view work recover --work-id W --request-id R --reason T` runs for a pending same-project request
- THEN the result equals the browser recovery, with the audit recording `local_viewer`, `origin: cli` and reason T.

### Requirement: Durable client publication recovery
The manager SHALL persist versioned, permission-restricted publisher bookkeeping outside the server-owned viewer inventory. Every attempted mutating request SHALL have its exact bytes and operation identity durably recorded before transmission. Uncertain outcomes SHALL replay unchanged requests before accepting replacement work. Successful publication SHALL require exact revision/document readback; partial uploads SHALL not be described as a published revision. Unconfirmed inputs SHALL survive process exit and SHALL not expire by age. The local management commands (`hive view work rename`, `archive`, `restore`, `pin`, `unpin`, `recover` and `delete`) are exempt from this bookkeeping: after an uncertain outcome they SHALL re-read the work and report whether the action holds, and their expected version or a deleted work SHALL keep a retry from applying twice.

#### Scenario: Response lost after server commit
- WHEN the CLI exits after a committed snapshot but before recording its acknowledgment
- THEN resume replays the original operation and confirms the same revision without a duplicate.

#### Scenario: Competing publisher or external change
- WHEN a second publisher holds the client lock, or another actor changes the selected stage/chain incompatibly
- THEN the operation reports busy/conflict without silently overwriting, closing or reopening work.

#### Scenario: Recover from a definitive rejection
- WHEN the server has definitively rejected the exact pending request, every earlier attempted outcome is resolved, and an authorized caller explicitly reconciles that pending intent with a reason
- THEN the client atomically archives its bytes/receipts/rejection evidence locally, leaves remote content untouched and permits a new explicit publication against the current chain.
- WHEN any attempted outcome remains unknown
- THEN archival is refused and exact replay remains the recovery path.
- WHEN the work was closed or deleted
- THEN reconciliation never reopens or recreates it.
