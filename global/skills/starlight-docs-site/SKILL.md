---
name: starlight-docs-site
description: >
  Build, extend, or audit an Astro Starlight documentation site — a user manual
  (manual de usuario) or a specs/BRD site (docs-as-code). Use when scaffolding a
  new Starlight docs site, adding a page/section (rebanada) to one, or checking an
  existing one against the house convention. Triggers: "manual de usuario con
  Starlight", "sitio de specs", "agrega una página al manual", "audita el sitio de
  docs", or editing files in a Starlight project (astro.config, src/content/docs).
user-invocable: true
paths: "**/astro.config.*,**/src/content/docs/**"
argument-hint: "[scaffold|page|audit] <path> [user-manual|spec-site]"
---

# Starlight docs site

Captures the house convention for Astro Starlight documentation sites so it is
applied organically — you should not have to re-explain it each time. Two content
**profiles** share one technical **chassis**; three **modes** create, extend, or
verify a site. Templates live in `templates/`; the detailed canon lives in
`references/` — read the reference for the mode/profile you are in before acting.

Post-deploy, this skill lives at `~/.claude/skills/starlight-docs-site/`; refer to
bundled files by relative path (`templates/...`, `references/...`).

## Language rule (always)

Identifiers are English (config keys, component names, frontmatter fields). Page
**content is Mexican Spanish** with correct accents. Slugs follow the profile's
established practice (see the profiles). This mirrors the global code-vs-prose split.

## Profiles

| Profile | For | Dominant Diátaxis type | Detailed canon |
| --- | --- | --- | --- |
| `user-manual` | Operator/QA manual: screens, flows, roles | how-to + reference | `references/profile-user-manual.md` |
| `spec-site` | Specs/BRD site: business rules, gates | reference + explanation | `references/profile-spec-site.md` |

Detect the profile from the target when not given: a `starlight-image-zoom`
dependency, screenshots, or role×capability tables → `user-manual`; a `workflow`
frontmatter block + `SpecRubric.astro` → `spec-site`. Ambiguous → ask.

## Modes

Read `references/chassis.md` for the shared technical canon (versions, config,
theme strategy) and `references/best-practices.md` for the authoring/quality layer
before generating content.

### `scaffold <path> [profile]`

Bootstrap a new site. Honor **grows-by-slice**: create the chassis + a home page +
ONE example page, never empty section trees.

1. Confirm the profile (ask if not given). Copy `templates/chassis/*` (theme.css,
   biome.json, tsconfig.json, .gitignore) and the profile's `package.json`,
   `astro.config.mjs`, `content.config.ts` into `<path>`. Put content under
   `src/content/docs/`, the theme under `src/styles/theme.css`.
2. Replace every `REPLACE:`/`<Project>` marker: site title, brand accent family in
   theme.css (dark + light + the Mermaid node fill), `name` in package.json.
3. Seed content:
   - `user-manual`: `index.mdx` (home) + one page from `page-template.md` at the
     slug referenced in the sidebar (e.g. `introduccion/que-es`).
   - `spec-site`: `index.mdx` (home) + `product-workflow.mdx` at
     `workflow/product-workflow`, and copy `SpecRubric.astro` into `src/components/`.
4. **Ask two setup questions** (fold into one block):
   - **Git hooks (lefthook)?** If yes, add `lefthook` to devDependencies, a
     `"prepare": "lefthook install"` script, and copy `templates/chassis/lefthook.yml`.
   - **Basic Auth (Vercel)?** If yes, copy `templates/chassis/middleware.ts` and
     document the `DOCS_BASIC_AUTH_USER`/`DOCS_BASIC_AUTH_PASSWORD` env vars (never commit values).
5. `pnpm install`, then verify from a production build: `pnpm build && pnpm check`
   (not the dev server — search/pagefind is build-only). Report the result.

### `page <path>`

Add one page/slice to an existing site. Detect the profile from the site.

1. Instantiate the profile's page template (`page-template.md`, or
   `brd-template.mdx` / `module-index.mdx` for `spec-site`) under `src/content/docs/`.
2. Fill placeholders; keep the profile's fixed section skeleton and voice.
3. **Add the page's entry to the `sidebar` in `astro.config.mjs`** (grows-by-slice —
   one section at a time). Add screenshots as WebP under `src/assets/` only when real.
4. `pnpm build && pnpm check` (the links validator gates broken internal links).

### `audit <path>`

Validate an existing Starlight site against the canon. **Read-only: report, don't
modify.** Follow `references/audit-checklist.md` exactly — it lists every check, its
severity (build error vs. style observation), and the report format. Covers three
static layers: chassis/correctness, architecture/authoring, and theme/styles (theme
strategy, token discipline, and a computed WCAG-AA contrast check via
`scripts/contrast-check.py`). Rendered visual review (browser screenshots) is out of
scope of the static audit. Group findings by severity; the voice rule (usted vs.
imperative) is an observation, never an error.

## Reference map

| File | What it holds |
| --- | --- |
| `references/chassis.md` | Shared technical canon: pinned versions, config, theme strategy, locale, assets, quality gates, optional deploy |
| `references/profile-user-manual.md` | Manual archetype: 6-section page skeleton, role×capability tables, aside/screenshot patterns, voice |
| `references/profile-spec-site.md` | Specs archetype: BRD anatomy, gates/states, the workflow rubric, voice |
| `references/best-practices.md` | Research-backed quality layer (Diátaxis, structure, authoring, screenshots, anti-patterns) with sources |
| `references/audit-checklist.md` | The `audit` checklist: every check, severity, and report format |
