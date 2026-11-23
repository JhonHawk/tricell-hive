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

### Requirement: Declared source footprint without token claims

The size check SHALL report the global core separately from explicit simple-query, question, retained plan/build and E2E scenarios. Each scenario SHALL name its read conditions and source files, include the entrypoints and mandatory procedures selected by those conditions, and count each source once. A scenario that reads a flow's generated stage packet SHALL count it as one read whose sources are its members, and SHALL report its number of reads, its bytes without repeats, the members the packets bring that the scenario's earlier list did not read, and its bytes with members repeated across packets; the comparison against the earlier baseline SHALL apply to the bytes without repeats of the members the earlier list read, and SHALL NOT fail for the brought members or the repeats. Missing required sources SHALL fail measurement. The change report SHALL distinguish prose, references and executable code, and SHALL retain a before/after comparison using the same source set so a scenario-definition change is not misreported as savings. Existing scenario measurements SHALL remain identifiable. Source bytes and characters SHALL NOT be presented as actual tokens, observed native delivery or model reliability. Native metadata/rendering, repeated reads, project inputs, history and host policy SHALL remain explicit exclusions. Each skill entrypoint SHALL measure at most 16,000 bytes, each skill reference at most 24,000 bytes, and each stage packet at most 24,000 bytes and 1,000 lines; no new aggregate quota is implied.

#### Scenario: Standalone user decision
- **WHEN** the declared scenario is a user decision without an activity, delegation or artifact
- **THEN** its source union includes the global core and only the procedures required by its current read conditions, without unconditionally adding every packaged reference.

#### Scenario: Coordination reference is added
- **WHEN** E2E requires a new situational reference
- **THEN** the E2E scenario includes it, and the report does not count moving prose from an entrypoint into that reference as a reduction in the required source union.

#### Scenario: Comparable baseline
- **WHEN** the measured scenario's source list changes
- **THEN** the report separates the methodology change from the before/after measurement on matching paths and reports new sources explicitly.

#### Scenario: A route that reads stage packets
- **WHEN** the plan/build scenario reads the stage packets of its flows
- **THEN** the report gives its reads, its bytes without repeats, the brought members and its bytes with repeats, the scenario declares at most 26 reads, and only growth in the members the earlier list read can fail the comparison.

### Requirement: Layered global core with owning skills
The global core (`content/guidance/global.md`) SHALL hold the invariants that must act in any flow or outside one without another file being read: authorization and critical effects, including the tracker writes, the close-cleanup authorization and the immediate escalation of a P0, a security finding or an exposed secret, preservation, secrets, basic artifact placement, the current-project boundary, the session-start read of local instructions, the delivery modes and what each authorizes, and the minimal line of each question rule. For each of these whose procedure lives in a rule leaf, the core SHALL state only the invariant and link the leaf. It SHALL also hold an index, «Rule leaves», that links each rule leaf once, on a line that states literally when to read it, naming an observable act or event; the index SHALL list one rule leaf per line under one `###` heading per domain that `hive-core/assets/domains.json` lists, in that order, with the principle index and the code principle of verifying current state listed under `principles`. A rule that applies at an identifiable moment SHALL live in exactly one rule leaf, and every flow step where it applies SHALL link that leaf. A procedure that belongs to one activity SHALL live in the skill that owns that activity, or in a reference that skill links with its read condition, each with one canonical home. No rule SHALL depend on a model choosing to load a skill from its description. The global core SHALL measure at most 30,000 bytes, and a content test SHALL fail above that size with a message that names where a new rule goes: a rule leaf, with only its invariant in the core when it must act unread.

#### Scenario: A rule moved out of the global core
- **WHEN** a rule moves from the global core to a rule leaf
- **THEN** the global core keeps its minimal invariant, when it must act unread, and an index line with its read condition, and the rule text appears once under `content/`, in the leaf.

#### Scenario: A host that never selects an activity skill
- **WHEN** a supported host lists the activity skills but the model selects none of them
- **THEN** authorization, preservation, secret handling, basic artifact placement and the minimal question rule still act from the global core, and every rule leaf the agent needs is reached through the index or a flow step's link.

