import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act, StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { WalkthroughDocument, type WalkthroughRenderer } from "../document";

const key = (n: number) => String(n).padStart(64, "0");
const session = (n = 1, extra = {}) => ({
  key: key(n),
  source: "/fixture",
  source_status: "local",
  content_hash: "hash",
  created_at: "2026-10-19",
  ingested_at: "2026-10-19",
  session_id: `fixture-${n}`,
  title: `Review ${n}`,
  fileBase: `/files/${key(n)}/${"a".repeat(64)}/`,
  ...extra,
});
const _index = (items = [session()], extra = {}) => ({
  projects: [
    {
      key: "project-1",
      name: "Hive",
      tickets: [{ ticket: "#89", sessions: items }],
    },
  ],
  excluded_sessions: [],
  repositories: [],
  errors: [],
  ttl_seconds: 86400,
  ...extra,
});
const _response = (data: unknown) => new Response(JSON.stringify(data));
const settle = async () => {
  for (let i = 0; i < 20; i++) await Promise.resolve();
};

test("compact navigation preserves full titles and selects sessions within collapsible projects", async () => {
  const { ViewerShell } = await import("../shell");
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const titles = [
    ["#89 · Review configuration", "#89", "Review configuration"],
    ["COM-761 — Preserve browser focus", "COM-761", "Preserve browser focus"],
    ["#90 · Different ticket remains", "#89", "#90 · Different ticket remains"],
    ["Local: Semantic title remains", "Local", "Local: Semantic title remains"],
    ["#89 without delimiter", "#89", "#89 without delimiter"],
    ["#89: ", "#89", "#89: "],
  ];
  const sessions = titles.map(([title, ticket], i) => ({
    id: key(i + 1),
    project: "Hive",
    projectKey: "hive",
    ticket,
    title,
    createdAt: "23 Oct",
    classification: "local" as const,
    pinned: i === 0,
  }));
  let selected = "";
  try {
    await act(async () => {
      root.render(
        <ViewerShell
          sessions={sessions}
          selectedId={key(1)}
          state="ready"
          onSelect={(id) => {
            selected = id;
          }}
          onRetry={() => {}}
        />,
      );
    });
    const projectToggle = container.querySelector<HTMLButtonElement>(
      '[data-slot="collapsible-trigger"]',
    );
    assert(projectToggle);
    await act(async () => projectToggle.click());
    const rows = [
      ...container.querySelectorAll<HTMLButtonElement>("[data-session-key]"),
    ];
    for (const [i, row] of rows.entries()) {
      assert(row.title.startsWith(`${titles[i][0]} — `));
      assert.equal(
        row.querySelector(".session-navigation-title")?.textContent,
        titles[i][2],
      );
      assert(row.textContent?.includes(titles[i][1]));
      assert(row.textContent?.includes("23 Oct"));
    }
    assert.equal(container.querySelectorAll('[aria-current="page"]').length, 1);
    assert.equal(container.querySelector("h1")?.textContent, titles[0][0]);
    assert(rows[0].querySelector(".lucide-pin"));
    assert(container.querySelector('[data-sidebar="header"] .lucide-route'));
    await act(async () => rows[1].click());
    assert.equal(selected, key(2));
    const project = container.querySelector<HTMLButtonElement>(
      '[data-slot="collapsible-trigger"]',
    );
    assert(project);
    assert(project.textContent?.includes("6"));
    assert.equal(project.getAttribute("aria-expanded"), "true");
    await act(async () => project.click());
    assert.equal(project.getAttribute("aria-expanded"), "false");
    await act(async () => project.click());
    assert.equal(project.getAttribute("aria-expanded"), "true");
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

test("renderer ref owns its classes, mounting, cleanup, focus and same-session scroll", async () => {
  const container = document.createElement("div");
  container.innerHTML = '<section id="document"></section>';
  document.body.append(container);
  const mount = container.querySelector("section");
  assert(mount);
  const root = createRoot(mount);
  let renders = 0;
  let closes = 0;
  let pauses = 0;
  const render: WalkthroughRenderer = (node, manifest, { fileBase }) => {
    renders++;
    node.classList.add("renderer-active");
    const button = document.createElement("button");
    button.textContent = manifest.title;
    button.dataset.base = fileBase;
    node.append(button);
    const video = document.createElement("video");
    video.pause = () => {
      pauses++;
    };
    video.load = () => {};
    video.src = "/test.mp4";
    node.append(video);
    const dialog = document.createElement("dialog");
    dialog.open = true;
    dialog.close = () => {
      closes++;
      dialog.open = false;
    };
    node.append(dialog);
  };
  const loader = () => Promise.resolve(render);
  const onError = () => assert.fail("Renderer failed");
  const data = {
    key: key(1),
    hash: "v1",
    fileBase: session().fileBase,
    manifest: { title: "First" },
  };
  try {
    await act(async () => {
      root.render(
        <StrictMode>
          <WalkthroughDocument
            data={data}
            loadRenderer={loader}
            onError={onError}
          />
        </StrictMode>,
      );
      await settle();
    });
    assert.equal(renders, 1);
    assert.equal(
      mount.querySelectorAll(".walkthrough.renderer-active").length,
      1,
    );
    assert.equal(mount.querySelector("button")?.dataset.base, data.fileBase);
    container.scrollTop = 173;
    mount.querySelector("button")?.focus();
    await act(async () => {
      root.render(
        <StrictMode>
          <WalkthroughDocument
            data={{ ...data, hash: "v2", manifest: { title: "Second" } }}
            loadRenderer={loader}
            onError={onError}
          />
        </StrictMode>,
      );
      await settle();
    });
    assert.equal(mount.querySelectorAll("button").length, 1);
    assert.equal(container.scrollTop, 173);
    assert.equal(document.activeElement, mount.querySelector("button"));
    assert.equal(document.activeElement?.textContent, "Second");
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
  assert(pauses >= 1);
  assert(closes >= 1);
});

test("shared renderer keeps capture URLs, carousel close and print hooks across repeated mounts", async () => {
  const modulePath = new URL("../../../static/walkthrough.mjs", import.meta.url)
    .href;
  const { renderWalkthrough } = await import(modulePath);
  const originalShow = HTMLDialogElement.prototype.showModal;
  const originalClose = HTMLDialogElement.prototype.close;
  HTMLDialogElement.prototype.showModal = function () {
    this.open = true;
  };
  HTMLDialogElement.prototype.close = function () {
    this.open = false;
    this.dispatchEvent(new Event("close"));
  };
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const loadRenderer = () =>
    Promise.resolve(renderWalkthrough as WalkthroughRenderer);
  const base = session().fileBase;
  const manifest = {
    title: "Captures and print",
    kind: "walkthrough",
    lang: "en",
    created_at: "2026-10-19",
    candidate: { branch: "fixture", commit: "fixture" },
    session: { id: "fixture", host: "test" },
    flows: [
      {
        title: "Open capture",
        steps: [],
        captures: [
          { path: "images/view.png", caption: "View", viewport: "desktop" },
        ],
        recording: { status: "none" },
      },
    ],
    checks: [{ name: "Review", status: "pass", result: "Pass" }],
  };
  const onError = () => assert.fail("Shared renderer failed");
  try {
    for (let i = 0; i < 2; i++) {
      await act(async () => {
        root.render(
          <StrictMode>
            <WalkthroughDocument
              data={{ key: key(1), hash: String(i), fileBase: base, manifest }}
              loadRenderer={loadRenderer}
              onError={onError}
            />
          </StrictMode>,
        );
        await settle();
      });
      const thumb = container.querySelector<HTMLButtonElement>(".wt-thumb");
      assert(thumb);
      thumb.click();
      const dialog = container.querySelector<HTMLDialogElement>(".wt-carousel");
      assert(dialog?.open);
      assert.equal(
        dialog.querySelector("img")?.getAttribute("src"),
        `${base}images/view.png`,
      );
      dialog.querySelector<HTMLButtonElement>(".wt-close")?.click();
      assert.equal(dialog.open, false);
      assert.equal(document.activeElement, thumb);
      const details =
        container.querySelector<HTMLDetailsElement>("details.wt-checks");
      assert(details);
      details.open = false;
      window.dispatchEvent(new Event("beforeprint"));
      assert.equal(details.open, true);
      window.dispatchEvent(new Event("afterprint"));
      assert.equal(details.open, false);
      assert.equal(container.querySelectorAll(".wt-carousel").length, 1);
    }
  } finally {
    await act(async () => root.unmount());
    container.remove();
    HTMLDialogElement.prototype.showModal = originalShow;
    HTMLDialogElement.prototype.close = originalClose;
  }
});

test("renderer and delivery helper share module loading and retry rejected attempts", async () => {
  const { loadWalkthroughRenderer, loadWalkthroughModule } = await import(
    "../document"
  );
  const firstModule = loadWalkthroughModule();
  assert.equal(loadWalkthroughModule(), firstModule);
  const failed = loadWalkthroughRenderer();
  assert.equal(loadWalkthroughRenderer(), failed);
  await Promise.all([assert.rejects(firstModule), assert.rejects(failed)]);
  const nextModule = loadWalkthroughModule();
  assert.notEqual(nextModule, firstModule);
  assert.equal(loadWalkthroughModule(), nextModule);
  const retry = loadWalkthroughRenderer();
  assert.notEqual(retry, failed);
  await Promise.all([assert.rejects(nextModule), assert.rejects(retry)]);
});

test("project defaults follow lifecycle while manual project and sidebar choices survive remount", async () => {
  const { ViewerShell } = await import("../shell");
  const previous = sessionStorage.getItem("hive-view-navigation");
  sessionStorage.removeItem("hive-view-navigation");
  const container = document.createElement("div");
  document.body.append(container);
  let root = createRoot(container);
  const sessions = [
    {
      id: key(1),
      project: "Running",
      projectKey: "running",
      ticket: "#1",
      title: "Active work",
      createdAt: "23 Oct",
      classification: "local" as const,
      status: "in_progress" as const,
    },
    {
      id: key(2),
      project: "Closed",
      projectKey: "closed",
      ticket: "#2",
      title: "Completed work",
      createdAt: "22 Oct",
      classification: "local" as const,
      status: "closed" as const,
    },
    {
      id: key(3),
      project: "Unknown",
      projectKey: "unknown",
      ticket: "#3",
      title: "Historical work",
      createdAt: "21 Oct",
      classification: "local" as const,
    },
  ];
  const render = async (items = sessions) => {
    await act(async () =>
      root.render(
        <ViewerShell sessions={items} state="ready" onRetry={() => {}} />,
      ),
    );
  };
  const group = (name: string) => {
    const button = container.querySelector<HTMLButtonElement>(
      `[data-slot="collapsible-trigger"][title="${name}"]`,
    );
    assert(button);
    return button;
  };
  try {
    await render();
    assert.equal(group("Running").getAttribute("aria-expanded"), "true");
    assert.equal(group("Closed").getAttribute("aria-expanded"), "false");
    assert.equal(group("Unknown").getAttribute("aria-expanded"), "false");
    await act(async () => {
      group("Running").click();
      group("Closed").click();
    });
    const toggle = container.querySelector<HTMLButtonElement>(
      '[data-slot="sidebar-inset"] [data-sidebar="trigger"]',
    );
    assert(toggle);
    await act(async () => toggle.click());
    await render(
      sessions.map((session) => ({
        ...session,
        title: `${session.title} refreshed`,
      })),
    );
    assert.equal(group("Running").getAttribute("aria-expanded"), "false");
    assert.equal(group("Closed").getAttribute("aria-expanded"), "true");
    assert.equal(
      container
        .querySelector('[data-slot="sidebar"]')
        ?.getAttribute("data-state"),
      "collapsed",
    );
    await act(async () => root.unmount());
    root = createRoot(container);
    await render();
    assert.equal(group("Running").getAttribute("aria-expanded"), "false");
    assert.equal(group("Closed").getAttribute("aria-expanded"), "true");
    assert.equal(
      container
        .querySelector('[data-slot="sidebar"]')
        ?.getAttribute("data-state"),
      "collapsed",
    );
    assert.deepEqual(
      JSON.parse(sessionStorage.getItem("hive-view-navigation") || "{}")
        .projects,
      { running: false, closed: true },
    );
  } finally {
    await act(async () => root.unmount());
    container.remove();
    if (previous === null) sessionStorage.removeItem("hive-view-navigation");
    else sessionStorage.setItem("hive-view-navigation", previous);
  }
});

test("archive filter is icon-only with an accessible toggle and grouped actions keep their names", async () => {
  const { ViewerShell } = await import("../shell");
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  let shown = false;
  const item = {
    id: key(1),
    project: "Hive",
    ticket: "#89",
    title: "Archive action",
    createdAt: "23 Oct",
    classification: "local" as const,
    status: "closed" as const,
  };
  try {
    await act(async () =>
      root.render(
        <ViewerShell
          sessions={[item]}
          selectedId={item.id}
          state="ready"
          onRetry={() => {}}
          onShowArchivedChange={(value) => {
            shown = value;
          }}
          onArchive={() => {}}
          onPin={() => {}}
          onDelete={() => {}}
        />,
      ),
    );
    const filter = container.querySelector<HTMLButtonElement>(
      'button[aria-label="Show archived"]',
    );
    assert(filter);
    assert.equal(filter.textContent, "");
    assert.equal(filter.getAttribute("aria-pressed"), "false");
    await act(async () => filter.click());
    assert.equal(shown, true);
    const actions = container.querySelector(
      '[role="group"][aria-label="Session actions"]',
    );
    assert(actions);
    assert.deepEqual(
      [...actions.querySelectorAll("button")].map((button) =>
        button.getAttribute("aria-label"),
      ),
      ["Archive session", "Keep session", "Delete session"],
    );
    assert(!actions.textContent?.includes("Closed"));
    assert.equal(actions.querySelector('[aria-haspopup="menu"]'), null);
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});
