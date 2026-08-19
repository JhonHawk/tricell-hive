# Agent Configuration Hub

Source of truth for global agent configuration. Claude Code config lives under `global/` and deploys to `~/.claude/` via `/deploy-global`.

Generic AGENTS-compatible harness config lives under `harness/`. Root `AGENTS.md` is only the local guide for working on this repo from harnesses that do not read `CLAUDE.md`.

## Quick Start

```bash
# Inside this repo with Claude Code:
/deploy-global        # Syncs global/ → ~/.claude/
/manage-agents report # Live analysis of all agents
```

## Prerequisites

The config assumes these are installed; nothing here installs them for you.

### Per harness (one-time)

| Harness | What | How |
|---|---|---|
| Claude Code | Engram plugin (memory protocol) | Engram's own plugin channel (`plugin:engram`) |
| Codex | Engram Codex plugin (`engram@engram`, bundled hooks) | Plugin cache under `~/.codex/plugins/`. **Never run `engram setup` for Codex** — it sets `model_instructions_file`, which replaces Codex's base system prompt (`harness/codex/README.md`) |
| Codex | Config additions (`project_doc_max_bytes = 65536`, `commit_attribution = ""`, subagent limits) | Merge `harness/codex/config.toml.snippet` into `~/.codex/config.toml`, once |
| opencode | `opencode-rules@0.6.4` plugin (glob-conditional language rules; pinned, audited) + flow-skill gating | Merge `harness/opencode/opencode.jsonc.snippet` into `~/.config/opencode/opencode.json`, once |
| opencode | Single skill-discovery source | `export OPENCODE_DISABLE_CLAUDE_CODE_SKILLS=1` in the shell profile |

Merge/verify commands and the reasoning live in `harness/README.md` (snippets are the one
layer `/deploy-global` cannot automate). After any merge, start a fresh session.

### CLIs and MCPs the rules assume present

| Tool | Used by | Check |
|---|---|---|
| `agent-browser` CLI (+ `AGENT_BROWSER_PROFILE=Tricell` in `~/.zshenv`) | browser-automation rule, in-vivo verification, `in-vivo-qa-tester` | `agent-browser --version` |
| `uv` | Python rule (pip is banned) | `uv --version` |
| `pnpm` | default package manager | `pnpm --version` |
| `gh` | GitHub operations (PRs, API) | `gh auth status` |
| `cwebp` | evidence retention (WebP lossless) | `cwebp -version` |
| context7 MCP | context7 rule (library docs at write time) | plugin/MCP config per harness |
| chrome-devtools / playwright MCP | Lighthouse/perf; browser fallback | MCP config per harness |

### Terminal workspace: herdr (recommended)

