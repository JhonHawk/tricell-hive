import assert from "node:assert/strict";
import { readFileSync } from "node:fs";
import test from "node:test";
import {
  decodeTimelineContract,
  isSafeTimelineArtifactURL,
  isSafeTimelineDocumentURL,
  isSafeTimelineURL,
  parseTimelineDocumentResponse,
  parseTimelineIndexEntry,
  parseTimelineJSON,
  parseTimelineOperationReceipt,
  parseTimelineOperationResult,
  parseTimelineRequest,
  parseTimelineSession,
  parseTimelineSessionResponse,
  type TimelineAppendEventRequest,
  TimelineContractError,
  type TimelineContractKind,
  type TimelineEvent,
  type TimelineSession,
  type TimelineView,
  timelineTicketIdentity,
  validateTimelineDocumentBytes,
  validateTimelineDocumentSnapshot,
  validateTimelineEventPatch,
  validateTimelineEventReferences,
  validateTimelineEventWithdrawal,
  validateTimelineOperationRetry,
  validateTimelineSessionPatch,
} from "../timeline-contract";

const fixture = (name: string) =>
  readFileSync(
    new URL(`../../../testdata/timeline_contract/${name}`, import.meta.url),
    "utf8",
  );
const session = () => parseTimelineSession(JSON.parse(fixture("session.json")));
function evidenceViews(event: TimelineEvent): TimelineView[] {
  assert.ok(event.views);
  return event.views;
}
function predecessor(view: TimelineView) {
  assert.ok(view.supersedes);
  return view.supersedes;
}
function requestViews(request: TimelineAppendEventRequest) {
  assert.ok(request.views);
  return request.views;
}
const cases = JSON.parse(fixture("cases.json")) as {
  name: string;
  kind: TimelineContractKind;
  valid: boolean;
  json: string;
}[];
for (const item of cases)
  test(`timeline shared contract: ${item.name}`, () => {
    if (item.valid)
      assert.doesNotThrow(() => decodeTimelineContract(item.kind, item.json));
    else
      assert.throws(
        () => decodeTimelineContract(item.kind, item.json),
        TimelineContractError,
      );
  });
test("timeline output validates shared separate-view and focused-iteration fixture", () => {
  const s = session();
  assert.equal(s.validation_count, 2);
  assert.equal(s.events.length, 6);
  assert.equal(s.state, "active");
  assert.equal(s.events[5].delivery_refs?.[0].snapshot.status, "open");
  assert.equal(s.events[2].views?.length, 2);
  assert.equal(s.events[2].views?.[1].ticket_refs.length, 2);
  assert.equal(s.events[3].views?.length, 1);
  assert.deepEqual(parseTimelineSession(JSON.parse(JSON.stringify(s))), s);
});
test("copy changes preserve immutable evidence identity and references after roundtrip", () => {
  const s = session();
  const original = s.events[2];
  const evidenceID = original.evidence_id;
  original.title = "Corrected title";
  original.version++;
  original.updated_at = "2026-10-31T19:00:00-06:00";
  const restored = parseTimelineSession(JSON.parse(JSON.stringify(s)));
  assert.equal(restored.events[2].evidence_id, evidenceID);
  assert.equal(
    restored.events[3].views?.[0].supersedes?.evidence_id,
    evidenceID,
  );
  assert.equal(restored.events[2].sequence, 3);
  original.withdrawn = true;
  s.validation_count--;
  assert.doesNotThrow(
    () => parseTimelineSession(s),
    "historical references survive withdrawal",
  );
});
const corruptions: [string, (s: TimelineSession) => void][] = [
  [
    "foreign artifact owner",
    (s) => {
      s.artifacts[0].session_id = "2".repeat(64);
    },
  ],
  [
    "unknown ticket",
    (s) => {
      evidenceViews(s.events[2])[0].ticket_refs = [
        { ticket_id: "unknown", snapshot: { key: "OTHER" } },
      ];
    },
  ],
  [
    "foreign image",
    (s) => {
      evidenceViews(s.events[2])[0].image_refs[0].artifact_id = "foreign_image";
    },
  ],
  [
    "wrong evidence",
    (s) => {
      predecessor(evidenceViews(s.events[3])[0]).evidence_id = "wrong_evidence";
    },
  ],
  [
    "wrong view",
    (s) => {
      predecessor(evidenceViews(s.events[3])[0]).view_id = "wrong_view";
    },
  ],
  [
    "future event target",
    (s) => {
      predecessor(evidenceViews(s.events[3])[0]).event_id = "event_approval";
    },
  ],
  [
    "duplicate event",
    (s) => {
      s.events[3].event_id = s.events[2].event_id;
    },
  ],
  [
    "duplicate evidence",
    (s) => {
      s.events[3].evidence_id = s.events[2].evidence_id;
    },
  ],
  [
    "out of order",
    (s) => {
      s.events[3].sequence = 2;
    },
  ],
  [
    "human origin",
    (s) => {
      (s.events[2] as { origin: string }).origin = "human";
    },
  ],
  [
    "count mismatch",
    (s) => {
      s.validation_count = 0;
    },
  ],
  [
    "oversize image",
    (s) => {
      s.artifacts[0].size = 21 << 20;
    },
  ],
];
for (const [name, change] of corruptions)
  test(`timeline rejects corrupt output: ${name}`, () => {
    const s = session();
    change(s);
    assert.throws(() => parseTimelineSession(s));
  });
