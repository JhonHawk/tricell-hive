# Audit checklist

`audit` is **read-only**: report findings, propose fixes, modify nothing. Walk the layers
below. Assign each finding a severity and group the report by it.

## Severity

- **error** — breaks or should break the build (broken internal link, missing required
  frontmatter, second H1 in body, image import that won't resolve).
- **warning** — correctness/convention drift that won't fail the build but should be fixed
  (image in `public/` losing optimization, version drift, missing `sidebar.order` under autogenerate).
- **observation** — style/architecture guidance (voice mismatch, Diátaxis mixing, deep nesting,
  colors bypassing the token system). Never blocks; especially the voice rule (usted vs.
  imperative) is always an observation.

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

## Layer C — theme / styles (static)

Convention + contrast, not aesthetics — validates that the theme follows the house strategy and
is accessible. No browser (that is the out-of-scope rendered review below). Anchors:
`references/chassis.md > Theme strategy` and `~/.claude/rules/languages/ui-visual-design.md > Color`.

21. **Theme strategy**: `src/styles/theme.css` exists and overrides `--sl-color-*` with **Dark =
    `:root`** and **Light = `:root[data-theme="light"]`**. error if the file is missing or the
    dark/light split is absent.
22. **Accent replaced**: no leftover `REPLACE:` markers or the template's neutral placeholder
    accent (`oklch(70% 0.16 260)` / node fill `#4f6bed`) — those mean the site was never
    rebranded. warning.
23. **Token discipline**: colors come through `--sl-color-*` overrides, not hardcoded hex/rgb that
    bypasses the token system. The fixed Mermaid node-fill hex is exempt (brand identity by
    design, per chassis). observation.
24. **Mermaid framing**: theme-aware `.mermaid` overrides present (panel background/border via
    Starlight variables), so diagrams read in both themes. warning if absent.
25. **customCss wired**: `customCss: ["./src/styles/theme.css"]` set in `astro.config.mjs`. error
    if the theme file exists but is not wired.
26. **Contrast (WCAG AA)**: compute the accent-vs-background ratio for both themes and report it.
    Pull `--sl-color-accent` (and `--sl-color-accent-high` in light) and the background
    (`--sl-color-black` in dark; the light `--sl-color-black`/surface in light) from `theme.css`,
    then run the bundled helper:
    ```bash
    uv run --with coloraide python scripts/contrast-check.py "<accent>" "<background>"
    ```
    (OSV-check `coloraide` on first use — PyPI, clean as of this writing.) Ratio < 4.5:1 for
    body-size link text → warning; always report the ratio. Never meaning by color alone.
27. **image-zoom backdrop**: `--starlight-image-zoom-backdrop-bg` present iff `starlight-image-zoom`
    is installed. observation.

**Out of scope here — rendered visual review.** Actual rendered appearance (dark/light
screenshots, unstyled-flash, real on-screen contrast) needs a running build + a browser: build,
`pnpm preview`, and drive it with the `agent-browser` CLI, or route to `ux-flow-reviewer`. This
static layer does not open a browser.

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
