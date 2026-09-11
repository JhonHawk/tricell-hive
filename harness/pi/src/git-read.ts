import { spawn, type ChildProcessWithoutNullStreams } from "node:child_process";
import { Type, type Static } from "typebox";
import type { ExtensionAPI, ExtensionContext, ToolDefinition } from "@earendil-works/pi-coding-agent";
import type { GitReadOperation, GitReadRequest } from "./types.ts";

const MAX_LOG_ENTRIES = 200;
const MAX_OUTPUT_BYTES = 200_000;
const MAX_COMBINED_OUTPUT_BYTES = MAX_OUTPUT_BYTES * 2;
const GIT_TIMEOUT_MS = 10_000;
const TERMINATION_GRACE_MS = 250;

const OPERATION_VALUES: GitReadOperation[] = ["status", "diff", "log", "show", "blame", "ls-files", "branch", "remote"];

export const gitReadSchema = Type.Object({
  operation: Type.Union(OPERATION_VALUES.map((operation) => Type.Literal(operation))),
  path: Type.Optional(Type.String()),
  revision: Type.Optional(Type.String()),
  limit: Type.Optional(Type.Integer({ minimum: 1, maximum: MAX_LOG_ENTRIES })),
  staged: Type.Optional(Type.Boolean()),
  line: Type.Optional(Type.String()),
});

export type GitReadParams = Static<typeof gitReadSchema>;

function isSafeArgument(value: string): boolean {
  return value.length > 0 && value.length <= 4_096 && !/[\u0000-\u001f\u007f]/u.test(value) && !value.startsWith("-");
}

function assertSafePath(path: string | undefined): void {
  if (path !== undefined && !isSafeArgument(path)) throw new Error("Git path must be non-empty, non-control text and cannot start with '-'.");
}

function assertSafeRevision(revision: string | undefined): void {
  if (revision !== undefined && (!isSafeArgument(revision) || !/^[A-Za-z0-9][A-Za-z0-9._/@^~:+-]*$/u.test(revision))) {
    throw new Error("Git revision contains unsupported characters.");
  }
}

function assertSafeLine(line: string | undefined): void {
  if (line !== undefined && !/^\d+(?:,\d+)?$/u.test(line)) throw new Error("Git blame line must be N or N,M.");
}

function commonGitArguments(): string[] {
  return ["--no-pager", "-c", "core.fsmonitor=false", "-c", "core.pager=cat"];
}

export function buildGitArguments(request: GitReadRequest): string[] {
  assertSafePath(request.path);
  assertSafeRevision(request.revision);
  assertSafeLine(request.line);

  const args = commonGitArguments();
  switch (request.operation) {
    case "status":
      args.push("status", "--short", "--branch");
      break;
    case "diff":
      args.push("diff", "--no-ext-diff", "--no-textconv");
      if (request.staged) args.push("--cached");
      args.push("--");
      if (request.path) args.push(request.path);
      break;
    case "log": {
      const limit = Math.min(Math.max(request.limit ?? 20, 1), MAX_LOG_ENTRIES);
      args.push("log", "--oneline", `-${String(limit)}`);
      if (request.path) args.push("--", request.path);
      break;
    }
    case "show":
      args.push("show", "--no-ext-diff", "--no-textconv", request.revision ?? "HEAD");
      if (request.path) args.push("--", request.path);
      break;
    case "blame":
      args.push("blame");
      if (request.line) args.push("-L", request.line);
      args.push("--", request.path ?? ".");
      break;
    case "ls-files":
      args.push("ls-files", "--", request.path ?? ".");
      break;
    case "branch":
      args.push("branch", "--list", "--no-color");
      break;
    case "remote":
      args.push("remote");
      break;
  }
  return args;
}

function truncateOutput(output: string): string {
  const bytes = Buffer.byteLength(output, "utf8");
  if (bytes <= MAX_OUTPUT_BYTES) return output;
  return `${Buffer.from(output, "utf8").subarray(0, MAX_OUTPUT_BYTES).toString("utf8")}\n\n[Output truncated at ${String(MAX_OUTPUT_BYTES)} bytes]`;
}

interface GitProcessResult {
  readonly stdout: string;
  readonly stderr: string;
  readonly code: number | null;
  readonly killed: boolean;
  readonly timedOut: boolean;
  readonly outputLimit: boolean;
}

type GitSpawn = (
  command: string,
  args: readonly string[],
  options: {
    readonly cwd: string;
    readonly shell: false;
    readonly detached: boolean;
    readonly stdio: ["pipe", "pipe", "pipe"];
  },
) => ChildProcessWithoutNullStreams;

function killProcessGroup(child: ChildProcessWithoutNullStreams, signal: NodeJS.Signals): void {
  if (process.platform !== "win32" && child.pid !== undefined) {
    try {
      process.kill(-child.pid, signal);
      return;
    } catch {
      // Fall back to the direct process when the group is already gone.
    }
  }
  if (child.exitCode === null && child.signalCode === null) child.kill(signal);
}

