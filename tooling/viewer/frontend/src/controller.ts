import {
  type AccessClient,
  ARCHIVED_KEY,
  HttpError,
  type LibrarySession,
  mapIndex,
  OPEN_KEY,
  type RawIndex,
  save,
  stored,
  validateIndex,
} from "./api";
import { groupSessions, type SessionGroup, sessionIdentity } from "./sessions";
export type Delivery = {
  kind: "pull_request" | "commit";
  label: string;
  url: string;
};
export type WalkthroughManifest = {
  title: string;
  delivery?: Delivery[];
  [key: string]: unknown;
};
export type DocumentData = {
  key: string;
  hash: string;
  fileBase: string;
  manifest: WalkthroughManifest;
};
export type SegmentDocument = {
  revision: string;
  state: "loading" | "ready" | "error" | "excluded";
  document: DocumentData | null;
};
export type ViewerState = {
  access: "pending" | "ready" | "missing" | "rejected";
  index: RawIndex | null;
  sessions: LibrarySession[];
  selectedId?: string;
  groups: SessionGroup[];
  documents: Record<string, SegmentDocument>;
  writingId?: string;
  down: boolean;
  indexError: boolean;
  notice?:
    | "reconnected"
    | "gone"
    | "pinned"
    | "unpinned"
    | "deleted"
    | "archived"
    | "unarchived"
    | "writeError";
  actionError?: {
    id: string;
    groupId: string;
    operation: "archive" | "pin" | "title" | "delete";
    value?: string | boolean;
    code?: string;
    deleted?: number;
    remaining?: number;
    cleanup_pending?: boolean;
    cleanupGroups?: string[];
  };
  undoArchive?: { id: string; groupId: string; archived: boolean };
  writing: boolean;
  showArchived: boolean;
};
const initial: ViewerState = {
  access: "pending",
  index: null,
  sessions: [],
  groups: [],
  documents: {},
  down: false,
  indexError: false,
  writing: false,
  showArchived: false,
};
export class ViewerController {
  private state: ViewerState = { ...initial };
  private listeners = new Set<() => void>();
  private epoch = 0;
  private readSeq = 0;
  private requests = new Map<string, object>();
  private cache = new Map<string, SegmentDocument>();
  private reading = false;
  private timer: ReturnType<typeof setTimeout> | undefined;
  private active = false;
  constructor(
    public client: AccessClient,
    private isHidden = () => document.hidden,
  ) {}
  getSnapshot = () => this.state;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };
  private update(patch: Partial<ViewerState>) {
    this.state = { ...this.state, ...patch };
    for (const listener of this.listeners) listener();
  }
  async open(code: string | null) {
    const epoch = ++this.epoch;
    this.readSeq++;
    this.requests.clear();
    this.cache.clear();
    this.reading = false;
    this.update({
      ...initial,
      showArchived: stored(this.client.storage, ARCHIVED_KEY) === "true",
    });
    try {
      await this.client.establish(code);
      if (epoch !== this.epoch) return;
      if (!this.client.token) {
        this.update({ access: "missing" });
        return;
      }
      this.update({ access: "ready" });
      await this.refresh();
    } catch {
      if (epoch === this.epoch) this.reject();
    } finally {
      this.schedule();
    }
  }
  private reject() {
    this.epoch++;
    this.readSeq++;
    this.requests.clear();
    this.cache.clear();
    this.reading = false;
    this.client.reject();
    this.update({ ...initial, access: "rejected" });
    clearTimeout(this.timer);
  }
  private failure(error: unknown) {
    if (error instanceof HttpError && error.status === 403) {
      this.reject();
      return true;
    }
    return false;
  }
  async refresh() {
    if (this.state.access !== "ready") return;
    const epoch = this.epoch;
    const seq = ++this.readSeq;
    this.reading = true;
    try {
      const received = validateIndex(await this.client.request("/api/index"));
      if (epoch !== this.epoch || seq !== this.readSeq) return;
      const wasDown = this.state.down;
      this.update({
        down: false,
        indexError: false,
        notice: wasDown ? "reconnected" : this.state.notice,
      });
      this.apply(received);
      const actionError = this.state.actionError;
      if (
        actionError?.code === "refresh_required" ||
        (actionError?.remaining === 0 &&
          !received.errors.some(
            (error) =>
              error.code === "delete_cleanup_pending" &&
              (!actionError.cleanupGroups?.length ||
                actionError.cleanupGroups.includes(
                  error.session_group ?? error.path,
                )),
          ))
      )
        this.update({ actionError: undefined, notice: undefined });
      return true;
    } catch (error) {
      if (epoch !== this.epoch || seq !== this.readSeq) return;
      if (!this.failure(error))
        this.update({ down: !(error instanceof HttpError), indexError: true });
    } finally {
      if (epoch === this.epoch && seq === this.readSeq) this.reading = false;
      this.schedule();
    }
  }
  private revision(segment: LibrarySession) {
    return JSON.stringify([
      segment.raw.content_hash,
      segment.fileBase,
      segment.classification,
      segment.excluded,
    ]);
  }
  private publishDocuments() {
    const selected = this.state.sessions.find(
      (s) => s.id === this.state.selectedId,
    );
    const group = this.state.groups.find((g) =>
      g.segments.some((s) => s.id === selected?.id),
    );
    const segments = group?.segments ?? (selected ? [selected] : []);
    const documents = Object.fromEntries(
      segments.flatMap((s) => {
        const entry = this.cache.get(s.id);
        return entry ? [[s.id, entry]] : [];
      }),
    );
    this.update({
      documents,
    });
  }
  private apply(index: RawIndex, notifyMissing = true) {
    const mapped = mapIndex(index);
    const sessions = mapped.filter(
      (s) => this.state.showArchived || !s.archived,
    );
    const previous = this.state.sessions.find(
      (s) => s.id === this.state.selectedId,
    );
    const oldGroup = this.state.groups.find((g) =>
      g.segments.some((s) => s.id === previous?.id),
    );
    for (const [key, entry] of this.cache) {
      const segment = sessions.find((s) => s.id === key);
      if (!segment || entry.revision !== this.revision(segment)) {
        this.cache.delete(key);
        this.requests.delete(key);
      }
    }
    const groups = groupSessions(sessions);
    this.update({ index, sessions, groups });
    if (!sessions.length) {
      this.requests.clear();
      this.update({
        selectedId: undefined,
        documents: {},
        notice: previous && notifyMissing ? "gone" : this.state.notice,
      });
      return;
    }
    let selected = sessions.find((s) => s.id === this.state.selectedId);
    if (!selected && oldGroup && previous) {
      const at = oldGroup.segments.findIndex((s) => s.id === previous.id);
      const adjacent = [
        ...oldGroup.segments.slice(at + 1),
        ...oldGroup.segments.slice(0, at).reverse(),
      ];
      selected = adjacent
        .map((s) => sessions.find((current) => current.id === s.id))
        .find(Boolean);
    }
    if (!selected) {
      const remembered = stored(this.client.storage, OPEN_KEY);
      selected =
        sessions.find((s) => s.id === remembered) ??
        groups[0]?.segments[0] ??
        sessions[0];
    }
    this.select(selected.id);
    if (previous && previous.id !== selected.id && notifyMissing)
      this.update({ notice: "gone" });
  }
  selectSession = (groupId: string) => {
    const group = this.state.groups.find((g) => g.id === groupId);
    if (!group) return;
    this.select(group.segments[0].id);
  };
  select = (id: string) => {
    const selected = this.state.sessions.find((s) => s.id === id);
    if (!selected) return;
    save(this.client.storage, OPEN_KEY, id);
    this.update({ selectedId: id });
    const group = this.state.groups.find((g) =>
      g.segments.some((s) => s.id === id),
    );
    const segments = group?.segments ?? [selected];
    for (const key of this.requests.keys()) {
      if (!segments.some((s) => s.id === key)) {
        this.requests.delete(key);
        this.cache.delete(key);
      }
    }
    for (const segment of segments) {
      if (!this.cache.has(segment.id)) void this.load(segment);
    }
    this.publishDocuments();
  };
  private async load(selected: LibrarySession) {
    const request = {};
    const epoch = this.epoch;
    const revision = this.revision(selected);
    if (selected.excluded || selected.classification === "tracked") {
      this.cache.set(selected.id, {
        revision,
        document: null,
        state: "excluded",
      });
      this.publishDocuments();
      return;
    }
    this.requests.set(selected.id, request);
    this.cache.set(selected.id, { revision, document: null, state: "loading" });
    this.publishDocuments();
    try {
      const manifest = await this.client.request(
        `/api/sessions/${encodeURIComponent(selected.id)}`,
      );
      if (epoch !== this.epoch || this.requests.get(selected.id) !== request)
        return;
      this.requests.delete(selected.id);
      this.cache.set(selected.id, {
        revision,
        state: "ready",
        document: {
          key: selected.id,
          hash: selected.raw.content_hash,
          fileBase: selected.fileBase,
          manifest,
        },
      });
      this.publishDocuments();
    } catch (error) {
      if (epoch !== this.epoch || this.requests.get(selected.id) !== request)
        return;
      this.requests.delete(selected.id);
      if (this.failure(error)) return;
      this.cache.set(selected.id, {
        revision,
        document: null,
        state:
          error instanceof HttpError && error.status === 409
            ? "excluded"
            : "error",
      });
      this.publishDocuments();
      if (
        error instanceof HttpError &&
        (error.status === 409 || error.status === 404)
      )
        void this.refresh();
    }
  }
  retryDocument = (id = this.state.selectedId) => {
    const selected = this.state.sessions.find((s) => s.id === id);
    if (selected) void this.load(selected);
  };
  rendererFailed = (id = this.state.selectedId) => {
    const entry = id ? this.cache.get(id) : undefined;
    if (id && entry) {
      this.cache.set(id, { ...entry, document: null, state: "error" });
      this.publishDocuments();
    }
  };
  setShowArchived = (showArchived: boolean) => {
    save(this.client.storage, ARCHIVED_KEY, String(showArchived));
    this.update({ showArchived });
    if (this.state.index) this.apply(this.state.index, false);
  };
  private allSessions() {
    return this.state.index ? mapIndex(this.state.index) : [];
  }
  private async write(
    id: string,
    operation: "archive" | "pin" | "title" | "delete",
    value?: string | boolean,
  ) {
    const selected = this.allSessions().find((s) => s.id === id);
    if (!selected || this.state.writing || this.state.indexError) return false;
    const epoch = this.epoch;
    const groupId = sessionIdentity(selected);
    this.readSeq++;
    this.update({
      writing: true,
      writingId: id,
      actionError: undefined,
      notice: undefined,
    });
    try {
      await this.client.request(
        `/api/sessions/${id}${operation === "delete" ? "" : `/${operation}`}`,
        operation === "title"
          ? "PUT"
          : operation === "delete" || value === false
            ? "DELETE"
            : "POST",
        operation === "title" ? { title: value } : undefined,
      );
      if (epoch !== this.epoch) return false;
      this.readSeq++;
      if (!(await this.refresh())) {
        if (epoch === this.epoch)
          this.update({
            actionError: {
              id,
              groupId,
              operation,
              value,
              code: "refresh_required",
            },
          });
        return false;
      }
      if (epoch !== this.epoch) return false;
      this.update({
        notice:
          operation === "delete"
            ? "deleted"
            : operation === "archive"
              ? value
                ? "archived"
                : "unarchived"
              : operation === "pin"
                ? value
                  ? "pinned"
                  : "unpinned"
                : undefined,
      });
      return true;
    } catch (error) {
      if (epoch !== this.epoch || this.failure(error)) return false;
      const details = error instanceof HttpError ? error.details : {};
      if (operation === "delete") await this.refresh();
      if (epoch === this.epoch)
        this.update({
          notice: "writeError",
          actionError: {
            id,
            groupId,
            operation,
            value,
            ...details,
            cleanupGroups: this.state.index?.errors
              .filter((error) => error.code === "delete_cleanup_pending")
              .map((error) => error.session_group ?? error.path),
          },
        });
      return false;
    } finally {
      if (epoch === this.epoch)
        this.update({ writing: false, writingId: undefined });
    }
  }
  async archive(id = this.state.selectedId) {
    const selected = this.allSessions().find((s) => s.id === id);
    if (!selected) return false;
    const archived = !selected.archived;
    const ok = await this.write(selected.id, "archive", archived);
    if (ok)
      this.update({
        undoArchive: {
          id: selected.id,
          groupId: sessionIdentity(selected),
          archived: !archived,
        },
      });
    return ok;
  }
  async undoArchive() {
    const undo = this.state.undoArchive;
    if (!undo) return;
    const member = this.allSessions().find(
      (s) => sessionIdentity(s) === undo.groupId,
    );
    if (!member) return;
    if (await this.write(member.id, "archive", undo.archived)) {
      this.update({ undoArchive: undefined });
      if (!undo.archived) this.select(member.id);
    }
  }
  clearUndo = () =>
    this.update({
      undoArchive: undefined,
      notice: this.state.actionError ? this.state.notice : undefined,
    });
  async pin(id = this.state.selectedId) {
    const selected = this.allSessions().find((s) => s.id === id);
    return selected ? this.write(selected.id, "pin", !selected.pinned) : false;
  }
  rename(title: string, id = this.state.selectedId) {
    return id ? this.write(id, "title", title) : Promise.resolve(false);
  }
  remove(id = this.state.selectedId) {
    return id ? this.write(id, "delete") : Promise.resolve(false);
  }
  retryAction = () => {
    const error = this.state.actionError;
    if (!error) return;
    if (error.code === "reload_required") return;
    if (error.code === "refresh_required" || error.remaining === 0) {
      void this.refresh();
      return;
    }
    const member = this.allSessions().find(
      (s) => sessionIdentity(s) === error.groupId,
    );
    if (member) void this.write(member.id, error.operation, error.value);
  };
  dismiss = () => this.update({ notice: undefined });
  private schedule() {
    clearTimeout(this.timer);
    if (!this.active || this.state.access !== "ready") return;
    this.timer = setTimeout(
      async () => {
        if (!this.isHidden() && !this.reading) await this.refresh();
        this.schedule();
      },
      this.state.down ? 5000 : 10000,
    );
  }
  private visibility = () => {
    if (!this.isHidden() && !this.reading) void this.refresh();
  };
  private hashChange = () => {
    if (new URLSearchParams(location.hash.slice(1)).has("code"))
      void this.open(takeOpeningCode());
  };
  start() {
    this.active = true;
    document.addEventListener("visibilitychange", this.visibility);
    window.addEventListener("hashchange", this.hashChange);
    this.schedule();
    return () => {
      this.active = false;
      clearTimeout(this.timer);
      document.removeEventListener("visibilitychange", this.visibility);
      window.removeEventListener("hashchange", this.hashChange);
    };
  }
}

import { takeCode as takeOpeningCode } from "./api";
