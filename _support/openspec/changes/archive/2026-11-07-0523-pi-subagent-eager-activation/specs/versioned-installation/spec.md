## ADDED Requirements

### Requirement: Pi subagent tool active from the start
When a user-scope install, package installation or committed-source update plan includes the Pi host, the manager SHALL plan setting `toolActivation` to `"eager"` in pi-subagents' `extensions/subagent/config.json` under the plan's Pi home when that key is missing, so that the `subagent` tool is active from a session's first request; the summary SHALL show the path, the key, the value and that open Pi sessions need a restart. When Hive's state records that it wrote the key for that same path and the file still holds `"eager"`, the plan SHALL report it as already applied and SHALL carry the record forward. When the key is present with any other value, or holds `"eager"` that Hive's state does not record, the plan SHALL write nothing, SHALL record nothing, and SHALL report the value as set by the user. When the file is not a JSON object with unique keys, is a symbolic link or not a regular file, exceeds 1 MiB, or its folder cannot be written safely, the plan SHALL write nothing and SHALL report the key and value as a manual step. When Hive's record is for another Pi home, the plan SHALL write nothing for the new home and SHALL report the record it keeps. A migration from the legacy installer that edits the same file SHALL be planned together with this step, so that apply does not report its own migration as a concurrent change. Apply SHALL revalidate the file's fingerprint, copy an existing file to `config.json.hive-backup` when no such copy exists, insert only that key while keeping the other keys, their values and their order, keep an existing file's mode and create a new file with mode 0600 and its folders with mode 0700, and record in Hive's state the path, the value written, whether Hive created the file, the folders it created and whether it created the backup, never the file's other content. A failed apply or an interrupted removal SHALL be recovered at the key level from the journal: recovery removes a key that apply inserted, or restores a key that removal deleted, and never rewrites the file's other content. Removing Pi with `hive plan remove` SHALL delete the key only when the state records that Hive wrote it and the value is still `"eager"`, SHALL delete the file and the folders only when Hive created them and nothing else is left in them, and SHALL delete the backup only when Hive created it and the key was removed; it SHALL report each thing it leaves. Complete release updates SHALL NOT plan this step and SHALL reject a plan that carries it. A committed-source update whose only difference is this step SHALL NOT be treated as unchanged. Tests SHALL use a synthetic Pi home.

#### Scenario: Key missing
- **WHEN** a user-scope install that includes `pi` finds `extensions/subagent/config.json` absent, or a JSON object without `toolActivation`
- **THEN** the plan shows the step, apply writes `"toolActivation": "eager"` and leaves every other key, value and their order unchanged, and the state records what Hive wrote.

#### Scenario: Key Hive wrote earlier
- **WHEN** a later committed-source update finds `"toolActivation": "eager"` and the state records that Hive wrote it for that path
- **THEN** the plan reports the step as already applied, writes nothing and keeps the record.

#### Scenario: Key set by the user
- **WHEN** the file holds `toolActivation` with a value Hive's state does not record
- **THEN** the plan writes nothing, records nothing, and reports the value as set by the user.

#### Scenario: Removal after the user changed the value
- **WHEN** `hive plan remove` targets Pi and the state records the key, but the file now holds another value
- **THEN** the key and the file are left as they are and the report says so.

## MODIFIED Requirements

### Requirement: Read-only diagnostics
The manager SHALL offer read-only diagnostics, both as the `hive doctor` and `hive models` text commands and as the Diagnostics, Models, Integrations, and Project views of the terminal interface, with the same content in both forms. They SHALL report each supported host's detection, version, release, and installation state; every resource whose status is other than `installed`, `retained_shared`, and `not_installed`, and any pending operation; open sessions that probably predate the installed release; the effective model and effort per role and host, marking roles with a model override and listing overrides that are not applied; the local evidence and last onboarding status of Engram, Context7, pi-subagents, `agent-browser`, and CodeGraph, and, when Codex is a registered host or its registration cannot be read, the Codex question card, and then the Codex wait floor, which has no onboarding status; when Pi is a registered host, the Pi subagent activation, which has no onboarding status; when Pi is a registered host, how to keep Pi's `codemode` tool enabled, without checking whether it is; and the validity of a repository's `## Hive` section, where a retired key (`Review`, `UI verification`, or `Delivery` with `automatic` or `interactive`) is reported as retired, asked each session, with its old value only informing that question's Recommended option, without blocking. `hive doctor`, `hive models` without a subcommand, and the views while no edit is confirmed SHALL NOT write files or state and SHALL NOT read host or provider configuration, except Codex's `config.toml` for the Codex wait floor row, which SHALL report whether the floor is set, partly set, not set, refused, or not checked because the file could not be read, whether Hive's state records lines it wrote there, and whether `config.toml.hive-backup` exists, without showing any text from the file, and pi-subagents' `extensions/subagent/config.json` under the Pi home for the Pi subagent activation row, which SHALL report whether `toolActivation` is eager as written by Hive, set by the user with its value, not set, or not read because the file could not be read safely, and SHALL name `hive install` when the key is not set, without showing any other text from the file. They SHALL execute no program other than `--version` of a detected host and read-only `git` queries, except that opening an edit panel in the Models view MAY run that host's model-listing command under the Per-role model overrides requirement. CodeGraph's evidence SHALL come only from finding its program on the `PATH`, without running it. Diagnostics SHALL detect `agent-browser` and a browser for it by reading `PATH` and files only; when `agent-browser` is absent, or present without a browser, its next step SHALL name `hive install`, and its record SHALL show the worst status of its onboarding steps. A session marked as predating the installed release SHALL be presented as an estimate, registered hosts whose sessions cannot be observed SHALL show a restart notice instead, except a host whose documentation states that it reloads its instructions, which SHALL say so. Text read from outside the manager SHALL be stripped of control characters before it is shown. `hive doctor` SHALL exit with a non-zero status, naming the Project section on standard error, when the checked directory is inside a Git repository whose root `AGENTS.md` exists and either cannot be read as a regular file within the size limit or holds a `## Hive` section with a blocking finding: a missing or empty required value, a base branch name Git rejects, an invalid `Delivery` or `Hive guidance` value, or a duplicated section. A base branch that could not be checked SHALL be a warning. When the repository has no `AGENTS.md` or no section, or the checked path does not exist, is not inside a Git repository, or `git` cannot run, `hive doctor` SHALL exit with a non-zero status only when `--project` was given explicitly. It SHALL exit with status 0 otherwise.

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

#### Scenario: Pi subagent activation read without repeating the file
- **WHEN** `hive doctor` runs with Pi registered and pi-subagents' `config.json` holds `toolActivation` together with other settings
- **THEN** the Integrations section shows a `Pi subagent activation` row with the key's state and, when Hive wrote it, says so, and shows no other text from the file.
