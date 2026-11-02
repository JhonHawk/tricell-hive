import {
  type AccessClient,
  ARCHIVED_KEY,
  HttpError,
  OPEN_KEY,
  type RawIndex,
  save,
  stored,
  takeCode as takeOpeningCode,
  takeTarget,
  type ViewerOpenTarget,
  validateIndex,
} from "./api";
import {
  parseTimelineDocumentResponse,
  parseTimelineOperationResult,
  parseTimelineSessionResponse,
  type TimelineAPIIndexEntry,
  type TimelineSessionResponse,
} from "./timeline-contract";
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
export type TimelineRead = {
  id: string;
  state: "loading" | "ready" | "error";
  data: TimelineSessionResponse | null;
};
export type ViewerState = {
  timelineSessions: TimelineAPIIndexEntry[];
  timeline: TimelineRead | null;
  openTarget?: ViewerOpenTarget;
  targetError: boolean;
  access: "pending" | "ready" | "missing" | "rejected";
  index: RawIndex | null;
  selectedId?: string;
  writingId?: string;
  pendingTimelineDelete?: {
    id: string;
    title: string;
    expectedVersion: number;
  };
  down: boolean;
  indexError: boolean;
  notice?: "reconnected" | "gone" | "deleted" | "archived";
  undoArchive?: { id: string; expectedVersion: number };
  writing: boolean;
  showArchived: boolean;
};
const initial: ViewerState = {
  timelineSessions: [],
  timeline: null,
  targetError: false,
  access: "pending",
  index: null,
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
  private reading = false;
  private timer: ReturnType<typeof setTimeout> | undefined;
  private active = false;
  private timelineRequest: object | undefined;
  private timelineRequestedVersion = 0;
  private timelineTask: Promise<void> | undefined;
  private temporaryTargetID: string | undefined;
  private timelineMutation:
    | { key: string; body: Record<string, unknown> }
    | undefined;
  private pendingTimelineDeleteMutation:
    | { key: string; body: Record<string, unknown> }
    | undefined;
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
    for (const l of this.listeners) l();
  }
  async open(code: string | null, target?: ViewerOpenTarget) {
    const epoch = ++this.epoch;
    this.readSeq++;
    this.timelineRequest = undefined;
    this.timelineTask = undefined;
    this.timelineMutation = undefined;
    this.pendingTimelineDeleteMutation = undefined;
    this.temporaryTargetID =
      target && "sessionID" in target ? target.sessionID : undefined;
    this.reading = false;
    this.update({
      ...initial,
      pendingTimelineDelete: undefined,
      undoArchive: undefined,
      notice: undefined,
      selectedId: undefined,
      openTarget: target,
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
    this.reading = false;
    this.timelineRequest = undefined;
    this.timelineTask = undefined;
    this.timelineMutation = undefined;
    this.pendingTimelineDeleteMutation = undefined;
    this.client.reject();
    this.update({
      ...initial,
      pendingTimelineDelete: undefined,
      undoArchive: undefined,
      notice: undefined,
      selectedId: undefined,
      openTarget: undefined,
      access: "rejected",
    });
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
    const epoch = this.epoch,
      seq = ++this.readSeq;
    this.reading = true;
    try {
      const index = validateIndex(await this.client.request("/api/index"));
      if (epoch !== this.epoch || seq !== this.readSeq) return;
      this.update({
        down: false,
        indexError: false,
        notice: this.state.down ? "reconnected" : this.state.notice,
      });
      this.apply(index);
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
  private apply(index: RawIndex) {
    const previous = this.state.selectedId;
    const timelineSessions = index.timeline_sessions.filter(
      (s) =>
        this.state.showArchived ||
        !s.archived ||
        s.viewer_session_id === this.temporaryTargetID,
    );
    this.update({ index, timelineSessions });
    const target = this.state.openTarget;
    if (target) {
      if (
        !("sessionID" in target) ||
        !timelineSessions.some((s) => s.viewer_session_id === target.sessionID)
      ) {
        this.timelineRequest = undefined;
        this.update({
          selectedId: undefined,
          timeline: null,
          targetError: true,
        });
        return;
      }
      this.selectEntry(target.sessionID, true);
      return;
    }
    const id =
      timelineSessions.find((s) => s.viewer_session_id === previous)
        ?.viewer_session_id ??
      timelineSessions.find(
        (s) => s.viewer_session_id === stored(this.client.storage, OPEN_KEY),
      )?.viewer_session_id ??
      timelineSessions[0]?.viewer_session_id;
    if (id) this.selectEntry(id, true);
    else {
      this.timelineRequest = undefined;
      this.update({
        selectedId: undefined,
        timeline: null,
        notice: previous ? "gone" : this.state.notice,
      });
    }
  }
  select = (id: string) => this.selectEntry(id);
  private selectEntry(id: string, preserveTarget = false) {
    const entry = this.state.timelineSessions.find(
      (s) => s.viewer_session_id === id,
    );
    if (!entry) return;
    if (!preserveTarget) {
      this.temporaryTargetID = undefined;
      this.update({ openTarget: undefined, targetError: false });
    }
    save(this.client.storage, OPEN_KEY, id);
    const cached = this.state.timeline?.id === id ? this.state.timeline : null;
    this.update({ selectedId: id, timeline: cached });
    if (
      !cached ||
      cached.state === "error" ||
      (entry.version > (cached.data?.session.version ?? 0) &&
        (cached.state !== "loading" ||
          entry.version > this.timelineRequestedVersion))
    )
      this.timelineTask = this.loadTimeline(entry);
  }
  setShowArchived = (showArchived: boolean) => {
    save(this.client.storage, ARCHIVED_KEY, String(showArchived));
    this.update({ showArchived });
    if (this.state.index) this.apply(this.state.index);
  };
  dismiss = () => this.update({ notice: undefined });
  clearUndo = () => this.update({ undoArchive: undefined });
  undoArchive = async () => {
    const undo = this.state.undoArchive;
    if (!undo) return;
    if (!(await this.refresh())) throw new Error("Refresh required");
    const current = this.state.index?.timeline_sessions.find(
      (s) => s.viewer_session_id === undo.id,
    );
    if (!current) throw new Error("Session unavailable");
    await this.updateTimelineSession(
      undo.id,
      { archived: false },
      current.version,
    );
    this.update({ undoArchive: undefined });
  };
  async recoverTimelineLead(
    id: string,
    requestID: string,
    reason: string,
    expectedVersion: number,
  ) {
    if (
      this.state.access !== "ready" ||
      this.state.writing ||
      this.state.indexError ||
      this.state.selectedId !== id
    )
      throw new Error("Session unavailable");
    const epoch = this.epoch;
    this.readSeq++;
    this.update({ writing: true, writingId: id });
    const operationID = crypto.randomUUID();
    const abort = new AbortController();
    const deadline = setTimeout(() => abort.abort(), 15000);
    try {
      const result = parseTimelineOperationResult(
        await this.client.request(
          `/api/timeline/sessions/${id}/recover-lead`,
          "POST",
          {
            schema_version: 3,
            operation_id: operationID,
            actor: { kind: "local_viewer" },
            expected_session_version: expectedVersion,
            request_id: requestID,
            reason,
          },
          abort.signal,
          true,
        ),
      );
      if (epoch !== this.epoch) throw new Error("Access changed");
      if (
        result.kind !== "recover_lead" ||
        result.operation_id !== operationID ||
        result.viewer_session_id !== id ||
        result.request_id !== requestID
      )
        throw new Error("Invalid recovery receipt");
      if (!(await this.refresh())) throw new Error("Refresh required");
      await this.timelineTask;
    } catch (error) {
      if (epoch === this.epoch && !this.failure(error)) await this.refresh();
      throw error;
    } finally {
      clearTimeout(deadline);
      if (epoch === this.epoch)
        this.update({ writing: false, writingId: undefined });
    }
  }
  private async loadTimeline(entry: TimelineAPIIndexEntry) {
    const request = {};
    const epoch = this.epoch;
    const id = entry.viewer_session_id;
    this.timelineRequest = request;
    this.timelineRequestedVersion = entry.version;
    const previous =
      this.state.timeline?.id === id ? this.state.timeline.data : null;
    this.update({ timeline: { id, state: "loading", data: previous } });
    try {
      const data = parseTimelineSessionResponse(
        await this.client.request(`/api/timeline/sessions/${id}`),
      );
      if (
        epoch !== this.epoch ||
        this.timelineRequest !== request ||
        this.state.selectedId !== id
      )
        return;
      if (
        data.session.viewer_session_id !== id ||
        data.session.version < entry.version
      )
        throw new Error("Invalid timeline session read");
      const target = this.state.openTarget;
      const targetError = !!(
        target &&
        "sessionID" in target &&
        target.sessionID === id &&
        target.eventID &&
        !data.session.events.some((e) => e.event_id === target.eventID)
      );
      this.update({ timeline: { id, state: "ready", data }, targetError });
    } catch (error) {
      if (
        epoch !== this.epoch ||
        this.timelineRequest !== request ||
        this.state.selectedId !== id
      )
        return;
      if (!this.failure(error))
        this.update({ timeline: { id, state: "error", data: previous } });
    }
  }
  retryTimeline = async () => {
    if (
      this.state.indexError ||
      this.state.targetError ||
      !this.state.selectedId
    ) {
      if (!(await this.refresh())) throw new Error("Refresh required");
    }
    const selected = this.state.timelineSessions.find(
      (s) => s.viewer_session_id === this.state.selectedId,
    );
    if (selected) {
      this.timelineTask = this.loadTimeline(selected);
      await this.timelineTask;
      if (this.state.timeline?.state !== "ready")
        throw new Error("Refresh required");
    }
  };
  async loadTimelineDocument(sessionID: string, documentID: string) {
    const epoch = this.epoch;
    const session = this.state.timeline?.data?.session;
    if (
      this.state.access !== "ready" ||
      session?.viewer_session_id !== sessionID ||
      !session.documents.some((d) => d.document_id === documentID)
    )
      throw new Error("Document unavailable");
    try {
      const data = await parseTimelineDocumentResponse(
        await this.client.request(
          `/api/timeline/sessions/${sessionID}/documents/${encodeURIComponent(documentID)}`,
        ),
      );
      const expected = session.documents.find(
        (d) => d.document_id === documentID,
      );
      if (
        epoch !== this.epoch ||
        data.document.session_id !== sessionID ||
        data.document.document_id !== documentID ||
        data.document.sha256 !== expected?.sha256 ||
        data.document.size !== expected?.size
      )
        throw new Error("Document unavailable");
      return data;
    } catch (error) {
      if (epoch === this.epoch) this.failure(error);
      throw error;
    }
  }
  updateTimelineSession = (
    id: string,
    patch: { title?: string; archived?: boolean; pinned?: boolean },
    expectedVersion: number,
  ) => this.writeTimeline(id, "PATCH", expectedVersion, patch);
  deleteTimelineSession = (id: string, expectedVersion: number) =>
    this.writeTimeline(id, "DELETE", expectedVersion);
  retryTimelineDelete = () => {
    const pending = this.state.pendingTimelineDelete;
    if (!pending) return Promise.resolve();
    return this.deleteTimelineSession(pending.id, pending.expectedVersion);
  };
  private async writeTimeline(
    id: string,
    method: "PATCH" | "DELETE",
    expectedVersion: number,
    patch?: { title?: string; archived?: boolean; pinned?: boolean },
  ) {
    const pending = this.state.pendingTimelineDelete;
    const retryDelete =
      method === "DELETE" &&
      pending?.id === id &&
      pending.expectedVersion === expectedVersion &&
      this.pendingTimelineDeleteMutation !== undefined;
    if (
      this.state.access !== "ready" ||
      this.state.writing ||
      (method === "DELETE" && pending && !retryDelete) ||
      (!retryDelete &&
        (this.state.indexError ||
          !this.state.index?.timeline_sessions?.some(
            (s) => s.viewer_session_id === id,
          )))
    )
      throw new Error("Session unavailable");
    const epoch = this.epoch;
    const key = JSON.stringify([
      id,
      method,
      expectedVersion,
      patch?.title,
      patch?.archived,
      patch?.pinned,
    ]);
    const mutation =
      retryDelete && this.pendingTimelineDeleteMutation
        ? this.pendingTimelineDeleteMutation
        : this.timelineMutation?.key === key
          ? this.timelineMutation
          : {
              key,
              body: {
                schema_version: 3,
                actor: { kind: "local_viewer" },
                operation_id: crypto.randomUUID(),
                expected_version: expectedVersion,
                ...patch,
              },
            };
    const deleteTitle =
      pending?.title ??
      this.state.index?.timeline_sessions?.find(
        (s) => s.viewer_session_id === id,
      )?.title ??
      id;
    let committed = false;
    this.timelineMutation = mutation;
    this.readSeq++;
    this.update({ writing: true, writingId: id });
    const abort = new AbortController();
    const deadline = setTimeout(() => abort.abort(), 15000);
    try {
      const result = parseTimelineOperationResult(
        await this.client.request(
          `/api/timeline/sessions/${id}`,
          method,
          mutation.body,
          abort.signal,
          true,
        ),
      );
      if (epoch !== this.epoch) throw new Error("Access changed");
      if (
        result.viewer_session_id !== id ||
        result.operation_id !== mutation.body.operation_id ||
        result.kind !==
          (method === "PATCH" ? "session_patch" : "session_delete")
      )
        throw new Error("Invalid mutation receipt");
      committed = true;
      if (method === "DELETE") {
        this.pendingTimelineDeleteMutation = undefined;
        this.update({ pendingTimelineDelete: undefined });
      }
      if (method === "DELETE" || patch?.archived !== undefined) {
        this.temporaryTargetID = undefined;
        this.update({ openTarget: undefined, targetError: false });
      }
      this.readSeq++;
      if (!(await this.refresh())) throw new Error("Refresh required");
      if (epoch !== this.epoch) throw new Error("Access changed");
      if (method === "DELETE") {
        if (
          this.state.index?.timeline_sessions?.some(
            (s) => s.viewer_session_id === id,
          )
        )
          throw new Error("Refresh required");
        this.update({ notice: "deleted" });
      } else if (this.state.selectedId === id) {
        await this.timelineTask;
        if (
          this.state.timeline?.id !== id ||
          !this.state.timeline.data ||
          this.state.timeline.data.session.version < result.session_version
        )
          throw new Error("Refresh required");
      }
      if (patch?.archived === true)
        this.update({
          notice: "archived",
          undoArchive: { id, expectedVersion: result.session_version },
        });
      this.timelineMutation = undefined;
    } catch (error) {
      if (epoch === this.epoch) {
        if (this.failure(error)) {
          // Access reset clears any recovery operation owned by this epoch.
        } else if (
          error instanceof HttpError &&
          error.status >= 400 &&
          error.status < 500 &&
          error.status !== 408 &&
          error.status !== 429 &&
          error.details.code !== "delete_pending"
        ) {
          this.timelineMutation = undefined;
          if (method === "DELETE") {
            this.pendingTimelineDeleteMutation = undefined;
            this.update({ pendingTimelineDelete: undefined });
          }
          await this.refresh();
        } else {
          if (method === "DELETE" && !committed) {
            this.pendingTimelineDeleteMutation = mutation;
            this.update({
              pendingTimelineDelete: {
                id,
                title: deleteTitle,
                expectedVersion,
              },
            });
          }
          await this.refresh();
        }
      }
      throw error;
    } finally {
      clearTimeout(deadline);
      if (epoch === this.epoch)
        this.update({ writing: false, writingId: undefined });
    }
  }
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
      void this.open(takeOpeningCode(), takeTarget());
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
