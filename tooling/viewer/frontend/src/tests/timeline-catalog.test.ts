import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import {
  canonicalTimelineIdentityURL,
  captureTimelineTicketReferences,
  parseTimelineCatalogRequest,
  parseTimelineJSON,
  parseTimelineOperationResult,
  parseTimelineRequest,
  parseTimelineSession,
  type TimelineCatalogRequestMap,
  TimelineContractError,
  type TimelineV3RequestMap,
  validateTimelineCatalogReferences,
  validateTimelineEventReferences,
  validateTimelineV3OperationReferences,
} from "../timeline-contract";

const fixture = (name: string) =>
  readFileSync(
    new URL(`../../../testdata/timeline_contract/${name}`, import.meta.url),
    "utf8",
  );
const session = () =>
  parseTimelineSession(JSON.parse(fixture("catalog-session.json")));
type Case = {
  name: string;
  action: string;
  kind: keyof TimelineCatalogRequestMap;
  object_id: string;
  code?: string;
  valid: boolean;
  json: string;
};
const cases: Case[] = JSON.parse(fixture("catalog-cases.json"));
for (const c of cases)
  test(`canonical catalog: ${c.name}`, () => {
    const run = () => {
      const value = parseTimelineJSON(c.json);
      switch (c.action) {
        case "shape":
          return parseTimelineCatalogRequest(c.kind, value);
        case "references":
          return validateTimelineCatalogReferences(
            session(),
            c.kind,
            c.object_id,
            value,
          );
        case "result":
          return parseTimelineOperationResult(value);
        case "event":
          return validateTimelineEventReferences(
            session(),
            parseTimelineRequest("append_event", value),
          );
        case "operation_references":
          return validateTimelineV3OperationReferences(
            session(),
            c.kind as keyof TimelineV3RequestMap,
            c.object_id,
            value,
          );
        case "session":
          return parseTimelineSession(value);
        default:
          assert.fail("unsupported corpus action");
      }
    };
    if (c.valid) assert.doesNotThrow(run);
    else
      assert.throws(
        run,
        (err: unknown) =>
          err instanceof TimelineContractError &&
          (!c.code || err.message.endsWith(`: ${c.code}`)),
      );
  });
test("captured reference remains immutable after canonical metadata changes", () => {
  const s = session(),
    refs = captureTimelineTicketReferences(s, ["catalog-ticket"]),
    snapshot = structuredClone(refs);
  const ticket = s.tickets.find((t) => t.ticket_id === "catalog-ticket");
  assert.ok(ticket);
  ticket.key = "Corrected";
  assert.deepEqual(refs, snapshot);
  assert.equal(
    captureTimelineTicketReferences(s, ["catalog-ticket"])[0].snapshot.key,
    "Corrected",
  );
  for (const ids of [
    ["catalog-ticket", "catalog-ticket"],
    ["candidate-1"],
    ["foreign"],
  ])
    assert.throws(
      () => captureTimelineTicketReferences(s, ids),
      TimelineContractError,
    );
});
test("natural identity keeps path case and rejects normalizing away dot segments", () => {
  assert.equal(
    canonicalTimelineIdentityURL("HTTPS://EXAMPLE.test:443/Tracker/"),
    "https://example.test/Tracker",
  );
  assert.equal(
    canonicalTimelineIdentityURL("https://example.test/a%2Fb/"),
    "https://example.test/a%2Fb",
  );
  assert.equal(
    canonicalTimelineIdentityURL("https://[::1]:443/repository/"),
    "https://[::1]/repository",
  );
  for (const url of [
    "https://example.test/a/../b",
    "https://example.test/a/%2e%2e/b",
    "https://example.test/path#",
    "https://example.test/path?",
  ])
    assert.throws(
      () => canonicalTimelineIdentityURL(url),
      TimelineContractError,
    );
});
