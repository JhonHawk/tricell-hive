import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { loadPreviewDocument } from "../dev/timeline-documents-fixture";
import { timelineFixture } from "../dev/timeline-fixture";
import { TimelineContent } from "../timeline-content";
import type { TimelineDocumentResponse } from "../timeline-contract";
import { TimelineMarkdown } from "../timeline-markdown";

async function settle() {
  for (let i = 0; i < 25; i++) await Promise.resolve();
}
function button(label: string) {
  const value = [
    ...document.querySelectorAll<HTMLButtonElement>("button"),
  ].find(
    (b) =>
      b.getAttribute("aria-label") === label || b.textContent?.trim() === label,
  );
  assert.ok(value, `Missing button ${label}`);
  return value;
}
async function click(label: string) {
  await act(async () => {
    button(label).click();
    await settle();
  });
}
function mount() {
  const container = document.createElement("div");
  document.body.append(container);
  return { container, root: createRoot(container) };
}

test("stage documents default to latest snapshot; native tabs, revision selection and refresh preserve context", async () => {
  const { root, container } = mount();
  const data = structuredClone(timelineFixture);
  await act(async () =>
    root.render(
      <TimelineContent
        data={data}
        onRetry={() => {}}
        loadDocument={loadPreviewDocument}
      />,
    ),
  );
  const trigger = button("View plan"); // First plan button belongs to the newest occurrence.
  await act(async () => {
    trigger.focus();
    trigger.click();
    await settle();
  });
  let dialog = document.querySelector<HTMLElement>('[role="dialog"]');
  assert.ok(dialog);
  assert.ok(dialog.textContent?.includes("Propuesta · R3"));
  assert.equal(dialog.querySelectorAll('[role="tab"]').length, 4);
  assert.equal(dialog.style.maxWidth, "40rem");
  await click("Diseño");
  assert.ok(dialog.textContent?.includes("Native composition and sketches"));
  await act(async () => {
    button("Diseño").focus();
    button("Diseño").dispatchEvent(
      new KeyboardEvent("keydown", { key: "ArrowRight", bubbles: true }),
    );
    await settle();
  });
  assert.equal(document.activeElement?.textContent, "Tareas");
  const newer = structuredClone(data);
  newer.session.version++;
  const stage = newer.session.events.find(
    (e) => e.event_id === "event_plan_r3",
  );
  assert.ok(stage);
  stage.title = "Replacement title";
  stage.document_snapshots = [];
  await act(async () =>
    root.render(
      <TimelineContent
        data={newer}
        onRetry={() => {}}
        loadDocument={loadPreviewDocument}
      />,
    ),
  );
  dialog = document.querySelector<HTMLElement>('[role="dialog"]');
  assert.ok(dialog);
  assert.ok(dialog.textContent?.includes("An update is available"));
  assert.ok(dialog.textContent?.includes("Plan · revisión R3"));
  assert.equal(dialog.querySelectorAll('[role="tab"]').length, 4);
  await click("Revision");
  const option = [
    ...document.querySelectorAll<HTMLElement>('[role="option"]'),
  ].at(-1);
  assert.ok(option);
  await act(async () => {
    option.click();
    await settle();
  });
  assert.equal(dialog.querySelectorAll('[role="tab"]').length, 1);
  assert.ok(dialog.textContent?.includes("borrador inicial"));
  await click("Close");
  await act(async () => {
    await new Promise((resolve) => setTimeout(resolve, 10));
  });
  assert.equal(document.querySelector('[role="dialog"]'), null);
  // The native close restores focus when the trigger remains available.
  await act(async () => root.unmount());
  container.remove();
});

