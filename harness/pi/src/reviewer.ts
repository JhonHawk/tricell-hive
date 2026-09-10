import type { ExtensionAPI, ExtensionContext, ToolCallEvent } from "@earendil-works/pi-coding-agent";
import { Type } from "typebox";
import type { ToolDefinition } from "@earendil-works/pi-coding-agent";
import { defaultHookPaths, markHiveHookWired, recordHiveHookError, type HiveHookExtensionOptions } from "./hooks.ts";
import { checkHookReadiness, runHook } from "./hook-runner.ts";

const readinessSchema = Type.Object({});

function createReviewerReadinessTool(reviewerGuard: string): ToolDefinition<typeof readinessSchema> {
  return {
    name: "hive_reviewer_readiness",
    label: "Hive reviewer readiness",
    description: "Verify that the read-only reviewer guard is present and executable.",
    promptSnippet: "Verify the reviewer guard is loaded before reviewing",
    parameters: readinessSchema,
    executionMode: "sequential",
    async execute() {
      const readiness = checkHookReadiness([reviewerGuard]);
      return {
        content: [{ type: "text" as const, text: readiness.ready ? "Hive reviewer guard ready." : `Missing or non-executable reviewer guard:\n${readiness.missing.join("\n")}` }],
        details: readiness,
        isError: !readiness.ready,
      };
    },
  };
}

function currentSessionId(ctx: ExtensionContext): string {
  return ctx.sessionManager.getSessionId() || ctx.sessionManager.getSessionFile() || "pi-session";
}

function payload(ctx: ExtensionContext, event: ToolCallEvent): Record<string, unknown> {
  return {
    harness: "pi",
    cwd: ctx.cwd,
    session_id: currentSessionId(ctx),
    tool_name: "Bash",
    tool_input: event.input,
  };
}

export function registerHiveReviewerGuard(pi: ExtensionAPI, options: HiveHookExtensionOptions = {}): void {
  const paths = { ...defaultHookPaths(), ...options.paths };
  markHiveHookWired(pi, paths, "reviewerGuard");
  pi.registerTool(createReviewerReadinessTool(paths.reviewerGuard));
  pi.on("tool_call", async (event, ctx) => {
    if (event.toolName.toLowerCase() !== "bash") return undefined;
    const result = await runHook(
      {
        scriptPath: paths.reviewerGuard,
        cwd: ctx.cwd,
        payload: payload(ctx, event),
        mode: "blocking",
        signal: ctx.signal,
      },
      options.runnerOptions,
    );
    if (result.outcome === "block") {
      recordHiveHookError(pi, paths, "reviewerGuard", result.reason ?? "Reviewer guard blocked this command.");
      return {
        block: true,
        reason: result.reason ?? "Reviewer guard blocked this command.",
      };
    }
    recordHiveHookError(pi, paths, "reviewerGuard", undefined);
    return undefined;
  });
}

export function createHiveReviewerGuardExtension(options: HiveHookExtensionOptions = {}) {
  return (pi: ExtensionAPI): void => registerHiveReviewerGuard(pi, options);
}

export default createHiveReviewerGuardExtension();
