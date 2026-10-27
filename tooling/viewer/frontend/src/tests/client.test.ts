import assert from "node:assert/strict";
import test from "node:test";
import {
  AccessClient,
  ARCHIVED_KEY,
  mapIndex,
  OPEN_KEY,
  TOKEN_KEY,
  takeCode,
} from "../api";
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
  assert.equal(
    (
      controller.getSnapshot().documents[
        controller.getSnapshot().selectedId ?? ""
      ]?.document ?? null
    )?.fileBase,
    session().fileBase,
  );
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
    (
      controller.getSnapshot().documents[
        controller.getSnapshot().selectedId ?? ""
      ]?.document ?? null
    )?.manifest.title,
    "new selection",
  );
});

test("delete invalidates an in-flight poll and trusts the reloaded session inventory", async () => {
  const stale = deferred();
  let reads = 0;
  let deleted = false;
  const controller = setup(async (url, init) => {
    if (init?.method === "DELETE") {
      deleted = true;
      return new Response(null, { status: 204 });
    }
    if (String(url) === "/api/index") {
      if (++reads === 2) return stale.promise;
      return response(index(deleted ? [] : [session(1), session(2)]));
    }
    return response({ title: "Review", flows: [] });
  });
  await controller.open(null);
  const poll = controller.refresh();
  await controller.remove();
  stale.resolve(response(index([session(1), session(2)])));
  await poll;
  assert.equal(controller.getSnapshot().sessions.length, 0);
  assert.equal(controller.getSnapshot().selectedId, undefined);
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
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.document ?? null,
    null,
  );
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.state ?? "empty",
    "excluded",
  );
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
  const document =
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.document ?? null;
  await controller.pin();
  await settle();
  assert.equal(controller.getSnapshot().sessions[0].pinned, true);
  await controller.pin();
  await settle();
  assert.equal(controller.getSnapshot().sessions[0].pinned, false);
  assert.deepEqual(methods, ["POST", "DELETE"]);
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.document ?? null,
    document,
  );
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
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.state ?? "empty",
    "excluded",
  );
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.document ?? null,
    null,
  );
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
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.state ?? "empty",
    "loading",
  );
  deletion.resolve(new Response(null, { status: 204 }));
  await deleting;
  await settle();
  replacement.resolve(response({ title: "Replacement", flows: [] }));
  await settle();
  await controller.refresh();
  assert.equal(controller.getSnapshot().selectedId, key(2));
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.state ?? "empty",
    "ready",
  );
  assert.equal(
    (
      controller.getSnapshot().documents[
        controller.getSnapshot().selectedId ?? ""
      ]?.document ?? null
    )?.manifest.title,
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
  const open =
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.document ?? null;
  assert(open);
  const refreshing = controller.refresh();
  await settle();
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.document ?? null,
    open,
  );
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.state ?? "empty",
    "ready",
  );
  poll.resolve(response({}, 500));
  await refreshing;
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.document ?? null,
    open,
  );
  assert.equal(controller.getSnapshot().indexError, true);
});

