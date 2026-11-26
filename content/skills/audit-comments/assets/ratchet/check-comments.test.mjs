import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import { copyFileSync, mkdirSync, mkdtempSync, readFileSync, rmSync, symlinkSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { test } from "node:test";
import { fileURLToPath } from "node:url";

const script = fileURLToPath(new URL("./check-comments.mjs", import.meta.url));
const env = {
  ...Object.fromEntries(Object.entries(process.env).filter(([key]) => !key.startsWith("GIT_"))),
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

test("adding the ratchet to a project passes its own lint", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  for (const name of ["check-comments.mjs", "check-comments.test.mjs"]) {
    copyFileSync(fileURLToPath(new URL(`./${name}`, import.meta.url)), join(r.root, name));
  }
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("public API docs stay on configured declarations without exempting private prose", (t) => {
  const r = repo(t);
  const harness = mkdtempSync(join(tmpdir(), "public-doc-lint-"));
  t.after(() => rmSync(harness, { recursive: true, force: true }));
  const scope = join(harness, "node_modules", "@babel");
  mkdirSync(scope, { recursive: true });
  symlinkSync(dirname(fileURLToPath(import.meta.resolve("@babel/parser/package.json"))), join(scope, "parser"));
  const configured = join(harness, "check-comments.mjs");
  const template = readFileSync(script, "utf8");
  writeFileSync(configured, template.replace("function publicApiDocStarts(",
    'PUBLIC_API_SYMBOLS.set("api.ts", new Set(["openStream"]));\nfunction publicApiDocStarts('));
  const base = r.commit("empty");
  const doc = "/** The caller must close the returned stream. */";
  const api = "export function openStream() {}";
  r.write("api.ts", [doc, api].join("\n"));
  const lint = (rev) => r.run(["--base", rev], r.root, configured);
  assert.equal(lint(base).status, 0);
  const published = r.commit("public API");
  r.write("api.ts", [api, doc, "function helper() {}"].join("\n"));
  r.write("internal.ts", [doc, api].join("\n"));
  const rejected = lint(published);
  assert.equal(rejected.status, 1, rejected.stderr);
  assert.match(rejected.stderr, /api\.ts:2:1: no-new-comments/);
  assert.match(rejected.stderr, /internal\.ts:1:1: no-new-comments/);
  r.write("api.ts", [doc, api, "// internal note"].join("\n"));
  const inventory = r.run(["--inventory"], r.root, configured);
  assert.equal(inventory.status, 0, inventory.stderr);
  const rows = inventory.stdout.trim().split("\n").map((row) => JSON.parse(row));
  assert.deepEqual(rows.filter((row) => row.publicApiDoc).map((row) => [row.file, row.text]), [["api.ts", doc]]);
});

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
    '/// <reference types="node" />\n// workaround: ABC-1\nexport const a = 1;\n',
  );
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("license headers and JSX forms of directives and the marker pass", (t) => {
  const r = repo(t);
  r.write("src/a.tsx", "export const A = () => <div>x</div>;\n");
  const base = r.commit("base");
  r.write("src/b.ts", "// SPDX-License-Identifier: MIT\nexport const b = 1;\n");
  r.write("src/c.ts", "/* Copyright 2026 Example Ltd. */\nexport const c = 1;\n");
  r.write(
    "src/a.tsx",
    "declare const X: (p: { a: string }) => null;\nexport const A = () => (\n  <div>\n    {/* @ts-expect-error */}\n    <X a={1} />\n    {/* workaround: ABC-1 */}\n    <p>x</p>\n  </div>\n);\n",
  );
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("license words below the file header are a new comment", (t) => {
  const r = repo(t);
  r.write("src/a.tsx", "export const A = () => <p>x</p>;\n");
  const base = r.commit("base");
  const prose = [
    "// we skip the @license check here",
    "/*! remember to update this before 2027 */",
    "// Copyright notice is broken, fix later",
    "/* workaround: ABC-2 */",
    "/*! TODO: check license header later */",
    "// (c) fall back to the cache",
  ];
  r.write("src/b.ts", `export const b = 1;\n${prose.join("\n")}\n`);
  const result = r.run(["--base", base]);
  assert.equal(result.status, 1);
  prose.forEach((_, index) => {
    assert.match(result.stderr, new RegExp(`^src/b\\.ts:${index + 2}:1: no-new-comments$`, "m"));
  });
  r.write("src/c.ts", "// Copyright 2026 Example Ltd.\n\n// TODO: split this module\nexport const c = 1;\n");
  r.write("src/b.ts", "export const b = 1;\n");
  const separated = r.run(["--base", base]);
  assert.equal(separated.status, 1);
  assert.match(separated.stderr, /^src\/c\.ts:3:1: no-new-comments$/m);
  assert.doesNotMatch(separated.stderr, /^src\/c\.ts:1:/m);
});

test("a file header that is not a license, or a task note inside one, is a new comment", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write("src/b.ts", "// Validates the customer's license key before checkout\nexport const b = 1;\n");
  r.write("src/c.ts", "// Steps: (a) parse, (b) map, (c) reduce\nexport const c = 1;\n");
  r.write("src/d.ts", "// Copyright 2026 Acme\n// HACK: works around the race in the cache\nexport const d = 1;\n");
  r.write("src/e.js", "export const e = 1;\n/*! tiny-lib v1.0 | MIT */\nexport const f = 2;\n");
  const result = r.run(["--base", base]);
  assert.equal(result.status, 1);
  assert.match(result.stderr, /^src\/b\.ts:1:1: no-new-comments$/m);
  assert.match(result.stderr, /^src\/c\.ts:1:1: no-new-comments$/m);
  assert.match(result.stderr, /^src\/d\.ts:2:1: no-new-comments$/m);
  assert.doesNotMatch(result.stderr, /^src\/d\.ts:1:/m);
  assert.match(result.stderr, /^src\/e\.js:2:1: no-new-comments$/m);
});

test("a license header after a shebang or a directive, and an explicit license tag anywhere, are directives", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write("bin/cli.js", "#!/usr/bin/env node\n// Copyright 2026 X\n// SPDX-License-Identifier: MIT\nconsole.log(1);\n");
  r.write("src/client.js", '"use client";\n// Copyright 2026 X\nexport const c = 1;\n');
  r.write("src/bundle.js", "export const a = 1;\n/*! @license tiny-lib v1.0 MIT */\nexport const b = 2;\n");
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("a license header before the first code token is a directive", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  const headers = [
    "/** @license MIT */",
    "/**\n * @license\n * Copyright Google LLC All Rights Reserved.\n */",
    "/*\n * Foo lib v1.0\n * @license MIT\n */",
    "// @license MIT",
    "/*! MIT License */",
    "/*! foo v1.2 | MIT License | github.com/x/foo */",
    "/*! Copyright Example Ltd. */",
    "/*! (c) 2026 Foo Inc. */",
    "// Copyright (C) 2026 Free Software Foundation",
    "// Copyright The Kubernetes Authors.",
    "// SPDX-License-Identifier: MIT",
    "// Copyright 2026 Google LLC\n//\n// Licensed under the Apache License, Version 2.0 (the \"License\");\n// you may not use this file except in compliance with the License.",
  ];
  headers.forEach((header, index) => r.write(`src/h${index}.ts`, `${header}\nexport const h = ${index};\n`));
  const result = r.run(["--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("a symlinked source file is skipped, not fatal", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  symlinkSync("a.ts", join(r.root, "src/link.ts"));
  r.write("src/b.ts", "export const b = 1;\n// added\n");
  const lint = r.run(["--base", base]);
  assert.equal(lint.status, 1, lint.stderr);
  assert.match(lint.stderr, /^src\/b\.ts:2:1: no-new-comments$/m);
  assert.equal(r.run(["--inventory"]).status, 0);
});

test("the inventory does not read through a directory symlink out of the repository", (t) => {
  const r = repo(t);
  const outside = mkdtempSync(join(tmpdir(), "comment-ratchet-outside-"));
  t.after(() => rmSync(outside, { recursive: true, force: true }));
  writeFileSync(join(outside, "x.ts"), "// leaked\nexport const x = 1;\n");
  r.write("d/x.ts", "export const x = 1;\n");
  r.commit("base");
  rmSync(join(r.root, "d"), { recursive: true });
  symlinkSync(outside, join(r.root, "d"));
  const result = r.run(["--inventory"]);
  assert.equal(result.status, 0, result.stderr);
  assert.doesNotMatch(result.stdout, /leaked/);
});

test("a name that starts with two dots is still checked", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write("..lib/x.ts", "export const x = 1;\n// added\n");
  const result = r.run(["--base", base]);
  assert.equal(result.status, 1, result.stderr);
  assert.match(result.stderr, /^\.\.lib\/x\.ts:2:1: no-new-comments$/m);
});

test("control characters in a file name are escaped in findings", (t) => {
  const r = repo(t);
  r.write("src/a.ts", "export const a = 1;\n");
  const base = r.commit("base");
  r.write("src/x\n::error::forged.ts", "export const x = 1;\n// added\n");
  const result = r.run(["--base", base]);
  assert.equal(result.status, 1);
  assert.doesNotMatch(result.stderr, /^::error/m);
  assert.match(result.stderr, /x\\x0a::error::forged\.ts:2:1: no-new-comments/);
});

test("--same-code rejects a removal that changes line breaks between tokens", (t) => {
  const r = repo(t);
  r.write("src/a.js", "export function f() {\n  return /*\n  */ 1;\n}\n");
  const base = r.commit("base");
  r.write("src/a.js", "export function f() {\n  return 1;\n}\n");
  assert.equal(r.run(["--same-code", "--base", base]).status, 1);
});

test("--same-code accepts removing a multi-line JSX comment between elements", (t) => {
  const r = repo(t);
  r.write("src/a.jsx", "export const A = () => <div><a/>{/* x\n y */}<b/></div>;\n");
  const base = r.commit("base");
  r.write("src/a.jsx", "export const A = () => <div><a/><b/></div>;\n");
  const result = r.run(["--same-code", "--base", base]);
  assert.equal(result.status, 0, result.stderr);
});

test("--same-code rejects a file replaced by a symlink, renamed or moved out of the check", (t) => {
  const r = repo(t);
  r.write("src/a.js", "export const a = 1;\n");
  r.write("src/b.js", "export const b = 1;\n");
  const base = r.commit("base");
  rmSync(join(r.root, "src/b.js"));
  symlinkSync("a.js", join(r.root, "src/b.js"));
  const linked = r.run(["--same-code", "--base", base]);
  assert.equal(linked.status, 1);
  assert.match(linked.stderr, /^src\/b\.js: /m);
  rmSync(join(r.root, "src/b.js"));
  r.write("src/b.js", "export const b = 1;\n");
  r.git("mv", "src/a.js", "src/z.js");
  const renamed = r.run(["--same-code", "--base", base]);
  assert.equal(renamed.status, 1);
  assert.match(renamed.stderr, /^src\/z\.js: /m);
  mkdirSync(join(r.root, "dist"));
  r.git("mv", "src/z.js", "dist/a.js");
  const moved = r.run(["--same-code", "--base", base]);
  assert.equal(moved.status, 1);
  assert.match(moved.stderr, /^dist\/a\.js: /m);
});

test("--same-code rejects a deleted or added source file", (t) => {
  const r = repo(t);
  r.write("src/a.js", "export const a = 1;\n");
  r.write("src/gone.js", "export const g = 1;\n");
  const base = r.commit("base");
  rmSync(join(r.root, "src/gone.js"));
  const deleted = r.run(["--same-code", "--base", base]);
  assert.equal(deleted.status, 1);
  assert.match(deleted.stderr, /^src\/gone\.js: /m);
  r.write("src/gone.js", "export const g = 1;\n");
  r.write("src/new.js", "export const n = 1;\n");
  const added = r.run(["--same-code", "--base", base]);
  assert.equal(added.status, 1);
  assert.match(added.stderr, /^src\/new\.js: /m);
});

test("--same-code rejects a change under an excluded path, of any type, even one that does not parse", (t) => {
  const r = repo(t);
  r.write("src/a.js", "// note\nexport const a = 1;\n");
  r.write("dist/a.js", "export const d = 1;\n");
  r.write("dist/style.css", "a { color: red; }\n");
  r.write("dist/bad.js", "export const b = 1;\n");
  const base = r.commit("base");
  r.write("src/a.js", "export const a = 1;\n");
  assert.equal(r.run(["--same-code", "--base", base]).status, 0);
  r.write("dist/a.js", "export const d = 2;\n");
  const edited = r.run(["--same-code", "--base", base]);
  assert.equal(edited.status, 1);
  assert.match(edited.stderr, /^dist\/a\.js: changed under an excluded path$/m);
  r.write("dist/a.js", "export const d = 1;\n");
  r.write("dist/style.css", "a { color: blue; }\n");
  const styled = r.run(["--same-code", "--base", base]);
  assert.equal(styled.status, 1);
  assert.match(styled.stderr, /^dist\/style\.css: changed under an excluded path$/m);
  r.write("dist/style.css", "a { color: red; }\n");
  r.write("dist/bad.js", "export const = ;\n");
  const unparsable = r.run(["--same-code", "--base", base]);
  assert.equal(unparsable.status, 1, unparsable.stderr);
  assert.match(unparsable.stderr, /^dist\/bad\.js: changed under an excluded path$/m);
  r.write("dist/bad.js", "export const b = 1;\n");
  r.write("dist/new.js", "export const n = 1;\n");
  r.git("add", "dist/new.js");
  const added = r.run(["--same-code", "--base", base]);
  assert.equal(added.status, 1);
  assert.match(added.stderr, /^dist\/new\.js: changed under an excluded path$/m);
  r.git("rm", "-q", "--cached", "dist/new.js");
  assert.equal(r.run(["--same-code", "--base", base]).status, 0);
});

test("--same-code ignores untracked files under an excluded path", (t) => {
  const r = repo(t);
  r.write("src/a.js", "export const a = 1;\n");
  const base = r.commit("base");
  r.write("dist/out.js", "export const o = 1;\n");
  assert.equal(r.run(["--same-code", "--base", base]).status, 0);
});

test("--same-code names both sides of a move across an excluded path once", (t) => {
  const r = repo(t);
  r.write("dist/a.js", "export const a = 1;\n");
  r.write("src/s.js", "export const s = 1;\n");
  const base = r.commit("base");
  r.git("mv", "dist/a.js", "src/a2.js");
  const out = r.run(["--same-code", "--base", base]);
  assert.equal(out.status, 1);
  assert.match(out.stderr, /^dist\/a\.js: changed under an excluded path$/m);
  assert.doesNotMatch(out.stderr, /^src\/a2\.js: /m);
  r.git("mv", "src/a2.js", "dist/a.js");
  r.git("mv", "src/s.js", "dist/s.js");
  const into = r.run(["--same-code", "--base", base]);
  assert.equal(into.status, 1);
  assert.match(into.stderr, /^dist\/s\.js: changed under an excluded path$/m);
  assert.match(into.stderr, /^dist\/s\.js: moved from src\/s\.js$/m);
});

test("the inventory skips excluded paths", (t) => {
  const r = repo(t);
  r.write("src/a.js", "// kept\nexport const a = 1;\n");
  r.write("dist/a.js", "// skipped\nexport const d = 1;\n");
  r.commit("base");
  const result = r.run(["--inventory"]);
  assert.equal(result.status, 0, result.stderr);
  const files = result.stdout.trim().split("\n").map((line) => JSON.parse(line).file);
  assert.deepEqual(files, ["src/a.js"]);
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
