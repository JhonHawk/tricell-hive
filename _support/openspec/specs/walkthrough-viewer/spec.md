# walkthrough-viewer

## Purpose

A local server, `hive view`, that the user starts and stops to keep and browse the UI walkthroughs agents write, grouped by project, ticket and session.

## Requirements

### Requirement: User-controlled walkthrough server
The manager SHALL provide `hive view start`, `stop`, `status`, `open`, `serve` and `register`. The server SHALL bind only to 127.0.0.1, SHALL run until the user stops it, and SHALL never be started by `register`, `status` or `open`.

#### Scenario: Second start
- **WHEN** a live server is recorded and the user runs `hive view start`
- **THEN** the command reports the running server and starts no second process.

#### Scenario: Workspace root that is not a Git repository
- **WHEN** `hive view register --dir` names a folder inside `<root>/_support/workspace/` and `<root>` is not a Git repository
- **THEN** the registry records `<root>`, and its walkthroughs are grouped under the project named by the root's `## Hive` `Project:` value.

#### Scenario: Idempotent registration
- **WHEN** a repository is already registered and `hive view register` runs in it
- **THEN** the command exits successfully without writing the registry.

### Requirement: Walkthrough ingestion and grouping
The server SHALL copy valid `walkthrough.json` files found under `_support/workspace/*/` of registered roots, with their referenced files, into its own store and group copies by project, ticket and session. It SHALL NOT ingest a manifest confirmed as tracked by a complete check of its owning Git index. Tracking SHALL be checked before unchanged-source shortcuts, including ancestor repositories and worktrees. A confirmed non-Git workspace SHALL remain supported. An unavailable or incomplete tracking check SHALL produce unknown status and a warning, never confirmed untracked status; ingestion and access in that unknown state SHALL NOT claim tracking exclusion or verified authorship. The project key SHALL be the normalized Git remote, or the repository root without one; the project label SHALL come from the root's `## Hive` `Project:` value, then `.engram/config.json` `project_name`, then the remote's last segment or the folder name.

#### Scenario: Invalid manifest
- **WHEN** a manifest has an unknown field or a path outside its folder
- **THEN** it is not copied, the error is listed in the index, and the server keeps serving.

#### Scenario: Force-added manifest
- **WHEN** a complete owning-index check confirms that a valid manifest is tracked even under an ignored workspace directory
- **THEN** no new stored copy is created from it.

### Requirement: Store retention and deletion
The server SHALL purge copies older than the configured age, 14 days by default, except pinned ones, and SHALL delete a copy on a request carrying its token and its own origin. Deleting SHALL never modify files in a repository. Every API read SHALL require the server token, and copied files SHALL be served only under a per-session capability derived from it.

#### Scenario: Read without token
- **WHEN** a request for the index, a session or a copied file lacks the token or the matching capability
- **THEN** the server answers 403 and returns no data.

#### Scenario: Delete without token
- **WHEN** a delete request lacks the server token or carries a foreign `Origin` or `Host`
- **THEN** the server answers 403 and the copy remains.

### Requirement: Single-use viewer access handoff
The CLI SHALL obtain a cryptographically random access code through a token-authenticated same-origin POST and pass only that code, never the server token, in the system opener URL fragment. Codes SHALL expire after 60 seconds, SHALL be consumed atomically once, and SHALL exist only in server memory. The exchange POST SHALL require a valid code and the server's own Host and Origin instead of the server token; this exception SHALL NOT relax authentication on any other existing API route. Exchange responses SHALL be non-cacheable. Tokens and codes SHALL NOT appear in CLI output, errors or logs.

#### Scenario: Open and reload
- **WHEN** the user opens the running viewer through the CLI
- **THEN** the page removes the code fragment before awaiting its exchange, receives the server token, and preserves authenticated access across a normal tab reload.

#### Scenario: Reuse or expiry
- **WHEN** a code is exchanged twice, concurrently, or at least 60 seconds after issuance
- **THEN** at most one exchange before expiry succeeds and all others grant no access.

#### Scenario: Invalid origin
- **WHEN** exchange has an absent, null or foreign Origin, or a disallowed Host
- **THEN** it returns 403 and grants no access.

#### Scenario: Overlapping opens
- **WHEN** an older exchange finishes after a newer open begins
- **THEN** the older result does not overwrite the newer access state; a failed current exchange offers reopening through the CLI.

### Requirement: Viewer command group help
The manager SHALL display the six viewer subcommands and exit successfully for `hive view --help`, `hive view -h` and `hive view help`, without resolving or writing server state or launching processes.

#### Scenario: Help without a running server
- **WHEN** the user requests viewer group help without a server or state directory
- **THEN** the command lists start, stop, status, open, register and serve and exits successfully without creating state.

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
