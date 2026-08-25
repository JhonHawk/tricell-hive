---
name: deploy-global
description: Deploy the full global/ directory (CLAUDE.md, rules, agents, skills, hooks) to ~/.claude/ with automatic backup, plus the multi-harness layer — universal skills to ~/.agents/skills/, generated agents to ~/.codex/agents/, ~/.config/opencode/agents/, and ~/.grok/agents/, opencode commands, harness/AGENTS.md to ~/.codex/AGENTS.md + ~/.config/opencode/AGENTS.md, and the always-on rules symlinked flat into ~/.grok/rules/. Also idempotently merges hook blocks into ~/.claude/settings.json (additive, never overwriting preferences). Use when ready to deploy config changes.
disable-model-invocation: true
---

Deploys `global/` to `~/.claude/`, then the multi-harness layer (Codex + opencode, derived
from the same sources under `harness/`) — via `scripts/deploy-global.sh`, a real, idempotent
shell script. This skill is the interface over it; it does not restate the procedure.

**User-initiated only.** Never invoke this skill or the script on your own — deploying to
`~/.claude/`, `~/.codex/`, `~/.config/opencode/`, or `~/.grok/` is always the user's explicit call.

## The `grok` scope

Grok reads `~/.claude/CLAUDE.md` natively, but its rules discovery is **not recursive** and
does **not** honor `paths:`. Every rule in `global/rules/` lives one level down, so none of them
ever reached it. The scope closes that gap in two parts:

1. **Rules** — one flat **file** symlink per always-on rule into `$GROK_HOME/rules/`
   (default `~/.grok/rules/`), named `<dir>__<file>.md`, pointing at the deployed rule under
   `~/.claude/rules/` — so Claude and Grok read the same bytes. Directory symlinks do not
   work; the scan still refuses to recurse. Path-scoped rules are deliberately excluded
   (Grok would load them always-on); they reach it through the router skills
   (`language-rules`, `workspace-conventions`) like on Codex. A rule that later gains
   `paths:` turns its link into a manifest-detected orphan.

   That router path only works because the claude scope now also deploys the
   **build-injected `references/`** into `~/.claude/skills/` (`deploy_injected_references`).
   Grok scans `~/.claude/skills/` and never `~/.agents/skills/`, so while the injected sets
   existed only under `harness/agents-skills`, every router on Grok pointed at reference
   files reachable from no path it knew — the routing table loaded, the targets did not
   exist, and the model guessed. Keep the two in step: a new entry in
   `SKILL_REFERENCE_INJECTIONS` reaches Grok only after a deploy of the claude scope.
2. **Agents** — copies generated `harness/grok/agents/*.md` into `$GROK_HOME/agents/`
   (real files; Grok frontmatter differs from Claude). Built by `harness/build.py` from
   `global/agents/`. Hive hooks stay single-source under `~/.claude/hooks/` (Grok merges
   `settings.json` via compat); scripts are dual-runtime for Claude + Grok payloads.

**Herdr note:** Herdr installs its own SessionStart under both `~/.claude/settings.json` and
`~/.grok/hooks/` — that double registration is third-party, not this deploy.

## Before running

`harness/` is generated from `global/agents` and `global/skills` by `harness/build.py`,
which also assembles the two always-on cores (`global/CLAUDE.md` and `harness/AGENTS.md`)
from `global/core-sections/`. If canonical sources changed since the last build, rebuild
and commit first:
```bash
python3 harness/build.py && git status --porcelain harness/ global/CLAUDE.md
```
The script also rebuilds automatically under `--apply` (every scope, before deploying
`global/CLAUDE.md`, so the deployed core is the fresh assembly) and warns if `harness/`
or `global/CLAUDE.md` comes out dirty — but a clean, committed tree going in is the
judgment call the script can't make for you. A hand-edited core aborts the rebuild (and
with it the deploy) before anything is copied. **The dry-run diff for harness-derived
categories (`agents-skills`, `codex-agents`, `opencode-agents`, etc.) is labeled
`(pre-rebuild)`**: dry-run never rebuilds `harness/` (it never writes anything, full stop),
so that diff reflects whatever `harness/` held on disk at run time, not what a fresh
`build.py` would produce. Rebuild first if you need the diff to be current.

## Machine preflight — agent shell config (report-only, never mutates)

The zsh-mine mitigation (`global/CLAUDE.md > Shell`) assumes machine state this repo cannot
deploy. Check it on every deploy and report ✓/✗ per line; a ✗ is a drift warning for the
user — this skill never edits shell config or settings to fix it:

```bash
S=$(jq -r '.env.CLAUDE_CODE_SHELL // empty' ~/.claude/settings.json)
echo "CLAUDE_CODE_SHELL=${S:-MISSING}"                    # expect /opt/homebrew/bin/bash
[ -x "$S" ] && "$S" -c 'echo "bash=$BASH_VERSION BASHPID=${BASHPID:-EMPTY}"'
    # expect major >= 5 AND a real PID — bash 3.2 is actively broken with the
    # snapshot's BASHPID branch, and an empty BASHPID means that broken path
grep -c 'zsh_special_re' "$HOME/.claude/hooks/bash-policy.sh"
    # expect >= 1: the path=/status= deny (Grok scope) survived the deploy
grep -c 'unsetopt nomatch' "$HOME/.zshenv"
    # expect >= 1: the agent-shell nomatch guard (covers Codex `zsh -lc` and
    # Claude's zsh fallback) is still present in the user's zshenv
P=$(command -v pnpm) && readlink -f "$P"
    # expect a path NOT containing "corepack" (e.g. ~/Library/pnpm/bin/pnpm —
    # the standalone install). A corepack shim means pnpm is corepack-managed,
    # which pnpm upstream deprecated (ago-2026) and Node 25+ no longer bundles
grep -c '^[^#]*--corepack-enabled' "$HOME/.zshrc"
    # expect 0 (non-comment lines only): fnm's --corepack-enabled flag re-shims
    # pnpm through corepack in every new shell, shadowing the standalone install
```

