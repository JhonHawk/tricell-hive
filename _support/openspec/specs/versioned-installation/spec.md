# versioned-installation

## Purpose

Versioned, consented installation of Hive from an offline package or a verified online bootstrap, with recoverable transactions and optional capabilities offered as manual instructions until each passes a native gate.

## Requirements

### Requirement: Product version and content identity
The installer SHALL distinguish the Hive product version, package schema, artifact digest, and historical payload release ID. Product labels SHALL NOT change historical payload hashes. A published product version SHALL NOT be rebound to different artifact content in the installation's version index.

#### Scenario: Identical payload with a new product version
- **WHEN** a selected consumer installs a new product version with identical payload bytes
- **THEN** its requested-version receipt changes transactionally while physical payload identities remain unchanged.

#### Scenario: Legacy installation
- **WHEN** no trustworthy product-version receipt exists
- **THEN** status shows the historical content identity without inventing a product version.

### Requirement: Explicit destinations and shared resources
The installer SHALL require explicit confirmation of selected hosts and every additional registered consumer required by a changed shared resource. Deselection SHALL NOT uninstall a host. Shared native discovery SHALL NOT be presented as host isolation.

#### Scenario: Partial shared update
- **WHEN** a selected host requires changes to a resource also owned by an unselected consumer
- **THEN** the installer explains the required cohort and performs no update unless the user confirms all affected consumers.

### Requirement: Deleted managed content is reinstalled
The system SHALL treat managed content that the user deleted by hand as absent: `hive install` and `hive update` SHALL recreate it as a fresh installation would, and removal SHALL complete without writing when what it would delete is already gone. The system SHALL keep refusing content that was edited rather than deleted.

#### Scenario: Deleted managed file
- **WHEN** a skill, agent or link that Hive installed was deleted and the user runs `hive install` or `hive update`
- **THEN** the file is written again with its managed bytes and mode, and `hive status` no longer reports drift for it.

#### Scenario: Deleted managed block
- **WHEN** the Hive or voice block was removed from a user file, or the user file that held it was deleted
- **THEN** `hive install` adds the block again, creating the file when needed, and preserves the user's text outside the block.

#### Scenario: Removal of deleted content
- **WHEN** the user removes a host whose managed content was already deleted
- **THEN** removal completes without writing and drops the record; a shared resource still used by another host stays absent until the next install or update.

#### Scenario: Edited content still refused
- **WHEN** managed content was edited, its permissions changed, or its markers were damaged
- **THEN** install, update and remove refuse with a message naming the path, and preserve the edit.

### Requirement: Pi user-scope install declares pi-subagents
When a user-scope plan includes the Pi host, the manager SHALL leave `pi-subagents` selectable by Pi's native package mechanism without rewriting unrelated `settings.json` entries. If the user settings do not declare the package by npm name, apply SHALL run `pi install npm:pi-subagents@0.74.0` with `PI_CODING_AGENT_DIR` set to the plan's Pi home. If the package is already declared and the extensions filter is omitted, the plan SHALL omit install and SHALL NOT change the existing pin. A declaration with `extensions: []`, or a non-empty filter that does not demonstrably include the extension, SHALL be a conflict or an unverified-availability report, not success, and SHALL NOT change settings. Missing files or a pin/version mismatch on a pre-existing declaration SHALL be reported without reinstalling. Recover SHALL remove the package only when this operation added the exact source string and that string is still present. Tests SHALL use a fake `pi` and a synthetic home and SHALL NOT call npm or write the real user settings.

#### Scenario: Package absent
- **WHEN** a user-scope install or update plan includes `pi` and the planned Pi settings do not declare `pi-subagents`
- **THEN** the plan shows install of `npm:pi-subagents@0.74.0`, apply adds that source, and other package entries remain.

#### Scenario: Package already declared
- **WHEN** the planned Pi settings already declare `pi-subagents` with a loadable extension
- **THEN** the plan omits install, apply does not invoke `pi install`, and recover does not remove the entry.

#### Scenario: Empty extensions filter
- **WHEN** the planned Pi settings declare `pi-subagents` with `extensions: []`
- **THEN** the plan reports a conflict and apply does not change settings.

### Requirement: Optional operation capabilities
The optional capability catalog SHALL be limited to Engram, Context7, and pi-subagents. Engram and Context7 SHALL remain manual until each recipe passes a native gate for the host and platform; a selected capability without a verified recipe SHALL be reported as manual, with its official source and reason, and SHALL NOT be presented as installed or verified. pi-subagents on a user-scope Pi host follows the Pi user-scope install declares pi-subagents requirement instead of that manual path. Development-of-Hive requirements SHALL remain separate from installation and operation requirements. Hosts, runtimes and other tools SHALL NOT be installed implicitly.

#### Scenario: Capability without a verified recipe
- **WHEN** the user selects Engram or Context7, whose recipes have not passed their native gate
- **THEN** the core installs, no provider process runs, and the result lists the capability as manual with its official instructions and a non-zero exit status.

### Requirement: External operation recovery
The installer SHALL persist optional steps in a parent onboarding journal before the core transaction and distinguish core commit from optional step completion. Unknown outcomes SHALL block further changes until reconciliation. Recovery SHALL preserve pre-existing tools, memory and unrelated configuration.

