# Global Configuration

## Build & Lint
- Run lint/format scripts after implementation, scoped to modified files only — not the entire codebase.
- **Lint/format covers every modified file a formatter owns — not just code.** Biome/Prettier reflow CSS, JSON, Markdown, YAML, and config too; never treat a change as "just CSS/docs" and skip the run. Before pushing to any repo whose CI gates lint/format, run the repo's lint (or `format --write`) over the touched files and confirm clean FIRST — a formatter reflow CI would reject (e.g. wrapping a long CSS value) is precisely what the local run catches, and the pre-push reminder hook is not a substitute for actually running it.
- **Build must pass before task completion.** Run the project's build command after implementation. A task that breaks the build is not complete — fix before proceeding. Do not defer build verification to a later task or to the user. If no build command is configured, skip and note it.
- **Scale the build gate to the change.** A trivial/localized change → a fast typecheck (`tsc --noEmit`, `mypy`) is the local verification; reserve the full production build for build-graph changes (config, routes, deps, assets) and the CI/pre-deploy gate. "Build must pass" means don't *ship* a broken build, not run the heaviest command on every one-line fix. Proportionality canon: `rules/quality/testing.md > Execution Scope`.

## Destructive Operations
- **Always ask before executing any destructive or hard-to-reverse action** — unless the user has already authorized it or a rollback path is known. This includes: dropping/truncating database tables, force-pushing, resetting git history, modifying production infrastructure, scaling/restarting services, changing DNS or security groups. If not pre-authorized, present what will happen, what could break, and how to revert before executing.
- **Carve-out — non-production deploy operations:** a deploy/restart/scaling change on a non-prod environment with a documented rollback path, or a non-prod deploy executed under a declared flow skill or pipeline, is standing-authorized — declare it, don't ask. **Production always confirms.**
- **Claude Code CLI commands and flags with broad blast radius** require the same per-invocation confirmation, including: `claude project purge [path]` (deletes all transcripts, tasks, file history, and config entries for a project — added in 2.1.126), and any use of `--dangerously-skip-permissions` (since 2.1.126 it bypasses writes to `.claude/`, `.git/`, `.vscode/`, and shell config files; catastrophic removal commands still prompt as a safety net). Never run these autonomously, even if the session is in autonomous commit mode — autonomous mode covers `commit` only, never destructive CLI operations.
- **Explicitly-delegated unattended runs** ("tienes control total esta noche", "don't ask until I'm back") follow `rules/workflow/unattended-autonomy.md`: proceed-and-log on reversible in-scope work, identical absolute gates, gated decisions queued for review, all work on a dedicated reversible branch, run-bound expiry revoked by the user's first message.

## Execution
- **Dev servers: start freely, never orphan.** Starting a local dev server for verification or manual UI checks needs no prior authorization. Declare it when started, and never leave a started server orphaned at end of turn — stop it when done, or ask whether to leave it running when the user is mid-verification.
- **Validate from a production build, not the dev server.** For an in-vivo gate or any final visual/behavior check, serve the compiled output (`build` then `start`) rather than the dev server — in any stack that distinguishes the two (Next, Vite, SvelteKit, Astro, Remix; Bun/Deno with a bundling step). It is faithful to what gets deployed and avoids the dev-mode overhead and worker runaway (e.g. `next dev --turbopack` spawning an unbounded PostCSS worker per CSS module) that can OOM the machine. Start one app at a time — never a monorepo-root `dev` that launches every app at once. Reserve the dev server / HMR for active iteration. **Exempt:** pure node/bun/deno services with no build step — dev and prod are the same process, so there is nothing to compile.

## Package Manager
- **pnpm is the default.** For new projects or when no lock file exists, use pnpm.
- **Detect before installing:** check for `pnpm-lock.yaml` → pnpm, `yarn.lock` → yarn, `package-lock.json` → npm. Fallback: `packageManager` field in `package.json`. If nothing is detected, use pnpm.
- **Never mix package managers.** Use only the detected one for all operations.
- **Never delete or regenerate lock files** unless explicitly requested.

## System Installations
- **Verify before assuming absence.** Before proposing to install a tool/CLI or routing around its absence, check it isn't already present (`which`, `--version`, or a project-local variant). Don't suggest an install — or an alternative — for something that's already there.
- **Confirm before any system-wide install** (Homebrew, global npm/pnpm/uv, system Python, etc.) — explain what installs, where it lands, and how to uninstall cleanly before proceeding.

