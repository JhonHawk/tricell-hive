## MODIFIED Requirements

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
