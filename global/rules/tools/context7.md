---
alwaysApply: true
---

When working with libraries, frameworks, or APIs — use Context7 MCP to fetch current documentation instead of relying on training data. This includes setup questions, code generation, API references, and anything involving specific packages.

## Mandatory Triggers

Query context7 BEFORE writing implementation code when the surface is version-sensitive:

- **Installing or integrating a new dependency** — even a stable library may ship new defaults or breaking changes between versions.
- **File conventions tied to a framework version** — `middleware.ts` vs `proxy.ts`, App Router vs Pages Router layout, Angular standalone vs NgModule, server actions, route handlers.
- **Routing, data-fetching, or rendering APIs** when the project's framework is on a recent major (or was bumped this session). Examples: Next.js 15→16 async params, Angular signals, Spring Boot 3 record controllers, React 19 form actions.
- **Auth flows, middleware, transaction boundaries.** Configuration shape changes enough across versions that training data is unreliable.
- **Deprecation warnings or build errors** that reference framework conventions — include the version in the query.
- **Writing the call, not just choosing the pattern.** Querying at planning time does not discharge the obligation at implementation time. Re-query for the exact signature, options/props, and config shape at the moment you write the version-sensitive code. A pattern confirmed while planning (App Router, Angular signals, Spring record controllers) is a different surface from the API detail you then write (the exact async-params signature, the option keys, the config block) — the latter is where first-try failures and silent omissions come from.

### Does NOT apply to

- **Stable APIs that haven't changed in years.** `useState`, `useEffect`, `Promise.all`, Express `app.get()`, Spring `@Service`, basic class definitions, standard Zod schemas — write directly.
- **Language-level constructs.** Control flow, loops, math, string manipulation, standard data structures — no framework involved.
- **Same API surface, same version, same session.** Reuse a prior answer only for the *same surface* — the same hook, component, config block, or call signature — at the same version. Scope the carve-out to the surface, not the whole library: a planning-phase query about which pattern to use does NOT cover the implementation-phase detail of the call you then write. That detail is a new surface — re-query it.
- **Internal codebase navigation.** Use Grep/Read/Explore for "where is X used in this repo"; context7 is for upstream docs, not local code.

## Priority Over Exploration

- When investigating framework-specific errors: **query context7 FIRST**, then explore the codebase. Do not use Explore agents to research framework conventions — that's what context7 is for.
- Never assume file conventions (`middleware.ts`, `proxy.ts`, `app/api/` structure) — verify against current docs for the detected version.
- If a build log or error message references a specific framework version, include that version in your context7 query.

## Anchor to the right version — by phase

The correct anchor depends on the phase. Never anchor to the version you remember from training data either way (it lags releases, and majors rewrite syntax wholesale, e.g. lefthook 1.x `commands:` → 2.x `jobs:`).

- **New dependency, integration, or upgrade decision.** Determine the current latest stable from an authoritative source (`npm view <pkg> version`, `pip index versions <pkg>`, the registry, or context7's resolved version) AND whether it is compatible with the project as it stands now (peer deps, engine/runtime, the framework major already in use). Validate plan, config, and examples against the docs for the version you will install. Greenfield with no version mentioned defaults to latest.
- **Working on an already-installed dependency.** Do NOT look up latest. Read the installed version from the lockfile or `package.json`/manifest, and anchor every context7 query, example, and config to THAT version. Latest is noise here — it induces APIs the project doesn't have; bumping the version is a separate, deliberate decision, never a side effect of implementing a feature.

If you must proceed unverified, state which version you assumed so the user can catch the drift.

## Steps

1. Call `resolve-library-id` with the library name and the user's question
2. Pick the best match — prefer exact names and version-specific IDs when a version is mentioned. If results look off, retry with alternate names or a rephrased query (e.g., `next.js` not `nextjs`)
3. Call `query-docs` with the selected library ID and the user's question
4. Answer using the fetched docs — include code examples and cite the version

## Fallback

If context7 returns no results or is unavailable: use web search as fallback. If neither is available, proceed with training data but explicitly state the version assumptions being made so the user can verify. For framework-critical patterns (auth flows, middleware, transaction management), do not proceed with training-data assumptions alone — ask the user to confirm the target version before implementing.
