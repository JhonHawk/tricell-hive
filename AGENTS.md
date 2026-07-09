# Agent Configuration Hub

## Purpose
This is the canonical guide for all harnesses. `CLAUDE.md` imports it via `@AGENTS.md` and adds Claude Code-specific content below the import.

This repo is the **source of truth** for the user's global agent configuration. Claude Code configuration lives under `global/`: global CLAUDE.md, path-scoped rules, and optimized subagent definitions. Everything under `global/` deploys to `~/.claude/` via the `/deploy-global` skill. **Deploy is always user-initiated** — never execute `/deploy-global` or copy files to `~/.claude/` without explicit user instruction.

The reusable generic harness configuration lives in `harness/AGENTS.md`. Edit that file when changing cross-project AGENTS guidance. Do not duplicate its full contents here.

## Context
The user is a software architect and developer working across 6 client groups with ~50 repositories spanning Next.js, Angular, NestJS, Express, Java/Spring, Kotlin, Python, and DevOps (GitHub Actions, AWS, Hetzner, Vercel, Dokploy). Agents must not repeat global rules; they inherit them automatically.

## Repository Boundaries
- `global/` is the deployable Claude Code source of truth for `~/.claude/`.
- `harness/` is the source of truth for generic AGENTS-compatible harness configuration.
- Root `CLAUDE.md` imports this file via `@AGENTS.md` and adds Claude Code-specific content.
- Root `AGENTS.md` is the canonical guide for all harnesses.

## Critical Constraints
- **Never add new agents without user approval.** Only optimize existing ones unless explicitly asked.
- **Never modify `~/.claude/` files autonomously.** All deploys are user-initiated via `/deploy-global`.
- **Agent prompts in English.** User-facing strings follow the project's language.

## Rules
- Never run `/deploy-global` or copy files to `~/.claude/` unless the user explicitly asks for deployment.
- Do not treat `harness/AGENTS.md` as part of `/deploy-global`; it belongs to the generic harness track.
- `harness/agents-skills/`, `harness/codex/agents/`, and `harness/opencode/agents/` are GENERATED trees. Never hand-edit them; edit the canonical source under `global/` and run `python3 harness/build.py`, committing the regenerated output.
- Do not alias or symlink `AGENTS.md` to `harness/AGENTS.md`; they serve different scopes.
- When changing routing or domains, check `global/rules/workflow/agent-routing.md`.
- When a rule or convention is grounded in external authority (standards, canonical books, official docs), record the source in `_support/docs/methodology-bibliography.md` and consult it before re-researching.
- Use conventional commit prefixes if asked to commit.

## Agent Design Principles

### What Makes a Good Agent
- **Under 120 lines.** Every line must change the agent's output vs default behavior.
- **Unique rules only.** If the global CLAUDE.md already covers it (e.g., "no `any`", "thin controllers"), don't repeat it.
- **Concrete, not generic.** "Use `class-validator` for DTOs" is good. "Follow best practices" is filler.
- **Description controls routing.** The `description` field must be specific and action-oriented — not a resume.
- **Restricted tools.** Only include tools the agent needs. Review agents (cyan) are read-only: Read, Glob, Grep. Quality agents (yellow) may be remediation-oriented (Write/Edit) or audit-oriented (read-only plus Bash when they orchestrate external analysis).
- **`model: inherit` by default.** The agent uses the session's active model. Only override if there's a strong reason (e.g., `haiku` for a read-only explorer).
- **Calibrate to the floor model, not the ceiling.** Rules and agents must work on the least capable model the user runs day-to-day (as of jun-2026: Opus 4.8 once the Fable 5 preview ends on Jun 22). Before cutting a rule as "the model does this by default", verify the *floor* model does it — preview-model capability is not a pruning criterion.
- **Path-scoped rules only load when matching files are touched; do not duplicate them into agents.**

### Naming Rules
- 3-50 characters, lowercase letters, numbers, and hyphens only.
- Must start and end with alphanumeric. No underscores, spaces, or special characters.
- `code-reviewer` ✅ — `my_agent` ❌ — `ag` ❌ (too short) — `-agent-` ❌ (starts/ends with hyphen)

