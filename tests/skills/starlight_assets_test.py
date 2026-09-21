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
            self.assertEqual(package["dependencies"]["astro"], "7.0.7")
            self.assertEqual(package["dependencies"]["@astrojs/starlight"], "0.41.3")
            self.assertEqual(package["devDependencies"]["typescript"], "6.0.3")
            self.assertEqual(package["packageManager"], "pnpm@11.21.0")
            self.assertTrue((folder / "astro.config.mjs").is_file())
            self.assertTrue((folder / "src/content.config.ts").is_file())
            self.assertTrue((folder / "src/content/docs/index.mdx").is_file())

    def test_assets_do_not_impose_delivery_or_fixed_locale(self) -> None:
        for profile in ("user-manual", "spec-site"):
            text = (ASSETS / profile / "astro.config.mjs").read_text(encoding="utf-8")
            self.assertNotIn("middleware", text)
            self.assertNotIn("locale", text)
            self.assertNotIn("lefthook", text)
