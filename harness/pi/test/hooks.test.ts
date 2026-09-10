import { chmodSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import assert from "node:assert/strict";
import type { ExtensionAPI, ExtensionContext } from "@earendil-works/pi-coding-agent";
import {
  guardHiveMcpInput,
  HIVE_CONTEXT7_MCP_TOOLS,
  registerGeneralHiveHooks,
} from "../src/hooks.ts";
import type { HookPaths } from "../src/types.ts";

type StoredHandler = (event: unknown, ctx: ExtensionContext) => unknown | Promise<unknown>;

interface FakeEvents {
  readonly on: (channel: string, handler: (payload: unknown) => void) => () => void;
}

interface FakePi {
  readonly api: ExtensionAPI;
  readonly handlers: Map<string, StoredHandler>;
}

function createFakePi(): FakePi {
  const handlers = new Map<string, StoredHandler>();
  const listeners = new Map<string, Set<(payload: unknown) => void>>();
  const events: FakeEvents = {
    on(channel, handler) {
      const channelListeners = listeners.get(channel) ?? new Set<(payload: unknown) => void>();
      channelListeners.add(handler);
      listeners.set(channel, channelListeners);
      return () => channelListeners.delete(handler);
    },
  };
  const api = {
    events,
    on(event: string, handler: unknown) {
      handlers.set(event, handler as StoredHandler);
    },
    registerTool: () => undefined,
  } as unknown as ExtensionAPI;
  return { api, handlers };
}

function makeContext(directory: string, sessionId: string, systemPrompt = "base"): ExtensionContext {
  return {
    cwd: directory,
    signal: undefined,
    sessionManager: {
      getSessionId: () => sessionId,
      getSessionFile: () => undefined,
    } as unknown as ExtensionContext["sessionManager"],
    ui: {} as ExtensionContext["ui"],
  } as unknown as ExtensionContext;
}

async function invoke(
  fake: FakePi,
  name: string,
  event: unknown,
  context: ExtensionContext,
): Promise<unknown> {
  const handler = fake.handlers.get(name);
  assert.ok(handler, `missing ${name} handler`);
  return handler(event, context);
}

function pathsFor(directory: string): HookPaths {
  return {
    bashPolicy: join(directory, "missing-bash-policy.sh"),
    reviewerGuard: join(directory, "missing-reviewer-guard.sh"),
    postToolHub: join(directory, "missing-post-tool-hub.sh"),
    flowContext: join(directory, "missing-flow-context.sh"),
    flowPlanCapture: join(directory, "missing-plan-capture.sh"),
  };
}

test("advisory hook failures warn once per session and script while context remains reusable", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-advisory-dedupe-"));
  const fake = createFakePi();
  const contextA = makeContext(directory, "session-a");
  const contextB = makeContext(directory, "session-b");
  const paths = pathsFor(directory);
  const previousParent = process.env.PI_SUBAGENT_PARENT_SESSION;
  try {
    delete process.env.PI_SUBAGENT_PARENT_SESSION;
    registerGeneralHiveHooks(fake.api, { paths });
    await invoke(fake, "session_start", { type: "session_start", reason: "startup" }, contextA);

    const beforeFirst = await invoke(fake, "before_agent_start", { systemPrompt: "base" }, contextA);
    assert.match(JSON.stringify(beforeFirst), /Required hook is missing/u);
    const beforeSecond = await invoke(fake, "before_agent_start", { systemPrompt: "base" }, contextA);
    assert.doesNotMatch(JSON.stringify(beforeSecond), /Required hook is missing/u);
    assert.match(JSON.stringify(beforeSecond), /context7_query-docs/u);

    const toolResult = {
      type: "tool_result",
      toolCallId: "tool-1",
      toolName: "read",
      content: [{ type: "text", text: "result" }],
      isError: false,
      details: undefined,
    };
    const postFirst = await invoke(fake, "tool_result", toolResult, contextA);
    assert.match(JSON.stringify(postFirst), /Required hook is missing/u);
    const postSecond = await invoke(fake, "tool_result", toolResult, contextA);
    assert.equal(postSecond, undefined);

    await invoke(fake, "session_start", { type: "session_start", reason: "new" }, contextB);
    const newSessionWarning = await invoke(fake, "before_agent_start", { systemPrompt: "base" }, contextB);
    assert.match(JSON.stringify(newSessionWarning), /Required hook is missing/u);
  } finally {
    if (previousParent === undefined) delete process.env.PI_SUBAGENT_PARENT_SESSION;
    else process.env.PI_SUBAGENT_PARENT_SESSION = previousParent;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("ordinary advisory additional context is preserved on every invocation", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-advisory-context-"));
  const flowContext = join(directory, "flow-context.sh");
  writeFileSync(flowContext, "#!/usr/bin/env node\nprocess.stdout.write(JSON.stringify({ additionalContext: 'ordinary context' }));\n");
  chmodSync(flowContext, 0o755);
  const fake = createFakePi();
  const context = makeContext(directory, "session-a");
  const paths = { ...pathsFor(directory), flowContext };
  try {
    registerGeneralHiveHooks(fake.api, { paths });
    const first = await invoke(fake, "before_agent_start", { systemPrompt: "base" }, context);
    const second = await invoke(fake, "before_agent_start", { systemPrompt: "base" }, context);
    assert.match(JSON.stringify(first), /ordinary context/u);
    assert.match(JSON.stringify(second), /ordinary context/u);
    assert.match(JSON.stringify(first), new RegExp(HIVE_CONTEXT7_MCP_TOOLS.resolveLibraryId, "u"));
    assert.match(JSON.stringify(first), new RegExp(HIVE_CONTEXT7_MCP_TOOLS.queryDocs, "u"));
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test("Hive MCP guard accepts only the two closed Context7 calls", () => {
  const resolveInput = {
    tool: HIVE_CONTEXT7_MCP_TOOLS.resolveLibraryId,
    args: { query: "react", libraryName: "react" },
  };
  const queryInput = {
    tool: HIVE_CONTEXT7_MCP_TOOLS.queryDocs,
    args: { libraryId: "/facebook/react", query: "hooks" },
  };
  assert.deepEqual(guardHiveMcpInput(resolveInput), { allowed: true, input: resolveInput });
  assert.deepEqual(guardHiveMcpInput(queryInput), { allowed: true, input: queryInput });

  const rejected: unknown[] = [
    { search: "context7" },
    { list: true },
    { server: "context7" },
    { action: "auth-start" },
    { code: "await tools.search({ query: 'x' })" },
    { scripts: ["context7"] },
    { tool: "unknown_tool", args: { query: "x", libraryName: "x" } },
    { tool: HIVE_CONTEXT7_MCP_TOOLS.resolveLibraryId, args: "{}" },
    { tool: HIVE_CONTEXT7_MCP_TOOLS.resolveLibraryId, args: { query: "x" } },
    { tool: HIVE_CONTEXT7_MCP_TOOLS.resolveLibraryId, args: { query: "x", libraryName: "x", server: "x" } },
    { tool: HIVE_CONTEXT7_MCP_TOOLS.queryDocs, args: { libraryId: "x", query: "" } },
    { tool: HIVE_CONTEXT7_MCP_TOOLS.queryDocs, args: { libraryId: "x", query: "x" }, extra: true },
  ];
  for (const input of rejected) {
    const result = guardHiveMcpInput(input);
    assert.equal(result.allowed, false, JSON.stringify(input));
  }
});

test("general Hive hooks reject invalid MCP gateway input before adapter execution", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-mcp-hook-"));
  const fake = createFakePi();
  const context = makeContext(directory, "session-a");
  const previousParent = process.env.PI_SUBAGENT_PARENT_SESSION;
  try {
    delete process.env.PI_SUBAGENT_PARENT_SESSION;
    registerGeneralHiveHooks(fake.api, { paths: pathsFor(directory) });
    const blocked = await invoke(fake, "tool_call", {
      type: "tool_call",
      toolCallId: "mcp-1",
      toolName: "mcp",
      input: { search: "context7" },
    }, context);
    assert.deepEqual(blocked, {
      block: true,
      reason: "Hive MCP bridge only allows Context7 resolve-library-id or query-docs with their required string arguments.",
      terminate: true,
    });
    const allowed = await invoke(fake, "tool_call", {
      type: "tool_call",
      toolCallId: "mcp-2",
      toolName: "mcp",
      input: {
        tool: HIVE_CONTEXT7_MCP_TOOLS.queryDocs,
        args: { libraryId: "/facebook/react", query: "hooks" },
      },
    }, context);
    assert.equal(allowed, undefined);
  } finally {
    if (previousParent === undefined) delete process.env.PI_SUBAGENT_PARENT_SESSION;
    else process.env.PI_SUBAGENT_PARENT_SESSION = previousParent;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("general Hive hooks reject the generic MCP script tool in child processes", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-mcp-child-"));
  const fake = createFakePi();
  const context = makeContext(directory, "child-session");
  const previousChild = process.env.PI_SUBAGENT_CHILD;
  try {
    process.env.PI_SUBAGENT_CHILD = "1";
    registerGeneralHiveHooks(fake.api, { paths: pathsFor(directory) });
    const blocked = await invoke(fake, "tool_call", {
      type: "tool_call",
      toolCallId: "mcp-script-1",
      toolName: "mcpScript",
      input: { code: "await tools.search({ query: 'context7' })" },
    }, context);
    assert.deepEqual(blocked, {
      block: true,
      reason: "Hive child agents cannot use the generic MCP script tool.",
      terminate: true,
    });
  } finally {
    if (previousChild === undefined) delete process.env.PI_SUBAGENT_CHILD;
    else process.env.PI_SUBAGENT_CHILD = previousChild;
    rmSync(directory, { recursive: true, force: true });
  }
});
