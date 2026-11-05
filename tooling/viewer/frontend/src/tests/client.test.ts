import assert from "node:assert/strict";
import test from "node:test";
import { AccessClient, takeCode, validateIndex } from "../api";

const response = (data: unknown) =>
  new Response(JSON.stringify(data), {
    headers: { "X-Hive-Timeline-Revision": "3" },
  });
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

test("the index validator accepts only protocol 3 with manual retention", () => {
  const valid = {
    protocol_version: 3,
    retention: "manual",
    ttl_seconds: 0,
    timeline_sessions: [],
    repositories: [],
    errors: [],
  };
  assert.doesNotThrow(() => validateIndex(valid as never));
  for (const value of [
    { ...valid, protocol_version: 1 },
    { ...valid, projects: [] },
    { ...valid, excluded_sessions: [] },
    { ...valid, timeline_sessions: undefined },
    { ...valid, ttl_seconds: 86400 },
  ])
    assert.throws(() => validateIndex(value as never));
});

test("direct writes require server revision acknowledgment and preserve the request", async () => {
  const methods: string[] = [];
  const access = new AccessClient(async (_path, options) => {
    methods.push(options?.method ?? "GET");
    return new Response("{}");
  });
  const body = {
    schema_version: 3,
    operation_id: "original-operation",
    expected_version: 1,
  };
  const before = JSON.stringify(body);
  await assert.rejects(
    access.request("/api/timeline/sessions/example", "PATCH", body),
    (error: unknown) =>
      typeof error === "object" &&
      error !== null &&
      "details" in error &&
      (error.details as { code: string }).code === "reload_required",
  );
  assert.deepEqual(methods, ["GET"]);
  assert.equal(JSON.stringify(body), before);
});
