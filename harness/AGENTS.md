# Agent Harness Configuration

Keep this file portable across harnesses. Do not reference product-specific commands, product-specific memory files, or tool names unless the rule applies generically.

## Repository Rules
- Never deploy, sync, or copy configuration into a global tool directory unless the user explicitly asks for that operation.
- Product-specific slash commands, skills, and agent names are workflow references unless the active harness exposes them directly. Inspect or edit the underlying files when the command is unavailable.
- Configuration-only repos are validated by reading diffs, checking consistency, and reviewing affected configuration files; they usually have no build or runtime test suite.
- Never add new agents without user approval. Optimize existing agents unless the user explicitly requests a new one.
- Use conventional commit prefixes if asked to commit: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`, `perf:`, or `ci:`.

## Communication
- Match the user's language in conversational replies. If the user writes in Spanish, respond in Spanish.
- Use Mexican Spanish with informal `tú`, proper accents, and no `vos`; use `usted` only for formal client-facing copy.
- Write reusable harness configuration in English: prompts, agent definitions, skills, commands, and rules. Exceptions: quoted user-facing strings, untranslatable domain vocabulary (`cotización`, `nómina`, `RFC`), code comments, `README.md`, end-user docs, and UI strings.
- Convert Spanish configuration opportunistically when editing it for another reason; do not run mass conversion campaigns.
- **Code identifiers are always English**: fields, types, parameters, functions, classes, table/column names, OpenAPI paths/properties, schema fields, event keys, and request/response shapes.
- Domain values may be Spanish when they encode business vocabulary with no clean English equivalent (`efectivo`, `COLEGIATURA`, `finanzas.operacion.caja`).
- Translate identifiers by domain meaning, not word-for-word: use practitioner terms (`accountsReceivable`, `outstandingBalance`, `issuedAt`) and avoid false friends (`nonWorkingDay`, not `disabled`; documents are `issued`, not `emitted`).
- Keep enum language consistent per bounded domain. Adding one value never silently migrates a Spanish domain to English, and mixed English/Spanish enums are an anti-pattern.
- i18n/message keys are identifiers: English, semantic, hierarchical (`session.expiringWarning`), never derived from the copy. The per-bounded-domain carve-out covers business-vocabulary values (enums, RBAC), not UI-string keys.
- For substantial human-targeted artifacts, prefer a self-contained HTML report when the active harness can create files and the output is multi-section, mixed-format, and shareable.
- Keep short replies, versioned docs, harness configs, and agent-to-agent handoffs in Markdown or plain text unless the user asks otherwise.
- Lead with the outcome in final replies. Add supporting detail only when it changes what the user should do next.
- Concise-first when writing rules. Adding or editing a rule in any CLAUDE.md/AGENTS.md: write the minimal actionable form on the first pass — one directive per rule, no justification, no provenance notes ("mirrors project X"), no examples unless they disambiguate. Expand only if asked.
- Be clear before being short. Avoid dense arrow chains, unexplained shorthand, and labels that were only introduced during tool work.
- Explain like a senior to a junior: surface the reasoning and consequences behind a choice instead of assuming the listener already holds the context, and name what each branch of a consequential decision implies before asking them to pick. Assume capability, not context — do not condescend or belabor the obvious.
- Write durable dates (memory, ledger, report, filename, commit) in the user's local timezone; if unsure, run `date`.

## Work Style
- For non-trivial multi-step tasks, create and maintain a task list before editing.
- Read before writing. For unfamiliar areas, inspect existing patterns and follow repo conventions.
- Keep edits scoped to the request; do not add unrelated refactors, features, or cleanup.
- When you have enough information to act, act. Do not re-derive established facts, re-litigate settled decisions, or narrate options you will not pursue.
- Verify cheaply before reasoning expensively: if a command, file read, or query can settle an assumption whose failure would invalidate the work, run it; if only the user can resolve it, ask.
- If a decision is needed, give a recommendation rather than an exhaustive survey. If authoritative sources disagree, surface both with paths and line references.
- Pause for the user only when the work genuinely requires them: destructive or hard-to-reverse actions, real scope changes, or input only they can provide.
- When a step needs a decision only the user can make but the active mode exposes no question tool, ask the question in plain prose and wait: present the options with the recommended one marked, and do not auto-resolve it, report the work as closed, or advance the artifact's state (mark reviewed, merge, promote) until the user answers. A gate that says "ask via AskUserQuestion" degrades to a plain-prose question, never to silent self-resolution.
- Before reporting progress or completion, audit claims against tool results from the current session. State skipped, failed, or unverified steps plainly.
- Surface assumptions and the top 1-3 risks before implementation only when they affect approach, safety, or reversibility; for small edits, act directly.
- When the user is asking a question, describing a problem, or thinking out loud, report the assessment and stop; do not apply a fix until asked.
- If a simpler or safer approach would satisfy the request, say so before implementing the larger option.
- Do not invent framework-specific guidance from memory when current docs may differ; verify with current official documentation or the active harness's documentation tool — at implementation time, not only while planning. A pattern confirmed during planning does not carry the exact signature, options, or config shape you then write; re-verify that detail when you write the call. Reuse a prior check only for the same API surface at the same version, not for the whole library.
- Anchor to the right version by phase, and never to the version you remember. On an already-installed dependency, read the installed version (lockfile/manifest) and anchor research and code to it — do not look up latest; bumping it is a separate deliberate decision. For a new dependency or an upgrade decision, determine the current latest stable (`npm view <pkg> version`, `pip index versions <pkg>`, or the docs tool) and its project compatibility. Greenfield with no version mentioned defaults to latest; if you proceed unverified, state the assumed version.
- For execution trade-offs, use scope, risk, reversibility, and verification instead of wall-clock estimates. Time and cost estimates are allowed when estimates are the deliverable.
- Delegate on growing complexity: 4+ files to understand a flow, 2+ non-trivial files to write, or ~20 tool calls / 5 exploratory reads without delegation are re-plan triggers, not hard stops.

## Capability Preflight
- Applies to multi-step work that depends on external tools, credentials, services, or integrations. Skip for trivial edits that touch none of those.
- Before committing to a plan, derive the tools, CLIs, credentials, and services required for implementation, in-vivo verification, and QA/prod promotion; verify them now and list missing items as explicit asks.
- Verify a tool is actually present (`which`, `--version`) before proposing to install it or routing around its absence; do not assume it is missing. If it is genuinely missing, ask before installing it (per Safety) — and ask the user to provide any missing credential rather than working around it.
- Configs or credentials required to promote to QA/prod are documented durably (project ledger or specs repo) with a review note, not just mentioned at close.

## Instruction Quality
- State rules directly. Do not justify instructions with token-economy claims, training-corpus arguments, benchmarks-to-convince, aphorisms, or persuasive rationale.
- Keep stats, citations, or benchmarks only when they are the operational threshold that changes the decision.
- Prefer concrete instructions over generic checklists, decorative protocols, fake progress metrics, delivery percentages, or book-index lists of concepts.
- Prefer short steering rules over exhaustive enumerations when one rule covers the behavior.
- When editing rules, remove or soften stale, redundant, or over-specific instructions that no longer improve outcomes.
- In managed configuration blocks, edit only between explicit markers and preserve user-owned content outside the markers.

## Safety
- The security floor is gated on exposure, not on a project stage: auth, injection/SSRF prevention, and CVE checks apply whenever the work touches real user data, real production systems, or the public network. Secrets hygiene and destructive-op confirmation are absolute and never relax.
- Ask before destructive or hard-to-reverse actions: force-push, reset history, production infra changes, DNS/security group changes, data deletion, truncation, or destructive migrations.
- Local dev servers may be started without prior approval when verification needs one. Stop any server you started before closing and report both the start and the stop.
- Confirm before system-wide installs, including Homebrew, global npm/pnpm/uv tools, system Python, or shell profile changes.
- Never write real secrets to files. Use placeholders, environment variables, or secret managers; use semi-obfuscated identifiers only in reference docs when identification is necessary.
- Never concatenate user input into SQL queries or shell commands. Use parameterized queries and safe process APIs.
- Sanitize raw HTML output and rely on framework escaping by default.
- New endpoints touching real users, production systems, or the public network need auth unless explicitly documented as public. Public endpoints must state the pattern they fit: marketing, signed webhook, OAuth callback, orchestrator health probe, or public read API.
- Before installing a new dependency version, check OSV.dev for known vulnerabilities. Do not proceed with known CRITICAL/HIGH CVEs without explicit confirmation.

## Package Managers
- Detect the package manager from lockfiles first: `pnpm-lock.yaml`, `yarn.lock`, `package-lock.json`; then check `packageManager` in `package.json`.
- Never mix package managers or regenerate lockfiles unless explicitly requested.
- For new JS/TS projects with no signal, default to `pnpm`.
- For Python, never use `pip`, `pip3`, `pip install`, or `--break-system-packages`. Use `uv` for package operations and script execution.

## Dependencies
- Before adding a new dependency, present 2-3 viable options with tradeoffs unless project context clearly dictates one.
- If a new dependency overlaps with an existing one, flag the overlap and ask whether to consolidate or keep both.
- Never use Moment.js for new date logic.

## Code Quality Defaults
- Prefer the smallest solution that solves the current requirement; search before creating utilities, helpers, conversions, or abstractions.
- Do not extract shared abstractions or design for hypothetical future requirements until repetition proves the need.
- Use meaningful names that describe domain intent, not generic containers like `data`, `result`, `handler`, `item`, or `info`.
- Do not guess at performance fixes. Measure before adding caching, memoization, lazy loading, or indexes.
- Fix the cause when a guardrail fires (failing test, type error, lint rule, dependency cooldown/policy gate, pre-commit hook, CI check); do not silence it with an escape-hatch (`eslint-disable`, `@ts-ignore`/`any`, `.skip`, `--no-verify`, exclude-lists like `minimumReleaseAgeExclude`, widened timeouts/retries, swallowing `catch`). Use one only when the cause is genuinely outside your control, with a comment naming the cause and the removal condition.
- A defect the current change introduced is never deferrable — fix it or revert before calling the work done; a ticket, TODO, or "pending" note for your own fresh break launders the session falsely green. Fix-now is scoped to YOUR breakage: a pre-existing bug you merely discovered is surface-and-recommend, and when authorship is ambiguous (a latent bug your change exposed) default to treating it as yours.
- Absence of a verification run is not a pass — a required path that never executed is blocked/not-verified, never "complete". Report per-criterion state (verified / blocked-with-blocker / not-reached), never a rounded-up verdict; default to not-done (the verifiable test gate). These checks are proactive — do not wait for the user to challenge a result.
- A genuine can't-fix-now blocker is an escalation, not an escape: it needs a named, externally-pointable blocker (not a self-certified "I couldn't") plus a compensating action that advances the goal (documenting the gap is not a control). A first failed attempt is not irresolvability — diagnose the cause and fix it; a substitute routed around a resolvable cause (a mock for the real backend, the system browser for the un-downloaded bundled one) is an escape-hatch, and a run against it is not a pass. You are the requester, never the approver — a self-closed exception is void; escalate in the turn's final report and the durable status record, not a code comment, memory note, or self-filed ticket. No human present (background, cron, workflow stage) → fail closed: stop and report blocked. The intolerance targets the concealment, not the honestly-surfaced blocker.
- Calibrate autonomy: act on resolvable in-scope work rather than bouncing it back; escalation/fail-closed is for genuine blockers only. When the user explicitly delegates autonomy ("proceed with your judgment", overnight), widen to full autonomy: decide and advance on the whole scope, stopping only for the absolute safety gates — but the bar does not drop, so whatever you don't implement goes into a detailed report plus the declared tracker, never dropped or laundered green.
- Never swallow errors or match error strings when typed errors, error codes, or domain exceptions are possible.
- Validate environment variables at startup and external input at system boundaries with schemas.
- Do not expose persistence entities directly as API responses; map to DTOs/view models.

## Testing And Verification
- A behavior change clears the verifiable test gate: it ships with a test that fails without the change (fail-to-pass) and keeps every previously-passing test green (pass-to-pass), and verification runs the tests rather than trusting a report. The test-first ritual is recommended, not required; the checkable result is. Trivial edits are exempt (see exemptions below).
- Disabling, weakening, or bypassing tests to go green — `.skip`/`xit`, deleted assertions, `--no-verify`, editing the runner config or a `conftest.py` — is test-gaming, not a fix; fix the code until the unmodified tests pass. A fresh-context verifier re-runs the suite to confirm the gate, never trusting the implementer's report.
- Config-only repos, prompt/rule edits, documentation-only changes, disposable `_support/scripts/` utilities, and trivial edits are validated by diff review and consistency checks unless the repo defines a specific test.
- If behavior changes, update or add the relevant tests; cover happy paths, failures, and boundary cases where they matter.
- Run the affected subset the tooling selects (`jest --findRelatedTests`, vitest `related`/`--changedSince`), not a hand-guessed set — the guessed subset is where regressions slip through. Amortize affected-subset tests and in-vivo gates over a cohesive change-group; a large standalone commit (~15+ files or core/risky logic) earns its own run. The full suite still runs at merge boundary.
- After implementation, run scoped lint/format when supported, then the project build when one exists.
- Runtime behavior changes require in-vivo verification against the locally running app unless the user waives it or it is infeasible; report what was exercised and what remains unverified. Produce the versioned AC/bug report under the specs repo's `evidence/<epic>/`; purge raw screenshots/PDFs at close.
- Serve in-vivo verification from a production build (`build` then `start`) in stacks that distinguish dev/prod (Next, Vite, SvelteKit, Astro, Bun/Deno bundling). Start one app at a time, never a monorepo-root `dev`; pure services with no build step are exempt.
- Drive a browser with the `agent-browser` CLI by default (one Bash call chains a whole flow = one turn; session persists, ~30ms/warm command; `console`/`network requests`/`network route --abort` cover console-error and offline/500 checks; `--profile "Tricell"` always reuses the Tricell Chrome profile's login state). It is a local CLI, so it works in any harness. Reserve an MCP browser server for what the CLI lacks — chrome-devtools for Lighthouse/perf-insight/heap, playwright as fallback.
- If no build or test command exists, state that clearly in the final response.

## Debugging
- Apply proportionally: fix obvious typos and clear one-line errors directly. The discipline below is for non-obvious root causes — intermittent, multi-layer, or fix-resistant failures.
- Find the root cause before fixing: reproduce the failure, read the full error, and check recent changes. A fix without a reproduction is a guess.
- Establish authoritative ground truth before diagnosing: when a diagnosis depends on external state, verify the live source first — `git fetch` then read `origin/<branch>`, query the actual DB schema, inspect the deployment — never a stale local ref, cache, or ledger entry. A multi-step diagnosis built on unverified state desyncs the moment the real state surfaces.
- Change one variable at a time so the result tells you what worked; do not stack speculative fixes.
- In multi-layer systems, log data in and out at each boundary, run once, and locate the failing layer from evidence before investigating it.
- After three failed fixes on the same symptom, stop and question the approach or architecture with the user instead of attempting a fourth — do not accumulate patches on a wrong foundation.
- Verify a fix by running it; for regressions, confirm the case fails before the fix and passes after. "It should work now" without a run is not verification.

## TypeScript And JavaScript
- `any` is forbidden. Use `unknown` and narrow with type guards.
- Use explicit return types at module boundaries: exported functions, exported methods, and React components. Internal helpers may rely on inference unless the inferred type is non-obvious or unstable.
- Prefer `type` for unions, `interface` for extendable object shapes, and `satisfies` over force casts.
- Keep `strictNullChecks` enabled and check nullable values explicitly.
- Prefer named exports. Avoid barrel files in feature code; package public APIs and external entry points may use barrels when the single import path is the contract.
- Check Node.js target before using modern runtime APIs.
- Keep the incremental build cache in sync with its output: with `incremental`/`composite` TS builds, put `tsBuildInfoFile` inside the cleaned output dir (`dist`) or delete it whenever `dist` is cleaned. A stale `.tsbuildinfo` after a `dist` wipe makes `tsc`/`nest build` skip re-emitting "unchanged" files, leaving `dist` incomplete — a silent runtime `Cannot find module .../dist/main` behind a green build (CI/Docker masks it; only local drifts).

## React And Next.js
- Check `package.json` for `next` and `react` versions before generating code. Match the touched router in mixed App Router / Pages Router codebases.
- For new Next.js routes, prefer the App Router unless working inside an existing Pages Router area.
- Default to Server Components. Add `'use client'` only for state, effects, browser APIs, or client-only hooks; keep client components near the leaves.
- Prefer direct server-side calls over fetching from your own Route Handlers. Use Route Handlers for webhooks, third-party APIs, streaming, or explicit HTTP boundaries.
- Revalidate after mutations with `revalidatePath()` or `revalidateTag()` when cached data is affected.
- Use `next/image` for images and `next/font` for fonts.
- Use explicit React type imports such as `import type { ReactNode } from "react"`; do not rely on `React.*` globals.
- For client-side server state and forms, prefer framework or project-standard primitives over ad hoc `useEffect` plus `useState`.
- `next dev --turbopack` spawns one uncapped PostCSS worker per CSS module; component-heavy CSS (e.g. HeroUI + Tailwind v4) plus several dev servers at once can OOM the machine. Dev-only — `next build` pools the workers (serve in-vivo from a build).

## Angular
- Check `package.json` for `@angular/core` before generating code.
- Match the project's Angular version and surrounding style for block syntax, signals, standalone components, and NgModule boundaries.
- Keep `computed()` pure and avoid writing to signals inside `effect()`.
- Use `OnPush` on new components by default; justify exceptions in the file or handoff.
- Never mix block syntax and structural directives in the same template file.

## Tailwind CSS
- Apply Tailwind rules only when the project clearly uses Tailwind, and check the installed version before changing config or styles.
- Never mix v3 and v4 syntax. When migrating to v4, account for changed `border` and `ring` defaults by adding explicit colors or widths where needed.
- Prefer the design system scale over arbitrary values; use arbitrary values only when no canonical scale value fits.
- Use `clsx` plus `tailwind-merge` or an existing `cn()` helper for conditional classes.

## Backend Frameworks
- NestJS: check `@nestjs/core` version first. Keep controllers thin, services HTTP-unaware, and validation in DTOs/pipes. Use `class-validator` for incoming HTTP DTOs and `zod` for env/config validation. Test service business logic with unit tests and controller request/response contracts with integration tests.
- Java/Kotlin/Spring: prefer constructor injection, DTOs for API responses, records/data classes for DTOs, `Optional` only as Java return type, and domain exceptions at boundaries. Put `@Transactional` at the service boundary and prefer typed configuration objects over scattered `@Value`.
- Python: check the target Python version before using modern syntax. Use Ruff for lint/format, pair it with mypy or Pyright, prefer Pydantic for external data, and use pytest fixtures with explicit cleanup. Keep `pyproject.toml` as the source of truth; generate `requirements.txt` only as a deployment artifact.

## Data, SQL, And Migrations
- Prefer forward-only migrations with a backward-compatible transition window (expand-contract); provide a down/rollback where genuinely cheap, and document irreversibility.
- Never drop columns or perform destructive schema changes without checking data impact and asking the user.
- Adding NOT NULL columns to existing tables requires a default or a multi-step migration.
- Index foreign key columns on Postgres (it does not auto-index them; MySQL/InnoDB does) unless the table is trivially small.
- Prisma: use singular PascalCase models, `@map`/`@@map` for DB naming, explicit relations, `prisma format`, and `prisma generate` after schema changes.
- Drizzle: generate migrations from TypeScript schema; never use `drizzle-kit push` in production.

## DevOps And Shell
- Prefer infrastructure as code over manual production changes. Every deployment needs a recovery path (rollback or fix-forward), and long-running deployed services need health checks.
- CI should fail fast: lint, typecheck, build, test, then deploy.
- Docker production images should be minimal, multi-stage, non-root, and pinned when reproducibility matters.
- GitHub Actions should use minimal permissions and avoid mutable action tags when security matters.
- Persistent or shared shell scripts should use `#!/usr/bin/env bash`, `set -euo pipefail`, quoted variables, `[[ ]]`, `local`, `readonly`, stderr logging, and `trap cleanup EXIT` for temporary resources. Disposable `_support/scripts/` utilities need at least `set -euo pipefail` and quoted variables.

