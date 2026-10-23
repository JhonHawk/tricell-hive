import { defineCollection } from "astro:content";
import { readFile } from "node:fs/promises";
import { fileURLToPath } from "node:url";
import { docsLoader } from "@astrojs/starlight/loaders";
import { docsSchema } from "@astrojs/starlight/schema";
import type { Loader } from "astro/loaders";

const docsWithChangelog: Loader = {
  name: "hive-docs-with-changelog",
  async load(context) {
    await docsLoader().load(context);
    const source = new URL("../CHANGELOG.md", context.config.root);
    const sourcePath = fileURLToPath(source);

    async function loadChangelog() {
      const markdown = await readFile(source, "utf8");
      const body =
        "Unreleased contains pending changes. Dated entries are published releases.\n\n" +
        markdown
          .replace(/^# Changelog\r?\n/, "")
          .replace(
            /^## (\[\d+\.\d+\.\d+\]) - (\d{4}-\d{2}-\d{2})$/gm,
            '## $1 - <time datetime="$2" class="release-date">$2</time>',
          )
          .replaceAll(
            "](VERSION)",
            "](https://github.com/JhonHawk/tricell-hive/blob/master/VERSION)",
          )
          .replaceAll(
            "](CONTRIBUTING.md#releases-and-version-numbers)",
            "](https://github.com/JhonHawk/tricell-hive/blob/master/CONTRIBUTING.md#releases-and-version-numbers)",
          );
      const data = await context.parseData({
        id: "changelog",
        data: {
          title: "Changelog",
          description:
            "Published Hive releases and pending changes, from the project's canonical changelog.",
        },
      });
      context.store.set({
        id: "changelog",
        data,
        body,
        rendered: await context.renderMarkdown(body, { fileURL: source }),
        digest: context.generateDigest(body),
      });
    }

    await loadChangelog();
    context.watcher?.add(sourcePath);
    context.watcher?.on("change", async (changedPath) => {
      if (changedPath === sourcePath) await loadChangelog();
    });
  },
};

export const collections = {
  docs: defineCollection({ loader: docsWithChangelog, schema: docsSchema() }),
};
