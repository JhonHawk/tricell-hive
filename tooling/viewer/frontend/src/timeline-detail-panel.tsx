import {
  ArrowLeft,
  Check,
  ChevronLeft,
  ChevronRight,
  Circle,
  CircleAlert,
  CircleHelp,
  Pause,
  X,
} from "lucide-react";
import { type ReactNode, useLayoutEffect, useRef, useState } from "react";
import { Alert, AlertAction, AlertDescription } from "@/components/reui/alert";
import { Badge } from "@/components/reui/badge";
import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet";
import { Spinner } from "@/components/ui/spinner";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { timelineCandidateByID } from "./timeline-catalog";
import {
  isSafeTimelineArtifactURL,
  isSafeTimelineURL,
  type TimelineEvent,
  type TimelineSessionResponse,
  type TimelineView,
} from "./timeline-contract";
import {
  aggregateTimelineCIWait,
  type TimelineCIObservation,
  type TimelineCIWait,
  type TimelineTaskState,
  type TimelineWaitState,
} from "./timeline-contract-v3";
import {
  isHostTask,
  localViewerLabel,
  type TimelineLanguage,
  timelineCopy,
  timelineDate,
  timelineEventTitle,
  timelineTaskLabel,
} from "./timeline-copy";
import {
  compareRevisionTabs,
  type DocumentRevision,
  documentChains,
  latestSuccessor,
  revisionTabs,
} from "./timeline-document-chains";
import {
  DocumentBody,
  type TimelineDocumentLoader,
  type TimelineDocumentOpener,
} from "./timeline-documents";
import { CaptureImage, CaptureThumbnail } from "./timeline-images";
import { TimelineMarkdown } from "./timeline-markdown";
import {
  TimelineDeliveryLinks,
  TimelineTicketLinks,
} from "./timeline-references";
import {
  HostObservedBadge,
  type TimelineDetailOpener,
  timelineStateColor,
} from "./timeline-tasks";

type BasePanel = { data: TimelineSessionResponse; back?: Panel };
type DocumentPanel = BasePanel & {
  kind: "document";
  eventID: string;
  key: string;
  keys: string[];
  revisionID: string;
  label: string;
  announcement?: string;
};
type TaskPanel = BasePanel & {
  kind: "task";
  taskID: string;
  segmentID: string;
  tab: string;
  unblock?: boolean;
};
type WaitPanel = BasePanel & { kind: "wait"; waitID: string; tab: string };
type ImagePanel = BasePanel & {
  kind: "images";
  event: TimelineEvent;
  view: TimelineView;
  index: number;
};
type Panel = DocumentPanel | TaskPanel | WaitPanel | ImagePanel;
function keyedText(values: string[]) {
  const counts = new Map<string, number>();
  return values.map((value) => {
    const count = (counts.get(value) ?? 0) + 1;
    counts.set(value, count);
    return { key: `${value}:${count}`, value };
  });
}
function captured(data: TimelineSessionResponse) {
  return structuredClone(data);
}
export function currentWaitObservations(
  wait: TimelineCIWait,
  candidate: { commit: string },
): Record<string, TimelineCIObservation> {
  const result: Record<string, TimelineCIObservation> = {};
  for (const observation of wait.observations) {
    const attempt = wait.attempts.find(
      (a) => a.attempt_id === observation.attempt_id,
    );
    if (
      observation.attempt_id !==
        wait.applicable_attempt_id[observation.check_key] ||
      observation.external_run.commit !== candidate.commit ||
      attempt?.external_run.commit !== candidate.commit
    )
      continue;
    if (
      !result[observation.check_key] ||
      result[observation.check_key].observed_at < observation.observed_at
    )
      result[observation.check_key] = observation;
  }
  return result;
}
function latestWait(data: TimelineSessionResponse, id: string) {
  let wait = data.session.waits.find((w) => w.wait_id === id);
  const seen = new Set<string>();
  while (wait && !seen.has(wait.wait_id)) {
    seen.add(wait.wait_id);
    const successor = data.session.waits.find(
      (w) => w.previous_wait_id === wait?.wait_id,
    );
    if (!successor) break;
    wait = successor;
  }
  return wait;
}
function StateSymbol({
  state,
  label,
}: {
  state: TimelineTaskState | TimelineWaitState;
  label: string;
}) {
  const Icon =
    state === "completed" || state === "succeeded"
      ? Check
      : state === "in_progress"
        ? Spinner
        : state === "in_review" || state === "waiting"
          ? Pause
          : state === "blocked" || state === "failed" || state === "cancelled"
            ? CircleAlert
            : state === "unknown"
              ? CircleHelp
              : Circle;
  return (
    <span
      role="img"
      aria-label={label}
      className={`inline-flex shrink-0 ${timelineStateColor(state)}`}
    >
      <Icon aria-hidden className="size-4 motion-reduce:animate-none" />
    </span>
  );
}
function SafeLink({ url, children }: { url?: string; children: ReactNode }) {
  return isSafeTimelineURL(url) ? (
    <a
      href={url}
      target="_blank"
      rel="noopener noreferrer"
      className="break-all underline underline-offset-4"
    >
      {children}
    </a>
  ) : (
    <span>{children}</span>
  );
}
function ReaderSelect({
  label,
  value,
  items,
  onChange,
}: {
  label: string;
  value: string;
  items: { value: string; label: string }[];
  onChange: (value: string) => void;
}) {
  return (
    <Select
      value={value}
      items={items}
      onValueChange={(value) => {
        if (value) onChange(value);
      }}
    >
      <SelectTrigger size="sm" aria-label={label}>
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
  );
}
function EmptyDetail({
  text,
  retry,
  label,
}: {
  text: string;
  retry: () => void;
  label: string;
}) {
  return (
    <div role="status" className="flex flex-col items-start gap-3">
      <p className="text-sm text-muted-foreground">{text}</p>
      <Button type="button" variant="outline" size="sm" onClick={retry}>
        {label}
      </Button>
    </div>
  );
}

