## ADDED Requirements

### Requirement: Layered global core with a shared core skill
The global core (`content/guidance/global.md`) SHALL hold only the minimal invariant of each rule that must act without any skill or reference being read, and each read pointer in it SHALL state literally when to read its target, naming an observable act or event, such as "Before a commit, pull request, or merge, read …". Procedures that act only after such a condition SHALL live in the shared `hive-core` skill or in the skill that owns the activity, each with one canonical home. The shared references used across activities, `questions`, `scope-changes`, `effect-authorization` and `project-settings`, SHALL live in `hive-core/references/`. A rule SHALL NOT depend on a model choosing to load `hive-core` from its description.

#### Scenario: A rule moved out of the global core
- **WHEN** a procedure moves from the global core to `hive-core` or an owning skill
- **THEN** the global core keeps the rule's minimal invariant, or a literal read condition for it, and the procedure text appears once under `content/`, in its destination.

#### Scenario: A host that does not load hive-core by description
- **WHEN** a supported host lists `hive-core` but the model never selects it
- **THEN** authorization, preservation, secret handling and basic artifact placement still act from the global core, and every procedure the agent needs is reached through a read condition stated in the global core or in the active skill.