test("archiving hides a copy, reconciles selection, and survives refresh without changing pin or status", async () => {
  let data = index([
    session(1, { pinned: true, session_status: "in_progress" }),
    session(2),
  ]);
  const storage = new Map<string, string>([[TOKEN_KEY, "test-access"]]);
  const methods: string[] = [];
  const controller = new ViewerController(
    new AccessClient(
      async (url, options) => {
        if (String(url).endsWith("/archive")) {
          methods.push(options?.method ?? "GET");
          data = index([
            session(1, {
              pinned: true,
              session_status: "in_progress",
              archived: options?.method === "POST",
            }),
            session(2),
          ]);
          return new Response(null, { status: 204 });
        }
        return response(
          String(url) === "/api/index" ? data : { title: "Walkthrough" },
        );
      },
      {
        getItem: (key) => storage.get(key) ?? null,
        setItem: (key, value) => {
          storage.set(key, value);
        },
        removeItem: (key) => {
          storage.delete(key);
        },
      },
    ),
  );
  await controller.open(null);
  controller.select(key(1));
  await settle();
  await controller.archive();
  await settle();
  assert.deepEqual(
    controller.getSnapshot().sessions.map((s) => s.id),
    [key(2)],
  );
  assert.equal(controller.getSnapshot().selectedId, key(2));
  assert.equal(controller.getSnapshot().notice, "archived");
  await controller.refresh();
  controller.setShowArchived(true);
  assert.equal(storage.get(ARCHIVED_KEY), "true");
  controller.select(key(1));
  await settle();
  const archived = controller
    .getSnapshot()
    .sessions.find((s) => s.id === key(1));
  assert.equal(archived?.archived, true);
  assert.equal(archived?.pinned, true);
  assert.equal(archived?.status, "in_progress");
  await controller.open(null);
  await settle();
  assert.equal(controller.getSnapshot().showArchived, true);
  assert.equal(controller.getSnapshot().selectedId, key(1));
  assert.equal(storage.get(OPEN_KEY), key(1));
  await controller.archive();
  await settle();
  assert.equal(
    controller.getSnapshot().sessions.find((s) => s.id === key(1))?.archived,
    false,
  );
  assert.deepEqual(methods, ["POST", "DELETE"]);
  controller.setShowArchived(false);
  assert.equal(controller.getSnapshot().selectedId, key(1));
});

test("failed archive leaves selection and copy unchanged and invalidates stale polling on success", async () => {
  const stale = deferred();
  let reads = 0;
  let fail = true;
  let archived = false;
  const controller = setup(async (url) => {
    if (String(url).endsWith("/archive")) {
      if (fail) return response({}, 500);
      archived = true;
      return new Response(null, { status: 204 });
    }
    if (String(url) === "/api/index") {
      reads++;
      if (reads === 2) return stale.promise;
      return response(index([session(1, { archived })]));
    }
    return response({ title: "Walkthrough" });
  });
  await controller.open(null);
  await settle();
  await controller.archive();
  assert.equal(controller.getSnapshot().selectedId, key(1));
  assert.equal(controller.getSnapshot().sessions[0].archived, false);
  assert.equal(controller.getSnapshot().notice, "writeError");
  assert.equal(controller.getSnapshot().writing, false);
  const poll = controller.refresh();
  fail = false;
  await controller.archive();
  await settle();
  assert.equal(controller.getSnapshot().sessions.length, 0);
  assert.equal(
    controller.getSnapshot().documents[
      controller.getSnapshot().selectedId ?? ""
    ]?.document ?? null,
    null,
  );
  stale.resolve(response(index([session(1, { archived: false })])));
  await poll;
  assert.equal(controller.getSnapshot().sessions.length, 0);
  controller.setShowArchived(true);
  assert.equal(controller.getSnapshot().sessions[0].archived, true);
});

test("groups preserve exact identity, order, tickets and search targets", async () => {
  const { groupSessions, matchingSegments } = await import("../sessions");
  const mapped = mapIndex(
    index([], {
      projects: [
        {
          key: "p",
          name: "One",
          tickets: [
            {
              ticket: "#1",
              sessions: [
                session(2, { host: "codex" }),
                session(1, {
                  host: "codex",
                  created_at: "",
                  ingested_at: "2026-10-18T00:00:00Z",
                }),
              ],
            },
            {
              ticket: "#2",
              sessions: [
                session(3, { host: "codex" }),
                session(4, { host: "claude" }),
                session(5),
                session(6, { session_id: "other" }),
                session(7, { host: "" }),
              ],
            },
          ],
        },
        {
          key: "q",
          name: "Two",
          tickets: [
            { ticket: "#1", sessions: [session(8, { host: "codex" })] },
          ],
        },
      ],
    }),
  );
  const groups = groupSessions(mapped);
  assert.equal(groups.length, 6);
  const shared = groups.find((group) => group.segments.length === 3);
  assert(shared);
  assert.deepEqual(
    shared.segments.map((s) => s.id),
    [key(1), key(2), key(3)],
  );
  assert.deepEqual(shared.tickets, ["#1", "#2"]);
  assert.deepEqual(
    matchingSegments(shared, "#2").map((s) => s.id),
    [key(3)],
  );
  assert.equal(shared.title, "Review 1");
});