## Python Dependency Management
- **`pip`, `pip3`, `pip install`, and `--break-system-packages` are forbidden.** No exceptions.
- **Use `uv`** (`~/.local/bin/uv`) as the sole Python package manager and runner:
```bash
uv run --with "reportlab,openpyxl,lxml" python script.py
uv venv .venv && source .venv/bin/activate && uv pip install reportlab
uv tool install <package>
```
- `uv` manages its own Python downloads; do not install Python via Homebrew or any other manager.

## Shell (zsh on macOS)
- The interactive shell is zsh on macOS — never assume bash semantics or GNU userland (BSD `sed -i ''`, `grep`, `awk` differ).
- Unquoted `$var` does NOT word-split: pass file lists via `find -print0 | xargs -0` or arrays — never `cmd $list`.
- Unmatched globs ERROR (`nomatch`): use `find` for cleanups, or guard the glob.
- Never unquoted `=word`/`===` as separators or arguments (zsh `=cmd` expansion).
- Complex quoting or multiline text → a `python3` heredoc, not heroic shell escaping.
- After a state-mutating one-liner, verify the post-state — never trust the success banner (a zsh pipeline can exit 0 having silently no-oped).

## Spanish
- **Orthography is mandatory.** Always include proper accents (á, é, í, ó, ú, ñ, ü) in user-facing strings, error messages, labels, and comments written in Spanish. Common mistakes to avoid: `reservacion` → `reservación`, `vehiculo` → `vehículo`, `sesion` → `sesión`, `informacion` → `información`, `numero` → `número`.
- **Dialect: Mexican Spanish.** Default to `tú` (informal) for all conversational communication with the user. Use `usted` only for formal client-facing copy (cotizaciones, RFP responses, proposals, formal emails to unknown audiences). Never use `vos` — voseo is not Mexican Spanish and reads as foreign.
- **Lexical preference:** when a Mexican-standard term differs from another regional term, prefer the Mexican one (e.g., `computadora` over `ordenador`). When both are accepted in Mexican usage, follow the user's lead.

## Code Layer — identifiers always English

The `Spanish` rule above governs prose and UI strings; this governs code identifiers — a separate axis.

- **Identifiers are always English.** Fields, types, parameters, functions, classes, files, table/column names: `cashMethod`, `studentId`, `amount`. Never a Spanish-named identifier, whatever the project's language.
- **Translate by domain meaning, not word-for-word.** Choosing the English identifier is a domain decision, not a dictionary lookup. Pick the term a native-speaking practitioner of that domain uses — `accountsReceivable` not `portfolio` for *cartera*, `outstandingBalance`/`amountDue` not `debt` for *adeudo*, `issuedAt` not `emittedAt` for *emitido*. Watch for false friends: a word that exists in English but means something else in this domain (`inhábil` → `nonWorkingDay`, never `disabled`; documents are `issued`, never `emitted`). A literal 1:1 translation passes the English-only rule yet produces identifiers a domain developer wouldn't recognize — the rule above governs the *language*, this one the *terminology*.
- **Exceptions — domain values MAY be Spanish.** Enum values, RBAC/permission keys, status constants encoding business vocabulary with no clean English equivalent: `efectivo`, `COLEGIATURA`, `condonación`, `finanzas.operacion.caja`.
  - **Per bounded domain, not per value.** Once a domain's enums are Spanish, every new value stays Spanish. A **mixed enum** (English + Spanish in one domain) is the anti-pattern — consistency with the domain outranks a technically-valid alternative. Migrating a Spanish domain to English is a deliberate migration, never a side effect of adding one value.
- **Specs define identifiers too — English at spec-writing time.** An OpenAPI `path`/`property`, schema field, table/column/FK name, or event payload key in a spec is a code-layer identifier and must be English even when the prose is Spanish. Catch it there: a Spanish identifier slipped through a spec gets implemented verbatim and persists into a column/migration before anyone notices — then it costs a migration, not an edit. (A Spanish *value* in a spec follows the per-bounded-domain rule — not a violation for being Spanish.)
- **i18n/message keys are identifiers, not domain values.** A translation key (`t("session.expiringWarning")`) is English, semantic, and hierarchical; the Spanish lives in the value (the copy), never in a key derived from it. The per-bounded-domain carve-out covers business-vocabulary values (enums, RBAC), not UI-string keys.

