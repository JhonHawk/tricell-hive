## ADDED Requirements

### Requirement: Deployed skill resource links resolve to installed paths
The installer SHALL rewrite every inline `skill:owner/path` Markdown link outside code fences, in deployed role files and in the deployed global guidance block, to the resource's path in the skills directory for the install scope: the absolute path under `<home>/.agents/skills` for user scope, and the path relative to the project root of the host's project skills directory for project scope. Source content SHALL keep the `skill:` locator. A file shared by several hosts SHALL receive identical bytes for each of them, or the plan SHALL fail before any write.

#### Scenario: Role link at user scope
- **WHEN** a role whose source links `[browser automation](skill:flow-build/references/browser-automation.md)` is installed or updated at user scope
- **THEN** the deployed role file, including a Codex TOML role, links `<home>/.agents/skills/flow-build/references/browser-automation.md` and contains no `skill:` link

#### Scenario: Project scope stays portable
- **WHEN** the same role is installed at project scope for Claude
- **THEN** the deployed role links `.claude/skills/flow-build/references/browser-automation.md` and contains no absolute home path

#### Scenario: Unchanged second update
- **WHEN** `hive update` runs twice from the same committed revision
- **THEN** the second plan proposes no writes for the rewritten roles or block
