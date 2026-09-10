import { mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import assert from "node:assert/strict";
import type { ExtensionAPI } from "@earendil-works/pi-coding-agent";
import { attachChildLifecycle, createChildRegistry, reconcileActiveChildren } from "../src/child-registry.ts";

interface FakeEvents {
  readonly on: (channel: string, handler: (payload: unknown) => void) => () => void;
  readonly emit: (channel: string, payload: unknown) => void;
}

function fakeEvents(): FakeEvents {
  const listeners = new Map<string, Set<(payload: unknown) => void>>();
  return {
    on(channel, handler) {
      const channelListeners = listeners.get(channel) ?? new Set<(payload: unknown) => void>();
      channelListeners.add(handler);
      listeners.set(channel, channelListeners);
      return () => channelListeners.delete(handler);
    },
    emit(channel, payload) {
      for (const handler of listeners.get(channel) ?? []) handler(payload);
    },
  };
}

test("child registry tracks async children until a terminal lifecycle event", () => {
  const events = fakeEvents();
  const registry = createChildRegistry();
  const pi = { events } as unknown as ExtensionAPI;
  const detach = attachChildLifecycle(pi, registry, () => "parent-session");
  try {
    events.emit("subagent:async-started", { id: "run-1", sessionId: "parent-session" });
    assert.equal(registry.hasActiveChildren("parent-session"), true);
    events.emit("subagent:async-complete", { id: "run-1", sessionId: "parent-session" });
    assert.equal(registry.hasActiveChildren("parent-session"), false);

    events.emit("subagent:async-started", { id: "run-2" });
    assert.equal(registry.hasActiveChildren("parent-session"), true);
    events.emit("subagent:process-terminal", { runId: "run-2" });
    assert.equal(registry.hasActiveChildren("parent-session"), false);
  } finally {
    detach();
  }
});

test("child registry isolates parents and ignores self-registration", () => {
  const registry = createChildRegistry();
  registry.register("parent-a", "child-a");
  registry.register("parent-b", "child-b");
  registry.register("parent-a", "parent-a");
  assert.equal(registry.hasActiveChildren("parent-a"), true);
  assert.equal(registry.hasActiveChildren("parent-b"), true);
  registry.settle("child-a");
  assert.equal(registry.hasActiveChildren("parent-a"), false);
  assert.equal(registry.hasActiveChildren("parent-b"), true);
});

test("child registry reconciles persisted active runs for the current parent after a restart", async () => {
  const events = fakeEvents();
  const asyncRoot = mkdtempSync(join(tmpdir(), "hive-pi-active-runs-"));
  const runId = "run-persisted";
  mkdirSync(join(asyncRoot, ".active-runs"));
  mkdirSync(join(asyncRoot, ".active-runs", "tool-calls"));
  mkdirSync(join(asyncRoot, runId));
  writeFileSync(join(asyncRoot, ".active-runs", runId), "");
  writeFileSync(join(asyncRoot, runId, "status.json"), JSON.stringify({ runId, sessionId: "parent-session", state: "running" }));
  try {
    const first = createChildRegistry();
    const pi = { events } as unknown as ExtensionAPI;
    const firstReconciliation = await reconcileActiveChildren(pi, first, "parent-session", { asyncRoot, timeoutMs: 5 });
    assert.equal(firstReconciliation.active, true);
    assert.equal(first.hasActiveChildren("parent-session"), true);

    const resumed = createChildRegistry();
    const resumedReconciliation = await reconcileActiveChildren(pi, resumed, "parent-session", { asyncRoot, timeoutMs: 5 });
    assert.equal(resumedReconciliation.active, true);
    assert.equal(resumed.hasActiveChildren("parent-session"), true);

    const otherSession = createChildRegistry();
    const otherReconciliation = await reconcileActiveChildren(pi, otherSession, "other-session", { asyncRoot, timeoutMs: 5 });
    assert.equal(otherReconciliation.active, false);
    assert.equal(otherSession.hasActiveChildren("other-session"), false);
  } finally {
    rmSync(asyncRoot, { recursive: true, force: true });
  }
});

test("child registry fails closed when the current session status marker is malformed", async () => {
  const events = fakeEvents();
  const asyncRoot = mkdtempSync(join(tmpdir(), "hive-pi-active-corrupt-"));
  const runId = "run-corrupt";
  mkdirSync(join(asyncRoot, ".active-runs"));
  mkdirSync(join(asyncRoot, runId));
  writeFileSync(join(asyncRoot, ".active-runs", runId), "");
  writeFileSync(join(asyncRoot, runId, "status.json"), JSON.stringify({ runId, sessionId: "parent-session" }));
  try {
    const registry = createChildRegistry();
    const pi = { events } as unknown as ExtensionAPI;
    const reconciliation = await reconcileActiveChildren(pi, registry, "parent-session", { asyncRoot, timeoutMs: 5 });
    assert.equal(reconciliation.active, true);
    assert.equal(reconciliation.indeterminate, true);
    assert.equal(registry.hasActiveChildren("parent-session"), true);
  } finally {
    rmSync(asyncRoot, { recursive: true, force: true });
  }
});

test("child registry scans beyond the old marker boundary for the current session", async () => {
  const events = fakeEvents();
  const asyncRoot = mkdtempSync(join(tmpdir(), "hive-pi-active-runs-many-"));
  const indexRoot = join(asyncRoot, ".active-runs");
  mkdirSync(indexRoot);
  try {
    for (let index = 0; index < 256; index += 1) {
      const runId = `foreign-${String(index).padStart(3, "0")}`;
      mkdirSync(join(asyncRoot, runId));
      writeFileSync(join(indexRoot, runId), "");
      writeFileSync(
        join(asyncRoot, runId, "status.json"),
        JSON.stringify({ runId, sessionId: "other-session", state: "running" }),
      );
    }
    const currentRunId = "z-current";
    mkdirSync(join(asyncRoot, currentRunId));
    writeFileSync(join(indexRoot, currentRunId), "");
    writeFileSync(
      join(asyncRoot, currentRunId, "status.json"),
      JSON.stringify({ runId: currentRunId, sessionId: "parent-session", state: "running" }),
    );

    const registry = createChildRegistry();
    const pi = { events } as unknown as ExtensionAPI;
    const reconciliation = await reconcileActiveChildren(pi, registry, "parent-session", { asyncRoot, timeoutMs: 5 });
    assert.equal(reconciliation.active, true);
    assert.equal(registry.hasActiveChildren("parent-session"), true);
  } finally {
    rmSync(asyncRoot, { recursive: true, force: true });
  }
});