- Missing/non-executable `CLAUDE_CODE_SHELL` → Claude Code degrades to zsh auto-detection
  (benign — the zsh rules in `CLAUDE.md > Shell` reapply; still report the drift).
- pnpm resolving to a corepack shim, or `--corepack-enabled` present in `.zshrc` → warn:
  installs will break on Node 25+ (corepack no longer bundled) and go through corepack's
  integrity layer meanwhile. Fix (user-run, never automatic): install the standalone
  (`curl -fsSL https://get.pnpm.io/install.sh | sh -`), drop `--corepack-enabled` from the
  fnm env line, and remove stale shims with `corepack disable pnpm`.
- After changing `settings.json` `env`, a FULL Claude Code restart is required (the daemon
  caches the environment); a new session under the old daemon keeps the old shell.

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
| `--only SCOPE[,SCOPE...]` | Restrict to `claude` (→ `~/.claude`), `codex` (→ `~/.codex`), `opencode` (→ `~/.config/opencode`), `grok` (→ `~/.grok/rules` + `~/.grok/agents` + shared skills), `harness` (alias for `codex,opencode,grok`), or `all` (default). Repeatable or comma-separated. Note: under `--apply`, `--only harness`/`--only grok` deletes that scope's orphans by default with NO backup snapshot — backups cover only `~/.claude`; the content is regenerable from the repo, but the deletion runs unshielded. |
| `--keep-orphans` | Under `--apply`, list manifest-confirmed orphans WITHOUT deleting them (the old default). By default orphan deletion is part of the apply flow: an orphan is by construction a source the user already removed from `global/`/`harness/`, deploys are user-initiated, and the full list prints in the report before deletion — so a separate confirmation flag re-asked what the repo's own history already answered. The deletion also purges the orphans' `settings.json`/`hooks.json` entries. Refused (not deleted) if the set exceeds 20 entries or 25% of the manifest — that volume looks like a broken checkout, not a routine cleanup; the refusal is the anomaly brake. Undeleted orphans (dry-run, kept, or refused) keep their manifest entries, so a later run can still remove them. |
| `--force-delete-orphans` | Bypasses the size-based refusal above. Only pass this when a large, deliberate orphan set is genuinely expected (e.g. a major agent restructuring) — never as a default response to the refusal message. |
| `--delete-orphans` | Deprecated no-op alias, accepted for compatibility — deletion is now the `--apply` default. |
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

- Deletes in exactly three bounded classes, nothing else — and nothing in a dry run:
  (1) **manifest-confirmed orphans**, under `--apply` — the full list prints in the report
  before deletion, the size refusal brakes it, and `--keep-orphans` exempts exactly this
  class; (2) **stale flat duplicates** — a rule/agent file sitting flat under
  `~/.claude/rules`/`~/.claude/agents` superseded by a same-named file in a subdirectory
  (legacy of the pre-nested layout), removed under `--apply`; (3) **backup rotation** —
  keeps the 5 most recent `~/.claude` snapshots, removes older ones. `--keep-orphans` does
  NOT exempt classes 2-3.
- Backs up `~/.claude` (tar, keeps 5 most recent) before any overwrite — but only when the
  `claude` scope is active; there's no equivalent snapshot for `~/.codex` or
  `~/.config/opencode` in this procedure. The `grok` scope needs none: it writes only
  symlinks, so removing them can never destroy a rule.
- The `grok` scope skips (with a WARNING, never a dangling link) any rule not yet present
  under `~/.claude/rules/` — run the `claude` scope at least once first.
- Never overwrites `~/.claude/settings.json`, `~/.codex/hooks.json`, or
  `~/.config/opencode/opencode.json` wholesale — only surgical, validated (temp-file +
  `jq empty` before move) merges of repo-managed blocks, via the bundled
  `filters/hook-merge.jq` and `filters/hook-purge.jq` for hooks, and
  `harness/opencode/permission-config.json` for the opencode permission keys. That
  permission merge adds a repo rule only when the user has not already declared the same
  pattern (`~` expanded on both sides for the comparison), never rewrites their spelling,
  and skips a `permission` set to a bare action string. Only `opencode.json` is merged —
  a comments-bearing `opencode.jsonc` is reported and left alone, since `jq` cannot
  round-trip comments.
- A partial run (`--only claude`, `--only codex`, etc.) preserves the untouched scopes'
  entries in `~/.claude/.deploy-manifest` rather than dropping them — so orphan detection
  for scopes you didn't just run stays accurate on the next deploy.
- Refuses to run at all against a checkout that doesn't look like tricell-hive (missing
  `global/CLAUDE.md`, `global/rules`, or `global/agents`), and refuses an oversized orphan
  deletion (see `--keep-orphans` above) — both are the compensating controls for a
  mis-resolved repo root turning "nothing to deploy" into "delete everything".

## After a deploy

Remind the user to restart Claude Code / Codex / opencode (or start a new session) to
reload — a running session does not pick up the new files. **On a fresh machine (or any run
touching the `codex`/`opencode` scopes for the first time)**, also point them at
`harness/{codex,opencode}/README.md`: the `*.snippet` config merges (plugin/hook
registration in `opencode.jsonc` / `config.toml`) are one-time and manual — this script writes
no part of those files except the opencode `permission` keys above, so hooks and rules can land
on disk with nothing registering them until that merge happens. The script's own final report repeats this reminder whenever the
`codex` or `opencode` scope ran.