#### Scenario: Crash between core commit and optional steps
- **WHEN** the installer stops after the core transaction commits but before optional steps are recorded
- **THEN** recovery keeps the committed core and does not run any optional step.

### Requirement: Verified online bootstrap and offline installation
The online entry SHALL verify its downloaded manager before executing it, then validate and safely extract the package. The local offline package SHALL remain usable without Go, Git, GitHub CLI, or network downloads for the core. Production downloads SHALL use the configured HTTPS origin and bounded inputs.

#### Scenario: Invalid archive
- **WHEN** the package checksum, archive paths, entry types, or limits fail validation
- **THEN** installation does not begin and no entry is written outside the task-owned extraction directory.

#### Scenario: Cancellation or unavailable terminal
- **WHEN** the user cancels before application or a controlling terminal is unavailable
- **THEN** no persistent installation changes occur; the online bootstrap removes only its disposable temporary download.

### Requirement: Recovery executable retention
An online installation SHALL retain its verified compatible manager and package privately after confirmation and before applying changes, and SHALL identify how to recover from a later terminal without network access.

#### Scenario: Interrupted online installation
- **WHEN** the bootstrap process exits during installation
- **THEN** the retained manager and durable operation receipts remain available for recovery.

### Requirement: Update from a committed revision
From a Git checkout of Hive, the manager SHALL build an install plan from the content of one resolved commit, excluding uncommitted changes, for every consumer already installed in user scope. It SHALL apply a plan that changes files only after interactive confirmation, and without a terminal SHALL only preview or save that plan for a separate `apply`. A plan that changes no file MAY be applied without confirmation, since its only write is the source-commit record; when a plan file is requested, the plan SHALL still be saved. It SHALL remove its temporary extraction on every return path. The offline package SHALL NOT require Git.

#### Scenario: Uncommitted changes in the checkout
- **WHEN** the checkout has uncommitted changes under `content/` and the user runs `hive update`
- **THEN** the applied release contains the committed content only.

#### Scenario: No terminal
- **WHEN** `hive update` runs without a terminal and without `--dry-run` or `--out`
- **THEN** it changes nothing and names the options that preview or save the plan.

#### Scenario: Invalid revision
- **WHEN** the revision starts with `-`, names no commit, or the source is not a Git checkout
- **THEN** the command fails before extracting or writing anything.

### Requirement: Retained release listing
The manager SHALL list every retained release snapshot with its identifier, recording time, the source commits recorded for it, and the consumers where it is installed, without modifying state. Applying a plan that carries a source commit SHALL record that commit beside the release, including when the plan changes no file, without changing the snapshot or its identifier.

#### Scenario: Release installed before commit recording
- **WHEN** a retained release has no commit record
- **THEN** it is listed with an empty commit list.

#### Scenario: Same content from two commits
- **WHEN** two commits produce the same release identifier and both are applied
- **THEN** the release lists both commits once each.

### Requirement: Optional voice layer
The manager SHALL offer an optional voice layer, off by default, as a managed span separate from the Hive guidance block in each registered user-scope instruction file. The voice SHALL NOT change release identifiers or the communication rules, and its text SHALL limit itself to the messages the user reads in the main conversation. Installing a new release SHALL keep the chosen voice and regenerate its span when the voice source changed. Removing Hive from the last consumer of a file SHALL remove that file's voice span.

#### Scenario: Voice set and removed
- **WHEN** the user sets a voice and later turns it off
- **THEN** each instruction file returns byte for byte to its content before the voice was set.

#### Scenario: Release update with a voice
- **WHEN** a new release is installed while a voice is active
- **THEN** the voice span remains, regenerated with the same choice if its source text changed.

#### Scenario: Edited voice span
- **WHEN** the user edited the voice span by hand
- **THEN** any operation on that file reports a conflict and preserves the edit.

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

### Requirement: Deployed skill resource links resolve to installed paths
The installer SHALL rewrite every inline `skill:owner/path` Markdown link outside code fences, in deployed role files and in the deployed global guidance block, to the resource's path in the skills directory for the install scope: the absolute path under `<home>/.agents/skills` for user scope, and the path relative to the project root of the host's project skills directory for project scope. Source content SHALL keep the `skill:` locator. A file shared by several hosts SHALL receive identical bytes for each of them, or the plan SHALL fail before any write.

#### Scenario: Role link at user scope
- **WHEN** a role whose source links `[browser automation](skill:flow-build/references/browser-automation.md)` is installed or updated at user scope
- **THEN** the deployed role file, including a Codex TOML role, links `<home>/.agents/skills/flow-build/references/browser-automation.md` and contains no `skill:` link

#### Scenario: Project scope stays portable
- **WHEN** the same role is installed at project scope for Claude
- **THEN** the deployed role links `.claude/skills/flow-build/references/browser-automation.md` and contains no absolute home path

#### Scenario: Unchanged second update
- **WHEN** `hive update` runs twice from the same committed revision
- **THEN** the second plan proposes no writes for the rewritten roles or block

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
