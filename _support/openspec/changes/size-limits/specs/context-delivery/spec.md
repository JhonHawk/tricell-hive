## MODIFIED Requirements

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

## ADDED Requirements

### Requirement: Size limits rest on a host limit or a loading budget
Every size limit a content test enforces SHALL belong to one of three kinds, and its constant SHALL name its source. A host limit SHALL rest on a documented cut of a supported host and change only when that host changes: a skill entrypoint at most 16,000 bytes, because Claude Code keeps only the first 5,000 tokens of an invoked skill after compaction. A whole-read limit of 24,000 bytes SHALL apply to every other file a model reads in one pass: a skill reference, a rule leaf, a principle leaf and a playbook, the same limit as a stage packet; no whole-read file SHALL have a limit of its own or an exception, and a leaf a stage packet holds SHALL also fit that packet, which the stage packet margin test keeps at least 500 bytes under its 24,000. A loading budget SHALL cap what a session loads without a read of its own, the global core (30,000 bytes), the voice text (3,072 bytes) and the language leaf `hive-typescript` (4,000 bytes, loaded by `paths` in every session that edits TypeScript), and SHALL change only by the user's decision. A failing size test SHALL name the file, its size, the limit and what to do: for an entrypoint, move a situational section to a reference linked with its read condition; for a whole-read file, split it by responsibility; for the global core, move the rule's procedure to its leaf. A change SHALL NOT raise a limit to make room for its own content, and SHALL NOT cut text a rule needs to fit a limit.

#### Scenario: A leaf past the whole-read limit
- **WHEN** a rule leaf, principle leaf, playbook or skill reference measures more than 24,000 bytes
- **THEN** `go test ./tests/content/...` fails, naming the file, its size, the limit and that it splits by responsibility.

#### Scenario: A leaf under the whole-read limit
- **WHEN** a rule leaf measures between 4,001 and 24,000 bytes
- **THEN** the size tests pass, and no exception list names it.

#### Scenario: An entrypoint past the host limit
- **WHEN** a skill's `SKILL.md` measures more than 16,000 bytes
- **THEN** `go test ./tests/content/...` fails, naming the compaction cut and that a situational section moves to a reference.
