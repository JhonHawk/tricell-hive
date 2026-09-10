import { createHash } from "node:crypto";
import { closeSync, constants, fstatSync, lstatSync, openSync, readSync, type Stats } from "node:fs";
import { isAbsolute, resolve } from "node:path";
import type { ExtensionAPI, ExtensionCommandContext, ExtensionContext, ToolCallEvent } from "@earendil-works/pi-coding-agent";
import { checkHookReadiness, runHook } from "./hook-runner.ts";
import { createGitReadTool } from "./git-read.ts";
import { defaultHookPaths, guardHiveMcpInput, type HiveHookExtensionOptions } from "./hooks.ts";
import { attachChildLifecycle, defaultChildRegistry, reconcileActiveChildren } from "./child-registry.ts";
import type { ChildRegistry, HookPaths, PlanCaptureResult, PlanModeState } from "./types.ts";

const STATE_ENTRY = "hive-plan-state";
const PLAN_ENV = "PI_HIVE_PLAN_MODE";
const PLAN_ALLOWED_TOOLS = new Set([
  "read",
  "grep",
  "find",
  "ls",
  "hive_git_read",
  "hive_hook_readiness",
  "hive_reviewer_readiness",
  "subagent",
  "bg_wait",
  "contact_supervisor",
  "mem_search",
  "mem_context",
  "mem_get_observation",
  "ask_user_question",
  "web_search",
  "source_check",
  "fetch_content",
  "get_search_content",
  "mcp",
]);
const MAX_PLAN_BYTES = 2_000_000;

interface MutablePlanState {
  enabled: boolean;
  executionStarted: boolean;
  toolsBeforePlanMode?: string[];
  candidate?: string;
  candidateAfterEntryId?: string;
  candidateSha256?: string;
  capturePath?: string;
  lastCaptureStatus?: PlanModeState["lastCaptureStatus"];
}

interface JsonRecord {
  readonly [key: string]: unknown;
}

