import {
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  FileText,
  Flag,
  GitBranch,
  GitPullRequest,
  Images,
  MessageCircle,
  MessageSquare,
  Monitor,
  PauseCircle,
  RefreshCw,
  Smartphone,
  X,
} from "lucide-react";
import { type ComponentType, useRef, useState } from "react";
import {
  Alert,
  AlertAction,
  AlertDescription,
  AlertTitle,
} from "@/components/reui/alert";
import { Badge } from "@/components/reui/badge";
import {
  TimelineContent as EventContent,
  Timeline,
  TimelineDate,
  TimelineHeader,
  TimelineIndicator,
  TimelineItem,
  TimelineSeparator,
  TimelineTitle,
} from "@/components/reui/timeline";
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
import { Separator } from "@/components/ui/separator";
import { Skeleton } from "@/components/ui/skeleton";
import { Spinner } from "@/components/ui/spinner";
import {
  isSafeTimelineArtifactURL,
  isSafeTimelineURL,
  type TimelineArtifact,
  type TimelineEvent,
  type TimelineEventKind,
  type TimelineImageRef,
  type TimelineSessionResponse,
  type TimelineTicketRef,
  type TimelineView,
  timelineTicketIdentity,
} from "./timeline-contract";
import {
  type TimelineLanguage,
  timelineCopy,
  timelineDate,
} from "./timeline-copy";
import {
  type TimelineDocumentLoader,
  TimelineDocuments,
} from "./timeline-documents";

const icons: Record<
  TimelineEventKind,
  ComponentType<{ className?: string }>
> = {
  request: MessageSquare,
  work_started: GitBranch,
  validation: Images,
  changes_reported: MessageCircle,
  iteration: RefreshCw,
  approval_reported: CheckCircle2,
  delivery: GitPullRequest,
  closure: Flag,
  stage: FileText,
};
type Snapshot = {
  sessionID: string;
  sessionVersion: number;
  event: TimelineEvent;
  view: TimelineView;
  images: { ref: TimelineImageRef; artifact: TimelineArtifact; url: string }[];
  mode: "instructions" | "images";
  index: number;
};
export type TimelineContentProps = {
  data: TimelineSessionResponse;
  onRetry: () => void;
  loading?: boolean;
  error?: boolean;
  lang?: TimelineLanguage;
  loadDocument?: TimelineDocumentLoader;
};

function TicketLinks({ refs }: { refs: TimelineTicketRef[] }) {
  return (
    <div className="flex flex-wrap items-center gap-2">
      {refs.map((ref) =>
        isSafeTimelineURL(ref.url) ? (
          <a
            key={timelineTicketIdentity(ref)}
            href={ref.url}
            target="_blank"
            rel="noopener noreferrer"
            className="text-xs underline underline-offset-4 transition-colors hover:text-foreground"
          >
            {ref.key}
          </a>
        ) : (
          <span key={timelineTicketIdentity(ref)} className="text-xs">
            {ref.key}
          </span>
        ),
      )}
    </div>
  );
}

function keyedSteps(steps: string[]) {
  const occurrences = new Map<string, number>();
  return steps.map((text) => {
    const count = (occurrences.get(text) ?? 0) + 1;
    occurrences.set(text, count);
    return { key: `${count}:${text}`, text };
  });
}

function CaptureImage({
  src,
  alt,
  onRetry,
  label,
  className,
  retryLabel,
}: {
  src: string;
  alt: string;
  onRetry: () => void;
  label: string;
  retryLabel: string;
  className?: string;
}) {
  const [failed, setFailed] = useState(false);
  const [attempt, setAttempt] = useState(0);
  if (failed)
    return (
      <div className="flex flex-col items-start gap-2">
        <p className="text-xs text-muted-foreground">{label}</p>
        <Button
          type="button"
          variant="outline"
          size="sm"
          aria-label={retryLabel}
          onClick={() => {
            setFailed(false);
            setAttempt((n) => n + 1);
            onRetry();
          }}
        >
          ↻
        </Button>
      </div>
    );
  return (
    <img
      key={`${src}:${attempt}`}
      src={src}
      alt={alt}
      className={className}
      onError={() => setFailed(true)}
    />
  );
}

