---
name: manage-agents
description: >
  Validate Claude Code agent definitions in global/agents/ against the hub's design
  principles — the judgment checks no hook or build step covers. Use with:
  /manage-agents validate [--all] [--deep]. Triggers after editing agent files,
  after routing changes, or when reviewing agent quality.
---

Validate agent definitions in this workspace. Parse `$ARGUMENTS` to determine the subcommand.

## Subcommands

### `validate` — Validate all agents

Default scope: the agent files changed in the working tree / recent commits, or named by the user; `validate --all` scans the full inventory. Check each in-scope agent against the design principles in CLAUDE.md:

1. **Line count** — Must be ≤120 lines. Flag any that exceed.
2. **Naming** — kebab-case, 3-50 chars, starts/ends with alphanumeric.
3. **Frontmatter** — Must have: name, description, model, color. `tools` is recommended but optional; if omitted, report that the agent inherits the default toolset and verify that this is intentional for the role. Recognized optional fields: `tools`, `effort`, `permissionMode`, `maxTurns`, `skills`, `disallowedTools`, `mcpServers`, `hooks`, `background`, `isolation`, `initialPrompt`, `memory`.
4. **Description quality** — Must be 1-3 action-oriented sentences stating WHEN to invoke (triggers). Flag generic descriptions, and flag descriptions that compress the body's procedure into steps — a workflow-summarizing description risks being followed instead of the body.
5. **Tool restriction** — Read-only agents (review/) should not have Write, Edit, or Bash. Quality agents may be remediation-oriented (Write/Edit expected) or audit-oriented (read-only plus Bash for external analyzers); validate the tool surface against the role described. `permissionMode: plan` is a positive signal for read-only review agents.
6. **Color vs directory** — Agent must be in the correct subdirectory for its color (design/=blue, development/=green, review/=cyan, quality/=yellow, ops/=red, docs/=magenta).
7. **No global rule duplication** — Flag rules that repeat content from `global/rules/`.
8. **Role rules resolution** — every path cited in an agent's `## Role rules` table resolves to a real file: the canonical `~/.claude/...` form exists as deployed state, and the same reference exists in the generated `harness/agents-skills/` tree. `build.py` rebases the table's path root per harness, but only for reference files `SKILL_REFERENCE_INJECTIONS` actually emits — a row pointing at a reference the build never injects is a broken pointer in all four harnesses. Hand-added rows are the typical source of drift.
9. **README inventory sync** (always on `--all`; otherwise when any agent file changed) — `README.md`'s agents table matches disk: heading count == `ls global/agents/*/*.md | wc -l`, one row per agent, tool-surface cell consistent with the agent's `tools`/`disallowedTools` frontmatter.
10. **Cross-harness emission** — the agent is consumed by four harnesses, not one. Run `python3 harness/build.py`: (a) `git status --porcelain harness/` comes back clean — dirty means sources changed without a rebuild; (b) every `WARNING` the build prints for an in-scope agent is a reported finding, never left in scroll-off; (c) per in-scope agent, intent survives translation — a read-only reviewer (`permissionMode: plan`, or no Write/Edit in its surface) emits `sandbox_mode = "read-only"` in its Codex TOML, `permission_mode: plan` in its Grok copy, and `edit: "deny"` in opencode; an `Agent` denial (allowlist omission or `disallowedTools`) emits `task: "deny"` (opencode) and the do-not-spawn clause (Codex/Grok); (d) a `~/.claude/...` path cited in the body must exist as deployed machine state reachable from every harness (the claude scope deploys it), or the citation is Claude-only by accident.
11. **Industry alignment (deep pass — only on `validate --deep` or explicit request)** — check the in-scope agents' Rules sections against current industry consensus (web search, context7); flag outdated or missing practices. The default validate skips this pass.

Output a summary table, then list specific issues per agent with suggestions.

### No arguments — Show help

If `$ARGUMENTS` is empty, show `validate [--all] [--deep]` with usage examples.
