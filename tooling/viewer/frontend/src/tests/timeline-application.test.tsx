import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { AccessClient } from "../api";
import { ViewerApplication } from "../application";
import { ViewerController } from "../controller";
import { loadPreviewDocument } from "../dev/timeline-documents-fixture";
import { timelineFixture } from "../dev/timeline-fixture";

async function settle() {
  for (let i = 0; i < 60; i++) await Promise.resolve();
}
function button(label: string) {
  const node = [...document.querySelectorAll<HTMLButtonElement>("button")].find(
    (b) =>
      b.getAttribute("aria-label") === label || b.textContent?.trim() === label,
  );
  assert.ok(node, `Missing button ${label}`);
  return node;
}
async function click(label: string) {
  await act(async () => {
    button(label).click();
    await settle();
  });
}

test("real application wires API timeline, documents, conditional title and v2 navigation with one controller", async (t) => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  const data = structuredClone(timelineFixture);
  const sid = data.session.viewer_session_id;
  let savedVersion: number | undefined;
  const paths: string[] = [];
  const client = new AccessClient(
    async (path, options) => {
      paths.push(String(path));
      if (path === "/api/index") {
        const {
          schema_version: _s,
          events: _e,
          artifacts: _a,
          documents: _d,
          producers: _producers,
          foreground_producer_id: _lead,
          pending_join_requests: _joins,
          tasks: _tasks,
          waits: _waits,
          deliveries: _deliveries,
          candidates: _candidates,
          work_key: _workKey,
          ...entry
        } = data.session;
        return new Response(
          JSON.stringify({
            protocol_version: 3,
            retention: "manual",
            repositories: [],
            errors: [],
            ttl_seconds: 0,
            timeline_sessions: [{ mode: "timeline_api", ...entry }],
          }),
        );
      }
      if (String(path).includes("/documents/"))
        return new Response(
          JSON.stringify(
            await loadPreviewDocument(String(path).split("/").at(-1) as string),
          ),
        );
      if (options?.method === "PATCH") {
        const body = JSON.parse(String(options.body));
        savedVersion = body.expected_version;
        data.session.title = body.title;
        data.session.version++;
        return new Response(
          JSON.stringify({
            schema_version: 3,
            kind: "session_patch",
            viewer_session_id: sid,
            operation_id: body.operation_id,
            session_version: data.session.version,
            current_stage_event_id: data.session.current_stage_event_id,
          }),
        );
      }
      return new Response(JSON.stringify(data));
    },
    { getItem: () => "test-token", setItem() {}, removeItem() {} },
  );
  const controller = new ViewerController(client, () => false);
  await controller.open(null, { sessionID: sid, eventID: "event_plan_r3" });
  await settle();
  await act(async () => {
    root.render(<ViewerApplication controller={controller} />);
    await new Promise((resolve) => setTimeout(resolve, 15));
  });
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 15));
  });
  assert.ok(container.querySelector('[aria-label="Session timeline"]'));
  assert.equal(document.activeElement?.id, "timeline-event-event_plan_r3");
  await click("Edit title");
  const input = container.querySelector<HTMLInputElement>(
    '[aria-label="Session title"]',
  );
  assert.ok(input);
  await act(async () => {
    const setter = Object.getOwnPropertyDescriptor(
      window.HTMLInputElement.prototype,
      "value",
    )?.set;
    setter?.call(input, "Conditional API title");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  const version = data.session.version;
  await click("Save title");
  assert.equal(savedVersion, version);
  assert.ok(container.textContent?.includes("Conditional API title"));
  await act(async () => {
    const target = container.querySelector<HTMLButtonElement>(
      "#timeline-event-event_plan_r3 button",
    );
    const plan = [
      ...container.querySelectorAll<HTMLButtonElement>(
        "#timeline-event-event_plan_r3 button",
      ),
    ].find((b) => b.textContent?.trim() === "View plan");
    assert.ok(target);
    assert.ok(plan);
    plan.click();
    await settle();
  });
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 10));
  });
  assert.ok(
    document
      .querySelector('[role="dialog"]')
      ?.textContent?.includes("Propuesta · R3"),
  );
  assert.ok(
    paths.some((path) => path.includes("/documents/document_proposal")),
  );
  await click("Close");
  assert.equal(
    paths.some(
      (path) => path.startsWith("/api/sessions/") || path.startsWith("/files/"),
    ),
    false,
  );
});

test("pending API deletion exposes recovery outside the removed session header", async (t) => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  const data = structuredClone(timelineFixture),
    sid = data.session.viewer_session_id;
  let tombstoned = false,
    attempts = 0;
  const bodies: unknown[] = [];
  const client = new AccessClient(
    async (path, options) => {
      if (path === "/api/index") {
        const {
          schema_version: _s,
          events: _e,
          artifacts: _a,
          documents: _d,
          producers: _producers,
          foreground_producer_id: _lead,
          pending_join_requests: _joins,
          tasks: _tasks,
          waits: _waits,
          deliveries: _deliveries,
          candidates: _candidates,
          work_key: _workKey,
          ...entry
        } = data.session;
        return new Response(
          JSON.stringify({
            protocol_version: 3,
            retention: "manual",
            repositories: [],
            errors: [],
            ttl_seconds: 0,
            timeline_sessions: tombstoned
              ? []
              : [{ mode: "timeline_api", ...entry }],
          }),
        );
      }
      if (options?.method === "DELETE") {
        const body = JSON.parse(String(options.body));
        bodies.push(body);
        tombstoned = true;
        attempts++;
        return new Response(
          JSON.stringify(
            attempts === 1
              ? { code: "delete_pending", cleanup_pending: true }
              : {
                  schema_version: 3,
                  kind: "session_delete",
                  viewer_session_id: sid,
                  operation_id: body.operation_id,
                  session_version: data.session.version + 1,
                  current_stage_event_id: null,
                  deleted: true,
                },
          ),
          { status: attempts === 1 ? 503 : 200 },
        );
      }
      return new Response(JSON.stringify(data));
    },
    { getItem: () => "test-token", setItem: () => {}, removeItem: () => {} },
  );
  const controller = new ViewerController(client, () => false);
  await controller.open(null, { sessionID: sid });
  await settle();
  await act(async () => {
    root.render(<ViewerApplication controller={controller} />);
    await settle();
  });
  await act(async () => {
    await assert.rejects(
      controller.deleteTimelineSession(sid, data.session.version),
    );
    await settle();
  });
  assert.equal(controller.getSnapshot().selectedId, undefined);
  const notice = [...container.querySelectorAll('[role="status"]')].find(
    (node) => node.textContent?.includes("Deletion incomplete"),
  );
  assert.ok(
    notice,
    "Recovery remains visible after the session header unmounts",
  );
  const retry = notice.querySelector<HTMLButtonElement>("button");
  assert.ok(retry);
  await act(async () => {
    retry.click();
    await settle();
  });
  assert.equal(attempts, 2);
  assert.deepEqual(bodies[0], bodies[1]);
  assert.equal(controller.getSnapshot().pendingTimelineDelete, undefined);
  assert.ok(!container.textContent?.includes("Deletion incomplete"));
});
