import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { TimelineContent } from "../timeline-content";
import type {
  TimelineEvent,
  TimelineSessionResponse,
} from "../timeline-contract";
import { TimelineHeader } from "../timeline-header";

const sid = "a".repeat(64);
const candidate = {
  branch: "fix/example",
  commit: "abc1234",
  base: "development",
};
const stamp = "2026-10-31T18:00:00-06:00";
const ticket = { key: "COM-1", url: "https://example.com/tickets/COM-1" };
function fixture(): TimelineSessionResponse {
  const event: TimelineEvent = {
    actor: { kind: "native_producer", producer_id: "producer-original" },
    event_id: "event-first",
    sequence: 1,
    version: 1,
    kind: "validation",
    title: "Initial validation",
    summary: "Check the shared view.",
    created_at: stamp,
    updated_at: stamp,
    origin: "agent_reported",
    ticket_refs: [ticket, { key: "COM-2" }],
    evidence_id: "evidence-first",
    candidate,
    views: [
      {
        view_id: "view-chat",
        title: "Chat",
        ticket_refs: [ticket, { key: "COM-2" }],
        steps: ["Open Chat.", "Choose a filter."],
        expected_result: "The icon stays beside search.",
        image_refs: [
          { artifact_id: "image-one", caption: "390 px · light" },
          { artifact_id: "image-two", caption: "320 px · dark" },
        ],
      },
    ],
  };
  return {
    schema_version: 2,
    mode: "timeline_api",
    session: {
      schema_version: 2,
      producers: [
        {
          producer_id: "producer-original",
          source: {
            project_root: "/fixture",
            host: "codex",
            native_session_id: "native-example",
          },
          joined_at: stamp,
        },
      ],
      foreground_producer_id: "producer-original",
      pending_join_requests: [],
      tasks: [],
      waits: [],
      documents: [],
      current_stage_event_id: null,
      viewer_session_id: sid,
      project_key: "example/project",
      project_name: "Example",
      source: {
        project_root: "/fixture",
        host: "codex",
        native_session_id: "native-example",
      },
      version: 1,
      title: "Profile and Chat",
      summary: "Improve profile and filters.",
      state: "awaiting_validation",
      archived: false,
      pinned: false,
      created_at: stamp,
      updated_at: stamp,
      validation_count: 1,
      tickets: [ticket, { key: "COM-2" }],
      events: [event],
      artifacts: ["image-one", "image-two"].map((artifact_id) => ({
        artifact_id,
        actor: { kind: "native_producer", producer_id: "producer-original" },
        session_id: sid,
        sha256: "b".repeat(64),
        size: 100,
        media_type: "image/png",
        received_at: stamp,
        captured_at: stamp,
        candidate,
        viewport: "mobile",
        theme: "light",
      })),
    },
    artifact_urls: {
      "image-one": `/files/${sid}/${"b".repeat(64)}/one.png`,
      "image-two": `/files/${sid}/${"b".repeat(64)}/two.png`,
    },
  };
}
async function settle() {
  for (let i = 0; i < 20; i++) await Promise.resolve();
}
function button(label: string) {
  const found = [
    ...document.querySelectorAll<HTMLButtonElement>("button"),
  ].find(
    (b) =>
      b.getAttribute("aria-label") === label || b.textContent?.trim() === label,
  );
  assert.ok(found, `Missing button ${label}`);
  return found;
}
async function click(label: string) {
  await act(async () => {
    button(label).click();
    await settle();
  });
}

test("timeline is ascending, shared tickets do not duplicate a view, and steps stay behind Instructions", async () => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const data = fixture();
  data.session.events.push({
    ...data.session.events[0],
    actor: { kind: "native_producer", producer_id: "producer-original" },
    event_id: "event-later",
    evidence_id: "evidence-later",
    sequence: 2,
    kind: "iteration",
    title: "Compact filter",
    views: [],
  });
  await act(async () => {
    root.render(<TimelineContent data={data} onRetry={() => {}} />);
  });
  assert.deepEqual(
    [...container.querySelectorAll("[data-event-id]")].map((e) => e.id),
    ["timeline-event-event-first", "timeline-event-event-later"],
  );
  assert.equal(
    container.querySelectorAll('section[aria-label="Chat"]').length,
    1,
  );
  assert.equal(
    container.querySelectorAll('[data-slot="timeline-item"][data-completed]')
      .length,
    0,
  );
  for (const item of container.querySelectorAll("[data-event-id]")) {
    assert.equal(item.querySelectorAll("time").length, 1);
    assert.ok(item.querySelector('[data-slot="timeline-event-footer"] time'));
  }
  assert.equal(container.textContent?.includes("Open Chat."), false);
  assert.equal(
    container.querySelectorAll('a[href="https://example.com/tickets/COM-1"]')
      .length,
    2,
  );
  await act(async () => root.unmount());
  container.remove();
});

