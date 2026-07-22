---
name: manage-agents
description: >
  Manage Claude Code agent definitions — validate, optimize, or generate reports.
  Use with: /manage-agents validate, /manage-agents optimize <agent-name>,
  or /manage-agents report. Triggers when working with agent files in global/agents/,
  reviewing agent quality, creating new agents, or analyzing agent metrics.
---

Manage agent definitions in this workspace. Parse `$ARGUMENTS` to determine the subcommand.

## Subcommands

### `validate` — Validate all agents

Default scope: the agent files changed in the working tree / recent commits, or named by the user; `validate --all` scans the full inventory. Check each in-scope agent against the design principles in CLAUDE.md:

1. **Line count** — Must be ≤120 lines. Flag any that exceed.
2. **Naming** — kebab-case, 3-50 chars, starts/ends with alphanumeric.
3. **Frontmatter** — Must have: name, description, model, color. `tools` is recommended but optional; if omitted, report that the agent inherits the default toolset and verify that this is intentional for the role. Recognized optional fields: `tools`, `effort`, `permissionMode`, `maxTurns`, `skills`, `disallowedTools`, `mcpServers`, `hooks`, `background`, `isolation`, `initialPrompt`, `memory`.
4. **Description quality** — Must be 1-3 action-oriented sentences stating WHEN to invoke (triggers). Flag generic descriptions, and flag descriptions that compress the body's procedure into steps — a workflow-summarizing description risks being followed instead of the body.
5. **Tool restriction** — Read-only agents (review/) should not have Write, Edit, or Bash. Quality agents may be remediation-oriented (Write/Edit expected) or audit-oriented (read-only plus Bash for external analyzers); validate the tool surface against the role described. `permissionMode: plan` is a positive signal for read-only review agents.
6. **Color vs directory** — Agent must be in the correct subdirectory for its color (development/=green, review/=cyan, quality/=yellow, ops/=red, docs/=magenta).
7. **No global rule duplication** — Flag rules that repeat content from `global/rules/`.
8. **Industry alignment (deep pass — only on `validate --deep` or explicit request)** — check the in-scope agents' Rules sections against current industry consensus (web search, context7); flag outdated or missing practices. The default validate skips this pass.

Output a summary table, then list specific issues per agent with suggestions.

### `optimize <agent-name>` — Optimize a single agent

`<agent-name>` is the filename without `.md`, e.g., `backend-developer`.

1. **Read the original** from `_support/workspace/<agent-name>.md`. If not found, list available agents.
2. **Read the template** from `references/template.md` (bundled in this skill).
3. **Analyze** — identify: line count, filler sections (apply "What to eliminate" from CLAUDE.md), rules duplicating global config, frontmatter issues, description quality.
4. **Present findings** — show what will be cut and why, with educational explanation.
5. **Rewrite** using the template structure — aim for 30-120 lines.
6. **Save** to the correct subdirectory in `global/agents/` based on color role.
7. **Report** — original vs optimized line count, % reduction, what was cut, what was preserved.

Rules:
- Check if an optimized version already exists — show diff if so.
- Use context7 MCP for framework-specific docs when the agent covers a technology.
- Never add rules that belong in global rules.
- Ask the user before saving if reduction is >70%.
- **Research-driven**: before writing the agent's Rules section, research current industry best practices for that agent's domain (web search, context7, authoritative sources). Don't limit the agent's rules to what the global config already says — the agent should encode domain-specific expertise that goes beyond the global rules.

### `report` — Generate analysis report

Scan `global/agents/` and `_support/workspace/` to generate a live optimization report:

1. Count all optimized agents with line counts, grouped by subdirectory/color.
2. For each agent that has an original in `_support/workspace/`, calculate % reduction.
3. Identify the largest agent, smallest agent, and average lines.
4. List agents in `_support/workspace/` that have NO optimized version (candidates for optimization).
5. Output a markdown report with summary table and coverage analysis.

### No arguments — Show help

If `$ARGUMENTS` is empty, show the available subcommands with usage examples.
