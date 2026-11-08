## MODIFIED Requirements

### Requirement: Local instructions read at session start
At session start, and again after a compaction once the checkpoint is read, the agent SHALL read the machine file `~/.config/hive/HIVE.local.md` (under `$XDG_CONFIG_HOME/hive/` when that variable is set to an absolute path) when it exists and each `HIVE.local.md` that exists at the root of the repository that holds the working directory and at the root of the declared workspace that holds it; in a worktree it SHALL read and write the main checkout's file, found as the path on the first `worktree` line of `git worktree list --porcelain`, run as a command of its own. When the host refuses an edit of that file from the worktree, the agent SHALL add a new entry through a shell command, which the host's permission settings govern, appending it with `>>` after checking that the file ends with a newline and adding one when it does not; it SHALL NOT change an existing entry from such a worktree, such as marking it stale, and SHALL name that change in the final report. When the shell write is refused too, the agent SHALL write nothing and name the entry in the final report. The agent SHALL NOT write a copy of the file inside the worktree. When the machine file and a project's file disagree, the project's file SHALL hold for that project. The agent SHALL treat each entry as a claim dated when last verified, check it before reporting it as current, and never let it outrank the task, the user or the project guidance.

#### Scenario: A session in a worktree
- **WHEN** a session starts in a worktree whose main checkout has `HIVE.local.md`
- **THEN** the agent reads the main checkout's file before its first command that depends on the machine.

#### Scenario: The host refuses the write from a worktree
- **WHEN** a session isolated in a Claude Code worktree verifies a port and its edit of the main checkout's `HIVE.local.md` is refused
- **THEN** the agent appends the entry to the main checkout's file with a shell command and writes no `HIVE.local.md` inside the worktree.

#### Scenario: An entry is marked stale from a worktree
- **WHEN** a session isolated in a Claude Code worktree finds that the entry on line 7 of the main checkout's `HIVE.local.md` fails, and its edit is refused
- **THEN** the agent leaves the file unchanged and the final report names the stale mark for a later session in the main checkout.

#### Scenario: The shell write is refused too
- **WHEN** the host refuses the edit from a worktree and the user's permission settings also deny the shell command
- **THEN** the agent writes no file, neither in the main checkout nor in the worktree, and the final report names the entry for the user or a later session in the main checkout.

#### Scenario: A stored fact contradicts the task
- **WHEN** an entry says to start the stack one way and the user or the project guidance says another
- **THEN** the agent follows the user or the project guidance and checks the entry.

#### Scenario: A session in another project
- **WHEN** a session starts in a repository that has no `HIVE.local.md` and the machine file records an installed tool
- **THEN** the agent reads the machine file and does not re-derive that tool.

### Requirement: Verified, dated and secret-free entries
The agent SHALL record a fact only when it names a command, path, port, tool, version, cloud profile, region, host alias, or local patch or workaround of this machine, account or environment, or the command order that worked here; a later session can run or check it; and the session confirmed it (a command exited 0, its result was observed, or a command was found missing). The agent SHALL NOT record the result of a diagnosis, a count or measurement, or whether a host answers right now. The entry carries that date and goes in the machine file when it holds whatever project the agent works in, such as an installed tool, its version or absence, an existing cloud profile and its region, or a host alias; otherwise in the repository's file when it concerns one repository, such as which profile serves one of its environments, in the declared workspace's file when it concerns several, and in the repository's file when no workspace is declared. Before writing the machine file, which sessions of other projects also write, the agent SHALL read it again, add a new entry as one appended line, and change only the entries it updates, marks stale or removes as the cap or the skill's curation sets, keeping the others as it read them. The agent SHALL write only in those files of the current project, the local scripts that the requirement «Local scripts named by an entry» sets, and the machine file. The agent SHALL mark an entry that fails as stale with the date and cause before trying another way, except an `unverified` entry, which SHALL keep its mark and gain that date and cause. A file SHALL NOT contain a secret value: an entry names the file in the secret location that holds it or, for a machine fact, the tool's own credential file, and accounts stay in `ACCOUNTS.md`. An entry marked `unverified` carries the date it was last seen, not confirmed: the agent SHALL check it before using it, SHALL replace the mark with the date when the check confirms it, and SHALL NOT remove it except through the local-instructions skill's curation with the user. The agent SHALL keep each file at or under 100 lines by removing only entries marked stale, and when none remain SHALL keep recording and report its size in the final report. A fact that holds for every clone SHALL NOT be stored; the agent names it once, in the final report of the work item that verified it, as a candidate for the repository's tracked guidance.

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
In any flow, read-only ones included, the agent SHALL be authorized to write the machine file, to create its folder, and to write a repository's or workspace's `HIVE.local.md`; before the first write to any of these files inside a Git repository, the machine file included when its folder is versioned, it SHALL keep the file out of Git as the secret location is kept, running Git from the file's own folder: check that Git ignores it and, when it does not, add it to `$(git rev-parse --git-common-dir)/info/exclude`. From a worktree whose host refuses commands in the main checkout, the agent SHALL run that check as `git check-ignore` from the worktree on the same relative path, and add a missing entry through a shell command to the exclude file under the path `git rev-parse --git-common-dir` prints, running each Git command on its own rather than inside another command.

