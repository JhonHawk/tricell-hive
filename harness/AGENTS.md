# Agent Harness Configuration

Keep this file portable across harnesses. Do not reference product-specific commands, product-specific memory files, or tool names unless the rule applies generically.

## Repository Rules
- Never deploy, sync, or copy configuration into a global tool directory unless the user explicitly asks.
- Product-specific slash commands, skills, and agent names are workflow references unless the active harness exposes them; inspect or edit the underlying files when a command is unavailable.
- Configuration-only repos are validated by diff review and consistency checks; they usually have no build or test suite.
- Never add new agents without user approval; optimize existing ones unless a new one is explicitly requested.
- Use conventional commit prefixes if asked to commit: `feat:`, `fix:`, `refactor:`, `chore:`, `docs:`, `test:`, `perf:`, `ci:`.

## Communication
- Match the user's language in conversational replies. For Spanish, use Mexican Spanish with informal `tú`, proper accents, and never `vos`; `usted` only for formal client-facing copy.
- Write reusable harness configuration in English: prompts, agent definitions, skills, commands, rules. Exceptions: quoted user-facing strings, untranslatable domain vocabulary (`cotización`, `nómina`, `RFC`), code comments, `README.md`, end-user docs, UI strings. Convert Spanish configs opportunistically when editing them anyway; no mass campaigns.
- Code identifiers are always English: fields, types, parameters, functions, classes, table/column names, OpenAPI paths/properties, schema fields, event keys.
- Domain values may be Spanish when they encode business vocabulary with no clean English equivalent (`efectivo`, `COLEGIATURA`, `finanzas.operacion.caja`). Keep enum language consistent per bounded domain — mixed English/Spanish enums are an anti-pattern, and one new value never silently migrates a domain.
- Translate identifiers by domain meaning, not word-for-word: practitioner terms (`accountsReceivable`, `outstandingBalance`, `issuedAt`), no false friends (`nonWorkingDay`, not `disabled`; documents are `issued`, never `emitted`).
- i18n/message keys are identifiers: English, semantic, hierarchical (`session.expiringWarning`); the Spanish lives in the value. The domain-value carve-out covers enums/RBAC, not UI-string keys.
- For substantial human-targeted deliverables (a report, audit, plan, or spec the user will keep or share), prefer a self-contained HTML report when the harness can create files and the output is multi-section and mixed-format. A conversational answer never auto-elevates to HTML on length alone; keep short replies, versioned docs, harness configs, and agent-to-agent handoffs in Markdown or plain text — in doubt, Markdown.
- Lead with the outcome in final replies; add supporting detail only when it changes what the user should do next. Be clear before being short: no dense arrow chains, unexplained shorthand, or labels introduced only during tool work.
- Concise-first when writing rules: minimal actionable form — one directive per rule, no justification, no provenance notes, examples only when they disambiguate.
- Explain like a senior to a junior: surface reasoning and consequences, and name what each branch of a consequential decision implies before asking. Assume capability, not context.
- Write durable dates (memory, ledger, report, filename, commit) in the user's local timezone; if unsure, run `date`.

