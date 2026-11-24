# context-delivery

## Purpose

What the manager installs for each host and what it refuses before writing: validated instruction references, the optional inactive OpenCode loader, fixed-role context, the Antigravity rules and roles, and the skill registry check.

## Requirements

### Requirement: Kernel references validated before installation
The manager SHALL validate the global core's packaged instruction dependencies, including Markdown heading anchors (`#fragment`) on relative, in-file and `skill:` links across the global core, roles and skills, before preparing or applying installation writes, including saved plans with internally recomputed hashes. It SHALL retain the existing exclusions for fenced examples and external URLs.

#### Scenario: Missing global core dependency
- **WHEN** the global core links a skill resource absent from the release payload
- **THEN** preparing and applying the installation reject the payload before writing native destinations.

#### Scenario: Broken heading anchor in an instruction link
- **WHEN** the global core, a role or a skill links a Markdown file with a `#fragment` that matches no ATX heading slug in the target file
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
`flow-plan/SKILL.md` and `hive-review-harness` to
`audit-harness/SKILL.md`. The identifier SHALL be resolved from the native
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

### Requirement: Antigravity global core in size-bounded always-on rules
For the Antigravity host, Hive SHALL deliver the global core as one or more Hive-owned rule files `<AntigravityHome>/rules/hive-NN.md` under the resolved `~/.gemini/config`, each beginning with the frontmatter `trigger: always_on` and each at or below 24,000 bytes, because Antigravity truncates every rule file above that size, the global ones included. The parts SHALL be cut only at `##` headings outside fenced code blocks, in order, so that their managed texts joined in order reproduce the rewritten global core; a section that alone exceeds the limit SHALL fail the plan with an error naming it. The voice block SHALL appear only in the first part, and setting or changing the voice SHALL change only that part. Hive SHALL NOT write `~/.gemini/GEMINI.md` for this host.

#### Scenario: Guidance larger than one rule file
- **WHEN** a user-scope plan for Antigravity renders a global core larger than one rule file allows
- **THEN** it proposes as many `hive-NN.md` parts as needed, each with the `always_on` frontmatter and at or below 24,000 bytes, and their managed texts joined in order equal the rewritten global core.

#### Scenario: Part count changes on update
- **WHEN** an update renders the global core into fewer or more parts than the recorded installation
- **THEN** it writes the new parts and removes only the surplus Hive-owned parts, leaving every other file in `rules/` unchanged.

### Requirement: Antigravity roles inherit the session model
For the Antigravity host, Hive SHALL render each canonical role as `<AntigravityHome>/agents/<role>.md` with `name`, `description` and `model: "inherit"` unless the user set another model with `hive models set`, and SHALL NOT render a `tools` list or a reasoning effort. Read-only roles SHALL rely on their role instructions, since Antigravity offers no read-only field.

#### Scenario: Observe role rendered
- **WHEN** a role with `observe` access is rendered for Antigravity
- **THEN** its file carries `model: "inherit"` and no `tools` or effort field, and its body keeps the role's read-only instructions.

### Requirement: Project skill registry check

`skill_registry.py check` of `audit-harness` SHALL read the `## Skills` section of a repository's `AGENTS.md`, outside `## Hive`, where each line names an observable trigger and the path of a `SKILL.md`. It SHALL exit non-zero naming each listed path that does not exist and each installed skill, other than those Hive installed, whose name no line lists.

#### Scenario: An installed skill missing from the registry

- **GIVEN** a repository with `.agents/skills/expo-router/SKILL.md` and an `AGENTS.md` whose `## Skills` does not list it
- **WHEN** `skill_registry.py check` runs on the repository
- **THEN** it exits non-zero and names `expo-router`
