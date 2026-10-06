// @ts-check
import starlight from "@astrojs/starlight";
import { defineConfig } from "astro/config";

export default defineConfig({
  integrations: [
    starlight({
      title: "Project specifications",
      sidebar: [{ label: "Overview", items: [{ slug: "index" }] }],
    }),
  ],
});
