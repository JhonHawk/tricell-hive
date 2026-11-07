# local-instructions

## Purpose

The per-machine instructions file `HIVE.local.md`: agents read it at session start, record there the machine and environment facts they verify, and keep it untracked and free of secret values, so a later session does not re-derive them.

## ADDED Requirements

### Requirement: Local instructions read at session start
At session start, and again after a compaction once the checkpoint is read, the agent SHALL read each `HIVE.local.md` that exists at the root of the repository that holds the working directory and at the root of the declared workspace that holds it; in a worktree it SHALL read and write the main checkout's file. The agent SHALL treat each entry as a claim dated when last verified, check it before reporting it as current, and never let it outrank the task, the user or the project guidance.

#### Scenario: A session in a worktree
- **WHEN** a session starts in a worktree whose main checkout has `HIVE.local.md`
- **THEN** the agent reads the main checkout's file before its first command that depends on the machine.

#### Scenario: A stored fact contradicts the task
- **WHEN** an entry says to start the stack one way and the user or the project guidance says another
- **THEN** the agent follows the user or the project guidance and checks the entry.

### Requirement: Verified, dated and secret-free entries
The agent SHALL record a fact in `HIVE.local.md` when it names something of this machine, account or environment, such as a path, a port in use, an installed tool, its version or its absence, a cloud profile, region or host, a local patch or workaround, or the command order that worked here, and the session confirmed it (a command exited 0, its result was observed, or a command was found missing); the entry carries that date and goes in the repository's file when it concerns one repository, in the declared workspace's file when it concerns several or the whole machine, and in the repository's file when no workspace is declared. The agent SHALL write only in the current project. The agent SHALL mark an entry that fails as stale with the date and cause before trying another way. The file SHALL NOT contain a secret value: an entry names the file in the secret location that holds it, and accounts stay in `ACCOUNTS.md`. The agent SHALL keep the file at or under 100 lines by removing only entries marked stale, and when none remain SHALL keep recording and report its size in the final report. A fact that holds for every clone SHALL NOT be stored; the agent names it once, in the final report of the work item that verified it, as a candidate for the repository's tracked guidance.

#### Scenario: The local stack starts
- **WHEN** the agent starts the local stack on ports that differ from the documented ones and the health check passes
- **THEN** `HIVE.local.md` holds the commands and ports it used, with that date.

#### Scenario: A recorded command fails
- **WHEN** a command recorded in the file fails
- **THEN** the agent marks that entry stale with the date and cause before trying another way.

#### Scenario: A fact needs a credential
- **WHEN** the verified fact depends on a password or token
- **THEN** the entry names the file in the secret location that holds it and contains no value.

#### Scenario: The file is full of current entries
- **WHEN** the file has more than 100 lines and no entry is marked stale
- **THEN** the agent removes no current entry and reports the size in the final report.

### Requirement: Ignored before the first write
In any flow, read-only ones included, the agent SHALL be authorized to write `HIVE.local.md`; before its first write inside a Git repository it SHALL keep the file out of Git as the secret location is kept: check that Git ignores it and, when it does not, add it to `$(git rev-parse --git-common-dir)/info/exclude`.

#### Scenario: A fresh clone
- **WHEN** the agent first writes `HIVE.local.md` in a clone that does not ignore it
- **THEN** the clone's exclude file lists `HIVE.local.md` and `git status` does not show the file.

### Requirement: Children report instead of writing
A child SHALL read `HIVE.local.md` and SHALL NOT write it; it SHALL report each fact it verified or found stale, and the main thread SHALL update the file.

#### Scenario: A child starts the stack
- **WHEN** a delegated child starts the local stack with a command the file lacks
- **THEN** the child reports the command and its result, and the main thread records it.
