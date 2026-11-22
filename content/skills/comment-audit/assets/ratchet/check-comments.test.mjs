import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { copyFileSync, mkdirSync, mkdtempSync, rmSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";

const script = fileURLToPath(new URL("./check-comments.mjs", import.meta.url));
const env = {
  ...process.env,
  GIT_CONFIG_GLOBAL: "/dev/null",
  GIT_CONFIG_NOSYSTEM: "1",
  GIT_AUTHOR_NAME: "test",
  GIT_AUTHOR_EMAIL: "test@example.com",
  GIT_COMMITTER_NAME: "test",
  GIT_COMMITTER_EMAIL: "test@example.com",
};

function git(cwd, ...args) {
  const result = spawnSync("git", ["-c", "commit.gpgsign=false", ...args], {
    cwd,
    env,
    encoding: "utf8",
  });
  assert.equal(result.status, 0, result.stderr);
  return result.stdout.trim();
}

function repo(t) {
  const root = mkdtempSync(join(tmpdir(), "comment-ratchet-"));
  t.after(() => rmSync(root, { recursive: true, force: true }));
  git(root, "init", "-q", "-b", "main");
  return {
    root,
    write(path, text) {
      mkdirSync(dirname(join(root, path)), { recursive: true });
      writeFileSync(join(root, path), text);
    },
    commit(message) {
      git(root, "add", "-A");
      git(root, "commit", "-q", "--allow-empty", "-m", message);
      return git(root, "rev-parse", "HEAD");
    },
    git: (...args) => git(root, ...args),
    run(args, cwd = root, path = script) {
      return spawnSync(process.execPath, [path, ...args], { cwd, env, encoding: "utf8" });
    },
  };
}

test("a change that adds no comments passes", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "// kept\nexport const a = 1;\n");
  const base = r.commit("base");
  r.write("src/a.ts", "// kept\nexport const a = 2;\n");
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("a new comment fails with its file, line and column", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write("src/a.ts", "export const a = 1;\n  // added\n");
  const result = r.run(["--base", base]);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /^src\/a\.ts:2:3: no-new-comments$/m);
});

test("a duplicate of an existing comment fails", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "// same\nexport const a = 1;\n");
  const base = r.commit("base");
  r.write("src/a.ts", "// same\nexport const a = 1;\n// same\n");
  assert.equal(r.run(["--base", base]).status, 1);
});

test("a comment the base removed cannot come back", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "// old\nexport const a = 1;\n");
  r.commit("with comment");
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("cleanup");
  r.write("src/a.ts", "// old\nexport const a = 1;\n");
  assert.equal(r.run(["--base", base]).status, 1);
});

test("a renamed file keeps the comments it had", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "// kept\nexport const a = 1;\nexport const b = 2;\n");
  const base = r.commit("base");
  r.git("config", "diff.renames", "false");
  r.git("mv", "src/a.ts", "src/b.ts");
  r.write("src/b.ts", "// kept\nexport const a = 1;\nexport const b = 3;\n");
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("a non-ASCII path is reported as written", (t) => {
  const r = repo(t);
  r.write("src/año.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write("src/año.ts", "export const a = 1; // added\n");
  const result = r.run(["--base", base]);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /^src\/año\.ts:1:21: no-new-comments$/m);
});

test("running from a subfolder still checks the whole repository", (t) => {
  const r = repo(t);
  r.write("pkg/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write("pkg/a.ts", "export const a = 1;\n// added\n");
  const result = r.run(["--base", base], join(r.root, "pkg"));
  assert.equal(result.status, 1);
  assert.match(result.stderr, /^pkg\/a\.ts:2:1: no-new-comments$/m);
});

test("comment-like text in strings, templates and regexes is not a comment", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write(
    "src/a.ts",
    'export const a = "// no";\nexport const b = `/* no */ ${a}`;\nexport const c = /\\/\\/ no/;\n',
  );
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("JSX text is not a comment and a JSX comment is", (t) => {
  const r = repo(t);
  r.write("src/a.tsx", "export const A = () => <p>x</p>;\n");
  const base = r.commit("base");
  r.write("src/a.tsx", "export const A = () => <p>// text</p>;\n");
  assert.equal(r.run(["--base", base]).status, 0);
  r.write("src/a.tsx", "export const A = () => <p>{/* note */}x</p>;\n");
  assert.equal(r.run(["--base", base]).status, 1);
});

test("type assertions and decorators in .ts files parse", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write(
    "src/a.ts",
    "declare const x: unknown;\nexport const a = <string>x;\ndeclare function D(t: unknown): void;\n@D\nexport class C {}\n",
  );
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("excluded paths are skipped", (t) => {
  const r = repo(t);
  r.write("dist/a.js", "export const a = 1;\n");
  const base = r.commit("base");
  r.write("dist/a.js", "export const a = 1;\n// generated\n");
  assert.equal(r.run(["--base", base]).status, 0);
});

test("allowed directives and the workaround marker pass", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write(
    "src/a.ts",
    '/// <reference types="node" />\n// workaround: COM-1\nexport const a = 1;\n',
  );
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("a base that does not resolve exits 2", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  r.commit("base");
  const result = r.run(["--base", "no-such-revision"]);
  assert.equal(result.status, 2);
  assert.match(result.stderr, /no-such-revision/);
});

test("a missing parser exits 2", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  const lone = mkdtempSync(join(tmpdir(), "comment-ratchet-lone-"));
  t.after(() => rmSync(lone, { recursive: true, force: true }));
  const copy = join(lone, "check-comments.mjs");
  copyFileSync(script, copy);
  const result = r.run(["--base", base], r.root, copy);
  assert.equal(result.status, 2);
  assert.match(result.stderr, /@babel\/parser/);
});

test("the inventory lists each comment with its fields", (t) => {
  const r = repo(t);
  r.write(
    "src/a.ts",
    '/// <reference types="node" />\n// whole line\nexport const a = 1; // trailing\n',
  );
  r.commit("base");
  const result = r.run(["--inventory"]);
  assert.equal(result.status, 0, result.stderr);
  const rows = result.stdout.trim().split("\n").map((line) => JSON.parse(line));
  assert.deepEqual(
    rows.map(({ file, line, column, fullLine, directive, text }) => [
      file,
      line,
      column,
      fullLine,
      directive,
      text,
    ]),
    [
      ["src/a.ts", 1, 1, true, true, '/// <reference types="node" />'],
      ["src/a.ts", 2, 1, true, false, "// whole line"],
      ["src/a.ts", 3, 21, false, false, "// trailing"],
    ],
  );
  for (const row of rows) assert.ok(row.end > row.start);
});

test("--same-code accepts removed comments, also in JSX, and rejects other changes", (t) => {
  const r = repo(t);
  r.write(
    "src/a.tsx",
    [
      "// header",
      "export const A = () => (",
      "  <div>",
      "    {/* note */}",
      "    <p>a {/* x */} b</p>",
      "  </div>",
      "); // trailing",
      "",
    ].join("\n"),
  );
  const base = r.commit("base");
  r.write(
    "src/a.tsx",
    ["export const A = () => (", "  <div>", "    <p>a b</p>", "  </div>", ");", ""].join("\n"),
  );
  const clean = r.run(["--same-code", "--base", base]);
  assert.equal(clean.status, 0, clean.stderr);
  r.write(
    "src/a.tsx",
    ["export const A = () => (", "  <div>", "    <p>a c</p>", "  </div>", ");", ""].join("\n"),
  );
  const changed = r.run(["--same-code", "--base", base]);
  assert.equal(changed.status, 1);
  assert.match(changed.stderr, /^src\/a\.tsx: /m);
});
