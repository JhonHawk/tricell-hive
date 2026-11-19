## ADDED Requirements

### Requirement: Hidden skills per host
When a released skill's frontmatter sets `disable-model-invocation: true`, the manager SHALL generate `agents/openai.yaml` with `policy.allow_implicit_invocation: false` in that skill's installed directory, recorded as a generated managed target. A second update with unchanged content SHALL write nothing; the file SHALL be removed when the skill loses that frontmatter or leaves the release, and reinstalled when deleted, as any managed content. A hand-written `agents/openai.yaml` SHALL be rejected, from a stored release and from a source checkout alike. Skills without that frontmatter SHALL install unchanged, and `SKILL.md` SHALL be installed byte for byte. Every host's catalog link or path, Antigravity's included, SHALL be created as for any skill.

#### Scenario: A hidden leaf installed at user scope
- **WHEN** a release containing a skill with `disable-model-invocation: true` is applied at user scope
- **THEN** `~/.agents/skills/<skill>/agents/openai.yaml` holds the generated policy and every host's link or path for that skill exists.

#### Scenario: A hand-written host file
- **WHEN** a stored release or a source checkout includes `agents/openai.yaml` written by hand in a skill folder
- **THEN** validation rejects it before any managed write.

#### Scenario: A skill loses the flag
- **WHEN** an update ships a skill without `disable-model-invocation: true` that previously had it
- **THEN** the generated `agents/openai.yaml` is removed and nothing else in the skill changes.