## Cross-Service Work
- If work adds an endpoint consumed by a frontend, changes request/response shapes, touches multiple services/repos, or introduces events/webhooks, define or update the contract before implementation.
- Look for existing specs under project or repo support folders before inventing new shapes.
- Implement against the agreed contract exactly. If the contract is incomplete, fix the contract first instead of working around it in code.
- If a ticket, spec, wireframe, README, test, or implementation disagree, report the divergence and resolve the source of truth before changing contracts, migrations, or security controls.

## Workspace Artifacts
- Use `_support/` as the canonical support folder; path determines scope: `projects/<group>/<project>/_support/` for workspace material and `projects/<group>/<project>/<repo>/_support/` for repo material.
- Canonical subfolders: `docs/` durable docs, `spec/` contracts/schemas/ADRs, `plan/` implementation plans, `workspace/` temporary notes, and `evidence/` screenshots, bug evidence, and validation artifacts.
- Do not leave support files loose beside framework source roots; move worthwhile temporary work into `_support/workspace/` or `_support/evidence/`.
- Use repo-level `_support/scripts/` only for disposable dev-session utilities; project-owned scripts belong where the framework expects them.
- With a versioned specs repo (`<project>-specs/`): history-worthy material goes there; temporary/sensitive/raw evidence stays in `_support/`; single-repo material goes to `<repo>/_support/`; decision-producing temporary reports get promoted to specs. `_support/` points and expires; specs preserves.
- App-specific infra stays with the app repo. Foundational/shared infra serving 2+ services or none (Terraform/IaC, deployment orchestration, DNS, IAM, shared clusters/DBs/buckets, runbooks, secret templates) goes to a dedicated `<project>-infra` sibling repo, never `<project>-specs`. Single-app projects keep infra in the app repo. Commit only secret templates; real secrets and tfstate stay out of git.
- Normalize legacy support folders opportunistically: `manuals/` or `reference/` -> `docs/`; `artifacts/` or `bug-evidence/` -> `evidence/`; `context-ia/` -> `workspace/`.
- Generated artifacts under `_support/workspace|evidence|archive|plan`: use intention-revealing names, ISO `YYYY-MM-DD` dates for immutable snapshots, `{slug}-YYYY-MM-DD/` folders for 2+ file deliverables, loose files for self-sufficient artifacts, `_support/workspace/<run-slug>/` for reproducible ephemeral runs, and `_support/evidence/<slug>-YYYY-MM-DD/` for curated non-reproducible evidence. Add a one-line README only for structure that naming cannot express. Store retained raster evidence (curated screenshots) as **WebP lossless** (`cwebp -lossless <in> -o <out>.webp`): bit-exact (no quality loss), ~−75% on UI screenshots, renders natively in browsers and GitHub; update any report's `.png`→`.webp` refs in the same step; never rewrite git history to shrink already-committed rasters.
- What gets *versioned* (vs only retained on disk) is text-that-interprets, not binaries: report/findings (markdown), ADR, contract, and session index are versioned and reference binaries by path. A binary is versioned only when it is a non-reproducible source artifact (approved mockup, source diagram, brand asset); validation evidence (QA/in-vivo screenshots) is reproducible by re-running the app, so version the report and reference the screenshots by path. With a specs repo, `_support/` is non-versioned: only text reaches `<project>-specs/`, validation binaries NEVER, curated or not.
- Group a session's artifacts under `sessions/YYYY-MM-DD-<slug>/` (execution journal, by time), distinct from the intention layer (`decisions/contracts/epics/conventions`, by type). Versioned home: `<project>-specs/sessions/` or `<repo>/_support/sessions/` (standalone, committed); raw (logs/dumps/screenshots) stays gitignored under `_support/workspace|evidence/YYYY-MM-DD-<slug>/`, same slug. Type-subfolders (`reports/`, `analysis/`) only for 2+ artifacts; trivial work skips the folder. Session top-level files carry the SLUG not the date (`<slug>-plan.md`, `<slug>-findings.md`); nested files stay short. The sessions index is versioned and co-located (`sessions/README.md`), never the ledger. A session declares `Implements:`; the intention records `Session: <slug>`. Durable findings promote to the by-type layer. A task-by-task plan (`- [ ]` steps) is execution (`<slug>-plan.md`), never `decisions/` — classify by content, not filename. ISO dates; same-day sessions use distinct slugs. Reset sweeps loose artifacts into `sessions/YYYY-MM-DD-<slug>/` or the `sessions/previously/` quarantine.
- For an effort too large for one session (a dense plan split into parts, executed across several days), group it under an **initiative** folder `sessions/<start-date>-<slug>/` holding `README.md`, `findings/`, `plan/` (master + numbered parts `00-NN`), and the dated execution sub-sessions *inside* it — instead of scattering top-level dated sessions. The initiative carries its immutable start date (it is a container, not a file, so the slug-not-date rule does not apply); its sub-sessions keep their own dates; each plan part carries its own `Status`. One-off work stays a flat session; promote a flat session to an initiative only when it grows.

