# comment-audit

## ADDED Requirements

### Requirement: Comment audit of a repository

Hive SHALL ship a `comment-audit` skill that inventories the comments of a repository or an area, classifies each one as tautology, stale, excuse, decision record, directive or constraint, and removes them in batches, filing a ticket for each excuse and writing each decision record before its comment is removed.

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
