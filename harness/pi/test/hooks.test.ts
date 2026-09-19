import { chmodSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import assert from "node:assert/strict";
import type { ExtensionAPI, ExtensionContext } from "@earendil-works/pi-coding-agent";
import {
  createHiveMcpGuard,
  getHiveHookStatuses,
  guardHiveMcpInput,
  loadHiveMcpAllowlist,
  HIVE_CONTEXT7_MCP_TOOLS,
  PI_MCP_GUIDANCE,
  REQUIRED_HOOK_KEYS,
  registerGeneralHiveHooks,
  resolveHookRoot,
} from "../src/hooks.ts";
import type { HookPaths } from "../src/types.ts";

type StoredHandler = (event: unknown, ctx: ExtensionContext) => unknown | Promise<unknown>;

interface FakeEvents {
  readonly on: (channel: string, handler: (payload: unknown) => void) => () => void;
}

interface FakePi {
  readonly api: ExtensionAPI;
  readonly handlers: Map<string, StoredHandler>;
  readonly commands: Map<string, StoredHandler>;
  readonly messages: Array<{ readonly message: unknown; readonly options: unknown }>;
  readonly notifications: string[];
  readonly tools: unknown[];
}

function createFakePi(
  persistedEntries: unknown[] = [],
  activeTools: readonly string[] = ["read", "bash", "mcp"],
): FakePi {
  const handlers = new Map<string, StoredHandler>();
  const commands = new Map<string, StoredHandler>();
  const messages: Array<{ readonly message: unknown; readonly options: unknown }> = [];
  const notifications: string[] = [];
  const tools: unknown[] = [];
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
    registerCommand: (name: string, options: { readonly handler: StoredHandler }) => {
      commands.set(name, options.handler);
    },
    registerTool: (tool: unknown) => tools.push(tool),
    getActiveTools: () => [...activeTools],
    sendMessage: (message: unknown, options: unknown) => {
      messages.push({ message, options });
      const hasDeliveryMode = typeof options === "object" && options !== null && "deliverAs" in options;
      if (!hasDeliveryMode && typeof message === "object" && message !== null && "customType" in message) {
        persistedEntries.push({ type: "custom_message", customType: message.customType });
      }
    },
  } as unknown as ExtensionAPI;
  return { api, handlers, commands, messages, notifications, tools };
}

function makeContext(
  directory: string,
  sessionId: string,
  systemPrompt = "base",
  entries: unknown[] = [],
  notifications: string[] = [],
): ExtensionContext {
  return {
    cwd: directory,
    signal: undefined,
    sessionManager: {
      getSessionId: () => sessionId,
      getSessionFile: () => undefined,
      getBranch: () => entries,
      buildContextEntries: () => entries,
    } as unknown as ExtensionContext["sessionManager"],
    ui: { notify: (message: string) => notifications.push(message) } as unknown as ExtensionContext["ui"],
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
    flowSessionContext: join(directory, "missing-flow-session-context.sh"),
    ruleContext: join(directory, "missing-rule-context.sh"),
    ruleDelivery: join(directory, "missing-rule-delivery.py"),
    sessionHygieneReport: join(directory, "missing-session-hygiene-report.sh"),
  };
}

function makeAdvisoryHook(directory: string, name: string, context: string, extraSource = ""): string {
  const hook = join(directory, name);
  writeFileSync(hook, `#!/usr/bin/env node
const fs = require("node:fs");
let input = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => {
  ${extraSource}
  process.stdout.write(JSON.stringify({ additionalContext: ${JSON.stringify(context)} }));
});
`);
  chmodSync(hook, 0o755);
  return hook;
}

