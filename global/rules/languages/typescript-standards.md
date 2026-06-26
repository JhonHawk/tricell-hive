---
paths:
  - "**/*.{ts,tsx,js,jsx}"
---

## TypeScript & JS Standards

### Type Safety
- **`any` is strictly forbidden.** Use `unknown` and narrow with type guards, assertions, or conditional checks.
- **Explicit return types at module boundaries.** Required on exported functions, exported methods, and React components. Internal helpers may rely on inference unless the inferred type is non-obvious or unstable across refactors. Matches typescript-eslint's `explicit-module-boundary-types`.
- Prefer `type` for unions; `interface` for extendable object shapes. (TS Handbook heuristic: reach for `interface` until you need a `type`-only feature — unions, tuples, mapped/conditional types.)
- Prefer `satisfies` over type assertions.
- **`strictNullChecks` must be enabled.** Never allow implicit `null | undefined` — check explicitly before access.

### Code Style
- Co-locate types in `.types.ts`, `.models.ts`, or `.dto.ts` within feature folders.
- Prefer named exports over default exports.
- **Avoid barrels in feature code.** Do not create `index.ts` barrels inside feature folders — they slow incremental builds and hide imports. **Exception:** package public APIs (`packages/*/src/index.ts` in a monorepo) and external-facing entry points where a single import path is the contract.
- Import order: external → internal → relative.
- PascalCase: components, interfaces, types, classes. camelCase: variables, functions. SCREAMING_SNAKE_CASE: constants. kebab-case: files. Booleans: `is`/`has`/`can`/`should` prefix.
- **Identifiers in English** — variable, function, type, and file names are always English even in Spanish-domain projects; only domain *values* (enum literals, status strings) may be Spanish. See `CLAUDE.md > Code Layer`.
- JSDoc for public APIs; document "why", not "what".

### Linting & Formatting
- **Always respect existing ESLint and Prettier configs.** After completing an implementation, run lint and format **only on the files you modified** — never on the entire codebase.
- **New projects:** if no `.eslintrc.*`/`eslint.config.*` or `.prettierrc.*`/`prettier.config.*` exists, ask the user whether to configure them or skip. Do not assume either way.
- **Existing projects:** if configs exist, run them on modified files after implementation. If configs are missing, ask the user whether to add them or skip.
- **If the user chooses to skip:** document the decision in the project's `CLAUDE.md` (e.g., `## Constraints\n- ESLint/Prettier intentionally omitted`) so future sessions don't re-ask.

### Runtime Awareness
- **Check the project's Node.js version** (`.nvmrc`, `engines` in `package.json`, or Dockerfile) before using modern syntax.
- `?.` (optional chaining) and `??` (nullish coalescing) are ES2020 — they run natively on Node.js 14+ (V8 8.1); older runtimes need transpilation (TS/Babel down-level), so don't assume native support against older targets.
- `Array.prototype.at()`, `structuredClone()`, `Object.groupBy()` — verify runtime support before using.
- When in doubt, check compatibility against the project's target environment.

### Build & Compilation
- **Keep the incremental build cache in sync with the output it describes.** With
  `incremental`/`composite` TS builds, put `tsBuildInfoFile` inside the cleaned output dir
  (`dist`) — or delete it in the same step that cleans `dist`. A clean that wipes `dist` but
  leaves a stale `.tsbuildinfo` makes `tsc`/`nest build` skip re-emitting "unchanged" files,
  leaving `dist` incomplete: a silent runtime `Cannot find module .../dist/main` behind a
  green build. CI/Docker masks it (clean checkout); only local incremental builds drift.
