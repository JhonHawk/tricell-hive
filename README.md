# Agent Configuration Hub

Source of truth for global agent configuration. Claude Code config lives under `global/` and deploys to `~/.claude/` via `/deploy-global`. Pi is integrated through the adapter under `harness/pi/` and uses the same canonical agents, skills, and hooks.

Generic AGENTS-compatible harness config lives under `harness/`. **What each harness actually loads — and the official doc backing every loading mechanism, with a verified date — has one README per harness:** `global/README.md` (Claude Code), `harness/codex/README.md`, `harness/opencode/README.md`, `harness/grok/README.md`, `harness/cursor/README.md`, and `harness/pi/README.md`. Root `AGENTS.md` is only the local guide for working on this repo from harnesses that do not read `CLAUDE.md`.

## Quick Start

```bash
# Inside this repo with Claude Code:
/deploy-global        # Syncs global/ → ~/.claude/
/manage-agents validate --all  # Judgment checks on every agent
```

## Prerequisites

The config assumes these are installed; nothing here installs them for you.

### Per harness (one-time)

| Harness | What | How |
|---|---|---|
| Claude Code | Engram plugin (memory protocol) | Engram's own plugin channel (`plugin:engram`) |
| Codex | Engram Codex plugin (`engram@engram`, bundled hooks) | Plugin cache under `~/.codex/plugins/`. **Never run `engram setup` for Codex** — it sets `model_instructions_file`, which replaces Codex's base system prompt (`harness/codex/README.md`) |
| Codex | Config additions (`project_doc_max_bytes = 65536`, `commit_attribution = ""`, subagent limits) | Merge `harness/codex/config.toml.snippet` into `~/.codex/config.toml`, once |
| opencode | flow-skill gating | Merge `harness/opencode/opencode.jsonc.snippet` into `~/.config/opencode/opencode.json`, once |
| Pi | Pi 0.85.1, Node 22.19+, the five exact Hive package pins, and the reviewed `pi-subagents` patch | The default/`all` and `harness` selections include PI; `--only pi` previews the PI + shared-skills selection without writing other harnesses. Install the pins before `--apply`; the helper never auto-installs. |

Merge/verify commands and the reasoning live in `harness/README.md` (snippets are the one
layer `/deploy-global` cannot automate). After any merge, start a fresh session.

### Deployment selections

`/deploy-global` defaults to all five harnesses: Claude Code, Codex, opencode, Grok, and
Pi. `--only all` and `--only harness` use the same scope sets, and comma-separated
selectors may be mixed. Every selected root is preflighted before the first target write.

`--only pi` is the isolated write boundary: it preflights Pi and the neutral shared-skills
surface, writes only the Pi root and shared skills, and leaves Claude, Codex, opencode, and
Grok installations untouched. The shared deploy engine records source-commit and content
hash provenance; a legacy path-only manifest is migration evidence only. Matching bytes (or
matching bytes at its recorded source commit) can be adopted, while unknown differences are
preserved and reported as conflicts.

### CLIs and MCPs the rules assume present

| Tool | Used by | Check |
|---|---|---|
| `agent-browser` CLI (+ `AGENT_BROWSER_PROFILE=Tricell` in `~/.zshenv`) | browser-automation rule, in-vivo verification, `sdd-verify` | `agent-browser --version` |
| `uv` | Python rule (pip is banned) | `uv --version` |
| `pnpm` | default package manager | `pnpm --version` |
| `gh` | GitHub operations (PRs, API) | `gh auth status` |
| `cwebp` | evidence retention (WebP lossless) | `cwebp -version` |
| context7 | context7 rule (library docs at write time) | plugin/MCP config per harness; Pi uses `https://mcp.context7.com/mcp` |
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
├── CLAUDE.md                        # GENERATED always-on core (assembled from core-sections/ by harness/build.py)
├── core-sections/                   # Canonical section files for both always-on cores (global/CLAUDE.md + harness/AGENTS.md)
├── rules-situational/               # THE rule-text store (42 texts): 7 inlined into the core by a rule-* section, 26 held by the rule-delivery hook on a matching write or command, 8 router-only, 1 pack-only. Never deployed to ~/.claude/rules/ — see its README for the procedure
├── skills/                          # Global skills (deployed to ~/.claude/skills/)
│   ├── agents-md-primary/           # Convert projects to AGENTS.md-canonical + CLAUDE.md import; audit|apply dedups vs deployed canon + content quality
│   ├── engram-init-workspace/       # Unified Engram project for multi-repo workspaces
│   ├── flow-core/                   # Process library (non-invocable): contract, templates, playbooks
│   │                                #   (bootstrap, spec-writing, migration, workspace-hygiene, audit, promotion)
│   ├── flow-plan/                   # Plan and authorize a portable work contract
│   ├── flow-build/                  # Execute an authorized plan: reconciler + verify gate
│   ├── flow-report/                 # Self-contained HTML reports for artifacts that outlive the thread (6 archetypes; `paper` is the printable one)
│   ├── language-rules/              # Router skill: language rules for every harness but Claude Code; browser CLI reference for all (references injected by build.py)
│   ├── memory-policy/               # Router skill: Engram policy layer (every harness)
│   ├── memory-sync/                 # Reconcile Engram + native memory vs ground truth
│   ├── starlight-docs-site/         # Astro Starlight docs: scaffold | page | audit
│   ├── unattended-delegation/       # Router skill: explicitly-delegated unattended runs (every harness)
│   └── workspace-conventions/       # Router skill: workspace/session/contract conventions (every harness)
├── agents/                          # Optimized agents by role
│   ├── design/                      # blue    — cloud-architect, sdd-design, solution-architect, visual-designer
│   ├── development/                 # green   — angular, backend, database, kotlin-multiplatform, react
│   ├── review/                      # cyan    — review-code, sdd-explore, review-refuter, sdd-product-critic, review-security, sdd-spec-reviewer, review-ux
│   ├── quality/                     # yellow  — performance, prompt, secrets, state-fetcher, test, workspace-custodian
│   ├── ops/                         # red     — devops-engineer
│   └── docs/                        # magenta — sdd-spec-writer
└── README.md                        # What Claude Code loads + the official doc backing each mechanism

