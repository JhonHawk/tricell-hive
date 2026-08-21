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
| Glob-scoped rules | `harness/opencode/rules/` (generated, versioned) | `~/.config/opencode/rules/` | **Generated** by `harness/build.py` from `global/rules/{languages,workflow}/` — never edit |
| SessionStart plugin | `global/hooks/flow-session-context/flow-session-context.ts` | `~/.config/opencode/plugins/` | `/deploy-global` (copy) |
| Config additions | `opencode.jsonc.snippet` | merge into `~/.config/opencode/opencode.json` | Manual, once |
| Permission keys | `permission-config.json` | `permission.*` in `~/.config/opencode/opencode.json` | `/deploy-global` (surgical jq merge, idempotent) |

## One-time setup

1. Merge `opencode.jsonc.snippet` into `~/.config/opencode/opencode.json`
   (skill gating + Linear MCP). The `permission` keys are NOT part of this step —
   `permission-config.json` is merged by the deploy on every run.
2. Run `/deploy-global` from the hub; start a fresh opencode session.

## What this harness loads

Every URL below was fetched and its quote extracted from the page body on the date in the
last column. Companion files: `global/README.md` (Claude Code),
`harness/codex/README.md`, `harness/grok/README.md` — all four share this layout.

> **The upstream repo moved orgs.** `github.com/sst/opencode` now redirects to
> `github.com/anomalyco/opencode`. Write `anomalyco`.

| Layer | What | Where it lands | Mechanism | Official doc | Verified |
|---|---|---|---|---|---|
| Always-on core | `harness/AGENTS.md` (condensed cross-harness core) | `~/.config/opencode/AGENTS.md` | Global rules file, applied to every session | [opencode.ai/docs/rules](https://opencode.ai/docs/rules) — *"You can also have global rules in a `~/.config/opencode/AGENTS.md` file. This gets applied across all opencode sessions."* | 2026-08-20 |
| Glob-scoped rules | 18 rules, `paths:` rewritten to `globs:` + `match: any` | `~/.config/opencode/rules/` | Loaded conditionally by touched-file glob via the **`opencode-rules` plugin** (pinned 0.6.4) | [github.com/frap129/opencode-rules](https://github.com/frap129/opencode-rules) — *"`globs` (optional): Array of glob patterns for file-based matching"* | 2026-08-20 |
| Skills | 16 skills with injected `references/` | `~/.agents/skills/` | Agent-compatible global skill folder | [opencode.ai/docs/skills](https://opencode.ai/docs/skills) — *"Global agent-compatible: `~/.agents/skills/<name>/SKILL.md`"* | 2026-08-20 |
| Commands | 8 wrappers, one per user-invoked skill | `~/.config/opencode/commands/` | Global command folder | [opencode.ai/docs/commands](https://opencode.ai/docs/commands) — *"Global: ~/.config/opencode/commands/"* | 2026-08-20 |
| Agents | 25 generated `.md` | `~/.config/opencode/agents/` | Global agent folder | [opencode.ai/docs/agents](https://opencode.ai/docs/agents) — *"Global: ~/.config/opencode/agents/"* | 2026-08-20 |
| Plugin | `flow-session-context.ts` | `~/.config/opencode/plugins/` | Global plugin folder | [opencode.ai/docs/plugins](https://opencode.ai/docs/plugins) — *"`~/.config/opencode/plugins/` - Global plugins"* | 2026-08-20 |
| Permissions | `permission-config.json` | `permission.*` in `~/.config/opencode/opencode.json` | Per-tool permission keys, wildcard-overridable | [opencode.ai/docs/permissions](https://opencode.ai/docs/permissions) | 2026-08-20 |

The plugin's own directory resolution — `$OPENCODE_CONFIG_DIR/rules/` →
`$XDG_CONFIG_HOME/opencode/rules/` → `~/.config/opencode/rules/` — matches where the deploy
writes. **The 0.6.4 pin is the latest release** (2026-04-25), not a stale one; re-check
before assuming an upgrade is due.

## What does NOT reach it

- **`global/rules/quality/patterns-antipatterns.md`.** `build.py` builds the rules tree from
  `global/rules/languages/` and `global/rules/workflow/` only, so this is the one
  path-scoped rule with no glob channel here — it arrives injected inside the
  `language-rules` skill instead.
- **The 12 always-on rules, as files.** They are not duplicated into `rules/`; their
  condensed form is already in `harness/AGENTS.md`, which loads every session.
- **Claude Code's `disable-model-invocation`.** opencode does not honor it — the gate is
  reproduced by `permission.skill."flow-*": "ask"` in `opencode.json` plus the command
  wrappers, which are the only way a gated skill is exposed.

opencode also reads `~/.claude/CLAUDE.md` and `~/.claude/skills/` as a fallback, and needs
no env var to avoid double-listing: measured 2026-08-21 with `opencode debug skill`, the 33
pack skills all resolve from `~/.agents/skills` with zero duplicate names and
`~/.claude/skills` absent from every location. **Caveat that surfaced in the same run:** the
scan is RECURSIVE, so a skill directory that vendors its own `.codex/skills/` or nested
`.agents/skills/` publishes those too — `herdr` contributed three skills nobody chose.

## Unverified / undocumented dependencies

- **`experimental.chat.system.transform`** — the hook `flow-session-context.ts` uses to
  inject SessionStart context, because opencode has no real SessionStart event. It does
  **not** appear in [opencode.ai/docs/plugins](https://opencode.ai/docs/plugins); the only
  documented `experimental.*` hook there is `experimental.session.compacting`.
  **Status: undocumented surface, read from source.** It can break without a release note —
  the plugin degrades to not injecting on any failure, which is the intended failure mode,
  but a silent stop means the flow protocol quietly disappears from opencode sessions.
  Re-check this first when opencode sessions stop showing flow context.

## How to re-verify

```bash
ls ~/.config/opencode/rules/   | wc -l   # expect 18
ls ~/.config/opencode/agents/  | wc -l   # expect 25
ls ~/.config/opencode/commands/| wc -l   # expect 8
jq '.permission' ~/.config/opencode/opencode.json
```

## How it works

`/flow-build <args>` (opencode command) → instructs the agent to read
`harness-mechanics.md` (mechanic translation) + the skill body from
`~/.agents/skills/` → dispatches the generated subagents via the task tool.
Updating a skill or agent in the hub + `/deploy-global` updates this harness with no
adapter edits; only a renamed agent or a new skill requires touching this folder
(new command wrapper).

**Wrapper paths — no literal `~`.** Every `~/.agents/…` path in a wrapper is written
as a `` !`echo ~/…` `` shell injection, so opencode expands it to an absolute path at
prompt-build time. The Read tool does not expand tilde, and a model that passes the
literal `~` gets a failed read plus a recovery detour.