## Dates & Time
- **Write dates in the user's local timezone, never UTC.** When you type a date into durable text — a memory summary, a ledger entry, a report header, a date-suffixed filename, a commit — use the local date already in context (`currentDate`) or from `date`, never a mentally computed UTC one. After ~18:00 in the Americas, UTC has already rolled to the next day, so a calculated date silently lands one day ahead. The system clock and tool timestamps (`date`, Engram's `Created:`) are already local and correct — only hand-typed dates drift. Unsure → run `date`, don't compute.

## Config Files Language

- **Write Claude-facing config files in English** — `CLAUDE.md`, `AGENTS.md`, files under `.claude/agents/`, `.claude/skills/`, `.claude/commands/`, and rule files under `~/.claude/rules/` or `global/rules/`. `/init` follows this regardless of project language.
- **Exceptions:** quoted user-facing strings shown verbatim (`Reservación no encontrada`) and untranslatable domain vocabulary (`cotización`, `nómina`, `RFC`).
- **Does NOT apply to:** code comments, `README.md`, end-user documentation, UI strings.
- **Convert existing Spanish configs opportunistically.** When editing one for another reason, convert as part of the change. No mass conversion campaigns.

## Agent Behavior

### Task Execution
- **Multi-step tasks: create a task list before starting.** Break down the work with TaskCreate, update status during execution, and only mark tasks complete after verification.
- **Plan mode scope.** Allows reading, creating tasks, and writing plan files — never project edits or mutations.
- **Plan mode re-entry.** If a prior plan file exists, read it. Decide whether to extend it (continuation of the same task) or overwrite it (genuinely new task).

### Dependency Decisions
- **One decision point, one report.** Adding a dependency resolves by inference, not by a default question: a preferred library below or an established project convention decides the pick; the OSV check resolves the version; context7 validates the integration on version-sensitivity signals (`rules/tools/context7.md`). The close report names pick, version, and check results once.
- **Preferred libraries** — use without proposing alternatives unless project context warrants it:
  - `zod` — schema validation with TS type inference (over joi, yup, manual validation).
  - `date-fns` — date manipulation; **never moment.js** (over dayjs, for tree-shakeability).
  - `nanoid` — short client-side IDs (over uuid when full UUIDs are unnecessary).
  - `vitest` — test runner for Vite/Next.js projects (over jest in modern setups).
  - `playwright` — E2E testing (over cypress, for CI reliability).
- **Ask only at a real fork:** an architectural pick with no preferred default and no project convention — 2-3 curated options folded into the plan gate — or the OSV CRITICAL/HIGH gate below. Only one viable option → explain briefly and proceed.
- **Overlap with an existing dependency:** flag it with a consolidate-or-keep recommendation and proceed with the current change; consolidating existing usages is a separate, user-approved refactor.
- **OSV before any install** — `security.md > Supply Chain Security` owns the command, ecosystem mapping, and resolution tiers; resolve by its tiers and report at close. **CRITICAL/HIGH with no safe path → explicit user confirmation, always.**

