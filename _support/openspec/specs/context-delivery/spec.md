# context-delivery

## Purpose

How Hive delivers its guidance to each host: the global core through managed instruction channels, situational procedures through packaged skill resources, the validation of their references before installation, the optional OpenCode context loader, and the declared reading footprint of common scenarios.

## Requirements

### Requirement: Portable kernel and situational procedures
Hive SHALL deliver a canonical global core (`content/guidance/global.md`) through its existing managed instruction channels and situational procedures through packaged skill resources. Each host's discovery prerequisites, including Cursor's project pointer, SHALL remain explicit; delivery SHALL NOT be presented as proof of universal automatic loading. Authorization, preservation, secret handling and basic artifact placement SHALL remain in the global core without depending on a loader. Moving a procedure SHALL retain a canonical destination and an explicit read condition for its consumers. The content SHALL distinguish installation, discovery, context delivery and behavior evidence.

#### Scenario: A host without the optional loader
- **WHEN** a supported host uses the reorganized guidance without a native loader
- **THEN** the global core retains essential boundaries and directs the agent to the required packaged procedure before the dependent action.

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

The size check SHALL report the global core separately from explicit simple-query, question, retained plan/build and E2E scenarios. Each scenario SHALL name its read conditions and source files, include the entrypoints and mandatory procedures selected by those conditions, and count each source once. Missing required sources SHALL fail measurement. The change report SHALL distinguish prose, references and executable code, and SHALL retain a before/after comparison using the same source set so a scenario-definition change is not misreported as savings. Existing scenario measurements SHALL remain identifiable. Source bytes and characters SHALL NOT be presented as actual tokens, observed native delivery or model reliability. Native metadata/rendering, repeated reads, project inputs, history and host policy SHALL remain explicit exclusions. Each skill entrypoint and reference SHALL retain the existing 16,000-byte limit; no new aggregate quota is implied.

#### Scenario: Standalone user decision
- **WHEN** the declared scenario is a user decision without an activity, delegation or artifact
- **THEN** its source union includes the global core and only the procedures required by its current read conditions, without unconditionally adding every packaged reference.

#### Scenario: Coordination reference is added
- **WHEN** E2E requires a new situational reference
- **THEN** the E2E scenario includes it, and the report does not count moving prose from an entrypoint into that reference as a reduction in the required source union.

#### Scenario: Comparable baseline
- **WHEN** the measured scenario's source list changes
- **THEN** the report separates the methodology change from the before/after measurement on matching paths and reports new sources explicitly.

### Requirement: Layered global core with owning skills
The global core (`content/guidance/global.md`) SHALL hold the invariants that must act in any flow or outside one without another file being read: authorization and critical effects, including the tracker writes, the close-cleanup authorization and the immediate escalation of a P0, a security finding or an exposed secret, preservation, secrets, basic artifact placement, the current-project boundary, the session-start read of local instructions, the delivery modes and what each authorizes, and the minimal line of each question rule. It SHALL also hold an index, «Rule leaves», with one line per rule leaf that states literally when to read it, naming an observable act or event, and links the leaf. A rule that applies at an identifiable moment SHALL live in exactly one rule leaf, and every flow step where it applies SHALL link that leaf. A procedure that belongs to one activity SHALL live in the skill that owns that activity, or in a reference that skill links with its read condition, each with one canonical home. No rule SHALL depend on a model choosing to load a skill from its description.

#### Scenario: A rule moved out of the global core
- **WHEN** a rule moves from the global core to a rule leaf
- **THEN** the global core keeps its minimal invariant, when it must act unread, and an index line with its read condition, and the rule text appears once under `content/`, in the leaf.

#### Scenario: A host that never selects an activity skill
- **WHEN** a supported host lists the activity skills but the model selects none of them
- **THEN** authorization, preservation, secret handling, basic artifact placement and the minimal question rule still act from the global core, and every rule leaf the agent needs is reached through the index or a flow step's link.

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

### Requirement: Bounded native coordination observation

A behavior observation SHALL identify the exact candidate installed through the existing delivery mechanism, the host and model identities, evidence of selected source reading, owner/verifier assignments, human interventions, artifacts and terminal outcome. The newly authorized E2E case SHALL be one native run bounded by a total 20-minute deadline including setup and process termination, with no automatic retry or second host run. It SHALL exercise two independent units, a justified verifier correction, a controlled owner replacement and verified local integration in a disposable repository, without remote publication. Its launcher SHALL explicitly configure and test the native skills/delegation/communication permissions and session persistence the case needs, without inheriting assumptions from another case name or expanding every pilot's permissions. Injected fixture faults SHALL be labeled and SHALL invalidate affected evidence. A single run SHALL NOT establish reliability, qualification or cross-host support.

Temporary installation SHALL require its own covered effect, preserve the current trial and unrelated state, and use the manager's conflict-aware plans for installation and restoration. The pilot SHALL preserve authentication and native histories, isolate fictitious memory and verify cleanup of only its own resources. It SHALL NOT silently overwrite later changes to meet the deadline. If restoration or behavior evidence is incomplete, its status SHALL remain explicit and the affected acceptance criterion SHALL remain unverified.

#### Scenario: Catalog without source reading
- **WHEN** the host lists a skill or role but the trace contains no supported evidence that its relevant instructions were read
- **THEN** the observation does not report that instruction as exercised.

#### Scenario: Deadline or incomplete return
- **WHEN** the deadline is reached or a required behavior is not observed
- **THEN** the case stops its model work and known process groups, preserves necessary evidence, reports failed/blocked/inconclusive as supported, and launches no second attempt without renewed authorization.

### Requirement: Rule leaves hidden from model invocation
Each rule leaf (`hive-rule-*`) and principle leaf (`hive-principle-*`, except the index) SHALL be a skill whose frontmatter sets `disable-model-invocation: true`, whose description is one line of at most 200 characters, and whose `SKILL.md` is at most 4,000 bytes, or at most 7,000 bytes when the content test names it as an exception. The language leaf `hive-typescript` SHALL stay model-invocable, with `paths` for TypeScript files, and `hive-principle-index` SHALL link it with the condition of editing those files. Every rule leaf SHALL appear exactly once in the global core's index, and every principle leaf once in `hive-principle-index`. A leaf SHALL be read by the path its index line or flow link gives, never through a host's skill-invocation tool.

#### Scenario: A leaf missing its switch
- **WHEN** a leaf skill lacks `disable-model-invocation: true` or is listed twice in an index
- **THEN** `go test ./tests/content/...` fails and names the leaf.

#### Scenario: A rule restated outside its leaf
- **WHEN** a key phrase registered for a rule leaf appears in a second file under `content/`
- **THEN** the single-home test fails and names both files.
