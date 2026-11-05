import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { completeFixture } from "../dev/timeline-complete-fixture";
import { TimelineContent } from "../timeline-content";
import { TIMELINE_HOST_TASK_KEY_PREFIX } from "../timeline-contract-v3";
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

test("registry rows of host tasks show only the title and the host badge", async (t) => {
  const data = completeFixture();
  const host = data.session.tasks[0];
  host.task_key = `${TIMELINE_HOST_TASK_KEY_PREFIX}native-1`;
  const plan = data.session.tasks[1];
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
        lang="es"
        onRetry={() => {}}
      />,
    ),
  );
  await click("Tareas", container);
  const popover = document.querySelector('[data-slot="popover-content"]');
  assert.ok(popover);
  const row = popover.querySelector(
    `[data-registry-task-id="${host.task_id}"]`,
  );
  assert.ok(row);
  assert.equal(row.querySelector("p")?.textContent, host.title);
  assert.equal(row.querySelector("p")?.getAttribute("title"), host.title);
  assert.equal(row.textContent?.includes("host:native"), false);
  assert.equal(
    row
      .querySelector("[aria-label]")
      ?.getAttribute("aria-label")
      ?.startsWith(`${host.title}: `),
    true,
  );
  const badge = row.querySelector('[data-slot="host-observed"]');
  // Compact marker: an icon with the badge's name, no visible text that would truncate titles.
  assert.ok(badge);
  assert.equal(badge.textContent, "");
  assert.equal(badge.querySelector("svg")?.getAttribute("aria-hidden"), "true");
  assert.equal(badge.getAttribute("role"), "img");
  assert.equal(
    badge.getAttribute("aria-label"),
    "Del host: Tarea tomada de la lista de tareas del host",
  );
  assert.equal(badge.hasAttribute("tabindex"), false);
  const planRow = popover.querySelector(
    `[data-registry-task-id="${plan.task_id}"]`,
  );
  assert.equal(
    planRow?.querySelector("p")?.textContent,
    `${plan.task_key} — ${plan.title}`,
  );
  assert.equal(planRow?.querySelector('[data-slot="host-observed"]'), null);
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

test("plan order sorts the current registry without changing stage segment order or source array", () => {
  const data = completeFixture();
  const sourceIDs = data.session.tasks.map((t) => t.task_id);
  const last = data.session.tasks.at(-1);
  assert.ok(last);
  last.plan_order = 1;
  data.session.tasks[0].plan_order = 80;
  const rows = workTaskRegistry(data.session);
  assert.equal(rows[0].task, last);
  assert.deepEqual(
    data.session.tasks.map((t) => t.task_id),
    sourceIDs,
  );
  data.session.tasks[0].title = "Updated in the only task registry";
  assert.equal(
    workTaskRegistry(data.session)[6].task.title,
    "Updated in the only task registry",
  );
  assert.equal(data.session.tasks[0].segments[0].order, 1);
});

test("registry loading and error states retain an actionable retry without phantom tasks", async (t) => {
  const data = completeFixture();
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
        loading
        error
      />,
    ),
  );
  await click("Tasks", container);
  const popover = document.querySelector('[data-slot="popover-content"]');
  assert.ok(popover);
  assert.ok(popover.querySelector('[role="status"]'));
  assert.equal(popover.querySelectorAll("[data-registry-task-id]").length, 0);
  await click("Retry", popover);
  assert.equal(retries, 1);
});
