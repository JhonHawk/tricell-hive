import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { AccessClient } from "../api";
import { ViewerApplication } from "../application";
import { ViewerController } from "../controller";
import { timelineFixture } from "../dev/timeline-fixture";
import type { TimelineSessionResponse } from "../timeline-contract";

const settle = async () => {
  for (let i = 0; i < 60; i++) await Promise.resolve();
};
function indexEntry(data: TimelineSessionResponse) {
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
    work_key: _workKey,
    ...rest
  } = data.session;
  return { mode: "timeline_api", ...rest };
}
function fixture(id: string) {
  const d = structuredClone(timelineFixture);
  d.session.viewer_session_id = id;
  d.session.events = [d.session.events[0]];
  d.session.tasks = [];
  d.session.waits = [];
  d.session.documents = [];
  d.session.artifacts = [];
  d.session.current_stage_event_id = null;
  d.session.validation_count = 0;
  delete d.artifact_urls;
  delete d.document_urls;
  return d;
}
test("work A/B scroll restores independently and polling append never jumps or steals focus", async (t) => {
  const a = fixture("a".repeat(64)),
    b = fixture("b".repeat(64));
  const all = [a, b];
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const client = new AccessClient(
    async (path) =>
      new Response(
        JSON.stringify(
          path === "/api/index"
            ? {
                protocol_version: 3,
                retention: "manual",
                ttl_seconds: 0,
                repositories: [],
                errors: [],
                timeline_sessions: all.map(indexEntry),
              }
            : all.find((d) =>
                String(path).endsWith(d.session.viewer_session_id),
              ),
        ),
      ),
    { getItem: () => "test-token", setItem() {}, removeItem() {} },
  );
  const controller = new ViewerController(client, () => false);
  await controller.open(null);
  await settle();
  await act(async () => {
    root.render(<ViewerApplication controller={controller} />);
    await settle();
  });
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  const canvas = container.querySelector<HTMLElement>(
    '[data-slot="viewer-canvas"]',
  );
  assert.ok(canvas);
  canvas.scrollTop = 144;
  canvas.dispatchEvent(new Event("scroll"));
  await act(async () => {
    controller.select(b.session.viewer_session_id);
    await settle();
  });
  canvas.scrollTop = 222;
  canvas.dispatchEvent(new Event("scroll"));
  await act(async () => {
    controller.select(a.session.viewer_session_id);
    await settle();
  });
  assert.equal(canvas.scrollTop, 144);
  const current = [
    ...container.querySelectorAll<HTMLButtonElement>("button"),
  ].find((e) => e.textContent?.trim() === "Go to current");
  assert.ok(current);
  current.focus();
  a.session.version++;
  a.session.events.push({
    ...a.session.events[0],
    event_id: "append-new",
    sequence: 2,
    title: "New activity",
  });
  await act(async () => {
    await controller.refresh();
    await settle();
  });
  assert.equal(canvas.scrollTop, 144);
  assert.equal(document.activeElement, current);
  await act(async () => {
    current.click();
    await settle();
  });
  assert.equal(document.activeElement?.id, "timeline-event-append-new");
});
