import { chmodSync, existsSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import assert from "node:assert/strict";
import { checkHookReadiness, runHook } from "../src/hook-runner.ts";

function makeHook(source: string, executable = true): { readonly directory: string; readonly path: string } {
  const directory = mkdtempSync(join(tmpdir(), "hive-pi-hook-"));
  const path = join(directory, "hook-script;literal.sh");
  writeFileSync(path, source, "utf8");
  chmodSync(path, executable ? 0o755 : 0o644);
  return { directory, path };
}

test("runHook passes JSON and argv without a shell and allows valid blocking output", async () => {
  const hook = makeHook(`#!/usr/bin/env node
let input = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => { input += chunk; });
process.stdin.on("end", () => {
  const payload = JSON.parse(input);
process.stdout.write(JSON.stringify({ permissionDecision: "allow", additionalContext: process.argv.slice(2).join("|") + ":" + payload.value }));
});
`);
  try {
    const result = await runHook({
      scriptPath: hook.path,
      args: ["--from-pi-command", "literal;arg"],
      cwd: hook.directory,
      payload: { value: "exact-json" },
      mode: "blocking",
    });
    assert.equal(result.outcome, "warning");
    assert.equal(result.additionalContext, "--from-pi-command|literal;arg:exact-json");
    assert.equal(result.exitCode, 0);
  } finally {
    rmSync(hook.directory, { recursive: true, force: true });
  }
});

test("runHook fails closed for a blocking deny, invalid output and timeout", async () => {
  const deny = makeHook(`#!/usr/bin/env node
process.stdout.write(JSON.stringify({ decision: "deny", reason: "policy denied" }));
process.exit(2);
`);
  const invalid = makeHook(`#!/usr/bin/env node
process.stdout.write("not-json");
`);
  const unknown = makeHook(`#!/usr/bin/env node
process.stdout.write(JSON.stringify({ unexpected: true }));
`);
  const mixed = makeHook(`#!/usr/bin/env node
process.stdout.write("diagnostic\\n" + JSON.stringify({ decision: "allow" }));
`);
  const reasonOnly = makeHook(`#!/usr/bin/env node
process.stdout.write(JSON.stringify({ reason: "blocked" }));
`);
  const unknownDecision = makeHook(`#!/usr/bin/env node
process.stdout.write(JSON.stringify({ decision: "banana" }));
`);
  const capturedStatus = makeHook(`#!/usr/bin/env node
process.stdout.write(JSON.stringify({ status: "captured" }));
`);
  const uppercaseDeny = makeHook(`#!/usr/bin/env node
process.stdout.write(JSON.stringify({ decision: "DENY", reason: "policy denied" }));
`);
  const askDecision = makeHook(`#!/usr/bin/env node
process.stdout.write(JSON.stringify({ decision: "ask" }));
`);
  const timeout = makeHook(`#!/usr/bin/env node
setTimeout(() => process.stdout.write(JSON.stringify({ decision: "allow" })), 1000);
`);
  try {
    const denied = await runHook({ scriptPath: deny.path, cwd: deny.directory, payload: {}, mode: "blocking" });
    assert.equal(denied.outcome, "block");
    assert.equal(denied.reason, "policy denied");

    const malformed = await runHook({ scriptPath: invalid.path, cwd: invalid.directory, payload: {}, mode: "blocking" });
    assert.equal(malformed.outcome, "block");
    assert.match(malformed.reason ?? "", /invalid JSON/u);

    const unrecognized = await runHook({ scriptPath: unknown.path, cwd: unknown.directory, payload: {}, mode: "blocking" });
    assert.equal(unrecognized.outcome, "block");
    assert.match(unrecognized.reason ?? "", /invalid JSON/u);

    const mixedOutput = await runHook({ scriptPath: mixed.path, cwd: mixed.directory, payload: {}, mode: "blocking" });
    assert.equal(mixedOutput.outcome, "block");
    assert.match(mixedOutput.reason ?? "", /invalid JSON/u);

    const reasonOnlyOutput = await runHook({ scriptPath: reasonOnly.path, cwd: reasonOnly.directory, payload: {}, mode: "blocking" });
    assert.equal(reasonOnlyOutput.outcome, "block");
    assert.equal(reasonOnlyOutput.reason, "blocked");

    const unknownDecisionOutput = await runHook({ scriptPath: unknownDecision.path, cwd: unknownDecision.directory, payload: {}, mode: "blocking" });
    assert.equal(unknownDecisionOutput.outcome, "block");
    assert.match(unknownDecisionOutput.reason ?? "", /invalid JSON/u);

    const capturedStatusOutput = await runHook({ scriptPath: capturedStatus.path, cwd: capturedStatus.directory, payload: {}, mode: "blocking" });
    assert.equal(capturedStatusOutput.outcome, "block");
    assert.match(capturedStatusOutput.reason ?? "", /invalid JSON/u);

    const uppercaseDenyOutput = await runHook({ scriptPath: uppercaseDeny.path, cwd: uppercaseDeny.directory, payload: {}, mode: "blocking" });
    assert.equal(uppercaseDenyOutput.outcome, "block");
    assert.equal(uppercaseDenyOutput.reason, "policy denied");

    const askOutput = await runHook({ scriptPath: askDecision.path, cwd: askDecision.directory, payload: {}, mode: "blocking" });
    assert.equal(askOutput.outcome, "block");

    const timedOut = await runHook({
      scriptPath: timeout.path,
      cwd: timeout.directory,
      payload: {},
      mode: "blocking",
      timeoutMs: 25,
    });
    assert.equal(timedOut.outcome, "block");
    assert.equal(timedOut.timedOut, true);
  } finally {
    for (const hook of [deny, invalid, unknown, mixed, reasonOnly, unknownDecision, capturedStatus, uppercaseDeny, askDecision, timeout]) rmSync(hook.directory, { recursive: true, force: true });
  }
});

test("runHook kills timed-out descendants with the hook process group", { skip: process.platform === "win32" }, async () => {
  const hook = makeHook([
    "#!/bin/sh",
    '(sleep 2; printf late > "$HIVE_LATE_MARKER") &',
    `printf '{"decision":"allow"}'`,
    "exit 0",
    "",
  ].join("\n"));
  const marker = join(hook.directory, "late-marker");
  const started = Date.now();
  try {
    const result = await runHook({
      scriptPath: hook.path,
      cwd: hook.directory,
      payload: {},
      mode: "blocking",
      timeoutMs: 100,
    }, { env: { HIVE_LATE_MARKER: marker } });
    assert.equal(result.outcome, "block");
    assert.equal(result.timedOut, true);
    assert.ok(Date.now() - started < 1_500, `timeout took ${String(Date.now() - started)}ms`);
    await new Promise((resolve) => setTimeout(resolve, 300));
    assert.equal(existsSync(marker), false);
  } finally {
    rmSync(hook.directory, { recursive: true, force: true });
  }
});

test("runHook bounds hook output before returning", async () => {
  const hook = makeHook("#!/usr/bin/env node\nprocess.stdout.write(\"x\".repeat(400000));\nsetTimeout(() => undefined, 2000);\n");
  const started = Date.now();
  try {
    const result = await runHook({ scriptPath: hook.path, cwd: hook.directory, payload: {}, mode: "blocking", timeoutMs: 5_000 });
    assert.equal(result.outcome, "block");
    assert.match(result.reason ?? "", /output exceeded|invalid JSON/u);
    assert.ok(result.stdout.length <= 262_144);
    assert.ok(Date.now() - started < 1_500, `output limit took ${String(Date.now() - started)}ms`);
  } finally {
    rmSync(hook.directory, { recursive: true, force: true });
  }
});

test("runHook handles a pre-aborted signal without an uncaught stdin EPIPE", async () => {
  const hook = makeHook("#!/usr/bin/env node\nsetTimeout(() => undefined, 2_000);\n");
  const controller = new AbortController();
  controller.abort();
  try {
    const result = await runHook({
      scriptPath: hook.path,
      cwd: hook.directory,
      payload: { large: "x".repeat(100_000) },
      mode: "blocking",
      signal: controller.signal,
      timeoutMs: 500,
    });
    assert.equal(result.outcome, "block");
  } finally {
    rmSync(hook.directory, { recursive: true, force: true });
  }
});

test("runHook applies per-invocation environment and removes requested variables", async () => {
  const hook = makeHook(`#!/usr/bin/env node
process.stdout.write(JSON.stringify({ additionalContext: JSON.stringify({ claude: process.env.CLAUDECODE ?? null, repo: process.env.HIVE_REPO ?? null }) }));
`);
  const previousClaude = process.env.CLAUDECODE;
  try {
    process.env.CLAUDECODE = "inherited";
    const result = await runHook({
      scriptPath: hook.path,
      cwd: hook.directory,
      payload: {},
      mode: "advisory",
      env: { HIVE_REPO: "/known/repo" },
      unsetEnv: ["CLAUDECODE"],
    });
    assert.equal(result.outcome, "warning");
    assert.equal(result.additionalContext, JSON.stringify({ claude: null, repo: "/known/repo" }));
  } finally {
    if (previousClaude === undefined) delete process.env.CLAUDECODE;
    else process.env.CLAUDECODE = previousClaude;
    rmSync(hook.directory, { recursive: true, force: true });
  }
});

test("advisory failures warn while missing hooks block only in blocking mode", async () => {
  const advisory = makeHook(`#!/usr/bin/env node
process.stderr.write("advisory failure");
process.exit(1);
`);
  const missingPath = join(advisory.directory, "missing-hook.sh");
  try {
    const warning = await runHook({ scriptPath: advisory.path, cwd: advisory.directory, payload: {}, mode: "advisory" });
    assert.equal(warning.outcome, "warning");
    assert.equal(warning.reason, "advisory failure");

    const missingBlocking = await runHook({ scriptPath: missingPath, cwd: advisory.directory, payload: {}, mode: "blocking" });
    assert.equal(missingBlocking.outcome, "block");
    const missingAdvisory = await runHook({ scriptPath: missingPath, cwd: advisory.directory, payload: {}, mode: "advisory" });
    assert.equal(missingAdvisory.outcome, "warning");
  } finally {
    rmSync(advisory.directory, { recursive: true, force: true });
  }
});

test("hook readiness requires executable files", () => {
  const executable = makeHook("#!/bin/sh\nexit 0\n");
  const notExecutable = makeHook("#!/bin/sh\nexit 0\n", false);
  try {
    assert.deepEqual(checkHookReadiness([executable.path]), { ready: true, missing: [] });
    assert.deepEqual(checkHookReadiness([executable.path, notExecutable.path]), {
      ready: false,
      missing: [notExecutable.path],
    });
  } finally {
    rmSync(executable.directory, { recursive: true, force: true });
    rmSync(notExecutable.directory, { recursive: true, force: true });
  }
});
