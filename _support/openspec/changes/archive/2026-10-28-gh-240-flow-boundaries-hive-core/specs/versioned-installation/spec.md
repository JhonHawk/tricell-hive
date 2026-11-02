## MODIFIED Requirements

### Requirement: Read-only diagnostics
The manager SHALL offer read-only diagnostics, both as the `hive doctor` and `hive models` text commands and as the Diagnostics, Models, Integrations, and Project views of the terminal interface, with the same content in both forms. They SHALL report each supported host's detection, version, release, and installation state; every resource whose status is other than `installed`, `retained_shared`, and `not_installed`, and any pending operation; open sessions that probably predate the installed release; the effective model and effort per role and host, marking roles with a model override and listing overrides that are not applied; the local evidence and last onboarding status of Engram, Context7, pi-subagents, `agent-browser`, and CodeGraph, and, when Codex is a registered host or its registration cannot be read, the Codex question card, and then the Codex wait floor, which has no onboarding status; when Pi is a registered host, how to keep Pi's `codemode` tool enabled, without checking whether it is; and the validity of a repository's `## Hive` section, where a retired key (`Review`, `UI verification`, or `Delivery` with `automatic` or `interactive`) is reported as retired, asked each session, with its old value only informing that question's Recommended option, without blocking. `hive doctor`, `hive models` without a subcommand, and the views while no edit is confirmed SHALL NOT write files or state and SHALL NOT read host or provider configuration, except Codex's `config.toml` for the Codex wait floor row, which SHALL report whether the floor is set, partly set, not set, refused, or not checked because the file could not be read, whether Hive's state records lines it wrote there, and whether `config.toml.hive-backup` exists, without showing any text from the file. They SHALL execute no program other than `--version` of a detected host and read-only `git` queries, except that opening an edit panel in the Models view MAY run that host's model-listing command under the Per-role model overrides requirement. CodeGraph's evidence SHALL come only from finding its program on the `PATH`, without running it. Diagnostics SHALL detect `agent-browser` and a browser for it by reading `PATH` and files only; when `agent-browser` is absent, or present without a browser, its next step SHALL name `hive install`, and its record SHALL show the worst status of its onboarding steps. A session marked as predating the installed release SHALL be presented as an estimate, registered hosts whose sessions cannot be observed SHALL show a restart notice instead, except a host whose documentation states that it reloads its instructions, which SHALL say so. Text read from outside the manager SHALL be stripped of control characters before it is shown. `hive doctor` SHALL exit with a non-zero status, naming the Project section on standard error, when the checked directory is inside a Git repository whose root `AGENTS.md` exists and either cannot be read as a regular file within the size limit or holds a `## Hive` section with a blocking finding: a missing or empty required value, a base branch name Git rejects, an invalid `Delivery` or `Hive guidance` value, or a duplicated section. A base branch that could not be checked SHALL be a warning. When the repository has no `AGENTS.md` or no section, or the checked path does not exist, is not inside a Git repository, or `git` cannot run, `hive doctor` SHALL exit with a non-zero status only when `--project` was given explicitly. It SHALL exit with status 0 otherwise.

#### Scenario: Stale session
- **WHEN** a Claude Code or Grok session whose process is alive started before the installed release was last written
- **THEN** diagnostics marks it as started before the installed release and suggests restarting it.

#### Scenario: Unreadable session record
- **WHEN** a host's session record is unreadable, larger than 1 MiB, or in an unrecognized format
- **THEN** that host's session check is reported as unavailable with its reason, and the other sections still show.

#### Scenario: Incomplete project section
- **WHEN** a repository's `AGENTS.md` lacks a required `## Hive` value or its `Specs` path does not exist
- **THEN** the Project check names each missing or invalid value and changes nothing.

#### Scenario: Blocking project finding fails the command
- **WHEN** `hive doctor` runs inside a Git repository whose `## Hive` section lacks a required value or has an invalid `Delivery` value
- **THEN** it prints the finding and exits with a non-zero status.

#### Scenario: Duplicated or oversized section file fails the command
- **WHEN** `hive doctor` runs inside a Git repository whose `AGENTS.md` holds two `## Hive` sections, or is larger than 1 MiB
- **THEN** it reports the finding and exits with a non-zero status.

#### Scenario: Other findings do not fail the command
- **WHEN** `hive doctor` finds drift, stale sessions, a retired key, or only warnings in the project section, or runs without `--project` outside a Git repository or in a repository without `AGENTS.md`
- **THEN** it prints them and exits with status 0.

#### Scenario: Explicit project without a section fails the command
- **WHEN** `hive doctor --project <repo>` runs on a Git repository whose `AGENTS.md` is missing or has no `## Hive` section, or `<repo>` does not exist or is outside a Git repository
- **THEN** it reports the finding and exits with a non-zero status.

#### Scenario: Retired key warns
- **WHEN** a repository's `## Hive` section declares `Review: /code-review` and `Delivery: automatic`
- **THEN** the Project check reports each as retired and asked each session, with the old value only informing the Recommended option, without blocking.

#### Scenario: Pi codemode guidance without reading Pi
- **WHEN** `hive doctor` runs with Pi registered while Pi's settings already enable `codemode`
- **THEN** the Integrations section shows a `Pi codemode` row that stays `not checked` and whose next step names `"defaultTools": ["+codemode"]`

#### Scenario: No Pi codemode row without Pi
- **WHEN** `hive doctor` runs and Pi is not a registered host
- **THEN** the Integrations section shows no `Pi codemode` row

#### Scenario: Codex wait floor read without repeating the file
- **WHEN** `hive doctor` runs with Codex registered and Codex's `config.toml` holds the wait floor together with other settings
- **THEN** the Integrations section shows a `Codex wait floor` row as applied, says whether Hive wrote those lines and whether the backup exists, and shows no other text from the file, and no other host or provider configuration is read.

#### Scenario: CodeGraph detected without running it
- **WHEN** `hive doctor` runs and a `codegraph` program is on the `PATH`
- **THEN** the Integrations section shows a CodeGraph row as detected with the program's path, and the program is not executed.

#### Scenario: agent-browser present without a browser
- **WHEN** `hive doctor` runs with `agent-browser` on `PATH`, no browser in the locations the installer checks, and its last Chrome step failed
- **THEN** the Integrations row names `hive install` as the next step, its record shows the failed step, and no program runs.
