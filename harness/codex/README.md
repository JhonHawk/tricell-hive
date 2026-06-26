# Codex CLI adapter — flow pack

Thin layer for Codex. Skills need NO wrapper here: Codex reads `~/.agents/skills/`
natively and invokes with `$flow-plan <free text>` (or `/skills`). Each skill ships
`agents/openai.yaml` with `allow_implicit_invocation: false`, so the user-gate
semantics travel with the skill.

## Pieces

| Piece | Source | Deploy target | Maintained how |
|---|---|---|---|
| Skills + rubrics | `global/skills/` | `~/.agents/skills/` | `/deploy-global` (copy) |
| Subagents | `harness/codex/agents/` (generated, versioned) | `~/.codex/agents/` | **Generated** by `harness/build.py` from `global/agents/` — never edit |
| Config additions | `config.toml.snippet` | merge into `~/.codex/config.toml` | Manual, once |
| Engram memory protocol | `engram-memory-tail.md` | appended to `~/.codex/AGENTS.md` | `/deploy-global` (cat after the shared AGENTS.md) |

## One-time setup

1. Merge `config.toml.snippet` into `~/.codex/config.toml` (Linear MCP + subagent
   limits).
2. Run `/deploy-global` from the hub; start a fresh Codex session.

## Engram memory — DO NOT run `engram setup` for Codex

Engram reaches Codex as the `engram-memory-tail.md` section of `~/.codex/AGENTS.md`
(deployed by `/deploy-global`). This is additive (preserves Codex's base prompt) and
invisible in the TUI — unlike a SessionStart hook, which renders its injected block
on screen every session.

`engram setup` (option 5, Codex) **breaks this**: it rewrites `~/.codex/engram-instructions.md`
to a verbose 90-line protocol AND re-adds `model_instructions_file = .../engram-instructions.md`
to `config.toml`, which **replaces** Codex's base system prompt (verified in codex-rs). The
Engram MCP server (the tools) is registered separately and does NOT need `engram setup`.

If `engram setup` Codex ever runs again: remove the `model_instructions_file` line from
`config.toml`, then re-run `/deploy-global` to restore the AGENTS.md tail. Check with:
`grep -n model_instructions_file ~/.codex/config.toml` (should return nothing).

## Behavioral notes (from harness-mechanics.md)

- **No argument substitution in Codex skills** — `$flow-specs review epics/E07` passes
  "review epics/E07" as free text; the skill bodies interpret it as the subcommand.
- **Subagent spawning is explicit-only** in Codex; the skills' "dispatch X" instructions
  count as the explicit request. `max_depth = 1` suffices (orchestrator → workers).
- **Read-only reviewers** are enforced via `sandbox_mode = "read-only"` in the generated
  TOMLs (no per-tool allowlists in Codex).

## Claude → Codex agent parity

`global/agents/**/*.md` is Claude Code-first. The generator preserves the shared
parts in Codex TOML and makes non-equivalent fields visible:

- Native Codex fields: `name`, `description`, `developer_instructions`,
  `sandbox_mode`, and compatible `model_reasoning_effort`.
- Claude-only or lossy fields are emitted as `# Harness compatibility` comments.
- `disallowedTools: Agent` becomes a developer instruction not to spawn or
  delegate agents; Codex has no native per-agent tool deny equivalent here.
- `skills:` becomes a developer instruction to use the named skill when
  available; do not assume it is preloaded.
- Claude model aliases such as `sonnet` stay as comments; only Codex-compatible
  model IDs are emitted as `model`.

`python3 harness/build.py` prints warnings for lossy fields so review catches
semantic drift before deploy.
