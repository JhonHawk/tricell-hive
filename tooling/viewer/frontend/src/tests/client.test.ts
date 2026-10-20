import assert from "node:assert/strict";
import test from "node:test";
import { AccessClient, mapIndex, takeCode } from "../api";
import { ViewerController } from "../controller";

const key = (n: number) => String(n).padStart(64, "0");
const session = (n = 1, extra = {}) => ({
  key: key(n),
  source: "/repo/_support/workspace/run",
  source_status: "local",
  content_hash: "hash",
  created_at: "2026-10-19T15:00:00Z",
  ingested_at: "2026-10-19T15:01:00Z",
  session_id: "run",
  title: `Review ${n}`,
  fileBase: `/files/${key(n)}/${"a".repeat(64)}/`,
  ...extra,
});
const index = (sessions = [session()], extra = {}) => ({
  projects: [
    { key: "project-1", name: "Hive", tickets: [{ ticket: "#89", sessions }] },
  ],
  excluded_sessions: [],
  repositories: [{ root: "/repo", absent: false }],
  errors: [],
  ttl_seconds: 86400,
  ...extra,
});
const response = (data: unknown, status = 200) =>
  new Response(JSON.stringify(data), { status });
const deferred = () => {
  let resolve!: (value: Response) => void;
  const promise = new Promise<Response>((r) => {
    resolve = r;
  });
  return { promise, resolve };
};
const settle = async () => {
  for (let i = 0; i < 12; i++) await Promise.resolve();
};
function setup(fetcher: typeof fetch) {
  const access = new AccessClient(fetcher, {
    getItem: () => "test-access",
    setItem() {},
    removeItem() {},
  });
  return new ViewerController(access, () => false);
}

test("opening code is removed synchronously and exchange is reused with unavailable storage", async () => {
  const location = {
    hash: "#code=one-use&token=ignored",
    pathname: "/",
    search: "",
  };
  let replaced = false;
  assert.equal(
    takeCode(location, {
      replaceState() {
        replaced = true;
      },
    }),
    "one-use",
  );
  assert.equal(replaced, true);
  let calls = 0;
  const access = new AccessClient(
    async () => {
      calls++;
      return response({ token: "test-access" });
    },
    {
      getItem() {
        throw Error("disabled");
      },
      setItem() {
        throw Error("disabled");
      },
      removeItem() {
        throw Error("disabled");
      },
    },
  );
  const opening = access.establish("one-use");
  await Promise.all([opening, opening]);
  assert.equal(calls, 1);
  assert.equal(access.token, "test-access");
});

test("adapter maps explicit status independently from kind/provenance and keeps project keys", () => {
  const data = index([
    session(1, {
      session_status: "closed",
      source_status: "unknown",
      kind: "first_view",
    }),
    session(2, { session_status: "in_progress", kind: "final" }),
    session(3, { kind: "final" }),
  ]);
  data.projects.push({
    key: "project-2",
    name: "Hive",
    tickets: [{ ticket: "#89", sessions: [session(4)] }],
  });
  const mapped = mapIndex(data);
  assert.deepEqual(
    mapped.map((s) => s.status),
    ["closed", "in_progress", "unknown", "unknown"],
  );
  assert.equal(mapped[0].classification, "unknown");
  assert.notEqual(mapped[0].projectKey, mapped[3].projectKey);
});

test("authenticated reads use headers and keep capability separate from token", async () => {
  const calls: { url: string; init?: RequestInit }[] = [];
  const controller = setup(async (url, init) => {
    calls.push({ url: String(url), init });
    return response(
      String(url) === "/api/index" ? index() : { title: "Review", flows: [] },
    );
  });
  await controller.open(null);
  await settle();
  assert.equal(controller.getSnapshot().document?.fileBase, session().fileBase);
  assert.equal(
    new Headers(calls[0].init?.headers).get("X-Hive-Token"),
    "test-access",
  );
  assert(calls.every((c) => !c.url.includes("test-access")));
});

