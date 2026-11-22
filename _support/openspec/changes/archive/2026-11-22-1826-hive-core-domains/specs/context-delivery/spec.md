## MODIFIED Requirements

### Requirement: Rule leaves hidden from model invocation
The skill `hive-core` SHALL hold every rule leaf and every principle leaf in domain folders at its top level, `content/skills/hive-core/<domain>/<name>.md`, and SHALL NOT hold a `references/` folder. The domains SHALL be the ones `content/skills/hive-core/assets/domains.json` lists, each folder SHALL hold at least one leaf, and a leaf's name SHALL be unique across domains. Each rule leaf SHALL have no frontmatter, a first body line after its `#` title that starts with "Read" and states when to read it, and at most 4,000 bytes, or at most 7,000 bytes when the content test names it as an exception; a rule leaf's own references SHALL live under `hive-core/<domain>/<name>/` and its scripts under `hive-core/scripts/<name>/`. The principle leaves SHALL form the domain `principles`: each a file `content/skills/hive-core/principles/<name>.md` with no frontmatter, a first body line after its `#` title that starts with "Read", and at most 4,000 bytes, and the principle index SHALL be `content/skills/hive-core/principles/index.md`. No directory `content/skills/hive-rule-*` or `content/skills/hive-principle-*` SHALL exist. `hive-core/SKILL.md` SHALL set `disable-model-invocation: true` and SHALL NOT list or route to the leaves, since the global core's index and the principle index do. The language leaf `hive-typescript` SHALL stay model-invocable, with `paths` for TypeScript files, and the principle index SHALL link it with the condition of editing those files. Every rule leaf SHALL appear exactly once in the global core's index, on a line whose trigger is literal text of the leaf's first body line, compared in lower case with markdown links reduced to their text, and every principle leaf once in the principle index. A leaf SHALL be read by the path its index line, its flow link or its flow's stage packet gives, never through a host's skill-invocation tool.

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
