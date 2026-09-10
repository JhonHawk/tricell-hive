import { randomUUID } from "node:crypto";
import * as fs from "node:fs";
import * as os from "node:os";
import * as path from "node:path";
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import type { ChildRegistry } from "./types.ts";

const CHILD_START_EVENTS = ["subagent:async-started"] as const;
const CHILD_SETTLED_EVENTS = [
  "subagent:async-complete",
  "subagent:foreground-complete",
  "subagent:process-terminal",
] as const;
const SUBAGENT_RPC_REQUEST_EVENT = "subagents:rpc:v1:request";
const SUBAGENT_RPC_REPLY_EVENT_PREFIX = "subagents:rpc:v1:reply:";
const ACTIVE_RUN_INDEX_DIR = ".active-runs";
const RPC_RECONCILIATION_TIMEOUT_MS = 100;
const MAX_STATUS_BYTES = 128 * 1024;

interface JsonRecord {
  readonly [key: string]: unknown;
}

export interface ChildReconciliation {
  readonly active: boolean;
  readonly indeterminate: boolean;
  readonly source: "memory" | "durable" | "none";
}

interface EventBus {
  on(event: string, handler: (data: unknown) => void): (() => void) | void;
  emit(event: string, data: unknown): void;
}

function isRecord(value: unknown): value is JsonRecord {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function nonEmptyString(value: unknown): string | undefined {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

function eventId(value: unknown): string | undefined {
  if (!isRecord(value)) return undefined;
  return nonEmptyString(value.id) ?? nonEmptyString(value.runId) ?? nonEmptyString(value.childId);
}

function eventParent(value: unknown): string | undefined {
  if (!isRecord(value)) return undefined;
  return nonEmptyString(value.parentSessionId) ?? nonEmptyString(value.sessionId);
}

class InMemoryChildRegistry implements ChildRegistry {
  private readonly parents = new Map<string, Set<string>>();
  private readonly observed = new Map<string, Set<string>>();

  public hasActiveChildren(parentSession: string): boolean {
    return (this.parents.get(parentSession)?.size ?? 0) > 0
      || (this.observed.get(parentSession)?.size ?? 0) > 0;
  }

  public register(parentSession: string, childSession: string): void {
    if (parentSession.length === 0 || childSession.length === 0 || parentSession === childSession) return;
    const children = this.parents.get(parentSession) ?? new Set<string>();
    children.add(childSession);
    this.parents.set(parentSession, children);
  }

  public settle(childSession: string): void {
    for (const [parent, children] of this.parents) {
      children.delete(childSession);
      if (children.size === 0) this.parents.delete(parent);
    }
    for (const [parent, children] of this.observed) {
      children.delete(childSession);
      if (children.size === 0) this.observed.delete(parent);
    }
  }

  public replaceObservedChildren(parentSession: string, childSessions: readonly string[]): void {
    const children = new Set(childSessions.filter((child) => child.length > 0 && child !== parentSession));
    if (children.size === 0) this.observed.delete(parentSession);
    else this.observed.set(parentSession, children);
  }

  public clear(parentSession?: string): void {
    if (parentSession) {
      this.parents.delete(parentSession);
      this.observed.delete(parentSession);
    } else {
      this.parents.clear();
      this.observed.clear();
    }
  }
}

export function createChildRegistry(): ChildRegistry {
  return new InMemoryChildRegistry();
}

export const defaultChildRegistry: ChildRegistry = createChildRegistry();

function activeAsyncRoot(): string {
  const configured = process.env.PI_SUBAGENTS_TEMP_ROOT?.trim();
  if (configured) return path.resolve(configured, "async-subagent-runs");
  const uid = typeof process.getuid === "function" ? `uid-${process.getuid()}` : "shared";
  return path.join(os.tmpdir(), `pi-subagents-${uid}`, "async-subagent-runs");
}

function activeRunState(value: unknown): value is "queued" | "running" {
  return value === "queued" || value === "running";
}

function statusForMarker(asyncRoot: string, runId: string, parentSession: string): "active" | "irrelevant" | "indeterminate" {
  if (!/^[A-Za-z0-9._-]{1,255}$/u.test(runId)) return "irrelevant";
  const statusPath = path.join(asyncRoot, runId, "status.json");
  try {
    const stat = fs.lstatSync(statusPath);
    if (!stat.isFile() || stat.size > MAX_STATUS_BYTES) return "irrelevant";
    let parsed: unknown;
    try {
      parsed = JSON.parse(fs.readFileSync(statusPath, "utf8"));
    } catch {
      return "irrelevant";
    }
    if (!isRecord(parsed) || typeof parsed.sessionId !== "string") return "irrelevant";
    if (parsed.sessionId !== parentSession) return "irrelevant";
    if (typeof parsed.runId !== "string" || parsed.runId !== runId || typeof parsed.state !== "string") return "indeterminate";
    return activeRunState(parsed.state) ? "active" : "irrelevant";
  } catch {
    return "irrelevant";
  }
}

function readDurableActiveChildren(parentSession: string, asyncRoot = activeAsyncRoot()): { ids: string[]; indeterminate: boolean } {
  const indexRoot = path.join(asyncRoot, ACTIVE_RUN_INDEX_DIR);
  let entries: fs.Dirent[];
  try {
    entries = fs.readdirSync(indexRoot, { withFileTypes: true });
  } catch (error) {
    const code = (error as NodeJS.ErrnoException).code;
    if (code === "ENOENT" || code === "ENOTDIR") return { ids: [], indeterminate: false };
    return { ids: ["hive-durable-uncertain"], indeterminate: true };
  }

  const ids: string[] = [];
  let indeterminate = false;
  for (const entry of entries) {
    if (!entry.isFile()) continue;
    const result = statusForMarker(asyncRoot, entry.name, parentSession);
    if (result === "active") ids.push(`durable:${entry.name}`);
    else if (result === "indeterminate") indeterminate = true;
  }
  // Markers without a valid owner cannot be assigned to this session. Valid
  // status records with this session ID remain fail-closed above.
  if (indeterminate) ids.push(`hive-durable-uncertain:${parentSession}`);
  return { ids, indeterminate };
}

function rpcFleetChildren(value: unknown, parentSession: string): { ids: string[]; indeterminate: boolean } | undefined {
  if (!isRecord(value) || value.success !== true || !isRecord(value.data) || !isRecord(value.data.fleet)) return undefined;
  const fleet = value.data.fleet;
  if (typeof fleet.totalActive !== "number" || !Number.isSafeInteger(fleet.totalActive) || fleet.totalActive < 0) {
    return { ids: ["hive-rpc-uncertain"], indeterminate: true };
  }
  const ids: string[] = [];
  if (Array.isArray(fleet.entries)) {
    for (const entry of fleet.entries) {
      if (!isRecord(entry) || typeof entry.key !== "string" || entry.key.length === 0) {
        return { ids: ["hive-rpc-uncertain"], indeterminate: true };
      }
      ids.push(`rpc:${parentSession}:${entry.key}`);
    }
  } else if (fleet.totalActive > 0) {
    return { ids: ["hive-rpc-uncertain"], indeterminate: true };
  }
  if (ids.length > fleet.totalActive) return { ids: ["hive-rpc-uncertain"], indeterminate: true };
  for (let index = ids.length; index < fleet.totalActive; index += 1) ids.push(`rpc:${parentSession}:omitted-${index}`);
  return { ids, indeterminate: false };
}

async function querySubagentStatus(events: EventBus, parentSession: string, timeoutMs: number): Promise<{ ids: string[]; indeterminate: boolean } | undefined> {
  const requestId = `hive-reconcile-${randomUUID()}`;
  const replyEvent = `${SUBAGENT_RPC_REPLY_EVENT_PREFIX}${requestId}`;
  return new Promise((resolve) => {
    let settled = false;
    let timer: ReturnType<typeof setTimeout> | undefined;
    const unsubscribe = events.on(replyEvent, (payload) => {
      if (settled) return;
      settled = true;
      if (timer) clearTimeout(timer);
      if (typeof unsubscribe === "function") unsubscribe();
      resolve(rpcFleetChildren(payload, parentSession));
    });
    timer = setTimeout(() => {
      if (settled) return;
      settled = true;
      if (typeof unsubscribe === "function") unsubscribe();
      resolve(undefined);
    }, timeoutMs);
    events.emit(SUBAGENT_RPC_REQUEST_EVENT, {
      version: 1,
      requestId,
      method: "status",
      source: { extension: "hive-plan" },
    });
  });
}

/** Refresh durable child state for one parent session before plan entry. */
export async function reconcileActiveChildren(
  pi: Pick<ExtensionAPI, "events">,
  registry: ChildRegistry,
  parentSession: string,
  options: { readonly timeoutMs?: number; readonly asyncRoot?: string } = {},
): Promise<ChildReconciliation> {
  if (parentSession.length === 0) return { active: registry.hasActiveChildren(parentSession), indeterminate: false, source: "memory" };
  const memoryActive = registry.hasActiveChildren(parentSession);
  const [rpc, durable] = await Promise.all([
    querySubagentStatus(pi.events, parentSession, options.timeoutMs ?? RPC_RECONCILIATION_TIMEOUT_MS),
    Promise.resolve(readDurableActiveChildren(parentSession, options.asyncRoot)),
  ]);
  const ids = new Set(durable.ids);
  if (rpc) for (const id of rpc.ids) ids.add(id);
  registry.replaceObservedChildren(parentSession, [...ids]);
  const active = memoryActive || ids.size > 0;
  return {
    active,
    indeterminate: durable.indeterminate || rpc?.indeterminate === true,
    source: ids.size > 0 ? "durable" : memoryActive ? "memory" : "none",
  };
}

/**
 * Project pi-subagents lifecycle events into the plan-mode registry.
 * Event names are intentionally string literals so this bridge remains
 * optional when users load the Hive extension without pi-subagents.
 */
export function attachChildLifecycle(
  pi: ExtensionAPI,
  registry: ChildRegistry,
  parentSession: () => string,
): () => void {
  const unsubscribers: Array<() => void> = [];
  for (const channel of CHILD_START_EVENTS) {
    unsubscribers.push(
      pi.events.on(channel, (payload: unknown) => {
        const child = eventId(payload);
        if (!child) return;
        registry.register(eventParent(payload) ?? parentSession(), child);
      }),
    );
  }
  for (const channel of CHILD_SETTLED_EVENTS) {
    unsubscribers.push(
      pi.events.on(channel, (payload: unknown) => {
        const child = eventId(payload);
        if (child) registry.settle(child);
      }),
    );
  }
  return () => {
    for (const unsubscribe of unsubscribers) unsubscribe();
  };
}
