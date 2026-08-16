@AGENTS.md

## Claude Code

### Flow Pack Manual

When editing any `flow-*` skill (or `flow-core`), evaluate whether `_support/docs/flow-pack-manual.html` needs updating to match — it documents the pack's user-facing model, so a user-visible change (new/renamed command, changed phase model, new gate, new artifact convention) drifts it. Update it in the same change-group when the change is user-visible; skip for internal-only edits. The update is a coherent revision of the affected prose, never a blind string-swap.

### Agent Management

Use `/manage-agents` for all agent operations:
- `/manage-agents validate` — check all agents against design principles; also run after routing changes
- `/manage-agents optimize <name>` — guided optimization of a single agent
- `/manage-agents report` — generate live analysis report with metrics

After editing files under `global/agents/**`, offer `/manage-agents validate` at close; after editing `global/rules/**`, offer `/manage-rules validate`. Offer, don't run uninvited.

### Agent Frontmatter Reference

```yaml
---
# --- Required ---
name: agent-name                    # kebab-case, 3-50 chars, unique
description: >                      # 1-3 sentences. CRITICAL for auto-invocation routing.
  When to use this agent. Be specific. Include trigger phrases. Add <example> blocks if routing is ambiguous.

# --- Tool control ---
tools: Read, Glob, Grep             # Allowlist. Inherits all if omitted. Supports Agent(type1, type2)
                                    # and MCP patterns: mcp__<server>, mcp__* — the Claude-side
                                    # analog of Grok's mcpInheritance.
disallowedTools: Write, Edit        # Denylist. Applied before tools allowlist. Same MCP patterns.

# --- Model & execution ---
model: inherit                      # sonnet | opus | haiku | fable | full model ID | inherit (default)
effort: high                        # low | medium | high | xhigh | max. Overrides session default.
                                    # Available levels depend on the model. Setting the level the
                                    # session already runs at is a no-op — check settings first.
maxTurns: 30                        # Max agentic turns before stopping.
permissionMode: default             # default | acceptEdits | auto | dontAsk | bypassPermissions |
                                    # plan (manual = alias of default). NOT enforced when the parent
                                    # session is in auto mode — see AGENTS.md "Restricted tools".

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
                                    # Model-visible skill listing caps description at
                                    # 1,536 chars/skill (listing budget ~1% of context,
                                    # least-invoked truncated first) — highest-signal
                                    # triggers up front.
# --- Execution control ---
effort: max                         # low | medium | high | xhigh | max. Overrides session default.
context: fork                       # Runs in isolated subagent context. Main sees only result.
                                    # The fork has NO conversation history — only for skills whose
                                    # input is fully specified by $ARGUMENTS + disk.
agent: Explore                      # Subagent type when context: fork. Default: general-purpose.
allowed-tools: Read, Bash(python3 ${CLAUDE_SKILL_DIR}/scripts/*)
                                    # GRANTS unprompted permission for the listed tools during
                                    # the invoking turn. Does NOT restrict: every tool stays
                                    # callable. Scope Bash to command patterns, never bare `Bash`
                                    # — the grant covers the whole turn, not just this skill's
                                    # own commands. ${CLAUDE_SKILL_DIR} substitutes here too.
disallowed-tools: Write, Edit       # Removes tools from the pool while the skill is active.
                                    # This is the field that restricts.
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
- **No AGENTS.md fallback:** Claude Code does NOT load a repo's `AGENTS.md` when `CLAUDE.md` is absent — an `AGENTS.md` without its companion `CLAUDE.md` (`@AGENTS.md`) is invisible to Claude Code. Audit periodically with `find ~/Development/projects -maxdepth 4 -name AGENTS.md | while read f; do [ -f "$(dirname "$f")/CLAUDE.md" ] || echo "INVISIBLE TO CLAUDE CODE: $f"; done`.