export function runBoundedGit(
  args: readonly string[],
  cwd: string,
  signal: AbortSignal | undefined,
  spawnProcess: GitSpawn = spawn,
): Promise<GitProcessResult> {
  return new Promise((resolve) => {
    const stdoutChunks: Buffer[] = [];
    const stderrChunks: Buffer[] = [];
    let stdoutBytes = 0;
    let stderrBytes = 0;
    let settled = false;
    let terminated = false;
    let timedOut = false;
    let outputLimit = false;
    let terminationTimer: ReturnType<typeof setTimeout> | undefined;
    let timeoutTimer: ReturnType<typeof setTimeout> | undefined;
    let child: ChildProcessWithoutNullStreams;

    const finish = (code: number | null): void => {
      if (settled) return;
      settled = true;
      if (terminationTimer) clearTimeout(terminationTimer);
      if (timeoutTimer) clearTimeout(timeoutTimer);
      signal?.removeEventListener("abort", abort);
      resolve({
        stdout: Buffer.concat(stdoutChunks).toString("utf8"),
        stderr: Buffer.concat(stderrChunks).toString("utf8"),
        code,
        killed: terminated,
        timedOut,
        outputLimit,
      });
    };

    const terminate = (reason: "abort" | "timeout" | "output"): void => {
      if (terminated || settled) return;
      terminated = true;
      timedOut ||= reason === "timeout";
      outputLimit ||= reason === "output";
      killProcessGroup(child, "SIGTERM");
      terminationTimer = setTimeout(() => {
        killProcessGroup(child, "SIGKILL");
        finish(null);
      }, TERMINATION_GRACE_MS);
    };

    const abort = (): void => terminate("abort");
    try {
      child = spawnProcess("git", args, {
        cwd,
        shell: false,
        detached: process.platform !== "win32",
        stdio: ["pipe", "pipe", "pipe"],
      });
    } catch {
      finish(null);
      return;
    }

    const collect = (target: "stdout" | "stderr", value: Buffer | string): void => {
      if (settled) return;
      const chunk = Buffer.isBuffer(value) ? value : Buffer.from(value);
      const nextStdoutBytes = target === "stdout" ? stdoutBytes + chunk.byteLength : stdoutBytes;
      const nextStderrBytes = target === "stderr" ? stderrBytes + chunk.byteLength : stderrBytes;
      if (nextStdoutBytes > MAX_OUTPUT_BYTES || nextStderrBytes > MAX_OUTPUT_BYTES || nextStdoutBytes + nextStderrBytes > MAX_COMBINED_OUTPUT_BYTES) {
        outputLimit = true;
        terminate("output");
        return;
      }
      if (target === "stdout") {
        stdoutBytes = nextStdoutBytes;
        stdoutChunks.push(chunk);
      } else {
        stderrBytes = nextStderrBytes;
        stderrChunks.push(chunk);
      }
    };

    child.stdout.on("data", (chunk: Buffer | string) => collect("stdout", chunk));
    child.stderr.on("data", (chunk: Buffer | string) => collect("stderr", chunk));
    child.stdin.on("error", () => undefined);
    child.once("error", () => finish(null));
    child.once("close", (code) => finish(code));
    child.stdin.end();
    timeoutTimer = setTimeout(() => terminate("timeout"), GIT_TIMEOUT_MS);
    if (signal?.aborted) terminate("abort");
    else signal?.addEventListener("abort", abort, { once: true });
  });
}

export function createGitReadTool(pi: ExtensionAPI): ToolDefinition<typeof gitReadSchema> {
  return {
    name: "hive_git_read",
    label: "Hive Git read",
    description: "Read bounded Git state without invoking a shell or mutating the repository.",
    promptSnippet: "Read Git state through Hive's bounded, read-only interface",
    promptGuidelines: ["Use hive_git_read for bounded, read-only Git inspection.", "Do not use generic bash for Git operations."],
    parameters: gitReadSchema,
    executionMode: "sequential",
    async execute(_toolCallId, params, signal, _onUpdate, ctx: ExtensionContext) {
      let args: string[];
      try {
        args = buildGitArguments(params);
      } catch (error) {
        return {
          content: [{ type: "text" as const, text: error instanceof Error ? error.message : String(error) }],
          details: { operation: params.operation, blocked: true },
          isError: true,
        };
      }

      const result = await runBoundedGit(args, ctx.cwd, signal);
      const output = truncateOutput(result.stdout.length > 0 ? result.stdout : result.stderr);
      return {
        content: [{ type: "text" as const, text: output }],
        details: { operation: params.operation, args, exitCode: result.code, killed: result.killed, timedOut: result.timedOut, outputLimit: result.outputLimit },
        isError: result.code !== 0 || result.killed,
      };
    },
  };
}
