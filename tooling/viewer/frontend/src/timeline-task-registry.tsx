import { ClipboardList } from "lucide-react";
import { Alert, AlertAction, AlertDescription } from "@/components/reui/alert";
import { Button } from "@/components/ui/button";
import {
  Popover,
  PopoverContent,
  PopoverTitle,
  PopoverTrigger,
} from "@/components/ui/popover";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import type { TimelineSession } from "./timeline-contract";
import {
  isHostTask,
  type TimelineLanguage,
  timelineCopy,
  timelineTaskLabel,
  timelineTaskName,
} from "./timeline-copy";
import { taskProgress } from "./timeline-grouping";
import {
  HostObservedMarker,
  TimelineStateIcon,
  timelineStateColor,
} from "./timeline-tasks";

/** This read-only projection retains the canonical registry objects and IDs. */
export function workTaskRegistry(session: TimelineSession) {
  return session.tasks
    .filter((task) => !task.withdrawn)
    .sort((a, b) => a.plan_order - b.plan_order)
    .map((task) => {
      const segments = task.segments.filter((segment) => !segment.withdrawn);
      return {
        task,
        completed: segments.filter((segment) => segment.state === "completed")
          .length,
        total: segments.length,
      };
    });
}
export function TimelineTaskRegistry({
  session,
  lang,
  onRetry,
  loading = false,
  error = false,
}: {
  session: TimelineSession;
  lang: TimelineLanguage;
  onRetry: () => void;
  loading?: boolean;
  error?: boolean;
}) {
  const t = timelineCopy(lang),
    registry = workTaskRegistry(session),
    progress = taskProgress(session.tasks);
  return (
    <Popover>
      <PopoverTrigger
        render={<Button type="button" variant="outline" size="sm" />}
      >
        <ClipboardList aria-hidden />
        {t.tasks}
      </PopoverTrigger>
      <PopoverContent
        align="start"
        className="min-h-0 max-h-[var(--available-height)]"
      >
        <PopoverTitle>{t.tasks}</PopoverTitle>
        {progress.total > 0 && (
          <p className="text-xs tabular-nums">
            {progress.completed}/{progress.total} {t.completedTasks}
          </p>
        )}
        {error && (
          <Alert variant="destructive">
            <AlertDescription>{t.loadFailed}</AlertDescription>
            <AlertAction>
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={onRetry}
              >
                {t.retry}
              </Button>
            </AlertAction>
          </Alert>
        )}
        {loading ? (
          <div
            role="status"
            aria-label={t.loading}
            className="flex flex-col gap-3"
          >
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
            <Skeleton className="h-8 w-full" />
          </div>
        ) : registry.length ? (
          <section
            aria-label={t.taskRegistry}
            // biome-ignore lint/a11y/noNoninteractiveTabindex: The bounded registry is a keyboard-scrollable reading region.
            tabIndex={0}
            className="min-h-0 overflow-auto"
          >
            <ul className="flex flex-col gap-4 py-1">
              {registry.map(({ task, completed, total }) => (
                <li
                  key={task.task_id}
                  data-registry-task-id={task.task_id}
                  className="flex min-w-0 items-center gap-2"
                >
                  <Tooltip>
                    <TooltipTrigger
                      render={
                        <span
                          role="img"
                          aria-label={`${timelineTaskName(task)}: ${t.taskStates[task.state]}`}
                          // biome-ignore lint/a11y/noNoninteractiveTabindex: Keyboard focus exposes the read-only state tooltip.
                          tabIndex={0}
                          className={`inline-flex shrink-0 ${timelineStateColor(task.state)}`}
                        />
                      }
                    >
                      <TimelineStateIcon
                        state={task.state}
                        aria-hidden
                        className="size-4 motion-reduce:animate-none"
                      />
                    </TooltipTrigger>
                    <TooltipContent>
                      {t.taskStates[task.state]}
                      {isHostTask(task) && ` · ${t.hostObservedHint}`}
                    </TooltipContent>
                  </Tooltip>
                  <p
                    className="min-w-0 flex-1 truncate text-sm"
                    title={timelineTaskLabel(task)}
                  >
                    {timelineTaskLabel(task)}
                  </p>
                  {isHostTask(task) && <HostObservedMarker t={t} />}
                  {total > 0 && (
                    <span
                      data-task-segment-progress
                      title={`${completed}/${total} ${t.segmentProgress}`}
                      className="shrink-0 text-xs text-muted-foreground tabular-nums"
                    >
                      {completed}/{total}
                      <span className="sr-only"> {t.segmentProgress}</span>
                    </span>
                  )}
                </li>
              ))}
            </ul>
          </section>
        ) : (
          <div role="status" className="flex flex-col items-start gap-3">
            <p className="text-sm text-muted-foreground">{t.noTasks}</p>
            <Button type="button" variant="outline" size="sm" onClick={onRetry}>
              {t.reload}
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  );
}
