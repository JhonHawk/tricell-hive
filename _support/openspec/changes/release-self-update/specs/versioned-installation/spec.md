# versioned-installation delta

## ADDED Requirements

### Requirement: Complete update from a published release
For a migrated release installation, `hive update` without an explicit source SHALL resolve a stable published release from Hive's fixed origin and update the executable and all registered user-scope hosts without requiring Git, Go or a checkout. The result SHALL identify the previous and target product versions and SHALL report success only after both the core and command activation complete. It SHALL preserve the core's conflict and shared-resource protections and SHALL NOT execute optional provider installation steps implicitly.

#### Scenario: Complete release update
- **WHEN** a migrated installation runs update and the operator confirms a compatible newer stable release
- **THEN** the retained target package supplies the executable and content, registered hooks point to that package's real executable, and the stable user command selects that executable.

#### Scenario: Exact version and compatibility
- **WHEN** the operator supplies `--version X.Y.Z`
- **THEN** update selects exactly that version, shows an explicit downgrade when applicable, and refuses an absent or incompatible update protocol or state schema before applying changes; default selection never implicitly downgrades or installs a prerelease.

### Requirement: Verified update preparation and frozen plans
Release update SHALL verify HTTPS and permitted hosts on every redirect, bounded downloads, archive checksum, entry paths/types/limits, platform and bootstrap executable identity before executing a downloaded binary. The target executor SHALL validate the complete package and installed state before applying. The launcher SHALL invoke only a fixed internal protocol with bounded structured input/output, without a shell or executable commands supplied by package metadata. A preview or cancellation SHALL leave active installation, state, command and persistent package storage unchanged and remove its disposable inputs. Without a terminal, a non-preview update without `--out` SHALL fail before downloading. `--out` MAY retain a verified package and private parent plan; it SHALL NOT activate them.

#### Scenario: Invalid remote input
- **WHEN** a download, redirect, checksum, package identity, platform, bootstrap protocol or archive boundary is invalid
- **THEN** update rejects it before executing its binary or changing active installation files.

#### Scenario: Offline saved plan
- **WHEN** `hive apply --plan` receives a complete-update plan
- **THEN** it uses only the pinned package, executor, core plan and fingerprints, without resolving latest or downloading; changed bytes or stale state/destinations are refused before core writes.

### Requirement: Managed stable command and package migration
A managed release installation SHALL retain immutable versioned packages and register `~/.local/bin/hive` as its stable command without editing PATH. The online bootstrap and offline package installer SHALL support migration to that command using the same activation coordinator. Source-checkout installation SHALL NOT implicitly adopt or create the release command. A non-owned command destination SHALL be a conflict. Removal SHALL remove only an owned unchanged command and SHALL preserve retained packages and user modifications.

#### Scenario: Initial migration
- **WHEN** the operator runs the updated installer for an existing package installation and the command destination is absent or owned unchanged
- **THEN** one confirmed installation migrates hooks and command to the new package, and repeating it is idempotent.

#### Scenario: Foreign executable
- **WHEN** the stable command destination contains an unowned file or link
- **THEN** installation reports the conflict and a usable package path without overwriting or adopting the destination.

### Requirement: Package-bound preview executable destination
Release preparation SHALL distinguish its temporary preparation executable from the final executable destination bound to the verified package manifest. A preview MAY render hooks for that final destination before the package exists there, but SHALL NOT write or retain that destination, put a temporary executable path in hooks, or accept a generic unverified executable-path override. Before applying hooks, the target package SHALL exist at the pinned canonical destination and its executable SHALL satisfy the regular-file, path and digest validations. Source and ordinary install validation SHALL retain their current executable checks.

#### Scenario: Preview before first package retention
- **WHEN** a compatible target package is prepared temporarily and its final store directory does not exist
- **THEN** preview renders the final hook executable paths without persistent package writes, and application refuses until the exact package is materialized and verified there.

### Requirement: Update coordination and commit-boundary recovery
A versioned parent journal independent of the core state schema SHALL bind the target package, verified executor, frozen plan, command fingerprint and core transaction before mutation. Its association with durable child commit evidence SHALL survive a lost process response; required journals and packages SHALL remain retained while pending. Before core commit, recovery SHALL recover incomplete core writes and retain the previous command. After core commit, recovery SHALL preserve committed integrations and finish command activation using the retained target executor without network access. It SHALL NOT claim rollback of a committed core. Altered command or package bytes SHALL be conflicts rather than overwritten or accepted. A pending update SHALL block competing aware core mutators; lock acquisition SHALL have one documented order. Reads and recovery routing SHALL inspect the parent before decoding an unsupported core schema.

#### Scenario: Crash before child commit
- **WHEN** update stops with incomplete child writes and no durable commit
- **THEN** recovery uses the recorded executor and association to recover those writes and leaves the previous command selected.