test("siblings load independently, retry by key, and pin preserves cache", async () => {
  let fail = true;
  const calls: string[] = [];
  const controller = setup(async (url, options) => {
    calls.push(`${options?.method ?? "GET"} ${url}`);
    if (String(url) === "/api/index")
      return response(index([session(1), session(2)]));
    if (String(url).endsWith("/pin"))
      return new Response(null, { status: 204 });
    if (String(url).endsWith(key(2)) && fail) return response({}, 500);
    return response({ title: String(url) });
  });
  await controller.open(null);
  await settle();
  assert.equal(controller.getSnapshot().groups.length, 1);
  const first = controller.getSnapshot().documents[key(1)].document;
  assert(first);
  assert.equal(controller.getSnapshot().documents[key(2)].state, "error");
  await controller.pin(key(2));
  await settle();
  assert.equal(controller.getSnapshot().documents[key(1)].document, first);
  assert.equal(
    calls.filter((call) => call === `GET /api/sessions/${key(1)}`).length,
    1,
  );
  assert(calls.includes(`POST /api/sessions/${key(2)}/pin`));
  fail = false;
  controller.retryDocument(key(2));
  await settle();
  assert.equal(controller.getSnapshot().documents[key(2)].state, "ready");
});

test("archive and delete select the next visible session across ticket groups", async () => {
  let data = index([
    session(1),
    session(2),
    session(3, { session_id: "other" }),
  ]);
  const controller = setup(async (url, options) => {
    if (String(url).endsWith("/archive")) {
      data = index([
        session(1, { archived: true }),
        session(2, { archived: true }),
        session(3, { session_id: "other" }),
      ]);
      return new Response(null, { status: 204 });
    }
    if (options?.method === "DELETE") {
      data = index([
        session(1, { archived: true }),
        session(2, { archived: true }),
      ]);
      return new Response(null, { status: 204 });
    }
    return response(
      String(url) === "/api/index" ? data : { title: "Document" },
    );
  });
  await controller.open(null);
  controller.select(key(2));
  await controller.archive();
  assert.equal(controller.getSnapshot().selectedId, key(3));
  await controller.remove();
  assert.equal(controller.getSnapshot().selectedId, undefined);
  controller.setShowArchived(true);
  assert.equal(controller.getSnapshot().groups[0].segments.length, 2);
});

test("revision and classification changes invalidate only affected reads; late archived responses stay discarded", async () => {
  let data = index([session(1), session(2)]);
  const stale = deferred();
  const fresh = deferred();
  let secondReads = 0;
  const controller = setup(async (url) => {
    if (String(url) === "/api/index") return response(data);
    if (String(url).endsWith(key(2)))
      return ++secondReads === 1 ? stale.promise : fresh.promise;
    return response({ title: "stable" });
  });
  await controller.open(null);
  await settle();
  const stable = controller.getSnapshot().documents[key(1)].document;
  data = index([session(1), session(2, { content_hash: "new" })]);
  await controller.refresh();
  fresh.resolve(response({ title: "fresh" }));
  await settle();
  stale.resolve(response({ title: "stale" }));
  await settle();
  assert.equal(
    controller.getSnapshot().documents[key(2)].document?.manifest.title,
    "fresh",
  );
  assert.equal(controller.getSnapshot().documents[key(1)].document, stable);
  data = index([session(1), session(2, { archived: true })]);
  await controller.refresh();
  assert.equal(controller.getSnapshot().documents[key(2)], undefined);
  data = index([session(1)], {
    excluded_sessions: [session(2, { source_status: "tracked" })],
  });
  await controller.refresh();
  assert.equal(controller.getSnapshot().groups[0].segments.length, 1);
  controller.select(key(2));
  assert.equal(controller.getSnapshot().documents[key(2)].state, "excluded");
  assert.equal(secondReads, 2);
});

