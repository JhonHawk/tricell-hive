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
The server SHALL copy each valid `walkthrough.json` found under `_support/workspace/*/` of a registered repository, with the files it references, into its own store, and SHALL group copies by project, ticket and session. The project key SHALL be the normalized Git remote, or the repository root without one; the project label SHALL come from the `## Hive` `Project:` value, then `.engram/config.json` `project_name`, then the remote's last segment or the folder name.

#### Scenario: Invalid manifest
- **WHEN** a manifest has an unknown field or a path outside its folder
- **THEN** it is not copied, the error is listed in the index, and the server keeps serving.

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