test("runtime reference checking uses session membership, not matching names", () => {
  const s = session();
  const r = parseTimelineRequest(
    "append_event",
    JSON.parse(fixture("append-validation.json")),
  ) as TimelineAppendEventRequest;
  assert.doesNotThrow(() => validateTimelineEventReferences(s, r));
  requestViews(r)[0].image_refs[0].artifact_id = "foreign_image";
  assert.throws(() => validateTimelineEventReferences(s, r));
  requestViews(r)[0].image_refs[0].artifact_id = "artifact_chat";
  requestViews(r)[0].ticket_ids = ["unknown-ticket"];
  assert.throws(() => validateTimelineEventReferences(s, r));
  assert.notEqual(
    timelineTicketIdentity(s.tickets[0]),
    requestViews(r)[0].ticket_ids[0],
  );
});
test("artifact URLs are transport-only, bound to known session artifacts", () => {
  const s = session();
  const url = `/files/${s.viewer_session_id}/${"a".repeat(64)}/artifact_profile.png`;
  assert.equal(isSafeTimelineArtifactURL(url, s.viewer_session_id), true);
  assert.equal(isSafeTimelineArtifactURL(url, "2".repeat(64)), false);
  assert.equal(
    isSafeTimelineArtifactURL(
      url.replace("artifact_profile.png", "../outside"),
      s.viewer_session_id,
    ),
    false,
  );
  assert.equal(
    isSafeTimelineArtifactURL(
      "https://example.com/image.png",
      s.viewer_session_id,
    ),
    false,
  );
  assert.equal(
    isSafeTimelineArtifactURL(`${url}?token=value`, s.viewer_session_id),
    false,
  );
  assert.doesNotThrow(() =>
    parseTimelineSessionResponse({
      schema_version: 3,
      mode: "timeline_api",
      session: s,
      artifact_urls: { artifact_profile: url },
    }),
  );
  assert.throws(() =>
    parseTimelineSessionResponse({
      schema_version: 3,
      mode: "timeline_api",
      session: s,
      artifact_urls: { foreign: url },
    }),
  );
  assert.equal(Object.hasOwn(s.artifacts[0], "url"), false);
});
test("safe external URL helper rejects credentials, non-HTTPS and lexical normalization tricks", () => {
  for (const url of [
    "https://example.com/pull/1",
    "HTTPS://example.com/a%20b",
    "https://[::1]:443/pull/1",
  ])
    assert.equal(isSafeTimelineURL(url), true, url);
  for (const url of [
    "http://example.com/",
    "javascript:alert(1)",
    "https://user:secret@example.com/",
    "https://example.com:/",
    "https://example.com:65536/",
    "https://example.com/%zz",
    "https://example.com/\\path",
    "https://example.com/\npath",
  ])
    assert.equal(isSafeTimelineURL(url), false, url);
});
test("strict scanner rejects nested duplicate escaped keys, nulls, trailing data and deep input", () => {
  for (const raw of [
    '{"a":{"x":1,"\\u0078":2}}',
    '{"a":null}',
    "{} {}",
    `${"[".repeat(34)}1${"]".repeat(34)}`,
  ])
    assert.throws(() => parseTimelineJSON(raw), TimelineContractError);
  assert.throws(
    () => parseTimelineJSON(" ".repeat((2 << 20) + 1)),
    TimelineContractError,
  );
  const s = JSON.parse(fixture("session.json"));
  delete s.pinned;
  assert.throws(() => parseTimelineSession(s), TimelineContractError);
});