function CaptureThumbnail({
  url,
  alt,
  label,
  onOpen,
  onRetry,
  retryLabel,
  failureLabel,
}: {
  url: string;
  alt: string;
  label: string;
  onOpen: (trigger: HTMLElement) => void;
  onRetry: () => void;
  retryLabel: string;
  failureLabel: string;
}) {
  const [failed, setFailed] = useState(false);
  const [attempt, setAttempt] = useState(0);
  if (failed)
    return (
      <div className="flex flex-col items-start gap-2">
        <p className="text-xs text-muted-foreground">{failureLabel}</p>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => {
            setFailed(false);
            setAttempt((n) => n + 1);
            onRetry();
          }}
        >
          {retryLabel}
        </Button>
      </div>
    );
  return (
    <button
      type="button"
      className="group block w-full min-w-0 overflow-hidden border border-border text-left transition-colors hover:border-foreground focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      aria-label={label}
      onClick={(e) => onOpen(e.currentTarget)}
    >
      <img
        key={`${url}:${attempt}`}
        src={url}
        alt={alt}
        loading="lazy"
        onError={() => setFailed(true)}
        className="block aspect-[4/3] w-full object-cover object-top"
      />
    </button>
  );
}

export function TimelineContent({
  data,
  onRetry,
  loading = false,
  error = false,
  lang = "en",
  loadDocument,
}: TimelineContentProps) {
  const { session } = data;
  const t = timelineCopy(lang);
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null);
  const returnFocus = useRef<HTMLElement | null>(null);
  const events = [...session.events].sort((a, b) => b.sequence - a.sequence);
  const open = (
    event: TimelineEvent,
    view: TimelineView,
    mode: Snapshot["mode"],
    index: number,
    trigger: HTMLElement,
  ) => {
    const images = view.image_refs.flatMap((ref) => {
      const artifact = session.artifacts.find(
        (a) => a.artifact_id === ref.artifact_id,
      );
      const url = data.artifact_urls?.[ref.artifact_id];
      return artifact &&
        isSafeTimelineArtifactURL(url, session.viewer_session_id)
        ? [
            {
              ref: { ...ref },
              artifact: { ...artifact, candidate: { ...artifact.candidate } },
              url,
            },
          ]
        : [];
    });
    returnFocus.current = trigger;
    setSnapshot({
      sessionID: session.viewer_session_id,
      sessionVersion: session.version,
      event: { ...event },
      view: { ...view, steps: [...view.steps] },
      images,
      mode,
      index,
    });
  };
  const isUpdated =
    snapshot &&
    (snapshot.sessionID !== session.viewer_session_id ||
      snapshot.sessionVersion !== session.version);
  const selected = snapshot?.images[snapshot.index];
  const move = (direction: number) =>
    setSnapshot((current) =>
      current?.images.length
        ? {
            ...current,
            index:
              (current.index + direction + current.images.length) %
              current.images.length,
          }
        : current,
    );
  return (
    <div className="mx-auto flex w-full min-w-0 max-w-2xl flex-col gap-6">
      {session.summary && (
        <p className="text-sm text-muted-foreground">{session.summary}</p>
      )}
      {error && (
        <Alert variant="destructive">
          <AlertTitle>{t.loadFailed}</AlertTitle>
          <AlertAction>
            <Button type="button" variant="outline" size="sm" onClick={onRetry}>
              {t.retry}
            </Button>
          </AlertAction>
        </Alert>
      )}
      {loading && (
        <div
          role="status"
          aria-label={t.loading}
          className="flex flex-col gap-4"
        >
          <Skeleton className="h-4 w-2/3" />
          <Skeleton className="h-32 w-full" />
        </div>
      )}
      {!loading && events.length === 0 && (
        <div role="status" className="flex flex-col items-start gap-3">
          <p className="text-sm text-muted-foreground">{t.empty}</p>
          <Button type="button" variant="outline" size="sm" onClick={onRetry}>
            {t.reload}
          </Button>
        </div>
      )}
      {events.length > 0 && (
        <>
          <h2 className="sr-only">{t.timeline}</h2>
          <Timeline value={0} role="region" aria-label={t.timeline}>
            {events.map((event, index) => {
              const Icon =
                event.kind === "stage"
                  ? event.stage_state === "in_progress"
                    ? Spinner
                    : event.stage_state === "completed"
                      ? CheckCircle2
                      : event.stage_state === "paused"
                        ? PauseCircle
                        : event.stage_state === "failed"
                          ? CircleAlert
                          : FileText
                  : icons[event.kind];
              return (
                <TimelineItem
                  key={event.event_id}
                  step={index + 1}
                  id={`timeline-event-${event.event_id}`}
                  className="group-data-[orientation=vertical]/timeline:ms-9"
                >
                  <TimelineSeparator className="bg-border group-data-[orientation=vertical]/timeline:top-1 group-data-[orientation=vertical]/timeline:-left-7 group-data-[orientation=vertical]/timeline:h-[calc(100%-2.25rem)] group-data-[orientation=vertical]/timeline:translate-y-6" />
                  <TimelineIndicator className="flex size-6 items-center justify-center border bg-background group-data-[orientation=vertical]/timeline:-left-7 group-data-[orientation=vertical]/timeline:-top-0.5">
                    <Icon
                      className="size-4 motion-reduce:animate-none"
                      aria-hidden
                    />
                  </TimelineIndicator>
                  <TimelineHeader className="flex min-w-0 items-start justify-between gap-2.5">
                    <div className="flex min-w-0 flex-wrap items-center gap-2">
                      <TimelineTitle
                        className="min-w-0 truncate"
                        title={event.title}
                      >
                        {event.withdrawn ? t.withdrawn : event.title}
                      </TimelineTitle>
                      {!event.views?.length && (
                        <TicketLinks refs={event.ticket_refs} />
                      )}
                    </div>
                  </TimelineHeader>
                  <EventContent className="flex min-w-0 flex-col items-start gap-2 pb-1">
                    {!event.withdrawn && event.summary && (
                      <p className="text-sm">{event.summary}</p>
                    )}
                    {!event.withdrawn &&
                      event.candidate?.branch &&
                      event.kind === "work_started" && (
                        <p className="break-words text-xs">
                          <code>{event.candidate.branch}</code>
                        </p>
                      )}
                    {!event.withdrawn &&
                      event.views?.map((view) => {
                        const newer = events.find(
                          (e) =>
                            !e.withdrawn &&
                            e.views?.some(
                              (v) =>
                                v.supersedes?.event_id === event.event_id &&
                                v.supersedes.evidence_id ===
                                  event.evidence_id &&
                                v.supersedes.view_id === view.view_id,
                            ),
                        );
                        return (
                          <section
                            key={view.view_id}
                            aria-label={view.title}
                            className="flex w-full min-w-0 flex-col gap-3"
                          >
                            <div className="flex min-w-0 flex-wrap items-center gap-2">
                              <h4
                                className="min-w-0 truncate text-sm font-medium text-foreground"
                                title={view.title}
                              >
                                {view.title}
                              </h4>
                              <TicketLinks refs={view.ticket_refs} />
                              {newer && (
                                <a
                                  href={`#timeline-event-${newer.event_id}`}
                                  className="text-xs underline underline-offset-4"
                                >
                                  {t.superseded} {newer.title}
                                </a>
                              )}
                            </div>
                            <div className="grid min-w-0 grid-cols-2 gap-3 sm:grid-cols-3 xl:grid-cols-4">
                              {view.image_refs.map((ref) => {
                                const artifact = session.artifacts.find(
                                  (a) => a.artifact_id === ref.artifact_id,
                                );
                                const url =
                                  data.artifact_urls?.[ref.artifact_id];
                                if (
                                  !artifact ||
                                  !isSafeTimelineArtifactURL(
                                    url,
                                    session.viewer_session_id,
                                  )
                                )
                                  return (
                                    <p
                                      key={ref.artifact_id}
                                      className="text-xs text-muted-foreground"
                                    >
                                      {ref.caption} · {t.unavailable}
                                    </p>
                                  );
                                const validIndex = view.image_refs
                                  .slice(0, view.image_refs.indexOf(ref))
                                  .filter(
                                    (r) =>
                                      session.artifacts.some(
                                        (a) => a.artifact_id === r.artifact_id,
                                      ) &&
                                      isSafeTimelineArtifactURL(
                                        data.artifact_urls?.[r.artifact_id],
                                        session.viewer_session_id,
                                      ),
                                  ).length;
                                return (
                                  <figure
                                    key={ref.artifact_id}
                                    className="flex min-w-0 flex-col gap-1"
                                  >
                                    <CaptureThumbnail
                                      url={url}
                                      alt={ref.alt ?? ref.caption}
                                      label={`${t.captures}: ${view.title} · ${ref.caption}`}
                                      onOpen={(trigger) =>
                                        open(
                                          event,
                                          view,
                                          "images",
                                          validIndex,
                                          trigger,
                                        )
                                      }
                                      onRetry={onRetry}
                                      retryLabel={t.retry}
                                      failureLabel={t.imageFailed}
                                    />
                                    <figcaption className="flex min-w-0 items-center gap-1 text-xs">
                                      <span aria-hidden>
                                        {artifact.viewport === "mobile" ? (
                                          <Smartphone className="size-3" />
                                        ) : (
                                          <Monitor className="size-3" />
                                        )}
                                      </span>
                                      <span
                                        className="truncate"
                                        title={ref.caption}
                                      >
                                        {ref.caption}
                                      </span>
                                    </figcaption>
                                  </figure>
                                );
                              })}
                            </div>
                            <div>
                              <Button
                                type="button"
                                variant="outline"
                                size="sm"
                                onClick={(e) =>
                                  open(
                                    event,
                                    view,
                                    "instructions",
                                    0,
                                    e.currentTarget,
                                  )
                                }
                              >
                                <FileText aria-hidden />
                                {t.instructions}
                              </Button>
                            </div>
                          </section>
                        );
                      })}
                    {event.kind === "stage" && (
                      <TimelineDocuments
                        data={data}
                        event={event}
                        lang={lang}
                        loadDocument={loadDocument}
                      />
                    )}
                    {!event.withdrawn && event.closure && (
                      <p className="text-sm">{event.closure.result}</p>
                    )}
                    <div
                      data-slot="timeline-event-footer"
                      className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground"
                    >
                      <TimelineDate
                        className="mb-0"
                        dateTime={event.created_at}
                      >
                        {timelineDate(event.created_at, lang)}
                      </TimelineDate>
                      {!event.withdrawn &&
                        event.kind === "stage" &&
                        event.stage_state && (
                          <Badge
                            variant={
                              (
                                {
                                  in_progress: "warning-light",
                                  completed: "success-light",
                                  paused: "outline",
                                  failed: "destructive-light",
                                } as const
                              )[event.stage_state]
                            }
                          >
                            {t.stageStates[event.stage_state]}
                          </Badge>
                        )}
                      {(event.kind === "changes_reported" ||
                        event.kind === "approval_reported") && (
                        <span>{t.reported}</span>
                      )}
                      {!event.withdrawn &&
                        event.deliveries?.map((delivery) => (
                          <div
                            key={delivery.url}
                            className="flex min-w-0 flex-wrap items-center gap-2"
                          >
                            {isSafeTimelineURL(delivery.url) ? (
                              <a
                                href={delivery.url}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-sm underline underline-offset-4"
                              >
                                {delivery.label}
                              </a>
                            ) : (
                              <span>{delivery.label}</span>
                            )}
                            <Badge
                              variant={
                                delivery.status === "merged"
                                  ? "success-light"
                                  : "outline"
                              }
                            >
                              {t[delivery.status]}
                            </Badge>
                          </div>
                        ))}
                    </div>
                  </EventContent>
                </TimelineItem>
              );
            })}
          </Timeline>
        </>
      )}
      <Dialog
        open={Boolean(snapshot)}
        onOpenChange={(open) => {
          if (!open) setSnapshot(null);
        }}
      >
        <DialogContent
          finalFocus={returnFocus}
          style={{
            width: "calc(100vw - 2rem)",
            maxWidth: snapshot?.mode === "images" ? "72rem" : "40rem",
          }}
          onKeyDown={(e) => {
            if (snapshot?.mode !== "images") return;
            if (e.key === "ArrowRight") {
              e.preventDefault();
              move(1);
            }
            if (e.key === "ArrowLeft") {
              e.preventDefault();
              move(-1);
            }
          }}
        >
          <DialogHeader>
            <DialogTitle>{snapshot?.view.title}</DialogTitle>
            <DialogDescription className="whitespace-pre-line">
              {snapshot?.event.title}
            </DialogDescription>
          </DialogHeader>
          {isUpdated && (
            <Alert variant="info">
              <AlertDescription>{t.updateAvailable}</AlertDescription>
            </Alert>
          )}
          {snapshot?.mode === "instructions" ? (
            <div className="flex min-h-0 flex-col gap-4 overflow-auto">
              {snapshot.view.steps.length ? (
                <ol className="flex list-decimal flex-col gap-3 pl-5 text-sm leading-relaxed">
                  {keyedSteps(snapshot.view.steps).map((step) => (
                    <li className="whitespace-pre-line" key={step.key}>
                      {step.text}
                    </li>
                  ))}
                </ol>
              ) : (
                <p className="text-sm text-muted-foreground">{t.noSteps}</p>
              )}
              <Separator />
              <div className="flex flex-col gap-1">
                <p className="text-xs font-medium text-muted-foreground">
                  {t.expected}
                </p>
                <p className="whitespace-pre-line text-sm leading-relaxed">
                  {snapshot.view.expected_result}
                </p>
              </div>
            </div>
          ) : selected ? (
            <div className="flex min-h-0 flex-col gap-3 overflow-auto">
              <CaptureImage
                src={selected.url}
                alt={selected.ref.alt ?? selected.ref.caption}
                label={t.imageFailed}
                retryLabel={t.retry}
                onRetry={onRetry}
                className="mx-auto max-w-full"
              />
              <p className="text-sm">{selected.ref.caption}</p>
              <p className="text-xs text-muted-foreground">
                {t[selected.artifact.viewport]} · {t[selected.artifact.theme]} ·{" "}
                {t.captured}:{" "}
                {timelineDate(selected.artifact.captured_at, lang)} ·{" "}
                <code
                  className="inline-block max-w-full truncate align-bottom"
                  title={selected.artifact.candidate.commit}
                >
                  {selected.artifact.candidate.commit}
                </code>
              </p>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">{t.unavailable}</p>
          )}
          <DialogFooter>
            {snapshot?.mode === "images" && snapshot.images.length > 0 ? (
              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="icon-sm"
                  aria-label={t.previous}
                  disabled={snapshot.images.length < 2}
                  onClick={() => move(-1)}
                >
                  <ChevronLeft aria-hidden />
                </Button>
                <span aria-live="polite" className="text-xs tabular-nums">
                  {snapshot.index + 1} / {snapshot.images.length}
                </span>
                <Button
                  type="button"
                  variant="outline"
                  size="icon-sm"
                  aria-label={t.next}
                  disabled={snapshot.images.length < 2}
                  onClick={() => move(1)}
                >
                  <ChevronRight aria-hidden />
                </Button>
              </div>
            ) : null}
            <DialogCancel variant="outline" size="sm">
              <X aria-hidden />
              {t.close}
            </DialogCancel>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
