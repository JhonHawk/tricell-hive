import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { completeFixture } from "../dev/timeline-complete-fixture";
import { TimelineContent } from "../timeline-content";
import {
  TimelineTaskRegistry,
  workTaskRegistry,
} from "../timeline-task-registry";

const settle = async () => {
  for (let i = 0; i < 50; i++) await Promise.resolve();
};
function button(label: string, within: ParentNode = document) {
  const found = [...within.querySelectorAll<HTMLButtonElement>("button")].find(
    (node) =>
      node.getAttribute("aria-label") === label ||
      node.textContent?.trim() === label,
  );
  assert.ok(found, label);
  return found;
}
async function click(label: string, within: ParentNode = document) {
  await act(async () => {
    button(label, within).click();
    await settle();
  });
}

test("global task projection retains canonical objects and counts a multistage task only once", () => {
  const data = completeFixture(),
    registry = workTaskRegistry(data.session);
  assert.equal(registry.length, 7);
  const t7 = registry.find((row) => row.task.task_key === "T7");
  assert.ok(t7);
  assert.equal(
    t7.task,
    data.session.tasks.find((task) => task.task_key === "T7"),
  );
  assert.equal(t7.completed, 1);
  assert.equal(t7.total, 2);
});

test("toolbar opens the canonical task registry and a single origin marker never writes an API event", async (t) => {
  const data = completeFixture(),
    events = JSON.stringify(data.session.events);
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  await act(async () =>
    root.render(<TimelineContent data={data} onRetry={() => {}} />),
  );
  const toolbar = container.querySelector('[role="toolbar"]');
  assert.ok(toolbar);
  assert.deepEqual(
    [...toolbar.querySelectorAll("button")].map((button) =>
      button.textContent?.trim(),
    ),
    ["Tasks", "Go to current"],
  );
  assert.equal(container.textContent?.includes("2/7 completed"), false);
  const origins = container.querySelectorAll('[data-ui-origin="true"]');
  assert.equal(origins.length, 1);
  assert.ok(origins[0].textContent?.includes("Start"));
  assert.equal(
    origins[0].querySelector("time")?.dateTime,
    data.session.created_at,
  );
  assert.equal(
    origins[0].textContent?.includes(data.session.events[0].title),
    false,
  );
  assert.equal(
    container.querySelector('[aria-label="Session timeline"]')
      ?.firstElementChild,
    origins[0],
  );
  await click("Tasks", toolbar);
  const popover = document.querySelector('[data-slot="popover-content"]');
  assert.ok(popover);
  assert.ok(popover.textContent?.includes("2/7 completed"));
  assert.equal(popover.querySelectorAll("[data-registry-task-id]").length, 7);
  const t7 = popover.querySelector('[data-registry-task-id="task-7"]');
  assert.ok(t7);
  assert.equal(
    t7.querySelector("[data-task-segment-progress]")?.textContent,
    "1/2 segments completed",
  );
  assert.equal(
    popover.querySelectorAll('[data-registry-task-id="task-7"]').length,
    1,
  );
  await act(async () => {
    popover.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Escape", bubbles: true }),
    );
    await settle();
  });
  assert.equal(JSON.stringify(data.session.events), events);
});

test("empty task registry exposes refresh instead of a fabricated zero progress", async (t) => {
  const data = completeFixture();
  data.session.tasks = [];
  let retries = 0;
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  await act(async () =>
    root.render(
      <TimelineTaskRegistry
        session={data.session}
        lang="en"
        onRetry={() => {
          retries++;
        }}
      />,
    ),
  );
  await click("Tasks", container);
  const popover = document.querySelector('[data-slot="popover-content"]');
  assert.ok(popover);
  assert.ok(popover.textContent?.includes("No tasks published"));
  assert.equal(popover.textContent?.includes("0/0"), false);
  await click("Reload", popover);
  assert.equal(retries, 1);
});
