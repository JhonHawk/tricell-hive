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

test("real application wires API timeline, documents, conditional title and mixed legacy navigation with one controller", async (t) => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  const data = structuredClone(timelineFixture);
  const sid = data.session.viewer_session_id;
  const legacyID = "1".repeat(64);
  let legacyRenders = 0;
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
          ...entry
        } = data.session;
        return new Response(
          JSON.stringify({
            projects: [
              {
                key: "legacy-project",
                name: "Legacy",
                tickets: [
                  {
                    ticket: "#1",
                    sessions: [
                      {
                        key: legacyID,
                        title: "Legacy session",
                        session_id: "legacy-native",
                        source: "/fixture",
                        source_status: "local",
                        content_hash: "hash",
                        created_at: "2026-10-31",
                        ingested_at: "2026-10-31",
                        fileBase: `/files/${legacyID}/${"c".repeat(64)}/`,
                      },
                    ],
                  },
                ],
              },
            ],
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
      if (path === `/api/sessions/${legacyID}`)
        return new Response(JSON.stringify({ title: "Legacy content" }));
      if (options?.method === "PATCH") {
        const body = JSON.parse(String(options.body));
        savedVersion = body.expected_version;
        data.session.title = body.title;
        data.session.version++;
        return new Response(
          JSON.stringify({
            schema_version: 1,
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
    root.render(
      <ViewerApplication
        controller={controller}
        loadModule={async () => ({
          validDeliveries: () => [],
          renderWalkthrough: (node) => {
            legacyRenders++;
            node.textContent = "Legacy renderer";
          },
        })}
      />,
    );
    await new Promise((resolve) => setTimeout(resolve, 15));
  });
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 15));
  });
  assert.ok(container.querySelector('[aria-label="Session timeline"]'));
  assert.equal(legacyRenders, 0);
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
  await click("View plan");
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
  await act(async () => {
    const project = container.querySelector<HTMLButtonElement>(
      'button[title="Legacy"]',
    );
    assert.ok(project);
    project.click();
    await settle();
  });
  await act(async () => {
    const row = container.querySelector<HTMLButtonElement>(
      `[data-session-key="${legacyID}"]`,
    );
    assert.ok(row);
    row.click();
    await settle();
  });
  assert.equal(controller.getSnapshot().selectedId, legacyID);
  assert.ok(container.textContent?.includes("Legacy renderer"));
  assert.ok(legacyRenders > 0);
  await act(async () => {
    container
      .querySelector<HTMLButtonElement>(`[data-session-key="${sid}"]`)
      ?.click();
    await settle();
  });
  assert.equal(controller.getSnapshot().selectedId, sid);
  assert.ok(container.querySelector('[aria-label="Session timeline"]'));
  assert.equal(
    paths.some((path) => path === `/api/sessions/${sid}`),
    false,
  );
});
