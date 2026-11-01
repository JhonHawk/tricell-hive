import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test, { type TestContext } from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { HttpError } from "../api";
import { workFixture } from "../dev/timeline-work-fixture";
import { TimelineContent } from "../timeline-content";
import { TimelineHeader } from "../timeline-header";
import type { TimelineDetailTarget } from "../timeline-tasks";

const settle = async () => {
  for (let i = 0; i < 50; i++) await Promise.resolve();
};
function region(selector: string, within: ParentNode = document) {
  const node = within.querySelector(selector);
  assert.ok(node);
  return node;
}
function button(label: string, within: ParentNode = document) {
  const b = [...within.querySelectorAll<HTMLButtonElement>("button")].find(
    (b) =>
      b.getAttribute("aria-label") === label || b.textContent?.trim() === label,
  );
  assert.ok(b, `Missing ${label}`);
  return b;
}
async function click(label: string, within: ParentNode = document) {
  await act(async () => {
    button(label, within).click();
    await settle();
  });
}
function mount(t: TestContext) {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  return { root, container };
}
test("CAM stage counts use bound segments and work total deduplicates T7", async (t) => {
  const { root, container } = mount(t);
  const data = workFixture();
  let target: TimelineDetailTarget | undefined;
  await act(async () =>
    root.render(
      <TimelineContent
        data={data}
        lang="es"
        onRetry={() => {}}
        onOpenDetail={(value) => {
          target = value;
        }}
      />,
    ),
  );
  assert.ok(container.textContent?.includes("6/7 finalizadas"));
  for (const [id, count] of [
    ["build-stage", "6/6"],
    ["verify-stage", "1/1"],
    ["delivery-stage", "0/1"],
  ])
    assert.ok(
      container
        .querySelector(`#timeline-event-${id}`)
        ?.textContent?.includes(`${count} finalizadas`),
    );
  assert.equal(
    container.querySelectorAll(
      '#timeline-event-build-stage [data-task-id="task-7"]',
    ).length,
    0,
  );
  await click(
    "Expandir T7",
    region("#timeline-event-delivery-stage", container),
  );
  assert.ok(container.textContent?.includes("CI — Checks del pull request"));
  await click(
    "T7 — Verificación y entrega del candidato: No iniciada",
    region("#timeline-event-delivery-stage", container),
  );
  assert.deepEqual(target, {
    kind: "task",
    taskID: "task-7",
    segmentID: "delivery-7",
  });
});
test("independent indicators, disclosures and focus survive polling updates for thirty realistic tasks", async (t) => {
  const { root, container } = mount(t);
  const data = workFixture("volume");
  await act(async () =>
    root.render(<TimelineContent data={data} lang="es" onRetry={() => {}} />),
  );
  const rows = [...container.querySelectorAll("[data-task-id]")];
  assert.equal(rows.length, 30);
  assert.ok(rows[0].hasAttribute("data-completed"));
  assert.ok(!rows[1].hasAttribute("data-completed"));
  assert.ok(rows[2].hasAttribute("data-completed"));
  assert.ok(rows[1].querySelector('[aria-label="T2: Bloqueada"]'));
  assert.ok(rows[4].querySelector('[aria-label="T5: No iniciada"]'));
  await click("Expandir T2");
  await click("Expandir T4");
  const trigger = button("Contraer T4");
  trigger.focus();
  const next = structuredClone(data);
  next.session.version++;
  next.session.tasks[1].task_version++;
  next.session.tasks[1].segments[0].state = "in_progress";
  delete next.session.tasks[1].segments[0].block;
  await act(async () =>
    root.render(<TimelineContent data={next} lang="es" onRetry={() => {}} />),
  );
  assert.ok(button("Contraer T2"));
  assert.equal(document.activeElement, trigger);
  assert.ok(button("Contraer T4"));
  assert.equal(
    container.querySelector('[data-task-id="task-2"] [data-slot="badge"]'),
    null,
  );
  assert.ok(
    container
      .querySelector('[data-slot="spinner"]')
      ?.classList.contains("motion-reduce:animate-none"),
  );
});
test("pending future segments never fabricate a delivery stage", async (t) => {
  const { root, container } = mount(t);
  await act(async () =>
    root.render(
      <TimelineContent data={workFixture("future")} onRetry={() => {}} />,
    ),
  );
  assert.equal(container.querySelector("#timeline-event-delivery-stage"), null);
  assert.equal(
    container.querySelector(
      '#timeline-event-build-stage [data-task-id="task-7"]',
    ),
    null,
  );
});
test("lead recovery requires explicit reason, preserves input after conflict and never auto-submits", async (t) => {
  const { root, container } = mount(t);
  let calls = 0;
  await act(async () =>
    root.render(
      <TimelineHeader
        data={workFixture()}
        onRetry={async () => {}}
        onUpdate={async () => {}}
        onDelete={async () => {}}
        onRecover={async () => {
          calls++;
          throw new HttpError(409, { code: "version_conflict" });
        }}
      />,
    ),
  );
  await click("Sessions: 1");
  await click("Reassign");
  assert.equal(calls, 0);
  const input = document.querySelector<HTMLInputElement>(
    '[role="dialog"] input',
  );
  assert.ok(input);
  assert.ok(
    button("Reassign", region('[data-slot="dialog-content"]')).disabled,
  );
  await act(async () => {
    Object.getOwnPropertyDescriptor(
      window.HTMLInputElement.prototype,
      "value",
    )?.set?.call(input, "Continue authorized work");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  await click("Reassign", region('[role="dialog"]'));
  assert.equal(calls, 1);
  assert.equal(input.value, "Continue authorized work");
  assert.ok(
    document
      .querySelector('[role="dialog"]')
      ?.textContent?.includes("The lead changed"),
  );
  assert.equal(calls, 1);
  assert.ok(container);
});

test("failed opening reread blocks lead recovery until a successful manual reload", async (t) => {
  const { root } = mount(t);
  let reads = 0,
    calls = 0;
  let readable = false;
  await act(async () =>
    root.render(
      <TimelineHeader
        data={workFixture()}
        onRetry={async () => {
          reads++;
          if (!readable) throw new Error("read_unavailable");
        }}
        onUpdate={async () => {}}
        onDelete={async () => {}}
        onRecover={async () => {
          calls++;
        }}
      />,
    ),
  );
  await click("Sessions: 1");
  await click("Reassign");
  assert.equal(reads, 1);
  const dialog = region('[data-slot="dialog-content"]');
  const input = dialog.querySelector<HTMLInputElement>("input");
  assert.ok(input);
  await act(async () => {
    Object.getOwnPropertyDescriptor(
      window.HTMLInputElement.prototype,
      "value",
    )?.set?.call(input, "Resume the authorized work");
    input.dispatchEvent(new Event("input", { bubbles: true }));
  });
  assert.ok(button("Reassign", dialog).disabled);
  await click("Reassign", dialog);
  assert.equal(calls, 0);
  readable = true;
  await click("Reload", dialog);
  assert.equal(reads, 2);
  assert.equal(input.value, "Resume the authorized work");
  assert.equal(calls, 0);
  assert.equal(button("Reassign", dialog).disabled, false);
  await click("Reassign", dialog);
  assert.equal(calls, 1);
});
