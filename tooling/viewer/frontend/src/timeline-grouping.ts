import type { TimelineEvent, TimelineSession } from "./timeline-contract";
import type {
  TimelineCIWait,
  TimelineTask,
  TimelineTaskSegment,
} from "./timeline-contract-v2";
export type TimelineTaskProjection = {
  task: TimelineTask;
  segment: TimelineTaskSegment;
  events: TimelineEvent[];
  waits: TimelineCIWait[];
};
export type TimelineGroup = {
  id: string;
  stage?: TimelineEvent;
  events: TimelineEvent[];
  tasks: TimelineTaskProjection[];
  waits: TimelineCIWait[];
};
export function groupTimeline(session: TimelineSession): TimelineGroup[] {
  const events = [...session.events].sort((a, b) => a.sequence - b.sequence);
  const stages = events.filter((e) => e.kind === "stage");
  const project = (stage: TimelineEvent): TimelineGroup => {
    const tasks = session.tasks
      .filter((t) => !t.withdrawn)
      .flatMap((task) =>
        task.segments
          .filter(
            (s) => !s.withdrawn && s.display_stage_event_id === stage.event_id,
          )
          .map((segment) => ({
            task,
            segment,
            events: events.filter(
              (e) =>
                e.task_id === task.task_id &&
                e.segment_id === segment.segment_id &&
                e.stage_event_id === stage.event_id,
            ),
            waits: session.waits.filter(
              (w) =>
                w.task_id === task.task_id &&
                w.segment_id === segment.segment_id &&
                w.stage_event_id === stage.event_id,
            ),
          })),
      )
      .sort((a, b) => a.segment.order - b.segment.order);
    return {
      id: stage.event_id,
      stage,
      tasks,
      events: events.filter(
        (e) => e.stage_event_id === stage.event_id && !e.task_id,
      ),
      waits: session.waits.filter(
        (w) => w.stage_event_id === stage.event_id && !w.task_id,
      ),
    };
  };
  const standalone = events.filter(
    (e) => e.kind !== "stage" && !e.stage_event_id,
  );
  return [
    ...standalone
      .filter((e) => e.kind !== "closure")
      .map((e) => ({ id: e.event_id, events: [e], tasks: [], waits: [] })),
    ...stages.map(project),
    ...standalone
      .filter((e) => e.kind === "closure")
      .map((e) => ({ id: e.event_id, events: [e], tasks: [], waits: [] })),
  ];
}
export function taskProgress(tasks: TimelineTask[]) {
  const active = tasks.filter((t) => !t.withdrawn);
  return {
    completed: active.filter((t) => t.state === "completed").length,
    total: active.length,
  };
}
export function distinctCaptions(
  captions: string[],
  fallback: string,
): string[] {
  const clean = captions.map((s) => s.trim());
  const parts = clean.map((s) => s.split(/\s+[·|—]\s+/));
  let prefix = 0;
  while (
    parts.length > 1 &&
    parts.every((p) => p.length > prefix + 1 && p[prefix] === parts[0][prefix])
  )
    prefix++;
  const labels = clean.map((s, i) =>
    s
      ? prefix
        ? parts[i].slice(prefix).join(" · ")
        : s
      : `${fallback} ${i + 1}`,
  );
  return labels.map((s, i) =>
    labels.filter((other) => other === s).length > 1
      ? `${s} · ${labels.slice(0, i + 1).filter((other) => other === s).length}`
      : s,
  );
}
export function groupMustStayOpen(
  group: TimelineGroup,
  current: string | null,
) {
  return (
    group.stage?.event_id === current ||
    ["in_progress", "in_review", "awaiting_validation", "paused"].includes(
      group.stage?.stage_state ?? "",
    ) ||
    group.tasks.some((p) =>
      ["in_progress", "in_review", "blocked"].includes(p.segment.state),
    )
  );
}
