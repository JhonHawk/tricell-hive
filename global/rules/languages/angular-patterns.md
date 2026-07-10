---
paths:
  - "**/*.{component,directive,pipe,service,guard,resolver,interceptor,module}.ts"
  - "**/*.component.html"
  - "**/app.config.ts"
  - "**/main.ts"
  - "angular.json"
---

> **Applies when:** `angular.json` exists in the project. Skip if working on NestJS or other non-Angular TypeScript projects.

## Angular

### Tooling
- **Lint with `angular-eslint`, including template linting** (`@angular-eslint/template-parser` over `*.component.html`). Do not adopt Biome in Angular repos — it does not lint Angular templates.

### Version Detection (FIRST STEP — NON-NEGOTIABLE)
- **Check `package.json` → `@angular/core` BEFORE generating any code.** Angular guidance in this repo is version-banded, not one-size-fits-all.

### Control Flow (VERSION-DEPENDENT)
- **v20+**: Use block syntax for new code — `@if`, `@for`, `@switch`, `@defer`.
- **v17-19**: Block syntax is available but optional. Match the surrounding file or feature style; do not force migrations inside unrelated work.
- **v16 and lower**: Use structural directives only — `*ngIf`, `*ngFor`, `*ngSwitch`.
- **Never mix** block syntax with structural directives inside the same template file.

### State Management
- **v17+**: Prefer `signal()` for writable state, `computed()` for derived, and `effect()` for side-effects.
- **v17-19**: Input signals are available, but only use them when the surrounding codebase already leans into the signals API.
- **v20+**: Input/output function APIs are the default direction for new code.
- **`computed()` must be pure** — no side effects, no DOM manipulation, no signal writes. It's the safest and most performant reactive primitive.
- **Never write to signals inside `effect()`** — causes circular dependencies and infinite loops. Effects are for external side-effects only (DOM, logging, API calls). Use `computed()` for derived state instead.
- **v16 and lower**: Use `@Input()`/`@Output()` decorators and RxJS observables for reactive state.

### Change Detection
- **Default to `OnPush` on new components.** With signals (v17+) and especially zoneless mode (developer preview in v18, on track for stable in v19/v20), `OnPush` becomes less critical but remains the safer baseline. Justify exceptions in the file.
- With signal inputs (v17+): automatic updates reduce the need for `markForCheck()`.

### Dependency Injection
- **v16+**: `inject()` is available and preferred when it improves readability. Match the surrounding file style inside mature codebases.
- **v15 and lower**: Constructor injection is standard.

### Observables & Cleanup
- **v17+**: Convert observables to signals with `toSignal()` when the local component state is signal-first.
- **v16 and lower**: Use `AsyncPipe` for display observables. Never `.subscribe()` manually for display data.
- **Unsubscription**: `takeUntilDestroyed()` is available in v16+. Below that, use the established `OnDestroy`/`Subject` cleanup pattern.

### Components
- **v20+**: Standalone components are the default direction for new code.
- **v17-19**: Prefer standalone components when the project already uses them; respect existing NgModule boundaries during incremental migrations.
- **v16 and lower**: NgModule declarations remain the normal baseline unless the project has already opted into standalone.
