# comment-audit

## Purpose

TBD: created by archiving comment-audit; update the purpose.

## Requirements

### Requirement: Comment audit of a repository

Hive SHALL ship a `audit-comments` skill that inventories the comments of a repository or an area, classifies each one as tautology, stale, excuse, decision record, directive or constraint, and removes them in batches, filing a ticket for each excuse and writing each decision record before its comment is removed.

#### Scenario: An excuse comment
- **GIVEN** a comment that justifies a workaround
- **WHEN** its batch runs
- **THEN** a ticket for the workaround exists before the comment is removed
- **AND** the report links that ticket

### Requirement: Comment ratchet

The skill SHALL leave a lint that compares each changed file's comments with the same file at the Git base and rejects a new, duplicated or reintroduced comment, so a partial cleanup can land; directives and the workaround marker are allowed.

#### Scenario: A change adds a comment
- **GIVEN** the ratchet wired into the project's lint
- **WHEN** a change adds a comment to a file
- **THEN** the lint fails and names the file, line and column

### Requirement: Project skill

The skill SHALL generate the project's own skill to repeat the cleanup, in `.agents/skills/<name>/` with a relative link in `.claude/skills/<name>`, a link in `.grok/skills/<name>` when the project has `.grok/`, a line in the project's `## Skills` section, and the same line with the child's prefix in a declared workspace's `## Skills`.

#### Scenario: Registry check after generation
- **GIVEN** the generated skill in a repository that is a child of a declared workspace
- **WHEN** `skill_registry.py check` runs on the repository and on the workspace
- **THEN** both exit 0

### Requirement: Workaround marker under a comment ban

When a project's lint runs the comment check, an accepted workaround SHALL be marked with exactly one line holding the language's line-comment token followed by `workaround: <ticket>`, or the decision-record entry while no ticket exists, which the lint allows, instead of a prose comment naming its cause.

#### Scenario: An excuse over a workaround in a project with the ratchet
- **GIVEN** an excuse comment over workaround code, in a project whose lint runs the comment check
- **WHEN** its batch runs
- **THEN** the workaround's ticket exists, the excuse is gone, and the line above the workaround is the marker naming that ticket

### Requirement: Narrating comments under a comment ban

When a project's lint or CI runs the comment check, the review of a candidate SHALL move a comment it would keep, and a constraint left without a check, to the decision record the check's error message or the generated project skill names, with the affected path, instead of keeping it in the code; directives, license headers and the workaround marker stay. When no record is named, the review reports the comment as open.

#### Scenario: A why-comment in a project with the ratchet
- **GIVEN** a candidate that adds a comment explaining a non-obvious choice, in a project whose lint runs the comment check and names its decision record
- **WHEN** the review applies the narrating-comments rule
- **THEN** the explanation is in that record with the path it concerns, and the comment is gone

### Requirement: Vendored code stays out of the audit and inside the batch check

The comment audit SHALL leave vendored third-party code out of its inventory and its lanes, listed by its path from the repository root in the ratchet's `EXCLUDED`, and the ratchet's `--same-code` check SHALL fail when a batch changes, adds, deletes or moves out any tracked file of any type under an excluded path, so leaving code out of the audit never leaves it out of the proof that a batch removed only comments.

#### Scenario: A batch damages vendored code
- **GIVEN** `vendor/lib.js` listed in `EXCLUDED`
- **WHEN** a removal batch changes `vendor/lib.js` and `--same-code` runs against the record commit
- **THEN** it exits 1 and names `vendor/lib.js`

#### Scenario: A lane child does not read vendored comments
- **GIVEN** a lane whose package holds a vendored folder listed in `EXCLUDED`
- **WHEN** the skill briefs the lane's child
- **THEN** the child marks only the comments of the lane's inventory file, which leaves out that folder

### Requirement: A comment that restates a rule the code expresses is removed

The narrating-comments review and the comment audit SHALL classify as removable a comment that restates a rule the code, its tests or its configuration in the repository already express, even when it states only part of the rule, when the evidence names the line, test or setting that expresses it; a part of the comment the code does not express, a detail the code contradicts included, SHALL remain, the comment cut down to that part, which is classified on its own.

#### Scenario: A half-stated rule the code enforces
- **GIVEN** a comment `// only admins can delete` over a check that rejects non-admin users and also rejects archived records
- **WHEN** the review or an audit batch classifies it
- **THEN** it is a Kill whose evidence names the check, and the comment is removed

#### Scenario: A rule the code does not show
- **GIVEN** a comment that states a vendor's rate limit that no line, test or setting in the repository expresses
- **WHEN** it is classified
- **THEN** it is not a Kill under this rule and goes through the Keep test

#### Scenario: Half a rule the code enforces plus a vendor fact
- **GIVEN** a comment that restates the admin check above it and adds the vendor's rate limit, which the code does not express
- **WHEN** it is classified
- **THEN** the restated part is removed, the comment is cut down to the rate limit, and that part goes through the Keep test

### Requirement: Comment rule in the project's instructions

The comment audit SHALL write a `## Comments` section in the `AGENTS.md` of each repository whose lint runs the comment check, from the template of its project-skill reference, naming the comment lint command, the paths it checks, the project's decision record and tracker; it SHALL update an existing `## Comments` section instead of adding another, SHALL NOT write it to a declared workspace, and SHALL report a `CLAUDE.md` at that level that does not import `@AGENTS.md`.

#### Scenario: A repository audited for the first time
- **GIVEN** a repository whose `AGENTS.md` has no `## Comments` section
- **WHEN** the audit reaches its project-skill step
- **THEN** the file holds one `## Comments` section that bans new comments, names the lint command and the paths it checks, says where a decision, a shortcut and a constraint go (a constraint without a check to the decision record), lists directives, license headers and the workaround marker as the only comments that stay, and the project skill points to that section for the marker's form
