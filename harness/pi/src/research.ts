import type { ExtensionAPI, ToolDefinition } from "@earendil-works/pi-coding-agent";
import { Type } from "typebox";

export const HIVE_RESEARCH_PROFILES = {
  documentation: ["fetch_content", "get_search_content"],
  web: ["fetch_content", "get_search_content", "web_search", "source_check"],
} as const;

export type HiveResearchProfile = keyof typeof HIVE_RESEARCH_PROFILES;

export interface HiveResearchReadiness {
  readonly profile: HiveResearchProfile;
  readonly available: readonly string[];
  readonly missing: readonly string[];
  readonly ready: boolean;
}

const researchProfileSchema = Type.Union([
  Type.Literal("documentation"),
  Type.Literal("web"),
]);

const researchReadinessSchema = Type.Object({
  profile: Type.Optional(researchProfileSchema),
});

export function inspectHiveResearchReadiness(
  activeTools: readonly string[],
  profile: HiveResearchProfile,
): HiveResearchReadiness {
  const active = new Set(activeTools);
  const required = HIVE_RESEARCH_PROFILES[profile];
  const available = required.filter((tool) => active.has(tool));
  const missing = required.filter((tool) => !active.has(tool));
  return {
    profile,
    available,
    missing,
    ready: missing.length === 0,
  };
}

export function formatHiveResearchReadiness(readiness: HiveResearchReadiness): string {
  const available = readiness.available.join(", ") || "none";
  const missing = readiness.missing.join(", ") || "none";
  return `${readiness.profile}: available=${available}; missing=${missing}; ready=${String(readiness.ready)}`;
}

export function createResearchReadinessTool(
  activeToolsReader: () => readonly string[],
): ToolDefinition<typeof researchReadinessSchema, HiveResearchReadiness> {
  return {
    name: "hive_research_readiness",
    label: "Hive research readiness",
    description: "Inspect the active PI agent tools required for documentation or web research.",
    promptSnippet: "Check active research tools before external research; local tasks remain unaffected",
    parameters: researchReadinessSchema,
    executionMode: "sequential",
    async execute(_toolCallId, params) {
      const profile = params.profile ?? "web";
      const readiness = inspectHiveResearchReadiness(activeToolsReader(), profile);
      return {
        content: [{ type: "text" as const, text: formatHiveResearchReadiness(readiness) }],
        details: readiness,
        // Missing provider tools are an informational readiness result. They do
        // not turn a local task into a failed tool call.
        isError: false,
      };
    },
  };
}

export function registerHiveResearchReadiness(pi: ExtensionAPI): void {
  pi.registerTool(createResearchReadinessTool(() => pi.getActiveTools()));
}

export function researchReadinessForActiveTools(
  activeTools: readonly string[],
): Readonly<Record<HiveResearchProfile, HiveResearchReadiness>> {
  return {
    documentation: inspectHiveResearchReadiness(activeTools, "documentation"),
    web: inspectHiveResearchReadiness(activeTools, "web"),
  };
}

export function formatHiveResearchStatus(activeTools: readonly string[]): string {
  const statuses = researchReadinessForActiveTools(activeTools);
  return [
    "research_readiness:",
    formatHiveResearchReadiness(statuses.documentation),
    formatHiveResearchReadiness(statuses.web),
  ].join("\n");
}
