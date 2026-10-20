// Local production-bundle probe with the viewer's unchanged CSP.

import { readFile } from "node:fs/promises";
import { createServer } from "node:http";
import { fileURLToPath } from "node:url";

const output = new URL("../../static/app/", import.meta.url);
const manifest = JSON.parse(
  await readFile(new URL("bundle-manifest.json", output), "utf8"),
);
const assets = new Map(
  manifest.files.map((file) => [
    `/assets/app/${file.path}`,
    { url: new URL(file.path, output), mime: file.mime },
  ]),
);
assets.set("/", {
  url: new URL("index.html", output),
  mime: "text/html; charset=utf-8",
});
for (const file of ["walkthrough.mjs", "walkthrough.css"])
  assets.set(`/assets/${file}`, {
    url: new URL(`../../static/${file}`, import.meta.url),
    mime: file.endsWith(".css")
      ? "text/css; charset=utf-8"
      : "text/javascript; charset=utf-8",
  });
const server = createServer(async (req, res) => {
  res.setHeader(
    "Content-Security-Policy",
    "default-src 'self'; img-src 'self'; media-src 'self'; object-src 'none'; frame-ancestors 'none'; base-uri 'none'",
  );
  const asset = assets.get(req.url?.split("?")[0]);
  if (!asset) {
    res.writeHead(404);
    res.end();
    return;
  }
  res.setHeader("Content-Type", asset.mime);
  res.end(
    req.method === "HEAD"
      ? undefined
      : await readFile(fileURLToPath(asset.url)),
  );
});
server.listen(5190, "127.0.0.1", () =>
  console.log("CSP probe http://127.0.0.1:5190"),
);
