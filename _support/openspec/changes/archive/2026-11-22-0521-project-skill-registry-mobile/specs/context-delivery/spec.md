## ADDED Requirements

### Requirement: Project skill registry in AGENTS.md

A project MAY declare in a `## Skills` section of its `AGENTS.md`, outside `## Hive`, which skill files an agent reads and when: each line SHALL name an observable trigger, a path or a kind of task, and the path of a `SKILL.md` to read whole before that work. A declared workspace SHALL repeat each child repository's lines in its own `## Skills` with the child's path as prefix on both the trigger's paths and the skill's path. `harness-audit` SHALL check the section with `skill_registry.py check`, which SHALL exit non-zero naming each listed path that does not exist and each installed skill, other than those Hive installed, whose name no line lists. When onboarding a project, `harness-audit` SHALL propose the official skills of the stack the project's files show, and SHALL install only those the user approves, through the `skills` installer after the package check, each source pinned to a commit and each `SKILL.md` read before installing; it SHALL then write the section. The main thread SHALL copy into each child's brief the lines whose trigger applies to the child's task.

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