test("hook root resolution prefers explicit, deployed, cwd and source roots in that order", () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-hook-roots-"));
  const configuredDir = join(directory, "configured");
  const deployedDir = join(directory, "deployed");
  const cwdDir = join(directory, "cwd");
  const sourceDir = join(directory, "source");
  for (const root of [configuredDir, deployedDir, cwdDir, sourceDir]) mkdirSync(root);
  const candidates = { configuredDir, deployedDir, cwdDir, sourceDir };
  try {
    assert.equal(resolveHookRoot(candidates), configuredDir);
    assert.equal(resolveHookRoot({ ...candidates, configuredDir: undefined }), deployedDir);
    rmSync(deployedDir, { recursive: true, force: true });
    assert.equal(resolveHookRoot({ ...candidates, configuredDir: undefined }), cwdDir);
    rmSync(cwdDir, { recursive: true, force: true });
    assert.equal(resolveHookRoot({ ...candidates, configuredDir: undefined }), sourceDir);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

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

test("Hive MCP guard keeps the Context7 required-argument contract", () => {
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

test("Hive MCP guard admits allowlisted Linear and HeroUI tools and passes their args through", () => {
  const listIssues = { tool: "linear_list_issues", args: { team: "FAC", limit: 20 } };
  const listIssuesResult = guardHiveMcpInput(listIssues);
  assert.equal(listIssuesResult.allowed, true);
  assert.ok(listIssuesResult.allowed);
  assert.deepEqual(listIssuesResult.input.args, { team: "FAC", limit: 20 });

  const allowed: unknown[] = [
    { tool: "linear_get_issue", args: { id: "FAC-12" } },
    { tool: "linear_save_issue", args: { title: "x", team: "FAC" } },
    { tool: "linear_save_comment", args: { issueId: "FAC-12", body: "done" } },
    { tool: "linear_list_comments", args: {} },
    { tool: "heroui_get_component_docs", args: { component: "Button" } },
    { tool: "heroui_list_components", args: {} },
    {
      tool: "linear_save_issue",
      args: { nested: { level2: { level3: ["a", 1, true, null] } } },
    },
  ];
  for (const input of allowed) {
    assert.equal(guardHiveMcpInput(input).allowed, true, JSON.stringify(input));
  }
});

test("Hive MCP guard suggests the prefixed name when a bare tool name is unambiguous", () => {
  const bare = guardHiveMcpInput({ tool: "list_issues", args: { team: "FAC" } });
  assert.equal(bare.allowed, false);
  assert.match(bare.allowed ? "" : bare.reason, /did you mean linear_list_issues/iu);
  const docs = guardHiveMcpInput({ tool: "get_docs", args: {} });
  assert.match(docs.allowed ? "" : docs.reason, /did you mean heroui_get_docs/iu);
  const none = guardHiveMcpInput({ tool: "run_sql", args: {} });
  assert.doesNotMatch(none.allowed ? "" : none.reason, /did you mean/iu);
  assert.match(PI_MCP_GUIDANCE, /always <server>_<tool>/u);
  assert.match(PI_MCP_GUIDANCE, /never list_issues/u);
});

test("Hive MCP guard rejects non-allowlisted servers, write tools and unsafe argument payloads", () => {
  const rejected: ReadonlyArray<{ readonly input: unknown; readonly reason: RegExp }> = [
    // invalid {tool, args} shape
    { input: { search: "context7" }, reason: /call mcp with exactly \{tool, args\}/u },
    { input: { tool: "linear_list_issues", args: { team: "FAC" }, extra: true }, reason: /exactly \{tool, args\}/u },
    // tool outside the allowlist — the reason names the allowlisted servers
    { input: { tool: "linear_delete_attachment", args: { id: "x" } }, reason: /context7_\*, linear_\*, heroui_\*/u },
    { input: { tool: "linear_merge_diff", args: { id: "x" } }, reason: /unknown tool/u },
    { input: { tool: "linear_create_issue", args: { title: "x" } }, reason: /unknown tool/u },
    { input: { tool: "heroui_nonexistent", args: {} }, reason: /unknown tool/u },
    { input: { tool: "neon_run_sql", args: { sql: "select 1" } }, reason: /unknown tool/u },
    { input: { tool: "toString", args: {} }, reason: /unknown tool/u },
    { input: { tool: "constructor", args: {} }, reason: /unknown tool/u },
    { input: { tool: "valueOf", args: {} }, reason: /unknown tool/u },
    { input: { tool: "hasOwnProperty", args: {} }, reason: /unknown tool/u },
    { input: { tool: "__proto__", args: {} }, reason: /unknown tool/u },
    { input: { tool: "isPrototypeOf", args: {} }, reason: /unknown tool/u },
    // args is not a plain object
    { input: { tool: "linear_list_issues", args: "{}" }, reason: /args must be a plain object/u },
    { input: { tool: "linear_list_issues", args: [1, 2] }, reason: /args must be a plain object/u },
    // Context7 required-argument contract
    {
      input: { tool: HIVE_CONTEXT7_MCP_TOOLS.queryDocs, args: { libraryId: "x", query: "" } },
      reason: /context7_query-docs requires exactly \{libraryId, query\} as non-empty strings/u,
    },
    {
      input: { tool: HIVE_CONTEXT7_MCP_TOOLS.resolveLibraryId, args: { query: "x" } },
      reason: /context7_resolve-library-id requires exactly \{query, libraryName\}/u,
    },
    // args carry data the bridge cannot serialize
    { input: { tool: "linear_list_issues", args: { onDone: () => undefined } }, reason: /only JSON data/u },
    { input: { tool: "linear_list_issues", args: { when: new Date(0) } }, reason: /only JSON data/u },
    // nesting and size budgets name their limit
    {
      input: { tool: "linear_list_issues", args: { deep: { a: { b: { c: { d: 1 } } } } } },
      reason: /maximum depth of 4/u,
    },
    { input: { tool: "linear_list_issues", args: { big: "x".repeat(16_385) } }, reason: /16384 bytes/u },
  ];
  for (const { input, reason } of rejected) {
    const result = guardHiveMcpInput(input, { childSession: false });
    assert.equal(result.allowed, false, JSON.stringify(input));
    assert.ok(!result.allowed);
    assert.match(result.reason, reason, JSON.stringify(input));
  }
});

test("Hive MCP guard rejects prototype-polluting argument keys at any depth", () => {
  const payloads = [
    '{"tool":"linear_list_issues","args":{"__proto__":{"x":1}}}',
    '{"tool":"linear_list_issues","args":{"constructor":{"x":1}}}',
    '{"tool":"linear_list_issues","args":{"prototype":{"x":1}}}',
    '{"tool":"linear_list_issues","args":{"filter":{"__proto__":{"x":1}}}}',
    '{"tool":"linear_list_issues","args":{"filter":[{"constructor":1}]}}',
    '{"tool":"linear_save_issue","args":{"a":{"b":{"prototype":1}}}}',
  ];
  for (const payload of payloads) {
    const result = guardHiveMcpInput(JSON.parse(payload), { childSession: false });
    assert.equal(result.allowed, false, payload);
    assert.ok(!result.allowed);
    assert.match(result.reason, /must not contain the keys __proto__, constructor or prototype/u, payload);
  }
});

test("an unreadable allowlist makes the Hive MCP guard refuse every call", () => {
  const missing = join(tmpdir(), "hive-pi-absent-allowlist.json");
  const allowlist = loadHiveMcpAllowlist(missing);
  assert.ok(allowlist.error !== undefined);
  assert.deepEqual(allowlist.servers, {});

  const guard = createHiveMcpGuard(allowlist);
  for (const input of [
    { tool: HIVE_CONTEXT7_MCP_TOOLS.queryDocs, args: { libraryId: "/facebook/react", query: "hooks" } },
    { tool: "linear_list_issues", args: {} },
  ]) {
    const result = guard(input, { childSession: false });
    assert.equal(result.allowed, false, JSON.stringify(input));
    assert.ok(!result.allowed);
    assert.match(result.reason, /allowlist unavailable \(.+\); all MCP calls are refused\./u);
  }
});

test("a failed allowlist load blocks MCP tool calls and warns once per session", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-allowlist-failure-"));
  const fake = createFakePi();
  const context = makeContext(directory, "session-a", "base", [], fake.notifications);
  const previousChild = process.env.PI_SUBAGENT_CHILD;
  try {
    delete process.env.PI_SUBAGENT_CHILD;
    registerGeneralHiveHooks(fake.api, {
      paths: pathsFor(directory),
      mcpAllowlist: loadHiveMcpAllowlist(join(directory, "absent.json")),
    });
    const blocked = await invoke(fake, "tool_call", {
      type: "tool_call",
      toolCallId: "mcp-1",
      toolName: "mcp",
      input: {
        tool: HIVE_CONTEXT7_MCP_TOOLS.queryDocs,
        args: { libraryId: "/facebook/react", query: "hooks" },
      },
    }, context);
    assert.match(JSON.stringify(blocked), /allowlist unavailable/u);

    await invoke(fake, "before_agent_start", { systemPrompt: "base" }, context);
    await invoke(fake, "before_agent_start", { systemPrompt: "base" }, context);
    const warnings = fake.notifications.filter((message) => /Hive MCP warning/u.test(message));
    assert.equal(warnings.length, 1, JSON.stringify(fake.notifications));
    assert.match(warnings[0] ?? "", /allowlist/u);
  } finally {
    if (previousChild === undefined) delete process.env.PI_SUBAGENT_CHILD;
    else process.env.PI_SUBAGENT_CHILD = previousChild;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("Hive MCP guard confines Linear write tools to the parent session", () => {
  const saveIssue = { tool: "linear_save_issue", args: { title: "Bridge", team: "FAC" } };
  const saveComment = { tool: "linear_save_comment", args: { issueId: "FAC-12", body: "done" } };
  const listIssues = { tool: "linear_list_issues", args: { team: "FAC" } };

  assert.equal(guardHiveMcpInput(saveIssue, { childSession: false }).allowed, true);
  assert.equal(guardHiveMcpInput(saveComment, { childSession: false }).allowed, true);
  assert.equal(guardHiveMcpInput(listIssues, { childSession: false }).allowed, true);
  assert.equal(guardHiveMcpInput(listIssues, { childSession: true }).allowed, true);

  for (const input of [saveIssue, saveComment]) {
    assert.deepEqual(guardHiveMcpInput(input, { childSession: true }), {
      allowed: false,
      reason: "Hive MCP bridge: write tools (save_issue, save_comment) are only callable from the parent session.",
    });
  }
});

test("the Hive MCP guard reads the PI child marker when no session option is given", () => {
  const saveIssue = { tool: "linear_save_issue", args: { title: "Bridge", team: "FAC" } };
  const previousChild = process.env.PI_SUBAGENT_CHILD;
  try {
    process.env.PI_SUBAGENT_CHILD = "1";
    assert.equal(guardHiveMcpInput(saveIssue).allowed, false);
    assert.equal(guardHiveMcpInput({ tool: "linear_list_issues", args: {} }).allowed, true);
    delete process.env.PI_SUBAGENT_CHILD;
    assert.equal(guardHiveMcpInput(saveIssue).allowed, true);
  } finally {
    if (previousChild === undefined) delete process.env.PI_SUBAGENT_CHILD;
    else process.env.PI_SUBAGENT_CHILD = previousChild;
  }
});

test("the PI MCP guidance names every allowlisted server within its budget", async () => {
  assert.match(PI_MCP_GUIDANCE, /^PI MCP bridge: use mcp only with \{tool, args\}\./u);
  assert.match(PI_MCP_GUIDANCE, /linear_list_issues/u);
  assert.match(PI_MCP_GUIDANCE, /heroui_get_component_docs/u);
  assert.match(PI_MCP_GUIDANCE, /save_issue, save_comment/u);
  assert.match(PI_MCP_GUIDANCE, /parent session/u);
  // Budget: this text is injected once per session and per child, so it is capped.
  assert.ok(PI_MCP_GUIDANCE.length <= 1200, `guidance is ${PI_MCP_GUIDANCE.length} characters`);

  const directory = mkdtempSync(join(tmpdir(), "hive-pi-mcp-guidance-"));
  const fake = createFakePi();
  const context = makeContext(directory, "session-a");
  try {
    registerGeneralHiveHooks(fake.api, { paths: pathsFor(directory) });
    const started = JSON.stringify(await invoke(fake, "before_agent_start", { systemPrompt: "base" }, context));
    assert.match(started, /linear_list_issues/u);
    assert.match(started, /heroui_get_component_docs/u);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test("the MCP guidance reaches only roles whose active tools include mcp", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-mcp-guidance-scope-"));
  const withoutMcp = createFakePi([], ["read", "bash", "grep"]);
  const withMcp = createFakePi([], ["read", "bash", "mcp"]);
  const context = makeContext(directory, "session-a");
  try {
    registerGeneralHiveHooks(withoutMcp.api, { paths: pathsFor(directory) });
    const quiet = await invoke(withoutMcp, "before_agent_start", { systemPrompt: "base" }, context);
    assert.doesNotMatch(JSON.stringify(quiet), /PI MCP bridge/u);

    registerGeneralHiveHooks(withMcp.api, { paths: pathsFor(directory) });
    const guided = await invoke(withMcp, "before_agent_start", { systemPrompt: "base" }, context);
    assert.match(JSON.stringify(guided), /PI MCP bridge/u);
    // Skill-path guidance is model-facing for every role: a session read a
    // package skill at ~/.agents/skills/<name> after reading a Hive skill there.
    for (const result of [quiet, guided]) {
      const prompt = JSON.stringify(result);
      assert.match(prompt, /<available_skills>/u);
      assert.match(prompt, /never guess/u);
      assert.match(prompt, /~\/\.agents\/skills\//u);
    }
  } finally {
    rmSync(directory, { recursive: true, force: true });
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
      reason: "Hive MCP bridge: call mcp with exactly {tool, args}.",
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

test("parent session advisories preserve freshness, order, native pending context and compact delivery", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-parent-advisories-"));
  const flowLog = join(directory, "flow.log");
  const flowSessionContext = makeAdvisoryHook(directory, "flow-session-context.sh", "flow", `fs.appendFileSync(${JSON.stringify(flowLog)}, JSON.parse(input).source + "\\n");`);
  const sessionHygieneReport = join(directory, "session-hygiene-report.sh");
  writeFileSync(sessionHygieneReport, `#!/usr/bin/env node
let input = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => process.stdout.write(JSON.stringify({ additionalContext: "hygiene claude=" + (process.env.CLAUDECODE ?? "unset") + " repo=" + (process.env.HIVE_REPO ?? "unset") })));
`);
  chmodSync(sessionHygieneReport, 0o755);
  const paths = { ...pathsFor(directory), flowSessionContext, sessionHygieneReport };
  const entries: unknown[] = [];
  const fake = createFakePi(entries);
  const context = makeContext(directory, "session-a", "base", entries, fake.notifications);
  const previousChild = process.env.PI_SUBAGENT_CHILD;
  const previousParent = process.env.PI_SUBAGENT_PARENT_SESSION;
  const previousClaude = process.env.CLAUDECODE;
  const previousRepo = process.env.HIVE_REPO;
  try {
    delete process.env.PI_SUBAGENT_CHILD;
    delete process.env.PI_SUBAGENT_PARENT_SESSION;
    process.env.CLAUDECODE = "present";
    process.env.HIVE_REPO = "/known/hive";
    registerGeneralHiveHooks(fake.api, { paths });

    await invoke(fake, "session_start", { type: "session_start", reason: "startup" }, context);
    assert.equal(fake.messages.length, 1);
    const initial = JSON.stringify(fake.messages[0]);
    assert.ok(initial.indexOf("flow") < initial.indexOf("hygiene"));
    assert.match(initial, /claude=unset/u);
    assert.match(initial, /repo=\/known\/hive/u);
    assert.match(initial, /triggerTurn.*false/u);
    assert.doesNotMatch(initial, /deliverAs/u);
    assert.match(readFileSync(flowLog, "utf8"), /^startup\n$/u);

    await invoke(fake, "session_start", { type: "session_start", reason: "reload" }, context);
    assert.equal(fake.messages.length, 1, "reload before the first model turn must not enqueue a second native nextTurn message");

    const reloaded = createFakePi(entries);
    const resumedContext = makeContext(directory, "session-a", "base", entries, reloaded.notifications);
    registerGeneralHiveHooks(reloaded.api, { paths });
    await invoke(reloaded, "session_start", { type: "session_start", reason: "startup" }, resumedContext);
    assert.equal(reloaded.messages.length, 0, "a native hidden custom_message in resumed session state must suppress replay");

    entries.length = 0;
    await invoke(fake, "session_start", { type: "session_start", reason: "new" }, context);
    assert.equal(fake.messages.length, 2);
    assert.match(readFileSync(flowLog, "utf8"), /startup\nclear\n/u);

    await invoke(fake, "session_compact", { type: "session_compact", willRetry: true }, context);
    assert.equal(fake.messages.length, 3);
    assert.match(JSON.stringify(fake.messages[2]), /steer/u);
    assert.match(readFileSync(flowLog, "utf8"), /compact\n/u);
    await invoke(fake, "session_compact_failed", { type: "session_compact_failed" }, context);
    assert.equal(fake.messages.length, 3);
  } finally {
    if (previousChild === undefined) delete process.env.PI_SUBAGENT_CHILD;
    else process.env.PI_SUBAGENT_CHILD = previousChild;
    if (previousParent === undefined) delete process.env.PI_SUBAGENT_PARENT_SESSION;
    else process.env.PI_SUBAGENT_PARENT_SESSION = previousParent;
    if (previousClaude === undefined) delete process.env.CLAUDECODE;
    else process.env.CLAUDECODE = previousClaude;
    if (previousRepo === undefined) delete process.env.HIVE_REPO;
    else process.env.HIVE_REPO = previousRepo;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("parent-only advisory hooks do not run in PI child sessions and resumed startup stays quiet", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-parent-only-"));
  const flowSessionContext = makeAdvisoryHook(directory, "flow-session-context.sh", "flow");
  const sessionHygieneReport = makeAdvisoryHook(directory, "session-hygiene-report.sh", "hygiene");
  const paths = { ...pathsFor(directory), flowSessionContext, sessionHygieneReport };
  const entries: unknown[] = [{ type: "message", message: { role: "user", content: [] } }];
  const fake = createFakePi();
  const notifications: string[] = [];
  const context = makeContext(directory, "child-session", "base", entries, notifications);
  const previousChild = process.env.PI_SUBAGENT_CHILD;
  try {
    process.env.PI_SUBAGENT_CHILD = "1";
    registerGeneralHiveHooks(fake.api, { paths });
    await invoke(fake, "session_start", { type: "session_start", reason: "startup" }, context);
    assert.equal(fake.messages.length, 0);
    assert.equal(notifications.length, 0);

    delete process.env.PI_SUBAGENT_CHILD;
    const resumed = createFakePi();
    registerGeneralHiveHooks(resumed.api, { paths });
    await invoke(resumed, "session_start", { type: "session_start", reason: "startup" }, context);
    assert.equal(resumed.messages.length, 0, "startup with active message history is a CLI resume");
  } finally {
    if (previousChild === undefined) delete process.env.PI_SUBAGENT_CHILD;
    else process.env.PI_SUBAGENT_CHILD = previousChild;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("hook readiness fails when the declared gate is not installed", () => {
  // rule-delivery is a GATE: if it is missing, writes go through ungated and
  // nothing says so. A readiness check that still answers "ready" hides it.
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-gate-readiness-"));
  const fake = createFakePi();
  const paths = pathsFor(directory);
  const previousChild = process.env.PI_SUBAGENT_CHILD;
  try {
    delete process.env.PI_SUBAGENT_CHILD;
    assert.equal(REQUIRED_HOOK_KEYS.includes("ruleDelivery"), true, "a gate is not advisory");
    for (const key of REQUIRED_HOOK_KEYS) {
      if (key === "ruleDelivery") continue;
      writeFileSync(paths[key], "#!/bin/sh\nexit 0\n");
      chmodSync(paths[key], 0o755);
    }
    registerGeneralHiveHooks(fake.api, { paths });
    const readiness = fake.tools.find((tool): tool is { readonly name: string; readonly execute: (...args: readonly unknown[]) => Promise<unknown> } => (
      typeof tool === "object" && tool !== null && "name" in tool && tool.name === "hive_hook_readiness" && "execute" in tool && typeof tool.execute === "function"
    ));
    assert.ok(readiness);
    return readiness.execute({}).then((result) => {
      assert.match(JSON.stringify(result), /"ready":false/u);
      assert.match(JSON.stringify(result), /rule-delivery\.py/u);
    });
  } finally {
    if (previousChild === undefined) delete process.env.PI_SUBAGENT_CHILD;
    else process.env.PI_SUBAGENT_CHILD = previousChild;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("hook readiness keeps the required hooks authoritative while exposing all eight statuses", () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-hook-status-"));
  const fake = createFakePi();
  const paths = pathsFor(directory);
  const previousChild = process.env.PI_SUBAGENT_CHILD;
  try {
    delete process.env.PI_SUBAGENT_CHILD;
    for (const key of REQUIRED_HOOK_KEYS) {
      writeFileSync(paths[key], "#!/bin/sh\nexit 0\n");
      chmodSync(paths[key], 0o755);
    }
    registerGeneralHiveHooks(fake.api, { paths });
    const statuses = getHiveHookStatuses(fake.api, paths);
    assert.equal(Object.keys(statuses).length, 8);
    assert.equal(statuses.flowSessionContext?.wired, true);
    assert.equal(statuses.ruleContext?.wired, true);
    assert.equal(statuses.ruleDelivery?.wired, true);
    assert.equal(statuses.sessionHygieneReport?.wired, true);
    assert.equal(statuses.reviewerGuard?.wired, false);
    const readiness = fake.tools.find((tool): tool is { readonly name: string; readonly execute: (...args: readonly unknown[]) => Promise<unknown> } => (
      typeof tool === "object" && tool !== null && "name" in tool && tool.name === "hive_hook_readiness" && "execute" in tool && typeof tool.execute === "function"
    ));
    assert.ok(readiness);
    const execute = readiness.execute;
    return execute({}).then((result) => {
      assert.equal(typeof result, "object");
      assert.equal(JSON.stringify(result).includes("flowSessionContext"), true);
      assert.equal(JSON.stringify(result).includes("ready"), true);
      assert.match(JSON.stringify(result), /"ready":true/u, "missing advisory scripts must not block required readiness");
    });
  } finally {
    if (previousChild === undefined) delete process.env.PI_SUBAGENT_CHILD;
    else process.env.PI_SUBAGENT_CHILD = previousChild;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("general Hive extension keeps status and Git read without registering plan mode", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-general-surface-"));
  const fake = createFakePi();
  const context = makeContext(directory, "session-a", "base", [], fake.notifications);
  try {
    registerGeneralHiveHooks(fake.api, { paths: pathsFor(directory) });
    assert.equal(fake.commands.has("hive-plan"), false);
    assert.equal(fake.commands.has("hive-status"), true);
    assert.equal(fake.tools.some((tool) => typeof tool === "object" && tool !== null && "name" in tool && tool.name === "hive_hook_readiness"), true);
    assert.equal(fake.tools.some((tool) => typeof tool === "object" && tool !== null && "name" in tool && tool.name === "hive_research_readiness"), true);
    assert.equal(fake.tools.some((tool) => typeof tool === "object" && tool !== null && "name" in tool && tool.name === "hive_git_read"), true);

    const status = fake.commands.get("hive-status");
    assert.ok(status);
    await status("", context);
    assert.match(fake.notifications.join("\n"), /Hive status/u);
    assert.match(fake.notifications.join("\n"), /research_readiness:/u);
    assert.doesNotMatch(fake.notifications.join("\n"), /plan mode/u);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test("general flow context ignores the legacy PI plan environment", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-no-plan-env-"));
  const flowContext = join(directory, "flow-context.sh");
  writeFileSync(flowContext, `#!/usr/bin/env node
let input = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => {
const payload = JSON.parse(input);
process.stdout.write(JSON.stringify({ additionalContext: "permission_mode=" + String(payload.permission_mode ?? "absent") }));
});
`);
  chmodSync(flowContext, 0o755);
  const fake = createFakePi();
  const context = makeContext(directory, "session-a");
  const previousPlanMode = process.env.PI_HIVE_PLAN_MODE;
  try {
    process.env.PI_HIVE_PLAN_MODE = "1";
    registerGeneralHiveHooks(fake.api, { paths: { ...pathsFor(directory), flowContext } });
    const result = await invoke(fake, "before_agent_start", { systemPrompt: "base" }, context);
    assert.match(JSON.stringify(result), /permission_mode=absent/u);
  } finally {
    if (previousPlanMode === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanMode;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("rule context maps parent Write/Edit/Bash calls and serializes hidden steer messages", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-rule-context-"));
  const log = join(directory, "rule.log");
  const ruleContext = join(directory, "rule-context.sh");
  writeFileSync(ruleContext, `#!/usr/bin/env node
const fs = require("node:fs");
let input = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => {
  const payload = JSON.parse(input);
  fs.appendFileSync(${JSON.stringify(log)}, JSON.stringify(payload) + "\\n");
  process.stdout.write(JSON.stringify({ additionalContext: payload.tool_name }));
});
`);
  chmodSync(ruleContext, 0o755);
  const fake = createFakePi();
  const context = makeContext(directory, "session-a", "base", [], fake.notifications);
  try {
    registerGeneralHiveHooks(fake.api, { paths: { ...pathsFor(directory), ruleContext } });
    await Promise.all([
      invoke(fake, "tool_call", { type: "tool_call", toolCallId: "write-1", toolName: "write", input: { file_path: "a.ts", content: "a" } }, context),
      invoke(fake, "tool_call", { type: "tool_call", toolCallId: "edit-1", toolName: "edit", input: { file_path: "b.ts", old_string: "b", new_string: "c" } }, context),
    ]);
    const payloads = readFileSync(log, "utf8").trim().split("\n").map((line) => JSON.parse(line) as Record<string, unknown>);
    assert.deepEqual(payloads.map((payload) => payload.tool_name), ["Write", "Edit"]);
    assert.equal(payloads[0]?.harness, "pi");
    assert.deepEqual(payloads[0]?.tool_input, { file_path: "a.ts", content: "a" });
    assert.equal(fake.messages.length, 2);
    assert.equal(JSON.stringify(fake.messages[0]?.options).includes("steer"), true);
    assert.match(JSON.stringify(fake.messages[0]?.message), /Write/u);
    assert.match(JSON.stringify(fake.messages[1]?.message), /Edit/u);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

function writeRuleDeliveryStub(directory: string, log: string, denyTools: readonly string[]): string {
  const ruleDelivery = join(directory, "rule-delivery.py");
  writeFileSync(ruleDelivery, `#!/usr/bin/env node
const fs = require("node:fs");
let input = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => {
  const payload = JSON.parse(input);
  fs.appendFileSync(${JSON.stringify(log)}, JSON.stringify(payload) + "\\n");
  if (${JSON.stringify(denyTools)}.includes(payload.tool_name)) {
    process.stdout.write(JSON.stringify({ hookSpecificOutput: { hookEventName: "PreToolUse", permissionDecision: "deny", permissionDecisionReason: "Held: read /rules/ts.md first, then re-issue this call." } }));
  }
});
`);
  chmodSync(ruleDelivery, 0o755);
  return ruleDelivery;
}

test("rule delivery blocks a held Write/Edit with the hook's own reason", async () => {
  // PI *can* refuse a tool call (the bash-policy path returns `block`), so the
  // hook keeps its gate here instead of degrading to advice.
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-rule-delivery-"));
  const log = join(directory, "delivery.log");
  const ruleDelivery = writeRuleDeliveryStub(directory, log, ["Write", "Edit"]);
  const fake = createFakePi();
  const context = makeContext(directory, "session-a", "base", [], fake.notifications);
  try {
    registerGeneralHiveHooks(fake.api, { paths: { ...pathsFor(directory), ruleDelivery } });
    const held = await invoke(fake, "tool_call", { type: "tool_call", toolCallId: "write-1", toolName: "write", input: { file_path: "a.ts", content: "a" } }, context);
    assert.deepEqual(held, { block: true, reason: "Held: read /rules/ts.md first, then re-issue this call." });

    const payloads = readFileSync(log, "utf8").trim().split("\n").map((line) => JSON.parse(line) as Record<string, unknown>);
    assert.equal(payloads[0]?.harness, "pi");
    assert.equal(payloads[0]?.tool_name, "Write");
    assert.deepEqual(payloads[0]?.tool_input, { file_path: "a.ts", content: "a" });
    // The gate speaks through the block, never through steered context.
    assert.equal(fake.messages.some((entry) => JSON.stringify(entry.message).includes("Held: read")), false);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test("rule delivery observes Read and Bash calls and never blocks them", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-rule-delivery-observe-"));
  const log = join(directory, "delivery.log");
  // A stub that would deny everything: only the tools the bridge gates can block.
  const ruleDelivery = writeRuleDeliveryStub(directory, log, ["Write", "Edit", "Read", "Bash"]);
  // A permissive bash-policy stub: without it the bash call is blocked by the
  // missing required hook and says nothing about rule-delivery.
  const bashPolicy = join(directory, "bash-policy.sh");
  writeFileSync(bashPolicy, "#!/usr/bin/env node\nprocess.stdin.resume();\nprocess.stdin.on(\"end\", () => {});\n");
  chmodSync(bashPolicy, 0o755);
  const fake = createFakePi();
  const context = makeContext(directory, "session-a", "base", [], fake.notifications);
  try {
    registerGeneralHiveHooks(fake.api, { paths: { ...pathsFor(directory), ruleDelivery, bashPolicy } });
    const read = await invoke(fake, "tool_call", { type: "tool_call", toolCallId: "read-1", toolName: "read", input: { file_path: "/rules/ts.md" } }, context);
    const bash = await invoke(fake, "tool_call", { type: "tool_call", toolCallId: "bash-1", toolName: "bash", input: { command: "cat /rules/ts.md" } }, context);
    assert.equal(read, undefined);
    assert.equal(bash, undefined);

    const payloads = readFileSync(log, "utf8").trim().split("\n").map((line) => JSON.parse(line) as Record<string, unknown>);
    // Bash reaches the hook now: a `cat` of the rule file is how the gate is
    // released when the model does not use the read tool.
    assert.deepEqual(payloads.map((payload) => payload.tool_name), ["Read", "Bash"]);
    assert.deepEqual(payloads[1]?.tool_input, { command: "cat /rules/ts.md" });
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test("rule delivery gates child sessions too and tells the hook who is writing", async () => {
  // Children are the sessions that write code; the parent-only exclusion exists
  // for steered-context noise, and a block steers nothing.
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-rule-delivery-child-"));
  const log = join(directory, "delivery.log");
  const ruleDelivery = writeRuleDeliveryStub(directory, log, ["Write", "Edit"]);
  const fake = createFakePi();
  const context = makeContext(directory, "session-child", "base", [], fake.notifications);
  const previousChild = process.env.PI_SUBAGENT_CHILD;
  const previousAgent = process.env.PI_HIVE_AGENT;
  try {
    process.env.PI_SUBAGENT_CHILD = "1";
    process.env.PI_HIVE_AGENT = "ts-backend-developer";
    registerGeneralHiveHooks(fake.api, { paths: { ...pathsFor(directory), ruleDelivery } });
    const held = await invoke(fake, "tool_call", { type: "tool_call", toolCallId: "write-1", toolName: "write", input: { file_path: "a.ts", content: "a" } }, context);
    assert.deepEqual(held, { block: true, reason: "Held: read /rules/ts.md first, then re-issue this call." });

    const payloads = readFileSync(log, "utf8").trim().split("\n").map((line) => JSON.parse(line) as Record<string, unknown>);
    assert.equal(payloads[0]?.hook_event_name, "PreToolUse");
    assert.equal(payloads[0]?.agent_type, "ts-backend-developer", "packs and read-only scoping need the identity");
    assert.equal(payloads[0]?.session_id, "session-child");
    // Advisories stay parent-only: no steered context in a child.
    assert.equal(fake.messages.length, 0);
  } finally {
    if (previousChild === undefined) delete process.env.PI_SUBAGENT_CHILD;
    else process.env.PI_SUBAGENT_CHILD = previousChild;
    if (previousAgent === undefined) delete process.env.PI_HIVE_AGENT;
    else process.env.PI_HIVE_AGENT = previousAgent;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("rule delivery labels the write with the identity this session loaded with", async () => {
  // A runner process hosts more than one child session, and `PI_HIVE_AGENT` is
  // process-global: the next launch overwrites it. Read per event, this
  // session's write would carry the other agent's name — the one failure the
  // gate must never have, since a wrong identity can skip a rule the real
  // caller does not carry.
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-rule-delivery-identity-"));
  const log = join(directory, "delivery.log");
  const ruleDelivery = writeRuleDeliveryStub(directory, log, []);
  const fake = createFakePi();
  const context = makeContext(directory, "session-child", "base", [], fake.notifications);
  const previousAgent = process.env.PI_HIVE_AGENT;
  try {
    process.env.PI_HIVE_AGENT = "ts-backend-developer";
    registerGeneralHiveHooks(fake.api, { paths: { ...pathsFor(directory), ruleDelivery } });
    process.env.PI_HIVE_AGENT = "review-code";
    await invoke(fake, "tool_call", { type: "tool_call", toolCallId: "write-1", toolName: "write", input: { file_path: "a.ts", content: "a" } }, context);

    const payloads = readFileSync(log, "utf8").trim().split("\n").map((line) => JSON.parse(line) as Record<string, unknown>);
    assert.equal(payloads[0]?.agent_type, "ts-backend-developer", "a later launch must never relabel this session");
  } finally {
    if (previousAgent === undefined) delete process.env.PI_HIVE_AGENT;
    else process.env.PI_HIVE_AGENT = previousAgent;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("a blank PI_HIVE_AGENT is the main thread, never an agent with an empty name", async () => {
  // The launcher clears the key when it has no name; a blank value that reached
  // the payload would be an identity no manifest entry matches.
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-rule-delivery-blank-"));
  const log = join(directory, "delivery.log");
  const ruleDelivery = writeRuleDeliveryStub(directory, log, []);
  const fake = createFakePi();
  const context = makeContext(directory, "session-a", "base", [], fake.notifications);
  const previousAgent = process.env.PI_HIVE_AGENT;
  try {
    process.env.PI_HIVE_AGENT = "   ";
    registerGeneralHiveHooks(fake.api, { paths: { ...pathsFor(directory), ruleDelivery } });
    await invoke(fake, "tool_call", { type: "tool_call", toolCallId: "write-1", toolName: "write", input: { file_path: "a.ts", content: "a" } }, context);

    const payloads = readFileSync(log, "utf8").trim().split("\n").map((line) => JSON.parse(line) as Record<string, unknown>);
    assert.equal("agent_type" in (payloads[0] ?? {}), false, "no identity is main-thread treatment");
  } finally {
    if (previousAgent === undefined) delete process.env.PI_HIVE_AGENT;
    else process.env.PI_HIVE_AGENT = previousAgent;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("rule delivery observes completed reads on tool_result", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-rule-delivery-result-"));
  const log = join(directory, "delivery.log");
  const ruleDelivery = writeRuleDeliveryStub(directory, log, []);
  const fake = createFakePi();
  const context = makeContext(directory, "session-a", "base", [], fake.notifications);
  try {
    registerGeneralHiveHooks(fake.api, { paths: { ...pathsFor(directory), ruleDelivery } });
    await invoke(fake, "tool_result", { type: "tool_result", toolCallId: "read-1", toolName: "read", input: { file_path: "/rules/ts.md" }, content: [], isError: false }, context);
    await invoke(fake, "tool_result", { type: "tool_result", toolCallId: "bash-1", toolName: "bash", input: { command: "cat /rules/ts.md" }, content: [], isError: false }, context);
    await invoke(fake, "tool_result", { type: "tool_result", toolCallId: "write-1", toolName: "write", input: { file_path: "a.ts" }, content: [], isError: false }, context);

    const payloads = readFileSync(log, "utf8").trim().split("\n").map((line) => JSON.parse(line) as Record<string, unknown>);
    assert.deepEqual(payloads.map((payload) => payload.tool_name), ["Read", "Bash"], "a completed write is not an observation");
    assert.equal(payloads[0]?.hook_event_name, "PostToolUse");
    assert.deepEqual(payloads[1]?.tool_input, { command: "cat /rules/ts.md" });
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test("rule delivery allows the write when the hook is silent or cannot run", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-rule-delivery-open-"));
  const log = join(directory, "delivery.log");
  const ruleDelivery = writeRuleDeliveryStub(directory, log, []);
  const fake = createFakePi();
  const context = makeContext(directory, "session-a", "base", [], fake.notifications);
  try {
    registerGeneralHiveHooks(fake.api, { paths: { ...pathsFor(directory), ruleDelivery } });
    const allowed = await invoke(fake, "tool_call", { type: "tool_call", toolCallId: "write-1", toolName: "write", input: { file_path: "a.ts", content: "a" } }, context);
    assert.equal(allowed, undefined);
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }

  // An absent hook script must never turn into a blocked write.
  const empty = mkdtempSync(join(tmpdir(), "hive-pi-rule-delivery-absent-"));
  const fakeWithout = createFakePi();
  const contextWithout = makeContext(empty, "session-b", "base", [], fakeWithout.notifications);
  try {
    registerGeneralHiveHooks(fakeWithout.api, { paths: pathsFor(empty) });
    const allowed = await invoke(fakeWithout, "tool_call", { type: "tool_call", toolCallId: "write-2", toolName: "write", input: { file_path: "a.ts", content: "a" } }, contextWithout);
    assert.equal(allowed, undefined);
  } finally {
    rmSync(empty, { recursive: true, force: true });
  }
});