#### Scenario: A fresh clone
- **WHEN** the agent first writes `HIVE.local.md` in a clone that does not ignore it
- **THEN** the clone's exclude file lists `HIVE.local.md` and `git status` does not show the file.

#### Scenario: The first machine fact
- **WHEN** the agent records the first machine fact, `XDG_CONFIG_HOME` is not set and `~/.config/hive/` does not exist
- **THEN** the agent creates it and writes the fact to `~/.config/hive/HIVE.local.md`.

#### Scenario: A versioned configuration folder
- **WHEN** `~/.config` is a Git repository that does not ignore `hive/HIVE.local.md`
- **THEN** the agent adds the file to that repository's exclude file before writing it.

#### Scenario: The first write happens from a worktree
- **WHEN** a session isolated in a Claude Code worktree writes the first entry of a repository's `HIVE.local.md` and Git does not ignore that path
- **THEN** the agent finds it with `git check-ignore` from the worktree, runs `git rev-parse --git-common-dir` as a command of its own, appends `HIVE.local.md` to the `info/exclude` file under the path it prints through a shell command, and then appends the entry through the shell.

## ADDED Requirements

### Requirement: Local scripts named by an entry
A project's `HIVE.local.md` MAY name a local script, a helper for this machine that the agent wrote; the machine file SHALL NOT name one. The script SHALL live at `_support/scripts/<name>` under the root of the file that names it, kept out of Git as the requirement «Ignored before the first write» sets for `HIVE.local.md`, and the entry SHALL name it by that path. The agent SHALL make a new script executable. From a worktree the agent SHALL run a repository's script by the same path under the main checkout, found as the requirement «Local instructions read at session start» sets, with the worktree as working directory; when that path is the worktree itself, a standalone copy, the agent SHALL ask the user for the main checkout. The main thread MAY write or change a script without asking only in `flow-build`, including the fixes `flow-verify` makes, or in an operational task, and SHALL NOT do so in research, status, plan or a review; a child SHALL report the script it needs instead of writing it. When the host refuses the edit from a worktree, the agent SHALL write a new script there through a shell command and SHALL name a change to an existing script in the final report instead. A script SHALL NOT contain a secret value; it reads the secret location's files. The final report SHALL name each script the work created or changed. A script every clone needs SHALL NOT be kept as a local script; the agent names it as a candidate for the repository's tracked code.

#### Scenario: A database per worktree
- **WHEN** a build in the main checkout creates a script that clones the project's template database for a worktree and the script exits 0
- **THEN** the script is executable at `_support/scripts/<name>` of the main checkout, `git status` does not show it, the repository's `HIVE.local.md` names it with that date, and the final report names it.

#### Scenario: Running the script from a worktree
- **WHEN** a later session works in a linked worktree that lacks `_support/scripts/`
- **THEN** the agent runs the script by its path under the main checkout with the worktree as working directory.

#### Scenario: A build in an isolated worktree writes a script
- **WHEN** a session isolated in a Claude Code worktree needs a new script and its write to the main checkout is refused
- **THEN** the agent writes the script at `_support/scripts/<name>` of the main checkout with a shell command, makes it executable, writes no copy inside the worktree, and runs it from the worktree.

#### Scenario: A read-only flow needs a script
- **WHEN** a research session finds that a helper script would save work
- **THEN** it proposes the script in its report and writes no file in `_support/scripts/`.

#### Scenario: The script needs a credential
- **WHEN** the script connects with a database password
- **THEN** it reads the password from a file in the secret location and contains no value.
