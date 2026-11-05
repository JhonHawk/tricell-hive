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
import {
  navigationFixture,
  navigationIndexEntry,
} from "../dev/timeline-navigation-fixture";

function revisionResponse(body: BodyInit, init?: ResponseInit) {
  return new Response(body, {
    ...init,
    headers: { "X-Hive-Timeline-Revision": "3" },
  });
}
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
        return revisionResponse(
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
        return revisionResponse(
          JSON.stringify(
            await loadPreviewDocument(String(path).split("/").at(-1) as string),
          ),
        );
      if (options?.method === "PATCH") {
        const body = JSON.parse(String(options.body));
        savedVersion = body.expected_version;
        data.session.title = body.title;
        data.session.version++;
        return revisionResponse(
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
      return revisionResponse(JSON.stringify(data));
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
        return revisionResponse(
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
        return revisionResponse(
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
      return revisionResponse(JSON.stringify(data));
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
          return revisionResponse(JSON.stringify(projectIndex()));
        }
        return revisionResponse(
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
  assert.equal(
    rows()[0].querySelector('[data-slot="badge"]')?.textContent,
    "In progress",
  );
  await act(async () => {
    rows()[1].focus();
    await settle();
  });
  b.session.last_activity_at = "2026-11-01T22:00:00Z";
  b.session.updated_at = b.session.last_activity_at;
  b.session.state = "awaiting_validation";
  await act(async () => {
    await controller.refresh();
    await settle();
  });
  assert.equal(rows()[0].dataset.sessionKey, b.session.viewer_session_id);
  assert.equal(
    document.activeElement,
    rows()[0],
    "reordering preserves focused identity",
  );
  assert.equal(
    rows()[0].querySelector('[data-slot="badge"]')?.textContent,
    "Awaiting validation",
  );
  assert.ok(
    rows()[0]
      .querySelector("time")
      ?.getAttribute("title")
      ?.includes("Last activity"),
  );

  await act(async () => {
    rows()[0].click();
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

test("archive notifications expire without replay, restore their original work and clear on direct restoration", async (t) => {
  const hidden = Object.getOwnPropertyDescriptor(document, "hidden");
  Object.defineProperty(document, "hidden", {
    configurable: true,
    value: false,
  });
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
    if (hidden) Object.defineProperty(document, "hidden", hidden);
    else Reflect.deleteProperty(document, "hidden");
  });
  const a = structuredClone(timelineFixture),
    b = JSON.parse(
      JSON.stringify(timelineFixture).replaceAll(
        timelineFixture.session.viewer_session_id,
        "b".repeat(64),
      ),
    ) as typeof timelineFixture;
  a.session.title = "Archive notifications";
  b.session.viewer_session_id = "b".repeat(64);
  b.session.title = "Review original authority";
  let failRestore = false;
  let delayedIndex: Promise<void> | undefined;
  const writes: { id: string; archived: boolean; expectedVersion: number }[] =
    [];
  const controller = new ViewerController(
    new AccessClient(
      async (path, options) => {
        if (path === "/api/index?navigation=grouped") {
          await delayedIndex;
          return revisionResponse(
            JSON.stringify({
              protocol_version: 3,
              retention: "manual",
              repositories: [],
              errors: [],
              ttl_seconds: 0,
              timeline_sessions: [a, b].map(({ session }) => {
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
              }),
            }),
          );
        }
        const data = String(path).endsWith(a.session.viewer_session_id) ? a : b;
        if (options?.method === "PATCH") {
          const body = JSON.parse(String(options.body));
          writes.push({
            id: data.session.viewer_session_id,
            archived: body.archived,
            expectedVersion: body.expected_version,
          });
          if (failRestore && body.archived === false)
            return revisionResponse(JSON.stringify({ code: "unavailable" }), {
              status: 503,
            });
          assert.equal(body.expected_version, data.session.version);
          data.session.archived = body.archived;
          data.session.archive_origin = body.archived ? "manual" : null;
          data.session.version++;
          return revisionResponse(
            JSON.stringify({
              schema_version: 3,
              kind: "session_patch",
              viewer_session_id: data.session.viewer_session_id,
              operation_id: body.operation_id,
              session_version: data.session.version,
              current_stage_event_id: null,
            }),
          );
        }
        return revisionResponse(JSON.stringify(data));
      },
      { getItem: () => "test-token", setItem() {}, removeItem() {} },
    ),
    () => false,
  );
  await controller.open(null);
  const waitFor = async (condition: () => boolean, timeout = 1000) => {
    const deadline = performance.now() + timeout;
    while (!condition()) {
      assert.ok(
        performance.now() < deadline,
        "notification condition timed out",
      );
      await act(async () => {
        await new Promise((resolve) => setTimeout(resolve, 5));
      });
    }
  };
  const run = async (operation: () => unknown) => {
    await act(async () => {
      await operation();
      await settle();
    });
  };
  const notification = () =>
    container.querySelector<HTMLElement>(
      '[data-sonner-toast]:not([data-removed="true"])',
    );
  await run(() => root.render(<ViewerApplication controller={controller} />));
  const archivedAt = performance.now();
  await run(() =>
    controller.updateTimelineSession(
      a.session.viewer_session_id,
      { archived: true },
      a.session.version,
    ),
  );
  await waitFor(() => notification() !== null);
  assert.ok(
    notification(),
    "archive uses the real Sonner notification surface",
  );
  assert.ok(notification()?.textContent?.includes(a.session.title));
  assert.ok(notification()?.textContent?.includes("Undo"));
  assert.equal(
    controller.getSnapshot().selectedId,
    a.session.viewer_session_id,
  );
  await waitFor(() => notification() === null, 8500);
  assert.ok(
    performance.now() - archivedAt >= 8000,
    "Undo remains available for eight seconds",
  );
  assert.equal(notification(), null);
  assert.equal(controller.getSnapshot().undoArchive, undefined);
  await run(() => controller.refresh());
  await run(() => controller.select(b.session.viewer_session_id));
  assert.equal(
    notification(),
    null,
    "polling and navigation do not replay an expired archive",
  );

  await run(() =>
    controller.updateTimelineSession(
      b.session.viewer_session_id,
      { archived: true },
      b.session.version,
    ),
  );
  await run(() => controller.setShowArchived(true));
  await run(() => controller.select(a.session.viewer_session_id));
  b.session.version++;
  await waitFor(() => notification()?.textContent?.includes("Undo") === true);
  failRestore = true;
  await run(() => button("Undo").click());
  await waitFor(
    () =>
      notification()?.textContent?.includes(
        "Could not complete the operation. Try again.",
      ) === true,
  );
  assert.ok(
    notification()?.textContent?.includes(
      "Could not complete the operation. Try again.",
    ),
  );
  assert.ok(notification()?.textContent?.includes(b.session.title));
  failRestore = false;
  await run(() => button("Retry").click());
  await waitFor(() => notification() === null);
  assert.equal(b.session.archived, false);
  assert.equal(writes.at(-1)?.id, b.session.viewer_session_id);
  assert.equal(writes.at(-1)?.expectedVersion, b.session.version - 1);
  assert.equal(notification(), null);

  await run(() =>
    controller.updateTimelineSession(
      b.session.viewer_session_id,
      { archived: true },
      b.session.version,
    ),
  );
  await run(() =>
    controller.updateTimelineSession(
      b.session.viewer_session_id,
      { archived: false },
      b.session.version,
    ),
  );
  await waitFor(() => notification() === null);
  assert.equal(
    notification(),
    null,
    "direct restore removes the stale Undo action",
  );
  await run(() => controller.refresh());
  assert.equal(notification(), null);

  await run(() =>
    controller.updateTimelineSession(
      b.session.viewer_session_id,
      { archived: true },
      b.session.version,
    ),
  );
  await waitFor(() => notification() !== null);
  const close = notification()?.querySelector<HTMLButtonElement>(
    '[data-close-button="true"]',
  );
  assert.ok(close);
  await run(() => close.click());
  await waitFor(() => notification() === null);
  assert.equal(controller.getSnapshot().undoArchive, undefined);
  assert.equal(notification(), null);
  await run(() =>
    controller.updateTimelineSession(
      b.session.viewer_session_id,
      { archived: false },
      b.session.version,
    ),
  );
  await run(() =>
    controller.updateTimelineSession(
      b.session.viewer_session_id,
      { archived: true },
      b.session.version,
    ),
  );
  await waitFor(() => notification()?.textContent?.includes("Undo") === true);
  let releaseIndex!: () => void;
  delayedIndex = new Promise<void>((resolve) => {
    releaseIndex = resolve;
  });
  await run(() => button("Undo").click());
  // The external index response arrives after the original Undo window has closed.
  const responseTimer = setTimeout(() => {
    delayedIndex = undefined;
    releaseIndex();
  }, 8100);
  t.after(() => {
    clearTimeout(responseTimer);
    releaseIndex();
  });
  await waitFor(() => notification()?.dataset.type === "loading");
  assert.equal(notification()?.querySelector("[data-close-button]"), null);
  assert.equal(notification()?.querySelector("[data-action]"), null);
  await waitFor(() => b.session.archived === false, 9000);
  await waitFor(() => notification() === null);
  assert.equal(
    writes.at(-1)?.id,
    b.session.viewer_session_id,
    "an accepted Undo completes even when its authority refresh crosses expiry",
  );
});

test("six visible states, automatic archive retention and Restore remain local to selected work", async (t) => {
  sessionStorage.clear();
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
    sessionStorage.clear();
  });
  const works = navigationFixture();
  const work = works[0];
  let mismatch = false;
  const writes: unknown[] = [];
  const controller = new ViewerController(
    new AccessClient(
      async (path, options) => {
        if (mismatch)
          return revisionResponse(JSON.stringify({ code: "reload_required" }), {
            status: 409,
          });
        if (path === "/api/health") return revisionResponse("{}");
        if (path === "/api/index?navigation=grouped")
          return revisionResponse(
            JSON.stringify({
              protocol_version: 3,
              retention: "manual",
              ttl_seconds: 0,
              repositories: [],
              errors: [],
              timeline_sessions: works.map(navigationIndexEntry),
            }),
          );
        const selected =
          works.find(({ session }) =>
            String(path).endsWith(session.viewer_session_id),
          ) ?? work;
        if (options?.method === "PATCH") {
          const body = JSON.parse(String(options.body));
          writes.push(body);
          selected.session.archived = body.archived;
          selected.session.archive_origin = body.archived ? "manual" : null;
          selected.session.version++;
          return revisionResponse(
            JSON.stringify({
              schema_version: 3,
              kind: "session_patch",
              viewer_session_id: selected.session.viewer_session_id,
              operation_id: body.operation_id,
              session_version: selected.session.version,
              current_stage_event_id: selected.session.current_stage_event_id,
            }),
          );
        }
        return revisionResponse(JSON.stringify(selected));
      },
      {
        getItem: (key) => (key === "hive-view-token" ? "test-token" : null),
        setItem() {},
        removeItem() {},
      },
    ),
    () => false,
  );
  await controller.open(null);
  await act(async () => {
    root.render(<ViewerApplication controller={controller} />);
    await settle();
  });
  const rows = () => [
    ...container.querySelectorAll<HTMLButtonElement>("[data-session-key]"),
  ];
  assert.equal(rows().length, 23);
  for (const label of [
    "In progress",
    "In review",
    "Awaiting validation",
    "Paused",
    "Finished",
    "Failed",
  ])
    assert.ok(
      rows().some(
        (row) =>
          row.querySelector('[data-slot="badge"]')?.textContent === label,
      ),
      label,
    );
  await act(async () => {
    rows()[0].focus();
    await settle();
  });
  work.session.archived = true;
  work.session.archive_origin = "automatic";
  const stage = work.session.events.find(
    (event) => event.event_id === work.session.current_stage_event_id,
  );
  if (stage) {
    stage.stage_state = "completed";
    stage.finished_at = stage.updated_at;
  }
  work.session.state = "closed";
  work.session.current_stage_event_id = null;
  work.session.closed_at = work.session.last_activity_at;
  work.session.version++;
  await act(async () => {
    await controller.refresh();
    await settle();
  });
  assert.equal(
    controller.getSnapshot().selectedId,
    work.session.viewer_session_id,
  );
  assert.equal(rows().length, 22);
  assert.equal(
    document.activeElement?.id,
    "document",
    "archival retains keyboard access to the opened work",
  );
  assert.equal(controller.getSnapshot().undoArchive, undefined);
  assert.ok(
    container
      .querySelector('[data-slot="timeline-header"]')
      ?.textContent?.includes(
        "Automatically archived two hours after completion.",
      ),
  );
  assert.ok(container.querySelector('[aria-label="Session timeline"]'));
  await click("Restore");
  assert.equal(rows().length, 23);
  assert.equal(work.session.archived, false);
  assert.equal(writes.length, 1);
  mismatch = true;
  await act(async () => {
    await controller.refresh();
    await settle();
  });
  assert.ok(
    container.textContent?.includes(
      "Update Hive and reload the viewer to continue.",
    ),
  );
  assert.ok(button("Reload viewer"));
  assert.equal(button("Archive").disabled, true);
  await assert.rejects(
    controller.updateTimelineSession(
      work.session.viewer_session_id,
      { archived: true },
      work.session.version,
    ),
  );
  assert.equal(writes.length, 1, "revision mismatch never replays a mutation");
});

test("first index revision mismatch replaces cold loading with actionable reload and stops requests", async (t) => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  const requests: string[] = [];
  const controller = new ViewerController(
    new AccessClient(
      async (path) => {
        requests.push(String(path));
        return revisionResponse(JSON.stringify({ code: "reload_required" }), {
          status: 409,
        });
      },
      {
        getItem: (key) => (key === "hive-view-token" ? "test-token" : null),
        setItem() {},
        removeItem() {},
      },
    ),
    () => false,
  );
  await controller.open(null);
  await act(async () => {
    root.render(<ViewerApplication controller={controller} />);
    await settle();
  });
  assert.equal(controller.getSnapshot().index, null);
  assert.equal(controller.getSnapshot().reloadRequired, true);
  assert.ok(
    container.textContent?.includes(
      "Update Hive and reload the viewer to continue.",
    ),
  );
  assert.ok(button("Reload viewer"));
  assert.equal(
    container
      .querySelector('[data-slot="viewer-canvas"]')
      ?.getAttribute("aria-busy"),
    "false",
  );
  t.mock.timers.enable({ apis: ["setTimeout"] });
  await act(async () => {
    t.mock.timers.tick(60_000);
    await settle();
  });
  await controller.refresh();
  await assert.rejects(
    controller.updateTimelineSession("a".repeat(64), { pinned: true }, 1),
  );
  assert.deepEqual(
    requests,
    ["/api/index?navigation=grouped"],
    "mismatch stops polling and writes without replay",
  );
});