harness/                      # Multi-harness layer (Codex + opencode + Grok + Pi), deployed by /deploy-global with selected scopes
├── README.md                        # The layer as a whole + the manual-merge snippets
├── AGENTS.md                        # Always-on cross-harness core (depth behind router skills)
├── build.py                         # Regenerates generated trees + injects router-skill references
├── agents-skills/                   # GENERATED — universal skills → ~/.agents/skills
├── codex/                           # README + config.toml.snippet + GENERATED TOML agents
├── opencode/                        # README + opencode.jsonc.snippet + commands/ + GENERATED agents
├── grok/                            # README + GENERATED agents (rules reach Grok as flat symlinks)
├── pi/                              # Pi runtime, package manifest, tests, and Pi README
└── cursor/                          # README only — no generated tree, no deploy step

.claude/skills/
├── manage-agents/                   # /manage-agents — validate (agents + rule texts)
└── deploy-global/                   # /deploy-global — sync global/ + selected harness targets

_support/                            # Workspace material, not deployed
├── archive/                         # Dated historical snapshots (audits/, docs/)
├── archived-agents/                 # Retired agent versions kept for reference
├── docs/                            # Durable hub docs (methodology-bibliography, flow-pack-manual)
├── spec/                            # Design specs and decision records
└── workspace/                       # Ephemeral scratch (gitignored)
```

## Agents (26 agents)

| Agent | Category | Color | Tool surface |
|-------|----------|-------|--------------|
| `cloud-architect` | design | blue | Read, Write, Edit, Glob, Grep, WebSearch/WebFetch, context7 |
| `sdd-design` | design | blue | Read, Write, Edit, Glob, Grep |
| `solution-architect` | design | blue | Read, Write, Edit, Glob, Grep, WebSearch/WebFetch, context7 |
| `visual-designer` | design | blue | Read, Write, Edit, Bash, Glob, Grep, context7, heroui-pro |
| `angular-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep, context7 |
| `backend-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `database-specialist` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `kotlin-multiplatform-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `react-developer` | development | green | Read, Write, Edit, Bash, Grep, Glob, context7, heroui-pro |
| `ts-backend-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `sdd-explore` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only discovery & research) |
| `review-code` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only investigation; guard coverage varies by harness) |
| `review-refuter` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (executes claims, never modifies) |
| `sdd-product-critic` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only investigation; guard coverage varies by harness) |
| `review-security` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only investigation; guard coverage varies by harness) |
| `sdd-spec-reviewer` | review | cyan | Read, Glob, Grep, Bash, WebSearch/WebFetch, context7 (read-only investigation; guard coverage varies by harness) |
| `review-ux` | review | cyan | All except Write, Edit, NotebookEdit, Agent (needs browser MCP via ToolSearch) |
| `sdd-verify` | quality | yellow | All except Edit, NotebookEdit, Agent (drives a real browser; Write for the in-vivo report) |
| `performance-engineer` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `prompt-engineer` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `secrets-auditor` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `state-fetcher` | quality | yellow | All except Write, Edit, NotebookEdit, Agent (tracker MCP/CLI + gh + deploy CLIs; never mutates files) |
| `test-engineer` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `workspace-custodian` | quality | yellow | Read, Glob, Grep, Bash (read-only audit) |
| `devops-engineer` | ops | red | Read, Write, Edit, Bash, Glob, Grep (explicit allowlist; `Agent` denied) |
| `sdd-spec-writer` | docs | magenta | Read, Write, Edit, Glob, Grep |

Line counts live on disk (`wc -l global/agents/*/*.md`); `/manage-agents validate --all` checks this table against it.

## Rules (42 texts, one store)

`global/rules-situational/` holds every rule text; the channel decides how each reaches a model, and a rule with no channel is a build error:

| Delivery | Texts | How it reaches an agent |
|---|---:|---|
| Core include | 7 | A `global/core-sections/rule-<name>.md` section carries `include:`; `harness/build.py` inlines the body into `global/CLAUDE.md`, always-on in every session. Grok reads that same core natively; Codex and opencode read the condensed `harness/AGENTS.md` |
| Hook trigger | 26 | `globs:` (a matching write) and/or `commands:` (a matching command prefix) — the `rule-delivery` hook denies the call and names the reference to read. Claude Code, Grok, Codex and PI; a trigger counts only together with the router reference the hold names |
| Router only | 8 | Injected into a skill's `references/` (`SKILL_REFERENCE_INJECTIONS` in `harness/build.py`), present when the model invokes the router |
| Pack only | 1 | `agent-core-gates.md` — no trigger, no injection; inlined into every packed agent, which runs with `omitClaudeMd: true` |

The packed agents also carry their triggered rules inlined, so a pack is a build-time guarantee where the hook is a probability. Counts come from `harness/rule-manifest.json` (`python3 -c "import json;m=json.load(open('harness/rule-manifest.json'));print(len(m['rules']))"`), never from subtraction. The procedure for adding one — which shape, what to wire, what to check before finishing — is `global/rules-situational/README.md`.

**`paths:` is a build error.** It was Claude Code's native path-scoping key (*"rules without a `paths` field are loaded unconditionally"*); this repo retired that channel because it fires on a READ, misses the creation of the first file of a kind, charges every read-only agent that opens a `.ts`, and does not survive compaction.

Always-on footprint, now that the gates are inlined into the core: **405 lines / 86,768 B / ~21.7k tokens** — down from 37.5 KB of core plus 83 KB of always-on rule files. Paid on every session before any work starts; measure it with:

```sh
wc -lc global/CLAUDE.md
```

## Agent Design Criteria

1. **Every line changes the output** — past 120 lines of its own text, review for filler (a threshold, not a cap; rule texts carried through `packs:` do not count)
2. **Unique rules only** — inherited from global CLAUDE.md are not repeated
3. **Concrete, not generic** — every rule must change the agent's output
4. **Restricted tools** — only what the agent actually needs; audit-only quality agents may be read-only + Bash
5. **`model: inherit`** — override only with strong reason
6. **Color by role** — blue=design, cyan=review, green=build, yellow=validate, magenta=creative, red=critical
7. **Effective description** — triggers correct auto-invocation routing

## Skills

| Skill | Scope | Purpose |
|-------|-------|---------|
| `/adversarial-research` | global | N independent generators (one may be Codex) + review-refuter cross-exam → refuted/weakened/surviving/net-new canon |
| `/agents-md-primary` | global | Convert projects to AGENTS.md-canonical + CLAUDE.md `@AGENTS.md` import; `scan` finds candidates; `audit \| apply` dedups project rules against the deployed canon (harness-coverage matrix) + completeness checks (project pointers, Git Workflow declarations) + content quality (agent-discoverable rules, stale paths, instruction budget). Hub-destined proposals (promote-to-core, injection map) are reported, never executed by the skill |
| `flow-core` | global | Process library (non-invocable): flow contract, templates, and the playbooks — bootstrap, spec-writing, migration, workspace-hygiene, audit, promotion |
| `/flow-plan` | global | Prepare a portable plan with explicit approval, scoped authorization and separate execution evidence; also available through an explicit conversational planning request |
| `/flow-build` | global | Execute or resume an authorized plan; reconcile working-tree/Git evidence, verification and pending authorized delivery. `verify` never publishes |
| `/engram-init-workspace` | global | Unified `.engram/config.json` for multi-repo workspaces |
| `flow-report` | global | Renders substantial output as self-contained HTML in six archetypes — document, paper, explainer, review, comparison, deck (auto-invoked) |
| `/memory-sync` | global | `audit` \| `apply` — reconcile Engram + native memory vs ground truth |
| `/monorepo-cutover` | global | `intake` \| `execute` \| `verify` — migrate a multi-repo product to a monorepo (or bootstrap one greenfield); stack fork at intake (JS/TS lane fully specified; polyglot → orchestrator decision), candidate tracking via Engram upsert |
| `/workspace-archive` | global | `run` \| `normalize` — move `verified` sessions older than 15 days to `sessions/archived/` (git mv + index links); on request only, deterministic script + one confirmation |
| `/status-fetch` | global | Live external state (git, declared tracker, PRs, deploys) as compact facts — runs in a forked `state-fetcher` so the sweep stays out of the session |
| `/starlight-docs-site` | global | `scaffold` \| `page` \| `audit` — Astro Starlight docs sites |
| `language-rules` | global | Router: full language rules for Codex, Grok, opencode and PI; browser CLI reference for every harness (references injected by `build.py`) |
| `workspace-conventions` | global | Router: workspace/session/contract conventions, every harness (model-invoked) |
| `memory-policy` | global | Router: Engram policy layer, every harness (model-invoked) |
| `unattended-delegation` | global | Router: explicitly-delegated unattended runs and ledger-declared standing jobs, every harness (model-invoked) |
| `/manage-agents` | repo | `validate [--all] [--deep]` — judgment checks on agent definitions and on rule texts / core sections (trigger honesty, glob breadth, hold cost, pack coverage, enforcement honesty) |
| `/deploy-global` | repo | Sync `global/` to `~/.claude/` (user-initiated only) |
