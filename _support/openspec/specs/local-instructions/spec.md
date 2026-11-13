# local-instructions

## Purpose

The per-machine instructions files `HIVE.local.md`, one at a repository or workspace root for that project's facts and the machine file `~/.config/hive/HIVE.local.md` for facts that hold in any project: agents read them at session start, record there, one line each, what the machine or the environment has, uses or needs that a later session can run or check, keep the detail a fact needs in persistent memory behind a pointer, and keep the files untracked and free of secret values, so a later session does not re-derive them.

## Requirements

### Requirement: Local instructions read at session start
At session start, and again after a compaction once the checkpoint is read, the agent SHALL read the machine file `~/.config/hive/HIVE.local.md` (under `$XDG_CONFIG_HOME/hive/` when that variable is set to an absolute path) when it exists and each `HIVE.local.md` that exists at the root of the repository that holds the working directory and at the root of the declared workspace that holds it; in a worktree it SHALL read and write the main checkout's file, found as the path on the first `worktree` line of `git worktree list --porcelain`, run as a command of its own. When the host refuses an edit of that file from the worktree, the agent SHALL add a new entry through a shell command, which the host's permission settings govern, appending it with `>>`, first adding a newline when the file exists, is not empty and does not end with one; it SHALL NOT change or remove an existing entry from such a worktree, such as marking it stale or removing one past the cap, and SHALL name that change in the final report, which stands in for those steps until a session in the main checkout makes them. When the shell write is refused too, the agent SHALL write nothing and name the entry in the final report. The agent SHALL NOT write a copy of the file inside the worktree. When the machine file and a project's file disagree, the project's file SHALL hold for that project. The agent SHALL treat each entry as a claim dated when last verified, check it before reporting it as current, and never let it outrank the task, the user or the project guidance.

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
The agent SHALL record a fact only when it says what this machine, account or environment has, lacks, uses or needs, or a limitation of a tool here with the workaround that worked: a tool, an app, a container, a port, a host alias, a script, a cloud profile with its region, or the account or identity a tool uses, and what it serves; a later session can run or check it; and the session confirmed it (a command exited 0, its result was observed, or a command was found missing). An entry SHALL be one line of at most 200 characters, its pointer included, holding the fact and the date it was last verified, a stale mark and the date and cause a failing `unverified` entry gains not counting toward that length, and SHALL hold no sequence of steps, no account of what happened, and no version unless the project pins one this machine lacks. When the line cannot hold the command with its flags within that length, or reusing the fact takes two or more steps, the agent SHALL save that detail in persistent memory under a topic `local/<slug>`, named after the fact's subject and starting with the name of the folder that holds the file when that file sits below the folder that holds the project's Engram configuration, so files that share a project never share a topic; it SHALL search that project for the topic before saving, update it when it exists and create it otherwise, and SHALL end the line with the pointer to it. The detail of an entry of a project's file SHALL go to that project in persistent memory, and the detail of an entry of the machine file to the project `hive-machine`, which any session reads by that name and which the agent creates through Engram's command line when the memory tool rejects the name as unknown; when that fails too, the agent SHALL save the detail in the current project and name that project in the pointer. The agent SHALL read a detail only when the task needs its fact, and the main thread SHALL put in a child's brief the detail that child's task needs. The agent SHALL write the line alone when the machine's persistent memory is not Engram. When a detail fails, the agent SHALL mark its line stale and correct the topic once another way works, and before removing a stale line the agent SHALL write in its topic that it failed, with the date and cause, and SHALL leave the topic. The agent SHALL NOT record in these files the result of a diagnosis, a count or measurement, whether a host answers right now, an event, the state of a work item, such as its worktrees, its running processes or what its close kept, the app's design file or design library once the design record holds them, or a fact the repository's tracked files state. An instruction of the user about the machine or the project SHALL be saved in persistent memory as a decision, and the file SHALL keep only one line for what it constrains. The entry carries that date and goes in the machine file when it holds whatever project the agent works in, such as an installed tool or its absence, an existing cloud profile and its region, or a host alias; otherwise in the repository's file when it concerns one repository, such as which profile serves one of its environments, in the declared workspace's file when it concerns several, and in the repository's file when no workspace is declared. Before writing the machine file, which sessions of other projects also write, the agent SHALL read it again, add a new entry as one appended line, and change only the entries it updates, marks stale or removes as the cap or the skill's curation sets, keeping the others as it read them. The agent SHALL write only in those files of the current project, the local scripts that the requirement «Local scripts named by an entry» sets, and the machine file. The agent SHALL mark an entry that fails as stale with the date and cause before trying another way, except an `unverified` entry, which SHALL keep its mark and gain that date and cause. A file SHALL NOT contain a secret value: an entry names the file in the secret location that holds it or, for a machine fact, the tool's own credential file, and the test accounts of an environment stay in `ACCOUNTS.md`, while which account or identity a tool uses on this machine is a fact for these files, named without its credential. An entry marked `unverified` carries the date it was last seen, not confirmed: the agent SHALL check it before using it, SHALL replace the mark with the date when the check confirms it, and SHALL NOT remove it except through the local-instructions skill's curation with the user. The agent SHALL keep each file at or under 50 lines by removing only entries marked stale, and when none remain SHALL keep recording and report its size in the final report. A fact that holds for every clone SHALL NOT be stored; the agent names it once, in the final report of the work item that verified it, as a candidate for the repository's tracked guidance.

