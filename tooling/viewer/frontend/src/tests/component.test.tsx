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
  session_id: "fixture",
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
      { ticket: "#89", sessions: [session(5, { session_status: "closed" })] },
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
    assert.equal(container.querySelectorAll(".navigation-branch").length, 4);
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
    assert.equal(controller.getSnapshot().documentState, "error");
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

test("renderer loading shares in-flight work but does not keep a rejected attempt", async () => {
  const { loadWalkthroughRenderer } = await import("../document");
  const failed = loadWalkthroughRenderer();
  assert.equal(loadWalkthroughRenderer(), failed);
  await assert.rejects(failed);
  const retry = loadWalkthroughRenderer();
  assert.notEqual(retry, failed, "Retry must initiate another renderer load");
  await assert.rejects(retry);
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
    assert.equal(controller.getSnapshot().documentState, "excluded");
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
    assert.equal(controller.getSnapshot().documentState, "excluded");
    assert.equal(
      document.activeElement?.getAttribute("data-session-key"),
      key(1),
    );
    const keep = [
      ...container.querySelectorAll<HTMLButtonElement>("button"),
    ].find((button) => button.textContent === "Keep");
    assert(keep);
    assert.equal(keep.disabled, false);
    const remove = [
      ...container.querySelectorAll<HTMLButtonElement>("button"),
    ].find((button) => button.textContent === "Delete copy");
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
