import { chmodSync, existsSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import assert from "node:assert/strict";
import type {
  ExtensionAPI,
  ExtensionCommandContext,
  ExtensionContext,
  ToolCallEvent,
} from "@earendil-works/pi-coding-agent";
import { createChildRegistry } from "../src/child-registry.ts";
import { readPlanFile, registerHivePlanMode } from "../src/plan-mode.ts";
import type { HookPaths } from "../src/types.ts";

interface FakeEvents {
  readonly on: (channel: string, handler: (payload: unknown) => void) => () => void;
  readonly emit: (channel: string, payload: unknown) => void;
}

interface FakePi {
  api: ExtensionAPI;
  readonly events: FakeEvents;
  readonly commands: Map<string, (args: string, ctx: ExtensionCommandContext) => Promise<void>>;
  readonly handlers: Map<string, StoredHandler[]>;
  readonly sentMessages: string[];
  readonly notifications: string[];
  readonly appended: unknown[];
  activeTools: string[];
}

type StoredHandler = (event: unknown, ctx: ExtensionContext) => unknown | Promise<unknown>;

function fakeEvents(): FakeEvents {
  const listeners = new Map<string, Set<(payload: unknown) => void>>();
  return {
    on(channel, handler) {
      const channelListeners = listeners.get(channel) ?? new Set<(payload: unknown) => void>();
      channelListeners.add(handler);
      listeners.set(channel, channelListeners);
      return () => channelListeners.delete(handler);
    },
    emit(channel, payload) {
      for (const handler of listeners.get(channel) ?? []) handler(payload);
    },
  };
}

function makeContext(
  directory: string,
  entries: unknown[],
  notifications: string[],
  sessionId = "parent-session",
): ExtensionCommandContext {
  const sessionManager = {
    getEntries: () => entries,
    getBranch: () => entries,
    getSessionFile: () => join(directory, "session.jsonl"),
    getSessionId: () => sessionId,
  };
  const ui = {
    notify: (message: string) => notifications.push(message),
    setStatus: () => undefined,
  };
  const context: Partial<ExtensionContext> = {
    ui: ui as unknown as ExtensionContext["ui"],
    mode: "tui",
    hasUI: true,
    cwd: directory,
    sessionManager: sessionManager as unknown as ExtensionContext["sessionManager"],
    modelRegistry: {} as ExtensionContext["modelRegistry"],
    model: undefined,
    scopedModels: [],
    isIdle: () => true,
    isProjectTrusted: () => true,
    signal: undefined,
    abort: () => undefined,
    hasPendingMessages: () => false,
    shutdown: () => undefined,
    getContextUsage: () => undefined,
    compact: () => undefined,
    getSystemPrompt: () => "",
  };
  return context as ExtensionCommandContext;
}

function createFakePi(initialTools: string[]): FakePi {
  const events = fakeEvents();
  const commands = new Map<string, (args: string, ctx: ExtensionCommandContext) => Promise<void>>();
  const handlers = new Map<string, StoredHandler[]>();
  const sentMessages: string[] = [];
  const notifications: string[] = [];
  const appended: unknown[] = [];
  const fake: FakePi = {
    api: undefined as unknown as ExtensionAPI,
    events,
    commands,
    handlers,
    sentMessages,
    notifications,
    appended,
    activeTools: [...initialTools],
  };
  const on: ExtensionAPI["on"] = (event, handler) => {
    const registered = handlers.get(event) ?? [];
    registered.push(handler as unknown as StoredHandler);
    handlers.set(event, registered);
  };
  const registerCommand: ExtensionAPI["registerCommand"] = (name, options) => {
    commands.set(name, options.handler as (args: string, ctx: ExtensionCommandContext) => Promise<void>);
  };
  const sendUserMessage: ExtensionAPI["sendUserMessage"] = (content) => {
    sentMessages.push(typeof content === "string" ? content : "non-text");
  };
  const appendEntry: ExtensionAPI["appendEntry"] = (_type, data) => {
    appended.push(data);
  };
  const getActiveTools: ExtensionAPI["getActiveTools"] = () => [...fake.activeTools];
  const setActiveTools: ExtensionAPI["setActiveTools"] = (toolNames) => {
    fake.activeTools = [...toolNames];
  };
  fake.api = {
    on,
    registerTool: () => undefined,
    registerCommand,
    registerShortcut: () => undefined,
    registerFlag: () => undefined,
    registerMessageRenderer: () => undefined,
    registerMarkdownTransformer: () => undefined,
    registerEntryRenderer: () => undefined,
    getFlag: () => undefined,
    sendMessage: () => undefined,
    sendUserMessage,
    appendEntry,
    setSessionName: () => undefined,
    getSessionName: () => undefined,
    setLabel: () => undefined,
    exec: async () => ({ stdout: "", stderr: "", code: 0, killed: false }),
    getActiveTools,
    getAllTools: () => [],
    setActiveTools,
    getCommands: () => [],
    setModel: async () => false,
    getThinkingLevel: () => undefined,
    setThinkingLevel: () => undefined,
    registerProvider: () => undefined,
    unregisterProvider: () => undefined,
    events,
  } as unknown as ExtensionAPI;
  return fake;
}

async function invokeHandler(fake: FakePi, event: string, payload: unknown, context: ExtensionContext): Promise<unknown> {
  const handlers = fake.handlers.get(event) ?? [];
  let result: unknown;
  for (const handler of handlers) {
    result = await handler(payload, context);
  }
  return result;
}

function makeDeferredCaptureHook(directory: string, readyPath: string, releasePath: string): string {
  const hook = join(directory, "capture-deferred.sh");
  writeFileSync(hook, `#!/usr/bin/env node
const fs = require("node:fs");
let input = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => {
  JSON.parse(input);
  fs.writeFileSync(${JSON.stringify(readyPath)}, "ready");
  const timer = setInterval(() => {
    if (!fs.existsSync(${JSON.stringify(releasePath)})) return;
    clearInterval(timer);
    process.stdout.write(JSON.stringify({ status: "session_only", reason: "test session" }));
  }, 5);
});
`);
  chmodSync(hook, 0o755);
  return hook;
}

async function waitForFile(filePath: string): Promise<void> {
  for (let attempt = 0; attempt < 200; attempt += 1) {
    if (existsSync(filePath)) return;
    await new Promise((resolve) => setTimeout(resolve, 5));
  }
  assert.fail(`Timed out waiting for ${filePath}`);
}

test("plan mode snapshots and restores exact tools, blocks writes and approves once", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-plan-"));
  const captureHook = join(directory, "capture.sh");
  writeFileSync(captureHook, `#!/usr/bin/env node
const { createHash } = require("node:crypto");
let input = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => {
  const request = JSON.parse(input);
  if (request.harness !== "pi") process.exit(3);
  const plan = request.tool_response.plan;
  const sha256 = createHash("sha256").update(Buffer.from(plan, "utf8")).digest("hex");
  process.stdout.write(JSON.stringify({ status: "captured", path: process.cwd() + "/captured-plan.md", sha256 }));
});
`);
  chmodSync(captureHook, 0o755);
  const paths: Partial<HookPaths> = { flowPlanCapture: captureHook };
  const entries: unknown[] = [
    { id: "before-enter", type: "message", message: { role: "assistant", content: [{ type: "text", text: "# Old plan\n- must not be reused" }] } },
  ];
  const fake = createFakePi([
    "read",
    "bash",
    "edit",
    "write",
    "grep",
    "custom_read",
    "mem_search",
    "mem_save",
    "run_terminal_command",
    "mcp:github/create_issue",
    "mcp",
    "context7_resolve-library-id",
    "context7_query-docs",
  ]);
  const context = makeContext(directory, entries, fake.notifications);
  const previousPlanEnv = process.env.PI_HIVE_PLAN_MODE;
  delete process.env.PI_HIVE_PLAN_MODE;
  try {
    registerHivePlanMode(fake.api, { paths });
    const enter = fake.commands.get("hive-plan");
    const status = fake.commands.get("hive-status");
    assert.ok(enter);
    assert.ok(status);

    await enter("enter", context);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, "1");
    assert.equal(fake.activeTools.includes("bash"), false);
    assert.equal(fake.activeTools.includes("edit"), false);
    assert.equal(fake.activeTools.includes("write"), false);
    assert.equal(fake.activeTools.includes("read"), true);
    assert.equal(fake.activeTools.includes("custom_read"), false);
    assert.equal(fake.activeTools.includes("mem_search"), true);
    assert.equal(fake.activeTools.includes("mem_save"), false);
    assert.equal(fake.activeTools.includes("run_terminal_command"), false);
    assert.equal(fake.activeTools.includes("mcp:github/create_issue"), false);
    assert.equal(fake.activeTools.includes("context7_resolve-library-id"), false);
    assert.equal(fake.activeTools.includes("context7_query-docs"), false);
    assert.equal(fake.activeTools.includes("mcp"), true);
    const blocked = await invokeHandler(fake, "tool_call", {
      type: "tool_call",
      toolCallId: "bash-1",
      toolName: "bash",
      input: { command: "touch forbidden" },
    } satisfies ToolCallEvent, context);
    assert.deepEqual(blocked, {
      block: true,
      reason: "Hive plan mode: 'bash' is blocked until the plan is explicitly approved.",
      terminate: true,
    });
    const blockedMcp = await invokeHandler(fake, "tool_call", {
      type: "tool_call",
      toolCallId: "mcp-1",
      toolName: "mcp",
      input: { search: "context7" },
    } satisfies ToolCallEvent, context);
    assert.deepEqual(blockedMcp, {
      block: true,
      reason: "Hive MCP bridge only allows Context7 resolve-library-id or query-docs with their required string arguments.",
      terminate: true,
    });
    const allowedMcp = await invokeHandler(fake, "tool_call", {
      type: "tool_call",
      toolCallId: "mcp-2",
      toolName: "mcp",
      input: {
        tool: "context7_query-docs",
        args: { libraryId: "/facebook/react", query: "hooks" },
      },
    } satisfies ToolCallEvent, context);
    assert.equal(allowedMcp, undefined);
    const blockedMcpScript = await invokeHandler(fake, "tool_call", {
      type: "tool_call",
      toolCallId: "mcp-script-1",
      toolName: "mcpScript",
      input: { code: "await tools.search({ query: 'context7' })" },
    } satisfies ToolCallEvent, context);
    assert.deepEqual(blockedMcpScript, {
      block: true,
      reason: "Hive plan mode: 'mcpScript' is blocked until the plan is explicitly approved.",
      terminate: true,
    });

    await invokeHandler(fake, "agent_settled", { type: "agent_settled" }, context);
    await status("", context);
    assert.match(fake.notifications.at(-1) ?? "", /candidate: none/u);
    entries.push({
      id: "after-enter",
      type: "message",
      message: { role: "assistant", content: [{ type: "text", text: "# Approved plan\n- inspect" }] },
    });
    await invokeHandler(fake, "agent_settled", { type: "agent_settled" }, context);
    await status("", context);
    assert.match(fake.notifications.at(-1) ?? "", /candidate: \d+ bytes/u);
    assert.match(fake.notifications.at(-1) ?? "", /sha256: [a-f0-9]{64}/u);
    assert.match(fake.notifications.at(-1) ?? "", /hooks: /u);
    assert.match(fake.notifications.at(-1) ?? "", /tools: /u);

    await enter("approve", context);
    assert.equal(fake.sentMessages.length, 1);
    assert.match(fake.sentMessages[0] ?? "", /# Approved plan/u);
    assert.equal(fake.activeTools.includes("bash"), true);
    assert.equal(fake.activeTools.includes("edit"), true);
    assert.equal(fake.activeTools.includes("write"), true);
    assert.equal(fake.activeTools.includes("mcp"), true);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, undefined);
    assert.equal(fake.appended.length >= 3, true);
    await enter("approve", context);
    assert.equal(fake.sentMessages.length, 1);
  } finally {
    if (previousPlanEnv === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanEnv;
    assert.equal(readFileSync(captureHook).length > 0, true);
    rmSync(directory, { recursive: true, force: true });
  }
});

