import {
  type ReactNode,
  useEffect,
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
import { Skeleton } from "@/components/ui/skeleton";
import type { ViewerController } from "./controller";
import { WalkthroughDocument } from "./document";
import { messages as t } from "./messages";
import { ViewerShell } from "./shell";

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
}: {
  controller: ViewerController;
}) {
  const state = useSyncExternalStore(
    controller.subscribe,
    controller.getSnapshot,
  );
  const [deleting, setDeleting] = useState(false);
  const deleteKey = useRef<string | undefined>(undefined);
  const selected = state.sessions.find((s) => s.id === state.selectedId);
  useEffect(() => controller.start(), [controller]);
  useEffect(() => {
    document.title = selected
      ? `${selected.title} · Hive`
      : "Hive · Walkthroughs";
  }, [selected]);
  useEffect(() => {
    if (state.notice === "deleted") {
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
  else if (state.documentState === "excluded")
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
      {selected &&
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
        sessions={state.sessions}
        coldLoading={
          state.access === "pending" ||
          (state.access === "ready" && !state.index && !state.indexError)
        }
        selectedId={state.selectedId}
        onSelect={controller.select}
        showArchived={state.showArchived}
        onShowArchivedChange={controller.setShowArchived}
        onArchive={selected ? () => void controller.archive() : undefined}
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
        panel={panel}
        notices={notices}
        libraryFooter={libraryFooter}
        writing={state.writing}
        onPin={selected ? () => void controller.pin() : undefined}
        onDelete={
          selected
            ? () => {
                deleteKey.current = selected.id;
                setDeleting(true);
              }
            : undefined
        }
      >
        {state.documentState === "loading" && !state.document && (
          <div aria-busy="true" role="status" className="flex flex-col gap-4">
            <span className="sr-only">{t.loading}</span>
            <Skeleton className="h-8 w-2/3" />
            <Skeleton className="h-48 w-full" />
          </div>
        )}
        {state.documentState === "error" && (
          <Alert variant="destructive">
            <AlertTitle>{t.loadFailed}</AlertTitle>
            <AlertAction>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={controller.retryDocument}
              >
                {t.retry}
              </Button>
            </AlertAction>
          </Alert>
        )}
        {state.document && (
          <WalkthroughDocument
            data={state.document}
            onError={controller.rendererFailed}
          />
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
              disabled={state.writing || state.selectedId !== deleteKey.current}
              onClick={() => {
                void controller.remove().then(() => setDeleting(false));
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
