# walkthrough-viewer delta

Superseded on 2026-10-31 by the viewer-timeline-api change. Historical only; do not execute this contract or checkpoint.
## ADDED Requirements

### Requirement: Explicit screen revision evidence
Newly produced walkthroughs SHALL declare schema v2 with explicit logical-session-scoped screen identity, monotonically numbered publication rounds, per-screen revision history and immutable hashed capture provenance, including an explicit originating round. Counters SHALL be integers in 1..2147483647; revisions SHALL NOT exceed the enclosing publication and current-round timestamps SHALL agree. Capture origins SHALL NOT exceed their containing revision. The greatest unambiguous revision SHALL be current. Earlier revisions SHALL remain accessible under History and SHALL be labeled superseded. Updated and retained evidence SHALL be distinguished for the current publication round. The viewer SHALL NOT infer supersession from titles, ticket IDs, closure status, dates or equal file names.

#### Scenario: Focused correction
- WHEN R3 changes only the mobile filter and retains Profile and invitation
- THEN the current view labels Chat updated in R3, labels Profile/invitation retained from their original rounds, and shows old filter revisions only in History.

#### Scenario: Explicit continuation
- WHEN two source folders in one exact project/host/session identity contain revisions of the same explicit screen ID
- THEN the viewer presents one screen with its merged unambiguous revision history while preserving each resource owner's capability boundary.

#### Scenario: Conflict or incomplete loading
- WHEN equal publication/revision numbers disagree or a required source document has not loaded
- THEN the viewer shows loading or a scoped error rather than selecting a guessed current winner.

#### Scenario: Retained capture
- WHEN a new revision reuses an unaffected state capture
- THEN its original candidate/date/hash remain and the UI distinguishes it from newly captured evidence.

### Requirement: Exact active publication confirmation
The publisher SHALL support exact live-viewer confirmation for explicit in_progress and closed lifecycle states using the current manifest and referenced resource fingerprint. Active confirmation SHALL additionally require accessible, unambiguous current evidence across the relevant group; excluded copies, superseded target revisions and unresolved conflicts SHALL fail it. Existing closed-copy retention confirmation SHALL keep its exact-storage semantics. Server presentability SHALL NOT be described as observed browser rendering. Failed confirmation SHALL preserve the source evidence and SHALL NOT be reported as a successful visible update. Confirmation SHALL NOT start the server or enable an HTML fallback when a Hive executable exists.

#### Scenario: Stale live copy
- WHEN the source contains R4 but the viewer still stores R3
- THEN publication confirmation fails until the exact current source/resources/status are observed, regardless of a matching folder/key.


#### Scenario: Stored but not presentable
- WHEN the exact source copy is excluded or another relevant source supersedes/conflicts with its claimed current revision
- THEN active-publication confirmation fails even if stored_exact is true.

### Requirement: Proportional iterative UI verification
The initial verification SHALL cover the affected layout and states at the relevant supported desktop/mobile widths and themes. Subsequent rounds SHALL verify and recapture changed regions, affected neighboring layout, relevant states and responsive/theme-sensitive cases only. A navigation change SHALL exercise its affected sequence; a nonresponsive correction SHALL NOT trigger an unrelated full mobile walkthrough. Viewports SHALL share a screen walkthrough, and correction rounds SHALL revise it rather than add duplicate viewport walkthroughs. Unaffected evidence MAY be retained with explicit provenance. This procedure SHALL have one canonical home in flow-walkthrough; review and verification consumers SHALL follow it without a conflicting full-matrix iteration rule.

#### Scenario: Filter width correction
- WHEN an iteration changes the mobile filter width but not Profile, email or navigation
- THEN the producer checks the filter and resized neighbors at affected mobile widths, updates their evidence, retains unaffected evidence, and asks no per-viewport decision.

#### Scenario: Mobile navigation correction
- WHEN the change affects switching between list and conversation
- THEN verification walks the list, opens a conversation and returns with the relevant mobile state preserved.

## MODIFIED Requirements

### Requirement: Walkthrough ingestion and grouping
Existing logical session identity, authentication, tracking classification and resource-boundary requirements SHALL remain. New readers SHALL accept legacy v1 without migration or inferred revisions and SHALL accept bounded, strictly validated v2 history. Python and Go validators SHALL agree, including capture hash validation and retention of all history resources. V2 screen projection SHALL supplement logical session grouping rather than merge separate identities. The standalone HTML renderer SHALL show the same current/history and provenance behavior for the complete history present in its manifest. Old readers MAY reject v2; rollout SHALL update the binary before producing v2 content.

#### Scenario: Legacy walkthrough
- WHEN a v1 manifest has no revision metadata
- THEN new readers preserve its original rendering, identify revision information as unavailable and do not fabricate relationships.

#### Scenario: Standalone history
- WHEN Hive is unavailable and a self-contained v2 manifest has three screen revisions
- THEN its generated HTML defaults to the current one and exposes the previous two with original provenance and viewport selection.

### Requirement: Store retention and deletion
The existing supported logical-session deletion and tombstone behavior SHALL remain. The authorized historical cleanup SHALL use the fixed planning-time allowlist and an expected-members precondition comparing exact group member keys/content hashes inside the same store deletion lock before any mutation. A mismatch SHALL reject the entire group without tombstones/removal and return a bounded conflict response. It SHALL act only on that unchanged allowlist, preserve repository sources, registry, authentication and runtime settings, and leave this task's or subsequently produced evidence untouched. Unknown or concurrently changing groups SHALL remain for resolution. Partial deletion SHALL report surviving members and retain normal recovery behavior. Existing TTL/pin behavior outside the explicit cleanup SHALL remain.

#### Scenario: Authorized cleanup
- WHEN an allowlisted historical group is unchanged and its exact member keys/hashes match atomically
- THEN supported deletion removes its stored copies, including archived/pinned members, and tombstones prevent unchanged return without deleting any source files.

#### Scenario: New or changed evidence
- WHEN a group was created after the inventory or changed before cleanup
- THEN the cleanup leaves it untouched and reports why it was skipped.


#### Scenario: Ingestion races cleanup
- WHEN a member is added or changed after inventory but before the guarded deletion acquires the store lock
- THEN the entire group deletion is rejected without tombstones or removed members, preserving the new evidence.
