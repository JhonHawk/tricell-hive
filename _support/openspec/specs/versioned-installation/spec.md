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

### Requirement: Optional operation capabilities
The optional capability catalog SHALL be limited to Engram, Context7, and pi-subagents. The installer SHALL NOT execute a provider installer until that provider's recipe passes a native gate for the host and platform; until then a selected capability SHALL be reported as manual, with its official source and reason, and SHALL NOT be presented as installed or verified. Development-of-Hive requirements SHALL remain separate from installation and operation requirements. Hosts, runtimes and other tools SHALL NOT be installed implicitly.

#### Scenario: Capability without a verified recipe
- **WHEN** the user selects a capability whose recipe has not passed its native gate
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
