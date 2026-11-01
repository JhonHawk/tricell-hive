import assert from "node:assert/strict";
import test from "node:test";
import {
  AccessClient,
  ARCHIVED_KEY,
  mapTimelineIndex,
  OPEN_KEY,
  type RawIndex,
  takeCode,
  takeTarget,
  validateIndex,
} from "../api";
import { ViewerController } from "../controller";
import { loadPreviewDocument } from "../dev/timeline-documents-fixture";
import { timelineFixture } from "../dev/timeline-fixture";
import type {
  TimelineAPIIndexEntry,
  TimelineSessionResponse,
} from "../timeline-contract";

const sid = "a".repeat(64),
  other = "b".repeat(64);
const stamp = "2026-10-31T21:00:00Z";
const response = (data: unknown, status = 200) =>
  new Response(JSON.stringify(data), { status });
async function settle() {
  for (let i = 0; i < 60; i++) await Promise.resolve();
}
function fixture(id = sid): TimelineSessionResponse {
  return {
    schema_version: 2,
    mode: "timeline_api",
    session: {
      schema_version: 2,
      producers: [
        {
          producer_id: "producer-original",
          source: {
            project_root: "/fixture/project",
            host: "codex",
            native_session_id: `native-${id}`,
          },
          joined_at: stamp,
        },
      ],
      foreground_producer_id: "producer-original",
      pending_join_requests: [],
      tasks: [],
      waits: [],
      viewer_session_id: id,
      project_key: "example.test/project",
      project_name: "Example",
      source: {
        project_root: "/fixture/project",
        host: "codex",
        native_session_id: `native-${id}`,
      },
      version: 1,
      title: `API ${id[0]}`,
      summary: "Read-only example",
      state: "active",
      archived: false,
      pinned: false,
      created_at: stamp,
      updated_at: stamp,
      validation_count: 0,
      tickets: [],
      events: [
        {
          actor: { kind: "native_producer", producer_id: "producer-original" },
          event_id: "request_one",
          sequence: 1,
          version: 1,
          kind: "request",
          title: "Original request",
          summary: "Keep API data separate",
          created_at: stamp,
          updated_at: stamp,
          origin: "agent_reported",
          ticket_refs: [],
        },
      ],
      artifacts: [],
      documents: [],
      current_stage_event_id: null,
    },
  };
}
function entry(data: TimelineSessionResponse): TimelineAPIIndexEntry {
  const {
    schema_version: _schema,
    events: _events,
    artifacts: _artifacts,
    documents: _documents,
    producers: _producers,
    foreground_producer_id: _lead,
    pending_join_requests: _joins,
    tasks: _tasks,
    waits: _waits,
    ...rest
  } = data.session;
  return { mode: "timeline_api", ...rest };
}
function index(...data: TimelineSessionResponse[]): RawIndex {
  return {
    projects: [],
    repositories: [],
    errors: [],
    ttl_seconds: 0,
    timeline_sessions: data.map(entry),
  };
}
function setup(fetcher: typeof fetch, remembered = "") {
  const writes: [string, string | undefined][] = [];
  const client = new AccessClient(fetcher, {
    getItem: (key) =>
      key === OPEN_KEY
        ? remembered
        : key === ARCHIVED_KEY
          ? "false"
          : "test-token",
    setItem: (key, value) => writes.push([key, value]),
    removeItem: (key) => writes.push([key, undefined]),
  });
  return {
    controller: new ViewerController(client, () => false),
    client,
    writes,
  };
}

