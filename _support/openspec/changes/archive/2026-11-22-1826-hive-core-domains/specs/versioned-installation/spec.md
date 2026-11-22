## MODIFIED Requirements

### Requirement: Generated stage packets per flow
When a released skill holds `assets/stage.json`, a JSON object whose keys match `^stage(-[a-z]+)?$` and whose values are ordered arrays of paths relative to `content/skills/`, the manager SHALL generate in that skill's installed directory one packet `<key>.md` per key, for every host that installs the skill, recorded as a generated managed target. Each packet SHALL hold its members' text in array order, each without frontmatter and under a heading `## <member>`, where a member's name is its manifest path without `.md` (such as `hive-core/messages/questions`), and SHALL end with the line `<!-- hive packet end: <key> · <member>,<member>,… -->` naming every member in order. A link without an anchor to another member of the same packet SHALL point to that member's heading in the packet; a link with an anchor, to another member or within a member, SHALL point to the member's file in the installed store with its anchor; every other relative link, reference-style definitions included, SHALL be rewritten so it resolves from the packet's directory in the installed store; and the rendered packet SHALL be added to the files and anchors its links are validated against, as a release file's are. A release SHALL be rejected when the manifest is malformed, has an unknown key, names a missing member, or renders a packet over 24,000 bytes or 1,000 lines. A second update with unchanged content SHALL write nothing; a packet SHALL be removed when its key, its manifest or its skill leaves the release, rewritten when a member changes, and reinstalled when deleted, as any managed content. A hand-written `stage*.md` in a source skill directory SHALL be rejected, from a stored release and from a source checkout alike. When the manager retires a Hive skill directory, after the retirement is recorded, it SHALL also remove the regular `*.pyc` files of each `__pycache__` directory in it, checked without following links, and those directories once empty; it SHALL refuse a `__pycache__` that is a link, SHALL keep a directory that holds any other file, and SHALL report each kept directory in the apply output.

#### Scenario: A flow with one packet
- **WHEN** a release holds `content/skills/flow-build/assets/stage.json` as `{"stage": [three rule leaves]}`
- **THEN** each host's installed `flow-build` directory gets `stage.md` with the three leaves' text in that order, without frontmatter, and its links resolve in the installed store.

#### Scenario: A flow with two packets
- **WHEN** `flow-verify/assets/stage.json` holds the keys `stage-entry` and `stage-exit`
- **THEN** the installed `flow-verify` directory gets `stage-entry.md` and `stage-exit.md` with their own members.

#### Scenario: A member changes
- **WHEN** an update changes one rule leaf a manifest lists
- **THEN** that leaf and the packets that list it are rewritten and no other managed file changes.

#### Scenario: A packet over its limit or a bad manifest
- **WHEN** a packet would render to more than 24,000 bytes or 1,000 lines, or a manifest has an unknown key or a missing member
- **THEN** the plan fails before any write and names the manifest, the packet and the reason.

#### Scenario: A hand-written packet
- **WHEN** a source checkout holds `content/skills/flow-build/stage.md`
- **THEN** the release is rejected with a message that names the file.

#### Scenario: A retired skill with a Python cache
- **WHEN** an update retires a Hive skill directory whose `scripts/__pycache__/` holds only `.pyc` files
- **THEN** the directory is removed whole, a directory holding any other file is kept and named in the apply output, and a `__pycache__` that is a link is left untouched and reported.

## ADDED Requirements

### Requirement: Declared domain folders of a skill
A released skill MAY hold `assets/domains.json`, a JSON object whose one key `domains` is an array of unique kebab-case folder names, none of them `references`, `scripts`, `assets`, `playbooks` or `agents`, each an existing, non-empty folder at the top of that skill. The manager SHALL install, for every host that installs the skill, the resources below each listed folder as it installs those of `references/`, and SHALL validate the Markdown links in them as it validates those of `references/`. A first-level folder of a skill that is neither one of the four fixed folders nor listed SHALL NOT be installed. The manager SHALL check the manifest against the release's own files, whether the release comes from a source checkout or from a stored release, and SHALL reject the release before any write, naming the manifest and the reason, when the manifest is not that object, has another key, repeats a name, uses a reserved name, lists a folder that does not exist or holds no file, or when the release holds a file in a first-level folder of that skill that is neither fixed nor listed. Recorded state that names a resource under a listed folder SHALL be accepted as any recorded skill resource.

#### Scenario: A skill with two domains
- **WHEN** `content/skills/hive-core/assets/domains.json` is `{"domains": ["git", "review"]}` and the skill holds `git/git-commit.md`, `review/blast-radius.md` and `notes/draft.md`
- **THEN** each host's installed `hive-core` gets `git/git-commit.md` and `review/blast-radius.md`, their links are validated, and `notes/draft.md` is not installed.

#### Scenario: A bad domain manifest
- **WHEN** a domain manifest names a missing or empty folder, repeats a name, uses `references` or `agents`, or has a key other than `domains`, or a stored release holds a file in an unlisted first-level folder
- **THEN** the plan fails before any write and names the manifest and the reason.
