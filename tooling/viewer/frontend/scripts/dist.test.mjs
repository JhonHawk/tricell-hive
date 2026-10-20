import assert from "node:assert/strict";
import { mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { generate, verify } from "./dist.mjs";

test("manifest detects source changes, asset tampering and undeclared assets", async () => {
  const root = await mkdtemp(join(tmpdir(), "hive-viewer-dist-"));
  const source = join(root, "frontend");
  const output = join(root, "app");
  try {
    await mkdir(source);
    await mkdir(output);
    await writeFile(join(source, "package.json"), "{}\n");
    await writeFile(
      join(output, "index.html"),
      '<script type="module" src="/assets/app/main.js"></script>',
    );
    await writeFile(join(output, "main.js"), "export {};");
    await generate(source, output);
    await verify(source, output);
    await writeFile(join(source, "new.ts"), "export {};");
    await assert.rejects(verify(source, output), /sources/);
    await generate(source, output);
    await writeFile(join(output, "main.js"), "tampered");
    await assert.rejects(verify(source, output), /hash/);
    await generate(source, output);
    await writeFile(join(output, "extra.js"), "extra");
    await assert.rejects(verify(source, output), /undeclared/);
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});

test("private environment files do not enter source digest", async () => {
  const root = await mkdtemp(join(tmpdir(), "hive-viewer-private-"));
  try {
    const source = join(root, "frontend");
    const output = join(root, "app");
    await mkdir(source);
    await mkdir(output);
    await writeFile(join(source, "package.json"), "{}");
    await writeFile(join(output, "index.html"), "<html></html>");
    await generate(source, output);
    await writeFile(join(source, ".env.local"), "PRIVATE=not-a-real-value");
    await verify(source, output);
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});

test("repeated generation produces identical manifest bytes", async () => {
  const root = await mkdtemp(join(tmpdir(), "hive-viewer-repeat-"));
  try {
    const source = join(root, "frontend");
    const output = join(root, "app");
    await mkdir(source);
    await mkdir(output);
    await writeFile(join(source, "package.json"), "{}");
    await writeFile(join(output, "index.html"), "<html></html>");
    await generate(source, output);
    const first = await readFile(join(output, "bundle-manifest.json"));
    await generate(source, output);
    assert.deepEqual(
      await readFile(join(output, "bundle-manifest.json")),
      first,
    );
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});
