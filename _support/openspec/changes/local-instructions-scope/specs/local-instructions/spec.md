## MODIFIED Requirements

### Requirement: Local instructions read at session start
At session start, and again after a compaction once the checkpoint is read, the agent SHALL read the machine file `~/.config/hive/HIVE.local.md` (under `$XDG_CONFIG_HOME/hive/` when that variable is set) when it exists and each `HIVE.local.md` that exists at the root of the repository that holds the working directory and at the root of the declared workspace that holds it; in a worktree it SHALL read and write the main checkout's file. The agent SHALL treat each entry as a claim dated when last verified, check it before reporting it as current, and never let it outrank the task, the user or the project guidance.

#### Scenario: A session in a worktree
- **WHEN** a session starts in a worktree whose main checkout has `HIVE.local.md`
- **THEN** the agent reads the main checkout's file before its first command that depends on the machine.

#### Scenario: A stored fact contradicts the task
- **WHEN** an entry says to start the stack one way and the user or the project guidance says another
- **THEN** the agent follows the user or the project guidance and checks the entry.

#### Scenario: A session in another project
- **WHEN** a session starts in a repository that has no `HIVE.local.md` and the machine file records an installed tool
- **THEN** the agent reads the machine file and does not re-derive that tool.

### Requirement: Verified, dated and secret-free entries
The agent SHALL record a fact only when it names a command, path, port, tool, version, cloud profile, region, host alias, or local patch or workaround of this machine, account or environment, or the command order that worked here; a later session can run or check it; and the session confirmed it (a command exited 0, its result was observed, or a command was found missing). The agent SHALL NOT record the result of a diagnosis, a count or measurement, or whether a host answers right now. The entry carries that date and goes in the machine file when it holds whatever project the agent works in, such as an installed tool, its version or absence, an existing cloud profile and its region, or a host alias; otherwise in the repository's file when it concerns one repository, such as which profile serves one of its environments, in the declared workspace's file when it concerns several, and in the repository's file when no workspace is declared. Before writing the machine file, which sessions of other projects also write, the agent SHALL read it again and change only its own entry. The agent SHALL write only in those files of the current project and in the machine file. The agent SHALL mark an entry that fails as stale with the date and cause before trying another way, except an `unverified` entry, which SHALL keep its mark and gain that date and cause. A file SHALL NOT contain a secret value: an entry names the file in the secret location that holds it or, for a machine fact, the tool's own credential file, and accounts stay in `ACCOUNTS.md`. An entry marked `unverified` carries the date it was last seen, not confirmed: the agent SHALL check it before using it, SHALL replace the mark with the date when the check confirms it, and SHALL NOT remove it except through the local-instructions skill's curation with the user. The agent SHALL keep each file at or under 100 lines by removing only entries marked stale, and when none remain SHALL keep recording and report its size in the final report. A fact that holds for every clone SHALL NOT be stored; the agent names it once, in the final report of the work item that verified it, as a candidate for the repository's tracked guidance.

#### Scenario: The local stack starts
- **WHEN** the agent starts the local stack on ports that differ from the documented ones and the health check passes
- **THEN** the repository's `HIVE.local.md` holds the commands and ports it used, with that date.

#### Scenario: A tool is missing on the machine
- **WHEN** a command fails because `timeout` is not installed
- **THEN** the machine file records that `timeout` is missing, with the date and the replacement that worked.

#### Scenario: A project's cloud environment
- **WHEN** the session confirms that a repository's QA environment uses an existing profile in a region
- **THEN** the machine file records the profile and its region, and the repository's file records that QA uses that profile.

#### Scenario: Two sessions write the machine file
- **WHEN** a session in another project added an entry to the machine file after this session read it
- **THEN** this session reads the file again before writing and keeps that entry.

#### Scenario: A diagnosis or a momentary state
- **WHEN** the session finds a host unreachable, counts running processes, or diagnoses an operating-system fault
- **THEN** neither file records it.