[herdr](https://herdr.dev) is the recommended terminal multiplexer for running agent
sessions (Claude Code / Codex / opencode) side by side — workspaces per project,
worktree-per-agent panes. Not required by any rule; it is the daily driver this config
is operated from.

```bash
curl -fsSL https://herdr.dev/install.sh | sh    # or: brew install herdr
```

Adopted plugins (install with `herdr plugin install <owner/repo> --yes`):

| Plugin | Why | Key |
|---|---|---|
| `persiyanov/herdr-reviewr` | Review an agent's diff beside the chat, line comments sent back to its input — the review-before-push loop as a pane. Config sets `base_branches = ["development", "main", "master"]` to match environment-branch repos | `cmd+r` toggle |
| `smarzban/herdr-file-viewer` | Git-aware read-only tree + content browsing with diffs and rendered markdown | `prefix+f` split, `prefix+shift+f` tab |
| `Numbered-com/herdr-ports` | `$ports` badge on every space with active TCP listeners — permanent dev-server visibility, the visual complement of the `session-hygiene-report` hook | — |

herdr config is machine-local, NOT deployed by this repo: keybindings live in
`~/.config/herdr/config.toml`; per-plugin config in `herdr plugin config-dir <id>`.
Candidates evaluated for a second wave (worktree bootstrap, notifications, workspace
templates, Linear/browser panes): see the 2026-07-23 plugin review in Engram.

## Structure

```
AGENTS.md                            # Local guide for AGENTS-compatible harnesses working on this repo
CLAUDE.md                            # Claude-facing project guide, not deployed
global/                              # Mirrors ~/.claude/ — deployable source of truth
├── CLAUDE.md                        # Core config (always loaded)
├── rules/                           # Path-scoped and alwaysApply rules
│   ├── quality/                     # Code principles (8 files)
│   ├── languages/                   # Language/framework standards (11 files, path-scoped)
│   ├── workflow/                    # Git, deploys, structure, routing, naming (8 files)
│   └── tools/                       # External tools & MCP protocols (1 file)
├── skills/                          # Global skills (deployed to ~/.claude/skills/)
│   ├── agents-md-primary/           # Convert projects to AGENTS.md-canonical + CLAUDE.md import; audit|apply dedups vs deployed canon + content quality
│   ├── engram-init-workspace/       # Unified Engram project for multi-repo workspaces
│   ├── flow-core/                   # Process library (non-invocable): contract, templates, playbooks
│   │                                #   (bootstrap, spec-writing, migration, workspace-hygiene, audit, promotion)
│   ├── flow-build/                  # Execute a captured plan: reconciler + verify gate
│   ├── flow-report/                 # Self-contained HTML reports for artifacts that outlive the thread
│   ├── language-rules/              # Router skill: language rules for Codex (references injected by build.py)
│   ├── memory-policy/               # Router skill: Engram policy layer (Codex/opencode)
│   ├── memory-sync/                 # Reconcile Engram + native memory vs ground truth
│   ├── starlight-docs-site/         # Astro Starlight docs: scaffold | page | audit
│   ├── unattended-delegation/       # Router skill: explicitly-delegated unattended runs (Codex/opencode)
│   └── workspace-conventions/       # Router skill: workspace/session/contract conventions (Codex/opencode)
└── agents/                          # Optimized agents by role
    ├── design/                      # blue    — cloud-architect, requirement-analyst, system-designer, visual-designer
    ├── development/                 # green   — angular, backend, database, kotlin-multiplatform, react
    ├── review/                      # cyan    — code-reviewer, code-scout, finding-refuter, product-critic, security-reviewer, spec-quality-reviewer, ui-reviewer
    ├── quality/                     # yellow  — performance, prompt, secrets, state-fetcher, test, workspace-custodian
    ├── ops/                         # red     — devops-engineer
    └── docs/                        # magenta — technical-writer

harness/                      # Multi-harness layer (Codex + opencode), deployed by /deploy-global step 13b
├── AGENTS.md                        # Always-on cross-harness core (~18 KiB; depth behind router skills)
├── build.py                         # Regenerates generated trees + injects router-skill references
├── agents-skills/                   # GENERATED — universal skills → ~/.agents/skills
├── codex/                           # config.toml.snippet + GENERATED TOML agents
└── opencode/                        # opencode.jsonc.snippet + commands/ + GENERATED agents & rules

.claude/skills/
├── manage-agents/                   # /manage-agents — validate, optimize, report
├── manage-rules/                    # /manage-rules — validate, audit, create
└── deploy-global/                   # /deploy-global — sync global/ → ~/.claude/

_support/                            # Workspace material, not deployed
├── archive/                         # Dated historical snapshots (audits/, docs/)
├── archived-agents/                 # Retired agent versions kept for reference
├── docs/                            # Durable hub docs (methodology-bibliography, flow-pack-manual)
├── spec/                            # Design specs and decision records
└── workspace/                       # Ephemeral scratch (gitignored)
```

## Agents (25 agents)

| Agent | Category | Color | Tool surface |
|-------|----------|-------|--------------|
| `cloud-architect` | design | blue | Read, Write, Edit, Glob, Grep |
| `requirement-analyst` | design | blue | Read, Glob, Grep |
| `system-designer` | design | blue | Read, Write, Edit, Glob, Grep |
| `visual-designer` | design | blue | Read, Write, Edit, Bash, Glob, Grep |
| `angular-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `backend-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `database-specialist` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `kotlin-multiplatform-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `react-developer` | development | green | Read, Write, Edit, Bash, Grep, Glob |
| `code-scout` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only discovery) |
| `code-reviewer` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only investigation under plan mode) |
| `finding-refuter` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (executes claims, never modifies) |
| `product-critic` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only investigation under plan mode) |
| `security-reviewer` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only investigation under plan mode) |
| `spec-quality-reviewer` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only investigation under plan mode) |
| `ui-reviewer` | review | cyan | All except Write, Edit, NotebookEdit, Agent (needs browser MCP via ToolSearch) |
| `in-vivo-qa-tester` | quality | yellow | All except Edit, NotebookEdit, Agent (drives a real browser; Write for the in-vivo report) |
| `performance-engineer` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `prompt-engineer` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `secrets-auditor` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `state-fetcher` | quality | yellow | All except Write, Edit, NotebookEdit, Agent (tracker MCP/CLI + gh + deploy CLIs; never mutates files) |
| `test-engineer` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `workspace-custodian` | quality | yellow | Read, Glob, Grep, Bash (read-only audit) |
| `devops-engineer` | ops | red | Read, Write, Edit, Bash, Glob, Grep (explicit allowlist; `Agent` denied) |
| `technical-writer` | docs | magenta | Read, Write, Edit, Glob, Grep |

Use `/manage-agents report` for live line counts and reduction metrics instead of relying on static README totals.

## Rules (32 files)

**`paths:` is the only frontmatter key Claude Code reads.** Per the official docs, *"rules without a `paths` field are loaded unconditionally"* — so there are two states, not three:

- **`paths: [...]`** — loads only when a matching file is touched. The only way to keep a rule out of the always-on set.
- **No `paths:`** — always-on. We write `alwaysApply: true` to state the intent, but it is documentation, not mechanism: the file loads identically without it. An unrecognized key does **not** suppress loading.

| Category     | Files | always-on | path-scoped |
|--------------|------:|----------:|------------:|
| `quality/`   |     7 |         6 |           1 |
| `languages/` |    11 |         0 |          11 |
| `workflow/`  |    11 |         9 |           2 |
| `tools/`     |     3 |         3 |           0 |

Always-on footprint (`global/CLAUDE.md` + the 18 rules without `paths:`): **913 lines / 121 KB / ~30k tokens**, paid on every session before any work starts. Measure it with:

```sh
cd global/rules && for f in $(find . -name '*.md'); do grep -q '^paths:' "$f" || cat "$f"; done | wc -c
```

## Agent Design Criteria

1. **Under 120 lines** — if longer, it probably has filler
2. **Unique rules only** — inherited from global CLAUDE.md are not repeated
3. **Concrete, not generic** — every rule must change the agent's output
4. **Restricted tools** — only what the agent actually needs; audit-only quality agents may be read-only + Bash
5. **`model: inherit`** — override only with strong reason
6. **Color by role** — blue=design, cyan=review, green=build, yellow=validate, magenta=creative, red=critical
7. **Effective description** — triggers correct auto-invocation routing

## Skills

| Skill | Scope | Purpose |
|-------|-------|---------|
| `/adversarial-research` | global | N independent generators (one may be Codex) + finding-refuter cross-exam → refuted/weakened/surviving/net-new canon |
| `/agents-md-primary` | global | Convert projects to AGENTS.md-canonical + CLAUDE.md `@AGENTS.md` import; `scan` finds candidates; `audit \| apply` dedups project rules against the deployed canon (harness-coverage matrix) + completeness checks (project pointers, Git Workflow declarations) + content quality (agent-discoverable rules, stale paths, instruction budget). Hub-destined proposals (promote-to-core, injection map) are reported, never executed by the skill |
| `flow-core` | global | Process library (non-invocable): flow contract, templates, and the playbooks — bootstrap, spec-writing, migration, workspace-hygiene, audit, promotion |
| `/flow-build` | global | Execute or resume a captured plan: state-driven reconciler + verify gate. The one remaining process command — everything else is a playbook |
| `/engram-init-workspace` | global | Unified `.engram/config.json` for multi-repo workspaces |
| `flow-report` | global | Renders substantial output as self-contained HTML in five archetypes — document, explainer, review, comparison, deck (auto-invoked) |
| `/memory-sync` | global | `audit` \| `apply` — reconcile Engram + native memory vs ground truth |
| `/monorepo-cutover` | global | `intake` \| `execute` \| `verify` — migrate a multi-repo product to a monorepo (or bootstrap one greenfield); stack fork at intake (JS/TS lane fully specified; polyglot → orchestrator decision), candidate tracking via Engram upsert |
| `/status-fetch` | global | Live external state (git, declared tracker, PRs, deploys) as compact facts — runs in a forked `state-fetcher` so the sweep stays out of the session |
| `/starlight-docs-site` | global | `scaffold` \| `page` \| `audit` — Astro Starlight docs sites |
| `language-rules` | global | Router: full language rules for Codex; quality depth rows for Codex + opencode (references injected by `build.py`) |
| `workspace-conventions` | global | Router: workspace/session/contract conventions for Codex + opencode (model-invoked) |
| `memory-policy` | global | Router: Engram policy layer for Codex + opencode (model-invoked) |
| `unattended-delegation` | global | Router: explicitly-delegated unattended runs for Codex + opencode (model-invoked) |
| `/manage-agents` | repo | `validate` \| `optimize <name>` \| `report` — agent lifecycle management |
| `/manage-rules` | repo | `validate` \| `audit` \| `create` — rule lifecycle management, incl. `global/CLAUDE.md` and the always-on rule corpus |
| `/deploy-global` | repo | Sync `global/` to `~/.claude/` (user-initiated only) |