#### Scenario: The local stack starts
- **WHEN** the agent starts the local stack on ports that differ from the documented ones and the health check passes
- **THEN** the repository's `HIVE.local.md` holds one line with the ports it used and that date, ending with the pointer to the persistent-memory entry that holds the commands in the order that worked.

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
- **WHEN** an entry marked `unverified` fails its check and the file is past 50 lines
- **THEN** the entry keeps its mark with the failure date and cause, is not marked stale, and stays in the file until the skill's curation with the user.

#### Scenario: A fact needs a credential
- **WHEN** the verified fact depends on a password or token
- **THEN** the entry names the file in the secret location that holds it and contains no value.

#### Scenario: The file is full of current entries
- **WHEN** a file has more than 50 lines and no entry is marked stale
- **THEN** the agent removes no current entry and reports the size in the final report.

#### Scenario: An account that applies to a project
- **WHEN** the session confirms that this machine holds three Git accounts and the repository uses one of them
- **THEN** the repository's file holds one line naming the account the repository uses, with that date, no steps and no credential.

#### Scenario: A workaround with several steps
- **WHEN** the session finds that attaching a file to a ticket needs a REST call with a header and a credential file instead of the tracker's CLI
- **THEN** the file holds one line saying that attachments go through the REST API, ending with the pointer to the persistent-memory entry that holds the call
- **AND** that entry names the credential's file and holds no secret value.

#### Scenario: A limitation of a tool on this machine
- **WHEN** the session finds that a host's shell refuses a compound command in a worktree and that separate commands work
- **THEN** the machine file holds one line stating the limitation and the workaround
- **AND** when the workaround takes several steps, its detail is saved in the project `hive-machine` and the line ends with its pointer.

#### Scenario: An instruction of the user
- **WHEN** the user says a local container must be kept for the whole of an epic
- **THEN** the instruction is saved in persistent memory as a decision
- **AND** the file holds one line saying that the container this project uses must stay.

#### Scenario: The state of a work item
- **WHEN** the session creates worktrees, starts a demo on two ports or closes a ticket keeping its local configuration
- **THEN** neither file records the worktrees, the running processes or what the close kept.

#### Scenario: A detail that stopped working
- **WHEN** the commands a `local/` topic holds fail
- **THEN** the agent marks the line stale with the date and cause, and corrects the topic once another way works.

#### Scenario: No shared persistent memory
- **WHEN** the machine's persistent memory is the host's own, not Engram, and a fact needs more than its line
- **THEN** the agent writes the line alone, without a pointer.

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

### Requirement: Children report instead of writing
A child SHALL read the machine file and the project's `HIVE.local.md` files and SHALL NOT write them; it SHALL report each fact it verified or found stale, and the main thread SHALL update the right file and the detail in persistent memory. A child SHALL NOT be expected to read a detail from persistent memory: the main thread SHALL put in its brief the detail its task needs.

