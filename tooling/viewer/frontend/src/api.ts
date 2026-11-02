import { resolveLanguage } from "./language";
import type { ViewerSession } from "./shell";
import {
  parseTimelineIndexEntry,
  type TimelineAPIIndexEntry,
  TimelineContractError,
} from "./timeline-contract";
export const TOKEN_KEY = "hive-view-token";
export const OPEN_KEY = "hive-view-open";
export const ARCHIVED_KEY = "hive-view-show-archived";
export type StorageLike = Pick<Storage, "getItem" | "setItem" | "removeItem">;
export function optionalStorage(): StorageLike | undefined {
  try {
    return sessionStorage;
  } catch {
    return undefined;
  }
}
export function stored(storage: StorageLike | undefined, key: string) {
  try {
    return storage?.getItem(key) || "";
  } catch {
    return "";
  }
}
export function save(
  storage: StorageLike | undefined,
  key: string,
  value?: string,
) {
  try {
    if (value === undefined) storage?.removeItem(key);
    else storage?.setItem(key, value);
  } catch {
    /* Storage is optional. */
  }
}
export function takeCode(
  location: Pick<Location, "hash" | "pathname" | "search"> = window.location,
  history: Pick<History, "replaceState"> = window.history,
) {
  const fragment = new URLSearchParams(location.hash.slice(1));
  const code = fragment.get("code");
  if (fragment.has("code") || fragment.has("token"))
    history.replaceState(null, "", location.pathname + location.search);
  return code;
}
export class HttpError extends Error {
  constructor(
    public status: number,
    public details: {
      code?: string;
      deleted?: number;
      remaining?: number;
      cleanup_pending?: boolean;
    } = {},
  ) {
    super(`Viewer request failed (HTTP ${status})`);
  }
}
export class AccessClient {
  token = "";
  private opening: Promise<void> | undefined;
  private openingCode: string | null = null;
  private epoch = 0;
  constructor(
    private fetcher: typeof fetch = globalThis.fetch.bind(globalThis),
    public storage: StorageLike | undefined = optionalStorage(),
  ) {}
  establish(code: string | null) {
    if (code !== null && code === this.openingCode && this.opening)
      return this.opening;
    const epoch = ++this.epoch;
    this.openingCode = code;
    this.token = "";
    if (code === null) {
      this.token = stored(this.storage, TOKEN_KEY);
      return Promise.resolve();
    }
    save(this.storage, TOKEN_KEY);
    this.opening = (async () => {
      const response = await this.fetcher("/api/access-codes/exchange", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ code }),
        cache: "no-store",
        credentials: "same-origin",
      });
      if (epoch !== this.epoch) return;
      if (!response.ok) throw new HttpError(response.status);
      const result = await response.json();
      if (epoch !== this.epoch) return;
      if (typeof result.token !== "string" || !result.token)
        throw new Error("Invalid access response");
      this.token = result.token;
      save(this.storage, TOKEN_KEY, this.token);
    })();
    return this.opening;
  }
  reject() {
    this.epoch++;
    this.token = "";
    save(this.storage, TOKEN_KEY);
  }
  async request(
    path: string,
    method = "GET",
    body?: unknown,
    signal?: AbortSignal,
    responseJSON = method === "GET",
  ) {
    const response = await this.fetcher(path, {
      method,
      ...(signal ? { signal } : {}),
      headers: {
        "X-Hive-Token": this.token,
        ...(method !== "GET" ? { "X-Hive-Action-Scope": "session" } : {}),
        ...(body !== undefined ? { "Content-Type": "application/json" } : {}),
      },
      ...(body !== undefined ? { body: JSON.stringify(body) } : {}),
      cache: "no-store",
      credentials: "same-origin",
    });
    if (!response.ok) {
      let details = {};
      try {
        details = await response.json();
      } catch {
        /* Older errors may be plain text. */
      }
      const fields =
        details && typeof details === "object"
          ? (details as Record<string, unknown>)
          : {};
      throw new HttpError(response.status, {
        code: typeof fields.code === "string" ? fields.code : undefined,
        deleted:
          typeof fields.deleted === "number" &&
          Number.isInteger(fields.deleted) &&
          fields.deleted >= 0
            ? fields.deleted
            : undefined,
        remaining:
          typeof fields.remaining === "number" &&
          Number.isInteger(fields.remaining) &&
          fields.remaining >= 0
            ? fields.remaining
            : undefined,
        cleanup_pending:
          typeof fields.cleanup_pending === "boolean"
            ? fields.cleanup_pending
            : undefined,
      });
    }
    return responseJSON ? response.json() : undefined;
  }
}
export type RawSession = {
  key: string;
  source: string;
  source_status: string;
  source_reason?: string;
  content_hash: string;
  ingested_at: string;
  pinned?: boolean;
  archived?: boolean;
  session_id: string;
  session_title?: string;
  session_status?: string;
  host?: string;
  ticket?: string;
  title: string;
  kind?: string;
  created_at: string;
  project_key?: string;
  project_name?: string;
  repository?: string;
  fileBase?: string;
};
export type RawIndex = {
  protocol_version: 3;
  retention: "manual";
  ttl_seconds: 0;
  timeline_sessions: TimelineAPIIndexEntry[];
  repositories: { root: string; absent: boolean }[];
  errors: { path: string; reason: string; code?: string }[];
  /** Set by validateIndex, never by the server: entries the client skipped. */
  skipped_sessions?: SkippedSession[];
};
export type SkippedSession = { position: number; code: string; id?: string };
export function validateIndex(
  value: RawIndex,
  previous?: RawIndex | null,
): RawIndex {
  if (
    value?.protocol_version !== 3 ||
    value.retention !== "manual" ||
    value.ttl_seconds !== 0 ||
    !Array.isArray(value.timeline_sessions) ||
    !Array.isArray(value.repositories) ||
    !Array.isArray(value.errors) ||
    Object.keys(value).some(
      (k) =>
        ![
          "protocol_version",
          "retention",
          "ttl_seconds",
          "timeline_sessions",
          "repositories",
          "errors",
        ].includes(k),
    )
  )
    throw new Error("Incompatible viewer index; update the viewer and reload");
  for (const r of value.repositories)
    if (!r || typeof r.root !== "string" || typeof r.absent !== "boolean")
      throw new Error("Invalid registered repository");
  for (const e of value.errors)
    if (
      !e ||
      typeof e.path !== "string" ||
      typeof e.reason !== "string" ||
      (e.code !== undefined && typeof e.code !== "string")
    )
      throw new Error("Invalid viewer error");
  // One entry the client cannot read must not hide every other session: skip
  // it, keep the last copy read when there is one, and record it apart from
  // the server's errors. Duplicates and non-contract failures still fail.
  const ids = new Set<string>();
  const sessions: TimelineAPIIndexEntry[] = [];
  const skipped: SkippedSession[] = [];
  value.timeline_sessions.forEach((e, position) => {
    let entry: TimelineAPIIndexEntry | undefined;
    try {
      const parsed = parseTimelineIndexEntry(e);
      if (parsed.mode !== "timeline_api")
        throw new Error("Unsupported viewer work");
      entry = parsed;
    } catch (error) {
      if (!(error instanceof TimelineContractError)) throw error;
      const raw = (e as { viewer_session_id?: unknown } | null)
        ?.viewer_session_id;
      const id =
        typeof raw === "string" && VIEWER_SESSION_ID.test(raw)
          ? raw
          : undefined;
      skipped.push({ position, code: error.code, ...(id ? { id } : {}) });
      entry = id
        ? previous?.timeline_sessions.find((s) => s.viewer_session_id === id)
        : undefined;
    }
    if (!entry) return;
    if (ids.has(entry.viewer_session_id))
      throw new Error("Duplicate viewer work");
    ids.add(entry.viewer_session_id);
    sessions.push(entry);
  });
  if (!skipped.length) return value;
  return { ...value, timeline_sessions: sessions, skipped_sessions: skipped };
}
const VIEWER_SESSION_ID = /^[0-9a-f]{64}$/;

