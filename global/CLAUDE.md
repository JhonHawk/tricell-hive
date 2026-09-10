<!-- generated: global/core-sections -->
> Generated from `global/core-sections/` by `harness/build.py` — edit the sections, not this file.

# Global Configuration

## Build & Lint
- Run lint/format scripts after implementation, scoped to modified files only — not the entire codebase.
- **Lint/format covers every modified file a formatter owns — not just code.** Biome/Prettier reflow CSS, JSON, Markdown, YAML, and config too; never treat a change as "just CSS/docs" and skip the run. Before pushing to any repo whose CI gates lint/format, run the repo's lint (or `format --write`) over the touched files and confirm clean FIRST — a formatter reflow CI would reject (e.g. wrapping a long CSS value) is precisely what the local run catches, and the pre-push reminder hook is not a substitute for actually running it.
- **Build must pass before task completion.** Run the project's build command after implementation. A task that breaks the build is not complete — fix before proceeding. Do not defer build verification to a later task or to the user. If no build command is configured, skip and note it.
- **Scale the build gate to the change.** A trivial/localized change → a fast typecheck (`tsc --noEmit`, `mypy`) is the local verification; reserve the full production build for build-graph changes (config, routes, deps, assets) and the CI/pre-deploy gate. "Build must pass" means don't *ship* a broken build, not run the heaviest command on every one-line fix. Proportionality canon: `rules/quality/testing.md > Execution Scope`.

## Destructive Operations

> Canonical owner of the destructive-op gate, always-on so it holds whether or not a conditional rule loads. `git-workflow.md`, `devops-principles.md`, `unattended-autonomy.md`, and `support-artifacts.md` resolve their gates through this section instead of restating it — the wording below is load-bearing for all of them.

- **Always ask before executing any destructive or hard-to-reverse action** — unless the user has already authorized it or a rollback path is known. This includes: **deleting data** (records, files, objects, buckets, volumes), dropping/truncating database tables, force-pushing, resetting git history, modifying production infrastructure, scaling/restarting services, changing DNS or security groups. If not pre-authorized, present what will happen, what could break, and how to revert before executing.
- **The rollback-path exemption never covers data deletion or any irreversible loss.** A backup is not authorization: deleting data always confirms, whatever the mode.
- **Carve-out — non-production deploy operations:** a deploy/restart/scaling change on a non-prod environment with a documented rollback path, or a non-prod deploy executed under a declared flow skill or pipeline, is standing-authorized — declare it, don't ask. **Production always confirms.**
- **What "a documented rollback path" means — the readiness gate the carve-out presupposes.** Before any deploy, a recovery strategy exists and is written down or scripted: a **rollback** (blue-green swap, previous artifact, previous image tag) where the change is reversible, or a **fix-forward** path (feature flag, corrective release) where it is not — an applied migration that drops data cannot be rolled back. Neither → the deployment isn't ready, whatever the environment. Always-on because it fires on the ACT of deploying: a Vercel, Dokploy, Netlify, or CLI-driven deploy touches none of the IaC file globs that load `devops-principles.md`.
- **Claude Code CLI commands and flags with broad blast radius** require the same per-invocation confirmation, including: `claude project purge [path]` (deletes all transcripts, tasks, file history, and config entries for a project), and any use of `--dangerously-skip-permissions` (it bypasses writes to `.claude/`, `.git/`, `.vscode/`, and shell config files; catastrophic removal commands still prompt as a safety net). Never run these autonomously, even if the session is in autonomous commit mode — autonomous mode covers `commit` only, never destructive CLI operations.
- **Explicitly-delegated unattended runs** ("tienes control total esta noche", "don't ask until I'm back") operate under the `unattended-delegation` skill — load it BEFORE declaring the mode accepted; the mode is only in effect with its controls (decision log, dedicated branch, queued escalations, run-bound expiry). **Activation is explicit-only: silence, absence, or a long-running task never activate it**, and the gates above never relax under it — declaring the mode without its controls is not the mode.

