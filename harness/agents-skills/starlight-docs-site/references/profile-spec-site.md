# Profile — spec-site

A specs/BRD docs-as-code site: the source of truth for business rules. Diátaxis:
**reference (dominant) + explanation**. Reference project: `acme-next-specs`
(distilled from its `_support/docs/specs-site-replication-guide.md`). Template:
`templates/spec-site/`.

## Flow-core mapping (canonical specs repos)

When the site sits over a flow-core specs repo (`flow-core/references/specs-structure.md`),
the two vocabularies reconcile like this — flow-core stays the dev source of truth; the
site presents the **product topology**, never the delivery taxonomy:

| This profile | flow-core specs repo |
|---|---|
| Module | `product/<module>/` — a business capability |
| Slice (rebanada) | ≈ a **vista** — `product/<module>/<vista>.md` |
| BRD per slice | the vista page: business rules IN FORCE, accumulated across epics |
| Gates on a slice | the **business gate** (the spec-writing playbook's review step) on the epics that shaped it |
| — | `epics/` — planning deltas; appendix material on the site, never the nav axis |

Epics are *deltas*; vistas are the *state*. A vista page lists the epics that shaped it in
its `influencedBy` traceability block — the epic's status is read from the repo's epic
index (one source), never duplicated into the vista.

## Sidebar — the product topology

The sidebar IS the product map, in this order:

1. **Introducción** — qué es el producto, actores y roles, flujo completo (the map,
   `product/README.md`).
2. **One group per module**, its vistas as items — groups appear only when the module
   exists (grows-by-slice; never scaffold empty groups).
3. **Appendix, last** — épicas, decisiones, requisitos: reference material for planning
   readers.

Two hard rules: delivery taxonomy (epic IDs `E01…`) is **never** the navigation axis, and
workflow status **never** rides in sidebar labels (`"draft (bloqueada en Q-03)"` is the
anti-pattern) — status renders in-page via `SpecRubric`.

## What makes it this profile

- `SpecRubric.astro` + a `workflow` frontmatter block validated by a Zod schema in
  `content.config.ts` (the schema guarantees no spec publishes without a status).
- Governance by gates; no image-zoom (wireframes/diagrams, not UI screenshots).
- Voice: **impersonal, third person** — declarative and precise for reference, argumentative
  for the rationale (explanation), kept separate. Ambiguity is a measurable defect here.
- Slugs: English (`school-signup/public-signup`, `workflow/product-workflow`). Kebab-case, stable.

## Governance — grows by slice, gated

A **module** groups a business capability; a **slice** (rebanada) is a vertical, small,
independently approvable+verifiable delivery. Publishing ≠ approving — the rubric states the
status with dated evidence. States: `draft` → `gate1-approved` → `gate2-approved`. The
governance page (`product-workflow.mdx`) lives under "Guías", outside the specs navigation.

On a flow-core repo the gates map, they don't multiply: `gate1-approved` ≈ the business
gate passed (the spec-writing playbook's review step — rules decided, vista updated);
`gate2-approved` ≈ delivered/UAT signed. Technical resolution is TECH.md's gate (native
plan mode) and never blocks `gate1`.

## The workflow rubric (frontmatter + component)

Every spec page carries a `workflow` block rendered by `<SpecRubric {...frontmatter.workflow} />`
right under the title. Fields (schema in `templates/spec-site/content.config.ts`):
`module`, `slice?`, `artifact`
(product-map|module-index|brd|vista|technical-spec|acceptance-criteria|uat-report),
`status`, `stage`, `lastUpdated` (date), and optional `nextGate`, `owner`, `evidence`,
`gate1Evidence`, `gate2Evidence`, `tracker`, `influencedBy`. Copy `SpecRubric.astro` into
`src/components/`; the import path from a slice page is `../../../components/SpecRubric.astro`.

**Traceability (`influencedBy`)** — on vista/BRD pages, the epics that created or modified
the vista: `{ epic, contribution, gateDate }` per entry, rendered by `SpecRubric` as a
compact "Influenciada por" list. It records who contributed what and when the business
gate passed; the epic's *current status* lives in the repo's epic index and is not
repeated here.

## BRD anatomy (`brd-template.mdx`)

Reference structure — reference + explanation, never how-to:

1. Opening aside "qué es y qué no es este documento" (business vs. technical).
2. `## Propósito` (with a success metric).
3. `## Actores y precondiciones`.
4. `## Flujo end-to-end` (`<Steps>` and/or mermaid `flowchart TD`).
5. `## Pantallas` (if UI) — wireframes in `src/assets/<slice>/` (WebP).
6. `## Reglas de negocio`.
7. `## Estados y transiciones` (mermaid `stateDiagram-v2`).
8. Cross-cutting sections that apply (e.g. anti-abuse).
9. `## Escenarios de aceptación` — numbered, Dado/Cuando/Entonces: happy path + every negative path.
10. `## Decisiones pendientes del PO` — open questions documented, not hidden.
11. `## Referencia visual`.

Module indexes (`module-index.mdx`, `artifact: module-index`) are short: present the module,
list slices with status, declare convergence rules. They do not replace BRDs.