export type ViewerOpenTarget =
  | { sessionID: string; eventID?: string }
  | { invalid: true };
export function takeTarget(
  search = window.location.search,
): ViewerOpenTarget | undefined {
  const query = new URLSearchParams(search);
  if (!query.has("session_id") && !query.has("event_id")) return;
  const sessionID = query.get("session_id");
  const eventID = query.get("event_id");
  if (
    query.getAll("session_id").length !== 1 ||
    query.getAll("event_id").length > 1 ||
    !sessionID ||
    !/^[a-f0-9]{64}$/.test(sessionID) ||
    (eventID !== null && !/^[A-Za-z0-9][A-Za-z0-9_-]{0,127}$/.test(eventID))
  )
    return { invalid: true };
  return { sessionID, ...(eventID ? { eventID } : {}) };
}
export function mapTimelineIndex(
  entry: TimelineAPIIndexEntry,
  language = typeof navigator === "undefined" ? "en" : navigator.language,
): ViewerSession {
  const date = new Date(entry.created_at);
  return {
    id: entry.viewer_session_id,
    project: entry.project_name,
    projectKey: entry.project_key,
    title: entry.title,
    localTitle: true,
    ticket:
      entry.tickets.map((t) => t.key).join(" · ") ||
      (resolveLanguage(language) === "es" ? "Sin ticket" : "No ticket"),
    createdAt: new Intl.DateTimeFormat(
      resolveLanguage(language) === "es" ? "es" : "en",
      { day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" },
    ).format(date),
    classification: "unknown",
    pinned: entry.pinned,
    archived: entry.archived,
    status: ["active", "in_review", "awaiting_validation", "paused"].includes(
      entry.state,
    )
      ? "in_progress"
      : entry.state === "closed"
        ? "closed"
        : "unknown",
  };
}
