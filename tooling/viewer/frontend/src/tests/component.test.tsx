import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test from "node:test";
import { act, StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { AccessClient } from "../api";
import { ViewerApplication } from "../application";
import { ViewerController } from "../controller";
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
const index = (items = [session()], extra = {}) => ({
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
const response = (data: unknown) => new Response(JSON.stringify(data));
const settle = async () => {
  for (let i = 0; i < 20; i++) await Promise.resolve();
};

test("StrictMode mounts exchange once, renders grouped explicit statuses and updates by polling", async () => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  let exchanges = 0;
  let data = index(
    [
      session(1, {
        session_status: "closed",
        source_status: "unknown",
        kind: "first_view",
      }),
      session(2, { session_status: "in_progress" }),
      session(3, { kind: "final" }),
    ],
    { excluded_sessions: [session(4, { source_status: "tracked" })] },
  );
  data.projects.push({
    key: "project-2",
    name: "Hive",
    tickets: [
      {
        ticket: "#89",
        sessions: [session(5, { session_status: "closed" })],
      },
    ],
  });
  const client = new AccessClient(
    async (url) => {
      if (String(url).includes("exchange")) {
        exchanges++;
        return response({ token: "test-access" });
      }
      return response(data);
    },
    {
      getItem() {
        throw Error("unavailable");
      },
      setItem() {
        throw Error("unavailable");
      },
      removeItem() {
        throw Error("unavailable");
      },
    },
  );
  const controller = new ViewerController(client);
  const opening = controller.open("one-use");
  try {
    await act(async () => {
      root.render(
        <StrictMode>
          <ViewerApplication controller={controller} />
        </StrictMode>,
      );
      await opening;
      await settle();
    });
    assert.equal(exchanges, 1);
    const closedProject = container.querySelectorAll<HTMLButtonElement>(
      '[data-slot="collapsible-trigger"]',
    )[1];
    assert(closedProject);
    await act(async () => closedProject.click());
    const rows = [
      ...container.querySelectorAll<HTMLElement>("[data-session-key]"),
    ];
    assert.equal(rows.length, 5);
    assert.match(
      rows.find((row) => row.dataset.sessionKey === key(1))?.textContent ?? "",
      /Closed/,
    );
    assert.match(
      rows.find((row) => row.dataset.sessionKey === key(3))?.textContent ?? "",
      /Not reported/,
    );
    assert.equal(container.querySelectorAll(".navigation-branch").length, 0);
    assert.equal(
      container.querySelectorAll('[data-slot="collapsible"]').length,
      2,
    );
    assert.equal(
      rows[0]
        .querySelector("[data-session-status]")
        ?.getAttribute("data-session-status"),
      "closed",
    );
    data = index([
      session(1, {
        session_status: "in_progress",
        kind: "final",
        source_status: "unknown",
      }),
    ]);
    await act(async () => {
      await controller.refresh();
      await settle();
    });
    assert.match(
      container.querySelector("[data-session-key]")?.textContent ?? "",
      /In progress/,
    );
    assert.equal(
      controller.getSnapshot().sessions[0].classification,
      "unknown",
    );
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

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

test("visibility suspends polling and reconnect uses five second retry", async (t) => {
  t.mock.timers.enable({ apis: ["setTimeout"] });
  let hidden = false;
  let calls = 0;
  let online = false;
  const controller = new ViewerController(
    new AccessClient(
      async () => {
        calls++;
        if (!online) throw TypeError("offline");
        return response(index([]));
      },
      { getItem: () => "test-access", setItem() {}, removeItem() {} },
    ),
    () => hidden,
  );
  const stop = controller.start();
  try {
    await controller.open(null);
    assert.equal(controller.getSnapshot().down, true);
    hidden = true;
    t.mock.timers.tick(5000);
    await settle();
    assert.equal(calls, 1);
    hidden = false;
    online = true;
    document.dispatchEvent(new Event("visibilitychange"));
    await settle();
    assert.equal(calls, 2);
    assert.equal(controller.getSnapshot().down, false);
    t.mock.timers.tick(9999);
    await settle();
    assert.equal(calls, 2);
    t.mock.timers.tick(1);
    await settle();
    assert.equal(calls, 3);
  } finally {
    stop();
    t.mock.timers.reset();
  }
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

test("an obsolete unresolved index cannot suspend subsequent polling", async (t) => {
  t.mock.timers.enable({ apis: ["setTimeout"] });
  let calls = 0;
  let resolveStale!: (value: Response) => void;
  const stale = new Promise<Response>((resolve) => {
    resolveStale = resolve;
  });
  const controller = new ViewerController(
    new AccessClient(
      async () => {
        calls++;
        return calls === 2 ? stale : response(index([]));
      },
      { getItem: () => "test-access", setItem() {}, removeItem() {} },
    ),
    () => false,
  );
  const stop = controller.start();
  try {
    await controller.open(null);
    const old = controller.refresh();
    await controller.refresh();
    t.mock.timers.tick(10000);
    await settle();
    assert.equal(calls, 4);
    resolveStale(response(index([])));
    await old;
  } finally {
    resolveStale(response(index([])));
    stop();
    t.mock.timers.reset();
  }
});

test("alert recovery actions occupy the installed action region in every error state", async () => {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  let mode: "network" | "document" = "network";
  const controller = new ViewerController(
    new AccessClient(
      async (url, init) => {
        if (mode === "network") throw TypeError("offline");
        if (init?.method === "POST") return new Response(null, { status: 500 });
        if (String(url) === "/api/index") return response(index());
        return new Response(null, { status: 500 });
      },
      { getItem: () => "test-access", setItem() {}, removeItem() {} },
    ),
    () => false,
  );
  const assertActions = () => {
    const buttons = [
      ...container.querySelectorAll<HTMLButtonElement>(
        '[data-slot="alert"] button',
      ),
    ];
    assert(buttons.length > 0);
    for (const button of buttons)
      assert(
        button.closest('[data-slot="alert-action"]'),
        `Alert action ${button.textContent} must not occupy the icon grid column`,
      );
  };
  try {
    await act(async () => {
      await controller.open(null);
      root.render(<ViewerApplication controller={controller} />);
      await settle();
    });
    assertActions();
    mode = "document";
    await act(async () => {
      await controller.refresh();
      await settle();
      await controller.pin();
      await settle();
    });
    assert.equal(
      controller.getSnapshot().documents[
        controller.getSnapshot().selectedId ?? ""
      ]?.state ?? "empty",
      "error",
    );
    assert.equal(controller.getSnapshot().notice, "writeError");
    assertActions();
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

test("selected session exposes its complete identifier to keyboard users without an opening prompt", async () => {
  const { ViewerShell } = await import("../shell");
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const id = "abcdef0123456789".repeat(4);
  const item = {
    id,
    project: "Hive",
    projectKey: "hive",
    ticket: "#89",
    title: "Comprobación de configuración",
    createdAt: "19 Oct",
    classification: "local" as const,
  };
  try {
    await act(async () => {
      root.render(
        <ViewerShell
          sessions={[item]}
          selectedId={id}
          state="ready"
          onRetry={() => {}}
        />,
      );
    });
    const region = container.querySelector<HTMLElement>(
      '[aria-label="Session identifier"]',
    );
    assert(
      region,
      "Complete session identifier must have an accessible region",
    );
    assert.equal(region.textContent, id);
    region.focus();
    assert.equal(document.activeElement, region);
    assert(
      !container.textContent?.includes("Open a session in the navigation"),
    );
    await act(async () => {
      root.render(
        <ViewerShell sessions={[item]} state="ready" onRetry={() => {}} />,
      );
    });
    assert(container.textContent?.includes("Open a session in the navigation"));
    assert.equal(
      container.querySelector('[aria-label="Session identifier"]'),
      null,
    );
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

test("cold index loading replaces the full canvas without speculative session chrome", async () => {
  let release!: (response: Response) => void;
  const delayed = new Promise<Response>((resolve) => {
    release = resolve;
  });
  let indexReads = 0;
  const controller = new ViewerController(
    new AccessClient(
      async (url) => {
        if (String(url) === "/api/index") {
          indexReads++;
          return delayed;
        }
        return new Promise<Response>(() => {});
      },
      { getItem: () => "test-access", setItem() {}, removeItem() {} },
    ),
    () => false,
  );
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  try {
    const opening = controller.open(null);
    await act(async () => {
      root.render(<ViewerApplication controller={controller} />);
      await settle();
    });
    const canvas = container.querySelector('[data-slot="viewer-canvas"]');
    assert(canvas, "The loading state must own the content canvas");
    assert.equal(canvas.getAttribute("aria-busy"), "true");
    assert(
      canvas.querySelector(
        '[role="status"][aria-label="Loading walkthroughs"]',
      ),
    );
    assert.equal(canvas.querySelector('[data-slot="frame"]'), null);
    assert.equal(canvas.querySelector("h1"), null);
    const title =
      "Comprobación de configuración y revisión de navegación con nombres completos";
    await act(async () => {
      release(response(index([session(1, { title })])));
      await opening;
      await settle();
    });
    assert.equal(
      container.querySelector('[data-slot="viewer-canvas"]'),
      canvas,
    );
    assert.equal(canvas.getAttribute("aria-busy"), "false");
    assert(canvas.querySelector('[data-slot="frame"]'));
    assert.equal(canvas.querySelector("h1")?.textContent, title);
    const frame = canvas.querySelector('[data-slot="frame"]');
    await act(async () => {
      await controller.refresh();
      await settle();
    });
    assert.equal(canvas.querySelector('[data-slot="frame"]'), frame);
    assert.equal(canvas.getAttribute("aria-busy"), "false");
    assert.equal(indexReads, 2);
  } finally {
    await act(async () => root.unmount());
    container.remove();
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

test("tracked copies stay outside the normal tree through polling and retain focus and actions", async () => {
  let data = index([session(1)], {
    excluded_sessions: [
      session(2, {
        source_status: "tracked",
        project_key: "project-1",
        project_name: "Hive",
        ticket: "#89",
        fileBase: undefined,
      }),
    ],
  });
  const writes: string[] = [];
  const controller = new ViewerController(
    new AccessClient(
      async (url, init) => {
        if (init?.method === "POST") {
          writes.push(String(url));
          return new Response(null, { status: 204 });
        }
        if (init?.method === "DELETE") {
          writes.push(String(url));
          data = index([], { excluded_sessions: [] });
          return new Response(null, { status: 204 });
        }
        return String(url) === "/api/index"
          ? response(data)
          : new Promise<Response>(() => {});
      },
      { getItem: () => "test-access", setItem() {}, removeItem() {} },
    ),
    () => false,
  );
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  try {
    await act(async () => {
      await controller.open(null);
      root.render(<ViewerApplication controller={controller} />);
      await settle();
    });
    const normal = container.querySelector('nav[aria-label="Walkthroughs"]');
    const excluded = container.querySelector(
      'section[aria-label="Set-aside copies"]',
    );
    assert(normal);
    const projectToggle = normal.querySelector<HTMLButtonElement>(
      '[data-slot="collapsible-trigger"]',
    );
    assert(projectToggle);
    if (projectToggle.getAttribute("aria-expanded") !== "true")
      await act(async () => projectToggle.click());
    assert(excluded, "Tracked copies require a separate explanatory region");
    assert(normal.querySelector(`[data-session-key="${key(1)}"]`));
    assert.equal(normal.querySelector(`[data-session-key="${key(2)}"]`), null);
    assert(excluded.querySelector(`[data-session-key="${key(2)}"]`));
    const current = normal.querySelector<HTMLButtonElement>(
      `[data-session-key="${key(1)}"]`,
    );
    assert(current);
    current.focus();
    data = index([], {
      excluded_sessions: [
        session(1, {
          source_status: "tracked",
          project_key: "project-1",
          project_name: "Hive",
          ticket: "#89",
          fileBase: undefined,
        }),
      ],
    });
    await act(async () => {
      await controller.refresh();
      await settle();
    });
    assert.equal(
      controller.getSnapshot().documents[
        controller.getSnapshot().selectedId ?? ""
      ]?.state ?? "empty",
      "excluded",
    );
    assert.equal(
      container.querySelector(
        'nav[aria-label="Walkthroughs"] [data-session-key]',
      ),
      null,
    );
    assert.equal(
      document.activeElement?.getAttribute("data-session-key"),
      key(1),
    );
    assert(
      container.textContent?.includes("Only set-aside copies are available"),
    );
    data = index([session(1)], { excluded_sessions: [] });
    await act(async () => {
      await controller.refresh();
      await settle();
    });
    assert(
      container.querySelector(
        'nav[aria-label="Walkthroughs"] [data-session-key]',
      ),
    );
    assert.equal(
      container.querySelector('section[aria-label="Set-aside copies"]'),
      null,
    );
    assert.equal(
      document.activeElement?.getAttribute("data-session-key"),
      key(1),
    );
    data = index([], {
      excluded_sessions: [
        session(1, { source_status: "tracked", fileBase: undefined }),
      ],
    });
    await act(async () => {
      await controller.refresh();
      await settle();
    });
    assert.equal(
      controller.getSnapshot().documents[
        controller.getSnapshot().selectedId ?? ""
      ]?.state ?? "empty",
      "excluded",
    );
    assert.equal(
      document.activeElement?.getAttribute("data-session-key"),
      key(1),
    );
    const keep = [
      ...container.querySelectorAll<HTMLButtonElement>("button"),
    ].find((button) => button.getAttribute("aria-label") === "Keep");
    assert(keep);
    assert.equal(keep.disabled, false);
    const remove = [
      ...container.querySelectorAll<HTMLButtonElement>("button"),
    ].find((button) => button.getAttribute("aria-label") === "Delete copy");
    assert(remove);
    assert.equal(remove.disabled, false);
    await act(async () => {
      await controller.pin();
      await settle();
    });
    assert(writes.some((path) => path.endsWith("/pin")));
    await act(async () => {
      await controller.remove();
      await settle();
    });
    assert(writes.includes(`/api/sessions/${key(1)}`));
    assert.equal(container.querySelector("[data-session-key]"), null);
    assert.equal(document.activeElement?.id, "document");
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
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
      ["Archive copy", "Keep", "Delete copy"],
    );
    assert(actions.textContent?.includes("Closed"));
    assert.equal(actions.querySelector('[aria-haspopup="menu"]'), null);
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

test("session composition groups tickets, stacks siblings, deduplicates deliveries and targets segment actions", async () => {
  const canonical = await import(
    new URL(
      "../../../../../content/skills/flow-walkthrough/assets/walkthrough.mjs",
      import.meta.url,
    ).href
  );
  const loadModule = () => Promise.resolve(canonical);
  const data = index([
    session(1, {
      session_id: "conversation",
      session_status: "in_progress",
      created_at: "2026-10-21",
    }),
    session(2, { session_id: "conversation", created_at: "2026-10-22" }),
  ]);
  data.projects[0].tickets = [
    { ticket: "#89", sessions: [data.projects[0].tickets[0].sessions[0]] },
    { ticket: "#90", sessions: [data.projects[0].tickets[0].sessions[1]] },
  ];
  const writes: string[] = [];
  const controller = new ViewerController(
    new AccessClient(
      async (url, init) => {
        if (init?.method === "POST") {
          writes.push(String(url));
          return new Response(null, { status: 204 });
        }
        if (String(url) === "/api/index") return response(data);
        return response({
          title: "Review",
          candidate: { branch: "fixture", commit: "working tree" },
          session: { id: "conversation", host: "Codex" },
          created_at: "2026-10-21",
          lang: "en",
          kind: "first_view",
          flows: [],
          checks: [],
          delivery: [1, 2, ...(String(url).endsWith(key(2)) ? [3] : [])].map(
            (n) => ({
              kind: "pull_request",
              label: `Change ${n}`,
              url: `https://example.com/pull/${n}`,
            }),
          ),
        });
      },
      { getItem: () => "test-access", setItem() {}, removeItem() {} },
    ),
    () => false,
  );
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  try {
    await act(async () => {
      await controller.open(null);
      root.render(
        <ViewerApplication controller={controller} loadModule={loadModule} />,
      );
      await settle();
    });
    assert.equal(
      container.querySelectorAll(
        'nav[aria-label="Walkthroughs"] [data-session-key]',
      ).length,
      1,
    );
    assert.equal(container.querySelectorAll("[data-segment-key]").length, 2);
    assert.match(container.textContent ?? "", /Initial work/);
    assert.match(container.textContent ?? "", /Continuation/);
    const deliveries = container.querySelector(
      'section[aria-label="Deliveries"]',
    );
    assert(deliveries);
    assert.equal(deliveries.querySelectorAll("a").length, 2);
    assert.match(deliveries.textContent ?? "", /View all \(3\)/);
    const allButton = [
      ...deliveries.querySelectorAll<HTMLButtonElement>("button"),
    ].find((button) => button.textContent?.includes("View all"));
    assert(allButton);
    allButton.focus();
    await act(async () => {
      allButton.click();
      await settle();
    });
    const dialog = document.querySelector('[role="dialog"]');
    assert(dialog);
    assert.equal(dialog.querySelectorAll("a").length, 3);
    const close = [
      ...dialog.querySelectorAll<HTMLButtonElement>("button"),
    ].find((button) => button.textContent === "Close");
    assert(close);
    await act(async () => {
      close.click();
      await new Promise((resolve) => setTimeout(resolve, 20));
    });
    assert.equal(document.activeElement, allButton);

    const search = container.querySelector<HTMLInputElement>(
      'input[type="search"]',
    );
    assert(search);
    await act(async () => {
      Object.getOwnPropertyDescriptor(
        window.HTMLInputElement.prototype,
        "value",
      )?.set?.call(search, "#90");
      search.dispatchEvent(new Event("input", { bubbles: true }));
    });
    assert.equal(
      container.querySelectorAll(
        'nav[aria-label="Walkthroughs"] [data-session-key]',
      ).length,
      1,
    );
    const jump = [
      ...container.querySelectorAll<HTMLButtonElement>("button"),
    ].find((button) => button.textContent === "#90 · Review 2");
    assert(jump);
    await act(async () => {
      jump.click();
      await settle();
    });
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 10));
    });
    assert.equal(controller.getSnapshot().selectedId, key(2));
    assert.equal(document.activeElement?.id, `heading-${key(2)}`);
    assert.equal(container.querySelectorAll("[data-segment-key]").length, 2);
    const keep = container.querySelector<HTMLButtonElement>(
      `[data-segment-key="${key(1)}"] button[aria-label="Keep"]`,
    );
    assert(keep);
    await act(async () => {
      keep.click();
      await settle();
    });
    assert(writes.includes(`/api/sessions/${key(1)}/pin`));
    assert.equal(controller.getSnapshot().selectedId, key(2));
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

test("development preview passes the production index adapter and displays two session segments", async () => {
  const { DevelopmentPreview } = await import("../dev/preview");
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  try {
    await act(async () => {
      root.render(<DevelopmentPreview />);
      await settle();
    });
    assert.equal(container.querySelectorAll("[data-segment-key]").length, 2);
    assert.equal(
      container.querySelectorAll(
        'nav[aria-label="Walkthroughs"] [data-session-key]',
      ).length,
      2,
    );
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

test("delivery metadata waits for the canonical helper, reports failure and retries without replacing documents", async () => {
  const { SessionDeliveries } = await import("../deliveries");
  const { groupSessions } = await import("../sessions");
  const { mapIndex } = await import("../api");
  const canonical = await import(
    new URL(
      "../../../../../content/skills/flow-walkthrough/assets/walkthrough.mjs",
      import.meta.url,
    ).href
  );
  const group = groupSessions(mapIndex(index()))[0];
  const documents = {
    [key(1)]: {
      revision: "one",
      state: "ready" as const,
      document: {
        key: key(1),
        hash: "hash",
        fileBase: session().fileBase,
        manifest: {
          title: "Delivery",
          delivery: [
            {
              kind: "pull_request" as const,
              label: "Canonical link",
              url: "https://example.com/pull/1",
            },
            {
              kind: "pull_request" as const,
              label: "Unsafe link",
              url: "https://bücher.example/pull/1",
            },
          ],
        },
      },
    },
  };
  let calls = 0;
  let validations = 0;
  const loaded = {
    ...canonical,
    validDeliveries(value: unknown) {
      validations++;
      return canonical.validDeliveries(value);
    },
  };
  let reject!: (error: Error) => void;
  const pending = new Promise<never>((_resolve, rejectPromise) => {
    reject = rejectPromise;
  });
  const loadModule = () => {
    calls++;
    return calls === 1 ? pending : Promise.resolve(loaded);
  };
  const retried: string[] = [];
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  try {
    await act(async () => {
      root.render(
        <SessionDeliveries
          group={group}
          documents={documents}
          onRetry={(key) => {
            retried.push(key);
          }}
          loadModule={loadModule}
        />,
      );
      await settle();
    });
    assert.equal(container.querySelectorAll("a").length, 0);
    assert(container.querySelector('[aria-label="Loading deliveries"]'));
    assert(!container.textContent?.includes("No delivery recorded"));
    await act(async () => {
      reject(Error("Module unavailable"));
      await settle();
    });
    assert(container.textContent?.includes("Could not read all deliveries"));
    assert(!container.textContent?.includes("No delivery recorded"));
    const retry = container.querySelector<HTMLButtonElement>("button");
    assert(retry);
    await act(async () => {
      retry.click();
      await settle();
    });
    assert.equal(calls, 2);
    assert(validations > 0);
    assert.equal(container.querySelectorAll("a").length, 1);
    assert.equal(container.querySelector("a")?.textContent, "Canonical link");
    assert.deepEqual(retried, []);
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

test("reloading a remembered continuation waits for preceding documents, restores its position and focus once, and leaves polling focus alone", async () => {
  const { OPEN_KEY, TOKEN_KEY } = await import("../api");
  let data = index([
    session(1, { session_id: "remembered", created_at: "2026-10-20" }),
    session(2, { session_id: "remembered", created_at: "2026-10-21" }),
  ]);
  let releaseFirst!: (value: Response) => void;
  const first = new Promise<Response>((resolve) => {
    releaseFirst = resolve;
  });
  let releaseModule!: (module: {
    renderWalkthrough: WalkthroughRenderer;
    validDeliveries: () => [];
  }) => void;
  const module = new Promise<{
    renderWalkthrough: WalkthroughRenderer;
    validDeliveries: () => [];
  }>((resolve) => {
    releaseModule = resolve;
  });
  const loadModule = () => module;
  let firstRead = true;
  const controller = new ViewerController(
    new AccessClient(
      async (url) => {
        if (String(url) === "/api/index") return response(data);
        if (String(url).endsWith(key(1)) && firstRead) {
          firstRead = false;
          return first;
        }
        return response({
          title: String(url).endsWith(key(1))
            ? "Initial content"
            : "Continuation content",
        });
      },
      {
        getItem: (name) =>
          name === TOKEN_KEY
            ? "test-access"
            : name === OPEN_KEY
              ? key(2)
              : null,
        setItem() {},
        removeItem() {},
      },
    ),
    () => false,
  );
  const render: WalkthroughRenderer = (root, manifest) => {
    const paragraph = document.createElement("p");
    paragraph.textContent = manifest.title;
    root.append(paragraph);
  };
  const scrolled: number[] = [];
  let scrollOffset = 0;
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  try {
    await act(async () => {
      await controller.open(null);
      root.render(
        <ViewerApplication controller={controller} loadModule={loadModule} />,
      );
      await settle();
    });
    const canvas = container.querySelector<HTMLElement>(
      '[data-slot="viewer-canvas"]',
    );
    assert(canvas);
    const section = container.querySelector<HTMLElement>(
      `[data-segment-key="${key(2)}"]`,
    );
    assert(section);
    const actions = section.querySelector<HTMLElement>('[role="group"]');
    assert(actions);
    canvas.getBoundingClientRect = () => new window.DOMRect(0, 83, 400, 700);
    section.getBoundingClientRect = () =>
      new window.DOMRect(0, 83 + 3000 - scrollOffset, 400, 500);
    actions.getBoundingClientRect = () =>
      new window.DOMRect(0, section.getBoundingClientRect().top + 8, 200, 32);
    Object.defineProperty(canvas, "scrollTop", {
      configurable: true,
      get: () => scrollOffset,
      set: (value: number) => {
        scrollOffset = value;
        scrolled.push(value);
        assert(
          document.body.textContent?.includes("Initial content"),
          "Restore must wait for the preceding document to render",
        );
      },
    });
    assert.equal(controller.getSnapshot().selectedId, key(2));
    assert.equal(scrolled.length, 0);
    await act(async () => {
      releaseModule({ renderWalkthrough: render, validDeliveries: () => [] });
      await settle();
    });
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 10));
    });
    assert.equal(
      scrolled.length,
      0,
      "A ready continuation must wait for its preceding segment",
    );
    await act(async () => {
      releaseFirst(response({ title: "Initial content" }));
      await settle();
    });
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 10));
    });
    const heading = container.querySelector<HTMLElement>(`#heading-${key(2)}`);
    assert(heading);
    assert.equal(scrolled.length, 1);
    assert.equal(canvas.scrollTop, 3000);
    assert(
      actions.getBoundingClientRect().top >= canvas.getBoundingClientRect().top,
    );
    assert(
      actions.getBoundingClientRect().bottom <=
        canvas.getBoundingClientRect().bottom,
    );
    assert.equal(document.activeElement, heading);
    const search = container.querySelector<HTMLInputElement>(
      'input[type="search"]',
    );
    assert(search);
    search.focus();
    data = index([
      session(1, {
        session_id: "remembered",
        created_at: "2026-10-20",
        content_hash: "new-content",
      }),
      session(2, { session_id: "remembered", created_at: "2026-10-21" }),
    ]);
    await act(async () => {
      await controller.refresh();
      await settle();
    });
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 10));
    });
    assert.equal(document.activeElement, search);
    assert.equal(
      scrolled.length,
      1,
      "Polling must not repeat the restored jump",
    );
    scrollOffset = 2500;
    const jumps = container.querySelectorAll<HTMLButtonElement>(
      'nav[aria-label="Jump to a segment"] button',
    );
    await act(async () => {
      jumps[1].click();
      await settle();
    });
    await act(async () => {
      await new Promise((resolve) => setTimeout(resolve, 10));
    });
    assert.equal(canvas.scrollTop, 3000);
    assert.equal(scrolled.length, 2);
    assert.equal(document.activeElement, heading);
    assert(
      actions.getBoundingClientRect().top >= canvas.getBoundingClientRect().top,
    );
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

test("grouped session indicators preserve explicit closed, active and unknown states", async () => {
  const data = index([
    session(1, { session_id: "closed", session_status: "closed" }),
    session(2, { session_id: "closed", session_status: "closed" }),
    session(3, { session_id: "active", session_status: "closed" }),
    session(4, { session_id: "active", session_status: "in_progress" }),
    session(5, { session_id: "unknown", session_status: "closed" }),
    session(6, { session_id: "unknown" }),
  ]);
  const controller = new ViewerController(
    new AccessClient(async () => response(data), {
      getItem: () => "test-access",
      setItem() {},
      removeItem() {},
    }),
    () => false,
  );
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  try {
    await act(async () => {
      await controller.open(null);
      root.render(<ViewerApplication controller={controller} />);
      await settle();
    });
    for (const [n, expected] of [
      [1, "closed"],
      [3, "in_progress"],
      [5, "unknown"],
    ] as const) {
      const row = container.querySelector(`[data-session-key="${key(n)}"]`);
      assert(row);
      assert.equal(
        row
          .querySelector("[data-session-status]")
          ?.getAttribute("data-session-status"),
        expected,
      );
      assert.match(
        row.getAttribute("title") ?? "",
        /Closed does not mean approved/,
      );
      assert.match(row.textContent ?? "", /Closed does not mean approved/);
    }
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});

test("sidebar opens session headers while explicit segment jumps still target their headings", async () => {
  const data = index([
    session(1, { session_id: "first", created_at: "2026-10-20" }),
    session(2, { session_id: "first", created_at: "2026-10-21" }),
    session(3, {
      session_id: "other",
      session_status: "in_progress",
      created_at: "2026-10-22",
    }),
  ]);
  const controller = new ViewerController(
    new AccessClient(
      async (url) =>
        String(url) === "/api/index"
          ? response(data)
          : response({ title: "Content" }),
      {
        getItem: () => "test-access",
        setItem() {},
        removeItem() {},
      },
    ),
    () => false,
  );
  const loadModule = async () => ({
    renderWalkthrough: ((root) => {
      root.textContent = "Content";
    }) as WalkthroughRenderer,
    validDeliveries: () => [],
  });
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  const tick = () => new Promise((resolve) => setTimeout(resolve, 20));
  try {
    await act(async () => {
      await controller.open(null);
      controller.select(key(1));
      root.render(
        <ViewerApplication controller={controller} loadModule={loadModule} />,
      );
      await settle();
    });
    await act(tick);
    const canvas = container.querySelector<HTMLElement>(
      '[data-slot="viewer-canvas"]',
    );
    assert(canvas);
    let offset = 800;
    Object.defineProperty(canvas, "scrollTop", {
      configurable: true,
      get: () => offset,
      set: (value: number) => {
        offset = value;
      },
    });
    canvas.getBoundingClientRect = () => new window.DOMRect(0, 80, 600, 700);
    const measureSections = () => {
      for (const section of container.querySelectorAll<HTMLElement>(
        "[data-segment-key]",
      )) {
        const top = section.dataset.segmentKey === key(2) ? 1800 : 200;
        section.getBoundingClientRect = () =>
          new window.DOMRect(0, 80 + top - offset, 600, 500);
      }
    };
    const click = async (selector: string) => {
      const button = container.querySelector<HTMLButtonElement>(selector);
      assert(button, selector);
      await act(async () => {
        button.click();
        await settle();
      });
      measureSections();
      await act(tick);
    };
    await click(`[data-session-key="${key(3)}"]`);
    assert.equal(
      canvas.scrollTop,
      0,
      "Changing session must retain the session header",
    );
    await click(`[data-session-key="${key(1)}"]`);
    assert.equal(canvas.scrollTop, 0);
    await click('nav[aria-label="Jump to a segment"] button:nth-child(2)');
    assert.equal(canvas.scrollTop, 1800);
    assert.equal(document.activeElement?.id, `heading-${key(2)}`);
    await click(`[data-session-key="${key(3)}"]`);
    await click(`[data-session-key="${key(1)}"]`);
    assert.equal(
      canvas.scrollTop,
      0,
      "A remembered continuation must not hide the session header",
    );
    assert.equal(controller.getSnapshot().selectedId, key(1));
    await click('nav[aria-label="Jump to a segment"] button:nth-child(2)');
    await click(`[data-session-key="${key(1)}"]`);
    assert.equal(
      canvas.scrollTop,
      0,
      "Reopening the current session must also return to its header",
    );
  } finally {
    await act(async () => root.unmount());
    container.remove();
  }
});
