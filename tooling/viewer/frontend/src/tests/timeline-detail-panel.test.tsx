import { timelineCandidateByID } from "../timeline-catalog";
import "./dom-setup.mjs";
import assert from "node:assert/strict";
import test, { type TestContext } from "node:test";
import { act } from "react";
import { createRoot } from "react-dom/client";
import { loadPreviewDocument } from "../dev/timeline-documents-fixture";
import { timelineFixture } from "../dev/timeline-fixture";
import { workFixture } from "../dev/timeline-work-fixture";
import type { TimelineSessionResponse } from "../timeline-contract";
import { aggregateTimelineCIWait } from "../timeline-contract-v3";
import {
  currentWaitObservations,
  useTimelinePanel,
} from "../timeline-detail-panel";

const settle = async () => {
  for (let i = 0; i < 50; i++) await Promise.resolve();
};
function region(selector: string, within: ParentNode = document) {
  const node = within.querySelector<HTMLElement>(selector);
  assert.ok(node, selector);
  return node;
}
function button(label: string, within: ParentNode = document) {
  const node = [...within.querySelectorAll<HTMLButtonElement>("button")].find(
    (b) =>
      b.getAttribute("aria-label") === label || b.textContent?.trim() === label,
  );
  assert.ok(node, label);
  return node;
}
async function click(label: string, within: ParentNode = document) {
  await act(async () => {
    button(label, within).click();
    await settle();
  });
}
function Harness({ data }: { data: TimelineSessionResponse }) {
  const panel = useTimelinePanel({
    data,
    lang: "en",
    loadDocument: loadPreviewDocument,
    onRetry: () => {},
  });
  return (
    <>
      <button
        type="button"
        onClick={(e) =>
          panel.openDetail(
            {
              kind: "task",
              taskID: "task-2",
              segmentID: "build-2",
              unblock: true,
            },
            e.currentTarget,
          )
        }
      >
        Open blocked task
      </button>
      <button
        type="button"
        onClick={(e) =>
          panel.openDetail(
            { kind: "task", taskID: "task-1", segmentID: "build-1" },
            e.currentTarget,
          )
        }
      >
        Open empty task
      </button>
      <button
        type="button"
        onClick={(e) =>
          panel.openDetail(
            { kind: "task", taskID: "task-7", segmentID: "verify-7" },
            e.currentTarget,
          )
        }
      >
        Open verify segment
      </button>
      <button
        type="button"
        onClick={(e) =>
          panel.openDetail({ kind: "wait", waitID: "ci-wait" }, e.currentTarget)
        }
      >
        Open CI
      </button>
      {panel.sheet}
    </>
  );
}
function mount(t: TestContext) {
  const container = document.createElement("div");
  document.body.append(container);
  const root = createRoot(container);
  t.after(async () => {
    await act(async () => root.unmount());
    container.remove();
  });
  return { root, container };
}
function taskFixture() {
  const data = workFixture("volume");
  const validation = timelineFixture.session.events.find(
    (e) => e.views?.length,
  );
  assert.ok(validation);
  data.session.events.push({
    ...structuredClone(validation),
    event_id: "task-evidence",
    task_id: "task-2",
    segment_id: "build-2",
    stage_event_id: "build-stage",
    sequence: 3,
    title: "Original linked evidence",
    summary: "Original task evidence remains readable.",
  });
  const stage = data.session.events.find((e) => e.event_id === "build-stage");
  assert.ok(stage);
  const source = timelineFixture.session.events
    .find((e) => e.event_id === "event_plan_r3")
    ?.document_snapshots?.at(-1);
  assert.ok(source);
  stage.document_snapshots = [
    {
      ...structuredClone(source),
      document_key: "implementation",
      document_revision_id: "task-document-v1",
      predecessor_revision_id: undefined,
      task_id: "task-2",
      segment_id: "build-2",
    },
  ];
  return data;
}
test("the task sheet of a host task is titled by the title alone", async (t) => {
  const { root } = mount(t);
  const data = taskFixture();
  const task = data.session.tasks.find((x) => x.task_id === "task-2");
  assert.ok(task);
  task.task_key = "host:native-2";
  await act(async () => root.render(<Harness data={data} />));
  await click("Open blocked task");
  const title = region('[data-slot="sheet-title"]');
  assert.equal(title.textContent, task.title);
  // The sheet carries the same provenance marker, outside the title text.
  const marker = region('[data-slot="host-observed"]');
  assert.equal(marker.getAttribute("role"), "img");
  assert.ok(marker.getAttribute("aria-label")?.startsWith("From host: "));
  assert.equal(marker.hasAttribute("tabindex"), false);
  assert.equal(title.contains(marker), false);
});

