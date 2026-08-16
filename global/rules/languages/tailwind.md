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
- **Always prefer the canonical scale class over arbitrary values.** The spacing/sizing scale is 4px = 1 unit (`p-3` = 12px, `h-10` = 40px, `min-w-60` = 240px). Before writing `min-w-[240px]` or `p-[12px]`, check whether a canonical class maps to that value. The IDE's `suggestCanonicalClasses` inspection flags violations — treat its suggestions as authoritative and convert on touch.
- **v4's scale is dynamic: every number resolves, decimals included.** Divide the pixel value by 4 and write it — `w-[50px]` → `w-12.5`, `min-h-[550px]` → `min-h-137.5`, `p-[3px]` → `p-0.75`. An odd-looking decimal is the correct class, not a typo. v3 resolves only its fixed set of keys, so an off-scale value stays arbitrary there.
- **Arbitrary values use `[...]` in both v3 and v4** — reserved for what the scale cannot express: `w-[calc(100%-2rem)]`, `w-[clamp(...)]`, `bg-[url('...')]`, and non-numeric scales (`text-[10px]` — the text scale is semantic: `text-sm`, `text-base`). v4's `(...)` is only the CSS-variable shorthand: `bg-(--brand)` ≡ `bg-[var(--brand)]`.
- **Opacity modifiers are percentages, not fractions:** `bg-black/50`, never `bg-black/[0.5]`.
- **v4 replaces common arbitrary selectors with variants:** `[&:has(…)]:` → `has-[…]:`, `[&>*]:` → `*:`, `[&_*]:` → `**:`.
- **Conditional classes:** use `clsx` + `tailwind-merge` (or `cn()` helper). Never concatenate class strings manually.
