import {
  Check,
  ChevronDown,
  Circle,
  CircleAlert,
  CircleHelp,
  Pause,
} from "lucide-react";
import { type ReactNode, useState } from "react";
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
import type { TimelineEvent } from "./timeline-contract";
import type { TimelineCIWait, TimelineTaskState } from "./timeline-contract-v2";
import { type TimelineLanguage, timelineCopy } from "./timeline-copy";
import type { TimelineTaskProjection } from "./timeline-grouping";
import { TimelineSummary } from "./timeline-summary";
export type TimelineDetailTarget =
  | { kind: "task"; taskID: string; segmentID: string; unblock?: boolean }
  | { kind: "wait"; waitID: string };
export type TimelineDetailOpener = (
  target: TimelineDetailTarget,
  trigger: HTMLElement,
) => void;
export function TaskIndicator({
  state,
  label,
}: {
  state: TimelineTaskState | TimelineCIWait["state"];
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
    <Tooltip>
      <TooltipTrigger
        render={
          <TimelineIndicator
            aria-hidden={false}
            role="img"
            aria-label={label}
            tabIndex={0}
          />
        }
      >
        <Icon aria-hidden className="size-3 motion-reduce:animate-none" />
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}
export function TimelineWaitRow({
  wait,
  index,
  lang,
  onOpen,
}: {
  wait: TimelineCIWait;
  index: number;
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
      >
        <TimelineSeparator />
        <TaskIndicator
          state={wait.state}
          label={`CI: ${t.waitStates[wait.state]}`}
        />
        <TimelineHeader className="flex min-w-0 items-start gap-2">
          <TimelineTitle className="min-w-0 flex-1">
            <Button
              type="button"
              variant="ghost"
              className="max-w-full whitespace-normal text-left"
              onClick={(e) =>
                onOpen?.(
                  { kind: "wait", waitID: wait.wait_id },
                  e.currentTarget,
                )
              }
            >
              {t.ciTitle}
            </Button>
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
  index,
  lang,
  onOpen,
  renderActivity,
  targetEventID,
}: {
  projection: TimelineTaskProjection;
  targetEventID?: string;
  index: number;
  lang: TimelineLanguage;
  onOpen?: TimelineDetailOpener;
  renderActivity: (event: TimelineEvent) => ReactNode;
}) {
  const { task, segment, events, waits } = projection;
  const [expanded, setExpanded] = useState(() =>
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
      >
        <TimelineSeparator />
        <TaskIndicator
          state={segment.state}
          label={`${task.task_key}: ${t.taskStates[segment.state]}`}
        />
        <TimelineHeader className="flex min-w-0 items-start gap-2">
          <TimelineTitle className="min-w-0 flex-1">
            <Button
              type="button"
              variant="ghost"
              className="max-w-full whitespace-normal break-words text-left"
              aria-label={`${task.task_key} — ${task.title}: ${t.taskStates[segment.state]}`}
              onClick={(e) =>
                onOpen?.(
                  {
                    kind: "task",
                    taskID: task.task_id,
                    segmentID: segment.segment_id,
                  },
                  e.currentTarget,
                )
              }
            >
              {task.task_key} — {task.title}
            </Button>
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
        <CollapsibleContent>
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