test("the task sheet of an agent task shows no host marker", async (t) => {
  const { root } = mount(t);
  await act(async () => root.render(<Harness data={taskFixture()} />));
  await click("Open blocked task");
  region('[data-slot="sheet-title"]');
  assert.equal(document.querySelector('[data-slot="host-observed"]'), null);
});

test("task snapshot focuses unblock instructions and only its own version offers manual refresh", async (t) => {
  const { root } = mount(t);
  let data = taskFixture();
  const render = async () => {
    await act(async () => root.render(<Harness data={data} />));
  };
  await render();
  await click("Open blocked task");
  let sheet = region('[data-slot="sheet-content"]');
  assert.ok(sheet.textContent?.includes("Solicitar acceso"));
  assert.ok(sheet.textContent?.includes("Equipo del entorno"));
  assert.equal(document.activeElement?.textContent, "Unblocking steps");
  await click("Evidence", sheet);
  assert.ok(sheet.textContent?.includes("Original linked evidence"));
  data = structuredClone(data);
  data.session.version++;
  data.session.tasks[0].task_version++;
  await render();
  assert.equal(sheet.textContent?.includes("An update is available"), false);
  data = structuredClone(data);
  data.session.version++;
  data.session.tasks[1].task_version++;
  const evidence = data.session.events.find(
    (e) => e.event_id === "task-evidence",
  );
  assert.ok(evidence);
  evidence.title = "New linked evidence without a state change";
  await render();
  assert.ok(sheet.textContent?.includes("An update is available"));
  assert.ok(sheet.textContent?.includes("Original linked evidence"));
  assert.equal(sheet.textContent?.includes("New linked evidence"), false);
  await click("Show latest", sheet);
  sheet = region('[data-slot="sheet-content"]');
  assert.ok(
    sheet.textContent?.includes("New linked evidence without a state change"),
  );
  assert.equal(
    sheet.querySelector('[role="tab"][aria-selected="true"]')?.textContent,
    "Evidence",
  );
});
test("task documents and images reuse one Sheet and return to the preserved evidence context", async (t) => {
  const { root } = mount(t);
  const data = taskFixture();
  await act(async () => root.render(<Harness data={data} />));
  await click("Open blocked task");
  await click("Evidence");
  let sheet = region('[data-slot="sheet-content"]');
  const attachment = region('[data-document-key="implementation"]', sheet);
  await act(async () => {
    attachment.click();
    await settle();
  });
  assert.equal(
    document.querySelectorAll('[data-slot="sheet-content"]').length,
    1,
  );
  assert.equal(document.querySelectorAll('[role="dialog"]').length, 1);
  assert.ok(sheet.textContent?.includes("Propuesta · R3"));
  await click("Back to task");
  sheet = region('[data-slot="sheet-content"]');
  assert.equal(
    sheet.querySelector('[role="tab"][aria-selected="true"]')?.textContent,
    "Evidence",
  );
  assert.equal(
    document.activeElement?.getAttribute("data-document-key"),
    "implementation",
  );
  const image = sheet.querySelector<HTMLButtonElement>(
    '[aria-label^="Perfil personal"]',
  );
  assert.ok(image);
  await act(async () => {
    image.click();
    await settle();
  });
  assert.equal(document.querySelectorAll('[role="dialog"]').length, 1);
  assert.ok(sheet.querySelector("img"));
  await click("Back to task");
  assert.ok(sheet.textContent?.includes("Original linked evidence"));
  await click("Close");
});
test("tasks without attachments show an actionable empty region and initial history", async (t) => {
  const { root } = mount(t);
  await act(async () => root.render(<Harness data={taskFixture()} />));
  await click("Open empty task");
  await click("Evidence");
  assert.ok(
    region('[data-slot="sheet-content"]').textContent?.includes(
      "No attachments",
    ),
  );
  assert.ok(button("Reload"));
  await click("History");
  assert.ok(
    region('[data-slot="sheet-content"]').textContent?.includes(
      "Initial record",
    ),
  );
});
test("CI snapshots display only applicable attempts and update through explicit version or successor", async (t) => {
  const { root } = mount(t);
  let data = workFixture();
  const wait = data.session.waits[0];
  wait.attempts.push({
    ...structuredClone(wait.attempts[0]),
    attempt_id: "rerun",
    generation: 2,
    check_keys: ["tests"],
    external_run: {
      run_id: "rerun-2",
      commit: timelineCandidateByID(data.session, wait.candidate_id).commit,
      url: "https://example.test/actions/2",
    },
  });
  wait.applicable_attempt_id.tests = "rerun";
  const prior = wait.observations.find((o) => o.check_key === "tests");
  assert.ok(prior);
  wait.observations.push(
    {
      ...structuredClone(prior),
      observation_id: "late-old-success",
      state: "succeeded",
      observed_at: "2026-11-01T23:00:00Z",
    },
    {
      ...structuredClone(prior),
      observation_id: "new-failure",
      attempt_id: "rerun",
      state: "failed",
      external_run: wait.attempts[1].external_run,
    },
  );
  wait.state = "failed";
  assert.equal(
    currentWaitObservations(
      wait,
      timelineCandidateByID(data.session, wait.candidate_id),
    ).tests.state,
    "failed",
  );
  assert.equal(
    aggregateTimelineCIWait(
      wait.policy,
      wait.applicable_attempt_id,
      currentWaitObservations(
        wait,
        timelineCandidateByID(data.session, wait.candidate_id),
      ),
    ),
    "failed",
  );
  const render = async () => {
    await act(async () => root.render(<Harness data={data} />));
  };
  await render();
  await click("Open CI");
  let sheet = region('[data-slot="sheet-content"]');
  assert.ok(sheet.textContent?.includes("abc1234"));
  assert.ok(sheet.querySelector('a[href="https://example.test/pull/42"]'));
  await click("Checks");
  assert.ok(sheet.textContent?.includes("Pruebas · Failed"));
  await click("Attempts");
  assert.ok(sheet.textContent?.includes("rerun-2"));
  assert.ok(sheet.textContent?.includes("Finished"));
  data = structuredClone(data);
  data.session.version++;
  data.session.waits.push({
    ...structuredClone(data.session.waits[0]),
    wait_id: "unrelated",
    wait_version: 1,
  });
  await render();
  assert.equal(sheet.textContent?.includes("An update is available"), false);
  data = structuredClone(data);
  data.session.waits[0].wait_version++;
  data.session.waits[0].next_action = "Inspect the failing test log";
  await render();
  assert.ok(sheet.textContent?.includes("An update is available"));
  assert.equal(
    sheet.textContent?.includes("Inspect the failing test log"),
    false,
  );
  await click("Show latest");
  await click("Summary");
  assert.ok(sheet.textContent?.includes("Inspect the failing test log"));
  data = structuredClone(data);
  const successor = {
    ...structuredClone(data.session.waits[0]),
    wait_id: "next-candidate",
    previous_wait_id: "ci-wait",
    candidate_id: "candidate-next",
    policy: { ...data.session.waits[0].policy, confirmed: false },
    state: "unknown" as const,
    attempts: [],
    observations: [],
    applicable_attempt_id: {},
  };
  data.session.candidates.push({
    candidate_id: "candidate-next",
    branch: "fix/dashboard",
    commit: "def5678",
    base: "development",
  });
  data.session.waits.push(successor);
  await render();
  assert.ok(sheet.textContent?.includes("abc1234"));
  await click("Show latest");
  sheet = region('[data-slot="sheet-content"]');
  assert.ok(sheet.textContent?.includes("def5678"));
  assert.ok(sheet.textContent?.includes("Unconfirmed"));
  await click("Checks");
  assert.ok(
    sheet.textContent?.includes("The required checks are not confirmed"),
  );
  assert.equal(sheet.querySelector('button[aria-label*="Merge"]'), null);
});

