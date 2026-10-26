## MODIFIED Requirements
### Requirement: Optional operation capabilities
The optional capability catalog SHALL be limited to Engram, Context7, pi-subagents, agent-browser, the Codex question card (`codex-question-card`), and CodeGraph. Engram and Context7 SHALL remain manual until each recipe passes a native gate for the host and platform. CodeGraph SHALL remain manual-only: Hive SHALL NOT run CodeGraph's installer or edit the host configuration it writes. A selected capability without a verified recipe SHALL be reported as manual, with its official source and reason, and SHALL NOT be presented as installed or verified. pi-subagents on a user-scope Pi host follows the Pi user-scope install declares pi-subagents requirement instead of that manual path. `agent-browser` does not depend on the host, so its native gate is per platform: it SHALL be offered as installable only on macOS arm64 with Homebrew (`brew install agent-browser`) and on Linux amd64 with Node 24 or later and a writable global npm prefix (`npm install -g agent-browser`), and otherwise SHALL be reported as manual with the reason and the official command. Its gate passed on Linux amd64 for both steps and on macOS arm64 for the Chrome step only; the Homebrew step is covered by tests with doubles, a risk the user accepted. It installs the latest version, with no version prompt, and runs `agent-browser install` to download Chrome only when no browser exists in the locations agent-browser checks. Each step SHALL run only after interactive consent, without a shell or `sudo`, with an empty standard input, in its own process group, with a time limit, and with its saved output truncated; the step validator SHALL accept for it only the command lines bound to each step ID. Before consent, the summary SHALL show each exact command and SHALL state that Homebrew installs Node and its dependencies, so that runtime is declared rather than implicit. The Codex question card SHALL run only its fixed recipe, after the operator's consent. Development-of-Hive requirements SHALL remain separate from installation and operation requirements. Hosts, runtimes and other tools SHALL NOT be installed implicitly.

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

### Requirement: Verified release package installation
A complete package published as a GitHub Release asset SHALL be installable after the user verifies it against its published `.sha256` file and runs the package's `install.sh`. The installer SHALL validate and safely extract the package. The core installation SHALL work without Go, Git, GitHub CLI, or network downloads. Only two optional steps may download: the Pi host's pi-subagents step, from npm, and, after the user's consent, the `agent-browser` capability, from Homebrew or npm (whose postinstall downloads a native binary from GitHub without a checksum) and, when no browser is found, Chrome for Testing. Every package SHALL include `LICENSE` and `THIRD_PARTY_NOTICES.md`.

#### Scenario: Invalid archive
- **WHEN** the package checksum, archive paths, entry types, or limits fail validation
- **THEN** installation does not begin and no entry is written outside the task-owned extraction directory.

#### Scenario: Cancellation or unavailable terminal
- **WHEN** the user cancels before application or a controlling terminal is unavailable
- **THEN** no persistent installation changes occur.

#### Scenario: Pi host selected without pi-subagents
- **WHEN** a user-scope package installation includes the Pi host and Pi's settings do not declare pi-subagents
- **THEN** the plan runs `pi install npm:pi-subagents`, which downloads from npm; no other host step downloads, and capability steps download only as the next scenario sets.

#### Scenario: agent-browser capability consented
- **WHEN** the operator consents to the agent-browser capability in an interactive package installation
- **THEN** on a supported platform the plan runs only that capability's fixed recipe commands, which download from Homebrew or npm (whose postinstall downloads a native binary from GitHub) and, when no browser is found, Chrome for Testing; without that consent no agent-browser step runs.

### Requirement: Read-only diagnostics
The manager SHALL offer read-only diagnostics, both as the `hive doctor` and `hive models` text commands and as the Diagnostics, Models, Integrations, and Project views of the terminal interface, with the same content in both forms. They SHALL report each supported host's detection, version, release, and installation state; every resource whose status is other than `installed`, `retained_shared`, and `not_installed`, and any pending operation; open sessions that probably predate the installed release; the effective model and effort per role and host, marking roles with a model override and listing overrides that are not applied; the local evidence and last onboarding status of Engram, Context7, pi-subagents, `agent-browser`, and CodeGraph, and, when Codex is a registered host or its registration cannot be read, the Codex question card; when Pi is a registered host, how to keep Pi's `codemode` tool enabled, without checking whether it is; and the validity of a repository's `## Hive` section. `hive doctor`, `hive models` without a subcommand, and the views while no edit is confirmed SHALL NOT write files or state and SHALL NOT read host or provider configuration. They SHALL execute no program other than `--version` of a detected host and read-only `git` queries, except that opening an edit panel in the Models view MAY run that host's model-listing command under the Per-role model overrides requirement. CodeGraph's evidence SHALL come only from finding its program on the `PATH`, without running it. Diagnostics SHALL detect `agent-browser` and a browser for it by reading `PATH` and files only; when `agent-browser` is absent, or present without a browser, its next step SHALL name `hive install`, and its record SHALL show the worst status of its onboarding steps. A session marked as predating the installed release SHALL be presented as an estimate, registered hosts whose sessions cannot be observed SHALL show a restart notice instead, except a host whose documentation states that it reloads its instructions, which SHALL say so. Text read from outside the manager SHALL be stripped of control characters before it is shown.

#### Scenario: Stale session
- **WHEN** a Claude Code or Grok session whose process is alive started before the installed release was last written
- **THEN** diagnostics marks it as started before the installed release and suggests restarting it.

#### Scenario: Unreadable session record
- **WHEN** a host's session record is unreadable, larger than 1 MiB, or in an unrecognized format
- **THEN** that host's session check is reported as unavailable with its reason, and the other sections still show.

#### Scenario: Incomplete project section
- **WHEN** a repository's `AGENTS.md` lacks a required `## Hive` value or its `Specs` path does not exist
- **THEN** the Project check names each missing or invalid value and changes nothing.

#### Scenario: Findings do not fail the command
- **WHEN** `hive doctor` finds drift, stale sessions, or an invalid project section
- **THEN** it prints them and exits with status 0.

#### Scenario: Pi codemode guidance without reading Pi
- **WHEN** `hive doctor` runs with Pi registered while Pi's settings already enable `codemode`
- **THEN** the Integrations section shows a `Pi codemode` row that stays `not checked` and whose next step names `"defaultTools": ["+codemode"]`

#### Scenario: No Pi codemode row without Pi
- **WHEN** `hive doctor` runs and Pi is not a registered host
- **THEN** the Integrations section shows no `Pi codemode` row

#### Scenario: CodeGraph detected without running it
- **WHEN** `hive doctor` runs and a `codegraph` program is on the `PATH`
- **THEN** the Integrations section shows a CodeGraph row as detected with the program's path, and the program is not executed.

#### Scenario: agent-browser present without a browser
- **WHEN** `hive doctor` runs with `agent-browser` on `PATH`, no browser in the locations the installer checks, and its last Chrome step failed
- **THEN** the Integrations row names `hive install` as the next step, its record shows the failed step, and no program runs.