test("opening target validates opaque IDs, takes precedence, and remains in the URL after secret removal", async () => {
  const target = takeTarget(`?session_id=${sid}&event_id=request_one`);
  assert.deepEqual(target, { sessionID: sid, eventID: "request_one" });
  for (const query of [
    "?event_id=request_one",
    "?session_id=bad",
    `?session_id=${sid}&event_id=../bad`,
    `?session_id=${sid}&session_id=${other}`,
  ])
    assert.deepEqual(takeTarget(query), { invalid: true });
  let replaced = "";
  assert.equal(
    takeCode(
      {
        hash: "#code=one-use",
        pathname: "/",
        search: `?session_id=${sid}&event_id=request_one`,
      },
      {
        replaceState: (_a, _b, url) => {
          replaced = String(url);
        },
      },
    ),
    "one-use",
  );
  assert.equal(replaced, `/?session_id=${sid}&event_id=request_one`);
  const a = fixture(),
    b = fixture(other);
  const { controller } = setup(
    async (path) =>
      response(
        path === "/api/index"
          ? index(a, b)
          : String(path).endsWith(sid)
            ? a
            : b,
      ),
    other,
  );
  await controller.open(null, target);
  await settle();
  assert.equal(controller.getSnapshot().selectedId, sid);
  assert.equal(
    controller.getSnapshot().timeline?.data?.session.viewer_session_id,
    sid,
  );
  const mapped = mapTimelineIndex(entry(a));
  assert.equal("raw" in mapped, false);
  assert.equal("fileBase" in mapped, false);
  assert.throws(() =>
    validateIndex({
      ...index(a),
      timeline_sessions: [
        { ...entry(a), fileBase: "/invented" },
      ] as unknown as TimelineAPIIndexEntry[],
    }),
  );
});

test("missing and invalid destinations do not select another session; explicit archived target does not change preference", async () => {
  const a = fixture();
  a.session.archived = true;
  const b = fixture(other);
  const { controller, writes } = setup(
    async (path) =>
      response(
        path === "/api/index"
          ? index(a, b)
          : String(path).endsWith(sid)
            ? a
            : b,
      ),
    other,
  );
  await controller.open(null, { sessionID: sid });
  await settle();
  assert.equal(controller.getSnapshot().selectedId, sid);
  assert.equal(controller.getSnapshot().showArchived, false);
  assert.equal(
    writes.some(([key]) => key === ARCHIVED_KEY),
    false,
  );
  await controller.open(null, { sessionID: "c".repeat(64) });
  assert.equal(controller.getSnapshot().selectedId, undefined);
  assert.equal(controller.getSnapshot().targetError, true);
  controller.select(other);
  await settle();
  assert.equal(controller.getSnapshot().selectedId, other);
  assert.equal(controller.getSnapshot().targetError, false);
  await controller.open(null, { invalid: true });
  assert.equal(controller.getSnapshot().selectedId, undefined);
  assert.equal(controller.getSnapshot().targetError, true);
  await controller.open(null, { sessionID: other, eventID: "missing_event" });
  await settle();
  assert.equal(controller.getSnapshot().selectedId, other);
  assert.equal(controller.getSnapshot().targetError, true);
});

test("late API session reads cannot replace a newer selection; polling failures retain the last readable API data", async () => {
  const a = fixture(),
    b = fixture(other);
  let release: ((r: Response) => void) | undefined;
  let offline = false;
  const { controller } = setup(async (path) => {
    if (offline) throw new TypeError("offline");
    if (path === "/api/index") return response(index(a, b));
    if (String(path).endsWith(sid))
      return new Promise<Response>((resolve) => {
        release = resolve;
      });
    return response(b);
  });
  await controller.open(null, { sessionID: sid });
  controller.select(other);
  await settle();
  assert.equal(
    controller.getSnapshot().timeline?.data?.session.viewer_session_id,
    other,
  );
  assert.ok(release);
  release(response(a));
  await settle();
  assert.equal(
    controller.getSnapshot().timeline?.data?.session.viewer_session_id,
    other,
  );
  offline = true;
  await controller.refresh();
  assert.equal(controller.getSnapshot().down, true);
  assert.equal(
    controller.getSnapshot().timeline?.data?.session.viewer_session_id,
    other,
  );
  offline = false;
  controller.retryTimeline();
  await settle();
  assert.equal(controller.getSnapshot().down, false);
  assert.equal(controller.getSnapshot().indexError, false);
});

