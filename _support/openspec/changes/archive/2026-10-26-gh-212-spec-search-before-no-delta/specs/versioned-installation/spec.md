## MODIFIED Requirements

### Requirement: Optional operation capabilities
The optional capability catalog SHALL be limited to Engram, Context7, pi-subagents, agent-browser, the Codex question card (`codex-question-card`), the Codex wait floor (`codex-wait-floor`), and CodeGraph. Engram and Context7 SHALL remain manual until each recipe passes a native gate for the host and platform. CodeGraph SHALL remain manual-only: Hive SHALL NOT run CodeGraph's installer or edit the host configuration it writes. A selected capability without a verified recipe SHALL be reported as manual, with its official source and reason, and SHALL NOT be presented as installed or verified. pi-subagents on a user-scope Pi host follows the Pi user-scope install declares pi-subagents requirement instead of that manual path. `agent-browser` does not depend on the host, so its native gate is per platform: it SHALL be offered as installable only on macOS arm64 with Homebrew (`brew install agent-browser`) and on Linux amd64 with Node 24 or later and a writable global npm prefix (`npm install -g agent-browser`), and otherwise SHALL be reported as manual with the reason and the official command. Its gate passed on Linux amd64 for both steps and on macOS arm64 for the Chrome step only; the Homebrew step is covered by tests with doubles, a risk the user accepted. It installs the latest version, with no version prompt, and runs `agent-browser install` to download Chrome only when no browser exists in the locations agent-browser checks. Each step SHALL run only after interactive consent, without a shell or `sudo`, with an empty standard input, in its own process group, with a time limit, and with its saved output truncated; the step validator SHALL accept for it only the command lines bound to each step ID. Before consent, the summary SHALL show each exact command and SHALL state that Homebrew installs Node and its dependencies, so that runtime is declared rather than implicit. The Codex question card SHALL run only its fixed recipe, after the operator's consent. The Codex wait floor SHALL run no external step: it SHALL be offered only in an interactive user-scope `hive install` or `hive update` that targets Codex, with a `codex` program on `PATH` whose version loads `[features.multi_agent_v2]`, and only when the floor is not already set. With the operator's consent, the installation plan SHALL insert the missing `min_wait_timeout_ms` and `default_wait_timeout_ms` keys into that table of Codex's `config.toml`, creating the table or the file when absent and leaving the file's other lines unchanged, after copying an existing file to `config.toml.hive-backup` when no such copy exists. It SHALL be offered only as a manual step, writing nothing, when the file holds a conflicting value or a form it does not edit, when the file or Codex's home cannot be written safely, or when Hive's record of the floor is for another Codex home. Hive's state and journal SHALL record only the file's fingerprints and what Hive inserted, never the file's other content. Removing Codex with `hive plan remove` SHALL delete only the recorded keys whose lines still hold the values Hive wrote; it SHALL delete the table or the file that Hive created only when nothing else is left in it, and the backup that Hive created only when that removal deleted every recorded key. Development-of-Hive requirements SHALL remain separate from installation and operation requirements. Hosts, runtimes and other tools SHALL NOT be installed implicitly.

#### Scenario: Capability without a verified recipe
- **WHEN** the user selects Engram or Context7, whose recipes have not passed their native gate, or CodeGraph, which has no recipe
- **THEN** the core installs, no provider process runs, and the result lists the capability as manual with its official instructions and a non-zero exit status.

#### Scenario: Planned CodeGraph step rejected
- **WHEN** an onboarding record holds a planned step with a command for CodeGraph
- **THEN** step validation rejects it and nothing runs.

#### Scenario: agent-browser on a supported platform
- **WHEN** an interactive `hive install` runs on Linux amd64 with Node 24 and a writable npm prefix, `agent-browser` is not on `PATH`, no browser is found, and the user consents
- **THEN** the installer runs `npm install -g agent-browser` and then `agent-browser install`, and each step is recorded as verified by its own check.

#### Scenario: agent-browser step fails
- **WHEN** the agent-browser install step fails, times out, or is interrupted
- **THEN** the step is recorded as failed with its message shown in the result detail, the Chrome step is skipped, the core installation is kept, and the journal never records the outcome as unknown.

#### Scenario: Unsupported platform or missing prerequisite
- **WHEN** the platform is not macOS arm64 or Linux amd64, or Homebrew, Node 24, or a writable npm prefix is missing
- **THEN** `agent-browser` is listed as manual with the reason and the official command, and no program runs.

#### Scenario: Safe modes run nothing
- **WHEN** the session is not interactive, or the user runs `--dry-run` or uses a synthetic `--home`
- **THEN** detection runs no program and no installation step runs.

