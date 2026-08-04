---
name: deploy-global
description: Deploy the full global/ directory (CLAUDE.md, rules, agents, skills, hooks) to ~/.claude/ with automatic backup, plus the multi-harness layer — universal skills to ~/.agents/skills/, generated agents to ~/.codex/agents/ and ~/.config/opencode/agents/, opencode commands, harness/AGENTS.md to ~/.codex/AGENTS.md + ~/.config/opencode/AGENTS.md, and the always-on rules symlinked flat into ~/.grok/rules/. Also idempotently merges hook blocks into ~/.claude/settings.json (additive, never overwriting preferences). Use when ready to deploy config changes.
disable-model-invocation: true
---

Deploys `global/` to `~/.claude/`, then the multi-harness layer (Codex + opencode, derived
from the same sources under `harness/`) — via `scripts/deploy-global.sh`, a real, idempotent
shell script. This skill is the interface over it; it does not restate the procedure.

**User-initiated only.** Never invoke this skill or the script on your own — deploying to
`~/.claude/`, `~/.codex/`, `~/.config/opencode/`, or `~/.grok/` is always the user's explicit call.

## The `grok` scope

Grok reads `~/.claude/CLAUDE.md` natively, but its rules discovery is **not recursive** and
does **not** honor `paths:` (both verified against grok 0.2.118 with marker files under a
temporary `GROK_HOME`). Every rule in `global/rules/` lives one level down, so none of them
ever reached it. The scope closes that gap: one flat **file** symlink per always-on rule into
`$GROK_HOME/rules/` (default `~/.grok/rules/`), named `<dir>__<file>.md`, pointing at the
deployed rule under `~/.claude/rules/` — so both harnesses read the same bytes and one
redeploy updates both. Directory symlinks do not work; the scan still refuses to recurse.

Path-scoped rules are deliberately excluded: Grok would load them always-on. They reach it
through the router skills (`language-rules`, `workspace-conventions`) like on Codex.
A rule that later gains `paths:` turns its link into a manifest-detected orphan.

## Before running

`harness/` is generated from `global/agents` and `global/skills` by `harness/build.py`.
If canonical sources changed since the last build, rebuild and commit first:
```bash
python3 harness/build.py && git status --porcelain harness/
```
The script also rebuilds automatically under `--apply` (for the `codex`/`opencode` scopes)
and warns if `harness/` comes out dirty — but a clean, committed `harness/` going in is the
judgment call the script can't make for you. **The dry-run diff for harness-derived
categories (`agents-skills`, `codex-agents`, `opencode-agents`, etc.) is labeled
`(pre-rebuild)`**: dry-run never rebuilds `harness/` (it never writes anything, full stop),
so that diff reflects whatever `harness/` held on disk at run time, not what a fresh
`build.py` would produce. Rebuild first if you need the diff to be current.

## Running it

Always run from the repo root, or let the script find it (it resolves the repo root from
its own location regardless of cwd):
```bash
.claude/skills/deploy-global/scripts/deploy-global.sh [flags]
```

**Two-step confirmation flow:**
1. Run with no flags (or `--dry-run` explicitly) first — this is the default and touches
   nothing. Review the diff/orphan/backup report with the user.
2. Only on explicit user confirmation, re-run with `--apply`.

### Flags

| Flag | Effect |
|---|---|
| `--dry-run` | Report what would change; write nothing. **Default.** |
| `--apply` | Perform the deploy for real. |
| `--only SCOPE[,SCOPE...]` | Restrict to `claude` (→ `~/.claude`), `codex` (→ `~/.codex`), `opencode` (→ `~/.config/opencode`), `grok` (→ `~/.grok/rules`), `harness` (alias for `codex,opencode,grok`), or `all` (default). Repeatable or comma-separated. |
| `--delete-orphans` | Under `--apply`, actually delete manifest-confirmed orphans (files whose source was removed from `global/`/`harness/`) and purge their `settings.json`/`hooks.json` entries. **Without this flag orphans are only listed, never removed** — this is the confirmation gate; present the orphan list to the user before ever passing it. Refused (not deleted) if the orphan set exceeds 20 entries or 25% of the manifest — that volume looks like a broken checkout, not a routine cleanup. |
| `--force-delete-orphans` | Implies `--delete-orphans` and bypasses the size-based refusal above. Only pass this when a large, deliberate orphan set is genuinely expected (e.g. a major agent restructuring) — never as a default response to the refusal message. |
| `--verbose`, `-v` | Per-file logging instead of per-category summaries. |
| `--help`, `-h` | Flag reference. |

## Reading the output

Narration streams to stderr as the script works; a structured report prints to stdout at
the end — diff summary (new/modified/unchanged per category), backup path, orphans found
(and whether they were deleted), per-scope deploy counts, the hook-merge outcome for
`settings.json` and `~/.codex/hooks.json`, the Engram #555 hotfix result, and a restore
command. A `WARNING:` line means that one step degraded safely (e.g. a hook merge left
`settings.json` untouched, or `harness/` came out dirty) — the rest of the deploy still ran;
surface every `WARNING:` to the user, don't just report success.

## What the script guarantees

- Never deletes anything outside the manifest-confirmed orphan set, and never deletes even
  those without `--apply --delete-orphans` explicitly passed.
- Backs up `~/.claude` (tar, keeps 5 most recent) before any overwrite — but only when the
  `claude` scope is active; there's no equivalent snapshot for `~/.codex` or
  `~/.config/opencode` in this procedure. The `grok` scope needs none: it writes only
  symlinks, so removing them can never destroy a rule.
- The `grok` scope skips (with a WARNING, never a dangling link) any rule not yet present
  under `~/.claude/rules/` — run the `claude` scope at least once first.
- Never overwrites `~/.claude/settings.json` or `~/.codex/hooks.json` wholesale — only
  surgical, validated (temp-file + `jq empty` before move) merges of repo-managed hook
  blocks, via the bundled `filters/hook-merge.jq` and `filters/hook-purge.jq`.
- A partial run (`--only claude`, `--only codex`, etc.) preserves the untouched scopes'
  entries in `~/.claude/.deploy-manifest` rather than dropping them — so orphan detection
  for scopes you didn't just run stays accurate on the next deploy.
- Refuses to run at all against a checkout that doesn't look like tricell-hive (missing
  `global/CLAUDE.md`, `global/rules`, or `global/agents`), and refuses an oversized orphan
  deletion (see `--delete-orphans` above) — both are the compensating controls for a
  mis-resolved repo root turning "nothing to deploy" into "delete everything".

## After a deploy

Remind the user to restart Claude Code / Codex / opencode (or start a new session) to
reload — a running session does not pick up the new files. **On a fresh machine (or any run
touching the `codex`/`opencode` scopes for the first time)**, also point them at
`harness/{codex,opencode}/README.md`: the `*.snippet` config merges (plugin/hook
registration in `opencode.jsonc` / `config.toml`) are one-time and manual — this script never
writes those config files, so hooks and rules can land on disk with nothing registering them
until that merge happens. The script's own final report repeats this reminder whenever the
`codex` or `opencode` scope ran.
