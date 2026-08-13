# opencode adapter — flow pack

Thin layer that lets opencode run the flow pack. The content lives once (skills +
rubrics deploy to `~/.agents/skills/`); this folder holds only what opencode needs
natively.

## Pieces

| Piece | Source | Deploy target | Maintained how |
|---|---|---|---|
| Skills + rubrics | `global/skills/` | `~/.agents/skills/` | `/deploy-global` (copy) |
| Subagents | `harness/opencode/agents/` (generated, versioned) | `~/.config/opencode/agents/` | **Generated** by `harness/build.py` from `global/agents/` — never edit |
| Command wrappers (every user-invoked skill, gated or not; model-invoked routers get none) | `harness/opencode/commands/` | `~/.config/opencode/commands/` | `/deploy-global` (copy) |
| Config additions | `opencode.jsonc.snippet` | merge into `~/.config/opencode/opencode.json` | Manual, once |

## One-time setup

1. Merge `opencode.jsonc.snippet` into `~/.config/opencode/opencode.json`
   (skill gating + Linear MCP).
2. Add to your shell profile: `export OPENCODE_DISABLE_CLAUDE_CODE_SKILLS=1`
   (single skill-discovery source — see snippet comments).
3. Run `/deploy-global` from the hub; start a fresh opencode session.

## How it works

`/flow-plan <args>` (opencode command) → instructs the agent to read
`harness-mechanics.md` (mechanic translation) + the skill body from
`~/.agents/skills/` → dispatches the generated subagents via the task tool.
Updating a skill or agent in the hub + `/deploy-global` updates this harness with no
adapter edits; only a renamed agent or a new skill requires touching this folder
(new command wrapper).