test("403 from one sibling clears every protected document and discards late siblings", async () => {
  const denied = deferred();
  const late = deferred();
  const controller = setup(async (url) => {
    if (String(url) === "/api/index")
      return response(index([session(1), session(2), session(3)]));
    if (String(url).endsWith(key(2))) return denied.promise;
    if (String(url).endsWith(key(3))) return late.promise;
    return response({ title: "protected" });
  });
  await controller.open(null);
  await settle();
  assert(controller.getSnapshot().documents[key(1)].document);
  denied.resolve(response({}, 403));
  await settle();
  late.resolve(response({ title: "late protected" }));
  await settle();
  assert.equal(controller.getSnapshot().access, "rejected");
  assert.deepEqual(controller.getSnapshot().documents, {});
  assert.deepEqual(controller.getSnapshot().groups, []);
});

test("group opening selects initial work and retains cached siblings; all archived groups disappear", async () => {
  let data = index([
    session(1),
    session(2),
    session(3, { session_id: "other" }),
  ]);
  let reads = 0;
  const controller = setup(async (url) => {
    if (String(url) === "/api/index") return response(data);
    reads++;
    return response({ title: String(url) });
  });
  await controller.open(null);
  await settle();
  const groupId = controller
    .getSnapshot()
    .groups.find((g) => g.segments.length === 2)?.id;
  assert(groupId);
  controller.select(key(2));
  await settle();
  const cached = controller.getSnapshot().documents[key(1)].document;
  controller.select(key(3));
  await settle();
  controller.selectSession(groupId);
  await settle();
  assert.equal(controller.getSnapshot().selectedId, key(1));
  assert.equal(controller.getSnapshot().documents[key(1)].document, cached);
  assert.equal(reads, 3);
  data = index([
    session(1, { archived: true }),
    session(2, { archived: true }),
  ]);
  await controller.refresh();
  assert.equal(controller.getSnapshot().groups.length, 0);
  assert.equal(controller.getSnapshot().selectedId, undefined);
  controller.setShowArchived(true);
  await settle();
  assert.equal(controller.getSnapshot().groups.length, 1);
});

test("deleting a pending sibling discards its response without replacing loaded resources", async () => {
  let data = index([session(1), session(2)]);
  const pending = deferred();
  const controller = setup(async (url, options) => {
    if (options?.method === "DELETE") {
      data = index([session(1)]);
      return new Response(null, { status: 204 });
    }
    if (String(url) === "/api/index") return response(data);
    if (String(url).endsWith(key(2))) return pending.promise;
    return response({ title: "stable" });
  });
  await controller.open(null);
  await settle();
  const stable = controller.getSnapshot().documents[key(1)].document;
  controller.select(key(2));
  await controller.remove(key(2));
  await settle();
  pending.resolve(response({ title: "deleted" }));
  await settle();
  assert.equal(controller.getSnapshot().selectedId, key(1));
  assert.equal(controller.getSnapshot().documents[key(1)].document, stable);
  assert.equal(controller.getSnapshot().documents[key(2)], undefined);
});

test("session names are searchable across every sibling and writes declare group scope", async () => {
  const { groupSessions, matchingSegments } = await import("../sessions");
  const group = groupSessions(
    mapIndex(
      index([session(1, { session_title: "Release review" }), session(2)]),
    ),
  )[0];
  assert.equal(group.title, "Release review");
  assert.equal(matchingSegments(group, "release review").length, 2);
  let headers: Headers | undefined;
  const client = new AccessClient(async (_url, init) => {
    headers = new Headers(init?.headers);
    return new Response(null, { status: 204 });
  });
  client.token = "test-access";
  await client.request(`/api/sessions/${key(1)}/archive`, "POST");
  assert.equal(headers?.get("X-Hive-Action-Scope"), "session");
  assert.equal(headers?.get("X-Hive-Token"), "test-access");
});

