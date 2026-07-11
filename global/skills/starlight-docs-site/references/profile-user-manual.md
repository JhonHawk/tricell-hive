# Profile — user-manual

An operator/QA manual. Diátaxis: **how-to (dominant) + reference**. Reference project:
`globex-specs/docs/manual-usuario`. Template: `templates/user-manual/`.

## What makes it this profile

- `starlight-image-zoom` plugin + `unified()` markdown processor (screenshots need zoom).
- Screenshots of the real UI (WebP), role×capability tables, business-rule callouts.
- Voice: descriptive. The established convention is **formal "usted"** — keep it by default
  (see the voice note in `best-practices.md`; it is a configurable convention, not a law).
- Slugs: Spanish, matching the manual's content language (`introduccion/que-es`,
  `portal/mensajeria/chat`) — URLs are user-facing content here. Kebab-case, stable.

## Page skeleton (screen pages)

Every screen page uses these **6 H2 sections, fixed order** (`page-template.md`):

1. `## Propósito` — why the screen exists, the problem it solves.
2. `## Quién puede acceder` — capability×role table (`✓`/`✗`/`—`), plus the "la UI esconde;
   el backend niega" rule.
3. `## Tour de la pantalla` — UI areas in **bold**, then the screenshot. Text must stand alone.
4. `## Flujos paso a paso` — `###` per flow, numbered steps, UI actions in **bold**, observable results.
5. `## Reglas de negocio` — named rules; the épicas are the source of truth, explained here for the operator.
6. `## Estados y errores` — Starlight asides with titles (`:::note[…]`, `:::caution[…]`,
   `:::danger[…]`); do not stack consecutive asides.

Intro/overview pages (home, "qué es", "roles") are free-form — not every page is a screen page.

## Content conventions

- Body starts at `##` (H1 = frontmatter `title`). Front-load each section's first sentence.
- Frontmatter is minimal: `title` + `description` (required). `index.mdx` uses `template: splash` + `hero`.
- Screenshots: `NN-slug.webp` narrative order, under `src/assets/<area>/`, relative path,
  full functional Spanish alt. Add only when the real capture exists — the template ships the
  image line commented so a fresh page builds without it.
- Mermaid diagrams wrapped in `<figure><figcaption>…</figcaption></figure>`.
- Internal links relative with trailing slash: `[Clientes](../clientes/)`.
