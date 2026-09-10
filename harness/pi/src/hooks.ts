import { constants, existsSync, statSync } from "node:fs";
import { join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import type {
  ExtensionAPI,
  ExtensionContext,
  ToolCallEvent,
  ToolDefinition,
  ToolResultEvent,
} from "@earendil-works/pi-coding-agent";
import { Type } from "typebox";
import { attachChildLifecycle, defaultChildRegistry } from "./child-registry.ts";
import { checkHookReadiness, runHook } from "./hook-runner.ts";
import type { ChildRegistry, HookPaths, HookRunnerOptions } from "./types.ts";

export const HIVE_CONTEXT7_MCP_TOOLS = {
  resolveLibraryId: "context7_resolve-library-id",
  queryDocs: "context7_query-docs",
} as const;

export type HiveContext7McpToolName = (typeof HIVE_CONTEXT7_MCP_TOOLS)[keyof typeof HIVE_CONTEXT7_MCP_TOOLS];

export interface HiveContext7McpInput {
  readonly tool: HiveContext7McpToolName;
  readonly args: Readonly<Record<string, string>>;
}

export type HiveMcpGuardResult =
  | { readonly allowed: true; readonly input: HiveContext7McpInput }
  | { readonly allowed: false; readonly reason: string };

const HIVE_MCP_REJECTION =
  "Hive MCP bridge only allows Context7 resolve-library-id or query-docs with their required string arguments.";
const HIVE_MCP_SCRIPT_CHILD_REJECTION = "Hive child agents cannot use the generic MCP script tool.";

function isPlainObject(value: unknown): value is Record<string, unknown> {
  if (typeof value !== "object" || value === null || Array.isArray(value)) return false;
  try {
    const prototype = Object.getPrototypeOf(value);
    return prototype === Object.prototype || prototype === null;
  } catch {
    return false;
  }
}

function hasExactlyKeys(value: Record<string, unknown>, keys: readonly string[]): boolean {
  const actual = Object.keys(value);
  return actual.length === keys.length && keys.every((key) => Object.prototype.hasOwnProperty.call(value, key));
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

/** Validate the closed input accepted by the PI MCP bridge before adapter execution. */
export function guardHiveMcpInput(value: unknown): HiveMcpGuardResult {
  if (!isPlainObject(value) || !hasExactlyKeys(value, ["tool", "args"])) {
    return { allowed: false, reason: HIVE_MCP_REJECTION };
  }

  const tool = value.tool;
  if (tool !== HIVE_CONTEXT7_MCP_TOOLS.resolveLibraryId && tool !== HIVE_CONTEXT7_MCP_TOOLS.queryDocs) {
    return { allowed: false, reason: HIVE_MCP_REJECTION };
  }
  const args = value.args;
  if (!isPlainObject(args)) return { allowed: false, reason: HIVE_MCP_REJECTION };

  const argKeys = tool === HIVE_CONTEXT7_MCP_TOOLS.resolveLibraryId
    ? ["query", "libraryName"]
    : ["libraryId", "query"];
  if (!hasExactlyKeys(args, argKeys) || !argKeys.every((key) => isNonEmptyString(args[key]))) {
    return { allowed: false, reason: HIVE_MCP_REJECTION };
  }

  return {
    allowed: true,
    input: {
      tool,
      args: Object.fromEntries(argKeys.map((key) => [key, args[key] as string])),
    },
  };
}

const PI_MCP_GUIDANCE = [
  "PI MCP bridge: use mcp only with {tool, args}.",
  `Allowed calls: ${HIVE_CONTEXT7_MCP_TOOLS.resolveLibraryId} with {query, libraryName}; ${HIVE_CONTEXT7_MCP_TOOLS.queryDocs} with {libraryId, query}.`,
].join(" ");

export interface HiveHookExtensionOptions {
  readonly paths?: Partial<HookPaths>;
  readonly runnerOptions?: HookRunnerOptions;
  readonly childRegistry?: ChildRegistry;
}

function firstExistingDirectory(candidates: readonly string[]): string {
  return candidates.find((candidate) => existsSync(candidate)) ?? candidates[0] ?? process.cwd();
}

export function defaultHookPaths(): HookPaths {
  const sourceDir = resolve(fileURLToPath(new URL("../../..", import.meta.url)), "global/hooks");
  const deployedDir = resolve(fileURLToPath(new URL("..", import.meta.url)), "global/hooks");
  const configuredDir = process.env.PI_HIVE_HOOK_DIR;
  const root = firstExistingDirectory([
    configuredDir ?? "",
    resolve(process.cwd(), "global/hooks"),
    deployedDir,
    sourceDir,
  ]);
  return {
    bashPolicy: join(root, "bash-policy/bash-policy.sh"),
    reviewerGuard: join(root, "reviewer-guard/reviewer-guard.sh"),
    postToolHub: join(root, "post-tool-hub/post-tool-hub.sh"),
    flowContext: join(root, "flow-context/flow-context.sh"),
    flowPlanCapture: join(root, "flow-plan-capture/flow-plan-capture.sh"),
  };
}

function mergePaths(overrides: Partial<HookPaths> | undefined): HookPaths {
  return { ...defaultHookPaths(), ...overrides };
}

function canonicalToolName(toolName: string): string {
  return toolName.toLowerCase() === "bash" ? "Bash" : toolName;
}

function currentSessionId(ctx: ExtensionContext): string {
  return ctx.sessionManager.getSessionId() || ctx.sessionManager.getSessionFile() || "pi-session";
}

function hookPayload(ctx: ExtensionContext, event: ToolCallEvent | ToolResultEvent): Record<string, unknown> {
  const input = "input" in event ? event.input : {};
  const payload: Record<string, unknown> = {
    harness: "pi",
    cwd: ctx.cwd,
    session_id: currentSessionId(ctx),
    tool_name: canonicalToolName(event.toolName),
    tool_input: input,
  };
  if (event.type === "tool_result") {
    payload.tool_response = {
      content: event.content,
      details: event.details,
      isError: event.isError,
    };
  }
  return payload;
}

function isExecutable(path: string): boolean {
  try {
    return statSync(path).isFile() && (statSync(path).mode & constants.S_IXUSR) !== 0;
  } catch {
    return false;
  }
}

const readinessSchema = Type.Object({});

export function createHookReadinessTool(paths: HookPaths): ToolDefinition<typeof readinessSchema> {
  return {
    name: "hive_hook_readiness",
    label: "Hive hook readiness",
    description: "Verify that all required Hive PI hook scripts are present and executable.",
    promptSnippet: "Verify Hive hooks are installed before delegating or reviewing",
    parameters: readinessSchema,
    executionMode: "sequential",
    async execute() {
      const required = Object.values(paths);
      const missing = required.filter((path) => !isExecutable(path));
      const ready = missing.length === 0;
      return {
        content: [{ type: "text" as const, text: ready ? "Hive hooks ready." : `Missing or non-executable hooks:\n${missing.join("\n")}` }],
        details: { ready, missing },
        isError: !ready,
      };
    },
  };
}

interface AdvisoryHookResult {
  readonly outcome: "allow" | "block" | "warning" | "error";
  readonly additionalContext?: string;
  readonly reason?: string;
}

function advisoryText(
  result: AdvisoryHookResult,
  scriptPath: string,
  sessionId: string,
  failureKeys: Set<string>,
): string | undefined {
  const warning = result.reason ? `[Hive hook warning] ${result.reason}` : undefined;
  const failed = result.outcome === "warning" || result.outcome === "error";
  if (!failed || !warning) return result.additionalContext;

  const failureKey = `${sessionId}\u0000${scriptPath}`;
  const firstFailure = !failureKeys.has(failureKey);
  failureKeys.add(failureKey);
  if (!firstFailure) return result.additionalContext;
  return result.additionalContext ? `${warning}\n${result.additionalContext}` : warning;
}

export function registerGeneralHiveHooks(
  pi: ExtensionAPI,
  options: HiveHookExtensionOptions = {},
): void {
  const paths = mergePaths(options.paths);
  const runnerOptions = options.runnerOptions;
  const childRegistry = options.childRegistry ?? defaultChildRegistry;
  const advisoryFailureKeys = new Set<string>();
  let parentSession = process.env.PI_SUBAGENT_PARENT_SESSION ?? "";
  attachChildLifecycle(pi, childRegistry, () => parentSession || process.env.PI_SUBAGENT_PARENT_SESSION || "");
  pi.registerTool(createHookReadinessTool(paths));

  pi.on("session_start", async (_event, ctx) => {
    const session = currentSessionId(ctx);
    parentSession = session;
    for (const key of advisoryFailureKeys) {
      if (!key.startsWith(`${session}\u0000`)) advisoryFailureKeys.delete(key);
    }
    if (!process.env.PI_SUBAGENT_PARENT_SESSION) process.env.PI_SUBAGENT_PARENT_SESSION = session;
    process.env.PI_HIVE_SESSION = session;
  });

  pi.on("tool_call", async (event, ctx) => {
    const toolName = event.toolName.toLowerCase();
    if (toolName === "mcpscript" && process.env.PI_SUBAGENT_CHILD === "1") {
      return { block: true, reason: HIVE_MCP_SCRIPT_CHILD_REJECTION, terminate: true };
    }
    if (toolName === "mcp") {
      const guard = guardHiveMcpInput(event.input);
      if (!guard.allowed) {
        return { block: true, reason: guard.reason, terminate: true };
      }
    }
    if (/^(?:subagent|delegate|spawn)(?:[-_]|$)/iu.test(event.toolName)) {
      childRegistry.register(parentSession || currentSessionId(ctx), event.toolCallId);
    }
    if (event.toolName.toLowerCase() !== "bash") return undefined;
    const result = await runHook(
      {
        scriptPath: paths.bashPolicy,
        cwd: ctx.cwd,
        payload: hookPayload(ctx, event),
        mode: "blocking",
        signal: ctx.signal,
      },
      runnerOptions,
    );
    if (result.outcome === "block") {
      return { block: true, reason: result.reason ?? "Blocked by Hive bash policy." };
    }
    return undefined;
  });

  pi.on("tool_result", async (event, ctx) => {
    if (/^(?:subagent|delegate|spawn)(?:[-_]|$)/iu.test(event.toolName)) {
      childRegistry.settle(event.toolCallId);
    }
    const result = await runHook(
      {
        scriptPath: paths.postToolHub,
        cwd: ctx.cwd,
        payload: hookPayload(ctx, event),
        mode: "advisory",
        signal: ctx.signal,
      },
      runnerOptions,
    );
    const context = advisoryText(result, paths.postToolHub, currentSessionId(ctx), advisoryFailureKeys);
    if (!context) return undefined;
    return {
      content: [...event.content, { type: "text" as const, text: context }],
      isError: event.isError,
      details: event.details,
    };
  });

  pi.on("before_agent_start", async (event, ctx) => {
    const result = await runHook(
      {
        scriptPath: paths.flowContext,
        cwd: ctx.cwd,
        payload: {
          harness: "pi",
          cwd: ctx.cwd,
          session_id: currentSessionId(ctx),
          permission_mode: process.env.PI_HIVE_PLAN_MODE === "1" ? "plan" : "normal",
        },
        mode: "advisory",
        signal: ctx.signal,
      },
      runnerOptions,
    );
    const context = advisoryText(result, paths.flowContext, currentSessionId(ctx), advisoryFailureKeys);
    const additions = [context, PI_MCP_GUIDANCE].filter((value): value is string => value !== undefined);
    return { systemPrompt: `${event.systemPrompt}\n\n${additions.join("\n\n")}` };
  });
}

export function createGeneralHiveHooksExtension(options: HiveHookExtensionOptions = {}) {
  return (pi: ExtensionAPI): void => registerGeneralHiveHooks(pi, options);
}

export default createGeneralHiveHooksExtension();
