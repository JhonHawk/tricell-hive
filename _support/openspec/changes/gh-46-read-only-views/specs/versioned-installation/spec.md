## ADDED Requirements

### Requirement: Read-only diagnostics
The manager SHALL offer read-only diagnostics, both as the `hive doctor` and `hive models` text commands and as the Diagnostics, Models, Integrations, and Project views of the terminal interface, with the same content in both forms. They SHALL report each supported host's detection, version, release, and installation state; every resource whose status is other than `installed`, `retained_shared`, and `not_installed`, and any pending operation; open sessions that probably predate the installed release; the effective model and effort per role and host; the local evidence and last onboarding status of Engram, Context7, pi-subagents, and `agent-browser`; and the validity of a repository's `## Hive` section. They SHALL NOT write files or state, SHALL execute no program other than `--version` of a detected host and read-only `git` queries, and SHALL NOT read host or provider configuration. Detecting `agent-browser` SHALL NOT add it to the optional capability catalog. A session marked as predating the installed release SHALL be presented as an estimate, registered hosts whose sessions cannot be observed SHALL show a restart notice instead, except a host whose documentation states that it reloads its instructions, which SHALL say so. Text read from outside the manager SHALL be stripped of control characters before it is shown.

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

## MODIFIED Requirements

### Requirement: Interactive terminal interface
Running the manager without arguments in a terminal, or `hive tui`, SHALL open a full-screen interface in the terminal's alternate screen for host status, installing and removing hosts, updating from a commit, returning to a retained release, the voice layer, and the read-only diagnostics, models, integrations, and project views. Changing views SHALL redraw the whole screen, and leaving the interface SHALL restore the terminal's previous content. Every action that writes SHALL show a preview and ask for confirmation first, except an update whose plan changes nothing, which records the source commit without asking as the update command does, and SHALL produce the same files and state as the equivalent command. Declining, going back, or an error inside an action SHALL change nothing further and return to the originating view with the message shown inside it. Without a terminal, running the manager without arguments SHALL keep its usage error, and the text commands SHALL remain the interface for scripts and screen readers. The core management and distribution packages SHALL NOT depend on third-party modules.

#### Scenario: No terminal
- **WHEN** the manager runs without arguments and without a terminal
- **THEN** it prints the usage error and changes nothing.

#### Scenario: Going back
- **WHEN** the user presses Esc in any view, or Backspace while no text field has focus
- **THEN** the previous view returns, redrawn in place.

#### Scenario: Hosts changed in one pass
- **WHEN** the user checks some hosts and unchecks others in the hosts view and applies
- **THEN** the removal is previewed, confirmed and applied first, and the installation is then planned on the resulting state, previewed and confirmed separately.

#### Scenario: Declined or cancelled action
- **WHEN** the user declines or goes back from a confirmation
- **THEN** no file or state changes and the originating view shows that nothing was applied.

#### Scenario: Read-only views
- **WHEN** the user opens the Diagnostics, Models, Integrations, or Project view and leaves it
- **THEN** no file, state, or project content changes.
