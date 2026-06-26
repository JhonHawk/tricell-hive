---
paths:
  - "**/*.{tsx,jsx,html,vue,astro,svelte}"
  - "tailwind.config.*"
  - "**/globals.css"
  - "**/app.css"
  - "**/index.css"
  - "**/styles/**/*.{css,pcss,scss}"
---

> **Activation check.** Path-scoping in Claude Code cannot inspect sibling files, so the globs above will sometimes match a JSX/CSS file in a project that does NOT use Tailwind. When that happens, exit silently — don't apply Tailwind rules to non-Tailwind projects.
> - **v3 signal**: `tailwind.config.*` present, classic `@tailwind` directives in a stylesheet, or `tailwindcss` in `package.json`.
> - **v4 signal**: `@import "tailwindcss"` and/or `@theme` in a stylesheet.
> - If none of those markers exist, ignore this rule even if the file path matched.

## Tailwind CSS
- **Detect, then check version.** First confirm a Tailwind marker is present (per the activation check above); if no marker, exit silently. When detected, identify version from the markers: v4 uses CSS-first `@theme` in stylesheets with `@import "tailwindcss"`; v3 uses `tailwind.config.js` with `@tailwind` directives. Never mix syntaxes.
- **v4 default changes**: `border` defaults to `currentColor` (was `gray-200`), `ring` defaults to `1px` (was `3px`). Add explicit colors/widths when migrating.
- **Always prefer the canonical scale class over arbitrary values.** Tailwind's spacing/sizing scale uses 4px = 1 unit (e.g., `min-w-60` = 240px, `w-80` = 320px, `p-3` = 12px, `gap-1.5` = 6px, `h-10` = 40px). Before writing `min-w-[240px]`, `w-[80px]`, `p-[12px]`: check if a canonical class maps to that value. **Arbitrary values `[...]` (v3) or `(...)` (v4) are reserved for values that don't fit the scale** (e.g., `top-[37px]`, `w-[clamp(...)]`, `bg-[url('...')]`). The IDE's `suggestCanonicalClasses` inspection flags violations — treat its suggestions as authoritative and convert on touch.
- **Conditional classes:** use `clsx` + `tailwind-merge` (or `cn()` helper). Never concatenate class strings manually.
