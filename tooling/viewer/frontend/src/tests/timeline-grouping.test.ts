import assert from "node:assert/strict";
import test from "node:test";
import { completeFixture } from "../dev/timeline-complete-fixture";
import { timelineFixture } from "../dev/timeline-fixture";
import type { TimelineTask } from "../timeline-contract-v3";
import {
  distinctCaptions,
  eventBodyTexts,
  groupTimeline,
  taskProgress,
} from "../timeline-grouping";

test("work-level activity after a stage sits below it, not above every stage", () => {
  const s = structuredClone(timelineFixture.session);
  const source = s.events[0];
  s.events = [
    { ...source, event_id: "plan", kind: "stage", stage: "plan", sequence: 1 },
    {
      ...source,
      event_id: "plan-done",
      kind: "state_transition",
      stage_event_id: "plan",
      sequence: 2,
    },
    { ...source, event_id: "paused", kind: "state_transition", sequence: 3 },
    { ...source, event_id: "build", kind: "stage", stage: "build", sequence: 4 },
  ];
  assert.deepEqual(
    groupTimeline(s).map((x) => x.stage?.event_id ?? x.events[0]?.event_id),
    ["plan", "paused", "build"],
  );
});
test("an audit reason copied into title and summary is shown once", () => {
  const reason = "Delivery paused at the user's request";
  assert.deepEqual(
    eventBodyTexts({ title: reason, summary: reason, audit: { reason } }),
    { summary: undefined, reason: undefined, auditReason: undefined },
  );
  assert.deepEqual(
    eventBodyTexts({ title: "", summary: "Plan rewritten", reason: "Done" }),
    { summary: "Plan rewritten", reason: "Done", auditReason: undefined },
  );
});

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
    plan_order: i,
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

for (const change of ["rebind", "task withdrawal", "segment withdrawal"]) {
  test(`historical task activity remains visible after ${change}`, () => {
    const session = completeFixture().session;
    const task = session.tasks.find((item) => item.task_key === "T7");
    assert.ok(task);
    const segment = task.segments.find(
      (item) => item.planned_stage === "delivery",
    );
    assert.ok(segment?.display_stage_event_id);
    const stage = session.events.find(
      (item) => item.event_id === segment.display_stage_event_id,
    );
    assert.ok(stage);
    const historicalEvents = session.events.filter(
      (item) =>
        item.task_id === task.task_id &&
        item.segment_id === segment.segment_id &&
        item.stage_event_id === stage.event_id,
    );
    const historicalWaits = session.waits.filter(
      (item) =>
        item.task_id === task.task_id &&
        item.segment_id === segment.segment_id &&
        item.stage_event_id === stage.event_id,
    );
    assert.ok(historicalEvents.length > 0);
    assert.ok(historicalWaits.length > 0);
    if (change === "rebind") {
      session.events.push({
        ...structuredClone(stage),
        event_id: "delivery-reentry",
        sequence:
          Math.max(...session.events.map((event) => event.sequence)) + 1,
      });
      segment.display_stage_event_id = "delivery-reentry";
    } else if (change === "task withdrawal") {
      task.withdrawn = true;
    } else {
      segment.withdrawn = true;
      task.state = "completed";
    }
    const progress = taskProgress(session.tasks);
    const groups = groupTimeline(session);
    const historical = groups.find((group) => group.id === stage.event_id);
    assert.ok(historical);
    assert.equal(historical.tasks.length, 0);
    for (const event of historicalEvents) {
      assert.ok(
        historical.events.some((item) => item.event_id === event.event_id),
      );
    }
    for (const wait of historicalWaits) {
      assert.ok(historical.waits.some((item) => item.wait_id === wait.wait_id));
    }
    const visibleEvents = groups.flatMap((group) => [
      ...group.events,
      ...group.tasks.flatMap((projection) => projection.events),
    ]);
    const visibleWaits = groups.flatMap((group) => [
      ...group.waits,
      ...group.tasks.flatMap((projection) => projection.waits),
    ]);
    assert.equal(
      new Set(visibleEvents.map((event) => event.event_id)).size,
      visibleEvents.length,
    );
    assert.equal(
      new Set(visibleWaits.map((wait) => wait.wait_id)).size,
      visibleWaits.length,
    );
    assert.deepEqual(taskProgress(session.tasks), progress);
    assert.equal(progress.total, change === "task withdrawal" ? 6 : 7);
  });
}
