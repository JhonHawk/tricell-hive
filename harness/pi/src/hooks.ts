import { constants, existsSync, readFileSync, statSync } from "node:fs";
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
import { createGitReadTool } from "./git-read.ts";
import { checkHookReadiness, runHook } from "./hook-runner.ts";
import { formatHiveResearchStatus, registerHiveResearchReadiness } from "./research.ts";
import type { ChildRegistry, HookPaths, HookRunnerOptions } from "./types.ts";

export const HIVE_CONTEXT7_MCP_TOOLS = {
  resolveLibraryId: "context7_resolve-library-id",
  queryDocs: "context7_query-docs",
} as const;

export type HiveContext7McpToolName = (typeof HIVE_CONTEXT7_MCP_TOOLS)[keyof typeof HIVE_CONTEXT7_MCP_TOOLS];

interface HiveMcpToolPolicy {
  readonly requiredStringArgs?: readonly string[];
  /** Mutating tool: callable from the parent session only. */
  readonly write?: boolean;
}

interface HiveMcpServerPolicy {
  readonly url: string;
  readonly auth?: string;
  readonly headers?: Readonly<Record<string, string>>;
  readonly tools: Readonly<Record<string, HiveMcpToolPolicy>>;
}

export interface HiveMcpAllowlist {
  readonly servers: Readonly<Record<string, HiveMcpServerPolicy>>;
  /** Set when the allowlist could not be read: the guard then refuses everything. */
  readonly error?: string;
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  if (typeof value !== "object" || value === null || Array.isArray(value)) return false;
  try {
    const prototype = Object.getPrototypeOf(value);
    return prototype === Object.prototype || prototype === null;
  } catch {
    return false;
  }
}

/**
 * Read the allowlist that both the guard and the deploy-time `includeTools` cut
 * derive from. A read or shape failure never throws: it yields an empty
 * allowlist, which makes the guard refuse every MCP call.
 */
export function loadHiveMcpAllowlist(source: string | URL): HiveMcpAllowlist {
  let parsed: unknown;
  try {
    parsed = JSON.parse(readFileSync(source, "utf8")) as unknown;
  } catch (error) {
    return { servers: {}, error: error instanceof Error ? error.message : String(error) };
  }
  if (!isPlainObject(parsed)) {
    return { servers: {}, error: `${String(source)} does not contain a JSON object` };
  }
  for (const [server, policy] of Object.entries(parsed)) {
    if (!isPlainObject(policy) || typeof policy.url !== "string" || !isPlainObject(policy.tools)) {
      return { servers: {}, error: `${String(source)} has a malformed "${server}" entry` };
    }
  }
  return { servers: parsed as Readonly<Record<string, HiveMcpServerPolicy>> };
}

const HIVE_MCP_ALLOWLIST = loadHiveMcpAllowlist(new URL("./mcp-allowlist.json", import.meta.url));

interface HiveMcpServerGuidance {
  readonly label: string;
  /** Proxy name shown as the call example, so the prefix is unambiguous. */
  readonly example: string;
}

const MCP_SERVER_GUIDANCE: Readonly<Record<string, HiveMcpServerGuidance>> = {
  linear: { label: "Linear", example: "linear_list_issues" },
  heroui: { label: "HeroUI", example: "heroui_get_component_docs" },
};

function allowedToolsFor(allowlist: HiveMcpAllowlist): Readonly<Record<string, HiveMcpToolPolicy>> {
  return Object.freeze(
    Object.fromEntries(
      Object.entries(allowlist.servers).flatMap(([server, policy]) =>
        Object.entries(policy.tools).map(([tool, toolPolicy]) => [`${server}_${tool}`, toolPolicy] as const),
      ),
    ),
  );
}

export const HIVE_MCP_ALLOWED_TOOLS = allowedToolsFor(HIVE_MCP_ALLOWLIST);

export interface HiveMcpInput {
  readonly tool: string;
  readonly args: Readonly<Record<string, unknown>>;
}

/** @deprecated Retained for importers written against the Context7-only bridge. */
export type HiveContext7McpInput = HiveMcpInput;

export type HiveMcpGuardResult =
  | { readonly allowed: true; readonly input: HiveMcpInput }
  | { readonly allowed: false; readonly reason: string };