#### Scenario: The global core grows past its budget
- **WHEN** a change makes `content/guidance/global.md` larger than 30,000 bytes
- **THEN** `go test ./tests/content/...` fails, and its message says to move the rule's procedure to a leaf and keep only its invariant in the core.

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
The skill `hive-core` SHALL hold every rule leaf and every principle leaf in domain folders at its top level, `content/skills/hive-core/<domain>/<name>.md`, and SHALL NOT hold a `references/` folder. The domains SHALL be the ones `content/skills/hive-core/assets/domains.json` lists, each folder SHALL hold at least one leaf, and a leaf's name SHALL be unique across domains. Each rule leaf SHALL have no frontmatter, a first body line after its `#` title that starts with "Read" and states when to read it, and at most the 24,000-byte whole-read limit; a rule leaf's own references SHALL live under `hive-core/<domain>/<name>/` and its scripts under `hive-core/scripts/<name>/`. The principle leaves SHALL form the domain `principles`: each a file `content/skills/hive-core/principles/<name>.md` with no frontmatter, a first body line after its `#` title that starts with "Read", and at most the 24,000-byte whole-read limit, and the principle index SHALL be `content/skills/hive-core/principles/index.md`. No directory `content/skills/hive-rule-*` or `content/skills/hive-principle-*` SHALL exist. `hive-core/SKILL.md` SHALL set `disable-model-invocation: true` and SHALL NOT list or route to the leaves, since the global core's index and the principle index do. The language leaf `hive-typescript` SHALL stay model-invocable, with `paths` for TypeScript files, and the principle index SHALL link it with the condition of editing those files. Every rule leaf SHALL appear exactly once in the global core's index, on a line whose trigger is literal text of the leaf's first body line, compared in lower case with markdown links reduced to their text, and every principle leaf once in the principle index. A leaf SHALL be read by the path its index line, its flow link or its flow's stage packet gives, never through a host's skill-invocation tool.

#### Scenario: A leaf missing its switch
- **WHEN** `hive-core/SKILL.md` lacks `disable-model-invocation: true`, a leaf has frontmatter, or a leaf is listed twice in an index
- **THEN** `go test ./tests/content/...` fails and names the leaf.

#### Scenario: An index line that does not quote its leaf
- **WHEN** the trigger of a line of the global core's «Rule leaves» is not literal text of the first body line of the leaf it links
- **THEN** `go test ./tests/content/...` fails and names the line and the leaf.

#### Scenario: A rule restated outside its leaf
- **WHEN** a key phrase registered for a rule leaf appears in a second file under `content/`
- **THEN** the single-home test fails and names both files.

#### Scenario: A leaf outside its domains
- **WHEN** `content/skills/hive-core/references/` exists, a folder at the top of `hive-core` is neither a domain `domains.json` lists nor `assets` or `scripts`, a domain holds no leaf, a leaf name repeats across domains, or a directory `content/skills/hive-rule-*` or `content/skills/hive-principle-*` exists
- **THEN** `go test ./tests/content/...` fails and names it.

### Requirement: Project skill registry in AGENTS.md

A project MAY declare in a `## Skills` section of its `AGENTS.md`, outside `## Hive`, which skill files an agent reads and when: each line SHALL name an observable trigger, a path or a kind of task, and the path of a `SKILL.md` to read whole before that work. A declared workspace SHALL repeat each child repository's lines in its own `## Skills` with the child's path as prefix on both the trigger's paths and the skill's path. `audit-harness` SHALL check the section with `skill_registry.py check`, which SHALL exit non-zero naming each listed path that does not exist and each installed skill, other than those Hive installed, whose name no line lists. When onboarding a project, `audit-harness` SHALL propose the official skills of the stack the project's files show, and SHALL install only those the user approves, through the `skills` installer after the package check, each source pinned to a commit and each `SKILL.md` read before installing; it SHALL then write the section. The main thread SHALL copy into each child's brief the lines whose trigger applies to the child's task.

