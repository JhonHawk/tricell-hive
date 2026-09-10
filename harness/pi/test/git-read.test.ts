import { spawn } from "node:child_process";
import test from "node:test";
import assert from "node:assert/strict";
import { buildGitArguments, runBoundedGit } from "../src/git-read.ts";

test("git read disables pager, fsmonitor and external diff and keeps paths after --", () => {
  const args = buildGitArguments({ operation: "diff", path: "folder with spaces/file.ts", staged: true });
  assert.deepEqual(args, [
    "--no-pager",
    "-c",
    "core.fsmonitor=false",
    "-c",
    "core.pager=cat",
    "diff",
    "--no-ext-diff",
    "--no-textconv",
    "--cached",
    "--",
    "folder with spaces/file.ts",
  ]);
});

test("git read caps log requests and bounds revisions", () => {
  const args = buildGitArguments({ operation: "log", limit: 999 });
  assert.equal(args.at(-1), "-200");
  assert.deepEqual(buildGitArguments({ operation: "show", revision: "HEAD~2", path: "src/main.ts" }).slice(-6), [
    "show",
    "--no-ext-diff",
    "--no-textconv",
    "HEAD~2",
    "--",
    "src/main.ts",
  ]);
});

test("git read rejects option injection, controls and unsupported blame ranges", () => {
  assert.throws(() => buildGitArguments({ operation: "diff", path: "--output=evil" }), /cannot start/u);
  assert.throws(() => buildGitArguments({ operation: "show", revision: "HEAD;touch" }), /unsupported/u);
  assert.throws(() => buildGitArguments({ operation: "blame", line: "1 -L 2" }), /N or N,M/u);
  assert.throws(() => buildGitArguments({ operation: "status", path: "line\nfeed" }), /control/u);
});

test("git read uses explicit read-only forms for each operation", () => {
  assert.deepEqual(buildGitArguments({ operation: "status" }).slice(-3), ["status", "--short", "--branch"]);
  assert.deepEqual(buildGitArguments({ operation: "blame", line: "4,2", path: "file.ts" }).slice(-5), ["blame", "-L", "4,2", "--", "file.ts"]);
  assert.deepEqual(buildGitArguments({ operation: "ls-files", path: "src" }).slice(-2), ["--", "src"]);
  assert.deepEqual(buildGitArguments({ operation: "branch" }).slice(-3), ["branch", "--list", "--no-color"]);
  assert.deepEqual(buildGitArguments({ operation: "remote" }).slice(-1), ["remote"]);
});

test("git read terminates and bounds an over-producing process", async () => {
  const result = await runBoundedGit([], process.cwd(), undefined, (_command, _args, options) => spawn(
    process.execPath,
    ["-e", "process.stdout.write('x'.repeat(500000))"],
    options,
  ));
  assert.equal(result.outputLimit, true);
  assert.equal(result.killed, true);
  assert.ok(Buffer.byteLength(result.stdout, "utf8") <= 200_000);
});
