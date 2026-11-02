import {
  CheckCircle2,
  ChevronDown,
  ChevronLeft,
  ChevronRight,
  CircleAlert,
  CircleDot,
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
import { type ComponentType, type ReactNode, useRef, useState } from "react";
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
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from "@/components/ui/collapsible";
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
  type TimelineArtifact,
  type TimelineEvent,
  type TimelineEventKind,
  type TimelineImageRef,
  type TimelineSessionResponse,
  type TimelineView,
} from "./timeline-contract";
import {
  type TimelineLanguage,
  timelineCopy,
  timelineDate,
  timelineEventTitle,
} from "./timeline-copy";
import { useTimelinePanel } from "./timeline-detail-panel";
import {
  type TimelineDocumentLoader,
  TimelineDocuments,
} from "./timeline-documents";
import {
  distinctCaptions,
  groupMustStayOpen,
  groupTimeline,
} from "./timeline-grouping";
import { CaptureImage, CaptureThumbnail } from "./timeline-images";
import {
  TimelineDeliveryLinks,
  TimelineTicketLinks,
} from "./timeline-references";
import { TimelineSummary } from "./timeline-summary";
import { TimelineTaskRegistry } from "./timeline-task-registry";
import {
  type TimelineDetailTarget,
  TimelineTaskRow,
  TimelineWaitRow,
  timelineStateColor,
} from "./timeline-tasks";

const icons: Record<
  TimelineEventKind,
  ComponentType<{ className?: string }>
> = {
  review: CheckCircle2,
  state_transition: CheckCircle2,
  producer_joined: CheckCircle2,
  foreground_transferred: CheckCircle2,
  task_state_transition: CheckCircle2,
  local_action: CheckCircle2,
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
  candidates: TimelineSessionResponse["session"]["candidates"];
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
  onOpenDetail?: (target: TimelineDetailTarget, trigger: HTMLElement) => void;
  targetEventID?: string;
  defaultExpandedTaskIDs?: string[];
};

function keyedSteps(steps: string[]) {
  const occurrences = new Map<string, number>();
  return steps.map((text) => {
    const count = (occurrences.get(text) ?? 0) + 1;
    occurrences.set(text, count);
    return { key: `${count}:${text}`, text };
  });
}