test("conditional API mutations read JSON receipts, refresh authority and reuse the same operation after a lost response", async () => {
  const a = fixture();
  const operations: string[] = [];
  let lost = true;
  let receipt: unknown;
  const headers: Headers[] = [];
  const { controller } = setup(async (path, options) => {
    headers.push(new Headers(options?.headers));
    if (path === "/api/index") return response(index(a));
    if (options?.method === "PATCH") {
      const body = JSON.parse(String(options.body));
      operations.push(body.operation_id);
      if (!receipt) {
        a.session.version++;
        a.session.title = body.title;
        receipt = {
          schema_version: 2,
          kind: "session_patch",
          operation_id: body.operation_id,
          viewer_session_id: sid,
          session_version: a.session.version,
          current_stage_event_id: null,
        };
      }
      if (lost) {
        lost = false;
        throw new TypeError("lost response");
      }
      return response(receipt);
    }
    return response(a);
  });
  await controller.open(null, { sessionID: sid });
  await settle();
  await assert.rejects(
    controller.updateTimelineSession(sid, { title: "Saved title" }, 1),
  );
  await settle();
  await controller.updateTimelineSession(sid, { title: "Saved title" }, 1);
  assert.equal(operations.length, 2);
  assert.equal(operations[0], operations[1]);
  assert.equal(a.session.version, 2);
  assert.equal(
    controller.getSnapshot().timeline?.data?.session.title,
    "Saved title",
  );
  assert.ok(headers.every((h) => h.get("X-Hive-Token") === "test-token"));
  assert.ok(headers.some((h) => h.get("X-Hive-Action-Scope") === "session"));
});

test("a stale API write refreshes without replay or silently overwriting newer data", async () => {
  const a = fixture();
  let patches = 0;
  const { controller } = setup(async (path, options) => {
    if (path === "/api/index") return response(index(a));
    if (options?.method === "PATCH") {
      patches++;
      a.session.version = 2;
      a.session.title = "Newer title";
      return response({ code: "version_conflict" }, 409);
    }
    return response(a);
  });
  await controller.open(null, { sessionID: sid });
  await settle();
  await assert.rejects(
    controller.updateTimelineSession(sid, { title: "Stale draft" }, 1),
  );
  await settle();
  assert.equal(patches, 1);
  assert.equal(
    controller.getSnapshot().timeline?.data?.session.title,
    "Newer title",
  );
});

test("document reads use authenticated JSON and async hash validation; forbidden reads revoke all protected state", async () => {
  const data = structuredClone(timelineFixture);
  let corrupt = false,
    forbidden = false;
  const paths: string[] = [];
  const { controller, client } = setup(async (path) => {
    paths.push(String(path));
    if (path === "/api/index") return response(index(data));
    if (String(path).includes("/documents/")) {
      if (forbidden) return response({}, 403);
      const result = await loadPreviewDocument("document_proposal");
      return response(
        corrupt ? { ...result, content: "altered bytes" } : result,
      );
    }
    return response(data);
  });
  await controller.open(null, { sessionID: sid });
  await settle();
  const read = await controller.loadTimelineDocument(sid, "document_proposal");
  assert.ok(read.content.includes("Propuesta"));
  corrupt = true;
  await assert.rejects(
    controller.loadTimelineDocument(sid, "document_proposal"),
  );
  assert.equal(controller.getSnapshot().access, "ready");
  forbidden = true;
  await assert.rejects(
    controller.loadTimelineDocument(sid, "document_proposal"),
  );
  assert.equal(controller.getSnapshot().access, "rejected");
  assert.equal(controller.getSnapshot().timeline, null);
  assert.equal(client.token, "");
  assert.ok(paths.every((path) => !path.includes("test-token")));
});

