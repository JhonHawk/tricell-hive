// Illustrative development data; never a measurement of live viewer work.
import type {
  TimelineAPIIndexEntry,
  TimelineSessionResponse,
  TimelineState,
} from "../timeline-contract";
import { timelineFixture } from "./timeline-fixture";

const titles = [
  "Portal choice borders and keyboard selection",
  "Transactional delivery evidence and validation",
  "Graph API catalog and provider synchronization",
  "Conversation filters with long customer labels",
  "Invitation recovery and profile permissions",
  "Release checklist and integration follow-ups",
  "Repository navigation across the workspace",
  "Document snapshots and revision comparison",
];
const states: TimelineState[] = [
  "active",
  "in_review",
  "awaiting_validation",
  "paused",
  "closed",
  "failed",
];
export function navigationFixture(): TimelineSessionResponse[] {
  return Array.from({ length: 24 }, (_, i) => {
    const data = JSON.parse(
      JSON.stringify(timelineFixture).replaceAll(
        timelineFixture.session.viewer_session_id,
        (i + 1).toString(16).padStart(64, "0"),
      ),
    ) as TimelineSessionResponse;
    const group = i % 3;
    const session = data.session;
    session.project_name = [
      "Customer portal",
      "Delivery tools",
      "Workspace services",
    ][group];
    session.project_key = `fixture-project-${group}`;
    session.title = `${titles[i % titles.length]}${i >= titles.length ? ` · ${i + 1}` : ""}`;
    session.state = states[i % states.length];
    const stage = session.events.find(
      (event) => event.event_id === session.current_stage_event_id,
    );
    if (stage) {
      stage.stage_state =
        session.state === "active"
          ? "in_progress"
          : session.state === "closed"
            ? "completed"
            : session.state;
      if (["completed", "failed"].includes(stage.stage_state))
        stage.finished_at = stage.updated_at;
    }
    if (["paused", "closed", "failed"].includes(session.state))
      session.current_stage_event_id = null;
    session.created_at = "2026-11-01T08:00:00Z";
    session.last_activity_at = new Date(
      Date.UTC(2026, 10, 4, 12, 44) - i * 37 * 60_000,
    ).toISOString();
    session.updated_at = session.last_activity_at;
    session.closed_at =
      session.state === "closed" ? session.last_activity_at : null;
    session.autoarchive_at = session.closed_at
      ? new Date(Date.parse(session.closed_at) + 7_200_000).toISOString()
      : null;
    session.archived = i === 23;
    session.archive_origin = session.archived ? "manual" : null;
    return data;
  });
}
export function navigationIndexEntry({
  session,
}: TimelineSessionResponse): TimelineAPIIndexEntry {
  const {
    schema_version: _schema,
    events: _events,
    artifacts: _artifacts,
    documents: _documents,
    producers: _producers,
    foreground_producer_id: _lead,
    pending_join_requests: _joins,
    tasks: _tasks,
    waits: _waits,
    deliveries: _deliveries,
    candidates: _candidates,
    work_key: _work,
    ...entry
  } = session;
  return { mode: "timeline_api", ...entry };
}