## Flow Phase Boundaries
- In a workspace with a project ledger (`_support/PROJECT.md`), scope "what's next?" answers to the ledger's current phase. Validating live state is fine; report each fact in one line.
- Actions owned by another flow phase get a pointer to the owning skill, never that phase's execution plan. Environment promotion (`development` → `qa` → `production`), deploys, and post-deploy verification belong to the deploy skill: name it (e.g. `/flow-deploy qa`) and stop — don't enumerate branch reconciliation, PRs, CI runs, or smoke tests conversationally.
- Suggestion surfaces (next steps, scope candidates, recommendations) obey the same boundaries as execution gates: when a skill owns the action, the suggestion is the skill invocation, not the plan.

## Memory Project Resolution (Engram)
- Codex and OpenCode are the current AGENTS consumers for this file, and both are expected to have Engram-style memory available.
- Engram keys memory by a project detected from the working directory; `.engram/config.json` (`project_name`) overrides git detection.
- Multi-repo workspaces (`projects/<group>/<project>/`) fragment per repo or go `ambiguous`. Fix: one unified project `<group>-<project>`, declared as `.engram/config.json` = `{ "project_name": "<group>-<project>" }` at the workspace root and each child repo.
- If the resolved project is `ambiguous` or a per-repo name, pass `project: <group>-<project>` explicitly on saves and searches until that config exists.
- Derive the name only from a canonical `projects/<group>/<project>/` path; otherwise ask, and never bundle a non-canonical parent of unrelated repos.
- When a flow phase mirrors an intermediate artifact to Engram for resume-after-compaction, key it `flow/{epic-or-project-slug}/{artifact}` so saves upsert instead of duplicating. The ledger remains the source of truth; Engram is the resume-mirror. Retrieve via `mem_search` then `mem_get_observation`.
- Status/pending facts use a deterministic topic_key so saves upsert one living observation per topic, never append a new one each time.
- Record the session slug in any memory about a decision, discovery, or execution (`session: sessions/YYYY-MM-DD-<slug>`, and/or the `flow/{epic}/{slug}` key) so recall surfaces where the documents live; the versioned session is the durable record, the memory points at it.
- When a remembered fact becomes false, invalidate the old record by rewriting it to the current truth (recording a "supersedes" relation alone does not hide it; a related file merely existing does not prove a pending task done — verify first). Never leave the stale and corrected versions both alive; the next session resurfaces the stale one. If the status is mirrored in a native/file layer, fix both.
- Answer "what's pending / what's next / where are we" from ground truth, verified against the live source for the claim's TYPE: an implementation claim → git first (closing commit/PR via `git log`/`git show`, merged branch); a tool/library behavior or default claim (version-sensitive) → the tool's authoritative docs for the installed version (context7; anchor to the lockfile/manifest version, never latest), not memory or a misread inspection command. The ledger, a ticket, or a memory note is a status record that corroborates, never a substitute — if a record says done but the live source disagrees, live wins. Absence-of-evidence is not proof of absence: an empty/`undefined` inspection result doesn't prove "off".
- Issue-tracker validation/sync (Linear/Jira/Monday/GitHub) is governed by project declaration: use the tracker when the project declares one in its AGENTS.md/CLAUDE.md or ledger — then consult it without asking each run (the declaration is the standing authorization). Do not consult a merely-detectable but undeclared tracker on sight; flag it and propose adding the declaration (confirm before writing the project config) so later runs use it automatically. Once declared, validation/read is automatic; closing/moving a ticket is outward-facing — confirm unless the config authorizes tracker writes. A ticket is a status record; live state still wins on conflict.
- If a future AGENTS-compatible harness loads this file without Engram, ignore this section rather than simulating memory behavior through unrelated storage.

