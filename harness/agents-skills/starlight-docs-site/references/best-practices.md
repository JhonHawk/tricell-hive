# Best practices — the quality layer

Research-backed rules for structure, authoring, and correctness. **Not UI/branding** — the
theme is fixed. Applied when generating content and checked in `audit`. Sources at the end.

## Diátaxis — classify every page

Four distinct user needs, one per page. Two axes: action vs. cognition, acquisition vs.
application. Classify with the compass, per page — never mix types on one page.

- **user-manual** → **how-to** (task procedures) dominant + **reference** (role/state/error
  tables for lookup). An onboarding "first full flow" can be a tutorial. Explanation is marginal.
- **spec-site** → **reference** (the contract: rules, acceptance criteria) dominant +
  **explanation** (rationale/decisions, kept separate). How-to/tutorial barely apply.

## Structure & authoring (enforce / flag)

- **H1 = frontmatter `title`; body starts at `##`**, never skip levels, never a second H1 in body.
- **Front-load purpose**: the first sentence of each page/section carries the distinctive info.
- **Unique, descriptive headings** (no generic "Introducción"/"Detalles") — scannability + deep-links.
- **Tree depth ≤ 2 levels** of sub-pages; finer hierarchy is headings, not more nesting.
- **Slugs**: kebab-case, short, stable (renaming breaks URLs); dates only on immutable snapshots.
- **Explicit state**: `draft: true` (Starlight drops it from the production build) for in-progress
  pages; `lastUpdated` as a freshness signal. spec-site does this via the gate rubric.
- **Required sections make completeness verifiable** (Good Docs Project): the profile skeletons
  are the required set; the generator flags missing required sections.

## Voice (per Diátaxis type — configurable)

- how-to/tutorial → imperative, 2nd person ("Selecciona…", "Verifica…"); reference → declarative,
  neutral; explanation → argumentative.
- **Divergence note**: `ark` user-manual uses formal "usted" descriptive; Google/Microsoft
  recommend imperative 2nd person for manuals. **Respect the project's established convention by
  default** (usted for the manual); expose voice as a configurable rule. `audit` reports a voice
  mismatch as an **observation, never an error**.

## Screenshots & diagrams (discipline, not redesign)

- Only when they add over the text; **the text must suffice alone** (critical info never depends
  on an image). Never text/code/terminal output as an image — use real, searchable text.
- **Alt text** = purpose in context (capitalized, period). Purely decorative → `alt=""`.
- Prefer architecture diagrams (Mermaid, age well) over UI screenshots where possible.
- `NN-slug.webp` narrative order, WebP in `src/assets/` (Sharp optimizes). Cover PII with a solid
  100%-opacity overlay, never blur/mosaic.

## Maintainability

- **Grow incrementally, never scaffold empty trees** — structure emerges from healthy content.
- **Single source of truth**: each fact in one place; everything else links. Duplication → drift.
- Update docs in the same change-group as the code they describe.

## Anti-patterns (audit flags these)

Mixing Diátaxis types on one page · catch-all "dumping ground" sections · wall-of-text hiding
ambiguity · empty scaffolding · organizing by audience instead of content type · nesting > 2
levels · depending on screenshots for critical info · text-as-image · duplicated facts · generic/
duplicate headings · docs out of sync with code · stacked consecutive asides.

## Not baked in prematurely

Doc versioning: Starlight has **no** official versioning. Default recommendation = **branch-based**
(a branch/deploy per version); `starlight-versions` (plugin, early-stage) only if multi-version is
a hard requirement. Do not scaffold either by default.

## Sources

- Diátaxis — diataxis.fr (start-here, compass, reference-explanation).
- Google developers style guide — headings, paragraph-structure, highlights, images, alt-text.
- Microsoft Writing Style Guide — top-10-tips, person, accessibility/alt-text.
- The Good Docs Project — templates. Write the Docs / GitBook / Document360 — information architecture.
- Starlight docs (v0.41.x) via context7 — sidebar, frontmatter, i18n, site-search, authoring-content,
  components. `starlight-links-validator`; versioning discussions #957/#372; `starlight-versions` plugin.
