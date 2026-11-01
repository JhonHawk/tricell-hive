import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import {
  aggregateTimelineTaskState,
  parseTimelineJSON,
  parseTimelineSession,
  parseTimelineV2Request,
  type TimelineActor,
  type TimelineCIWait,
  TimelineContractError,
  type TimelineSource,
  type TimelineTaskSegment,
  type TimelineV2RequestMap,
  validateTimelineActor,
  validateTimelineAuthority,
  validateTimelineCIWait,
  validateTimelineEventCapacity,
  validateTimelineResolvedProject,
  validateTimelineV2OperationReferences,
} from "../timeline-contract";

const fixture = (name: string) =>
  readFileSync(
    new URL(`../../../testdata/timeline_contract/${name}`, import.meta.url),
    "utf8",
  );
const session = () =>
  parseTimelineSession(parseTimelineJSON(fixture("v2-session.json")));
const cases = JSON.parse(fixture("v2-cases.json")) as Array<{
  name: string;
  action: string;
  kind: string;
  object_id: string;
  valid: boolean;
  json: string;
}>;
for (const c of cases)
  test(`timeline v2 shared: ${c.name}`, () => {
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
          parseTimelineV2Request(c.kind as keyof TimelineV2RequestMap, value);
          break;
        case "references":
          validateTimelineV2OperationReferences(
            session(),
            c.kind as keyof TimelineV2RequestMap,
            c.object_id,
            value,
          );
          break;
        case "wait":
          validateTimelineCIWait(value as TimelineCIWait);
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
