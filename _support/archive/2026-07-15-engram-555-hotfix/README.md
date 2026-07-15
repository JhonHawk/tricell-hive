# Engram detect_project hotfix — issue #555 (portable, all three harnesses)

Upstream bug: the Engram harness plugins' `detect_project` ignores `.engram/config.json`
and derives the project from git remote / git root / cwd basename, splitting workspace
memories across alias projects (worktrees even yield task-slug names).
https://github.com/Gentleman-Programming/engram/issues/555 — open as of 2026-07-15,
affects v1.19.0 and main `be4b613`. **Delete this folder once upstream ships the fix.**

The fix (same logic, one patch per harness): `ENGRAM_PROJECT` env override → nearest
`.engram/config.json` walking ancestor directories (`project_name`, lowercased) →
original git/dirname fallbacks. Shell versions require `jq` (bundled on macOS).

## Apply on a machine (idempotent — check first)

```bash
# Already applied? (each listed file has the fix; missing file = not applied there)
grep -l "ENGRAM_PROJECT" \
  ~/.codex/plugins/cache/engram/engram/*/scripts/_helpers.sh \
  ~/.claude/plugins/cache/engram/engram/*/scripts/_helpers.sh \
  ~/.config/opencode/plugins/engram.ts 2>/dev/null

# Upstream already fixed? If the pristine script reads config.json natively, do NOT apply.

# Codex — both copies (cache + marketplace source, so cache rebuilds keep it):
patch -d ~/.codex/plugins/cache/engram/engram/<ver> -p1 < helpers-issue555.patch
patch -d ~/.codex/.tmp/marketplaces/engram/plugin/codex -p1 < helpers-issue555.patch

# Claude Code — both copies (the marketplace clone goes dirty on purpose:
# a plugin update then CONFLICTS instead of silently dropping the fix):
patch -d ~/.claude/plugins/cache/engram/engram/<ver> -p1 < helpers-issue555-claude-code.patch
patch -d ~/.claude/plugins/marketplaces/engram/plugin/claude-code -p1 < helpers-issue555-claude-code.patch

# opencode — single file, not a cache:
patch -d ~/.config/opencode -p1 < engram-ts-issue555-opencode.patch

# Verify from any workspace that has a .engram/config.json (expect its project_name):
bash -c 'source ~/.claude/plugins/cache/engram/engram/<ver>/scripts/_helpers.sh; \
  detect_project <path-inside-a-configured-workspace>'
```

Re-apply triggers: the harness's engram plugin updates (new version dir / marketplace
`last_updated` change / hooks ask re-trust), or memories start splitting again.
If a patch stops applying cleanly, upstream rewrote the script — re-derive from the
pristine upstream source instead of blind-copying old fixed files.

## engram setup side-effects (any machine, same vintage)

- **Codex:** `engram setup` re-adds `model_instructions_file` to `~/.codex/config.toml`,
  which REPLACES Codex's base system prompt — delete that line.
- **Claude Code:** setup may create `~/.claude/mcp/engram.json`, a duplicate standalone
  MCP registration with a brew-Cellar-versioned path — delete it (the plugin already
  provides the MCP server).
- **opencode:** setup is a net no-op.

Machine-local incident record (this machine only, gitignored):
`_support/backup/2026-07-15-engram-codex-hook-hotfix/` — this archive folder is the
portable mirror of its patches.
