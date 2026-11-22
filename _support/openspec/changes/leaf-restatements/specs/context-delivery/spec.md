## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: A rule is cited outside its home, never restated
Outside its canonical home, distributed guidance SHALL refer to a rule only by the moment it applies and a link to its home, to the exact section with an anchor when the rule is one section of a longer file, and SHALL NOT restate, summarize or paraphrase any step, condition or exception of that rule. A sentence that must act before the home is read SHALL state its invariant whole and short, and an exception that applies at one point of use SHALL stay there naming each element of the rule it changes. A section that other files cite SHALL keep its heading, which serves as its citable identifier.

#### Scenario: A flow step that needs a rule
- **WHEN** a flow step depends on a rule leaf
- **THEN** it names the moment and links the leaf or its anchored section, with no list of the leaf's steps.
