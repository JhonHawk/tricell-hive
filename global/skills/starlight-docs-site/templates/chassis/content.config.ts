import { defineCollection } from "astro:content";
import { docsLoader } from "@astrojs/starlight/loaders";
import { docsSchema } from "@astrojs/starlight/schema";
import { z } from "astro/zod";

// Base collection (user-manual profile). `description` is made REQUIRED so every
// page ships SEO/social metadata and the sidebar/search stay meaningful — a
// missing description is the most-missed Starlight authoring gap. The spec-site
// profile replaces this file with one that also carries the `workflow` rubric.
export const collections = {
  docs: defineCollection({
    loader: docsLoader(),
    schema: docsSchema({
      extend: z.object({
        description: z.string(),
      }),
    }),
  }),
};