test("group archive reloads authority before selection and undo retains its original destination", async () => {
  let data = index([
    session(1),
    session(2),
    session(3, { session_id: "other" }),
  ]);
  const controller = setup(async (url, init) => {
    if (String(url).endsWith("/archive")) {
      data = index([
        session(1, { archived: init?.method === "POST" }),
        session(2, { archived: init?.method === "POST" }),
        session(3, { session_id: "other" }),
      ]);
      return new Response(null, { status: 204 });
    }
    return response(
      String(url) === "/api/index" ? data : { title: "Document" },
    );
  });
  await controller.open(null);
  controller.select(key(1));
  await controller.archive();
  assert.deepEqual(
    controller.getSnapshot().sessions.map((s) => s.id),
    [key(3)],
  );
  assert.equal(controller.getSnapshot().undoArchive?.id, key(1));
  await controller.undoArchive();
  assert.equal(controller.getSnapshot().sessions.length, 3);
  assert.equal(controller.getSnapshot().selectedId, key(1));
});

test("partial deletion keeps structured recovery state and read-only checks clear completed cleanup", async () => {
  let data = index([session(1), session(2)]);
  let deletes = 0;
  const controller = setup(async (url, init) => {
    if (init?.method === "DELETE") {
      deletes++;
      data = index([], {
        errors: [
          {
            path: "groups/pending",
            reason: "cleanup",
            code: "delete_cleanup_pending",
            session_group: "group",
          },
        ],
      });
      return response(
        {
          code: "partial_delete",
          deleted: 2,
          remaining: 0,
          cleanup_pending: true,
        },
        500,
      );
    }
    return response(
      String(url) === "/api/index" ? data : { title: "Document" },
    );
  });
  await controller.open(null);
  assert.equal(await controller.remove(), false);
  assert.equal(controller.getSnapshot().actionError?.remaining, 0);
  assert.equal(controller.getSnapshot().sessions.length, 0);
  data = index([]);
  await controller.refresh();
  assert.equal(controller.getSnapshot().actionError, undefined);
  assert.equal(deletes, 1);
});

test("an unanswered rename aborts within fifteen seconds and refreshes without replay", async () => {
  let data = index();
  let writes = 0;
  let reads = 0;
  let signal: AbortSignal | null | undefined;
  const controller = setup(async (url, init) => {
    if (init?.method === "PUT") {
      writes++;
      signal = init.signal;
      return new Promise<Response>((_resolve, reject) => {
        signal?.addEventListener(
          "abort",
          () => {
            data = index([
              session(1, { session_title: "Saved before connection ended" }),
            ]);
            reject(new DOMException("The request was aborted", "AbortError"));
          },
          { once: true },
        );
      });
    }
    if (String(url) === "/api/index") reads++;
    return response(
      String(url) === "/api/index" ? data : { title: "Document" },
    );
  });
  await controller.open(null);
  const saving = controller.rename("Saved before connection ended");
  assert.equal(controller.getSnapshot().writing, true);
  let deadline: ReturnType<typeof setTimeout> | undefined;
  try {
    const result = await Promise.race([
      saving,
      new Promise<string>((resolve) => {
        deadline = setTimeout(() => resolve("unresolved"), 16000);
      }),
    ]);
    assert.equal(
      result,
      false,
      "An unanswered rename must leave the pending state after its deadline",
    );
    await settle();
    assert.equal(signal?.aborted, true);
    assert.equal(controller.getSnapshot().writing, false);
    assert.equal(controller.getSnapshot().actionError?.code, "rename_timeout");
    assert.equal(
      controller.getSnapshot().groups[0].title,
      "Saved before connection ended",
    );
    assert.equal(writes, 1);
    assert.equal(reads, 2);
  } finally {
    clearTimeout(deadline);
  }
});