const HIVE_MCP_SCRIPT_CHILD_REJECTION = "Hive child agents cannot use the generic MCP script tool.";
const HIVE_MCP_SHAPE_REJECTION = "Hive MCP bridge: call mcp with exactly {tool, args}.";
const HIVE_MCP_ARGS_OBJECT_REJECTION = "Hive MCP bridge: args must be a plain object.";
const HIVE_MCP_WRITE_CHILD_REJECTION =
  "Hive MCP bridge: write tools (save_issue, save_comment) are only callable from the parent session.";
const HIVE_MCP_MAX_ARGS_DEPTH = 4;
const HIVE_MCP_MAX_ARGS_LENGTH = 16384;
const HIVE_MCP_FORBIDDEN_ARG_KEYS: readonly string[] = ["__proto__", "constructor", "prototype"];

function unknownToolRejection(allowlist: HiveMcpAllowlist, tool?: string): string {
  if (allowlist.error !== undefined) {
    return `Hive MCP bridge: allowlist unavailable (${allowlist.error}); all MCP calls are refused.`;
  }
  const servers = Object.keys(allowlist.servers).map((server) => `${server}_*`);
  const base = `Hive MCP bridge: unknown tool; allowlisted servers are ${servers.join(", ")}.`;
  // A bare name (list_issues) that belongs to exactly one server gets the prefixed spelling back.
  const owners = tool === undefined
    ? []
    : Object.entries(allowlist.servers)
        .filter(([, policy]) => Object.prototype.hasOwnProperty.call(policy.tools, tool))
        .map(([server]) => `${server}_${tool}`);
  return owners.length === 1 ? `${base} Did you mean ${owners[0]}? Tool names are always <server>_<tool>.` : base;
}

function requiredArgsRejection(tool: string, keys: readonly string[]): string {
  return `Hive MCP bridge: ${tool} requires exactly {${keys.join(", ")}} as non-empty strings.`;
}

function hasExactlyKeys(value: Record<string, unknown>, keys: readonly string[]): boolean {
  const actual = Object.keys(value);
  return actual.length === keys.length && keys.every((key) => Object.prototype.hasOwnProperty.call(value, key));
}

function isNonEmptyString(value: unknown): value is string {
  return typeof value === "string" && value.trim().length > 0;
}

type HiveMcpArgsFault = "data" | "key" | "depth";

/** Locate the first reason the args are not plain JSON data within budget. */
function findArgsFault(value: unknown, depth: number): HiveMcpArgsFault | undefined {
  if (value === null) return undefined;
  const kind = typeof value;
  if (kind === "string" || kind === "boolean") return undefined;
  if (kind === "number") return Number.isFinite(value) ? undefined : "data";
  if (depth > HIVE_MCP_MAX_ARGS_DEPTH) return "depth";
  if (Array.isArray(value)) {
    for (const item of value) {
      const fault = findArgsFault(item, depth + 1);
      if (fault !== undefined) return fault;
    }
    return undefined;
  }
  if (isPlainObject(value)) {
    for (const key of Object.keys(value)) {
      if (HIVE_MCP_FORBIDDEN_ARG_KEYS.includes(key)) return "key";
      const fault = findArgsFault(value[key], depth + 1);
      if (fault !== undefined) return fault;
    }
    return undefined;
  }
  return "data";
}

function argsFaultRejection(fault: HiveMcpArgsFault): string {
  if (fault === "key") {
    return `Hive MCP bridge: args must not contain the keys ${HIVE_MCP_FORBIDDEN_ARG_KEYS.join(", ").replace(/, ([^,]*)$/u, " or $1")}.`;
  }
  if (fault === "depth") {
    return `Hive MCP bridge: args nesting exceeds the maximum depth of ${HIVE_MCP_MAX_ARGS_DEPTH}.`;
  }
  return "Hive MCP bridge: args must contain only JSON data (objects, arrays, strings, finite numbers, booleans, null).";
}

export interface HiveMcpGuardOptions {
  /** Defaults to the PI child marker in the environment. */
  readonly childSession?: boolean;
}

export type HiveMcpGuard = (value: unknown, options?: HiveMcpGuardOptions) => HiveMcpGuardResult;

