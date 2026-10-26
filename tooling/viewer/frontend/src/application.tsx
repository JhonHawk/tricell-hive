import { Archive, ArchiveRestore, Pin, Trash2 } from "lucide-react";
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
import { Badge } from "@/components/reui/badge";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Button } from "@/components/ui/button";
import { ButtonGroup, ButtonGroupText } from "@/components/ui/button-group";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import type { DocumentData, ViewerController } from "./controller";
import { SessionDeliveries } from "./deliveries";
import {
  loadWalkthroughRenderer,
  WalkthroughDocument,
  type WalkthroughModule,
} from "./document";
import { messages as t } from "./messages";
import { ActionButton, ViewerShell } from "./shell";

function SegmentDocument({
  data,
  controller,
  onReady,
  loadModule,
}: {
  data: DocumentData;
  controller: ViewerController;
  onReady: (data: DocumentData) => void;
  loadModule?: () => Promise<WalkthroughModule>;
}) {
  const onError = useCallback(
    () => controller.rendererFailed(data.key),
    [controller, data.key],
  );
  const loadRenderer = useCallback(
    () =>
      loadModule
        ? loadModule().then((module) => module.renderWalkthrough)
        : loadWalkthroughRenderer(),
    [loadModule],
  );
  return (
    <WalkthroughDocument
      data={data}
      onError={onError}
      onReady={onReady}
      loadRenderer={loadRenderer}
    />
  );
}
function Command({ command }: { command: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <div className="flex flex-wrap items-center gap-2">
      <code className="select-all text-sm">{command}</code>
      <Button
        type="button"
        variant="outline"
        size="sm"
        onClick={() => {
          void navigator.clipboard
            ?.writeText(command)
            .then(() => setCopied(true))
            .catch(() => {});
        }}
      >
        {copied ? t.copied : t.copy}
      </Button>
    </div>
  );
}
export function ViewerApplication({
  controller,
  loadModule,
}: {
  controller: ViewerController;
  loadModule?: () => Promise<WalkthroughModule>;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const [deleting, setDeleting] = useState(false);
  const deleteKey = useRef<string | undefined>(undefined);
  const selected = state.sessions.find((s) => s.id === state.selectedId);
  const group = state.groups.find((group) =>
    group.segments.some((segment) => segment.id === state.selectedId),
  );
  const segmentRoots = useRef(new Map<string, HTMLElement>());
  const [renderedDocuments, setRenderedDocuments] = useState<
    Record<string, DocumentData>
  >({});
  const documentReady = useCallback((data: DocumentData) => {
    setRenderedDocuments((previous) =>
      previous[data.key] === data
        ? previous
        : { ...previous, [data.key]: data },
    );
  }, []);
  const pendingJump = useRef<{ groupId: string; key: string } | null>(null);
  const previousGroup = useRef<string | undefined>(undefined);
  const jump = (key: string) => {
    const target = state.groups.find((group) =>
      group.segments.some((segment) => segment.id === key),
    );
    if (target) pendingJump.current = { groupId: target.id, key };
    controller.select(key);
  };
  const activeGroup = group?.id ?? selected?.id;
  useLayoutEffect(() => {
    if (previousGroup.current === activeGroup) return;
    previousGroup.current = activeGroup;
    setRenderedDocuments({});
    const requested = pendingJump.current;
    const continuation =
      group &&
      group.segments.findIndex((segment) => segment.id === state.selectedId) >
        0;
    pendingJump.current =
      requested?.groupId === group?.id && requested?.key === state.selectedId
        ? requested
        : group && state.selectedId && continuation
          ? { groupId: group.id, key: state.selectedId }
          : null;
    const canvas = document.querySelector('[data-slot="viewer-canvas"]');
    if (canvas) canvas.scrollTop = 0;
  }, [activeGroup, group, state.selectedId]);
  useEffect(() => {
    const target = pendingJump.current;
    if (
      !target ||
      !group ||
      target.groupId !== group.id ||
      target.key !== state.selectedId
    )
      return;
    const index = group.segments.findIndex(
      (segment) => segment.id === target.key,
    );
    if (index < 0) return;
    const settled = group.segments.slice(0, index + 1).every((segment) => {
      const entry = state.documents[segment.id];
      return (
        entry &&
        (entry.state === "error" ||
          entry.state === "excluded" ||
          (entry.state === "ready" &&
            entry.document === renderedDocuments[segment.id]))
      );
    });
    if (!settled) return;
    const frame = requestAnimationFrame(() => {
      if (pendingJump.current !== target) return;
      const heading = segmentRoots.current.get(target.key);
      if (!heading) return;
      pendingJump.current = null;
      const section = heading.closest<HTMLElement>("[data-segment-key]");
      const canvas = section?.closest<HTMLElement>(
        '[data-slot="viewer-canvas"]',
      );
      if (section && canvas)
        canvas.scrollTop +=
          section.getBoundingClientRect().top -
          canvas.getBoundingClientRect().top;
      heading.focus({ preventScroll: true });
    });
    return () => cancelAnimationFrame(frame);
  }, [group, state.documents, state.selectedId, renderedDocuments]);
  const navigation = [
    ...state.groups.map((group) => ({
      ...group.segments[0],
      title: group.title,
      ticket:
        group.tickets.length === 1
          ? group.tickets[0]
          : `${group.tickets.length} ${t.tickets}`,
      segments: group.segments,
      status: group.inProgress
        ? ("in_progress" as const)
        : group.segments.every((segment) => segment.status === "closed")
          ? ("closed" as const)
          : ("unknown" as const),
    })),
    ...state.sessions.filter(
      (segment) => segment.excluded || segment.classification === "tracked",
    ),
  ];
  useEffect(() => controller.start(), [controller]);
  useEffect(() => {
    document.title = selected
      ? `${selected.title} · Hive`
      : "Hive · Walkthroughs";
  }, [selected]);
  useEffect(() => {
    if (
      state.notice === "deleted" ||
      (state.notice === "archived" && document.activeElement === document.body)
    ) {
      const heading = segmentRoots.current.get(state.selectedId ?? "");
      if (heading) {
        heading.focus({ preventScroll: true });
        return;
      }
      const target = [
        ...document.querySelectorAll<HTMLElement>("[data-session-key]"),
      ].find((node) => node.dataset.sessionKey === state.selectedId);
      if (target?.getClientRects().length)
        target.focus({ preventScroll: true });
      else document.getElementById("document")?.focus({ preventScroll: true });
    }
  }, [state.notice, state.selectedId]);
  useEffect(() => {
    if (state.access === "missing" || state.access === "rejected")
      document.getElementById("viewer-access")?.focus({ preventScroll: true });
  }, [state.access]);
  const retry = () => void controller.refresh();
  let panel: ReactNode;
  if (state.access === "pending") panel = <p role="status">{t.pending}</p>;
  else if (state.access !== "ready")
    panel = (
      <div
        id="viewer-access"
        tabIndex={-1}
        role="alert"
        className="flex flex-col gap-3"
      >
        <h2 className="font-medium">{t.access}</h2>
        <p>{state.access === "rejected" ? t.rejected : t.missing}</p>
        <Command command="hive view open" />
      </div>
    );
  else if (
    !state.sessions.length &&
    state.index &&
    !state.showArchived &&
    (state.index.projects.some((project) =>
      project.tickets.some((ticket) =>
        ticket.sessions.some((session) => session.archived),
      ),
    ) ||
      state.index.excluded_sessions?.some((session) => session.archived))
  )
    panel = (
      <div role="status" className="flex flex-col items-start gap-3">
        <h2 className="font-medium">{t.archivedOnly}</h2>
        <p className="text-muted-foreground">{t.archivedOnlyBody}</p>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => controller.setShowArchived(true)}
        >
          {t.showArchived}
        </Button>
      </div>
    );
  else if (!state.sessions.length && state.index)
    panel = (
      <div role="status" className="flex flex-col items-start gap-3">
        <h2 className="font-medium">
          {state.index.repositories.length ? t.empty : t.noRepositories}
        </h2>
        <p className="text-muted-foreground">
          {state.index.repositories.length ? t.emptyBody : t.register}
        </p>
        <Command command="hive view register" />
        {state.index.repositories.length > 0 && (
          <>
            <h3 className="font-medium">{t.repositories}</h3>
            <ul>
              {state.index.repositories.map((repo) => (
                <li key={repo.root}>
                  <code>{repo.root}</code>
                  {repo.absent && (
                    <p className="text-muted-foreground">{t.absent}</p>
                  )}
                </li>
              ))}
            </ul>
          </>
        )}
        <Button type="button" size="sm" variant="outline" onClick={retry}>
          {t.retry}
        </Button>
      </div>
    );
  else if (
    selected &&
    (selected.excluded || selected.classification === "tracked")
  )
    panel = (
      <div role="status" className="flex flex-col items-start gap-3">
        <Badge variant="outline">{t.excluded}</Badge>
        <p>{t.excludedBody}</p>
        <p className="text-xs text-muted-foreground">{t.source}</p>
        <code className="max-w-full break-all text-sm">
          {selected?.raw.source}
        </code>
        <Button
          type="button"
          size="sm"
          variant="outline"
          onClick={() => {
            controller.retryDocument();
            retry();
          }}
        >
          {t.retry}
        </Button>
      </div>
    );
  const libraryFooter = state.index && (
    <div className="flex flex-col gap-3">
      <p className="text-xs text-muted-foreground">
        {t.ttl}{" "}
        {state.index.ttl_seconds % 86400 === 0
          ? `${state.index.ttl_seconds / 86400} d`
          : `${Math.max(1, Math.round(state.index.ttl_seconds / 3600))} h`}{" "}
        {t.unlessKept}
      </p>
      {state.index.errors.length > 0 && (
        <details>
          <summary className="cursor-pointer text-xs">
            {t.ingestErrors} ({state.index.errors.length})
          </summary>
          <ul className="flex flex-col gap-2">
            {state.index.errors.map((error) => (
              <li
                key={`${error.path}:${error.reason}`}
                className="break-all text-xs"
              >
                <code>{error.path}</code>
                <p>{error.reason}</p>
              </li>
            ))}
          </ul>
        </details>
      )}
    </div>
  );
  const notices = (
    <>
      {(state.down || state.indexError) && (
        <Alert variant="destructive">
          <AlertTitle>{t.loadFailed}</AlertTitle>
          <AlertDescription>
            {state.down ? t.down : t.indexFailure}
          </AlertDescription>
          <AlertAction>
            <Button type="button" size="sm" variant="outline" onClick={retry}>
              {t.retry}
            </Button>
          </AlertAction>
        </Alert>
      )}
      {state.notice && (
        <Alert
          variant={state.notice === "writeError" ? "destructive" : "default"}
          role="status"
        >
          <AlertTitle>{t[state.notice]}</AlertTitle>
          <AlertAction>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={controller.dismiss}
            >
              {t.dismiss}
            </Button>
          </AlertAction>
        </Alert>
      )}
      {!group &&
        selected &&
        !selected.excluded &&
        selected.classification !== "tracked" && (
          <Alert
            variant={
              selected.classification === "unknown" ? "warning" : "default"
            }
            className={
              selected.classification === "local" ? "border-border" : undefined
            }
            role="status"
          >
            <AlertTitle>
              {selected.classification === "local"
                ? t.sourceLocal
                : t.sourceUnknown}
            </AlertTitle>
            {selected.classification === "unknown" && (
              <AlertDescription className="text-foreground">
                {t.sourceReasons[
                  selected.raw.source_reason as keyof typeof t.sourceReasons
                ] ?? t.sourceReasons.tracking_unavailable}
              </AlertDescription>
            )}
            {selected.absent && <AlertDescription>{t.absent}</AlertDescription>}
          </Alert>
        )}
    </>
  );
  return (
    <>
      <ViewerShell
        sessions={navigation}
        coldLoading={
          state.access === "pending" ||
          (state.access === "ready" && !state.index && !state.indexError)
        }
        selectedId={group?.segments[0].id ?? state.selectedId}
        onSelect={(key) => {
          const target = state.groups.find(
            (group) => group.segments[0].id === key,
          );
          if (target) {
            pendingJump.current = null;
            controller.selectSession(target.id);
            const canvas = document.querySelector(
              '[data-slot="viewer-canvas"]',
            );
            if (canvas) canvas.scrollTop = 0;
          } else controller.select(key);
        }}
        onJump={jump}
        showArchived={state.showArchived}
        onShowArchivedChange={controller.setShowArchived}
        onArchive={
          !group && selected ? () => void controller.archive() : undefined
        }
        state={
          state.access === "pending" ||
          (state.access === "ready" && !state.index && !state.indexError)
            ? "loading"
            : state.indexError && !state.index
              ? "error"
              : state.sessions.length
                ? "ready"
                : "empty"
        }
        onRetry={retry}
        panel={
          panel ??
          (group ? (
            <div className="flex flex-col gap-4">
              <SessionDeliveries
                group={group}
                documents={state.documents}
                onRetry={controller.retryDocument}
                loadModule={loadModule}
              />
              <nav aria-label={t.jumpTo} className="flex flex-wrap gap-2">
                {group.segments.map((segment, index) => (
                  <Button
                    key={segment.id}
                    type="button"
                    variant="outline"
                    size="sm"
                    className="min-w-0 max-w-full"
                    aria-current={
                      segment.id === state.selectedId ? "location" : undefined
                    }
                    onClick={() => jump(segment.id)}
                  >
                    <span className="min-w-0 max-w-64 truncate">
                      {index + 1} · {segment.ticket} · {segment.title}
                    </span>
                  </Button>
                ))}
              </nav>
            </div>
          ) : undefined)
        }
        notices={notices}
        libraryFooter={libraryFooter}
        writing={state.writing}
        onPin={!group && selected ? () => void controller.pin() : undefined}
        onDelete={
          !group && selected
            ? () => {
                deleteKey.current = selected.id;
                setDeleting(true);
              }
            : undefined
        }
      >
        {group && (
          <div className="flex flex-col gap-8">
            {group.segments.map((segment, index) => {
              const entry = state.documents[segment.id];
              return (
                <section
                  key={segment.id}
                  aria-labelledby={`heading-${segment.id}`}
                  className="flex min-w-0 flex-col gap-4"
                  data-segment-key={segment.id}
                >
                  {index > 0 && <Separator />}
                  <div className="flex flex-wrap items-start justify-between gap-3">
                    <div className="flex min-w-0 flex-col gap-1">
                      <p className="text-xs text-muted-foreground">
                        {index === 0 ? t.initialWork : t.continuation} ·{" "}
                        {segment.ticket} · {segment.createdAt}
                      </p>
                      <h2
                        id={`heading-${segment.id}`}
                        tabIndex={-1}
                        ref={(node) => {
                          if (node) segmentRoots.current.set(segment.id, node);
                          else segmentRoots.current.delete(segment.id);
                        }}
                        className="break-words font-semibold tracking-tight"
                      >
                        {segment.title}
                      </h2>
                    </div>
                    <ButtonGroup
                      aria-label={`${segment.title} · ${t.segmentActions}`}
                      aria-busy={state.writingId === segment.id}
                    >
                      <ButtonGroupText>
                        <span className="text-xs">
                          {t.status[segment.status ?? "unknown"]}
                        </span>
                      </ButtonGroupText>
                      <ActionButton
                        label={segment.archived ? t.unarchive : t.archive}
                        disabled={state.writing}
                        onClick={() => void controller.archive(segment.id)}
                      >
                        {segment.archived ? (
                          <ArchiveRestore aria-hidden />
                        ) : (
                          <Archive aria-hidden />
                        )}
                      </ActionButton>
                      <ActionButton
                        label={segment.pinned ? t.unpin : t.pin}
                        aria-pressed={Boolean(segment.pinned)}
                        disabled={state.writing}
                        onClick={() => void controller.pin(segment.id)}
                      >
                        <Pin aria-hidden />
                      </ActionButton>
                      <ActionButton
                        label={t.remove}
                        disabled={state.writing}
                        onClick={() => {
                          deleteKey.current = segment.id;
                          setDeleting(true);
                        }}
                      >
                        <Trash2 aria-hidden />
                      </ActionButton>
                    </ButtonGroup>
                  </div>
                  <p className="text-xs text-muted-foreground">
                    {segment.classification === "local"
                      ? t.sourceLocal
                      : t.sourceUnknown}
                    {segment.absent ? ` · ${t.absent}` : ""}
                  </p>
                  {(!entry || entry.state === "loading") && (
                    <div
                      aria-busy="true"
                      role="status"
                      className="flex flex-col gap-4"
                    >
                      <span className="sr-only">{t.loading}</span>
                      <Skeleton className="h-8 w-2/3" />
                      <Skeleton className="h-48 w-full" />
                    </div>
                  )}
                  {(entry?.state === "error" ||
                    entry?.state === "excluded") && (
                    <Alert variant="destructive">
                      <AlertTitle>
                        {entry.state === "excluded" ? t.excluded : t.loadFailed}
                      </AlertTitle>
                      <AlertAction>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={() => controller.retryDocument(segment.id)}
                        >
                          {t.retry}
                        </Button>
                      </AlertAction>
                    </Alert>
                  )}
                  {entry?.document && (
                    <SegmentDocument
                      data={entry.document}
                      controller={controller}
                      onReady={documentReady}
                      loadModule={loadModule}
                    />
                  )}
                </section>
              );
            })}
          </div>
        )}
      </ViewerShell>
      <AlertDialog
        open={deleting && state.access === "ready"}
        onOpenChange={(open) => {
          if (!state.writing) setDeleting(open);
        }}
      >
        <AlertDialogContent
          finalFocus={() => {
            if (controller.getSnapshot().notice !== "deleted") return true;
            const key = controller.getSnapshot().selectedId;
            const heading = segmentRoots.current.get(key ?? "");
            if (heading) return heading;
            const target = [
              ...document.querySelectorAll<HTMLElement>("[data-session-key]"),
            ].find((node) => node.dataset.sessionKey === key);
            return target?.getClientRects().length
              ? target
              : document.getElementById("document");
          }}
        >
          <AlertDialogHeader>
            <AlertDialogTitle>{t.confirm}</AlertDialogTitle>
            <AlertDialogDescription>{t.confirmBody}</AlertDialogDescription>
          </AlertDialogHeader>
          <p className="break-words text-sm">
            {state.sessions.find((s) => s.id === deleteKey.current)?.title}
          </p>
          <AlertDialogFooter>
            <AlertDialogCancel type="button" disabled={state.writing}>
              {t.cancel}
            </AlertDialogCancel>
            <AlertDialogAction
              type="button"
              disabled={
                state.writing ||
                !state.sessions.some(
                  (segment) => segment.id === deleteKey.current,
                )
              }
              onClick={() => {
                void controller
                  .remove(deleteKey.current)
                  .then(() => setDeleting(false));
              }}
            >
              {t.remove}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </>
  );
}
