import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import { timelineCandidateByID } from "../timeline-catalog";
import {
  aggregateTimelineTaskState,
  parseTimelineHostEvent,
  parseTimelineJSON,
  parseTimelineSession,
  parseTimelineV3Request,
  TIMELINE_CONTRACT_REVISION,
  type TimelineActor,
  type TimelineCIWait,
  TimelineContractError,
  type TimelineSource,
  type TimelineTaskSegment,
  type TimelineV3RequestMap,
  validateTimelineActor,
  validateTimelineAuthority,
  validateTimelineCIWait,
  validateTimelineEventCapacity,
  validateTimelineResolvedProject,
  validateTimelineV3OperationReferences,
} from "../timeline-contract";

const fixture = (name: string) =>
  readFileSync(
    new URL(`../../../testdata/timeline_contract/${name}`, import.meta.url),
    "utf8",
  );
const session = () =>
  parseTimelineSession(parseTimelineJSON(fixture("v3-session.json")));
const cases = JSON.parse(fixture("v3-cases.json")) as Array<{
  name: string;
  action: string;
  kind: string;
  object_id: string;
  valid: boolean;
  json: string;
}>;
for (const c of cases)
  test(`timeline v3 shared: ${c.name}`, () => {
    const check = () => {
      const value = parseTimelineJSON(c.json);
      switch (c.action) {
        case "resolved_project": {
          const context = value as {
            source: TimelineSource;
            resolved_key: string;
          };
          validateTimelineResolvedProject(
            session(),
            context.source,
            context.resolved_key,
          );
          break;
        }
        case "graph":
          parseTimelineSession(value);
          break;
        case "shape":
          parseTimelineV3Request(c.kind as keyof TimelineV3RequestMap, value);
          break;
        case "references":
          validateTimelineV3OperationReferences(
            session(),
            c.kind as keyof TimelineV3RequestMap,
            c.object_id,
            value,
          );
          break;
        case "wait":
          validateTimelineCIWait(
            value as TimelineCIWait,
            timelineCandidateByID(
              session(),
              (value as TimelineCIWait).candidate_id,
            ),
          );
          break;
        case "authority":
          validateTimelineAuthority(
            session(),
            c.kind,
            value as TimelineActor,
            "producer-original",
          );
          break;
        default:
          assert.fail("unknown fixture action");
      }
    };
    if (c.valid) assert.doesNotThrow(check);
    else assert.throws(check, TimelineContractError);
  });
test("task aggregate counts a multi-stage task once and retains independent states", () => {
  for (const [states, want] of [
    [["completed", "not_started"], "in_progress"],
    [["completed", "completed"], "completed"],
    [["blocked", "in_review"], "blocked"],
    [["completed", "in_review"], "in_review"],
    [["not_started"], "not_started"],
  ] as const) {
    assert.equal(
      aggregateTimelineTaskState(
        states.map((state) => ({ state }) as TimelineTaskSegment),
      ),
      want,
    );
  }
});
test("closure capacity is reserved without pruning", () => {
  assert.doesNotThrow(() => validateTimelineEventCapacity(499, 1, false));
  assert.throws(() => validateTimelineEventCapacity(500, 1, false));
  assert.doesNotThrow(() => validateTimelineEventCapacity(500, 8, true));
  assert.throws(() => validateTimelineEventCapacity(508, 1, true));
});
test("actor union rejects mixed branches", () => {
  assert.throws(() =>
    validateTimelineActor({
      kind: "local_viewer",
      producer_id: "producer-original",
    }),
  );
  assert.throws(() => validateTimelineActor({ kind: "native_producer" }));
});

test("local_viewer actor accepts only a declared origin", () => {
  for (const origin of ["browser", "cli"])
    assert.doesNotThrow(() =>
      validateTimelineActor({ kind: "local_viewer", origin }),
    );
  assert.doesNotThrow(() => validateTimelineActor({ kind: "local_viewer" }));
  for (const origin of ["robot", "", null, 1])
    assert.throws(() =>
      validateTimelineActor({ kind: "local_viewer", origin }),
    );
  assert.throws(() =>
    validateTimelineActor({
      kind: "native_producer",
      producer_id: "producer-original",
      origin: "cli",
    }),
  );
});

test("canonical project validation preserves actual first source and alias roots", () => {
  const s = session(),
    before = structuredClone(s);
  assert.equal(s.producers[1].source.project_root, "/fixture/project-worktree");
  assert.equal(
    s.pending_join_requests[0].source.project_root,
    "/fixture/project-alias",
  );
  validateTimelineResolvedProject(s, s.producers[1].source, s.project_key);
  assert.deepEqual(s, before);
});

test("historical task evidence survives rebinding while new inputs require current binding", async () => {
  const { validateTimelineTaskReference } = await import(
    "../timeline-contract-v3"
  );
  const s = session();
  const event = s.events[2];
  event.task_id = "task-seven";
  event.segment_id = "segment-verify";
  event.stage_event_id = "stage-verify";
  const revision = s.events[0].document_snapshots?.[0];
  assert.ok(revision);
  revision.task_id = "task-seven";
  revision.segment_id = "segment-verify";
  s.events.push({
    ...structuredClone(s.events[0]),
    event_id: "stage-verify-next",
    related_event_id: "stage-verify",
    sequence: 4,
    document_snapshots: [],
  });
  s.tasks[0].segments[0].display_stage_event_id = "stage-verify-next";
  assert.doesNotThrow(() => parseTimelineSession(s));
  assert.throws(
    () =>
      validateTimelineTaskReference(
        s,
        "task-seven",
        "segment-verify",
        "stage-verify",
      ),
    TimelineContractError,
  );
  s.tasks[0].segments[0].withdrawn = true;
  s.tasks[0].state = aggregateTimelineTaskState(s.tasks[0].segments);
  assert.doesNotThrow(() => parseTimelineSession(s));
  assert.throws(
    () =>
      validateTimelineTaskReference(
        s,
        "task-seven",
        "segment-verify",
        "stage-verify-next",
      ),
    TimelineContractError,
  );
});