function isRecord(value: unknown): value is JsonRecord {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function stringArray(value: unknown): string[] | undefined {
  if (!Array.isArray(value)) return undefined;
  const items = value.filter((item): item is string => typeof item === "string");
  return items.length === value.length ? items : undefined;
}

function parsePersistedState(value: unknown): MutablePlanState | undefined {
  if (!isRecord(value) || typeof value.enabled !== "boolean" || typeof value.executionStarted !== "boolean") return undefined;
  const tools = stringArray(value.toolsBeforePlanMode);
  return {
    enabled: value.enabled,
    executionStarted: value.executionStarted,
    ...(tools ? { toolsBeforePlanMode: tools } : {}),
    ...(typeof value.candidate === "string" ? { candidate: value.candidate } : {}),
    ...(typeof value.candidateAfterEntryId === "string" ? { candidateAfterEntryId: value.candidateAfterEntryId } : {}),
    ...(typeof value.candidateSha256 === "string" ? { candidateSha256: value.candidateSha256 } : {}),
    ...(typeof value.capturePath === "string" ? { capturePath: value.capturePath } : {}),
    ...(value.lastCaptureStatus === "captured" || value.lastCaptureStatus === "session_only" || value.lastCaptureStatus === "skipped"
      ? { lastCaptureStatus: value.lastCaptureStatus }
      : {}),
  };
}

function isAssistantMessage(value: unknown): value is { role: "assistant"; content: unknown[] } {
  return isRecord(value) && value.role === "assistant" && Array.isArray(value.content);
}

function messageText(value: unknown): string | undefined {
  if (!isAssistantMessage(value)) return undefined;
  const parts = value.content
    .filter(isRecord)
    .filter((part) => part.type === "text" && typeof part.text === "string")
    .map((part) => String(part.text));
  return parts.length > 0 ? parts.join("") : undefined;
}

function planSha256(plan: string): string {
  return createHash("sha256").update(Buffer.from(plan, "utf8")).digest("hex");
}

function latestAssistantText(ctx: ExtensionContext, markerId: string | undefined): string | undefined {
  const entries = ctx.sessionManager.getBranch();
  for (let index = entries.length - 1; index >= 0; index -= 1) {
    const entry = entries[index];
    if (markerId && entry?.id === markerId) return undefined;
    if (entry?.type !== "message" || !("message" in entry)) continue;
    const text = messageText(entry.message);
    if (text !== undefined) return text;
  }
  return undefined;
}

function currentSession(ctx: ExtensionContext): string {
  return ctx.sessionManager.getSessionId() || ctx.sessionManager.getSessionFile() || "pi-session";
}

function isPlanToolAllowed(toolName: string): boolean {
  return PLAN_ALLOWED_TOOLS.has(toolName.toLowerCase());
}

function planToolSet(activeTools: readonly string[]): string[] {
  const allowed = activeTools.filter((name) => isPlanToolAllowed(name));
  for (const tool of ["hive_git_read", "hive_hook_readiness"]) {
    if (!allowed.some((name) => name.toLowerCase() === tool)) allowed.push(tool);
  }
  return [...new Set(allowed)];
}

function notifyStatus(ctx: ExtensionContext, state: MutablePlanState): void {
  if (state.enabled) ctx.ui.setStatus("hive-plan", "⏸ plan");
  else if (state.executionStarted) ctx.ui.setStatus("hive-plan", "▶ plan");
  else ctx.ui.setStatus("hive-plan", undefined);
}

function parseCommand(args: string): { readonly command: string; readonly path?: string } {
  const tokens = args.trim().split(/\s+/u).filter(Boolean);
  return { command: tokens[0] ?? "status", ...(tokens[1] ? { path: tokens.slice(1).join(" ") } : {}) };
}

function parseCaptureOutput(stdout: string): PlanCaptureResult | undefined {
  const lines = stdout.trim().split(/\r?\n/u).filter(Boolean);
  const candidates = [stdout.trim(), ...lines];
  for (const candidate of candidates) {
    try {
      const parsed: unknown = JSON.parse(candidate) as unknown;
      if (!isRecord(parsed) || typeof parsed.status !== "string") continue;
      if (
        parsed.status === "captured" &&
        typeof parsed.path === "string" &&
        isAbsolute(parsed.path) &&
        typeof parsed.sha256 === "string" &&
        /^[a-f0-9]{64}$/u.test(parsed.sha256)
      ) {
        return { status: "captured", path: parsed.path, sha256: parsed.sha256 };
      }
      if (parsed.status === "session_only" && typeof parsed.reason === "string" && parsed.reason.length > 0) {
        return { status: "session_only", reason: parsed.reason };
      }
      if (parsed.status === "skipped" && typeof parsed.reason === "string" && parsed.reason.length > 0) {
        return { status: "skipped", reason: parsed.reason };
      }
    } catch {
      // Continue until a complete JSON response is found.
    }
  }
  return undefined;
}

function stateFromEntries(ctx: ExtensionContext): MutablePlanState | undefined {
  const entries = ctx.sessionManager.getEntries();
  for (let index = entries.length - 1; index >= 0; index -= 1) {
    const entry = entries[index];
    if (entry?.type !== "custom" || !("customType" in entry) || entry.customType !== STATE_ENTRY) continue;
    return parsePersistedState("data" in entry ? entry.data : undefined);
  }
  return undefined;
}

function persistState(pi: ExtensionAPI, state: MutablePlanState): void {
  pi.appendEntry(STATE_ENTRY, {
    enabled: state.enabled,
    executionStarted: state.executionStarted,
    toolsBeforePlanMode: state.toolsBeforePlanMode,
    candidate: state.candidate,
    candidateAfterEntryId: state.candidateAfterEntryId,
    candidateSha256: state.candidateSha256,
    capturePath: state.capturePath,
    lastCaptureStatus: state.lastCaptureStatus,
  });
}

export async function readPlanFile(path: string, cwd: string): Promise<string> {
  const absolute = resolve(cwd, path);
  let linkStat: Stats;
  try {
    linkStat = lstatSync(absolute);
  } catch (error) {
    throw new Error(`Plan file does not exist: ${absolute}`);
  }
  if (linkStat.isSymbolicLink()) throw new Error(`Plan file must not be a symbolic link: ${absolute}`);
  if (!linkStat.isFile()) throw new Error(`Plan path must be a regular file: ${absolute}`);

  const noFollow = typeof constants.O_NOFOLLOW === "number" ? constants.O_NOFOLLOW : 0;
  let descriptor: number | undefined;
  try {
    descriptor = openSync(absolute, constants.O_RDONLY | noFollow);
    const stat = fstatSync(descriptor);
    if (!stat.isFile()) throw new Error(`Plan path must be a regular file: ${absolute}`);
    if (stat.size > MAX_PLAN_BYTES) throw new Error(`Plan file exceeds ${String(MAX_PLAN_BYTES)} bytes.`);
    const buffer = Buffer.alloc(MAX_PLAN_BYTES + 1);
    let offset = 0;
    while (offset < buffer.byteLength) {
      const read = readSync(descriptor, buffer, offset, buffer.byteLength - offset, null);
      if (read === 0) break;
      offset += read;
    }
    if (offset > MAX_PLAN_BYTES || fstatSync(descriptor).size > MAX_PLAN_BYTES) {
      throw new Error(`Plan file exceeds ${String(MAX_PLAN_BYTES)} bytes.`);
    }
    return buffer.subarray(0, offset).toString("utf8");
  } finally {
    if (descriptor !== undefined) closeSync(descriptor);
  }
}

async function capturePlan(
  paths: HookPaths,
  cwd: string,
  plan: string,
  signal: AbortSignal | undefined,
  runnerOptions: HiveHookExtensionOptions["runnerOptions"],
): Promise<PlanCaptureResult> {
  const result = await runHook(
    {
      scriptPath: paths.flowPlanCapture,
      args: ["--from-pi-command"],
      cwd,
      payload: { harness: "pi", cwd, tool_response: { plan } },
      mode: "blocking",
      outputKind: "capture",
      signal,
    },
    runnerOptions,
  );
  if (result.outcome === "block") throw new Error(result.reason ?? "Plan capture hook failed.");
  const captured = parseCaptureOutput(result.stdout);
  if (!captured) throw new Error("Plan capture hook returned invalid or missing JSON.");
  return captured;
}

export interface HivePlanExtensionOptions extends HiveHookExtensionOptions {
  readonly childRegistry?: ChildRegistry;
}

export function registerHivePlanMode(pi: ExtensionAPI, options: HivePlanExtensionOptions = {}): void {
  const paths = { ...defaultHookPaths(), ...options.paths };
  const childRegistry = options.childRegistry ?? defaultChildRegistry;
  let parentSession = process.env.PI_SUBAGENT_PARENT_SESSION ?? "";
  let inheritedChildPlan = false;
  let activeSession = "";
  let processChild = false;
  let initialized = false;
  let approvalInFlight = false;
  attachChildLifecycle(pi, childRegistry, () => parentSession || process.env.PI_SUBAGENT_PARENT_SESSION || "");
  let state: MutablePlanState = { enabled: false, executionStarted: false };

  const applyPlanTools = (): void => {
    if (state.toolsBeforePlanMode === undefined) state.toolsBeforePlanMode = pi.getActiveTools();
    pi.setActiveTools(planToolSet(state.toolsBeforePlanMode));
  };

  const restoreTools = (): void => {
    if (state.toolsBeforePlanMode) pi.setActiveTools(state.toolsBeforePlanMode);
    state.toolsBeforePlanMode = undefined;
  };

  const enter = async (ctx: ExtensionCommandContext): Promise<void> => {
    if (!ctx.isIdle()) {
      ctx.ui.notify("Cannot enter plan mode while the agent is active.", "warning");
      return;
    }
    const session = currentSession(ctx);
    const entryState = state;
    if (childRegistry.hasActiveChildren(session)) {
      ctx.ui.notify("Cannot enter plan mode while child agents are still active.", "warning");
      return;
    }
    await reconcileActiveChildren(pi, childRegistry, session);
    if (!ctx.isIdle() || currentSession(ctx) !== session || state !== entryState) {
      ctx.ui.notify("Plan mode entry was cancelled because the session or agent state changed.", "warning");
      return;
    }
    if (childRegistry.hasActiveChildren(session)) {
      ctx.ui.notify("Cannot enter plan mode while child agents are still active.", "warning");
      return;
    }
    if (state.enabled) {
      ctx.ui.notify("Plan mode is already active.", "info");
      return;
    }
    state = {
      enabled: true,
      executionStarted: false,
      candidateAfterEntryId: ctx.sessionManager.getBranch().at(-1)?.id ?? "",
    };
    applyPlanTools();
    process.env[PLAN_ENV] = "1";
    persistState(pi, state);
    notifyStatus(ctx, state);
    ctx.ui.notify("Hive plan mode enabled: read/search tools only.", "info");
  };

  const cancel = (ctx: ExtensionCommandContext): void => {
    if (inheritedChildPlan) {
      ctx.ui.notify("Child agents keep inherited plan restrictions for their lifetime.", "warning");
      return;
    }
    if (!state.enabled) {
      ctx.ui.notify("Plan mode is not active.", "info");
      return;
    }
    restoreTools();
    state = { enabled: false, executionStarted: false };
    delete process.env[PLAN_ENV];
    persistState(pi, state);
    notifyStatus(ctx, state);
    ctx.ui.notify("Hive plan mode cancelled; tools restored.", "info");
  };

  const status = (ctx: ExtensionCommandContext): void => {
    const mode = state.enabled ? "plan" : state.executionStarted ? "execution" : "normal";
    const candidate = state.candidate
      ? `${String(Buffer.byteLength(state.candidate, "utf8"))} bytes${state.candidateSha256 ? ` (sha256: ${state.candidateSha256})` : ""}`
      : "none";
    const capture = state.capturePath ?? state.lastCaptureStatus ?? "none";
    const readiness = checkHookReadiness(Object.values(paths));
    const hooks = readiness.ready ? "ready" : `not_ready (${readiness.missing.join(", ")})`;
    const tools = [...pi.getActiveTools()].sort().join(", ") || "none";
    ctx.ui.notify(
      `Hive plan status\nmode: ${mode}\ncandidate: ${candidate}\ncapture: ${capture}\nhooks: ${hooks}\ntools: ${tools}`,
      "info",
    );
  };

  const approve = async (path: string | undefined, ctx: ExtensionCommandContext): Promise<void> => {
    if (approvalInFlight) {
      ctx.ui.notify("Plan approval is already in progress.", "warning");
      return;
    }
    if (inheritedChildPlan) {
      ctx.ui.notify("Child agents cannot approve or execute an inherited plan.", "warning");
      return;
    }
    if (!state.enabled) {
      ctx.ui.notify("Plan approval requires active plan mode.", "warning");
      return;
    }
    if (!ctx.isIdle()) {
      ctx.ui.notify("Cannot approve while the agent is active.", "warning");
      return;
    }
    try {
      approvalInFlight = true;
      const approvalState = state;
      const approvalSession = currentSession(ctx);
      let plan: string;
      try {
        plan = path ? await readPlanFile(path, ctx.cwd) : state.candidate ?? "";
      } catch (error) {
        ctx.ui.notify(error instanceof Error ? error.message : String(error), "error");
        return;
      }
      if (plan.length === 0) {
        ctx.ui.notify("No plan candidate is available. Provide /hive-plan approve <path>.", "warning");
        return;
      }
      if (Buffer.byteLength(plan, "utf8") > MAX_PLAN_BYTES) {
        ctx.ui.notify(`Plan exceeds ${String(MAX_PLAN_BYTES)} bytes.`, "error");
        return;
      }

      let capture: PlanCaptureResult;
      try {
        capture = await capturePlan(paths, ctx.cwd, plan, ctx.signal, options.runnerOptions);
      } catch (error) {
        ctx.ui.notify(error instanceof Error ? error.message : String(error), "error");
        return;
      }

      if (state !== approvalState || !state.enabled || currentSession(ctx) !== approvalSession) {
        ctx.ui.notify("Plan approval was cancelled or the session changed; execution remains blocked.", "warning");
        return;
      }
      const candidateSha256 = planSha256(plan);
      if (capture.status === "captured" && capture.sha256 !== candidateSha256) {
        ctx.ui.notify("Plan capture hash does not match the approved plan; execution remains blocked.", "error");
        return;
      }
      restoreTools();
      state = {
        ...state,
        enabled: false,
        executionStarted: true,
        candidate: plan,
        candidateSha256,
        ...(capture.status === "captured" ? { capturePath: capture.path } : {}),
        lastCaptureStatus: capture.status,
      };
      delete process.env[PLAN_ENV];
      persistState(pi, state);
      notifyStatus(ctx, state);
      const suffix = capture.status === "captured" ? ` (${capture.path})` : ` (${capture.reason})`;
      ctx.ui.notify(`Approved plan captured${suffix}; starting execution.`, "info");
      pi.sendUserMessage(`Execute the approved plan exactly as captured below.\n\n${plan}`, {
        deliverAs: "followUp",
        expandPromptTemplates: false,
      });
    } finally {
      approvalInFlight = false;
    }
  };

  pi.registerCommand("hive-plan", {
    description: "Enter, approve, cancel or inspect Hive plan mode",
    handler: async (args, ctx) => {
      const parsed = parseCommand(args);
      switch (parsed.command) {
        case "enter":
          await enter(ctx);
          break;
        case "approve":
          await approve(parsed.path, ctx);
          break;
        case "cancel":
          cancel(ctx);
          break;
        case "status":
        case "":
          status(ctx);
          break;
        default:
          ctx.ui.notify("Usage: /hive-plan enter | approve [path] | cancel | status", "warning");
      }
    },
  });
  pi.registerCommand("hive-status", {
    description: "Inspect Hive plan mode, hook readiness and active tools",
    handler: async (_args, ctx) => status(ctx),
  });

  pi.registerTool(createGitReadTool(pi));

  pi.on("session_start", async (event, ctx) => {
    const session = currentSession(ctx);
    const inheritedParent = process.env.PI_SUBAGENT_PARENT_SESSION;
    if (!initialized) {
      const nativeNewRoot = (event.reason === "new" || event.reason === "fork") && process.env.PI_SUBAGENT_CHILD !== "1";
      processChild = process.env.PI_SUBAGENT_CHILD === "1"
        || (!nativeNewRoot && process.env[PLAN_ENV] === "1" && inheritedParent !== undefined && inheritedParent !== session);
      initialized = true;
    }
    if (activeSession !== "" && activeSession !== session) {
      if (state.enabled && state.toolsBeforePlanMode) pi.setActiveTools(state.toolsBeforePlanMode);
      delete process.env[PLAN_ENV];
    }
    activeSession = session;
    parentSession = session;
    inheritedChildPlan = processChild
      && process.env[PLAN_ENV] === "1"
      && (process.env.PI_SUBAGENT_CHILD === "1" || (inheritedParent !== undefined && inheritedParent !== session));
    if (!processChild) process.env.PI_SUBAGENT_PARENT_SESSION = parentSession;
    const persisted = stateFromEntries(ctx);
    state = persisted ?? { enabled: false, executionStarted: false };
    if (!processChild && (event.reason === "new" || event.reason === "fork") && persisted?.enabled !== true) {
      delete process.env[PLAN_ENV];
    }
    if (process.env[PLAN_ENV] === "1" && !state.enabled && !state.executionStarted) {
      state.enabled = true;
      state.toolsBeforePlanMode = pi.getActiveTools();
      state.candidateAfterEntryId = ctx.sessionManager.getBranch().at(-1)?.id ?? "";
    }
    if (state.enabled) {
      process.env[PLAN_ENV] = "1";
      applyPlanTools();
    }
    await reconcileActiveChildren(pi, childRegistry, parentSession);
    notifyStatus(ctx, state);
  });

  pi.on("agent_settled", async (_event, ctx) => {
    if (!state.enabled || state.executionStarted) return;
    const candidate = latestAssistantText(ctx, state.candidateAfterEntryId);
    if (candidate !== undefined) {
      const candidateSha256 = planSha256(candidate);
      if (state.candidate === candidate && state.candidateSha256 === candidateSha256) return;
      state.candidate = candidate;
      state.candidateSha256 = candidateSha256;
      persistState(pi, state);
      ctx.ui.setStatus("hive-plan", "⏸ plan · candidate ready");
    }
  });

  pi.on("tool_call", async (event: ToolCallEvent) => {
    if (!state.enabled) return undefined;
    if (event.toolName.toLowerCase() === "mcp") {
      const guard = guardHiveMcpInput(event.input);
      if (!guard.allowed) {
        return { block: true, reason: guard.reason, terminate: true };
      }
      return undefined;
    }
    if (isPlanToolAllowed(event.toolName)) return undefined;
    return {
      block: true,
      reason: `Hive plan mode: '${event.toolName}' is blocked until the plan is explicitly approved.`,
      terminate: true,
    };
  });

  pi.on("context", async (event) => {
    if (!state.enabled) return undefined;
    return {
      messages: event.messages,
    };
  });
}

export function createHivePlanModeExtension(options: HivePlanExtensionOptions = {}) {
  return (pi: ExtensionAPI): void => registerHivePlanMode(pi, options);
}

export default createHivePlanModeExtension();
