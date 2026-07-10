# Agent Configuration Hub

Source of truth for global agent configuration. Claude Code config lives under `global/` and deploys to `~/.claude/` via `/deploy-global`.

Generic AGENTS-compatible harness config lives under `harness/`. Root `AGENTS.md` is only the local guide for working on this repo from harnesses that do not read `CLAUDE.md`.

## Quick Start

```bash
# Inside this repo with Claude Code:
/deploy-global        # Syncs global/ → ~/.claude/
/manage-agents report # Live analysis of all agents
```

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
│   ├── flow-hygiene/                # Workspace hygiene: audit | apply
│   └── flow-report/                 # Self-contained HTML reports for substantial output
└── agents/                          # Optimized agents by role
    ├── design/                      # blue    — cloud-architect, requirement-analyst, system-designer
    ├── development/                 # green   — angular, backend, database, kotlin-multiplatform, nextjs
    ├── review/                      # cyan    — code-reviewer, product-critic, security-reviewer, spec-quality-reviewer, ux-flow-reviewer
    ├── quality/                     # yellow  — performance, prompt, secrets, test, workspace-custodian
    ├── ops/                         # red     — devops-engineer
    └── docs/                        # magenta — technical-writer

harness/                      # Generic AGENTS-compatible config, not deployed by /deploy-global
└── AGENTS.md                        # Condensed cross-harness guidance

.claude/skills/
├── manage-agents/                   # /manage-agents — validate, optimize, report
├── manage-rules/                    # /manage-rules — validate, audit, create
└── deploy-global/                   # /deploy-global — sync global/ → ~/.claude/
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

## Rules (27 files)

| Category     | Files | Scope |
|--------------|------:|-------|
| `quality/`   |     8 | alwaysApply — code principles, security, testing, verifiable test gate, communication format |
| `languages/` |    10 | path-scoped — TypeScript, React/Next.js, Angular, Java/Kotlin, Python, SQL, Tailwind, shell, IaC, NestJS |
| `workflow/`  |     8 | alwaysApply — git, routing, project structure, infra naming, cross-service, gap resolution, memory routing, devops |
| `tools/`     |     1 | alwaysApply — context7 query protocol |

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
| `/manage-agents` | repo | `validate` \| `optimize <name>` \| `report` — agent lifecycle management |
| `/manage-rules` | repo | `validate` \| `audit` \| `create` — rule lifecycle management |
| `/deploy-global` | repo | Sync `global/` to `~/.claude/` (user-initiated only) |
