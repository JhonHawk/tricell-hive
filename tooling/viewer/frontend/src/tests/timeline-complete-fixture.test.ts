import assert from "node:assert/strict";
import test from "node:test";
import { completeFixture } from "../dev/timeline-complete-fixture";
import { loadPreviewDocument } from "../dev/timeline-documents-fixture";
import {
  parseTimelineDocumentResponse,
  parseTimelineSessionResponse,
} from "../timeline-contract";
import { aggregateTimelineCIWait } from "../timeline-contract-v2";
import { currentWaitObservations } from "../timeline-detail-panel";
import { groupTimeline, taskProgress } from "../timeline-grouping";

test("comprehensive review is one valid v2 work with five stages and existing asset presentations", async () => {
  const data = parseTimelineSessionResponse(completeFixture()),
    s = data.session,
    groups = groupTimeline(s);
  const visualExamples = s.events.find(
    (event) => event.event_id === "visual-examples",
  );
  assert.ok(
    visualExamples && !visualExamples.stage_event_id && !visualExamples.task_id,
  );
  const firstStage = s.events.find((event) => event.kind === "stage");
  assert.ok(firstStage);
  assert.ok(visualExamples.sequence < firstStage.sequence);
  assert.deepEqual(
    visualExamples.views?.map((view) => view.image_refs.length),
    [2, 1],
  );
  assert.deepEqual(
    groups.flatMap((g) => g.stage?.stage ?? []),
    ["research", "plan", "build", "verify", "delivery"],
  );
  assert.deepEqual(
    new Set(s.tasks.flatMap((t) => t.segments.map((s) => s.state))),
    new Set([
      "not_started",
      "in_progress",
      "in_review",
      "blocked",
      "completed",
    ]),
  );
  assert.deepEqual(taskProgress(s.tasks), { completed: 4, total: 9 });
  assert.equal(groups.find((g) => g.stage?.stage === "build")?.tasks.length, 6);
  assert.equal(
    groups.find((g) => g.stage?.stage === "verify")?.tasks[0].task.task_id,
    "task-7",
  );
  assert.equal(
    groups.find((g) => g.stage?.stage === "delivery")?.tasks[0].task.task_id,
    "task-7",
  );
  assert.ok(
    s.events.some(
      (e) => !e.task_id && e.views?.some((v) => v.image_refs.length === 1),
    ),
  );
  assert.ok(
    s.events.some(
      (e) => e.task_id && e.views?.some((v) => v.image_refs.length > 1),
    ),
  );
  assert.ok(s.events.some((e) => e.views?.some((v) => v.supersedes)));
  assert.ok(
    s.events.some((e) => e.views?.some((v) => v.image_refs.some((r) => r.alt))),
  );
  assert.ok(
    s.events.some((e) =>
      e.views?.some((v) =>
        v.image_refs.some((r) => !data.artifact_urls?.[r.artifact_id]),
      ),
    ),
  );
  assert.ok(s.artifacts.some((a) => a.viewport === "mobile"));
  assert.ok(s.artifacts.some((a) => a.viewport === "desktop"));
  assert.ok(
    s.tasks.some((t) => !s.events.some((e) => e.task_id === t.task_id)),
  );
  for (const stage of s.events.filter((e) => e.kind === "stage"))
    assert.ok(stage.document_snapshots?.some((r) => !r.task_id));
  const policy = s.waits[0];
  assert.equal(policy.attempts.length, 2);
  assert.equal(
    aggregateTimelineCIWait(
      policy.policy,
      policy.applicable_attempt_id,
      currentWaitObservations(policy),
    ),
    "waiting",
  );
  assert.ok(policy.observations.some((o) => o.state === "failed"));
  assert.ok(policy.observations.some((o) => o.state === "cancelled"));
  for (const doc of s.documents) {
    const parsed = await parseTimelineDocumentResponse(
      await loadPreviewDocument(doc.document_id),
    );
    assert.equal(parsed.document.sha256, doc.sha256);
  }
});
