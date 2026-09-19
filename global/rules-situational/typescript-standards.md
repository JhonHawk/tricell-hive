---
globs:
  - "**/*.{ts,tsx,js,jsx}"
---

## TypeScript & JS Standards

### Type Safety
- **`any` is strictly forbidden.** Use `unknown` and narrow with type guards, assertions, or conditional checks.
- **Explicit return types at module boundaries.** Required on exported functions, exported methods, and React components. Internal helpers may rely on inference unless the inferred type is non-obvious or unstable across refactors. Matches typescript-eslint's `explicit-module-boundary-types`.
- Prefer `type` for unions; `interface` for extendable object shapes. (TS Handbook heuristic: reach for `interface` until you need a `type`-only feature — unions, tuples, mapped/conditional types.)
- Prefer `satisfies` over type assertions.
- **`strictNullChecks` must be enabled.** Never allow implicit `null | undefined` — check explicitly before access.
- **The type system is the source of truth — derive from it and trust it.** A shape that already exists is derived (`Partial`/`Pick`/`Omit`, mapped types), never re-declared by hand (NestJS DTO variants: `nestjs-patterns.md > DTOs & Validation`); a state the declared type excludes is never re-checked at runtime (`!== null` on a `string | undefined` field) — intentional defense-in-depth widens the type or carries a one-line comment naming the unvalidated path; `value != null` is the nullish idiom over `!== null && !== undefined` chains.

### Code Style
- Co-locate types in `.types.ts`, `.models.ts`, or `.dto.ts` within feature folders.
- Prefer named exports over default exports.
- **Avoid barrels in feature code.** Do not create `index.ts` barrels inside feature folders — they slow incremental builds and hide imports. **Exception:** package public APIs (`packages/*/src/index.ts` in a monorepo) and external-facing entry points where a single import path is the contract.
- Import order: external → internal → relative.
- PascalCase: components, interfaces, types, classes. camelCase: variables, functions. SCREAMING_SNAKE_CASE: constants. kebab-case: files. Booleans: `is`/`has`/`can`/`should` prefix.
- JSDoc for public APIs; document "why", not "what".

### Linting & Formatting
- **Always respect existing lint/format configs** (Biome, ESLint, Prettier). Execution scope and timing after implementation: `CLAUDE.md > Build & Lint`.
- **New projects — the user's stack:** frontend repos use **Biome + ESLint**; backend repos use **ESLint + Prettier**. Never introduce a different lint/format tool (Ultracite or otherwise) without asking. Angular is the exception — see `angular-patterns.md > Tooling`.
- **Existing projects:** if configs exist, run them on modified files after implementation. If configs are missing, ask the user whether to add them or skip.
- **If the user chooses to skip:** document the decision in the project's `CLAUDE.md` (e.g., `## Constraints\n- ESLint/Prettier intentionally omitted`) so future sessions don't re-ask.
- **Smell backstop — the deterministic layer.** When configuring lint for a repo, prefer the typed presets: `strict-type-checked` + `stylistic-type-checked` (`no-unnecessary-condition` requires `strictNullChecks`; expect false positives where ORM types overclaim non-nullability), plus cherry-picked `eslint-plugin-sonarjs` (`cognitive-complexity`, `no-identical-functions`, `no-gratuitous-expressions`, `no-selector-parameter`) and `jscpd` in CI for duplication. Biome frontends: `noUnnecessaryConditions` and complexity rules are opt-in, and Biome has no duplication detection — pair with jscpd. Adopting this stack in an existing repo is a deliberate change, never a side effect of another task.

### Runtime Awareness
- **Check the project's Node.js version** (`.nvmrc`, `engines` in `package.json`, or Dockerfile) before using modern syntax.
- `Array.prototype.at()`, `structuredClone()`, `Object.groupBy()` — verify runtime support before using.

### Build & Compilation
- **Keep the incremental build cache in sync with the output it describes.** With
  `incremental`/`composite` TS builds, put `tsBuildInfoFile` inside the cleaned output dir
  (`dist`) — or delete it in the same step that cleans `dist`. A clean that wipes `dist` but
  leaves a stale `.tsbuildinfo` makes `tsc`/`nest build` skip re-emitting "unchanged" files,
  leaving `dist` incomplete: a silent runtime `Cannot find module .../dist/main` behind a
  green build. CI/Docker masks it (clean checkout); only local incremental builds drift.