/** Build the input guard for one allowlist; an unreadable allowlist refuses everything. */
export function createHiveMcpGuard(allowlist: HiveMcpAllowlist): HiveMcpGuard {
  const allowedTools = allowedToolsFor(allowlist);
  return (value, options) => {
    if (!isPlainObject(value) || !hasExactlyKeys(value, ["tool", "args"])) {
      return { allowed: false, reason: HIVE_MCP_SHAPE_REJECTION };
    }

    const tool = value.tool;
    const policy = typeof tool === "string" && Object.prototype.hasOwnProperty.call(allowedTools, tool)
      ? allowedTools[tool]
      : undefined;
    if (typeof tool !== "string" || policy === undefined) {
      return { allowed: false, reason: unknownToolRejection(allowlist, typeof tool === "string" ? tool : undefined) };
    }
    const childSession = options?.childSession ?? process.env.PI_SUBAGENT_CHILD === "1";
    if (policy.write === true && childSession) {
      return { allowed: false, reason: HIVE_MCP_WRITE_CHILD_REJECTION };
    }
    const args = value.args;
    if (!isPlainObject(args)) return { allowed: false, reason: HIVE_MCP_ARGS_OBJECT_REJECTION };

    const argKeys = policy.requiredStringArgs;
    if (argKeys !== undefined) {
      if (!hasExactlyKeys(args, argKeys) || !argKeys.every((key) => isNonEmptyString(args[key]))) {
        return { allowed: false, reason: requiredArgsRejection(tool, argKeys) };
      }
      return {
        allowed: true,
        input: {
          tool,
          args: Object.fromEntries(argKeys.map((key) => [key, args[key] as string])),
        },
      };
    }

    const fault = findArgsFault(args, 1);
    if (fault !== undefined) return { allowed: false, reason: argsFaultRejection(fault) };
    if (JSON.stringify(args).length > HIVE_MCP_MAX_ARGS_LENGTH) {
      return {
        allowed: false,
        reason: `Hive MCP bridge: args exceed the maximum serialized size of ${HIVE_MCP_MAX_ARGS_LENGTH} bytes.`,
      };
    }
    return { allowed: true, input: { tool, args } };
  };
}

/** Validate the allowlisted input accepted by the PI MCP bridge before adapter execution. */
export const guardHiveMcpInput: HiveMcpGuard = createHiveMcpGuard(HIVE_MCP_ALLOWLIST);

function context7GuidanceSentence(): string {
  const calls = Object.entries(HIVE_MCP_ALLOWLIST.servers.context7?.tools ?? {}).map(
    ([tool, policy]) => `context7_${tool} with {${(policy.requiredStringArgs ?? []).join(", ")}}`,
  );
  return `Allowed calls: ${calls.join("; ")}.`;
}

function serverGuidanceSentence(server: string, policy: HiveMcpServerPolicy): string {
  const tools = Object.keys(policy.tools);
  const guidance = MCP_SERVER_GUIDANCE[server];
  const label = guidance?.label ?? server;
  const example = guidance?.example ?? `${server}_${tools[0] ?? ""}`;
  return `${label} (e.g. ${example}): ${tools.join(", ")}.`;
}

/** Model-facing: a session guessed a package skill under ~/.agents/skills after reading a Hive skill there. */
export const PI_SKILL_PATH_GUIDANCE =
  "PI skills: read a skill only at the <location> listed in <available_skills>; never guess its path. " +
  "~/.agents/skills/ holds only Hive's shared skills; package skills (e.g. pi-subagents, council-mode) live under the PI agent directory's npm/node_modules/<package>/skills/.";

export const PI_MCP_GUIDANCE = [
  "PI MCP bridge: use mcp only with {tool, args}. Tool names are always <server>_<tool> (call linear_list_issues, never list_issues).",
  context7GuidanceSentence(),
  ...Object.entries(HIVE_MCP_ALLOWLIST.servers)
    .filter(([server]) => server !== "context7")
    .map(([server, policy]) => serverGuidanceSentence(server, policy)),
  "Linear writes (save_issue, save_comment) are callable from the parent session only.",
].join(" ");

export interface HiveHookExtensionOptions {
  readonly paths?: Partial<HookPaths>;
  readonly runnerOptions?: HookRunnerOptions;
  readonly childRegistry?: ChildRegistry;
  /** Defaults to the allowlist shipped next to this module. */
  readonly mcpAllowlist?: HiveMcpAllowlist;
}

export type HiveHookKey = keyof HookPaths;

export interface HiveHookStatus {
  readonly installed: boolean;
  readonly wired: boolean;
  readonly error?: string;
}

export type HiveHookStatusSnapshot = Readonly<Record<HiveHookKey, HiveHookStatus>>;

export const REQUIRED_HOOK_KEYS: readonly HiveHookKey[] = [
  "bashPolicy",
  "reviewerGuard",
  "postToolHub",
  "flowContext",
];

export const ADVISORY_HOOK_KEYS: readonly HiveHookKey[] = [
  "flowSessionContext",
  "ruleContext",
  "sessionHygieneReport",
];