test("CI result labels distinguish success, failure, cancellation and unknown without remote controls", async (t) => {
  const { root } = mount(t);
  for (const [state, label] of [
    ["succeeded", "Finished"],
    ["failed", "Failed"],
    ["cancelled", "Cancelled"],
    ["unknown", "Unconfirmed"],
  ] as const) {
    const data = workFixture();
    const wait = data.session.waits[0];
    wait.state = state;
    for (const observation of wait.observations) observation.state = state;
    await act(async () => root.render(<Harness data={data} />));
    await click("Open CI");
    const sheet = region('[data-slot="sheet-content"]');
    assert.ok(sheet.textContent?.includes(label));
    assert.equal(
      [...sheet.querySelectorAll("button")].some((b) =>
        /rerun|merge|approve/i.test(b.textContent ?? ""),
      ),
      false,
    );
    await click("Close");
  }
});

test("task header indicator names the opened segment and summary keeps global state explicit", async (t) => {
  const { root } = mount(t);
  await act(async () => root.render(<Harness data={workFixture()} />));
  await click("Open verify segment");
  const sheet = region('[data-slot="sheet-content"]');
  assert.equal(
    sheet.querySelector('[role="img"]')?.getAttribute("aria-label"),
    "Completed",
  );
  assert.ok(sheet.textContent?.includes("Overall state: In progress"));
});