#### Scenario: Crash after commit before activation
- **WHEN** the child commits and the process stops before recording its response or activating the command
- **THEN** recovery determines commit from the bound durable journal and completes activation with the target executor, including when the previous launcher cannot read the new core schema.

#### Scenario: Concurrent operation or changed link
- **WHEN** another aware mutator encounters the pending parent or activation sees a different command fingerprint
- **THEN** it does not overwrite the competing state and explains the pending recovery or path conflict.

### Requirement: Unified terminal update experience
The Update terminal view SHALL offer Release and Source modes using the same preparation, confirmation, frozen plans, activation and recovery contracts as the text commands. It SHALL show current/target versions for a release, preserve existing terminal component patterns and expose actionable no-host, network, conflict and pending-recovery states. Success SHALL explain session reload requirements and report a still-running viewer with a different revision without restarting it.

#### Scenario: Release view and retry
- **WHEN** release preparation fails and the operator retries or switches to Source
- **THEN** the view prepares through the shared backend without applying incomplete bytes or silently changing its source.

## MODIFIED Requirements

### Requirement: Pi user-scope install declares pi-subagents
When a user-scope plan includes the Pi host, the manager SHALL leave `pi-subagents` selectable by Pi's native package mechanism without rewriting unrelated `settings.json` entries. If the user settings do not declare the package by npm name, apply SHALL run `pi install npm:pi-subagents` with `PI_CODING_AGENT_DIR` set to the plan's Pi home. If the package is already declared and the extensions filter is omitted, the plan SHALL omit install and SHALL NOT change the existing pin. A declaration with `extensions: []`, or a non-empty filter that does not demonstrably include the extension, SHALL be a conflict or an unverified-availability report, not success, and SHALL NOT change settings. Missing files or a pin/version mismatch on a pre-existing declaration SHALL be reported without reinstalling. Recover SHALL remove the package only when this operation added the exact source string and that string is still present. Tests SHALL use a fake `pi` and a synthetic home and SHALL NOT call npm or write the real user settings. For complete release updates, this requirement SHALL use a read-only preflight instead of scheduling Pi installation: when the existing declaration is absent or requires repair, preparation and application SHALL refuse before core writes and command activation, preserve settings, and identify separate repair through `hive install`. A saved release-update plan SHALL NOT introduce a Pi installation step. Install and committed-source update SHALL retain the behavior above.

#### Scenario: Package absent
- **WHEN** a user-scope install or committed-source update plan includes `pi` and the planned Pi settings do not declare `pi-subagents`
- **THEN** the plan shows install of `npm:pi-subagents`, apply adds that source, and other package entries remain.

#### Scenario: Package already declared
- **WHEN** the planned Pi settings already declare `pi-subagents` with a loadable extension
- **THEN** the plan omits install, apply does not invoke `pi install`, and recover does not remove the entry.

#### Scenario: Empty extensions filter
- **WHEN** the planned Pi settings declare `pi-subagents` with `extensions: []`
- **THEN** the plan reports a conflict and apply does not change settings.


#### Scenario: Pi declaration missing during release update
- **WHEN** a complete release update targets an already registered Pi host whose pi-subagents declaration is absent or requires repair
- **THEN** preparation or application refuses before core writes and command activation, leaves settings unchanged, does not invoke Pi, and explains separate repair through `hive install`.


### Requirement: Update from a committed revision
With explicit `--source DIR` or `--rev REV`, the manager SHALL build an install plan from the content of one resolved commit, excluding uncommitted changes, for every consumer already installed in user scope. `--rev` alone SHALL use the current directory; `--source` alone SHALL use HEAD. These flags SHALL conflict with `--version`. Source update SHALL neither download releases nor build or activate executables. A development binary invoked without an explicit source selector SHALL refuse release replacement before downloading and explain the source route. Source update SHALL apply a plan that changes files only after interactive confirmation, and without a terminal SHALL only preview or save that plan for a separate `apply`. A plan that changes no file MAY be applied without confirmation, since its only write is the source-commit record; when a plan file is requested, the plan SHALL still be saved. It SHALL remove temporary extraction on every return path. The offline package SHALL NOT require Git.

#### Scenario: Uncommitted changes in the checkout
- **WHEN** the checkout has uncommitted changes under `content/` and the user runs `hive update --source .`
- **THEN** the applied release contains the committed content only and the executable remains unchanged.

#### Scenario: No terminal
- **WHEN** source update runs without a terminal and without `--dry-run` or `--out`
- **THEN** it changes nothing and names the options that preview or save the plan.

#### Scenario: Invalid revision
- **WHEN** the revision starts with `-`, names no commit, or the source is not a Git checkout
- **THEN** the command fails before extracting or writing anything.

#### Scenario: Development binary without explicit source
- **WHEN** a development binary receives bare `hive update`
- **THEN** it explains the explicit source command without downloading or replacing itself.
