import { FileText } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@/components/reui/alert";
import { Badge } from "@/components/reui/badge";
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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import type {
  TimelineDocument,
  TimelineDocumentResponse,
  TimelineEvent,
  TimelineSessionResponse,
} from "./timeline-contract";
import {
  type TimelineLanguage,
  timelineCopy,
  timelineDate,
} from "./timeline-copy";
import { TimelineMarkdown } from "./timeline-markdown";

/** Transport and authentication belong to the caller, never to the renderer. */
export type TimelineDocumentLoader = (
  documentID: string,
) => Promise<TimelineDocumentResponse>;
type CapturedDocuments = {
  sessionID: string;
  version: number;
  event: TimelineEvent;
  documents: TimelineDocument[];
  loadDocument?: TimelineDocumentLoader;
};

function DocumentBody({
  document,
  loadDocument,
  lang,
}: {
  document?: TimelineDocument;
  loadDocument?: TimelineDocumentLoader;
  lang: TimelineLanguage;
}) {
  const t = timelineCopy(lang);
  const [attempt, setAttempt] = useState(0);
  const [state, setState] = useState<{
    kind: "loading" | "error" | "ready";
    content?: string;
  }>({ kind: "loading" });
  useEffect(() => {
    void attempt; // An explicit retry starts a fresh read of the same immutable document.
    let disposed = false;
    setState({ kind: "loading" });
    if (!document || !loadDocument) return;
    void Promise.resolve()
      .then(() => loadDocument(document.document_id))
      .then((response) => {
        if (
          response.schema_version !== 2 ||
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
  if (!document || !loadDocument)
    return (
      <p className="text-sm text-muted-foreground">{t.documentUnavailable}</p>
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
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={() => setAttempt((n) => n + 1)}
          >
            {t.retry}
          </Button>
        </AlertAction>
      </Alert>
    );
  if (!state.content?.trim())
    return (
      <div role="status" className="flex flex-col items-start gap-3">
        <p className="text-sm text-muted-foreground">{t.documentEmpty}</p>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => setAttempt((n) => n + 1)}
        >
          {t.reload}
        </Button>
      </div>
    );
  return <TimelineMarkdown content={state.content} />;
}

function DocumentRevisions({
  snapshot,
  lang,
}: {
  snapshot: CapturedDocuments;
  lang: TimelineLanguage;
}) {
  const t = timelineCopy(lang);
  const revisions = snapshot.event.document_snapshots ?? [];
  const [revisionID, setRevisionID] = useState(
    revisions.at(-1)?.document_revision_id ?? "",
  );
  const revision = revisions.find((r) => r.document_revision_id === revisionID);
  const items = [...revisions].reverse().map((r) => ({
    value: r.document_revision_id,
    label: `${timelineDate(r.created_at, lang)} · ${r.status === "draft" ? t.draft : t.complete}`,
  }));
  return (
    <div className="flex min-h-0 min-w-0 flex-col gap-4 overflow-y-auto">
      <div className="flex flex-wrap items-center gap-2">
        <Select
          value={revisionID}
          onValueChange={(value) => {
            if (value) setRevisionID(value);
          }}
          items={items}
        >
          <SelectTrigger size="sm" aria-label={t.revision}>
            <SelectValue />
          </SelectTrigger>
          <SelectContent alignItemWithTrigger={false}>
            {items.map((item) => (
              <SelectItem key={item.value} value={item.value}>
                {item.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        {revision && (
          <Badge variant="outline">
            {revision.status === "draft" ? t.draft : t.complete}
          </Badge>
        )}
      </div>
      {revision && (
        <Tabs key={revision.document_revision_id} defaultValue="0">
          <div className="min-w-0 overflow-x-auto">
            <TabsList>
              {revision.tabs.map((tab, index) => (
                <TabsTrigger
                  key={`${tab.label}:${tab.document_id}`}
                  value={String(index)}
                >
                  {tab.label}
                </TabsTrigger>
              ))}
            </TabsList>
          </div>
          {revision.tabs.map((tab, index) => (
            <TabsContent
              key={`${tab.label}:${tab.document_id}`}
              value={String(index)}
            >
              <DocumentBody
                document={snapshot.documents.find(
                  (d) => d.document_id === tab.document_id,
                )}
                loadDocument={snapshot.loadDocument}
                lang={lang}
              />
            </TabsContent>
          ))}
        </Tabs>
      )}
    </div>
  );
}

export function TimelineDocuments({
  data,
  event,
  loadDocument,
  lang,
}: {
  data: TimelineSessionResponse;
  event: TimelineEvent;
  loadDocument?: TimelineDocumentLoader;
  lang: TimelineLanguage;
}) {
  const t = timelineCopy(lang);
  const [snapshot, setSnapshot] = useState<CapturedDocuments | null>(null);
  const returnFocus = useRef<HTMLElement | null>(null);
  if ((!event.document_snapshots?.length || !event.stage) && !snapshot)
    return null;
  const updated =
    snapshot &&
    (snapshot.version !== data.session.version ||
      snapshot.sessionID !== data.session.viewer_session_id);
  return (
    <>
      {!event.withdrawn &&
        event.stage &&
        !!event.document_snapshots?.length && (
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={(e) => {
              returnFocus.current = e.currentTarget;
              setSnapshot({
                sessionID: data.session.viewer_session_id,
                version: data.session.version,
                event: structuredClone(event),
                documents: structuredClone(data.session.documents),
                loadDocument,
              });
            }}
          >
            <FileText aria-hidden />
            {t.viewStage[event.stage]}
          </Button>
        )}
      <Dialog
        open={!!snapshot}
        onOpenChange={(open) => {
          if (!open) setSnapshot(null);
        }}
      >
        {snapshot && (
          <DialogContent
            style={{ width: "calc(100vw - 2rem)", maxWidth: "40rem" }}
            finalFocus={returnFocus}
          >
            <DialogHeader>
              <DialogTitle>{snapshot.event.title}</DialogTitle>
              <DialogDescription>{t.documents}</DialogDescription>
            </DialogHeader>
            {updated && (
              <Alert>
                <AlertDescription>{t.documentUpdated}</AlertDescription>
              </Alert>
            )}
            <DocumentRevisions snapshot={snapshot} lang={lang} />
            <DialogFooter>
              <DialogCancel>{t.close}</DialogCancel>
            </DialogFooter>
          </DialogContent>
        )}
      </Dialog>
    </>
  );
}