### Communication
- **Exploratory questions get prose first.** When the user asks "how should we...", "what could we do about...", "what do you think?" — respond with a 2-3 sentence recommendation and the main tradeoff, presented as something the user can redirect, not a decided plan. Only escalate to `AskUserQuestion` if the user signals they want to commit ("let me decide", "give me options"), or the question blocks work until answered.
- **Mid-discussion agreement is not implementation authorization.** In an exploratory conversation (the user asking "¿podemos…?", reacting to an assessment), an "ok"/"sí, pero…" followed by a consideration is design feedback — incorporate it and re-present, don't build. Implement only on an explicit verb ("hazlo", "implementa", "aplica", "procede") or an approved plan; unsure which mode the conversation is in → ask "¿lo aplico?". The bias-for-action duty (`development-principles.md > Calibrate autonomy`) operates *within* an authorized task, never as license to exit a design conversation into implementation.
- **Use `AskUserQuestion`** for decisions that need an answer now (2+ discrete options, work blocked until decided). Never present those decisions as plain numbered text.
- **Never dump 10 options.** Narrow to 2-3 curated recommendations with a clear opinion on which is preferred and why.
- **Always mark the recommended option.** Whenever an `AskUserQuestion` call has a preferred choice, make it the FIRST option and suffix its label with `(Recomendado)` — `(Recommended)` only when the question itself is in English. This is mandatory, not optional: a question without a flagged recommendation is only acceptable when the options are genuinely equivalent and you hold no opinion. State the recommendation in the label, not just the description.
- **`AskUserQuestion` on genuinely ambiguous requests** — where the wrong choice would require redoing work, don't assume user intent; otherwise infer, proceed, and report the choice at close.
- **Be direct, not diplomatic.** State problems plainly. Don't soften bad news with compliments or qualifiers. The user prefers honest challenge over polite agreement.
- **Explain like a senior to a junior.** Surface the reasoning and consequences behind a choice instead of assuming the listener already holds the context — name what each branch of a consequential decision implies before asking them to pick. Assume capability, not context: don't condescend or belabor the obvious.
- **Concise-first when writing rules — rules are not debates.** Adding or editing a rule in any CLAUDE.md/AGENTS.md, rule file, agent prompt, or hook message: write the minimal actionable form on the first pass — one directive per rule, no justification, no provenance notes ("mirrors project X"), no evidence citations (benchmarks, measurements, "v3/v4", version history), no examples unless they disambiguate. The consuming agent has no access to the evidence and the citation only inflates context; evidence lives in the repo's bibliography/README, never in the directive. **Carve-out:** a stat stays only when it IS the operational threshold the rule turns on — test: would removing the number change the decision? Expand only if asked.
- **Name the enforcement layer.** A rule that states a gate says what enforces it — deterministic (hook, deny permission, allowlist), confirm-gated (the user's explicit confirmation is the key), or prompt-convention. Never phrase a prompt-convention as mechanical impossibility; if a gate must be unbreakable, that's a request for a deterministic backstop, not stronger wording.
- **Flag contradictions.** If a codebase pattern contradicts a global rule, OR two authoritative sources (spec, ticket, wireframe, code, docs, tests) disagree about the same decision, flag the contradiction rather than silently following either. See `rules/workflow/gap-resolution.md > Divergence Between Sources` for the resolution protocol.
- **No human-time estimates for work Claude will do.** When presenting options or trade-offs for tasks the agent will execute in-session, omit wall-clock estimates ("~2-3 horas", "medio día", "1 día de trabajo"). Use scope, risk, and reversibility instead. Exception: client-facing sales work (cotizaciones, retainers, staffing), where hour/day estimates are the deliverable.
- **Quality over token thrift.** Prefer the best result over the cheapest path; spend more when it materially improves the outcome, hold back only when the cost buys nothing or a marginal gain. Governs process effort, not product scope. Detail: `rules/quality/critical-thinking.md`.

### Skill Auto-invocation

- **Design skills may auto-invoke** when the user's intent clearly matches a skill's own description. Many independent skills cover this space (`/polish`, `/critique`, `/audit`, `/bolder`, `/quieter`, `/colorize`, `/distill`, `/delight`, `/normalize`, `/arrange`, `/typeset`, `/optimize`, `/extract`, `/overdrive`, `/shape`, `/onboard`, plus the broader `/impeccable`). Each routes on its own description — no separate intent-to-skill table is maintained.

- **`flow-report` skill auto-invokes** when `rules/quality/communication-format.md` trigger conditions are met. That rule is the canonical source for the trigger — don't restate the conditions elsewhere.

- **User-gated skills are OFFERED, never invoked.** A `disable-model-invocation` skill (the flow pack, deploys, maintenance commands) is never executed uninvited. In a flow workspace, when the ledger's `Current phase` / `Next suggested` state (ambient via the flow-context hook) matches the conversation, offer the corresponding `/flow-*` command in prose ("plan approved — run `/flow-build`?") and let the user run it. One exception in the offer *mechanics* — `flow-brainstorming`: on clear feature/business-idea exploration intent, ask ONCE via a yes/no question ("¿Iniciamos modo brainstorming?"); yes invokes it, a No is sticky for the rest of the session (manual invocation only, never re-ask). Prompt-convention, carried by the flow-session-context hook block. Never offer a production promotion or `/deploy-global` as an automatic next step — promotions to production are git-workflow gates that always confirm, and `/deploy-global` is always the user's explicit call.

### Deferred Tools

- **Deferred tools surface only by name at session start** when tool search is enabled — this includes most MCP server tools and several built-ins (e.g., `LSP`, `WebFetch`, `WebSearch`, `Monitor`, `NotebookEdit`, `computer-use`, `playwright`, `chrome-devtools`). Calling a deferred tool directly returns `InputValidationError`.
- **Load via `ToolSearch` before first use.** Use `ToolSearch({query: "select:Name1,Name2"})` for direct selection, or a keyword query to discover relevant tools. When you'll use several tools from one MCP server, batch the load into a single call.

### Delegation & Context Hygiene

Keep the main thread focused: delegate executable work, reason in the main thread. Governing question: does this inflate my context without need? Yes → delegate; no → inline. Gates and the inline-vs-delegate table: `rules/workflow/agent-routing.md > Delegation Gates`.

- **Delegate to subagent `Explore`** when you need 3+ search queries across the codebase, or the question is open-ended ("how does X work?", "where is Y used?"). Explore returns a summarized report; tool-use noise stays out of the main thread.
- **Delegate to specialized agents** per `rules/workflow/agent-routing.md` disambiguation table before handling the task yourself in the main thread. If the task falls clearly in a domain (security review, DB schema, frontend component), the specialist is preferred — both for quality and for context hygiene.
- **Parallelize independent subagent calls.** When 2+ queries have no data dependency between them, dispatch in a single message with multiple `Agent` tool uses. Sequential dispatch of independent work wastes wall-clock and main-thread turns.
- **Delegate with the intent, not only the task.** Subagent prompts state the why — the larger goal, who or what consumes the output, and what it enables — so the agent connects the task to relevant context instead of inferring it.
- **Never relaunch what you already delegated.** Before spawning, check no equivalent subagent is already running or answered — wait for its result. Hygiene detail: `rules/workflow/agent-routing.md > Delegation Gates`.
- **Default to subagents; escalate only on a clear signal.** Agent Teams when the work needs persistent parallel workers or the user asks; the Workflow tool only on explicit opt-in — never on inferred intent. The opt-in list, disambiguation, and the `agentType` rule: `rules/workflow/agent-routing.md > Workflow Tool vs Subagents vs Agent Teams`.

<!-- CODEGRAPH_START -->
## CodeGraph

In repositories indexed by CodeGraph (a `.codegraph/` directory exists at the repo root), reach for it BEFORE grep/find or reading files when you need to understand or locate code:

- **MCP tool** (when available): `codegraph_explore` answers most code questions in one call — the relevant symbols' verbatim source plus the call paths between them, including dynamic-dispatch hops grep can't follow. Name a file or symbol in the query to read its current line-numbered source. If it's listed but deferred, load it by name via tool search.
- **Shell** (always works): `codegraph explore "<symbol names or question>"` prints the same output.

If there is no `.codegraph/` directory, skip CodeGraph entirely — indexing is the user's decision.
<!-- CODEGRAPH_END -->

The managed block above is owned by `codegraph install` (markers kept so upgrades report "Unchanged" instead of appending a duplicate). Hive specifics on top of it:

- **Earn it by shape:** the payoff is round-trips — relations or multi-file context in one call. A one-shot question a single grep or Read answers does NOT earn it. Bonus signal from `codegraph_explore`: it marks affected symbols with no covering tests. Multi-repo workspace: pass `projectPath` to the child repo (the root has no index).
- **Specialized CLI commands when you already know the target** (shell only — not exposed as MCP): `codegraph query <name>` to locate a symbol, `node <name>` to read one symbol's source + caller/callee trail, `callers`/`callees <symbol>` for direct relations, `impact <symbol>` for blast radius before touching shared UI/services, `affected --stdin --depth 2 --json` to pick the tests a changed file hits (feeds the `testing.md > Execution Scope` test selection), `files --filter <dir>` for an indexed-area inventory (no positional args). Multi-repo workspace: query commands take `-p <repo>`; maintenance commands (`index`, `sync`, `status`) take the path positionally instead (`codegraph status <repo>`) — no `--path` there.
- **`rg` still wins on literal, exhaustive textual work** — confirming a string exists, counting usages, full-coverage audits. The split is *coverage, not speed*: warm MCP latency ≈ `rg`. CodeGraph for relevance and relations; `rg` for exact text and total coverage. Unknown-name bug: CodeGraph to orient, then `rg` to confirm the pattern.
- **Skip CodeGraph in legacy/untyped JS repos** — its anchors mislead without types; route those to jbcontext/rg per `rules/tools/code-search.md`, which owns the full code-search routing policy and the anti-conclusion discipline.
- **Name exact symbols/files when you know them** — a broad query can return the high-level caller instead of the detail you meant; duplicate names → disambiguate with a file-specific query or `node -f`.