test("API pin, archive, restore and delete use versioned session mutations and preserve selection policy", async () => {
  const a = fixture(),
    b = fixture(other);
  let deleted = false;
  const { controller } = setup(async (path, options) => {
    if (path === "/api/index")
      return response(index(...(deleted ? [b] : [a, b])));
    if (options?.method === "PATCH" || options?.method === "DELETE") {
      const body = JSON.parse(String(options.body));
      assert.equal(body.expected_version, a.session.version);
      a.session.version++;
      if (options.method === "DELETE") deleted = true;
      else {
        if (body.pinned !== undefined) a.session.pinned = body.pinned;
        if (body.archived !== undefined) a.session.archived = body.archived;
      }
      return response({
        schema_version: 2,
        kind: options.method === "DELETE" ? "session_delete" : "session_patch",
        operation_id: body.operation_id,
        viewer_session_id: sid,
        session_version: a.session.version,
        current_stage_event_id: null,
        ...(deleted ? { deleted: true } : {}),
      });
    }
    return response(String(path).endsWith(sid) ? a : b);
  });
  await controller.open(null, { sessionID: sid });
  await settle();
  await controller.updateTimelineSession(sid, { pinned: true }, 1);
  assert.equal(controller.getSnapshot().timeline?.data?.session.pinned, true);
  await controller.updateTimelineSession(sid, { archived: true }, 2);
  await settle();
  assert.equal(controller.getSnapshot().selectedId, other);
  assert.equal(controller.getSnapshot().showArchived, false);
  controller.setShowArchived(true);
  controller.select(sid);
  await settle();
  await controller.updateTimelineSession(sid, { archived: false }, 3);
  assert.equal(
    controller.getSnapshot().timeline?.data?.session.archived,
    false,
  );
  await controller.deleteTimelineSession(sid, 4);
  await settle();
  assert.equal(controller.getSnapshot().selectedId, other);
  assert.equal(controller.getSnapshot().notice, "deleted");
  assert.equal(
    controller
      .getSnapshot()
      .timelineSessions.some((s) => s.viewer_session_id === sid),
    false,
  );
});

test("unchanged polls reuse an outstanding API read while a newer version supersedes it", async () => {
  const a = fixture();
  const pending: ((value: Response) => void)[] = [];
  let reads = 0;
  const { controller } = setup(async (path) => {
    if (path === "/api/index") return response(index(a));
    reads++;
    return reads === 1
      ? response(a)
      : new Promise<Response>((resolve) => pending.push(resolve));
  });
  await controller.open(null, { sessionID: sid });
  await settle();
  a.session.version = 2;
  await controller.refresh();
  await controller.refresh();
  assert.equal(reads, 2);
  const older = structuredClone(a);
  a.session.version = 3;
  await controller.refresh();
  assert.equal(reads, 3);
  pending[1](response(a));
  await settle();
  assert.equal(controller.getSnapshot().timeline?.data?.session.version, 3);
  pending[0](response(older));
  await settle();
  assert.equal(controller.getSnapshot().timeline?.data?.session.version, 3);
});

test("switching to API invalidates pending legacy reads without leaving a stuck loading cache on return", async () => {
  const a = fixture(),
    legacyID = "1".repeat(64);
  const pending: ((r: Response) => void)[] = [];
  const mixed = index(a);
  mixed.projects = [
    {
      key: "legacy",
      name: "Legacy",
      tickets: [
        {
          ticket: "#1",
          sessions: [
            {
              key: legacyID,
              title: "Legacy",
              session_id: "legacy",
              source: "/fixture",
              source_status: "local",
              content_hash: "hash",
              created_at: stamp,
              ingested_at: stamp,
              fileBase: `/files/${legacyID}/${"c".repeat(64)}/`,
            },
          ],
        },
      ],
    },
  ];
  const { controller } = setup(async (path) => {
    if (path === "/api/index") return response(mixed);
    if (path === `/api/sessions/${legacyID}`)
      return new Promise<Response>((resolve) => pending.push(resolve));
    return response(a);
  });
  await controller.open(null);
  assert.equal(pending.length, 1);
  controller.select(sid);
  await settle();
  controller.select(legacyID);
  assert.equal(pending.length, 2);
  pending[0](response({ title: "Obsolete legacy" }));
  await settle();
  assert.equal(controller.getSnapshot().documents[legacyID]?.state, "loading");
  pending[1](response({ title: "Current legacy" }));
  await settle();
  assert.equal(controller.getSnapshot().documents[legacyID]?.state, "ready");
  assert.equal(
    controller.getSnapshot().documents[legacyID]?.document?.manifest.title,
    "Current legacy",
  );
});

