# Grok CLI adapter — what it loads and where that is documented

The thinnest adapter of the three. Grok reads Claude Code's own files through its compat
layer, so most of the pack reaches it with **no** generated tree: `~/.claude/CLAUDE.md`,
`~/.claude/skills/`, and `~/.claude/settings.json` are consumed directly. Only two things
need a Grok-specific shape — the rules (its scan cannot use `paths:`) and the agents (its
frontmatter differs).

Companion files: `global/README.md` (Claude Code), `harness/codex/README.md`,
`harness/opencode/README.md`. All four share the section layout below.

## Pieces

| Piece | Source | Deploy target | Maintained how |
|---|---|---|---|
| Always-on rules | `~/.claude/rules/**` (already deployed) | `~/.grok/rules/<dir>__<file>.md` | `/deploy-global --only grok` — one **flat file symlink** per rule |
| Subagents | `harness/grok/agents/` (generated, versioned) | `~/.grok/agents/` | **Generated** by `harness/build.py` from `global/agents/` — never edit |
| Skills | `~/.claude/skills/` | *(read in place)* | Nothing to deploy — `[compat.claude] skills = true` |
| Global instructions | `~/.claude/CLAUDE.md` | *(read in place)* | Nothing to deploy |
| Hooks | `~/.claude/hooks/` + `~/.claude/settings.json` | *(read in place)* | Nothing to deploy; `bash-policy.sh` and `post-tool-hub.sh` are dual-runtime |

There is deliberately **no** `harness/grok/skills/` tree: it would duplicate what
`~/.claude/skills/` already provides. Contrast with Codex, which needs a generated
`openai.yaml` to carry the same gate.

## What this harness loads

Every URL below was fetched and its quote extracted from the page body on the date in the
last column.

| Layer | What | Where it lands | Mechanism | Official doc | Verified |
|---|---|---|---|---|---|
| Always-on core | `global/CLAUDE.md` | `~/.claude/CLAUDE.md` (read in place) | Claude-compat instruction-file discovery | [docs.x.ai/build/features/skills-plugins-marketplaces](https://docs.x.ai/build/features/skills-plugins-marketplaces) — *"Grok automatically reads Claude Code marketplaces, plugins, skills, MCPs, agents, hooks, and instruction files"* | 2026-08-20 |
| Always-on rules | 12 flat symlinks `<dir>__<file>.md` | `~/.grok/rules/` | Every `*.md` in a rules dir is loaded regardless of name | [docs.x.ai/build/features/project-rules](https://docs.x.ai/build/features/project-rules) — *"every `*.md` file in a `.grok/rules/` directory"*; `.claude/rules/` also read for compat | 2026-08-20 |
| Path-scoped rules | **none** | — | No file-pattern scoping key exists — see *What does NOT reach it* | [docs.x.ai/build/features/project-rules](https://docs.x.ai/build/features/project-rules) (absence) | 2026-08-20 |
| Skills | 16 skills, incl. the router `references/` | `~/.claude/skills/` (read in place) | Claude-compat scan at the **lowest** precedence; `disable-model-invocation` honored natively | [docs.x.ai/build/features/skills-plugins-marketplaces](https://docs.x.ai/build/features/skills-plugins-marketplaces) — *"`disable-model-invocation`: Slash command only; no automatic invoke. Default `false`."* | 2026-08-20 |
| Agents | 25 Grok-shaped `.md` (real files, not symlinks) | `~/.grok/agents/` | `.md` agent definitions in `~/.grok/agents/` | [docs.x.ai/build/features/subagents](https://docs.x.ai/build/features/subagents) | 2026-08-20 |
| Hooks | shared with Claude Code | `~/.claude/settings.json` (merged by compat) | Grok hooks + Claude settings compat | [docs.x.ai/build/features/hooks](https://docs.x.ai/build/features/hooks) | 2026-08-20 |
| Scope & precedence table | — | — | Which dirs are scanned, in what order | bundled doc `~/.grok/docs/user-guide/12-project-rules.md` + `08-skills.md` (grok 1.0.5) — local file, not a URL | 2026-08-20 |

**Dead URLs — never write these** (all verified 404 on 2026-08-20):
`github.com/xai-org/grok-cli` · `docs.x.ai/docs/grok-cli` ·
`docs.x.ai/build/project-rules` (the real path carries `/features/`) ·
`github.com/xai-org/grok-build/blob/main/docs/user-guide/...` — that repo is public but is
the Rust harness and has **no `docs/` directory**; the user guide exists only on disk.

## What does NOT reach it

- **The 19 path-scoped rules.** Grok has no `paths:`-equivalent, so shipping them into
  `~/.grok/rules/` would load all of them, always. They are excluded on purpose and reach
  Grok the same way they reach Codex: injected into a router skill's `references/`
  (`language-rules`, `workspace-conventions`, `task-routing`, …) plus the advisory
  `rule-context` hook. **Consequence:** giving an existing always-on rule a `paths:` key
  silently removes it from Grok — route that content through a router skill in the same
  change.
- **Directory symlinks.** `~/.grok/rules/` is populated with one symlink **per file**, not
  a single link to `~/.claude/rules/`. See the recursion caveat below.
- **Nothing yet deployed to `~/.claude/rules/`.** The grok scope skips a missing rule with
  a WARNING rather than leaving a dangling link — run `--only claude` at least once first.

## Unverified / undocumented dependencies

Two claims this adapter is built on that **no official source states positively**. Both are
recorded here rather than dressed up as citations:

- **"The rules scan is not recursive."** This is the entire reason for the flat
  `<dir>__<file>.md` naming. Neither the public docs nor the bundled user guide say
  whether the scan descends into subdirectories. Compare Claude Code, whose docs are
  explicit that `.md` files are discovered recursively. **Status: assumed, unconfirmed.**
  Settle it empirically — drop a `.md` into a subdirectory of `~/.grok/rules/`, run
  `grok inspect`, and see whether it is listed.
- **"Grok ignores `paths:`."** Supported as *documentary absence*: no source mentions any
  file-pattern scoping key at all (`paths:`, `globs:`, or otherwise); the only documented
  scoping is per-directory. Reasonable, but weaker than the flat assertion sounds.

## How to re-verify

```bash
grok inspect          # lists every config source, rules file, skill, plugin, hook and MCP
                      # server Grok discovered here — the deterministic auditor,
                      # analogous to `codex debug prompt-input`

ls -la ~/.grok/rules/ # should be 12 symlinks named <dir>__<file>.md → ~/.claude/rules/...
ls    ~/.grok/agents/ # should be 25 real .md files
```

The bundled user guide under `~/.grok/docs/user-guide/` ships with the CLI and is version-
stamped: check `grok --version` before trusting a quote from it.
