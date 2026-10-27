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
import {
  Dialog,
  DialogCancel,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { mapIndex } from "./api";
import type { DocumentData, ViewerController } from "./controller";
import { SessionDeliveries } from "./deliveries";
import {
  loadWalkthroughRenderer,
  WalkthroughDocument,
  type WalkthroughModule,
} from "./document";
import { messages as t } from "./messages";
import { sessionIdentity } from "./sessions";
import { ViewerShell } from "./shell";

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
  const [deleteTarget, setDeleteTarget] = useState<{
    id: string;
    groupId: string;
    title: string;
    count: number;
  }>();
  const [renameTarget, setRenameTarget] = useState<{
    id: string;
    title: string;
  }>();
  const [name, setName] = useState("");
  const [nameError, setNameError] = useState(false);
  const [checkingName, setCheckingName] = useState(false);
  const editButton = useRef<HTMLElement | null>(null);
  const deleteCancel = useRef<HTMLButtonElement>(null);
  const deleteProgress = useRef<HTMLParagraphElement>(null);
  const wasDeleting = useRef(false);
  const nameInput = useRef<HTMLInputElement>(null);
  const savingNameStatus = useRef<HTMLParagraphElement>(null);
  const wasSavingName = useRef(false);
  useLayoutEffect(() => {
    const saving = Boolean(
      renameTarget && state.writing && state.writingId === renameTarget.id,
    );
    if (saving) savingNameStatus.current?.focus();
    else if (
      wasSavingName.current &&
      renameTarget &&
      state.actionError?.operation === "title"
    )
      nameInput.current?.focus();
    wasSavingName.current = saving;
  }, [renameTarget, state.writing, state.writingId, state.actionError]);
  const undoButton = useRef<HTMLButtonElement>(null);
  const allSessions = state.index ? mapIndex(state.index) : [];
  const deleteMembers = deleteTarget
    ? allSessions.filter((s) => sessionIdentity(s) === deleteTarget.groupId)
    : [];
  const deleteError =
    state.actionError?.operation === "delete" &&
    state.actionError.groupId === deleteTarget?.groupId
      ? state.actionError
      : undefined;
  useLayoutEffect(() => {
    const pending = deleting && state.writing;
    if (pending) deleteProgress.current?.focus();
    else if (wasDeleting.current && deleting && deleteError)
      deleteCancel.current?.focus();
    wasDeleting.current = pending;
  }, [deleting, state.writing, deleteError]);
  const savedNameSession = allSessions.find(
    (session) => session.id === renameTarget?.id,
  );
  const currentSavedName =
    state.groups.find((group) =>
      group.segments.some((segment) => segment.id === renameTarget?.id),
    )?.title ??
    savedNameSession?.raw.session_title ??
    savedNameSession?.title;
  const checkSavedName = async () => {
    if (checkingName) return;
    setCheckingName(true);
    try {
      await controller.refresh();
    } finally {
      setCheckingName(false);
    }
  };
  const openDelete = () => {
    if (!selected) return;
    const groupId = sessionIdentity(selected);
    setDeleteTarget({
      id: selected.id,
      groupId,
      title: group?.title ?? selected.raw.session_title ?? selected.title,
      count: allSessions.filter((s) => sessionIdentity(s) === groupId).length,
    });
    setDeleting(true);
  };
  useEffect(() => {
    if (!state.undoArchive) return;
    let timer: ReturnType<typeof setTimeout>;
    const expire = () => {
      if (document.activeElement === undoButton.current)
        timer = setTimeout(expire, 1000);
      else controller.clearUndo();
    };
    timer = setTimeout(expire, 10000);
    return () => clearTimeout(timer);
  }, [state.undoArchive, controller]);
  const saveName = async (value: string) => {
    if (/\p{Cc}/u.test(value) || [...value.trim()].length > 200) {
      setNameError(true);
      return;
    }
    setNameError(false);
    if (state.actionError?.code === "reload_required") return;
    if (await controller.rename(value.trim(), renameTarget?.id))
      setRenameTarget(undefined);
  };
  useEffect(() => {
    if (
      deleting &&
      deleteTarget &&
      !state.writing &&
      !deleteError &&
      !allSessions.some((s) => sessionIdentity(s) === deleteTarget.groupId)
    )
      setDeleting(false);
  }, [deleting, deleteTarget, state.writing, deleteError, allSessions]);
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
      localTitle: Boolean(group.segments[0].raw.session_title),
      ticket:
        group.tickets.length === 1
          ? group.tickets[0]
          : `${group.tickets.length} ${t.tickets}`,
      segments: group.segments,
      status: group.inProgress
        ? ("in_progress" as const)
        : ("unknown" as const),
    })),
    ...state.sessions
      .filter(
        (segment) => segment.excluded || segment.classification === "tracked",
      )
      .map((segment) => ({
        ...segment,
        title: segment.raw.session_title || segment.title,
        localTitle: Boolean(segment.raw.session_title),
      })),
  ];
  useEffect(() => controller.start(), [controller]);
  useEffect(() => {
    document.title = selected
      ? `${group?.title ?? selected.raw.session_title ?? selected.title} · Hive`
      : "Hive · Walkthroughs";
  }, [selected, group?.title]);
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
      {state.notice &&
        !state.actionError &&
        !state.undoArchive &&
        state.notice !== "archived" &&
        state.notice !== "unarchived" && (
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
        onArchive={selected ? () => void controller.archive() : undefined}
        onRename={
          selected
            ? () => {
                editButton.current = document.activeElement as HTMLElement;
                setName(
                  group?.title ?? selected.raw.session_title ?? selected.title,
                );
                setNameError(false);
                setRenameTarget({
                  id: selected.id,
                  title: group?.title ?? selected.title,
                });
              }
            : undefined
        }
        feedback={
          (state.undoArchive ||
            state.actionError ||
            state.index?.errors.some(
              (error) => error.code === "delete_cleanup_pending",
            )) && (
            <>
              {state.undoArchive && (
                <div
                  role="status"
                  className="flex flex-wrap items-center gap-2"
                >
                  <span>
                    {state.undoArchive.archived ? t.unarchived : t.archived}
                  </span>
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    disabled={state.writing}
                    ref={undoButton}
                    onClick={() => void controller.undoArchive()}
                  >
                    {t.undo}
                  </Button>
                </div>
              )}
              {state.actionError && (
                <Alert variant="destructive" role="alert">
                  <AlertTitle>
                    {state.actionError.code === "rename_timeout"
                      ? t.nameSaveTimeout
                      : state.actionError.code === "reload_required"
                        ? t.reloadRequired
                        : state.actionError.remaining === 0 &&
                            state.actionError.cleanup_pending
                          ? t.cleanupPending
                          : state.actionError.code === "partial_delete"
                            ? t.partialDelete
                            : t.writeError}
                  </AlertTitle>
                  {state.actionError.code === "partial_delete" && (
                    <AlertDescription>
                      {t.deletedCount}: {state.actionError.deleted ?? 0} ·{" "}
                      {t.remainingCount}: {state.actionError.remaining ?? 0}
                    </AlertDescription>
                  )}
                  <AlertAction>
                    <Button
                      type="button"
                      size="sm"
                      variant="outline"
                      disabled={state.writing}
                      onClick={() => {
                        if (state.actionError?.code === "reload_required")
                          window.location.reload();
                        else if (
                          state.actionError?.operation === "delete" &&
                          state.actionError.remaining !== 0
                        )
                          setDeleting(true);
                        else controller.retryAction();
                      }}
                    >
                      {state.actionError.code === "rename_timeout"
                        ? t.checkSavedName
                        : state.actionError.code === "reload_required"
                          ? t.reload
                          : state.actionError.remaining === 0
                            ? t.checkRecovery
                            : t.retry}
                    </Button>
                  </AlertAction>
                </Alert>
              )}
              {state.index?.errors
                .filter(
                  (error) =>
                    error.code === "delete_cleanup_pending" &&
                    !state.actionError?.cleanupGroups?.includes(
                      error.session_group ?? error.path,
                    ),
                )
                .map((error) => (
                  <Alert
                    key={error.session_group ?? error.path}
                    variant="destructive"
                    role="status"
                  >
                    <AlertTitle>{t.cleanupPending}</AlertTitle>
                    <AlertAction>
                      <Button
                        type="button"
                        size="sm"
                        variant="outline"
                        onClick={retry}
                      >
                        {t.checkRecovery}
                      </Button>
                    </AlertAction>
                  </Alert>
                ))}
            </>
          )
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
        writing={state.writing || state.indexError}
        onPin={selected ? () => void controller.pin() : undefined}
        onDelete={selected ? openDelete : undefined}
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
                    <Badge
                      variant={
                        segment.status === "in_progress"
                          ? "info-light"
                          : segment.status === "closed"
                            ? "success-light"
                            : "outline"
                      }
                    >
                      {t.status[segment.status ?? "unknown"]}
                    </Badge>
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
          initialFocus={() => deleteCancel.current}
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
          <section
            aria-label={t.sessionName}
            aria-describedby="delete-session-name-reading-hint"
            style={{
              whiteSpace: /\s/u.test(deleteTarget?.title ?? "")
                ? "normal"
                : "nowrap",
            }}
            // biome-ignore lint/a11y/noNoninteractiveTabindex: The name is a deliberate keyboard-accessible horizontal reading region.
            tabIndex={0}
            className="min-w-0 overflow-x-auto whitespace-normal break-normal text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
          >
            {deleteTarget?.title}
          </section>
          <span id="delete-session-name-reading-hint" className="sr-only">
            {t.nameReadingHint}
          </span>
          <p
            ref={deleteProgress}
            tabIndex={-1}
            className="text-sm focus:outline-2 focus:outline-offset-2 focus:outline-ring"
          >
            {t.storedCount}:{" "}
            {deleteMembers.length ||
              (deleteError?.remaining === 0 ? 0 : deleteTarget?.count)}
          </p>
          {deleteError && (
            <p role="alert" className="text-sm">
              {deleteError.code === "reload_required"
                ? t.reloadRequired
                : deleteError.remaining === 0 && deleteError.cleanup_pending
                  ? t.cleanupPending
                  : t.writeError}{" "}
              {deleteError.code === "partial_delete"
                ? `${t.deletedCount}: ${deleteError.deleted ?? 0} · ${t.remainingCount}: ${deleteError.remaining ?? 0}`
                : ""}
            </p>
          )}
          <AlertDialogFooter>
            <AlertDialogCancel
              ref={deleteCancel}
              type="button"
              disabled={state.writing}
            >
              {t.cancel}
            </AlertDialogCancel>
            <AlertDialogAction
              type="button"
              disabled={
                state.writing ||
                state.indexError ||
                (deleteMembers.length === 0 && deleteError?.remaining !== 0)
              }
              onClick={() => {
                if (deleteError?.code === "reload_required") {
                  window.location.reload();
                  return;
                }
                if (deleteError?.remaining === 0) {
                  void controller.refresh();
                  return;
                }
                const member = deleteMembers[0];
                if (member)
                  void controller.remove(member.id).then((ok) => {
                    if (ok) setDeleting(false);
                  });
              }}
            >
              {deleteError?.code === "reload_required"
                ? t.reload
                : deleteError?.remaining === 0
                  ? t.checkRecovery
                  : deleteError
                    ? t.retry
                    : t.remove}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
      <Dialog
        open={Boolean(renameTarget) && state.access === "ready"}
        onOpenChange={(open) => {
          if (!open && !state.writing) setRenameTarget(undefined);
        }}
      >
        <DialogContent
          finalFocus={() =>
            editButton.current?.isConnected
              ? editButton.current
              : document.getElementById("document")
          }
        >
          <DialogHeader>
            <DialogTitle>{t.sessionName}</DialogTitle>
            <DialogDescription>{t.nameHelp}</DialogDescription>
          </DialogHeader>
          <form
            className="flex flex-col gap-4"
            aria-busy={state.writing}
            onSubmit={(event) => {
              event.preventDefault();
              void saveName(name);
            }}
          >
            <label htmlFor="session-name" className="text-sm">
              {t.sessionName}
            </label>
            <Input
              id="session-name"
              ref={nameInput}
              value={name}
              disabled={state.writing}
              aria-invalid={nameError}
              aria-describedby={nameError ? "session-name-error" : undefined}
              onChange={(event) => {
                setName(event.target.value);
                setNameError(false);
              }}
            />
            {state.writing && (
              <p
                role="status"
                aria-live="polite"
                tabIndex={-1}
                ref={savingNameStatus}
                className="text-sm focus:outline-2 focus:outline-offset-2 focus:outline-ring"
              >
                {t.savingName}
              </p>
            )}
            {nameError && (
              <p id="session-name-error" role="alert" className="text-sm">
                {t.invalidName}
              </p>
            )}
            {state.actionError?.operation === "title" && (
              <p role="alert" className="text-sm">
                {state.actionError.code === "rename_timeout"
                  ? t.nameSaveTimeout
                  : state.actionError.code === "reload_required"
                    ? t.reloadRequired
                    : t.writeError}
              </p>
            )}
            {state.actionError?.operation === "title" &&
              state.actionError.code === "rename_timeout" && (
                <div className="flex min-w-0 flex-col gap-2">
                  <p className="text-xs text-muted-foreground">
                    {t.lastLoadedSavedName}
                  </p>
                  <section
                    aria-label={t.lastLoadedSavedName}
                    aria-describedby="saved-session-name-reading-hint"
                    aria-live="polite"
                    // biome-ignore lint/a11y/noNoninteractiveTabindex: The saved name is a deliberate keyboard-accessible horizontal reading region.
                    tabIndex={0}
                    style={{
                      whiteSpace: /\s/u.test(currentSavedName ?? "")
                        ? "normal"
                        : "nowrap",
                    }}
                    className="min-w-0 overflow-x-auto text-sm focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
                  >
                    {currentSavedName ?? t.savedNameUnavailable}
                  </section>
                  <span
                    id="saved-session-name-reading-hint"
                    className="sr-only"
                  >
                    {t.nameReadingHint}
                  </span>
                  {checkingName && (
                    <p role="status" aria-live="polite" className="text-sm">
                      {t.checkingSavedName}
                    </p>
                  )}
                  {state.indexError && !checkingName && (
                    <p role="alert" className="text-sm">
                      {t.nameCheckFailed}
                    </p>
                  )}
                  <Button
                    type="button"
                    size="sm"
                    variant="outline"
                    aria-disabled={checkingName}
                    aria-busy={checkingName}
                    onClick={() => void checkSavedName()}
                  >
                    {t.checkSavedName}
                  </Button>
                </div>
              )}
            <Button
              type="button"
              size="sm"
              variant="outline"
              disabled={
                state.writing ||
                state.indexError ||
                state.actionError?.code === "reload_required"
              }
              onClick={() => void saveName("")}
            >
              {t.resetName}
            </Button>
            <DialogFooter>
              <DialogCancel type="button" disabled={state.writing}>
                {t.cancel}
              </DialogCancel>
              {state.actionError?.code === "reload_required" && (
                <Button
                  type="button"
                  size="sm"
                  variant="outline"
                  onClick={() => window.location.reload()}
                >
                  {t.reload}
                </Button>
              )}
              <Button
                type="submit"
                size="sm"
                disabled={
                  state.writing ||
                  state.indexError ||
                  state.actionError?.code === "reload_required"
                }
              >
                {t.saveName}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </>
  );
}