test("canonical ticket and PR changes offer manual refresh scoped to the open task and wait", async (t) => {
  const { root } = mount(t);
  let data = taskFixture();
  const render = async () =>
    act(async () => {
      root.render(<Harness data={data} />);
      await settle();
    });
  await render();
  await click("Open blocked task");
  await click("Evidence");
  const ticketID = data.session.events.find(
    (e) => e.event_id === "task-evidence",
  )?.views?.[0].ticket_refs[0].ticket_id;
  assert.ok(ticketID);
  data = structuredClone(data);
  const ticket = data.session.tickets.find((t) => t.ticket_id === ticketID);
  assert.ok(ticket);
  ticket.key = "UPDATED-7";
  ticket.url = "https://example.test/tickets/updated";
  data.session.version++;
  data.session.tasks[1].task_version++;
  await render();
  let sheet = region('[data-slot="sheet-content"]');
  assert.ok(sheet.textContent?.includes("An update is available"));
  assert.equal(sheet.textContent?.includes("UPDATED-7"), false);
  await click("Show latest", sheet);
  assert.ok(sheet.textContent?.includes("UPDATED-7"));
  assert.ok(sheet.textContent?.includes("Published"));
  assert.ok(
    sheet.querySelector('a[href="https://example.test/tickets/updated"]'),
  );
  await click("Close");
  data = workFixture("cam");
  await render();
  await click("Open CI");
  sheet = region('[data-slot="sheet-content"]');
  const originalCandidate = data.session.waits[0].candidate_id;
  data = structuredClone(data);
  const delivery = data.session.deliveries.find(
    (d) => d.delivery_id === data.session.waits[0].pull_request_ref.delivery_id,
  );
  assert.ok(delivery);
  delivery.url = "https://example.test/pull/corrected";
  delivery.label = "Corrected PR";
  data.session.waits[0].wait_version++;
  data.session.version++;
  await render();
  assert.equal(sheet.textContent?.includes("Corrected PR"), false);
  assert.ok(sheet.textContent?.includes("An update is available"));
  await click("Show latest", sheet);
  assert.ok(sheet.textContent?.includes("Corrected PR"));
  assert.ok(
    sheet.querySelector('a[href="https://example.test/pull/corrected"]'),
  );
  assert.equal(data.session.waits[0].candidate_id, originalCandidate);
  assert.equal(
    data.session.waits[0].pull_request_ref.snapshot.url,
    "https://example.test/pull/42",
  );
});
