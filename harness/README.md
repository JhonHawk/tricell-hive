# Multi-harness layer

Generic AGENTS-compatible configuration derived from `global/`, so Codex and opencode
run the same agents, skills, and rules as Claude Code. Claude Code reads `global/`
directly (deployed to `~/.claude/` by `/deploy-global`); this layer adapts the same
canonical sources for the other two CLIs.

## Layout

| Path | What | Maintained how |
|---|---|---|
| `AGENTS.md` | Condensed cross-harness guidance → `~/.codex/AGENTS.md` + `~/.config/opencode/AGENTS.md` | Hand-edited; budget ~30 KiB (`build.py` warns), hard limit 48 KiB |
| `build.py` | Regenerates every generated tree below from `global/` | Run after any agent/skill edit |
| `agents-skills/` | Cleaned universal skills → `~/.agents/skills/` | **Generated** |
| `codex/agents/` | TOML subagents → `~/.codex/agents/` | **Generated** |
| `opencode/agents/` | Markdown subagents → `~/.config/opencode/agents/` | **Generated** |
| `opencode/commands/` | `/flow-*` wrappers → `~/.config/opencode/commands/` | Hand-edited (one per skill) |
| `codex/config.toml.snippet` | Codex config additions | **Manual merge, once** |
| `opencode/opencode.jsonc.snippet` | opencode config additions | **Manual merge, once** |

Per-harness detail lives in `codex/README.md`, `opencode/README.md`, and
`agents-skills/README.md`. This file covers the layer as a whole and the snippets in
depth.

## Three maintenance classes

1. **Generated** (`build.py`) — never hand-edit; change the `global/` source and rebuild.
2. **Copied** (`/deploy-global`) — skills, command wrappers, `AGENTS.md`. The deploy
   overwrites these targets because they are fully repo-managed.
3. **Manual merge** — the two `*.snippet` files. This is the one class the deploy
   **cannot** automate; the rest of this README explains why and how.

## Why snippets are not deployed

`/deploy-global` freely replaces the files it fully owns (`~/.codex/AGENTS.md`, the
generated agents, the skills). But `~/.codex/config.toml` and
`~/.config/opencode/opencode.json` are **your** files — they hold your model choice, your
MCP servers, your trusted-project list, your secrets. Overwriting them would destroy
configuration the repo knows nothing about. So the flow-pack additions ship as `.snippet`
fragments you merge by hand, once.

They are **idempotent in intent**: once merged you don't repeat them on each deploy —
unless a snippet itself changes (e.g. the `AGENTS.md` hard limit in `build.py` moves, so
`project_doc_max_bytes` must follow to stay in sync).

## What each snippet adds

### `codex/config.toml.snippet` → `~/.codex/config.toml`

- **`project_doc_max_bytes = 49152`** — the one that matters. Codex truncates `AGENTS.md`
  **silently** past its 32 KiB default, dropping the last sections. This raises the
  ceiling to 48 KiB. Keep it in sync with `AGENTS_HARD_LIMIT_BYTES` in `build.py`.
  TOML note: it is a top-level key — it must stay **above** the `[agents]` table.
- **`[agents] max_threads / max_depth`** — subagent coordination for flow-build-style
  sessions (`max_depth = 1` is enough: orchestrator → workers).
- **Linear MCP** (commented) — only if you drive flow-specs/flow-plan/flow-build with Linear and
  don't already have Codex's native Linear plugin.

### `opencode/opencode.jsonc.snippet` → `~/.config/opencode/opencode.json`

- **`permission.skill."flow-*": "ask"`** — opencode ignores Claude Code's
  `disable-model-invocation`, so this gates the flow skills behind your approval.
- **`instructions`** (commented, optional) — makes the shared rubrics ambient;
  reinforcement, not a requirement.
- **Shell env** (not a JSON key): `export OPENCODE_DISABLE_CLAUDE_CODE_SKILLS=1` in your
  shell profile. Skills deploy to both `~/.claude/skills` and `~/.agents/skills`; this
  keeps a single discovery source so opencode doesn't list them twice.

## Apply / verify

Merging is "add only what's missing" — applying twice must not duplicate keys (TOML and
JSON both reject duplicates / silently last-wins). Check current state before editing:

```bash
grep -q 'project_doc_max_bytes' ~/.codex/config.toml          && echo "codex: AGENTS airbag ok"
grep -q '^\[agents\]'           ~/.codex/config.toml          && echo "codex: subagents ok"
grep -q 'flow-\*'  ~/.config/opencode/opencode.json           && echo "opencode: skill gate ok"
grep -q 'OPENCODE_DISABLE_CLAUDE_CODE_SKILLS' ~/.zshrc        && echo "shell: discovery flag ok"
```

Each line that prints is already applied — skip it. After merging anything new, start a
fresh Codex / opencode session to reload.

> Secrets: these config files commonly hold API tokens in plaintext (MCP headers). They
> are personal files outside this repo — never copy their literal values into the repo,
> documentation, or a backup you might share. Prefer environment variables where the CLI
> supports them.

## Rebuild + deploy

```bash
python3 harness/build.py      # regenerate the generated trees from global/
# then, from the hub:
/deploy-global                # copies everything; reports a dirty harness/ if you forgot to rebuild
```

A **new skill** needs an opencode command wrapper in `opencode/commands/`; a **renamed**
agent or skill needs a grep through `harness/`. Rules and `global/CLAUDE.md` do **not**
pass through `build.py` — a cross-harness rule change must be mirrored manually into
`AGENTS.md`.
