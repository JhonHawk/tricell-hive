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
| Engram memory protocol | Engram Codex plugin (`engram@engram`, bundled hooks) | plugin cache under `~/.codex/plugins/` | Engram's own plugin channel — nothing to deploy from the hub |

## One-time setup

1. Merge `config.toml.snippet` into `~/.codex/config.toml` (Linear MCP + subagent
   limits).
2. Run `/deploy-global` from the hub; start a fresh Codex session.

## Engram memory — plugin-injected; DO NOT run `engram setup` for Codex

The Engram protocol reaches Codex through the **Engram Codex plugin** (bundled hooks:
SessionStart injects the protocol + live memory context and recovers after compaction;
UserPromptSubmit forces the MCP ToolSearch; Stop/SubagentStop close sessions). The former
`engram-memory-tail.md` appended to `~/.codex/AGENTS.md` duplicated this and was removed
2026-07-10 — history preserves it if the plugin ever goes away.

`engram setup` (option 5, Codex) is still harmful: it rewrites `~/.codex/engram-instructions.md`
AND re-adds `model_instructions_file = .../engram-instructions.md` to `config.toml`, which
**replaces** Codex's base system prompt (verified in codex-rs). The plugin makes it
unnecessary. If it ever runs: remove the `model_instructions_file` line from `config.toml`.
Check with: `grep -n model_instructions_file ~/.codex/config.toml` (should return nothing).

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
- Claude model aliases are translated by the tier map in
  `harness/build/convert-agents.py`: `opus` → `gpt-5.6-sol` @ `high` (judgment),
  `sonnet` → `gpt-5.6-luna` @ `max` (execution), `haiku` → `gpt-5.6-luna` @ `high`.
  The tier's effort overrides the Claude `effort` frontmatter, which is calibrated
  for Claude's models; the original value stays as a comment.
- `inherit` emits no `model`, so those agents resolve to `[agents]
  default_subagent_model` in `config.toml` before falling back to the session model.

`python3 harness/build.py` prints warnings for lossy fields so review catches
semantic drift before deploy.