#### Scenario: A recorded command fails
- **WHEN** a command recorded in a file fails
- **THEN** the agent marks that entry stale with the date and cause before trying another way.

#### Scenario: A kept unverified entry fails
- **WHEN** an entry marked `unverified` fails its check and the file is past 100 lines
- **THEN** the entry keeps its mark with the failure date and cause, is not marked stale, and stays in the file until the skill's curation with the user.

#### Scenario: A fact needs a credential
- **WHEN** the verified fact depends on a password or token
- **THEN** the entry names the file in the secret location that holds it and contains no value.

#### Scenario: The file is full of current entries
- **WHEN** a file has more than 100 lines and no entry is marked stale
- **THEN** the agent removes no current entry and reports the size in the final report.

### Requirement: Ignored before the first write
In any flow, read-only ones included, the agent SHALL be authorized to write the machine file, to create its folder, and to write a repository's or workspace's `HIVE.local.md`; before the first write to any of these files inside a Git repository, the machine file included when its folder is versioned, it SHALL keep the file out of Git as the secret location is kept: check that Git ignores it and, when it does not, add it to `$(git rev-parse --git-common-dir)/info/exclude`.

#### Scenario: A fresh clone
- **WHEN** the agent first writes `HIVE.local.md` in a clone that does not ignore it
- **THEN** the clone's exclude file lists `HIVE.local.md` and `git status` does not show the file.

#### Scenario: The first machine fact
- **WHEN** the agent records the first machine fact, `XDG_CONFIG_HOME` is not set and `~/.config/hive/` does not exist
- **THEN** the agent creates it and writes the fact to `~/.config/hive/HIVE.local.md`.

#### Scenario: A versioned configuration folder
- **WHEN** `~/.config` is a Git repository that does not ignore `hive/HIVE.local.md`
- **THEN** the agent adds the file to that repository's exclude file before writing it.

### Requirement: Children report instead of writing
A child SHALL read the machine file and the project's `HIVE.local.md` files and SHALL NOT write them; it SHALL report each fact it verified or found stale, and the main thread SHALL update the right file.

#### Scenario: A child starts the stack
- **WHEN** a delegated child starts the local stack with a command the file lacks
- **THEN** the child reports the command and its result, and the main thread records it.

### Requirement: Bootstrap and curation through the local-instructions skill
When the user asks to create, bootstrap, review or prune `HIVE.local.md`, the `local-instructions` skill SHALL gather candidates from persistent memory, host memory and local notes, the project's tracked guidance, the names of the secret location's files, and the skill's `extract_candidates.py` run over the last 30 days of this project's top-level Claude Code and Codex sessions, which SHALL print no secret value; SHALL discard diagnosis results, counts and whether a host answered; SHALL send the `missing` candidates and the existing cloud profiles with their regions to the machine file, and the `commands`, the `ports` and which profile serves an environment to the file their `where` names; SHALL ask the user which candidates to keep, group by group, before writing anything, keeping the answers in the files and not in persistent memory; SHALL confirm each kept fact with a read-only check from its fixed list when one applies and write the rest with its source date marked `unverified`; SHALL never run a project command that starts, stops, seeds, migrates, installs, signs in or writes; and SHALL report, without storing them, the facts that hold for every clone.

#### Scenario: First run on a machine with history
- **WHEN** the user asks to create `HIVE.local.md` and recent transcripts show a start command that exited 0, a port in use and a missing tool
- **THEN** the skill shows them as candidates, writes only those the user keeps, puts the kept start command and port in the repository's file and the kept missing tool in the machine file, and dates each by the read-only check that confirmed it or marks it `unverified` with its source date.

#### Scenario: A transcript command carries a credential
- **WHEN** a transcript command includes a password in an environment assignment, a URL or a header
- **THEN** the script's output, the question to the user and the file contain the command without that value.

#### Scenario: Curating a full file
- **WHEN** a file passes 100 lines and the user asks to prune it
- **THEN** the skill asks which entries to keep and removes only those the user discards.
