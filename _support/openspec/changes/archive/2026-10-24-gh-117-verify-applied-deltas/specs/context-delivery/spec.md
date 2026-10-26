# context-delivery

## Purpose

How Hive delivers its guidance to each host: the global core through managed instruction channels, situational procedures through packaged skill resources, the validation of their references before installation, the optional OpenCode context loader, and the declared reading footprint of common scenarios.

## ADDED Requirements

### Requirement: Portable kernel and situational procedures
Hive SHALL deliver a canonical global core (`content/guidance/global.md`) through its existing managed instruction channels and situational procedures through packaged skill resources. Each host's discovery prerequisites, including Cursor's project pointer, SHALL remain explicit; delivery SHALL NOT be presented as proof of universal automatic loading. Authorization, preservation, secret handling and basic artifact placement SHALL remain in the global core without depending on a loader. Moving a procedure SHALL retain a canonical destination and an explicit read condition for its consumers. The content SHALL distinguish installation, discovery, context delivery and behavior evidence.

#### Scenario: A host without the optional loader
- **WHEN** a supported host uses the reorganized guidance without a native loader
- **THEN** the global core retains essential boundaries and directs the agent to the required packaged procedure before the dependent action.

### Requirement: Kernel references validated before installation
The manager SHALL validate the global core's packaged instruction dependencies before preparing or applying installation writes, including saved plans with internally recomputed hashes. It SHALL retain the existing exclusions for fenced examples and external URLs.

#### Scenario: Missing global core dependency
- **WHEN** the global core links a skill resource absent from the release payload
- **THEN** preparing and applying the installation reject the payload before writing native destinations.

### Requirement: Optional inactive OpenCode context resource
Hive SHALL deliver the optional OpenCode loader only as inactive managed resources at user scope. It SHALL NOT register or enable the plugin or edit user OpenCode configuration merely because Hive is installed or updated. Native activation remains an explicit user-controlled configuration action. Removal and downgrade SHALL preserve unrelated configuration and disclose the prerequisite to unregister a user-owned reference before removing the loader resources.

The resources SHALL be `hive/context-loader/index.js` and
`hive/context-loader/context-loader.json`, with no `package.json`,
under the resolved OpenCode home, outside the native auto-discovered `plugins/`
directory. Activation SHALL add the absolute installed
`<OpenCodeHome>/hive/context-loader` directory path to the user configuration’s
`plugins` array, preserving existing entries, followed by a new session.
Deactivation SHALL remove that entry or use the native `-hive.context-loader`
directive, then reload through the native host. Refresh SHALL regenerate from the
desired validated release payload with the manager, then reload the plugin.
Metadata SHALL use schema `2` and the ordered fixed-role reference
identities and SHA-256 digests of their exact canonical UTF-8 payload bytes;
it SHALL NOT grant a `skillsDirectory` capability or be read in the request loop.
Saved-plan validation SHALL recompute both the installed JS snapshot and that
derived metadata before writes. Installed local resources SHALL NOT be described
as cryptographically authenticated. New offline packages SHALL use manifest version `2` and include
`integrations/opencode/context-loader.js`. The current verifier SHALL retain
manifest version `1` support without requiring that source in old packages.
Documentation SHALL distinguish an old content release used with the current
manager from an old manager binary that cannot verify manifest version `2`.

#### Scenario: Normal installation
- **WHEN** Hive is installed for OpenCode user scope
- **THEN** its loader resources can be delivered outside native autodiscovery, with managed ownership, without activating them or changing `opencode.json(c)`.

### Requirement: Fixed-role context delivery without authority expansion
The manager SHALL generate the two fixed reference snapshots from its validated release payload into the installed plugin, serializing content as data and validating derived identity before writes. When explicitly active, the loader SHALL resolve the native agent's identifier and supply that generated snapshot for `hive-review-plan` or `hive-review-harness` before model dispatch without opening reference or metadata files in its request loop. It SHALL NOT infer prompt intent, alter tools, permissions or generation settings, or persist message changes. Source resources missing or invalid at generation SHALL reject delivery before writes; runtime missing-snapshot diagnostics SHALL be bounded and SHALL NOT claim native request abortion. The global core's obligation to report dependencies remains. Unknown roles SHALL remain unchanged.

The fixed mappings SHALL be `hive-review-plan` to
`flow-plan/references/plan-review.md` and `hive-review-harness` to
`harness-audit/SKILL.md`. The identifier SHALL be resolved from the native
agent lookup's `data.id`, then its `id`, then the hook event's agent value (also
when the lookup fails), and SHALL NOT be taken from the display name.
Deduplication SHALL affect only parts added by this plugin to the current
request, not foreign content or persistent session flags.

#### Scenario: Covered native role
- **WHEN** a context request belongs to a covered role and its generated snapshot is present
- **THEN** one copy is supplied by the loader to the outgoing request, and a later request receives that snapshot without depending on persistent session flags or reopening mutable reference paths.

#### Scenario: Reference files change after generation
- **WHEN** a local reference is edited or its directory replaced after the plugin snapshot was generated
- **THEN** the request loop supplies the previously generated snapshot, performs no live read of that reference, and changes only after regeneration and plugin reload.

### Requirement: Declared source footprint without token claims
The size check SHALL report the global core separately from explicit simple-query,
question and retained plan/build scenarios. Each scenario SHALL name its read
conditions and source files, include selected activity entrypoints and required
procedures, and count each source once. Missing required sources SHALL fail the
measurement. The simple-query scenario SHALL count the reporting reference as a
conservative upper bound; this SHALL NOT be read as a rule that every reply
requires reading it, since the global core requires no read for a greeting or a
short direct answer. Source bytes and Unicode characters SHALL NOT be presented
as actual tokens, observed native delivery or model reliability. Unmeasured
native metadata/rendering, repetition, project inputs, history and host policy
SHALL remain explicit exclusions, without a new quota.

#### Scenario: Standalone user decision
- **WHEN** the declared scenario is a user decision without an activity, delegation or artifact
- **THEN** its current source union includes the global core, reporting and questions, without unconditionally adding every packaged reference.
