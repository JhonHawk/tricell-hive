@AGENTS.md

## Claude Code

### Flow Pack Manual

When editing any `flow-*` skill (or `flow-core`), evaluate whether `_support/docs/flow-pack-manual.html` needs updating to match — it documents the pack's user-facing model, so a user-visible change (new/renamed command, changed phase model, new gate, new artifact convention) drifts it. Update it in the same change-group when the change is user-visible; skip for internal-only edits. The update is a coherent revision of the affected prose, never a blind string-swap.

### Agent Management

Use `/manage-agents` for all agent operations:
- `/manage-agents validate` — check all agents against design principles; also run after routing changes
- `/manage-agents optimize <name>` — guided optimization of a single agent
- `/manage-agents report` — generate live analysis report with metrics

### Agent Frontmatter Reference

```yaml
---
# --- Required ---
name: agent-name                    # kebab-case, 3-50 chars, unique
description: >                      # 1-3 sentences. CRITICAL for auto-invocation routing.
  When to use this agent. Be specific. Include trigger phrases. Add <example> blocks if routing is ambiguous.

# --- Tool control ---
tools: Read, Glob, Grep             # Allowlist. Inherits all if omitted. Supports Agent(type1, type2).
disallowedTools: Write, Edit        # Denylist. Applied before tools allowlist.

# --- Model & execution ---
model: inherit                      # sonnet | opus | haiku | full model ID | inherit (default)
effort: high                        # low | medium | high | max. Overrides session default.
maxTurns: 30                        # Max agentic turns before stopping.
permissionMode: default             # default | acceptEdits | dontAsk | bypassPermissions | plan

# --- Context injection ---
skills: skill-name                  # Skills preloaded at startup. NOT inherited from parent.
mcpServers: server-name             # MCP servers scoped to this agent.
memory: user                        # Persistent memory: user | project | local.

# --- Isolation & lifecycle ---
background: false                   # Always run as background task.
isolation: worktree                 # Git worktree isolation.
hooks:                              # Lifecycle hooks: PreToolUse, PostToolUse, Stop.
  PreToolUse:
    - matcher: "Bash"
      hooks: [{ type: command, command: "./script.sh" }]

# --- UI ---
color: green                        # blue=design, cyan=review, green=build, yellow=validate, magenta=creative, red=critical

# --- Main-agent mode (--agent flag) ---
initialPrompt: "Analyze this repo"  # Auto-submitted first prompt.
---
```

### Skill Frontmatter Reference

```yaml
---
name: skill-name                    # kebab-case, max 64 chars, unique
description: >                      # 1-3 sentences. Claude uses this for auto-invocation.
                                    # Truncated to 250 chars in listings.
# --- Execution control ---
effort: max                         # low | medium | high | max. Overrides session default.
context: fork                       # Runs in isolated subagent context. Main sees only result.
agent: Explore                      # Subagent type when context: fork. Default: general-purpose.
allowed-tools: Read, Glob, Grep     # Restricts tools available to the skill.
model: inherit                      # Override model. Default: session model.
shell: bash                         # Shell for !`command` blocks. Default: bash.

# --- Invocation control ---
disable-model-invocation: true      # Only user can invoke (deploy, commit). Not in context.
user-invocable: false               # Only Claude can invoke (background knowledge). In context.
paths: "src/**/*.tsx"               # Limit auto-activation to matching files.
argument-hint: "[issue-number]"     # Hint shown in autocomplete.

# --- Metadata ---
compatibility: ">=1.0.0"           # Claude Code version compatibility.
license: MIT                        # License for the skill.
metadata:                           # Arbitrary key-value metadata.
  author: team-name
---

# String substitutions available in skill body:
# $ARGUMENTS — all args passed to the skill
# $ARGUMENTS[N] or $N — specific argument by index (0-based)
# ${CLAUDE_SESSION_ID} — current session ID
# ${CLAUDE_SKILL_DIR} — directory containing SKILL.md
# !`command` — shell command output injected BEFORE Claude sees the prompt
```

### Template

```markdown
---
name: [agent-name]
description: [1-3 sentences: WHEN to invoke. Add <example> blocks if routing is ambiguous.]
tools: [only needed tools]
model: inherit
color: [by role: blue/cyan=review, green=build, yellow=validate, magenta=creative, red=critical]
---

You are [1 sentence: who this agent is and its core expertise].

## Focus
- [3-7 bullets: specific areas of expertise — not generic concepts]

## Rules
- [Only rules UNIQUE to this role]
- [Don't repeat global CLAUDE.md rules]
- [Each rule must change the agent's output vs default behavior]

## Output
- [What this agent delivers — format, structure, artifacts]
```

### Notes
- **Claude Code harness injection:** Claude Code injects much of Anthropic's model-specific prompting guidance (act-don't-overplan, lead-with-outcome, autonomous-operation, faithful progress reporting) into the system prompt directly — never duplicate harness-injected guidance into `global/`.
- **AGENTS.md fallback:** In repos *without* `CLAUDE.md`, Claude Code loads `AGENTS.md` as memory fallback — audit periodically with `find ~/Development/projects -maxdepth 4 -name AGENTS.md | while read f; do [ -f "$(dirname "$f")/CLAUDE.md" ] || echo "FALLBACK ACTIVE: $f"; done`.
