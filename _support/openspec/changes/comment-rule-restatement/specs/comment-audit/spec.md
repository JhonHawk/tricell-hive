# comment-audit

## ADDED Requirements

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
