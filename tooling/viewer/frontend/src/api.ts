import type { ViewerSession } from "./shell";
export const TOKEN_KEY = "hive-view-token";
export const OPEN_KEY = "hive-view-open";
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
  constructor(public status: number) {
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
  async request(path: string, method = "GET") {
    const response = await this.fetcher(path, {
      method,
      headers: { "X-Hive-Token": this.token },
      cache: "no-store",
      credentials: "same-origin",
    });
    if (!response.ok) throw new HttpError(response.status);
    return method === "GET" ? response.json() : undefined;
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
  session_id: string;
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
  projects: {
    key: string;
    name: string;
    absent?: boolean;
    tickets: { ticket: string; sessions: RawSession[] }[];
  }[];
  excluded_sessions?: RawSession[];
  repositories: { root: string; absent: boolean }[];
  errors: { path: string; reason: string }[];
  ttl_seconds: number;
};
export type LibrarySession = ViewerSession & {
  raw: RawSession;
  excluded: boolean;
  absent: boolean;
  fileBase: string;
};
export function validateIndex(value: RawIndex): RawIndex {
  if (
    !value ||
    !Array.isArray(value.projects) ||
    !Array.isArray(value.repositories) ||
    !Array.isArray(value.errors) ||
    !Number.isFinite(value.ttl_seconds) ||
    value.ttl_seconds < 0
  )
    throw new Error("Invalid viewer index");
  for (const repository of value.repositories) {
    if (
      !repository ||
      typeof repository.root !== "string" ||
      typeof repository.absent !== "boolean"
    )
      throw new Error("Invalid viewer repository");
  }
  for (const error of value.errors) {
    if (
      !error ||
      typeof error.path !== "string" ||
      typeof error.reason !== "string"
    )
      throw new Error("Invalid viewer ingest error");
  }
  for (const project of value.projects) {
    if (
      typeof project.key !== "string" ||
      typeof project.name !== "string" ||
      !Array.isArray(project.tickets)
    )
      throw new Error("Invalid viewer project");
    for (const ticket of project.tickets) {
      if (typeof ticket.ticket !== "string" || !Array.isArray(ticket.sessions))
        throw new Error("Invalid viewer ticket");
    }
  }
  if (
    value.excluded_sessions !== undefined &&
    !Array.isArray(value.excluded_sessions)
  )
    throw new Error("Invalid excluded sessions");
  for (const session of [
    ...value.projects.flatMap((p) => p.tickets.flatMap((t) => t.sessions)),
    ...(value.excluded_sessions ?? []),
  ]) {
    if (
      !session ||
      !/^[a-f0-9]{64}$/.test(session.key) ||
      typeof session.title !== "string" ||
      typeof session.session_id !== "string" ||
      typeof session.content_hash !== "string" ||
      typeof session.source !== "string"
    )
      throw new Error("Invalid viewer session");
    for (const field of [
      "project_key",
      "project_name",
      "repository",
      "ticket",
      "source_reason",
    ] as const) {
      if (session[field] !== undefined && typeof session[field] !== "string")
        throw new Error("Invalid viewer session metadata");
    }
  }
  return value;
}
export function mapIndex(
  index: RawIndex,
  language = typeof navigator === "undefined" ? "en" : navigator.language,
): LibrarySession[] {
  const format = new Intl.DateTimeFormat(
    language.startsWith("es") ? "es" : "en",
    { day: "numeric", month: "short", hour: "2-digit", minute: "2-digit" },
  );
  const map = (
    raw: RawSession,
    project: string,
    projectKey: string,
    ticket: string,
    excluded: boolean,
    absent = false,
  ): LibrarySession => {
    const time = Date.parse(raw.created_at) || Date.parse(raw.ingested_at);
    const classification =
      raw.source_status === "tracked"
        ? "tracked"
        : raw.source_status === "local"
          ? "local"
          : "unknown";
    const fileBase = raw.fileBase ?? "";
    if (
      !excluded &&
      !new RegExp(`^/files/${raw.key}/[a-f0-9]{64}/$`).test(fileBase)
    )
      throw new Error("Invalid session capability path");
    return {
      id: raw.key,
      project,
      projectKey,
      ticket:
        ticket || (language.startsWith("es") ? "Sin ticket" : "No ticket"),
      title: raw.title || raw.session_id,
      createdAt: Number.isNaN(time) ? "" : format.format(time),
      classification,
      pinned: Boolean(raw.pinned),
      status:
        raw.session_status === "in_progress" || raw.session_status === "closed"
          ? raw.session_status
          : "unknown",
      raw,
      excluded,
      absent,
      fileBase,
    };
  };
  return [
    ...index.projects.flatMap((p) =>
      p.tickets.flatMap((t) =>
        t.sessions.map((s) =>
          map(s, p.name || p.key, p.key, t.ticket, false, p.absent),
        ),
      ),
    ),
    ...(index.excluded_sessions ?? []).map((s) =>
      map(
        s,
        s.project_name || s.repository || s.source,
        s.project_key || s.repository || s.source,
        s.ticket || "",
        true,
      ),
    ),
  ];
}
