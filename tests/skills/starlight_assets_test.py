#!/usr/bin/env python3
"""Structural invariants for the two small, runnable Starlight assets."""

from __future__ import annotations

import json
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
ASSETS = ROOT / "content/skills/starlight-docs-site/assets"


class StarlightAssetsTest(unittest.TestCase):
    def test_profiles_have_runnable_minimum(self) -> None:
        for profile in ("user-manual", "spec-site"):
            folder = ASSETS / profile
            package = json.loads((folder / "package.json").read_text(encoding="utf-8"))
            self.assertEqual(package["dependencies"]["astro"], "7.3.5")
            self.assertEqual(package["dependencies"]["@astrojs/starlight"], "0.42.4")
            self.assertEqual(package["dependencies"]["sharp"], "0.35.5")
            self.assertEqual(package["devDependencies"]["@astrojs/check"], "0.9.10")
            self.assertEqual(package["devDependencies"]["@biomejs/biome"], "2.5.14")
            self.assertEqual(package["devDependencies"]["typescript"], "6.0.3")
            self.assertEqual(package["packageManager"], "pnpm@12.8.1")
            self.assertTrue((folder / "astro.config.mjs").is_file())
            self.assertTrue((folder / "src/content.config.ts").is_file())
            self.assertTrue((folder / "src/content/docs/index.mdx").is_file())
            workspace = (folder / "pnpm-workspace.yaml").read_text(encoding="utf-8")
            self.assertIn("allowBuilds:\n  esbuild: false\n", workspace)
            biome = json.loads((folder / "biome.json").read_text(encoding="utf-8"))
            self.assertIn("!!**/dist", biome["files"]["includes"])
            self.assertIn("!!**/.astro", biome["files"]["includes"])

    def test_assets_do_not_impose_delivery_or_fixed_locale(self) -> None:
        for profile in ("user-manual", "spec-site"):
            text = (ASSETS / profile / "astro.config.mjs").read_text(encoding="utf-8")
            self.assertNotIn("middleware", text)
            self.assertNotIn("locale", text)
            self.assertNotIn("lefthook", text)


if __name__ == "__main__":
    unittest.main()
