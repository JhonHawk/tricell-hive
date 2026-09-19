---
name: deploy-global
description: Deploy the full global/ directory (CLAUDE.md, agents, skills, hooks) to ~/.claude/ with automatic backup, plus the multi-harness layer — universal skills to ~/.agents/skills/, generated agents to ~/.codex/agents/, ~/.config/opencode/agents/, ~/.grok/agents/, and the PI agent root, opencode commands, harness/AGENTS.md to ~/.codex/AGENTS.md + ~/.config/opencode/AGENTS.md, and the build-injected router references into ~/.claude/skills/. Also idempotently merges hook blocks into ~/.claude/settings.json (additive, never overwriting preferences). PI is included by default/all/harness selections and can be selected with shared skills through PI_CODING_AGENT_DIR; PI and shared-skills manifests and rollback backups preserve user-owned state while other harnesses retain their legacy backup paths. Use when ready to deploy config changes.
disable-model-invocation: true
---

Deploys `global/` to `~/.claude/`, then the multi-harness layer (Codex + opencode + Grok + Pi, derived
from the same sources under `harness/`) — via `scripts/deploy-global.sh`, a real, idempotent
shell script. This skill is the interface over it; it does not restate the procedure.

**User-initiated only.** Never invoke this skill or the script on your own — deploying to
`~/.claude/`, `~/.codex/`, `~/.config/opencode/`, `~/.grok/`, `~/.agents/`, or the PI root
is always the user's explicit call.

## The `grok` scope

**There is no rules step.** Grok reads `~/.claude/CLAUDE.md` natively, so the 7 rules the
core inlines arrive with it; every other rule reaches Grok through a router skill under
`~/.claude/skills/` or the `rule-delivery` hold, exactly as on Codex. The scope writes agents
and nothing else.

Flat `<dir>__<file>.md` symlinks a pre-2026-09 deploy left under `$GROK_HOME/rules/` (default
`~/.grok/rules/`) now resolve to no source and are swept as manifest-detected orphans on the
next `--apply`. Together with the retired `~/.claude/rules/` files that is a large one-time
orphan set: the size refusal fires, the report says so in as many words, and the run is
re-issued **once** with `--force-delete-orphans`. That is the only occasion for that flag.

The router path only works because the claude scope also deploys the **build-injected
`references/`** into `~/.claude/skills/` (`deploy_injected_references`). Grok scans
`~/.claude/skills/` and never `~/.agents/skills/`, so while the injected sets existed only
under `harness/agents-skills`, every router on Grok pointed at reference files reachable from
no path it knew — the routing table loaded, the targets did not exist, and the model guessed.
Keep the two in step: a new entry in `SKILL_REFERENCE_INJECTIONS` reaches Grok only after a
deploy of the claude scope.

**Agents** — the scope copies generated `harness/grok/agents/*.md` into `$GROK_HOME/agents/`
(real files; Grok frontmatter differs from Claude). Built by `harness/build.py` from
`global/agents/`. Hive hooks stay single-source under `~/.claude/hooks/` (Grok merges
`settings.json` via compat); scripts are dual-runtime for Claude + Grok payloads.

**Herdr note:** Herdr installs its own SessionStart under both `~/.claude/settings.json` and
`~/.grok/hooks/` — that double registration is third-party, not this deploy.

## Before running

`harness/` is generated from `global/agents` and `global/skills` by `harness/build.py`,
which also assembles the two always-on cores (`global/CLAUDE.md` and `harness/AGENTS.md`)
from `global/core-sections/`. The versioned `.githooks/pre-commit` regenerates and stages
those outputs with any commit that touches a canonical source (enable once per clone:
`git config core.hooksPath .githooks`). A "dirty after rebuild" warning from this script
means the hook is not enabled or sources changed without a commit — rebuild and commit first:
```bash
python3 harness/build.py && git status --porcelain harness/ global/CLAUDE.md
```

