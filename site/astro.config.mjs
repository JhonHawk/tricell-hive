// @ts-check
import starlight from "@astrojs/starlight";
import { defineConfig } from "astro/config";

export default defineConfig({
  site: "https://hive.tricell.tech",
  integrations: [
    starlight({
      title: "Hive",
      head: [
        {
          tag: "meta",
          attrs: {
            property: "og:image",
            content: "https://hive.tricell.tech/social-preview.png",
          },
        },
        {
          tag: "meta",
          attrs: { property: "og:image:width", content: "1734" },
        },
        {
          tag: "meta",
          attrs: { property: "og:image:height", content: "907" },
        },
        {
          tag: "meta",
          attrs: {
            property: "og:image:alt",
            content:
              "Hive: portable guidance for coding agents, with seven connected hexagonal modules lit in red.",
          },
        },
        {
          tag: "meta",
          attrs: {
            name: "twitter:image",
            content: "https://hive.tricell.tech/social-preview.png",
          },
        },
        {
          tag: "meta",
          attrs: {
            name: "twitter:image:alt",
            content:
              "Hive: portable guidance for coding agents, with seven connected hexagonal modules lit in red.",
          },
        },
      ],
      logo: {
        dark: "./src/assets/hive-mark.svg",
        light: "./src/assets/hive-mark-light.svg",
        alt: "",
        replacesTitle: false,
      },
      social: [
        {
          icon: "github",
          label: "GitHub",
          href: "https://github.com/JhonHawk/tricell-hive",
        },
      ],
      customCss: ["./src/styles/theme.css"],
      sidebar: [
        { label: "Start", items: [{ slug: "index" }] },
        {
          label: "Flows",
          items: [
            { slug: "flows/research" },
            { slug: "flows/plan" },
            { slug: "flows/build" },
            { slug: "flows/close" },
          ],
        },
      ],
    }),
  ],
});
