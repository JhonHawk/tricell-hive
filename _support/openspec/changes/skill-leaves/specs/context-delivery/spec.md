## MODIFIED Requirements

### Requirement: Layered global core with owning skills
The global core (`content/guidance/global.md`) SHALL hold the invariants that must act in any flow or outside one without another file being read: authorization and critical effects, including the tracker writes, the close-cleanup authorization and the immediate escalation of a P0, a security finding or an exposed secret, preservation, secrets, basic artifact placement, the current-project boundary, the session-start read of local instructions, the delivery modes and what each authorizes, and the minimal line of each question rule. It SHALL also hold an index, «Rule leaves», with one line per rule leaf that states literally when to read it, naming an observable act or event, and links the leaf. A rule that applies at an identifiable moment SHALL live in exactly one rule leaf, and every flow step where it applies SHALL link that leaf. A procedure that belongs to one activity SHALL live in the skill that owns that activity, or in a reference that skill links with its read condition, each with one canonical home. No rule SHALL depend on a model choosing to load a skill from its description.

#### Scenario: A rule moved out of the global core
- **WHEN** a rule moves from the global core to a rule leaf
- **THEN** the global core keeps its minimal invariant, when it must act unread, and an index line with its read condition, and the rule text appears once under `content/`, in the leaf.

#### Scenario: A host that never selects an activity skill
- **WHEN** a supported host lists the activity skills but the model selects none of them
- **THEN** authorization, preservation, secret handling, basic artifact placement and the minimal question rule still act from the global core, and every rule leaf the agent needs is reached through the index or a flow step's link.

## ADDED Requirements

### Requirement: Rule leaves hidden from model invocation
Each rule leaf (`hive-rule-*`) and principle leaf (`hive-principle-*`, except the index) SHALL be a skill whose frontmatter sets `disable-model-invocation: true`, whose description is one line of at most 200 characters, and whose `SKILL.md` is at most 4,000 bytes, or at most 7,000 bytes when the content test names it as an exception. The language leaf `hive-typescript` SHALL stay model-invocable, with `paths` for TypeScript files, and `hive-principle-index` SHALL link it with the condition of editing those files. Every rule leaf SHALL appear exactly once in the global core's index, and every principle leaf once in `hive-principle-index`. A leaf SHALL be read by the path its index line or flow link gives, never through a host's skill-invocation tool.

#### Scenario: A leaf missing its switch
- **WHEN** a leaf skill lacks `disable-model-invocation: true` or is listed twice in an index
- **THEN** `go test ./tests/content/...` fails and names the leaf.

#### Scenario: A rule restated outside its leaf
- **WHEN** a key phrase registered for a rule leaf appears in a second file under `content/`
- **THEN** the single-home test fails and names both files.