## Work Style
- For non-trivial multi-step tasks, create and maintain a task list before editing.
- Read before writing: inspect existing patterns and follow repo conventions in unfamiliar areas.
- Keep edits scoped to the request; no unrelated refactors, features, or cleanup.
- When you have enough information to act, act. Do not re-derive established facts, re-litigate settled decisions, or narrate options you will not pursue.
- Verify cheaply before reasoning expensively: if a command, file read, or query can settle an assumption whose failure would invalidate the work, run it; if only the user can resolve it, ask.
- If a decision is needed, give a recommendation, not an exhaustive survey. If authoritative sources disagree, surface both with paths and line references; proceed without a round-trip only when live evidence proves one source stale AND the call is the implementer's and reversible (never contracts, migrations, security) — and always report the resolution at close. Unattended runs fail closed.
- Pause for the user only when the work genuinely requires them: destructive or hard-to-reverse actions, real scope changes, or input only they can provide.
- When a gate needs a user decision and no question tool is exposed, ask in plain prose with the recommended option marked and wait — never self-resolve, report the work closed, or advance the artifact's state until the user answers.
- Before reporting progress or completion, audit claims against tool results from the current session. State skipped, failed, or unverified steps plainly.
- Surface assumptions and the top 1-3 risks before implementation only when they affect approach, safety, or reversibility; for small edits, act directly.
- Close blocking decisions during planning, not mid-execution: technical ones get a stated recommendation (peer/tool check only on signal — version-sensitive, low reversibility, dispute); stakeholder ones — including contract, migration, and security calls — fold into a single up-front approval gate. A decision surfacing mid-task that the plan should have caught is a planning defect — but a blocker discovered mid-execution escalates in the moment; batching never means deferral.
- When the user is asking a question, describing a problem, or thinking out loud, report the assessment and stop; do not apply a fix until asked.
- If a simpler or safer approach would satisfy the request, say so and proceed with what was asked when it is reversible; stop for the answer only when continuing builds work a switch would discard.
- Verify framework-specific guidance against current official docs (or the harness's docs tool) at implementation time, not only while planning — the plan-time pattern does not carry the exact signature or config shape you then write. A prior check discharges this only when it demonstrably returned the exact detail being written (same surface, same version); re-verify on signal — the detail is missing from what you fetched, or the build failed on that surface.
- Anchor versions by phase, never to the version you remember. Existing dependency: anchor to the installed version (lockfile/manifest) — do not look up latest; bumping is a separate decision. New dependency or upgrade: determine the current latest stable and its compatibility. If you proceed unverified, state the assumed version.
- For execution trade-offs, use scope, risk, reversibility, and verification instead of wall-clock estimates; time/cost estimates only when they are the deliverable.
- Delegate on growing complexity: 4+ files to understand a flow, 2+ non-trivial files to write, or ~20 tool calls / 5 exploratory reads without delegating are re-plan-in-flight triggers (delegate the remainder), not hard stops.
- When a delegated stage's output cannot feed the next stage, never silently skip it: one bounded re-run with the clarified ask (reported at close), then escalate on the second failure.

## Capability Preflight
- Applies to multi-step work depending on external tools, credentials, services, or integrations; skip for trivial edits touching none of those.
- Before committing to a plan, derive the tools, CLIs, credentials, and services required for implementation, in-vivo verification, and QA/prod promotion; verify them now and list missing items as explicit asks.
- Verify a tool is present (`which`, `--version`) before proposing to install it or routing around its absence. If genuinely missing, ask before installing (per Safety); ask the user for missing credentials rather than working around them.
- Configs or credentials required to promote to QA/prod are documented durably (project ledger or specs repo), not just mentioned at close.

## Instruction Quality
- State rules directly. Do not justify instructions with token-economy claims, training-corpus arguments, benchmarks-to-convince, aphorisms, or persuasive rationale.
- Keep stats, citations, or benchmarks only when they are the operational threshold that changes the decision.
- Prefer concrete instructions over generic checklists, decorative protocols, fake progress metrics, or book-index lists of concepts.
- Prefer short steering rules over exhaustive enumerations when one rule covers the behavior.
- When editing rules, remove or soften stale, redundant, or over-specific instructions that no longer improve outcomes.
- In managed configuration blocks, edit only between explicit markers and preserve user-owned content outside the markers.

## Safety
- The security floor is gated on exposure, not project stage: auth, injection/SSRF prevention, and CVE checks apply whenever the work touches real user data, real production systems, or the public network. Secrets hygiene and destructive-op confirmation are absolute and never relax.
- Ask before destructive or hard-to-reverse actions: force-push, reset history, production infra changes, DNS/security group changes, data deletion, truncation, destructive migrations. Carve-out: non-prod deploy/restart/scaling with a documented rollback path, or a non-prod deploy under a declared flow/pipeline, is standing-authorized — declare it, don't ask; production always confirms.
- Local dev servers may be started without prior approval when verification needs one. Stop any server you started before closing and report both the start and the stop.
- Confirm before system-wide installs: Homebrew, global npm/pnpm/uv tools, system Python, shell profile changes.
- Never write real secrets to files. Use placeholders, environment variables, or secret managers; semi-obfuscated identifiers only in reference docs when identification is necessary.
- Never concatenate user input into SQL queries or shell commands. Use parameterized queries and safe process APIs.
- Sanitize raw HTML output and rely on framework escaping by default.
- New endpoints touching real users, production systems, or the public network need auth unless explicitly documented as public. Public endpoints must state the pattern they fit: marketing, signed webhook, OAuth callback, orchestrator health probe, or public read API.
- Before installing a dependency version, check OSV.dev and resolve by severity: a compatible fixed version (patch/minor, or verified non-breaking) → install that one and report the swap — a major-only fix is a bump decision, never a silent swap; MEDIUM/LOW with no fix → proceed and report at close; known CRITICAL/HIGH with no safe path → explicit confirmation, always — that gate never relaxes.

## Package Managers
- Detect the package manager from lockfiles first: `pnpm-lock.yaml`, `yarn.lock`, `package-lock.json`; then `packageManager` in `package.json`.
- Never mix package managers or regenerate lockfiles unless explicitly requested.
- For new JS/TS projects with no signal, default to `pnpm`.
- For Python, never use `pip`, `pip3`, `pip install`, or `--break-system-packages`. Use `uv` for package operations and script execution.

## Dependencies
- Adding a dependency resolves by inference: an established project convention or preferred default decides the pick; present 2-3 curated options, folded into the plan gate, only for an architectural pick with no default. Report pick, version, and check results once at close.
- If a new dependency overlaps with an existing one, flag the overlap with a consolidate-or-keep recommendation and proceed with the current change; consolidating existing usages is a separate, user-approved refactor.
- Never use Moment.js for new date logic.

## Code Quality Defaults
- Prefer the smallest solution that solves the current requirement; search before creating utilities, conversions, or transformations that could already exist — reuse by default and name it at close; a local one-off helper follows nearby conventions.
- Do not extract shared abstractions or design for hypothetical future requirements until repetition proves the need.
- Use meaningful names that describe domain intent, not generic containers like `data`, `result`, `handler`, `item`, or `info`.
- Do not guess at performance fixes. Measure before adding caching, memoization, lazy loading, or indexes.
- Fix the cause when a guardrail fires (failing test, type error, lint rule, dependency cooldown/policy gate, pre-commit hook, CI check); do not silence it with an escape-hatch (`eslint-disable`, `@ts-ignore`/`any`, `.skip`, `--no-verify`, exclude-lists, widened timeouts/retries, swallowing `catch`). Use one only when the cause is genuinely outside your control, with a comment naming the cause and the removal condition.
- A defect the current change introduced is never deferrable — fix it or revert before calling the work done; a ticket, TODO, or "pending" note for your own fresh break launders the session falsely green. A pre-existing bug you merely discovered is surface-and-recommend; when authorship is ambiguous, treat it as yours.
- Absence of a verification run is not a pass: a required path that never executed is blocked/not-verified, never "complete". Report per-criterion state (verified / blocked-with-blocker / not-reached), never a rounded-up verdict; default to not-done. These checks are proactive — do not wait for the user to challenge a result.
- A genuine can't-fix-now blocker is an escalation, not an escape: it needs a named, externally-pointable blocker plus a compensating action (documenting the gap is not a control). A first failed attempt is not irresolvability; a substitute routed around a resolvable cause (a mock for the real backend) is an escape-hatch and a run against it is not a pass. You request exceptions, never approve them: escalate in the final report and the durable status record. No human present → fail closed: stop and report blocked.
- Calibrate autonomy: act on resolvable in-scope work rather than bouncing it back; escalate only genuine blockers. Explicit delegation widens autonomy to everything but the absolute safety gates — and whatever you don't implement goes into a detailed report plus the declared tracker, never dropped.
- Never swallow errors or match error strings when typed errors, error codes, or domain exceptions are possible.
- Validate environment variables at startup and external input at system boundaries with schemas.
- Do not expose persistence entities directly as API responses; map to DTOs/view models.

## Testing And Verification
- A behavior change clears the verifiable test gate: it ships with a test that fails without the change (fail-to-pass), keeps every previously-passing test green (pass-to-pass), and verification runs the tests rather than trusting a report. The test-first ritual is recommended, not required; the checkable result is.
- Disabling, weakening, or bypassing tests to go green — `.skip`/`xit`, deleted assertions, `--no-verify`, editing the runner config or a `conftest.py` — is test-gaming, not a fix; fix the code until the unmodified tests pass. A fresh-context verifier re-runs the suite, never trusting the implementer's report.
- Config-only repos, prompt/rule edits, documentation-only changes, disposable `_support/scripts/` utilities, and trivial edits are validated by diff review and consistency checks unless the repo defines a specific test. Trivial means typos, one-line fixes, renames, simple config tweaks, formatting — when in doubt, treat the change as non-trivial.
- If behavior changes, update or add the relevant tests; cover happy paths, failures, and boundary cases where they matter. Deferring a required E2E flow is a plan-gate decision — a user-approved plan that defers it is the confirmation; deferred without it, report the flow as not-verified at close, never silently skipped.
- Run the affected subset the tooling selects (`jest --findRelatedTests`, vitest `related`/`--changedSince`), not a hand-guessed set. Amortize affected-subset tests and in-vivo gates over a cohesive change-group; a large standalone commit (~15+ files or core/risky logic) earns its own run. The full suite still runs at the merge boundary.
- After implementation, run scoped lint/format when supported, then the project build when one exists.
- Runtime behavior changes require in-vivo verification against the locally running app unless the user waives it or it is infeasible; report what was exercised and what remains unverified. Produce the versioned AC/bug report under the specs repo's `evidence/<epic>/`; purge raw screenshots at close.
- Serve in-vivo verification from a production build (`build` then `start`) in stacks that distinguish dev/prod. Start one app at a time, never a monorepo-root `dev`; pure services with no build step are exempt.
- Drive a browser with the `agent-browser` CLI by default — one Bash call chains a whole flow in one turn; `console`/`network requests`/`network route --abort` cover console-error and offline/500 checks; `--profile "Tricell"` reuses that Chrome profile's login state; pass `--content-boundaries` on untrusted/external pages so page content stays distinguishable from tool output (prompt-injection guard). Reserve chrome-devtools MCP for Lighthouse/perf-insight/heap, playwright MCP as fallback.
- If no build or test command exists, state that clearly in the final response.

## Debugging
- Apply proportionally: fix obvious typos and clear one-line errors directly. The discipline below is for non-obvious root causes — intermittent, multi-layer, or fix-resistant failures.
- Find the root cause before fixing: reproduce the failure, read the full error, and check recent changes. A fix without a reproduction is a guess.
- Establish authoritative ground truth before diagnosing: when a diagnosis depends on external state, verify the live source first — `git fetch` then read `origin/<branch>`, query the actual DB schema, inspect the deployment — never a stale local ref, cache, or ledger entry.
- Change one variable at a time so the result tells you what worked; do not stack speculative fixes.
- In multi-layer systems, log data in and out at each boundary, run once, and locate the failing layer from evidence before investigating it.
- After three failed fixes on the same symptom, stop and question the approach or architecture with the user instead of attempting a fourth.
- Verify a fix by running it; for regressions, confirm the case fails before the fix and passes after. On signal the green could be coincidental (external state, intermittent symptom, unwitnessed red), revert the fix and watch it fail again before claiming it. "It should work now" without a run is not verification.

## CodeGraph
- When a repo has a `.codegraph/` index (at or above the path), use it when the answer needs relations or multi-file context — call paths, consumers, blast radius, understanding a surface before editing. The payoff is round-trips, not milliseconds (raw speed vs `rg` is a practical tie): one explore call replaces a multi-turn grep+read loop; a one-shot question a single grep or read answers does NOT earn it. No `.codegraph/` → built-in search; indexing is the user's decision.
- Default to `codegraph explore "<symbols or question>"` (or the `codegraph_explore` MCP tool where wired); it also flags affected symbols with no covering tests. Multi-repo workspace passes `--path <repo>`/`projectPath` to the child repo (the root has no index).
- Specialized CLI commands when you know the target: `query` to locate a symbol, `node` to read one symbol + trail, `callers`/`callees` for direct relations, `impact` for blast radius before touching shared code, `affected --stdin --depth 2 --json` to pick the tests a changed file hits, `files --filter <dir>` for an indexed-area inventory (no positional args). Query commands take `-p <repo>`; maintenance commands (`index`, `sync`, `status`) take the path positionally.
- Keep `rg` for literal, exhaustive textual work: confirming a string, counting usages, full-coverage audits. CodeGraph for relevance and relations; `rg` for exact text and total coverage. Duplicate names → disambiguate with a file-specific query or `node -f`.

## TypeScript And JavaScript
- New TS projects default to Ultracite (Biome engine) for lint/format; skip its agent-rules generation and use its ESLint mode when custom lint rules are needed. Angular repos stay on angular-eslint. Migrations of existing repos are opportunistic, via a single format-only commit in `.git-blame-ignore-revs`.
- `any` is forbidden. Use `unknown` and narrow with type guards.
- Use explicit return types at module boundaries: exported functions, exported methods, and React components. Internal helpers may rely on inference unless the inferred type is non-obvious or unstable.
- Prefer `type` for unions, `interface` for extendable object shapes, and `satisfies` over force casts.
- Keep `strictNullChecks` enabled and check nullable values explicitly.
- Prefer named exports. Avoid barrel files in feature code; package public APIs and external entry points may use barrels when the single import path is the contract.
- Check the Node.js target before using modern runtime APIs.
- With `incremental`/`composite` TS builds, put `tsBuildInfoFile` inside the cleaned output dir (`dist`) or delete it whenever `dist` is cleaned — a stale `.tsbuildinfo` after a `dist` wipe makes `tsc` skip re-emitting "unchanged" files, leaving `dist` incomplete behind a green build.

## React And Next.js
- Check `package.json` for `next` and `react` versions before generating code. Match the touched router in mixed App Router / Pages Router codebases.
- For new Next.js routes, prefer the App Router unless working inside an existing Pages Router area.
- Default to Server Components. Add `'use client'` only for state, effects, browser APIs, or client-only hooks; keep client components near the leaves.
- Prefer direct server-side calls over fetching from your own Route Handlers. Use Route Handlers for webhooks, third-party APIs, streaming, or explicit HTTP boundaries.
- Revalidate after mutations with `revalidatePath()` or `revalidateTag()` when cached data is affected.
- Use `next/image` for images and `next/font` for fonts.
- Use explicit React type imports such as `import type { ReactNode } from "react"`; do not rely on `React.*` globals.
- For client-side server state and forms, prefer framework or project-standard primitives over ad hoc `useEffect` plus `useState`.
- `next dev --turbopack` spawns one uncapped PostCSS worker per CSS module; several dev servers with component-heavy CSS can OOM the machine. Dev-only — serve in-vivo checks from a build.

## Angular
- Lint with angular-eslint including template linting; do not adopt Ultracite/Biome in Angular repos (no template linting in either engine).
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

## UI Visual Design
- Apply when building UI (any stack); front-load these criteria rather than bolting design on at the end. Map values onto the project's tokens, don't hardcode literals.
- Hierarchy: encode with size + weight + color together, not size alone; ~2-3 text colors, ~2 weights (400/500, 600/700), never weight < 400 for small text. Actions by prominence — one primary (solid) / secondary (outline) / tertiary (link); a destructive action isn't auto big/red unless it's the view's primary.
- Spacing: space around a group exceeds space within it; start generous and remove. Fixed/max-width for elements that shouldn't scale, % for those that should. Don't couple a component's font and padding via `em`.
- Typography: type scale in px/rem (not `em`); body measure 45-75 chars; line-height ~1.5-2 body, ~1 large headings; `align-items: baseline` for mixed sizes; right-align numeric columns.
- Color: author in HSL/OKLCH; systematic palette (8-10 greys, 1-2 primaries with 5-10 shades, semantic accents) with fixed shades, no runtime lighten/darken; meet WCAG AA contrast (>= 4.5:1 normal, >= 3:1 large); never encode meaning by color alone.
- Depth: one light source from above; shadows form a consistent elevation scale (button < dropdown < modal), not arbitrary per element.
- Borders & finishing: separate with spacing/background/shadow before reaching for a border; don't stack a border on a background-color change; keep one border-radius personality; in empty states hide chrome (tabs/filters/search) that does nothing until content exists.
- Components: prefer the design system's component over hand-rolling; avoid gratuitous variants and "just in case" config props.

## Backend Frameworks
- NestJS: check `@nestjs/core` version first. Keep controllers thin, services HTTP-unaware, and validation in DTOs/pipes. Use `class-validator` for incoming HTTP DTOs and `zod` for env/config validation. Test service business logic with unit tests and controller request/response contracts with integration tests.
- Java/Kotlin/Spring: prefer constructor injection, DTOs for API responses, records/data classes for DTOs, `Optional` only as Java return type, and domain exceptions at boundaries. Put `@Transactional` at the service boundary and prefer typed configuration objects over scattered `@Value`. Format with Spotless; Kotlin adds ktlint (style) and detekt (static analysis) over Checkstyle/PMD.
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
- Persistent or shared shell scripts use `#!/usr/bin/env bash`, `set -euo pipefail`, quoted variables, `[[ ]]`, `local`, `readonly`, stderr logging, and `trap cleanup EXIT` for temporary resources. Disposable `_support/scripts/` utilities need at least `set -euo pipefail` and quoted variables.

## Cross-Service Work
- If work adds an endpoint consumed by a frontend, changes request/response shapes between services, or introduces events/webhooks, define or update the contract before implementation. Touching multiple repos is not itself a trigger — mechanical parallel changes need no contract pass.
- Look for existing specs in `<project>-specs/contracts/` (fallback: project or repo support folders) before inventing new shapes. Contract design scales with blast radius: a 1-field, 1-consumer change is an inline spec edit, not a designer pass.
- Contract confirmation is by signal: an approved plan that includes the contract IS the confirmation. Confirm up-front only for externally-visible or hard-to-change contracts (public/partner API, event schema, DB-persisted shape, third-party webhook, invented domain semantics); an internal contract with both sides in the same change-group proceeds with the spec written, flagged at close.
- Implement against the agreed contract exactly. If the contract is incomplete, fix the contract first instead of working around it in code.
- If a ticket, spec, wireframe, README, test, or implementation disagree, report the divergence and resolve the source of truth before changing contracts, migrations, or security controls.

## Workspace Artifacts
- Use `_support/` as the canonical support folder; path determines scope: `projects/<group>/<project>/_support/` for workspace material and `projects/<group>/<project>/<repo>/_support/` for repo material.
- Canonical subfolders: `docs/` durable docs, `spec/` contracts/schemas/ADRs, `plan/` implementation plans, `workspace/` temporary notes, `evidence/` screenshots and validation artifacts.
- Do not leave support files loose beside framework source roots; move worthwhile temporary work into `_support/workspace/` or `_support/evidence/`.
- Use repo-level `_support/scripts/` only for disposable dev-session utilities; project-owned scripts belong where the framework expects them — infer the home from existing scripts and repo convention; ask only when no convention exists or two homes are genuinely plausible.
- With a versioned specs repo (`<project>-specs/`): history-worthy material goes there; temporary/sensitive/raw evidence stays in `_support/`; single-repo material goes to `<repo>/_support/`; decision-producing temporary reports get promoted to specs. `_support/` points and expires; specs preserves.
- App-specific infra stays with the app repo. Foundational/shared infra serving 2+ services or none (Terraform/IaC, deployment orchestration, DNS, IAM, shared clusters/DBs/buckets, runbooks, secret templates) goes to a dedicated `<project>-infra` sibling repo, never `<project>-specs`. Single-app projects keep infra in the app repo. Commit only secret templates; real secrets and tfstate stay out of git.
- Normalize legacy support folders opportunistically: `manuals/`/`reference/` → `docs/`; `artifacts/`/`bug-evidence/` → `evidence/`; `context-ia/` → `workspace/`.
- Generated artifacts under `_support/workspace|evidence|archive|plan`: intention-revealing names; ISO `YYYY-MM-DD` dates on immutable snapshots only; `{slug}-YYYY-MM-DD/` folders for multi-file deliverables, loose files for self-sufficient artifacts; ephemeral runs under `_support/workspace/<run-slug>/`; curated evidence under `_support/evidence/<slug>-YYYY-MM-DD/`. README only for what naming cannot express. Retain screenshots as WebP lossless (`cwebp -lossless`; bit-exact, ~−75% on UI shots), updating report refs in the same step; never rewrite git history to shrink committed rasters.
- Version text-that-interprets (reports, ADRs, contracts, session index), referencing binaries by path. Version a binary only when it is a non-reproducible source artifact; QA/in-vivo screenshots are reproducible — version the report, not the images. With a specs repo, only text reaches `<project>-specs/`; validation binaries never.
- Group a session's artifacts under `sessions/YYYY-MM-DD-<slug>/` (execution, by time), distinct from the intention layer (`decisions/contracts/epics/conventions`, by type). Versioned home: `<project>-specs/sessions/`, or `<repo>/_support/sessions/` standalone; raw stays gitignored under `_support/workspace|evidence/` with the same slug. Top-level session files carry the SLUG, not the date (`<slug>-plan.md`); the versioned index is `sessions/README.md`, never the ledger. A session declares `Implements:`; the intention records `Session: <slug>`. A task-by-task plan is execution, never `decisions/`. Multi-session efforts group under an initiative folder `sessions/<start-date>-<slug>/` with dated sub-sessions inside; one-off work stays flat. Full convention: flow-core's specs-structure reference.

## Flow Phase Boundaries
- In a workspace with a project ledger (`_support/PROJECT.md`), scope "what's next?" answers to the ledger's current phase. Validating live state is fine; report each fact in one line.
- Actions owned by another flow phase get a pointer to the owning skill, never that phase's execution plan. Environment promotion, deploys, and post-deploy verification belong to the deploy skill: name it (e.g. `/flow-deploy qa`) and stop.
- Suggestion surfaces (next steps, scope candidates, recommendations) obey the same boundaries: when a skill owns the action, the suggestion is the skill invocation, not the plan.

## Memory Project Resolution (Engram)
- Codex and OpenCode are the current AGENTS consumers for this file, and both are expected to have Engram-style memory available.
- Engram keys memory by a project detected from the working directory; `.engram/config.json` (`project_name`) overrides git detection.
- Multi-repo workspaces (`projects/<group>/<project>/`) fragment per repo or go `ambiguous`. Fix: one unified project `<group>-<project>` declared in `.engram/config.json` at the workspace root and each child repo.
- If the resolved project is `ambiguous` or a per-repo name, pass `project: <group>-<project>` explicitly on saves and searches until that config exists. Derive the name only from a canonical `projects/<group>/<project>/` path; otherwise ask, and never bundle a non-canonical parent of unrelated repos.
- When a flow phase mirrors an intermediate artifact to Engram for resume-after-compaction, key it `flow/{epic-or-project-slug}/{artifact}` so saves upsert instead of duplicating. The ledger remains the source of truth. Retrieve via `mem_search` then `mem_get_observation`.
- Status/pending facts use a deterministic topic_key so saves upsert one living observation per topic, never append a new one each time.
- Record the session slug in any memory about a decision, discovery, or execution (`session: sessions/YYYY-MM-DD-<slug>`, and/or the `flow/{epic}/{slug}` key) so recall points at where the documents live.
- When a remembered fact becomes false, rewrite the old record to the current truth — a "supersedes" relation alone does not hide it, and a related file merely existing does not prove a pending task done. Never leave stale and corrected versions both alive; if mirrored in a native/file layer, fix both.
- Answer "what's pending / where are we" from ground truth verified by claim TYPE: implementation claims → git first (closing commit/PR, merged branch, file on disk); tool/library behavior claims → the docs for the installed version (lockfile/manifest, never latest). A ledger, ticket, or memory note corroborates but never substitutes; on conflict, live state wins. An empty inspection result does not prove "off".
- Consult an issue tracker only when the project declares one (AGENTS.md/CLAUDE.md/ledger) — the declaration is standing authorization to read; ticket writes are outward-facing, so batch every proposed update/close into the session-close confirmation (one approval, never per-ticket asks) unless the config authorizes writes. Flag a detectable-but-undeclared tracker and propose declaring it; do not consult it on sight.
- If a future AGENTS-compatible harness loads this file without Engram, ignore this section rather than simulating memory behavior through unrelated storage.

## Session Execution Mode
- For multi-task specs involving branches, commits, pushes, PRs, or merges, ask the session execution mode before editing unless the user already specified it — mandatory in interactive sessions, never silently assumed. It is a planning question, not a tool-approval request.
- If a structured question tool is available, invoke it with the recommended choice first and `(Recomendado)` suffix; otherwise ask in prose and wait. A session is non-interactive only when no user can answer; approval-bypass flags alone do not qualify. If no mode is stated and questions are unavailable, default to `publish-autonomous`.
- Supported modes: `publish-autonomous` (default: branch, implement, verify, commit, push, open PR; merges stay pending for the user) and `full-autonomous` (also merge PRs within the stated scope).
- You own CI for any PR you merge until the landed branch is green: wait for PR checks before merging, then watch the post-merge branch CI and fix any failure you triggered. Local lint/build is not a substitute for authoritative CI.
- Shared integration branches receive changes via PR, not local direct merge/push, when the project uses a PR gate. Solo repos without PR flow may merge locally.
- Higher-environment promotion moves the integration branch's whole current state, not only the session diff, unless the user explicitly scopes a partial promotion.
- If the user asks to keep work local, treat that as an instruction, not a mode: implement, verify, and commit locally, then close by listing the pending publish steps. Never leave verified work uncommitted.
- If the user already stated the mode, follow it without re-asking in the same task scope. A skill whose body declares its own git scope sets `publish-autonomous` within that scope; run its declared close steps without re-asking. Protected-branch promotions stay gated.
- In `publish-autonomous`, when a task depends on an unmerged prior PR, stack it: branch from the prior task's branch, chain the PRs, declare the merge order in the close report. Never wait for a merge mid-session. When the spec requires sequential merges into a non-protected integration branch, recommend `full-autonomous` in the mode question.
- No mode authorizes force-push, destructive history rewrites, production infrastructure changes, secrets handling, data deletion, or work outside the stated repositories or scope.

## Git
- Do not push, merge, force-push, rewrite published history, or delete branches unless the user explicitly authorizes that operation through a direct instruction or session execution mode. The session-close pruning of confirmed-merged branches (below) is standing-authorized; an explicit verb IS the authorization for that operation.
- At session close, prune branches confirmed 100% merged into their integration target (delete local + merged remote counterpart, return the checkout to the integration branch, report it), and in the same pass `git fetch --prune` and fast-forward local integration branches behind upstream — automatic, recoverable. For each unmerged local branch, check for real un-integrated work (`git log <target>..<branch>`): recommend integrating before closing, or ask if it looks abandoned — never silently leave or delete it. Deleting an unmerged branch, force-delete, unverifiable merge status, or reconciling a diverged integration branch stays gated.
- Do not commit directly to `main` or `master` unless the repo explicitly allows it.
- Follow the repository's documented branching model; do not assume Gitflow. If none is documented, use short-lived branches from the requested or default base and open PRs before merging.
- A branch's unit is a cohesive change-group. Keep one open thematic branch at a time unless concurrent thematic work is explicitly isolated in a worktree.
- On a dedicated work branch, commit autonomously and incrementally — N small conventional commits at natural seams keep the history auditable and bisectable; no per-commit confirmation. This is the default in every repo; a project may declare different semantics (e.g. squash-only) in its CLAUDE.md/AGENTS.md, and absence of a declaration means the default applies.
- Branch-name prefixes are not a branching model; for short-lived branches pick the prefix matching the change type, mirroring Conventional Commit types: `feat/`, `fix/`, `chore/`, `docs/`, `refactor/`, `test/`, `perf/`, `ci/`. Use `develop`, `release/*`, `hotfix/*`, or full Gitflow only when the repo already uses it or the user requests it; for client repos, confirm unclear release/hotfix/cross-branch bases.
- An authorizing session mode plus an explicit git verb authorizes matching branch/commit/push/PR/merge actions within the current task scope. Ask only when details are ambiguous, mode is missing, a safety gate applies, or repo/branch/remote/scope/risk changes.
- Do not run lint, typecheck, build, or tests as part of the git command itself. Run task-level verification separately before the git task.
- Keep commits small and focused, splitting unrelated concerns; use conventional commit and PR titles. PR descriptions cover what changed, why, and the verification performed.
- If a GitHub push returns `404 Repository not found` for a private repo, verify the active account before assuming the repo is missing.
- Never include AI attribution in commits, PRs, messages, or file outputs.