test("global pause can synchronize a foreground stage while closure still requires clearing it", async () => {
  const { validateTimelineSessionPatch } = await import("../timeline-contract");
  const s = session();
  s.current_stage_event_id = "stage-delivery";
  s.events[1].stage_state = "in_progress";
  const request = {
    schema_version: 3 as const,
    operation_id: "pause-foreground",
    expected_version: s.version,
    actor: {
      kind: "native_producer" as const,
      producer_id: s.foreground_producer_id,
    },
    state: "paused" as const,
    reason: "User paused this work",
  };
  assert.doesNotThrow(() => validateTimelineSessionPatch(s, request));
  assert.throws(
    () => validateTimelineSessionPatch(s, { ...request, state: "closed" }),
    TimelineContractError,
  );
});

test("a known CI run cannot become applicable again through a different display URL", () => {
  const s = session();
  const wait = s.waits[0];
  const run = wait.attempts[0].external_run;
  const request = {
    schema_version: 3,
    operation_id: "repromote-known-run",
    actor: { kind: "native_producer", producer_id: s.foreground_producer_id },
    expected_session_version: s.version,
    expected_wait_version: wait.wait_version,
    reason: "Observe provider retry",
    check_keys: ["tests"],
    external_run: { ...run, url: `${run.url}?view=checks` },
    source_url: run.url,
    observed_at: s.updated_at,
  };
  assert.throws(
    () =>
      validateTimelineV3OperationReferences(
        s,
        "wait_attempt_start",
        wait.wait_id,
        request,
      ),
    (error: unknown) =>
      error instanceof TimelineContractError &&
      error.message.includes("attempt_already_known"),
  );
});

test("stored CI attempts reject a known run with a different display URL", () => {
  const s = session();
  const wait = s.waits[0];
  wait.attempts.push({
    ...structuredClone(wait.attempts[0]),
    attempt_id: "attempt-repromoted",
    generation: 3,
    external_run: {
      ...wait.attempts[0].external_run,
      url: "https://example.test/runs/1?view=checks",
    },
  });
  assert.throws(
    () =>
      validateTimelineCIWait(
        wait,
        timelineCandidateByID(session(), wait.candidate_id),
      ),
    (error: unknown) =>
      error instanceof TimelineContractError &&
      error.message.includes("attempt_already_known"),
  );
});

test("lifecycle capacity always retains room to finish foreground and close", async () => {
  const { validateTimelineClosureReserve } = await import(
    "../timeline-contract-v3"
  );
  const s = session();
  s.events = Array.from({ length: 505 }, () => structuredClone(s.events[0]));
  s.current_stage_event_id = "stage-delivery";
  assert.doesNotThrow(() => validateTimelineClosureReserve(s));
  s.events.push(structuredClone(s.events[0]));
  assert.throws(() => validateTimelineClosureReserve(s), TimelineContractError);
  s.current_stage_event_id = null;
  assert.doesNotThrow(() => validateTimelineClosureReserve(s));
  s.state = "failed";
  assert.throws(() => validateTimelineClosureReserve(s), TimelineContractError);
  s.state = "closed";
  assert.doesNotThrow(() => validateTimelineClosureReserve(s));
});

test("timeline contract revision is 3 and host events are a host_observed origin", () => {
  assert.equal(TIMELINE_CONTRACT_REVISION, 3);
  const s = session();
  for (const e of s.events) (e as { origin: string }).origin = "host_observed";
  assert.doesNotThrow(() => parseTimelineSession(structuredClone(s)));
  (s.events[0] as { origin: string }).origin = "hook";
  assert.throws(
    () => parseTimelineSession(structuredClone(s)),
    TimelineContractError,
  );
});

test("a stored task may carry a host task key but an agent request may not", () => {
  const s = session();
  s.tasks[0].task_key = "host:7";
  assert.doesNotThrow(() => parseTimelineSession(structuredClone(s)));
  s.tasks[0].task_key = `host:${"a".repeat(101)}`;
  assert.throws(
    () => parseTimelineSession(structuredClone(s)),
    TimelineContractError,
  );
  s.tasks[0].task_key = "host:";
  assert.throws(
    () => parseTimelineSession(structuredClone(s)),
    TimelineContractError,
  );
  const create = {
    schema_version: 3,
    operation_id: "op-one",
    actor: { kind: "native_producer", producer_id: "producer-original" },
    expected_session_version: 5,
    reason: "Explicit authorized change",
    plan_order: 2,
    task_key: "host:7",
    title: "Mine",
    summary: "",
    segments: [{ planned_stage: "build", order: 1 }],
  };
  assert.throws(
    () => parseTimelineV3Request("task_create", create),
    TimelineContractError,
  );
});

const hostEventCases = JSON.parse(
  fixture("host-event-cases.json"),
) as typeof cases;
for (const c of hostEventCases)
  test(`timeline host event shared: ${c.name}`, () => {
    const check = () => {
      const value = parseTimelineJSON(c.json);
      if (c.action === "shape") parseTimelineHostEvent(value);
      else if (c.action === "authority")
        validateTimelineAuthority(
          session(),
          c.kind,
          value as TimelineActor,
          "",
        );
      else assert.fail("unknown fixture action");
    };
    if (c.valid) assert.doesNotThrow(check);
    else assert.throws(check, TimelineContractError);
  });
