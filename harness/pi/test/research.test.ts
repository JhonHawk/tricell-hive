import assert from "node:assert/strict";
import test from "node:test";
import type { ExtensionAPI, ExtensionContext } from "@earendil-works/pi-coding-agent";
import {
  createResearchReadinessTool,
  formatHiveResearchStatus,
  inspectHiveResearchReadiness,
  registerHiveResearchReadiness,
} from "../src/research.ts";

test("documentation readiness accepts required tools and ignores extras", () => {
  assert.deepEqual(
    inspectHiveResearchReadiness(["fetch_content", "get_search_content", "read"], "documentation"),
    {
      profile: "documentation",
      available: ["fetch_content", "get_search_content"],
      missing: [],
      ready: true,
    },
  );
});

test("web readiness inspects each active agent independently", () => {
  const parent = inspectHiveResearchReadiness(
    ["fetch_content", "get_search_content", "web_search", "source_check"],
    "web",
  );
  const child = inspectHiveResearchReadiness(
    ["fetch_content", "get_search_content", "source_check"],
    "web",
  );

  assert.equal(parent.ready, true);
  assert.deepEqual(child.available, ["fetch_content", "get_search_content", "source_check"]);
  assert.deepEqual(child.missing, ["web_search"]);
  assert.equal(child.ready, false);
});

test("registered readiness remains advisory when research tools are inactive", async () => {
  const activeTools = ["read", "hive_research_readiness"];
  const tool = createResearchReadinessTool(() => activeTools);
  const result = await tool.execute(
    "research-readiness",
    { profile: "web" },
    undefined,
    undefined,
    {} as ExtensionContext,
  );

  assert.deepEqual(result.details, {
    profile: "web",
    available: [],
    missing: ["fetch_content", "get_search_content", "web_search", "source_check"],
    ready: false,
  });
});

test("general extension registers the readiness tool without changing active tools", () => {
  const registered: Array<{ readonly name: string }> = [];
  const pi = {
    getActiveTools: () => ["read", "bash"],
    registerTool: (tool: { readonly name: string }) => registered.push(tool),
  } as unknown as ExtensionAPI;

  registerHiveResearchReadiness(pi);

  assert.deepEqual(registered.map((tool) => tool.name), ["hive_research_readiness"]);
  assert.deepEqual(pi.getActiveTools(), ["read", "bash"]);
});

test("status formatting reports both research profiles", () => {
  const status = formatHiveResearchStatus(["fetch_content", "get_search_content", "source_check"]);
  assert.match(status, /research_readiness:/u);
  assert.match(status, /documentation: .*ready=true/u);
  assert.match(status, /web: .*missing=web_search.*ready=false/u);
});