**Claude Code's agents ship from `harness/claude/agents/`, not from `global/agents/`** —
preview, diff, deploy, and orphan detection all resolve there. The generated tree mirrors
the source's role subfolders and differs from the source in exactly what the build adds: the
do-not-edit note, the rule texts an agent's `packs:` declare inlined into its body, and
`omitClaudeMd: true` on a packed agent (its packs ARE its rule corpus, so the global one is
dropped rather than stacked underneath). An agent with no packs renders its source body
unchanged. Deploying `global/agents/` would hand Claude Code the unpacked source — an agent
whose conventions silently went missing.

The PI source tree is generated alongside the other harness outputs. All selected roots
complete preflight before the first target write. A dry run remains read-only: it does not
rebuild or write generated files, and it runs `harness/build.py --check` for generated-tree
parity. Under `--apply`, the script runs `harness/build.py` exactly once before the diff,
then preflights the selected roots before backups and copies. Apply does not run a second
generated-tree check. A stale or hand-edited generated output aborts the selected apply
before target writes; run the generator explicitly first when you want to inspect the
generated diff in advance.

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
jq -r '.shell // "MISSING"' ~/.config/opencode/opencode.json
jq -r '.shellPath // "MISSING"' "${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}/settings.json"
    # expect /opt/homebrew/bin/bash on both: opencode `shell` and Pi `shellPath` are
    # deploy-managed (user value wins); MISSING after a deploy means the merge was
    # skipped (jsonc-only opencode config, or the binary is absent)
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
| `--only SCOPE[,SCOPE...]` | Restrict to `claude` (→ `~/.claude`), `codex` (→ `~/.codex`), `opencode` (→ `~/.config/opencode`), `grok` (→ `~/.grok/agents` + shared skills), `pi` (→ `${PI_CODING_AGENT_DIR:-~/.pi/agent}`), `harness` (alias for `codex,opencode,grok,pi`), or `all` (default: every harness). Repeatable or comma-separated; mixed selections are allowed. `pi` includes the neutral shared-skills preflight and writes only PI plus shared skills, never other harness installations or the legacy Claude manifest. The neutral engine owns Pi/shared-skills manifests and backups; other selected harnesses retain their legacy writer and backup path. |
| `--keep-orphans` | Under `--apply`, list manifest-confirmed orphans WITHOUT deleting them (the old default). By default orphan deletion is part of the apply flow: an orphan is by construction a source the user already removed from `global/`/`harness/`, deploys are user-initiated, and the full list prints in the report before deletion — so a separate confirmation flag re-asked what the repo's own history already answered. The deletion also purges the orphans' `settings.json`/`hooks.json` entries. Refused (not deleted) if the set exceeds 20 entries or 25% of the manifest — that volume looks like a broken checkout, not a routine cleanup; the refusal is the anomaly brake. Undeleted orphans (dry-run, kept, or refused) keep their manifest entries, so a later run can still remove them. |
| `--force-delete-orphans` | Bypasses the size-based refusal above. Only pass this when a large, deliberate orphan set is genuinely expected — never as a default response to the refusal message. The one standing case is the first `--apply` after the rule store dissolved: the retired `~/.claude/rules/` files and the Grok flat symlinks that pointed at them are all orphans at once, and the report names that class explicitly before you re-run. |
| `--delete-orphans` | Deprecated no-op alias, accepted for compatibility — deletion is now the `--apply` default. |
| `--verbose`, `-v` | Per-file logging instead of per-category summaries. |
| `--help`, `-h` | Flag reference. |

When a managed hook is retired, orphan cleanup removes only the exact Hive command
registration for that hook and prunes an event block only after its inner hook list is
empty. Co-located user hooks, commands at another path, and commands with other arguments
remain. The retired plan-capture script is removed only when its historical Hive bytes still
match; an edited or unprovable copy is preserved and reported while its exact legacy
registration is still removed so native approval cannot keep running the retired bridge.