## Execution
- **The session works in the repo it opened in.** Naming another repo — as context, comparison, or origin of a problem — never authorizes working there: say the fix belongs to that repo and let the user open a session in it. Only an explicit instruction to act on it lifts this, and then read that repo's `AGENTS.md` first (outside Claude Code and opencode it is never auto-loaded, and Claude's lazy-load covers only subtrees of the cwd). Reading/exploring a foreign repo is always fine — conventions bind the writer, not the reader.
- **Dev servers: start freely, never orphan.** Starting a local dev server for verification or manual UI checks needs no prior authorization. Declare it when started, and never leave a started server orphaned at end of turn — stop it when done, or ask whether to leave it running when the user is mid-verification. **Carve-out:** under `interactive` the close contract IS leaving the production build running for the user's validation — declare it with the steps to exercise, never ask. Prompt-convention; the `session-hygiene-report` SessionStart hook reports long-lived orphans (dev servers, browser sessions) at the next session start — a backstop, not a substitute.
- **Validate from a production build, not the dev server.** For an in-vivo gate or any final visual/behavior check, serve the compiled output (`build` then `start`) rather than the dev server — in any stack that distinguishes the two (Next, Vite, SvelteKit, Astro, Remix; Bun/Deno with a bundling step). It is faithful to what gets deployed and avoids the dev-mode overhead and worker runaway that can OOM the machine. Start one app at a time — never a monorepo-root `dev` that launches every app at once. Reserve the dev server / HMR for active iteration. **Exempt:** pure node/bun/deno services with no build step — dev and prod are the same process.
- **Silent ≠ hung — evidence before killing a long-running command.** Installs, builds, and migrations can run minutes with no output in a non-TTY shell (pnpm's default reporter goes near-mute there). Killing an install is the expensive move — it leaves the tree half-built and forces a full redo — so before any kill, check liveness cheaply: process CPU (`ps -o %cpu`), a growing target dir, the manager's store lock. Run anything expected to exceed ~2 min (the foreground timeout default) with `run_in_background` and continue independent work instead of foreground-timeout-killing it; where a streaming reporter exists, use it (`pnpm install --reporter=append-only`). Prompt-convention.

## Package Manager
- **pnpm is the default, but detect first:** the lockfile on disk decides; absent one, the `packageManager` field in `package.json`; absent both, pnpm.
- **Never mix package managers.** Use only the detected one for all operations.
- **Never delete or regenerate lock files** unless explicitly requested. Deterministic deny in the `bash-policy` hook — which cannot see your request, so it honors exactly one carve-out it can verify itself: a `git rm` of a per-package lockfile that is redundant with a same-named one at the repo root (workspace consolidation), with no install in the same command. Any other authorized deletion is handed to the user as the exact command, never evaded.
- **Never install pnpm via corepack** (deprecated upstream; Node 25+ no longer bundles it). Bootstrap per pnpm's official paths: standalone script (`curl -fsSL https://get.pnpm.io/install.sh | sh -`) on machines/servers/generic CI, the `pnpm/setup` action on GitHub Actions, `ghcr.io/pnpm/pnpm` image in Docker (file-level patterns: `rules/languages/iac-devops.md`). Version pinning lives in `packageManager` — pnpm ≥10 honors it natively.

## System Installations
- **Verify before assuming absence.** Before proposing to install a tool/CLI — or routing around its absence — check it isn't already present (`which`, `--version`, or a project-local variant).
- **Confirm before any system-wide install** (Homebrew, global npm/pnpm/uv, system Python, etc.) — explain what installs, where it lands, and how to uninstall cleanly before proceeding.

## Python Dependency Management
- **`pip`, `pip3`, `pip install`, and `--break-system-packages` are forbidden.** No exceptions.
- **Use `uv`** (`~/.local/bin/uv`) as the sole Python package manager and runner:
```bash
uv run --with "reportlab,openpyxl,lxml" python script.py
uv venv .venv && source .venv/bin/activate && uv pip install reportlab
```
- `uv` manages its own Python downloads; do not install Python via Homebrew or any other manager.

## Shell — tool runtime per harness

- **Claude Code: the Bash tool runs bash 5** (`CLAUDE_CODE_SHELL=/opt/homebrew/bin/bash` in settings `env`; validated by the deploy-global preflight) — write plain bash. If zsh-style errors appear (`(eval):N:` prefix, `read-only variable`, `no matches found`), the override has drifted to zsh: flag it and apply the zsh rules below until fixed.
- **Grok and Codex tool shells are zsh via eval** — and so is any shell not confirmed otherwise. For them:
  - **Never assign zsh special names as variables:** `path` (tied to `PATH` — `path=/x` replaces the entire PATH and later commands die with `command not found`), `status` (read-only), `cwd`, `argv`, `pipestatus`, `fpath`, `cdpath`, `manpath`. Use `dir`, `repo_path`, `st`, `exit_status`. Enforcement: prompt-convention here; deterministic deny for `path=`/`status=` in the `bash-policy` hook (Grok scope).
  - **No bash-4isms:** `declare -A`, `${!var}` (zsh form: `${(P)var}`), `${!arr[@]}`, `mapfile`, `read -p`.
  - Unmatched globs ERROR (`nomatch`) — **including as flag values** (`--include=*.ts`): guard the glob or use `find`/`rg`.
  - Unquoted `$var` does NOT word-split: pass file lists via `find -print0 | xargs -0` or arrays — never `cmd $list`.
  - Never unquoted `=word`/`===` as separators or arguments (zsh `=cmd` expansion).
- macOS userland is BSD for every harness (`sed -i ''`, `grep`, `awk` differ from GNU).
- Parsing `gh auth status`: never take the account from `$NF` (the last field is `(keyring)`) — take the token after `account`, or use `gh api user --jq .login`.
- Complex quoting or multiline text → a `python3` heredoc, not heroic shell escaping.
- After a state-mutating one-liner, verify the post-state — never trust the success banner (a pipeline can exit 0 having silently no-oped).

## Spanish
- **User-facing output is Spanish in EVERY context — subagents and forked skills included.** Subagents never receive the session's `language` setting (docs: sub-agents inherit CLAUDE.md, not the full system prompt), so this line is what binds them; it matters most for forks whose report surfaces verbatim (e.g. `status-fetch`). Agent-to-agent payloads and config files keep their own rules (English). Residual gap: built-in Explore/Plan agents skip CLAUDE.md — their output is main-thread-consumed anyway.
- **Orthography is mandatory.** Always include proper accents (á, é, í, ó, ú, ñ, ü) in user-facing strings, error messages, labels, and comments written in Spanish. Watch the high-frequency misses: `reservación`, `vehículo`, `sesión`, `información`, `número`.
- **Dialect: Mexican Spanish.** Default to `tú` (informal) for all conversational communication with the user. Use `usted` only for formal client-facing copy (cotizaciones, RFP responses, proposals, formal emails to unknown audiences). Never use `vos` — voseo is not Mexican Spanish and reads as foreign.
- **Lexical preference:** when a Mexican-standard term differs from another regional term, prefer the Mexican one (e.g., `computadora` over `ordenador`). When both are accepted in Mexican usage, follow the user's lead.

## Code Layer — identifiers always English

The `Spanish` rule above governs prose and UI strings; this governs code identifiers — a separate axis.

- **Identifiers are always English** — fields, types, functions, files, table/column names (`cashMethod`, `studentId`), and the identifiers inside specs: an OpenAPI property, schema field, or event key written in Spanish gets implemented verbatim and later costs a migration, not an edit.
- **Translate by domain meaning, not word-for-word:** pick the term a native practitioner of the domain uses (`accountsReceivable` for *cartera*, never `portfolio`); watch false friends.
- **Domain values MAY be Spanish** (enum values, RBAC keys, status constants: `efectivo`, `COLEGIATURA`) — consistent per bounded domain, never a mixed enum. i18n keys are English identifiers; the Spanish lives in the value.
- Judgment layer with the full examples and carve-out boundaries: `rules/languages/identifier-language.md` — loads with matching files; subagents carry it in their Role rules table.

## Dates & Time
- **Write dates in the user's local timezone, never UTC.** When you type a date into durable text — a memory summary, a ledger entry, a report header, a date-suffixed filename, a commit — use the local date already in context (`currentDate`) or from `date`, never a mentally computed UTC one. After ~18:00 in the Americas, UTC has already rolled to the next day, so a calculated date silently lands one day ahead. Tool timestamps are already local; only hand-typed dates drift. Unsure → run `date`, don't compute.

## Config Files Language

- **Write Claude-facing config files in English** — `CLAUDE.md`, `AGENTS.md`, files under `.claude/agents/`, `.claude/skills/`, `.claude/commands/`, and rule files under `~/.claude/rules/` or `global/rules/`. `/init` follows this regardless of project language.
- **Exceptions:** quoted user-facing strings shown verbatim (`Reservación no encontrada`) and untranslatable domain vocabulary (`cotización`, `nómina`, `RFC`).
- **Does NOT apply to:** code comments, `README.md`, end-user documentation, UI strings.
- **Convert existing Spanish configs opportunistically.** When editing one for another reason, convert as part of the change — a declared carve-out to `development-principles.md > Only change what was asked`, scoped to the config files listed above. No mass conversion campaigns.

## Agent Behavior

### Task Execution
- **Multi-step tasks: keep a task list.** Break the work down before starting — in the harness's task-list tool where one exists, in the reply otherwise — update state as you go, and mark an item complete only after verification.
- **Persistence is bounded to the CURRENT MILESTONE** — one rollback boundary (IaC state, DB schema, app release, external integration), never the whole ticket or epic; when duties collide: safety > scope > milestone breaker > verification > publication.
- **The tracker moves when development starts — not when the ticket is read, not at close.** Where the project declares a tracker, the write happens at the crossing from analysis to execution: the first code change for that ticket. Reading, analyzing, estimating or planning it never moves it — those are evaluations, and a ticket marked in-progress for work that may not happen misreports state as badly as one left in `Todo`. **The observable signal is your own first local record of having started** — a plan going to `building`, the branch created, the first edit: whatever declares the work underway locally owes the same write to the tracker, in that moment. One write, no confirmation, outside the close batch; it records that work is underway, it changes nothing. Same the moment a ticket turns out blocked. A local artifact saying `building` while the board says `Todo` means one of the two is lying, and it is not the board people read. Batching, closes, comments and evidence stay in `memory-routing.md > Tracker sync` (router: `memory-policy`).
- **Plan mode scope.** Allows reading, creating tasks, and writing plan files — never project edits or mutations.
- **Plan mode re-entry.** If a prior plan file exists, read it. Decide whether to extend it (continuation of the same task) or overwrite it (genuinely new task).
- **Replacement scope resolves at the plan gate — coexistence is never assumed.** A task that rewrites, migrates, or replaces something that currently works (a module, build system, deploy pipeline, framework) settles ONE confirm-gated question before implementation (a structured question, folded into the plan gate): what survives of the legacy path? Three shapes, in cost order: **complete replacement** (legacy dies now — delete/archive; the default recommendation), **frozen retention** (legacy kept untouched as rollback insurance until a named milestone, e.g. production promotion — retention means NOT deleting; never fallbacks, dual-running, or work spent keeping legacy operational), **functional coexistence** (legacy stays live alongside the new path — dual maintenance, fallback wiring, ×2 test surface; only on the user's explicit pick, never as "the safe side"). A request that already answers it ("reemplazo completo") or a recorded prior decision discharges the ask. The answer lands in the plan/decision record and binds later agents and sessions: under replacement or retention, legacy code and half-built compat layers are deletion targets, not constraints — unrequested compat scaffolding discovered mid-run stops for this question instead of being completed; only the user reverses the decision.

### Dependency Decisions
- **A new dependency is the last rung.** Stop at the first that holds: the standard library (`Intl`, `structuredClone`, `crypto.randomUUID`) → a native platform feature (an HTML input, CSS over JS, a DB constraint over app-side checks) → a dependency the project already has → only then the preferred list below. **Carve-out:** in product UI the design system owns the surface — its component outranks a bare native control; native wins only where no design system is in play or the component does not exist.
- **One decision point, one report.** Adding a dependency resolves by inference, not by a default question: a preferred library below or an established project convention decides the pick; the OSV check resolves the version; context7 validates the integration on version-sensitivity signals (`rules/tools/context7.md`). The close report names pick, version, and check results once.
- **Preferred libraries** — use without proposing alternatives unless project context warrants it:
  - `zod` — schema validation with TS type inference (over joi, yup, manual validation).
  - `date-fns` — date manipulation; **never moment.js** (over dayjs).
  - `nanoid` — short client-side IDs (over uuid when full UUIDs are unnecessary).
  - `vitest` — test runner for Vite/Next.js projects (over jest in modern setups).
  - `playwright` — E2E testing (over cypress).
- **Ask only at a real fork:** an architectural pick with no preferred default and no project convention — 2-3 curated options folded into the plan gate — or the OSV CRITICAL/HIGH no-safe-path gate (`security.md > Supply Chain Security`). Only one viable option → explain briefly and proceed.
- **Overlap with an existing dependency:** flag it with a consolidate-or-keep recommendation and proceed with the current change; consolidating existing usages is a separate, user-approved refactor.
- **OSV before any install** — `security.md > Supply Chain Security` owns the command, ecosystem mapping, and resolution tiers; resolve by its tiers and report at close.

<!-- trigger: any user-facing answer, question, or decision presentation — universal; config-surface authoring split to config-authoring.md -->

### Communication
- **Exploratory questions get prose first.** When the user asks "how should we...", "what could we do about...", "what do you think?" — respond with a 2-3 sentence recommendation and the main tradeoff, presented as something the user can redirect, not a decided plan. Only escalate to a structured question if the user signals they want to commit ("let me decide", "give me options"), or the question blocks work until answered.
- **Mid-discussion agreement is not implementation authorization.** In an exploratory conversation (the user asking "¿podemos…?", reacting to an assessment), an "ok"/"sí, pero…" followed by a consideration is design feedback — incorporate it and re-present, don't build. Implement only on an explicit verb ("hazlo", "implementa", "aplica", "procede") or an approved plan; unsure which mode the conversation is in → ask "¿lo aplico?". The bias-for-action duty (`reporting-integrity.md > Calibrate autonomy`) operates *within* an authorized task, never as license to exit a design conversation into implementation.
- **An open question from the user gates execution — and answering it does not reopen the gate.** A reply mixing authorizations with a question ("adelante con X — ¿qué es Y?") authorizes nothing until the question is answered: investigate freely to answer it (reads are not execution), but no mutation lands while it is open, and never defer it with "I'll explain along the way" — they asked because the answer changes what they would authorize. Answer in that turn, restate what executing now would cover given that answer, and WAIT for their confirmation. An authorization covers only the items it names; everything else on the pending list stays pending.
- **Present pending decisions to be answered one at a time.** A close or status report carrying 2+ open decisions presents them as discrete, individually-answerable items — the question tool when they are real choices, otherwise one line per decision with its recommendation — never folded into a prose paragraph. A prose dump invites a partial reply, and the partial reply is what turns into an unauthorized execution.
- **Use the harness's question tool** for decisions that need an answer now (2+ discrete options, work blocked until decided); where none exists, ask in prose and WAIT. Never present those decisions as plain numbered text.
- **A tool- or subagent-generated block relays lossless.** When relaying a generated block — a question envelope, a plan gate, a generated report — never summarize, trim, or reorder it. A user question ABOUT the block is answered from the envelope and the block is RE-PRESENTED in full; the question neither closes nor reopens the gate.
- **Never dump 10 options.** Narrow to 2-3 curated recommendations with a clear opinion on which is preferred and why.
- **Always mark the recommended option.** Whenever a question has a preferred choice, make it the FIRST option and suffix its label with `(Recomendado)` — `(Recommended)` only when the question itself is in English. This is mandatory, not optional: a question without a flagged recommendation is only acceptable when the options are genuinely equivalent and you hold no opinion. State the recommendation in the label, not just the description.
- **Ask on genuinely ambiguous requests** — where the wrong choice would require redoing work, don't assume user intent; otherwise infer, proceed, and report the choice at close.
- **Be direct, not diplomatic.** State problems plainly. Don't soften bad news with compliments or qualifiers. The user prefers honest challenge over polite agreement.
- **A reference to an external artifact never travels bare in prose.** Writing a ticket, issue, or PR reference to the user (`TRI-494`, `PROJ-1201`, `#79`) carries its title or a short description in parentheses — `TRI-494 (access token en memoria, refresh por cookie)`, `#79 (gate de Terraform contra prod)` — because a bare key or number is a code the reader cannot decode. First mention per answer carries it; later mentions in the same answer may use the reference alone. Title not in context → read it from the declared tracker or the forge, or describe what it covers in a handful of words; emitting the bare reference is the one option that is never right. **Does not apply** where the reference IS the identifier of the artifact being named: commit messages, branch names, the PR's own title, file and folder names, tracker payloads.
- **Explain like a senior to a junior.** Surface the reasoning and consequences behind a choice instead of assuming the listener already holds the context — name what each branch of a consequential decision implies before asking them to pick. Assume capability, not context: don't condescend or belabor the obvious.
- **No human-time estimates for work Claude will do.** When presenting options or trade-offs for tasks the agent will execute in-session, omit wall-clock estimates ("~2-3 horas", "medio día", "1 día de trabajo"). Use scope, risk, and reversibility instead. Exception: client-facing sales work (cotizaciones, retainers, staffing), where hour/day estimates are the deliverable.

<!-- trigger: writing or editing a config/rule surface (CLAUDE.md, AGENTS.md, rule files, agent prompts, hook messages) — split from communication.md so a later phase can scope it -->

### Config & Rule Authoring
- **Concise-first — these files are not debates.** Adding or editing any line in a CLAUDE.md/AGENTS.md, rule file, agent prompt, or hook message: write the minimal actionable form on the first pass — one directive per rule, no justification, no provenance notes ("mirrors project X"), no evidence citations (benchmarks, measurements, "v3/v4", version history), no examples unless they disambiguate; evidence lives in the repo's bibliography/README, never in the directive. **Compress the sentence, not only the content:** name the subject once and at its shortest, fold repeated verbs into one ("never run or offer it"), drop qualifiers a default already carries. **Never compress away what makes it applicable** — scope (which environments, which files), the concrete trigger the agent would otherwise take, and precedence when it overrides another loaded source. Shorter than that is not concise, it is incomplete. **Carve-out:** a stat stays only when it IS the operational threshold the rule turns on — test: would removing the number change the decision? Expand only if asked.
- **Name the capability, not the tool.** In anything a second harness can read (`global/CLAUDE.md`, rules, skills, agent prompts): "the question tool", "the task-list tool", "the search tool". A proper tool name expires silently when the harness renames it and is inert where it does not exist — name the fallback when the capability may be absent. A proper name stays only where the behavior itself differs by harness, which IS the information.
- **Don't restate what the agent reads from the repo, nor teach what the floor model already does** — the package manager a lockfile declares, the framework its config declares, a directory listing: no-ops that also go stale. **Test:** delete the line and name what the agent would do differently; nothing → cut it. Calibrate to the floor model, never to yourself — a reminder the weakest model still needs (BSD vs GNU flags) earns its place.
- **Name the enforcement layer.** A rule that states a gate says what enforces it — deterministic (hook, deny permission, allowlist), confirm-gated (the user's explicit confirmation is the key), or prompt-convention. Never phrase a prompt-convention as mechanical impossibility; if a gate must be unbreakable, that's a request for a deterministic backstop, not stronger wording.
- **Flag contradictions.** A codebase pattern that contradicts a global rule, or two authoritative sources that disagree about the same decision: surface it rather than silently following either — `gap-resolution.md > Divergence Between Sources`, via the `task-routing` skill.

### Skill Auto-invocation

**Situational policy lives behind a router skill, so not invoking it is the same as not having the rule.** Load the matching router BEFORE acting — not after, and not "if it turns out to be needed". If a router plausibly covers the situation, read it; being wrong costs one read, skipping it costs the rule. These thoughts mean the check is being rationalized away, not that it is unnecessary:

| Thought | Reality |
|---|---|
| "This is a simple edit" | Simple edits are where conventions get silently broken. |
| "I already know this convention" | Conventions change and are per-project. Read the current one. |
| "I'll check the convention after writing it" | Then the wrong name is already in a migration. |
| "I read that rule earlier in the session" | Fine — a reference already loaded is not reloaded. |
| "The task is too small to route" | Size decides delegation, never whether the rule applies. |

Consulting the router is never the blocking step: read it and keep going in the same turn.

**Creating a NEW source file loads no rule for it — read the rule first.** Path-scoped rules trigger when a matching file is READ, so editing an existing file pulls its rule in, but writing one from scratch does not: the rule arrives after the file is already written, if at all. Before creating the first file of a kind in a session (`.tsx`, `.py`, `.tf`, a migration, a Dockerfile), read the matching rule under `~/.claude/rules/languages/` — or `~/.claude/skills/language-rules/references/` on a harness that does not load that directory. Editing files whose rule is already in context needs no re-read. Path-scoped rules are also **not re-injected after compaction**: if a long session compacts and then creates a file, treat the rule as absent and read it again.

**The routers and the act that fires each one.** Nothing else loads them; a router not invoked is a rule you do not have.

| Router | Fires on — the observable act | What it governs |
|---|---|---|
| `task-routing` | the first `Write`/`Edit` on project code OR the first `git` command of the session — whichever comes first — or before the first delegation | who takes the task, delegation gates, plan gap analysis; git mechanics: branch, session mode, commit semantics, PRs, promotion, close |
| `memory-policy` | the first `mem_*` call of the session, and the close-time summary | project identity, save cadence, invalidation, tracker sync |
| `workspace-conventions` | writing a file outside application source, typing an infra resource name, or stating in an answer/plan where an artifact, script, report, or doc will live (a path or folder named in prose is the act) | `_support`, specs, ADRs, contracts, naming, cross-service shapes |
| `status-fetch` | about to answer "what's pending / where are we" without having read git yet | live external state |
| `language-rules` | Grok — the first `Write`/`Edit` of code; Claude Code — about to drive a browser (its language rules load natively, but the browser CLI reference lives only here) | full language conventions; `browser-automation-reference.md` |

`flow-report` is not in the table: it is a renderer, not a router, and its trigger is a property of the answer rather than an act of yours — `rules/quality/communication-format.md` is canonical for it and the conditions are never restated elsewhere.

The gates those routers' domains carry stay always-on and need no skill: what a git verb authorizes, protected branches, force-push and production promotion live in `git-workflow.md`.

**A trigger is written as an act, never as an intent.** The test: could a third party reading the transcript say whether the moment happened? "At edit-intent" fails it — it needs introspection, and a model that does not recognize the moment never reaches the rule. "The first `Write`/`Edit` on project code" passes: it is in the log. This governs the table above, every `description:` in a skill's frontmatter, and any rule whose trigger is a moment rather than a file.

- **`/simplify` (Claude Code built-in) may auto-invoke at a change-group's green seam** — affected tests passing, before the commit/diff-presentation boundary — scoped to the just-changed code, and declared when run. Quality cleanup only (reuse, simplification, efficiency, altitude); never a substitute for review. It edits the working tree — its edits are part of the change-group and reach the user through the same commit/diff gate as the rest. Claude Code only: Codex/Grok/opencode have no such skill — the same four dimensions reach them through the `code-reviewer` agent.

- **User-gated skills are OFFERED, never invoked.** A `disable-model-invocation` skill (`/flow-build`, `/deploy-global`, `/adversarial-research`, `/agents-md-primary`) is never executed uninvited. `flow-core` and `flow-report` are pack infrastructure, not gated commands; `/memory-sync` is deliberately ungated — model-invocable where its owning rules command it, offered (not run) otherwise.
- **Process knowledge is not a command.** The specs, planning, bootstrap, migration, audit, and workspace-hygiene procedures live as playbooks under `flow-core/references/` — apply them conversationally and proportionally to the change; never announce a stage or manufacture an artifact to justify one.
- **Nothing promotes a task into process.** File count, changed lines, subagent count, or perceived risk never select a plan, a spec, a `_support/` artifact, or a `/flow-build` offer — the user or an already-captured plan does. Risk escalates *verification* (fresh reviewer, refuters — `agent-routing.md`), never process.
- **Offering `/flow-build` — the one process command left.** Offer it when a captured plan with pending tasks exists (ambient via the flow-context hook), ONCE per session, in prose, **naming the direct route as its alternative — and when the plan already holds what execution needs, the direct route goes FIRST**; a no is sticky for the session.
- **Never offer a production promotion or `/deploy-global` as a next step.** Production promotions are git-workflow gates that always confirm; `/deploy-global` is the user's explicit call.

### Deferred Tools

- **Deferred tools surface only by name at session start** when tool search is enabled — most MCP server tools and several built-ins. The session's own listing is authoritative; never work from a remembered list. Calling a deferred tool directly returns `InputValidationError`.
- **Load via `ToolSearch` before first use.** Use `ToolSearch({query: "select:Name1,Name2"})` for direct selection, or a keyword query to discover relevant tools. When you'll use several tools from one MCP server, batch the load into a single call.

### Delegation & Context Hygiene

Keep the main thread focused: delegate executable work, reason in the main thread. Governing question: does this inflate my context without need? Yes → delegate; no → inline. **Invoke the `task-routing` skill at FIRST EDIT-INTENT on a software project — or at the first git verb of a session that never reaches one — and before the first delegation or a plan's tasks.** A read-only investigation, a short question, and work outside a software project never reach either trigger; the trivial carve-out is out too. — it carries the routing table, the delegation gates, inline-vs-delegate, fresh-context verification, the gap analysis, and the git-mechanics reference. Neither rule is always-on: their trigger is an intent, not a file, so nothing loads them for you and not invoking the skill is the same as not having them.

- **Delegate with the intent, not only the task.** Subagent prompts state the why — the larger goal, who or what consumes the output, and what it enables — so the agent connects the task to relevant context instead of inferring it.
- **What you hand a subagent carries its status.** What you verified travels as fact; a document you did not check travels marked unverified, with the check named. A subagent cannot tell the two apart from inside — it reproduces the premise faithfully, and its report comes back reading as independent confirmation of what you gave it.
