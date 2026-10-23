import { readFileSync } from "node:fs";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import tailwindcss from "@tailwindcss/vite";
import react from "@vitejs/plugin-react";
import { defineConfig } from "vite";
export default defineConfig({
  base: "/assets/app/",
  plugins: [
    react(),
    tailwindcss(),
    {
      name: "shared-walkthrough-development",
      apply: "serve",
      configureServer(server) {
        server.middlewares.use(async (req, res, next) => {
          const asset = req.url
            ?.split("?")[0]
            ?.replace("/assets/app/assets/", "/assets/");
          if (
            asset !== "/assets/walkthrough.mjs" &&
            asset !== "/assets/walkthrough.css" &&
            asset !== "/assets/hive-mark.svg"
          )
            return next();
          res.setHeader(
            "Content-Type",
            asset.endsWith(".svg")
              ? "image/svg+xml"
              : asset.endsWith(".css")
                ? "text/css"
                : "text/javascript",
          );
          res.end(
            await readFile(
              fileURLToPath(
                new URL(`../static/${asset.slice(8)}`, import.meta.url),
              ),
            ),
          );
        });
      },
    },
  ],
  resolve: { alias: { "@": fileURLToPath(new URL("./src", import.meta.url)) } },
  build: {
    outDir: "../static/app",
    emptyOutDir: true,
    assetsInlineLimit: 0,
    sourcemap: false,
    rolldownOptions: {
      output: {
        comments: { legal: true },
        banner: `/*!
${readFileSync(new URL("./THIRD_PARTY_NOTICES.md", import.meta.url), "utf8").replaceAll("*/", "* /")}
*/`,
      },
    },
  },
});