export const ALL_HOOK_KEYS: readonly HiveHookKey[] = [...REQUIRED_HOOK_KEYS, ...ADVISORY_HOOK_KEYS];

interface MutableHiveHookStatus {
  installed: boolean;
  wired: boolean;
  error?: string;
}

interface HiveHookRuntimeStatus {
  readonly paths: HookPaths;
  readonly statuses: Record<HiveHookKey, MutableHiveHookStatus>;
}

const hookRuntimeStatuses = new WeakMap<object, HiveHookRuntimeStatus>();

function statusForPaths(paths: HookPaths): Record<HiveHookKey, MutableHiveHookStatus> {
  return Object.fromEntries(ALL_HOOK_KEYS.map((key) => [key, { installed: isExecutable(paths[key]), wired: false }])) as Record<HiveHookKey, MutableHiveHookStatus>;
}

function ensureHookRuntimeStatus(pi: ExtensionAPI, paths: HookPaths): HiveHookRuntimeStatus {
  const existing = hookRuntimeStatuses.get(pi as object);
  if (existing) {
    Object.assign(existing.paths, paths);
    for (const key of ALL_HOOK_KEYS) existing.statuses[key].installed = isExecutable(paths[key]);
    return existing;
  }
  const created: HiveHookRuntimeStatus = { paths: { ...paths }, statuses: statusForPaths(paths) };
  hookRuntimeStatuses.set(pi as object, created);
  return created;
}

export function markHiveHookWired(pi: ExtensionAPI, paths: HookPaths, key: HiveHookKey): void {
  const runtime = ensureHookRuntimeStatus(pi, paths);
  runtime.statuses[key].wired = true;
}

export function recordHiveHookError(pi: ExtensionAPI, paths: HookPaths, key: HiveHookKey, error: string | undefined): void {
  const runtime = ensureHookRuntimeStatus(pi, paths);
  runtime.statuses[key].error = error;
}

export function getHiveHookStatuses(pi: ExtensionAPI, paths: HookPaths): HiveHookStatusSnapshot {
  const runtime = ensureHookRuntimeStatus(pi, paths);
  return Object.fromEntries(
    ALL_HOOK_KEYS.map((key) => {
      const status = runtime.statuses[key];
      return [key, {
        installed: status.installed,
        wired: status.wired,
        ...(status.error ? { error: status.error } : {}),
      }];
    }),
  ) as HiveHookStatusSnapshot;
}

function firstExistingDirectory(candidates: readonly string[]): string {
  return candidates.find((candidate) => existsSync(candidate)) ?? candidates[0] ?? process.cwd();
}

export interface HookRootCandidates {
  readonly configuredDir?: string;
  readonly deployedDir: string;
  readonly cwdDir: string;
  readonly sourceDir: string;
}

export function resolveHookRoot(candidates: HookRootCandidates): string {
  return firstExistingDirectory([
    ...(candidates.configuredDir ? [candidates.configuredDir] : []),
    candidates.deployedDir,
    candidates.cwdDir,
    candidates.sourceDir,
  ]);
}

