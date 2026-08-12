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
- Gates in rules and docs name their enforcement layer — deterministic (hook, deny permission, allowlist), confirm-gated (the user's confirmation is the key), or prompt-convention; never phrase a convention as mechanical impossibility. Taxonomy: `_support/docs/enforcement-layers.md`.
- The `<!-- CODEGRAPH_START/END -->` block in `global/CLAUDE.md` and `harness/AGENTS.md` is owned by `codegraph install` — absorbed into the sources (2026-07-10) so installer upgrades report "Unchanged" and `/deploy-global` doesn't delete it. Never edit, dedupe, or reflow content between the markers; hive-specific CodeGraph guidance lives in the bullets AFTER the block. If a codegraph upgrade rewrites its block in the DEPLOYED files, re-absorb the new content into both sources instead of letting them drift.
- Use conventional commit prefixes if asked to commit.
- Never add temporary or incident-specific details (local hotfixes, dated machine state, pending workarounds) to versioned convention docs (`AGENTS.md`, `CLAUDE.md`, `global/`, `harness/` READMEs) unless the user explicitly asks. Route them to machine-local notes (`_support/backup/`, Engram) instead.

## Agent Design Principles

### What Makes a Good Agent
- **Under 120 lines.** Every line must change the agent's output vs default behavior.
- **Unique rules only.** If the global CLAUDE.md already covers it (e.g., "no `any`", "thin controllers"), don't repeat it.
- **Concrete, not generic.** "Use `class-validator` for DTOs" is good. "Follow best practices" is filler.
- **Description controls routing.** The `description` field must be specific and action-oriented — not a resume.
- **Restricted tools.** Only include tools the agent needs. Review agents (cyan) are read-only in effect: Read, Glob, Grep, plus Bash for read-only investigation (git, codegraph, dependency audits) under harness-enforced `permissionMode: plan` — never Write/Edit. Exception: execution-verification reviewers (finding-refuter) run Bash without plan mode to execute claims (tests, repro commands), constrained by prompt to never mutate. Quality agents (yellow) may be remediation-oriented (Write/Edit) or audit-oriented (read-only plus Bash when they orchestrate external analysis).
- **`model: inherit` by default.** The agent uses the session's active model. Only override if there's a strong reason (e.g., `haiku` for a read-only explorer).
- **Calibrate to the floor model, not the ceiling.** Rules and agents must work on the least capable model the user runs day-to-day (as of jul-2026: Sonnet 5 — the executor-tier agents; sessions and top-tier agents run Fable 5, permanent on the plan, with a pinned `opus` tier between). Before cutting a rule as "the model does this by default", verify the *floor* model does it — top-model capability is not a pruning criterion.
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
- Multi-harness layer: after editing any agent or skill, run `python3 harness/build.py` and commit the regenerated trees (deploy also runs it and flags a dirty `harness/`); a NEW **user-invoked** skill needs an opencode command wrapper in `harness/opencode/commands/` (model-invoked router/reference skills need none); a renamed agent/skill needs a grep through `harness/`. Changing a skill's invocation gate (adding/removing `disable-model-invocation`) is also a cross-harness change: it flips the generated Codex `openai.yaml` policy, and it does NOT make the skill organic in opencode (which only exposes gated skills via its command wrappers) — verify the wrapper still matches the intended exposure.
- **`paths:` is the only frontmatter key Claude Code reads.** The docs are explicit: *"Rules without a `paths` field are loaded unconditionally."* So a rule is conditional if and only if it has `paths:`. `alwaysApply: true` is **documentation of intent, not a switch** — the file loads identically without it, and inventing a third scope key does NOT suppress loading (learned the hard way on 2026-07-31: a whole wave shipped on the opposite assumption and reduced nothing). Always-on is the expensive default, paid every session before any work, so keep it for safety gates and for policy whose trigger is an action rather than a file. Before adding `paths:` to an existing rule, apply the three reachability tests in `/manage-rules validate` — a real glob, nothing safety-bearing, and consumers that can still reach it (an executor agent whose `tools:` allowlist omits `Skill` cannot invoke a router skill, and skills are not inherited).
- **AlwaysApply rules don't pass through `build.py` — but router-skill references DO.** `global/rules/` and `global/CLAUDE.md` deploy only to `~/.claude/` (Claude Code); Codex and opencode read the condensed `harness/AGENTS.md` always-on core. A new or changed alwaysApply rule that applies to all harnesses is reflected MANUALLY in `harness/AGENTS.md` when it belongs to the core (gates, every-session procedure) — or, for situational policy, added to a router skill's injected references (`SKILL_REFERENCE_INJECTIONS` in `build.py`: workspace-conventions ← project-structure/session-capture/support-artifacts/cross-service/infra-naming, memory-policy ← memory-routing, unattended-delegation ← unattended-autonomy, language-rules ← languages/* + quality depth + browser-automation), where `build.py` regenerates it automatically. **Exception — path-scoped language rules DO pass through `build.py`:** `global/rules/languages/*.md` → `harness/opencode/rules/` (opencode-rules plugin format, `paths:`→`globs:`), deployed to `~/.config/opencode/rules/` where the `opencode-rules` plugin (pinned 0.6.4, security-audited 2026-07-10) loads them conditionally by touched-file glob — the opencode analog of Claude Code path-scoping. Codex has no equivalent; it keeps only the condensed sections. Rule changes scoped to Claude Code's own mechanics (skill authoring, agent frontmatter) stay in `global/` only.
- **Grok scope (`deploy-global --only grok` / part of `harness`):** (1) always-on rules — Grok's scan is NOT recursive and ignores `paths:` (verified against grok 0.2.118), so each always-on rule is symlinked flat into `~/.grok/rules/` as `<dir>__<file>.md` → `~/.claude/rules/` copy; a new always-on rule is picked up on the next deploy; **giving an existing rule `paths:` removes it from Grok**, so situational content must reach Grok via a router skill's `SKILL_REFERENCE_INJECTIONS` (same as Codex); rule filenames must never contain `__` (flatten separator). (2) agents — `harness/build.py` emits `harness/grok/agents/*.md` from `global/agents/`; deploy copies them to `~/.grok/agents/` so `spawn_subagent` can use the hub roster by name. (3) hooks — **not** re-copied under `~/.grok/hooks/`; Grok merges `~/.claude/settings.json` via compat, and hive hook scripts are dual-runtime (Claude + Grok payloads). Grok also loads `~/.claude/CLAUDE.md` natively.

## File Structure

```
AGENTS.md                          # Canonical guide for all harnesses
CLAUDE.md                          # Imports AGENTS.md via @AGENTS.md; adds Claude Code-specific content
global/                            # Mirrors ~/.claude/ — deployable source of truth
├── CLAUDE.md                      # Core config (always loaded)
├── hooks/                         # Hook scripts + settings-config.json blocks, deployed/merged by /deploy-global (bash-policy, post-tool-hub, flow-session-context, flow-context, flow-plan-capture, session-hygiene-report)
├── rules/                         # Organized by function, discovered recursively
│   ├── quality/                   # Code principles (6 alwaysApply, 1 path-scoped)
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
│   │   ├── typescript-standards.md
│   │   └── ui-visual-design.md    # Visual craft: type scale, spacing, contrast, action hierarchy
│   ├── workflow/                  # Git, routing, coordination (9 always-on, 2 path-scoped)
│   │   ├── agent-routing.md
│   │   ├── cross-service-workflow.md
│   │   ├── devops-principles.md  # path-scoped (Dockerfile/tf/workflows)
│   │   ├── gap-resolution.md
│   │   ├── git-workflow.md
│   │   ├── infra-naming.md        # Generic infra naming; projects instantiate it in their specs repo
│   │   ├── memory-routing.md      # Engram vs native file-memory boundary
│   │   ├── project-structure.md   # 3-level hierarchy + file-routing (_support vs specs repo)
│   │   ├── session-capture.md     # Session layer + subfolder vocabulary (split out of project-structure)
│   │   ├── support-artifacts.md   # Path-scoped (_support/**): generated-artifact naming, retention, versioning, legacy mappings
│   │   └── unattended-autonomy.md # The delegated-run mode
│   └── tools/                     # External tools & MCP protocols (3 always-on)
│       ├── browser-automation.md  # agent-browser CLI vs MCP browser servers
│       ├── code-search.md         # rg vs jbcontext vs codegraph routing + anti-conclusion discipline
│       └── context7.md            # Context7 MCP query protocol (installed via plugin)
├── skills/                        # Global skills (deployed to ~/.claude/skills/)
│   ├── adversarial-research/      # /adversarial-research — N independent generators + finding-refuter cross-exam → refuted/weakened/surviving/net-new canon
│   ├── agents-md-primary/         # /agents-md-primary — convert projects to AGENTS.md-canonical + CLAUDE.md import
│   ├── engram-init-workspace/     # /engram-init-workspace — unified .engram/config.json for multi-repo workspaces
│   │   ├── SKILL.md
│   │   └── bootstrap-workspace.sh
│   ├── flow-core/                 # Flow pack shared library (non-invocable): contract + templates
│   │   ├── SKILL.md
│   │   └── references/            # ledger-template, handoff-protocol, naming-template, specs-structure, migration-playbook, promotion-playbook, ux-rubric
│   ├── flow-brainstorming/        # /flow-brainstorming — business-idea iteration into a decision
│   ├── flow-start/                # /flow-start — greenfield wizard: intake + workspace bootstrap + technical foundation
│   ├── flow-specs/                # /flow-specs — write/review specs: init | epic | review (+ spec rubric)
│   ├── flow-plan/                 # /flow-plan — write the session plan (research | write) with agent-routing table
│   ├── flow-build/                # /flow-build — execute the plan (state-driven reconciler + verify gate)
│   ├── flow-hygiene/              # /flow-hygiene — audit | apply | migrate workspace hygiene
│   ├── flow-report/               # Renders substantial output as self-contained HTML
│   │   └── SKILL.md
│   ├── memory-sync/               # /memory-sync — audit | apply: reconcile Engram + native memory vs ground truth
│   └── starlight-docs-site/       # /starlight-docs-site — scaffold | page | audit Astro Starlight docs (user-manual | spec-site)
│       ├── SKILL.md
│       ├── references/            # chassis, profile-*, best-practices, audit-checklist
│       └── templates/            # chassis + user-manual + spec-site (copy-ready, pinned versions)
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

`_support/workspace/` is **git-ignored** scratch — never a commit target. When work concludes, each artifact either moves to `_support/archive/<audits|docs>/` with a date-prefixed name (`YYYY-MM-DD-{slug}`) if worth keeping, or is deleted. To commit a generated file, relocate it to `archive/` first. Existing archive entries keep their legacy names — no renames.

## Git Conventions
- **Conventional commits:** `feat:`, `fix:`, `chore:`, `docs:` prefixes required.
- **Direct commits to `master` are this hub's declared workflow** (no PR gate, no CI on branches). **Push is confirm-gated: never push until the user confirms they agree with the changes** — their explicit instruction after reviewing them is the confirmation. Force-push and history rewrites stay gated.
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
- **Both harnesses get the shared file verbatim.** The Engram protocol reaches every harness through its own plugin — Codex included since the Engram Codex plugin (bundled SessionStart/UserPromptSubmit/Stop hooks) shipped; the former `engram-memory-tail.md` concat was removed 2026-07-10 as a duplicate.
- After changing the source, replicate it with:
  ```bash
  cp harness/AGENTS.md ~/.config/opencode/AGENTS.md
  cp harness/AGENTS.md ~/.codex/AGENTS.md
  ```
- Validate with:
  ```bash
  cmp -s harness/AGENTS.md ~/.config/opencode/AGENTS.md && echo "OpenCode MATCH"
  cmp -s harness/AGENTS.md ~/.codex/AGENTS.md && echo "Codex MATCH"
  wc -c ~/.codex/AGENTS.md  # must be well under 65536 (project_doc_max_bytes) — the budget is
                            # COMBINED across the chain (global + workspace + repo AGENTS.md),
                            # so this check is a lower bound, not the full budget
  ```
- Start a new Codex/OpenCode session after replication. Do not rely on resumed sessions to reflect changed global instructions.

## Validation
- This is a configuration-only repo; validation is mostly diff review and consistency checks.
- **Shell is the exception — it has a real linter.** After touching any `.sh`, run `find global .claude harness -name '*.sh' -print0 | xargs -0 shellcheck -S style` and keep it at zero findings. A genuine false positive gets `# shellcheck disable=SC####` with the reason inline, placed **before the compound command** (`while`/`if`), never before its `done`/`fi` — a misplaced directive makes shellcheck skip the whole block instead of one line. `_support/backup|archive/**` is third-party or frozen; leave it out of scope.
- For changes to `AGENTS.md` or `harness/AGENTS.md`, verify the files do not reference Claude-only tools as if they were available in other harnesses.
- For changes to deploy behavior, verify the deploy skill still only targets `global/` unless the user explicitly requests a new deployment workflow.