### Description Best Practices
- 1-3 sentences for agents with obvious routing (e.g., "backend" tasks go to `backend-developer`).
- **Triggers, not workflow.** A description states WHEN to invoke — it never compresses the body's procedure into steps. The body loads only on invocation; a description that summarizes the workflow gets followed instead of the body. Enumerating subcommands and what each delivers is fine; enumerating the execution sequence is not. Applies to agents and skills equally.
- For agents where routing is ambiguous, include `<example>` blocks with `<commentary>` to help Claude decide when to invoke:
```
description: >
  Use this agent when evaluating architecture decisions...

  <example>
  Context: Team is migrating from monolith to microservices.
  user: "Review our proposed service boundaries"
  assistant: "I'll evaluate the architecture..."
  <commentary>
  Invoke for macro-level design, not code-level review.
  </commentary>
  </example>
```

### Color Assignment by Role
- **blue** — design, analysis (system-designer)
- **cyan** — review, research (code-reviewer, security-reviewer)
- **green** — implementation, building (backend-developer, angular-developer, nextjs-architecture-expert)
- **yellow** — validation, quality, testing (test-engineer, prompt-engineer)
- **magenta** — creative, documentation, content (technical-writer)
- **red** — critical operations, security, devops (devops-engineer)

### What to Eliminate
- "Communication Protocol" sections with JSON templates — they don't execute.
- "Integration with other agents" lists — routing is handled by the harness or filesystem.
- "Progress tracking" with fake JSON metrics — decorative.
- "Delivery notification" with invented percentages — filler.
- Generic checklists ("Documentation complete, Security robust") — too vague to change behavior.
- Bullet lists of concepts ("Horizontal scaling, Vertical scaling, Data partitioning") — book indexes, not instructions.
- **Rules are not debates.** Don't justify a rule with token-economy stats, training-corpus arguments, or benchmarks-to-convince. State the rule. **Carve-out:** stats, citations, or benchmarks are allowed when they ARE the operational threshold the rule turns on (e.g., "2-4× HTML cost" deciding HTML-vs-Markdown, "400-line PR" triggering split). The test: would removing the stat change the decision? If yes, keep it.

## Routing Maintenance
When creating, editing, or deleting agents or rules, review and adjust impacted files:
- `global/rules/workflow/agent-routing.md` — update the disambiguation table if the new agent overlaps with an existing one, or remove the entry if an agent is deleted.
- Multi-harness layer: after editing any agent or skill, run `python3 harness/build.py` and commit the regenerated trees (deploy also runs it and flags a dirty `harness/`); a NEW skill needs an opencode command wrapper in `harness/opencode/commands/`; a renamed agent/skill needs a grep through `harness/`.
- **Rules don't pass through `build.py`.** `global/rules/` and `global/CLAUDE.md` deploy only to `~/.claude/` (Claude Code). Codex and opencode read only the condensed `harness/AGENTS.md`. So a new or changed rule that applies to all harnesses must be reflected MANUALLY in `harness/AGENTS.md` — there is no generator for this. Rule changes scoped to Claude Code's own mechanics (skill authoring, agent frontmatter) stay in `global/` only.

## File Structure

