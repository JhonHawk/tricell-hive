## ADDED Requirements

### Requirement: Antigravity global core in size-bounded always-on rules
For the Antigravity host, Hive SHALL deliver the global core as one or more Hive-owned rule files `<AntigravityHome>/rules/hive-NN.md` under the resolved `~/.gemini/config`, each beginning with the frontmatter `trigger: always_on` and each at or below 24,000 bytes, because Antigravity truncates every rule file above that size, the global ones included. The parts SHALL be cut only at `##` headings outside fenced code blocks, in order, so that their managed texts joined in order reproduce the rewritten global core; a section that alone exceeds the limit SHALL fail the plan with an error naming it. The voice block SHALL appear only in the first part, and setting or changing the voice SHALL change only that part. Hive SHALL NOT write `~/.gemini/GEMINI.md` for this host.

#### Scenario: Guidance larger than one rule file
- **WHEN** a user-scope plan for Antigravity renders a global core larger than one rule file allows
- **THEN** it proposes as many `hive-NN.md` parts as needed, each with the `always_on` frontmatter and at or below 24,000 bytes, and their managed texts joined in order equal the rewritten global core.

#### Scenario: Part count changes on update
- **WHEN** an update renders the global core into fewer or more parts than the recorded installation
- **THEN** it writes the new parts and removes only the surplus Hive-owned parts, leaving every other file in `rules/` unchanged.

### Requirement: Antigravity roles inherit the session model
For the Antigravity host, Hive SHALL render each canonical role as `<AntigravityHome>/agents/<role>.md` with `name`, `description` and `model: "inherit"` unless the user set another model with `hive models set`, and SHALL NOT render a `tools` list or a reasoning effort. Read-only roles SHALL rely on their role instructions, since Antigravity offers no read-only field.

#### Scenario: Observe role rendered
- **WHEN** a role with `observe` access is rendered for Antigravity
- **THEN** its file carries `model: "inherit"` and no `tools` or effort field, and its body keeps the role's read-only instructions.
