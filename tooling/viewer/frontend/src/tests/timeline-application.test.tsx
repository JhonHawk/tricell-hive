import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { AccessClient, type NavigationProjection } from "../api";
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
      if (path === "/api/index?navigation=grouped") {
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
      if (path === "/api/index?navigation=grouped") {
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

test("grouped navigation preserves selection, preferences, search and isolated fallback through refresh", async (t) => {
  sessionStorage.clear();
  sessionStorage.setItem(
    "hive-view-navigation",
    JSON.stringify({
      sidebarOpen: true,
      projects: {
        "legacy-api": true,
        "legacy-workspace": false,
        untouched: false,
      },
    }),
  );
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
    sessionStorage.clear();
  });
  const a = structuredClone(timelineFixture),
    b = structuredClone(timelineFixture);
  a.session.project_key = "legacy-api";
  a.session.title = "Review integrations";
  b.session.viewer_session_id = "b".repeat(64);
  b.session.project_key = "legacy-workspace";
  b.session.title = "Audit navigation";
  let group = `explicit:${"c".repeat(64)}`;
  let broken = false;
  let indexReads = 0;
  const projectIndex = () => {
    const timeline_sessions = [a, b].map(({ session }) => {
      const {
        schema_version: _s,
        events: _e,
        artifacts: _a,
        documents: _d,
        producers: _p,
        foreground_producer_id: _lead,
        pending_join_requests: _j,
        tasks: _t,
        waits: _w,
        deliveries: _deliveries,
        candidates: _c,
        work_key: _work,
        ...entry
      } = session;
      return { mode: "timeline_api", ...entry };
    });
    const navigation: NavigationProjection = {
      version: 1,
      entries: [a, b].map(({ session }, i) => ({
        viewer_session_id: session.viewer_session_id,
        group_key: group,
        group_name: "Explicit workspace",
        origin_kind: i ? "workspace" : "repository",
        origin_label: i ? "workspace-root" : "project-api",
        config_status: "valid",
      })),
    };
    return {
      protocol_version: 3,
      retention: "manual",
      ttl_seconds: 0,
      repositories: [],
      errors: [],
      timeline_sessions,
      navigation: broken ? { ...navigation, version: 99 } : navigation,
    };
  };
  const controller = new ViewerController(
    new AccessClient(
      async (path) => {
        if (path === "/api/index?navigation=grouped") {
          indexReads++;
          return new Response(JSON.stringify(projectIndex()));
        }
        return new Response(
          JSON.stringify(
            String(path).endsWith(a.session.viewer_session_id) ? a : b,
          ),
        );
      },
      { getItem: () => "test-token", setItem() {}, removeItem() {} },
    ),
    () => false,
  );
  await controller.open(null);
  await act(async () => {
    root.render(<ViewerApplication controller={controller} />);
    await settle();
  });
  const nav = () => container.querySelector("nav");
  const rows = () => [
    ...container.querySelectorAll<HTMLButtonElement>("[data-session-key]"),
  ];
  assert.equal(
    nav()?.querySelectorAll('[data-slot="collapsible-trigger"]').length,
    1,
  );
  assert.equal(rows().length, 2);
  assert.ok(nav()?.textContent?.includes("Workspace"));
  assert.ok(nav()?.textContent?.includes("project-api"));
  assert.ok(rows()[0].title.includes(a.session.source.project_root));
  await act(async () => {
    rows()[1].click();
    await settle();
  });
  assert.equal(
    controller.getSnapshot().selectedId,
    b.session.viewer_session_id,
  );
  const trigger = () =>
    nav()?.querySelector<HTMLButtonElement>(
      '[data-slot="collapsible-trigger"]',
    );
  await act(async () => {
    trigger()?.click();
    await settle();
  });
  assert.equal(trigger()?.getAttribute("aria-expanded"), "false");
  await act(async () => {
    await controller.refresh();
    await settle();
  });
  assert.equal(trigger()?.getAttribute("aria-expanded"), "false");
  group = `explicit:${"d".repeat(64)}`;
  sessionStorage.setItem("unused", "preserved");
  await act(async () => {
    await controller.refresh();
    await settle();
  });
  assert.equal(
    controller.getSnapshot().selectedId,
    b.session.viewer_session_id,
  );
  assert.equal(trigger()?.getAttribute("aria-expanded"), "true");
  const input = nav()?.querySelector<HTMLInputElement>('input[type="search"]');
  assert.ok(input);
  const search = async (value: string) =>
    act(async () => {
      Object.getOwnPropertyDescriptor(
        window.HTMLInputElement.prototype,
        "value",
      )?.set?.call(input, value);
      input.dispatchEvent(new Event("input", { bubbles: true }));
      await settle();
    });
  await search("Audit navigation");
  assert.equal(rows().length, 1);
  assert.ok(trigger()?.textContent?.includes("1"));
  await search("project-api");
  assert.equal(rows().length, 0, "origin does not broaden title/ticket search");
  await search("");
  broken = true;
  await act(async () => {
    await controller.refresh();
    await settle();
  });
  assert.ok(nav()?.textContent?.includes("Could not read project grouping"));
  assert.equal(
    nav()?.querySelectorAll('[data-slot="collapsible-trigger"]').length,
    2,
  );
  assert.equal(
    controller.getSnapshot().selectedId,
    b.session.viewer_session_id,
  );
  broken = false;
  const retry = [
    ...(nav()?.querySelectorAll<HTMLButtonElement>("button") ?? []),
  ].find((node) => node.textContent === "Retry");
  assert.ok(retry);
  const before = indexReads;
  await act(async () => {
    retry.click();
    await settle();
  });
  assert.ok(indexReads > before);
  assert.equal(
    nav()?.querySelectorAll('[data-slot="collapsible-trigger"]').length,
    1,
  );
  assert.ok(!nav()?.textContent?.includes("Could not read project grouping"));
  const preferences = JSON.parse(
    sessionStorage.getItem("hive-view-navigation") ?? "{}",
  );
  assert.equal(preferences.projects.untouched, false);
  assert.equal(sessionStorage.getItem("unused"), "preserved");
});
