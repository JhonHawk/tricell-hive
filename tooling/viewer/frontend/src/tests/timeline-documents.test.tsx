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
      !b.closest("[hidden]") &&
      (b.getAttribute("aria-label") === label ||
        b.textContent?.trim() === label),
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

test("common Sheet keeps document snapshots, compares the same chain and updates tabs manually", async (t) => {
  const { root, container } = mount();
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  let data = structuredClone(timelineFixture);
  const render = async () => {
    await act(async () =>
      root.render(
        <TimelineContent
          data={data}
          onRetry={() => {}}
          loadDocument={loadPreviewDocument}
        />,
      ),
    );
  };
  await render();
  await click("View plan");
  const sheet = document.querySelector<HTMLElement>(
    '[data-slot="sheet-content"]',
  );
  assert.ok(sheet);
  assert.ok(sheet.className.includes("sm:w-2/3"));
  assert.equal(sheet.querySelectorAll('[aria-label="Close"]').length, 1);
  assert.ok(sheet.textContent?.includes("Propuesta · R3"));
  assert.ok(sheet.textContent?.includes("Published"));
  assert.equal(sheet.querySelectorAll('[role="tab"]').length, 4);
  await click("Diseño");
  assert.ok(sheet.textContent?.includes("Native composition and sketches"));
  data = structuredClone(data);
  data.session.version++;
  data.session.title = "Unrelated title";
  await render();
  assert.equal(sheet.textContent?.includes("A newer revision"), false);
  const stage = data.session.events.find((e) => e.event_id === "event_plan_r3");
  assert.ok(stage);
  const previous = stage.document_snapshots?.at(-1);
  assert.ok(previous);
  stage.document_snapshots?.push({
    ...previous,
    document_revision_id: "revision-plan-four",
    predecessor_revision_id: previous.document_revision_id,
    tabs: previous.tabs.map((tab) =>
      tab.label === "Diseño"
        ? { ...tab, document_id: "document_earlier" }
        : tab,
    ),
  });
  await render();
  assert.ok(sheet.textContent?.includes("A newer revision"));
  assert.ok(sheet.textContent?.includes("Native composition and sketches"));
  await click("Show latest");
  assert.ok(sheet.textContent?.includes("Plan · revisión R2"));
  assert.ok(sheet.textContent?.includes("Changed: Diseño"));
  data = structuredClone(data);
  const current = data.session.events.find(
    (e) => e.event_id === "event_plan_r3",
  );
  assert.ok(current);
  const fourth = current.document_snapshots?.at(-1);
  assert.ok(fourth);
  current.document_snapshots?.push({
    ...fourth,
    document_revision_id: "revision-plan-five",
    predecessor_revision_id: fourth.document_revision_id,
    tabs: fourth.tabs.filter((tab) => tab.label !== "Diseño"),
  });
  await render();
  await click("Show latest");
  assert.ok(sheet.textContent?.includes("The previous tab no longer exists"));
  assert.ok(sheet.textContent?.includes("Removed: Diseño"));
  await click("Close");
});

test("documents from all five stages use the same Sheet and only structured plans group specifications", async (t) => {
  for (const kind of [
    "research",
    "plan",
    "build",
    "verify",
    "delivery",
  ] as const) {
    const { root, container } = mount();
    const data = structuredClone(timelineFixture);
    const stage = data.session.events.find(
      (e) => e.event_id === "event_plan_r3",
    );
    assert.ok(stage);
    stage.stage = kind;
    await act(async () =>
      root.render(
        <TimelineContent
          data={data}
          onRetry={() => {}}
          loadDocument={loadPreviewDocument}
        />,
      ),
    );
    const target = container.querySelector("#timeline-event-event_plan_r3");
    assert.ok(target);
    const label = {
      research: "View research",
      plan: "View plan",
      build: "View implementation",
      verify: "View verification",
      delivery: "View delivery",
    }[kind];
    const trigger = [
      ...target.querySelectorAll<HTMLButtonElement>("button"),
    ].find((b) => b.textContent?.trim() === label);
    assert.ok(trigger);
    await act(async () => {
      trigger.click();
      await settle();
    });
    const sheet = document.querySelector('[data-slot="sheet-content"]');
    assert.ok(sheet);
    assert.equal(
      sheet.textContent?.includes("Specifications"),
      kind === "plan",
    );
    assert.equal(sheet.querySelectorAll('[data-slot="sheet-close"]').length, 1);
    await click("Close");
    await act(async () => root.unmount());
    container.remove();
  }
  assert.ok(t);
});

test("document errors retry locally, empty content reloads and inactive reads cannot replace the selected tab", async (t) => {
  const { root, container } = mount();
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  let count = 0;
  let release: ((value: TimelineDocumentResponse) => void) | undefined;
  const loader = async (id: string) => {
    if (id !== "document_proposal") return loadPreviewDocument(id);
    count++;
    if (count === 1) throw new Error("unavailable");
    if (count === 2) return { ...(await loadPreviewDocument(id)), content: "" };
    return new Promise<TimelineDocumentResponse>((resolve) => {
      release = resolve;
    });
  };
  await act(async () =>
    root.render(
      <TimelineContent
        data={timelineFixture}
        onRetry={() => {}}
        loadDocument={loader}
      />,
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
  await click("Diseño");
  assert.ok(
    document
      .querySelector('[role="dialog"]')
      ?.textContent?.includes("Native composition and sketches"),
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

test("stage completion markers follow explicit states instead of feed indices", async () => {
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
      data.session.events.filter((e) => e.stage_state === "completed").length,
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

test("document selector isolates chains with identical labels and ignores a successor in another chain", async (t) => {
  const { root, container } = mount();
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  let data = structuredClone(timelineFixture);
  const stage = data.session.events.find((e) => e.event_id === "event_plan_r3");
  assert.ok(stage);
  const current = stage.document_snapshots?.at(-1);
  assert.ok(current);
  stage.document_snapshots?.push({
    ...structuredClone(current),
    document_key: "notes",
    document_revision_id: "notes-first",
    predecessor_revision_id: undefined,
    tabs: [{ label: "Diseño", document_id: "document_research" }],
  });
  const render = async () => {
    await act(async () =>
      root.render(
        <TimelineContent
          data={data}
          onRetry={() => {}}
          loadDocument={loadPreviewDocument}
        />,
      ),
    );
  };
  await render();
  await click("View plan");
  await click("Document");
  const option = [
    ...document.querySelectorAll<HTMLElement>('[role="option"]'),
  ].find((n) => n.textContent?.trim() === "notes");
  assert.ok(option);
  await act(async () => {
    option.click();
    await settle();
  });
  const sheet = document.querySelector('[data-slot="sheet-content"]');
  assert.ok(sheet);
  assert.ok(sheet.textContent?.includes("Initial revision"));
  assert.ok(sheet.textContent?.includes("Hallazgos"));
  assert.equal(sheet.querySelectorAll('[role="tab"]').length, 1);
  data = structuredClone(data);
  const live = data.session.events.find((e) => e.event_id === "event_plan_r3");
  assert.ok(live);
  live.document_snapshots?.push({
    ...structuredClone(current),
    document_revision_id: "plan-unrelated-successor",
    predecessor_revision_id: current.document_revision_id,
  });
  data.session.version++;
  await render();
  assert.equal(sheet.textContent?.includes("A newer revision"), false);
  assert.ok(sheet.textContent?.includes("Hallazgos"));
  await click("Close");
});