test("instructions keep the opened evidence snapshot across live session updates", async () => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const data = fixture();
  await act(async () =>
    root.render(<TimelineContent data={data} onRetry={() => {}} />),
  );
  await click("Instructions");
  assert.equal(
    (document.querySelector('[role="dialog"]') as HTMLElement)?.style.maxWidth,
    "40rem",
  );
  assert.ok(document.querySelector('[role="dialog"] [data-slot="separator"]'));
  assert.ok(
    document
      .querySelector('[role="dialog"]')
      ?.textContent?.includes("Open Chat."),
  );
  const newer = structuredClone(data);
  newer.session.version = 2;
  const changedView = newer.session.events[0].views?.[0];
  assert.ok(changedView);
  changedView.steps = ["New steps must not replace the open snapshot."];
  await act(async () => {
    root.render(<TimelineContent data={newer} onRetry={() => {}} />);
    await settle();
  });
  const dialog = document.querySelector('[role="dialog"]');
  assert.ok(dialog?.textContent?.includes("Open Chat."));
  assert.equal(
    dialog?.textContent?.includes("New steps must not replace"),
    false,
  );
  assert.ok(dialog?.textContent?.includes("An update is available"));
  await click("Close");
  await act(async () => root.unmount());
  container.remove();
});

test("capture carousel remains scoped to its view and ignores unsafe capability URLs", async () => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const data = fixture();
  await act(async () =>
    root.render(<TimelineContent data={data} onRetry={() => {}} />),
  );
  await click("Captures: Chat · 390 px · light");
  assert.ok(
    document
      .querySelector('[role="dialog"] img')
      ?.getAttribute("src")
      ?.endsWith("/one.png"),
  );
  await act(async () => {
    document
      .querySelector('[role="dialog"] img')
      ?.dispatchEvent(new Event("error"));
    await settle();
  });
  assert.equal(document.querySelector('[role="dialog"] img'), null);
  await click("Next capture");
  assert.ok(
    document
      .querySelector('[role="dialog"] img')
      ?.getAttribute("src")
      ?.endsWith("/two.png"),
  );
  await click("Close");
  await click("Captures: Chat · 390 px · light");
  assert.ok(
    document
      .querySelector('[role="dialog"] img')
      ?.getAttribute("src")
      ?.endsWith("/one.png"),
  );
  await click("Close");
  const unsafe = fixture();
  assert.ok(unsafe.artifact_urls);
  unsafe.artifact_urls["image-one"] =
    "https://example.com/not-a-local-capture.png";
  await act(async () =>
    root.render(<TimelineContent data={unsafe} onRetry={() => {}} />),
  );
  assert.equal(container.querySelector('img[src^="https:"]'), null);
  assert.ok(container.textContent?.includes("Capture unavailable"));
  await act(async () => root.unmount());
  container.remove();
});

test("header saves title against its opened version and keeps input on failure", async () => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const data = fixture();
  let expected = 0;
  let updates = 0;
  const update = async (_patch: { title?: string }, version: number) => {
    expected = version;
    updates++;
    throw new Error("conflict");
  };
  await act(async () =>
    root.render(
      <TimelineHeader
        data={data}
        onUpdate={update}
        onDelete={async () => {}}
        onRetry={() => {}}
      />,
    ),
  );
  await click("Edit title");
  const newer = structuredClone(data);
  newer.session.version = 4;
  await act(async () =>
    root.render(
      <TimelineHeader
        data={newer}
        onUpdate={update}
        onDelete={async () => {}}
        onRetry={() => {}}
      />,
    ),
  );
  await click("Save title");
  assert.equal(expected, 1);
  assert.equal(updates, 1);
  assert.ok(container.querySelector('input[aria-label="Session title"]'));
  assert.ok(container.textContent?.includes("Could not save"));
  await click("Cancel");
  await act(async () => root.unmount());
  container.remove();
});

test("empty, loading and error regions expose reload/retry instead of fabricated evidence", async () => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const data = fixture();
  data.session.events = [];
  let retries = 0;
  await act(async () =>
    root.render(
      <TimelineContent
        data={data}
        onRetry={() => {
          retries++;
        }}
      />,
    ),
  );
  assert.ok(container.textContent?.includes("No milestones yet"));
  await click("Reload");
  assert.equal(retries, 1);
  await act(async () =>
    root.render(
      <TimelineContent
        data={data}
        onRetry={() => {
          retries++;
        }}
        loading
        error
      />,
    ),
  );
  assert.ok(container.querySelector('[aria-label="Loading evidence"]'));
  await click("Retry");
  assert.equal(retries, 2);
  await act(async () => root.unmount());
  container.remove();
});

test("state audit and validation icons follow their declared state without an extra node ring", async (t) => {
  const { completeFixture } = await import("../dev/timeline-complete-fixture");
  const data = completeFixture();
  const stage = data.session.events.find((e) => e.stage === "delivery");
  assert.ok(stage);
  stage.stage_state = "awaiting_validation";
  data.session.state = "awaiting_validation";
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  await act(async () =>
    root.render(
      <TimelineContent
        data={data}
        defaultExpandedTaskIDs={["task-2"]}
        onRetry={() => {}}
      />,
    ),
  );
  const blocked = container.querySelector(
    '#timeline-event-task-block-audit [data-slot="timeline-indicator"]',
  );
  assert.ok(blocked);
  assert.match(
    blocked.querySelector("svg")?.getAttribute("class") ?? "",
    /circle-alert/,
  );
  assert.ok(blocked.classList.contains("border-0"));
  const validation = container.querySelector(
    '#timeline-event-delivery-stage [data-slot="timeline-indicator"]',
  );
  assert.ok(validation);
  assert.match(
    validation.querySelector("svg")?.getAttribute("class") ?? "",
    /circle-pause/,
  );
});
