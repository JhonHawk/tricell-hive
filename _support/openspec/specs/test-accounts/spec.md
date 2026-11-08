# test-accounts

## Purpose

The project's index of test accounts: reuse an account that fits before creating one, create only where the environment allows it, and keep the index rows in memory at close.

## Requirements

### Requirement: Account index before signing in
Before a check through the real path that signs in or uses records the environment already holds, in any flow or outside one, the agent SHALL read the project's account index, `ACCOUNTS.md` in the project's declared untracked secret location (`_support/secrets/` unless the project declares another; in a worktree, the main checkout's file, which is also where it writes; when the host refuses the edit, it appends a new row through a shell command and names a change to an existing row in the final report, as the local-instructions capability sets for an entry), rebuilding the file from the persistent-memory entry `accounts/index` when it is missing, and SHALL reuse an account whose environment, role and reach fit the check. Each row SHALL name the environment, role, identifier, secret source (a file in the secret location or a password-manager reference), what the account reaches and its status with the date it was last verified, and SHALL NOT contain a secret value. A child SHALL receive the row and its secret source, never the value, and SHALL report the accounts it created, found stale or used that the index lacks instead of writing the index.

#### Scenario: An account that fits already exists
- **WHEN** a check needs a signed-in user and the index lists an account of that environment and role whose reach covers the data under test
- **THEN** the agent uses that account and creates none.

#### Scenario: An indexed account fails
- **WHEN** an indexed account's sign-in fails
- **THEN** the agent marks its row stale with the date and cause before choosing or creating another account.

#### Scenario: A worktree session creates a local account
- **WHEN** a session isolated in a Claude Code worktree creates a local account and its edit of the main checkout's `ACCOUNTS.md` is refused
- **THEN** the agent appends the row to the main checkout's index with a shell command and writes no index inside the worktree.

### Requirement: Creation by environment
In a local environment, one whose database and services run on this machine, the agent SHALL create only the accounts and records the check needs that the index and local data do not provide, and SHALL add each account it creates or verifies to the index. In QA, a shared or cloud development database, and deployed environments the agent SHALL NOT create accounts, records or mock data on its own initiative; it SHALL use what exists, add each existing account it uses to the index, and ask the user before creating anything.

#### Scenario: Local check without a fitting account
- **WHEN** a local check needs a role the index does not list
- **THEN** the agent creates that account locally, without asking, and adds its row with the secret source.

#### Scenario: QA check without a fitting account
- **WHEN** a QA or deployed check needs an account or record that does not exist
- **THEN** the agent asks the user before creating it and creates nothing in the meantime.

### Requirement: Index rows kept in memory at close
At a work item's close the agent SHALL copy the index's rows, without values, to the project's persistent-memory entry `accounts/index`.

#### Scenario: A worktree without the index
- **WHEN** a later session works in a worktree without `ACCOUNTS.md` and the main checkout's file is unavailable
- **THEN** the persistent-memory entry gives each account's existence and where its secret is stored.

#### Scenario: A work item closes after using accounts
- **WHEN** a work item that created or used accounts reaches its close
- **THEN** the persistent-memory entry `accounts/index` holds each index row's environment, role, identifier, secret source, reach and status, and no secret value.
