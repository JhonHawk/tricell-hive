## MODIFIED Requirements
### Requirement: Read-only diagnostics
The manager SHALL offer read-only diagnostics, both as the `hive doctor` and `hive models` text commands and as the Diagnostics, Models, Integrations, and Project views of the terminal interface, with the same content in both forms. They SHALL report each supported host's detection, version, release, and installation state; every resource whose status is other than `installed`, `retained_shared`, and `not_installed`, and any pending operation; open sessions that probably predate the installed release; the effective model and effort per role and host, marking roles with a model override and listing overrides that are not applied; the local evidence and last onboarding status of Engram, Context7, pi-subagents, `agent-browser`, and CodeGraph, and, when Codex is a registered host or its registration cannot be read, the Codex question card; when Pi is a registered host, how to keep Pi's `codemode` tool enabled, without checking whether it is; and the validity of a repository's `## Hive` section, where a retired key (`Review`, `UI verification`, or `Delivery` with `automatic` or `interactive`) is reported as no longer read without blocking. `hive doctor`, `hive models` without a subcommand, and the views while no edit is confirmed SHALL NOT write files or state and SHALL NOT read host or provider configuration. They SHALL execute no program other than `--version` of a detected host and read-only `git` queries, except that opening an edit panel in the Models view MAY run that host's model-listing command under the Per-role model overrides requirement. CodeGraph's evidence SHALL come only from finding its program on the `PATH`, without running it. Diagnostics SHALL detect `agent-browser` and a browser for it by reading `PATH` and files only; when `agent-browser` is absent, or present without a browser, its next step SHALL name `hive install`, and its record SHALL show the worst status of its onboarding steps. A session marked as predating the installed release SHALL be presented as an estimate, registered hosts whose sessions cannot be observed SHALL show a restart notice instead, except a host whose documentation states that it reloads its instructions, which SHALL say so. Text read from outside the manager SHALL be stripped of control characters before it is shown. `hive doctor` SHALL exit with a non-zero status, naming the Project section on standard error, when the checked directory is inside a Git repository whose root `AGENTS.md` exists and either cannot be read as a regular file within the size limit or holds a `## Hive` section with a blocking finding: a missing or empty required value, a base branch name Git rejects, an invalid `Delivery` or `Hive guidance` value, or a duplicated section. A base branch that could not be checked SHALL be a warning. When the repository has no `AGENTS.md` or no section, or the checked path does not exist, is not inside a Git repository, or `git` cannot run, `hive doctor` SHALL exit with a non-zero status only when `--project` was given explicitly. It SHALL exit with status 0 otherwise.

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

#### Scenario: CodeGraph detected without running it
- **WHEN** `hive doctor` runs and a `codegraph` program is on the `PATH`
- **THEN** the Integrations section shows a CodeGraph row as detected with the program's path, and the program is not executed.

#### Scenario: agent-browser present without a browser
- **WHEN** `hive doctor` runs with `agent-browser` on `PATH`, no browser in the locations the installer checks, and its last Chrome step failed
- **THEN** the Integrations row names `hive install` as the next step, its record shows the failed step, and no program runs.

### Requirement: Project settings section editing
The manager SHALL let the user create or edit the `## Hive` section of the `AGENTS.md` at the root of a Git repository, through `hive project set` and through the Project view, after showing the section's lines before and after the change and asking for confirmation. The editable keys SHALL be `Project`, `Base branch`, `Tracker`, `Specs`, `Environments`, `Delivery` (only `direct-base`), `Completion`, and `Hive guidance`. It SHALL change only the given keys, preserve every other line of the file, its line endings and its permissions, append a missing section at the end of the file, and create `AGENTS.md` with only the section when the file does not exist. It SHALL refuse, without writing, a duplicate section, a symbolic link or non-regular file, a file larger than 1 MiB, a directory outside a Git repository, a missing or empty required value, a value with line breaks, control characters, or characters that would be stripped when shown, an unknown or retired key given by the user, a `Delivery` value other than `direct-base` given by the user, an invalid `Delivery` or `Hive guidance` value, a base branch name Git rejects, and a file that changed after the preview. A retired key already in the file SHALL be preserved unless the user unsets it. A base branch or `Specs` path that does not exist yet, and a `Completion` value without the `<state> on <ref> · <state> on <ref>` shape, SHALL be reported as a warning without blocking. Values suggested from the repository SHALL be labeled as suggestions until the user confirms them. When the repository root holds a `CLAUDE.md` that does not import `@AGENTS.md`, the preview SHALL warn about it and the manager SHALL NOT modify `CLAUDE.md`. This write SHALL NOT enter Hive's transaction journal.

#### Scenario: Section added to an existing file
- **WHEN** the user saves the section for a repository whose `AGENTS.md` has none
- **THEN** the section is appended at the end and every earlier byte of the file is unchanged.

#### Scenario: File changed after the preview
- **WHEN** `AGENTS.md` changes between the preview and the confirmation
- **THEN** nothing is written and the manager says the file changed since the preview.

#### Scenario: Declined edit
- **WHEN** the user cancels the confirmation
- **THEN** `AGENTS.md` and `CLAUDE.md` are unchanged.

#### Scenario: Retired key refused on set, removable on unset
- **WHEN** the user runs `hive project set --set "Review: /code-review"` on a section that already holds `UI verification: full`
- **THEN** the manager refuses the set naming `Review` as retired, and `hive project set --unset "UI verification"` removes the existing line.

#### Scenario: Retired Delivery mode refused on set
- **WHEN** the user runs `hive project set --set "Delivery: automatic"`
- **THEN** the manager refuses it naming `automatic` as a retired `Delivery` value and writes nothing.
