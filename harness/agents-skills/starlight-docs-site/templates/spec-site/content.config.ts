import { defineCollection } from "astro:content";
import { docsLoader } from "@astrojs/starlight/loaders";
import { docsSchema } from "@astrojs/starlight/schema";
import { z } from "astro/zod";

const workflowSchema = z.object({
  module: z.string(),
  slice: z.string().optional(),
  artifact: z.enum([
    "product-map",
    "module-index",
    "brd",
    "vista",
    "technical-spec",
    "acceptance-criteria",
    "uat-report",
  ]),
  status: z.enum(["draft", "gate1-approved", "gate2-approved"]),
  stage: z.enum([
    "discovery",
    "brd-draft",
    "gate1-approved",
    "technical-spec",
    "mock",
    "acceptance-criteria",
    "build-qa",
    "uat",
    "gate2-approved",
  ]),
  lastUpdated: z.date(),
  nextGate: z.string().optional(),
  owner: z.string().optional(),
  evidence: z.array(z.string()).optional(),
  gate1Evidence: z.array(z.string()).optional(),
  gate2Evidence: z.array(z.string()).optional(),
  tracker: z.array(z.string()).optional(),
  // Vista pages only: the epics (deltas) that created or modified this vista.
  // Epic STATUS is read from the repo's epic index — never duplicated here.
  influencedBy: z
    .array(
      z.object({
        epic: z.string(), // e.g. "E04 — roles operativos"
        contribution: z.string(), // one line: what it added/changed
        gateDate: z.date(), // when the business gate passed
      }),
    )
    .optional(),
});

export const collections = {
  docs: defineCollection({
    loader: docsLoader(),
    schema: docsSchema({
      extend: z.object({
        // `description` is REQUIRED (SEO/social + no metadata drift); `workflow`
        // is the spec-status rubric rendered by SpecRubric.astro.
        description: z.string(),
        workflow: workflowSchema.optional(),
      }),
    }),
  }),
};
