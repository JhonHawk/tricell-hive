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