test("document error retries, empty reloads, and a delayed inactive tab cannot replace the selected document", async () => {
  const { root, container } = mount();
  const data = structuredClone(timelineFixture);
  let calls = 0;
  let release: ((response: TimelineDocumentResponse) => void) | undefined;
  const loader = async (id: string) => {
    if (id === "document_proposal") {
      calls++;
      if (calls === 1) throw new Error("unavailable");
      if (calls === 2)
        return { ...(await loadPreviewDocument(id)), content: "" };
      return new Promise<TimelineDocumentResponse>((resolve) => {
        release = resolve;
      });
    }
    return loadPreviewDocument(id);
  };
  await act(async () =>
    root.render(
      <TimelineContent data={data} onRetry={() => {}} loadDocument={loader} />,
    ),
  );
  await click("View plan");
  assert.ok(
    document
      .querySelector('[role="dialog"]')
      ?.textContent?.includes("Could not load document"),
  );
  await click("Retry");
  assert.ok(
    document
      .querySelector('[role="dialog"]')
      ?.textContent?.includes("The document is empty"),
  );
  await click("Reload");
  assert.ok(document.querySelector('[aria-label="Loading document"]'));
  await click("Criterios");
  assert.ok(
    document
      .querySelector('[role="dialog"]')
      ?.textContent?.includes("Criterios · R3"),
  );
  assert.ok(release);
  await act(async () => {
    release?.({
      ...(await loadPreviewDocument("document_proposal")),
      content: "STALE CONTENT",
    });
    await settle();
  });
  assert.equal(
    document
      .querySelector('[role="tabpanel"]')
      ?.textContent?.includes("STALE CONTENT"),
    false,
  );
  await click("Close");
  await act(async () => root.unmount());
  container.remove();
});

test("Markdown preserves document semantics while HTML, images and unsafe destinations stay inert", async () => {
  const { root, container } = mount();
  await act(async () =>
    root.render(
      <TimelineMarkdown
        content={
          '# Heading\n\n1. First\n2. Second\n\n| A | B |\n| - | - |\n| one | two |\n\n```html\n<script>alert(1)</script>\n```\n\n[Safe](https://example.com/docs) [Unsafe](javascript:alert%281%29) [HTTP](http://example.com)\n\n![Remote](https://example.com/tracker.png)\n\n<script>window.compromised=true</script><iframe src="https://example.com"></iframe><img src="https://example.com/other.png" />'
        }
      />,
    ),
  );
  assert.equal(container.querySelector("h1")?.textContent, "Heading");
  assert.equal(container.querySelectorAll("ol li").length, 2);
  assert.equal(container.querySelectorAll("table tbody tr").length, 1);
  assert.ok(
    container.querySelector("pre code")?.textContent?.includes("<script>"),
  );
  assert.equal(container.querySelectorAll("a").length, 1);
  assert.equal(
    container.querySelector("a")?.getAttribute("rel"),
    "noopener noreferrer",
  );
  assert.equal(container.querySelectorAll("img,iframe,script").length, 0);
  assert.ok(container.textContent?.includes("Unsafe"));
  await act(async () => root.unmount());
  container.remove();
});

test("stage indicators use declared paused/failed/completed states without marking the feed steps completed", async () => {
  const { root, container } = mount();
  const data = structuredClone(timelineFixture);
  const stage = data.session.events.find((e) => e.event_id === "event_plan_r3");
  assert.ok(stage);
  for (const state of ["paused", "failed", "completed"] as const) {
    stage.stage_state = state;
    data.session.current_stage_event_id = null;
    await act(async () =>
      root.render(<TimelineContent data={data} onRetry={() => {}} />),
    );
    const node = container.querySelector("#timeline-event-event_plan_r3");
    assert.ok(node);
    assert.ok(
      node
        .querySelector('[data-slot="timeline-event-footer"]')
        ?.textContent?.includes(
          state === "paused"
            ? "Paused"
            : state === "failed"
              ? "Failed"
              : "Completed",
        ),
    );
    assert.equal(
      container.querySelectorAll('[data-slot="timeline-item"][data-completed]')
        .length,
      0,
    );
  }
  await act(async () => root.unmount());
  container.remove();
});

test("native document Dialog closes with Escape and restores its trigger focus", async () => {
  const { root, container } = mount();
  await act(async () =>
    root.render(
      <TimelineContent
        data={timelineFixture}
        onRetry={() => {}}
        loadDocument={loadPreviewDocument}
      />,
    ),
  );
  const trigger = button("View plan");
  await act(async () => {
    trigger.focus();
    trigger.click();
    await settle();
  });
  assert.ok(document.querySelector('[role="dialog"]'));
  await act(async () => {
    document.activeElement?.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Escape", bubbles: true }),
    );
    await new Promise((resolve) => setTimeout(resolve, 10));
  });
  assert.equal(document.querySelector('[role="dialog"]'), null);
  assert.equal(document.activeElement, trigger);
  await act(async () => root.unmount());
  container.remove();
});
