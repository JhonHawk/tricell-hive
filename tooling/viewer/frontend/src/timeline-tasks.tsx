import { Button as ButtonPrimitive } from "@base-ui/react/button";
import {
  Check,
  ChevronDown,
  Circle,
  CircleAlert,
  CircleHelp,
  FileText,
  History,
  Images,
  MessageSquare,
  Pause,
} from "lucide-react";
import { type ComponentProps, type ReactNode, useId, useState } from "react";
import {
  Timeline,
  TimelineContent,
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
import { Spinner } from "@/components/ui/spinner";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import {
  isSafeTimelineArtifactURL,
  type TimelineEvent,
  type TimelineSessionResponse,
} from "./timeline-contract";
import type { TimelineCIWait, TimelineTaskState } from "./timeline-contract-v3";
import { type TimelineLanguage, timelineCopy } from "./timeline-copy";
import type { TimelineTaskProjection } from "./timeline-grouping";
import { TimelineSummary } from "./timeline-summary";
export type TimelineDetailTarget =
  | {
      kind: "task";
      taskID: string;
      segmentID: string;
      unblock?: boolean;
      tab?: "summary" | "evidence" | "history";
    }
  | { kind: "wait"; waitID: string };
export type TimelineDetailOpener = (
  target: TimelineDetailTarget,
  trigger: HTMLElement,
) => void;
// The styled Button has a fixed control height. Text titles use the same installed
// primitive so each line participates in normal layout without overriding its sizes.
function TimelineTextAction(props: ComponentProps<typeof ButtonPrimitive>) {
  return (
    <ButtonPrimitive
      type="button"
      data-slot="timeline-text-action"
      className="block w-full whitespace-normal break-words text-left text-sm font-medium transition-colors hover:text-primary focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ring"
      {...props}
    />
  );
}
// State color is carried by the point itself, as requested; row copy stays neutral.
export function timelineStateColor(state: string | undefined): string {
  if (state === "completed" || state === "succeeded" || state === "closed")
    return "text-success-foreground";
  if (state === "failed" || state === "cancelled")
    return "text-destructive-foreground";
  if (
    state === "blocked" ||
    state === "paused" ||
    state === "awaiting_validation" ||
    state === "waiting"
  )
    return "text-warning-foreground";
  if (state === "in_review") return "text-info-foreground";
  if (state === "in_progress" || state === "active") return "text-primary";
  return "text-muted-foreground";
}
export function taskContent(
  data: TimelineSessionResponse,
  taskID: string,
  segmentID?: string,
) {
  const task = data.session.tasks.find((task) => task.task_id === taskID);
  const events = data.session.events.filter(
    (event) =>
      !event.withdrawn &&
      event.task_id === taskID &&
      (!segmentID || event.segment_id === segmentID),
  );
  const documents = new Set(
    data.session.events.flatMap((event) =>
      (event.document_snapshots ?? [])
        // The reader retains immutable document chains from withdrawn occurrences.
        .filter(
          (revision) =>
            revision.task_id === taskID &&
            (!segmentID || revision.segment_id === segmentID),
        )
        .map((revision) => revision.document_key),
    ),
  );
  const refs = events.flatMap(
    (event) => event.views?.flatMap((view) => view.image_refs) ?? [],
  );
  const images = refs.filter(
    (ref) =>
      data.session.artifacts.some((a) => a.artifact_id === ref.artifact_id) &&
      isSafeTimelineArtifactURL(
        data.artifact_urls?.[ref.artifact_id],
        data.session.viewer_session_id,
      ),
  ).length;
  const history = events.filter(
    (event) => event.kind === "task_state_transition",
  ).length;
  const reviews = events.filter((event) => event.kind === "review").length;
  const blocked =
    task?.segments.some(
      (segment) =>
        Boolean(segment.block) &&
        (!segmentID || segment.segment_id === segmentID),
    ) ?? false;
  const waits = data.session.waits.filter(
    (wait) =>
      wait.task_id === taskID && (!segmentID || wait.segment_id === segmentID),
  ).length;
  const substantive =
    blocked ||
    images > 0 ||
    refs.length > 0 ||
    documents.size > 0 ||
    history > 0 ||
    reviews > 0 ||
    waits > 0 ||
    events.some((e) => Boolean(e.delivery_refs?.length) || Boolean(e.target));
  return {
    images,
    unavailable: refs.length - images,
    documents: documents.size,
    history,
    reviews,
    blocked,
    waits,
    substantive,
  };
}
export function TimelineStateIcon({
  state,
  ...props
}: { state: TimelineTaskState | TimelineCIWait["state"] } & ComponentProps<
  typeof Check
>) {
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
  return <Icon {...props} />;
}
export function TaskIndicator({
  state,
  label,
}: {
  state: TimelineTaskState | TimelineCIWait["state"];
  label: string;
}) {
  return (
    <Tooltip>
      <TooltipTrigger
        render={
          <TimelineIndicator
            aria-hidden={false}
            role="img"
            aria-label={label}
            tabIndex={0}
            className={`flex size-6 items-center justify-center border-0 bg-background group-data-[orientation=vertical]/timeline:-top-0.5 ${timelineStateColor(state)}`}
          />
        }
      >
        <TimelineStateIcon
          state={state}
          aria-hidden
          className="size-4 motion-reduce:animate-none"
        />
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}
export function TimelineWaitRow({
  wait,
  index,
  last = false,
  lang,
  onOpen,
}: {
  wait: TimelineCIWait;
  index: number;
  last?: boolean;
  lang: TimelineLanguage;
  onOpen?: TimelineDetailOpener;
}) {
  const t = timelineCopy(lang);
  const [expanded, setExpanded] = useState(false);
  const current = wait.policy.required_checks.map((c) =>
    wait.observations
      .filter(
        (o) =>
          o.check_key === c.check_key &&
          o.attempt_id === wait.applicable_attempt_id[c.check_key],
      )
      .sort((a, b) => a.observed_at.localeCompare(b.observed_at))
      .at(-1),
  );
  const remaining = current.filter((o) => o?.state !== "succeeded").length;
  return (
    <Collapsible open={expanded} onOpenChange={setExpanded}>
      <TimelineItem
        step={index + 1}
        data-wait-id={wait.wait_id}
        data-completed={wait.state === "succeeded" || undefined}
        className={
          last ? undefined : "group-data-[orientation=vertical]/timeline:pb-6"
        }
      >
        {!last && (
          <TimelineSeparator className="bg-border group-last/timeline-item:block group-data-[orientation=vertical]/timeline:top-1 group-data-[orientation=vertical]/timeline:h-[calc(100%-2.25rem)] group-data-[orientation=vertical]/timeline:translate-y-6" />
        )}
        <TaskIndicator
          state={wait.state}
          label={`CI: ${t.waitStates[wait.state]}`}
        />
        <TimelineHeader className="flex min-w-0 items-start gap-2">
          <TimelineTitle className="min-w-0 flex-1">
            <TimelineTextAction
              onClick={(e) =>
                onOpen?.(
                  { kind: "wait", waitID: wait.wait_id },
                  e.currentTarget,
                )
              }
            >
              {t.ciTitle}
            </TimelineTextAction>
          </TimelineTitle>
          <CollapsibleTrigger
            render={
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label={`${expanded ? t.collapse : t.expand} CI`}
              />
            }
          >
            <ChevronDown
              aria-hidden
              className={expanded ? "size-4" : "size-4 -rotate-90"}
            />
          </CollapsibleTrigger>
        </TimelineHeader>
        <CollapsibleContent>
          <TimelineContent className="flex flex-col items-start gap-2">
            <p className="text-sm">
              {wait.policy.confirmed
                ? `${remaining}/${wait.policy.required_checks.length} ${t.checksRemaining}`
                : t.unconfirmed}
            </p>
            {wait.next_action && <p className="text-sm">{wait.next_action}</p>}
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={(e) =>
                onOpen?.(
                  { kind: "wait", waitID: wait.wait_id },
                  e.currentTarget,
                )
              }
            >
              {t.viewChecks}
            </Button>
          </TimelineContent>
        </CollapsibleContent>
      </TimelineItem>
    </Collapsible>
  );
}
export function TimelineTaskRow({
  projection,
  data,
  initiallyExpanded = false,
  index,
  last = false,
  lang,
  onOpen,
  renderActivity,
  targetEventID,
}: {
  projection: TimelineTaskProjection;
  data: TimelineSessionResponse;
  initiallyExpanded?: boolean;
  targetEventID?: string;
  index: number;
  last?: boolean;
  lang: TimelineLanguage;
  onOpen?: TimelineDetailOpener;
  renderActivity: (event: TimelineEvent) => ReactNode;
}) {
  const { task, segment, events, waits } = projection;
  const details = taskContent(data, task.task_id, segment.segment_id);
  const hasDetail = taskContent(data, task.task_id).substantive;
  const disclosureID = useId();
  const [expanded, setExpanded] = useState(
    () =>
      initiallyExpanded ||
      projection.events.some((e) => e.event_id === targetEventID),
  );
  const t = timelineCopy(lang);
  return (
    <Collapsible open={expanded} onOpenChange={setExpanded}>
      <TimelineItem
        step={index + 1}
        data-task-id={task.task_id}
        data-segment-id={segment.segment_id}
        data-completed={segment.state === "completed" || undefined}
        className={
          last ? undefined : "group-data-[orientation=vertical]/timeline:pb-6"
        }
      >
        {!last && (
          <TimelineSeparator className="bg-border group-last/timeline-item:block group-data-[orientation=vertical]/timeline:top-1 group-data-[orientation=vertical]/timeline:h-[calc(100%-2.25rem)] group-data-[orientation=vertical]/timeline:translate-y-6" />
        )}
        <TaskIndicator
          state={segment.state}
          label={`${task.task_key}: ${t.taskStates[segment.state]}`}
        />
        <TimelineHeader className="flex min-w-0 items-start gap-2">
          <TimelineTitle className="min-w-0 flex-1">
            <TimelineTextAction
              aria-label={`${task.task_key} — ${task.title}: ${t.taskStates[segment.state]}`}
              aria-expanded={hasDetail ? undefined : expanded}
              aria-controls={hasDetail ? undefined : disclosureID}
              onClick={(e) =>
                hasDetail
                  ? onOpen?.(
                      {
                        kind: "task",
                        taskID: task.task_id,
                        segmentID: segment.segment_id,
                      },
                      e.currentTarget,
                    )
                  : setExpanded((value) => !value)
              }
            >
              {task.task_key} — {task.title}
            </TimelineTextAction>
          </TimelineTitle>
          <CollapsibleTrigger
            render={
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                aria-label={`${expanded ? t.collapse : t.expand} ${task.task_key}`}
              />
            }
          >
            <ChevronDown
              aria-hidden
              className={expanded ? "size-4" : "size-4 -rotate-90"}
            />
          </CollapsibleTrigger>
        </TimelineHeader>
        <div
          className="flex flex-wrap items-center gap-1"
          data-slot="task-content-affordance"
        >
          {details.images > 0 && (
            <Button
              type="button"
              variant="link"
              size="sm"
              onClick={(e) =>
                onOpen?.(
                  {
                    kind: "task",
                    taskID: task.task_id,
                    segmentID: segment.segment_id,
                    tab: "evidence",
                  },
                  e.currentTarget,
                )
              }
            >
              <Images aria-hidden />
              {t.imageCount(details.images)}
            </Button>
          )}
          {details.unavailable > 0 && (
            <Button
              type="button"
              variant="link"
              size="sm"
              onClick={(e) =>
                onOpen?.(
                  {
                    kind: "task",
                    taskID: task.task_id,
                    segmentID: segment.segment_id,
                    tab: "evidence",
                  },
                  e.currentTarget,
                )
              }
            >
              <Images aria-hidden />
              {t.unavailableCount(details.unavailable)}
            </Button>
          )}
          {details.documents > 0 && (
            <Button
              type="button"
              variant="link"
              size="sm"
              onClick={(e) =>
                onOpen?.(
                  {
                    kind: "task",
                    taskID: task.task_id,
                    segmentID: segment.segment_id,
                    tab: "evidence",
                  },
                  e.currentTarget,
                )
              }
            >
              <FileText aria-hidden />
              {t.documentCount(details.documents)}
            </Button>
          )}
          {details.reviews > 0 && (
            <Button
              type="button"
              variant="link"
              size="sm"
              onClick={(e) =>
                onOpen?.(
                  {
                    kind: "task",
                    taskID: task.task_id,
                    segmentID: segment.segment_id,
                    tab: "evidence",
                  },
                  e.currentTarget,
                )
              }
            >
              <MessageSquare aria-hidden />
              {t.reviewCount(details.reviews)}
            </Button>
          )}
          {details.blocked && (
            <Button
              type="button"
              variant="link"
              size="sm"
              onClick={(e) =>
                onOpen?.(
                  {
                    kind: "task",
                    taskID: task.task_id,
                    segmentID: segment.segment_id,
                    unblock: true,
                  },
                  e.currentTarget,
                )
              }
            >
              {t.unblocking}
            </Button>
          )}
          {details.waits > 0 && (
            <Button
              type="button"
              variant="link"
              size="sm"
              onClick={(e) => {
                const localWaits = data.session.waits.filter(
                  (wait) =>
                    wait.task_id === task.task_id &&
                    wait.segment_id === segment.segment_id,
                );
                onOpen?.(
                  localWaits.length === 1
                    ? { kind: "wait", waitID: localWaits[0].wait_id }
                    : {
                        kind: "task",
                        taskID: task.task_id,
                        segmentID: segment.segment_id,
                      },
                  e.currentTarget,
                );
              }}
            >
              {t.waitCount(details.waits)}
            </Button>
          )}
          {details.history > 0 && (
            <Button
              type="button"
              variant="link"
              size="sm"
              onClick={(e) =>
                onOpen?.(
                  {
                    kind: "task",
                    taskID: task.task_id,
                    segmentID: segment.segment_id,
                    tab: "history",
                  },
                  e.currentTarget,
                )
              }
            >
              <History aria-hidden />
              {t.changeCount(details.history)}
            </Button>
          )}
        </div>
        <CollapsibleContent id={disclosureID}>
          <TimelineContent className="flex min-w-0 flex-col items-start gap-3">
            {(segment.block?.reason || task.summary) && (
              <TimelineSummary
                content={segment.block?.reason ?? task.summary}
                lines={2}
                lang={lang}
              />
            )}
            {segment.block && (
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={(e) =>
                  onOpen?.(
                    {
                      kind: "task",
                      taskID: task.task_id,
                      segmentID: segment.segment_id,
                      unblock: true,
                    },
                    e.currentTarget,
                  )
                }
              >
                {t.unblock}
              </Button>
            )}
            {waits.length > 0 && (
              <Timeline value={0} className="w-full">
                {waits.map((wait, i) => (
                  <TimelineWaitRow
                    key={wait.wait_id}
                    wait={wait}
                    index={i}
                    last={i === waits.length - 1}
                    lang={lang}
                    onOpen={onOpen}
                  />
                ))}
              </Timeline>
            )}
            {events.length > 0 && (
              <Timeline value={0} className="w-full">
                {events.map(renderActivity)}
              </Timeline>
            )}
          </TimelineContent>
        </CollapsibleContent>
      </TimelineItem>
    </Collapsible>
  );
}
