## ADDED Requirements

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
