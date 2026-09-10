import type { ChildProcessWithoutNullStreams } from "node:child_process";

export const HOOK_TIMEOUT_MS = 10_000;
export const HOOK_OUTPUT_LIMIT_BYTES = 262_144;
export const HOOK_COMBINED_OUTPUT_LIMIT_BYTES = HOOK_OUTPUT_LIMIT_BYTES * 2;

export type HookMode = "blocking" | "advisory";
export type HookOutputKind = "policy" | "capture";

export type HookOutcome = "allow" | "block" | "warning" | "error";

export interface HookInvocation {
  readonly scriptPath: string;
  readonly args?: readonly string[];
  readonly cwd: string;
  readonly payload: unknown;
  readonly mode: HookMode;
  readonly outputKind?: HookOutputKind;
  readonly env?: NodeJS.ProcessEnv;
  readonly unsetEnv?: readonly string[];
  readonly signal?: AbortSignal;
  readonly timeoutMs?: number;
}

export interface HookResult {
  readonly outcome: HookOutcome;
  readonly reason?: string;
  readonly additionalContext?: string;
  readonly stdout: string;
  readonly stderr: string;
  readonly exitCode: number | null;
  readonly timedOut: boolean;
}

export type HookSpawn = (
  command: string,
  args: readonly string[],
  options: {
    readonly cwd: string;
    readonly env: NodeJS.ProcessEnv;
    readonly stdio: ["pipe", "pipe", "pipe"];
    readonly shell: false;
    readonly detached: boolean;
  },
) => ChildProcessWithoutNullStreams;

export interface HookRunnerOptions {
  readonly spawnProcess?: HookSpawn;
  readonly env?: NodeJS.ProcessEnv;
}

export interface HookPaths {
  readonly bashPolicy: string;
  readonly reviewerGuard: string;
  readonly postToolHub: string;
  readonly flowContext: string;
  readonly flowPlanCapture: string;
  readonly flowSessionContext: string;
  readonly ruleContext: string;
  readonly sessionHygieneReport: string;
}

export interface HookReadiness {
  readonly ready: boolean;
  readonly missing: readonly string[];
}

export interface PlanModeState {
  readonly enabled: boolean;
  readonly executionStarted: boolean;
  readonly toolsBeforePlanMode?: readonly string[];
  readonly candidate?: string;
  readonly candidateAfterEntryId?: string;
  readonly candidateSha256?: string;
  readonly capturePath?: string;
  readonly lastCaptureStatus?: "captured" | "session_only" | "skipped";
}

export type PlanCaptureResult =
  | { readonly status: "captured"; readonly path: string; readonly sha256: string }
  | { readonly status: "session_only"; readonly reason: string }
  | { readonly status: "skipped"; readonly reason: string };

export type GitReadOperation =
  | "status"
  | "diff"
  | "log"
  | "show"
  | "blame"
  | "ls-files"
  | "branch"
  | "remote";

export interface GitReadRequest {
  readonly operation: GitReadOperation;
  readonly path?: string;
  readonly revision?: string;
  readonly limit?: number;
  readonly staged?: boolean;
  readonly line?: string;
}

export interface ChildRegistry {
  hasActiveChildren(parentSession: string): boolean;
  register(parentSession: string, childSession: string): void;
  settle(childSession: string): void;
  replaceObservedChildren(parentSession: string, childSessions: readonly string[]): void;
  clear(parentSession?: string): void;
}
