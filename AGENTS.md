# Agent Configuration Hub

## Purpose
This is the canonical guide for all harnesses. `CLAUDE.md` imports it via `@AGENTS.md` and adds Claude Code-specific content below the import.

This repo is the **source of truth** for the user's global agent configuration. Claude Code configuration lives under `global/`: global CLAUDE.md, the rule store, and optimized subagent definitions. Everything under `global/` deploys to `~/.claude/` via the `/deploy-global` skill. **Deploy is always user-initiated** — never execute `/deploy-global` or copy files to `~/.claude/` without explicit user instruction.

The reusable generic harness configuration lives in `harness/AGENTS.md`. Edit that file when changing cross-project AGENTS guidance. Do not duplicate its full contents here.

## Context
The user is a software architect and developer. Agents must not repeat global rules; they inherit them automatically.

**Portfolio snapshot — 2026-08-20.** 54 repos on disk, 44 with a commit in the last 90 days:

| Category (of the 44 active) | # | Notes |
|---|---|---|
| Product, high personal volume | 9 | ark-monorepo, sample-project ×4, globex-backend-client, globex-web-client-reforge, umbrella-backoffice, educavita-next-monorepo |
| Product, team-led (he reviews more than he writes) | ~13 | hooli, fleetsystem, initech ×6, remuneri-reforge-* |
| Specs | 6 | documentation repos, no runtime |
| Mocks / prototypes | 4 | |
| Cloned reference repos | 3 | optional reference project, optional reference project, improve — none of his commits |
| Config / meta | 4 | tricell-hive, tricell-hive-private (archived), rfp-estimations, terminal-ui-cv |
| Marketing / landing | 3 | |
| Dormant (>90d) | 10 | all of `ebitware`, untouched since 2026-03 |

Three live client groups — `sample-workspace`, `sample-organization`, and Tricell's own work; `ebitware/` frozen. Global Hitss ended 2026-08-26 (no local `hitss/` tree). Work concentrates hard: `ark-monorepo` alone carries ~40% of his commit volume.

**Monorepos are the default shape from here on** — Turborepo + workspaces with a shared contracts package between front and back. Seven active repos already follow it (ark, hooli-frontend, sample-project, fleetsystem-frontend, internal-apps, educavita, umbrella); two more are Nx (globex-web-client-reforge, e-hub-web-frontend). Assume new work is a monorepo unless told otherwise.

Stack across the portfolio: Next.js, Angular, NestJS, Express, Java/Spring, Kotlin, Python, and DevOps (GitHub Actions, AWS, Hetzner, Vercel, Dokploy). Day-to-day volume is TypeScript-dominant (Next/Angular/Nest, Prisma/Drizzle) with long-running Node workers (BullMQ + Redis + S3) in production; Java/Kotlin/Python are in the working set but idle in the repos on disk at this snapshot.

Refresh the snapshot with:

```bash
find ~/Development/projects -maxdepth 4 -name .git | sed 's|/.git$||' | while read -r r; do printf "%s %s\n" "$(git -C "$r" log -1 --format=%cs)" "$r"; done | sort -r
```

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
- `harness/agents-skills/`, `harness/codex/agents/`, and `harness/opencode/agents/` are GENERATED trees. Never hand-edit them; edit the canonical source under `global/`. The versioned `.githooks/pre-commit` regenerates and stages every generated output (and the hive profile) whenever a canonical source is staged — enable it once per clone with `git config core.hooksPath .githooks`; without it, run `python3 harness/build.py` and commit the output by hand.
- Do not alias or symlink `AGENTS.md` to `harness/AGENTS.md`; they serve different scopes.
- When changing routing or domains, check `global/rules-situational/agent-routing.md` (router: `task-routing`).
- When a rule or convention is grounded in external authority (standards, canonical books, official docs), record the source in `_support/docs/methodology-bibliography.md` and consult it before re-researching.
- Gates in rules and docs name their enforcement layer — deterministic (hook, deny permission, allowlist), confirm-gated (the user's confirmation is the key), or prompt-convention; never phrase a convention as mechanical impossibility. Taxonomy: `_support/docs/enforcement-layers.md`.
- **`flow-<x>` names a user-gated command of the flow pack** (`disable-model-invocation: true`). Pack infrastructure is the declared exception and carries no gate: `flow-core` (non-invocable library) and `flow-report` (model-invoked renderer; format will grow beyond HTML). A new `flow-*` skill that is not a gated command takes a non-flow name.
- Use conventional commit prefixes if asked to commit.
- Never add temporary or incident-specific details (local hotfixes, dated machine state, pending workarounds) to versioned convention docs (`AGENTS.md`, `CLAUDE.md`, `global/`, `harness/` READMEs) unless the user explicitly asks. Route them to machine-local notes (`_support/backup/`, Engram) instead.

## Agent Design Principles

### What Makes a Good Agent
- **Every line must change the agent's output vs default behavior.** 120 lines of the agent's own text is a review threshold, never a cap: past it, look for filler; lines that carry weight stay. The rule texts an agent carries through `packs:` are inlined at build and never count.
- **Unique rules only.** If the global CLAUDE.md already covers it (e.g., "no `any`", "thin controllers"), don't repeat it.
- **Concrete, not generic.** "Use `class-validator` for DTOs" is good. "Follow best practices" is filler.
- **Description controls routing.** The `description` field must be specific and action-oriented — not a resume.
- **Restricted tools.** Only include tools the agent needs. Review agents (cyan) are read-only in effect: Read, Glob, Grep, plus Bash for read-only investigation (git, `rg`, dependency audits), plus WebSearch/WebFetch and the context7 MCP tools for external verification (version-sensitive APIs, CVEs/advisories, ecosystem claims — read-only in effect) — never Write/Edit. They also carry `permissionMode: plan`, but **the `tools:` allowlist is what actually enforces read-only**: a parent session in auto mode (the default on Pro/Max/Team) makes a subagent inherit auto mode and ignore its frontmatter `permissionMode` entirely (`_support/docs/enforcement-layers.md`). Exception — reviewers and verifiers (cyan or yellow) that must EXECUTE to observe run Bash outside plan mode, constrained by a `tools:` allowlist without Write/Edit (review-refuter runs tests/repro commands) or an explicit `disallowedTools` when the agent needs the inherited surface (review-ux and sdd-verify drive a browser), plus a never-mutate prompt clause. Quality agents (yellow) may be remediation-oriented (Write/Edit) or audit-oriented (read-only plus Bash when they orchestrate external analysis).
- **Model by tier, not by default.** Three tiers, matching the roster in force: `inherit` for judgment roles that must match the session ceiling (designers, refuter, security); `opus` for deep-reasoning specialists; `sonnet` for executor, discovery, and the judgment roles deliberately kept at the floor (sdd-spec-reviewer, review-ux, sdd-verify) — the discovery floor per `code-search.md > Model floor for discovery agents`, never haiku there. `effort: high` accompanies every judgment role regardless of tier. Pick the tier when creating the agent; escalate per-invocation when a task proves reasoning-heavy.
- **Calibrate to the floor model, not the ceiling.** Rules and agents must work on the least capable model the user runs day-to-day (as of jul-2026: Sonnet 5 — the executor-tier agents; sessions and top-tier agents run Fable 5, permanent on the plan, with a pinned `opus` tier between). Before cutting a rule as "the model does this by default", verify the *floor* model does it — top-model capability is not a pruning criterion.
- **No rule loads by file kind.** A rule reaches an agent inlined via `packs:`, held by
  `rule-delivery` on a matching write, or read from a router's `references/` — pick one
  deliberately, never all three.

### Naming Rules
- 3-50 characters, lowercase letters, numbers, and hyphens only.
- Must start and end with alphanumeric. No underscores, spaces, or special characters.
- `review-code` ✅ — `my_agent` ❌ — `ag` ❌ (too short) — `-agent-` ❌ (starts/ends with hyphen)

### Description Best Practices
- 1-3 sentences for agents with obvious routing (e.g., a repo with `angular.json` goes to `angular-developer`).
- **Triggers, not workflow.** A description states WHEN to invoke — it never compresses the body's procedure into steps. The body loads only on invocation; a description that summarizes the workflow gets followed instead of the body. Enumerating subcommands and what each delivers is fine; enumerating the execution sequence is not. Applies to agents and skills equally.
- For agents where routing is ambiguous, include `<example>` blocks with `<commentary>` to help Claude decide when to invoke:
```
description: >
  Use this agent when evaluating architecture decisions...

  <example>
  Context: Team is designing service boundaries before implementation begins.
  user: "Design the service boundaries for the new module"
  assistant: "I'll design the contracts and boundaries..."
  <commentary>
  Invoke sdd-design for pre-implementation macro design, not code-level review.
  </commentary>
  </example>
```

### Color Assignment by Role
- **blue** — design, analysis (sdd-design)
- **cyan** — review, research (review-code, review-security)
- **green** — implementation, building (backend-developer, angular-developer, react-developer)
- **yellow** — validation, quality, testing (test-engineer, prompt-engineer)
- **magenta** — creative, documentation, content (sdd-spec-writer)
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
- `global/rules-situational/agent-routing.md` — update the disambiguation table if the new agent overlaps with an existing one, or remove the entry if an agent is deleted.
- `README.md` — the inventory of record for humans: keep the agents table (heading count + one row per agent + tool surface) and the skills table in sync with disk.
- **Per-harness loading READMEs** (`global/README.md`, `harness/{codex,opencode,grok}/README.md`) — each carries a *What this harness loads* table whose rows cite an official doc URL with a verification date. A change to what a harness receives (a rule changing delivery channel — core include, hook trigger, router, pack — a new generated tree, a hook target) updates the affected table; a claim about upstream behavior carries a URL that was fetched, or is marked **undocumented** rather than given a plausible-looking link. Prompt-convention. Enforcement: `/manage-agents validate` checks the agents table; the skills table is prompt-convention.
- Multi-harness layer: after editing any agent, skill, or core section under `global/core-sections/`, the pre-commit hook regenerates and stages the trees with the commit (deploy also runs `build.py` and flags a dirty `harness/` — the sign the hook is not enabled in this clone); a NEW **user-invoked** skill needs an opencode command wrapper in `harness/opencode/commands/` (model-invoked router/reference skills need none); a renamed agent/skill needs a grep through `harness/`. Changing a skill's invocation gate (adding/removing `disable-model-invocation`) is also a cross-harness change: it flips the generated Codex `openai.yaml` policy, and it does NOT make the skill organic in opencode (which only exposes gated skills via its command wrappers) — verify the wrapper still matches the intended exposure.
- **Adding or changing a rule — the short form; the procedure is `global/rules-situational/README.md`.** One store: every rule text lives in `global/rules-situational/`, and the CHANNEL decides how it arrives. Pick the shape first: policy that governs how the agent thinks or answers in EVERY session, with no observable act to hang it on → **core include** (the text here + a `global/core-sections/rule-<name>.md` carrying only `order:`/`targets:`/`include:`, no body, no trigger); an observable act — writing a file of a kind, running a command → **trigger** (`globs:` and/or `commands:`) which counts ONLY together with a `SKILL_REFERENCE_INJECTIONS` entry, plus that router's routing-table row, `RULE_READER_NAMES` if read-only agents need it, the `packs:` line on the specialized agents, and `exclusive-with:` where another rule claims the same file conventions; an intent with no act → **router-only**, with a one-line gate in the core when absence is unsafe; executor-only → **pack-only**. `paths:` is a build error, and so is a rule no channel delivers, a trigger with no reference, and an `include:` over a triggered rule. A trigger is written as an act a third party could point to in the transcript — never as an intent, never as a sentinel glob. Always-on is the expensive default, paid every session before any work: keep it for safety gates and for policy whose trigger is an action rather than a file. Before finishing: `python3 harness/build.py --check`, the hook's own test against the real manifest, the cost of the rule (KB for a core include; denial rounds and Grok's ~264-character reason budget for a trigger), and an in-vivo run in a sandbox OUTSIDE the temp roots, which are ungated by design.

- **Placement in `harness/AGENTS.md`: gate and pointer here, mechanics in the router — never both.** The core is paid in full at the start of every session in every project, so it holds only **safety gates** and **policy whose trigger is an action rather than a file**, plus the one-line pointer to the router that carries the rest. When the same content lives in the core AND in a router reference, the core is paying twice for what the router already delivers — that duplication — not prose length — is what makes the core expensive. **There is no size threshold on the file and none is wanted:** no harness caps it (verified 2026-08-18; bibliography), so `build.py` reports its size and per-session token cost and enforces nothing. Adding to the core: decide the layer first. A rising number is a cue to audit placement, never a reason to reword paragraphs that earned their place.
- **The two always-on cores are GENERATED from `global/core-sections/`.** `global/CLAUDE.md` (deploys to `~/.claude/`, Claude Code) and `harness/AGENTS.md` (the condensed core Codex and opencode read) are assembled by `harness/build.py` from the section files in `global/core-sections/` (format: its README) — edit a section, rebuild, commit both outputs. The build regenerates a stale output silently and REFUSES a hand-edited one (differs from both the regeneration and HEAD — the edit stays on disk); `python3 harness/build.py --check` runs the parity checks without writing (deploy preflight / pre-commit). A policy stated by both cores CAN live in one shared section (`targets: [claude, agents]`) and four do today; the remaining condensed twins are still per-target section pairs synced by hand — only the delegation thresholds carry their own parity check — so prefer promoting a twin to shared when editing it. A rule the Claude core includes reaches Claude Code and Grok as core text and no one else: a new or changed rule that applies to all harnesses is condensed into a core section targeting `agents` when it belongs to the core (gates, every-session procedure) — or, for situational policy, added to a router skill's injected references — `SKILL_REFERENCE_INJECTIONS` in `harness/build.py` is the authoritative skill←rule mapping (do not restate it here; it drifts) — where `build.py` regenerates it automatically. **Everything under `global/rules-situational/` DOES pass through `build.py`:** injected into the router skills' `references/`, inlined into the agents that declare it in `packs:`, and indexed in `harness/rule-manifest.json` for the `rule-delivery` hook. opencode's glob channel (the `opencode-rules` plugin and `harness/opencode/rules/`) was retired in M3 — it never loaded a rule in the deployed state. Rule changes scoped to Claude Code's own mechanics (skill authoring, agent frontmatter) stay in `global/` only.
- **Grok scope (`deploy-global --only grok` / part of `harness`):** (1) rules — **no rules scope any more**: `~/.grok/rules/` receives nothing and the next deploy empties it. Grok gets rule text exactly three ways, all of them already deployed by other scopes: the always-on core, because it loads `~/.claude/CLAUDE.md` natively (the 7 included rules ride inside it); a router skill's `SKILL_REFERENCE_INJECTIONS` under `~/.claude/skills/`, which Grok scans; and the `rule-delivery` hold on a matching write or command — same as Codex. (2) agents — `harness/build.py` emits `harness/grok/agents/*.md` from `global/agents/`; deploy copies them to `~/.grok/agents/` so `spawn_subagent` can use the hub roster by name. (3) hooks — **not** re-copied under `~/.grok/hooks/`; Grok merges `~/.claude/settings.json` via compat, and hive hook scripts are dual-runtime (Claude + Grok payloads). (4) skills — **no deploy step and no generated tree**: Grok scans `~/.claude/skills/` by default (`[compat.claude] skills = true`), so `global/skills/` reaches it through the Claude deploy alone, and Grok honors `disable-model-invocation` natively, so the `flow-*` gate holds there without translation. Two consequences: never build a `harness/grok/skills/` tree (pure duplication — unlike Codex, which needs a generated `openai.yaml` for the same gate), and note that `~/.claude/skills/` is Grok's LOWEST-precedence source, so a same-named Grok-native or repo skill wins. Grok also loads `~/.claude/CLAUDE.md` natively (undocumented upstream; verified empirically via `grok inspect`, which lists every rules file Grok loads — the deterministic scope auditor, analog of `codex debug prompt-input`).

## File Structure

```
AGENTS.md                          # Canonical guide for all harnesses
CLAUDE.md                          # Imports AGENTS.md via @AGENTS.md; adds Claude Code-specific content
global/                            # Mirrors ~/.claude/ — deployable source of truth
├── README.md                      # What Claude Code loads + the official doc backing each mechanism (verified URLs)
├── CLAUDE.md                      # GENERATED always-on core (assembled from core-sections/ by harness/build.py)
├── core-sections/                 # Canonical section files for BOTH always-on cores (global/CLAUDE.md + harness/AGENTS.md);
│                                  # a `rule-<name>.md` section is body-less and carries `include:` — the build
│                                  # inlines that rule text from rules-situational/ into the Claude core
├── hooks/                         # Hook scripts + settings-config.json blocks, deployed/merged by /deploy-global (bash-policy, rule-delivery, post-tool-hub, flow-session-context, flow-context, reviewer-guard, executor-dispatch-gate, session-hygiene-report)
├── rules-situational/             # THE rule-text store (42 texts) — NEVER deployed to ~/.claude/rules.
│                                  # 7 inlined into the core by a `rule-*` section; 26 carrying
│                                  # `globs:`/`commands:` → held by the rule-delivery hook on a matching
│                                  # write or command; 8 router-only; 1 pack-only. Injection into a
│                                  # router's references/ and inlining via `packs:` are build steps.
│                                  # Its README carries the shapes and the procedure for adding one.
│   ├── README.md
│   ├── agent-core-gates.md               # Pack-only: the gates a packed agent loses with the global corpus (required in every `packs:`)
│   ├── agent-routing.md
│   ├── angular-patterns.md
│   ├── browser-automation.md             # Gate block: delegation, profile, viewport (commands: agent-browser)
│   ├── browser-automation-reference.md   # CLI mechanics + MCP escalation (via language-rules)
│   ├── code-search.md                    # Search routing + anti-conclusion discipline (router-only)
│   ├── communication-format.md           # CORE INCLUDE: flow-report trigger + carve-outs
│   ├── communication-format-mechanics.md # Layout floor, in-thread form, diagram norm (via flow-report)
│   ├── config-authoring.md               # How a directive is written into a config/rule surface (globs: CLAUDE.md, AGENTS.md, rules, skills, agents)
│   ├── context7.md                       # Context7 MCP query protocol (commands: dependency installs)
│   ├── critical-thinking.md              # CORE INCLUDE: trivial carve-out, decision ownership, pre-ship test
│   ├── cross-service-workflow.md
│   ├── debugging.md                      # CORE INCLUDE: root-cause discipline, 3-fix breaker, incident response
│   ├── development-principles.md         # How to write code (globs: code + IaC files)
│   ├── devops-principles.md              # Deploy/CI doctrine (globs: Dockerfile/tf/workflows)
│   ├── gap-resolution.md
│   ├── git-mechanics.md                  # Branching, commits, PRs, promotion (commands: git/gh verbs)
│   ├── git-workflow.md                   # CORE INCLUDE: git authorization + safety gates
│   ├── iac-devops.md                     # Docker, Terraform, GH Actions
│   ├── identifier-language.md            # Domain-translation judgment layer (its gate is in CLAUDE.md > Code Layer)
│   ├── infra-naming.md                   # Generic infra naming; projects instantiate it in their specs repo
│   ├── java-kotlin.md
│   ├── memory-routing.md
│   ├── nestjs-patterns.md
│   ├── patterns-antipatterns.md
│   ├── project-structure.md              # 3-level hierarchy + file-routing (_support vs specs repo)
│   ├── python-standards.md
│   ├── react-nextjs.md
│   ├── reporting-integrity.md            # CORE INCLUDE: ground-truth state claims + Fix at the Root
│   ├── security-floor.md                 # CORE INCLUDE: exposure floor + authentication & secrets
│   ├── security.md                       # The code/install half: validation, injection, errors, supply chain
│   ├── session-capture.md                # Session layer + subfolder vocabulary (split out of project-structure)
│   ├── shell-standards.md
│   ├── sql-migrations.md                 # SQL, Prisma, Drizzle
│   ├── support-artifacts.md              # Generated-artifact naming, retention, versioning, legacy mappings (_support/**)
│   ├── tailwind.md
│   ├── test-gate.md                      # The verifiable test gate condensed for executors (globs: code files)
│   ├── testing.md                        # Full testing policy: approach, coverage, execution scope (router-only)
│   ├── typescript-standards.md
│   ├── ui-visual-design.md               # Visual craft: type scale, spacing, contrast, action hierarchy
│   ├── unattended-autonomy.md            # CORE INCLUDE: activation guard + gate pointers
│   └── unattended-autonomy-mode.md       # Full delegated-run mechanics (via unattended-delegation)
├── skills/                        # Global skills (deployed to ~/.claude/skills/)
│   ├── adversarial-research/      # /adversarial-research — N independent generators + review-refuter cross-exam → refuted/weakened/surviving/net-new canon
│   ├── agents-md-primary/         # /agents-md-primary — convert projects to AGENTS.md-canonical + CLAUDE.md import
│   ├── engram-init-workspace/     # /engram-init-workspace — unified .engram/config.json for multi-repo workspaces
│   │   ├── SKILL.md
│   │   └── bootstrap-workspace.sh
│   ├── flow-core/                 # Process library (non-invocable): contract, templates, playbooks
│   │   ├── SKILL.md
│   │   ├── references/            # ledger-template, handoff-protocol, naming-template, specs-structure, plan-format,
│   │   │                          #   judgment-criteria, ux-rubric, spec-rubric, requirements-rubric + the playbooks:
│   │   │                          #   bootstrap, spec-writing, migration, workspace-hygiene, audit, promotion
│   │   └── templates/             # workspace-agents, workspace-claude, sessions-permissions
│   ├── flow-plan/                 # /flow-plan — portable planning, approval and scoped authorization
│   ├── flow-build/                # /flow-build — execute or resume an authorized plan (reconciler + verify gate)
│   ├── flow-report/               # Renders substantial output as self-contained HTML
│   │   └── SKILL.md
│   ├── memory-sync/               # /memory-sync — audit | apply: reconcile Engram + native memory vs ground truth
│   ├── monorepo-cutover/          # /monorepo-cutover — cutover playbook (multi-repo → monorepo)
│   ├── workspace-archive/         # /workspace-archive — run | normalize: archive verified sessions > 15 days to sessions/archived/ (on request)
│   ├── status-fetch/              # Fetches live external state in an isolated subagent
│   ├── task-routing/              # Router (model-invoked): who gets the task + plan gap analysis + git mechanics (branching, commits, PRs, promotion, close)
│   ├── language-rules/            # Router (model-invoked): language rules for Codex/Grok — references injected by build.py
│   ├── memory-policy/             # Router (model-invoked): Engram policy for Codex/opencode — references injected by build.py
│   ├── workspace-conventions/     # Router (model-invoked): workspace/session/contract conventions — references injected by build.py
│   ├── unattended-delegation/     # Router (model-invoked): delegated unattended runs — references injected by build.py
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
├── README.md                      # The layer as a whole + the two manual-merge snippets
├── AGENTS.md                      # GENERATED condensed cross-harness core (assembled from global/core-sections/; deployed by /deploy-global 13b to ~/.codex/AGENTS.md + ~/.config/opencode/AGENTS.md)
├── build.py                       # Regenerates every generated tree below from global/ — run after agent/skill edits
├── build/                         # convert-agents.py + convert-skills.py (build tooling)
├── agents-skills/                 # GENERATED — cleaned universal skills → ~/.agents/skills (Codex + opencode)
├── rule-manifest.json             # GENERATED — every rule text: globs, deployed reference paths, `readers`, which agents carry it as a pack, read-only agents; read by the rule-delivery hook
├── claude/                        # README (generated marker)
│   └── agents/                    # GENERATED — Claude Code subagents with their `packs:` inlined → ~/.claude/agents (mirrors global/agents/<role>/)
├── codex/                         # README + config.toml.snippet (sources)
│   └── agents/                    # GENERATED — TOML subagents → ~/.codex/agents
├── opencode/                      # README + opencode.jsonc.snippet + commands/ + permission-config.json (sources)
│   └── agents/                    # GENERATED — markdown subagents → ~/.config/opencode/agents
├── grok/                          # README (rules reach Grok as flat symlinks; no skills tree by design)
│   └── agents/                    # GENERATED — Grok-shaped markdown subagents → ~/.grok/agents
└── cursor/                        # README only — reads ~/.agents/skills natively and ~/.claude/{agents,settings.json} via compat; nothing generated, nothing deployed
_support/                          # Workspace material (see global/rules-situational/project-structure.md)
├── archive/                       # Dated historical snapshots (audits/, docs/) — superseded reports kept for reference
├── archived-agents/               # Retired agents kept for reference
├── docs/                          # Durable docs about this hub
│   └── methodology-bibliography.md # Sources backing the rule and core conventions (append-only, consult before re-researching)
├── spec/                          # Design specs and decision records (e.g., dual-mode git workflow)
└── workspace/                     # Ephemeral scratch — content is relocated or deleted when work concludes
.githooks/pre-commit               # Regenerates + stages harness/, global/CLAUDE.md and the hive profile with the commit (enable: git config core.hooksPath .githooks)
.claude/skills/                    # Project-only skills (NOT deployed to ~/.claude/)
├── manage-agents/                 # /manage-agents — validate (agents + rule texts / core sections)
│   └── SKILL.md
└── deploy-global/                 # /deploy-global — meta-skill that deploys global/ to ~/.claude/
.agents/skills/                    # Symlinks -> .claude/skills/* so a non-Claude harness in THIS repo
                                   # runs the same two skills (copies drifted; symlinks cannot)
.codex/                            # hooks.json + hooks/agent-line-count.sh -> ../../.claude/hooks/ (same file)
```

**opencode is not used in this repo, and that is what closes an otherwise real exposure.** opencode
discovers `.agents/skills/<name>/SKILL.md` project-locally (walking up to the git worktree) and
*disregards frontmatter fields it does not recognize* — `disable-model-invocation` among them — so the
`deploy-global` symlink above would be model-invocable there, where Claude Code and Grok both honor the
gate. Documented rather than mitigated: the exposure needs opencode to be run here to exist. Revisit if
that changes. (Per its published docs, 2026-08-27; not verified by running opencode.)

**Repo-local Codex hooks require trust before they run.** `.codex/hooks.json` is read (verified by
experiment, Codex 0.149.1: the same hook fires with `--dangerously-bypass-hook-trust` and stays silent
without it, matcher held constant). Trust is persisted per hook in `~/.codex/config.toml` under
`[hooks.state."<abs path>:<event>:<i>:<j>"]` as `trusted_hash` + `enabled = true`, keyed to a hash of the
hook config — so editing `hooks.json` invalidates it. Until the trust prompt is accepted in an
interactive Codex session here, the hook is inert and fails silently: no error, no notice.

No rule loads by file kind: every text in `global/rules-situational/` arrives inlined into the always-on core by a `rule-*` section, by hook hold, by router skill, or by pack. Agents are discovered recursively — subdirectories provide organizational namespace for humans, not routing logic. Deploy with `/deploy-global`.

### Ephemeral Workspace

`_support/workspace/` is **git-ignored** scratch — never a commit target. When work concludes, each artifact either moves to `_support/archive/<audits|docs>/` with a date-prefixed name (`YYYY-MM-DD-{slug}`) if worth keeping, or is deleted. To commit a generated file, relocate it to `archive/` first. Existing archive entries keep their legacy names — no renames.

## Git Conventions
- **Tracker: GitHub Issues** (this repo, via `gh`). Declared per `memory-routing.md > Tracker sync`: reads are standing-authorized; a ticket moves to in-progress when its work starts, closes/comments batch to the plan's close confirmation.
- **Conventional commits:** `feat:`, `fix:`, `chore:`, `docs:` prefixes required.
- **Direct commits to `master` are this hub's declared workflow** (no PR gate, no CI on branches). **The review gate sits before the commit, not before the push:** present the diff, commit on the user's approval, and push in the same step — that one approval covers both, so never ask a second time for the push. Confirm-gated; force-push and history rewrites stay separately gated.
- This is a **configuration-only repo** — no build system, no CI/CD, no runtime. Changes are validated by reading/reviewing agent files, not by running builds or tests.

## Rule Exclusions (this repo)

Carried by the compiled hive profile at the end of this file (`hive-profile` block, class: config-hub); the pre-commit hook refreshes it with any commit that touches `global/rules-situational/` or the classifier (by hand: `python3 harness/hive-compile.py . --apply`). Staleness is content-based — `hive-compile.py --check` compares the profile it would compile today against the block on disk, stamp excluded — so a rules commit that leaves the profile identical is not stale; the `session-hygiene-report` hook delegates its advisory to that check. Residual nuance the generator does not express: the exclusions cover agent prompts, rule files, and skills — none of them "production code" — and `critical-thinking.md`'s risk-surfacing and tradeoff-flagging still apply here.

**Per-repo override — `.hive-profile.yaml` at the target repo's root** (flat `key: value` lines, no nesting; keys: `class`, `tracker`, `exclusions` (comma-separated rule paths), `notes`). Overrides win over detection; `tracker` is emitted only from this file, never inferred. Read by `hive-compile.py`; no repo declares one today — it exists for the case detection gets a repo wrong or a repo needs an extra exclusion.

## Replicating Global Harness Config
- Source file: `harness/AGENTS.md` — itself GENERATED from `global/core-sections/`: edit the sections and run `python3 harness/build.py` first, never the output.
- Codex global target: `~/.codex/AGENTS.md`.
- OpenCode global target: `~/.config/opencode/AGENTS.md`.
- Before changing behavior that depends on how Codex or OpenCode loads, scopes, resumes, or prioritizes `AGENTS.md`, validate against current Codex and OpenCode documentation when applicable.
- **Both harnesses get the shared file verbatim.** The Engram protocol reaches every harness through its own plugin — Codex included via the Engram Codex plugin (bundled SessionStart/UserPromptSubmit/Stop hooks).
- After changing the source, replicate it with:
  ```bash
  cp harness/AGENTS.md ~/.config/opencode/AGENTS.md
  cp harness/AGENTS.md ~/.codex/AGENTS.md
  ```
- Validate with:
  ```bash
  cmp -s harness/AGENTS.md ~/.config/opencode/AGENTS.md && echo "OpenCode MATCH"
  cmp -s harness/AGENTS.md ~/.codex/AGENTS.md && echo "Codex MATCH"
  wc -c ~/.codex/AGENTS.md  # context-size sanity only — NO harness caps the GLOBAL file
                            # (Codex, opencode, Grok: verified 2026-08-18 in source and
                            # reproduced empirically — see the bibliography). The one real
                            # cap, project_doc_max_bytes, covers only the repo chain
                            # (git root → cwd). On overflow Codex truncates the crossing file
                            # mid-content and silently drops deeper files — keep per-repo
                            # AGENTS.md chains small and audit them with:
                            #   find <repo> -name AGENTS.md -exec cat {} + | wc -c
  ```
- Start a new Codex/OpenCode session after replication. Do not rely on resumed sessions to reflect changed global instructions.
- **A Codex session opened under `harness/` loads the shared core twice** — once as the deployed global, once as a project doc in the chain (that sum exceeds the 32 KiB default cap; it fits only because `~/.codex/config.toml` raises `project_doc_max_bytes`). A session at the repo ROOT loads only the root file and is unaffected — never quote the under-`harness/` figure as the repo's current chain size. Codex offers no exclusion mechanism — `project_doc_fallback_filenames` adds names, never removes them. Open Codex at the repo root; `build.py` reports BOTH chain sizes — at the root and under `harness/` — on every run.

## Validation
- This is a configuration-only repo; validation is mostly diff review and consistency checks.
- **Shell is the exception — it has a real linter.** After touching any `.sh`, run `find global .claude harness -name '*.sh' -print0 | xargs -0 shellcheck -S style` and keep it at zero findings. A genuine false positive gets `# shellcheck disable=SC####` with the reason inline, placed **before the compound command** (`while`/`if`), never before its `done`/`fi` — a misplaced directive makes shellcheck skip the whole block instead of one line. `_support/backup|archive/**` is third-party or frozen; leave it out of scope.
- For changes to `AGENTS.md` or `harness/AGENTS.md`, verify the files do not reference Claude-only tools as if they were available in other harnesses.
- For changes to deploy behavior, verify the deploy skill still only targets `global/` unless the user explicitly requests a new deployment workflow.

<!-- hive-profile:start -->
Hive profile v1 · hive@9b341d5 · 2026-09-19 · class: config-hub

## Hive Profile

- **Class:** config-hub
- **Stack detected:** Python, shell scripts

**Rule families that apply here** (delivered by the rule-delivery hold on a matching write, or read from the language-rules references): `python-standards`, `shell-standards`.

**Rule Exclusions (class: config-hub — no runtime):**
- `testing.md` + `test-gate.md` — no runtime code to test; files here are reviewed by reading. The hold on `test-gate.md` still fires on a `.py`/`.sh` write — it is not suppressible; read it once and apply this exclusion, which outranks it here.
- `Build & Lint` (global `CLAUDE.md`) — no build system; validation is read-review/diff review.
- `security.md` > Supply Chain Security — no installable dependencies; no OSV checks to run.
- `critical-thinking.md` > Pre-ship ownership test (questions 1-2) — no runtime load or customer impact; risk-surfacing and tradeoffs still apply.

These exclusions are a declaration you READ, not a filter something applies: the gates above are inlined into the always-on core or delivered by the rule-delivery hook, and neither is suppressible file by file. Honor them here; they are restored automatically in any repo with runtime code.

**Creating the FIRST file of a kind in a session:** read its rule from `~/.claude/skills/language-rules/references/` first — the rule-delivery hook holds the write until you do, and nothing fires before the file exists (naming identifiers, choosing a path).
<!-- hive-profile:end -->
