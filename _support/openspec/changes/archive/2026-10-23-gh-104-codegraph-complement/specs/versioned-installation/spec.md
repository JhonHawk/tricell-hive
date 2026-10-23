## MODIFIED Requirements

### Requirement: Optional operation capabilities
The optional capability catalog SHALL be limited to Engram, Context7, pi-subagents, agent-browser, the Codex question card (`codex-question-card`), and CodeGraph. Engram and Context7 SHALL remain manual until each recipe passes a native gate for the host and platform. CodeGraph SHALL remain manual-only: Hive SHALL NOT run CodeGraph's installer or edit the host configuration it writes. A selected capability without a verified recipe SHALL be reported as manual, with its official source and reason, and SHALL NOT be presented as installed or verified. pi-subagents on a user-scope Pi host follows the Pi user-scope install declares pi-subagents requirement instead of that manual path. agent-browser and the Codex question card SHALL run only their fixed recipes, after the operator's consent. Development-of-Hive requirements SHALL remain separate from installation and operation requirements. Hosts, runtimes and other tools SHALL NOT be installed implicitly.

#### Scenario: Capability without a verified recipe
- **WHEN** the user selects Engram or Context7, whose recipes have not passed their native gate, or CodeGraph, which has no recipe
- **THEN** the core installs, no provider process runs, and the result lists the capability as manual with its official instructions and a non-zero exit status.

#### Scenario: Planned CodeGraph step rejected
- **WHEN** an onboarding record holds a planned step with a command for CodeGraph
- **THEN** step validation rejects it and nothing runs.

### Requirement: Read-only diagnostics
The manager SHALL offer read-only diagnostics, both as the `hive doctor` and `hive models` text commands and as the Diagnostics, Models, Integrations, and Project views of the terminal interface, with the same content in both forms. They SHALL report each supported host's detection, version, release, and installation state; every resource whose status is other than `installed`, `retained_shared`, and `not_installed`, and any pending operation; open sessions that probably predate the installed release; the effective model and effort per role and host, marking roles with a model override and listing overrides that are not applied; the local evidence and last onboarding status of Engram, Context7, pi-subagents, `agent-browser`, and CodeGraph, and, when Codex is a registered host or its registration cannot be read, the Codex question card; when Pi is a registered host, how to keep Pi's `codemode` tool enabled, without checking whether it is; and the validity of a repository's `## Hive` section. `hive doctor`, `hive models` without a subcommand, and the views while no edit is confirmed SHALL NOT write files or state and SHALL NOT read host or provider configuration. They SHALL execute no program other than `--version` of a detected host and read-only `git` queries, except that opening an edit panel in the Models view MAY run that host's model-listing command under the Per-role model overrides requirement. CodeGraph's evidence SHALL come only from finding its program on the `PATH`, without running it. A session marked as predating the installed release SHALL be presented as an estimate, registered hosts whose sessions cannot be observed SHALL show a restart notice instead, except a host whose documentation states that it reloads its instructions, which SHALL say so. Text read from outside the manager SHALL be stripped of control characters before it is shown.

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
