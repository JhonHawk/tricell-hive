// @ts-check
import starlight from "@astrojs/starlight";
import { defineConfig } from "astro/config";

export default defineConfig({
  site: "https://hive.tricell.tech",
  integrations: [
    starlight({
      title: "Hive",
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