export function TimelineContent({
  data,
  onRetry,
  loading = false,
  error = false,
  lang = "en",
  loadDocument,
  onOpenDetail,
  targetEventID,
  defaultExpandedTaskIDs = [],
}: TimelineContentProps) {
  const { session } = data;
  const panel = useTimelinePanel({ data, lang, loadDocument, onRetry });
  const openDetail: NonNullable<TimelineContentProps["onOpenDetail"]> = (
    target,
    trigger,
  ) => {
    onOpenDetail?.(target, trigger);
    panel.openDetail(target, trigger);
  };
  const t = timelineCopy(lang);
  const [snapshot, setSnapshot] = useState<Snapshot | null>(null);
  const returnFocus = useRef<HTMLElement | null>(null);
  const events = [...session.events].sort((a, b) => a.sequence - b.sequence);
  const groups = groupTimeline(session);
  const [expandedStages, setExpandedStages] = useState<Record<string, boolean>>(
    {},
  );
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
              artifact: { ...artifact },
              url,
            },
          ]
        : [];
    });
    returnFocus.current = trigger;
    setSnapshot({
      sessionID: session.viewer_session_id,
      sessionVersion: session.version,
      candidates: structuredClone(session.candidates),
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
  const renderEvent = (
    event: TimelineEvent,
    index: number,
    children?: ReactNode,
    stageOpen = true,
    toggleStage?: () => void,
  ) => {
    const producer =
      event.actor.kind === "native_producer"
        ? session.producers.find(
            (p) =>
              p.producer_id ===
              (event.actor.kind === "native_producer"
                ? event.actor.producer_id
                : ""),
          )
        : undefined;
    const labels: Record<string, string> = {
      ...t.stageStates,
      ...t.states,
      ...t.taskStates,
    };
    const stageTasks =
      event.kind === "stage"
        ? (groups.find((g) => g.stage?.event_id === event.event_id)?.tasks ??
          [])
        : [];
    const nodeState =
      event.kind === "stage"
        ? event.stage_state
        : event.kind === "review"
          ? event.review_status === "completed"
            ? "completed"
            : event.review_status === "failed"
              ? "failed"
              : "in_review"
          : (event.audit?.next_state ??
            (event.kind === "closure"
              ? event.closure?.reason === "completed"
                ? "completed"
                : "paused"
              : undefined));
    const stateDriven =
      event.kind === "stage" ||
      event.kind === "state_transition" ||
      event.kind === "task_state_transition" ||
      event.kind === "review";
    const Icon = stateDriven
      ? nodeState === "blocked" || nodeState === "failed"
        ? CircleAlert
        : ["in_review", "awaiting_validation", "paused", "waiting"].includes(
              nodeState ?? "",
            )
          ? PauseCircle
          : nodeState === "in_progress" || nodeState === "active"
            ? Spinner
            : ["completed", "closed", "succeeded"].includes(nodeState ?? "")
              ? CheckCircle2
              : FileText
      : icons[event.kind];
    return (
      <TimelineItem
        key={event.event_id}
        step={index + 1}
        id={`timeline-event-${event.event_id}`}
        data-event-id={event.event_id}
        data-completed={event.stage_state === "completed" || undefined}
        className="scroll-mt-4 group-data-[orientation=vertical]/timeline:ms-9"
      >
        <TimelineSeparator className="bg-border group-data-[orientation=vertical]/timeline:top-1 group-data-[orientation=vertical]/timeline:-left-7 group-data-[orientation=vertical]/timeline:h-[calc(100%-2.25rem)] group-data-[orientation=vertical]/timeline:translate-y-6" />
        <TimelineIndicator
          className={`flex size-6 items-center justify-center border-0 bg-background group-data-[orientation=vertical]/timeline:-left-7 group-data-[orientation=vertical]/timeline:-top-0.5 ${timelineStateColor(nodeState)}`}
        >
          <Icon className="size-4 motion-reduce:animate-none" aria-hidden />
        </TimelineIndicator>
        <TimelineHeader className="flex min-w-0 items-start justify-between gap-2.5">
          <div className="flex min-w-0 flex-wrap items-center gap-2">
            <TimelineTitle
              className="min-w-0 truncate"
              title={timelineEventTitle(event, lang)}
            >
              {event.withdrawn ? t.withdrawn : timelineEventTitle(event, lang)}
            </TimelineTitle>
            {stageTasks.length > 0 && (
              <span className="text-xs tabular-nums">
                {
                  stageTasks.filter((p) => p.segment.state === "completed")
                    .length
                }
                /{stageTasks.length} {t.completedTasks}
              </span>
            )}
            {toggleStage && (
              <CollapsibleTrigger
                render={
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    aria-label={`${stageOpen ? t.collapse : t.expand} ${timelineEventTitle(event, lang)}`}
                  />
                }
                onClick={toggleStage}
              >
                <ChevronDown
                  aria-hidden
                  className={stageOpen ? "size-4" : "size-4 -rotate-90"}
                />
              </CollapsibleTrigger>
            )}
            {!event.views?.length && (
              <TimelineTicketLinks
                session={session}
                refs={event.ticket_refs}
                lang={lang}
              />
            )}
          </div>
        </TimelineHeader>
        <EventContent
          hidden={!stageOpen}
          className="flex min-w-0 flex-col items-start gap-2 pb-1"
        >
          {!event.withdrawn && event.summary && (
            <TimelineSummary content={event.summary} lines={3} lang={lang} />
          )}
          {!event.withdrawn &&
            session.candidates.find(
              (c) => c.candidate_id === event.candidate_id,
            )?.branch &&
            event.kind === "work_started" && (
              <p className="break-words text-xs">
                <code>
                  {
                    session.candidates.find(
                      (c) => c.candidate_id === event.candidate_id,
                    )?.branch
                  }
                </code>
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
                      v.supersedes.evidence_id === event.evidence_id &&
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
                    <TimelineTicketLinks
                      session={session}
                      refs={view.ticket_refs}
                      lang={lang}
                    />
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
                      const url = data.artifact_urls?.[ref.artifact_id];
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
                            alt={
                              ref.alt ||
                              ref.caption ||
                              `${t.captures} ${view.image_refs.indexOf(ref) + 1}`
                            }
                            label={`${t.captures}: ${view.title} · ${ref.caption}`}
                            onOpen={(trigger) =>
                              open(event, view, "images", validIndex, trigger)
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
                              title={
                                ref.caption ||
                                `${t.captures} ${view.image_refs.indexOf(ref) + 1}`
                              }
                            >
                              {
                                distinctCaptions(
                                  view.image_refs.map((r) => r.caption),
                                  t.captures,
                                )[view.image_refs.indexOf(ref)]
                              }
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
                        open(event, view, "instructions", 0, e.currentTarget)
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
              onOpen={panel.openDocument}
            />
          )}
          {event.reason && <p className="text-sm">{event.reason}</p>}
          {event.audit && <p className="text-sm">{event.audit.reason}</p>}
          {children}
          {!event.withdrawn && event.closure && (
            <p className="text-sm">{event.closure.result}</p>
          )}
          <div
            data-slot="timeline-event-footer"
            className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground"
          >
            <TimelineDate className="mb-0" dateTime={event.created_at}>
              {timelineDate(event.created_at, lang)}
            </TimelineDate>
            {!event.withdrawn &&
              event.kind === "stage" &&
              event.stage_state && (
                <Badge
                  variant={
                    (
                      {
                        in_review: "info-light",
                        awaiting_validation: "warning-light",
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
            <span title={producer?.source.native_session_id}>
              {event.actor.kind === "local_viewer"
                ? t.localViewer
                : producer
                  ? `${producer.source.host} · ${producer.source.native_session_id.slice(0, 8)}`
                  : t.producerUnavailable}
            </span>
            {event.audit?.previous_state && event.audit.next_state && (
              <span>
                {labels[event.audit.previous_state] ??
                  event.audit.previous_state}{" "}
                → {labels[event.audit.next_state] ?? event.audit.next_state}
              </span>
            )}
            {(event.kind === "changes_reported" ||
              event.kind === "approval_reported") && <span>{t.reported}</span>}
            {!event.withdrawn && event.delivery_refs && (
              <TimelineDeliveryLinks
                session={session}
                refs={event.delivery_refs}
                lang={lang}
              />
            )}
          </div>
        </EventContent>
      </TimelineItem>
    );
  };
  return (
    <div className="mx-auto flex w-full min-w-0 max-w-2xl flex-col gap-6 pb-8">
      {session.summary && (
        <TimelineSummary content={session.summary} lines={2} lang={lang} />
      )}
      <div
        role="toolbar"
        aria-label={t.timelineActions}
        className="flex flex-wrap items-center gap-2"
      >
        <TimelineTaskRegistry
          session={session}
          lang={lang}
          onRetry={onRetry}
          loading={loading}
          error={error}
        />
        <Button
          type="button"
          variant="outline"
          size="sm"
          className="self-start"
          disabled={!events.length}
          onClick={() => {
            const visible = [
              ...document.querySelectorAll<HTMLElement>("[data-event-id]"),
            ].filter((node) => !node.closest("[hidden]"));
            const node = session.current_stage_event_id
              ? document.getElementById(
                  `timeline-event-${session.current_stage_event_id}`,
                )
              : visible.at(-1);
            const canvas = node?.closest<HTMLElement>(
              '[data-slot="viewer-canvas"]',
            );
            if (node && canvas)
              canvas.scrollTop +=
                node.getBoundingClientRect().top -
                canvas.getBoundingClientRect().top;
            if (node) {
              node.tabIndex = -1;
              node.focus({ preventScroll: true });
            }
          }}
        >
          {t.goCurrent}
        </Button>
      </div>
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
      <h2 className="sr-only">{t.timeline}</h2>
      <Timeline value={0} role="region" aria-label={t.timeline}>
        <TimelineItem
          step={1}
          id={`timeline-origin-${session.viewer_session_id}`}
          data-ui-origin="true"
          className="group-data-[orientation=vertical]/timeline:ms-9"
        >
          {events.length > 0 && (
            <TimelineSeparator className="bg-border group-data-[orientation=vertical]/timeline:top-1 group-data-[orientation=vertical]/timeline:-left-7 group-data-[orientation=vertical]/timeline:h-[calc(100%-2.25rem)] group-data-[orientation=vertical]/timeline:translate-y-6" />
          )}
          <TimelineIndicator className="flex size-6 items-center justify-center border-0 bg-background text-primary group-data-[orientation=vertical]/timeline:-left-7 group-data-[orientation=vertical]/timeline:-top-0.5">
            <CircleDot aria-hidden className="size-4" />
          </TimelineIndicator>
          <TimelineHeader>
            <TimelineTitle>{t.start}</TimelineTitle>
          </TimelineHeader>
          <EventContent className="flex flex-wrap items-center gap-2">
            <TimelineDate dateTime={session.created_at}>
              {timelineDate(session.created_at, lang)}
            </TimelineDate>
            <span
              className="text-xs text-muted-foreground"
              title={session.source.native_session_id}
            >
              {session.source.host} ·{" "}
              {session.source.native_session_id.slice(0, 8)}
            </span>
          </EventContent>
        </TimelineItem>
        {groups.map((group, index) => {
          if (!group.stage)
            return group.events.map((e) => renderEvent(e, index));
          const stage = group.stage;
          const targeted =
            stage.event_id === targetEventID ||
            group.events.some((e) => e.event_id === targetEventID) ||
            group.tasks.some((p) =>
              p.events.some((e) => e.event_id === targetEventID),
            );
          const forced =
            groupMustStayOpen(group, session.current_stage_event_id) ||
            targeted;
          const stageOpen = forced || (expandedStages[stage.event_id] ?? false);
          return (
            <Collapsible key={stage.event_id} open={stageOpen}>
              {renderEvent(
                stage,
                index,
                <CollapsibleContent keepMounted className="w-full min-w-0">
                  {!group.tasks.length &&
                    ["build", "verify", "delivery"].includes(
                      stage.stage ?? "",
                    ) && (
                      <div
                        role="status"
                        className="flex flex-col items-start gap-2"
                      >
                        <p className="text-sm">{t.noTasks}</p>
                        <Button
                          type="button"
                          variant="outline"
                          size="sm"
                          onClick={onRetry}
                        >
                          {t.reload}
                        </Button>
                      </div>
                    )}
                  {group.tasks.length > 0 && (
                    <Timeline
                      value={0}
                      className="mt-4 w-full"
                      role="group"
                      aria-label={t.tasks}
                    >
                      {group.tasks.map((projection, i) => (
                        <TimelineTaskRow
                          key={projection.segment.segment_id}
                          projection={projection}
                          data={data}
                          initiallyExpanded={defaultExpandedTaskIDs.includes(
                            projection.task.task_id,
                          )}
                          targetEventID={targetEventID}
                          index={i}
                          last={i === group.tasks.length - 1}
                          lang={lang}
                          onOpen={openDetail}
                          renderActivity={(e) => renderEvent(e, i)}
                        />
                      ))}
                    </Timeline>
                  )}
                  {group.waits.length > 0 && (
                    <Timeline value={0} role="group" aria-label={t.checks}>
                      {group.waits.map((wait, i) => (
                        <TimelineWaitRow
                          key={wait.wait_id}
                          wait={wait}
                          index={i}
                          last={i === group.waits.length - 1}
                          lang={lang}
                          onOpen={openDetail}
                        />
                      ))}
                    </Timeline>
                  )}
                  {group.events.length > 0 && (
                    <Timeline value={0} role="group" aria-label={t.activity}>
                      {group.events.map((e, i) => renderEvent(e, i))}
                    </Timeline>
                  )}
                </CollapsibleContent>,
                stageOpen,
                forced
                  ? undefined
                  : () =>
                      setExpandedStages((old) => ({
                        ...old,
                        [stage.event_id]: !stageOpen,
                      })),
              )}
            </Collapsible>
          );
        })}
      </Timeline>
      {panel.sheet}
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
              {snapshot && timelineEventTitle(snapshot.event, lang)}
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
                key={selected.url}
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
                  title={
                    snapshot?.candidates.find(
                      (c) => c.candidate_id === selected.artifact.candidate_id,
                    )?.commit
                  }
                >
                  {
                    snapshot?.candidates.find(
                      (c) => c.candidate_id === selected.artifact.candidate_id,
                    )?.commit
                  }
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
