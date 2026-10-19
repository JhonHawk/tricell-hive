## MODIFIED Requirements

### Requirement: Walkthrough ingestion and grouping
The server SHALL copy valid `walkthrough.json` files found under `_support/workspace/*/` of registered roots, with their referenced files, into its own store and group copies by project, ticket and session. It SHALL NOT ingest a manifest confirmed as tracked by a complete check of its owning Git index. Tracking SHALL be checked before unchanged-source shortcuts, including ancestor repositories and worktrees. A confirmed non-Git workspace SHALL remain supported. An unavailable or incomplete tracking check SHALL produce unknown status and a warning, never confirmed untracked status; ingestion and access in that unknown state SHALL NOT claim tracking exclusion or verified authorship. The project key SHALL be the normalized Git remote, or the repository root without one; the project label SHALL come from the root's `## Hive` `Project:` value, then `.engram/config.json` `project_name`, then the remote's last segment or the folder name.

#### Scenario: Invalid manifest
- **WHEN** a manifest has an unknown field or a path outside its folder
- **THEN** it is not copied, the error is listed in the index, and the server keeps serving.

#### Scenario: Force-added manifest
- **WHEN** a complete owning-index check confirms that a valid manifest is tracked even under an ignored workspace directory
- **THEN** no new stored copy is created from it.

## ADDED Requirements

### Requirement: Current source classification
The server SHALL derive `tracked`, `local` or `unknown` source status for every stored copy before publishing its initial or reconstructed view and on each complete scan. It SHALL place tracked copies outside the normal project/ticket tree, preserve their files, keys, ingestion time, pins and existing retention behavior, and show an explanatory entry without rendering the document's instructions. Authenticated normal reads of a tracked copy's manifest or files SHALL return 409 without document bytes; existing authenticated pin and delete operations SHALL remain available. Missing sources without positive tracking evidence and failed classification SHALL be shown as unknown with a warning. Local status SHALL NOT claim verified agent authorship. The list and open document SHALL update source warnings when source status or reason changes, even when content hashes remain unchanged. Classification SHALL NOT require changes to stored per-copy metadata.

#### Scenario: Tracking changes without content changes
- **WHEN** an existing source becomes tracked or stops being tracked while its bytes remain unchanged and its file remains present
- **THEN** the next complete scan changes its placement without resetting its ingestion time or pin.

#### Scenario: Source missing or tracking unavailable
- **WHEN** the source is missing without positive tracking evidence or its tracking cannot be determined
- **THEN** its retained copy remains accessible with an explicit unknown-origin warning rather than a verified-authorship claim.

### Requirement: Safe and deterministic Git tracking lookup
The source classifier SHALL explicitly disable repository-configured `core.fsmonitor` callbacks for its Git read processes without changing stored Git configuration. Inherited pathspec environment variables SHALL NOT alter the classifier's selected pathspec interpretation. Index paths resolving to the same filesystem object as a manifest SHALL be recognized as tracked despite case differences; distinct files on case-sensitive filesystems SHALL NOT be conflated. Prefix selection SHALL account for case aliases in parent directories. These controls SHALL preserve the existing bounded-output and unknown-status policies.

#### Scenario: Repository filesystem monitor
- **WHEN** Git accepts a repository configured with a filesystem-monitor callback and the classifier reads its index
- **THEN** the callback is not executed and the index still determines tracking status.

#### Scenario: Inherited literal pathspec semantics
- **WHEN** the parent environment sets `GIT_LITERAL_PATHSPECS=1` and a manifest is tracked
- **THEN** the complete query still identifies that manifest as tracked rather than producing a false local result.

### Requirement: Open-file hardlink rejection
Before reading an ingestion manifest or referenced resource, including each hashing and copying read, the server SHALL inspect the opened file descriptor and reject a regular file with more than one hard link. It SHALL reject unavailable link-count metadata explicitly rather than claiming the check passed. This policy SHALL NOT require the file owner to equal the server's effective user. Rejection SHALL NOT install a partial copy or replace the last valid stored copy.

#### Scenario: Resource replaced after validation
- **WHEN** a validated resource is replaced with a hard-linked regular file before a hashing or copying open
- **THEN** descriptor inspection rejects it before reading its bytes and leaves the last valid stored copy intact.

### Requirement: Full-width desktop screen capture preview
On screen, a desktop capture preview in a strip wider than 520 CSS pixels SHALL occupy the strip's usable width. Narrow strips SHALL retain the existing next-capture peek, and additional captures SHALL remain horizontally accessible. Mobile capture sizing, carousel access and the existing print arrangement SHALL remain unchanged. The skill asset and embedded viewer stylesheet SHALL remain byte-identical. Previously generated standalone pages SHALL NOT be rewritten automatically.

#### Scenario: Wide screen strip
- **WHEN** a newly generated standalone walkthrough or the current viewer displays desktop captures in a wide strip
- **THEN** one capture occupies its usable width rather than half, without adding page-level horizontal overflow.