test("out of order indices and manifests cannot replace newer selection", async () => {
  const reads = [deferred(), deferred()];
  let indexCalls = 0;
  const docs = new Map<string, ReturnType<typeof deferred>>();
  const controller = setup(async (url) =>
    String(url) === "/api/index"
      ? reads[indexCalls++].promise
      : (
          docs.get(String(url)) ??
          (() => {
            const d = deferred();
            docs.set(String(url), d);
            return d;
          })()
        ).promise,
  );
  const first = controller.open(null);
  await settle();
  const second = controller.refresh();
  reads[1].resolve(response(index([session(2), session(1)])));
  await second;
  await settle();
  controller.select(key(1));
  await settle();
  docs
    .get(`/api/sessions/${key(1)}`)
    ?.resolve(response({ title: "new selection", flows: [] }));
  await settle();
  docs
    .get(`/api/sessions/${key(2)}`)
    ?.resolve(response({ title: "old selection", flows: [] }));
  reads[0].resolve(response(index([session(3)])));
  await first;
  await settle();
  assert.equal(controller.getSnapshot().selectedId, key(1));
  assert.equal(
    controller.getSnapshot().document?.manifest.title,
    "new selection",
  );
});

test("delete invalidates an in-flight poll and retains a local tombstone", async () => {
  let polling = false;
  const stale = deferred();
  const controller = setup(async (url, init) =>
    init?.method === "DELETE"
      ? new Response(null, { status: 204 })
      : String(url) === "/api/index"
        ? polling
          ? stale.promise
          : response(index([session(1), session(2)]))
        : response({ title: "Review", flows: [] }),
  );
  await controller.open(null);
  await settle();
  polling = true;
  const poll = controller.refresh();
  await controller.remove();
  stale.resolve(response(index([session(1), session(2)])));
  await poll;
  await settle();
  assert(!controller.getSnapshot().sessions.some((s) => s.id === key(1)));
  assert.equal(controller.getSnapshot().selectedId, key(2));
});

test("classification changes without hash remove the document; status updates arrive by polling", async () => {
  let data = index();
  const controller = setup(async (url) =>
    response(
      String(url) === "/api/index" ? data : { title: "Review", flows: [] },
    ),
  );
  await controller.open(null);
  await settle();
  data = index([
    session(1, { session_status: "closed", source_status: "unknown" }),
  ]);
  await controller.refresh();
  assert.equal(controller.getSnapshot().sessions[0].status, "closed");
  assert.equal(controller.getSnapshot().sessions[0].classification, "unknown");
  data = index([], {
    excluded_sessions: [session(1, { source_status: "tracked" })],
  });
  await controller.refresh();
  await settle();
  assert.equal(controller.getSnapshot().document, null);
  assert.equal(controller.getSnapshot().documentState, "excluded");
});

test("403 blocks access, clears stored token and prevents further reads", async () => {
  let calls = 0;
  const controller = setup(async () => {
    calls++;
    return response({}, 403);
  });
  await controller.open(null);
  await controller.refresh();
  assert.equal(controller.getSnapshot().access, "rejected");
  assert.equal(calls, 1);
});

test("pin and unpin reconcile with fresh index while preserving the open document", async () => {
  let pinned = false;
  const methods: string[] = [];
  let documents = 0;
  const controller = setup(async (url, init) => {
    if (String(url).endsWith("/pin")) {
      methods.push(init?.method ?? "GET");
      pinned = init?.method === "POST";
      return new Response(null, { status: 204 });
    }
    if (String(url) === "/api/index")
      return response(index([session(1, { pinned })]));
    documents++;
    return response({ title: "Review", flows: [] });
  });
  await controller.open(null);
  await settle();
  const document = controller.getSnapshot().document;
  await controller.pin();
  await settle();
  assert.equal(controller.getSnapshot().sessions[0].pinned, true);
  await controller.pin();
  await settle();
  assert.equal(controller.getSnapshot().sessions[0].pinned, false);
  assert.deepEqual(methods, ["POST", "DELETE"]);
  assert.equal(controller.getSnapshot().document, document);
  assert.equal(documents, 1);
});

test("missing access makes no requests and remembered selection is optional", async () => {
  let requests = 0;
  const empty = new ViewerController(
    new AccessClient(
      async () => {
        requests++;
        return response(index());
      },
      { getItem: () => "", setItem() {}, removeItem() {} },
    ),
    () => false,
  );
  await empty.open(null);
  assert.equal(empty.getSnapshot().access, "missing");
  assert.equal(requests, 0);
  const remembered = new ViewerController(
    new AccessClient(
      async (url) =>
        response(
          String(url) === "/api/index"
            ? index([session(1), session(2)])
            : { title: "remembered", flows: [] },
        ),
      {
        getItem: (key) =>
          key === "hive-view-open" ? session(2).key : "test-access",
        setItem() {},
        removeItem() {},
      },
    ),
    () => false,
  );
  await remembered.open(null);
  await settle();
  assert.equal(remembered.getSnapshot().selectedId, session(2).key);
});

