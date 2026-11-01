import assert from "node:assert/strict";
import test from "node:test";
import { AccessClient, takeCode, validateIndex } from "../api";

const response = (data: unknown) => new Response(JSON.stringify(data));
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

test("v2 cutoff rejects old or mixed indices without interpreting legacy content", () => {
  const valid = {
    protocol_version: 2,
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