test("plan mode captures the first plan from an empty branch", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-empty-branch-"));
  const captureHook = join(directory, "capture.sh");
  writeFileSync(captureHook, "#!/usr/bin/env node\nprocess.stdin.resume(); process.stdin.on('end', () => process.stdout.write(JSON.stringify({ status: 'session_only', reason: 'test session' })));\n");
  chmodSync(captureHook, 0o755);
  const entries: unknown[] = [];
  const fake = createFakePi(["read", "bash"]);
  const context = makeContext(directory, entries, fake.notifications);
  const previousPlanEnv = process.env.PI_HIVE_PLAN_MODE;
  delete process.env.PI_HIVE_PLAN_MODE;
  try {
    registerHivePlanMode(fake.api, { paths: { flowPlanCapture: captureHook } });
    const command = fake.commands.get("hive-plan");
    assert.ok(command);
    await command("enter", context);
    entries.push({ id: "first-plan", type: "message", message: { role: "assistant", content: [{ type: "text", text: "# First plan" }] } });
    await invokeHandler(fake, "agent_settled", { type: "agent_settled" }, context);
    await command("approve", context);
    assert.equal(fake.sentMessages.length, 1);
    assert.match(fake.sentMessages[0] ?? "", /# First plan/u);
  } finally {
    if (previousPlanEnv === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanEnv;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("plan approval serializes concurrent approvals and executes once", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-approve-race-"));
  const readyPath = join(directory, "capture-ready");
  const releasePath = join(directory, "capture-release");
  const captureHook = makeDeferredCaptureHook(directory, readyPath, releasePath);
  const entries: unknown[] = [{ id: "before-enter", type: "message", message: { role: "assistant", content: [{ type: "text", text: "old" }] } }];
  const fake = createFakePi(["read", "bash"]);
  const context = makeContext(directory, entries, fake.notifications);
  const previousPlanEnv = process.env.PI_HIVE_PLAN_MODE;
  delete process.env.PI_HIVE_PLAN_MODE;
  try {
    registerHivePlanMode(fake.api, { paths: { flowPlanCapture: captureHook } });
    const command = fake.commands.get("hive-plan");
    assert.ok(command);
    await command("enter", context);
    entries.push({ id: "after-enter", type: "message", message: { role: "assistant", content: [{ type: "text", text: "# Race plan" }] } });
    await invokeHandler(fake, "agent_settled", { type: "agent_settled" }, context);
    const first = command("approve", context);
    await waitForFile(readyPath);
    await command("approve", context);
    assert.match(fake.notifications.at(-1) ?? "", /already in progress/u);
    writeFileSync(releasePath, "release");
    await first;
    assert.equal(fake.sentMessages.length, 1);
    assert.equal(fake.activeTools.includes("bash"), true);
  } finally {
    if (previousPlanEnv === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanEnv;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("cancelling during capture invalidates the pending approval", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-approve-cancel-"));
  const readyPath = join(directory, "capture-ready");
  const releasePath = join(directory, "capture-release");
  const captureHook = makeDeferredCaptureHook(directory, readyPath, releasePath);
  const entries: unknown[] = [{ id: "before-enter", type: "message", message: { role: "assistant", content: [{ type: "text", text: "old" }] } }];
  const fake = createFakePi(["read", "bash"]);
  const context = makeContext(directory, entries, fake.notifications);
  const previousPlanEnv = process.env.PI_HIVE_PLAN_MODE;
  delete process.env.PI_HIVE_PLAN_MODE;
  try {
    registerHivePlanMode(fake.api, { paths: { flowPlanCapture: captureHook } });
    const command = fake.commands.get("hive-plan");
    assert.ok(command);
    await command("enter", context);
    entries.push({ id: "after-enter", type: "message", message: { role: "assistant", content: [{ type: "text", text: "# Cancel plan" }] } });
    await invokeHandler(fake, "agent_settled", { type: "agent_settled" }, context);
    const pending = command("approve", context);
    await waitForFile(readyPath);
    await command("cancel", context);
    writeFileSync(releasePath, "release");
    await pending;
    assert.equal(fake.sentMessages.length, 0);
    assert.equal(fake.activeTools.includes("bash"), true);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, undefined);
  } finally {
    if (previousPlanEnv === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanEnv;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("plan mode waits for registered children before enter while parent approval keeps child work restricted", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-child-plan-"));
  const captureHook = join(directory, "capture.sh");
  writeFileSync(captureHook, `#!/usr/bin/env node
process.stdin.resume();
process.stdin.on("end", () => process.stdout.write(JSON.stringify({ status: "session_only", reason: "test session" })));
`);
  chmodSync(captureHook, 0o755);
  const fake = createFakePi(["read", "bash"]);
  const context = makeContext(directory, [
    { id: "before-enter", type: "message", message: { role: "assistant", content: [{ type: "text", text: "# Old child plan" }] } },
  ], fake.notifications);
  const registry = createChildRegistry();
  const previousPlanEnv = process.env.PI_HIVE_PLAN_MODE;
  delete process.env.PI_HIVE_PLAN_MODE;
  try {
    registerHivePlanMode(fake.api, {
      paths: { flowPlanCapture: captureHook },
      childRegistry: registry,
    });
    const command = fake.commands.get("hive-plan");
    assert.ok(command);

    fake.events.emit("subagent:async-started", { id: "child-1", sessionId: "parent-session" });
    await command("enter", context);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, undefined);
    assert.match(fake.notifications.at(-1) ?? "", /child agents/u);

    fake.events.emit("subagent:async-complete", { runId: "child-1" });
    await command("enter", context);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, "1");
    (context.sessionManager.getBranch() as unknown[]).push({
      id: "after-enter",
      type: "message",
      message: { role: "assistant", content: [{ type: "text", text: "# Child-safe plan" }] },
    });
    await invokeHandler(fake, "agent_settled", { type: "agent_settled" }, context);

    fake.events.emit("subagent:async-started", { id: "child-2", sessionId: "parent-session" });
    await command("approve", context);
    assert.equal(fake.sentMessages.length, 1);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, undefined);
    assert.equal(fake.activeTools.includes("bash"), true);
    assert.equal(registry.hasActiveChildren("parent-session"), true);
  } finally {
    if (previousPlanEnv === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanEnv;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("plan mode aborts entry if the agent becomes active during child reconciliation", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-enter-race-"));
  const fake = createFakePi(["read", "bash"]);
  const context = makeContext(directory, [], fake.notifications);
  let idle = true;
  (context as unknown as { isIdle: () => boolean }).isIdle = () => idle;
  const previousPlanEnv = process.env.PI_HIVE_PLAN_MODE;
  delete process.env.PI_HIVE_PLAN_MODE;
  try {
    registerHivePlanMode(fake.api, { childRegistry: createChildRegistry() });
    const command = fake.commands.get("hive-plan");
    assert.ok(command);
    const entering = command("enter", context);
    await new Promise((resolve) => setTimeout(resolve, 10));
    idle = false;
    await entering;
    assert.equal(process.env.PI_HIVE_PLAN_MODE, undefined);
    assert.equal(fake.activeTools.includes("bash"), true);
    assert.match(fake.notifications.at(-1) ?? "", /state changed/u);
  } finally {
    if (previousPlanEnv === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanEnv;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("plan mode keeps restrictions when capture hash does not match", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-hash-"));
  const captureHook = join(directory, "capture.sh");
  writeFileSync(captureHook, `#!/usr/bin/env node
process.stdout.write(JSON.stringify({
  status: "captured",
  path: "/tmp/hive-captured-plan.md",
  sha256: "0000000000000000000000000000000000000000000000000000000000000000"
}));
`);
  chmodSync(captureHook, 0o755);
  const entries: unknown[] = [{ id: "before-enter", type: "message", message: { role: "assistant", content: [{ type: "text", text: "old" }] } }];
  const fake = createFakePi(["read", "bash"]);
  const context = makeContext(directory, entries, fake.notifications);
  const previousPlanEnv = process.env.PI_HIVE_PLAN_MODE;
  delete process.env.PI_HIVE_PLAN_MODE;
  try {
    registerHivePlanMode(fake.api, { paths: { flowPlanCapture: captureHook } });
    const command = fake.commands.get("hive-plan");
    assert.ok(command);
    await command("enter", context);
    entries.push({
      id: "after-enter",
      type: "message",
      message: { role: "assistant", content: [{ type: "text", text: "# Approved plan\n- inspect" }] },
    });
    await invokeHandler(fake, "agent_settled", { type: "agent_settled" }, context);
    await command("approve", context);
    assert.equal(fake.sentMessages.length, 0);
    assert.equal(fake.activeTools.includes("bash"), false);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, "1");
    assert.match(fake.notifications.at(-1) ?? "", /hash does not match/u);
  } finally {
    if (previousPlanEnv === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanEnv;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("inherited child plan mode cannot cancel or approve itself", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-inherited-plan-"));
  const fake = createFakePi(["read", "bash", "write"]);
  const context = makeContext(directory, [
    { type: "message", message: { role: "assistant", content: [{ type: "text", text: "# Inherited plan" }] } },
  ], fake.notifications, "child-session");
  const previousPlanEnv = process.env.PI_HIVE_PLAN_MODE;
  const previousParentEnv = process.env.PI_SUBAGENT_PARENT_SESSION;
  process.env.PI_HIVE_PLAN_MODE = "1";
  process.env.PI_SUBAGENT_PARENT_SESSION = "root-session";
  try {
    registerHivePlanMode(fake.api);
    const command = fake.commands.get("hive-plan");
    assert.ok(command);
    await invokeHandler(fake, "session_start", { type: "session_start" }, context);
    assert.equal(fake.activeTools.includes("bash"), false);
    assert.equal(fake.activeTools.includes("write"), false);

    await command("cancel", context);
    await command("approve", context);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, "1");
    assert.equal(fake.sentMessages.length, 0);
    assert.match(fake.notifications.at(-1) ?? "", /cannot approve/u);
  } finally {
    if (previousPlanEnv === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanEnv;
    if (previousParentEnv === undefined) delete process.env.PI_SUBAGENT_PARENT_SESSION;
    else process.env.PI_SUBAGENT_PARENT_SESSION = previousParentEnv;
    rmSync(directory, { recursive: true, force: true });
  }
});

test("readPlanFile rejects oversized and non-regular paths before reading", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-plan-file-"));
  const oversized = join(directory, "oversized.md");
  const link = join(directory, "plan-link.md");
  const fifo = join(directory, "plan.fifo");
  writeFileSync(oversized, Buffer.alloc(2_000_001, 120));
  try {
    await assert.rejects(readPlanFile("oversized.md", directory), /exceeds/u);
    await assert.rejects(readPlanFile(".", directory), /regular file/u);
    if (process.platform !== "win32") {
      const fifoResult = spawnSync("mkfifo", [fifo]);
      assert.equal(fifoResult.status, 0);
      const started = Date.now();
      await assert.rejects(readPlanFile("plan.fifo", directory), /regular file/u);
      assert.ok(Date.now() - started < 500, "FIFO rejection must not block on open");
    }
    if (process.platform !== "win32") {
      const symlinkResult = spawnSync("ln", ["-s", oversized, link]);
      assert.equal(symlinkResult.status, 0);
      await assert.rejects(readPlanFile("plan-link.md", directory), /symbolic link/u);
    }
  } finally {
    rmSync(directory, { recursive: true, force: true });
  }
});

test("switching sessions resets plan state and resuming the old session restores its persisted restrictions", async () => {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-session-switch-"));
  const entriesA: unknown[] = [];
  const entriesB: unknown[] = [];
  const fake = createFakePi(["read", "bash", "write"]);
  const contextA = makeContext(directory, entriesA, fake.notifications, "session-a");
  const contextB = makeContext(directory, entriesB, fake.notifications, "session-b");
  const previousPlanEnv = process.env.PI_HIVE_PLAN_MODE;
  const previousParentEnv = process.env.PI_SUBAGENT_PARENT_SESSION;
  const previousChildEnv = process.env.PI_SUBAGENT_CHILD;
  delete process.env.PI_HIVE_PLAN_MODE;
  delete process.env.PI_SUBAGENT_PARENT_SESSION;
  delete process.env.PI_SUBAGENT_CHILD;
  try {
    registerHivePlanMode(fake.api);
    const command = fake.commands.get("hive-plan");
    assert.ok(command);
    await invokeHandler(fake, "session_start", { type: "session_start" }, contextA);
    await command("enter", contextA);
    const persisted = fake.appended.at(-1);
    assert.ok(persisted);
    entriesA.push({ type: "custom", customType: "hive-plan-state", data: persisted });
    assert.equal(fake.activeTools.includes("bash"), false);

    await invokeHandler(fake, "session_start", { type: "session_start" }, contextB);
    assert.equal(fake.activeTools.includes("bash"), true);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, undefined);

    await invokeHandler(fake, "session_start", { type: "session_start" }, contextA);
    assert.equal(fake.activeTools.includes("bash"), false);
    assert.equal(process.env.PI_HIVE_PLAN_MODE, "1");
  } finally {
    if (previousPlanEnv === undefined) delete process.env.PI_HIVE_PLAN_MODE;
    else process.env.PI_HIVE_PLAN_MODE = previousPlanEnv;
    if (previousParentEnv === undefined) delete process.env.PI_SUBAGENT_PARENT_SESSION;
    else process.env.PI_SUBAGENT_PARENT_SESSION = previousParentEnv;
    if (previousChildEnv === undefined) delete process.env.PI_SUBAGENT_CHILD;
    else process.env.PI_SUBAGENT_CHILD = previousChildEnv;
    rmSync(directory, { recursive: true, force: true });
  }
});