test("pending deletion remains retryable with its original operation after tombstoning removes the index entry", async () => {
  const a = fixture(),
    b = fixture(other);
  let tombstoned = false;
  let attempts = 0;
  const bodies: unknown[] = [];
  const { controller } = setup(async (path, options) => {
    if (path === "/api/index")
      return response(index(...(tombstoned ? [b] : [a, b])));
    if (options?.method === "DELETE") {
      const body = JSON.parse(String(options.body));
      bodies.push(body);
      tombstoned = true;
      attempts++;
      if (attempts < 3)
        return response(
          { code: "delete_pending", cleanup_pending: true },
          attempts === 1 ? 503 : 409,
        );
      return response({
        schema_version: 2,
        kind: "session_delete",
        operation_id: body.operation_id,
        viewer_session_id: sid,
        session_version: 2,
        current_stage_event_id: null,
        deleted: true,
      });
    }
    return response(String(path).endsWith(sid) ? a : b);
  });
  await controller.open(null, { sessionID: sid });
  await settle();
  await assert.rejects(controller.deleteTimelineSession(sid, 1));
  assert.equal(controller.getSnapshot().selectedId, undefined);
  assert.deepEqual(controller.getSnapshot().pendingTimelineDelete, {
    id: sid,
    title: a.session.title,
    expectedVersion: 1,
  });
  await assert.rejects(controller.retryTimelineDelete());
  assert.equal(controller.getSnapshot().pendingTimelineDelete?.id, sid);
  await controller.retryTimelineDelete();
  assert.deepEqual(bodies[0], bodies[1]);
  assert.deepEqual(bodies[0], bodies[2]);
  assert.equal(controller.getSnapshot().pendingTimelineDelete, undefined);
  assert.equal(controller.getSnapshot().notice, "deleted");
});

test("pending deletion recovery is cleared when access changes", async () => {
  const a = fixture();
  let tombstoned = false;
  let deletes = 0;
  const { controller } = setup(async (path, options) => {
    if (path === "/api/index")
      return response(index(...(tombstoned ? [] : [a])));
    if (options?.method === "DELETE") {
      deletes++;
      tombstoned = true;
      return response({ code: "delete_pending" }, 503);
    }
    return response(a);
  });
  await controller.open(null, { sessionID: sid });
  await settle();
  await assert.rejects(controller.deleteTimelineSession(sid, 1));
  assert.ok(controller.getSnapshot().pendingTimelineDelete);
  await controller.open(null);
  await settle();
  assert.equal(controller.getSnapshot().pendingTimelineDelete, undefined);
  await controller.retryTimelineDelete();
  assert.equal(deletes, 1);
});

test("definitive delete rejection clears recovery and permits deleting another session", async (t) => {
  for (const status of [400, 404, 410]) {
    await t.test(`HTTP ${status}`, async () => {
      const a = fixture(),
        b = fixture(other);
      let aGone = false,
        bGone = false,
        firstDelete = true;
      const deletedIDs: string[] = [];
      const { controller } = setup(async (path, options) => {
        if (path === "/api/index")
          return response(index(...(aGone ? [] : [a]), ...(bGone ? [] : [b])));
        if (options?.method === "DELETE") {
          const id = String(path).split("/").at(-1) as string;
          deletedIDs.push(id);
          if (id === sid) {
            if (firstDelete) {
              firstDelete = false;
              return response({ code: "delete_pending" }, 503);
            }
            aGone = true;
            return response(
              { code: status === 410 ? "session_deleted" : "invalid_request" },
              status,
            );
          }
          bGone = true;
          const body = JSON.parse(String(options.body));
          return response({
            schema_version: 2,
            kind: "session_delete",
            operation_id: body.operation_id,
            viewer_session_id: other,
            session_version: 2,
            current_stage_event_id: null,
            deleted: true,
          });
        }
        return response(String(path).endsWith(sid) ? a : b);
      });
      await controller.open(null);
      await settle();
      await assert.rejects(controller.deleteTimelineSession(sid, 1));
      assert.ok(controller.getSnapshot().pendingTimelineDelete);
      await assert.rejects(controller.retryTimelineDelete());
      assert.equal(controller.getSnapshot().pendingTimelineDelete, undefined);
      assert.equal(
        controller
          .getSnapshot()
          .index?.timeline_sessions?.some((s) => s.viewer_session_id === sid),
        false,
      );
      await controller.deleteTimelineSession(other, 1);
      assert.deepEqual(deletedIDs, [sid, sid, other]);
      assert.equal(controller.getSnapshot().notice, "deleted");
    });
  }
});
