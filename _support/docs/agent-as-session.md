# Agent as Session — Reference

## What it is

The `agent` setting in `settings.json` runs the entire Claude Code session as a named subagent. The agent's system prompt, tool restrictions, and model **replace** the default Claude Code system prompt. CLAUDE.md files and project memory still load normally.

**Source:** [Invoke subagents explicitly](https://code.claude.com/docs/en/sub-agents#invoke-subagents-explicitly)

## Configuration

### Via settings.json (persistent)

```json
{
  "agent": "review-code"
}
```

Works at any scope:

| Scope | File | Effect |
|-------|------|--------|
| Global | `~/.claude/settings.json` | All sessions, all projects |
| Project | `.claude/settings.json` | All sessions in this repo (shared via git) |
| Local | `.claude/settings.local.json` | Personal, this repo only |

### Via CLI flag (one-off)

```bash
claude --agent review-code
```

CLI flag overrides the settings.json value if both are present.

### Via CLI JSON (ephemeral, not saved to disk)

```bash
claude --agents '{
  "review-code": {
    "description": "Expert code reviewer",
    "prompt": "You are a senior code reviewer...",
    "tools": ["Read", "Grep", "Glob", "Bash"],
    "model": "sonnet"
  }
}'
```

Accepts the same fields as file-based agent frontmatter. Useful for CI/CD or quick testing.

## What the agent inherits vs. replaces

| Aspect | Behavior |
|--------|----------|
| System prompt | **Replaced** by agent's markdown body |
| CLAUDE.md | Loaded normally |
| Rules (`~/.claude/rules/`) | Loaded normally |
| Tools | Defined by agent's `tools` field (or inherits all if omitted) |
| Model | Defined by agent's `model` field (or inherits if omitted) |
| Permissions | Inherited from session, overridable via `permissionMode` |
| MCP servers | Agent can scope its own via `mcpServers` field |
| Hooks | Agent can define its own via `hooks` field |
| Skills | Must be listed explicitly in `skills` field (not inherited) |
| Memory | Optional via `memory` field (`user`, `project`, `local`) |

## Agent frontmatter fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Unique identifier, lowercase + hyphens |
| `description` | Yes | When to delegate to this agent |
| `tools` | No | Allowlist of tools. Inherits all if omitted |
| `disallowedTools` | No | Denylist, removed from inherited set |
| `model` | No | `sonnet`, `opus`, `haiku`, full model ID, or `inherit` |
| `permissionMode` | No | `default`, `acceptEdits`, `dontAsk`, `bypassPermissions`, `plan` |
| `maxTurns` | No | Max agentic turns before stopping |
| `skills` | No | Skills injected at startup (not inherited from parent) |
| `mcpServers` | No | Scoped MCP servers (inline definition or name reference) |
| `hooks` | No | Lifecycle hooks scoped to this agent |
| `memory` | No | Persistent memory: `user`, `project`, `local` |
| `background` | No | `true` to always run as background task |
| `effort` | No | `low`, `medium`, `high`, `max` (Opus 4.6 only) |
| `isolation` | No | `worktree` for isolated git worktree |
| `initialPrompt` | No | Auto-submitted as first user turn when running as session agent |

**Source:** [Configure subagents](https://code.claude.com/docs/en/sub-agents#configure-subagents)

## Restricting subagent spawning

When running as session agent, `tools` controls which subagents can be spawned:

```yaml
# Only worker and researcher can be spawned
tools: Agent(worker, researcher), Read, Bash

# Any subagent allowed
tools: Agent, Read, Bash

# No subagents (Agent omitted entirely)
tools: Read, Bash
```

**Source:** [Restrict which subagents can be spawned](https://code.claude.com/docs/en/sub-agents#restrict-which-subagents-can-be-spawned)

## Potential use cases (not implemented)

### Per-project agent defaults

| Project type | Agent | Effect |
|-------------|-------|--------|
| NestJS backend | `backend-developer` | Session with NestJS rules, thin controllers, DTOs |
| Angular frontend | `angular-developer` | Session with Angular patterns, signals, standalone |
| Docs-heavy repo | `sdd-spec-writer` | Session focused on documentation |
| Config repo (tricell-hive) | None (default) | Normal session, automatic agent routing |

### Session agent with initialPrompt

An agent that auto-starts with a specific action:

```yaml
---
name: pr-reviewer
description: Reviews the current branch against main
initialPrompt: "Review all commits on this branch vs main. Focus on security, correctness, and maintainability."
tools: Read, Glob, Grep, Bash
model: inherit
---
```

Opening a session with `claude --agent pr-reviewer` would automatically start the review.

## Risks and tradeoffs

| Risk | Detail |
|------|--------|
| Losing default system prompt | The default prompt includes tool usage patterns, git workflows, parallelism guidance. An agent replaces all of this. The agent must be self-sufficient or rely on CLAUDE.md/rules. |
| Reduced versatility | Automatic routing to other agents is lost. A `backend-developer` session won't auto-delegate to `review-security`. |
| Global scope is risky | Setting `agent` in `~/.claude/settings.json` affects every project. Only for an "orchestrator" agent. |
| Project scope is the sweet spot | Each repo gets the agent matching its domain. |

## Related features

| Feature | How it differs | Docs |
|---------|---------------|------|
| Hook type `agent` | Spawns a subagent as reaction to an event, doesn't replace the session | [Hooks](https://code.claude.com/docs/en/hooks) |
| `initialPrompt` | Auto-submitted first turn when running as session agent | [Subagents](https://code.claude.com/docs/en/sub-agents#supported-frontmatter-fields) |
| Agent teams | Parallel agents with separate sessions, communicating with each other | [Agent teams](https://code.claude.com/docs/en/agent-teams) |
| `--agents` CLI flag | Ephemeral JSON agents, not saved to disk | [Subagents](https://code.claude.com/docs/en/sub-agents#choose-the-subagent-scope) |
| `@`-mention | One-off delegation to a specific subagent | [Subagents](https://code.claude.com/docs/en/sub-agents#invoke-subagents-explicitly) |
