# walkthrough-viewer delta

## ADDED Requirements

### Requirement: On-demand high-level publication
The manager SHALL provide an additive `hive view work` interface that reads an explicitly bounded local source, freezes selected Markdown bytes and publishes through existing authenticated protocol v3 operations. It SHALL return compact confirmed results without requiring callers to author transport IDs, versions, hashes or request bodies. Existing low-level interfaces and the Markdown renderer SHALL remain compatible. Publication SHALL NOT start/restart a viewer or open a browser.

#### Scenario: Publish and revise a plan
- WHEN the caller publishes a valid plan folder and later publishes its edited contents to the same work
- THEN the viewer exposes two immutable revisions with the original bytes, and the CLI confirms the respective revision IDs.

#### Scenario: Equal content and context
- WHEN current remote content and semantic publication context match the requested set
- THEN the command returns `unchanged` without remote mutation.

### Requirement: Explicit work and directory bindings
The manager SHALL bind canonical source directories to explicit work keys/IDs and actual native producers, without treating cwd, title or content hash as work identity. Distinct works in one native session SHALL remain separate. Listing/showing SHALL use existing read APIs and SHALL NOT create, reopen or mutate works. A changed binding SHALL require explicit bind with a reason, matching project/authority and no pending publication. Closed/deleted works SHALL not reopen through discovery, publication or technical recovery.

#### Scenario: Two changes in one conversation
- WHEN one native session opens change A and change B with distinct keys and publishes updates to each
- THEN list returns both identities, updates remain in their own histories, and changing cwd changes neither binding.

#### Scenario: Mismatched folder or producer
- WHEN a caller supplies a different directory for a bound work or lacks accepted membership
- THEN mutation is rejected without selecting another work or impersonating its lead.

### Requirement: Safe bounded document collection
The manager SHALL validate all selected files and their containment/types/limits before any publication mutation. Default plan collection SHALL include required proposal/tasks, optional design and deterministically ordered spec deltas. Explicit Markdown selection SHALL support retained documents outside that default shape without implicit linked-file traversal. It SHALL reject escapes, symlink components beneath the canonical root, shared inodes, nonregular files, unstable reads, invalid UTF-8 and exceeded existing limits. No document SHALL be silently truncated or omitted to fit a limit. Frozen bytes SHALL not change when source files are later edited.

#### Scenario: Invalid member of a set
- WHEN any selected member fails preflight
- THEN no review snapshot is published and the error names the bounded cause and corrective action.

### Requirement: Durable client publication recovery
The manager SHALL persist versioned, permission-restricted publisher bookkeeping outside the server-owned viewer inventory. Every attempted mutating request SHALL have its exact bytes and operation identity durably recorded before transmission. Uncertain outcomes SHALL replay unchanged requests before accepting replacement work. Successful publication SHALL require exact revision/document readback; partial uploads SHALL not be described as a published revision. Unconfirmed inputs SHALL survive process exit and SHALL not expire by age.

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

### Requirement: Semantic task and lifecycle operations
The manager SHALL accept structured task keys, titles, orders and stage segments and resolve their internal identities/versions for registration and updates. Registration SHALL prevalidate the entire input, preserve partial-batch receipts and skip identical entries without resetting progress. Metadata changes SHALL be explicit. Markdown publication SHALL not infer state, approval, assignment or completion. Stage/task/closure authority and semantic preconditions SHALL remain those of protocol v3.

#### Scenario: Markdown refresh after progress
- WHEN a plan's tasks document is republished after a task starts
- THEN the document revision changes while the live task state remains unchanged.

#### Scenario: Compact transition and separate closure
- WHEN an authorized producer updates task T1 in build using semantic fields and later closes work A
- THEN the correct segment changes without requiring technical version fields, and work B in the same native session remains unaffected.

### Requirement: Bounded publication interface and compatibility
Successful mutation responses SHALL fit 4096 UTF-8 bytes and SHALL not include full source documents or credentials. The work list SHALL provide bounded pages and honest missing-binding state. The high-level client SHALL function against the existing v3 server without a new endpoint or server inventory migration. Its compact contract and canonical procedure SHALL preserve publication timing, producer authority, human validation and evidence checks while delegating protocol bookkeeping to the binary.

#### Scenario: Existing server and advanced client
- WHEN the high-level client and existing low-level commands share an unmodified v3 server
- THEN both operate on compatible work histories, publisher files remain outside viewer storage, and unsupported advanced operations retain their existing explicit interface.
