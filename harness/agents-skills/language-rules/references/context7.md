
## Context7 — version-correct docs in one query

The context7 docs tool (MCP) returns the API surface of the version the project's lockfile/manifest pins: signatures, options and props, config shapes, migration notes. It replaces a web search for any library, framework, SDK, API or CLI surface, and it is both faster and version-correct where training memory is neither. Reach for it first; web search is the fallback, never the peer.

**The asymmetry is the threshold, not your confidence.** One unnecessary query costs seconds; one wrong signature costs a build failure or a silent bug. At the edge, query.

## The moments that call for it

Acts, each visible in the transcript:

- **Writing a call, decorator, config object, or schema against a library the project already has** — anchored to the installed version, never to latest.
- **Picking options, props, flags, or CLI arguments** — the names and the defaults are what move between versions.
- **Touching a version-banded API:** Angular signals, Nest v10→v11 (Express 5), Next App Router and async params, React 19 form actions, Prisma/Drizzle client surfaces, Playwright routing, provider structured-output parameters.
- **A typecheck or build failure naming a library surface** — pin the installed version from the lockfile FIRST, then query anchored to that version, then explore the codebase. Include the version in the query.
- **Installing or integrating a new dependency** — validate plan, config, and examples against the docs for the version being installed.
- **File conventions tied to a framework version** (`middleware.ts` vs `proxy.ts`, App Router layout, server actions, route handlers) — in an existing repo the disk is the primary signal (the convention already chosen is visible); query when CREATING the convention or when the framework major changed.
- **Auth flows, middleware, transaction boundaries** — configuration shape changes enough across versions that training data is unreliable.
- **Before PROPOSING a library-specific approach in an answer** — a recommendation carries the same version risk as the code, and it is acted on the same way.

**A pattern-level answer you already hold does not cover the call detail you are about to write.** "Use App Router" is not the signature of the function you are typing.

**The query that counts is the write-time one.** The obligation attaches to the exact signature, options/props, and config shape at the moment you write the version-sensitive code. A prior query — planning-phase or earlier in the session — discharges it only when it demonstrably returned that exact detail (same surface, same version). Re-query on signal: the detail isn't in what you already fetched, or the build/typecheck failed on that surface.

Enforcement: prompt-convention. The hook holds this file on a dependency-install command; nothing verifies that a given query happened.

### Does NOT apply to

- **Stable library on a stable surface.** APIs that haven't changed in years (`useState`, `Promise.all`, Express `app.get()`, Spring `@Service`, standard Zod schemas): anchor to the lockfile version and write directly.
- **Language-level constructs.** Control flow, loops, string manipulation, standard data structures — no framework involved.
- **Internal codebase navigation.** Grep/Read/Explore for "where is X used in this repo"; context7 is for upstream docs, not local code.

## Anchor to the right version — by phase

Never anchor to the version you remember from training data (it lags releases, and majors rewrite syntax wholesale — lefthook 1.x `commands:` → 2.x `jobs:`).

- **New dependency, integration, or upgrade decision:** determine the current latest stable from an authoritative source (`npm view <pkg> version`, the registry, context7's resolved version) AND whether it is compatible with the project as it stands (peer deps, engine/runtime, the framework major in use). Greenfield with no version mentioned defaults to latest.
- **Working on an already-installed dependency:** do NOT look up latest. Anchor every query, example, and config to the lockfile/manifest version — latest induces APIs the project doesn't have. Bumping is a separate, deliberate decision, never a side effect of implementing a feature.

## Mechanics

`resolve-library-id` — it takes BOTH the library name AND the question (a name-only call is rejected); prefer exact names and version-specific IDs, retry alternate spellings (`next.js`, not `nextjs`) → `query-docs` with the library id and the question → answer citing the version.

## Fallback

Context7 returns nothing or is unavailable → web search. Neither available → anchor to the lockfile/manifest version, proceed, and declare the assumed version prominently in the close report — the declared assumption does not replace the missing docs, so flag the unverified surface too. A hard stop remains only for framework-critical patterns (auth flows, middleware, transaction management) with NO docs AND NO version anchor (greenfield without a manifest) — and that single question folds into the plan gate, never a mid-run stop.
