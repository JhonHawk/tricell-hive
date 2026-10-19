## ADDED Requirements

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