#### Scenario: An installed skill missing from the registry

- **GIVEN** a repository with `.agents/skills/expo-router/SKILL.md` and an `AGENTS.md` whose `## Skills` does not list it
- **WHEN** `skill_registry.py check` runs on the repository
- **THEN** it exits non-zero and names `expo-router`

#### Scenario: A build child for a mobile screen

- **GIVEN** a registry line ``- Before editing files under `apps/mobile/`: `.agents/skills/expo-router/SKILL.md` ``
- **WHEN** the main thread briefs a child whose task edits `apps/mobile/`
- **THEN** the brief carries that line

### Requirement: Stack-neutral mobile build role

Hive SHALL deliver one mobile implementation role, `hive-build-mobile`, for native and cross-platform mobile apps, which SHALL read the registry lines its brief carries and the repository's `## Skills`, SHALL build with the stack's own tooling, and SHALL verify on the device or emulator its brief assigns by serial, reporting the targets it did not exercise. Hive SHALL NOT deliver a role bound to a single mobile technology.

#### Scenario: The Kotlin Multiplatform role is retired

- **WHEN** a host is updated from a release with this change
- **THEN** it has `hive-build-mobile` and no `hive-build-kmp`

### Requirement: A rule is cited outside its home, never restated
Outside its canonical home, distributed guidance SHALL refer to a rule only by the moment it applies and a link to its home, to the exact section with an anchor when the rule is one section of a longer file, and SHALL NOT restate, summarize or paraphrase any step, condition or exception of that rule. A sentence that must act before the home is read SHALL state its invariant whole and short, and an exception that applies at one point of use SHALL stay there naming each element of the rule it changes. A section that other files cite SHALL keep its heading, which serves as its citable identifier.

#### Scenario: A flow step that needs a rule
- **WHEN** a flow step depends on a rule leaf
- **THEN** it names the moment and links the leaf or its anchored section, with no list of the leaf's steps.

### Requirement: Size limits rest on a host limit or a loading budget
Every size limit a content test enforces SHALL belong to one of three kinds, and its constant SHALL name its source. A host limit SHALL rest on a documented cut of a supported host and change only when that host changes: a skill entrypoint at most 16,000 bytes, because Claude Code keeps only the first 5,000 tokens of an invoked skill after compaction. A whole-read limit of 24,000 bytes SHALL apply to every other file a model reads in one pass: a skill reference, a rule leaf, a principle leaf, the principle index and a playbook, the same limit as a stage packet; no whole-read file SHALL have a limit of its own or an exception, and a leaf a stage packet holds SHALL also fit that packet, which the stage packet margin test keeps at least 500 bytes under its 24,000. A loading budget SHALL cap what a session loads without a read of its own, the global core (30,000 bytes), the voice text (3,072 bytes) and the language leaf `hive-typescript` (4,000 bytes, loaded by `paths` in every session that edits TypeScript), and SHALL change only by the user's decision. A failing size test SHALL name the file, its size, the limit and what to do: for an entrypoint, move a situational section to a reference linked with its read condition; for a whole-read file, split it by responsibility; for the global core, move the rule's procedure to its leaf. A change SHALL NOT raise a limit to make room for its own content, and SHALL NOT cut text a rule needs to fit a limit.

#### Scenario: A leaf past the whole-read limit
- **WHEN** a rule leaf, principle leaf, playbook or skill reference measures more than 24,000 bytes
- **THEN** `go test ./tests/content/...` fails, naming the file, its size, the limit and that it splits by responsibility.

#### Scenario: A leaf under the whole-read limit
- **WHEN** a rule leaf measures between 4,001 and 24,000 bytes
- **THEN** the size tests pass, and no exception list names it.

#### Scenario: An entrypoint past the host limit
- **WHEN** a skill's `SKILL.md` measures more than 16,000 bytes
- **THEN** `go test ./tests/content/...` fails, naming the compaction cut and that a situational section moves to a reference.