```
AGENTS.md                          # Canonical guide for all harnesses
CLAUDE.md                          # Imports AGENTS.md via @AGENTS.md; adds Claude Code-specific content
global/                            # Mirrors ~/.claude/ — deployable source of truth
├── CLAUDE.md                      # Core config (always loaded)
├── hooks/                         # Hook scripts (none currently — harness enforces plan mode since 2.1.136)
├── rules/                         # Organized by function, discovered recursively
│   ├── quality/                   # Code principles (alwaysApply)
│   │   ├── communication-format.md # HTML-first policy for substantial human-targeted output
│   │   ├── critical-thinking.md
│   │   ├── debugging.md           # Root-cause discipline: reproduce before fix, one change at a time, 3-fix circuit breaker
│   │   ├── development-principles.md
│   │   ├── patterns-antipatterns.md
│   │   ├── security.md
│   │   └── testing.md
│   ├── languages/                 # Language/framework standards (path-scoped)
│   │   ├── angular-patterns.md
│   │   ├── iac-devops.md          # Docker, Terraform, GH Actions
│   │   ├── java-kotlin.md
│   │   ├── nestjs-patterns.md
│   │   ├── python-standards.md
│   │   ├── react-nextjs.md
│   │   ├── shell-standards.md
│   │   ├── sql-migrations.md      # SQL, Prisma, Drizzle
│   │   ├── tailwind.md
│   │   └── typescript-standards.md
│   ├── workflow/                  # Git, deploys, routing, coordination (alwaysApply)
│   │   ├── agent-routing.md
│   │   ├── cross-service-workflow.md
│   │   ├── devops-principles.md
│   │   ├── gap-resolution.md
│   │   ├── git-workflow.md
│   │   ├── infra-naming.md        # Generic infra naming layer; projects instantiate it in their specs repo
│   │   ├── memory-routing.md      # Engram (work-record) vs native file-memory (always-hot) boundary
│   │   └── project-structure.md   # 3-level hierarchy + file-routing (_support vs specs repo) + session capture layer
│   └── tools/                     # External tools & MCP plugin protocols (alwaysApply)
│       └── context7.md            # Context7 MCP query protocol (installed via plugin)
├── skills/                        # Global skills (deployed to ~/.claude/skills/)
│   ├── agents-md-primary/         # /agents-md-primary — convert projects to AGENTS.md-canonical + CLAUDE.md import
│   ├── engram-init-workspace/     # /engram-init-workspace — unified .engram/config.json for multi-repo workspaces
│   │   ├── SKILL.md
│   │   └── bootstrap-workspace.sh
│   ├── flow-core/                 # Flow pack shared library (non-invocable): contract + templates
│   │   ├── SKILL.md
│   │   └── references/            # ledger-template, handoff-protocol, naming-template, specs-structure, migration-playbook
│   ├── flow-intake/               # /flow-intake — F1: requirements analysis (+ requirements rubric)
│   ├── flow-kickoff/              # /flow-kickoff — F2: workspace bootstrap + ledger + workspace CLAUDE.md
│   ├── flow-specs/                # /flow-specs — F3: init | epic | review (+ spec rubric)
│   ├── flow-mock/                 # /flow-mock — F4: build | review (+ UX rubric)
│   ├── flow-foundation/           # /flow-foundation — F5: naming table, repos, contracts, CI/CD-first
│   ├── flow-plan/                 # /flow-plan — F6: plan the dev session (research | write) with agent-routing table
│   ├── flow-build/                # /flow-build — F6: execute the plan (state-driven reconciler + verify gate)
│   ├── flow-deploy/               # /flow-deploy — F7: qa | prod | verify (+ naming audit)
│   ├── flow-hygiene/              # /flow-hygiene — audit | apply | migrate workspace hygiene
│   ├── flow-report/               # Renders substantial output as self-contained HTML
│   │   └── SKILL.md
│   └── memory-sync/               # /memory-sync — audit | apply: reconcile Engram + native memory vs ground truth
└── agents/                        # Optimized agents, organized by role/color
    ├── design/                    # blue
    ├── development/               # green
    ├── review/                    # cyan
    ├── quality/                   # yellow
    ├── ops/                       # red
    └── docs/                      # magenta
harness/                           # Per-CLI layer — sources + VERSIONED generated trees (deployed by /deploy-global 13b)
├── AGENTS.md                      # Condensed cross-harness guidance (deployed by /deploy-global 13b to ~/.codex/AGENTS.md + ~/.config/opencode/AGENTS.md)
├── build.py                       # Regenerates every generated tree below from global/ — run after agent/skill edits
├── build/                         # convert-agents.py + convert-skills.py (build tooling)
├── agents-skills/                 # GENERATED — cleaned universal skills → ~/.agents/skills (Codex + opencode)
├── codex/                         # README + config.toml.snippet (sources)
│   └── agents/                    # GENERATED — TOML subagents → ~/.codex/agents
└── opencode/                      # README + opencode.jsonc.snippet + commands/ (sources)
    └── agents/                    # GENERATED — markdown subagents → ~/.config/opencode/agents
_support/                          # Workspace material (see global/rules/workflow/project-structure.md)
├── archive/                       # Dated historical snapshots (audits/, docs/) — superseded reports kept for reference
├── archived-agents/               # Retired agents kept for reference
├── docs/                          # Durable docs about this hub
│   └── methodology-bibliography.md # Sources backing global/rules conventions (append-only, consult before re-researching)
├── spec/                          # Design specs and decision records (e.g., dual-mode git workflow)
└── workspace/                     # Ephemeral scratch — content is relocated or deleted when work concludes
.claude/skills/                    # Project-only skills (NOT deployed to ~/.claude/)
├── manage-agents/                 # /manage-agents — validate, optimize, report
│   ├── SKILL.md
│   └── references/template.md     # Agent template (bundled resource)
├── manage-rules/                  # /manage-rules — validate, audit, create
│   └── SKILL.md
└── deploy-global/                 # /deploy-global — meta-skill that deploys global/ to ~/.claude/
```

