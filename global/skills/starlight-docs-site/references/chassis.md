# Chassis — shared technical canon

The ~70% both profiles share. Templates in `templates/chassis/` and each profile
dir. Versions are **pinned exact** for reproducible replication.

## Stack & pinned versions

| Piece | Choice | Version |
| --- | --- | --- |
| Generator | Astro + Starlight | `astro` 7.0.7, `@astrojs/starlight` 0.41.3 |
| Diagrams | astro-mermaid + mermaid | `astro-mermaid` 2.1.0, `mermaid` 11.16.0 |
| Images | Sharp (Astro built-in optimizer) | `sharp` 0.35.3 |
| Link check | starlight-links-validator | `starlight-links-validator` 0.25.2 |
| Zoom (user-manual only) | starlight-image-zoom | `starlight-image-zoom` 0.15.0 + `@astrojs/markdown-remark` 7.2.1 |
| Format/lint | Biome (JS/TS/JSON/CSS only) | `@biomejs/biome` 2.5.3 |
| Type/content check | astro check | `@astrojs/check` 0.9.9, `typescript` 6.0.3 |
| Git hooks (opt-in at scaffold) | lefthook 2.x `jobs` | `lefthook` 2.1.10 |
| Package manager | pnpm (exclusive) | `pnpm@11.8.0` |

- **TypeScript stays at 6.0.3** even though 7.x is released: the 6→7 native-port
  jump is unverified against `@astrojs/check` 0.9.9. Bumping is a deliberate decision.
- **Re-verify latest-compatible on any new scaffold** (`npm view <pkg> version`) —
  but bump majors deliberately, never as a silent side effect. Astro ↔ Starlight
  upgrade in lockstep (Starlight 0.41 requires Astro 7); check plugin compatibility.

## Config invariants

- **`astro-mermaid` runs BEFORE `starlight()` in `integrations[]`** — its mdast/rehype
  transform must process the diagram code fences before Starlight's pipeline. `autoTheme:
  true` follows `html[data-theme]`. Brand colors set via `mermaidConfig.themeVariables`.
- **`markdown.processor: unified()`** is set ONLY in `user-manual` (image-zoom does not
  support Astro 7's default Sätteri processor yet — issue #63). `spec-site` omits it.
- **Plugins**: `starlightLinksValidator({ errorOnRelativeLinks: false })` in both (fails
  the build on broken internal links — Starlight core does not check links). `user-manual`
  also adds `starlightImageZoom()`.
- **Locale**: monolingual root — `defaultLocale: "root"`, `locales: { root: { label:
  "Español", lang: "es-MX" } }`. Root-locale pages live directly under `src/content/docs/`
  (no `/es/` prefix), so adding a language later is a config diff, not a restructure.
- **`content.config.ts`**: `docsLoader()` + `docsSchema()`. `description` is REQUIRED via
  `extend` (SEO + no metadata drift). `spec-site` also extends with the `workflow` rubric schema.
- **Sidebar**: hand-curated, grows one section at a time (never scaffold empty groups). If
  `autogenerate` is ever used, set `sidebar.order` on every page (default sort is alphabetical).

## Theme strategy (do not redesign — rebrand only)

`src/styles/theme.css` overrides Starlight's `--sl-color-*`. **Dark = `:root`**, **Light =
`:root[data-theme="light"]`**. Neutrals are OKLCH chroma-0; the accent is an OKLCH family
(low/accent/high) tuned per theme. To rebrand, replace ONLY the `REPLACE:` accent lines
(dark, light, and the fixed Mermaid node fill hex) — the Mermaid framing, figure/figcaption,
and image-zoom backdrop blocks are brand-neutral and stay. Typography is Inter via `--sl-font`.

## Assets

- Content images → `src/assets/` (Astro/Sharp optimizes them). **Never `public/`** for content
  images (bypasses optimization). `public/` only for favicon/robots/verbatim files.
- Format **WebP**; reference with relative Markdown paths; alt text is a full functional
  Spanish description. Grouped to mirror the content tree; `NN-slug.webp` for narrative order.

## Quality gates

- Local: Biome on touched JS/TS/JSON/CSS (`pnpm lint` / `pnpm format`). Markdown/MDX is
  reviewed by reading, not formatted.
- CI/pre-push order: `astro check` → `astro build` → link validation (the validator runs
  inside the build). Search (pagefind) is build-only and prerender-only — keep the site static.

## Optional deploy layer (documented, not forced)

- **Git hooks** (`lefthook.yml`): asked at scaffold. pre-commit Biome on staged, pre-push `pnpm check`.
- **Basic Auth** (`middleware.ts`, Vercel Edge): asked at scaffold. Reads `DOCS_BASIC_AUTH_USER`
  / `DOCS_BASIC_AUTH_PASSWORD` — env vars only, never committed. 401 on mismatch, 503 if unset.
- **Branches**: `development` → `master` (Vercel production branch). Never push directly to `master`.