export function defaultHookPaths(): HookPaths {
  const sourceDir = resolve(fileURLToPath(new URL("../../..", import.meta.url)), "global/hooks");
  const deployedDir = resolve(fileURLToPath(new URL("..", import.meta.url)), "global/hooks");
  const configuredDir = process.env.PI_HIVE_HOOK_DIR;
  const root = resolveHookRoot({
    configuredDir,
    deployedDir,
    cwdDir: resolve(process.cwd(), "global/hooks"),
    sourceDir,
  });
  return {
    bashPolicy: join(root, "bash-policy/bash-policy.sh"),
    reviewerGuard: join(root, "reviewer-guard/reviewer-guard.sh"),
    postToolHub: join(root, "post-tool-hub/post-tool-hub.sh"),
    flowContext: join(root, "flow-context/flow-context.sh"),
    flowSessionContext: join(root, "flow-session-context/flow-session-context.sh"),
    ruleContext: join(root, "rule-context/rule-context.sh"),
    sessionHygieneReport: join(root, "session-hygiene-report/session-hygiene-report.sh"),
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

export function requiredHookPaths(paths: HookPaths): readonly string[] {
  return REQUIRED_HOOK_KEYS.map((key) => paths[key]);
}

function hookStatusLines(statuses: HiveHookStatusSnapshot): string {
  return ALL_HOOK_KEYS.map((key) => {
    const status = statuses[key];
    return `${key}: installed=${String(status.installed)} wired=${String(status.wired)} error=${status.error ?? "none"}`;
  }).join("\n");
}

const readinessSchema = Type.Object({});

export function createHookReadinessTool(
  paths: HookPaths,
  statusReader?: () => HiveHookStatusSnapshot,
): ToolDefinition<typeof readinessSchema> {
  return {
    name: "hive_hook_readiness",
    label: "Hive hook readiness",
    description: "Verify that all required Hive PI hook scripts are present and executable.",
    promptSnippet: "Verify Hive hooks are installed before delegating or reviewing",
    parameters: readinessSchema,
    executionMode: "sequential",
    async execute() {
      const required = requiredHookPaths(paths);
      const missing = required.filter((path) => !isExecutable(path));
      const ready = missing.length === 0;
      const hooks = statusReader?.() ?? getStandaloneHookStatuses(paths);
      return {
        content: [{ type: "text" as const, text: ready ? `Hive hooks ready.\n${hookStatusLines(hooks)}` : `Missing or non-executable hooks:\n${missing.join("\n")}\n${hookStatusLines(hooks)}` }],
        details: { ready, missing, hooks },
        isError: !ready,
      };
    },
  };
}

function registerHiveStatusCommand(pi: ExtensionAPI, paths: HookPaths): void {
  pi.registerCommand("hive-status", {
    description: "Inspect Hive hook, research, and active tool readiness",
    handler: async (_args, ctx) => {
      const readiness = checkHookReadiness(requiredHookPaths(paths));
      const hooks = readiness.ready ? "ready" : `not_ready (${readiness.missing.join(", ")})`;
      const activeTools = pi.getActiveTools();
      const tools = [...activeTools].sort().join(", ") || "none";
      ctx.ui.notify(
        `Hive status\nhooks: ${hooks}\nhook_status:\n${hookStatusLines(getHiveHookStatuses(pi, paths))}\n${formatHiveResearchStatus(activeTools)}\ntools: ${tools}`,
        "info",
      );
    },
  });
}

function getStandaloneHookStatuses(paths: HookPaths): HiveHookStatusSnapshot {
  return Object.fromEntries(
    ALL_HOOK_KEYS.map((key) => [key, { installed: isExecutable(paths[key]), wired: false }]),
  ) as HiveHookStatusSnapshot;
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

const INITIAL_CONTEXT_MESSAGE_TYPE = "hive-pi-hook-context";
const ACTIVE_CONTEXT_ENTRY_TYPES = new Set(["message", "custom_message", "compaction", "branch_summary"]);

interface SessionEntryReader {
  buildContextEntries?: () => readonly unknown[];
  getBranch?: () => readonly unknown[];
}

function activeContextEntries(ctx: ExtensionContext): readonly unknown[] {
  const manager = ctx.sessionManager as unknown as SessionEntryReader;
  if (typeof manager.buildContextEntries === "function") return manager.buildContextEntries();
  if (typeof manager.getBranch === "function") return manager.getBranch();
  return [];
}

function hasActiveConversation(ctx: ExtensionContext): boolean {
  return activeContextEntries(ctx).some((entry) => isPlainObject(entry) && typeof entry.type === "string" && ACTIVE_CONTEXT_ENTRY_TYPES.has(entry.type));
}

function initialHookSource(reason: unknown, ctx: ExtensionContext): "startup" | "clear" | undefined {
  if (reason === "new" || reason === "clear") return "clear";
  if (reason !== "startup" || hasActiveConversation(ctx)) return undefined;
  return "startup";
}

function notifyAdvisoryFailure(ctx: ExtensionContext, key: HiveHookKey, reason: string): void {
  const notify = (ctx.ui as unknown as { notify?: (message: string, level?: "info" | "warning" | "error") => void }).notify;
  notify?.(`[Hive hook warning] ${key}: ${reason}`, "warning");
}

/** Unknown active tools (no API) keeps the guidance: only a confirmed absence skips it. */
function mcpToolActive(pi: ExtensionAPI): boolean {
  const getter = (pi as { getActiveTools?: () => readonly string[] }).getActiveTools;
  if (typeof getter !== "function") return true;
  return getter.call(pi).some((tool) => tool.toLowerCase() === "mcp");
}

function warnOnMcpAllowlistFailureOnce(
  ctx: ExtensionContext,
  allowlist: HiveMcpAllowlist,
  sessionId: string,
  warnedSessions: Set<string>,
): void {
  if (allowlist.error === undefined || warnedSessions.has(sessionId)) return;
  warnedSessions.add(sessionId);
  const notify = (ctx.ui as unknown as { notify?: (message: string, level?: "info" | "warning" | "error") => void }).notify;
  notify?.(
    `[Hive MCP warning] allowlist unavailable (${allowlist.error}); every MCP call is refused until it is restored.`,
    "warning",
  );
}

function notifyAdvisoryFailureOnce(
  ctx: ExtensionContext,
  key: HiveHookKey,
  scriptPath: string,
  reason: string,
  sessionId: string,
  failureKeys: Set<string>,
): void {
  const failureKey = `${sessionId}\u0000${scriptPath}`;
  if (failureKeys.has(failureKey)) return;
  failureKeys.add(failureKey);
  notifyAdvisoryFailure(ctx, key, reason);
}

async function runAdvisoryHook(
  pi: ExtensionAPI,
  paths: HookPaths,
  key: HiveHookKey,
  invocation: Parameters<typeof runHook>[0],
  runnerOptions: HookRunnerOptions | undefined,
  failureKeys: Set<string>,
  ctx: ExtensionContext,
): Promise<string | undefined> {
  let result: AdvisoryHookResult;
  try {
    result = await runHook(invocation, runnerOptions);
  } catch (error) {
    const reason = `Advisory hook failed: ${hookFailureReason(error)}`;
    recordHiveHookError(pi, paths, key, reason);
    const firstFailureKey = `${currentSessionId(ctx)}\u0000${paths[key]}`;
    if (!failureKeys.has(firstFailureKey)) notifyAdvisoryFailure(ctx, key, reason);
    return advisoryText({ outcome: "warning", reason }, paths[key], currentSessionId(ctx), failureKeys);
  }
  const session = currentSessionId(ctx);
  const firstFailureKey = `${session}\u0000${paths[key]}`;
  const failureReason = result.reason;
  const failed = (result.outcome === "warning" || result.outcome === "error") && failureReason !== undefined;
  if (failed && failureReason !== undefined) {
    recordHiveHookError(pi, paths, key, failureReason);
    if (!failureKeys.has(firstFailureKey)) notifyAdvisoryFailure(ctx, key, failureReason);
  } else {
    recordHiveHookError(pi, paths, key, undefined);
  }
  return advisoryText(result, paths[key], session, failureKeys);
}

async function sendHiddenContext(
  pi: ExtensionAPI,
  content: string | undefined,
  deliverAs: "steer" | "nextTurn" | undefined,
  details: Record<string, unknown>,
  triggerTurn?: boolean,
): Promise<void> {
  if (!content || content.trim().length === 0) return;
  await pi.sendMessage(
    {
      customType: INITIAL_CONTEXT_MESSAGE_TYPE,
      content,
      display: false,
      details,
    },
    { ...(deliverAs === undefined ? {} : { deliverAs }), ...(triggerTurn === undefined ? {} : { triggerTurn }) },
  );
}

function ruleToolName(toolName: string): "Write" | "Edit" | "Bash" | undefined {
  switch (toolName.toLowerCase()) {
    case "write":
      return "Write";
    case "edit":
      return "Edit";
    case "bash":
      return "Bash";
    default:
      return undefined;
  }
}

function ruleContextPayload(ctx: ExtensionContext, event: ToolCallEvent, toolName: "Write" | "Edit" | "Bash"): Record<string, unknown> {
  return {
    harness: "pi",
    cwd: ctx.cwd,
    session_id: currentSessionId(ctx),
    tool_name: toolName,
    tool_input: event.input,
  };
}

function hookFailureReason(error: unknown): string {
  return error instanceof Error ? error.message : String(error);
}

export function registerGeneralHiveHooks(
  pi: ExtensionAPI,
  options: HiveHookExtensionOptions = {},
): void {
  const paths = mergePaths(options.paths);
  const runnerOptions = options.runnerOptions;
  const childRegistry = options.childRegistry ?? defaultChildRegistry;
  const mcpAllowlist = options.mcpAllowlist ?? HIVE_MCP_ALLOWLIST;
  const guardMcpInput = options.mcpAllowlist ? createHiveMcpGuard(options.mcpAllowlist) : guardHiveMcpInput;
  const mcpAllowlistWarnedSessions = new Set<string>();
  const advisoryFailureKeys = new Set<string>();
  const parentAdvisoriesEnabled = process.env.PI_SUBAGENT_CHILD !== "1";
  let ruleQueue: Promise<void> = Promise.resolve();
  let initialContextQueuedSession: string | undefined;
  let parentSession = process.env.PI_SUBAGENT_PARENT_SESSION ?? "";
  attachChildLifecycle(pi, childRegistry, () => parentSession || process.env.PI_SUBAGENT_PARENT_SESSION || "");
  ensureHookRuntimeStatus(pi, paths);
  markHiveHookWired(pi, paths, "bashPolicy");
  markHiveHookWired(pi, paths, "postToolHub");
  markHiveHookWired(pi, paths, "flowContext");
  if (parentAdvisoriesEnabled) {
    markHiveHookWired(pi, paths, "flowSessionContext");
    markHiveHookWired(pi, paths, "ruleContext");
    markHiveHookWired(pi, paths, "sessionHygieneReport");
  }
  pi.registerTool(createHookReadinessTool(paths, () => getHiveHookStatuses(pi, paths)));
  registerHiveResearchReadiness(pi);
  pi.registerTool(createGitReadTool(pi));
  registerHiveStatusCommand(pi, paths);

  pi.on("session_start", async (event, ctx) => {
    const session = currentSessionId(ctx);
    parentSession = session;
    for (const key of advisoryFailureKeys) {
      if (!key.startsWith(`${session}\u0000`)) advisoryFailureKeys.delete(key);
    }
    if (!process.env.PI_SUBAGENT_PARENT_SESSION) process.env.PI_SUBAGENT_PARENT_SESSION = session;
    process.env.PI_HIVE_SESSION = session;
    if (!parentAdvisoriesEnabled) return;

    const reason = (event as unknown as { reason?: unknown }).reason;
    const source = initialHookSource(reason, ctx);
    if (source === undefined) return;
    if (reason === "startup" && initialContextQueuedSession === session) return;
    const payload = {
      harness: "pi",
      cwd: ctx.cwd,
      session_id: session,
      source,
    };
    const flowContext = await runAdvisoryHook(
      pi,
      paths,
      "flowSessionContext",
      {
        scriptPath: paths.flowSessionContext,
        cwd: ctx.cwd,
        payload,
        mode: "advisory",
        signal: ctx.signal,
      },
      runnerOptions,
      advisoryFailureKeys,
      ctx,
    );
    const hygieneContext = await runAdvisoryHook(
      pi,
      paths,
      "sessionHygieneReport",
      {
        scriptPath: paths.sessionHygieneReport,
        cwd: ctx.cwd,
        payload,
        mode: "advisory",
        env: process.env.HIVE_REPO ? { HIVE_REPO: process.env.HIVE_REPO } : undefined,
        unsetEnv: ["CLAUDECODE"],
        signal: ctx.signal,
      },
      runnerOptions,
      advisoryFailureKeys,
      ctx,
    );
    try {
      await sendHiddenContext(
        pi,
        [flowContext, hygieneContext].filter((value): value is string => value !== undefined).join("\n\n"),
        undefined,
        { source, sessionId: session, hooks: ["flowSessionContext", "sessionHygieneReport"] },
        false,
      );
      initialContextQueuedSession = session;
    } catch (error) {
      const reason = `Unable to queue initial PI hook context: ${hookFailureReason(error)}`;
      recordHiveHookError(pi, paths, "flowSessionContext", reason);
      notifyAdvisoryFailureOnce(ctx, "flowSessionContext", paths.flowSessionContext, reason, session, advisoryFailureKeys);
    }
  });

  pi.on("tool_call", async (event, ctx) => {
    const toolName = event.toolName.toLowerCase();
    if (toolName === "mcpscript" && process.env.PI_SUBAGENT_CHILD === "1") {
      return { block: true, reason: HIVE_MCP_SCRIPT_CHILD_REJECTION, terminate: true };
    }
    if (toolName === "mcp") {
      const guard = guardMcpInput(event.input);
      if (!guard.allowed) {
        return { block: true, reason: guard.reason, terminate: true };
      }
    }
    if (/^(?:subagent|delegate|spawn)(?:[-_]|$)/iu.test(event.toolName)) {
      childRegistry.register(parentSession || currentSessionId(ctx), event.toolCallId);
    }

    if (parentAdvisoriesEnabled) {
      const mappedTool = ruleToolName(event.toolName);
      if (mappedTool) {
        const previous = ruleQueue;
        const current = previous.then(async () => {
          const context = await runAdvisoryHook(
            pi,
            paths,
            "ruleContext",
            {
              scriptPath: paths.ruleContext,
              cwd: ctx.cwd,
              payload: ruleContextPayload(ctx, event, mappedTool),
              mode: "advisory",
              signal: ctx.signal,
            },
            runnerOptions,
            advisoryFailureKeys,
            ctx,
          );
          try {
            await sendHiddenContext(
              pi,
              context,
              "steer",
              { source: "tool_call", sessionId: currentSessionId(ctx), hook: "ruleContext", tool: mappedTool },
            );
          } catch (error) {
            const reason = `Unable to queue rule context: ${hookFailureReason(error)}`;
            recordHiveHookError(pi, paths, "ruleContext", reason);
            notifyAdvisoryFailureOnce(ctx, "ruleContext", paths.ruleContext, reason, currentSessionId(ctx), advisoryFailureKeys);
          }
        }, async () => undefined);
        ruleQueue = current.then(() => undefined, () => undefined);
        await current;
      }
    }

    if (toolName !== "bash") return undefined;
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
      recordHiveHookError(pi, paths, "bashPolicy", result.reason ?? "Blocked by Hive bash policy.");
      return { block: true, reason: result.reason ?? "Blocked by Hive bash policy." };
    }
    recordHiveHookError(pi, paths, "bashPolicy", undefined);
    return undefined;
  });

  pi.on("tool_result", async (event, ctx) => {
    if (/^(?:subagent|delegate|spawn)(?:[-_]|$)/iu.test(event.toolName)) {
      childRegistry.settle(event.toolCallId);
    }
    const context = await runAdvisoryHook(
      pi,
      paths,
      "postToolHub",
      {
        scriptPath: paths.postToolHub,
        cwd: ctx.cwd,
        payload: hookPayload(ctx, event),
        mode: "advisory",
        signal: ctx.signal,
      },
      runnerOptions,
      advisoryFailureKeys,
      ctx,
    );
    if (!context) return undefined;
    return {
      content: [...event.content, { type: "text" as const, text: context }],
      isError: event.isError,
      details: event.details,
    };
  });

  pi.on("before_agent_start", async (event, ctx) => {
    const context = await runAdvisoryHook(
      pi,
      paths,
      "flowContext",
      {
        scriptPath: paths.flowContext,
        cwd: ctx.cwd,
        payload: {
          harness: "pi",
          cwd: ctx.cwd,
          session_id: currentSessionId(ctx),
        },
        mode: "advisory",
        signal: ctx.signal,
      },
      runnerOptions,
      advisoryFailureKeys,
      ctx,
    );
    // The guidance costs context in every session and child, so only roles that
    // can actually call the proxy receive it.
    const mcpActive = mcpToolActive(pi);
    if (mcpActive) {
      warnOnMcpAllowlistFailureOnce(ctx, mcpAllowlist, currentSessionId(ctx), mcpAllowlistWarnedSessions);
    }
    const additions = [context, PI_SKILL_PATH_GUIDANCE, mcpActive ? PI_MCP_GUIDANCE : undefined].filter(
      (value): value is string => value !== undefined,
    );
    return { systemPrompt: `${event.systemPrompt}\n\n${additions.join("\n\n")}` };
  });

  if (parentAdvisoriesEnabled) {
    pi.on("session_compact", async (event, ctx) => {
      const session = currentSessionId(ctx);
      const context = await runAdvisoryHook(
        pi,
        paths,
        "flowSessionContext",
        {
          scriptPath: paths.flowSessionContext,
          cwd: ctx.cwd,
          payload: { harness: "pi", cwd: ctx.cwd, session_id: session, source: "compact" },
          mode: "advisory",
          signal: ctx.signal,
        },
        runnerOptions,
        advisoryFailureKeys,
        ctx,
      );
      try {
        await sendHiddenContext(
          pi,
          context,
          event.willRetry ? "steer" : "nextTurn",
          { source: "compact", sessionId: session, hook: "flowSessionContext" },
        );
      } catch (error) {
        const reason = `Unable to queue compact context: ${hookFailureReason(error)}`;
        recordHiveHookError(pi, paths, "flowSessionContext", reason);
        notifyAdvisoryFailureOnce(ctx, "flowSessionContext", paths.flowSessionContext, reason, session, advisoryFailureKeys);
      }
    });
    pi.on("session_compact_failed", async () => undefined);
  }
}

export function createGeneralHiveHooksExtension(options: HiveHookExtensionOptions = {}) {
  return (pi: ExtensionAPI): void => registerGeneralHiveHooks(pi, options);
}

export default createGeneralHiveHooksExtension();