For PI, `--apply` deploys `harness/AGENTS.md`, generated `harness/pi/agents/`, the
TypeScript runtime and extension entrypoints (`src/` and `extensions/`), and canonical
shell hooks under `global/hooks/` in the PI agent directory. It
also merges only the owned fields: the five pinned package entries and `shellPath`
(Homebrew bash 5) in `settings.json`,
including PI package objects with their existing filters,
`forceTopLevelAsync` in `extensions/subagent/config.json`, the managed Context7, Linear, and
HeroUI proxies in `mcp.json` (each lazy, direct tools disabled, `includeTools` from
`harness/pi/src/mcp-allowlist.json`) plus `settings.scriptMode: false`, and `provider: openai` with
`openaiSearchProviders: [openai-codex]` plus `workflow: none` in `web-search.json`.
Existing PI config, credentials,
sessions, trust state, unrelated providers, and unrelated packages remain in place.

Install the five exact native packages into `${PI_CODING_AGENT_DIR:-~/.pi/agent}/npm`
before applying: `pi-subagents@0.67.0`, `gentle-engram@0.1.13`,
`pi-mcp-adapter@2.33.0`, `@juicesharp/rpiv-ask-user-question@2.9.0`, and
`pi-web-access@0.29.0`. The helper never installs packages: it fails closed unless
all five identities and versions are present, PI 0.85.1 is the selected runtime, and
the reviewed patch metadata and target hashes match. Apply atomically rewrites the three
reviewed `pi-subagents` files only when their bytes match the recorded pristine hashes,
records the patched hashes in the PI manifest, and re-applies the patch after a pristine
same-version reinstall. A modified or symlinked package target is preserved as an error;
the same backup and conflict-aware rollback engine covers package files as the rest of
the PI root.

The PI route defaults to dry-run just like the other scopes:

```bash
.claude/skills/deploy-global/scripts/deploy-global.sh --only pi
PI_CODING_AGENT_DIR="$HOME/.pi/agent" \
  .claude/skills/deploy-global/scripts/deploy-global.sh --only pi --apply
```

The deploy helper also exposes conflict-aware rollback without invoking a package
installer:

```bash
python3 harness/pi/deploy.py rollback \
  --pi-dir "${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}" \
  --backup-dir "${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}/.hive-deploy-backups/<stamp>"
python3 harness/pi/deploy.py rollback --apply \
  --pi-dir "${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}" \
  --backup-dir "${PI_CODING_AGENT_DIR:-$HOME/.pi/agent}/.hive-deploy-backups/<stamp>"
```

### PI deployment matrix

| Surface | Copied or staged | Wired or registered | Failure policy |
|---|---|---|---|
| Core and roles | `harness/AGENTS.md`, generated PI roles, runtime, extensions, and canonical hook scripts | Pi agent discovery, extension entries, and `pi-subagents` role selection | Missing or stale generated output blocks the selected apply before target writes |
| Managed settings | Owned fields in `settings.json`, `extensions/subagent/config.json`, `mcp.json`, and `web-search.json` | Five package pins, `forceTopLevelAsync`, the managed Context7, Linear, and HeroUI `mcp` proxies, and OpenAI web search | Invalid JSON or an ownership conflict blocks; unrelated user fields remain unchanged |
| Packages and patch | Preinstalled exact packages plus the reviewed `pi-subagents` patch metadata | Package identities and before/after target hashes | No auto-install; missing pins, symlinks, third hashes, or patch mismatch block |
| Shared skills | Generated universal skills in the neutral shared-skills target | Shared manifest and neutral backup/rollback records | Legacy conflicts are preserved and reported with hash/source-commit provenance |
| Advisory hooks | Canonical scripts and PI adapter wiring | Parent-only flow and hygiene advisories | Missing advisory output warns and continues; blocking guards remain fail-closed |

