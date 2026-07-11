// @ts-check

import { unified } from "@astrojs/markdown-remark";
import starlight from "@astrojs/starlight";
import { defineConfig } from "astro/config";
import mermaid from "astro-mermaid";
import starlightImageZoom from "starlight-image-zoom";
import starlightLinksValidator from "starlight-links-validator";

// https://astro.build/config
export default defineConfig({
  // starlight-image-zoom does not yet support the Sätteri processor (Astro 7's
  // default) — see github.com/HiDeoo/starlight-image-zoom/issues/63. unified()/
  // remark is the official fallback the plugin's own error suggests. Only needed
  // in the user-manual profile because it is the only one that ships image-zoom.
  markdown: {
    processor: unified(),
  },
  integrations: [
    // Mermaid must run BEFORE Starlight: its mdast/rehype transform has to
    // process the diagram code fences before Starlight's pipeline touches the
    // markdown. `autoTheme` follows the site's `html[data-theme]`; the brand
    // color is injected as the node fill so diagrams stay on-brand in both
    // themes. REPLACE: the four brand hex values below to match theme.css.
    mermaid({
      autoTheme: true,
      mermaidConfig: {
        // fontSize/nodeSpacing/rankSpacing raise visual density over mermaid 11
        // defaults — more air between nodes without overflowing the ~72ch column.
        fontSize: 16,
        flowchart: { curve: "basis", nodeSpacing: 60, rankSpacing: 70 },
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
      title: "<Project> — Manual de usuario",
      // Brand theme. See src/styles/theme.css.
      customCss: ["./src/styles/theme.css"],
      plugins: [
        // Click-to-zoom on the manual's screenshots.
        starlightImageZoom(),
        // Fail the build on broken internal links/anchors (Starlight core does
        // not check links). errorOnRelativeLinks:false allows relative links,
        // which this manual uses for cross-references.
        starlightLinksValidator({ errorOnRelativeLinks: false }),
      ],
      // Single-language site: Mexican Spanish.
      defaultLocale: "root",
      locales: {
        root: { label: "Español", lang: "es-MX" },
      },
      social: [],
      // The sidebar grows one section at a time — never scaffold empty groups.
      // Add each page's entry here as it is authored.
      sidebar: [
        {
          label: "Introducción",
          items: [{ slug: "introduccion/que-es" }],
        },
      ],
    }),
  ],
});
