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
| `codegraph` CLI + MCP | CodeGraph rule (indexed repos only) | `codegraph --version` |
| `uv` | Python rule (pip is banned) | `uv --version` |
| `pnpm` | default package manager | `pnpm --version` |
| `gh` | GitHub operations (PRs, API) | `gh auth status` |
| `cwebp` | evidence retention (WebP lossless) | `cwebp -version` |
| context7 MCP | context7 rule (library docs at write time) | plugin/MCP config per harness |
| chrome-devtools / playwright MCP | Lighthouse/perf; browser fallback | MCP config per harness |

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
│   ├── agents-md-primary/           # Convert projects to AGENTS.md-canonical + CLAUDE.md import
│   ├── engram-init-workspace/       # Unified Engram project for multi-repo workspaces
│   ├── flow-core/                   # Flow pack shared library (non-invocable)
│   ├── flow-intake/                 # F1 — requirements analysis
│   ├── flow-kickoff/                # F2 — workspace bootstrap
│   ├── flow-specs/                  # F3 — specs repo: init | epic | review
│   ├── flow-mock/                   # F4 — prototype: build | review
│   ├── flow-foundation/             # F5 — repos, contracts, CI/CD-first
│   ├── flow-plan/                   # F6 — plan the dev session: research | write
│   ├── flow-build/                  # F6 — execute the plan: reconciler + verify gate
│   ├── flow-deploy/                 # F7 — qa | prod | verify
│   ├── flow-hygiene/                # Workspace hygiene: audit | apply | migrate
│   ├── flow-report/                 # Self-contained HTML reports for substantial output
│   ├── language-rules/              # Router skill: language rules for Codex (references injected by build.py)
│   ├── memory-policy/               # Router skill: Engram policy layer (Codex/opencode)
│   ├── memory-sync/                 # Reconcile Engram + native memory vs ground truth
│   ├── starlight-docs-site/         # Astro Starlight docs: scaffold | page | audit
│   ├── unattended-delegation/       # Router skill: explicitly-delegated unattended runs (Codex/opencode)
│   └── workspace-conventions/       # Router skill: workspace/session/contract conventions (Codex/opencode)
└── agents/                          # Optimized agents by role
    ├── design/                      # blue    — cloud-architect, requirement-analyst, system-designer
    ├── development/                 # green   — angular, backend, database, kotlin-multiplatform, nextjs
    ├── review/                      # cyan    — code-reviewer, product-critic, security-reviewer, spec-quality-reviewer, ux-flow-reviewer
    ├── quality/                     # yellow  — performance, prompt, secrets, test, workspace-custodian
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

## Agents (21 agents)

| Agent | Category | Color | Tool surface |
|-------|----------|-------|--------------|
| `cloud-architect` | design | blue | Read, Write, Edit, Glob, Grep |
| `requirement-analyst` | design | blue | Read, Glob, Grep |
| `system-designer` | design | blue | Read, Write, Edit, Bash, Glob, Grep |
| `angular-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `backend-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `database-specialist` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `kotlin-multiplatform-developer` | development | green | Read, Write, Edit, Bash, Glob, Grep |
| `nextjs-architecture-expert` | development | green | Read, Write, Edit, Bash, Grep, Glob |
| `code-reviewer` | review | cyan | Read, Glob, Grep |
| `product-critic` | review | cyan | Read, Glob, Grep |
| `security-reviewer` | review | cyan | Read, Glob, Grep |
| `spec-quality-reviewer` | review | cyan | Read, Glob, Grep |
| `ux-flow-reviewer` | review | cyan | All except Write/Edit (needs browser MCP via ToolSearch) |
| `in-vivo-qa-tester` | quality | yellow | All except Edit, NotebookEdit (drives a real browser) |
| `performance-engineer` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `prompt-engineer` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `secrets-auditor` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `test-engineer` | quality | yellow | Read, Write, Edit, Bash, Glob, Grep |
| `workspace-custodian` | quality | yellow | Read, Glob, Grep, Bash (read-only audit) |
| `devops-engineer` | ops | red | Inherited toolset except `Agent` (intentional) |
| `technical-writer` | docs | magenta | Read, Write, Edit, Glob, Grep |

Use `/manage-agents report` for live line counts and reduction metrics instead of relying on static README totals.

## Rules (29 files)

| Category     | Files | Scope |
|--------------|------:|-------|
| `quality/`   |     7 | alwaysApply — code principles, security, testing, debugging, critical thinking, communication format |
| `languages/` |    11 | path-scoped — TypeScript, React/Next.js, Angular, Java/Kotlin, Python, SQL, Tailwind, shell, IaC, NestJS, UI visual design |
| `workflow/`  |     9 | alwaysApply — git, routing, project structure, infra naming, cross-service, gap resolution, memory routing, unattended autonomy, devops |
| `tools/`     |     2 | alwaysApply — context7 query protocol, browser automation |

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
| `/agents-md-primary` | global | Convert projects to AGENTS.md-canonical + CLAUDE.md `@AGENTS.md` import; `scan` finds candidates |
| `flow-core` | global | Flow pack shared library: contract + templates (non-invocable) |
| `/flow-intake` … `/flow-deploy` | global | The 7 phase gates of the client project flow (F1–F7) |
| `/flow-hygiene` | global | Workspace hygiene: `audit` \| `apply` |
| `/engram-init-workspace` | global | Unified `.engram/config.json` for multi-repo workspaces |
| `flow-report` | global | Renders substantial output as self-contained HTML (auto-invoked) |
| `/memory-sync` | global | `audit` \| `apply` — reconcile Engram + native memory vs ground truth |
| `/starlight-docs-site` | global | `scaffold` \| `page` \| `audit` — Astro Starlight docs sites |
| `language-rules` | global | Router: full language rules for Codex; quality depth rows for Codex + opencode (references injected by `build.py`) |
| `workspace-conventions` | global | Router: workspace/session/contract conventions for Codex + opencode (model-invoked) |
| `memory-policy` | global | Router: Engram policy layer for Codex + opencode (model-invoked) |
| `unattended-delegation` | global | Router: explicitly-delegated unattended runs for Codex + opencode (model-invoked) |
| `/manage-agents` | repo | `validate` \| `optimize <name>` \| `report` — agent lifecycle management |
| `/manage-rules` | repo | `validate` \| `audit` \| `create` — rule lifecycle management |
| `/deploy-global` | repo | Sync `global/` to `~/.claude/` (user-initiated only) |