test("receipt retry contracts recover committed IDs and reject changed request digests", () => {
  const replay = JSON.parse(fixture("operation-replay.json"));
  const receipt = parseTimelineOperationReceipt(replay.receipt);
  const result = validateTimelineOperationRetry(
    receipt,
    replay.matching_retry.operation_id,
    replay.matching_retry.request_sha256,
  );
  assert.equal(result.kind, "append_event");
  if (result.kind !== "append_event")
    assert.fail("Expected event publication result");
  assert.equal(result.event_id, "event_validation");
  assert.equal(result.evidence_id, "evidence_initial");
  assert.throws(
    () =>
      validateTimelineOperationRetry(
        receipt,
        replay.conflicting_retry.operation_id,
        replay.conflicting_retry.request_sha256,
      ),
    TimelineContractError,
  );
  assert.equal(
    Object.hasOwn(parseTimelineOperationResult(result), "request_sha256"),
    false,
  );
  assert.equal(Object.hasOwn(session(), "operation_receipts"), false);
});
test("v2 index rejects legacy input", () => {
  assert.throws(
    () =>
      parseTimelineIndexEntry(JSON.parse(fixture("legacy-index-entry.json"))),
    TimelineContractError,
  );
  const api = parseTimelineIndexEntry(
    JSON.parse(fixture("api-index-entry.json")),
  );
  assert.equal(api.mode, "timeline_api");
  if (api.mode !== "timeline_api") assert.fail("Expected API index variant");
  assert.equal(api.validation_count, 2);
  assert.equal(api.source.native_session_id, "native-session-alpha");
});

const transitionCases = JSON.parse(fixture("transition-cases.json")) as {
  name: string;
  action: string;
  session: unknown;
  event_id: string;
  request: unknown;
  valid: boolean;
}[];
for (const c of transitionCases)
  test(`timeline shared transition: ${c.name}`, () => {
    const s = parseTimelineSession(c.session);
    const check = () => {
      switch (c.action) {
        case "event_patch":
          return validateTimelineEventPatch(
            s,
            c.event_id,
            parseTimelineRequest("event_patch", c.request),
          );
        case "session_patch":
          return validateTimelineSessionPatch(
            s,
            parseTimelineRequest("session_patch", c.request),
          );
        case "withdraw":
          return validateTimelineEventWithdrawal(
            s,
            c.event_id,
            parseTimelineRequest("mutation", c.request),
          );
        case "append":
          return validateTimelineEventReferences(
            s,
            parseTimelineRequest("append_event", c.request),
          );
        case "snapshot":
          return validateTimelineDocumentSnapshot(
            s,
            c.event_id,
            parseTimelineRequest("document_snapshot", c.request),
          );
        default:
          throw Error("unknown scenario");
      }
    };
    if (c.valid) assert.doesNotThrow(check);
    else assert.throws(check, TimelineContractError);
  });
test("document byte bounds and immutable readback", async () => {
  validateTimelineDocumentBytes(new Uint8Array(262144));
  validateTimelineDocumentBytes(new Uint8Array());
  assert.throws(
    () => validateTimelineDocumentBytes(new Uint8Array(262145)),
    TimelineContractError,
  );
  assert.throws(
    () => validateTimelineDocumentBytes(new Uint8Array([255])),
    TimelineContractError,
  );
  const response = JSON.parse(fixture("document-response.json"));
  await assert.doesNotReject(() => parseTimelineDocumentResponse(response));
  response.content += "modified";
  await assert.rejects(
    () => parseTimelineDocumentResponse(response),
    TimelineContractError,
  );
});
test("document route map accepts only scoped known immutable IDs", () => {
  const s = parseTimelineSession(JSON.parse(fixture("stage-session.json")));
  const path = `/api/timeline/sessions/${s.viewer_session_id}/documents/doc-plan`;
  assert.ok(isSafeTimelineDocumentURL(path, s.viewer_session_id, "doc-plan"));
  const response = {
    schema_version: 3,
    mode: "timeline_api",
    session: s,
    document_urls: { "doc-plan": path },
  };
  assert.doesNotThrow(() => parseTimelineSessionResponse(response));
  for (const url of [
    `${path}?token=private`,
    `https://example.com${path}`,
    `${path}/../other`,
  ]) {
    response.document_urls["doc-plan"] = url;
    assert.throws(
      () => parseTimelineSessionResponse(response),
      TimelineContractError,
    );
  }
  assert.throws(
    () =>
      parseTimelineSessionResponse({
        ...response,
        document_urls: { unknown: path },
      }),
    TimelineContractError,
  );
});

test("shared document response preserves UTF8 BOM bytes, size and digest", async () => {
  const fixtureResponse = JSON.parse(fixture("document-response-bom.json"));
  const response = await parseTimelineDocumentResponse(fixtureResponse);
  assert.equal(response.content.charCodeAt(0), 0xfeff);
  assert.equal(response.document.size, 32);
  assert.equal(
    new TextEncoder().encode(response.content).length,
    response.document.size,
  );
  assert.equal(
    response.document.sha256,
    "d58e2f6d5d846c4960387b9afdac5ec7aba37d6785782cd7d6f9d057990eeb43",
  );
  assert.equal(response.content, fixtureResponse.content);
  for (const content of ["\ud800", "\udfff"]) {
    await assert.rejects(
      () => parseTimelineDocumentResponse({ ...fixtureResponse, content }),
      TimelineContractError,
    );
  }
});
