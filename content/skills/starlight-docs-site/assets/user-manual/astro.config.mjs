// @ts-check
import starlight from "@astrojs/starlight";
import { defineConfig } from "astro/config";

export default defineConfig({
  integrations: [
    starlight({
      title: "Project manual",
      sidebar: [{ label: "Start", items: [{ slug: "index" }] }],
    }),
  ],
});
