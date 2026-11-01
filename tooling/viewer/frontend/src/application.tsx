import {
  type ReactNode,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
  useSyncExternalStore,
} from "react";
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@/components/reui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import { mapTimelineIndex, optionalStorage, save, stored } from "./api";
import type { ViewerController } from "./controller";
import { messages as t } from "./messages";
import { ViewerShell } from "./shell";
import { TimelineContent } from "./timeline-content";
import { TimelineHeader } from "./timeline-header";
import type { TimelineDetailOpener } from "./timeline-tasks";

export function ViewerApplication({
  controller,
  onOpenDetail,
}: {
  controller: ViewerController;
  onOpenDetail?: TimelineDetailOpener;
}) {
  const [undoError, setUndoError] = useState(false);
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const read = state.timeline?.id === state.selectedId ? state.timeline : null,
    data = read?.data;
  const selected = state.timelineSessions.find(
    (s) => s.viewer_session_id === state.selectedId,
  );
  const lang = navigator.language.startsWith("es") ? "es" : "en";
  const scrolls = useRef(new Map<string, number>()),
    last = useRef<string | undefined>(undefined);
  const focusedTarget = useRef<object | undefined>(undefined);
  const loadDocument = useCallback(
    (id: string) =>
      controller.loadTimelineDocument(selected?.viewer_session_id ?? "", id),
    [controller, selected?.viewer_session_id],
  );
  useEffect(() => controller.start(), [controller]);
  useEffect(() => {
    document.title = selected
      ? `${selected.title} · Hive`
      : "Hive · Walkthroughs";
  }, [selected]);
  useLayoutEffect(() => {
    const canvas = document.querySelector<HTMLElement>(
      '[data-slot="viewer-canvas"]',
    );
    if (!canvas) return;
    if (last.current !== state.selectedId) {
      last.current = state.selectedId;
      canvas.scrollTop =
        scrolls.current.get(state.selectedId ?? "") ??
        Number(
          stored(optionalStorage(), `hive-view-scroll:${state.selectedId}`),
        );
    }
    const id = state.selectedId;
    const remember = () => {
      if (id) {
        scrolls.current.set(id, canvas.scrollTop);
        save(
          optionalStorage(),
          `hive-view-scroll:${id}`,
          String(canvas.scrollTop),
        );
      }
    };
    canvas.addEventListener("scroll", remember);
    return () => {
      remember();
      canvas.removeEventListener("scroll", remember);
    };
  }, [state.selectedId]);
  useLayoutEffect(() => {
    const canvas = document.querySelector<HTMLElement>(
      '[data-slot="viewer-canvas"]',
    );
    if (canvas && state.selectedId) {
      canvas.scrollTop =
        scrolls.current.get(state.selectedId) ??
        Number(
          stored(optionalStorage(), `hive-view-scroll:${state.selectedId}`),
        );
    }
  }, [state.selectedId]);
  useEffect(() => {
    const target = state.openTarget;
    if (
      !data ||
      state.targetError ||
      !target ||
      !("sessionID" in target) ||
      !target.eventID ||
      focusedTarget.current === target
    )
      return;
    const frame = requestAnimationFrame(() => {
      const node = document.getElementById(`timeline-event-${target.eventID}`);
      const canvas = node?.closest<HTMLElement>('[data-slot="viewer-canvas"]');
      if (!node || !canvas) return;
      canvas.scrollTop +=
        node.getBoundingClientRect().top - canvas.getBoundingClientRect().top;
      node.tabIndex = -1;
      node.focus({ preventScroll: true });
      focusedTarget.current = target;
    });
    return () => cancelAnimationFrame(frame);
  }, [data, state.openTarget, state.targetError]);
  const retry = () => void controller.refresh();
  let panel: ReactNode;
  if (state.access === "pending") panel = <p role="status">{t.pending}</p>;
  else if (state.access !== "ready")
    panel = (
      <div
        id="viewer-access"
        role="alert"
        tabIndex={-1}
        className="flex flex-col gap-3"
      >
        <h2 className="font-medium">{t.access}</h2>
        <p>{state.access === "rejected" ? t.rejected : t.missing}</p>
        <code>hive view open</code>
      </div>
    );
  else if (state.targetError)
    panel = (
      <Alert variant="destructive">
        <AlertTitle>{t.loadFailed}</AlertTitle>
        <AlertAction>
          <Button
            type="button"
            variant="outline"
            onClick={controller.retryTimeline}
          >
            {t.retry}
          </Button>
        </AlertAction>
      </Alert>
    );
  else if (!state.index && state.indexError)
    panel = (
      <Alert variant="destructive">
        <AlertTitle>{t.indexFailure}</AlertTitle>
        <AlertAction>
          <Button type="button" variant="outline" onClick={retry}>
            {t.retry}
          </Button>
        </AlertAction>
      </Alert>
    );
  else if (state.index && !state.timelineSessions.length)
    panel = (
      <div role="status" className="flex flex-col items-start gap-3">
        <h2>
          {state.index.timeline_sessions.some((s) => s.archived)
            ? t.archivedOnly
            : t.empty}
        </h2>
        <p>
          {state.index.repositories.length
            ? lang === "es"
              ? "Publica un trabajo mediante la API v2 para verlo aquí."
              : "Publish a work through the v2 API to see it here."
            : t.register}
        </p>
        {!state.index.repositories.length && <code>hive view register</code>}
        {state.index.timeline_sessions.some((s) => s.archived) && (
          <Button
            type="button"
            variant="outline"
            onClick={() => controller.setShowArchived(true)}
          >
            {t.showArchived}
          </Button>
        )}
        <Button type="button" variant="outline" onClick={retry}>
          {t.retry}
        </Button>
      </div>
    );
  const notices = (
    <>
      {state.indexError && state.index && (
        <Alert variant="destructive">
          <AlertDescription>{t.indexFailure}</AlertDescription>
          <AlertAction>
            <Button type="button" variant="outline" onClick={retry}>
              {t.retry}
            </Button>
          </AlertAction>
        </Alert>
      )}
      {state.pendingTimelineDelete && (
        <Alert variant="warning" role="status">
          <AlertTitle>{t.partialDelete}</AlertTitle>
          <AlertDescription>
            {state.pendingTimelineDelete.title}
          </AlertDescription>
          <AlertAction>
            <Button
              type="button"
              variant="outline"
              disabled={state.writing}
              onClick={() =>
                void controller.retryTimelineDelete().catch(() => {})
              }
            >
              {t.retry}
            </Button>
          </AlertAction>
        </Alert>
      )}
      {state.undoArchive && (
        <Alert variant="info">
          <AlertDescription>
            {undoError ? t.writeError : t.archived}
          </AlertDescription>
          <AlertAction>
            <Button
              type="button"
              variant="outline"
              onClick={() =>
                void controller
                  .undoArchive()
                  .then(() => setUndoError(false))
                  .catch(() => setUndoError(true))
              }
            >
              {t.undo}
            </Button>
          </AlertAction>
        </Alert>
      )}
    </>
  );
  return (
    <ViewerShell
      sessions={state.timelineSessions.map((s) => mapTimelineIndex(s))}
      selectedId={state.selectedId}
      state={
        state.access !== "ready"
          ? "ready"
          : !state.index && !state.indexError
            ? "loading"
            : "ready"
      }
      onSelect={controller.select}
      onRetry={retry}
      showArchived={state.showArchived}
      onShowArchivedChange={controller.setShowArchived}
      panel={panel}
      notices={notices}
      coldLoading={
        state.access === "ready" && !state.index && !state.indexError
      }
      sessionHeader={
        data && selected ? (
          <TimelineHeader
            key={selected.viewer_session_id}
            data={data}
            lang={lang}
            onRetry={controller.retryTimeline}
            onUpdate={(patch, version) =>
              controller.updateTimelineSession(
                selected.viewer_session_id,
                patch,
                version,
              )
            }
            onDelete={(version) =>
              controller.deleteTimelineSession(
                selected.viewer_session_id,
                version,
              )
            }
            onRecover={(request, reason, version) =>
              controller.recoverTimelineLead(
                selected.viewer_session_id,
                request,
                reason,
                version,
              )
            }
          />
        ) : undefined
      }
    >
      {data ? (
        <TimelineContent
          key={data.session.viewer_session_id}
          data={data}
          lang={lang}
          loadDocument={loadDocument}
          onRetry={controller.retryTimeline}
          loading={read?.state === "loading" && !data}
          error={read?.state === "error"}
          targetEventID={
            state.openTarget && "sessionID" in state.openTarget
              ? state.openTarget.eventID
              : undefined
          }
          onOpenDetail={onOpenDetail}
        />
      ) : read?.state === "error" ? (
        <Alert variant="destructive">
          <AlertTitle>{t.loadFailed}</AlertTitle>
          <AlertAction>
            <Button
              type="button"
              variant="outline"
              onClick={controller.retryTimeline}
            >
              {t.retry}
            </Button>
          </AlertAction>
        </Alert>
      ) : selected ? (
        <div role="status" className="flex flex-col gap-4">
          <Skeleton className="h-4 w-2/3" />
          <Skeleton className="h-32 w-full" />
        </div>
      ) : null}
    </ViewerShell>
  );
}
