## MODIFIED Requirements

### Requirement: Interactive terminal interface
Running the manager without arguments in a terminal, or `hive tui`, SHALL open a full-screen interface in the terminal's alternate screen for host status, installing and removing hosts, updating from a commit, returning to a retained release, the voice layer, the read-only diagnostics and integrations views, the models view with per-role and per-group model overrides and a model list taken from each host, and the project view with editing of a repository's `## Hive` section. Changing views SHALL redraw the whole screen, and leaving the interface SHALL restore the terminal's previous content. Every action that writes SHALL show a preview and ask for confirmation first, except an update whose plan changes nothing, which records the source commit without asking as the update command does, and SHALL produce the same files and state as the equivalent command. Declining, going back, or an error inside an action SHALL change nothing further and return to the originating view with the message shown inside it. Without a terminal, running the manager without arguments SHALL keep its usage error, and the text commands SHALL remain the interface for scripts and screen readers. The core management and distribution packages SHALL NOT depend on third-party modules.

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
- **WHEN** the user opens the Diagnostics, Models, Integrations, or Project view and leaves it without confirming an edit
- **THEN** no file, state, or project content changes.

### Requirement: Read-only diagnostics
The manager SHALL offer read-only diagnostics, both as the `hive doctor` and `hive models` text commands and as the Diagnostics, Models, Integrations, and Project views of the terminal interface, with the same content in both forms. They SHALL report each supported host's detection, version, release, and installation state; every resource whose status is other than `installed`, `retained_shared`, and `not_installed`, and any pending operation; open sessions that probably predate the installed release; the effective model and effort per role and host, marking roles with a model override and listing overrides that are not applied; the local evidence and last onboarding status of Engram, Context7, pi-subagents, and `agent-browser`; and the validity of a repository's `## Hive` section. `hive doctor`, `hive models` without a subcommand, and the views while no edit is confirmed SHALL NOT write files or state and SHALL NOT read host or provider configuration. They SHALL execute no program other than `--version` of a detected host and read-only `git` queries, except that opening an edit panel in the Models view MAY run that host's model-listing command under the Per-role model overrides requirement. Detecting `agent-browser` SHALL NOT add it to the optional capability catalog. A session marked as predating the installed release SHALL be presented as an estimate, registered hosts whose sessions cannot be observed SHALL show a restart notice instead, except a host whose documentation states that it reloads its instructions, which SHALL say so. Text read from outside the manager SHALL be stripped of control characters before it is shown.

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

## ADDED Requirements

### Requirement: Per-role model overrides
The manager SHALL let the user override the model and effort of one agent role on one registered user-scope host, through `hive models set` and `hive models reset` and through the Models view. Overrides SHALL be stored in Hive's state, SHALL NOT change release identifiers or release snapshots, and SHALL be applied by every later installation plan for that host, including updates and returns to a retained release. An override SHALL take precedence over the release's profile and the role's own effort. Applying or removing an override SHALL rewrite only that role's agent file on that host and SHALL keep the host's installation receipt consistent with the written files. Removing Hive from a host SHALL drop its overrides. An override SHALL be refused, without writing, when its host is not registered, its role is not in the host's installed release, its effort is set on a host that accepts none or is not one of `low`, `medium`, `high`, `xhigh`, `max`, and `ultra`, its model is empty, longer than 200 characters, or contains a character outside letters, digits, and `._:/@+=[]-`, or applying it would change anything on the host other than agent files. The text commands SHALL accept any model the validation allows, whether or not the host lists it. Roles SHALL be grouped by the folder of their source (`content/agents/<group>/`); setting a group SHALL write an override with only the given parts to every role of the group, replacing each role's own override, and SHALL name before confirmation the roles that lose a part; resetting a group SHALL remove the overrides of its roles. On OpenCode, a new model without a chosen effort SHALL keep the profile's effort, including a variant embedded in the profile's model, without storing that effort in the override. The Models view SHALL offer, for the model, a filterable list made of `release default`, the effective model, the host's models, and a free-text entry. The host's models SHALL come from its listing command (`codex debug models`, `opencode models`, `pi --list-models`, `grok models`, `cursor-agent models`) or, for Claude Code, from the fixed aliases `fable`, `opus`, `sonnet`, and `haiku`; the command SHALL run only when an edit panel of that host is first opened in a session, as the binary found on `PATH`, without a shell, in a neutral working directory, detached from the terminal, with stdin from the null device, with a time limit after which its process group is terminated, and with a capped output whose excess is an error. With an explicit synthetic home no host command SHALL run. Internal host model caches SHALL NOT be read.

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
- **WHEN** the user sets an effort for a role on Grok or Cursor
- **THEN** the manager refuses it, names the reason, and changes nothing.

### Requirement: Project settings section editing
The manager SHALL let the user create or edit the `## Hive` section of the `AGENTS.md` at the root of a Git repository, through `hive project set` and through the Project view, after showing the section's lines before and after the change and asking for confirmation. It SHALL change only the given keys, preserve every other line of the file, its line endings and its permissions, append a missing section at the end of the file, and create `AGENTS.md` with only the section when the file does not exist. It SHALL refuse, without writing, a duplicate section, a symbolic link or non-regular file, a file larger than 1 MiB, a directory outside a Git repository, a missing or empty required value, a value with line breaks, control characters, or characters that would be stripped when shown, an unknown key given by the user, an invalid `Delivery` or `Hive guidance` value, a base branch name Git rejects, and a file that changed after the preview. A base branch or `Specs` path that does not exist yet SHALL be reported as a warning without blocking. Values suggested from the repository SHALL be labeled as suggestions until the user confirms them. When the repository root holds a `CLAUDE.md` that does not import `@AGENTS.md`, the preview SHALL warn about it and the manager SHALL NOT modify `CLAUDE.md`. This write SHALL NOT enter Hive's transaction journal.

#### Scenario: Section added to an existing file
- **WHEN** the user saves the section for a repository whose `AGENTS.md` has none
- **THEN** the section is appended at the end and every earlier byte of the file is unchanged.

#### Scenario: File changed after the preview
- **WHEN** `AGENTS.md` changes between the preview and the confirmation
- **THEN** nothing is written and the manager says the file changed since the preview.

#### Scenario: Declined edit
- **WHEN** the user cancels the confirmation
- **THEN** `AGENTS.md` and `CLAUDE.md` are unchanged.
