## REMOVED Requirements

### Requirement: Layered global core with a shared core skill
Retired because the shared `hive-core` skill was dissolved into the global guidance in `5e6e8907` (2026-10-31); its rules now live in the global core, and the replacement requirement below states that.

## ADDED Requirements

### Requirement: Layered global core with owning skills
The global core (`content/guidance/global.md`) SHALL hold every rule that applies in any flow or outside one, such as authorization, preservation, secrets, questions and routes, delivery, project settings, persistent memory and scope changes, because each host loads it with its system instructions and keeps it after compaction. Each read pointer in it SHALL state literally when to read its target, naming an observable act or event, such as "Before a commit, pull request, or merge, read …". A procedure that belongs to one activity SHALL live in the skill that owns that activity, or in a reference that skill links with its read condition, each with one canonical home. No rule SHALL depend on a model choosing to load a skill from its description; there is no shared core skill.

#### Scenario: A rule moved out of the global core
- **WHEN** a procedure moves from the global core to an owning skill
- **THEN** the global core keeps the rule's minimal invariant, or a literal read condition for it, and the procedure text appears once under `content/`, in its destination.

#### Scenario: A host that never selects an activity skill
- **WHEN** a supported host lists the activity skills but the model selects none of them
- **THEN** authorization, preservation, secret handling, questions, delivery and basic artifact placement still act from the global core, and every procedure the agent needs is reached through a read condition stated in the global core or in the active skill.