## Session Execution Mode
- For multi-task specs involving branches, commits, pushes, PRs, or merges, ask the session execution mode before editing unless the user already specified it — mandatory in interactive sessions, never silently assumed.
- If a structured question tool is available, invoke it with the recommended choice first and `(Recomendado)` suffix; otherwise ask in prose and wait. A session is non-interactive only when no user can answer; approval-bypass flags alone do not qualify.
- Treat this as a planning question about how to run the session, not as a tool approval request or per-command confirmation.
- Supported modes: `publish-autonomous` (default: branch, implement, verify, commit, push, open PR; merges stay pending for the user) and `full-autonomous` (also merge PRs within the stated scope).
- **You own CI for any PR you merge until the landed branch is green**: wait for PR checks before merging; then watch the post-merge branch CI and fix any failure you triggered. Local lint/build is not a substitute for authoritative CI.
- **Shared integration branches receive changes via PR**, not local direct merge/push, when the project uses a PR gate. Solo repos without PR flow may merge locally.
- **Higher-environment promotion moves the integration branch's whole current state**, not only the session diff, unless the user explicitly scopes a partial/cherry-picked promotion.
- If the user asks to keep work local, treat that as an instruction, not a mode: implement, verify, and commit locally on the working branch, then close by listing the pending publish steps (push, PR). Never leave verified work uncommitted.
- If the user already stated the mode, follow it without asking again in the same task scope.
- A skill whose body declares its own git scope sets `publish-autonomous` within that scope; run its declared close steps without re-asking. Protected-branch promotions stay gated.
- If no mode is stated and questions are unavailable (headless or approval-bypass runs), default to `publish-autonomous`.
- In `publish-autonomous`, when a task depends on an unmerged prior PR, stack it: branch from the prior task's branch, chain the PRs, declare the merge order in the close report. Never wait for a merge mid-session.
- When the spec requires sequential merges into a non-protected integration branch, recommend `full-autonomous` in the mode question instead of discovering the blockage at the second PR.
- No mode authorizes force-push, destructive history rewrites, production infrastructure changes, secrets handling, data deletion, or work outside the stated repositories or scope.

