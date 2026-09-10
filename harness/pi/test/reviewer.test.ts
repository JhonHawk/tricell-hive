import test from "node:test";
import assert from "node:assert/strict";
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { registerHiveReviewerGuard } from "../src/reviewer.ts";

test("reviewer extension registers its load and guard readiness sentinel", () => {
  const registered: Array<{ readonly name: string }> = [];
  const api = {
    registerTool: (tool: { readonly name: string }) => registered.push(tool),
    on: () => undefined,
  } as unknown as ExtensionAPI;

  registerHiveReviewerGuard(api, { paths: { reviewerGuard: "/tmp/reviewer-guard.sh" } });

  assert.deepEqual(registered.map((tool) => tool.name), ["hive_reviewer_readiness"]);
});
