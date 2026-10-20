import { createHash } from "node:crypto";
import { readdir, readFile, rename, writeFile } from "node:fs/promises";
import { dirname, extname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const manifestName = "bundle-manifest.json";
const excluded = new Set([
  "node_modules",
  ".git",
  "coverage",
  "playwright-report",
  "test-results",
  ".DS_Store",
]);
const mime = {
  ".html": "text/html; charset=utf-8",
  ".js": "text/javascript; charset=utf-8",
  ".css": "text/css; charset=utf-8",
};
const sha = (bytes) => createHash("sha256").update(bytes).digest("hex");
async function list(root, source = false, prefix = "") {
  const entries = await readdir(join(root, prefix), { withFileTypes: true });
  const files = [];
  for (const entry of entries) {
    if (
      source &&
      (excluded.has(entry.name) ||
        entry.name.startsWith(".env") ||
        entry.name.endsWith(".local"))
    )
      continue;
    const path = prefix ? `${prefix}/${entry.name}` : entry.name;
    if (entry.isSymbolicLink())
      throw new Error(`Symbolic link is not allowed: ${path}`);
    if (entry.isDirectory()) files.push(...(await list(root, source, path)));
    else if (entry.isFile()) files.push(path);
  }
  return files.sort();
}
export async function sourceDigest(source) {
  const lines = [];
  for (const path of await list(source, true))
    lines.push(`${path}\0${sha(await readFile(join(source, path)))}\n`);
  return sha(lines.join(""));
}
export async function generate(source, output) {
  const files = [];
  for (const path of await list(output)) {
    if (path === manifestName) continue;
    const type = mime[extname(path)];
    if (!type) throw new Error(`Unsupported asset: ${path}`);
    files.push({
      path,
      sha256: sha(await readFile(join(output, path))),
      mime: type,
    });
  }
  if (!files.some((file) => file.path === "index.html"))
    throw new Error("Missing entrypoint");
  const manifest = {
    version: 1,
    sourceDigest: await sourceDigest(source),
    entrypoint: "index.html",
    files,
  };
  await writeFile(
    join(output, `${manifestName}.tmp`),
    `${JSON.stringify(manifest, null, 2)}\n`,
  );
  await rename(join(output, `${manifestName}.tmp`), join(output, manifestName));
}
export async function verify(source, output) {
  const manifest = JSON.parse(
    await readFile(join(output, manifestName), "utf8"),
  );
  if (manifest.version !== 1 || manifest.entrypoint !== "index.html")
    throw new Error("Invalid manifest");
  if (manifest.sourceDigest !== (await sourceDigest(source)))
    throw new Error("Frontend sources are stale; rebuild the bundle");
  if (!manifest.files.some((file) => file.path === manifest.entrypoint))
    throw new Error("Missing entrypoint");
  const actual = (await list(output)).filter((path) => path !== manifestName);
  if (
    JSON.stringify(actual) !==
    JSON.stringify(manifest.files.map((file) => file.path))
  )
    throw new Error("Missing or undeclared output files");
  for (const file of manifest.files) {
    if (
      !mime[extname(file.path)] ||
      file.mime !== mime[extname(file.path)] ||
      file.path.includes("..") ||
      file.path.startsWith("/")
    )
      throw new Error("Invalid asset path or MIME");
    if (file.sha256 !== sha(await readFile(join(output, file.path))))
      throw new Error(`Asset hash mismatch: ${file.path}`);
  }
}
if (
  process.argv[1] &&
  resolve(process.argv[1]) === fileURLToPath(import.meta.url)
) {
  const source = resolve(dirname(fileURLToPath(import.meta.url)), "..");
  const output = resolve(source, "../static/app");
  const command = process.argv[2];
  if (command === "generate") await generate(source, output);
  else if (command === "verify") await verify(source, output);
  else throw new Error("Expected generate or verify");
  console.log(`Bundle ${command} passed`);
}
