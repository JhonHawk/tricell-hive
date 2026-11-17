## ADDED Requirements

### Requirement: Bounded E2E coordination through existing flows

Hive SHALL offer explicitly activated E2E coordination through its existing flows and one canonical coordination procedure. Its activation SHALL bind an objective, acceptance criteria, delivery endpoint, allowed effects, isolation authorization and stopping conditions. It SHALL NOT infer implementation authority from a research or planning request. It SHALL continue already authorized stage transitions without requiring a separate user command at each stage. Scope changes and effects outside the grant SHALL retain their existing user decision. Unattended status SHALL remain separate from E2E activation.

#### Scenario: Authorized multiunit objective
- **WHEN** an E2E grant and ready plan cover independent deliverables
- **THEN** the coordinator assigns owners, selects required skills, arranges independent verification and continues toward the authorized endpoint without asking the user which tool or stage comes next.

#### Scenario: Missing isolation authority
- **WHEN** the proposed coordination needs worktrees but the activation did not cover their creation
- **THEN** the coordinator resolves that decision before creation, and the existing ordinary-worktree rule remains in force.

### Requirement: Local deliverable owner exception

Only a designated E2E local-owner role SHALL be allowed to create local commits within its assigned repository, worktree, branch and write paths, using the existing scoped-commit and secret-scan mechanism. This named exception SHALL NOT expand ordinary implementation roles. The coordinator SHALL create and assign worktrees, arrange exclusive resources, launch independent verifiers and retain all publication, pull-request, tracker and integration operations. Owners SHALL NOT launch children, publish, integrate remotely, alter another assignment or infer authority from record fields.

#### Scenario: Local handoff
- **WHEN** an owner finishes an assigned unit
- **THEN** it returns the exact local head, command results, evidence and limitations, stops its own processes, and waits for the coordinator's next assignment rather than publishing its branch.

### Requirement: Root-owned execution record

Multiunit execution SHALL have one versioned local execution record linked from the existing checkpoint. The coordinator SHALL be its only writer. Record operations SHALL reject unknown versions, duplicate keys/IDs, invalid task/criterion/dependency references, cycles, conflicting active assignments and stale generations without losing the prior valid state. Machine paths, native IDs and raw run evidence SHALL remain outside versioned change records. Existing progress/checkpoint commands SHALL remain compatible.

#### Scenario: Stale update
- **WHEN** an update names a generation that no longer matches the record
- **THEN** the operation fails with a correction naming the stale state and preserves the current record.

### Requirement: Exact-head handoff and independent evidence

Before a handoff may be recorded verified, the helper SHALL check local Git identity, assigned branch/worktree, full head, ancestry, cleanliness and allowed changed paths. Verification metadata SHALL identify the checked head, criteria, commands, results, available evidence and a verifier distinct from the author. Missing, failed, unavailable or stale evidence SHALL prevent verified status. The coordinator SHALL inspect evidence; structurally valid declarations SHALL NOT be treated as proof of execution or semantic acceptance. Changed code SHALL invalidate affected verdicts.

#### Scenario: Reused verdict after a fix
- **WHEN** an owner returns a new head with only an older head's verdict
- **THEN** the verified transition is rejected and the coordinator obtains evidence for the new head before continuing.

### Requirement: Recoverable replacement without overlapping writers

Replacement SHALL preserve the unit's work and history, require observable completion or stop acknowledgment for the former writer, reconcile Git and resources, and create a new assignment before another writer begins. A missing native ID or elapsed time alone SHALL NOT prove the old writer stopped. Active or ownership-uncertain branches, worktrees and resources SHALL remain protected from cleanup. This contract SHALL NOT promise background scheduling or persistence beyond the host's verified capabilities.

#### Scenario: Partial return
- **WHEN** an owner returns partial work and native completion is confirmed
- **THEN** the coordinator assigns a replacement from the recorded head, limitations and next action without asking the human to reconstruct the session.

### Requirement: Project verification contract and feature map

Hive SHALL generate and maintain a project-owned verification contract indexing existing Launch, Doctor, Drive, Evidence and Cleanup capabilities and the features their checks cover. Initialization SHALL refuse to overwrite. Validation SHALL check versions, required fields, feature references, referenced files and confined paths, without executing commands. Unknown or untested capabilities SHALL remain explicitly unverified; unsupported and non-applicable entries SHALL state reasons. The runtime flow SHALL inspect effects and actual outputs before accepting or reusing evidence, and SHALL avoid creating a competing contract when the project already has an equivalent source.

#### Scenario: Valid shape, unverified application
- **WHEN** a project contract passes structural validation but Doctor or a feature's Drive has not passed
- **THEN** the flow reports that behavior unverified and does not infer readiness from the validator's exit status.

### Requirement: Deterministic correction and suppression review

For a repeatable error addressed by the work, the correction procedure SHALL select the earliest suitable enforcement layer among architecture, types, actionable lint, tests and prose. It SHALL retain one scoped rule-to-enforcer-to-negative-check mapping and remove procedural prose replaced by the mechanism. Flow-verify SHALL require an explicit inventory of added or widened suppressions, skipped tests and configuration exclusions in the diff, and a justified disposition of each candidate, with any selected review mechanism including native review. This procedure SHALL have one canonical home, referenced by review roles and supplied in native review requirements. Missing review coverage SHALL remain unverified. A textual match alone SHALL NOT be a defect; justified exceptions and authorized workaround explanations SHALL remain permitted.

#### Scenario: Suppression hides the failing criterion
- **WHEN** a proposed fix suppresses the check that exposes the defect without fixing its cause
- **THEN** review demonstrates the remaining failure and returns the unit for correction rather than accepting green output alone.

#### Scenario: Legitimate exception
- **WHEN** an exception has a demonstrated scope and justification consistent with requirements
- **THEN** review does not reject it solely because it uses a suppression directive or explanatory comment.

### Requirement: Local integration of verified owner heads

The coordinator SHALL combine independently verified unit heads into the declared integration target, read back the resulting Git head and run the check spanning the units before marking them integrated. The execution record SHALL distinguish local integration from remote publication. The bounded pilot SHALL exercise local integration without a push, pull request or remote merge. Existing publication rules SHALL continue to govern actual delivery.

#### Scenario: Two individually verified units
- **WHEN** both unit heads have valid independent verdicts
- **THEN** the coordinator still checks their combined local result before recording integrated, and the record does not claim remote delivery from that local result.