#### Scenario: A child starts the stack
- **WHEN** a delegated child starts the local stack with a command the file lacks
- **THEN** the child reports the command and its result, and the main thread records it.

### Requirement: Bootstrap and curation through the local-instructions skill
When the user asks to create, bootstrap, review or prune `HIVE.local.md`, the `local-instructions` skill SHALL gather candidates from persistent memory, host memory and local notes, the project's tracked guidance, the names of the secret location's files, and the skill's `extract_candidates.py` run over the last 30 days of this project's top-level Claude Code and Codex sessions, which SHALL print no secret value; SHALL discard diagnosis results, counts and whether a host answered; SHALL discard a `missing` program the project's own dependencies provide; SHALL send the other `missing` candidates, the installed tools and the existing cloud profiles with their regions to the machine file, and the `commands`, the `ports` and which profile serves an environment to the file their `where` names, each as one line with its detail in persistent memory when it needs more than its line, as the requirement «Verified, dated and secret-free entries» sets; SHALL ask the user which candidates to keep, group by group, before writing anything, keeping the answers in the files, and in persistent memory only that detail and an instruction of the user saved as a decision; SHALL, when it curates, propose for each existing entry longer than its line the line that stays and the detail that moves, for each entry that requirement excludes its removal, and for each entry that belongs in another of the files its move, changing nothing without the user's answer; SHALL confirm each kept fact with a read-only check from its fixed list when one applies and write the rest with its source date marked `unverified`; SHALL never run a project command that starts, stops, seeds, migrates, installs, signs in or writes; and SHALL report, without storing them, the facts that hold for every clone.

#### Scenario: First run on a machine with history
- **WHEN** the user asks to create `HIVE.local.md` and recent transcripts show a start command that exited 0, a port in use and a missing tool
- **THEN** the skill shows them as candidates, writes only those the user keeps, puts one line for the kept start command and port in the repository's file, with the command in persistent memory when it does not fit its line, and the kept missing tool in the machine file, and dates each by the read-only check that confirmed it or marks it `unverified` with its source date.

#### Scenario: A transcript command carries a credential
- **WHEN** a transcript command includes a password in an environment assignment, a URL or a header
- **THEN** the script's output, the question to the user and the file contain the command without that value.

#### Scenario: Curating a full file
- **WHEN** a file passes 50 lines and the user asks to prune it
- **THEN** the skill asks, for that file, whether to apply its proposals, and changes only the entries whose proposal the user applies.

#### Scenario: Curating a file of long entries
- **GIVEN** a `HIVE.local.md` whose entries hold setup recipes, events and the state of finished work
- **WHEN** the user asks to curate it
- **THEN** the skill lists each entry with its proposal: the one line that stays, the line plus the detail that moves to persistent memory, its move to another of the files, or its removal
- **AND** it changes the file only after the user answers.

### Requirement: Local scripts named by an entry
A project's `HIVE.local.md` MAY name a local script, a helper for this machine that the agent wrote; the machine file SHALL NOT name one. The script SHALL live at `_support/scripts/<name>` under the root of the file that names it, kept out of Git as the requirement «Ignored before the first write» sets for `HIVE.local.md`, and the entry SHALL name it by that path. The agent SHALL make a new script executable. From a worktree the agent SHALL run a repository's script by the same path under the main checkout, found as the requirement «Local instructions read at session start» sets, with the worktree as working directory; when that path is the worktree itself, a standalone copy, or a bare repository, the agent SHALL ask the user for the main checkout. The main thread MAY write or change a script without asking only in `flow-build`, including the fixes `flow-verify` makes, or in an operational task, and SHALL NOT do so in research, status, plan or a review; a child SHALL report the script it needs instead of writing it. When the host refuses the edit from a worktree, the agent SHALL write a new script there through a shell command only when no file exists at its path, and SHALL name a change to an existing script in the final report instead. A script SHALL NOT contain a secret value; it reads the secret location's files. The final report SHALL name each script the work created or changed. A script every clone needs SHALL NOT be kept as a local script; the agent names it as a candidate for the repository's tracked code.

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