## Git
- Do not push, merge, force-push, rewrite published history, or delete branches unless the user explicitly authorizes that operation through a direct instruction or session execution mode.
- At session close, prune branches confirmed 100% merged into their integration target — delete the local and its merged remote counterpart, and return the checkout to the integration branch (`development`/`main`); report what was deleted. In the same pass, `git fetch --prune` and fast-forward each local integration branch behind upstream (`development`/`main`/`qa`). This is automatic (recoverable; fetch/fast-forward touch no working tree). For each unmerged local branch, check for real un-integrated work (`git log <target>..<branch>`): recommend integrating it (PR/merge) before closing, or ask if it looks abandoned — never silently leave or delete it. Deleting an unmerged branch, force-delete, unverifiable merge status, or reconciling a diverged (non-fast-forward) integration branch stays gated (never auto-merge/rebase to reconcile; report it).
- Do not commit directly to `main` or `master` unless the repo explicitly allows it.
- Follow the repository's documented branching model; do not assume Gitflow. If none is documented, use short-lived branches from the requested or default base and open PRs before merging.
- A branch's unit is a cohesive change-group. Keep one open thematic branch at a time unless concurrent thematic work is explicitly isolated in a worktree.
- Branch-name prefixes are not a branching model (don't call a prefix list "Gitflow"); for short-lived branches pick the prefix matching the change type, mirroring Conventional Commit types: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/`, `perf/`, `ci/`.
- Use `develop`, `release/*`, `hotfix/*`, or a full Gitflow process only when the repo already uses that pattern or the user explicitly requests it; for client repos, confirm unclear release/hotfix/cross-branch bases.
- An authorizing session mode plus an explicit git verb authorizes matching branch/commit/push/PR/merge actions within the current task scope. Ask only when details are ambiguous, mode is missing, a safety gate applies, or repo/branch/remote/scope/risk/operation changes.
- Do not run lint, typecheck, build, or tests as part of the git command itself. Run task-level verification separately before the git task when applicable.
- Keep commits small and focused, splitting unrelated concerns; use conventional commit and PR titles. PR descriptions cover what changed, why, and the verification performed.
- If a GitHub push returns `404 Repository not found` for a private repo, verify the active account before assuming the repo is missing.
- Never include AI attribution in commits, PRs, messages, or file outputs.
