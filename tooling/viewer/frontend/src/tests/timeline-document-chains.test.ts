import assert from "node:assert/strict";
import test from "node:test";
import { timelineFixture } from "../dev/timeline-fixture";
import type { TimelineDocumentSnapshot } from "../timeline-contract";
import {
  compareRevisionTabs,
  documentChains,
  latestSuccessor,
  revisionTabs,
} from "../timeline-document-chains";

const actor = {
  kind: "native_producer" as const,
  producer_id: "producer-original",
};
const revision = (
  id: string,
  key: string,
  tabs: TimelineDocumentSnapshot["tabs"],
  predecessor?: string,
): TimelineDocumentSnapshot => ({
  actor,
  document_revision_id: id,
  document_key: key,
  status: "complete",
  created_at: "2026-11-01T00:00:00Z",
  tabs,
  ...(predecessor ? { predecessor_revision_id: predecessor } : {}),
});
test("document keys separate identical labels and successor notices follow only explicit chain links", () => {
  const s = structuredClone(timelineFixture.session);
  const stage = s.events.find((e) => e.stage === "plan");
  assert.ok(stage);
  s.events = [
    {
      ...stage,
      event_id: "first",
      document_snapshots: [
        revision("a1", "alpha", [{ label: "Design", document_id: "one" }]),
        revision("b1", "beta", [{ label: "Design", document_id: "two" }]),
      ],
    },
    {
      ...stage,
      event_id: "continued",
      related_event_id: "first",
      sequence: stage.sequence + 1,
      document_snapshots: [
        revision(
          "a2",
          "alpha",
          [{ label: "Design", document_id: "three" }],
          "a1",
        ),
      ],
    },
  ];
  const chains = documentChains(s);
  assert.equal(chains.length, 2);
  assert.equal(chains[0].revisions.length, 2);
  assert.equal(chains[1].revisions.length, 1);
  assert.equal(
    latestSuccessor(chains[0], "a1")?.revision.document_revision_id,
    "a2",
  );
  assert.equal(latestSuccessor(chains[1], "b1"), undefined);
});
test("revision comparison is label/document identity; only structured plans group deltas", () => {
  const before = revision("one", "plan", [
    { label: "Proposal", document_id: "p" },
    { label: "Design", document_id: "d1" },
    { label: "Tasks", document_id: "t" },
    { label: "Notes", document_id: "n" },
  ]);
  const after = revision(
    "two",
    "plan",
    [
      { label: "Proposal", document_id: "p" },
      { label: "Design", document_id: "d2" },
      { label: "Tasks", document_id: "t" },
      { label: "specs/viewer", document_id: "s" },
    ],
    "one",
  );
  assert.deepEqual(compareRevisionTabs(after, before), {
    initial: false,
    added: ["specs/viewer"],
    changed: ["Design"],
    removed: ["Notes"],
  });
  assert.equal(compareRevisionTabs(before).initial, true);
  assert.equal(revisionTabs(after, "plan").at(-1)?.kind, "specs");
  assert.equal(revisionTabs(after, "build").length, 4);
  assert.equal(
    revisionTabs(
      revision("generic", "result", [
        { label: "Risks", document_id: "r" },
        { label: "Result", document_id: "x" },
      ]),
      "plan",
    ).length,
    2,
  );
});
