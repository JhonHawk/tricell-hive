# Test accounts

Read before a check that signs in or uses existing records. It applies in any flow or outside one.

- Read the project's account index, `ACCOUNTS.md` in the secret location «Secrets» names; in a worktree, read and write the main checkout's file (when the host refuses the edit, follow [worktree](local-write/worktree.md)). When the file is missing, rebuild it from the persistent-memory entry `accounts/index` when one exists, marking each row unverified until a sign-in confirms it. Reuse an account whose environment, role and reach fit the check.
- The index holds one row per account: environment, role, identifier, secret source (a file in the secret location or a password-manager reference, never a value), what it reaches, and its status with the date it was last verified working or failing; mark a row stale with date and cause when its sign-in fails.
- In a local environment, one whose database and services run on this machine, create only the accounts and records the check needs that the index and the local data do not provide, and add each account you create or verify to the index.
- In QA, a shared or cloud development database and deployed environments, create no accounts, records or mock data on your own initiative: use what exists, add each existing account you use to the index, and ask before creating anything, as «Scope and authorization» sets for shared data.
- A child does not read or write the index: it uses the rows and secret sources its brief gives, never a value, and reports each account it created, found stale or needed and did not get; the main thread updates the index.
- At a work item's close, copy the index's rows, without values, to persistent memory under `accounts/index`; this is the close write «Persistent memory» allows.
