## MODIFIED Requirements

### Requirement: Rule leaves hidden from model invocation
Each rule leaf SHALL be a file `content/skills/hive-core/references/<name>.md` with no frontmatter, whose first body line after its `#` title starts with "Read" and states when to read it, and which is at most 4,000 bytes, or at most 7,000 bytes when the content test names it as an exception; a rule leaf's own references and scripts SHALL live under `hive-core/references/<name>/` and `hive-core/scripts/<name>/`. The skill `hive-core` SHALL be the only skill that holds rule leaves; its `SKILL.md` SHALL set `disable-model-invocation: true` and SHALL NOT list or route to the leaves, since the global core's index does. Each principle leaf (`hive-principle-*`, except the index) SHALL be a skill whose frontmatter sets `disable-model-invocation: true`, whose description is one line of at most 200 characters, and whose `SKILL.md` is at most 4,000 bytes. The language leaf `hive-typescript` SHALL stay model-invocable, with `paths` for TypeScript files, and `hive-principle-index` SHALL link it with the condition of editing those files. Every rule leaf SHALL appear exactly once in the global core's index, on a line whose trigger is literal text of the leaf's first body line, compared in lower case with markdown links reduced to their text, and every principle leaf once in `hive-principle-index`. A leaf SHALL be read by the path its index line, its flow link or its flow's stage packet gives, never through a host's skill-invocation tool.

#### Scenario: A leaf missing its switch
- **WHEN** a principle leaf lacks `disable-model-invocation: true`, `hive-core/SKILL.md` lacks it, or a leaf is listed twice in an index
- **THEN** `go test ./tests/content/...` fails and names the leaf.

#### Scenario: An index line that does not quote its leaf
- **WHEN** the trigger of a line of the global core's «Rule leaves» is not literal text of the first body line of the leaf it links
- **THEN** `go test ./tests/content/...` fails and names the line and the leaf.

#### Scenario: A rule restated outside its leaf
- **WHEN** a key phrase registered for a rule leaf appears in a second file under `content/`
- **THEN** the single-home test fails and names both files.

#### Scenario: A rule leaf left as its own skill
- **WHEN** a directory `content/skills/hive-rule-*` exists
- **THEN** `go test ./tests/content/...` fails and names it.

### Requirement: Declared source footprint without token claims

The size check SHALL report the global core separately from explicit simple-query, question, retained plan/build and E2E scenarios. Each scenario SHALL name its read conditions and source files, include the entrypoints and mandatory procedures selected by those conditions, and count each source once. A scenario that reads a flow's generated stage packet SHALL count it as one read whose sources are its members, and SHALL report its number of reads, its bytes without repeats, the members the packets bring that the scenario's earlier list did not read, and its bytes with members repeated across packets; the comparison against the earlier baseline SHALL apply to the bytes without repeats of the members the earlier list read, and SHALL NOT fail for the brought members or the repeats. Missing required sources SHALL fail measurement. The change report SHALL distinguish prose, references and executable code, and SHALL retain a before/after comparison using the same source set so a scenario-definition change is not misreported as savings. Existing scenario measurements SHALL remain identifiable. Source bytes and characters SHALL NOT be presented as actual tokens, observed native delivery or model reliability. Native metadata/rendering, repeated reads, project inputs, history and host policy SHALL remain explicit exclusions. Each skill entrypoint and reference SHALL retain the existing 16,000-byte limit, and each stage packet SHALL be at most 24,000 bytes and 1,000 lines; no new aggregate quota is implied.

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