Path-scoped rules only load when matching files are touched. Agents are discovered recursively — subdirectories provide organizational namespace for humans, not routing logic. Deploy with `/deploy-global`.

### Ephemeral Workspace

`_support/workspace/` is **git-ignored** scratch — never a commit target. When work concludes, each artifact either moves to `_support/archive/<audits|docs>/` with a date-suffixed name (`{slug}-YYYY-MM-DD`) if worth keeping, or is deleted. To commit a generated file, relocate it to `archive/` first.

## Git Conventions
- **Conventional commits:** `feat:`, `fix:`, `chore:`, `docs:` prefixes required.
- **Direct commits and pushes to `master` are this hub's declared workflow** (no PR gate, no CI on branches). Force-push and history rewrites stay gated.
- This is a **configuration-only repo** — no build system, no CI/CD, no runtime. Changes are validated by reading/reviewing agent files, not by running builds or tests.

## Rule Exclusions (this repo)

This repo is config-only. The following inherited global rules do **not** apply to changes here:

- **`quality/testing.md`** — no runtime code to test. Agent prompts, rule files, and skills are not "production code".
- **`Build & Lint`** (from global `CLAUDE.md`) — no build system. Validation is read-review of rule/agent files.
- **`security.md` → Supply Chain Security** — no installable dependencies; no OSV checks to run.
- **`patterns-antipatterns.md`** — code-pattern rules (Promise.all, fs.readFileSync, useState) do not apply to markdown.
- **`critical-thinking.md` → Pre-ship ownership test** (questions 1-2 about runtime load and customer impact) — does not apply to agent/rule text. Risk-surfacing and tradeoff-flagging still apply.

Build/test/lint enforcement is restored automatically in any other repo with runtime code.

## Replicating Global Harness Config
- Source file: `harness/AGENTS.md`.
- Codex global target: `~/.codex/AGENTS.md`.
- OpenCode global target: `~/.config/opencode/AGENTS.md`.
- Before changing behavior that depends on how Codex or OpenCode loads, scopes, resumes, or prioritizes `AGENTS.md`, validate against current Codex and OpenCode documentation when applicable.
- **Codex carries a Codex-only tail.** opencode gets `harness/AGENTS.md` verbatim; Codex gets it PLUS `harness/codex/engram-memory-tail.md` (the Engram memory protocol). opencode and Claude Code inject that protocol through their own plugins, but Codex has none — appending the tail to its `AGENTS.md` is how the protocol reaches Codex (additive, TUI-invisible, survives `engram setup`). So the two deployed files are intentionally NOT identical.
- After changing the source, replicate it with:
  ```bash
  cp harness/AGENTS.md ~/.config/opencode/AGENTS.md
  cat harness/AGENTS.md harness/codex/engram-memory-tail.md > ~/.codex/AGENTS.md
  ```
- Validate with:
  ```bash
  cmp -s harness/AGENTS.md ~/.config/opencode/AGENTS.md && echo "OpenCode MATCH"
  # Codex = shared + tail; check the tail is present and the whole stays under project_doc_max_bytes (49152)
  tail -1 ~/.codex/AGENTS.md | grep -q mem_session_summary && echo "Codex tail OK"
  wc -c ~/.codex/AGENTS.md  # must be < 49152
  ```
- Start a new Codex/OpenCode session after replication. Do not rely on resumed sessions to reflect changed global instructions.

## Validation
- This is a configuration-only repo; validation is mostly diff review and consistency checks.
- For changes to `AGENTS.md` or `harness/AGENTS.md`, verify the files do not reference Claude-only tools as if they were available in other harnesses.
- For changes to deploy behavior, verify the deploy skill still only targets `global/` unless the user explicitly requests a new deployment workflow.
