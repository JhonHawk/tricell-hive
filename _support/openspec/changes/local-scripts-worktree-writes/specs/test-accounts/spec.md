## MODIFIED Requirements

### Requirement: Account index before signing in
Before a check through the real path that signs in or uses records the environment already holds, in any flow or outside one, the agent SHALL read the project's account index, `ACCOUNTS.md` in the project's declared untracked secret location (`_support/secrets/` unless the project declares another; in a worktree, the main checkout's file, which is also where it writes, through a shell command when the host refuses the edit, as the local-instructions capability sets), rebuilding the file from the persistent-memory entry `accounts/index` when it is missing, and SHALL reuse an account whose environment, role and reach fit the check. Each row SHALL name the environment, role, identifier, secret source (a file in the secret location or a password-manager reference), what the account reaches and its status with the date it was last verified, and SHALL NOT contain a secret value. A child SHALL receive the row and its secret source, never the value, and SHALL report the accounts it created, found stale or used that the index lacks instead of writing the index.

#### Scenario: An account that fits already exists
- **WHEN** a check needs a signed-in user and the index lists an account of that environment and role whose reach covers the data under test
- **THEN** the agent uses that account and creates none.

#### Scenario: An indexed account fails
- **WHEN** an indexed account's sign-in fails
- **THEN** the agent marks its row stale with the date and cause before choosing or creating another account.

#### Scenario: A worktree session creates a local account
- **WHEN** a session isolated in a Claude Code worktree creates a local account and its edit of the main checkout's `ACCOUNTS.md` is refused
- **THEN** the agent appends the row to the main checkout's index with a shell command and writes no index inside the worktree.
