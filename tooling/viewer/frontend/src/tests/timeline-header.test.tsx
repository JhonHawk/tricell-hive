import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { timelineFixture } from "../dev/timeline-fixture";
import { TimelineContent } from "../timeline-content";
import type { TimelineSessionResponse } from "../timeline-contract";
import { TimelineHeader } from "../timeline-header";
import { TimelineSummary } from "../timeline-summary";

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
  // A long token inside a spaced title wraps instead of being clipped, and the
  // clamped title is not offered as a horizontal reading region.
  assert.ok(h1?.className.includes("[overflow-wrap:anywhere]"));
  assert.equal(h1?.hasAttribute("tabindex"), false);
  assert.equal(h1?.hasAttribute("aria-describedby"), false);
  assert.equal(h1?.getAttribute("title"), "word ".repeat(40).trim());
  await spaced.unmount();
  const unbroken = await mount(header(withTitle("x".repeat(200)), "en"));
  const h1b = unbroken.container.querySelector("h1");
  assert.ok(h1b?.classList.contains("overflow-x-auto"));
  assert.equal(h1b?.getAttribute("tabindex"), "0");
  assert.equal(h1b?.classList.contains("line-clamp-2"), false);
  await unbroken.unmount();
});

test("a slow save does not take focus from where the user moved it", async () => {
  let resolve = () => {};
  const pending = new Promise<void>((r) => {
    resolve = r;
  });
  const elsewhere = document.createElement("button");
  document.body.append(elsewhere);
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  await act(async () =>
    root.render(
      <TimelineHeader
        data={timelineFixture}
        lang="en"
        onUpdate={() => pending}
        onDelete={async () => {}}
        onRetry={() => {}}
      />,
    ),
  );
  await act(async () =>
    container
      .querySelector<HTMLButtonElement>('button[aria-label="Edit title"]')
      ?.click(),
  );
  const input = container.querySelector<HTMLInputElement>("input");
  assert.ok(input);
  await act(async () => {
    input.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Enter", bubbles: true }),
    );
  });
  elsewhere.focus();
  await act(async () => {
    resolve();
    await pending;
  });
  assert.equal(document.activeElement, elsewhere);
  await act(async () => root.unmount());
  container.remove();
  elsewhere.remove();
});

async function withBoxHeights(
  scroll: number,
  client: number,
  run: () => Promise<void>,
) {
  const proto = HTMLElement.prototype;
  const saved = [
    Object.getOwnPropertyDescriptor(proto, "scrollHeight"),
    Object.getOwnPropertyDescriptor(proto, "clientHeight"),
  ];
  Object.defineProperty(proto, "scrollHeight", {
    configurable: true,
    get: () => scroll,
  });
  Object.defineProperty(proto, "clientHeight", {
    configurable: true,
    get: () => client,
  });
  try {
    await run();
  } finally {
    for (const [name, descriptor] of [
      ["scrollHeight", saved[0]],
      ["clientHeight", saved[1]],
    ] as const) {
      if (descriptor) Object.defineProperty(proto, name, descriptor);
      else delete (proto as unknown as Record<string, unknown>)[name];
    }
  }
}

test("show more appears only when the summary is actually clipped", async () => {
  const toggle = (c: HTMLElement) =>
    [...c.querySelectorAll("button")].find(
      (b) =>
        b.hasAttribute("aria-expanded") &&
        /more|less/i.test(b.textContent ?? ""),
    );
  await withBoxHeights(40, 40, async () => {
    const { container, unmount } = await mount(
      <TimelineSummary content={"word ".repeat(80)} lines={2} lang="en" />,
    );
    assert.equal(
      toggle(container),
      undefined,
      "fits: no toggle even for long text",
    );
    await unmount();
  });
  await withBoxHeights(120, 40, async () => {
    const { container, unmount } = await mount(
      <TimelineSummary content="short" lines={2} lang="en" />,
    );
    const button = toggle(container);
    assert.ok(button, "clipped: toggle shown");
    await act(async () => button.click());
    assert.equal(toggle(container)?.getAttribute("aria-expanded"), "true");
    await unmount();
  });
});
