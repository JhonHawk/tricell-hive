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
2. Run `/deploy-global` from the hub; start a fresh Codex session. The PI rollout
   selectively activates the five Astra role files listed below; it does not copy
   the rest of the Codex roster into the machine-level override.

## What this harness loads

Every URL below was fetched and its quote extracted from the page body on the date in the
last column. Companion files: `global/README.md` (Claude Code),
`harness/opencode/README.md`, `harness/grok/README.md`, and `harness/pi/README.md` —
all share this layout where the harness has an equivalent loading surface.

> **The documentation host moved.** Every `developers.openai.com/codex/*` URL now redirects
> to `learn.chatgpt.com/docs/*`. Write the destination host; the old one only resolves by
> redirect.

| Layer | What | Where it lands | Mechanism | Official doc | Verified |
|---|---|---|---|---|---|
| Always-on core | `harness/AGENTS.md` (condensed cross-harness core) | `~/.codex/AGENTS.md` | Global scope, read every session; separate from the project chain | [agent-configuration/agents-md](https://learn.chatgpt.com/docs/agent-configuration/agents-md) | 2026-08-20 |
| Project chain | every `AGENTS.md` from git root down to cwd | *(read in place)* | Concatenated until `project_doc_max_bytes` (**32 KiB default**) is reached, then **silently truncated** | [agents-md](https://learn.chatgpt.com/docs/agent-configuration/agents-md) — *"stops adding files once the combined size reaches the limit defined by project_doc_max_bytes (32 KiB by default)"* · [config-reference](https://learn.chatgpt.com/docs/config-file/config-reference) | 2026-08-20 |
| Delivered rules | 25 triggered texts (`globs:` and/or `commands:`) | `~/.agents/skills/<router>/references/` | `rule-delivery` holds the first `apply_patch` matching a rule's globs — or the first shell stage matching a declared command prefix — and names the reference; the hold does not re-arm after a compaction (no Codex compaction event) | (repo mechanism — `global/hooks/rule-delivery/README.md`) | 2026-09-19 |
| Skills | 17 skills with injected `references/` | `~/.agents/skills/` | User-scope skill folder, read natively | [build-skills](https://learn.chatgpt.com/docs/build-skills) — *"USER $HOME/.agents/skills — Any skills checked into the user's personal folder."* | 2026-08-20 |
| Skill gating | `agents/openai.yaml` per gated skill | inside each skill dir | Codex ignores Claude's `disable-model-invocation`; the YAML policy carries the gate | [build-skills](https://learn.chatgpt.com/docs/build-skills) — *"allow_implicit_invocation (default: true): When false, Codex won't implicitly invoke the skill based on user prompt"* | 2026-08-20 |
| Agents | 26 generated `.toml` | `~/.codex/agents/` | Standalone TOML files per agent | [agent-configuration/subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents) — *"add standalone TOML files under ~/.codex/agents/ for personal agents"* | 2026-08-20 |
| Hooks | 3 hooks shipping a `codex-hooks.json` | `~/.codex/hooks/` + merged into `~/.codex/hooks.json` | Native hooks config alongside `config.toml` | [hooks](https://learn.chatgpt.com/docs/hooks) — *"the four most useful locations are: ~/.codex/hooks.json ~/.codex/config.toml …"* | 2026-08-20 |
| Prompt auditor | — | — | `codex debug prompt-input` renders the model-visible prompt as JSON | [developer-commands](https://learn.chatgpt.com/docs/developer-commands?surface=cli) — *"Render the model-visible prompt input list as JSON"* | 2026-08-20 |

**`project_doc_max_bytes` default is 32 KiB, not 65536.** The `65536` in
`config.toml.snippet` is a value we *raise it to*; the `65536` seen in the official page is
likewise an example of raising it, never the default. The cap covers the **project chain
only** — the global `~/.codex/AGENTS.md` is not charged against it.

## What does NOT reach it

- **The 34 rule texts under `global/rules-situational/` that the core does not inline.**
  Codex has no `paths:`-equivalent and no conditional-rule channel of any kind. They reach
  it two ways instead: injected into router-skill `references/` by `harness/build.py`
  (`SKILL_REFERENCE_INJECTIONS`), and held by the `rule-delivery` hook on the first
  `apply_patch` matching a rule's `globs:` or the first shell stage matching its
  `commands:` (read the named reference under `~/.agents/skills/…`, then re-issue). The
  other 7 are not missing: they are inlined into the condensed core Codex already loads.
- **Claude Code's per-agent tool allowlist.** Read-only reviewers are enforced by
  `sandbox_mode = "read-only"` instead; there is no per-tool deny.
- **`$ARGUMENTS` substitution in skills.** Arguments arrive as free text.

## Unverified / undocumented dependencies

Two config surfaces this adapter relies on that the official docs do **not** mention. Both
were established empirically; neither is citable:

- **`[hooks.state]` enable slots.** A merged hook stays inert until its slot exists with
  `enabled = true`. The documented switch is `[features] hooks = false` (with `codex_hooks`
  as a deprecated alias); `[hooks.state]` appears nowhere in the docs.
  **Status: observed locally, undocumented.**
- **`multi_agent = true`.** Cited by the agent-resolution note below (verified 2026-08-15,
  codex 0.147.0). The documented subagent surface is the `[agents]` table
  (`max_concurrent_threads_per_session` et al.). **Status: observed locally, undocumented
  — may be an internal key or renamed since.**

## How to re-verify

```bash
codex debug prompt-input        # exact model-visible prompt: global AGENTS.md first, then
                                # `--- project-doc ---` with per-file headers
codex debug --help              # confirms the subcommand still exists in your build
grep -n project_doc_max_bytes ~/.codex/config.toml
ls ~/.codex/agents/ | wc -l     # expect 24
```

`python3 harness/build.py` reports **two** chain sizes on every run — at the repo root and
under `harness/` — because a session opened under `harness/` loads the shared core twice.

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
- **Subagent spawning triggers on explicit request OR an imperative AGENTS.md/skill
  instruction** (fully proactive delegation is gated to the Ultra intelligence tier). The
  always-on routing table in `harness/AGENTS.md` is that imperative instruction; the skills'
  "dispatch X" lines also count. `max_depth = 1` suffices (orchestrator → workers).
- **`codex debug prompt-input` renders the exact model-visible prompt** (global AGENTS.md
  first, then `--- project-doc ---` with per-file headers) without burning a model turn —
  the deterministic auditor for what Codex actually loads. Grok's analog is `grok inspect`.
- **Custom-agent name resolution verified working** (2026-08-15, codex 0.147.0,
  `multi_agent = true` — undocumented, see above): `spawn_agent(agent_type="review-refuter")` resolves and spawns —
  upstream issues #15250/#14579 report it broken in some tool-backed contexts; if it
  regresses, the fallback is reading the target `~/.codex/agents/<name>.toml` and inlining
  its `developer_instructions` into `spawn_agent(agent_type="worker")`.
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
  `harness/build/convert-agents.py`: `opus` → `gpt-6-astra` @ `medium` (the five approved judgment roles),
  `sonnet` → `gpt-5.6-luna` @ `high` (execution), `haiku` → `gpt-5.6-luna` @ `high`.
  A Claude `effort` declared in the frontmatter wins over the tier's effort, except
  `opus`/`high`, which the Astra policy pins to `medium`; the mapping is recorded as
  a comment in the TOML.
- `inherit` emits no `model`, so those agents resolve to `[agents]
  default_subagent_model` in `config.toml` before falling back to the session model.

### Astra role activation

The canonical Claude frontmatter remains `model: opus` and `effort: high`. The
Codex generator deliberately overrides that pair for these five roles:

| Role | Generated Codex model | Generated effort |
|---|---|---|
| `review-code` | `gpt-6-astra` | `medium` |
| `sdd-product-critic` | `gpt-6-astra` | `medium` |
| `database-specialist` | `gpt-6-astra` | `medium` |
| `performance-engineer` | `gpt-6-astra` | `medium` |
| `prompt-engineer` | `gpt-6-astra` | `medium` |

The source effort is retained in the generated compatibility comment for auditability.
Other roles and the Claude/Grok generated trees keep their existing policy. Rebuild
before checking the generated TOMLs:

```bash
python3 harness/build.py
for agent in review-code sdd-product-critic database-specialist performance-engineer prompt-engineer; do
  rg -n '^(model|model_reasoning_effort) = ' "harness/codex/agents/$agent.toml"
done
```

`python3 harness/build.py` prints warnings for lossy fields so review catches
semantic drift before deploy.

## Hive planning contract

`flow-plan` and `flow-build` use the shared plan artifact and read-only validator. Approval,
action/target scope and execution evidence travel with the plan; entering or leaving native
plan mode grants no Hive authority. During drafting, the parent's no-implementation boundary
is a prompt convention, not a universal write sandbox. `Session: no` retains conversation-only
operation without durable validation or cross-harness resume guarantees.

Source: `global/skills/flow-core/references/plan-format.md` (Hive convention, 2026-09-10).
