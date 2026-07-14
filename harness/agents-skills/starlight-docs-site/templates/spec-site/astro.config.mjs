// @ts-check

import starlight from "@astrojs/starlight";
import { defineConfig } from "astro/config";
import mermaid from "astro-mermaid";
import starlightLinksValidator from "starlight-links-validator";

// https://astro.build/config
export default defineConfig({
  integrations: [
    // Mermaid must run BEFORE Starlight: its mdast/rehype transform has to
    // process the diagram code fences before Starlight's pipeline touches the
    // markdown. `autoTheme` follows the site's `html[data-theme]`; the brand
    // color is injected as the node fill so diagrams stay on-brand in both
    // themes. REPLACE: the four brand hex values below to match theme.css.
    mermaid({
      autoTheme: true,
      mermaidConfig: {
        flowchart: { curve: "basis" },
        themeVariables: {
          primaryColor: "#4f6bed", // node fill = brand
          primaryTextColor: "#0a1230", // dark text, legible on the brand fill
          primaryBorderColor: "#3a52c4",
          lineColor: "#3a52c4", // connectors in the brand tone
        },
      },
    }),
    starlight({
      // REPLACE: site title.
      title: "<Project> — Especificaciones",
      // Brand theme. See src/styles/theme.css.
      customCss: ["./src/styles/theme.css"],
      plugins: [
        // Fail the build on broken internal links/anchors (Starlight core does
        // not check links). errorOnRelativeLinks:false allows relative links.
        starlightLinksValidator({ errorOnRelativeLinks: false }),
      ],
      // Single-language site: Mexican Spanish.
      defaultLocale: "root",
      locales: {
        root: { label: "Español", lang: "es-MX" },
      },
      social: [],
      // The sidebar IS the product topology, in this order:
      //   1. "Introducción" — the product map (qué es, actores y roles, flujo completo).
      //   2. One group per business module, its vistas as items.
      //   3. Appendix, last — épicas, decisiones, requisitos (reference material).
      // It grows one vista at a time — never scaffold empty groups. Delivery
      // taxonomy (epic IDs) is never the navigation axis, and workflow status
      // never rides in labels — it renders in-page via SpecRubric.
      // The governance page lives under "Guías", outside the specs navigation.
      sidebar: [
        {
          label: "Guías",
          items: [{ label: "Workflow de construcción", slug: "workflow/product-workflow" }],
        },
      ],
    }),
  ],
});
