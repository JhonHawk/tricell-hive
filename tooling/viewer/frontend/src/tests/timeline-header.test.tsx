import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { timelineFixture } from "../dev/timeline-fixture";
import { TimelineContent } from "../timeline-content";
import type { TimelineSessionResponse } from "../timeline-contract";
import { TimelineHeader } from "../timeline-header";

const summary = timelineFixture.session.summary;

async function mount(node: React.ReactNode) {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  await act(async () => root.render(node));
  return {
    container,
    unmount: async () => {
      await act(async () => root.unmount());
      container.remove();
    },
  };
}

function header(data: TimelineSessionResponse, lang: "es" | "en") {
  return (
    <TimelineHeader
      data={data}
      lang={lang}
      onUpdate={async () => {}}
      onDelete={async () => {}}
      onRetry={() => {}}
    />
  );
}

test("header shows neither the project line nor the date", async () => {
  const { container, unmount } = await mount(header(timelineFixture, "es"));
  assert.equal(
    container.textContent?.includes(timelineFixture.session.project_name),
    false,
  );
  assert.equal(container.querySelector("time"), null);
  assert.equal(
    container.querySelector("h1")?.textContent,
    timelineFixture.session.title,
  );
  await unmount();
});

test("the summary lives in the header, clamped to two lines, and not in the canvas", async () => {
  const head = await mount(header(timelineFixture, "es"));
  const slot = head.container.querySelector('[data-slot="timeline-header"]');
  assert.ok(slot?.textContent?.includes(summary));
  assert.ok(slot?.querySelector(".line-clamp-2"));
  await head.unmount();
  const canvas = await mount(
    <TimelineContent data={timelineFixture} onRetry={() => {}} />,
  );
  assert.equal(canvas.container.textContent?.includes(summary), false);
  await canvas.unmount();
});

test("the validation count is a toolbar counter that reads its name and is not a control", async () => {
  for (const [lang, name] of [
    ["es", "Validaciones: 2"],
    ["en", "Validations: 2"],
  ] as const) {
    const { container, unmount } = await mount(header(timelineFixture, lang));
    const counter = [
      ...container.querySelectorAll<HTMLElement>(
        '[data-slot="button-group-text"]',
      ),
    ].find((node) => node.textContent === name);
    assert.ok(counter, name);
    assert.equal(counter.tagName === "BUTTON", false);
    assert.equal(counter.closest("button"), null);
    await unmount();
  }
});

function withTitle(title: string): TimelineSessionResponse {
  return { ...timelineFixture, session: { ...timelineFixture.session, title } };
}

test("leaving title edit returns focus to the edit button", async () => {
  for (const key of ["Escape", "Enter"]) {
    const { container, unmount } = await mount(header(timelineFixture, "en"));
    const edit = () =>
      container.querySelector<HTMLButtonElement>(
        'button[aria-label="Edit title"]',
      );
    await act(async () => edit()?.click());
    const input = container.querySelector<HTMLInputElement>("input");
    assert.ok(input, "edit mode");
    await act(async () => {
      input.dispatchEvent(new KeyboardEvent("keydown", { key, bubbles: true }));
    });
    assert.equal(document.activeElement, edit(), key);
    await unmount();
  }
});

test("a long title with spaces is clamped to two lines; an unbroken one scrolls", async () => {
  const spaced = await mount(
    header(withTitle("word ".repeat(40).trim()), "en"),
  );
  const h1 = spaced.container.querySelector("h1");
  assert.ok(h1?.classList.contains("line-clamp-2"));
  assert.equal(h1?.getAttribute("title"), "word ".repeat(40).trim());
  await spaced.unmount();
  const unbroken = await mount(header(withTitle("x".repeat(200)), "en"));
  const h1b = unbroken.container.querySelector("h1");
  assert.ok(h1b?.classList.contains("overflow-x-auto"));
  assert.equal(h1b?.classList.contains("line-clamp-2"), false);
  await unbroken.unmount();
});
