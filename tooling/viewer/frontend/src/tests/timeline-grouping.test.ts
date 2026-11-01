import assert from "node:assert/strict";
import test from "node:test";
import { timelineFixture } from "../dev/timeline-fixture";
import type { TimelineTask } from "../timeline-contract-v2";
import {
  distinctCaptions,
  groupTimeline,
  taskProgress,
} from "../timeline-grouping";

test("ascending stages retain old occurrence activity and closure is last", () => {
  const s = structuredClone(timelineFixture.session);
  const source = s.events[0];
  s.events = [
    { ...source, event_id: "close", kind: "closure", sequence: 9 },
    {
      ...source,
      event_id: "old",
      kind: "stage",
      stage: "build",
      stage_state: "completed",
      sequence: 2,
    },
    {
      ...source,
      event_id: "new",
      kind: "stage",
      stage: "verify",
      stage_state: "in_progress",
      sequence: 4,
    },
    { ...source, event_id: "continued", stage_event_id: "old", sequence: 8 },
    { ...source, event_id: "pre", sequence: 1 },
  ];
  const g = groupTimeline(s);
  assert.deepEqual(
    g.map((x) => x.stage?.event_id ?? x.events[0]?.event_id),
    ["pre", "old", "new", "close"],
  );
  assert.deepEqual(
    g[1].events.map((e) => e.event_id),
    ["continued"],
  );
});
test("task projections use explicit stage bindings and count unique work tasks", () => {
  const s = structuredClone(timelineFixture.session);
  const source = s.events[0];
  s.events = [
    {
      ...source,
      event_id: "build",
      kind: "stage",
      stage: "build",
      sequence: 1,
    },
    {
      ...source,
      event_id: "verify",
      kind: "stage",
      stage: "verify",
      sequence: 2,
    },
  ];
  const task = (i: number): TimelineTask => ({
    task_id: `task-${i}`,
    task_key: `T${i}`,
    title: `Task ${i}`,
    summary: "",
    task_version: 1,
    state: "completed",
    segments: [
      {
        segment_id: `segment-${i}`,
        planned_stage: "build",
        display_stage_event_id: "build",
        order: i,
        state: "completed",
        assigned_producer_id: "producer-original",
      },
    ],
  });
  s.tasks = [1, 2, 3, 4, 5, 6].map(task);
  s.tasks.push({
    ...task(7),
    state: "in_progress",
    segments: [
      {
        ...task(7).segments[0],
        planned_stage: "verify",
        display_stage_event_id: "verify",
        state: "completed",
      },
      {
        ...task(7).segments[0],
        segment_id: "delivery-7",
        planned_stage: "delivery",
        display_stage_event_id: undefined,
        state: "not_started",
      },
    ],
  });
  const g = groupTimeline(s);
  assert.equal(g.length, 2);
  assert.equal(g[0].tasks.length, 6);
  assert.equal(g[1].tasks.length, 1);
  assert.deepEqual(taskProgress(s.tasks), { completed: 6, total: 7 });
  s.events.push({
    ...source,
    event_id: "evidence",
    stage_event_id: "verify",
    task_id: "task-7",
    segment_id: "segment-7",
    sequence: 3,
  });
  assert.equal(groupTimeline(s)[1].events.length, 0);
  assert.equal(groupTimeline(s)[1].tasks[0].events.length, 1);
});
test("captions strip common context while retaining nonempty duplicate and empty fallbacks", () => {
  assert.deepEqual(
    distinctCaptions(["Chat · Desktop", "Chat · Mobile"], "Capture"),
    ["Desktop", "Mobile"],
  );
  assert.deepEqual(distinctCaptions(["Same", "Same", ""], "Capture"), [
    "Same · 1",
    "Same · 2",
    "Capture 3",
  ]);
});