### Requirement: Read-only diagnostics
The manager SHALL offer read-only diagnostics, both as the `hive doctor` and `hive models` text commands and as the Diagnostics, Models, Integrations, and Project views of the terminal interface, with the same content in both forms. They SHALL report each supported host's detection, version, release, and installation state; every resource whose status is other than `installed`, `retained_shared`, and `not_installed`, and any pending operation; open sessions that probably predate the installed release; the effective model and effort per role and host, marking roles with a model override and listing overrides that are not applied; the local evidence and last onboarding status of Engram, Context7, pi-subagents, `agent-browser`, and CodeGraph, and, when Codex is a registered host or its registration cannot be read, the Codex question card, and then the Codex wait floor, which has no onboarding status; when Pi is a registered host, how to keep Pi's `codemode` tool enabled, without checking whether it is; and the validity of a repository's `## Hive` section, where a retired key (`Review`, `UI verification`, or `Delivery` with `automatic` or `interactive`) is reported as no longer read without blocking. `hive doctor`, `hive models` without a subcommand, and the views while no edit is confirmed SHALL NOT write files or state and SHALL NOT read host or provider configuration, except Codex's `config.toml` for the Codex wait floor row, which SHALL report whether the floor is set, partly set, not set, refused, or not checked because the file could not be read, whether Hive's state records lines it wrote there, and whether `config.toml.hive-backup` exists, without showing any text from the file. They SHALL execute no program other than `--version` of a detected host and read-only `git` queries, except that opening an edit panel in the Models view MAY run that host's model-listing command under the Per-role model overrides requirement. CodeGraph's evidence SHALL come only from finding its program on the `PATH`, without running it. Diagnostics SHALL detect `agent-browser` and a browser for it by reading `PATH` and files only; when `agent-browser` is absent, or present without a browser, its next step SHALL name `hive install`, and its record SHALL show the worst status of its onboarding steps. A session marked as predating the installed release SHALL be presented as an estimate, registered hosts whose sessions cannot be observed SHALL show a restart notice instead, except a host whose documentation states that it reloads its instructions, which SHALL say so. Text read from outside the manager SHALL be stripped of control characters before it is shown. `hive doctor` SHALL exit with a non-zero status, naming the Project section on standard error, when the checked directory is inside a Git repository whose root `AGENTS.md` exists and either cannot be read as a regular file within the size limit or holds a `## Hive` section with a blocking finding: a missing or empty required value, a base branch name Git rejects, an invalid `Delivery` or `Hive guidance` value, or a duplicated section. A base branch that could not be checked SHALL be a warning. When the repository has no `AGENTS.md` or no section, or the checked path does not exist, is not inside a Git repository, or `git` cannot run, `hive doctor` SHALL exit with a non-zero status only when `--project` was given explicitly. It SHALL exit with status 0 otherwise.

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
- **THEN** the Project check reports each as no longer read and asked each session, without blocking.

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

### Requirement: Per-role model overrides
The manager SHALL let the user override the model and effort of one agent role on one registered user-scope host, through `hive models set` and `hive models reset` and through the Models view. Overrides SHALL be stored in Hive's state, SHALL NOT change release identifiers or release snapshots, and SHALL be applied by every later installation plan for that host, including updates and returns to a retained release. An override SHALL take precedence over the release's profile and the role's own effort. Applying or removing an override SHALL rewrite only that role's agent file on that host and SHALL keep the host's installation receipt consistent with the written files. Removing Hive from a host SHALL drop its overrides. An override SHALL be refused, without writing, when its host is not registered, its role is not in the host's installed release, its effort is set on a host that accepts none, is not one of `low`, `medium`, `high`, `xhigh`, `max`, and `ultra`, or is `ultra` on Pi, whose thinking levels end at `max`, its model is empty, longer than 200 characters, or contains a character outside letters, digits, and `._:/@+=[]-`, or applying it would change anything on the host other than agent files. The text commands SHALL accept any model the validation allows, whether or not the host lists it. Roles SHALL be grouped by the folder of their source (`content/agents/<group>/`); setting a group SHALL write an override with only the given parts to every role of the group, replacing each role's own override, and SHALL name before confirmation the roles that lose a part; resetting a group SHALL remove the overrides of its roles. On OpenCode, a new model without a chosen effort SHALL keep the profile's effort, including a variant embedded in the profile's model, without storing that effort in the override. The Models view SHALL offer, for the model, a filterable list made of `release default`, the effective model, the host's models, and a free-text entry. The host's models SHALL come from its listing command (`codex debug models`, `opencode models`, `pi --list-models`, `grok models`, `cursor-agent models`) or, for Claude Code, from the fixed aliases `fable`, `opus`, `sonnet`, and `haiku`; the command SHALL run only when an edit panel of that host is first opened in a session, as the binary found on `PATH`, without a shell, in a neutral working directory, detached from the terminal, with stdin from the null device, with a time limit after which its process group is terminated, and with a capped output whose excess is an error. With an explicit synthetic home no host command SHALL run. Internal host model caches SHALL NOT be read.

#### Scenario: Group set replaces own overrides
- **WHEN** a role has its own override and the user sets a model for its group
- **THEN** the confirmation names that role, and after applying every role of the group carries only the group's parts.

#### Scenario: Model list unavailable
- **WHEN** a host's listing command fails, times out, or exceeds the output cap
- **THEN** the list shows the reason and still offers the effective model and the free-text entry, and nothing is written.

#### Scenario: Override survives an update
- **WHEN** a role has a model override and a new release is installed with `hive update`
- **THEN** that role's agent file still carries the overridden model and effort, and the host's installation is reported as verified.

#### Scenario: Override removed
- **WHEN** the user resets a role's override
- **THEN** the role's agent file becomes byte for byte the file an installation without overrides writes.

#### Scenario: Unsupported effort
- **WHEN** the user sets an effort for a role on Grok or Cursor, or the effort `ultra` on Pi
- **THEN** the manager refuses it, names the reason, and changes nothing.
