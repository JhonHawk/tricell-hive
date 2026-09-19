---
globs:
  - "**/*.{tsx,jsx}"
  - "next.config.*"
---

> **Applies when:** `react` is in `package.json` dependencies or `next.config.*` exists. Skip for Astro, Preact, or Solid projects that also use `.tsx`.
> **Plain React (no `next` in package.json — Vite, React Router, CRA legacy):** the Next-prescriptive sections below (Architecture routers/RSC/Server Actions, `next/image`, caching) do NOT apply; follow the project's own router and the framework-agnostic sections (Version Detection for React, Hooks, State, Preferred Libraries). Never introduce Next.js conventions (RSC, file routing, `use server`) into a plain React app.

## React / Next.js

### Version Detection (FIRST STEP — NON-NEGOTIABLE)
- **Check `package.json` → `next` and `react` BEFORE generating code.** Guidance below is version-banded: v14 (App Router stable but no async params), v15 (async `params`/`searchParams`, React 19 minimum), v16 (Turbopack as default, typed `PageProps` helper, `experimental.turbopack` removed). React behaviour also differs: v18 vs v19 form APIs, `useOptimistic`/`useActionState` availability.
- **Mixed-router codebases** (App Router + Pages Router during migration): match the touched file's router; don't migrate adjacent routes.

### Architecture
- **New routes**: App Router by default. Existing Pages Router code may coexist during incremental migration — do not rewrite unrelated routes just to enforce the new router.
- **Server Components by default** — `'use client'` only when state, effects, or browser APIs are needed.
- **Server Actions** are preferred for same-app form mutations and internal workflows. **Route Handlers** are appropriate for webhooks, third-party integrations, explicit HTTP APIs, streaming, or any case where you genuinely need an HTTP boundary.
- **Push `'use client'` to the leaves** of the component tree. A page can be a Server Component that passes data to thin Client Component shells for interactivity.
- **Project preference**: avoid fetching from your own Route Handlers in Server Components when you can call the underlying logic directly.
- **React Context is incompatible with Server Components** — use module-scope alternatives or pass data as props.

### Data Fetching & Caching
- **Server Components fetch directly** — no client-side fetching library needed. Fetch is auto-memoized within a render pass.
- **Cache strategies**: since Next 15, bare `fetch` is NOT persistently cached by default — `cache: 'force-cache'` opts in (static), `cache: 'no-store'` forces fresh, `next: { revalidate: N }` is ISR.
- **Tag-based revalidation**: Tag fetches with `next: { tags: ['resource'] }`, invalidate with `revalidateTag()` or `revalidatePath()`. **Always revalidate after mutations.**

### Streaming & Suspense
- **Use `<Suspense>` boundaries** for granular streaming — don't rely solely on `loading.tsx` for everything.
- **`loading.tsx`** is for route-level fallback. Use Suspense for finer-grained independent loading states.
- **`error.tsx`** must be a Client Component (`'use client'`). `global-error.tsx` must include `<html>` and `<body>`.

### Metadata & Assets
- **Metadata**: `generateMetadata` or static `metadata` export, not `<Head>`. Use `ResolvingMetadata` to extend parent metadata.
- **Images**: Always `next/image`, never raw `<img>` — handles responsive sizing, lazy loading, format conversion.
- **Fonts**: Self-host with `next/font/google` or `next/font/local` in root layout. Prevents FOUT and external requests.

### React 19+ / Next.js 16
- **Default for new projects**: Next.js 16 (current stable). React 19+ is the minimum since v15. Treat v15 as legacy-acceptable only for existing codebases mid-migration; do not start new projects on v15.
- **Async route props**: `params` and `searchParams` are Promises (since v15). On v16, prefer the typed helper `PageProps<'/route/[slug]'>` (run `next typegen` to generate). Fall back to manual `await` (Server Components) or `use()` (Client Components) when the helper is not yet generated.
- **Turbopack is default in v16** and configured at the top level of `next.config.ts` as `turbopack: { ... }`. The `experimental.turbopack` location was removed — do not introduce it in new configs and migrate it on touch.
- **`next dev --turbopack` worker runaway can OOM the machine** — dev-only; serve visual/in-vivo checks from a production build, one app at a time (full mechanics: global `Execution` rule).
- **`useActionState`** for form state with server validation — replaces manual `useState` + `useTransition` for forms.
- **`useOptimistic`** for optimistic UI updates while Server Actions complete.
- **Verify version-specific APIs against context7 before implementing** — `context7.md` mandates this for any Next.js feature tied to a major version.

### Type Imports
- **Never use `React.*` globals.** Always use explicit type imports: `import type { ReactNode } from "react"`.

### Preferred Libraries
- `@tanstack/react-query` — client-side server state. Prefer over `useEffect` + `useState` for data fetching. The latter is acceptable only for trivial demos, single-shot reads in throwaway code, or non-server data (URL hash, scroll position, IntersectionObserver, MediaQueryList).
- `@tanstack/react-table` — headless tables with sorting/filtering/pagination.
- `react-hook-form` + `@hookform/resolvers/zod` — forms. Preferred over formik or manual controlled inputs.
- `zustand` — client global state. Preferred over Redux/RTK when lightweight state is sufficient.
