import { spawn as nodeSpawn, type ChildProcessWithoutNullStreams } from "node:child_process";
import { constants, existsSync, statSync } from "node:fs";
import {
  HOOK_COMBINED_OUTPUT_LIMIT_BYTES,
  HOOK_OUTPUT_LIMIT_BYTES,
  HOOK_TIMEOUT_MS,
  type HookInvocation,
  type HookOutputKind,
  type HookResult,
  type HookRunnerOptions,
  type HookSpawn,
} from "./types.ts";

interface JsonRecord {
  readonly [key: string]: unknown;
}

interface ParsedHookOutput {
  readonly recognized: boolean;
  readonly blockingValid: boolean;
  readonly decision?: string;
  readonly status?: string;
  readonly reason?: string;
  readonly additionalContext?: string;
}

const KNOWN_DECISIONS = new Set(["allow", "deny", "blocked", "ask"]);
const KNOWN_STATUSES = new Set(["captured", "session_only", "skipped"]);

function isRecord(value: unknown): value is JsonRecord {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function textValue(value: unknown): string | undefined {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

function parseJsonRecord(text: string, outputKind: HookOutputKind = "policy"): ParsedHookOutput | undefined {
  const trimmed = text.trim();
  if (trimmed.length === 0) return undefined;

  let value: unknown;
  try {
    value = JSON.parse(trimmed) as unknown;
  } catch {
    return undefined;
  }

  const record = isRecord(value) ? value : undefined;
  if (!record) return undefined;

  const nested = isRecord(record.hookSpecificOutput) ? record.hookSpecificOutput : undefined;
  const rawDecision = textValue(record.decision) ?? textValue(record.permissionDecision) ?? textValue(nested?.permissionDecision);
  const decision = rawDecision?.toLowerCase();
  const status = textValue(record.status);
  const reason = textValue(record.reason) ?? textValue(nested?.reason);
  const additionalContext = textValue(record.additionalContext) ?? textValue(nested?.additionalContext);
  const normalizedDecision = decision?.toLowerCase();
  const validDecision = normalizedDecision !== undefined && KNOWN_DECISIONS.has(normalizedDecision);
  const validStatus = outputKind === "capture" && status !== undefined && KNOWN_STATUSES.has(status);
  const recognized = validDecision || validStatus || reason !== undefined || additionalContext !== undefined || nested !== undefined;
  return { recognized, blockingValid: validDecision || validStatus, decision, status, reason, additionalContext };
}

function signalProcessTree(child: ChildProcessWithoutNullStreams, signal: NodeJS.Signals): void {
  if (process.platform !== "win32" && child.pid !== undefined) {
    try {
      process.kill(-child.pid, signal);
      return;
    } catch {
      // Fall back to the direct child when a process group disappeared first.
    }
  }
  if (child.exitCode !== null || child.signalCode !== null) return;
  child.kill(signal);
}

function appendBounded(current: string, chunk: string, limit: number): { readonly value: string; readonly overflow: boolean } {
  const available = limit - Buffer.byteLength(current, "utf8");
  if (available <= 0) return { value: current, overflow: true };
  const chunkBytes = Buffer.byteLength(chunk, "utf8");
  if (chunkBytes <= available) return { value: current + chunk, overflow: false };
  return {
    value: current + Buffer.from(chunk, "utf8").subarray(0, available).toString("utf8"),
    overflow: true,
  };
}

function normalizeSpawn(spawnProcess: HookSpawn | undefined): HookSpawn {
  return spawnProcess ?? ((command, args, options) => nodeSpawn(command, args, options));
}

/**
 * Execute a canonical Hive hook without a shell and classify its result.
 * Blocking callers fail closed for timeouts, invalid output and non-zero exits.
 */
export function runHook(invocation: HookInvocation, options: HookRunnerOptions = {}): Promise<HookResult> {
  const timeoutMs = invocation.timeoutMs ?? HOOK_TIMEOUT_MS;
  const spawnProcess = normalizeSpawn(options.spawnProcess);

  if (!existsSync(invocation.scriptPath)) {
    const reason = `Required hook is missing: ${invocation.scriptPath}`;
    return Promise.resolve({
      outcome: invocation.mode === "blocking" ? "block" : "warning",
      reason,
      stdout: "",
      stderr: reason,
      exitCode: null,
      timedOut: false,
    });
  }

  return new Promise<HookResult>((resolve) => {
    let child: ChildProcessWithoutNullStreams;
    try {
      const env: NodeJS.ProcessEnv = {
        ...process.env,
        ...options.env,
        ...invocation.env,
      };
      for (const name of invocation.unsetEnv ?? []) delete env[name];
      child = spawnProcess(invocation.scriptPath, invocation.args ?? [], {
        cwd: invocation.cwd,
        env,
        stdio: ["pipe", "pipe", "pipe"],
        shell: false,
        detached: process.platform !== "win32",
      });
    } catch (error) {
      const reason = `Hook failed to start: ${error instanceof Error ? error.message : String(error)}`;
      resolve({
        outcome: invocation.mode === "blocking" ? "block" : "warning",
        reason,
        stdout: "",
        stderr: reason,
        exitCode: null,
        timedOut: false,
      });
      return;
    }

    let stdout = "";
    let stderr = "";
    let timedOut = false;
    let outputExceeded = false;
    let totalOutputBytes = 0;
    let settled = false;
    let timeoutHandle: NodeJS.Timeout | undefined;
    let forceKillHandle: NodeJS.Timeout | undefined;
    let abortListener: (() => void) | undefined;

    const finish = (exitCode: number | null): void => {
      if (settled) return;
      settled = true;
      if (timeoutHandle) clearTimeout(timeoutHandle);
      if (forceKillHandle) clearTimeout(forceKillHandle);
      if (abortListener && invocation.signal) invocation.signal.removeEventListener("abort", abortListener);

      const parsed = parseJsonRecord(stdout, invocation.outputKind);
      const explicitDeny = parsed?.decision === "deny" || parsed?.decision === "blocked";
      const confirmationRequired = invocation.mode === "blocking" && invocation.outputKind !== "capture" && parsed?.decision === "ask";
      const nonZero = exitCode !== 0;
      const invalidOutput = stdout.trim().length > 0 &&
        (parsed === undefined || !parsed.recognized || (invocation.mode === "blocking" && !parsed.blockingValid));
      const failed = timedOut || outputExceeded || nonZero || invalidOutput || explicitDeny || confirmationRequired;
      const reason = outputExceeded
        ? `Hook output exceeded ${String(HOOK_COMBINED_OUTPUT_LIMIT_BYTES)} bytes`
        : (parsed?.reason ?? stderr.trim()) ||
          (timedOut ? `Hook timed out after ${timeoutMs}ms` : undefined) ||
          (invalidOutput ? "Hook returned invalid JSON" : undefined) ||
          (nonZero ? `Hook exited with code ${String(exitCode)}` : undefined) ||
          (explicitDeny ? "Hook denied the operation" : undefined);
      const advisoryText = invocation.mode === "advisory" && invalidOutput ? stdout.trim() : undefined;

      resolve({
        outcome: failed ? (invocation.mode === "blocking" ? "block" : "warning") : parsed?.additionalContext ? "warning" : "allow",
        reason,
        additionalContext: parsed?.additionalContext ?? advisoryText,
        stdout,
        stderr,
        exitCode,
        timedOut,
      });
    };

    function terminate(kind: "timeout" | "output"): void {
      if (timedOut || outputExceeded) return;
      if (kind === "timeout") timedOut = true;
      else outputExceeded = true;
      signalProcessTree(child, "SIGTERM");
      forceKillHandle = setTimeout(() => {
        if (!settled) signalProcessTree(child, "SIGKILL");
      }, 250);
    }

    child.stdout.on("data", (chunk: Buffer | string) => {
      const text = chunk.toString();
      totalOutputBytes += Buffer.byteLength(text, "utf8");
      const bounded = appendBounded(stdout, text, HOOK_OUTPUT_LIMIT_BYTES);
      stdout = bounded.value;
      if (bounded.overflow || totalOutputBytes > HOOK_COMBINED_OUTPUT_LIMIT_BYTES) terminate("output");
    });
    child.stderr.on("data", (chunk: Buffer | string) => {
      const text = chunk.toString();
      totalOutputBytes += Buffer.byteLength(text, "utf8");
      const bounded = appendBounded(stderr, text, HOOK_OUTPUT_LIMIT_BYTES);
      stderr = bounded.value;
      if (bounded.overflow || totalOutputBytes > HOOK_COMBINED_OUTPUT_LIMIT_BYTES) terminate("output");
    });
    child.stdin.on("error", (error) => {
      if (settled) return;
      stderr = appendBounded(stderr, `${error.message}\n`, HOOK_OUTPUT_LIMIT_BYTES).value;
    });
    child.once("error", (error) => {
      stderr = appendBounded(stderr, `${error.message}\n`, HOOK_OUTPUT_LIMIT_BYTES).value;
      finish(child.exitCode);
    });
    child.once("close", (code) => finish(code));

    timeoutHandle = setTimeout(() => terminate("timeout"), timeoutMs);

    if (invocation.signal) {
      abortListener = () => {
        terminate("timeout");
      };
      if (invocation.signal.aborted) abortListener();
      else invocation.signal.addEventListener("abort", abortListener, { once: true });
    }

    child.stdin.end(JSON.stringify(invocation.payload));
  });
}

/** Execute an advisory hook whose output is plain context instead of JSON. */
export function runAdvisoryTextHook(invocation: Omit<HookInvocation, "mode">, options: HookRunnerOptions = {}): Promise<HookResult> {
  return runHook({ ...invocation, mode: "advisory" }, options);
}

/** Return a stable readiness result for the required hook scripts. */
export function checkHookReadiness(paths: readonly string[]): { ready: boolean; missing: readonly string[] } {
  const missing = paths.filter((path) => {
    try {
      const stats = statSync(path);
      return !stats.isFile() || (stats.mode & constants.S_IXUSR) === 0;
    } catch {
      return true;
    }
  });
  return { ready: missing.length === 0, missing };
}

export function parseHookJson(text: string): JsonRecord | undefined {
  const parsed = parseJsonRecord(text, "policy");
  return parsed ? { ...parsed } : undefined;
}
