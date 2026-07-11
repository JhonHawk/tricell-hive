# Audit checklist

`audit` is **read-only**: report findings, propose fixes, modify nothing. Walk both layers
below. Assign each finding a severity and group the report by it.

## Severity

- **error** — breaks or should break the build (broken internal link, missing required
  frontmatter, second H1 in body, image import that won't resolve).
- **warning** — correctness/convention drift that won't fail the build but should be fixed
  (image in `public/` losing optimization, version drift, missing `sidebar.order` under autogenerate).
- **observation** — style/architecture guidance (voice mismatch, Diátaxis mixing, deep nesting).
  Never blocks; especially the voice rule (usted vs. imperative) is always an observation.

## Layer A — chassis / correctness

1. **Versions**: compared to `chassis.md` pinned set — flag drift (e.g. caret ranges vs. exact
   pins, an off-canon Astro/Starlight pair, TypeScript jumped to 7.x unverified). warning.
2. **Mermaid before Starlight** in `integrations[]`. error if reversed (diagrams break).
3. **`unified()` processor** present iff `starlight-image-zoom` is installed. warning if mismatched.
4. **`starlight-links-validator`** present and wired; run `pnpm build` — broken internal
   links/anchors are errors.
5. **`description` required** in the schema and present on every page. error if missing on a page.
6. **Assets**: content images under `src/assets/` (not `public/`), WebP, non-empty alt text.
   `public/` content image → warning; missing/empty alt on a meaningful image → error.
7. **Locale**: single `root` locale `es-MX`; pages at the docs root (not pre-nested). warning if off.
8. **H1 discipline**: body starts at `##`, no second H1, no skipped heading levels. error.
9. **Sidebar**: hand-curated or, if `autogenerate`, `sidebar.order` set per page. warning.
10. **spec-site only**: `SpecRubric.astro` present, `workflow` schema in `content.config.ts`, every
    spec page carries a `workflow` block with a valid `status`/`stage`. error if a spec lacks status.

## Layer B — architecture / authoring

11. **Diátaxis coherence**: page matches its profile's dominant type; flag pages mixing types. observation.
12. **Tree depth ≤ 2** sub-page levels. observation (warning if clearly labyrinthine).
13. **Headings** unique and descriptive (no generic/duplicate). observation.
14. **Front-loaded purpose**: page/section opens with the distinctive info. observation.
15. **Stacked asides**: consecutive `:::` callouts with no prose between. observation.
16. **State marking**: in-progress pages use `draft`; `lastUpdated` present where the profile expects it. observation.
17. **Screenshots**: not the sole carrier of critical info; no text-as-image; PII covered with a
    solid overlay. observation (warning for exposed PII).
18. **Voice**: consistent with the profile's established convention. observation (never error).
19. **Empty scaffolding**: section trees/pages with placeholder-only content. observation.
20. **Duplication/drift**: the same fact stated in multiple places. observation.

## Report format

```
# Audit — <path> (<profile>)

## Errors (N)
- <file>:<loc> — <check> — <what's wrong> — <proposed fix>

## Warnings (N)
- ...

## Observations (N)
- ...

## Summary
<one paragraph: overall health, top 3 things to fix first>
```

Run `pnpm build && pnpm check` as part of the audit when the site installs — an actual build is
the ground truth for the correctness layer, not a static read.
