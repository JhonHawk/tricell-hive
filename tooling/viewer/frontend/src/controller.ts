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
export type WalkthroughManifest = { title: string; [key: string]: unknown };
export type DocumentData = {
  key: string;
  hash: string;
  fileBase: string;
  manifest: WalkthroughManifest;
};
export type ViewerState = {
  access: "pending" | "ready" | "missing" | "rejected";
  index: RawIndex | null;
  sessions: LibrarySession[];
  selectedId?: string;
  document: DocumentData | null;
  documentState: "empty" | "loading" | "ready" | "error" | "excluded";
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
  writing: boolean;
  showArchived: boolean;
};
const initial: ViewerState = {
  access: "pending",
  index: null,
  sessions: [],
  document: null,
  documentState: "empty",
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
  private loadSeq = 0;
  private reading = false;
  private hidden = new Set<string>();
  private timer: ReturnType<typeof setTimeout> | undefined;
  private active = false;
  private preferred: string | undefined;
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
    this.loadSeq++;
    this.hidden.clear();
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
    this.loadSeq++;
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
    } catch (error) {
      if (epoch !== this.epoch || seq !== this.readSeq) return;
      if (!this.failure(error))
        this.update({ down: !(error instanceof HttpError), indexError: true });
    } finally {
      if (epoch === this.epoch && seq === this.readSeq) this.reading = false;
      this.schedule();
    }
  }
  private apply(index: RawIndex, notifyMissing = true) {
    const mapped = mapIndex(index);
    for (const key of this.hidden)
      if (!mapped.some((s) => s.id === key)) this.hidden.delete(key);
    const sessions = mapped.filter(
      (s) => !this.hidden.has(s.id) && (this.state.showArchived || !s.archived),
    );
    const current = sessions.find((s) => s.id === this.state.selectedId);
    const previous = this.state.sessions.find(
      (s) => s.id === this.state.selectedId,
    );
    this.update({ index, sessions });
    if (!sessions.length) {
      this.loadSeq++;
      this.update({
        selectedId: undefined,
        document: null,
        documentState: "empty",
        notice: previous && notifyMissing ? "gone" : this.state.notice,
      });
      return;
    }
    if (current) {
      if (current.excluded || current.classification === "tracked") {
        this.loadSeq++;
        this.update({ document: null, documentState: "excluded" });
        return;
      }
      if (
        !previous ||
        previous.raw.content_hash !== current.raw.content_hash ||
        previous.classification === "tracked" ||
        previous.fileBase !== current.fileBase ||
        previous.excluded !== current.excluded
      )
        this.load(current);
      return;
    }
    const remembered = stored(this.client.storage, OPEN_KEY);
    const preferred =
      sessions.find((s) => s.id === this.preferred) ||
      sessions.find((s) => s.id === remembered);
    this.preferred = undefined;
    const normal = sessions.filter((s) => !s.excluded);
    const candidates = normal.length ? normal : sessions;
    const latest = [...candidates].sort(
      (a, b) =>
        (Date.parse(b.raw.created_at) || Date.parse(b.raw.ingested_at) || 0) -
          (Date.parse(a.raw.created_at) ||
            Date.parse(a.raw.ingested_at) ||
            0) ||
        (Date.parse(b.raw.ingested_at) || 0) -
          (Date.parse(a.raw.ingested_at) || 0),
    )[0];
    this.select((preferred || latest).id);
    if (previous && notifyMissing) this.update({ notice: "gone" });
  }
  select = (id: string) => {
    const selected = this.state.sessions.find((s) => s.id === id);
    if (!selected) return;
    if (id === this.state.selectedId && this.state.documentState !== "error")
      return;
    save(this.client.storage, OPEN_KEY, id);
    this.update({ selectedId: id });
    this.load(selected);
  };
  private async load(selected: LibrarySession) {
    const seq = ++this.loadSeq;
    const epoch = this.epoch;
    if (selected.excluded || selected.classification === "tracked") {
      this.update({ document: null, documentState: "excluded" });
      return;
    }
    this.update({
      document:
        this.state.document?.key === selected.id ? this.state.document : null,
      documentState: "loading",
    });
    try {
      const manifest = await this.client.request(
        `/api/sessions/${encodeURIComponent(selected.id)}`,
      );
      if (epoch !== this.epoch || seq !== this.loadSeq) return;
      this.update({
        document: {
          key: selected.id,
          hash: selected.raw.content_hash,
          fileBase: selected.fileBase,
          manifest,
        },
        documentState: "ready",
      });
    } catch (error) {
      if (epoch !== this.epoch || seq !== this.loadSeq) return;
      if (this.failure(error)) return;
      this.update({
        document: null,
        documentState:
          error instanceof HttpError && error.status === 409
            ? "excluded"
            : "error",
      });
      if (
        error instanceof HttpError &&
        (error.status === 409 || error.status === 404)
      )
        void this.refresh();
    }
  }
  retryDocument = () => {
    const selected = this.state.sessions.find(
      (s) => s.id === this.state.selectedId,
    );
    if (selected) this.load(selected);
  };
  rendererFailed = () =>
    this.update({ document: null, documentState: "error" });
  setShowArchived = (showArchived: boolean) => {
    save(this.client.storage, ARCHIVED_KEY, String(showArchived));
    this.update({ showArchived });
    if (this.state.index) this.apply(this.state.index, false);
  };
  async archive() {
    const selected = this.state.sessions.find(
      (s) => s.id === this.state.selectedId,
    );
    if (!selected || this.state.writing) return;
    const epoch = this.epoch;
    const archived = !selected.archived;
    this.update({ writing: true });
    try {
      await this.client.request(
        `/api/sessions/${selected.id}/archive`,
        archived ? "POST" : "DELETE",
      );
      if (epoch !== this.epoch) return;
      this.readSeq++;
      if (this.state.index) {
        const update = (
          session: RawIndex["projects"][number]["tickets"][number]["sessions"][number],
        ) => (session.key === selected.id ? { ...session, archived } : session);
        this.apply(
          {
            ...this.state.index,
            projects: this.state.index.projects.map((project) => ({
              ...project,
              tickets: project.tickets.map((ticket) => ({
                ...ticket,
                sessions: ticket.sessions.map(update),
              })),
            })),
            excluded_sessions: this.state.index.excluded_sessions?.map(update),
          },
          false,
        );
      }
      this.update({ notice: archived ? "archived" : "unarchived" });
      void this.refresh();
    } catch (error) {
      if (epoch === this.epoch && !this.failure(error))
        this.update({ notice: "writeError" });
    } finally {
      if (epoch === this.epoch) this.update({ writing: false });
    }
  }
  async pin() {
    const selected = this.state.sessions.find(
      (s) => s.id === this.state.selectedId,
    );
    if (!selected || this.state.writing) return;
    const epoch = this.epoch;
    this.update({ writing: true });
    try {
      try {
        await this.client.request(
          `/api/sessions/${selected.id}/pin`,
          selected.pinned ? "DELETE" : "POST",
        );
      } catch (error) {
        if (!(error instanceof HttpError && error.status === 404)) throw error;
      }
      if (epoch !== this.epoch) return;
      this.readSeq++;
      this.update({
        sessions: this.state.sessions.map((s) =>
          s.id === selected.id ? { ...s, pinned: !selected.pinned } : s,
        ),
        notice: selected.pinned ? "unpinned" : "pinned",
      });
      void this.refresh();
    } catch (error) {
      if (epoch === this.epoch && !this.failure(error))
        this.update({ notice: "writeError" });
    } finally {
      if (epoch === this.epoch) this.update({ writing: false });
    }
  }
  async remove() {
    const selected = this.state.sessions.find(
      (s) => s.id === this.state.selectedId,
    );
    if (!selected || this.state.writing) return;
    const epoch = this.epoch;
    this.update({ writing: true });
    try {
      try {
        await this.client.request(`/api/sessions/${selected.id}`, "DELETE");
      } catch (error) {
        if (!(error instanceof HttpError && error.status === 404)) throw error;
      }
      if (epoch !== this.epoch) return;
      this.readSeq++;
      this.hidden.add(selected.id);
      if (this.state.selectedId === selected.id) {
        const siblings = this.state.sessions.filter(
          (s) =>
            s.projectKey === selected.projectKey &&
            s.ticket === selected.ticket &&
            s.excluded === selected.excluded,
        );
        const at = siblings.findIndex((s) => s.id === selected.id);
        this.preferred = siblings[at + 1]?.id || siblings[at - 1]?.id;
        save(this.client.storage, OPEN_KEY);
        // Only the deleted document's load is superseded by this response.
        this.loadSeq++;
      }
      if (this.state.index) this.apply(this.state.index);
      this.update({ notice: "deleted" });
      void this.refresh();
    } catch (error) {
      if (epoch === this.epoch && !this.failure(error))
        this.update({ notice: "writeError" });
    } finally {
      if (epoch === this.epoch) this.update({ writing: false });
    }
  }
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