The neutral shared-skills manifest is `~/.agents/.hive-deploy-manifest.json` with adjacent
private backups. The old path-only Claude manifest is read-only migration evidence: only
matching current bytes or matching blobs at its recorded `source_commit` can be adopted.
Unknown differences remain untouched and are reported as conflicts. Pi and shared-skills roots
are backed up before their first neutral-engine write and have their own rollback records;
other selected harnesses retain their legacy backup path. A partial apply reports the affected
root instead of claiming a global transaction.

## Reading the output

Narration streams to stderr as the script works; a structured report prints to stdout at
the end — preflight and generated-tree parity, diff summary (new/modified/unchanged per
category), backup/rollback records, orphans found (and whether they were deleted),
per-scope deploy counts, and the hook/config merge outcomes. A `WARNING:` line means that
one advisory step degraded safely (for example, an advisory hook is unavailable) — the
rest of the deploy still ran; surface every `WARNING:` to the user, don't just report
success. A failed preflight or ownership conflict is a blocking error and occurs before
the selected apply writes targets.

## What the script guarantees

- Deletes in exactly three bounded classes, nothing else — and nothing in a dry run:
  (1) **manifest-confirmed orphans**, under `--apply` — the full list prints in the report
  before deletion, the size refusal brakes it, and `--keep-orphans` exempts exactly this
  class; (2) **stale flat duplicates** — an agent file sitting flat under `~/.claude/agents`
  superseded by a same-named file in a subdirectory (legacy of the pre-nested layout),
  removed under `--apply`; (3) **backup rotation** —
  keeps the 5 most recent `~/.claude` snapshots, removes older ones. `--keep-orphans` does
  NOT exempt classes 2-3.
- Uses the neutral deploy/rollback engine for the selected Pi and shared-skills roots. Each
  of those roots is backed up before its first neutral-engine write, and the report names the
  root-specific restore record. Shared skills use `~/.agents/.hive-deploy-manifest.json` and
  adjacent private backups; the Pi root keeps its managed package/config journal. Other
  selected harnesses retain their legacy writer and backup path. No engine snapshots or
  rewrites an unselected root.
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
- A partial run preserves untouched scope ownership and manifest entries rather than
  dropping them. Legacy path-only manifests are read-only migration evidence: only entries
  whose bytes match current source or the recorded `source_commit` are adopted; unknown
  differences remain in place and are reported as conflicts.
- Refuses to run at all against a checkout that doesn't look like tricell-hive (missing
  `global/CLAUDE.md`, `global/rules-situational`, or `global/agents`), and refuses an oversized orphan
  deletion (see `--keep-orphans` above) — both are the compensating controls for a
  mis-resolved repo root turning "nothing to deploy" into "delete everything".
- Never deletes outside a managed root. Every manifest entry is contained before it is
  reported or removed: a `..` segment, an absolute path, an unknown prefix, or a target whose
  resolved parent leaves the root for its scope is rejected — named in the report, never
  deleted, dropped from the rewritten manifest — and the run ends with exit code 3 so a
  corrupted manifest gets noticed while the rest of the deploy still lands. A final-component
  symlink is removed as a link; its destination is never followed. Deterministic
  (`test_orphan_containment.py`).

## After a deploy

Remind the user to restart Claude Code / Codex / opencode / Grok / Pi (or start a new
session) to reload — a running session does not pick up the new files. **On a fresh machine
(or any run touching the `codex`/`opencode` scopes for the first time)**, also point them at
`harness/{codex,opencode}/README.md`: the `*.snippet` config merges (plugin/hook
registration in `opencode.jsonc` / `config.toml`) are one-time and manual — this script writes
no part of those files except the opencode `permission` and `shell` keys above, so hooks and rules can land
on disk with nothing registering them until that merge happens. The script's own final report repeats this reminder whenever the
`codex` or `opencode` scope ran.