test("a tracked response cannot load files and access replacement invalidates old reads", async () => {
  let docCalls = 0;
  const controller = setup(async (url) => {
    if (String(url) === "/api/index") return response(index([session()]));
    docCalls++;
    return response({}, 409);
  });
  await controller.open(null);
  await settle();
  assert.equal(controller.getSnapshot().documentState, "excluded");
  assert.equal(controller.getSnapshot().document, null);
  assert.equal(docCalls, 1);
  const first = deferred();
  let calls = 0;
  const client = new AccessClient(
    async (url) =>
      String(url).includes("exchange")
        ? response({ token: "replacement-access" })
        : ++calls === 1
          ? first.promise
          : response(index([])),
    { getItem: () => "test-access", setItem() {}, removeItem() {} },
  );
  const replacement = new ViewerController(client, () => false);
  const old = replacement.open(null);
  await settle();
  await replacement.open("new-code");
  first.resolve(response(index()));
  await old;
  await settle();
  assert.equal(replacement.getSnapshot().sessions.length, 0);
  assert.equal(replacement.getSnapshot().access, "ready");
});

test("a late delete response preserves the replacement manifest already loading", async () => {
  const deletion = deferred();
  const replacement = deferred();
  let library = index([session(1), session(2)]);
  let replacementRequests = 0;
  const controller = setup(async (url, init) => {
    if (init?.method === "DELETE") return deletion.promise;
    if (String(url) === "/api/index") return response(library);
    if (String(url) === `/api/sessions/${key(2)}`) {
      replacementRequests++;
      return replacement.promise;
    }
    return response({ title: "Original", flows: [] });
  });
  await controller.open(null);
  await settle();
  const deleting = controller.remove();
  library = index([session(2)]);
  await controller.refresh();
  await settle();
  assert.equal(controller.getSnapshot().selectedId, key(2));
  assert.equal(controller.getSnapshot().documentState, "loading");
  deletion.resolve(new Response(null, { status: 204 }));
  await deleting;
  await settle();
  replacement.resolve(response({ title: "Replacement", flows: [] }));
  await settle();
  await controller.refresh();
  assert.equal(controller.getSnapshot().selectedId, key(2));
  assert.equal(controller.getSnapshot().documentState, "ready");
  assert.equal(
    controller.getSnapshot().document?.manifest.title,
    "Replacement",
  );
  assert.equal(replacementRequests, 1);
});

test("default fetch retains the browser global receiver for exchange and reads", async () => {
  const nativeFetch = globalThis.fetch;
  let requests = 0;
  globalThis.fetch = async function (this: unknown, url) {
    assert.equal(
      this,
      globalThis,
      "Browser fetch requires its global receiver",
    );
    requests++;
    return response(
      String(url).includes("exchange")
        ? { token: "test-access" }
        : { ready: true },
    );
  };
  try {
    const client = new AccessClient();
    await client.establish("one-use");
    assert.deepEqual(await client.request("/api/index"), { ready: true });
    assert.equal(requests, 2);
  } finally {
    globalThis.fetch = nativeFetch;
  }
});

test("refreshing or losing the index connection retains the open document", async () => {
  const poll = deferred();
  let reads = 0;
  const controller = setup(async (url) =>
    String(url) === "/api/index"
      ? ++reads === 1
        ? response(index())
        : poll.promise
      : response({ title: "Open document", flows: [] }),
  );
  await controller.open(null);
  await settle();
  const open = controller.getSnapshot().document;
  assert(open);
  const refreshing = controller.refresh();
  await settle();
  assert.equal(controller.getSnapshot().document, open);
  assert.equal(controller.getSnapshot().documentState, "ready");
  poll.resolve(response({}, 500));
  await refreshing;
  assert.equal(controller.getSnapshot().document, open);
  assert.equal(controller.getSnapshot().indexError, true);
});