export function useTimelinePanel({
  data,
  lang,
  loadDocument,
  onRetry,
}: {
  data: TimelineSessionResponse;
  lang: TimelineLanguage;
  loadDocument?: TimelineDocumentLoader;
  onRetry: () => void;
}) {
  const t = timelineCopy(lang);
  const [panel, setPanel] = useState<Panel | null>(null);
  const returnFocus = useRef<HTMLElement | null>(null),
    unblockHeading = useRef<HTMLHeadingElement | null>(null),
    body = useRef<HTMLDivElement | null>(null);
  const scrolls = useRef(new Map<string, number>());
  const restoreAttachment = useRef<string | undefined>(undefined);
  const live = useRef(data);
  live.current = data;
  const panelKey = panel
    ? `${panel.kind}:${panel.kind === "task" ? `${panel.taskID}:${panel.tab}` : panel.kind === "wait" ? `${panel.waitID}:${panel.tab}` : panel.kind === "document" ? `${panel.key}:${panel.revisionID}:${panel.label}` : `${panel.event.event_id}:${panel.view.view_id}`}`
    : "";
  useLayoutEffect(() => {
    if (body.current)
      body.current.scrollTop = scrolls.current.get(panelKey) ?? 0;
    if (panel?.kind === "task" && panel.unblock && panel.tab === "summary")
      unblockHeading.current?.focus({ preventScroll: true });
    if (restoreAttachment.current && panel?.kind === "task") {
      const key = restoreAttachment.current;
      const node = [
        ...document.querySelectorAll<HTMLElement>("[data-document-key]"),
      ].find((n) => n.dataset.documentKey === key);
      node?.focus({ preventScroll: true });
      restoreAttachment.current = undefined;
    }
  }, [panelKey, panel]);
  const openDetail: TimelineDetailOpener = (target, trigger) => {
    returnFocus.current = trigger;
    const snap = captured(live.current);
    setPanel(
      target.kind === "task"
        ? {
            kind: "task",
            data: snap,
            taskID: target.taskID,
            segmentID: target.segmentID,
            tab: target.tab ?? "summary",
            unblock: target.unblock,
          }
        : { kind: "wait", data: snap, waitID: target.waitID, tab: "summary" },
    );
  };
  const documentPanel = (
    snap: TimelineSessionResponse,
    eventID: string,
    key?: string,
    revisionID?: string,
    back?: Panel,
  ): DocumentPanel | undefined => {
    const event = snap.session.events.find((e) => e.event_id === eventID);
    const keys = key
      ? [key]
      : [
          ...new Set(
            event?.document_snapshots
              ?.filter((r) => !r.task_id)
              .map((r) => r.document_key) ?? [],
          ),
        ];
    const chosen = key ?? keys[0];
    const chain = documentChains(snap.session).find((c) => c.key === chosen);
    const selected =
      chain?.revisions.find(
        (r) => r.revision.document_revision_id === revisionID,
      ) ?? chain?.revisions.at(-1);
    if (!selected) return;
    return {
      kind: "document",
      data: snap,
      eventID: selected.event.event_id,
      key: chosen,
      keys,
      revisionID: selected.revision.document_revision_id,
      label: selected.revision.tabs[0]?.label ?? "",
      back,
    };
  };
  const openDocument: TimelineDocumentOpener = (
    eventID,
    trigger,
    key,
    revisionID,
  ) => {
    returnFocus.current = trigger;
    setPanel(
      documentPanel(captured(live.current), eventID, key, revisionID) ?? null,
    );
  };
  const attachedDocument = (
    eventID: string,
    key: string,
    revisionID: string,
  ) => {
    if (!panel) return;
    const next = documentPanel(panel.data, eventID, key, revisionID, panel);
    if (next) setPanel(next);
  };
  const back = () => {
    if (!panel?.back) return;
    if (panel.kind === "document") restoreAttachment.current = panel.key;
    setPanel(panel.back);
  };
  const revision =
    panel?.kind === "document"
      ? documentChains(panel.data.session)
          .find((c) => c.key === panel.key)
          ?.revisions.find(
            (r) => r.revision.document_revision_id === panel.revisionID,
          )
      : undefined;
  const task =
    panel?.kind === "task"
      ? panel.data.session.tasks.find((task) => task.task_id === panel.taskID)
      : undefined;
  const selectedTaskState =
    task && panel?.kind === "task"
      ? (task.segments.find((segment) => segment.segment_id === panel.segmentID)
          ?.state ?? task.state)
      : undefined;
  const wait =
    panel?.kind === "wait"
      ? panel.data.session.waits.find((wait) => wait.wait_id === panel.waitID)
      : undefined;
  const successor =
    panel?.kind === "document"
      ? latestSuccessor(
          documentChains(data.session).find((c) => c.key === panel.key),
          panel.revisionID,
        )
      : undefined;
  const newerTask =
    panel?.kind === "task"
      ? data.session.tasks.find((t) => t.task_id === panel.taskID)
      : undefined;
  const newerWait =
    panel?.kind === "wait" ? latestWait(data, panel.waitID) : undefined;
  const updated =
    !!panel &&
    panel.data.session.viewer_session_id === data.session.viewer_session_id &&
    (panel.kind === "document"
      ? !!successor
      : panel.kind === "task"
        ? !!task && !!newerTask && newerTask.task_version > task.task_version
        : panel.kind === "wait"
          ? !!wait &&
            !!newerWait &&
            (newerWait.wait_id !== wait.wait_id ||
              newerWait.wait_version > wait.wait_version)
          : false);
  const showLatest = () => {
    if (!panel) return;
    const snap = captured(data);
    if (panel.kind === "document" && successor) {
      const next = documentPanel(
        snap,
        successor.event.event_id,
        panel.key,
        successor.revision.document_revision_id,
        panel.back,
      );
      if (next) {
        const retained = successor.revision.tabs.some(
          (t) => t.label === panel.label,
        );
        setPanel({
          ...next,
          keys: panel.keys,
          label: retained ? panel.label : next.label,
          announcement: retained ? undefined : t.tabGone,
        });
      }
    } else if (panel.kind === "task" && newerTask)
      setPanel({ ...panel, data: snap });
    else if (panel.kind === "wait" && newerWait)
      setPanel({ ...panel, data: snap, waitID: newerWait.wait_id });
  };
  const selectRevision = (selected: DocumentRevision) => {
    if (panel?.kind !== "document") return;
    const retained = selected.revision.tabs.some(
      (tab) => tab.label === panel.label,
    );
    setPanel({
      ...panel,
      eventID: selected.event.event_id,
      revisionID: selected.revision.document_revision_id,
      label: retained ? panel.label : (selected.revision.tabs[0]?.label ?? ""),
      announcement: retained ? undefined : t.tabGone,
    });
  };
  const contentScroll = (node: HTMLDivElement | null) => {
    body.current = node;
  };
  const rememberScroll = () => {
    if (body.current) scrolls.current.set(panelKey, body.current.scrollTop);
  };
  const regionalEmpty = (text: string) => (
    <EmptyDetail text={text} label={t.reload} retry={onRetry} />
  );
  const renderImages = (event: TimelineEvent, view: TimelineView) => {
    if (!panel) return null;
    return (
      <div className="grid min-w-0 grid-cols-2 gap-3 sm:grid-cols-3">
        {view.image_refs.map((ref, index) => {
          const artifact = panel.data.session.artifacts.find(
              (a) => a.artifact_id === ref.artifact_id,
            ),
            url = panel.data.artifact_urls?.[ref.artifact_id];
          return artifact &&
            isSafeTimelineArtifactURL(
              url,
              panel.data.session.viewer_session_id,
            ) ? (
            <figure
              key={ref.artifact_id}
              className="flex min-w-0 flex-col gap-1"
            >
              <CaptureThumbnail
                url={url}
                alt={ref.alt || ref.caption || view.title}
                label={`${view.title} · ${ref.caption || `${t.captures} ${index + 1}`}`}
                onRetry={onRetry}
                retryLabel={t.retry}
                failureLabel={t.imageFailed}
                onOpen={() =>
                  setPanel({
                    kind: "images",
                    data: panel.data,
                    event,
                    view,
                    index: view.image_refs
                      .slice(0, index)
                      .filter(
                        (ref) =>
                          panel.data.session.artifacts.some(
                            (a) => a.artifact_id === ref.artifact_id,
                          ) &&
                          isSafeTimelineArtifactURL(
                            panel.data.artifact_urls?.[ref.artifact_id],
                            panel.data.session.viewer_session_id,
                          ),
                      ).length,
                    back: panel,
                  })
                }
              />
              <figcaption className="text-xs">
                {ref.caption || `${t.captures} ${index + 1}`}
              </figcaption>
            </figure>
          ) : (
            <div
              key={ref.artifact_id}
              role="status"
              className="flex flex-col gap-2"
            >
              <p className="text-sm">
                {ref.caption} · {t.unavailable}
              </p>
              <p className="text-xs text-muted-foreground">
                {t.assetUnavailableSnapshot}
              </p>
            </div>
          );
        })}
      </div>
    );
  };
  const taskEvidence = () => {
    if (panel?.kind !== "task" || !task) return null;
    const events = panel.data.session.events.filter(
        (e) =>
          e.task_id === task.task_id &&
          !e.withdrawn &&
          e.kind !== "task_state_transition",
      ),
      docs = panel.data.session.events.flatMap((event) =>
        (event.document_snapshots ?? [])
          .filter((r) => r.task_id === task.task_id)
          .map((revision) => ({ event, revision })),
      );
    const latestDocuments = [
      ...new Map(docs.map((doc) => [doc.revision.document_key, doc])).values(),
    ];
    return (
      <div className="flex flex-col gap-4">
        {!events.some((e) => e.views?.length || e.summary) &&
          !docs.length &&
          regionalEmpty(t.noAttachments)}
        {latestDocuments.map(({ event, revision }) => (
          <Button
            key={revision.document_revision_id}
            type="button"
            variant="outline"
            size="sm"
            data-document-key={revision.document_key}
            className="self-start"
            onClick={() =>
              attachedDocument(
                event.event_id,
                revision.document_key,
                revision.document_revision_id,
              )
            }
          >
            {revision.document_key} · {timelineDate(revision.created_at, lang)}
          </Button>
        ))}
        {events.map((event) => (
          <section key={event.event_id} className="flex min-w-0 flex-col gap-3">
            <h3 className="font-medium">{timelineEventTitle(event, lang)}</h3>
            <TimelineTicketLinks
              session={panel.data.session}
              refs={event.ticket_refs}
              lang={lang}
            />
            <TimelineDeliveryLinks
              session={panel.data.session}
              refs={event.delivery_refs ?? []}
              lang={lang}
            />
            <p className="text-xs text-muted-foreground">
              {(() => {
                const segment = task.segments.find(
                  (s) => s.segment_id === event.segment_id,
                );
                return segment
                  ? t.stageNames[segment.planned_stage]
                  : t.unavailableDetail;
              })()}
            </p>
            {event.summary && <TimelineMarkdown content={event.summary} />}{" "}
            {event.views?.map((view) => (
              <section
                key={view.view_id}
                className="flex min-w-0 flex-col gap-3"
              >
                <h4 className="font-medium">{view.title}</h4>
                <TimelineTicketLinks
                  session={panel.data.session}
                  refs={view.ticket_refs}
                  lang={lang}
                />
                {renderImages(event, view)}
              </section>
            ))}
          </section>
        ))}
      </div>
    );
  };
  const taskBody = (tab: string) => {
    if (!task || !panel || panel.kind !== "task")
      return regionalEmpty(t.unavailableDetail);
    const selected = task.segments.find(
        (s) => s.segment_id === panel.segmentID,
      ),
      history = panel.data.session.events
        .filter(
          (e) =>
            e.task_id === task.task_id && e.kind === "task_state_transition",
        )
        .sort((a, b) => a.sequence - b.sequence);
    return tab === "evidence" ? (
      taskEvidence()
    ) : tab === "history" ? (
      <div className="flex flex-col gap-4">
        {!history.length ? (
          <p>{t.initialRecord}</p>
        ) : (
          history.map((event) => (
            <section key={event.event_id} className="flex flex-col gap-2">
              <h3 className="font-medium">{timelineEventTitle(event, lang)}</h3>
              <p>
                {t.taskStates[
                  event.audit?.previous_state as TimelineTaskState
                ] ?? event.audit?.previous_state}{" "}
                →{" "}
                {t.taskStates[event.audit?.next_state as TimelineTaskState] ??
                  event.audit?.next_state}
              </p>
              <p>{event.audit?.reason ?? event.reason}</p>
              {event.audit?.block && (
                <ol className="list-decimal pl-5">
                  {keyedText(event.audit.block.steps).map((step) => (
                    <li key={step.key}>{step.value}</li>
                  ))}
                </ol>
              )}
              <p className="text-xs text-muted-foreground">
                {event.actor.kind === "native_producer"
                  ? panel.data.session.producers.find(
                      (p) =>
                        p.producer_id ===
                        (event.actor.kind === "native_producer"
                          ? event.actor.producer_id
                          : ""),
                    )?.source.native_session_id
                  : localViewerLabel(event.actor, t)}{" "}
                · {timelineDate(event.created_at, lang)}
              </p>
            </section>
          ))
        )}
      </div>
    ) : (
      <div className="flex flex-col gap-6">
        {task.segments
          .filter((segment) => segment.block)
          .sort(
            (a, b) =>
              Number(b.segment_id === selected?.segment_id) -
              Number(a.segment_id === selected?.segment_id),
          )
          .map(
            (segment) =>
              segment.block && (
                <section
                  key={segment.segment_id}
                  className="flex flex-col gap-3"
                >
                  <h3
                    ref={
                      segment.segment_id === selected?.segment_id
                        ? unblockHeading
                        : undefined
                    }
                    tabIndex={-1}
                    className="font-medium"
                  >
                    {t.unblocking}
                  </h3>
                  <p className="text-xs text-muted-foreground">
                    {t.stageNames[segment.planned_stage]}
                  </p>
                  <p>{segment.block.reason}</p>
                  <ol className="flex list-decimal flex-col gap-2 pl-5">
                    {keyedText(segment.block.steps).map((step) => (
                      <li key={step.key}>{step.value}</li>
                    ))}
                  </ol>
                  <p>
                    {t.assignee}: {segment.block.owner}
                  </p>
                  <p>
                    {t.resumeWhen}: {segment.block.resume_when}
                  </p>
                </section>
              ),
          )}
        {task.summary && <TimelineMarkdown content={task.summary} />}
        <p>
          {t.globalState}: {t.taskStates[task.state]}
        </p>
        <section className="flex flex-col gap-3">
          <h3 className="font-medium">{t.segments}</h3>
          <ul className="flex flex-col gap-3">
            {task.segments.map((segment) => (
              <li
                key={segment.segment_id}
                aria-current={
                  segment.segment_id === panel.segmentID ? "true" : undefined
                }
                className="flex flex-col gap-1"
              >
                <p>
                  {t.stageNames[segment.planned_stage]} ·{" "}
                  {t.taskStates[segment.state]}
                  {segment.withdrawn ? ` · ${t.withdrawn}` : ""}
                </p>
                <p className="break-all text-xs text-muted-foreground">
                  {t.assignee}:{" "}
                  {panel.data.session.producers.find(
                    (p) => p.producer_id === segment.assigned_producer_id,
                  )?.source.native_session_id ?? t.producerUnavailable}
                </p>
              </li>
            ))}
          </ul>
        </section>
        {panel.data.session.waits
          .filter((w) => w.task_id === task.task_id)
          .map((w) => (
            <Button
              key={w.wait_id}
              type="button"
              variant="outline"
              size="sm"
              className="self-start"
              onClick={() =>
                setPanel({
                  kind: "wait",
                  data: panel.data,
                  waitID: w.wait_id,
                  tab: "summary",
                  back: panel,
                })
              }
            >
              {t.ciTitle}
            </Button>
          ))}
      </div>
    );
  };
  const waitBody = (tab: string) => {
    if (!wait) return regionalEmpty(t.unavailableDetail);
    if (!panel) return null;
    const observations = currentWaitObservations(
      wait,
      timelineCandidateByID(panel.data.session, wait.candidate_id),
    );
    return tab === "checks" ? (
      <div className="flex flex-col gap-4">
        {(!wait.policy.confirmed || !wait.policy.required_checks.length) &&
          regionalEmpty(t.noChecks)}
        {wait.policy.required_checks.map((check) => {
          const observation = observations[check.check_key];
          const state =
            wait.policy.confirmed && observation
              ? observation.state
              : "unknown";
          return (
            <section key={check.check_key} className="flex flex-col gap-2">
              <h3 className="flex items-center gap-2 font-medium">
                <StateSymbol state={state} label={t.waitStates[state]} />
                {check.label} · {t.waitStates[state]}
              </h3>
              {observation && (
                <>
                  <p>{observation.summary}</p>
                  <SafeLink url={observation.external_run.url}>
                    {observation.external_run.run_id}
                  </SafeLink>
                  <p className="text-xs text-muted-foreground">
                    {t.observed}: {timelineDate(observation.observed_at, lang)}
                  </p>
                </>
              )}
            </section>
          );
        })}
      </div>
    ) : tab === "attempts" ? (
      <div className="flex flex-col gap-4">
        {!wait.attempts.length && regionalEmpty(t.noAttempts)}
        {wait.attempts.map((attempt) => (
          <section key={attempt.attempt_id} className="flex flex-col gap-2">
            <h3 className="font-medium">
              #{attempt.generation} ·{" "}
              <SafeLink url={attempt.external_run.url}>
                {attempt.external_run.run_id}
              </SafeLink>
            </h3>
            <p className="text-xs">
              {attempt.external_run.commit} ·{" "}
              {timelineDate(attempt.observed_at, lang)}
            </p>
            {wait.observations
              .filter((o) => o.attempt_id === attempt.attempt_id)
              .map((o) => (
                <p key={o.observation_id}>
                  {wait.policy.required_checks.find(
                    (c) => c.check_key === o.check_key,
                  )?.label ?? o.check_key}{" "}
                  · {t.waitStates[o.state]} · {o.summary}
                </p>
              ))}
          </section>
        ))}
      </div>
    ) : (
      <div className="flex flex-col gap-4">
        <p>
          {
            t.waitStates[
              aggregateTimelineCIWait(
                wait.policy,
                wait.applicable_attempt_id,
                observations,
              )
            ]
          }
        </p>
        {wait.next_action && <p>{wait.next_action}</p>}
        <section className="flex flex-col gap-2">
          <h3 className="font-medium">{t.requiredPolicy}</h3>
          <p>
            {wait.policy.confirmed && wait.policy.required_checks.length
              ? wait.policy.authority
              : t.noChecks}
          </p>
          <SafeLink url={wait.policy.source_url}>{t.authority}</SafeLink>
          <p className="text-xs">
            {timelineDate(wait.policy.observed_at, lang)}
          </p>
        </section>
        <p className="text-sm text-muted-foreground">{t.ciNotMerge}</p>
      </div>
    );
  };
  const documentBody = () => {
    if (panel?.kind !== "document" || !revision)
      return regionalEmpty(t.documentUnavailable);
    const chain = documentChains(panel.data.session).find(
      (c) => c.key === panel.key,
    );
    const previous = chain?.revisions.find(
      (r) =>
        r.revision.document_revision_id ===
        revision.revision.predecessor_revision_id,
    )?.revision;
    const comparison = compareRevisionTabs(revision.revision, previous);
    const tabs = revisionTabs(revision.revision, revision.event.stage);
    const tab = tabs.find((tab) =>
      tab.kind === "document"
        ? tab.tab.label === panel.label
        : tab.tabs.some((delta) => delta.label === panel.label),
    );
    const active = revision.revision.tabs.find(
      (tab) => tab.label === panel.label,
    );
    return (
      <>
        <div className="flex shrink-0 flex-wrap items-center gap-2">
          {panel.keys.length > 1 && (
            <ReaderSelect
              label={t.document}
              value={panel.key}
              items={panel.keys.map((key) => ({ value: key, label: key }))}
              onChange={(key) => {
                const next = documentPanel(
                  panel.data,
                  panel.eventID,
                  key,
                  undefined,
                  panel.back,
                );
                if (next) setPanel({ ...next, keys: panel.keys });
              }}
            />
          )}
          <ReaderSelect
            label={t.revision}
            value={panel.revisionID}
            items={[...(chain?.revisions ?? [])]
              .reverse()
              .map(({ revision: r }) => ({
                value: r.document_revision_id,
                label: `${timelineDate(r.created_at, lang)} · ${r.status === "draft" ? t.draft : t.complete}`,
              }))}
            onChange={(id) => {
              const selected = chain?.revisions.find(
                (r) => r.revision.document_revision_id === id,
              );
              if (selected) selectRevision(selected);
            }}
          />
          <Badge variant="outline">{t.published}</Badge>
        </div>
        <div className="flex shrink-0 flex-col gap-1 text-xs text-muted-foreground">
          {comparison.initial ? (
            <p>{t.initialRevision}</p>
          ) : (
            <>
              {comparison.changed.length > 0 && (
                <p>
                  {t.changed}: {comparison.changed.join(", ")}
                </p>
              )}
              {comparison.added.length > 0 && (
                <p>
                  {t.added}: {comparison.added.join(", ")}
                </p>
              )}
              {comparison.removed.length > 0 && (
                <p>
                  {t.removed}: {comparison.removed.join(", ")}
                </p>
              )}
            </>
          )}
          {panel.announcement && <p role="status">{panel.announcement}</p>}
        </div>
        <Tabs
          value={tab?.key ?? tabs[0]?.key}
          onValueChange={(key) => {
            const target = tabs.find((t) => t.key === key);
            if (target)
              setPanel({
                ...panel,
                label:
                  target.kind === "document"
                    ? target.tab.label
                    : target.tabs[0].label,
                announcement: undefined,
              });
          }}
          className="flex min-h-0 min-w-0 flex-1 flex-col"
        >
          <div className="min-w-0 shrink-0 overflow-x-auto">
            <TabsList>
              {tabs.map((tab) => (
                <TabsTrigger key={tab.key} value={tab.key}>
                  {tab.kind === "document" ? tab.tab.label : t.specifications}
                  {comparison.changed.includes(
                    tab.kind === "document" ? tab.tab.label : panel.label,
                  )
                    ? " *"
                    : ""}
                </TabsTrigger>
              ))}
            </TabsList>
          </div>
          {tabs.map((entry) => (
            <TabsContent
              key={entry.key}
              value={entry.key}
              className="flex min-h-0 min-w-0 flex-1 flex-col gap-3"
            >
              {entry.kind === "specs" && entry.tabs.length > 1 && (
                <div className="shrink-0">
                  <ReaderSelect
                    label={t.specification}
                    value={panel.label}
                    items={entry.tabs.map((t) => ({
                      value: t.label,
                      label: t.label,
                    }))}
                    onChange={(label) => setPanel({ ...panel, label })}
                  />
                </div>
              )}
              <div
                ref={contentScroll}
                onScroll={rememberScroll}
                className="min-h-0 min-w-0 flex-1 overflow-auto"
              >
                <DocumentBody
                  key={active?.document_id}
                  document={panel.data.session.documents.find(
                    (d) => d.document_id === active?.document_id,
                  )}
                  loadDocument={loadDocument}
                  lang={lang}
                  onRetry={onRetry}
                />
              </div>
            </TabsContent>
          ))}
        </Tabs>
      </>
    );
  };
  const tabbedDetail = (kind: "task" | "wait") => {
    if (!panel || panel.kind !== kind) return null;
    const tabs =
      kind === "task"
        ? [
            { key: "summary", label: t.summary },
            { key: "evidence", label: t.evidence },
            { key: "history", label: t.history },
          ]
        : [
            { key: "summary", label: t.summary },
            { key: "checks", label: t.checks },
            { key: "attempts", label: t.attempts },
          ];
    return (
      <Tabs
        value={panel.tab}
        onValueChange={(tab) => setPanel({ ...panel, tab })}
        className="flex min-h-0 min-w-0 flex-1 flex-col"
      >
        <div className="shrink-0 overflow-x-auto">
          <TabsList>
            {tabs.map((tab) => (
              <TabsTrigger key={tab.key} value={tab.key}>
                {tab.label}
              </TabsTrigger>
            ))}
          </TabsList>
        </div>
        {tabs.map((tab) => (
          <TabsContent
            key={tab.key}
            value={tab.key}
            className="flex min-h-0 flex-1 flex-col"
          >
            <div
              ref={contentScroll}
              onScroll={rememberScroll}
              className="min-h-0 flex-1 overflow-auto"
            >
              {kind === "task" ? taskBody(tab.key) : waitBody(tab.key)}
            </div>
          </TabsContent>
        ))}
      </Tabs>
    );
  };
  const images =
    panel?.kind === "images"
      ? panel.view.image_refs.flatMap((ref) => {
          const artifact = panel.data.session.artifacts.find(
              (a) => a.artifact_id === ref.artifact_id,
            ),
            url = panel.data.artifact_urls?.[ref.artifact_id];
          return artifact &&
            isSafeTimelineArtifactURL(url, panel.data.session.viewer_session_id)
            ? [{ ref, artifact, url }]
            : [];
        })
      : [];
  const selectedImage =
    panel?.kind === "images" ? images[panel.index] : undefined;
  const moveImage = (delta: number) => {
    if (panel?.kind === "images" && images.length)
      setPanel({
        ...panel,
        index: (panel.index + delta + images.length) % images.length,
      });
  };
  const title =
    panel?.kind === "document"
      ? revision && timelineEventTitle(revision.event, lang)
      : panel?.kind === "task"
        ? task
          ? timelineTaskLabel(task)
          : t.unavailableDetail
        : panel?.kind === "wait"
          ? t.ciTitle
          : panel?.kind === "images"
            ? panel.view.title
            : "";
  const sheet = (
    <Sheet
      open={!!panel}
      onOpenChange={(open) => {
        if (!open) setPanel(null);
      }}
    >
      {panel && (
        <SheetContent
          showCloseButton={false}
          finalFocus={returnFocus}
          initialFocus={
            panel.kind === "task" && panel.unblock ? unblockHeading : undefined
          }
          className="min-h-0 min-w-0 data-[side=right]:w-full data-[side=right]:sm:w-2/3 data-[side=right]:sm:max-w-none"
          onKeyDown={(event) => {
            if (panel.kind !== "images") return;
            if (event.key === "ArrowRight") {
              event.preventDefault();
              moveImage(1);
            }
            if (event.key === "ArrowLeft") {
              event.preventDefault();
              moveImage(-1);
            }
          }}
        >
          <SheetHeader className="shrink-0">
            <div className="flex min-w-0 items-start justify-between gap-2">
              <div className="flex min-w-0 items-center gap-2">
                {task && (
                  <StateSymbol
                    state={selectedTaskState ?? task.state}
                    label={t.taskStates[selectedTaskState ?? task.state]}
                  />
                )}
                <SheetTitle className="min-w-0 break-words">{title}</SheetTitle>
                {task && isHostTask(task) && <HostObservedBadge t={t} />}
              </div>
              <SheetClose
                render={
                  <Button
                    type="button"
                    variant="ghost"
                    size="icon-sm"
                    aria-label={t.close}
                  />
                }
              >
                <X aria-hidden />
              </SheetClose>
            </div>
            <SheetDescription>
              {panel.kind === "document"
                ? t.documents
                : panel.kind === "task"
                  ? t.tasks
                  : panel.kind === "wait"
                    ? t.checks
                    : t.captures}
            </SheetDescription>
          </SheetHeader>
          <div className="flex min-h-0 min-w-0 flex-1 flex-col gap-4 px-4 pb-4">
            {panel.back && (
              <Button
                type="button"
                variant="ghost"
                size="sm"
                className="self-start"
                onClick={back}
              >
                <ArrowLeft aria-hidden />
                {panel.back.kind === "task" ? t.backTask : t.backDetails}
              </Button>
            )}
            {updated && (
              <Alert variant="info">
                <AlertDescription>
                  {panel.kind === "document"
                    ? t.newRevision
                    : t.snapshotUpdated}
                </AlertDescription>
                <AlertAction>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={showLatest}
                  >
                    {t.showLatest}
                  </Button>
                </AlertAction>
              </Alert>
            )}
            {wait && (
              <div className="flex shrink-0 flex-col gap-2">
                <p className="break-all text-sm">
                  {t.candidate}:{" "}
                  <code>
                    {
                      panel.data.session.candidates.find(
                        (c) => c.candidate_id === wait.candidate_id,
                      )?.commit
                    }
                  </code>
                </p>
                <TimelineDeliveryLinks
                  session={panel.data.session}
                  refs={[wait.pull_request_ref]}
                  lang={lang}
                />
              </div>
            )}
            {panel.kind === "document" ? (
              documentBody()
            ) : panel.kind === "task" ? (
              tabbedDetail("task")
            ) : panel.kind === "wait" ? (
              tabbedDetail("wait")
            ) : selectedImage ? (
              <div
                ref={contentScroll}
                onScroll={rememberScroll}
                className="flex min-h-0 flex-1 flex-col gap-3 overflow-auto"
              >
                <CaptureImage
                  src={selectedImage.url}
                  alt={
                    selectedImage.ref.alt ||
                    selectedImage.ref.caption ||
                    panel.view.title
                  }
                  onRetry={onRetry}
                  label={t.imageFailed}
                  retryLabel={t.retry}
                  className="mx-auto max-w-full"
                />
                <p>{selectedImage.ref.caption}</p>
                <p className="text-xs text-muted-foreground">
                  {t[selectedImage.artifact.viewport]} ·{" "}
                  {t[selectedImage.artifact.theme]} ·{" "}
                  {timelineDate(selectedImage.artifact.captured_at, lang)}
                </p>
                <div className="flex items-center gap-2">
                  <Button
                    type="button"
                    variant="outline"
                    size="icon-sm"
                    aria-label={t.previous}
                    disabled={images.length < 2}
                    onClick={() => moveImage(-1)}
                  >
                    <ChevronLeft aria-hidden />
                  </Button>
                  <span className="text-xs tabular-nums">
                    {panel.index + 1}/{images.length}
                  </span>
                  <Button
                    type="button"
                    variant="outline"
                    size="icon-sm"
                    aria-label={t.next}
                    disabled={images.length < 2}
                    onClick={() => moveImage(1)}
                  >
                    <ChevronRight aria-hidden />
                  </Button>
                </div>
              </div>
            ) : (
              regionalEmpty(t.unavailable)
            )}
          </div>
        </SheetContent>
      )}
    </Sheet>
  );
  return { openDetail, openDocument, sheet };
}
