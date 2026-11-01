import { FileText } from "lucide-react";
import { useEffect, useState } from "react";
import { Alert, AlertAction, AlertTitle } from "@/components/reui/alert";
import { Button } from "@/components/ui/button";
import { Skeleton } from "@/components/ui/skeleton";
import type {
  TimelineDocument,
  TimelineDocumentResponse,
  TimelineEvent,
  TimelineSessionResponse,
} from "./timeline-contract";
import { type TimelineLanguage, timelineCopy } from "./timeline-copy";
import { TimelineMarkdown } from "./timeline-markdown";

/** The caller owns transport, authentication and immutable byte validation. */
export type TimelineDocumentLoader = (
  documentID: string,
) => Promise<TimelineDocumentResponse>;
export type TimelineDocumentOpener = (
  eventID: string,
  trigger: HTMLElement,
  documentKey?: string,
  revisionID?: string,
) => void;
export function DocumentBody({
  document,
  loadDocument,
  lang,
  onRetry,
}: {
  document?: TimelineDocument;
  loadDocument?: TimelineDocumentLoader;
  lang: TimelineLanguage;
  onRetry?: () => void;
}) {
  const t = timelineCopy(lang);
  const [attempt, setAttempt] = useState(0);
  const [state, setState] = useState<{
    kind: "loading" | "error" | "ready";
    content?: string;
  }>({ kind: "loading" });
  useEffect(() => {
    void attempt;
    let disposed = false;
    setState({ kind: "loading" });
    if (!document || !loadDocument) return;
    void Promise.resolve()
      .then(() => loadDocument(document.document_id))
      .then((response) => {
        if (
          response.schema_version !== 3 ||
          response.document.document_id !== document.document_id ||
          response.document.session_id !== document.session_id ||
          response.document.sha256 !== document.sha256 ||
          response.document.size !== document.size ||
          response.document.media_type !== "text/markdown" ||
          typeof response.content !== "string"
        )
          throw new Error("document_metadata_mismatch");
        if (!disposed) setState({ kind: "ready", content: response.content });
      })
      .catch(() => {
        if (!disposed) setState({ kind: "error" });
      });
    return () => {
      disposed = true;
    };
  }, [document, loadDocument, attempt]);
  const retry = () => {
    setAttempt((n) => n + 1);
    onRetry?.();
  };
  if (!document || !loadDocument)
    return (
      <div role="status" className="flex flex-col items-start gap-3">
        <p className="text-sm text-muted-foreground">{t.documentUnavailable}</p>
        <Button type="button" variant="outline" size="sm" onClick={retry}>
          {t.retry}
        </Button>
      </div>
    );
  if (state.kind === "loading")
    return (
      <div
        role="status"
        aria-label={t.documentLoading}
        className="flex flex-col gap-3"
      >
        <Skeleton className="h-4 w-2/3" />
        <Skeleton className="h-32 w-full" />
      </div>
    );
  if (state.kind === "error")
    return (
      <Alert variant="destructive">
        <AlertTitle>{t.documentFailed}</AlertTitle>
        <AlertAction>
          <Button type="button" variant="outline" size="sm" onClick={retry}>
            {t.retry}
          </Button>
        </AlertAction>
      </Alert>
    );
  if (!state.content?.trim())
    return (
      <div role="status" className="flex flex-col items-start gap-3">
        <p className="text-sm text-muted-foreground">{t.documentEmpty}</p>
        <Button type="button" variant="outline" size="sm" onClick={retry}>
          {t.reload}
        </Button>
      </div>
    );
  return <TimelineMarkdown content={state.content} />;
}
export function TimelineDocuments({
  event,
  lang,
  onOpen,
}: {
  data: TimelineSessionResponse;
  event: TimelineEvent;
  loadDocument?: TimelineDocumentLoader;
  lang: TimelineLanguage;
  onOpen?: TimelineDocumentOpener;
}) {
  const t = timelineCopy(lang);
  const revisions = event.document_snapshots?.filter((r) => !r.task_id) ?? [];
  if (event.withdrawn || !event.stage || !revisions.length) return null;
  return (
    <Button
      type="button"
      variant="outline"
      size="sm"
      onClick={(e) => onOpen?.(event.event_id, e.currentTarget)}
    >
      <FileText aria-hidden />
      {t.viewStage[event.stage]}
    </Button>
  );
}
