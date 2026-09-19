---
commands:
  - "pnpm add"
  - "pnpm install +"
  - "npm install +"
  - "npm i +"
  - "pnpm i +"
  - "npm add"
  - "yarn add"
  - "bun add"
  - "uv add"
  - "uv pip install"
  - "cargo add"
  - "go get"
---

When working with libraries, frameworks, or APIs — use Context7 MCP to fetch current documentation instead of relying on training data.

## Query on version-sensitivity signals

Query context7 before writing code whose exact shape depends on the library version:

- **Installing or integrating a new dependency** — validate plan, config, and examples against the docs for the version being installed.
- **File conventions tied to a framework version** (`middleware.ts` vs `proxy.ts`, App Router layout, server actions, route handlers) — in an existing repo the disk is the primary signal (the convention already chosen is visible); query the docs when CREATING the convention or when the framework major changed.
- **Routing, data-fetching, or rendering APIs on a recent major** (or bumped this session): Next.js 15→16 async params, Angular signals, React 19 form actions.
- **Auth flows, middleware, transaction boundaries** — configuration shape changes enough across versions that training data is unreliable.
- **Deprecation warnings or build errors naming framework conventions** — pin the installed version from the lockfile FIRST, then query context7 anchored to that version, then explore the codebase. Include the version in the query.

**The query that counts is the write-time one.** The obligation attaches to the exact signature, options/props, and config shape at the moment you write the version-sensitive code. A prior query — planning-phase or earlier in the session — discharges it only when it demonstrably returned that exact detail (same surface, same version); a pattern-level answer ("use App Router") does not cover the call detail you then write. Re-query on signal: the detail isn't in what you already fetched, or the build/typecheck failed on that surface.

### Does NOT apply to

- **Stable library on a stable surface.** APIs that haven't changed in years (`useState`, `Promise.all`, Express `app.get()`, Spring `@Service`, standard Zod schemas): anchor to the lockfile version and write directly.
- **Language-level constructs.** Control flow, loops, string manipulation, standard data structures — no framework involved.
- **Internal codebase navigation.** Grep/Read/Explore for "where is X used in this repo"; context7 is for upstream docs, not local code.

## Anchor to the right version — by phase

Never anchor to the version you remember from training data (it lags releases, and majors rewrite syntax wholesale — lefthook 1.x `commands:` → 2.x `jobs:`).

- **New dependency, integration, or upgrade decision:** determine the current latest stable from an authoritative source (`npm view <pkg> version`, the registry, context7's resolved version) AND whether it is compatible with the project as it stands (peer deps, engine/runtime, the framework major in use). Greenfield with no version mentioned defaults to latest.
- **Working on an already-installed dependency:** do NOT look up latest. Anchor every query, example, and config to the lockfile/manifest version — latest induces APIs the project doesn't have. Bumping is a separate, deliberate decision, never a side effect of implementing a feature.

## Mechanics

`resolve-library-id` (prefer exact names and version-specific IDs; retry alternate spellings — `next.js`, not `nextjs`) → `query-docs` with the question → answer citing the version.

## Fallback

Context7 returns nothing or is unavailable → web search. Neither available → anchor to the lockfile/manifest version, proceed, and declare the assumed version prominently in the close report — the declared assumption does not replace the missing docs, so flag the unverified surface too. A hard stop remains only for framework-critical patterns (auth flows, middleware, transaction management) with NO docs AND NO version anchor (greenfield without a manifest) — and that single question folds into the plan gate, never a mid-run stop.
